package topology

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/jsonformat"
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

func TestDeterminismOwnershipAndNoGuardEvaluation(t *testing.T) {
	m := fixture()
	raw, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	var original model.Model
	if err = json.Unmarshal(raw, &original); err != nil {
		t.Fatal(err)
	}
	first, err := Analyze(m)
	if err != nil {
		t.Fatal(err)
	}
	if first.FormatVersion != FormatVersion || first.Mode != "structural-only" || first.Authority != "none" || first.ModelID != "M" || first.Initial != "A" || len(first.ModelDigest) != 64 {
		t.Fatalf("envelope: %+v", first)
	}
	if !reflect.DeepEqual(m, original) {
		t.Fatal("input mutated")
	}
	for i, j := 0, len(m.States)-1; i < j; i, j = i+1, j-1 {
		m.States[i], m.States[j] = m.States[j], m.States[i]
	}
	for i, j := 0, len(m.Transitions)-1; i < j; i, j = i+1, j-1 {
		m.Transitions[i], m.Transitions[j] = m.Transitions[j], m.Transitions[i]
	}
	reordered, err := Analyze(m)
	if err != nil || !reflect.DeepEqual(first, reordered) {
		t.Fatal("set reordering changed output", err)
	}
	first.States[0].TransitionIDs[0] = "changed"
	first.Cycles[0].States[0] = "changed"
	first.Cycles[0].InternalTransitions[0] = "changed"
	again, err := Analyze(m)
	if err != nil || !reflect.DeepEqual(again, reordered) {
		t.Fatal("output aliases input or another result", err)
	}
	// An unsatisfiable expression must not be evaluated or treated as a proof
	// that the declared paths cannot be used. Its changed model digest remains visible.
	impossible := fixture()
	impossible.Guards[0].Expr = model.Expr{Kind: "op", Op: "eq", Args: []model.Expr{{Kind: "string", Value: "x"}, {Kind: "string", Value: "y"}}}
	structural, err := Analyze(impossible)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(structural.States, reordered.States) || !reflect.DeepEqual(structural.Cycles, reordered.Cycles) || structural.ModelDigest == reordered.ModelDigest {
		t.Fatal("guard handling changed structural graph or hid model identity")
	}
}

func TestParallelEdgeAndInvalidTypedModels(t *testing.T) {
	m := fixture()
	edge := m.Transitions[0]
	edge.ID = "T_A_B_alt"
	edge.Trigger = "alternative"
	m.Transitions = append(m.Transitions, edge)
	result, err := Analyze(m)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(result.States[0].TransitionIDs, []string{"T_A_B", "T_B_T"}) {
		t.Fatal("parallel edge tie", result.States[0])
	}
	invalid := []model.Model{model.Model{}, fixture(), fixture(), fixture()}
	invalid[1].States[0].ID = "bad id"
	invalid[2].Transitions[0].To = "Missing"
	invalid[3].Transitions = append(invalid[3].Transitions, invalid[3].Transitions[0])
	for _, bad := range invalid {
		got, err := Analyze(bad)
		if err == nil || !reflect.DeepEqual(got, Report{}) {
			t.Fatal("invalid model produced partial report")
		}
	}
}

