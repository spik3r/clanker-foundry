package main

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

func umlIdentity(profile string) *regexp.Regexp {
	return regexp.MustCompile(`^\{"contract":"foundry-generator/v1","profile":"` + regexp.QuoteMeta(profile) + `","sourceSHA256":"[0-9a-f]{64}","canonicalSHA256":"[0-9a-f]{64}"\}\n$`)
}

func exampleModel(t *testing.T, name string) string {
	t.Helper()
	p, err := filepath.Abs(filepath.Join("..", "..", "examples", name))
	if err != nil {
		t.Fatal(err)
	}
	return p
}

// TestUMLProfileContract runs validate, generate and check on one model of
// each UML-style profile, then checks drift and refusals.
func TestUMLProfileContract(t *testing.T) {
	for _, c := range []struct{ file, profile string }{
		{"checkout-sequence.json", "sequence-diagram/v1"},
		{"orders-class.json", "class-diagram/v1"},
	} {
		t.Run(c.file, func(t *testing.T) {
			m := exampleModel(t, c.file)
			code, vline, errs := do("validate", m)
			if code != 0 || !umlIdentity(c.profile).MatchString(vline) || errs != "" {
				t.Fatalf("validate: code %d stdout %q stderr %q", code, vline, errs)
			}
			out := filepath.Join(t.TempDir(), "out")
			code, line, errs := do("generate", m, out)
			if code != 0 || line != vline || errs != "" {
				t.Fatalf("generate: code %d stdout %q stderr %q", code, line, errs)
			}
			entries, err := os.ReadDir(out)
			if err != nil || len(entries) != 3 {
				t.Fatalf("output entries %v %v", entries, err)
			}
			for _, f := range []string{"diagram.svg", "model.json", "foundry-manifest.json"} {
				if _, err := os.Stat(filepath.Join(out, f)); err != nil {
					t.Errorf("missing %s: %v", f, err)
				}
			}
			if code, cline, errs := do("check", m, out); code != 0 || cline != line || errs != "" {
				t.Fatalf("check: code %d stdout %q stderr %q", code, cline, errs)
			}
			if err := os.WriteFile(filepath.Join(out, "diagram.svg"), []byte("x"), 0o644); err != nil {
				t.Fatal(err)
			}
			if err := os.Remove(filepath.Join(out, "model.json")); err != nil {
				t.Fatal(err)
			}
			code, stdout, stderr := do("check", m, out)
			got := strings.Split(strings.TrimSpace(stderr), "\n")
			sort.Strings(got)
			want := []string{"drift: missing model.json", "drift: modified diagram.svg"}
			if code != 3 || stdout != "" || strings.Join(got, "\n") != strings.Join(want, "\n") {
				t.Errorf("code %d stdout %q stderr %q", code, stdout, stderr)
			}
			// A second generate onto the same path is refused.
			if code, _, _ := do("generate", m, out); code != 2 {
				t.Errorf("generate onto existing output: exit %d", code)
			}
		})
	}
}

func TestUMLProfileRefusals(t *testing.T) {
	dir := t.TempDir()
	write := func(name, content string) string {
		p := filepath.Join(dir, name)
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
		return p
	}
	cases := map[string]string{
		"sequence without steps":  `{"formatVersion":"diagram-json/v1","profileVersion":"sequence-diagram/v1","id":"A","title":"T","participants":[{"id":"A","label":"a","kind":"actor"},{"id":"B","label":"b","kind":"actor"}],"steps":[]}`,
		"sequence unknown member": `{"formatVersion":"diagram-json/v1","profileVersion":"sequence-diagram/v1","id":"A","title":"T","colour":"red","participants":[],"steps":[]}`,
		"class self relation":     `{"formatVersion":"diagram-json/v1","profileVersion":"class-diagram/v1","id":"A","title":"T","classes":[{"id":"C","name":"C","kind":"class"}],"relations":[{"id":"R","kind":"association","from":"C","to":"C"}]}`,
		"class duplicate key":     `{"formatVersion":"diagram-json/v1","profileVersion":"class-diagram/v1","id":"A","id":"B","title":"T","classes":[]}`,
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
