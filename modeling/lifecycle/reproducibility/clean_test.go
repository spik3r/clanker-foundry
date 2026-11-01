// Package reproducibility verifies the committed lifecycle example in a scratch copy.
package reproducibility

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// This literal inventory is independent of either generator's output list.
var compilerOutputs = []string{
	"generated/lifecycle.go", "generated/model_paths_test.go", "generated/decision.svg",
	"generated/decision-table.json", "generated/trace.json",
}
var traceOutputs = []string{
	"traceability/view/trace.json", "traceability/view/trace.md", "traceability/view/trace.svg",
}

func copyModule(t *testing.T, src, dst string) map[string][]byte {
	t.Helper()
	original := map[string][]byte{}
	err := filepath.WalkDir(src, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, p)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		if !d.Type().IsRegular() {
			return fmt.Errorf("non-regular module input %s", rel)
		}
		b, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		original[filepath.ToSlash(rel)] = b
		return os.WriteFile(target, b, 0o644)
	})
	if err != nil {
		t.Fatal(err)
	}
	return original
}

func command(t *testing.T, dir, name string, args ...string) (string, error) {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	// Disable network resolution and ambient Go workspace/configuration.
	for _, e := range os.Environ() {
		key, _, _ := strings.Cut(e, "=")
		switch key {
		case "GOTOOLCHAIN", "GOPROXY", "GOSUMDB", "GOWORK", "GOENV", "GOFLAGS":
			continue
		}
		cmd.Env = append(cmd.Env, e)
	}
	cmd.Env = append(cmd.Env, "GOTOOLCHAIN=local", "GOPROXY=off", "GOSUMDB=off", "GOWORK=off", "GOENV=off", "GOFLAGS=")
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func mustRun(t *testing.T, dir, name string, args ...string) {
	t.Helper()
	out, err := command(t, dir, name, args...)
	if err != nil {
		t.Fatalf("%s %v: %v\n%s", name, args, err, out)
	}
}

func TestCleanRegenerationAndFrozenAcceptance(t *testing.T) {
	if runtime.Version() != "go1.27.3" {
		t.Fatalf("requires pinned go1.27.3; got %s", runtime.Version())
	}
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot locate module")
	}
	src := filepath.Dir(filepath.Dir(file))
	root := t.TempDir()
	dir := filepath.Join(root, "module")
	original := copyModule(t, src, dir)
	allOutputs := append(append([]string{}, compilerOutputs...), traceOutputs...)
	for _, p := range allOutputs {
		if _, ok := original[p]; !ok {
			t.Fatalf("committed output missing: %s", p)
		}
		if err := os.Remove(filepath.Join(dir, p)); err != nil {
			t.Fatal(err)
		}
	}
	modelc := filepath.Join(root, "modelc")
	modeltrace := filepath.Join(root, "modeltrace")
	mustRun(t, dir, "go", "build", "-o", modelc, "./cmd/modelc")
	mustRun(t, dir, "go", "build", "-o", modeltrace, "./cmd/modeltrace")
	for generation := 1; generation <= 2; generation++ {
		mustRun(t, dir, modelc, "generate")
		mustRun(t, dir, modeltrace, "generate")
		for _, p := range allOutputs {
			got, err := os.ReadFile(filepath.Join(dir, p))
			if err != nil || !bytes.Equal(got, original[p]) {
				t.Fatalf("generation %d differs for %s: %v", generation, p, err)
			}
		}
		mustRun(t, dir, modelc, "check")
		mustRun(t, dir, modeltrace, "check")
	}
	// Each check must notice a corrupt output; always restore the original bytes.
	for _, group := range []struct {
		tool  string
		paths []string
	}{{modelc, compilerOutputs}, {modeltrace, traceOutputs}} {
		for _, p := range group.paths {
			name := filepath.Join(dir, p)
			if err := os.WriteFile(name, []byte("deliberate drift\n"), 0o644); err != nil {
				t.Fatal(err)
			}
			if out, err := command(t, dir, group.tool, "check"); err == nil {
				t.Fatalf("check missed drift in %s: %s", p, out)
			}
			if err := os.WriteFile(name, original[p], 0o644); err != nil {
				t.Fatal(err)
			}
		}
		mustRun(t, dir, group.tool, "check")
	}
	mustRun(t, dir, "go", "test", "-count=1", "./acceptance", "./rawinputacceptance", "./scenarioacceptance", "./traceabilityacceptance", "./impactacceptance", "./authoring")
	// No source, schema, frozen fixture or unrelated file may be rewritten or added.
	after := copyModule(t, dir, filepath.Join(root, "after"))
	if len(original) != len(after) {
		t.Fatalf("module file count changed: %d -> %d", len(original), len(after))
	}
	for p, want := range original {
		if got, ok := after[p]; !ok || !bytes.Equal(got, want) {
			t.Errorf("module file changed or missing: %s", p)
		}
	}
}
