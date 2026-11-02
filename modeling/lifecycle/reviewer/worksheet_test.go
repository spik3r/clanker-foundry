// Package reviewer checks the frozen, hand-written review example.
package reviewer

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/intake"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/traceability"
)

const (
	intakeHash     = "312ea20f1c5eca4d281a09451097767de28c2c69cda3729a6e139b72d25f0779"
	reqHash        = "8f50fa3dbaf95ce1a5530a8256ecb42572ab96bb2ebb475177b1d07d446040d3"
	modelHash      = "a60916ca68771b28ee48712288decad7eb741d6a18e1cbc95b2a196eb10d12d8"
	catalogHash    = "bb886d8a8d63a00f6b6a225c9b4855c306e5cfd0dc45a4f262d410842076acae"
	reqNormalized  = "c951e96de70fc9be968b01395cfc459e562ebd0711e15ec7999543d3bf873bd3"
	modelCanonical = "948630cb7857199443b59b92d9b1ab0fadf2bf7a5bf8125f7a7e767db83958ed"
)

func read(t *testing.T, path string) []byte {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return b
}
func hash(b []byte) string { return fmt.Sprintf("%x", sha256.Sum256(b)) }

type intakeReport struct {
	RequestSHA256 string `json:"requestSha256"`
	Decisions     []struct {
		ID string `json:"id"`
	} `json:"decisions"`
	Questions []struct {
		ID string `json:"id"`
	} `json:"questions"`
	Bindings []struct {
		Kind     string  `json:"kind"`
		Status   string  `json:"status"`
		Observed *string `json:"observedSha256"`
		Declared *struct {
			SHA256 string `json:"sha256"`
			Author string `json:"author"`
		} `json:"declared"`
	} `json:"bindings"`
	Links []struct {
		From     string  `json:"from"`
		To       string  `json:"to"`
		Status   string  `json:"status"`
		Asserted *string `json:"assertedSha256"`
	} `json:"links"`
}

func report(t *testing.T, raw []byte) intakeReport {
	t.Helper()
	b, err := intake.Execute(raw)
	if err != nil {
		t.Fatal(err)
	}
	var r intakeReport
	if err := json.Unmarshal(b, &r); err != nil {
		t.Fatal(err)
	}
	return r
}

func TestFrozenIdentitiesAndIntake(t *testing.T) {
	request := read(t, "example-intake.json")
	doc := string(read(t, "example.md"))
	files := []struct{ path, want string }{
		{"example-intake.json", intakeHash}, {"../traceability/requirements.json", reqHash},
		{"../examples/lifecycle.json", modelHash}, {"../examples/oracle.json", catalogHash},
	}
	for _, f := range files {
		if got := hash(read(t, f.path)); got != f.want {
			t.Errorf("%s hash %s; want %s", f.path, got, f.want)
		}
		if !strings.Contains(doc, "`"+f.want+"`") {
			t.Errorf("worksheet omitted %s identity", f.path)
		}
	}
	for _, digest := range []string{reqNormalized, modelCanonical} {
		if !strings.Contains(doc, "`"+digest+"`") {
			t.Errorf("worksheet omitted normalized identity %s", digest)
		}
	}
	r := report(t, request)
	if r.RequestSHA256 != intakeHash || len(r.Bindings) != 3 || len(r.Links) != 3 {
		t.Fatalf("unexpected intake report: %+v", r)
	}
	for i, want := range []struct{ kind, status, digest, author string }{
		{"requirements", "missing-observation", reqHash, "Invented requirements author"},
		{"model", "missing-observation", modelHash, "Invented model author"},
		{"scenarioSuite", "missing-declaration", "", ""},
	} {
		b := r.Bindings[i]
		if b.Kind != want.kind || b.Status != want.status || b.Observed != nil {
			t.Fatalf("binding %d: %+v", i, b)
		}
		if want.digest == "" {
			if b.Declared != nil {
				t.Fatalf("unexpected declaration: %+v", b)
			}
			continue
		}
		if b.Declared == nil || b.Declared.SHA256 != want.digest || b.Declared.Author != want.author {
			t.Fatalf("declaration %d: %+v", i, b)
		}
	}
	for i, want := range []struct{ from, to, status, digest string }{
		{"model", "requirements", "matches-declaration", reqHash},
		{"scenarioSuite", "requirements", "missing-assertion", ""},
		{"scenarioSuite", "model", "missing-assertion", ""},
	} {
		l := r.Links[i]
		if l.From != want.from || l.To != want.to || l.Status != want.status {
			t.Fatalf("link %d: %+v", i, l)
		}
		if (l.Asserted == nil) != (want.digest == "") || l.Asserted != nil && *l.Asserted != want.digest {
			t.Fatalf("assertion %d: %+v", i, l)
		}
	}
	var decisions, questions []string
	for _, d := range r.Decisions {
		decisions = append(decisions, d.ID)
	}
	for _, q := range r.Questions {
		questions = append(questions, q.ID)
	}
	if !reflect.DeepEqual(decisions, []string{"D_input_boundary", "D_revision"}) || !reflect.DeepEqual(questions, []string{"Q_input_evidence", "Q_revision_evidence"}) {
		t.Fatalf("unexpected decisions/questions: %v %v", decisions, questions)
	}
	for _, text := range []string{
		"| `D_input_boundary` | `UR-010`", "| `Q_input_evidence` | `UR-010`", "| `D_revision` | `UR-012`", "| `Q_revision_evidence` | `UR-012`",
		"Both bindings report `missing-observation`", "`missing-declaration`", "`matches-declaration`", "`missing-assertion`", "`not run`", "`not performed`",
	} {
		if !strings.Contains(doc, text) {
			t.Errorf("worksheet omitted %q", text)
		}
	}
}

