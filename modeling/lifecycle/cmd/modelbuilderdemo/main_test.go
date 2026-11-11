package main

// Literal CLI acceptance frozen before the preparation implementation.
import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func cliRead(t *testing.T, p string) []byte {
	t.Helper()
	b, e := os.ReadFile(p)
	if e != nil {
		t.Fatal(e)
	}
	return b
}
func cliSource(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	var m struct{ Artifacts []struct{ Path string } }
	if e := json.Unmarshal(cliRead(t, "../../builderdemo/inputs.json"), &m); e != nil {
		t.Fatal(e)
	}
	for _, a := range m.Artifacts {
		p := filepath.Join(root, a.Path)
		if e := os.MkdirAll(filepath.Dir(p), 0700); e != nil {
			t.Fatal(e)
		}
		input := filepath.Join("../../../..", a.Path)
		if a.Path == "AGENTS.md" {
			input = "../../builderdemo/testdata/AGENTS-882b5002.txt"
		}
		if e := os.WriteFile(p, cliRead(t, input), 0600); e != nil {
			t.Fatal(e)
		}
	}
	return root
}
func cliJSON(t *testing.T, b []byte) map[string]any {
	t.Helper()
	var v map[string]any
	d := json.NewDecoder(bytes.NewReader(b))
	if e := d.Decode(&v); e != nil {
		t.Fatal(e)
	}
	var extra any
	if e := d.Decode(&extra); e != io.EOF {
		t.Fatal("trailing stdout", e)
	}
	return v
}
func cliArgs(root, parent string) []string {
	return []string{"prepare", "--root", root, "--work-parent", parent, "--go", "/record-only-go"}
}

func TestCLIValidPermutationsAndBlockedReport(t *testing.T) {
	root, parent := cliSource(t), t.TempDir()
	pairs := [][]string{{"--root", root}, {"--work-parent", parent}, {"--go", "/record-only-go"}}
	for _, order := range [][3]int{{0, 1, 2}, {0, 2, 1}, {1, 0, 2}, {1, 2, 0}, {2, 0, 1}, {2, 1, 0}} {
		args := []string{"prepare"}
		for _, i := range order {
			args = append(args, pairs[i]...)
		}
		var out, diag bytes.Buffer
		if code := run(args, &out, &diag); code != 0 || diag.Len() != 0 {
			t.Fatalf("exit=%d diag=%s", code, &diag)
		}
		got := cliJSON(t, out.Bytes())
		if got["outcome"] != "prepared-draft" || got["executionStatus"] != "not-run" || got["authority"] != "none" {
			t.Fatal(got)
		}
	}
	if e := os.WriteFile(filepath.Join(root, "AGENTS.md"), []byte("stale"), 0600); e != nil {
		t.Fatal(e)
	}
	before, e := os.ReadDir(parent)
	if e != nil {
		t.Fatal(e)
	}
	var out, diag bytes.Buffer
	if code := run(cliArgs(root, parent), &out, &diag); code != 1 {
		t.Fatalf("blocked exit=%d diag=%s", code, &diag)
	}
	got := cliJSON(t, out.Bytes())
	if got["outcome"] != "blocked" || got["workspace"] != nil || got["sourceReport"] == nil {
		t.Fatal(got)
	}
	after, e := os.ReadDir(parent)
	if e != nil || len(before) != len(after) {
		t.Fatal("blocked CLI wrote workspace", e)
	}
}
func TestCLIRejectsEveryNonexactInvocationWithoutReport(t *testing.T) {
	root, parent := cliSource(t), t.TempDir()
	good := cliArgs(root, parent)
	cases := [][]string{nil, {}, {"prepare"}, {"--help"}, {"--version"}, good[1:], {"run", "--root", root, "--work-parent", parent, "--go", "/go"},
		{"prepare", "--root", root, "--root", root, "--go", "/go"},
		{"prepare", "--root", root, "--work-parent", parent, "--work-parent", parent},
		{"prepare", "--root", root, "--work-parent", parent, "--unknown", "/go"},
		{"prepare", "--root=" + root, "--work-parent=" + parent, "--go=/go"},
		append(append([]string{}, good...), "extra"),
		append(append([]string{}, good...), "--go", "/go"),
		good[:len(good)-1],
	}
	for _, bad := range []string{"", ".", "relative", "/bad\npath", "/bad\xffpath"} {
		for _, index := range []int{2, 4, 6} {
			a := append([]string{}, good...)
			a[index] = bad
			cases = append(cases, a)
		}
	}
	cases = append(cases, cliArgs(root, root), cliArgs(root, filepath.Join(parent, "missing")))
	for _, args := range cases {
		var out, diag bytes.Buffer
		if code := run(args, &out, &diag); code != 2 || out.Len() != 0 || diag.Len() == 0 {
			t.Fatalf("args=%q exit=%d out=%s diag=%s", args, code, &out, &diag)
		}
	}
	entries, e := os.ReadDir(parent)
	if e != nil || len(entries) != 0 {
		t.Fatal("rejected invocation wrote output", entries, e)
	}
}

type failingOutput struct{ mode int }

