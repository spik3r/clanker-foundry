package modelpatch

import (
	"errors"
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/canonical"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/model"
)

func assertError(t *testing.T, b model.Model, p Patch, want error) {
	t.Helper()
	r, e := Plan(b, p)
	if !errors.Is(e, want) || !reflect.DeepEqual(r, Result{}) {
		t.Fatalf("want zero result and %v; got %+v, %v", want, r, e)
	}
}
func snapshotModel(m model.Model) model.Model {
	n := m
	n.States = append([]model.State{}, m.States...)
	n.Requirements = append([]model.Requirement{}, m.Requirements...)
	n.Guards = append([]model.Guard{}, m.Guards...)
	n.Transitions = append([]model.Transition{}, m.Transitions...)
	for i := range n.Guards {
		n.Guards[i].Expr = snapshotExpr(m.Guards[i].Expr)
	}
	for i := range n.Transitions {
		n.Transitions[i].Guards = append([]string{}, m.Transitions[i].Guards...)
		n.Transitions[i].Effects = append([]model.Effect{}, m.Transitions[i].Effects...)
	}
	return n
}
func TestPlanFrozenIndependentModelsAndReports(t *testing.T) {
	for _, f := range frozenFixtures() {
		t.Run(f.Name, func(t *testing.T) {
			b := snapshotModel(f.Before)
			pBytes, e := patchBytes(f.Patch)
			if e != nil {
				t.Fatal(e)
			}
			got, e := Plan(f.Before, f.Patch)
			if e != nil {
				t.Fatal(e)
			}
			if !reflect.DeepEqual(got.After, f.After) || got.BeforeDigest != frozenBeforeDigest || got.AfterDigest != f.AfterDigest {
				t.Fatalf("candidate differs from frozen literal\ngot %+v\nwant %+v", got.After, f.After)
			}
			if !reflect.DeepEqual(got.Review.Changes, f.Changes) || !reflect.DeepEqual(got.Review.LinkedTransitions, f.Links) {
				t.Fatal("review differs from frozen changes/links", got.Review)
			}
			if got.ReportVersion != ReportVersion || got.ProfileVersion != ProfileVersion || got.Mode != "example-only" || got.Authority != "none" || got.Review.BeforeDigest != frozenBeforeDigest || got.Review.AfterDigest != f.AfterDigest || got.Review.Mode != "example-only" || got.Review.Authority != "none" || got.Review.RequestDigest != nil || got.Review.BeforeSourceDigest != nil || got.Review.AfterSourceDigest != nil || got.Review.Bindings == nil || len(got.Review.Bindings) != 0 {
				t.Fatal("incorrect result envelope")
			}
			if got.Review.Changed != (frozenBeforeDigest != f.AfterDigest) || got.Review.ReviewRequired != (frozenBeforeDigest != f.AfterDigest) {
				t.Fatal("identity flag")
			}
			afterPatch, e := patchBytes(f.Patch)
			if e != nil || !reflect.DeepEqual(pBytes, afterPatch) || !reflect.DeepEqual(f.Before, b) {
				t.Fatal("input mutated")
			}
		})
	}
}
func TestFrozenMixedErrorStages(t *testing.T) {
	for _, f := range frozenMixedErrors() {
		t.Run(f.Name, func(t *testing.T) { assertError(t, frozenBefore(), f.Patch, f.Want) })
	}
}
func TestStaleDigestAndNoPartialResult(t *testing.T) {
	p := Patch{Version: 1, ProfileVersion: ProfileVersion, BeforeDigest: frozenStaleDigest, Operations: []Operation{}}
	assertError(t, frozenBefore(), p, ErrPrecondition)
	p.BeforeDigest = frozenBeforeDigest
	p.Operations = []Operation{{Action: Replace, Kind: State, ID: "A", State: &model.State{ID: "A", Label: "Private candidate"}}, {Action: Remove, Kind: State, ID: "Missing"}}
	before := frozenBefore()
	snapshot := snapshotModel(before)
	assertError(t, before, p, ErrConflict)
	if !reflect.DeepEqual(before, snapshot) {
		t.Fatal("partial apply mutated original")
	}
	p.Operations = []Operation{{Action: Remove, Kind: State, ID: "B"}}
	assertError(t, before, p, ErrInvalid)
	// Final-only profile failures become observable after the correct digest.
	for _, f := range frozenMixedErrors() {
		if !errors.Is(f.Want, ErrPrecondition) || f.Name == "stale missing target" {
			continue
		}
		f.Patch.BeforeDigest = frozenBeforeDigest
		assertError(t, before, f.Patch, ErrInvalid)
	}
}
func TestTargetPolicyAndCoordinatedRequirementStateEdits(t *testing.T) {
	p := Patch{Version: 1, ProfileVersion: ProfileVersion, BeforeDigest: frozenBeforeDigest, Operations: []Operation{{Action: Add, Kind: State, ID: "A", State: &model.State{ID: "A", Label: "Already exists"}}}}
	assertError(t, frozenBefore(), p, ErrConflict)
	p.Operations = []Operation{{Action: Remove, Kind: State, ID: "Missing"}}
	assertError(t, frozenBefore(), p, ErrConflict)
	p.Operations = []Operation{{Action: Remove, Kind: State, ID: "B"}, {Action: Add, Kind: State, ID: "B", State: &model.State{ID: "B", Label: "Done", Terminal: true}}}
	assertError(t, frozenBefore(), p, ErrConflict)
	// These full records and expected models are authored explicitly, not derived
	// by applying operations or observing planner results.
	p.Operations = []Operation{
		{Action: Remove, Kind: Requirement, ID: "R"},
		{Action: Add, Kind: Requirement, ID: "R2", Requirement: &model.Requirement{ID: "R2", Text: "Replacement requirement"}},
		{Action: Replace, Kind: Guard, ID: "G_old", Guard: &model.Guard{ID: "G_old", Requirement: "R2", Expr: model.Expr{Kind: "op", Op: "nonempty", Args: []model.Expr{{Kind: "ref", Ref: "task.outcome"}}}}},
	}
	want := model.Model{Version: 1, ID: "M", Initial: "A", States: []model.State{{ID: "A", Label: "Draft"}, {ID: "B", Label: "Done", Terminal: true}}, Requirements: []model.Requirement{{ID: "R2", Text: "Replacement requirement"}}, Guards: []model.Guard{{ID: "G_old", Requirement: "R2", Expr: model.Expr{Kind: "op", Op: "nonempty", Args: []model.Expr{{Kind: "ref", Ref: "task.outcome"}}}}}, Transitions: []model.Transition{{ID: "T1", From: "A", Trigger: "prepare", To: "B", Guards: []string{"G_old"}, Effects: []model.Effect{{Op: model.SetState}, {Op: model.IncrementRevision}}}}}
	r, e := Plan(frozenBefore(), p)
	if e != nil || !reflect.DeepEqual(r.After, want) {
		t.Fatal("explicit requirement repair", e, r.After)
	}
	p.Operations = []Operation{{Action: Remove, Kind: State, ID: "B"}, {Action: Add, Kind: State, ID: "C", State: &model.State{ID: "C", Label: "Done", Terminal: true}}, {Action: Replace, Kind: Transition, ID: "T1", Transition: &model.Transition{ID: "T1", From: "A", Trigger: "prepare", To: "C", Guards: []string{"G_old"}, Effects: []model.Effect{{Op: model.SetState}, {Op: model.IncrementRevision}}}}}
	want = model.Model{Version: 1, ID: "M", Initial: "A", States: []model.State{{ID: "A", Label: "Draft"}, {ID: "C", Label: "Done", Terminal: true}}, Requirements: []model.Requirement{{ID: "R", Text: "Invented requirement"}}, Guards: []model.Guard{{ID: "G_old", Requirement: "R", Expr: model.Expr{Kind: "op", Op: "nonempty", Args: []model.Expr{{Kind: "ref", Ref: "task.outcome"}}}}}, Transitions: []model.Transition{{ID: "T1", From: "A", Trigger: "prepare", To: "C", Guards: []string{"G_old"}, Effects: []model.Effect{{Op: model.SetState}, {Op: model.IncrementRevision}}}}}
	r, e = Plan(frozenBefore(), p)
	if e != nil || !reflect.DeepEqual(r.After, want) {
		t.Fatal("explicit state rename", e, r.After)
	}
	// Distinct targets can commute. No artificial order-dependent result is claimed.
	p.Operations[0], p.Operations[1] = p.Operations[1], p.Operations[0]
	other, e := Plan(frozenBefore(), p)
	if e != nil || !reflect.DeepEqual(r, other) {
		t.Fatal("independent edits did not commute")
	}
}
func TestExpressionBudgetsBeforeCopyAndEncoding(t *testing.T) {
	p := Patch{Version: 1, ProfileVersion: ProfileVersion, BeforeDigest: frozenStaleDigest, Operations: []Operation{}}
	guard := func(e model.Expr) Operation {
		return Operation{Action: Replace, Kind: Guard, ID: "G_old", Guard: &model.Guard{ID: "G_old", Requirement: "R", Expr: e}}
	}
	cyclic := model.Expr{Kind: "op", Op: "all"}
	cyclic.Args = make([]model.Expr, 1)
	cyclic.Args[0] = cyclic
	p.Operations = []Operation{guard(cyclic)}
	assertError(t, frozenBefore(), p, ErrLimit)
	deep := model.Expr{Kind: "op", Op: "nonempty", Args: []model.Expr{{Kind: "ref", Ref: "task.outcome"}}}
	for i := 0; i < MaxExprDepth+1; i++ {
		deep = model.Expr{Kind: "op", Op: "all", Args: []model.Expr{deep}}
	}
	p.Operations = []Operation{guard(deep)}
	assertError(t, frozenBefore(), p, ErrLimit)
	// 11*49=539 nodes, each individual guard well-typed, exceeds aggregate512.
	largeGuard := func(seed int) model.Expr {
		args := []model.Expr{}
		for j := 0; j < 16; j++ {
			a := strings.Repeat("x", 4090) + string(rune('A'+seed)) + string(rune('A'+j))
			args = append(args, model.Expr{Kind: "op", Op: "eq", Args: []model.Expr{{Kind: "string", Value: a}, {Kind: "string", Value: strings.Repeat("y", 4096)}}})
		}
		return model.Expr{Kind: "op", Op: "all", Args: args}
	}
	p.Operations = []Operation{}
	for i := 0; i < 11; i++ {
		id := "G" + string(rune('A'+i))
		p.Operations = append(p.Operations, Operation{Action: Add, Kind: Guard, ID: id, Guard: &model.Guard{ID: id, Requirement: "R", Expr: largeGuard(i)}})
	}
	assertError(t, frozenBefore(), p, ErrLimit)
	// 490 nodes fit aggregate bounds, but the compact DTO exceeds1MiB.
	p.Operations = p.Operations[:10]
	assertError(t, frozenBefore(), p, ErrLimit)
	p.Operations = []Operation{guard(model.Expr{Kind: "op", Op: "all", Args: make([]model.Expr, MaxExprArgs+1)})}
	assertError(t, frozenBefore(), p, ErrLimit)
}
func TestScalarOperationAndFinalCollectionBounds(t *testing.T) {
	p := Patch{Version: 1, ProfileVersion: ProfileVersion, BeforeDigest: frozenStaleDigest, Operations: make([]Operation, MaxOperations+1)}
	assertError(t, frozenBefore(), p, ErrLimit)
	p.Operations = nil
	assertError(t, frozenBefore(), p, ErrInvalid)
	p.Operations = []Operation{{Action: Replace, Kind: State, ID: "A", State: &model.State{ID: "A", Label: strings.Repeat("x", MaxStateLabelBytes+1)}}}
	assertError(t, frozenBefore(), p, ErrLimit)
	p.Operations = []Operation{}
	for i := 0; i < 15; i++ {
		id := "S" + string(rune('A'+i))
		p.Operations = append(p.Operations, Operation{Action: Add, Kind: State, ID: id, State: &model.State{ID: id, Label: "Terminal", Terminal: true}})
	}
	p.BeforeDigest = frozenBeforeDigest
	assertError(t, frozenBefore(), p, ErrInvalid)
	p.Version = 2
	assertError(t, frozenBefore(), p, ErrInvalid)
}
func TestOwnedResultsAndDeterministicConcurrentReplay(t *testing.T) {
	f := frozenFixtures()[0]
	before := snapshotModel(f.Before)
	patchBefore, e := patchBytes(f.Patch)
	if e != nil {
		t.Fatal(e)
	}
	a, e := Plan(f.Before, f.Patch)
	if e != nil {
		t.Fatal(e)
	}
	b, e := Plan(f.Before, f.Patch)
	if e != nil || !reflect.DeepEqual(a, b) {
		t.Fatal("replay differs")
	}
	a.After.Guards[0].Expr.Args[0].Ref = "changed"
	a.After.Transitions[0].Guards[0] = "changed"
	a.After.Transitions[0].Effects[0].Op = "changed"
	a.Review.Changes[0].Fields = append(a.Review.Changes[0].Fields, "changed")
	a.Review.LinkedTransitions[0].Causes[0] = "changed"
	if !reflect.DeepEqual(f.Before, before) || !reflect.DeepEqual(b.After, f.After) {
		t.Fatal("return aliases input/another result")
	}
	patchAfter, e := patchBytes(f.Patch)
	if e != nil || !reflect.DeepEqual(patchBefore, patchAfter) {
		t.Fatal("return aliases patch")
	}
	f.Patch.Operations[1].Guard.Expr.Args[0].Ref = "caller mutation"
	f.Patch.Operations[2].Transition.Guards[0] = "caller mutation"
	if !reflect.DeepEqual(b.After, f.After) {
		t.Fatal("result aliases caller patch")
	}
	f = frozenFixtures()[0]
	var wg sync.WaitGroup
	for i := 0; i < 6; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			r, e := Plan(f.Before, f.Patch)
			if e != nil || !reflect.DeepEqual(r, b) {
				t.Error("concurrent replay differs", e)
			}
		}()
	}
	wg.Wait()
}

