package modelmerge

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/model"
)

func copyModel(t *testing.T, input model.Model) model.Model {
	t.Helper()
	out, err := model.Normalize(input)
	if err != nil {
		t.Fatal(err)
	}
	return out
}

func TestValidationPrecedesConflicts(t *testing.T) {
	for _, position := range []int{0, 1, 2} {
		for _, fault := range []string{"zero", "cycle", "bounds", "unicode"} {
			t.Run(fault+string(rune('0'+position)), func(t *testing.T) {
				inputs := [3]model.Model{oracleBase(), oracleBase(), oracleBase()}
				inputs[1].ID = "Left"
				inputs[2].ID = "Right"
				switch fault {
				case "zero":
					inputs[position] = model.Model{}
				case "cycle":
					args := make([]model.Expr, 1)
					args[0] = model.Expr{Kind: "op", Op: "all", Args: args}
					inputs[position].Guards[0].Expr = args[0]
				case "bounds":
					inputs[position].Requirements[0].Text = strings.Repeat("x", 4097)
				case "unicode":
					inputs[position].States[0].Label = string([]byte{0xff})
				}
				got, err := Merge(inputs[0], inputs[1], inputs[2])
				if !errors.Is(err, ErrInvalid) || !reflect.DeepEqual(got, Result{}) {
					t.Fatalf("%+v %v", got, err)
				}
			})
		}
	}
}

func TestNormalizationPreservesOrderedEffects(t *testing.T) {
	base := oracleBase()
	base.Guards[0].Expr = model.Expr{Kind: "op", Op: "all", Args: []model.Expr{
		{Kind: "op", Op: "nonempty", Args: []model.Expr{{Kind: "ref", Ref: "task.outcome"}}},
		{Kind: "op", Op: "any", Args: []model.Expr{
			{Kind: "op", Op: "nonempty", Args: []model.Expr{{Kind: "ref", Ref: "command.actor"}}},
			{Kind: "op", Op: "absent", Args: []model.Expr{{Kind: "ref", Ref: "task.owner"}}},
		}},
	}}
	base.Guards = append(base.Guards, model.Guard{ID: "G_other", Requirement: "R", Expr: model.Expr{Kind: "op", Op: "nonempty", Args: []model.Expr{{Kind: "ref", Ref: "task.acceptance"}}}})
	base.Transitions[0].Guards = []string{"G_other", "G"}
	left := copyModel(t, base)
	right := copyModel(t, base)
	var reverse func(*model.Expr)
	reverse = func(e *model.Expr) {
		for i := range e.Args {
			reverse(&e.Args[i])
		}
		if e.Op == "all" || e.Op == "any" {
			for i, j := 0, len(e.Args)-1; i < j; i, j = i+1, j-1 {
				e.Args[i], e.Args[j] = e.Args[j], e.Args[i]
			}
		}
	}
	reverse(&left.Guards[0].Expr)
	left.Transitions[0].Guards = []string{"G_other", "G"}
	right.Requirements[0].Text = "Changed requirement"
	got, err := Merge(base, left, right)
	if err != nil || got.Status != Merged || !reflect.DeepEqual(got.Candidate, &right) {
		t.Fatalf("normalization produced %v / %+v", err, got)
	}
	left = copyModel(t, base)
	left.Transitions[0].Effects = []model.Effect{{Op: model.IncrementRevision}, {Op: model.SetState}}
	if got, err = Merge(base, left, right); !errors.Is(err, ErrInvalid) || !reflect.DeepEqual(got, Result{}) {
		t.Fatal("effect order was normalized or accepted")
	}
}

func TestOwnedOutputsAndConcurrentCalls(t *testing.T) {
	base := oracleBase()
	left := oracleAdded()
	right := oracleBase()
	snapshots := [3]model.Model{copyModel(t, base), copyModel(t, left), copyModel(t, right)}
	got, err := Merge(base, left, right)
	if err != nil || got.Candidate == nil {
		t.Fatal(err, got.Status)
	}
	got.Candidate.States[0].Label = "mutated"
	got.Candidate.Guards[0].Expr.Args[0].Ref = "command.actor"
	got.Candidate.Transitions[0].Guards[0] = "missing"
	got.Candidate.Transitions[0].Effects[0].Op = "missing"
	if !reflect.DeepEqual([3]model.Model{base, left, right}, snapshots) {
		t.Fatal("result aliases input")
	}
	want, err := Merge(base, left, right)
	if err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	for range 16 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			result, e := Merge(base, left, right)
			if e != nil || !reflect.DeepEqual(result, want) {
				t.Error("parallel immutable result differs", e)
			}
		}()
	}
	wg.Wait()
	base.States[0].Label = "changed after return"
	left.Guards[0].Expr.Args[0].Ref = "command.actor"
	if want.Candidate.States[0].Label != "Draft" || want.Candidate.Guards[0].Expr.Args[0].Ref != "task.outcome" {
		t.Fatal("input aliases result")
	}
}

func TestCombinedBoundsRefuseWithoutPartialCandidate(t *testing.T) {
	base := oracleBase()
	add := func(prefix string) model.Model {
		m := oracleBase()
		for i := 0; i < 8; i++ {
			id := prefix + string(rune('A'+i))
			m.States = append(m.States, model.State{ID: id, Label: id, Terminal: true})
			m.Transitions = append(m.Transitions, model.Transition{ID: "T_" + id, From: "A", To: id, Trigger: "go_" + id, Guards: []string{"G"}, Effects: []model.Effect{{Op: model.SetState}, {Op: model.IncrementRevision}}})
		}
		return m
	}
	left, right := add("L"), add("R")
	got, err := Merge(base, left, right)
	if err != nil || got.Status != InvalidCombination || got.Candidate != nil || got.CandidateDigest != nil || got.Conflicts == nil || len(got.Conflicts) != 0 {
		t.Fatalf("%+v %v", got, err)
	}
}

func TestTypedModelMustRemainJSONRepresentable(t *testing.T) {
	m := oracleBase()
	m.Requirements = nil
	m.Guards = nil
	m.Transitions[0].Guards = nil
	for i := 0; i < 64; i++ {
		r, g := fmt.Sprintf("R_%d", i), fmt.Sprintf("G_%d", i)
		m.Requirements = append(m.Requirements, model.Requirement{ID: r, Text: strings.Repeat("\x01", 4096)})
		m.Guards = append(m.Guards, model.Guard{ID: g, Requirement: r, Expr: model.Expr{Kind: "op", Op: "nonempty", Args: []model.Expr{{Kind: "ref", Ref: "task.outcome"}}}})
		m.Transitions[0].Guards = append(m.Transitions[0].Guards, g)
	}
	if err := model.Validate(m); err != nil {
		t.Fatal("fixture must pass semantic bounds", err)
	}
	got, err := Merge(m, oracleBase(), oracleBase())
	if !errors.Is(err, ErrInvalid) || !reflect.DeepEqual(got, Result{}) {
		t.Fatal("unrepresentable input accepted", err)
	}
}
