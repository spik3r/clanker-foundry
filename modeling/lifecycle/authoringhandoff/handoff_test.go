// Package authoringhandoff_test checks a proposal, not its behavior. Independent
// expectations stay frozen and unexecuted. No production API is added.
package authoringhandoff_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"go/parser"
	"go/token"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/acceptanceauthor"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/canonical"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/compiler"
	lifecycle "github.com/spik3r/clanker-foundry/modeling/lifecycle/generated"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/intakemodel"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/jsonformat"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/model"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/profilecatalog"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/scenariodiff"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/scenarios"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/views"
)

const inventorySHA = "8377777ad497b3c8901cf77738774498962e29bcfe26874d59efcb10d5c0755d"
const modelDigest = "4280fa4c999f7291c3325a997887c5e7b02b565f275a53df158afb55636d7db4"
const suiteDigest = "f0cea0b128e3e21b76c6ab2b15fefc841e72517fbfb37d12b8e41ee4d3e43ef9"

func hash(b []byte) string { s := sha256.Sum256(b); return hex.EncodeToString(s[:]) }
func fixtures(t *testing.T) map[string][]byte {
	t.Helper()
	root := filepath.Join("..", "..", "..", "skills", "model-driven-design", "examples")
	raw, e := os.ReadFile(filepath.Join(root, "inventory.json"))
	if e != nil {
		t.Fatal(e)
	}
	if hash(raw) != inventorySHA {
		t.Fatal("frozen inventory drift; independent review required, do not refresh from results")
	}
	var inv struct {
		FormatVersion, ProfileVersion, ModelCanonicalSHA256, SuiteSemanticSHA256 string
		Files                                                                    []struct {
			Name   string
			Bytes  int
			SHA256 string
		}
	}
	if e = json.Unmarshal(raw, &inv); e != nil {
		t.Fatal(e)
	}
	if inv.FormatVersion != "foundry-first-design-example-inventory/v1" || inv.ProfileVersion != model.ProfileVersion || inv.ModelCanonicalSHA256 != modelDigest || inv.SuiteSemanticSHA256 != suiteDigest {
		t.Fatal("inventory identity")
	}
	names := []string{"brief.md", "model.json", "suite.json", "manifest.json", "intake.json"}
	if len(inv.Files) != len(names) {
		t.Fatal("inventory count")
	}
	out := map[string][]byte{}
	for i, name := range names {
		f := inv.Files[i]
		if f.Name != name {
			t.Fatal("inventory path/order")
		}
		b, e := os.ReadFile(filepath.Join(root, name))
		if e != nil {
			t.Fatal(e)
		}
		if len(b) != f.Bytes || hash(b) != f.SHA256 {
			t.Fatalf("frozen input drift: %s", name)
		}
		out[name] = b
	}
	return out
}
func runner(t *testing.T) scenarios.Runner {
	t.Helper()
	// Validation adapter only: this supplied binding is not compiled-code identity.
	// The API requires a callback for its pure methods; executing it is forbidden.
	r, e := scenarios.NewRunner(scenarios.Binding{ProfileVersion: model.ProfileVersion, ModelDigest: modelDigest}, func(lifecycle.Task, lifecycle.Command) lifecycle.Result { panic("handoff must not execute a scenario") })
	if e != nil {
		t.Fatal(e)
	}
	return r
}
func observedRequest(t *testing.T, f map[string][]byte, mb, sb []byte) []byte {
	t.Helper()
	var r map[string]json.RawMessage
	if e := json.Unmarshal(f["intake.json"], &r); e != nil {
		t.Fatal(e)
	}
	o, e := json.Marshal(map[string]string{"requirements": string(f["brief.md"]), "model": string(mb), "scenarioSuite": string(sb)})
	if e != nil {
		t.Fatal(e)
	}
	r["observed"] = o
	b, e := json.Marshal(r)
	if e != nil {
		t.Fatal(e)
	}
	return b
}

type nativeIntake struct {
	Decisions []struct{ ID string }
	Questions []struct{ ID string }
	Bindings  []struct{ Kind, Status string }
	Links     []struct{ From, To, Status string }
}

