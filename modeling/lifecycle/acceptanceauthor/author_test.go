package acceptanceauthor

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/canonical"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/jsonformat"
)

type frozenExpected struct {
	MatchingCheck struct {
		ModelBinding    string `json:"modelBinding"`
		SuiteBinding    string `json:"suiteBinding"`
		SuiteValidation string `json:"suiteValidation"`
		Replay          string `json:"replay"`
	} `json:"matchingCheck"`
	MatchingReplay struct {
		Replay         string `json:"replay"`
		CasesCompleted int    `json:"casesCompleted"`
		StepsExecuted  int    `json:"stepsExecuted"`
	} `json:"matchingReplay"`
	ChangedModel struct {
		ModelBinding    string `json:"modelBinding"`
		SuiteValidation string `json:"suiteValidation"`
		Replay          string `json:"replay"`
	} `json:"changedModel"`
	ChangedSuite struct {
		SuiteBinding    string `json:"suiteBinding"`
		SuiteValidation string `json:"suiteValidation"`
		Replay          string `json:"replay"`
	} `json:"changedSuite"`
	WrongExpectation struct {
		Replay         string `json:"replay"`
		CasesCompleted int    `json:"casesCompleted"`
		StepsExecuted  int    `json:"stepsExecuted"`
	} `json:"wrongExpectation"`
}

