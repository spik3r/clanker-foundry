package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const sample = "../../examples/browser.json"

func do(args ...string) (int, string, string) {
	var out, errb bytes.Buffer
	code := run(args, &out, &errb)
	return code, out.String(), errb.String()
}

func TestValidateAndGenerateAndCheck(t *testing.T) {
	out := filepath.Join(t.TempDir(), "out")
	code, vline, _ := do("validate", sample)
	if code != 0 || !strings.HasPrefix(vline, `{"contract":"foundry-generator/v1","profile":"record-browser/v1","sourceSHA256":"`) {
		t.Fatalf("validate: %d %q", code, vline)
	}
	if code, line, _ := do("generate", sample, out); code != 0 || line != vline {
		t.Fatalf("generate: %d %q", code, line)
	}
	for _, p := range []string{"foundry-manifest.json", "cmd/demo/main.go", "records.go"} {
		if _, err := os.Stat(filepath.Join(out, p)); err != nil {
			t.Fatal(err)
		}
	}
	if code, line, e := do("check", sample, out); code != 0 || line != vline || e != "" {
		t.Fatalf("check: %d %q %q", code, line, e)
	}
}

func TestCheckReportsDriftAndNeverWrites(t *testing.T) {
	out := filepath.Join(t.TempDir(), "out")
	if code, _, _ := do("generate", sample, out); code != 0 {
		t.Fatal("generate failed")
	}
	if err := os.WriteFile(filepath.Join(out, "records.go"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(out, "go.mod")); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(out, "extra.txt"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(out, "cmd", "extra"), 0o755); err != nil {
		t.Fatal(err)
	}
	code, stdout, stderr := do("check", sample, out)
	if code != 3 || stdout != "" {
		t.Fatalf("check: %d %q", code, stdout)
	}
	for _, want := range []string{"drift: modified records.go", "drift: missing go.mod", "drift: unexpected extra.txt", "drift: unexpected cmd/extra"} {
		if !strings.Contains(stderr, want+"\n") {
			t.Errorf("missing %q in %q", want, stderr)
		}
	}
	if _, err := os.Stat(filepath.Join(out, "go.mod")); !os.IsNotExist(err) {
		t.Fatal("check repaired output")
	}
	if code, _, e := do("check", sample, filepath.Join(t.TempDir(), "none")); code != 3 || e != "drift: missing .\n" {
		t.Fatalf("absent: %d %q", code, e)
	}
}

func TestGenerateRefusesExistingOutAndMissingParent(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "dir")
	file := filepath.Join(root, "file")
	link := filepath.Join(root, "link")
	if err := os.Mkdir(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(file, []byte("keep"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(dir, link); err != nil {
		t.Skip(err)
	}
	for _, out := range []string{dir, file, link} {
		if code, stdout, stderr := do("generate", sample, out); code != 2 || stdout != "" || stderr == "" {
			t.Fatalf("%s: %d %q %q", out, code, stdout, stderr)
		}
	}
	if entries, _ := os.ReadDir(dir); len(entries) != 0 {
		t.Fatal("changed existing directory")
	}
	if b, _ := os.ReadFile(file); string(b) != "keep" {
		t.Fatal("changed existing file")
	}
	missing := filepath.Join(root, "no", "out")
	if code, _, _ := do("generate", sample, missing); code != 2 {
		t.Fatalf("missing parent: %d", code)
	}
	if _, err := os.Lstat(filepath.Join(root, "no")); !os.IsNotExist(err) {
		t.Fatal("created missing parent")
	}
}

func TestInvalidAndMissingModelRefusedBeforeOutput(t *testing.T) {
	root := t.TempDir()
	bad := filepath.Join(root, "invalid.json")
	if err := os.WriteFile(bad, []byte(`{"formatVersion":"wrong"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(root, "out")
	absent := filepath.Join(root, "absent.json")
	for _, args := range [][]string{{"validate", bad}, {"generate", bad, out}, {"check", bad, out},
		{"validate", absent}, {"generate", absent, out}} {
		if code, stdout, stderr := do(args...); code != 2 || stdout != "" || stderr == "" {
			t.Fatalf("%v: %d %q %q", args, code, stdout, stderr)
		}
	}
	if _, err := os.Lstat(out); !os.IsNotExist(err) {
		t.Fatal("created output for refused model")
	}
}

func TestUsageErrors(t *testing.T) {
	out := filepath.Join(t.TempDir(), "out")
	for _, args := range [][]string{nil, {"--help"}, {"validate"}, {"validate", sample, "x"}, {"generate", sample},
		{"check", sample}, {"bogus", sample}, {sample, out}} {
		if code, stdout, stderr := do(args...); code != 2 || stdout != "" || stderr == "" {
			t.Fatalf("%v: %d %q %q", args, code, stdout, stderr)
		}
	}
	if _, err := os.Lstat(out); !os.IsNotExist(err) {
		t.Fatal("old positional form wrote output")
	}
}
