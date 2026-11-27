package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestArchitectureDiff(t *testing.T) {
	dir := t.TempDir()
	before := filepath.Join(dir, "before.json")
	after := filepath.Join(dir, "after.json")
	base := `{"formatVersion":"diagram-json/v1","profileVersion":"architecture-diagram/v1","id":"Demo","title":"System","nodes":[{"id":"API","kind":"service","label":"API"}],"links":[]}`
	changed := strings.Replace(base, `"title":"System"`, `"title":"New title"`, 1)
	for path, raw := range map[string]string{before: base, after: changed} {
		if err := os.WriteFile(path, []byte(raw), 0600); err != nil {
			t.Fatal(err)
		}
	}
	const want = `{"changes":[{"kind":"title-changed","id":"Demo","field":"title","before":"System","after":"New title"}],"nodeOrderChanged":false}` + "\n"
	if code, out, errs := do("diff", before, after); code != 0 || out != want || errs != "" {
		t.Fatalf("diff: code %d stdout %q stderr %q", code, out, errs)
	}
	if code, out, errs := do("diff", before, before); code != 0 || out != "{\"changes\":[],\"nodeOrderChanged\":false}\n" || errs != "" {
		t.Fatalf("no-op: code %d stdout %q stderr %q", code, out, errs)
	}
	if code, out, errs := do("diff", before, model(t)); code != exitRefused || out != "" || !strings.Contains(errs, "after:") {
		t.Fatalf("bad profile: code %d stdout %q stderr %q", code, out, errs)
	}
	for path, raw := range map[string]string{before: base, after: changed} {
		data, err := os.ReadFile(path)
		if err != nil || !bytes.Equal(data, []byte(raw)) {
			t.Fatalf("diff changed input %s: %v", path, err)
		}
	}
}
