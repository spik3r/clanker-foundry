// Independent pre-adoption integration oracle. No Explain, Execute, or new
// DecodeReport output was used to author expectations. Only existing scenario
// digest computation supplies the canonical suite identity, not expected entries.
package scenarioexplain_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/canonical"
	transport "github.com/spik3r/clanker-foundry/modeling/lifecycle/generated"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/jsonformat"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/model"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/scenarioexplain"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/scenarios"
	"os"
	"reflect"
	"strings"
	"testing"
)

const integrationModelDigest = "948630cb7857199443b59b92d9b1ab0fadf2bf7a5bf8125f7a7e767db83958ed"
const integrationModelSource = "a60916ca68771b28ee48712288decad7eb741d6a18e1cbc95b2a196eb10d12d8"
const integrationSuiteSource = "ef199aa93c9e70887f3da8e72aa4a79dcaf130c4cec72d27b240d6780a3e4b26"

func integrationSHA(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }
func integrationFile(t *testing.T, p, hash string) []byte {
	t.Helper()
	b, e := os.ReadFile(p)
	if e != nil {
		t.Fatal(e)
	}
	if integrationSHA(b) != hash {
		t.Fatal("independent fixture drift", p)
	}
	return b
}
func integrationModel(t *testing.T) ([]byte, model.Model) {
	t.Helper()
	b := integrationFile(t, "../examples/lifecycle.json", integrationModelSource)
	m, e := jsonformat.Decode(b)
	if e != nil {
		t.Fatal(e)
	}
	d, e := canonical.Digest(m)
	if e != nil || d != integrationModelDigest {
		t.Fatal("baseline model identity", d, e)
	}
	return b, m
}
func integrationRef(ref string) scenarioexplain.Expression {
	return scenarioexplain.Expression{Kind: "ref", Ref: ref, Args: []scenarioexplain.Expression{}}
}