func checkSelected(r traceability.Report) error {
	if r.Inputs.RequirementsDigest != reqNormalized || r.Inputs.ModelDigest != modelCanonical || r.Inputs.CatalogSHA256 != catalogHash {
		return fmt.Errorf("trace identity differs: %+v", r.Inputs)
	}
	byID := map[string]traceability.RequirementTrace{}
	for _, q := range r.Requirements {
		byID[q.ID] = q
	}
	a, ok := byID["UR-010"]
	if !ok || a.Status != "incomplete" || !reflect.DeepEqual(a.Gaps, []string{"no-executable-acceptance", "no-model-link"}) || len(a.ModelLinks) != 0 || len(a.AcceptanceLinks) != 2 || a.AcceptanceLinks[0].CaseID != "special-03" || a.AcceptanceLinks[0].ExecutionStatus != "pending" || a.AcceptanceLinks[1].CaseID != "special-04" || a.AcceptanceLinks[1].ExecutionStatus != "pending" {
		return fmt.Errorf("UR-010 differs: %+v", a)
	}
	b, ok := byID["UR-012"]
	if !ok || b.Status != "incomplete" || !reflect.DeepEqual(b.Gaps, []string{"no-acceptance-link"}) || len(b.AcceptanceLinks) != 0 {
		return fmt.Errorf("UR-012 differs: %+v", b)
	}
	ids := []string{}
	for _, e := range b.ModelLinks {
		if e.Kind != "transition" {
			return fmt.Errorf("unexpected kind %s", e.Kind)
		}
		ids = append(ids, e.ID)
	}
	if !reflect.DeepEqual(ids, []string{"T_approve", "T_prepare", "T_requestChanges", "T_requestReview", "T_start"}) {
		return fmt.Errorf("UR-012 links %v", ids)
	}
	return nil
}

func TestTraceClaimsAndMutationSensitivity(t *testing.T) {
	model, catalog, req := read(t, "../examples/lifecycle.json"), read(t, "../examples/oracle.json"), read(t, "../traceability/requirements.json")
	trace, err := traceability.Build(model, catalog, req)
	if err != nil {
		t.Fatal(err)
	}
	if err := checkSelected(trace); err != nil {
		t.Fatal(err)
	}
	doc := string(read(t, "example.md"))
	for _, s := range []string{
		"| `UR-010` | none | `special-03` pending; `special-04` pending | `no-executable-acceptance`; `no-model-link` |",
		"| `UR-012` | `T_approve`, `T_prepare`, `T_requestChanges`, `T_requestReview`, `T_start` | none | `no-acceptance-link` |",
	} {
		if !strings.Contains(doc, s) {
			t.Errorf("worksheet trace row missing: %s", s)
		}
	}
	// Formatting changes exact source identity while leaving normalized trace identity unchanged.
	spaced := append(append([]byte{}, req...), '\n')
	if hash(spaced) == reqHash {
		t.Fatal("source-byte change not detected")
	}
	reformatted, err := traceability.Build(model, catalog, spaced)
	if err != nil || reformatted.Inputs.RequirementsDigest != reqNormalized {
		t.Fatalf("formatting changed normalized identity: %v", err)
	}
	// A still-valid link removal must change the independently stated trace facts.
	needle := []byte("        { \"caseId\": \"special-03\" },\n")
	if bytes.Count(req, needle) != 1 {
		t.Fatal("mutation target changed")
	}
	changed := bytes.Replace(req, needle, nil, 1)
	mutant, err := traceability.Build(model, catalog, changed)
	if err != nil {
		t.Fatalf("mutation must remain valid: %v", err)
	}
	for _, q := range mutant.Requirements {
		if q.ID == "UR-010" && (len(q.AcceptanceLinks) != 1 || q.AcceptanceLinks[0].CaseID != "special-04") {
			t.Fatalf("unexpected mutated links: %+v", q.AcceptanceLinks)
		}
	}
	if err := checkSelected(mutant); err == nil {
		t.Fatal("changed trace link escaped frozen expectations")
	}
}
