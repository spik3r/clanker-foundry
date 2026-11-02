package impact

import (
	"reflect"
	"sort"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/canonical"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/model"
)

type elementKey struct{ kind, id string }
type fields map[string]any

// These are structural values, never an expression evaluator. The finite
// vocabulary is part of this review profile; Normalize owns model semantics.
func elements(m model.Model) map[elementKey]fields {
	out := map[elementKey]fields{{"model", ""}: {"id": m.ID, "initial": m.Initial}}
	for _, s := range m.States {
		out[elementKey{"state", s.ID}] = fields{"label": s.Label, "terminal": s.Terminal}
	}
	for _, r := range m.Requirements {
		out[elementKey{"requirement", r.ID}] = fields{"text": r.Text}
	}
	for _, g := range m.Guards {
		out[elementKey{"guard", g.ID}] = fields{"requirement": g.Requirement, "expr": g.Expr}
	}
	for _, t := range m.Transitions {
		out[elementKey{"transition", t.ID}] = fields{"from": t.From, "trigger": t.Trigger, "to": t.To, "guards": t.Guards, "effects": t.Effects}
	}
	return out
}

// Compare returns a content review aid, not an equivalence or execution result.
// Both normalized graphs are retained for link projection, including removals.
func Compare(before, after model.Model, bindings []Binding) (Report, error) {
	b, err := model.Normalize(before)
	if err != nil {
		return Report{}, ErrInvalidInput
	}
	a, err := model.Normalize(after)
	if err != nil {
		return Report{}, ErrInvalidInput
	}
	if err := validateBindings(bindings); err != nil {
		return Report{}, err
	}
	bd, err := canonical.Digest(b)
	if err != nil {
		return Report{}, ErrInvalidInput
	}
	ad, err := canonical.Digest(a)
	if err != nil {
		return Report{}, ErrInvalidInput
	}
	r := Report{ReportVersion: ReportVersion, ProfileVersion: model.ProfileVersion, Mode: Mode, Authority: Authority, BeforeDigest: bd, AfterDigest: ad, Changed: bd != ad, ReviewRequired: bd != ad, Changes: []Change{}, LinkedTransitions: []LinkedTransition{}, Bindings: []BindingResult{}}
	be, ae := elements(b), elements(a)
	keys := make([]elementKey, 0, len(be)+len(ae))
	for k := range be {
		keys = append(keys, k)
	}
	for k := range ae {
		if _, exists := be[k]; !exists {
			keys = append(keys, k)
		}
	}
	sort.Slice(keys, func(i, j int) bool {
		if keys[i].kind != keys[j].kind {
			return keys[i].kind < keys[j].kind
		}
		return keys[i].id < keys[j].id
	})
	changed := map[elementKey]bool{}
	for _, k := range keys {
		old, had := be[k]
		next, has := ae[k]
		c := Change{Kind: k.kind, ID: k.id, Fields: []string{}}
		switch {
		case !had:
			c.Type = "added"
		case !has:
			c.Type = "removed"
		default:
			for field, value := range old {
				if !reflect.DeepEqual(value, next[field]) {
					c.Fields = append(c.Fields, field)
				}
			}
			if len(c.Fields) == 0 {
				continue
			}
			sort.Strings(c.Fields)
			c.Type = "modified"
		}
		r.Changes = append(r.Changes, c)
		changed[k] = true
	}
	// Refuse a digest-only difference with no represented change. Future schema
	// changes still require review of this fixed field vocabulary; this check
	// does not prove that every concurrent future-field change was represented.
	if r.Changed != (len(r.Changes) > 0) {
		return Report{}, ErrInvalidInput
	}
	r.LinkedTransitions = append(r.LinkedTransitions, links("before", b, changed)...)
	r.LinkedTransitions = append(r.LinkedTransitions, links("after", a, changed)...)
	ordered := append([]Binding(nil), bindings...)
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].ID < ordered[j].ID })
	for _, binding := range ordered {
		status := "matches-neither"
		if binding.ModelDigest == ad {
			status = "matches-after"
		} else if binding.ModelDigest == bd {
			status = "matches-before-only"
		}
		r.Bindings = append(r.Bindings, BindingResult{binding.ID, binding.ModelDigest, status})
	}
	return r, nil
}

func links(side string, m model.Model, changed map[elementKey]bool) []LinkedTransition {
	guards := map[string]model.Guard{}
	for _, g := range m.Guards {
		guards[g.ID] = g
	}
	out := []LinkedTransition{}
	// Normalize already sorts transitions. No state-graph traversal or cycle
	// closure is appropriate for this direct declared-link projection.
	for _, t := range m.Transitions {
		causes := map[string]bool{}
		add := func(kind, id string) {
			if changed[elementKey{kind, id}] {
				causes[kind+":"+id] = true
			}
		}
		add("model", "")
		add("transition", t.ID)
		add("state", t.From)
		add("state", t.To)
		for _, id := range t.Guards {
			add("guard", id)
			add("requirement", guards[id].Requirement)
		}
		if len(causes) == 0 {
			continue
		}
		list := make([]string, 0, len(causes))
		for cause := range causes {
			list = append(list, cause)
		}
		sort.Strings(list)
		out = append(out, LinkedTransition{side, t.ID, list})
	}
	return out
}
