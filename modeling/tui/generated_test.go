package tui

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func writeFiles(t *testing.T, dir string, files map[string][]byte) {
	t.Helper()
	for name, raw := range files {
		if err := os.MkdirAll(filepath.Dir(filepath.Join(dir, name)), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, name), raw, 0600); err != nil {
			t.Fatal(err)
		}
	}
}

// addAcceptance copies the acceptance tests into dir under the given package.
func addAcceptance(t *testing.T, dir, pkg string) {
	t.Helper()
	for _, name := range []string{"browser_test.go.txt", "wrap_test.go.txt"} {
		raw, err := os.ReadFile(filepath.Join("acceptance", name))
		if err != nil {
			t.Fatal(err)
		}
		raw = bytes.Replace(raw, []byte("package recordbrowser\n"), []byte("package "+pkg+"\n"), 1)
		if err := os.WriteFile(filepath.Join(dir, strings.TrimSuffix(name, ".txt")), raw, 0600); err != nil {
			t.Fatal(err)
		}
	}
}

// goIn runs the go command with the local toolchain in dir.
func goIn(t *testing.T, dir string, env []string, args ...string) string {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, "go", args...)
	cmd.Dir = dir
	cmd.Env = append(append(os.Environ(), "GOTOOLCHAIN=local", "GOMAXPROCS=2", "GOFLAGS=-mod=readonly"), env...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go %s: %v\n%s", strings.Join(args, " "), err, out)
	}
	return string(out)
}

