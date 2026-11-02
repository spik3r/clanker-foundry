package scenarioexplain_test

// These tests are frozen before implementation. They NEVER run a reducer.
// Suite expectations are fabricated into caller assertions, not replay evidence.
import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/canonical"
	lifecycle "github.com/spik3r/clanker-foundry/modeling/lifecycle/generated"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/jsonformat"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/model"
	sx "github.com/spik3r/clanker-foundry/modeling/lifecycle/scenarioexplain"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/scenarios"
)

type boundVector struct {
	Name     string          `json:"name"`
	ModelSHA string          `json:"modelSourceSHA256"`
	SuiteSHA string          `json:"suiteSourceSHA256"`
	Expected json.RawMessage `json:"expected"`
}

func boundVectors(t *testing.T) []boundVector {
	t.Helper()
	b, e := os.ReadFile("testdata/bound-vectors.json")
	if e != nil {
		t.Fatal(e)
	}
	if boundHash(b) != "c5e011aba6dd14f5eaf2de0016de10dd1359196decaf0eb8d6edf4463fc147f9" {
		t.Fatal("independent bound projection fixture bytes changed")
	}
	var v struct {
		Vectors []boundVector `json:"vectors"`
	}
	if e = json.Unmarshal(b, &v); e != nil {
		t.Fatal(e)
	}
	if len(v.Vectors) != 3 {
		t.Fatal("vector inventory")
	}
	return v.Vectors
}
func boundHash(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }
func boundCopy[T any](v T) T {
	b, e := json.Marshal(v)
	if e != nil {
		panic(e)
	}
	var out T
	if e = json.Unmarshal(b, &out); e != nil {
		panic(e)
	}
	return out
}
func boundJSON(t *testing.T, v any) []byte {
	t.Helper()
	var b bytes.Buffer
	e := json.NewEncoder(&b)
	e.SetEscapeHTML(false)
	if err := e.Encode(v); err != nil {
		t.Fatal(err)
	}
	return b.Bytes()
}
func boundRunner(t *testing.T, d string, calls *int) scenarios.Runner {
	t.Helper()
	r, e := scenarios.NewRunner(scenarios.Binding{ProfileVersion: "lifecycle/v1", ModelDigest: d}, func(lifecycle.Task, lifecycle.Command) lifecycle.Result {
		*calls++
		panic("explanation invoked callback")
	})
	if e != nil {
		t.Fatal(e)
	}
	return r
}
func boundAssert(t *testing.T, got sx.BoundExplanation, want json.RawMessage) {
	t.Helper()
	var compact bytes.Buffer
	if e := json.Compact(&compact, want); e != nil {
		t.Fatal(e)
	}
	compact.WriteByte('\n')
	if !bytes.Equal(boundJSON(t, got), compact.Bytes()) {
		t.Fatalf("complete independently authored projection differs\ngot %s\nwant %s", boundJSON(t, got), compact.Bytes())
	}
}
func boundReject(t *testing.T, m model.Model, r scenarios.Report, runner scenarios.Runner) {
	t.Helper()
	out, e := sx.ExplainWithRunner(m, r, runner)
	if !errors.Is(e, sx.ErrInvalid) || !reflect.DeepEqual(out, sx.BoundExplanation{}) {
		t.Fatalf("want fixed rejection and zero result; got %#v, %v", out, e)
	}
}

