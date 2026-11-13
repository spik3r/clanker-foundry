package main

import (
	"bytes"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

var identityLine = regexp.MustCompile(`^\{"contract":"foundry-generator/v1","profile":"process-diagram/v1","sourceSHA256":"[0-9a-f]{64}","canonicalSHA256":"[0-9a-f]{64}"\}\n$`)

func model(t *testing.T) string {
	t.Helper()
	p, err := filepath.Abs(filepath.Join("..", "..", "examples", "order.json"))
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func do(args ...string) (code int, stdout, stderr string) {
	var o, e bytes.Buffer
	code = run(args, &o, &e)
	return code, o.String(), e.String()
}

func TestValidate(t *testing.T) {
	code, out, errs := do("validate", model(t))
	if code != 0 || !identityLine.MatchString(out) || errs != "" {
		t.Fatalf("code %d stdout %q stderr %q", code, out, errs)
	}
}

func TestGenerateThenCheck(t *testing.T) {
	out := filepath.Join(t.TempDir(), "out")
	code, line, errs := do("generate", model(t), out)
	if code != 0 || !identityLine.MatchString(line) || errs != "" {
		t.Fatalf("generate: code %d stdout %q stderr %q", code, line, errs)
	}
	for _, f := range []string{"diagram.svg", "diagram.bpmn", "model.json", "foundry-manifest.json"} {
		if _, err := os.Stat(filepath.Join(out, f)); err != nil {
			t.Errorf("missing %s: %v", f, err)
		}
	}
	code, cline, errs := do("check", model(t), out)
	if code != 0 || cline != line || errs != "" {
		t.Fatalf("check: code %d stdout %q stderr %q", code, cline, errs)
	}
	_, vline, _ := do("validate", model(t))
	if vline != line {
		t.Errorf("validate and generate identity lines differ")
	}
}

func TestCheckReportsDrift(t *testing.T) {
	out := filepath.Join(t.TempDir(), "out")
	if code, _, e := do("generate", model(t), out); code != 0 {
		t.Fatal(e)
	}
	if err := os.WriteFile(filepath.Join(out, "diagram.svg"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(out, "diagram.bpmn")); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(out, "extra", "deeper"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(out, "notes.txt"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("model.json", filepath.Join(out, "link")); err != nil {
		t.Fatal(err)
	}
	code, stdout, stderr := do("check", model(t), out)
	if code != 3 || stdout != "" {
		t.Fatalf("code %d stdout %q", code, stdout)
	}
	got := strings.Split(strings.TrimSpace(stderr), "\n")
	want := []string{
		"drift: missing diagram.bpmn",
		"drift: modified diagram.svg",
		"drift: unexpected extra",
		"drift: unexpected link",
		"drift: unexpected notes.txt",
	}
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Errorf("drift lines:\n%s\nwant:\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
	}
	// check never repairs.
	if b, _ := os.ReadFile(filepath.Join(out, "diagram.svg")); string(b) != "x" {
		t.Errorf("check changed the output")
	}
	if _, err := os.Stat(filepath.Join(out, "diagram.bpmn")); err == nil {
		t.Errorf("check restored a deleted file")
	}
	// Tampering with the manifest is drift as well.
	out2 := filepath.Join(t.TempDir(), "out")
	if code, _, e := do("generate", model(t), out2); code != 0 {
		t.Fatal(e)
	}
	if err := os.WriteFile(filepath.Join(out2, "foundry-manifest.json"), []byte("{}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if code, _, e := do("check", model(t), out2); code != 3 || !strings.Contains(e, "drift: modified foundry-manifest.json") {
		t.Errorf("manifest drift: code %d stderr %q", code, e)
	}
}

func TestCheckMissingOut(t *testing.T) {
	code, stdout, stderr := do("check", model(t), filepath.Join(t.TempDir(), "nope"))
	if code != 3 || stdout != "" || strings.TrimSpace(stderr) != "drift: missing ." {
		t.Errorf("code %d stdout %q stderr %q", code, stdout, stderr)
	}
}

func TestRefusals(t *testing.T) {
	dir := t.TempDir()
	write := func(name, content string) string {
		p := filepath.Join(dir, name)
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
		return p
	}
	existing := filepath.Join(dir, "exists")
	if err := os.Mkdir(existing, 0o755); err != nil {
		t.Fatal(err)
	}
	file := write("afile", "keep")
	link := filepath.Join(dir, "link")
	if err := os.Symlink(existing, link); err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name string
		args []string
	}{
		{"no arguments", nil},
		{"unknown verb", []string{"build", model(t)}},
		{"help", []string{"--help"}},
		{"validate without model", []string{"validate"}},
		{"validate extra argument", []string{"validate", model(t), "x"}},
		{"generate without out", []string{"generate", model(t)}},
		{"check extra argument", []string{"check", model(t), "a", "b"}},
		{"missing model", []string{"validate", filepath.Join(dir, "none.json")}},
		{"model is a directory", []string{"validate", dir}},
		{"empty model", []string{"validate", write("empty.json", "")}},
		{"truncated model", []string{"validate", write("trunc.json", `{"formatVersion":"diagram-json/v1"`)}},
		{"empty object", []string{"validate", write("obj.json", `{}`)}},
		{"array", []string{"validate", write("arr.json", `[]`)}},
		{"unsupported profile", []string{"validate", write("prof.json", `{"profileVersion":"class-diagram/v9"}`)}},
		{"generate invalid model", []string{"generate", write("bad.json", `{}`), filepath.Join(dir, "bad-out")}},
		{"check invalid model", []string{"check", write("bad2.json", `[]`), existing}},
		{"generate onto existing directory", []string{"generate", model(t), existing}},
		{"generate onto file", []string{"generate", model(t), file}},
		{"generate onto symlink", []string{"generate", model(t), link}},
		{"generate with missing parent", []string{"generate", model(t), filepath.Join(dir, "no", "parent")}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			code, stdout, stderr := do(c.args...)
			if code != 2 {
				t.Errorf("exit %d, want 2 (stderr %q)", code, stderr)
			}
			if stdout != "" || stderr == "" {
				t.Errorf("stdout %q stderr %q", stdout, stderr)
			}
		})
	}
	if b, _ := os.ReadFile(file); string(b) != "keep" {
		t.Errorf("generate overwrote an existing file")
	}
	if entries, _ := os.ReadDir(existing); len(entries) != 0 {
		t.Errorf("generate wrote into an existing directory")
	}
	for _, p := range []string{"bad-out", "no"} {
		if _, err := os.Lstat(filepath.Join(dir, p)); err == nil {
			t.Errorf("%s was created", p)
		}
	}
}

func TestValidateWritesNothing(t *testing.T) {
	dir := t.TempDir()
	m := filepath.Join(dir, "m.json")
	b, err := os.ReadFile(model(t))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(m, b, 0o644); err != nil {
		t.Fatal(err)
	}
	if code, _, e := do("validate", m); code != 0 {
		t.Fatal(e)
	}
	entries, _ := os.ReadDir(dir)
	if len(entries) != 1 {
		t.Errorf("validate left %d entries", len(entries))
	}
}

var dashboardIdentity = regexp.MustCompile(`^\{"contract":"foundry-generator/v1","profile":"dashboard/v1","sourceSHA256":"[0-9a-f]{64}","canonicalSHA256":"[0-9a-f]{64}"\}\n$`)

func dashboardModel(t *testing.T, name string) string {
	t.Helper()
	p, err := filepath.Abs(filepath.Join("..", "..", "examples", name))
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func TestDashboardContract(t *testing.T) {
	for _, name := range []string{"ops-dashboard.json", "sales-dashboard.json", "minimal-dashboard.json"} {
		t.Run(name, func(t *testing.T) {
			m := dashboardModel(t, name)
			code, vline, errs := do("validate", m)
			if code != 0 || !dashboardIdentity.MatchString(vline) || errs != "" {
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
			for _, f := range []string{"dashboard.html", "model.json", "foundry-manifest.json"} {
				if _, err := os.Stat(filepath.Join(out, f)); err != nil {
					t.Errorf("missing %s: %v", f, err)
				}
			}
			if code, cline, errs := do("check", m, out); code != 0 || cline != line || errs != "" {
				t.Fatalf("check: code %d stdout %q stderr %q", code, cline, errs)
			}
			// Drift: edit the HTML, delete model.json, add a file.
			if err := os.WriteFile(filepath.Join(out, "dashboard.html"), []byte("x"), 0o644); err != nil {
				t.Fatal(err)
			}
			if err := os.Remove(filepath.Join(out, "model.json")); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(out, "notes.txt"), nil, 0o644); err != nil {
				t.Fatal(err)
			}
			code, stdout, stderr := do("check", m, out)
			want := "drift: modified dashboard.html\ndrift: missing model.json\ndrift: unexpected notes.txt"
			got := strings.Split(strings.TrimSpace(stderr), "\n")
			sort.Strings(got)
			wantLines := strings.Split(want, "\n")
			sort.Strings(wantLines)
			if code != 3 || stdout != "" || strings.Join(got, "\n") != strings.Join(wantLines, "\n") {
				t.Errorf("code %d stdout %q stderr %q", code, stdout, stderr)
			}
		})
	}
}

func TestDashboardRefusals(t *testing.T) {
	dir := t.TempDir()
	write := func(name, content string) string {
		p := filepath.Join(dir, name)
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
		return p
	}
	good, err := os.ReadFile(dashboardModel(t, "minimal-dashboard.json"))
	if err != nil {
		t.Fatal(err)
	}
	head := `{"formatVersion":"diagram-json/v1","profileVersion":"dashboard/v1","id":"A","title":"T","widgets":[{"id":"W","kind":"stat","label":"l","value":"v"}],"layout":[["W"]]`
	cases := map[string]string{
		"duplicate key":    strings.Replace(string(good), `"id": "Minimal",`, `"id": "Minimal", "id": "Again",`, 1),
		"case variant":     strings.Replace(string(good), `"title"`, `"Title"`, 1),
		"trailing data":    string(good) + "{}",
		"null":             strings.Replace(string(good), `"All systems normal"`, `null`, 1),
		"unknown member":   strings.Replace(string(good), `"layout"`, `"script": "x", "layout"`, 1),
		"exponent":         head + `,"datasets":[{"id":"D","columns":[{"id":"N","label":"N","type":"number"}],"rows":[[1e3]]}]}`,
		"bad reference":    strings.Replace(string(good), `["Status"]`, `["Nope"]`, 1),
		"invalid UTF-8":    strings.Replace(string(good), `Minimal dashboard`, "Minimal \xff dashboard", 1),
		"oversize":         string(good) + strings.Repeat(" ", 512<<10),
		"profile mismatch": strings.Replace(string(good), `dashboard/v1`, `dashboard/v9`, 1),
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
