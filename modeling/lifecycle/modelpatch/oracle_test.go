package modelpatch

import (
	"testing"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/canonical"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/impact"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/model"
)

// Independent frozen vectors: Python hashlib over hand-authored canonical model
// JSON, sorted map keys, UTF-8, no HTML escaping, no trailing newline. No planner,
// candidate execution or comparison implementation produced these expectations.
const frozenBeforeDigest = "0a6156d1f7c7f87a0022115fa75246ee3286ffc29046298cac43ad7baf077d9d"

func frozenBefore() model.Model {
	return model.Model{Version: 1, ID: "M", Initial: "A",
		States:       []model.State{{ID: "A", Label: "Draft"}, {ID: "B", Label: "Done", Terminal: true}},
		Requirements: []model.Requirement{{ID: "R", Text: "Invented requirement"}},
		Guards:       []model.Guard{{ID: "G_old", Requirement: "R", Expr: model.Expr{Kind: "op", Op: "nonempty", Args: []model.Expr{{Kind: "ref", Ref: "task.outcome"}}}}},
		Transitions:  []model.Transition{{ID: "T1", From: "A", Trigger: "prepare", To: "B", Guards: []string{"G_old"}, Effects: []model.Effect{{Op: model.SetState}, {Op: model.IncrementRevision}}}},
	}
}

type frozenFixture struct {
	Name        string
	Before      model.Model
	Patch       Patch
	After       model.Model
	AfterDigest string
	Changes     []impact.Change
	Links       []impact.LinkedTransition
}

