package modelmerge

import (
	"reflect"
	"testing"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/canonical"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/model"
)

// These vectors were specified before the merge implementation. Expected models
// and conflict lists are handwritten, not derived from Merge or another merger.
func oracleBase() model.Model {
	return model.Model{Version: 1, ID: "M", Initial: "A",
		States:       []model.State{{ID: "A", Label: "Draft"}, {ID: "B", Label: "Done", Terminal: true}},
		Requirements: []model.Requirement{{ID: "R", Text: "Invented requirement"}},
		Guards:       []model.Guard{{ID: "G", Requirement: "R", Expr: model.Expr{Kind: "op", Op: "nonempty", Args: []model.Expr{{Kind: "ref", Ref: "task.outcome"}}}}},
		Transitions:  []model.Transition{{ID: "T1", From: "A", Trigger: "prepare", To: "B", Guards: []string{"G"}, Effects: []model.Effect{{Op: model.SetState}, {Op: model.IncrementRevision}}}},
	}
}
func oracleBoth() model.Model {
	return model.Model{Version: 1, ID: "N", Initial: "A",
		States:       []model.State{{ID: "A", Label: "Edited draft"}, {ID: "B", Label: "Done", Terminal: true}},
		Requirements: []model.Requirement{{ID: "R", Text: "Edited requirement"}},
		Guards:       []model.Guard{{ID: "G", Requirement: "R", Expr: model.Expr{Kind: "op", Op: "nonempty", Args: []model.Expr{{Kind: "ref", Ref: "task.outcome"}}}}},
		Transitions:  []model.Transition{{ID: "T1", From: "A", Trigger: "prepare", To: "B", Guards: []string{"G"}, Effects: []model.Effect{{Op: model.SetState}, {Op: model.IncrementRevision}}}},
	}
}
func oracleAdded() model.Model {
	return model.Model{Version: 1, ID: "M", Initial: "A",
		States:       []model.State{{ID: "A", Label: "Draft"}, {ID: "B", Label: "Done", Terminal: true}, {ID: "C", Label: "Other", Terminal: true}},
		Requirements: []model.Requirement{{ID: "R", Text: "Invented requirement"}},
		Guards:       []model.Guard{{ID: "G", Requirement: "R", Expr: model.Expr{Kind: "op", Op: "nonempty", Args: []model.Expr{{Kind: "ref", Ref: "task.outcome"}}}}},
		Transitions: []model.Transition{
			{ID: "T1", From: "A", Trigger: "prepare", To: "B", Guards: []string{"G"}, Effects: []model.Effect{{Op: model.SetState}, {Op: model.IncrementRevision}}},
			{ID: "T2", From: "A", Trigger: "alternate", To: "C", Guards: []string{"G"}, Effects: []model.Effect{{Op: model.SetState}, {Op: model.IncrementRevision}}},
		},
	}
}
func oracleCycle(id, initial string) model.Model {
	return model.Model{Version: 1, ID: id, Initial: initial,
		States:       []model.State{{ID: "A", Label: "First"}, {ID: "B", Label: "Second"}, {ID: "C", Label: "Finished", Terminal: true}},
		Requirements: []model.Requirement{{ID: "R", Text: "Review"}},
		Guards:       []model.Guard{{ID: "G", Requirement: "R", Expr: model.Expr{Kind: "op", Op: "nonempty", Args: []model.Expr{{Kind: "ref", Ref: "task.outcome"}}}}},
		Transitions: []model.Transition{
			{ID: "T1", From: "A", Trigger: "next", To: "B", Guards: []string{"G"}, Effects: []model.Effect{{Op: model.SetState}, {Op: model.IncrementRevision}}},
			{ID: "T2", From: "B", Trigger: "back", To: "A", Guards: []string{"G"}, Effects: []model.Effect{{Op: model.SetState}, {Op: model.IncrementRevision}}},
			{ID: "T3", From: "A", Trigger: "finish", To: "C", Guards: []string{"G"}, Effects: []model.Effect{{Op: model.SetState}, {Op: model.IncrementRevision}}},
		},
	}
}

type oracleVector struct {
	name              string
	base, left, right model.Model
	status            string
	candidate         *model.Model
	conflicts         []Conflict
}

func oraclePtr(m model.Model) *model.Model { return &m }

