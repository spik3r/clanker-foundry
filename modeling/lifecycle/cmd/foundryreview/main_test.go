package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/acceptanceauthor"
)

func TestEscapeAndEvidenceStatus(t *testing.T) {
	h := strings.Repeat("a", 64)
	p := page{evidence: evidence{
		FormatVersion: "foundry-review-evidence/v1", Title: "<script>alert(1)</script>", Baseline: revision{"base", h}, Candidate: revision{"head", h},
		Requirements: []requirement{{"UR-101", "No linked case", "trace.json", "unlinked"}},
		Scenarios:    []scenario{{ID: "case", Title: "Case", Given: "Draft", When: "Prepare", Then: "Refuse", Expected: "refused", Status: "stale", Source: "suite.json"}},
	}}
	x, e := render(p)
	if e != nil {
		t.Fatal(e)
	}
	y, e := render(p)
	if e != nil || !bytes.Equal(x, y) {
		t.Fatal("render is not deterministic", e)
	}
	if bytes.Contains(x, []byte("<script>")) || !bytes.Contains(x, []byte("&lt;script&gt;")) || !bytes.Contains(x, []byte("UR-101 · unlinked")) {
		t.Fatal("review boundary lost")
	}
	p.Scenarios[0].Observed = "forged outcome"
	if _, e := render(p); e == nil {
		t.Fatal("stale evidence accepted with outcome")
	}
	p.Scenarios[0].Observed = ""
	p.Scenarios[0].RequirementIDs = []string{"UR-101"}
	if _, e := render(p); e == nil {
		t.Fatal("undeclared linkage accepted")
	}
}

func TestNativeReceiptBinding(t *testing.T) {
	t.Chdir("../..")
	before, err := loadDemo(1)
	if err != nil {
		t.Fatal(err)
	}
	report, err := acceptanceauthor.Check([]byte(before.Manifest), []byte(before.Model), []byte(before.Suite), "replay")
	if err != nil || report.Replay != "matched" || report.ScenarioDigest == nil {
		t.Fatalf("baseline replay: %v, %v", report.Replay, err)
	}
	original := *report.ScenarioDigest
	var suite literal
	if err := json.Unmarshal([]byte(before.Suite), &suite); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "replay.json")
	write := func() {
		t.Helper()
		if err := os.WriteFile(path, []byte(toJSON(report)), 0600); err != nil {
			t.Fatal(err)
		}
	}
	write()
	got, err := observations(before, suite, original, path)
	if err != nil || len(got) != 3 {
		t.Fatalf("valid replay: %v, %v", got, err)
	}
	if _, err := observations(before, suite, strings.Repeat("0", 64), path); err == nil {
		t.Fatal("receipt accepted with digest unrelated to native suite diff")
	}
	after, err := loadDemo(2)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal([]byte(after.Suite), &suite); err != nil {
		t.Fatal(err)
	}
	got, err = observations(after, suite, original, path)
	if err != nil || len(got) != 3 {
		t.Fatalf("mixed revision: %v, %v", got, err)
	}
	for _, row := range got {
		if row.status != "stale" || row.body != "" {
			t.Fatalf("mixed revision exposed an observation: %+v", row)
		}
	}
}

func TestTutorialRenders(t *testing.T) {
	t.Chdir("../..")
	html, err := tutorial(nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"UR-101 · unlinked", "G_actor", "not-run", "Baseline process projection", "Candidate process projection"} {
		if !bytes.Contains(html, []byte(want)) {
			t.Errorf("tutorial omitted %q", want)
		}
	}
}
