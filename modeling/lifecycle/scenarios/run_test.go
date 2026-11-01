package scenarios

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"testing"

	lifecycle "github.com/spik3r/clanker-foundry/modeling/lifecycle/generated"
)

func fixture(t *testing.T) ([]byte, Suite) {
	t.Helper()
	raw, e := os.ReadFile("testdata/lifecycle-v1.json")
	if e != nil {
		t.Fatal(e)
	}
	s, e := Decode(raw)
	if e != nil {
		t.Fatal(e)
	}
	return raw, s
}
func TestFrozenFixtureReplayAndDeterminism(t *testing.T) {
	raw, s := fixture(t)
	before, _ := Encode(s)
	r, e := Run(s)
	if e != nil || !r.Matched || r.CasesCompleted != 4 || r.StepsExecuted != 13 || r.SourceDigest != nil {
		t.Fatalf("typed run: %+v %v", r, e)
	}
	wire, e := Execute(raw)
	if e != nil || wire.SourceDigest == nil || *wire.SourceDigest != "ef199aa93c9e70887f3da8e72aa4a79dcaf130c4cec72d27b240d6780a3e4b26" {
		t.Fatal("source identity", e)
	}
	encoded, e := EncodeReport(wire)
	if e != nil {
		t.Fatal(e)
	}
	for range 3 {
		again, e := Execute(raw)
		if e != nil {
			t.Fatal(e)
		}
		b, e := EncodeReport(again)
		if e != nil || !bytes.Equal(encoded, b) {
			t.Fatal("nondeterministic bytes", e)
		}
	}
	after, _ := Encode(s)
	if !bytes.Equal(before, after) {
		t.Fatal("input changed")
	}
	if r.Entries[5].Actual.OK || r.Entries[5].Actual.Task != nil || !r.Entries[5].Matched || !r.Entries[6].Actual.OK {
		t.Fatal("INVALID_INPUT did not preserve current task")
	}
}
func TestPreflightRejectsLaterFaultBeforeAnyCall(t *testing.T) {
	s := minimalSuite()
	second := s.Cases[0]
	second.ID = "Other"
	second.Steps = append([]Step(nil), second.Steps...)
	second.Steps[0].ID = "Last"
	second.Steps[0].Expected = copyExpected(second.Steps[0].Expected)
	second.Steps[0].Expected.Task = nil
	s.Cases = append(s.Cases, second)
	calls := 0
	apply := func(t lifecycle.Task, c lifecycle.Command) lifecycle.Result {
		calls++
		return lifecycle.Transition(t, c)
	}
	r, e := runWith(s, apply)
	if !errors.Is(e, ErrInvalidScenario) || calls != 0 || !reflect.DeepEqual(r, Report{}) {
		t.Fatal("late malformed fixture executed", calls, e)
	}
	s = minimalSuite()
	s.ModelDigest = strings.Repeat("0", 64)
	if _, e := runWith(s, apply); !errors.Is(e, ErrModelBinding) || calls != 0 {
		t.Fatal("binding executed")
	}
	raw, _ := Encode(minimalSuite())
	raw = bytes.Replace(raw, []byte(`"actor":""`), []byte(`"actor":"","actor":"other"`), 1)
	if _, e := executeWith(raw, apply); e == nil || calls != 0 {
		t.Fatal("duplicate fixture executed")
	}
}
func TestFirstMismatchStopsAndDoesNotRefreshExpected(t *testing.T) {
	s := minimalSuite()
	s.Cases[0].Steps[0].Expected.Task.State = "Done"
	next := s.Cases[0].Steps[0]
	next.ID = "Later"
	s.Cases[0].Steps = append(s.Cases[0].Steps, next)
	before, _ := Encode(s)
	calls := 0
	r, e := runWith(s, func(t lifecycle.Task, c lifecycle.Command) lifecycle.Result {
		calls++
		return lifecycle.Transition(t, c)
	})
	if e != nil || r.Matched || calls != 1 || r.StepsExecuted != 1 || r.CasesCompleted != 0 || !reflect.DeepEqual(r.Entries[0].Differences, []string{"task"}) {
		t.Fatalf("mismatch: %+v %v", r, e)
	}
	after, _ := Encode(s)
	if !bytes.Equal(before, after) {
		t.Fatal("expected values refreshed")
	}
	if _, e := EncodeReport(r); e != nil {
		t.Fatal(e)
	}
}
func TestCompleteResultFieldsCompared(t *testing.T) {
	mutations := map[string]func(*lifecycle.Result){"mode": func(r *lifecycle.Result) { r.Mode = "production" }, "authority": func(r *lifecycle.Result) { r.Authority = "claimed" }, "modelDigest": func(r *lifecycle.Result) { r.ModelDigest = strings.Repeat("0", 64) }, "ok": func(r *lifecycle.Result) { r.OK = false }, "transitionId": func(r *lifecycle.Result) { r.TransitionID = "other" }, "task": func(r *lifecycle.Result) { r.Task.Revision++ }, "refusals": func(r *lifecycle.Result) { r.Refusals = []string{"other"} }}
	for field, mutate := range mutations {
		t.Run(field, func(t *testing.T) {
			r, e := runWith(minimalSuite(), func(t lifecycle.Task, c lifecycle.Command) lifecycle.Result {
				x := lifecycle.Transition(t, c)
				mutate(&x)
				return x
			})
			if e != nil || r.Matched || !reflect.DeepEqual(r.Entries[0].Differences, []string{field}) {
				t.Fatalf("ignored field %+v %v", r, e)
			}
		})
	}
	for _, change := range []func(*lifecycle.Result){func(r *lifecycle.Result) { r.Task.Acceptance = nil }, func(r *lifecycle.Result) { v := ""; r.Task.Owner = &v }} {
		r, e := runWith(minimalSuite(), func(t lifecycle.Task, c lifecycle.Command) lifecycle.Result {
			x := lifecycle.Transition(t, c)
			change(&x)
			return x
		})
		if e != nil || r.Matched {
			t.Fatal("representation difference ignored")
		}
	}
}
func TestExpectedRefusalOrderIsExact(t *testing.T) {
	s := minimalSuite()
	s.Cases[0].Initial.State = "Ready"
	x := &s.Cases[0].Steps[0]
	x.Command = lifecycle.Command{Type: "start", ExpectedRevision: 9}
	current := copyTask(s.Cases[0].Initial)
	x.Expected = Expected{TransitionID: "T_start", Task: &current, Refusals: []string{"G_revision", "G_actor"}}
	r, e := Run(s)
	if e != nil || r.Matched || !reflect.DeepEqual(r.Entries[0].Differences, []string{"refusals"}) {
		t.Fatal("refusal order normalized", e)
	}
}
func TestReportAndInputDoNotAlias(t *testing.T) {
	_, s := fixture(t)
	before, _ := Encode(s)
	r, e := Run(s)
	if e != nil {
		t.Fatal(e)
	}
	stable, e := EncodeReport(r)
	if e != nil {
		t.Fatal(e)
	}
	s.Cases[0].Initial.Acceptance[0] = "changed input"
	s.Cases[0].Steps[3].Command.Review.Reviewer = "changed reviewer"
	after, e := EncodeReport(r)
	if e != nil || !bytes.Equal(stable, after) {
		t.Fatal("report aliases suite")
	}
	_, s = fixture(t)
	r, e = Run(s)
	if e != nil {
		t.Fatal(e)
	}
	siblingBefore := copyTask(r.Entries[1].Input)
	r.Entries[0].Input.Acceptance[0] = "input copy"
	r.Entries[0].Expected.Task.Acceptance[0] = "expected copy"
	r.Entries[0].Actual.Task.Acceptance[0] = "actual copy"
	r.Entries[3].Command.Review.Reviewer = "review copy"
	after, _ = Encode(s)
	if !bytes.Equal(before, after) {
		t.Fatal("suite aliases report")
	}
	if !reflect.DeepEqual(r.Entries[1].Input, siblingBefore) {
		t.Fatal("cross-step result alias")
	}
}
func TestIncrementalReportLimitStopsEvaluation(t *testing.T) {
	s := minimalSuite()
	template := s.Cases[0]
	template.Initial.Acceptance = make([]string, 64)
	for i := range template.Initial.Acceptance {
		template.Initial.Acceptance[i] = strings.Repeat("x", 4096)
	}
	s.Cases = nil
	for c := 0; c < 2; c++ {
		v := Case{ID: fmt.Sprintf("Case%d", c), Initial: template.Initial}
		for step := 0; step < 32; step++ {
			v.Steps = append(v.Steps, Step{ID: fmt.Sprintf("Step%d_%d", c, step), Command: lifecycle.Command{Type: "prepare"}, Expected: Expected{Refusals: []string{"INVALID_INPUT"}}})
		}
		s.Cases = append(s.Cases, v)
	}
	if e := Validate(s); e != nil {
		t.Fatal(e)
	}
	calls := 0
	r, e := runWith(s, func(t lifecycle.Task, c lifecycle.Command) lifecycle.Result {
		calls++
		return lifecycle.Transition(t, c)
	})
	if !errors.Is(e, ErrReportLimit) || calls <= 1 || calls >= 64 || !reflect.DeepEqual(r, Report{}) {
		t.Fatal("report budget did not stop", calls, e)
	}
}
func TestAlwaysRefuseMutantIsDetected(t *testing.T) {
	s := minimalSuite()
	r, e := runWith(s, func(task lifecycle.Task, _ lifecycle.Command) lifecycle.Result {
		return lifecycle.Result{Mode: Mode, Authority: Authority, ModelDigest: lifecycle.ModelDigest, Task: &task, Refusals: []string{"MUTANT"}}
	})
	if e != nil || r.Matched {
		t.Fatal("always refuse admitted")
	}
}
func TestConcurrentReplayOnImmutableInput(t *testing.T) {
	_, s := fixture(t)
	var wg sync.WaitGroup
	for range 8 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range 4 {
				r, e := Run(s)
				if e != nil || !r.Matched {
					t.Errorf("concurrent replay %v", e)
					return
				}
			}
		}()
	}
	wg.Wait()
}
func TestReportEncoderBoundsAndEnvelope(t *testing.T) {
	r, e := Run(minimalSuite())
	if e != nil {
		t.Fatal(e)
	}
	raw, e := EncodeReport(r)
	if e != nil {
		t.Fatal(e)
	}
	var d map[string]any
	if e := json.Unmarshal(raw, &d); e != nil {
		t.Fatal(e)
	}
	if d["mode"] != Mode || d["authority"] != Authority || d["sourceDigest"] != nil {
		t.Fatal("report envelope")
	}
	r.Entries[0].Actual.Mode = strings.Repeat("x", MaxTextBytes+1)
	if b, e := EncodeReport(r); !errors.Is(e, ErrReportLimit) || b != nil {
		t.Fatal("unbounded report")
	}
	r, _ = Run(minimalSuite())
	r.Entries[0].Input.Acceptance = make([]string, 64)
	for i := range r.Entries[0].Input.Acceptance {
		r.Entries[0].Input.Acceptance[i] = strings.Repeat("\x01", 8192)
	}
	r.Entries = append(r.Entries, r.Entries[0], r.Entries[0])
	for i := range r.Entries {
		r.Entries[i].StepID = fmt.Sprintf("OutputStep%d", i)
		r.Entries[i].CaseID = fmt.Sprintf("OutputCase%d", i)
	}
	r.StepsExecuted = 3
	r.CasesCompleted = 3
	if b, e := EncodeReport(r); !errors.Is(e, ErrReportLimit) || b != nil {
		t.Fatal("encoded report cap")
	}
}

