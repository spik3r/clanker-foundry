// Package revisionreview checks a human walkthrough using existing APIs.
package revisionreview

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/acceptanceauthor"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/intakeimpact"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/intakemodel"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/traceimpact"
)

type acceptanceExpectation struct {
	ModelBinding      string `json:"modelBinding"`
	SuiteBinding      string `json:"suiteBinding"`
	SuiteModelBinding string `json:"suiteModelBinding"`
	RunnerBinding     string `json:"runnerBinding"`
	SuiteValidation   string `json:"suiteValidation"`
	Replay            string `json:"replay"`
}
type expectations struct {
	Question             string                `json:"question"`
	Requirement          string                `json:"requirement"`
	RequirementGap       string                `json:"requirementGap"`
	Transition           string                `json:"transition"`
	AddedGuard           string                `json:"addedGuard"`
	BeforeCanonicalModel string                `json:"beforeCanonicalModel"`
	AfterCanonicalModel  string                `json:"afterCanonicalModel"`
	IntakeModelStatus    string                `json:"intakeModelStatus"`
	SchemaValidation     string                `json:"schemaValidation"`
	Baseline             acceptanceExpectation `json:"baselineAcceptance"`
	OldCandidate         acceptanceExpectation `json:"oldAcceptanceAgainstCandidate"`
	NewCandidate         acceptanceExpectation `json:"candidateAcceptanceOnBaselineRunner"`
}

func file(t *testing.T, path string) []byte {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return b
}
func hash(b []byte) string { return fmt.Sprintf("%x", sha256.Sum256(b)) }

// withObservations fills only the three observation fields of a temporary request.
// The literal declarations and dependency assertions are never changed.
func withObservations(t *testing.T, declaration, requirements, model, suite []byte) []byte {
	t.Helper()
	var d map[string]json.RawMessage
	if err := json.Unmarshal(declaration, &d); err != nil {
		t.Fatal(err)
	}
	observed, err := json.Marshal(map[string]string{"requirements": string(requirements), "model": string(model), "scenarioSuite": string(suite)})
	if err != nil {
		t.Fatal(err)
	}
	d["observed"] = observed
	b, err := json.Marshal(d)
	if err != nil {
		t.Fatal(err)
	}
	return b
}
func gapsContain(gaps []intakeimpact.Gap, kind, id, status string) bool {
	for _, g := range gaps {
		if g.Kind == kind && g.ID == id && g.Status == status {
			return true
		}
	}
	return false
}
func changeContains(changes []intakeimpact.Change, kind, id string) bool {
	for _, c := range changes {
		if c.Kind == kind && c.ID == id && c.Change == "modified" {
			return true
		}
	}
	return false
}
func checkAcceptance(t *testing.T, manifest, model, suite []byte, want acceptanceExpectation) {
	t.Helper()
	r, err := acceptanceauthor.Check(manifest, model, suite, "check")
	if err != nil {
		t.Fatal(err)
	}
	got := acceptanceExpectation{r.ModelBinding, r.SuiteBinding, r.SuiteModelBinding, r.RunnerBinding, r.SuiteValidation, r.Replay}
	if got != want {
		t.Fatalf("acceptance got %+v; want %+v", got, want)
	}
	if r.ScenarioReport != nil {
		t.Fatalf("check produced an execution report: %s", r.ScenarioReport)
	}
}

