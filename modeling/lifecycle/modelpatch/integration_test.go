// Independent integration oracle, authored before local modelpatch adoption.
// Expected models/manifests/suites predate this test; no Plan/Execute output was
// used to derive an expectation. No reducer, compiler or process runner is used.
package modelpatch_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"testing"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/acceptanceauthor"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/canonical"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/impact"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/jsonformat"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/model"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/modelpatch"
)

const tutorialBeforeCanonical = "948630cb7857199443b59b92d9b1ab0fadf2bf7a5bf8125f7a7e767db83958ed"
const tutorialAfterCanonical = "c9bd33388e7595872bfab1f33edf905e232243f21f252bf3672fad4d340df1b8"

var tutorialFrozenFiles = map[string]string{
	"../examples/lifecycle.json":   "a60916ca68771b28ee48712288decad7eb741d6a18e1cbc95b2a196eb10d12d8",
	"../tutorial/model-v2.json":    "6c6820e6d8544136a5629aa2977aa5319d853a63a5f1363b86766d8204f73499",
	"../tutorial/manifest-v1.json": "57b88a554da0ff44d53d336324a295e4fe6eb6ddefb775cab996643eb7cdc3b9",
	"../tutorial/manifest-v2.json": "96df1a0eb978d30a9c59fad0a3049044c7fbd9610197b5c1a7e98dbbc5e37ae2",
	"../tutorial/suite-v1.json":    "2031108ec41558c01ae70e80b4e580f8af391213d4bb3db5e14ec2e080e3e2d8",
	"../tutorial/suite-v2.json":    "149aeb2ebe4ebbaea661ae944a7b30f083be406d97394189c2049d82b351baaf",
}

func tutorialSHA(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }
func tutorialRead(t *testing.T, path string) []byte {
	t.Helper()
	b, e := os.ReadFile(path)
	if e != nil {
		t.Fatal(e)
	}
	if tutorialSHA(b) != tutorialFrozenFiles[path] {
		t.Fatal("independent tutorial source drift", path)
	}
	return b
}
func tutorialDecode(t *testing.T, b []byte) model.Model {
	t.Helper()
	m, e := jsonformat.Decode(b)
	if e != nil {
		t.Fatal(e)
	}
	return m
}
func tutorialDigest(t *testing.T, m model.Model, want string) {
	t.Helper()
	d, e := canonical.Digest(m)
	if e != nil || d != want {
		t.Fatalf("canonical identity got %s/%v want %s", d, e, want)
	}
}

// Complete hand-authored replacement, not a record read from either model or
// copied from a patch result. Only the T_prepare guard-reference list changes.
func tutorialPreparePatch() modelpatch.Patch {
	return modelpatch.Patch{
		Version: 1, ProfileVersion: "lifecycle/v1", BeforeDigest: tutorialBeforeCanonical,
		Operations: []modelpatch.Operation{{Action: modelpatch.Replace, Kind: modelpatch.Transition, ID: "T_prepare", Transition: &model.Transition{
			ID: "T_prepare", From: "Draft", Trigger: "prepare", To: "Ready",
			Guards:  []string{"G_revision", "G_outcome", "G_acceptance", "G_actor"},
			Effects: []model.Effect{{Op: "set_state"}, {Op: "increment_revision"}},
		}}},
	}
}
func tutorialReviewExpected() impact.Report {
	return impact.Report{
		ReportVersion: "lifecycle-model-diff-report/v1", ProfileVersion: "lifecycle/v1", Mode: "example-only", Authority: "none",
		BeforeDigest: tutorialBeforeCanonical, AfterDigest: tutorialAfterCanonical, Changed: true, ReviewRequired: true,
		Changes:           []impact.Change{{Kind: "transition", ID: "T_prepare", Type: "modified", Fields: []string{"guards"}}},
		LinkedTransitions: []impact.LinkedTransition{{Side: "before", ID: "T_prepare", Causes: []string{"transition:T_prepare"}}, {Side: "after", ID: "T_prepare", Causes: []string{"transition:T_prepare"}}},
		Bindings:          []impact.BindingResult{},
	}
}
func tutorialAssertCandidate(t *testing.T, r modelpatch.Result, want model.Model) {
	t.Helper()
	if r.ReportVersion != "lifecycle-model-patch-report/v1" || r.ProfileVersion != "lifecycle/v1" || r.Mode != "example-only" || r.Authority != "none" || r.BeforeDigest != tutorialBeforeCanonical || r.AfterDigest != tutorialAfterCanonical {
		t.Fatal("proposal identity/authority mismatch", r)
	}
	if !reflect.DeepEqual(r.After, want) {
		t.Fatal("candidate differs from complete independently frozen tutorial model")
	}
	if !reflect.DeepEqual(r.Review, tutorialReviewExpected()) {
		t.Fatalf("declared-link review differs: %+v", r.Review)
	}
}

