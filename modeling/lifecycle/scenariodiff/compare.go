package scenariodiff

import (
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/model"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/scenarios"
	"reflect"
	"slices"
)

// Compare inventories literal edits after validating both complete model/suite
// pairs. It does not execute, verify expected behavior, or refresh expectations.
// Inputs must remain immutable during the call; output owns all mutable storage.
func Compare(beforeModel model.Model, before scenarios.Suite, afterModel model.Model, after scenarios.Suite) (Report, error) {
	br, bd, err := validator(beforeModel)
	if err != nil {
		return Report{}, ErrInvalid
	}
	bs, err := br.Digest(before)
	if err != nil {
		return Report{}, ErrInvalid
	}
	ar, ad, err := validator(afterModel)
	if err != nil {
		return Report{}, ErrInvalid
	}
	as, err := ar.Digest(after)
	if err != nil {
		return Report{}, ErrInvalid
	}
	out := Report{DiffVersion: Version, ProfileVersion: model.ProfileVersion, Mode: "example-only", Authority: "none",
		BeforeModelDigest: bd, AfterModelDigest: ad, BeforeSuiteDigest: bs, AfterSuiteDigest: as,
		ModelChanged: bd != ad, SuiteChanged: bs != as, ReviewRequired: bd != ad || bs != as,
		SuiteFields: []string{}, Cases: []CaseChange{}, Steps: []StepChange{},
		CaseOrder: Order{Before: caseIDs(before.Cases), After: caseIDs(after.Cases)}}
	out.CaseOrder.Changed = !slices.Equal(out.CaseOrder.Before, out.CaseOrder.After)
	if before.ID != after.ID {
		out.SuiteFields = append(out.SuiteFields, "id")
	}
	if before.ModelDigest != after.ModelDigest {
		out.SuiteFields = append(out.SuiteFields, "modelDigest")
	}
	bindex := index(before.Cases, func(c scenarios.Case) string { return c.ID })
	aindex := index(after.Cases, func(c scenarios.Case) string { return c.ID })
	for _, id := range keys(bindex, aindex) {
		b, hasBefore := bindex[id]
		a, hasAfter := aindex[id]
		change := CaseChange{ID: id, Kind: Modified, Fields: []string{}, BeforeSteps: stepIDs(b.Steps), AfterSteps: stepIDs(a.Steps)}
		switch {
		case !hasBefore:
			change.Kind = Added
		case !hasAfter:
			change.Kind = Removed
		default:
			if !reflect.DeepEqual(b.Initial, a.Initial) {
				change.Fields = append(change.Fields, "initial")
			}
			if !reflect.DeepEqual(b.Steps, a.Steps) {
				change.Fields = append(change.Fields, "steps")
			}
		}
		if change.Kind != Modified || len(change.Fields) != 0 {
			out.Cases = append(out.Cases, change)
		}
		out.Steps = append(out.Steps, compareSteps(id, b.Steps, a.Steps)...)
	}
	if _, err := encode(out); err != nil {
		return Report{}, err
	}
	return out, nil
}

func compareSteps(caseID string, before, after []scenarios.Step) []StepChange {
	out := []StepChange{}
	bi := index(before, func(s scenarios.Step) string { return s.ID })
	ai := index(after, func(s scenarios.Step) string { return s.ID })
	for _, id := range keys(bi, ai) {
		b, hasBefore := bi[id]
		a, hasAfter := ai[id]
		change := StepChange{CaseID: caseID, ID: id, Kind: Modified, Fields: []string{}}
		switch {
		case !hasBefore:
			change.Kind = Added
		case !hasAfter:
			change.Kind = Removed
		default:
			change.Fields = stepFields(b, a)
		}
		if change.Kind != Modified || len(change.Fields) != 0 {
			out = append(out, change)
		}
	}
	return out
}

func stepFields(before, after scenarios.Step) []string {
	fields := []string{}
	add := func(name string, changed bool) {
		if changed {
			fields = append(fields, name)
		}
	}
	// Full Review/Task values remain atomic; DeepEqual preserves pointer presence,
	// nil/empty acceptance and ordered slices rather than coercing fixture values.
	add("command.actor", before.Command.Actor != after.Command.Actor)
	add("command.expectedRevision", before.Command.ExpectedRevision != after.Command.ExpectedRevision)
	add("command.review", !reflect.DeepEqual(before.Command.Review, after.Command.Review))
	add("command.type", before.Command.Type != after.Command.Type)
	add("expected.ok", before.Expected.OK != after.Expected.OK)
	add("expected.refusals", !reflect.DeepEqual(before.Expected.Refusals, after.Expected.Refusals))
	add("expected.task", !reflect.DeepEqual(before.Expected.Task, after.Expected.Task))
	add("expected.transitionId", before.Expected.TransitionID != after.Expected.TransitionID)
	return fields
}

func caseIDs(cases []scenarios.Case) []string {
	out := make([]string, len(cases))
	for i, c := range cases {
		out[i] = c.ID
	}
	return out
}
func stepIDs(steps []scenarios.Step) []string {
	out := make([]string, len(steps))
	for i, s := range steps {
		out[i] = s.ID
	}
	return out
}
func index[T any](records []T, id func(T) string) map[string]T {
	out := make(map[string]T, len(records))
	for _, record := range records {
		out[id(record)] = record
	}
	return out
}
func keys[T any](before, after map[string]T) []string {
	out := make([]string, 0, len(before)+len(after))
	for id := range before {
		out = append(out, id)
	}
	for id := range after {
		if _, ok := before[id]; !ok {
			out = append(out, id)
		}
	}
	slices.Sort(out)
	return out
}
