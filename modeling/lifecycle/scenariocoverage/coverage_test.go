package scenariocoverage

import (
	"os"
	"reflect"
	"testing"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/jsonformat"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/scenarios"
)

// Literal expectations transcribed from the pre-existing independently authored
// scenario fixture and model before implementing the inventory mapping. The
// fixture's author is independent of its original runner; these new inventory
// expectations were authored by this implementation contributor, not an
// independent reviewer. They are assertions about the supplied report only.
func sample(t *testing.T) ([]byte, scenarios.Report) {
	t.Helper()
	modelBytes, err := os.ReadFile("../examples/lifecycle.json")
	if err != nil {
		t.Fatal(err)
	}
	suite, err := os.ReadFile("../scenarios/testdata/lifecycle-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	report, err := scenarios.Execute(suite)
	if err != nil {
		t.Fatal(err)
	}
	return modelBytes, report
}

func TestFrozenFixtureInventory(t *testing.T) {
	raw, report := sample(t)
	m, err := jsonformat.Decode(raw)
	if err != nil {
		t.Fatal(err)
	}
	got, err := Analyze(m, report)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != "caller-supplied-assertions" || !got.InputMatched || got.StepsExecuted != 13 {
		t.Fatalf("wrong envelope: %+v", got)
	}
	ids := []string{}
	for _, tr := range got.Transitions {
		ids = append(ids, tr.ID)
	}
	if !reflect.DeepEqual(ids, []string{"T_approve", "T_prepare", "T_requestChanges", "T_requestReview", "T_start"}) {
		t.Fatalf("transitions: %v", ids)
	}
	refs := func(v ...string) []Reference {
		out := []Reference{}
		for i := 0; i < len(v); i += 2 {
			out = append(out, Reference{CaseID: v[i], StepID: v[i+1]})
		}
		return out
	}
	check := func(label string, have, want []Reference) {
		t.Helper()
		if !reflect.DeepEqual(have, want) {
			t.Errorf("%s: got %#v want %#v", label, have, want)
		}
	}
	approve := got.Transitions[0]
	check("approve successes", approve.Successes, refs("happy", "approve", "rework_and_replay", "fresh_approval"))
	guards := map[string][]Reference{}
	for _, g := range approve.Guards {
		guards[g.ID] = g.Refusals
	}
	check("review content", guards["G_reviewContent"], refs("rework_and_replay", "stale_review_fresh_command", "rework_and_replay", "replayed_entire_command", "same_revision_content_mismatch", "old_content_review"))
	check("review revision", guards["G_reviewRevision"], refs("rework_and_replay", "stale_review_fresh_command", "rework_and_replay", "replayed_entire_command"))
	check("revision", guards["G_revision"], refs("rework_and_replay", "replayed_entire_command"))
	check("unlisted guard not inferred", guards["G_approved"], refs())
	start := got.Transitions[4]
	check("start successes", start.Successes, refs("happy", "start", "refusal_then_success", "start_after_both_refusals"))
	for _, g := range start.Guards {
		if g.ID == "G_actor" {
			check("actor refusal", g.Refusals, refs("refusal_then_success", "blank_actor"))
		}
	}
	check("invalid input", got.InvalidInput, refs("refusal_then_success", "unknown_command"))
	check("illegal transition", got.IllegalTransition, refs())
	check("excluded", got.Excluded, refs())
}

func TestRejectStaleModelAndInconsistentReport(t *testing.T) {
	raw, r := sample(t)
	m, err := jsonformat.Decode(raw)
	if err != nil {
		t.Fatal(err)
	}
	stale := m
	stale.ID += "_different"
	if _, err := Analyze(stale, r); err == nil {
		t.Fatal("accepted different model")
	}
	r.StepsExecuted++
	if _, err := Analyze(m, r); err == nil {
		t.Fatal("accepted inconsistent report")
	}
}

// Reports are assertions, not authenticated executions: deliberately fabricated
// but structurally consistent envelopes must remain labeled as assertions.
func TestFixedReasonsAndMismatchExclusion(t *testing.T) {
	raw, base := sample(t)
	m, err := jsonformat.Decode(raw)
	if err != nil {
		t.Fatal(err)
	}
	for _, reason := range []string{"INVALID_INPUT", "ILLEGAL_TRANSITION", "REVISION_LIMIT"} {
		t.Run(reason, func(t *testing.T) {
			r := base
			e := base.Entries[0]
			r.Entries = []scenarios.Entry{e}
			r.StepsExecuted = 1
			r.CasesCompleted = 1
			r.Matched = true
			e.Actual.OK = false
			e.Actual.Refusals = []string{reason}
			e.Expected.OK = false
			e.Expected.Refusals = []string{reason}
			e.Actual.TransitionID = ""
			e.Expected.TransitionID = ""
			if reason == "INVALID_INPUT" {
				e.Actual.Task = nil
				e.Expected.Task = nil
			} else {
				task := e.Input
				e.Actual.Task = &task
				e.Expected.Task = &task
			}
			if reason == "REVISION_LIMIT" {
				e.Actual.TransitionID = "T_prepare"
				e.Expected.TransitionID = "T_prepare"
			}
			e.Matched = true
			e.Differences = []string{}
			r.Entries[0] = e
			got, err := Analyze(m, r)
			if err != nil {
				t.Fatal(err)
			}
			ref := []Reference{{CaseID: "happy", StepID: "prepare"}}
			switch reason {
			case "INVALID_INPUT":
				if !reflect.DeepEqual(got.InvalidInput, ref) {
					t.Fatal(got.InvalidInput)
				}
			case "ILLEGAL_TRANSITION":
				if !reflect.DeepEqual(got.IllegalTransition, ref) {
					t.Fatal(got.IllegalTransition)
				}
			case "REVISION_LIMIT":
				if !reflect.DeepEqual(got.Transitions[1].RevisionLimit, ref) {
					t.Fatal(got.Transitions[1].RevisionLimit)
				}
			}
			if got.Status != "caller-supplied-assertions" {
				t.Fatal(got.Status)
			}
		})
	}
	// A final mismatch remains reported but contributes no matching evidence.
	r := base
	r.Entries = append([]scenarios.Entry(nil), base.Entries...)
	n := len(r.Entries) - 1
	r.Entries[n].Expected.TransitionID = "invented"
	r.Entries[n].Matched = false
	r.Entries[n].Differences = []string{"transitionId"}
	r.Matched = false
	r.CasesCompleted--
	got, err := Analyze(m, r)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got.Excluded, []Reference{{CaseID: "same_revision_content_mismatch", StepID: "old_content_review"}}) {
		t.Fatal(got.Excluded)
	}
	for _, g := range got.Transitions[0].Guards {
		if g.ID == "G_reviewContent" && len(g.Refusals) != 2 {
			t.Fatal(g.Refusals)
		}
	}
}