func TestRevisionReviewAcrossExistingTools(t *testing.T) {
	var want expectations
	statement := file(t, "testdata/expected.json")
	if hash(statement) != "92c22d736e649c336e34e29682e3846ba8dae34907819c393344c761f78abf2a" {
		t.Fatal("frozen expectations changed; review the original before revising them")
	}
	if err := json.Unmarshal(statement, &want); err != nil {
		t.Fatal(err)
	}
	oldDecl, newDecl := file(t, "testdata/intake-v1.json"), file(t, "testdata/intake-v2.json")
	oldReq, newReq := file(t, "testdata/requirements-v1.json"), file(t, "testdata/requirements-v2.json")
	oldModel, newModel := file(t, "../examples/lifecycle.json"), file(t, "../tutorial/model-v2.json")
	oldSuite, newSuite := file(t, "../tutorial/suite-v1.json"), file(t, "../tutorial/suite-v2.json")
	oldManifest, newManifest := file(t, "../tutorial/manifest-v1.json"), file(t, "../tutorial/manifest-v2.json")
	catalog := file(t, "../examples/oracle.json")
	inputs := [][]byte{oldDecl, newDecl, oldReq, newReq, oldModel, newModel, oldSuite, newSuite, oldManifest, newManifest, catalog}
	beforeHashes := make([]string, len(inputs))
	for i, b := range inputs {
		beforeHashes[i] = hash(b)
	}
	before := withObservations(t, oldDecl, oldReq, oldModel, oldSuite)
	stale := withObservations(t, oldDecl, newReq, newModel, newSuite)
	after := withObservations(t, newDecl, newReq, newModel, newSuite)

	// Both accepted models retain distinct source and canonical identities. The
	// former declaration remains mismatched even when the model is valid.
	for _, tc := range []struct {
		name                       string
		raw                        []byte
		source, canonical, binding string
	}{
		{"baseline", before, hash(oldModel), want.BeforeCanonicalModel, "matches-declaration"},
		{"old declarations, new observations", stale, hash(newModel), want.AfterCanonicalModel, "digest-mismatch"},
		{"candidate declarations", after, hash(newModel), want.AfterCanonicalModel, "matches-declaration"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r, err := intakemodel.Inspect(tc.raw)
			if err != nil {
				t.Fatal(err)
			}
			if r.Model.Status != want.IntakeModelStatus || r.Model.SourceSHA256 == nil || *r.Model.SourceSHA256 != tc.source || r.Model.CanonicalModelSHA256 == nil || *r.Model.CanonicalModelSHA256 != tc.canonical || r.Schema.Validation != want.SchemaValidation || r.ExecutionStatus != "not-run" || r.AcceptanceStatus != "not-evaluated" || r.ReviewStatus != "not-established" {
				t.Fatalf("unexpected bridge: %+v", r)
			}
			var native struct {
				Bindings  []struct{ Kind, Status string } `json:"bindings"`
				Questions []struct{ ID string }           `json:"questions"`
			}
			if err := json.Unmarshal([]byte(r.IntakeReportJSON), &native); err != nil {
				t.Fatal(err)
			}
			if r.IntakeReportSHA256 != hash([]byte(r.IntakeReportJSON)) || len(native.Questions) != 1 || native.Questions[0].ID != want.Question {
				t.Fatal("receipt identity or open question lost")
			}
			found := false
			for _, b := range native.Bindings {
				if b.Kind == "model" && b.Status == tc.binding {
					found = true
				}
			}
			if !found {
				t.Fatalf("model binding missing: %+v", native.Bindings)
			}
		})
	}
	upstream, err := intakeimpact.Compare(before, after)
	if err != nil {
		t.Fatal(err)
	}
	if !upstream.ContentChanged || !changeContains(upstream.Changes, "decision", "D_actor") || !changeContains(upstream.Changes, "intent", "intent") || !gapsContain(upstream.BeforeGaps, "question", want.Question, "unresolved") || !gapsContain(upstream.AfterGaps, "question", want.Question, "unresolved") || upstream.ArtifactValidation != "not-run" || upstream.ExecutionStatus != "not-run" || upstream.ReviewStatus != "not-established" {
		t.Fatalf("unexpected intake comparison: %+v", upstream)
	}
	if len(upstream.BeforeGaps) != 1 || len(upstream.AfterGaps) != 1 {
		t.Fatalf("expected only the open question after matching all observations and assertions: %+v / %+v", upstream.BeforeGaps, upstream.AfterGaps)
	}
	oldObserved, err := intakeimpact.Compare(before, stale)
	if err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{"requirements", "model", "scenarioSuite"} {
		if !gapsContain(oldObserved.AfterGaps, "artifact", id, "digest-mismatch") {
			t.Errorf("old %s binding did not remain stale", id)
		}
	}

	trace, err := traceimpact.Compare(traceimpact.Snapshot{Model: oldModel, Requirements: oldReq, Catalog: catalog}, traceimpact.Snapshot{Model: newModel, Requirements: newReq, Catalog: catalog})
	if err != nil {
		t.Fatal(err)
	}
	if !trace.ModelChanged || !trace.RequirementsChanged || trace.CatalogSourceChanged || !trace.ReviewRequired || trace.Before.ModelDigest != want.BeforeCanonicalModel || trace.After.ModelDigest != want.AfterCanonicalModel || len(trace.Candidates) != 1 {
		t.Fatalf("unexpected trace comparison: %+v", trace)
	}
	c := trace.Candidates[0]
	if c.ID != want.Requirement || c.Change != "modified" || !reflect.DeepEqual(c.Fields, []string{"modelLinks", "statement"}) || !reflect.DeepEqual(c.BeforeGaps, []string{want.RequirementGap}) || !reflect.DeepEqual(c.AfterGaps, []string{want.RequirementGap}) {
		t.Fatalf("unexpected candidate: %+v", c)
	}
	linked := false
	for _, r := range c.Reasons {
		if r.Code == "linked-transition-changed" && r.Target == want.Transition {
			linked = true
		}
	}
	if !linked {
		t.Fatalf("changed transition missing: %+v", c.Reasons)
	}
	if !bytes.Contains(newReq, []byte(want.AddedGuard)) {
		t.Fatal("candidate requirement no longer declares the actor guard")
	}

	checkAcceptance(t, oldManifest, oldModel, oldSuite, want.Baseline)
	checkAcceptance(t, oldManifest, newModel, oldSuite, want.OldCandidate)
	checkAcceptance(t, newManifest, newModel, newSuite, want.NewCandidate)
	// A byte-only suite edit must become stale; the original expectation is not refreshed.
	alteredSuite := append(bytes.Clone(oldSuite), '\n')
	r, err := acceptanceauthor.Check(oldManifest, oldModel, alteredSuite, "check")
	if err != nil || r.SuiteBinding != "stale" || r.SuiteValidation != "not-checked" || r.Replay != "not-run" {
		t.Fatalf("suite drift was hidden: %+v %v", r, err)
	}
	for i, b := range inputs {
		if hash(b) != beforeHashes[i] {
			t.Fatalf("input %d mutated", i)
		}
	}
}