func fixture(t *testing.T) ([]byte, []byte, []byte, frozenExpected) {
	t.Helper()
	read := func(p string) []byte {
		b, e := os.ReadFile(p)
		if e != nil {
			t.Fatal(e)
		}
		return b
	}
	var ex frozenExpected
	if e := json.Unmarshal(read("testdata/expected.json"), &ex); e != nil {
		t.Fatal(e)
	}
	return read("testdata/manifest.json"), read("../examples/lifecycle.json"), read("testdata/suite.json"), ex
}
func counts(t *testing.T, r Report) (int, int) {
	t.Helper()
	var s struct {
		Cases int `json:"casesCompleted"`
		Steps int `json:"stepsExecuted"`
	}
	if e := json.Unmarshal(r.ScenarioReport, &s); e != nil {
		t.Fatal(e)
	}
	return s.Cases, s.Steps
}
func TestFrozenIndependentCases(t *testing.T) {
	manifestRaw, model, suite, expected := fixture(t)
	original := append([]byte(nil), suite...)
	r, e := Check(manifestRaw, model, suite, "check")
	if e != nil {
		t.Fatal(e)
	}
	w := expected.MatchingCheck
	if r.ModelBinding != w.ModelBinding || r.SuiteBinding != w.SuiteBinding || r.SuiteValidation != w.SuiteValidation || r.Replay != w.Replay || r.ScenarioReport != nil || r.ScenarioDigest == nil {
		t.Fatalf("check: %+v", r)
	}
	r, e = Check(manifestRaw, model, suite, "replay")
	if e != nil {
		t.Fatal(e)
	}
	cases, steps := counts(t, r)
	x := expected.MatchingReplay
	if r.Replay != x.Replay || cases != x.CasesCompleted || steps != x.StepsExecuted {
		t.Fatalf("replay: %+v", r)
	}
	if !bytes.Equal(original, suite) {
		t.Fatal("suite mutated")
	}
	changedModel := bytes.Replace(model, []byte(`"label": "Draft"`), []byte(`"label": "Draft altered"`), 1)
	if bytes.Equal(model, changedModel) {
		t.Fatal("fixture edit failed")
	}
	r, e = Check(manifestRaw, changedModel, suite, "replay")
	if e != nil {
		t.Fatal(e)
	}
	cm := expected.ChangedModel
	if r.ModelBinding != cm.ModelBinding || r.SuiteValidation != cm.SuiteValidation || r.Replay != cm.Replay || r.ScenarioReport != nil || r.ScenarioDigest != nil {
		t.Fatalf("changed model: %+v", r)
	}
	changedSuite := append(append([]byte(nil), suite...), '\n')
	r, e = Check(manifestRaw, model, changedSuite, "replay")
	if e != nil {
		t.Fatal(e)
	}
	cs := expected.ChangedSuite
	if r.SuiteBinding != cs.SuiteBinding || r.SuiteValidation != cs.SuiteValidation || r.Replay != cs.Replay || r.ScenarioReport != nil {
		t.Fatalf("changed suite: %+v", r)
	}
	// Deliberately authored wrong literal with a newly asserted source identity;
	// no runner output is ever used to update the expected data.
	wrong := bytes.Replace(suite, []byte(`"state":"Ready","revision":"1"`), []byte(`"state":"Ready","revision":"2"`), 1)
	if bytes.Equal(suite, wrong) {
		t.Fatal("wrong expectation edit failed")
	}
	var m manifest
	if e = json.Unmarshal(manifestRaw, &m); e != nil {
		t.Fatal(e)
	}
	m.SuiteSourceSHA256 = hash(wrong)
	wrongManifest, _ := json.Marshal(m)
	r, e = Check(wrongManifest, model, wrong, "replay")
	if e != nil {
		t.Fatal(e)
	}
	cases, steps = counts(t, r)
	we := expected.WrongExpectation
	if r.Replay != we.Replay || cases != we.CasesCompleted || steps != we.StepsExecuted {
		t.Fatalf("wrong expected: %+v", r)
	}
}
func TestStrictAndCaseBinding(t *testing.T) {
	m, model, suite, _ := fixture(t)
	invalid := []struct {
		name                   string
		manifest, model, suite []byte
		mode                   string
	}{
		{"duplicate key", bytes.Replace(m, []byte(`"id": "independent_example"`), []byte(`"id": "independent_example", "id": "other"`), 1), model, suite, "check"},
		{"unknown key", bytes.Replace(m, []byte(`"id": "independent_example"`), []byte(`"unexpected": true, "id": "independent_example"`), 1), model, suite, "check"},
		{"missing intent", bytes.Replace(m, []byte(`"intent": "Preparing a valid draft reaches Ready and increments revision."`), []byte(`"intent": " "`), 1), model, suite, "check"},
		{"missing case", bytes.Replace(m, []byte(`"id": "unknown_refused"`), []byte(`"id": "other_case"`), 1), model, suite, "check"},
		{"bad model", m, []byte(`{}`), suite, "check"},
		{"bad mode", m, model, suite, "invalid"},
		{"bad suite", m, model, []byte(`{`), "check"},
	}
	for _, c := range invalid {
		t.Run(c.name, func(t *testing.T) {
			r, e := Check(c.manifest, c.model, c.suite, c.mode)
			if !errors.Is(e, ErrInvalid) || r.FormatVersion != "" {
				t.Fatalf("result=%+v error=%v", r, e)
			}
		})
	}
}
func TestRequestDeterminism(t *testing.T) {
	m, model, suite, _ := fixture(t)
	q := Request{RequestVersion, "replay", string(m), string(model), string(suite)}
	raw, _ := json.Marshal(q)
	a, e := Execute(raw)
	if e != nil {
		t.Fatal(e)
	}
	b, e := Execute(raw)
	if e != nil || !bytes.Equal(a, b) {
		t.Fatal("nondeterministic output")
	}
	if _, e = Execute(append(raw, []byte(` {}`)...)); !errors.Is(e, ErrInvalid) {
		t.Fatal("accepted trailing data")
	}
	if _, e = Execute([]byte(strings.Repeat(" ", MaxRequestBytes+1))); !errors.Is(e, ErrInvalid) {
		t.Fatal("accepted oversized request")
	}
}

