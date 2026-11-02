package modelmerge

import (
	"reflect"
	"sort"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/canonical"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/jsonformat"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/model"
)

// Merge compares complete normalized records. The caller supplies the common
// base; this function neither inspects ancestry nor authorizes the candidate.
func Merge(base, left, right model.Model) (Result, error) {
	inputs := [3]model.Model{base, left, right}
	var digests [3]string
	for i := range inputs {
		var err error
		inputs[i], digests[i], err = prepare(inputs[i])
		if err != nil {
			return Result{}, ErrInvalid
		}
	}
	base, left, right = inputs[0], inputs[1], inputs[2]
	result := Result{
		MergeVersion: Version, ProfileVersion: model.ProfileVersion,
		Mode: "example-only", Authority: "none",
		BaseDigest: digests[0], LeftDigest: digests[1], RightDigest: digests[2],
		Status: Conflicted, Conflicts: []Conflict{},
	}
	candidate := model.Model{Version: base.Version}
	root := func(field, b, l, r string) string {
		chosen, ok := choose(unit[string]{b, true}, unit[string]{l, true}, unit[string]{r, true})
		if !ok {
			result.Conflicts = append(result.Conflicts, Conflict{Kind: "model", Field: field, Reason: "change-change"})
		}
		return chosen.value
	}
	candidate.ID = root("id", base.ID, left.ID, right.ID)
	candidate.Initial = root("initial", base.Initial, left.Initial, right.Initial)
	var conflicts []Conflict
	candidate.States, conflicts = records("state", base.States, left.States, right.States, func(v model.State) string { return v.ID })
	result.Conflicts = append(result.Conflicts, conflicts...)
	candidate.Requirements, conflicts = records("requirement", base.Requirements, left.Requirements, right.Requirements, func(v model.Requirement) string { return v.ID })
	result.Conflicts = append(result.Conflicts, conflicts...)
	candidate.Guards, conflicts = records("guard", base.Guards, left.Guards, right.Guards, func(v model.Guard) string { return v.ID })
	result.Conflicts = append(result.Conflicts, conflicts...)
	candidate.Transitions, conflicts = records("transition", base.Transitions, left.Transitions, right.Transitions, func(v model.Transition) string { return v.ID })
	result.Conflicts = append(result.Conflicts, conflicts...)
	if len(result.Conflicts) != 0 {
		sort.Slice(result.Conflicts, func(i, j int) bool {
			a, b := result.Conflicts[i], result.Conflicts[j]
			if a.Kind != b.Kind {
				return a.Kind < b.Kind
			}
			if a.ID != b.ID {
				return a.ID < b.ID
			}
			return a.Field < b.Field
		})
		return result, nil
	}
	// Individually valid edits need not remain valid when combined. Validate the
	// whole private candidate; never repair links or leak a partial proposal.
	normalized, digest, err := prepare(candidate)
	if err != nil {
		result.Status = InvalidCombination
		return result, nil
	}
	result.Status = Merged
	result.Candidate = &normalized
	result.CandidateDigest = &digest
	return result, nil
}

func prepare(input model.Model) (model.Model, string, error) {
	normalized, err := model.Normalize(input)
	if err != nil {
		return model.Model{}, "", ErrInvalid
	}
	if _, err = jsonformat.Encode(normalized); err != nil {
		return model.Model{}, "", ErrInvalid
	}
	digest, err := canonical.Digest(normalized)
	if err != nil {
		return model.Model{}, "", ErrInvalid
	}
	return normalized, digest, nil
}

// Presence participates in equality, so removal cannot masquerade as a zero
// record. Only private normalized inputs reach this helper.
type unit[T any] struct {
	value   T
	present bool
}

func equal[T any](a, b unit[T]) bool {
	return a.present == b.present && (!a.present || reflect.DeepEqual(a.value, b.value))
}
func choose[T any](base, left, right unit[T]) (unit[T], bool) {
	switch {
	case equal(left, right):
		return left, true
	case equal(left, base):
		return right, true
	case equal(right, base):
		return left, true
	default:
		return unit[T]{}, false
	}
}

func records[T any](kind string, base, left, right []T, id func(T) string) ([]T, []Conflict) {
	index := func(values []T) map[string]unit[T] {
		out := make(map[string]unit[T], len(values))
		for _, value := range values {
			out[id(value)] = unit[T]{value, true}
		}
		return out
	}
	b, l, r := index(base), index(left), index(right)
	ids := make(map[string]bool, len(b)+len(l)+len(r))
	for _, input := range []map[string]unit[T]{b, l, r} {
		for key := range input {
			ids[key] = true
		}
	}
	keys := make([]string, 0, len(ids))
	for key := range ids {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	out := []T{}
	conflicts := []Conflict{}
	for _, key := range keys {
		value, ok := choose(b[key], l[key], r[key])
		if !ok {
			reason := "change-change"
			if !b[key].present {
				reason = "add-add"
			} else if !l[key].present || !r[key].present {
				reason = "remove-change"
			}
			conflicts = append(conflicts, Conflict{Kind: kind, ID: key, Reason: reason})
		} else if value.present {
			out = append(out, value.value)
		}
	}
	return out, conflicts
}
