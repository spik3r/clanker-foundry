package topology

import (
	"reflect"
	"testing"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/model"
)

// Hand-authored by the implementation contributor and frozen before mapping;
// these literals are not an independent reviewer's assertions. The graph has
// tied shortest routes, one closed two-state cycle, one cycle with exits, and a
// self-loop with an exit. Guard truth is deliberately not part of the oracle.
func fixture() model.Model {
	m := model.Model{Version: 1, ID: "M", Initial: "A",
		States:       []model.State{{ID: "A", Label: "A"}, {ID: "B", Label: "B"}, {ID: "C", Label: "C"}, {ID: "D", Label: "D"}, {ID: "E", Label: "E"}, {ID: "F", Label: "F"}, {ID: "T", Label: "T", Terminal: true}, {ID: "U", Label: "U", Terminal: true}},
		Requirements: []model.Requirement{{ID: "R", Text: "A condition"}},
		Guards:       []model.Guard{{ID: "G", Requirement: "R", Expr: model.Expr{Kind: "op", Op: "nonempty", Args: []model.Expr{{Kind: "ref", Ref: "task.id"}}}}},
	}
	edge := func(id, from, to string) model.Transition {
		return model.Transition{ID: id, From: from, To: to, Trigger: id + "_trigger", Guards: []string{"G"}, Effects: []model.Effect{{Op: model.SetState}, {Op: model.IncrementRevision}}}
	}
	m.Transitions = []model.Transition{edge("T_A_B", "A", "B"), edge("T_A_C", "A", "C"), edge("T_A_D", "A", "D"), edge("T_A_F", "A", "F"), edge("T_B_C", "B", "C"), edge("T_B_T", "B", "T"), edge("T_C_B", "C", "B"), edge("T_C_U", "C", "U"), edge("T_D_E", "D", "E"), edge("T_E_D", "E", "D"), edge("T_F_F", "F", "F"), edge("T_F_T", "F", "T")}
	return m
}
func TestFrozenGraphVectors(t *testing.T) {
	got, err := Analyze(fixture())
	if err != nil {
		t.Fatal(err)
	}
	wantStates := []State{
		{ID: "A", HasTerminalPath: true, TerminalID: "T", TransitionIDs: []string{"T_A_B", "T_B_T"}},
		{ID: "B", HasTerminalPath: true, TerminalID: "T", TransitionIDs: []string{"T_B_T"}},
		{ID: "C", HasTerminalPath: true, TerminalID: "U", TransitionIDs: []string{"T_C_U"}},
		{ID: "D", HasTerminalPath: false, TerminalID: "", TransitionIDs: []string{}},
		{ID: "E", HasTerminalPath: false, TerminalID: "", TransitionIDs: []string{}},
		{ID: "F", HasTerminalPath: true, TerminalID: "T", TransitionIDs: []string{"T_F_T"}},
		{ID: "T", HasTerminalPath: true, TerminalID: "T", TransitionIDs: []string{}},
		{ID: "U", HasTerminalPath: true, TerminalID: "U", TransitionIDs: []string{}},
	}
	wantCycles := []Cycle{
		{States: []string{"B", "C"}, InternalTransitions: []string{"T_B_C", "T_C_B"}, OutgoingTransitions: []string{"T_B_T", "T_C_U"}},
		{States: []string{"D", "E"}, InternalTransitions: []string{"T_D_E", "T_E_D"}, OutgoingTransitions: []string{}},
		{States: []string{"F"}, InternalTransitions: []string{"T_F_F"}, OutgoingTransitions: []string{"T_F_T"}},
	}
	if !reflect.DeepEqual(got.States, wantStates) {
		t.Fatalf("states: got %#v want %#v", got.States, wantStates)
	}
	if !reflect.DeepEqual(got.Cycles, wantCycles) {
		t.Fatalf("cycles: got %#v want %#v", got.Cycles, wantCycles)
	}
}