// All fixture changes below are explicit branch inputs. No expected model is
// computed by applying a proposed merge to another model.
func oracleVectors() []oracleVector {
	base := oracleBase()
	left := oracleBase()
	left.ID = "N"
	left.States[0].Label = "Edited draft"
	right := oracleBase()
	right.Requirements[0].Text = "Edited requirement"
	identical := oracleAdded()
	a1 := oracleAdded()
	a1.States[2].Label = "First addition"
	a2 := oracleAdded()
	a2.States[2].Label = "Second addition"
	remove := oracleBase()
	change := oracleAdded()
	change.States[2].Label = "Changed existing"
	fieldLeft := oracleBase()
	fieldLeft.States[0].Label = "Changed A"
	fieldRight := oracleBase()
	fieldRight.States[0].Terminal = true
	// For a valid independently changed state, changing Terminal on A would
	// violate the profile. Use B: one side changes label, the other makes B a
	// nonterminal and supplies its required edge. The atomic B record conflicts.
	fieldLeft = oracleBase()
	fieldLeft.States[1].Label = "Edited B"
	fieldRight = oracleBase()
	fieldRight.States[1].Terminal = false
	fieldRight.Transitions = append(fieldRight.Transitions, model.Transition{ID: "T2", From: "B", Trigger: "back", To: "A", Guards: []string{"G"}, Effects: []model.Effect{{Op: model.SetState}, {Op: model.IncrementRevision}}})
	// Keep a reachable terminal after B becomes nonterminal.
	fieldRight.States = append(fieldRight.States, model.State{ID: "C", Label: "Done C", Terminal: true})
	fieldRight.Transitions = append(fieldRight.Transitions, model.Transition{ID: "T3", From: "B", Trigger: "finish", To: "C", Guards: []string{"G"}, Effects: []model.Effect{{Op: model.SetState}, {Op: model.IncrementRevision}}})
	rootLeft := oracleCycle("N", "A")
	rootRight := oracleCycle("M", "B")
	rootConflictLeft := oracleBase()
	rootConflictLeft.ID = "N"
	rootConflictRight := oracleBase()
	rootConflictRight.ID = "O"
	globalLeft := oracleBase()
	globalLeft.ID = "C"
	globalRight := oracleAdded()
	dispatchLeft := oracleAdded()
	dispatchRight := oracleBase()
	dispatchRight.States = append(dispatchRight.States, model.State{ID: "D", Label: "Different", Terminal: true})
	dispatchRight.Transitions = append(dispatchRight.Transitions, model.Transition{ID: "T3", From: "A", Trigger: "alternate", To: "D", Guards: []string{"G"}, Effects: []model.Effect{{Op: model.SetState}, {Op: model.IncrementRevision}}})
	// Normalize set order before comparing: record collection order is not a
	// change. This case reverses state and transition order only.
	reordered := oracleAdded()
	reordered.States[0], reordered.States[2] = reordered.States[2], reordered.States[0]
	reordered.Transitions[0], reordered.Transitions[1] = reordered.Transitions[1], reordered.Transitions[0]
	// Several atomic units conflict at once; their expected order is lexical
	// by kind, ID and field, independent of branch order.
	multiLeft := oracleBase()
	multiLeft.ID = "N"
	multiLeft.Guards[0].Expr.Args[0].Ref = "task.id"
	multiLeft.Requirements[0].Text = "Left requirement"
	multiLeft.States[1].Label = "Left state"
	multiLeft.Transitions[0].Trigger = "left"
	multiRight := oracleBase()
	multiRight.ID = "O"
	multiRight.Guards[0].Expr.Args[0].Ref = "command.actor"
	multiRight.Requirements[0].Text = "Right requirement"
	multiRight.States[1].Label = "Right state"
	multiRight.Transitions[0].Trigger = "right"
	return []oracleVector{
		{"no-op", base, oracleBase(), oracleBase(), Merged, oraclePtr(oracleBase()), []Conflict{}},
		{"one-sided addition", base, oracleAdded(), oracleBase(), Merged, oraclePtr(oracleAdded()), []Conflict{}},
		{"identical addition", base, identical, oracleAdded(), Merged, oraclePtr(oracleAdded()), []Conflict{}},
		{"independent records and id", base, left, right, Merged, oraclePtr(oracleBoth()), []Conflict{}},
		{"independent root fields", oracleCycle("M", "A"), rootLeft, rootRight, Merged, oraclePtr(oracleCycle("N", "B")), []Conflict{}},
		{"normalization ignores set order", oracleAdded(), reordered, oracleAdded(), Merged, oraclePtr(oracleAdded()), []Conflict{}},
		{"add-add", base, a1, a2, Conflicted, nil, []Conflict{{Kind: "state", ID: "C", Reason: "add-add"}}},
		{"remove-change", oracleAdded(), remove, change, Conflicted, nil, []Conflict{{Kind: "state", ID: "C", Reason: "remove-change"}}},
		{"coordinated removal", oracleAdded(), oracleBase(), oracleBase(), Merged, oraclePtr(oracleBase()), []Conflict{}},
		{"atomic record not field merge", base, fieldLeft, fieldRight, Conflicted, nil, []Conflict{{Kind: "state", ID: "B", Reason: "change-change"}}},
		{"root replacement", base, rootConflictLeft, rootConflictRight, Conflicted, nil, []Conflict{{Kind: "model", Field: "id", Reason: "change-change"}}},
		{"complete conflict ordering", base, multiLeft, multiRight, Conflicted, nil, []Conflict{{Kind: "guard", ID: "G", Reason: "change-change"}, {Kind: "model", Field: "id", Reason: "change-change"}, {Kind: "requirement", ID: "R", Reason: "change-change"}, {Kind: "state", ID: "B", Reason: "change-change"}, {Kind: "transition", ID: "T1", Reason: "change-change"}}},
		{"duplicate global id after union", base, globalLeft, globalRight, InvalidCombination, nil, []Conflict{}},
		{"dispatch collision after union", base, dispatchLeft, dispatchRight, InvalidCombination, nil, []Conflict{}},
	}
}