// This handwritten oracle is intent, not a patch interpreter. Before and after
// records are specified separately; no expected references are auto-repaired.
// Preserve these literals when adding future Plan/Execute assertions.
func frozenFixtures() []frozenFixture {
	renameAfter := model.Model{Version: 1, ID: "M", Initial: "A",
		States:       []model.State{{ID: "A", Label: "Draft"}, {ID: "B", Label: "Done", Terminal: true}},
		Requirements: []model.Requirement{{ID: "R", Text: "Invented requirement"}},
		Guards:       []model.Guard{{ID: "G_new", Requirement: "R", Expr: model.Expr{Kind: "op", Op: "nonempty", Args: []model.Expr{{Kind: "ref", Ref: "task.outcome"}}}}},
		Transitions:  []model.Transition{{ID: "T1", From: "A", Trigger: "prepare", To: "B", Guards: []string{"G_new"}, Effects: []model.Effect{{Op: model.SetState}, {Op: model.IncrementRevision}}}},
	}
	rootAfter := model.Model{Version: 1, ID: "N", Initial: "A",
		States:       []model.State{{ID: "A", Label: "Draft"}, {ID: "B", Label: "Done", Terminal: true}},
		Requirements: []model.Requirement{{ID: "R", Text: "Invented requirement"}},
		Guards:       []model.Guard{{ID: "G_old", Requirement: "R", Expr: model.Expr{Kind: "op", Op: "nonempty", Args: []model.Expr{{Kind: "ref", Ref: "task.outcome"}}}}},
		Transitions:  []model.Transition{{ID: "T1", From: "A", Trigger: "prepare", To: "B", Guards: []string{"G_old"}, Effects: []model.Effect{{Op: model.SetState}, {Op: model.IncrementRevision}}}},
	}
	literalAfter := model.Model{Version: 1, ID: "M", Initial: "A",
		States:       []model.State{{ID: "A", Label: "Draft"}, {ID: "B", Label: "Done", Terminal: true}},
		Requirements: []model.Requirement{{ID: "R", Text: "Invented requirement"}},
		Guards:       []model.Guard{{ID: "G_old", Requirement: "R", Expr: model.Expr{Kind: "op", Op: "eq", Args: []model.Expr{{Kind: "ref", Ref: "task.outcome"}, {Kind: "string", Value: "literal <&> 雪\n\""}}}}},
		Transitions:  []model.Transition{{ID: "T1", From: "A", Trigger: "prepare", To: "B", Guards: []string{"G_old"}, Effects: []model.Effect{{Op: model.SetState}, {Op: model.IncrementRevision}}}},
	}
	crossAfter := model.Model{Version: 1, ID: "M", Initial: "A",
		States:       []model.State{{ID: "A", Label: "Draft"}, {ID: "B", Label: "Done", Terminal: true}, {ID: "G_old", Label: "Added terminal", Terminal: true}},
		Requirements: []model.Requirement{{ID: "R", Text: "Invented requirement"}},
		Guards:       []model.Guard{{ID: "G_new", Requirement: "R", Expr: model.Expr{Kind: "op", Op: "nonempty", Args: []model.Expr{{Kind: "ref", Ref: "task.outcome"}}}}},
		Transitions: []model.Transition{
			{ID: "T1", From: "A", Trigger: "prepare", To: "B", Guards: []string{"G_new"}, Effects: []model.Effect{{Op: model.SetState}, {Op: model.IncrementRevision}}},
			{ID: "T2", From: "A", Trigger: "alternate", To: "G_old", Guards: []string{"G_new"}, Effects: []model.Effect{{Op: model.SetState}, {Op: model.IncrementRevision}}},
		},
	}
	return []frozenFixture{
		{Name: "coordinated explicit guard rename", Before: frozenBefore(), Patch: Patch{Version: 1, ProfileVersion: "lifecycle/v1", BeforeDigest: frozenBeforeDigest, Operations: []Operation{
			{Action: Remove, Kind: Guard, ID: "G_old"},
			{Action: Add, Kind: Guard, ID: "G_new", Guard: &model.Guard{ID: "G_new", Requirement: "R", Expr: model.Expr{Kind: "op", Op: "nonempty", Args: []model.Expr{{Kind: "ref", Ref: "task.outcome"}}}}},
			{Action: Replace, Kind: Transition, ID: "T1", Transition: &model.Transition{ID: "T1", From: "A", Trigger: "prepare", To: "B", Guards: []string{"G_new"}, Effects: []model.Effect{{Op: model.SetState}, {Op: model.IncrementRevision}}}},
		}}, After: renameAfter, AfterDigest: "654370b558f03a682e7a17c516e4d5554616d8ab28aaf573256fee0de966c1a0",
			Changes: []impact.Change{{Kind: "guard", ID: "G_new", Type: "added", Fields: []string{}}, {Kind: "guard", ID: "G_old", Type: "removed", Fields: []string{}}, {Kind: "transition", ID: "T1", Type: "modified", Fields: []string{"guards"}}},
			Links:   []impact.LinkedTransition{{Side: "before", ID: "T1", Causes: []string{"guard:G_old", "transition:T1"}}, {Side: "after", ID: "T1", Causes: []string{"guard:G_new", "transition:T1"}}}},
		{Name: "zero operations no approval", Before: frozenBefore(), Patch: Patch{Version: 1, ProfileVersion: "lifecycle/v1", BeforeDigest: frozenBeforeDigest, Operations: []Operation{}}, After: frozenBefore(), AfterDigest: frozenBeforeDigest, Changes: []impact.Change{}, Links: []impact.LinkedTransition{}},
		{Name: "identical complete replacement", Before: frozenBefore(), Patch: Patch{Version: 1, ProfileVersion: "lifecycle/v1", BeforeDigest: frozenBeforeDigest, Operations: []Operation{{Action: Replace, Kind: State, ID: "A", State: &model.State{ID: "A", Label: "Draft"}}}}, After: frozenBefore(), AfterDigest: frozenBeforeDigest, Changes: []impact.Change{}, Links: []impact.LinkedTransition{}},
		{Name: "explicit root identity", Before: frozenBefore(), Patch: Patch{Version: 1, ProfileVersion: "lifecycle/v1", BeforeDigest: frozenBeforeDigest, Operations: []Operation{{Action: Replace, Kind: Model, ID: "", Root: &Root{ID: "N", Initial: "A"}}}}, After: rootAfter, AfterDigest: "7846a8112e78c31a75e9744dddfaecd4e780e7898ccb2697f97e5eb620133929",
			Changes: []impact.Change{{Kind: "model", ID: "", Type: "modified", Fields: []string{"id"}}}, Links: []impact.LinkedTransition{{Side: "before", ID: "T1", Causes: []string{"model:"}}, {Side: "after", ID: "T1", Causes: []string{"model:"}}}},
		{Name: "escaped literal is data", Before: frozenBefore(), Patch: Patch{Version: 1, ProfileVersion: "lifecycle/v1", BeforeDigest: frozenBeforeDigest, Operations: []Operation{{Action: Replace, Kind: Guard, ID: "G_old", Guard: &model.Guard{ID: "G_old", Requirement: "R", Expr: model.Expr{Kind: "op", Op: "eq", Args: []model.Expr{{Kind: "ref", Ref: "task.outcome"}, {Kind: "string", Value: "literal <&> 雪\n\""}}}}}}}, After: literalAfter, AfterDigest: "b5d4abe1bddd46a02f4a5620203a1aa1a94bef471243db4389455697ed23a766",
			Changes: []impact.Change{{Kind: "guard", ID: "G_old", Type: "modified", Fields: []string{"expr"}}}, Links: []impact.LinkedTransition{{Side: "before", ID: "T1", Causes: []string{"guard:G_old"}}, {Side: "after", ID: "T1", Causes: []string{"guard:G_old"}}}},
		{Name: "cross kind explicit reuse", Before: frozenBefore(), Patch: Patch{Version: 1, ProfileVersion: "lifecycle/v1", BeforeDigest: frozenBeforeDigest, Operations: []Operation{
			{Action: Remove, Kind: Guard, ID: "G_old"},
			{Action: Add, Kind: State, ID: "G_old", State: &model.State{ID: "G_old", Label: "Added terminal", Terminal: true}},
			{Action: Add, Kind: Guard, ID: "G_new", Guard: &model.Guard{ID: "G_new", Requirement: "R", Expr: model.Expr{Kind: "op", Op: "nonempty", Args: []model.Expr{{Kind: "ref", Ref: "task.outcome"}}}}},
			{Action: Replace, Kind: Transition, ID: "T1", Transition: &model.Transition{ID: "T1", From: "A", Trigger: "prepare", To: "B", Guards: []string{"G_new"}, Effects: []model.Effect{{Op: model.SetState}, {Op: model.IncrementRevision}}}},
			{Action: Add, Kind: Transition, ID: "T2", Transition: &model.Transition{ID: "T2", From: "A", Trigger: "alternate", To: "G_old", Guards: []string{"G_new"}, Effects: []model.Effect{{Op: model.SetState}, {Op: model.IncrementRevision}}}},
		}}, After: crossAfter, AfterDigest: "d35abf1b8fa388d039ad1676c6564cf2394d163d33ee90f87a3779dc966911c1",
			Changes: []impact.Change{{Kind: "guard", ID: "G_new", Type: "added", Fields: []string{}}, {Kind: "guard", ID: "G_old", Type: "removed", Fields: []string{}}, {Kind: "state", ID: "G_old", Type: "added", Fields: []string{}}, {Kind: "transition", ID: "T1", Type: "modified", Fields: []string{"guards"}}, {Kind: "transition", ID: "T2", Type: "added", Fields: []string{}}},
			Links:   []impact.LinkedTransition{{Side: "before", ID: "T1", Causes: []string{"guard:G_old", "transition:T1"}}, {Side: "after", ID: "T1", Causes: []string{"guard:G_new", "transition:T1"}}, {Side: "after", ID: "T2", Causes: []string{"guard:G_new", "state:G_old", "transition:T2"}}}},
	}
}

