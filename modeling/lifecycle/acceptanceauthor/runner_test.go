package acceptanceauthor_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/acceptanceauthor"
	lifecycle "github.com/spik3r/clanker-foundry/modeling/lifecycle/generated"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/scenarios"
)

func boundRead(t *testing.T, path string) []byte {
	t.Helper()
	b, e := os.ReadFile(path)
	if e != nil {
		t.Fatal(e)
	}
	return b
}
func boundFixture(t *testing.T) ([]byte, []byte, []byte) {
	return boundRead(t, "testdata/manifest.json"), boundRead(t, "../examples/lifecycle.json"), boundRead(t, "testdata/suite.json")
}
func boundHash(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }
func boundJSON(t *testing.T, v any) []byte {
	t.Helper()
	b, e := json.Marshal(v)
	if e != nil {
		t.Fatal(e)
	}
	return b
}
func boundObject(t *testing.T, b []byte) map[string]any {
	t.Helper()
	var v map[string]any
	if e := json.Unmarshal(b, &v); e != nil {
		t.Fatal(e)
	}
	return v
}
func boundRunner(t *testing.T, digest string, apply func(lifecycle.Task, lifecycle.Command) lifecycle.Result) scenarios.Runner {
	t.Helper()
	r, e := scenarios.NewRunner(scenarios.Binding{ProfileVersion: lifecycle.ProfileVersion, ModelDigest: digest}, apply)
	if e != nil {
		t.Fatal(e)
	}
	return r
}
func boundSuiteDeclaration(t *testing.T, m, s []byte) []byte {
	t.Helper()
	v := boundObject(t, m)
	v["suiteSourceSha256"] = boundHash(s)
	return boundJSON(t, v)
}

func TestBoundAcceptanceIdentityAndReplay(t *testing.T) {
	m, model, s := boundFixture(t)
	mc, modelc, sc := bytes.Clone(m), bytes.Clone(model), bytes.Clone(s)
	for _, mode := range []string{"check", "replay"} {
		t.Run(mode, func(t *testing.T) {
			calls := 0
			r := boundRunner(t, lifecycle.ModelDigest, func(task lifecycle.Task, c lifecycle.Command) lifecycle.Result {
				calls++
				return lifecycle.Transition(task, c)
			})
			got, e := acceptanceauthor.CheckWithRunner(m, model, s, mode, r)
			if e != nil {
				t.Fatal(e)
			}
			if got.FormatVersion != acceptanceauthor.BoundReportVersion || got.IdentitySource != "caller-supplied-go-binding" || got.RunnerModelDigest != lifecycle.ModelDigest || got.Authority != "none" {
				t.Fatalf("identity: %+v", got)
			}
			if got.ModelBinding != "current" || got.SuiteBinding != "current" || got.SuiteModelBinding != "current" || got.RunnerBinding != "current" || got.SuiteValidation != "valid" || got.ScenarioDigest == nil {
				t.Fatalf("bindings: %+v", got)
			}
			raw := boundJSON(t, got)
			if bytes.Contains(raw, []byte(`"compiledModelDigest"`)) {
				t.Fatal("custom identity presented as compiled evidence")
			}
			if mode == "check" {
				if calls != 0 || got.Replay != "not-run" || got.ScenarioReport != nil {
					t.Fatal("check mode executed")
				}
			} else {
				if calls != 2 || got.Replay != "matched" {
					t.Fatalf("replay calls=%d report=%+v", calls, got)
				}
				nested := boundObject(t, got.ScenarioReport)
				if nested["formatVersion"] != "lifecycle-bound-scenario-report/v1" || nested["identitySource"] != "caller-supplied-go-binding" {
					t.Fatalf("nested identity: %s", got.ScenarioReport)
				}
			}
			again, e := acceptanceauthor.CheckWithRunner(m, model, s, mode, r)
			if e != nil || !bytes.Equal(raw, boundJSON(t, again)) {
				t.Fatal("nondeterministic bound report")
			}
		})
	}
	if !bytes.Equal(m, mc) || !bytes.Equal(model, modelc) || !bytes.Equal(s, sc) {
		t.Fatal("input bytes changed")
	}
}