func boundFixture(t *testing.T, n int, calls *int) (model.Model, scenarios.Report, scenarios.Runner, boundVector) {
	t.Helper()
	v := boundVectors(t)[n-1]
	mb, e := os.ReadFile(fmt.Sprintf("../consumerintegration/testdata/model-v%d.json", n))
	if e != nil {
		t.Fatal(e)
	}
	sb, e := os.ReadFile(fmt.Sprintf("../consumerintegration/testdata/suite-v%d.json", n))
	if e != nil {
		t.Fatal(e)
	}
	if boundHash(mb) != v.ModelSHA || boundHash(sb) != v.SuiteSHA {
		t.Fatal("original public fixture bytes changed")
	}
	m, e := jsonformat.Decode(mb)
	if e != nil {
		t.Fatal(e)
	}
	d, e := canonical.Digest(m)
	if e != nil {
		t.Fatal(e)
	}
	runner := boundRunner(t, d, calls)
	s, e := runner.Decode(sb)
	if e != nil {
		t.Fatal(e)
	}
	sd, e := runner.Digest(s)
	if e != nil {
		t.Fatal(e)
	}
	// Deliberate fabrication from original literal expectations, never Run/Execute.
	r := scenarios.Report{ModelDigest: d, ScenarioDigest: sd, SourceDigest: &v.SuiteSHA, Matched: true, CasesCompleted: len(s.Cases), Entries: []scenarios.Entry{}}
	for _, c := range s.Cases {
		current := boundCopy(c.Initial)
		for _, step := range c.Steps {
			want := boundCopy(step.Expected)
			actual := lifecycle.Result{Mode: "example-only", Authority: "none", ModelDigest: d, OK: want.OK, TransitionID: want.TransitionID, Task: boundCopy(want.Task), Refusals: boundCopy(want.Refusals)}
			r.Entries = append(r.Entries, scenarios.Entry{CaseID: c.ID, StepID: step.ID, Input: boundCopy(current), Command: boundCopy(step.Command), Expected: want, Actual: actual, Matched: true, Differences: []string{}})
			if want.OK {
				current = boundCopy(*want.Task)
			}
		}
	}
	r.StepsExecuted = len(r.Entries)
	if _, e = runner.EncodeReport(r); e != nil {
		t.Fatalf("fabricated report must satisfy existing validator: %v", e)
	}
	return m, r, runner, v
}
func boundSingle(r scenarios.Report, index int) scenarios.Report {
	r = boundCopy(r)
	r.Entries = []scenarios.Entry{r.Entries[index]}
	r.CasesCompleted = 1
	r.StepsExecuted = 1
	r.Matched = true
	r.SourceDigest = nil
	return r
}
func boundMatchActual(e *scenarios.Entry) {
	e.Expected = scenarios.Expected{OK: e.Actual.OK, TransitionID: e.Actual.TransitionID, Task: boundCopy(e.Actual.Task), Refusals: boundCopy(e.Actual.Refusals)}
	e.Matched = true
	e.Differences = []string{}
}