func TestFrozenTutorialPrepareProposalAndAcceptanceSeparation(t *testing.T) {
	rawBefore := tutorialRead(t, "../examples/lifecycle.json")
	rawAfter := tutorialRead(t, "../tutorial/model-v2.json")
	before := tutorialDecode(t, rawBefore)
	wantAfter := tutorialDecode(t, rawAfter)
	tutorialDigest(t, before, tutorialBeforeCanonical)
	tutorialDigest(t, wantAfter, tutorialAfterCanonical)
	p := tutorialPreparePatch()
	r, e := modelpatch.Plan(before, p)
	if e != nil {
		t.Fatal(e)
	}
	tutorialAssertCandidate(t, r, wantAfter)
	// Check both original input graphs again: Plan must not mutate its source.
	if !reflect.DeepEqual(before, tutorialDecode(t, rawBefore)) || !reflect.DeepEqual(p, tutorialPreparePatch()) {
		t.Fatal("Plan mutated source or operation")
	}
	normalized, e := jsonformat.Encode(r.After)
	if e != nil {
		t.Fatal(e)
	}
	expectedNormalized, e := jsonformat.Encode(wantAfter)
	if e != nil {
		t.Fatal(e)
	}
	if !bytes.Equal(normalized, expectedNormalized) {
		t.Fatal("normalized output differs from independent frozen model encoding")
	}
	if bytes.Equal(normalized, rawAfter) || tutorialSHA(normalized) == tutorialFrozenFiles["../tutorial/model-v2.json"] {
		t.Fatal("fixture must distinguish raw tutorial source from normalized output")
	}
	tutorialDigest(t, tutorialDecode(t, normalized), tutorialAfterCanonical)
	// Formatting is source identity, not the semantic precondition; no raw hash is
	// silently promoted to the canonical BeforeDigest.
	formatted := append(bytes.Clone(rawBefore), '\n')
	if tutorialSHA(formatted) == tutorialSHA(rawBefore) {
		t.Fatal("raw formatting identity did not change")
	}
	same, e := modelpatch.Plan(tutorialDecode(t, formatted), tutorialPreparePatch())
	if e != nil {
		t.Fatal(e)
	}
	tutorialAssertCandidate(t, same, wantAfter)
	for _, v := range []string{"v1", "v2"} {
		manifest := tutorialRead(t, "../tutorial/manifest-"+v+".json")
		suite := tutorialRead(t, "../tutorial/suite-"+v+".json")
		a, e := acceptanceauthor.Check(manifest, normalized, suite, "check")
		if e != nil {
			t.Fatal(e)
		}
		declared := tutorialBeforeCanonical
		modelBinding := "stale"
		if v == "v2" {
			declared = tutorialAfterCanonical
			modelBinding = "current"
		}
		if a.ModelBinding != modelBinding || a.SuiteBinding != "current" || a.SuiteModelBinding != "current" || a.RunnerBinding != "stale" || a.DeclaredModelDigest != declared || a.CurrentModelDigest != tutorialAfterCanonical || a.CompiledModelDigest != tutorialBeforeCanonical {
			t.Fatalf("%s bindings were rebound: %+v", v, a)
		}
		if a.Authority != "none" || a.Replay != "not-run" || a.SuiteValidation != "not-checked" || a.ScenarioReport != nil || a.ScenarioDigest != nil {
			t.Fatalf("%s proposal/check promoted execution or authority: %+v", v, a)
		}
		if a.ManifestSHA256 != tutorialSHA(manifest) || a.DeclaredSuiteSHA256 != tutorialSHA(suite) || a.ObservedSuiteSHA256 != tutorialSHA(suite) {
			t.Fatal("frozen manifest/suite identities rebound")
		}
	}
	for path := range tutorialFrozenFiles {
		tutorialRead(t, path)
	}
}