func TestOwnedOutputAndInvalidMappings(t *testing.T) {
	raw, base := sample(t)
	m, err := jsonformat.Decode(raw)
	if err != nil {
		t.Fatal(err)
	}
	before, err := scenarios.EncodeReport(base)
	if err != nil {
		t.Fatal(err)
	}
	first, err := Analyze(m, base)
	if err != nil {
		t.Fatal(err)
	}
	first.Transitions[0].Guards[0].ID = "changed"
	first.Transitions[0].Successes[0].CaseID = "changed"
	if first.SourceDigest != nil {
		*first.SourceDigest = "changed"
	}
	second, err := Analyze(m, base)
	if err != nil {
		t.Fatal(err)
	}
	if second.Transitions[0].Guards[0].ID == "changed" || second.Transitions[0].Successes[0].CaseID == "changed" || *second.SourceDigest == "changed" {
		t.Fatal("output aliases input or later result")
	}
	after, err := scenarios.EncodeReport(base)
	if err != nil || !reflect.DeepEqual(before, after) {
		t.Fatal("input report changed")
	}
	// Keep the comparison internally consistent while claiming an unknown
	// actual guard; the existing explanation validation must reject it.
	bad := base
	bad.Entries = append([]scenarios.Entry(nil), base.Entries...)
	i := 5
	bad.Entries[i].Actual.Refusals = []string{"G_notDeclared"}
	bad.Entries[i].Expected.Refusals = []string{"G_notDeclared"}
	if _, err := Analyze(m, bad); err == nil {
		t.Fatal("unknown actual guard accepted")
	}
}

func TestReorderedModelAndLateFailure(t *testing.T) {
	raw, r := sample(t)
	m, err := jsonformat.Decode(raw)
	if err != nil {
		t.Fatal(err)
	}
	first, err := Analyze(m, r)
	if err != nil {
		t.Fatal(err)
	}
	for i, j := 0, len(m.Transitions)-1; i < j; i, j = i+1, j-1 {
		m.Transitions[i], m.Transitions[j] = m.Transitions[j], m.Transitions[i]
	}
	for i := range m.Transitions {
		g := m.Transitions[i].Guards
		for a, b := 0, len(g)-1; a < b; a, b = a+1, b-1 {
			g[a], g[b] = g[b], g[a]
		}
	}
	reordered, err := Analyze(m, r)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(first, reordered) {
		t.Fatal("set order changed inventory")
	}
	bad := r
	bad.Entries = append([]scenarios.Entry(nil), r.Entries...)
	last := len(bad.Entries) - 1
	bad.Entries[last].Actual.Refusals = []string{"G_reviewContent", "INVALID_INPUT"}
	bad.Entries[last].Expected.Refusals = []string{"G_reviewContent", "INVALID_INPUT"}
	// Keep report comparison internally consistent, then rely on whole-input
	// explanation classification to reject mixed fixed and model reasons.
	result, err := Analyze(m, bad)
	if err == nil || !reflect.DeepEqual(result, Inventory{}) {
		t.Fatal("late invalid entry returned inventory")
	}
}
