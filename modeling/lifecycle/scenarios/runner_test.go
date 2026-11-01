package scenarios

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"
	"sync"
	"testing"

	lifecycle "github.com/spik3r/clanker-foundry/modeling/lifecycle/generated"
)

func boundDefault(t *testing.T, apply func(lifecycle.Task, lifecycle.Command) lifecycle.Result) Runner {
	t.Helper()
	r, err := NewRunner(Binding{lifecycle.ProfileVersion, lifecycle.ModelDigest}, apply)
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func TestLegacyByteParity(t *testing.T) {
	golden, err := os.ReadFile("testdata/legacy-parity.json")
	if err != nil {
		t.Fatal(err)
	}
	if sum := sha256.Sum256(golden); hex.EncodeToString(sum[:]) != "58194fa32fc0d929248892533ca21e8188c6546def68a393cab630f4f9385fa8" {
		t.Fatal("pre-implementation compatibility snapshot changed")
	}
	var want map[string]string
	if json.Unmarshal(golden, &want) != nil {
		t.Fatal("invalid compatibility snapshot")
	}
	check := func(key string, raw []byte, err error) {
		t.Helper()
		if err != nil {
			t.Fatalf("%s: %v", key, err)
		}
		h := sha256.Sum256(raw)
		if hex.EncodeToString(h[:]) != want[key] {
			t.Fatalf("%s byte parity", key)
		}
	}
	js := func(key string, v any) { b, e := json.Marshal(v); check(key, b, e) }
	raw, err := os.ReadFile("testdata/lifecycle-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	s, err := Decode(raw)
	if err != nil || Validate(s) != nil {
		t.Fatal("legacy decode/validate", err)
	}
	js("decode", s)
	b, err := Encode(s)
	check("encode", b, err)
	digest, err := Digest(s)
	if err != nil || digest != want["digest"] {
		t.Fatal("legacy canonical identity", err)
	}
	r, err := Run(s)
	if err != nil {
		t.Fatal(err)
	}
	js("run", r)
	b, err = EncodeReport(r)
	check("encodeRunReport", b, err)
	r, err = Execute(raw)
	if err != nil {
		t.Fatal(err)
	}
	js("execute", r)
	b, err = EncodeReport(r)
	check("encodeExecuteReport", b, err)
	s.ModelDigest = strings.Repeat("0", 64)
	if e := Validate(s); e == nil || e.Error() != want["staleValidate"] {
		t.Fatal("legacy stale error", e)
	}
	if _, e := Decode([]byte("{}")); e == nil || e.Error() != want["emptyDecode"] {
		t.Fatal("legacy malformed error", e)
	}
}

func TestRunnerRejectsInvalidConstructionAndEveryZeroOperation(t *testing.T) {
	for _, b := range []Binding{{}, {"other", lifecycle.ModelDigest}, {lifecycle.ProfileVersion, "bad"}, {lifecycle.ProfileVersion, strings.Repeat("A", 64)}} {
		if _, err := NewRunner(b, lifecycle.Transition); !errors.Is(err, ErrInvalidRunner) {
			t.Fatal("invalid binding accepted", b, err)
		}
	}
	if _, err := NewRunner(defaultBinding(), nil); !errors.Is(err, ErrInvalidRunner) {
		t.Fatal("nil callback accepted")
	}
	var r Runner
	s := minimalSuite()
	_, e1 := r.ModelBinding()
	e2 := r.Validate(s)
	_, e3 := r.Decode(nil)
	_, e4 := r.Encode(s)
	_, e5 := r.Digest(s)
	_, e6 := r.Run(s)
	_, e7 := r.Execute(nil)
	_, e8 := r.EncodeReport(Report{})
	for _, err := range []error{e1, e2, e3, e4, e5, e6, e7, e8} {
		if !errors.Is(err, ErrInvalidRunner) {
			t.Fatal("zero runner silently defaulted", err)
		}
	}
}

func TestRunnerBindingSnapshotAndCrossRunnerIsolation(t *testing.T) {
	declared := defaultBinding()
	r, err := NewRunner(declared, lifecycle.Transition)
	if err != nil {
		t.Fatal(err)
	}
	declared.ModelDigest = strings.Repeat("b", 64)
	got, err := r.ModelBinding()
	if err != nil || got != defaultBinding() {
		t.Fatal("configuration was not snapshotted")
	}
	got.ModelDigest = declared.ModelDigest
	s := minimalSuite()
	calls := 0
	other, err := NewRunner(declared, func(t lifecycle.Task, c lifecycle.Command) lifecycle.Result {
		calls++
		return lifecycle.Transition(t, c)
	})
	if err != nil {
		t.Fatal(err)
	}
	b, _ := r.Encode(s)
	if other.Validate(s) != ErrModelBinding {
		t.Fatal("cross validation")
	}
	if _, err := other.Decode(b); err != ErrModelBinding {
		t.Fatal("cross decode")
	}
	if _, err := other.Encode(s); err != ErrModelBinding {
		t.Fatal("cross encode")
	}
	if _, err := other.Digest(s); err != ErrModelBinding {
		t.Fatal("cross digest")
	}
	if _, err := other.Run(s); err != ErrModelBinding {
		t.Fatal("cross run")
	}
	if _, err := other.Execute(b); err != ErrModelBinding {
		t.Fatal("cross execute")
	}
	result, err := r.Run(s)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := other.EncodeReport(result); err != ErrReportLimit {
		t.Fatal("cross report")
	}
	if calls != 0 {
		t.Fatal("cross-runner input executed")
	}
	// Accepting a different declaration does not authenticate its callback.
	s.ModelDigest = declared.ModelDigest
	if other.Validate(s) != nil {
		t.Fatal("custom fixture must bind to its configured runner")
	}
	if _, err := other.Run(s); err != ErrProviderProtocol || calls != 1 {
		t.Fatal("wrong actual identity must fail after invocation", calls, err)
	}
}

func TestRunnerWholeSuitePreflightAndFailFast(t *testing.T) {
	calls := 0
	r := boundDefault(t, func(task lifecycle.Task, c lifecycle.Command) lifecycle.Result {
		calls++
		return lifecycle.Transition(task, c)
	})
	s := minimalSuite()
	second := s.Cases[0]
	second.ID = "last_case"
	second.Steps = append([]Step{}, second.Steps...)
	second.Steps[0].ID = "last_step"
	second.Steps[0].Expected.Refusals = nil
	s.Cases = append(s.Cases, second)
	if _, err := r.Run(s); err != ErrInvalidScenario || calls != 0 {
		t.Fatal("invalid last step ran", calls, err)
	}
	s.Cases[1].Steps[0].Expected.Refusals = []string{}
	s.Cases[0].Steps[0].Expected.TransitionID = "wrong_literal"
	result, err := r.Run(s)
	if err != nil || result.Matched || calls != 1 || result.StepsExecuted != 1 || result.CasesCompleted != 0 {
		t.Fatal("ordinary mismatch must stop", result, calls, err)
	}
}

func TestRunnerProtocolFailuresBeforeRetainingResult(t *testing.T) {
	mutations := map[string]func(*lifecycle.Result){
		"model":            func(r *lifecycle.Result) { r.ModelDigest = strings.Repeat("b", 64) },
		"mode":             func(r *lifecycle.Result) { r.Mode = "production" },
		"authority":        func(r *lifecycle.Result) { r.Authority = "owner" },
		"missing task":     func(r *lifecycle.Result) { r.Task = nil },
		"missing refusals": func(r *lifecycle.Result) { r.Refusals = nil },
		"oversized list":   func(r *lifecycle.Result) { r.Task.Acceptance = make([]string, MaxItems+1) },
		"oversized text":   func(r *lifecycle.Result) { r.Task.Outcome = strings.Repeat("x", MaxTextBytes+1) },
		"invalid utf8":     func(r *lifecycle.Result) { r.Task.ID = string([]byte{0xff}) },
	}
	for name, change := range mutations {
		t.Run(name, func(t *testing.T) {
			calls := 0
			r := boundDefault(t, func(task lifecycle.Task, c lifecycle.Command) lifecycle.Result {
				calls++
				actual := lifecycle.Transition(task, c)
				change(&actual)
				return actual
			})
			result, err := r.Run(minimalSuite())
			if err != ErrProviderProtocol || calls != 1 || !reflect.DeepEqual(result, Report{}) {
				t.Fatal("invalid provider result retained or misclassified", calls, err)
			}
		})
	}
}

func TestRunnerMetadataOwnershipAndSourceIdentities(t *testing.T) {
	s := minimalSuite()
	before, _ := Encode(s)
	var retained lifecycle.Result
	r := boundDefault(t, func(task lifecycle.Task, c lifecycle.Command) lifecycle.Result {
		retained = lifecycle.Transition(task, c)
		task.Acceptance[0] = "callback mutation"
		return retained
	})
	result, err := r.Run(s)
	if err != nil || !result.Matched || result.SourceDigest != nil {
		t.Fatal(result, err)
	}
	retained.Task.Acceptance[0] = "late callback mutation"
	after, _ := Encode(s)
	if !bytes.Equal(before, after) || result.Entries[0].Actual.Task.Acceptance[0] != "Verified" {
		t.Fatal("input or callback output aliased")
	}
	encoded, err := r.EncodeReport(result)
	if err != nil || !bytes.Contains(encoded, []byte(`"identitySource":"caller-supplied-go-binding"`)) || !bytes.Contains(encoded, []byte(BoundReportVersion)) {
		t.Fatal("bound report declaration missing", err)
	}
	r = boundDefault(t, lifecycle.Transition)
	x, err := r.Execute(before)
	if err != nil {
		t.Fatal(err)
	}
	y, err := r.Execute(append(append([]byte{}, before...), '\n'))
	if err != nil || x.ScenarioDigest != y.ScenarioDigest || x.SourceDigest == nil || y.SourceDigest == nil || *x.SourceDigest == *y.SourceDigest {
		t.Fatal("source/canonical identity conflated", err)
	}
	x.Entries[0].Actual.Mode = "changed"
	if raw, err := r.EncodeReport(x); raw != nil || err != ErrReportLimit {
		t.Fatal("tampered provider report encoded", err)
	}
}

func TestRunnerPanicPolicyAndConcurrentPureUse(t *testing.T) {
	r := boundDefault(t, func(lifecycle.Task, lifecycle.Command) lifecycle.Result { panic("trusted-callback-panic") })
	func() {
		defer func() {
			if recover() != "trusted-callback-panic" {
				t.Fatal("callback panic must propagate unchanged")
			}
		}()
		_, _ = r.Run(minimalSuite())
	}()
	r = boundDefault(t, lifecycle.Transition)
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Go(func() {
			report, err := r.Run(minimalSuite())
			if err != nil || !report.Matched {
				t.Errorf("pure concurrent provider: %v", err)
			}
		})
	}
	wg.Wait()
}

func TestRunnerIncrementalReportAndSourceBounds(t *testing.T) {
	s := minimalSuite()
	initial := s.Cases[0].Initial
	initial.Acceptance = make([]string, MaxItems)
	for i := range initial.Acceptance {
		initial.Acceptance[i] = strings.Repeat("x", 4096)
	}
	s.Cases = nil
	for c := 0; c < 2; c++ {
		v := Case{ID: fmt.Sprintf("case_%d", c), Initial: initial}
		for step := 0; step < MaxStepsPerCase; step++ {
			v.Steps = append(v.Steps, Step{ID: fmt.Sprintf("step_%d_%d", c, step), Command: lifecycle.Command{Type: "prepare"}, Expected: Expected{Refusals: []string{"INVALID_INPUT"}}})
		}
		s.Cases = append(s.Cases, v)
	}
	calls := 0
	r := boundDefault(t, func(task lifecycle.Task, c lifecycle.Command) lifecycle.Result {
		calls++
		return lifecycle.Transition(task, c)
	})
	if err := r.Validate(s); err != nil {
		t.Fatal(err)
	}
	result, err := r.Run(s)
	if err != ErrReportLimit || calls <= 1 || calls >= 64 || !reflect.DeepEqual(result, Report{}) {
		t.Fatal("bound report budget did not stop subsequent calls", calls, err)
	}
	calls = 0
	if _, err := r.Execute(make([]byte, MaxSourceBytes+1)); err != ErrInvalidScenario || calls != 0 {
		t.Fatal("oversized source reached provider", calls, err)
	}
	s = minimalSuite()
	s.Cases[0].Initial.Outcome = strings.Repeat("x", MaxTextBytes+1)
	if _, err := r.Run(s); err != ErrInvalidScenario || calls != 0 {
		t.Fatal("oversized typed suite reached provider", calls, err)
	}
}

func TestRunnerFreezesLiteralsAndSourceBeforeEffectfulCallback(t *testing.T) {
	s := minimalSuite()
	second := minimalSuite().Cases[0]
	second.ID = "second_case"
	second.Steps[0].ID = "second_step"
	s.Cases = append(s.Cases, second)
	calls := 0
	r := boundDefault(t, func(task lifecycle.Task, command lifecycle.Command) lifecycle.Result {
		calls++
		// This mutates caller state deliberately, not the callback's copied args.
		s.Cases[1].Initial.State = "invalid"
		s.Cases[1].Steps[0].Expected.Task.Outcome = "changed after preflight"
		s.Cases[1].Steps[0].Expected.Refusals = nil
		s.Cases[1].ID = "rewritten_case"
		s.Cases[1].Steps[0].ID = "rewritten_step"
		s.ModelDigest = strings.Repeat("b", 64)
		return lifecycle.Transition(task, command)
	})
	result, err := r.Run(s)
	if err != nil || !result.Matched || calls != 2 || result.StepsExecuted != 2 || result.ModelDigest != lifecycle.ModelDigest || result.Entries[1].CaseID != "second_case" || result.Entries[1].StepID != "second_step" {
		t.Fatal("callback rewrote a later frozen literal", result, calls, err)
	}
	raw, err := Encode(minimalSuite())
	if err != nil {
		t.Fatal(err)
	}
	want := sha256.Sum256(raw)
	r = boundDefault(t, func(task lifecycle.Task, command lifecycle.Command) lifecycle.Result {
		raw[0] = '!'
		return lifecycle.Transition(task, command)
	})
	result, err = r.Execute(raw)
	if err != nil || !result.Matched || result.SourceDigest == nil || *result.SourceDigest != hex.EncodeToString(want[:]) {
		t.Fatal("source identity was computed after callback mutation", result, err)
	}
	// Rewriting a future expected result must not make an incorrect provider pass.
	s = minimalSuite()
	second = minimalSuite().Cases[0]
	second.ID, second.Steps[0].ID = "second_case", "second_step"
	s.Cases = append(s.Cases, second)
	calls = 0
	r = boundDefault(t, func(task lifecycle.Task, command lifecycle.Command) lifecycle.Result {
		calls++
		s.Cases[1].Steps[0].Expected.Task.Outcome = "wrong provider output"
		actual := lifecycle.Transition(task, command)
		if calls == 2 {
			actual.Task.Outcome = "wrong provider output"
		}
		return actual
	})
	result, err = r.Run(s)
	if err != nil || result.Matched || calls != 2 || result.StepsExecuted != 2 || result.Entries[1].Expected.Task.Outcome != "Ship outcome" {
		t.Fatal("callback expectation rewrite concealed a mismatch", result, err)
	}
}