func TestFrozenTutorialPreconditionAndNoop(t *testing.T) {
	before := tutorialDecode(t, tutorialRead(t, "../examples/lifecycle.json"))
	after := tutorialDecode(t, tutorialRead(t, "../tutorial/model-v2.json"))
	p := tutorialPreparePatch()
	p.BeforeDigest = tutorialAfterCanonical
	bad, e := modelpatch.Plan(before, p)
	if !errors.Is(e, modelpatch.ErrPrecondition) || !reflect.DeepEqual(bad, modelpatch.Result{}) {
		t.Fatal("stale precondition emitted candidate", bad, e)
	}
	bad, e = modelpatch.Plan(after, tutorialPreparePatch())
	if !errors.Is(e, modelpatch.ErrPrecondition) || !reflect.DeepEqual(bad, modelpatch.Result{}) {
		t.Fatal("reapply original patch bypassed stale precondition", bad, e)
	}
	p = tutorialPreparePatch()
	p.BeforeDigest = tutorialFrozenFiles["../examples/lifecycle.json"]
	bad, e = modelpatch.Plan(before, p)
	if !errors.Is(e, modelpatch.ErrPrecondition) || !reflect.DeepEqual(bad, modelpatch.Result{}) {
		t.Fatal("raw source digest accepted as semantic identity", bad, e)
	}
	noop, e := modelpatch.Plan(before, modelpatch.Patch{Version: 1, ProfileVersion: "lifecycle/v1", BeforeDigest: tutorialBeforeCanonical, Operations: []modelpatch.Operation{}})
	if e != nil || noop.Mode != "example-only" || noop.Authority != "none" || noop.Review.Authority != "none" || noop.BeforeDigest != tutorialBeforeCanonical || noop.AfterDigest != tutorialBeforeCanonical || noop.Review.Changed || noop.Review.ReviewRequired || !reflect.DeepEqual(noop.After, before) || len(noop.Review.Changes) != 0 || len(noop.Review.LinkedTransitions) != 0 {
		t.Fatal("noop changed model or promoted authority", noop, e)
	}
}

// This strict wire operation independently spells the same complete record.
const tutorialPrepareWireOperation = `{"action":"replace","kind":"transition","id":"T_prepare","root":null,"state":null,"requirement":null,"guard":null,"transition":{"id":"T_prepare","from":"Draft","trigger":"prepare","to":"Ready","guards":["G_revision","G_outcome","G_acceptance","G_actor"],"effects":[{"op":"set_state"},{"op":"increment_revision"}]}}`

func tutorialWireRequest(raw []byte, digest string) []byte {
	return []byte(`{"formatVersion":"lifecycle-model-patch-json/v1","patchVersion":"lifecycle-model-patch/v1","profileVersion":"lifecycle/v1","before":` + string(raw) + `,"beforeDigest":"` + digest + `","operations":[` + tutorialPrepareWireOperation + `]}`)
}
func TestFrozenTutorialStrictWireProposal(t *testing.T) {
	raw := tutorialRead(t, "../examples/lifecycle.json")
	want := tutorialDecode(t, tutorialRead(t, "../tutorial/model-v2.json"))
	out, e := modelpatch.Execute(tutorialWireRequest(raw, tutorialBeforeCanonical))
	if e != nil {
		t.Fatal(e)
	}
	var dto struct {
		ReportVersion  string
		ProfileVersion string
		Mode           string
		Authority      string
		BeforeDigest   string
		AfterDigest    string
		After          json.RawMessage
		Review         impact.Report
	}
	if e = json.Unmarshal(out, &dto); e != nil {
		t.Fatal(e)
	}
	tutorialAssertCandidate(t, modelpatch.Result{ReportVersion: dto.ReportVersion, ProfileVersion: dto.ProfileVersion, Mode: dto.Mode, Authority: dto.Authority, BeforeDigest: dto.BeforeDigest, AfterDigest: dto.AfterDigest, After: tutorialDecode(t, dto.After), Review: dto.Review}, want)
	for _, request := range [][]byte{tutorialWireRequest(raw, tutorialAfterCanonical), tutorialWireRequest(dto.After, tutorialBeforeCanonical)} {
		b, e := modelpatch.Execute(request)
		if !errors.Is(e, modelpatch.ErrPrecondition) || b != nil {
			t.Fatal("wire stale proposal emitted output", string(b), e)
		}
	}
}