func TestEncoderRejectsInternallyContradictoryReports(t *testing.T) {
	mutations := map[string]func(*Report){
		"actual":        func(r *Report) { r.Entries[0].Actual.Mode = "production" },
		"case count":    func(r *Report) { r.CasesCompleted = 0 },
		"step count":    func(r *Report) { r.StepsExecuted = 2 },
		"difference":    func(r *Report) { r.Entries[0].Differences = []string{"task"} },
		"entry match":   func(r *Report) { r.Entries[0].Matched = false },
		"overall match": func(r *Report) { r.Matched = false },
	}
	for name, mutate := range mutations {
		t.Run(name, func(t *testing.T) {
			r, e := Run(minimalSuite())
			if e != nil {
				t.Fatal(e)
			}
			mutate(&r)
			if b, e := EncodeReport(r); !errors.Is(e, ErrReportLimit) || b != nil {
				t.Fatal("contradictory report encoded")
			}
		})
	}
}

// These tests edit only scratch copies of the actual generated reducer. The
// independent fixture and acceptance file are copied unchanged, not regenerated.
func TestFrozenAcceptanceKillsReducerMutations(t *testing.T) {
	for _, mutation := range []string{"always-refuse", "G_reviewContent", "G_actor"} {
		t.Run(mutation, func(t *testing.T) {
			dir := t.TempDir()
			files := map[string]string{"go.mod": "../go.mod", "model/ir.go": "../model/ir.go", "strictjson/decode.go": "../strictjson/decode.go", "scenarios/types.go": "types.go", "scenarios/decode.go": "decode.go", "scenarios/run.go": "run.go", "scenarios/runner.go": "runner.go", "scenarioacceptance/oracle_test.go": "../scenarioacceptance/oracle_test.go", "scenarios/testdata/lifecycle-v1.json": "testdata/lifecycle-v1.json", "generated/lifecycle.go": "../generated/lifecycle.go"}
			for dest, source := range files {
				b, e := os.ReadFile(source)
				if e != nil {
					t.Fatal(e)
				}
				p := filepath.Join(dir, dest)
				if e := os.MkdirAll(filepath.Dir(p), 0755); e != nil {
					t.Fatal(e)
				}
				if e := os.WriteFile(p, b, 0644); e != nil {
					t.Fatal(e)
				}
			}
			p := filepath.Join(dir, "generated/lifecycle.go")
			b, e := os.ReadFile(p)
			if e != nil {
				t.Fatal(e)
			}
			if mutation == "always-refuse" {
				needle := []byte("func Transition(task Task, command Command) Result {")
				if bytes.Count(b, needle) != 1 {
					t.Fatal("mutation anchor")
				}
				b = bytes.Replace(b, needle, append(append([]byte{}, needle...), []byte(`
if len(task.ID)>0{return refused(task,"",[]string{"MUTANT"})}
`)...), 1)
			} else {
				re := regexp.MustCompile(`(?s)func guard_` + regexp.QuoteMeta(mutation) + `\(task Task, command Command, currentDigest string\) bool \{.*?\n\}`)
				if len(re.FindAll(b, -1)) != 1 {
					t.Fatal("guard mutation anchor")
				}
				b = re.ReplaceAll(b, []byte("func guard_"+mutation+"(task Task, command Command, currentDigest string) bool {return true}"))
			}
			if e := os.WriteFile(p, b, 0644); e != nil {
				t.Fatal(e)
			}
			cmd := exec.Command(filepath.Join(runtime.GOROOT(), "bin", "go"), "test", "-count=1", "./scenarioacceptance")
			cmd.Dir = dir
			cmd.Env = append(os.Environ(), "GOTOOLCHAIN=local", "GOPROXY=off", "GOSUMDB=off")
			out, e := cmd.CombinedOutput()
			if e == nil || !bytes.Contains(out, []byte("--- FAIL:")) || bytes.Contains(out, []byte("build failed")) {
				t.Fatalf("mutant was not killed behaviorally: %v\n%s", e, out)
			}
		})
	}
}