func snapshotExpr(e model.Expr) model.Expr {
	d := model.Expr{Kind: e.Kind, Ref: e.Ref, Value: e.Value, Op: e.Op}
	if e.Args != nil {
		d.Args = make([]model.Expr, len(e.Args))
		for i, x := range e.Args {
			d.Args[i] = snapshotExpr(x)
		}
	}
	return d
}

func TestMaximumOperationsAndScalarBoundaries(t *testing.T) {
	before := frozenBefore()
	before.Requirements = []model.Requirement{}
	before.Guards = []model.Guard{}
	before.Transitions[0].Guards = []string{}
	p := Patch{Version: 1, ProfileVersion: ProfileVersion, Operations: []Operation{}}
	for i := 0; i < MaxOperations; i++ {
		suffix := string(rune('A'+i/26)) + string(rune('A'+i%26))
		rid, gid := "R"+suffix, "G"+suffix
		before.Requirements = append(before.Requirements, model.Requirement{ID: rid, Text: "Before requirement"})
		before.Guards = append(before.Guards, model.Guard{ID: gid, Requirement: rid, Expr: model.Expr{Kind: "op", Op: "nonempty", Args: []model.Expr{{Kind: "ref", Ref: "task.outcome"}}}})
		before.Transitions[0].Guards = append(before.Transitions[0].Guards, gid)
		p.Operations = append(p.Operations, Operation{Action: Replace, Kind: Requirement, ID: rid, Requirement: &model.Requirement{ID: rid, Text: "After requirement"}})
	}
	d, e := canonical.Digest(before)
	if e != nil {
		t.Fatal(e)
	}
	p.BeforeDigest = d
	r, e := Plan(before, p)
	if e != nil || len(r.After.Requirements) != 64 || len(r.Review.Changes) != 64 {
		t.Fatal("exact64operations rejected", e)
	}
	p = Patch{Version: 1, ProfileVersion: ProfileVersion, BeforeDigest: frozenBeforeDigest, Operations: []Operation{{Action: Replace, Kind: State, ID: "A", State: &model.State{ID: "A", Label: strings.Repeat("x", 256)}}}}
	if _, e = Plan(frozenBefore(), p); e != nil {
		t.Fatal("maximum label rejected", e)
	}
	p.Operations = []Operation{{Action: Replace, Kind: Requirement, ID: "R", Requirement: &model.Requirement{ID: "R", Text: strings.Repeat("x", 4096)}}}
	if _, e = Plan(frozenBefore(), p); e != nil {
		t.Fatal("maximum text rejected", e)
	}
	p.BeforeDigest = frozenStaleDigest
	p.Operations = []Operation{{Action: Remove, Kind: State, ID: strings.Repeat("A", 48)}}
	assertError(t, frozenBefore(), p, ErrPrecondition)
	p.Operations[0].ID = strings.Repeat("A", 49)
	assertError(t, frozenBefore(), p, ErrLimit)
	p.Operations = []Operation{{Action: Replace, Kind: Transition, ID: "T1", Transition: &model.Transition{ID: "T1", From: "A", Trigger: "prepare", To: "B", Guards: []string{"G_old"}, Effects: []model.Effect{{Op: strings.Repeat("x", 64)}}}}}
	assertError(t, frozenBefore(), p, ErrPrecondition)
}