// SHA-256 of handwritten canonical JSON, independently produced with Python's
// sorted-key UTF-8 JSON and hashlib. Not computed from Merge output.
func oracleCandidateDigest(name string) string {
	switch name {
	case "no-op", "coordinated removal":
		return "db4563afc7984d60c6a958e4d09b2df23a367a56ed7c641737474e7b03470133"
	case "one-sided addition", "identical addition", "normalization ignores set order":
		return "894defaf008409ef5fbadbb33a6ca0bf482404fdf64a6229d00e675de7ffc828"
	case "independent records and id":
		return "92081c06679a0ceebe43f10c0874648ecf786edacc12c8d5102e2defba14facd"
	case "independent root fields":
		return "1e1f2191eec1e7c3c3a3bec3beedad348d75143b2f45ffbe0d33866630e6f462"
	default:
		return ""
	}
}

func TestOracleFixturesAreValid(t *testing.T) {
	for _, v := range oracleVectors() {
		t.Run(v.name, func(t *testing.T) {
			for _, m := range []model.Model{v.base, v.left, v.right} {
				if _, err := canonical.Digest(m); err != nil {
					t.Fatalf("invalid fixture: %v", err)
				}
			}
			if v.candidate != nil {
				n, err := model.Normalize(*v.candidate)
				if err != nil || !reflect.DeepEqual(n, *v.candidate) {
					t.Fatalf("expected candidate is not valid normalized model: %v", err)
				}
				digest, err := canonical.Digest(*v.candidate)
				if err != nil || digest != oracleCandidateDigest(v.name) {
					t.Fatalf("independent candidate digest differs: %s %v", digest, err)
				}
			}
			if v.conflicts == nil {
				t.Fatal("expected conflict array must be explicit")
			}
		})
	}
}

func TestFrozenThreeWayVectors(t *testing.T) {
	for _, v := range oracleVectors() {
		t.Run(v.name, func(t *testing.T) {
			for _, pair := range [][2]model.Model{{v.left, v.right}, {v.right, v.left}} {
				got, err := Merge(v.base, pair[0], pair[1])
				if err != nil {
					t.Fatalf("Merge: %v", err)
				}
				if got.MergeVersion != Version || got.ProfileVersion != "lifecycle/v1" || got.Mode != "example-only" || got.Authority != "none" || got.Sources != nil {
					t.Fatalf("wrong metadata: %+v", got)
				}
				if got.Status != v.status || !reflect.DeepEqual(got.Conflicts, v.conflicts) {
					t.Fatalf("got status/conflicts %q %#v; want %q %#v", got.Status, got.Conflicts, v.status, v.conflicts)
				}
				if !reflect.DeepEqual(got.Candidate, v.candidate) {
					t.Fatalf("got candidate %#v; want %#v", got.Candidate, v.candidate)
				}
				if (got.CandidateDigest != nil) != (v.candidate != nil) {
					t.Fatalf("candidate digest presence: %#v", got.CandidateDigest)
				}
				if got.CandidateDigest != nil && *got.CandidateDigest != oracleCandidateDigest(v.name) {
					t.Fatalf("candidate digest: %s", *got.CandidateDigest)
				}
			}
		})
	}
}