// This runs only existing profile validation and independently fixed digest
// vectors. It neither applies patches nor proves a planner that does not exist.
func TestFrozenFixtureModelsAndCanonicalVectors(t *testing.T) {
	fixtures := frozenFixtures()
	if len(fixtures) != 6 {
		t.Fatal("frozen fixture inventory changed")
	}
	for _, f := range fixtures {
		t.Run(f.Name, func(t *testing.T) {
			if err := model.Validate(f.Before); err != nil {
				t.Fatal("invalid handwritten before", err)
			}
			if err := model.Validate(f.After); err != nil {
				t.Fatal("invalid handwritten after", err)
			}
			before, err := canonical.Digest(f.Before)
			if err != nil || before != frozenBeforeDigest {
				t.Fatal("independent before vector differs", before, err)
			}
			after, err := canonical.Digest(f.After)
			if err != nil || after != f.AfterDigest {
				t.Fatal("independent after vector differs", after, err)
			}
			if f.Patch.Operations == nil || f.Changes == nil || f.Links == nil {
				t.Fatal("fixture unexpectedly uses nil collections")
			}
			for _, c := range f.Changes {
				if c.Fields == nil {
					t.Fatal("nil expected fields")
				}
			}
			for _, l := range f.Links {
				if len(l.Causes) == 0 {
					t.Fatal("empty expected causes")
				}
			}
		})
	}
}

// Mixed-error literals frozen after lead stage-2 clarification and BEFORE any
// planner implementation. These append to the byte-identical original oracle.
const frozenStaleDigest = "0000000000000000000000000000000000000000000000000000000000000000"

type mixedErrorFixture struct {
	Name  string
	Patch Patch
	Want  error
}

