package intakeimpact_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"testing"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/acceptanceauthor"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/intake"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/intakeimpact"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/traceimpact"
)

// These integration expectations are stated before the new comparator exists.
// Existing frozen suite literals are consumed unchanged, never regenerated.
func TestExistingPipelineKeepsSourceAndSemanticBindingsSeparate(t *testing.T) {
	read := func(path string) []byte {
		t.Helper()
		b, e := os.ReadFile(path)
		if e != nil {
			t.Fatal(e)
		}
		return b
	}
	model := read("../examples/lifecycle.json")
	requirements := read("../traceability/requirements.json")
	catalog := read("../examples/oracle.json")
	suite := read("../acceptanceauthor/testdata/suite.json")
	manifest := read("../acceptanceauthor/testdata/manifest.json")
	digest := func(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }
	declaration := func(b []byte) any {
		return map[string]string{"sha256": digest(b), "author": "unverified example author"}
	}
	request := func(m, r, s []byte, intent string) []byte {
		t.Helper()
		v := map[string]any{
			"formatVersion": intake.FormatVersion, "intent": intent, "decisions": []any{},
			"questions":    []any{map[string]string{"id": "Q_review", "text": "Who will review this intent?"}},
			"artifacts":    map[string]any{"model": declaration(model), "requirements": declaration(requirements), "scenarioSuite": declaration(suite)},
			"observed":     map[string]any{"model": string(m), "requirements": string(r), "scenarioSuite": string(s)},
			"dependencies": map[string]string{"modelRequirements": digest(requirements), "suiteRequirements": digest(requirements), "suiteModel": digest(model)},
		}
		b, e := json.Marshal(v)
		if e != nil {
			t.Fatal(e)
		}
		return b
	}
	before := request(model, requirements, suite, "Review the example lifecycle")
	compact := func(b []byte) []byte {
		var out bytes.Buffer
		if e := json.Compact(&out, b); e != nil {
			t.Fatal(e)
		}
		return out.Bytes()
	}
	changedModel := bytes.Replace(model, []byte(`"label": "Draft"`), []byte(`"label": "Initial draft"`), 1)
	if bytes.Equal(changedModel, model) {
		t.Fatal("model fixture label not found")
	}
	for _, tc := range []struct {
		name                                                                string
		model, requirements, suite                                          []byte
		intent, artifact, modelBinding, suiteBinding, runnerBinding, replay string
		traceChanged                                                        bool
	}{
		{"intent only", model, requirements, suite, "Review a changed intent", "", "current", "current", "current", "matched", false},
		{"model formatting", compact(model), requirements, suite, "Review the example lifecycle", "model", "current", "current", "current", "matched", false},
		{"requirements formatting", model, compact(requirements), suite, "Review the example lifecycle", "requirements", "current", "current", "current", "matched", false},
		{"suite formatting", model, requirements, append(bytes.Clone(suite), '\n'), "Review the example lifecycle", "scenarioSuite", "current", "stale", "current", "not-run", false},
		{"model editorial meaning", changedModel, requirements, suite, "Review the example lifecycle", "model", "stale", "current", "stale", "not-run", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			after := request(tc.model, tc.requirements, tc.suite, tc.intent)
			bcopy, acopy := bytes.Clone(before), bytes.Clone(after)
			upstream, e := intakeimpact.Compare(before, after)
			if e != nil {
				t.Fatal(e)
			}
			if !upstream.ContentChanged || !upstream.RequestSourceChanged {
				t.Fatal("upstream change was hidden")
			}
			foundQuestion := false
			foundMismatch := tc.artifact == ""
			for _, g := range upstream.AfterGaps {
				if g.Kind == "question" && g.ID == "Q_review" && g.Status == "unresolved" {
					foundQuestion = true
				}
				if g.Kind == "artifact" && g.ID == tc.artifact && g.Status == "digest-mismatch" {
					foundMismatch = true
				}
			}
			if !foundQuestion || !foundMismatch {
				t.Fatalf("lost unresolved question or original binding: %+v", upstream.AfterGaps)
			}
			trace, e := traceimpact.Compare(traceimpact.Snapshot{Model: model, Requirements: requirements, Catalog: catalog}, traceimpact.Snapshot{Model: tc.model, Requirements: tc.requirements, Catalog: catalog})
			if e != nil {
				t.Fatal(e)
			}
			if trace.ModelChanged != tc.traceChanged || trace.ReviewRequired != tc.traceChanged || trace.RequirementsChanged || trace.CatalogSourceChanged {
				t.Fatalf("wrong downstream identity change: %+v", trace)
			}
			// Replay here is an explicit independent test call, not performed by Compare.
			acceptance, e := acceptanceauthor.Check(manifest, tc.model, tc.suite, "replay")
			if e != nil {
				t.Fatal(e)
			}
			if acceptance.ModelBinding != tc.modelBinding || acceptance.SuiteBinding != tc.suiteBinding || acceptance.RunnerBinding != tc.runnerBinding || acceptance.Replay != tc.replay {
				t.Fatalf("wrong acceptance binding: %+v", acceptance)
			}
			if tc.replay == "not-run" && (acceptance.SuiteValidation != "not-checked" || acceptance.ScenarioReport != nil) {
				t.Fatal("stale suite was evaluated")
			}
			if upstream.ArtifactValidation != "not-run" || upstream.ExecutionStatus != "not-run" || upstream.ReviewStatus != "not-established" {
				t.Fatal("native execution leaked into upstream claims")
			}
			if !bytes.Equal(before, bcopy) || !bytes.Equal(after, acopy) {
				t.Fatal("inputs were rebound")
			}
		})
	}
	for path, want := range map[string][]byte{"../examples/lifecycle.json": model, "../traceability/requirements.json": requirements, "../examples/oracle.json": catalog, "../acceptanceauthor/testdata/suite.json": suite, "../acceptanceauthor/testdata/manifest.json": manifest} {
		if !bytes.Equal(read(path), want) {
			t.Fatalf("source mutated: %s", path)
		}
	}
}