func TestBoundFrozenCustomProjections(t *testing.T) {
	for n := 1; n <= 2; n++ {
		t.Run(fmt.Sprint(n), func(t *testing.T) {
			calls := 0
			defer func() {
				if calls != 0 {
					t.Fatal("callback count", calls)
				}
			}()
			m, r, runner, v := boundFixture(t, n, &calls)
			beforeM, beforeR := boundCopy(m), boundCopy(r)
			out, e := sx.ExplainWithRunner(m, r, runner)
			if e != nil {
				t.Fatal(e)
			}
			boundAssert(t, out, v.Expected)
			again, e := sx.ExplainWithRunner(m, r, runner)
			if e != nil || !bytes.Equal(boundJSON(t, out), boundJSON(t, again)) {
				t.Fatal("nondeterminism")
			}
			if !reflect.DeepEqual(m, beforeM) || !reflect.DeepEqual(r, beforeR) {
				t.Fatal("input mutation")
			}
			if _, e = sx.Explain(m, r); !errors.Is(e, sx.ErrInvalid) {
				t.Fatal("default must stay default-only")
			}
			if r.Entries[13].Input.Revision != math.MaxUint64 {
				t.Fatal("missing max revision control")
			}
		})
	}
}
func TestBoundFabricatedV1GuardAndFakeSuccess(t *testing.T) {
	calls := 0
	defer func() {
		if calls != 0 {
			t.Fatal(calls)
		}
	}()
	m, r, runner, _ := boundFixture(t, 1, &calls)
	r = boundSingle(r, 9)
	r.Entries[0].Actual.OK = false
	r.Entries[0].Actual.Task = &r.Entries[0].Input
	r.Entries[0].Actual.Refusals = []string{"G_submitActor"}
	boundMatchActual(&r.Entries[0])
	out, e := sx.ExplainWithRunner(m, r, runner)
	if e != nil {
		t.Fatal(e)
	}
	boundAssert(t, out, boundVectors(t)[2].Expected)
	// A nonblank guest would satisfy v1's guard, but no evaluation is promised.
	r.Entries[0].Actual.OK = true
	r.Entries[0].Actual.Refusals = []string{}
	boundMatchActual(&r.Entries[0])
	out, e = sx.ExplainWithRunner(m, r, runner)
	if e != nil || out.Status != "caller-supplied-assertions" || out.Entries[0].ActualOK != true || len(out.Entries[0].Guards) != 0 {
		t.Fatal("fabricated unchanged-task success must remain assertions", e)
	}
}
func TestBoundRejectIndependentMutations(t *testing.T) {
	mutations := map[string]func(*scenarios.Report){
		"report digest":       func(r *scenarios.Report) { r.ModelDigest = strings.Repeat("0", 64) },
		"source syntax":       func(r *scenarios.Report) { s := "private invalid digest"; r.SourceDigest = &s },
		"count":               func(r *scenarios.Report) { r.StepsExecuted++ },
		"completed":           func(r *scenarios.Report) { r.CasesCompleted = 0 },
		"false comparison":    func(r *scenarios.Report) { r.Entries[0].Differences = []string{"task"} },
		"nil actual refusals": func(r *scenarios.Report) { r.Entries[0].Actual.Refusals = nil; r.Entries[0].Expected.Refusals = nil },
		"wrong mode": func(r *scenarios.Report) {
			r.Entries[0].Actual.Mode = "production"
			r.Matched = false
			r.CasesCompleted = 0
			r.Entries[0].Matched = false
			r.Entries[0].Differences = []string{"mode"}
		},
		"wrong authority": func(r *scenarios.Report) {
			r.Entries[0].Actual.Authority = "approved"
			r.Matched = false
			r.CasesCompleted = 0
			r.Entries[0].Matched = false
			r.Entries[0].Differences = []string{"authority"}
		},
		"wrong actual identity": func(r *scenarios.Report) {
			r.Entries[0].Actual.ModelDigest = strings.Repeat("0", 64)
			r.Matched = false
			r.CasesCompleted = 0
			r.Entries[0].Matched = false
			r.Entries[0].Differences = []string{"modelDigest"}
		},
		"success nil task": func(r *scenarios.Report) { r.Entries[0].Actual.Task = nil; boundMatchActual(&r.Entries[0]) },
		"unknown transition": func(r *scenarios.Report) {
			r.Entries[0].Actual.TransitionID = "T_unknown"
			boundMatchActual(&r.Entries[0])
		},
		"wrong source state": func(r *scenarios.Report) { r.Entries[0].Input.State = "Reviewing" },
		"wrong trigger":      func(r *scenarios.Report) { r.Entries[0].Command.Type = "acceptCandidate" },
	}
	for name, change := range mutations {
		t.Run(name, func(t *testing.T) {
			calls := 0
			m, r, runner, _ := boundFixture(t, 1, &calls)
			r = boundSingle(r, 0)
			change(&r)
			boundReject(t, m, r, runner)
			if calls != 0 {
				t.Fatal(calls)
			}
		})
	}
	for _, reasons := range [][]string{{"G_unknown"}, {"G_reviewModel"}, {"G_revision", "G_acceptance"}, {"G_revision", "G_revision"}, {"G_revision", "INVALID_INPUT"}, {"INVALID_INPUT", "G_revision"}, {}} {
		t.Run(fmt.Sprint(reasons), func(t *testing.T) {
			calls := 0
			m, r, runner, _ := boundFixture(t, 1, &calls)
			r = boundSingle(r, 0)
			r.Entries[0].Actual.OK = false
			r.Entries[0].Actual.Task = &r.Entries[0].Input
			r.Entries[0].Actual.Refusals = reasons
			boundMatchActual(&r.Entries[0])
			boundReject(t, m, r, runner)
			if calls != 0 {
				t.Fatal(calls)
			}
		})
	}
	calls := 0
	m, r, runner, _ := boundFixture(t, 1, &calls)
	boundReject(t, m, r, scenarios.Runner{})
	_, _, other, _ := boundFixture(t, 2, &calls)
	boundReject(t, m, r, other)
	m.Requirements[0].Text += " changed"
	boundReject(t, m, r, runner)
	if calls != 0 {
		t.Fatal(calls)
	}
}
func TestBoundFixedRefusalStructure(t *testing.T) {
	for _, index := range []int{3, 12, 13} {
		t.Run(fmt.Sprint(index), func(t *testing.T) {
			calls := 0
			m, r, runner, _ := boundFixture(t, 1, &calls)
			r = boundSingle(r, index)
			switch index {
			case 3:
				r.Entries[0].Actual.TransitionID = "T_accept"
			case 12:
				r.Entries[0].Actual.Task = &r.Entries[0].Input
			case 13:
				r.Entries[0].Command.Type = "acceptCandidate"
			}
			boundMatchActual(&r.Entries[0])
			boundReject(t, m, r, runner)
			if calls != 0 {
				t.Fatal(calls)
			}
		})
	}
}