func frozenMixedErrors() []mixedErrorFixture {
	p := func(ops ...Operation) Patch {
		return Patch{Version: 1, ProfileVersion: "lifecycle/v1", BeforeDigest: frozenStaleDigest, Operations: ops}
	}
	transition := func(guards []string, effects []model.Effect, to, trigger string) Operation {
		return Operation{Action: Replace, Kind: Transition, ID: "T1", Transition: &model.Transition{ID: "T1", From: "A", Trigger: trigger, To: to, Guards: guards, Effects: effects}}
	}
	effects := []model.Effect{{Op: model.SetState}, {Op: model.IncrementRevision}}
	overlong := make([]byte, 65)
	for i := range overlong {
		overlong[i] = 'x'
	}
	return []mixedErrorFixture{
		{"stale blank state label", p(Operation{Action: Replace, Kind: State, ID: "A", State: &model.State{ID: "A", Label: " \t"}}), ErrPrecondition},
		{"stale blank requirement text", p(Operation{Action: Replace, Kind: Requirement, ID: "R", Requirement: &model.Requirement{ID: "R", Text: ""}}), ErrPrecondition},
		{"stale duplicate guard references", p(transition([]string{"G_old", "G_old"}, effects, "B", "prepare")), ErrPrecondition},
		{"stale empty guard references", p(transition([]string{}, effects, "B", "prepare")), ErrPrecondition},
		{"stale unknown bounded effect", p(transition([]string{"G_old"}, []model.Effect{{Op: "unknown_effect"}, {Op: model.IncrementRevision}}, "B", "prepare")), ErrPrecondition},
		{"stale reordered effects", p(transition([]string{"G_old"}, []model.Effect{{Op: model.IncrementRevision}, {Op: model.SetState}}, "B", "prepare")), ErrPrecondition},
		{"stale duplicate effects", p(transition([]string{"G_old"}, []model.Effect{{Op: model.SetState}, {Op: model.SetState}}, "B", "prepare")), ErrPrecondition},
		{"stale dangling valid state reference", p(transition([]string{"G_old"}, effects, "Missing", "prepare")), ErrPrecondition},
		{"stale dangling valid requirement reference", p(Operation{Action: Replace, Kind: Guard, ID: "G_old", Guard: &model.Guard{ID: "G_old", Requirement: "Missing", Expr: model.Expr{Kind: "op", Op: "nonempty", Args: []model.Expr{{Kind: "ref", Ref: "task.outcome"}}}}}), ErrPrecondition},
		{"stale invalid reference grammar", p(transition([]string{"G_old"}, effects, "bad-id", "prepare")), ErrInvalid},
		{"stale invalid guard id grammar", p(transition([]string{"bad-id"}, effects, "B", "prepare")), ErrInvalid},
		{"stale invalid trigger grammar", p(transition([]string{"G_old"}, effects, "B", "bad-trigger")), ErrInvalid},
		{"stale invalid root initial grammar", p(Operation{Action: Replace, Kind: Model, Root: &Root{ID: "M", Initial: "bad-id"}}), ErrInvalid},
		{"stale invalid guard typing", p(Operation{Action: Replace, Kind: Guard, ID: "G_old", Guard: &model.Guard{ID: "G_old", Requirement: "R", Expr: model.Expr{Kind: "ref", Ref: "task.outcome"}}}), ErrInvalid},
		{"stale record id mismatch", p(Operation{Action: Replace, Kind: State, ID: "A", State: &model.State{ID: "B", Label: "Draft"}}), ErrInvalid},
		{"stale unknown action", p(Operation{Action: Action("merge"), Kind: State, ID: "A", State: &model.State{ID: "A", Label: "Draft"}}), ErrInvalid},
		{"stale invalid union", p(Operation{Action: Replace, Kind: State, ID: "A", State: &model.State{ID: "A", Label: "Draft"}, Root: &Root{ID: "M", Initial: "A"}}), ErrInvalid},
		{"stale duplicate valid operation target", p(Operation{Action: Replace, Kind: State, ID: "A", State: &model.State{ID: "A", Label: "Draft"}}, Operation{Action: Replace, Kind: State, ID: "A", State: &model.State{ID: "A", Label: "Other"}}), ErrConflict},
		{"stale missing target", p(Operation{Action: Remove, Kind: State, ID: "Missing"}), ErrPrecondition},
		{"stale overlong effect op", p(transition([]string{"G_old"}, []model.Effect{{Op: string(overlong)}, {Op: model.IncrementRevision}}, "B", "prepare")), ErrLimit},
		{"stale invalid UTF8", p(Operation{Action: Replace, Kind: State, ID: "A", State: &model.State{ID: "A", Label: string([]byte{255})}}), ErrInvalid},
	}
}