func TestGeneratedBehaviorAndRequirementChange(t *testing.T) {
	if testing.Short() {
		t.Skip("generated Go acceptance invokes compiler")
	}
	for _, variant := range []string{"base", "changed", "wrap"} {
		t.Run(variant, func(t *testing.T) {
			m := sampleModel(t)
			run := "."
			switch variant {
			case "changed":
				m.Title = "Changed catalog"
				m.Layout.WideWidth = 100
				m.Layout.ListWidth = 32
				m.Details = []string{"Status", "Name"}
			case "wrap":
				m.DetailText = TextWrap
				run = "^TestWrap"
			}
			files, err := Compile(m)
			if err != nil {
				t.Fatal(err)
			}
			dir := t.TempDir()
			writeFiles(t, dir, files)
			addAcceptance(t, dir, DefaultPackage)
			out := goIn(t, dir, []string{"TUI_ACCEPTANCE_VARIANT=" + variant}, "test", "-race", "-count=1", "-p=1", "-run", run, "-v", "./...")
			if variant == "wrap" && (!strings.Contains(out, "--- PASS: TestWrapResizeRewrapsAndClamps") || strings.Contains(out, "--- SKIP")) {
				t.Fatalf("wrap tests did not run:\n%s", out)
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

const hostMain = `package main

import (
	"fmt"
	"os"
	"strings"

	tea "charm.land/bubbletea/v2"

	alpha "example.com/host/internal/a"
	beta "example.com/host/internal/b"
)

func view(m tea.Model) string {
	m, _ = m.Update(tea.WindowSizeMsg{Width: 100, Height: 12})
	return m.View().Content
}

func main() {
	a, err := alpha.New([]alpha.Record{{ID: "1", Name: "Alpha row", Status: "up", Description: "first"}})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	b, err := beta.New([]beta.Record{{Key: "k", Title: "Beta row", Body: strings.Repeat("long body ", 40) + "end"}})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	va, vb := view(a), view(b)
	if !strings.Contains(va, "Alpha row") || !strings.Contains(vb, "Beta row") || !strings.Contains(vb, "Notes") || strings.Contains(vb, "Alpha") {
		fmt.Fprintln(os.Stderr, "unexpected views")
		os.Exit(1)
	}
	fmt.Println("hosted both browsers")
}
`

// A package-mode browser must compile inside a host module that already requires
// compatible dependency versions, and two browsers must coexist there.
func TestPackageModeBuildsInsideHostModule(t *testing.T) {
	if testing.Short() {
		t.Skip("generated Go acceptance invokes compiler")
	}
	module, err := Compile(sampleModel(t))
	if err != nil {
		t.Fatal(err)
	}
	// The host takes the dependency requirements of the module-mode output but
	// has its own module path and a lower go directive than the generator's.
	gomod := strings.Replace(string(module["go.mod"]), "module example.com/generated/recordbrowser", "module example.com/host", 1)
	gomod = strings.Replace(gomod, "\ngo 1.25.5\n", "\ngo 1.26.0\n", 1)
	if !strings.HasPrefix(gomod, "module example.com/host\n") || !strings.Contains(gomod, "\ngo 1.26.0\n") {
		t.Fatalf("unexpected module-mode go.mod: %s", module["go.mod"])
	}

	a := sampleModel(t)
	a.Emit = &Emit{Mode: ModePackage, Package: "alpha"}
	b := sampleModel(t)
	b.ID = "NotesBrowser"
	b.Title = "Notes"
	b.Fields = []Field{{ID: "Key", Label: "Key"}, {ID: "Title", Label: "Title"}, {ID: "Body", Label: "Body"}}
	b.Identity, b.Summary, b.Details = "Key", "Title", []string{"Title", "Body"}
	b.Emit = &Emit{Mode: ModePackage, Package: "beta"}
	b.DetailText = TextWrap

	host := t.TempDir()
	writeFiles(t, host, map[string][]byte{"go.mod": []byte(gomod), "go.sum": module["go.sum"], "main.go": []byte(hostMain)})
	for dir, m := range map[string]Model{"internal/a": a, "internal/b": b} {
		files, err := Compile(m)
		if err != nil {
			t.Fatal(err)
		}
		for name := range files {
			if name == "go.mod" || name == "go.sum" || strings.HasPrefix(name, "cmd/") {
				t.Fatalf("package mode emitted %s", name)
			}
		}
		out := filepath.Join(host, filepath.FromSlash(dir))
		writeFiles(t, out, files)
		if m.Emit.Package == "alpha" {
			addAcceptance(t, out, "alpha")
		}
		head, _ := os.ReadFile(filepath.Join(out, "view.go"))
		if !strings.Contains(string(head), "\npackage "+m.Emit.Package+"\n") {
			t.Fatalf("%s: package clause is not %s", dir, m.Emit.Package)
		}
	}
	goIn(t, host, nil, "vet", "./...")
	if out := goIn(t, host, nil, "run", "."); !strings.Contains(out, "hosted both browsers") {
		t.Fatalf("host output: %s", out)
	}
	goIn(t, host, []string{"TUI_ACCEPTANCE_VARIANT=base"}, "test", "-race", "-count=1", "-p=1", "./...")
}

// filterModel is the sample model with a filter over Name and Description.
func filterModel(t *testing.T) Model {
	t.Helper()
	m := sampleModel(t)
	m.Filter = &Filter{Fields: []string{"Name", "Description"}}
	return m
}

// The list filter behaves as documented in both embed modes.
func TestGeneratedFilterBehavior(t *testing.T) {
	if testing.Short() {
		t.Skip("generated Go acceptance invokes compiler")
	}
	for _, variant := range []string{"standalone", "hosted"} {
		t.Run(variant, func(t *testing.T) {
			m := filterModel(t)
			if variant == "hosted" {
				m.Embed = EmbedHosted
			}
			files, err := Compile(m)
			if err != nil {
				t.Fatal(err)
			}
			dir := t.TempDir()
			writeFiles(t, dir, files)
			raw, err := os.ReadFile(filepath.Join("acceptance", "filter_test.go.txt"))
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(dir, "filter_test.go"), raw, 0600); err != nil {
				t.Fatal(err)
			}
			goIn(t, dir, nil, "vet", "./...")
			out := goIn(t, dir, []string{"TUI_ACCEPTANCE_VARIANT=" + variant}, "test", "-race", "-count=1", "-p=1", "-run", "^TestFilter", "-v", "./...")
			for _, want := range []string{"--- PASS: TestFilterOpenTypeAndMatchRule", "--- PASS: TestFilterEscPrecedence", "--- PASS: TestFilterInputCapturesEveryKey", "--- PASS: TestFilterBoundsAndSanitising", "--- PASS: TestFilterNavigationUsesTheFilteredList", "--- PASS: TestFilterKeyHintIDs", "--- PASS: TestFilterValidReplacementCharacter"} {
				if !strings.Contains(out, want) {
					t.Fatalf("missing %q in\n%s", want, out)
				}
			}
		})
	}
}
