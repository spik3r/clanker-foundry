package scenariodiff

import (
	"bytes"
	"fmt"
	"os"
	"reflect"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/canonical"
	lifecycle "github.com/spik3r/clanker-foundry/modeling/lifecycle/generated"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/jsonformat"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/model"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/scenarios"
)

func boundaryInputs(t *testing.T) (model.Model, scenarios.Suite) {
	t.Helper()
	raw, err := os.ReadFile("../examples/lifecycle.json")
	if err != nil {
		t.Fatal(err)
	}
	m, err := jsonformat.Decode(raw)
	if err != nil {
		t.Fatal(err)
	}
	d, err := canonical.Digest(m)
	if err != nil {
		t.Fatal(err)
	}
	return m, scenarios.Suite{Version: 1, ID: "ReviewSuite", ProfileVersion: model.ProfileVersion, ModelDigest: d,
		Cases: []scenarios.Case{{ID: "First", Initial: lifecycle.Task{ID: "not business valid", State: "InventedState"},
			Steps: []scenarios.Step{{ID: "FirstStep", Command: lifecycle.Command{Type: "NeverExecute", Actor: "<literal>& 雪", ExpectedRevision: ^uint64(0)},
				Expected: scenarios.Expected{OK: false, TransitionID: "NotDeclared", Refusals: []string{"NOT_A_MODEL_REASON"}}}}}}}
}

func TestTypedBoundaryAndNoExecution(t *testing.T) {
	m, s := boundaryInputs(t)
	got, err := Compare(m, s, m, s)
	if err != nil || got.ReviewRequired || got.Sources != nil {
		t.Fatalf("unexpected comparison: %+v / %v", got, err)
	}
	// These are intentionally impossible business inputs and fabricated expected
	// outcomes. Passing validates declarations only; invoking a reducer is a bug.
	bad := []func(*model.Model, *scenarios.Suite){
		func(m *model.Model, _ *scenarios.Suite) { m.Version = 0 },
		func(m *model.Model, _ *scenarios.Suite) { m.States = nil },
		func(_ *model.Model, s *scenarios.Suite) { s.Version = 0 },
		func(_ *model.Model, s *scenarios.Suite) { s.ModelDigest = strings.Repeat("0", 64) },
		func(_ *model.Model, s *scenarios.Suite) { s.ProfileVersion = "other/v1" },
		func(_ *model.Model, s *scenarios.Suite) { s.Cases = nil },
		func(_ *model.Model, s *scenarios.Suite) { s.Cases[0].Steps = nil },
		func(_ *model.Model, s *scenarios.Suite) { s.Cases[0].Steps[0].ID = s.Cases[0].ID },
		func(_ *model.Model, s *scenarios.Suite) { s.Cases[0].Steps[0].Expected.Refusals = nil },
		func(_ *model.Model, s *scenarios.Suite) { s.Cases[0].Initial.Outcome = string([]byte{0xff}) },
		func(_ *model.Model, s *scenarios.Suite) {
			s.Cases[0].Initial.Acceptance = make([]string, scenarios.MaxItems+1)
		},
		func(_ *model.Model, s *scenarios.Suite) {
			s.Cases[0].Steps[0].Command.Actor = strings.Repeat("x", scenarios.MaxTextBytes+1)
		},
	}
	for i, mutate := range bad {
		for _, side := range []string{"before", "after"} {
			t.Run(fmt.Sprintf("%d/%s", i, side), func(t *testing.T) {
				bm, bs := boundaryInputs(t)
				am, as := boundaryInputs(t)
				if side == "before" {
					mutate(&bm, &bs)
				} else {
					mutate(&am, &as)
				}
				r, err := Compare(bm, bs, am, as)
				if err != ErrInvalid || !reflect.DeepEqual(r, Report{}) {
					t.Fatal("partial or accepted invalid input", r, err)
				}
			})
		}
	}
}