func native(t *testing.T, r *intakemodel.Report) nativeIntake {
	t.Helper()
	var n nativeIntake
	if e := json.Unmarshal([]byte(r.IntakeReportJSON), &n); e != nil {
		t.Fatal(e)
	}
	if hash([]byte(r.IntakeReportJSON)) != r.IntakeReportSHA256 {
		t.Fatal("native receipt hash")
	}
	return n
}
func assertUnrun(t *testing.T, r *intakemodel.Report) {
	t.Helper()
	if r.Mode != "example-only" || r.Authority != "none" || r.ExecutionStatus != "not-run" || r.AcceptanceStatus != "not-evaluated" || r.ReviewStatus != "not-established" || r.Schema.Validation != "not-run" {
		t.Fatal("unsupported evidence claim", r)
	}
}

func TestFrozenFirstDesignDeclarations(t *testing.T) {
	f := fixtures(t)
	m, e := jsonformat.Decode(f["model.json"])
	if e != nil {
		t.Fatal(e)
	}
	d, e := canonical.Digest(m)
	if e != nil || d != modelDigest {
		t.Fatal("independent model identity", d, e)
	}
	p, e := profilecatalog.Describe()
	if e != nil || p.Identities.ProfileVersion != model.ProfileVersion {
		t.Fatal("profile unavailable", e)
	}
	r, e := intakemodel.Inspect(observedRequest(t, f, f["model.json"], f["suite.json"]))
	if e != nil {
		t.Fatal(e)
	}
	assertUnrun(t, r)
	if r.Model.Status != "accepted" || r.Model.CanonicalModelSHA256 == nil || *r.Model.CanonicalModelSHA256 != modelDigest || r.Model.SourceSHA256 == nil || *r.Model.SourceSHA256 != hash(f["model.json"]) {
		t.Fatal("model receipt", r.Model)
	}
	n := native(t, r)
	if len(n.Decisions) != 5 || len(n.Questions) != 1 || n.Questions[0].ID != "Q_use" {
		t.Fatal("human decisions/question lost", n)
	}
	if len(n.Bindings) != 3 || len(n.Links) != 3 {
		t.Fatal("missing identity evidence")
	}
	for _, b := range n.Bindings {
		if b.Status != "matches-declaration" {
			t.Fatal("source mismatch", b)
		}
	}
	for _, l := range n.Links {
		if l.Status != "matches-declaration" {
			t.Fatal("dependency mismatch", l)
		}
	}
	v := runner(t)
	s, e := v.Decode(f["suite.json"])
	if e != nil {
		t.Fatal(e)
	}
	sd, e := v.Digest(s)
	if e != nil || sd != suiteDigest || len(s.Cases) != 9 {
		t.Fatal("independent suite identity/count", sd, e)
	}
	a, e := acceptanceauthor.CheckWithRunner(f["manifest.json"], f["model.json"], f["suite.json"], "check", v)
	if e != nil || a.SuiteValidation != "valid" || a.Replay != "not-run" || a.ScenarioDigest == nil || *a.ScenarioDigest != suiteDigest || len(a.ScenarioReport) != 0 {
		t.Fatalf("check-only acceptance %+v %v", a, e)
	}
	cmp, e := scenariodiff.Compare(m, s, m, s)
	if e != nil || cmp.ReviewRequired || cmp.BeforeSuiteDigest != suiteDigest {
		t.Fatal("self-comparison declaration check", e)
	}
	// Identical copies do not answer Q_use or promote model validity to approval.
	if native(t, r).Questions[0].ID != "Q_use" || r.ReviewStatus != "not-established" {
		t.Fatal("comparison erased human gate")
	}
}