func TestBoundMismatchPreservesUnknownExpectedAndNull(t *testing.T) {
	calls := 0
	defer func() {
		if calls != 0 {
			t.Fatal(calls)
		}
	}()
	m, r, runner, _ := boundFixture(t, 1, &calls)
	r = boundSingle(r, 1)
	e := &r.Entries[0]
	e.Expected.TransitionID = "T_private_unknown"
	e.Expected.Refusals = []string{"G_private_unknown", "<&>"}
	e.Matched = false
	e.Differences = []string{"transitionId", "refusals"}
	r.Matched = false
	r.CasesCompleted = 0
	out, err := sx.ExplainWithRunner(m, r, runner)
	if err != nil {
		t.Fatal(err)
	}
	want := sx.Entry{CaseID: "happy", StepID: "happy_self_review", Matched: false, Differences: []string{"transitionId", "refusals"}, ExpectedTransitionID: "T_private_unknown", ExpectedRefusals: []string{"G_private_unknown", "<&>"}, ActualOK: false, ActualTransitionID: "T_accept", ActualRefusals: []string{"G_independent"}, RuntimeReason: "", Guards: []sx.GuardLink{{GuardID: "G_independent", RequirementID: "R_independent", RequirementText: "The reviewer label differs from the owner label.", Expression: sx.Expression{Kind: "op", Op: "ne", Args: []sx.Expression{{Kind: "ref", Ref: "review.reviewer", Args: []sx.Expression{}}, {Kind: "ref", Ref: "task.owner", Args: []sx.Expression{}}}}, TransitionID: "T_accept"}}}
	if !reflect.DeepEqual(out.Entries, []sx.Entry{want}) || out.SourceDigest != nil || out.Matched || out.CasesCompleted != 0 {
		t.Fatal("lost literal mismatch/null", out)
	}
	if !bytes.Contains(boundJSON(t, out), []byte("<&>")) {
		t.Fatal("HTML escaping policy")
	}
	// nil task acceptance differs from an empty list and must not be normalized.
	r = boundSingle(r, 0)
	r.Matched = false
	r.CasesCompleted = 0
	e = &r.Entries[0]
	e.Expected = scenarios.Expected{OK: false, TransitionID: "T_accept", Task: boundCopy(e.Actual.Task), Refusals: []string{"G_independent"}}
	e.Expected.Task.Acceptance = nil
	e.Actual.Task.Acceptance = []string{}
	e.Differences = []string{"task"}
	e.Matched = false
	out, err = sx.ExplainWithRunner(m, r, runner)
	if err != nil || !reflect.DeepEqual(out.Entries[0].Differences, []string{"task"}) {
		t.Fatal("accepted task nil/empty distinction lost", err)
	}
}