func TestExecuteStrictAndDeterministic(t *testing.T) {
	modelBytes, err := jsonformat.Encode(fixture())
	if err != nil {
		t.Fatal(err)
	}
	req, err := json.Marshal(struct {
		FormatVersion string          `json:"formatVersion"`
		Model         json.RawMessage `json:"model"`
	}{RequestFormatVersion, modelBytes})
	if err != nil {
		t.Fatal(err)
	}
	first, err := Execute(req)
	if err != nil {
		t.Fatal(err)
	}
	second, err := Execute(req)
	if err != nil || !bytes.Equal(first, second) {
		t.Fatal("nondeterministic output", err)
	}
	var decoded Report
	if err = json.Unmarshal(first, &decoded); err != nil || decoded.Mode != "structural-only" || len(decoded.Cycles) != 3 {
		t.Fatal("unexpected output", err)
	}
	cases := map[string][]byte{
		"empty":         []byte(`{}`),
		"extra":         bytes.Replace(req, []byte(`{"formatVersion"`), []byte(`{"unknown":1,"formatVersion"`), 1),
		"duplicate":     bytes.Replace(req, []byte(`"formatVersion":`), []byte(`"formatVersion":"x","formatVersion":`), 1),
		"case":          bytes.Replace(req, []byte(`"formatVersion"`), []byte(`"FormatVersion"`), 1),
		"null":          []byte(`{"formatVersion":"lifecycle-structural-topology-request/v1","model":null}`),
		"trailing":      append(append([]byte{}, req...), []byte(` {}`)...),
		"bad utf8":      append(append([]byte{}, req...), 0xff),
		"bad surrogate": bytes.Replace(req, []byte(`"label":"A"`), []byte(`"label":"\ud800"`), 1),
		"oversized":     []byte(strings.Repeat(" ", MaxRequestBytes+1)),
		"wrong version": bytes.Replace(req, []byte(RequestFormatVersion), []byte("other"), 1),
	}
	for name, b := range cases {
		t.Run(name, func(t *testing.T) {
			if bytes.Equal(req, b) {
				t.Fatal("mutation did not apply")
			}
			out, err := Execute(b)
			if err == nil || out != nil {
				t.Fatal("accepted invalid request")
			}
		})
	}
}

func TestMaximumDeclaredGraphBounds(t *testing.T) {
	m := model.Model{Version: 1, ID: "MaxModel", Initial: "S00", Requirements: []model.Requirement{{ID: "R", Text: "condition"}}, Guards: []model.Guard{{ID: "G", Requirement: "R", Expr: model.Expr{Kind: "op", Op: "nonempty", Args: []model.Expr{{Kind: "ref", Ref: "task.id"}}}}}}
	for i := 0; i < 16; i++ {
		m.States = append(m.States, model.State{ID: fmt.Sprintf("S%02d", i), Label: "state", Terminal: i == 15})
	}
	edge := func(id, from, to string) model.Transition {
		return model.Transition{ID: id, From: from, To: to, Trigger: id, Guards: []string{"G"}, Effects: []model.Effect{{Op: model.SetState}, {Op: model.IncrementRevision}}}
	}
	for i := 0; i < 15; i++ {
		from := fmt.Sprintf("S%02d", i)
		to := fmt.Sprintf("S%02d", i+1)
		m.Transitions = append(m.Transitions, edge(fmt.Sprintf("E%02d", i), from, to), edge(fmt.Sprintf("L%02d", i), from, from))
	}
	m.Transitions = append(m.Transitions, edge("Zalt1", "S00", "S01"), edge("Zalt2", "S00", "S01"))
	if len(m.Transitions) != 32 {
		t.Fatal("fixture is not at transition bound")
	}
	got, err := Analyze(m)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.States) != 16 || len(got.Cycles) != 15 || !got.States[0].HasTerminalPath || got.States[0].TerminalID != "S15" || len(got.States[0].TransitionIDs) != 15 || got.States[0].TransitionIDs[0] != "E00" {
		t.Fatal("unexpected bounded graph result")
	}
	over := m
	over.States = append(append([]model.State{}, m.States...), model.State{ID: "Extra", Label: "extra"})
	if result, err := Analyze(over); err == nil || !reflect.DeepEqual(result, Report{}) {
		t.Fatal("accepted graph above model bound")
	}
}