func TestConstructedReportGroupBounds(t *testing.T) {
	base, e := Run(minimalSuite())
	if e != nil {
		t.Fatal(e)
	}
	r := base
	r.Entries = nil
	for i := 0; i < 33; i++ {
		x := base.Entries[0]
		x.StepID = fmt.Sprintf("Step%d", i)
		if i > 0 {
			x.Input = copyTask(*base.Entries[0].Actual.Task)
		}
		r.Entries = append(r.Entries, x)
	}
	r.StepsExecuted = len(r.Entries)
	if b, e := EncodeReport(r); !errors.Is(e, ErrReportLimit) || b != nil {
		t.Fatal("oversized case group encoded")
	}
	r = base
	r.Entries = nil
	r.Matched = false
	r.CasesCompleted = 16
	for i := 0; i < 17; i++ {
		x := base.Entries[0]
		x.CaseID = fmt.Sprintf("Case%d", i)
		x.StepID = fmt.Sprintf("Step%d", i)
		if i == 16 {
			x.Actual.OK = false
			x.Matched = false
			x.Differences = []string{"ok"}
		}
		r.Entries = append(r.Entries, x)
	}
	r.StepsExecuted = len(r.Entries)
	if b, e := EncodeReport(r); !errors.Is(e, ErrReportLimit) || b != nil {
		t.Fatal("seventeenth mismatched group encoded")
	}
}

func TestEncoderRejectsBrokenWithinCaseContinuity(t *testing.T) {
	_, s := fixture(t)
	for _, index := range []int{1, 5} {
		r, e := Run(s)
		if e != nil {
			t.Fatal(e)
		}
		r.Entries[index].Input.Revision++
		if b, e := EncodeReport(r); !errors.Is(e, ErrReportLimit) || b != nil {
			t.Fatal("inconsistent next input encoded", index)
		}
	}
}