func TestBoundAcceptanceStaleAndInvalidNeverInvoke(t *testing.T) {
	m, model, s := boundFixture(t)
	badCase := boundObject(t, m)
	badCase["cases"].([]any)[0].(map[string]any)["id"] = "different_case"
	badSuite := boundObject(t, s)
	cases := badSuite["cases"].([]any)
	steps := cases[len(cases)-1].(map[string]any)["steps"].([]any)
	steps[len(steps)-1].(map[string]any)["command"].(map[string]any)["expectedRevision"] = "01"
	invalidLast := boundJSON(t, badSuite)
	badHeader := boundObject(t, s)
	badHeader["modelDigest"] = strings.Repeat("a", 64)
	staleHeader := boundJSON(t, badHeader)
	changedModel := bytes.Replace(model, []byte(`"label": "Draft"`), []byte(`"label": "Changed draft"`), 1)
	if bytes.Equal(changedModel, model) {
		t.Fatal("model fixture marker not found")
	}
	for _, tc := range []struct {
		name         string
		m, model, s  []byte
		digest, mode string
		invalid      bool
	}{
		{"suite bytes", m, model, append(bytes.Clone(s), '\n'), lifecycle.ModelDigest, "replay", false},
		{"model meaning", m, changedModel, s, lifecycle.ModelDigest, "replay", false},
		{"runner identity", m, model, s, strings.Repeat("a", 64), "replay", false},
		{"suite header", boundSuiteDeclaration(t, m, staleHeader), model, staleHeader, lifecycle.ModelDigest, "replay", false},
		{"case IDs", boundJSON(t, badCase), model, s, lifecycle.ModelDigest, "replay", true},
		{"invalid final step", boundSuiteDeclaration(t, m, invalidLast), model, invalidLast, lifecycle.ModelDigest, "replay", true},
		{"invalid final check-only step", boundSuiteDeclaration(t, m, invalidLast), model, invalidLast, lifecycle.ModelDigest, "check", true},
		{"wrong check-only case IDs", boundJSON(t, badCase), model, s, lifecycle.ModelDigest, "check", true},
		{"malformed manifest", []byte(`{`), model, s, lifecycle.ModelDigest, "replay", true},
		{"malformed model", m, []byte(`{`), s, lifecycle.ModelDigest, "replay", true},
		{"malformed suite", m, model, []byte(`{`), lifecycle.ModelDigest, "replay", true},
		{"invalid mode", m, model, s, lifecycle.ModelDigest, "execute", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			calls := 0
			r := boundRunner(t, tc.digest, func(task lifecycle.Task, c lifecycle.Command) lifecycle.Result {
				calls++
				return lifecycle.Transition(task, c)
			})
			got, e := acceptanceauthor.CheckWithRunner(tc.m, tc.model, tc.s, tc.mode, r)
			if calls != 0 {
				t.Fatal("preflight invoked provider")
			}
			if tc.invalid {
				if !errors.Is(e, acceptanceauthor.ErrInvalid) || !reflect.DeepEqual(got, acceptanceauthor.BoundReport{}) {
					t.Fatalf("invalid: %+v %v", got, e)
				}
			} else {
				if e != nil || got.SuiteValidation != "not-checked" || got.Replay != "not-run" || got.ScenarioDigest != nil || got.ScenarioReport != nil {
					t.Fatalf("stale: %+v %v", got, e)
				}
			}
		})
	}
	for _, mode := range []string{"check", "replay"} {
		got, e := acceptanceauthor.CheckWithRunner(m, model, s, mode, scenarios.Runner{})
		if !errors.Is(e, scenarios.ErrInvalidRunner) || !reflect.DeepEqual(got, acceptanceauthor.BoundReport{}) {
			t.Fatalf("zero runner defaulted: %+v %v", got, e)
		}
	}
}