// Exhaustively compare four-state graphs with an independently structured
// simple-path enumerator and per-source DFS connectivity oracle. This is a
// same-contributor test, not an independent-person attestation.
func TestAllSmallValidGraphsAgainstBruteForce(t *testing.T) {
	ids := []string{"A", "B", "C", "T"}
	tested := 0
	for mask := 0; mask < 1<<12; mask++ {
		m := model.Model{Version: 1, ID: "M", Initial: "A", States: []model.State{{ID: "A", Label: "A"}, {ID: "B", Label: "B"}, {ID: "C", Label: "C"}, {ID: "T", Label: "T", Terminal: true}}, Requirements: []model.Requirement{{ID: "R", Text: "condition"}}, Guards: []model.Guard{{ID: "G", Requirement: "R", Expr: model.Expr{Kind: "op", Op: "nonempty", Args: []model.Expr{{Kind: "ref", Ref: "task.id"}}}}}}
		for i := 0; i < 3; i++ {
			for j := 0; j < 4; j++ {
				if mask&(1<<(i*4+j)) != 0 {
					id := fmt.Sprintf("E%d%d", i, j)
					m.Transitions = append(m.Transitions, model.Transition{ID: id, From: ids[i], To: ids[j], Trigger: id, Guards: []string{"G"}, Effects: []model.Effect{{Op: model.SetState}, {Op: model.IncrementRevision}}})
				}
			}
		}
		if model.Validate(m) != nil {
			continue
		}
		tested++
		got, err := Analyze(m)
		if err != nil {
			t.Fatalf("mask %03x: %v", mask, err)
		}
		byID := map[string][]model.Transition{}
		for _, e := range m.Transitions {
			byID[e.From] = append(byID[e.From], e)
		}
		for i, start := range ids {
			var best []string
			found := false
			var enumerate func(string, map[string]bool, []string)
			enumerate = func(at string, seen map[string]bool, path []string) {
				if at == "T" {
					if !found || len(path) < len(best) || (len(path) == len(best) && strings.Join(path, "\x00") < strings.Join(best, "\x00")) {
						found = true
						best = append([]string{}, path...)
					}
					return
				}
				for _, e := range byID[at] {
					if !seen[e.To] {
						seen[e.To] = true
						enumerate(e.To, seen, append(path, e.ID))
						delete(seen, e.To)
					}
				}
			}
			enumerate(start, map[string]bool{start: true}, []string{})
			state := got.States[i]
			if state.ID != start || state.HasTerminalPath != found || (found && (state.TerminalID != "T" || !reflect.DeepEqual(state.TransitionIDs, best))) || (!found && (state.TerminalID != "" || len(state.TransitionIDs) != 0)) {
				t.Fatalf("mask %03x state %s: %+v expected %v %v", mask, start, state, found, best)
			}
		}
		// Compute connectivity by DFS from each state rather than the product's
		// all-pairs closure, then derive component membership and edge lists.
		paths := map[string]map[string]bool{}
		for _, start := range ids {
			seen := map[string]bool{}
			var visit func(string)
			visit = func(at string) {
				if seen[at] {
					return
				}
				seen[at] = true
				for _, e := range byID[at] {
					visit(e.To)
				}
			}
			visit(start)
			paths[start] = seen
		}
		used := map[string]bool{}
		want := []Cycle{}
		for _, start := range ids {
			if used[start] {
				continue
			}
			members := map[string]bool{}
			states := []string{}
			for _, other := range ids {
				if paths[start][other] && paths[other][start] {
					used[other] = true
					members[other] = true
					states = append(states, other)
				}
			}
			internal := []string{}
			outgoing := []string{}
			self := false
			for _, e := range m.Transitions {
				if members[e.From] {
					if members[e.To] {
						internal = append(internal, e.ID)
						if e.From == e.To {
							self = true
						}
					} else {
						outgoing = append(outgoing, e.ID)
					}
				}
			}
			if len(states) > 1 || self {
				want = append(want, Cycle{States: states, InternalTransitions: internal, OutgoingTransitions: outgoing})
			}
		}
		if !reflect.DeepEqual(got.Cycles, want) {
			t.Fatalf("mask %03x cycles got %#v want %#v", mask, got.Cycles, want)
		}
	}
	if tested < 100 {
		t.Fatalf("only %d valid graphs; sweep unexpectedly vacuous", tested)
	}
	t.Logf("checked %d valid four-state graphs", tested)
}
