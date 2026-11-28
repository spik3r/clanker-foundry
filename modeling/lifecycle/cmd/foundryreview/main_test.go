package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
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
	html, err := tutorial(nil, appearance{})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"UR-101 · unlinked", "G_actor", "not-run", "Baseline process projection", "Candidate process projection"} {
		if !bytes.Contains(html, []byte(want)) {
			t.Errorf("tutorial omitted %q", want)
		}
	}
}

func TestRejectsDuplicateRequirementIDs(t *testing.T) {
	h := strings.Repeat("a", 64)
	for _, tc := range []struct{ first, second string }{
		{"unlinked", "linked"},
		{"unlinked", "unlinked"},
		{"linked", "unlinked"},
		{"linked", "linked"},
	} {
		t.Run(tc.first+"-"+tc.second, func(t *testing.T) {
			p := page{evidence: evidence{
				FormatVersion: "foundry-review-evidence/v1", Title: "Duplicate requirements",
				Baseline: revision{"base", h}, Candidate: revision{"head", h},
				Requirements: []requirement{
					{"UR-101", "First", "trace.json", tc.first},
					{"UR-101", "Second", "trace.json", tc.second},
				},
			}}
			if tc.second == "linked" {
				p.Scenarios = []scenario{{
					ID: "case", Title: "Case", Given: "Draft", When: "Prepare", Then: "Refuse",
					Expected: "refused", Status: "not-run", Source: "suite.json",
					RequirementIDs: []string{"UR-101"},
				}}
			}
			if _, err := render(p); err == nil || err.Error() != `invalid requirement "UR-101"` {
				t.Fatalf("expected duplicate requirement rejection, got %v", err)
			}
		})
	}
}

func TestReportEmbedsPinnedTheme(t *testing.T) {
	const expected = "2265fdbee433afa1a46f4bb205ee11bf16780083e58fdf2a15a02de96f458e91"
	if got := fmt.Sprintf("%x", sha256.Sum256([]byte(tesseraeTokens))); got != expected {
		t.Fatalf("Tesserae token pin differs: %s", got)
	}
	// Use the renderer directly: the tutorial CLI runs from the module root.
	h := strings.Repeat("a", 64)
	b, err := render(page{evidence: evidence{
		FormatVersion: "foundry-review-evidence/v1", Title: "Theme check",
		Baseline: revision{"base", h}, Candidate: revision{"head", h},
	}})
	if err != nil {
		t.Fatal(err)
	}
	for _, fragment := range []string{`data-theme="slate" data-mode="auto"`, tesseraeTokens, "/*\n" + tesseraeLicense + "*/", "Copyright (c) 2026 Kai", "Permission is hereby granted", "color:var(--ts-text)", "default-src 'none'"} {
		if !bytes.Contains(b, []byte(fragment)) {
			t.Errorf("theme fragment missing: %q", fragment)
		}
	}
	if bytes.Contains(b, []byte("<script")) || bytes.Contains(b, []byte("<link")) {
		t.Fatal("report now contains a script or external stylesheet element")
	}
}

func TestAppearanceFlags(t *testing.T) {
	t.Chdir("../..")
	h := strings.Repeat("a", 64)
	packet := evidence{FormatVersion: "foundry-review-evidence/v1", Title: "Theme check",
		Baseline: revision{"base", h}, Candidate: revision{"head", h},
		Changes: []change{}, Requirements: []requirement{}, Scenarios: []scenario{}}
	path := filepath.Join(t.TempDir(), "evidence.json")
	b, err := json.Marshal(packet)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, b, 0600); err != nil {
		t.Fatal(err)
	}
	for _, input := range []string{"--tutorial", path} {
		for _, choice := range []appearance{{"slate", "auto"}, {"sand", "light"}, {"slate", "dark"}} {
			args := []string{"--theme", choice.Theme, "--mode", choice.Mode, input}
			html, err := run(args)
			want := fmt.Sprintf(`data-theme="%s" data-mode="%s"`, choice.Theme, choice.Mode)
			if err != nil || !bytes.Contains(html, []byte(want)) {
				t.Errorf("appearance %v via %q: expected %q, err=%v", choice, input, want, err)
			}
		}
	}
	for _, args := range [][]string{{"--theme"}, {"--theme", "example", path}, {"--mode", "x", path},
		{"--mode", "dark", "--mode", "light", path}, {"--theme", "sand", "--theme", "slate", path}} {
		if html, err := run(args); err == nil || len(html) != 0 {
			t.Errorf("invalid appearance %q: expected refusal with no report", args)
		}
	}
}
