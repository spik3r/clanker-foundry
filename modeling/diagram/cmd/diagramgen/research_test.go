package main

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

var researchIdentity = regexp.MustCompile(`^\{"contract":"foundry-generator/v1","profile":"research-page/v1","sourceSHA256":"[0-9a-f]{64}","canonicalSHA256":"[0-9a-f]{64}"\}\n$`)

func researchModel(t *testing.T) string {
	t.Helper()
	p, err := filepath.Abs(filepath.Join("..", "..", "examples", "evidence-atlas-page.json"))
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func TestResearchPageContract(t *testing.T) {
	m := researchModel(t)
	code, vline, errs := do("validate", m)
	if code != 0 || !researchIdentity.MatchString(vline) || errs != "" {
		t.Fatalf("validate: code %d stdout %q stderr %q", code, vline, errs)
	}
	out := filepath.Join(t.TempDir(), "out")
	code, line, errs := do("generate", m, out)
	if code != 0 || line != vline || errs != "" {
		t.Fatalf("generate: code %d stdout %q stderr %q", code, line, errs)
	}
	entries, err := os.ReadDir(out)
	if err != nil || len(entries) != 4 {
		t.Fatalf("output entries %v %v", entries, err)
	}
	for _, f := range []string{"ResearchPage.jsx", "slots.json", "model.json", "foundry-manifest.json"} {
		if _, err := os.Stat(filepath.Join(out, f)); err != nil {
			t.Errorf("missing %s: %v", f, err)
		}
	}
	// The committed example output is what the CLI writes.
	committed := filepath.Join("..", "..", "examples", "generated", "evidence-atlas-page")
	if code, cline, errs := do("check", m, committed); code != 0 || cline != line || errs != "" {
		t.Fatalf("check against committed output: code %d stdout %q stderr %q", code, cline, errs)
	}
	if code, cline, errs := do("check", m, out); code != 0 || cline != line || errs != "" {
		t.Fatalf("check: code %d stdout %q stderr %q", code, cline, errs)
	}
	// Drift: edit the component, delete slots.json, add a file.
	if err := os.WriteFile(filepath.Join(out, "ResearchPage.jsx"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(out, "slots.json")); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(out, "notes.txt"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	code, stdout, stderr := do("check", m, out)
	got := strings.Split(strings.TrimSpace(stderr), "\n")
	sort.Strings(got)
	want := []string{"drift: missing slots.json", "drift: modified ResearchPage.jsx", "drift: unexpected notes.txt"}
	if code != 3 || stdout != "" || strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Errorf("code %d stdout %q stderr %q", code, stdout, stderr)
	}
}

func TestResearchPageRefusals(t *testing.T) {
	dir := t.TempDir()
	write := func(name, content string) string {
		p := filepath.Join(dir, name)
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
		return p
	}
	good, err := os.ReadFile(researchModel(t))
	if err != nil {
		t.Fatal(err)
	}
	cases := map[string]string{
		"duplicate key":    strings.Replace(string(good), `"id": "EvidenceAtlas",`, `"id": "EvidenceAtlas", "id": "Again",`, 1),
		"case variant":     strings.Replace(string(good), `"title": "Evidence Atlas"`, `"Title": "Evidence Atlas"`, 1),
		"trailing data":    string(good) + "{}",
		"null":             strings.Replace(string(good), `"Exploratory"`, `null`, 1),
		"unknown member":   strings.Replace(string(good), `"sections"`, `"script": "x", "sections"`, 1),
		"unknown status":   strings.Replace(string(good), `"status": "mixed"`, `"status": "done"`, 1),
		"duplicate slot":   strings.Replace(string(good), `"slot": "detail"`, `"slot": "matrix"`, 1),
		"duplicate tab":    strings.Replace(string(good), `"tab": "method"`, `"tab": "explore"`, 1),
		"invalid UTF-8":    strings.Replace(string(good), `Evidence Atlas`, "Evidence \xff Atlas", 1),
		"oversize":         string(good) + strings.Repeat(" ", 512<<10),
		"profile mismatch": strings.Replace(string(good), `research-page/v1`, `research-page/v9`, 1),
	}
	for name, content := range cases {
		t.Run(name, func(t *testing.T) {
			p := write(strings.ReplaceAll(name, " ", "_")+".json", content)
			for _, args := range [][]string{{"validate", p}, {"generate", p, filepath.Join(dir, "never")}, {"check", p, dir}} {
				code, stdout, stderr := do(args...)
				if code != 2 || stdout != "" || stderr == "" {
					t.Errorf("%v: code %d stdout %q stderr %q", args, code, stdout, stderr)
				}
			}
		})
	}
	if _, err := os.Lstat(filepath.Join(dir, "never")); err == nil {
		t.Errorf("generate created output for an invalid model")
	}
}

func TestResearchPageValidateWritesNothing(t *testing.T) {
	dir := t.TempDir()
	m := filepath.Join(dir, "m.json")
	b, err := os.ReadFile(researchModel(t))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(m, b, 0o644); err != nil {
		t.Fatal(err)
	}
	if code, _, _ := do("validate", m); code != 0 {
		t.Fatalf("validate failed with %d", code)
	}
	entries, _ := os.ReadDir(dir)
	if len(entries) != 1 {
		t.Errorf("validate wrote files: %v", entries)
	}
}
