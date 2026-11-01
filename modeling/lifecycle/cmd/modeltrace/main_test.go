package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// fixture copies the three inputs into a fresh module-shaped directory.
func fixture(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	for _, p := range []string{modelPath, catalogPath, requirementsPath} {
		b, err := os.ReadFile(filepath.Join("..", "..", p))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(filepath.Join(dir, filepath.Dir(p)), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, p), b, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func exec(t *testing.T, dir string, args ...string) (int, string, string) {
	t.Helper()
	var out, errb bytes.Buffer
	code := run(args, dir, &out, &errb)
	return code, out.String(), errb.String()
}

func TestValidateDistinguishesGapsFromInvalid(t *testing.T) {
	dir := fixture(t)
	code, out, _ := exec(t, dir, "validate")
	if code != 0 || !strings.Contains(out, "gap UR-012 no-acceptance-link") || !strings.Contains(out, "unlinked case special-02") {
		t.Fatalf("validate: %d %q", code, out)
	}
	if code, _, _ := exec(t, dir, "validate", "-complete"); code != 3 {
		t.Fatalf("-complete: %d", code)
	}
	p := filepath.Join(dir, requirementsPath)
	b, _ := os.ReadFile(p)
	os.WriteFile(p, bytes.Replace(b, []byte(`"G_revision"`), []byte(`"G_missing"`), 1), 0o644)
	code, out, errs := exec(t, dir, "validate")
	if code != 2 || out != "" || !strings.Contains(errs, "dangling-model-link UR-003 -> guard:G_missing") {
		t.Fatalf("invalid: %d %q %q", code, out, errs)
	}
}

func TestGenerateThenCheck(t *testing.T) {
	dir := fixture(t)
	if code, _, _ := exec(t, dir, "check"); code != 1 {
		t.Fatalf("check before generate: %d", code)
	}
	if code, _, e := exec(t, dir, "generate"); code != 0 {
		t.Fatalf("generate: %d %s", code, e)
	}
	first, _ := os.ReadFile(filepath.Join(dir, viewDir, "trace.md"))
	if code, _, _ := exec(t, dir, "check"); code != 0 {
		t.Fatalf("check after generate: %d", code)
	}
	if code, _, _ := exec(t, dir, "generate"); code != 0 {
		t.Fatal("second generate")
	}
	second, _ := os.ReadFile(filepath.Join(dir, viewDir, "trace.md"))
	if !bytes.Equal(first, second) {
		t.Fatal("regeneration changed bytes")
	}
	os.WriteFile(filepath.Join(dir, viewDir, "trace.svg"), []byte("edited"), 0o644)
	code, _, errs := exec(t, dir, "check")
	if code != 1 || !strings.Contains(errs, "trace.svg") || strings.Contains(errs, "trace.md") {
		t.Fatalf("stale: %d %q", code, errs)
	}
	entries, _ := os.ReadDir(filepath.Join(dir, viewDir))
	if len(entries) != 3 {
		t.Fatalf("unexpected files: %v", entries)
	}
}

func TestCheckDoesNotWrite(t *testing.T) {
	dir := fixture(t)
	exec(t, dir, "check")
	if _, err := os.Stat(filepath.Join(dir, viewDir)); !os.IsNotExist(err) {
		t.Fatal("check created the view directory")
	}
}

func TestRefusesSymlinks(t *testing.T) {
	dir := fixture(t)
	outside := t.TempDir()
	if err := os.Symlink(outside, filepath.Join(dir, viewDir)); err != nil {
		t.Skip("symlinks unavailable")
	}
	if code, _, _ := exec(t, dir, "generate"); code != 2 {
		t.Fatalf("symlinked view dir: %d", code)
	}
	if entries, _ := os.ReadDir(outside); len(entries) != 0 {
		t.Fatal("wrote through symlink")
	}
	os.Remove(filepath.Join(dir, viewDir))
	os.MkdirAll(filepath.Join(dir, viewDir), 0o755)
	os.Symlink(filepath.Join(outside, "x"), filepath.Join(dir, viewDir, "trace.md"))
	if code, _, _ := exec(t, dir, "generate"); code != 2 {
		t.Fatalf("symlinked output: %d", code)
	}
	if entries, _ := os.ReadDir(outside); len(entries) != 0 {
		t.Fatal("wrote through output symlink")
	}
	req := filepath.Join(dir, requirementsPath)
	os.Rename(req, req+".real")
	os.Symlink(req+".real", req)
	if code, _, _ := exec(t, dir, "validate"); code != 2 {
		t.Fatalf("symlinked input: %d", code)
	}
}

func TestUsage(t *testing.T) {
	dir := fixture(t)
	for _, args := range [][]string{{}, {"nope"}, {"generate", "x"}, {"check", "-complete"}, {"validate", "-x"}, {"validate", "-complete", "x"}} {
		if code, out, _ := exec(t, dir, args...); code != 2 || out != "" {
			t.Errorf("%v: %d %q", args, code, out)
		}
	}
}