// Declared texts and expression structures are spelled from the frozen source,
// not obtained from the explanation mapper or an actual refusal report.
func integrationGuard(id, transition string) scenarioexplain.GuardLink {
	g := scenarioexplain.GuardLink{GuardID: id, TransitionID: transition}
	switch id {
	case "G_actor":
		g.RequirementID = "R_actor"
		g.RequirementText = "The named starting actor is nonblank."
		g.Expression = scenarioexplain.Expression{Kind: "op", Op: "nonempty", Args: []scenarioexplain.Expression{integrationRef("command.actor")}}
	case "G_reviewContent":
		g.RequirementID = "R_reviewContent"
		g.RequirementText = "Review refers to current task content."
		g.Expression = scenarioexplain.Expression{Kind: "op", Op: "eq", Args: []scenarioexplain.Expression{integrationRef("review.taskDigest"), integrationRef("task.digest")}}
	case "G_reviewRevision":
		g.RequirementID = "R_reviewRevision"
		g.RequirementText = "Review refers to the current revision."
		g.Expression = scenarioexplain.Expression{Kind: "op", Op: "eq", Args: []scenarioexplain.Expression{integrationRef("review.taskRevision"), integrationRef("task.revision")}}
	case "G_revision":
		g.RequirementID = "R_revision"
		g.RequirementText = "Command matches the current revision."
		g.Expression = scenarioexplain.Expression{Kind: "op", Op: "eq", Args: []scenarioexplain.Expression{integrationRef("task.revision"), integrationRef("command.expectedRevision")}}
	default:
		panic("oracle guard not explicitly specified")
	}
	return g
}
func integrationExpected(scenarioDigest string) scenarioexplain.Explanation {
	source := integrationSuiteSource
	out := scenarioexplain.Explanation{FormatVersion: "lifecycle-scenario-explanation/v1", Status: "caller-supplied-assertions", ModelDigest: integrationModelDigest, ScenarioDigest: scenarioDigest, SourceDigest: &source, Matched: true, CasesCompleted: 4, StepsExecuted: 13, Entries: []scenarioexplain.Entry{}}
	// Literal ordered inventory of the preexisting independent suite.
	rows := []struct {
		caseID, stepID, transition string
		ok                         bool
		reasons                    []string
		runtime                    string
	}{
		{"happy", "prepare", "T_prepare", true, []string{}, ""},
		{"happy", "start", "T_start", true, []string{}, ""},
		{"happy", "submit", "T_requestReview", true, []string{}, ""},
		{"happy", "approve", "T_approve", true, []string{}, ""},
		{"refusal_then_success", "blank_actor", "T_start", false, []string{"G_actor"}, ""},
		{"refusal_then_success", "unknown_command", "", false, []string{"INVALID_INPUT"}, "INVALID_INPUT"},
		{"refusal_then_success", "start_after_both_refusals", "T_start", true, []string{}, ""},
		{"rework_and_replay", "changes", "T_requestChanges", true, []string{}, ""},
		{"rework_and_replay", "resubmit", "T_requestReview", true, []string{}, ""},
		{"rework_and_replay", "stale_review_fresh_command", "T_approve", false, []string{"G_reviewContent", "G_reviewRevision"}, ""},
		{"rework_and_replay", "replayed_entire_command", "T_approve", false, []string{"G_reviewContent", "G_reviewRevision", "G_revision"}, ""},
		{"rework_and_replay", "fresh_approval", "T_approve", true, []string{}, ""},
		{"same_revision_content_mismatch", "old_content_review", "T_approve", false, []string{"G_reviewContent"}, ""},
	}
	for _, r := range rows {
		e := scenarioexplain.Entry{CaseID: r.caseID, StepID: r.stepID, Matched: true, Differences: []string{}, ExpectedTransitionID: r.transition, ExpectedRefusals: append([]string{}, r.reasons...), ActualOK: r.ok, ActualTransitionID: r.transition, ActualRefusals: append([]string{}, r.reasons...), RuntimeReason: r.runtime, Guards: []scenarioexplain.GuardLink{}}
		if r.runtime == "" {
			for _, id := range r.reasons {
				e.Guards = append(e.Guards, integrationGuard(id, r.transition))
			}
		}
		out.Entries = append(out.Entries, e)
	}
	return out
}
func integrationRequest(m, r []byte) []byte {
	return []byte(`{"formatVersion":"lifecycle-scenario-explain-request/v1","model":` + string(m) + `,"report":` + string(r) + `}`)
}
func integrationExplainBoth(t *testing.T, mRaw []byte, m model.Model, r scenarios.Report, want scenarioexplain.Explanation) {
	t.Helper()
	before, e := scenarios.EncodeReport(r)
	if e != nil {
		t.Fatal(e)
	}
	modelBefore, e := jsonformat.Encode(m)
	if e != nil {
		t.Fatal(e)
	}
	got, e := scenarioexplain.Explain(m, r)
	if e != nil || !reflect.DeepEqual(got, want) {
		t.Fatalf("typed explanation differs from literal oracle: %v\ngot %+v\nwant %+v", e, got, want)
	}
	decoded, e := scenarios.DecodeReport(before)
	if e != nil || !reflect.DeepEqual(decoded, r) {
		t.Fatal("report roundtrip lost a field or nil/empty distinction", e)
	}
	reencoded, e := scenarios.EncodeReport(decoded)
	if e != nil || !bytes.Equal(before, reencoded) {
		t.Fatal("baseline report roundtrip not byte stable", e)
	}
	got, e = scenarioexplain.Explain(m, decoded)
	if e != nil || !reflect.DeepEqual(got, want) {
		t.Fatal("decoded explanation differs", e)
	}
	out, e := scenarioexplain.Execute(integrationRequest(mRaw, before))
	if e != nil {
		t.Fatal(e)
	}
	var wire scenarioexplain.Explanation
	if e = json.Unmarshal(out, &wire); e != nil || !reflect.DeepEqual(wire, want) {
		t.Fatal("strict explanation differs from literal oracle", e)
	}
	again, e := scenarioexplain.Execute(integrationRequest(mRaw, before))
	if e != nil || !bytes.Equal(out, again) || !bytes.HasSuffix(out, []byte{'\n'}) {
		t.Fatal("wire nondeterminism or missing final LF", e)
	}
	after, e := scenarios.EncodeReport(r)
	if e != nil || !bytes.Equal(before, after) {
		t.Fatal("report input mutated", e)
	}
	modelAfter, e := jsonformat.Encode(m)
	if e != nil || !bytes.Equal(modelBefore, modelAfter) {
		t.Fatal("model input mutated", e)
	}
}
func TestIntegrationFrozenScenarioReportToExplanation(t *testing.T) {
	mRaw, m := integrationModel(t)
	suite := integrationFile(t, "../scenarios/testdata/lifecycle-v1.json", integrationSuiteSource)
	parsed, e := scenarios.Decode(suite)
	if e != nil {
		t.Fatal(e)
	}
	digest, e := scenarios.Digest(parsed)
	if e != nil {
		t.Fatal(e)
	}
	// Only the old scenario engine executes the old frozen suite. Explanation
	// expectations above are not derived from this report or execution.
	report, e := scenarios.Execute(suite)
	if e != nil {
		t.Fatal(e)
	}
	integrationExplainBoth(t, mRaw, m, report, integrationExpected(digest))
	if integrationSHA(suite) != integrationSuiteSource || integrationSHA(mRaw) != integrationModelSource {
		t.Fatal("raw input mutated")
	}
}

