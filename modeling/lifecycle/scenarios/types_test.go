package scenarios

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	lifecycle "github.com/spik3r/clanker-foundry/modeling/lifecycle/generated"
)

func minimalSuite() Suite {
	initial := lifecycle.Task{ID: "task-1", State: "Draft", Revision: 0, Outcome: "Ship outcome", Acceptance: []string{"Verified"}}
	expected := copyTask(initial)
	expected.State = "Ready"
	expected.Revision = 1
	return Suite{Version: 1, ID: "Example", ProfileVersion: lifecycle.ProfileVersion, ModelDigest: lifecycle.ModelDigest, Cases: []Case{{ID: "Happy", Initial: initial, Steps: []Step{{ID: "Prepare", Command: lifecycle.Command{Type: "prepare", ExpectedRevision: 0}, Expected: Expected{OK: true, TransitionID: "T_prepare", Task: &expected, Refusals: []string{}}}}}}}
}
func TestWholeFixtureStructuralRules(t *testing.T) {
	cases := map[string]func(*Suite){
		"version": func(s *Suite) { s.Version = 0 }, "model syntax": func(s *Suite) { s.ModelDigest = "not a digest" }, "root ID": func(s *Suite) { s.ID = "../bad" }, "global duplicate": func(s *Suite) { s.Cases[0].Steps[0].ID = s.ID }, "empty cases": func(s *Suite) { s.Cases = nil }, "empty steps": func(s *Suite) { s.Cases[0].Steps = nil }, "long scalar": func(s *Suite) { s.Cases[0].Initial.Outcome = strings.Repeat("x", MaxTextBytes+1) }, "invalid Unicode": func(s *Suite) { s.Cases[0].Initial.ID = string([]byte{255}) }, "acceptance count": func(s *Suite) { s.Cases[0].Initial.Acceptance = make([]string, MaxItems+1) }, "nil expected refusals": func(s *Suite) { s.Cases[0].Steps[0].Expected.Refusals = nil }, "OK without task": func(s *Suite) { s.Cases[0].Steps[0].Expected.Task = nil }, "OK with refusal": func(s *Suite) { s.Cases[0].Steps[0].Expected.Refusals = []string{"bad"} }, "OK without transition": func(s *Suite) { s.Cases[0].Steps[0].Expected.TransitionID = " " }, "refusal without reason": func(s *Suite) { s.Cases[0].Steps[0].Expected.OK = false }, "long refusal": func(s *Suite) {
			s.Cases[0].Steps[0].Expected.OK = false
			s.Cases[0].Steps[0].Expected.Refusals = []string{strings.Repeat("x", 257)}
		},
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			s := minimalSuite()
			mutate(&s)
			if !errors.Is(Validate(s), ErrInvalidScenario) {
				t.Fatal("invalid fixture accepted")
			}
			if b, e := Encode(s); e == nil || b != nil {
				t.Fatal("invalid fixture encoded")
			}
		})
	}
	for _, field := range []string{"profile", "digest"} {
		s := minimalSuite()
		if field == "profile" {
			s.ProfileVersion = "unknown"
		} else {
			s.ModelDigest = strings.Repeat("0", 64)
		}
		if !errors.Is(Validate(s), ErrModelBinding) {
			t.Fatal("wrong model binding accepted")
		}
	}
}
func TestRuntimeInvalidValuesRemainRepresentable(t *testing.T) {
	s := minimalSuite()
	s.Cases[0].Initial.State = "Unknown"
	s.Cases[0].Initial.Outcome = strings.Repeat("x", 4097)
	s.Cases[0].Initial.Acceptance = []string{""}
	s.Cases[0].Steps[0].Command.Type = "unknown"
	s.Cases[0].Steps[0].Command.Review = &lifecycle.Review{Decision: "unknown", ModelDigest: "invalid", TaskDigest: "invalid"}
	s.Cases[0].Steps[0].Expected = Expected{Task: nil, Refusals: []string{"INVALID_INPUT"}}
	if e := Validate(s); e != nil {
		t.Fatal("fixture duplicated business validation", e)
	}
}
func TestCountAndEncodedBounds(t *testing.T) {
	s := minimalSuite()
	s.Cases = make([]Case, MaxCases+1)
	if Validate(s) == nil {
		t.Fatal("case bound")
	}
	s = minimalSuite()
	s.Cases[0].Steps = make([]Step, MaxStepsPerCase+1)
	if Validate(s) == nil {
		t.Fatal("step bound")
	}

	s = minimalSuite()
	template := s.Cases[0]
	s.Cases = nil
	for i := 0; i < 5; i++ {
		c := template
		c.ID = fmt.Sprintf("Case%d", i)
		c.Steps = nil
		n := 32
		if i == 4 {
			n = 1
		}
		for j := 0; j < n; j++ {
			x := template.Steps[0]
			x.ID = fmt.Sprintf("Step%d_%d", i, j)
			c.Steps = append(c.Steps, x)
		}
		s.Cases = append(s.Cases, c)
	}
	if Validate(s) == nil {
		t.Fatal("aggregate step bound")
	}
	s.Cases = s.Cases[:4]
	if e := Validate(s); e != nil {
		t.Fatal("maximum aggregate step count refused", e)
	}
	s = minimalSuite()
	s.Cases[0].Steps[0].Expected.Task.Acceptance = make([]string, 64)
	for i := range s.Cases[0].Steps[0].Expected.Task.Acceptance {
		s.Cases[0].Steps[0].Expected.Task.Acceptance[i] = strings.Repeat("\x01", 8192)
	}
	if Validate(s) == nil {
		t.Fatal("encoded fixture exceeded source cap")
	}
}
func TestCopyHelpersPreserveRepresentationWithoutAliases(t *testing.T) {
	s := minimalSuite()
	owner := "owner"
	s.Cases[0].Initial.Owner = &owner
	c := copyTask(s.Cases[0].Initial)
	*c.Owner = "other"
	c.Acceptance[0] = "other"
	if owner != "owner" || s.Cases[0].Initial.Acceptance[0] != "Verified" {
		t.Fatal("task alias")
	}
	for _, a := range [][]string{nil, {}} {
		x := lifecycle.Task{Acceptance: a}
		if !reflect.DeepEqual(x, copyTask(x)) {
			t.Fatal("nil/empty changed")
		}
	}
	cmd := lifecycle.Command{Review: &lifecycle.Review{Reviewer: "one"}}
	v := copyCommand(cmd)
	v.Review.Reviewer = "two"
	if cmd.Review.Reviewer != "one" {
		t.Fatal("review alias")
	}
}
