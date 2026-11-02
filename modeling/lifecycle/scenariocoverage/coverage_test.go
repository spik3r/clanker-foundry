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
    modelBytes, err := os.ReadFile("../examples/lifecycle.json"); if err != nil { t.Fatal(err) }
    suite, err := os.ReadFile("../scenarios/testdata/lifecycle-v1.json"); if err != nil { t.Fatal(err) }
    report, err := scenarios.Execute(suite); if err != nil { t.Fatal(err) }
    return modelBytes, report
}

func TestFrozenFixtureInventory(t *testing.T) {
    raw, report := sample(t)
    m, err := jsonformat.Decode(raw); if err != nil { t.Fatal(err) }
    got, err := Analyze(m, report); if err != nil { t.Fatal(err) }
    if got.Status != "caller-supplied-assertions" || !got.InputMatched || got.StepsExecuted != 13 { t.Fatalf("wrong envelope: %+v", got) }
    ids := []string{}; for _, tr := range got.Transitions { ids = append(ids, tr.ID) }
    if !reflect.DeepEqual(ids, []string{"T_approve", "T_prepare", "T_requestChanges", "T_requestReview", "T_start"}) { t.Fatalf("transitions: %v", ids) }
    refs := func(v ...string) []Reference { out := []Reference{}; for i:=0;i<len(v);i+=2 { out=append(out, Reference{CaseID:v[i],StepID:v[i+1]}) }; return out }
    check := func(label string, have, want []Reference) { t.Helper(); if !reflect.DeepEqual(have,want) { t.Errorf("%s: got %#v want %#v",label,have,want) } }
    approve := got.Transitions[0]
    check("approve successes",approve.Successes,refs("happy","approve","rework_and_replay","fresh_approval"))
    guards := map[string][]Reference{}; for _, g := range approve.Guards { guards[g.ID]=g.Refusals }
    check("review content",guards["G_reviewContent"],refs("rework_and_replay","stale_review_fresh_command","rework_and_replay","replayed_entire_command","same_revision_content_mismatch","old_content_review"))
    check("review revision",guards["G_reviewRevision"],refs("rework_and_replay","stale_review_fresh_command","rework_and_replay","replayed_entire_command"))
    check("revision",guards["G_revision"],refs("rework_and_replay","replayed_entire_command"))
    check("unlisted guard not inferred",guards["G_approved"],refs())
    start := got.Transitions[4]
    check("start successes",start.Successes,refs("happy","start","refusal_then_success","start_after_both_refusals"))
    for _, g := range start.Guards { if g.ID == "G_actor" { check("actor refusal",g.Refusals,refs("refusal_then_success","blank_actor")) } }
    check("invalid input",got.InvalidInput,refs("refusal_then_success","unknown_command"))
    check("illegal transition",got.IllegalTransition,refs())
    check("excluded",got.Excluded,refs())
}

func TestRejectStaleModelAndInconsistentReport(t *testing.T) {
    raw, r := sample(t); m, err := jsonformat.Decode(raw); if err != nil { t.Fatal(err) }
    stale := m; stale.ID += "_different"
    if _, err := Analyze(stale,r); err == nil { t.Fatal("accepted different model") }
    r.StepsExecuted++
    if _, err := Analyze(m,r); err == nil { t.Fatal("accepted inconsistent report") }
}
