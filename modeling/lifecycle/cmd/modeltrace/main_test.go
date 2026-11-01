package main

import (
	"bytes"
	"fmt"
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

// A planted temporary-name symlink must not redirect generate's writes onto an
// input or an unrelated file inside the root.
func TestPlantedTemporarySymlinksAreNotFollowed(t *testing.T) {
	dir := fixture(t)
	model := filepath.Join(dir, modelPath)
	sentinel := filepath.Join(dir, "sentinel.txt")
	os.WriteFile(sentinel, []byte("unrelated"), 0o644)
	modelBefore, _ := os.ReadFile(model)
	os.MkdirAll(filepath.Join(dir, viewDir), 0o755)
	os.Symlink("../../"+modelPath, filepath.Join(dir, viewDir, "trace.json.tmp"))
	os.Symlink("../../sentinel.txt", filepath.Join(dir, viewDir, "trace.md.tmp"))
	if code, _, e := exec(t, dir, "generate"); code != 0 {
		t.Fatalf("generate: %d %s", code, e)
	}
	if b, _ := os.ReadFile(model); !bytes.Equal(b, modelBefore) {
		t.Fatal("model input overwritten")
	}
	if b, _ := os.ReadFile(sentinel); string(b) != "unrelated" {
		t.Fatal("unrelated file overwritten")
	}
	for _, name := range outputs {
		s, err := os.Lstat(filepath.Join(dir, viewDir, name))
		if err != nil || !s.Mode().IsRegular() {
			t.Fatalf("%s is not a regular output", name)
		}
	}
	for _, name := range []string{"trace.json.tmp", "trace.md.tmp"} {
		if s, err := os.Lstat(filepath.Join(dir, viewDir, name)); err != nil || s.Mode()&os.ModeSymlink == 0 {
			t.Fatalf("pre-existing %s was replaced or removed", name)
		}
	}
	if code, _, _ := exec(t, dir, "check"); code != 0 {
		t.Fatalf("check after generate: %d", code)
	}
}

// A failed write leaves no temporary files and does not touch final outputs.
func TestFailedGenerateLeavesNoTemporaries(t *testing.T) {
	dir := fixture(t)
	if code, _, _ := exec(t, dir, "generate"); code != 0 {
		t.Fatal("generate")
	}
	before, _ := os.ReadFile(filepath.Join(dir, viewDir, "trace.json"))
	os.Remove(filepath.Join(dir, viewDir, "trace.svg"))
	os.Mkdir(filepath.Join(dir, viewDir, "trace.svg"), 0o755) // non-regular output
	if code, _, _ := exec(t, dir, "generate"); code != 2 {
		t.Fatalf("generate over a directory output: %d", code)
	}
	entries, _ := os.ReadDir(filepath.Join(dir, viewDir))
	if len(entries) != 3 {
		t.Fatalf("leftover files: %v", entries)
	}
	if after, _ := os.ReadFile(filepath.Join(dir, viewDir, "trace.json")); !bytes.Equal(before, after) {
		t.Fatal("refused generate changed an output")
	}
}

// Inputs that render a view over the output limit are refused by every
// subcommand before anything is written, so generate never produces views that
// check then rejects.
func TestOversizeOutputIsRefusedBeforeWriting(t *testing.T) {
	dir := fixture(t)
	name := strings.Repeat("n", 17000)
	cat := `{"schema":"lifecycle-semantic-oracle-v2","sequences":[{"caseId":"big","name":"` + name + `","executionStatus":"executable"}],"refusals":[],"special":[]}`
	os.WriteFile(filepath.Join(dir, catalogPath), []byte(cat), 0o644)
	reqs := []string{}
	for i := 0; i < 256; i++ {
		reqs = append(reqs, fmt.Sprintf(`{"id":"UR-%d","title":"T","statement":"S%d","modelLinks":[{"kind":"state","id":"Done"}],"acceptanceLinks":[{"caseId":"big"}]}`, i, i))
	}
	doc := `{"formatVersion":"lifecycle-requirements-json/v1","modelId":"TaskLifecycleExample","acceptanceCatalog":"lifecycle-semantic-oracle-v2","requirements":[` + strings.Join(reqs, ",") + `]}`
	os.WriteFile(filepath.Join(dir, requirementsPath), []byte(doc), 0o644)
	for _, cmd := range []string{"validate", "generate", "check"} {
		code, out, errs := exec(t, dir, cmd)
		if code != 2 || out != "" || !strings.Contains(errs, "output limit") {
			t.Fatalf("%s: %d %q %q", cmd, code, out, errs)
		}
	}
	if _, err := os.Stat(filepath.Join(dir, viewDir)); !os.IsNotExist(err) {
		t.Fatal("refused generate created the view directory")
	}
}