func TestModelSetNormalization(t *testing.T) {
	m, s := boundaryInputs(t)
	am, as := boundaryInputs(t)
	slices.Reverse(am.States)
	slices.Reverse(am.Requirements)
	slices.Reverse(am.Guards)
	slices.Reverse(am.Transitions)
	for i := range am.Transitions {
		slices.Reverse(am.Transitions[i].Guards)
	}
	var reverse func(*model.Expr)
	reverse = func(e *model.Expr) {
		for i := range e.Args {
			reverse(&e.Args[i])
		}
		if e.Op == "all" || e.Op == "any" {
			slices.Reverse(e.Args)
		}
	}
	for i := range am.Guards {
		reverse(&am.Guards[i].Expr)
	}
	got, err := Compare(m, s, am, as)
	if err != nil || got.ModelChanged || got.SuiteChanged || got.ReviewRequired || len(got.Cases) != 0 || len(got.Steps) != 0 {
		t.Fatal(got, err)
	}
}

func TestOwnedDeterministicReadOnlyOutput(t *testing.T) {
	bm, bs := boundaryInputs(t)
	am, as := boundaryInputs(t)
	as.Cases[0].Initial.Acceptance = []string{}
	as.Cases[0].Steps[0].Command.Actor = "changed"
	br, _, err := validator(bm)
	if err != nil {
		t.Fatal(err)
	}
	ar, _, err := validator(am)
	if err != nil {
		t.Fatal(err)
	}
	braw, err := br.Encode(bs)
	if err != nil {
		t.Fatal(err)
	}
	araw, err := ar.Encode(as)
	if err != nil {
		t.Fatal(err)
	}
	got, err := Compare(bm, bs, am, as)
	if err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	for range 8 {
		wg.Go(func() {
			for range 8 {
				r, e := Compare(bm, bs, am, as)
				if e != nil || !reflect.DeepEqual(r, got) {
					t.Error("concurrent comparison differs", e)
				}
			}
		})
	}
	wg.Wait()
	got.CaseOrder.Before[0] = "mutated"
	if got.CaseOrder.After[0] == "mutated" {
		t.Fatal("case orders alias")
	}
	got.Cases[0].BeforeSteps[0] = "mutated"
	if got.Cases[0].AfterSteps[0] == "mutated" {
		t.Fatal("step orders alias")
	}
	got.Cases[0].Fields[0] = "mutated"
	got.Steps[0].Fields[0] = "mutated"
	bafter, err := br.Encode(bs)
	if err != nil {
		t.Fatal(err)
	}
	aafter, err := ar.Encode(as)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(braw, bafter) || !bytes.Equal(araw, aafter) {
		t.Fatal("input changed")
	}
}

func TestMaximumSuiteAndReportBound(t *testing.T) {
	m, s := boundaryInputs(t)
	prototype := s.Cases[0].Steps[0]
	s.Cases = make([]scenarios.Case, scenarios.MaxCases)
	for i := range s.Cases {
		s.Cases[i] = scenarios.Case{ID: fmt.Sprintf("Case_%040d", i), Steps: make([]scenarios.Step, 8)}
		for j := range s.Cases[i].Steps {
			s.Cases[i].Steps[j] = prototype
			s.Cases[i].Steps[j].ID = fmt.Sprintf("Step_%040d", 8*i+j)
		}
	}
	as := s
	as.Cases = slices.Clone(s.Cases)
	for i := range as.Cases {
		as.Cases[i].ID = "Z" + as.Cases[i].ID
		as.Cases[i].Steps = slices.Clone(s.Cases[i].Steps)
		for j := range as.Cases[i].Steps {
			as.Cases[i].Steps[j].ID = "Z" + as.Cases[i].Steps[j].ID
		}
	}
	got, err := Compare(m, s, m, as)
	if err != nil || len(got.Cases) != 32 || len(got.Steps) != 256 {
		t.Fatal("bounded suite rejected", err)
	}
	if out, err := encode(got); err != nil || len(out) >= MaxReportBytes {
		t.Fatal("report bounds", err)
	}
	if out, err := encode(Report{DiffVersion: strings.Repeat("x", MaxReportBytes)}); err != ErrLimit || out != nil {
		t.Fatal("output limit not enforced", err)
	}
}