func TestBoundNilActualRefusalsAndLegacyControl(t *testing.T) {
	calls := 0
	defer func() {
		if calls != 0 {
			t.Fatal(calls)
		}
	}()
	m, r, runner, _ := boundFixture(t, 1, &calls)
	r = boundSingle(r, 0)
	// Legacy comparison validation permits OK=true with nil actual refusals.
	// Bound provider validation refuses it even when comparisons are internally
	// consistent. Empty nonnil is accepted.
	r.Entries[0].Actual.Refusals = nil
	r.Entries[0].Expected.Refusals = []string{}
	r.Entries[0].Matched = false
	r.Entries[0].Differences = []string{"refusals"}
	r.Matched = false
	r.CasesCompleted = 0
	boundReject(t, m, r, runner)
	b, e := os.ReadFile("../examples/lifecycle.json")
	if e != nil {
		t.Fatal(e)
	}
	base, e := jsonformat.Decode(b)
	if e != nil {
		t.Fatal(e)
	}
	r.ModelDigest = lifecycle.ModelDigest
	r.Entries[0].Actual.ModelDigest = lifecycle.ModelDigest
	r.Entries[0].Input.State = "Draft"
	r.Entries[0].Command.Type = "prepare"
	r.Entries[0].Actual.TransitionID = "T_prepare"
	r.Entries[0].Expected.TransitionID = "T_prepare"
	if _, e = scenarios.EncodeReport(r); e != nil {
		t.Fatal("legacy encoder control", e)
	}
	if _, e = sx.Explain(base, r); e != nil {
		t.Fatal("legacy explanation nil behavior changed", e)
	}
	baseRunner := boundRunner(t, lifecycle.ModelDigest, &calls)
	boundReject(t, base, r, baseRunner)
	r.Entries[0].Actual.Refusals = []string{}
	r.Entries[0].Differences = []string{}
	r.Entries[0].Matched = true
	r.Matched = true
	r.CasesCompleted = 1
	out, e := sx.ExplainWithRunner(base, r, baseRunner)
	if e != nil || out.FormatVersion != "lifecycle-bound-scenario-explanation/v1" || out.IdentitySource != "caller-supplied-go-binding" {
		t.Fatal("explicit baseline-bound empty control", e)
	}
}
func TestBoundReservedUsedGuardCollisions(t *testing.T) {
	for _, reserved := range []string{"INVALID_INPUT", "ILLEGAL_TRANSITION", "REVISION_LIMIT"} {
		t.Run(reserved, func(t *testing.T) {
			calls := 0
			m, r, _, _ := boundFixture(t, 1, &calls)
			for i := range m.Guards {
				if m.Guards[i].ID == "G_revision" {
					m.Guards[i].ID = reserved
				}
			}
			for i := range m.Transitions {
				for j := range m.Transitions[i].Guards {
					if m.Transitions[i].Guards[j] == "G_revision" {
						m.Transitions[i].Guards[j] = reserved
					}
				}
			}
			normalized, e := model.Normalize(m)
			if e != nil {
				t.Fatal("collision model must be genuinely valid", e)
			}
			d, e := canonical.Digest(normalized)
			if e != nil {
				t.Fatal(e)
			}
			r = boundSingle(r, 0)
			r.ModelDigest = d
			r.Entries[0].Actual.ModelDigest = d
			runner := boundRunner(t, d, &calls)
			if _, e = runner.EncodeReport(r); e != nil {
				t.Fatal("collision report must pass native shape", e)
			}
			boundReject(t, m, r, runner)
			if calls != 0 {
				t.Fatal(calls)
			}
		})
	}
}
func TestBoundOwnedNestedOutput(t *testing.T) {
	calls := 0
	defer func() {
		if calls != 0 {
			t.Fatal(calls)
		}
	}()
	m, r, runner, v := boundFixture(t, 1, &calls)
	out, e := sx.ExplainWithRunner(m, r, runner)
	if e != nil {
		t.Fatal(e)
	}
	saved := boundCopy(out)
	*r.SourceDigest = strings.Repeat("f", 64)
	r.Entries[1].Actual.Refusals[0] = "G_changed"
	r.Entries[1].Expected.Refusals[0] = "G_changed"
	r.Entries[1].Differences = append(r.Entries[1].Differences, "task")
	for i := range m.Requirements {
		m.Requirements[i].Text = "changed"
	}
	for i := range m.Guards {
		if len(m.Guards[i].Expr.Args) > 0 {
			m.Guards[i].Expr.Args[0].Ref = "task.id"
		}
	}
	if !reflect.DeepEqual(saved, out) {
		t.Fatal("input aliases escaped to output")
	}
	boundAssert(t, out, v.Expected)
	m, r, runner, _ = boundFixture(t, 1, &calls)
	out, e = sx.ExplainWithRunner(m, r, runner)
	if e != nil {
		t.Fatal(e)
	}
	beforeM, beforeR := boundCopy(m), boundCopy(r)
	*out.SourceDigest = strings.Repeat("a", 64)
	out.Entries[1].ActualRefusals[0] = "changed"
	out.Entries[1].ExpectedRefusals[0] = "changed"
	out.Entries[1].Guards[0].Expression.Args[0].Ref = "changed"
	out.Entries[1].Guards[0].RequirementText = "changed"
	out.Entries[1].Differences = append(out.Entries[1].Differences, "task")
	if !reflect.DeepEqual(m, beforeM) || !reflect.DeepEqual(r, beforeR) {
		t.Fatal("output aliases caller storage")
	}
	again, e := sx.ExplainWithRunner(m, r, runner)
	if e != nil {
		t.Fatal(e)
	}
	boundAssert(t, again, v.Expected)
	// Nonempty mismatch differences require their own underlying-storage probe.
	r = boundSingle(r, 1)
	r.Entries[0].Expected.Refusals = []string{"unknown"}
	r.Entries[0].Differences = []string{"refusals"}
	r.Entries[0].Matched = false
	r.CasesCompleted = 0
	r.Matched = false
	out, e = sx.ExplainWithRunner(m, r, runner)
	if e != nil {
		t.Fatal(e)
	}
	r.Entries[0].Differences[0] = "task"
	if out.Entries[0].Differences[0] != "refusals" {
		t.Fatal("difference alias")
	}
	out.Entries[0].Differences[0] = "ok"
	if r.Entries[0].Differences[0] != "task" {
		t.Fatal("reverse difference alias")
	}
}
func TestBoundPrivateTextAndEscapedOutputBudget(t *testing.T) {
	calls := 0
	defer func() {
		if calls != 0 {
			t.Fatal(calls)
		}
	}()
	m, r, _, _ := boundFixture(t, 1, &calls)
	for i := range m.Requirements {
		m.Requirements[i].Text = "private <&> \"quoted\"\nline"
	}
	d, e := canonical.Digest(m)
	if e != nil {
		t.Fatal(e)
	}
	r = boundSingle(r, 1)
	r.ModelDigest = d
	r.Entries[0].Actual.ModelDigest = d
	runner := boundRunner(t, d, &calls)
	out, e := sx.ExplainWithRunner(m, r, runner)
	if e != nil {
		t.Fatal(e)
	}
	if out.Entries[0].Guards[0].RequirementText != "private <&> \"quoted\"\nline" || !bytes.Contains(boundJSON(t, out), []byte("private <&>")) {
		t.Fatal("text normalization/HTML escaping")
	}
	// A small structurally valid report can expand beyond 8 MiB by repeating
	// declared requirement text. NUL is valid UTF-8 and expands to six JSON bytes.
	for i := range m.Requirements {
		m.Requirements[i].Text = "x" + strings.Repeat("\x00", 4095)
	}
	d, e = canonical.Digest(m)
	if e != nil {
		t.Fatal(e)
	}
	runner = boundRunner(t, d, &calls)
	r.ModelDigest = d
	r.Entries[0].Actual.ModelDigest = d
	guards := []string{"G_revision", "G_outcome", "G_acceptance", "G_owner", "G_reviewer", "G_reviewTask", "G_reviewRevision", "G_reviewModel", "G_reviewContent", "G_reviewActor", "G_independent", "G_approved"}
	sort.Strings(guards)
	seed := boundCopy(r.Entries[0])
	seed.Actual.OK = false
	seed.Actual.Task = &seed.Input
	seed.Actual.Refusals = guards
	boundMatchActual(&seed)
	r.Entries = []scenarios.Entry{}
	for c := 0; c < 4; c++ {
		for j := 0; j < 32; j++ {
			entry := boundCopy(seed)
			entry.CaseID = fmt.Sprintf("budget_case_%d", c)
			entry.StepID = fmt.Sprintf("budget_step_%d_%d", c, j)
			r.Entries = append(r.Entries, entry)
		}
	}
	r.Matched = true
	r.CasesCompleted = 4
	r.StepsExecuted = 128
	if _, e = runner.EncodeReport(r); e != nil {
		t.Fatal("budget control must pass preexisting report validation", e)
	}
	boundReject(t, m, r, runner)
}