func TestGeneratedReviewArtifactsRemainDeclarations(t *testing.T) {
	f := fixtures(t)
	m, e := jsonformat.Decode(f["model.json"])
	if e != nil {
		t.Fatal(e)
	}
	implementation, derived, e := compiler.Generate(m)
	if e != nil {
		t.Fatal(e)
	}
	for name, b := range map[string][]byte{"lifecycle.go": implementation, "model_paths_test.go": derived} {
		if _, e = parser.ParseFile(token.NewFileSet(), name, b, parser.AllErrors); e != nil {
			t.Fatal("generated syntax", e)
		}
	}
	id := views.Identity{ModelDigest: modelDigest, FormatVersion: jsonformat.FormatVersion, ProfileVersion: model.ProfileVersion, GeneratorVersion: compiler.GeneratorVersion, ToolchainVersion: compiler.ToolchainVersion}
	v, e := views.Render(m, id)
	if e != nil {
		t.Fatal(e)
	}
	dec := xml.NewDecoder(bytes.NewReader(v.SVG))
	roots := 0
	for {
		tok, e := dec.Token()
		if e == io.EOF {
			break
		}
		if e != nil {
			t.Fatal("SVG syntax, not visual inspection", e)
		}
		if el, ok := tok.(xml.StartElement); ok && el.Name.Local == "svg" {
			roots++
		}
	}
	if roots != 1 {
		t.Fatal("missing SVG root")
	}
	var table struct {
		Identity views.Identity
		Model    model.Model
	}
	if e = json.Unmarshal(v.DecisionTable, &table); e != nil || !reflect.DeepEqual(table.Model, m) || table.Identity != id {
		t.Fatal("view/model mismatch", e)
	}
	trace, e := views.Trace(m, id, v, views.CodeArtifacts{Implementation: implementation, DerivedTests: derived})
	if e != nil {
		t.Fatal(e)
	}
	var tr struct {
		Artifacts []struct{ Path, SHA256 string }
	}
	if e = json.Unmarshal(trace, &tr); e != nil {
		t.Fatal(e)
	}
	expected := map[string]string{"generated/decision.svg": hash(v.SVG), "generated/decision-table.json": hash(v.DecisionTable), "generated/lifecycle.go": hash(implementation), "generated/model_paths_test.go": hash(derived)}
	if len(tr.Artifacts) != len(expected) {
		t.Fatal("trace artifact count", string(trace))
	}
	for _, a := range tr.Artifacts {
		if expected[a.Path] != a.SHA256 {
			t.Fatal("artifact identity", a)
		}
	}
	again, e := views.Render(m, id)
	if e != nil || !bytes.Equal(v.SVG, again.SVG) || !bytes.Equal(v.DecisionTable, again.DecisionTable) {
		t.Fatal("fixed-input view drift", e)
	}
}

func TestDriftAndInvalidityStayVisibleWithoutRepair(t *testing.T) {
	f := fixtures(t)
	t.Run("missing-observations", func(t *testing.T) {
		r, e := intakemodel.Inspect(f["intake.json"])
		if e != nil || r.Model.Status != "missing" {
			t.Fatal(e)
		}
		assertUnrun(t, r)
		for _, b := range native(t, r).Bindings {
			if b.Status != "missing-observation" {
				t.Fatal(b)
			}
		}
	})
	t.Run("model-source-only-drift", func(t *testing.T) {
		changed := append(bytes.Clone(f["model.json"]), '\n')
		r, e := intakemodel.Inspect(observedRequest(t, f, changed, f["suite.json"]))
		if e != nil || r.Model.Status != "accepted" || *r.Model.CanonicalModelSHA256 != modelDigest {
			t.Fatal(e)
		}
		assertUnrun(t, r)
		n := native(t, r)
		found := false
		for _, b := range n.Bindings {
			if b.Kind == "model" && b.Status == "digest-mismatch" {
				found = true
			}
		}
		if !found || len(n.Questions) != 1 {
			t.Fatal("source drift/question hidden", n)
		}
	})
	t.Run("invalid-model", func(t *testing.T) {
		r, e := intakemodel.Inspect(observedRequest(t, f, []byte(`{"id":"not a complete model"}`), f["suite.json"]))
		if e != nil || r.Model.Status != "rejected" || r.Model.Normalized != nil || r.Model.CanonicalModelSHA256 != nil {
			t.Fatal("invalid source accepted", e)
		}
		assertUnrun(t, r)
	})
	t.Run("suite-source-only-drift", func(t *testing.T) {
		changed := append(bytes.Clone(f["suite.json"]), '\n')
		a, e := acceptanceauthor.CheckWithRunner(f["manifest.json"], f["model.json"], changed, "check", runner(t))
		if e != nil || a.SuiteBinding != "stale" || a.SuiteValidation != "not-checked" || a.Replay != "not-run" {
			t.Fatalf("stale expectations hidden %+v %v", a, e)
		}
		if !bytes.Equal(f["suite.json"], fixtures(t)["suite.json"]) {
			t.Fatal("expectations changed")
		}
	})
	t.Run("stale-model-binding", func(t *testing.T) {
		changed := bytes.Replace(f["suite.json"], []byte(modelDigest), []byte(strings.Repeat("0", 64)), 1)
		if _, e := runner(t).Decode(changed); e != scenarios.ErrModelBinding {
			t.Fatal("stale binding accepted", e)
		}
		a, e := acceptanceauthor.CheckWithRunner(f["manifest.json"], f["model.json"], changed, "check", runner(t))
		if e != nil || a.SuiteModelBinding != "stale" || a.Replay != "not-run" {
			t.Fatal("stale header repaired", e)
		}
	})
}
