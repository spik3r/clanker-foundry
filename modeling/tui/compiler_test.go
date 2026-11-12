package tui

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

func sampleModel(t *testing.T) Model {
	t.Helper()
	raw, err := os.ReadFile("examples/browser.json")
	if err != nil {
		t.Fatal(err)
	}
	m, err := Decode(raw)
	if err != nil {
		t.Fatal(err)
	}
	return m
}
func TestCompilerDeterministicAndDoesNotMutateModel(t *testing.T) {
	m := sampleModel(t)
	before, _ := json.Marshal(m)
	first, err := Compile(m)
	if err != nil {
		t.Fatal(err)
	}
	second, err := Compile(m)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(first, second) {
		t.Fatal("non-deterministic emission")
	}
	after, _ := json.Marshal(m)
	if !bytes.Equal(before, after) {
		t.Fatal("compiler mutated its input")
	}
	m.Fields[0], m.Fields[3] = m.Fields[3], m.Fields[0]
	permuted, err := Compile(m)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(first, permuted) {
		t.Fatal("field declaration order changed output")
	}
}
func TestCompilerRejectsInvalidTypedModel(t *testing.T) {
	m := sampleModel(t)
	m.Identity = "Missing"
	if _, err := Compile(m); err == nil {
		t.Fatal("accepted invalid typed model")
	}
}
func TestGeneratedBehaviorAndRequirementChange(t *testing.T) {
	if testing.Short() {
		t.Skip("generated Go acceptance invokes compiler")
	}
	m := sampleModel(t)
	for _, variant := range []string{"base", "changed"} {
		t.Run(variant, func(t *testing.T) {
			if variant == "changed" {
				m.Title = "Changed catalog"
				m.Layout.WideWidth = 100
				m.Layout.ListWidth = 32
				m.Details = []string{"Status", "Name"}
			}
			files, err := Compile(m)
			if err != nil {
				t.Fatal(err)
			}
			dir := t.TempDir()
			for name, raw := range files {
				if err := os.MkdirAll(filepath.Dir(filepath.Join(dir, name)), 0700); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(dir, name), raw, 0600); err != nil {
					t.Fatal(err)
				}
			}
			acceptance, err := os.ReadFile("acceptance/browser_test.go.txt")
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(dir, "browser_acceptance_test.go"), acceptance, 0600); err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
			defer cancel()
			cmd := exec.CommandContext(ctx, "go", "test", "-mod=readonly", "-race", "-count=1", "-p=1", "./...")
			cmd.Dir = dir
			cmd.Env = append(os.Environ(), "GOTOOLCHAIN=local", "GOMAXPROCS=2", "TUI_ACCEPTANCE_VARIANT="+variant)
			if out, err := cmd.CombinedOutput(); err != nil {
				t.Fatalf("generated %s behavior: %v\n%s", variant, err, out)
			}
			// The accepted file set must remain unchanged after compilation/testing.
			for name, raw := range files {
				got, err := os.ReadFile(filepath.Join(dir, name))
				if err != nil || !bytes.Equal(raw, got) {
					t.Fatalf("generated file drift: %s", name)
				}
			}
		})
	}
}