// This independent size construction uses only literal declaration projections
// and standard encoding/json, never output from either explanation mapper.
func TestBoundEnvelopeOverheadIsIncludedInFinalLimit(t *testing.T) {
	calls := 0
	defer func() {
		if calls != 0 {
			t.Fatal(calls)
		}
	}()
	literal := "x" + strings.Repeat("\x00", 4095)
	expr := model.Expr{Kind: "op", Op: "eq", Args: []model.Expr{{Kind: "string", Value: literal}, {Kind: "string", Value: literal}}}
	projection := sx.Expression{Kind: "op", Op: "eq", Args: []sx.Expression{{Kind: "string", Value: literal, Args: []sx.Expression{}}, {Kind: "string", Value: literal, Args: []sx.Expression{}}}}
	m := model.Model{Version: 1, ID: "BoundaryModel", Initial: "A", States: []model.State{{ID: "A", Label: "A"}, {ID: "B", Label: "B", Terminal: true}}, Requirements: []model.Requirement{{ID: "R", Text: "x"}}, Guards: []model.Guard{{ID: "G", Requirement: "R", Expr: expr}}, Transitions: []model.Transition{{ID: "T", From: "A", Trigger: "go", To: "B", Guards: []string{"G"}, Effects: []model.Effect{{Op: model.SetState}, {Op: model.IncrementRevision}}}}}
	input := lifecycle.Task{ID: "task", State: "A", Outcome: "x", Acceptance: []string{}}
	report := scenarios.Report{ModelDigest: strings.Repeat("a", 64), ScenarioDigest: strings.Repeat("b", 64), Matched: false, CasesCompleted: 3, StepsExecuted: 128, Entries: []scenarios.Entry{}}
	expected := sx.BoundExplanation{FormatVersion: "lifecycle-bound-scenario-explanation/v1", IdentitySource: "caller-supplied-go-binding", Mode: "example-only", Authority: "none", Status: "caller-supplied-assertions", ModelDigest: report.ModelDigest, ScenarioDigest: report.ScenarioDigest, Matched: false, CasesCompleted: 3, StepsExecuted: 128, Entries: []sx.Entry{}}
	for i := 0; i < 128; i++ {
		caseID, stepID := fmt.Sprintf("c%d", i/32), fmt.Sprintf("s%d", i)
		actual := lifecycle.Result{Mode: "example-only", Authority: "none", ModelDigest: report.ModelDigest, OK: false, TransitionID: "T", Task: &input, Refusals: []string{"G"}}
		e := scenarios.Entry{CaseID: caseID, StepID: stepID, Input: input, Command: lifecycle.Command{Type: "go", Actor: "actor"}, Actual: actual}
		boundMatchActual(&e)
		report.Entries = append(report.Entries, boundCopy(e))
		expected.Entries = append(expected.Entries, sx.Entry{CaseID: caseID, StepID: stepID, Matched: true, Differences: []string{}, ExpectedTransitionID: "T", ExpectedRefusals: []string{"G"}, ActualOK: false, ActualTransitionID: "T", ActualRefusals: []string{"G"}, Guards: []sx.GuardLink{{GuardID: "G", RequirementID: "R", RequirementText: "x", Expression: boundCopy(projection), TransitionID: "T"}}})
	}
	report.Entries[127].Expected.Refusals = []string{"u", "v", "w", "z"}
	report.Entries[127].Matched = false
	report.Entries[127].Differences = []string{"refusals"}
	expected.Entries[127].ExpectedRefusals = []string{"u", "v", "w", "z"}
	expected.Entries[127].Matched = false
	expected.Entries[127].Differences = []string{"refusals"}
	legacySize := func() int {
		return len(boundJSON(t, sx.Explanation{FormatVersion: "lifecycle-scenario-explanation/v1", Status: "caller-supplied-assertions", ModelDigest: expected.ModelDigest, ScenarioDigest: expected.ScenarioDigest, Matched: false, CasesCompleted: 3, StepsExecuted: 128, Entries: expected.Entries}))
	}
	setText := func(n int) {
		text := "x" + strings.Repeat("\x00", n)
		m.Requirements[0].Text = text
		for i := range expected.Entries {
			expected.Entries[i].Guards[0].RequirementText = text
		}
	}
	// One NUL adds 6*128 bytes. Four independent expected strings permit
	// 1020 bytes of one-byte padding, enough to reach the exact legacy boundary.
	lo, hi := 0, 4095
	for lo < hi {
		mid := (lo + hi + 1) / 2
		setText(mid)
		if legacySize() <= sx.MaxOutputBytes {
			lo = mid
		} else {
			hi = mid - 1
		}
	}
	setText(lo)
	padding := sx.MaxOutputBytes - legacySize()
	if padding < 0 || padding > 1020 {
		t.Fatalf("independent boundary cannot be reached: padding %d", padding)
	}
	for i := 0; i < 4; i++ {
		n := padding
		if n > 255 {
			n = 255
		}
		text := expected.Entries[127].ExpectedRefusals[i] + strings.Repeat("p", n)
		expected.Entries[127].ExpectedRefusals[i] = text
		report.Entries[127].Expected.Refusals[i] = text
		padding -= n
	}
	if padding != 0 || legacySize() != sx.MaxOutputBytes || len(boundJSON(t, expected)) <= sx.MaxOutputBytes {
		t.Fatal("control does not isolate bound-envelope overhead")
	}
	d, e := canonical.Digest(m)
	if e != nil {
		t.Fatal("boundary model must validate", e)
	}
	report.ModelDigest = d
	expected.ModelDigest = d
	for i := range report.Entries {
		report.Entries[i].Actual.ModelDigest = d
	}
	runner := boundRunner(t, d, &calls)
	if _, e = runner.EncodeReport(report); e != nil {
		t.Fatal("boundary report must pass native validation", e)
	}
	if legacySize() != sx.MaxOutputBytes || len(boundJSON(t, expected)) <= sx.MaxOutputBytes {
		t.Fatal("identity width changed boundary")
	}
	boundReject(t, m, report, runner)
}