func TestSemanticAndCompiledBindings(t *testing.T) {
	m, source, suite, _ := fixture(t)
	formatted := append(append([]byte(nil), source...), '\n')
	r, e := Check(m, formatted, suite, "check")
	if e != nil || r.ModelBinding != "current" || r.SuiteValidation != "valid" {
		t.Fatalf("format-only model: %+v %v", r, e)
	}
	wrongHeader := bytes.Replace(suite, []byte(`948630cb7857199443b59b92d9b1ab0fadf2bf7a5bf8125f7a7e767db83958ed`), []byte(strings.Repeat("a", 64)), 1)
	var mm manifest
	if e = json.Unmarshal(m, &mm); e != nil {
		t.Fatal(e)
	}
	mm.SuiteSourceSHA256 = hash(wrongHeader)
	pinned, _ := json.Marshal(mm)
	r, e = Check(pinned, source, wrongHeader, "replay")
	if e != nil || r.SuiteModelBinding != "stale" || r.SuiteValidation != "not-checked" || r.Replay != "not-run" {
		t.Fatalf("suite model: %+v %v", r, e)
	}
	changed := bytes.Replace(source, []byte(`"label": "Draft"`), []byte(`"label": "New Draft"`), 1)
	ir, e := jsonformat.Decode(changed)
	if e != nil {
		t.Fatal(e)
	}
	digest, e := canonical.Digest(ir)
	if e != nil {
		t.Fatal(e)
	}
	futureSuite := bytes.Replace(suite, []byte(mm.ModelDigest), []byte(digest), 1)
	mm.ModelDigest = digest
	mm.SuiteSourceSHA256 = hash(futureSuite)
	futureManifest, _ := json.Marshal(mm)
	r, e = Check(futureManifest, changed, futureSuite, "replay")
	if e != nil || r.ModelBinding != "current" || r.SuiteModelBinding != "current" || r.RunnerBinding != "stale" || r.SuiteValidation != "not-checked" || r.Replay != "not-run" {
		t.Fatalf("unsupported runner: %+v %v", r, e)
	}
}

// Execute the documented recipe itself, with CRLF artifacts, so text-mode
// newline translation in a future edit cannot conceal a stale source identity.
func TestDocumentedRecipePreservesCRLF(t *testing.T) {
	python, err := exec.LookPath("python3")
	if err != nil {
		t.Skip("python3 is needed to exercise the documented Python recipe")
	}
	readme, err := os.ReadFile("README.md")
	if err != nil {
		t.Fatal(err)
	}
	_, after, ok := strings.Cut(string(readme), "python3 -c '")
	if !ok {
		t.Fatal("documented Python recipe missing")
	}
	script, _, ok := strings.Cut(after, "' | ")
	if !ok {
		t.Fatal("documented Python recipe terminator missing")
	}
	manifestBytes, modelBytes, suiteBytes, _ := fixture(t)
	crlf := func(b []byte) []byte { return bytes.ReplaceAll(b, []byte("\n"), []byte("\r\n")) }
	manifestBytes, modelBytes, suiteBytes = crlf(manifestBytes), crlf(modelBytes), crlf(suiteBytes)
	temp := t.TempDir()
	for _, entry := range []struct {
		path string
		data []byte
	}{
		{"acceptanceauthor/testdata/manifest.json", manifestBytes},
		{"examples/lifecycle.json", modelBytes},
		{"acceptanceauthor/testdata/suite.json", suiteBytes},
	} {
		path := filepath.Join(temp, entry.path)
		if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, entry.data, 0600); err != nil {
			t.Fatal(err)
		}
	}
	cmd := exec.Command(python, "-c", script)
	cmd.Dir = temp
	request, err := cmd.Output()
	if err != nil {
		t.Fatalf("documented recipe: %v", err)
	}
	var got Request
	if err := json.Unmarshal(request, &got); err != nil {
		t.Fatal(err)
	}
	for _, pair := range []struct {
		name, got string
		want      []byte
	}{
		{"manifest", got.Manifest, manifestBytes}, {"model", got.Model, modelBytes}, {"suite", got.Suite, suiteBytes},
	} {
		if !bytes.Equal([]byte(pair.got), pair.want) {
			t.Fatalf("%s bytes were changed by documented recipe", pair.name)
		}
	}
	output, err := Execute(request)
	if err != nil {
		t.Fatal(err)
	}
	var result Report
	if err := json.Unmarshal(output, &result); err != nil {
		t.Fatal(err)
	}
	if result.ManifestSHA256 != hash(manifestBytes) || result.ModelBinding != "current" || result.SuiteBinding != "stale" || result.SuiteValidation != "not-checked" || result.Replay != "not-run" {
		t.Fatalf("CRLF identities or freshness not preserved: %+v", result)
	}
}