// This invented report is structurally consistent but was never executed. It
// even asserts successful prepare without changing Draft/revision zero: input
// validation must not be misrepresented as execution or effect verification.
func integrationInventedReport(nullable bool) scenarios.Report {
	task := transport.Task{ID: "invented_task", State: "Draft", Revision: 0, Outcome: "invented outcome", Acceptance: []string{"invented criterion"}, Owner: nil}
	expectedTask := task
	actualTask := task
	actual := []string{}
	differences := []string{}
	matched := true
	completed := 1
	if nullable {
		actual = nil
		differences = []string{"refusals"}
		matched = false
		completed = 0
	}
	return scenarios.Report{ModelDigest: integrationModelDigest, ScenarioDigest: strings.Repeat("b", 64), SourceDigest: nil, Matched: matched, CasesCompleted: completed, StepsExecuted: 1, Entries: []scenarios.Entry{{CaseID: "invented_case", StepID: "invented_prepare", Input: task, Command: transport.Command{Type: "prepare", Actor: "invented_actor", ExpectedRevision: 0, Review: nil}, Expected: scenarios.Expected{OK: true, TransitionID: "T_prepare", Task: &expectedTask, Refusals: []string{}}, Actual: transport.Result{Mode: "example-only", Authority: "none", ModelDigest: integrationModelDigest, OK: true, TransitionID: "T_prepare", Task: &actualTask, Refusals: actual}, Matched: matched, Differences: differences}}}
}
func integrationInventedExpected(nullable bool) scenarioexplain.Explanation {
	actual := []string{}
	diff := []string{}
	matched := true
	completed := 1
	if nullable {
		actual = nil
		diff = []string{"refusals"}
		matched = false
		completed = 0
	}
	return scenarioexplain.Explanation{FormatVersion: "lifecycle-scenario-explanation/v1", Status: "caller-supplied-assertions", ModelDigest: integrationModelDigest, ScenarioDigest: strings.Repeat("b", 64), SourceDigest: nil, Matched: matched, CasesCompleted: completed, StepsExecuted: 1, Entries: []scenarioexplain.Entry{{CaseID: "invented_case", StepID: "invented_prepare", Matched: matched, Differences: diff, ExpectedTransitionID: "T_prepare", ExpectedRefusals: []string{}, ActualOK: true, ActualTransitionID: "T_prepare", ActualRefusals: actual, RuntimeReason: "", Guards: []scenarioexplain.GuardLink{}}}}
}
func TestIntegrationInventedAssertionsAndNullableActualRefusals(t *testing.T) {
	mRaw, m := integrationModel(t)
	for _, nullable := range []bool{false, true} {
		integrationExplainBoth(t, mRaw, m, integrationInventedReport(nullable), integrationInventedExpected(nullable))
	}
}