func TestBoundAcceptanceProviderErrorsAreNotMismatches(t *testing.T) {
	m, model, s := boundFixture(t)
	for _, field := range []string{"mode", "authority", "modelDigest", "nilTask", "nilRefusals", "oversizedTask"} {
		t.Run(field, func(t *testing.T) {
			calls := 0
			r := boundRunner(t, lifecycle.ModelDigest, func(task lifecycle.Task, c lifecycle.Command) lifecycle.Result {
				calls++
				actual := lifecycle.Transition(task, c)
				switch field {
				case "mode":
					actual.Mode = "production"
				case "authority":
					actual.Authority = "claimed"
				case "modelDigest":
					actual.ModelDigest = strings.Repeat("b", 64)
				case "nilTask":
					actual.Task = nil
				case "nilRefusals":
					actual.Refusals = nil
				case "oversizedTask":
					actual.Task.Outcome = strings.Repeat("x", scenarios.MaxTextBytes+1)
				}
				return actual
			})
			got, e := acceptanceauthor.CheckWithRunner(m, model, s, "replay", r)
			if !errors.Is(e, scenarios.ErrProviderProtocol) || calls != 1 || !reflect.DeepEqual(got, acceptanceauthor.BoundReport{}) {
				t.Fatalf("provider failure lost: calls=%d %+v %v", calls, got, e)
			}
		})
	}
	calls := 0
	r := boundRunner(t, lifecycle.ModelDigest, func(task lifecycle.Task, c lifecycle.Command) lifecycle.Result {
		calls++
		actual := lifecycle.Transition(task, c)
		actual.TransitionID = "DifferentTransition"
		return actual
	})
	got, e := acceptanceauthor.CheckWithRunner(m, model, s, "replay", r)
	if e != nil || calls != 1 || got.Replay != "mismatched" {
		t.Fatalf("literal mismatch: calls=%d %+v %v", calls, got, e)
	}
	var nested struct {
		Steps int `json:"stepsExecuted"`
		Cases int `json:"casesCompleted"`
	}
	if e := json.Unmarshal(got.ScenarioReport, &nested); e != nil || nested.Steps != 1 || nested.Cases != 0 {
		t.Fatalf("mismatch did not stop: %s %v", got.ScenarioReport, e)
	}
}

func TestBoundAcceptancePanicPropagates(t *testing.T) {
	m, model, s := boundFixture(t)
	marker := &struct{}{}
	r := boundRunner(t, lifecycle.ModelDigest, func(lifecycle.Task, lifecycle.Command) lifecycle.Result { panic(marker) })
	defer func() {
		if v := recover(); v != marker {
			t.Fatalf("panic replaced or swallowed: %v", v)
		}
	}()
	_, _ = acceptanceauthor.CheckWithRunner(m, model, s, "replay", r)
}

func TestBoundAcceptanceCustomFrozenPreflight(t *testing.T) {
	expected := boundObject(t, boundRead(t, "../consumerintegration/testdata/expected.json"))["versions"].(map[string]any)
	for _, v := range []string{"v1", "v2"} {
		m := boundRead(t, "../consumerintegration/testdata/manifest-"+v+".json")
		model := boundRead(t, "../consumerintegration/testdata/model-"+v+".json")
		s := boundRead(t, "../consumerintegration/testdata/suite-"+v+".json")
		declared := boundObject(t, m)["modelDigest"].(string)
		calls := 0
		r := boundRunner(t, declared, func(lifecycle.Task, lifecycle.Command) lifecycle.Result { calls++; panic("preflight called provider") })
		got, e := acceptanceauthor.CheckWithRunner(m, model, s, "check", r)
		if e != nil || calls != 0 || got.SuiteValidation != "valid" || got.Replay != "not-run" || got.ScenarioDigest == nil {
			t.Fatalf("custom current check: %+v %v", got, e)
		}
		if *got.ScenarioDigest != expected[v].(map[string]any)["scenarioDigest"].(string) {
			t.Fatal("custom scenario identity differs from frozen literal")
		}
		other := boundRunner(t, lifecycle.ModelDigest, func(lifecycle.Task, lifecycle.Command) lifecycle.Result { calls++; panic("stale called provider") })
		stale, e := acceptanceauthor.CheckWithRunner(m, model, s, "replay", other)
		if e != nil || calls != 0 || stale.RunnerBinding != "stale" || stale.SuiteValidation != "not-checked" || stale.Replay != "not-run" {
			t.Fatalf("custom stale: %+v %v", stale, e)
		}
	}
}

func TestLegacyAcceptanceByteParity(t *testing.T) {
	var want map[string]string
	if e := json.Unmarshal(boundRead(t, "../scenarios/testdata/legacy-parity.json"), &want); e != nil {
		t.Fatal(e)
	}
	m, model, s := boundFixture(t)
	for _, mode := range []string{"check", "replay"} {
		r, e := acceptanceauthor.Check(m, model, s, mode)
		if e != nil {
			t.Fatal(e)
		}
		if got := boundHash(boundJSON(t, r)); got != want["acceptanceCheck-"+mode] {
			t.Fatalf("legacy Check %s changed: %s", mode, got)
		}
		request := acceptanceauthor.Request{FormatVersion: acceptanceauthor.RequestVersion, Mode: mode, Manifest: string(m), Model: string(model), Suite: string(s)}
		raw, e := acceptanceauthor.Execute(boundJSON(t, request))
		if e != nil {
			t.Fatal(e)
		}
		if got := boundHash(raw); got != want["acceptanceExecute-"+mode] {
			t.Fatalf("legacy Execute %s changed: %s", mode, got)
		}
	}
}