func (w failingOutput) Write(p []byte) (int, error) {
	switch w.mode {
	case 0:
		return 0, nil
	case 1:
		return len(p) - 1, nil
	case 2:
		return len(p) / 2, errors.New("output failure")
	default:
		return len(p), errors.New("output failure")
	}
}
func TestCLIOutputFailurePreservesWorkspaceAndNamesIt(t *testing.T) {
	for mode := 0; mode < 4; mode++ {
		root, parent := cliSource(t), t.TempDir()
		var diag bytes.Buffer
		if code := run(cliArgs(root, parent), failingOutput{mode}, &diag); code != 3 {
			t.Fatalf("exit=%d diag=%s", code, &diag)
		}
		entries, e := os.ReadDir(parent)
		if e != nil || len(entries) != 1 {
			t.Fatal("workspace lost or retried", entries, e)
		}
		workspace := filepath.Join(parent, entries[0].Name())
		if !strings.Contains(diag.String(), workspace) {
			t.Fatal("lost recovery path", diag.String(), workspace)
		}
		if _, e := os.Stat(filepath.Join(workspace, "handoff.md")); e != nil {
			t.Fatal(e)
		}
	}
	root, parent := cliSource(t), t.TempDir()
	if e := os.Remove(filepath.Join(root, "AGENTS.md")); e != nil {
		t.Fatal(e)
	}
	var diag bytes.Buffer
	if code := run(cliArgs(root, parent), failingOutput{0}, &diag); code != 3 {
		t.Fatal("output failure must override blocked exit", code)
	}
}
func TestCLIHasNoStdinDependency(t *testing.T) {
	// A closed stdin detects accidental reads without running any candidate.
	old := os.Stdin
	f, e := os.CreateTemp(t.TempDir(), "closed-stdin")
	if e != nil {
		t.Fatal(e)
	}
	if e = f.Close(); e != nil {
		t.Fatal(e)
	}
	os.Stdin = f
	defer func() { os.Stdin = old }()
	root, parent := cliSource(t), t.TempDir()
	var out, diag bytes.Buffer
	if code := run(cliArgs(root, parent), &out, &diag); code != 0 {
		t.Fatalf("stdin affected preparation: %d %s", code, &diag)
	}
}

func TestCLIIncompleteExitFour(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root bypasses permission bits")
	}
	root, parent := cliSource(t), t.TempDir()
	if e := os.Chmod(parent, 0500); e != nil {
		t.Fatal(e)
	}
	defer os.Chmod(parent, 0700)
	var out, diag bytes.Buffer
	if code := run(cliArgs(root, parent), &out, &diag); code != 4 {
		t.Fatalf("exit=%d diag=%s", code, &diag)
	}
	got := cliJSON(t, out.Bytes())
	if got["outcome"] != "incomplete" || got["workspace"] != nil {
		t.Fatal(got)
	}
}

// Historical replay is test data selection, not a current-source exemption.
func TestCLIHistoricalInstructionsReplayAndChangedInstructionsRefuse(t *testing.T) {
	historical := cliRead(t, "../../builderdemo/testdata/AGENTS-882b5002.txt")
	current := cliRead(t, "../../../../AGENTS.md")
	mutated := append(append([]byte{}, historical...), []byte("\nfixture-only mutation\n")...)
	cases := []struct {
		name    string
		input   []byte
		allowed bool
	}{
		{"historical", historical, true},
		{"current", current, bytes.Equal(current, historical)},
		{"mutated", mutated, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root, parent := cliSource(t), t.TempDir()
			target := filepath.Join(root, "AGENTS.md")
			if !bytes.Equal(cliRead(t, target), historical) {
				t.Fatal("CLI fixture did not replay historical instructions")
			}
			if e := os.WriteFile(target, tc.input, 0600); e != nil {
				t.Fatal(e)
			}
			var out, diag bytes.Buffer
			code := run(cliArgs(root, parent), &out, &diag)
			got := cliJSON(t, out.Bytes())
			if got["executionStatus"] != "not-run" || got["authority"] != "none" {
				t.Fatal("unexpected execution or authority", got)
			}
			if tc.allowed {
				if code != 0 || diag.Len() != 0 || got["outcome"] != "prepared-draft" {
					t.Fatalf("historical preparation: exit=%d diag=%s report=%v", code, &diag, got)
				}
				workspace, ok := got["workspace"].(string)
				if !ok || !bytes.Equal(cliRead(t, filepath.Join(workspace, "inputs", "AGENTS.md")), historical) {
					t.Fatal("CLI prepared instructions differ from the historical snapshot")
				}
			} else {
				if code != 1 || got["outcome"] != "blocked" || got["workspace"] != nil || got["handoffPath"] != nil || got["sourceReport"] == nil || got["inputsReport"] != nil || got["exerciseReport"] != nil {
					t.Fatalf("changed instructions: exit=%d report=%v", code, got)
				}
				entries, e := os.ReadDir(parent)
				if e != nil || len(entries) != 0 {
					t.Fatalf("blocked CLI wrote output: %v %v", entries, e)
				}
			}
			if !bytes.Equal(cliRead(t, target), tc.input) {
				t.Fatal("CLI preparation rewrote the instruction input")
			}
		})
	}
}