func TestIntegrationBoundWireIsNotBaselineWire(t *testing.T) {
	mRaw, m := integrationModel(t)
	report := integrationInventedReport(false)
	calls := 0
	runner, e := scenarios.NewRunner(scenarios.Binding{ProfileVersion: "lifecycle/v1", ModelDigest: integrationModelDigest}, func(transport.Task, transport.Command) transport.Result {
		calls++
		t.Error("encoding or explanation invoked bound callback")
		return transport.Result{}
	})
	if e != nil {
		t.Fatal(e)
	}
	bound, e := runner.EncodeReport(report)
	if e != nil {
		t.Fatal(e)
	}
	var envelope struct {
		FormatVersion  string
		IdentitySource string
	}
	if e = json.Unmarshal(bound, &envelope); e != nil || envelope.FormatVersion != "lifecycle-bound-scenario-report/v1" || envelope.IdentitySource != "caller-supplied-go-binding" {
		t.Fatal("bound wire not distinguished", e)
	}
	decoded, e := scenarios.DecodeReport(bound)
	if e == nil || !reflect.DeepEqual(decoded, scenarios.Report{}) {
		t.Fatal("baseline decoder accepted bound wire", e)
	}
	out, e := scenarioexplain.Execute(integrationRequest(mRaw, bound))
	if e == nil || out != nil {
		t.Fatal("strict explanation accepted bound wire", e)
	}
	// Separate the version gate from the unknown-field gate: rejecting the
	// original bound envelope alone cannot establish that both are enforced.
	var fields map[string]json.RawMessage
	if e = json.Unmarshal(bound, &fields); e != nil {
		t.Fatal(e)
	}
	identity := bytes.Clone(fields["identitySource"])
	delete(fields, "identitySource")
	wrongVersionOnly, e := json.Marshal(fields)
	if e != nil {
		t.Fatal(e)
	}
	fields["identitySource"] = identity
	fields["formatVersion"] = json.RawMessage(`"lifecycle-scenario-report/v1"`)
	unknownFieldOnly, e := json.Marshal(fields)
	if e != nil {
		t.Fatal(e)
	}
	for _, raw := range [][]byte{wrongVersionOnly, unknownFieldOnly} {
		decoded, e := scenarios.DecodeReport(raw)
		if e == nil || !reflect.DeepEqual(decoded, scenarios.Report{}) {
			t.Fatal("independent report envelope gate bypassed", e)
		}
		out, e := scenarioexplain.Execute(integrationRequest(mRaw, raw))
		if e == nil || out != nil {
			t.Fatal("strict explanation envelope gate bypassed", e)
		}
	}
	// The typed Report has no origin discriminator. Passing it remains legal;
	// this does not launder it into execution proof or a baseline wire receipt.
	got, e := scenarioexplain.Explain(m, report)
	if e != nil || !reflect.DeepEqual(got, integrationInventedExpected(false)) {
		t.Fatal("compatible typed report rejected or promoted", e)
	}
	if calls != 0 {
		t.Fatal("callback executed", calls)
	}
}
func TestIntegrationChangedModelOrReportDigestRefused(t *testing.T) {
	mRaw, m := integrationModel(t)
	r := integrationInventedReport(false)
	changedRaw := integrationFile(t, "../tutorial/model-v2.json", "6c6820e6d8544136a5629aa2977aa5319d853a63a5f1363b86766d8204f73499")
	changed, e := jsonformat.Decode(changedRaw)
	if e != nil {
		t.Fatal(e)
	}
	got, e := scenarioexplain.Explain(changed, r)
	if e == nil || !reflect.DeepEqual(got, scenarioexplain.Explanation{}) {
		t.Fatal("candidate model accepted by baseline-only explainer")
	}
	wire, e := scenarios.EncodeReport(r)
	if e != nil {
		t.Fatal(e)
	}
	out, e := scenarioexplain.Execute(integrationRequest(changedRaw, wire))
	if e == nil || out != nil {
		t.Fatal("strict candidate model accepted")
	}
	r.ModelDigest = "c9bd33388e7595872bfab1f33edf905e232243f21f252bf3672fad4d340df1b8"
	got, e = scenarioexplain.Explain(m, r)
	if e == nil || !reflect.DeepEqual(got, scenarioexplain.Explanation{}) {
		t.Fatal("stale report digest accepted")
	}
	staleWire := bytes.Replace(wire, []byte(`"modelDigest":"`+integrationModelDigest+`"`), []byte(`"modelDigest":"`+r.ModelDigest+`"`), 1)
	out, e = scenarioexplain.Execute(integrationRequest(mRaw, staleWire))
	if e == nil || out != nil {
		t.Fatal("strict stale report accepted")
	}
}
