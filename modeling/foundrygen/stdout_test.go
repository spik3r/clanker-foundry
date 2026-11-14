package foundrygen

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

// The first delivery returns the configured result; later diagnostic writes
// work normally so the tests can observe the fixed fallback message.
type deliveryWriter struct {
	limit     int
	err       error
	calls     int
	delivered bytes.Buffer
}

func (w *deliveryWriter) Write(p []byte) (int, error) {
	w.calls++
	n, err := len(p), error(nil)
	if w.calls == 1 {
		if w.limit >= 0 {
			n = min(n, w.limit)
		}
		err = w.err
	}
	w.delivered.Write(p[:n])
	return n, err
}

// A pinned, local-only generator keeps delivery tests independent of warnings
// and real profile behavior. Check deliberately reports ordinary drift.
func deliveryArgs(t *testing.T) []string {
	t.Helper()
	foundry, _ := gitFixture(t)
	writeFile(t, filepath.Join(foundry, "modeling/fake/go.mod"), "module example.test/fake\n\ngo 1.27.0\n")
	h := strings.Repeat("a", 64)
	identity := fmt.Sprintf(`{"contract":"foundry-generator/v1","profile":"p/v1","sourceSHA256":"%s","canonicalSHA256":"%s"}`+"\n", h, h)
	source := fmt.Sprintf(`package main
import("fmt";"os")
func main(){if len(os.Args)>1 && os.Args[1]=="check" {fmt.Fprintln(os.Stderr,"drift: missing .");os.Exit(3)};fmt.Print(%q)}
`, identity)
	writeFile(t, filepath.Join(foundry, "modeling/fake/cmd/fake/main.go"), source)
	git := func(args ...string) string {
		t.Helper()
		cmd := exec.CommandContext(t.Context(), "git", append([]string{"-C", foundry, "-c", "user.name=t", "-c", "user.email=t@example.invalid", "-c", "commit.gpgsign=false"}, args...)...)
		b, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("fixture git %v: %v %s", args, err, b)
		}
		return strings.TrimSpace(string(b))
	}
	git("add", "modeling/fake")
	git("commit", "-q", "-m", "local generator fixture")
	head := git("rev-parse", "HEAD")
	project := t.TempDir()
	writeFile(t, filepath.Join(project, "model.json"), "{}")
	writeFile(t, filepath.Join(project, ProjectFile), strings.Replace(projectJSON(out("fake", "model.json", "out")), rev, head, 1))
	return []string{"plan", "--project", project, "--foundry", foundry}
}

func TestResultDelivery(t *testing.T) {
	args := deliveryArgs(t)
	failure := errors.New("private writer detail")
	var want, diagnostic bytes.Buffer
	if code := Run(args, &want, &diagnostic); code != 0 || want.Len() == 0 || diagnostic.Len() != 0 {
		t.Fatalf("normal delivery: %d %q %q", code, &want, &diagnostic)
	}
	for _, tc := range []struct {
		name  string
		limit int
		err   error
	}{
		{"error", 0, failure}, {"short", 1, nil}, {"partial error", 1, failure}, {"full count with error", want.Len(), failure},
	} {
		t.Run(tc.name, func(t *testing.T) {
			writer := &deliveryWriter{limit: tc.limit, err: tc.err}
			var stderr bytes.Buffer
			if code := Run(args, writer, &stderr); code != exitFailure || stderr.String() != "foundrygen: output write failed\n" {
				t.Fatalf("delivery: %d %q", code, &stderr)
			}
			if writer.calls != 1 || !bytes.Equal(writer.delivered.Bytes(), want.Bytes()[:tc.limit]) {
				t.Fatalf("delivered prefix/calls: %q %d", &writer.delivered, writer.calls)
			}
		})
	}
	t.Run("empty diagnostic skipped", func(t *testing.T) {
		stderr := &deliveryWriter{limit: 0, err: failure}
		var stdout bytes.Buffer
		if code := Run(args, &stdout, stderr); code != 0 || stderr.calls != 0 || !bytes.Equal(stdout.Bytes(), want.Bytes()) {
			t.Fatalf("code=%d stderr calls=%d stdout=%q", code, stderr.calls, &stdout)
		}
	})
}

func TestWarningDeliveryBeforeResult(t *testing.T) {
	args := append(deliveryArgs(t), "--unpinned")
	failure := errors.New("private diagnostic detail")
	var want, warning bytes.Buffer
	if code := Run(args, &want, &warning); code != 0 || want.Len() == 0 || warning.Len() == 0 {
		t.Fatalf("normal warning: %d %q %q", code, &want, &warning)
	}
	for _, tc := range []struct {
		name  string
		limit int
		err   error
	}{
		{"error", 0, failure}, {"short", 1, nil}, {"partial error", 1, failure}, {"full count with error", warning.Len(), failure},
	} {
		t.Run(tc.name, func(t *testing.T) {
			stdout := &deliveryWriter{limit: -1}
			stderr := &deliveryWriter{limit: tc.limit, err: tc.err}
			if code := Run(args, stdout, stderr); code != exitFailure || stdout.calls != 0 {
				t.Fatalf("code=%d stdout calls=%d", code, stdout.calls)
			}
			expected := string(warning.Bytes()[:tc.limit]) + "foundrygen: diagnostic write failed\n"
			if stderr.delivered.String() != expected {
				t.Fatalf("diagnostic=%q want=%q", &stderr.delivered, expected)
			}
		})
	}
}

func TestDiagnosticFailurePreservesPrimaryResult(t *testing.T) {
	drift := append(deliveryArgs(t), "--unpinned")
	drift[0] = "check"
	for _, tc := range []struct {
		name string
		args []string
		code int
	}{
		{"refusal", []string{"unknown"}, exitRefused}, {"drift", drift, exitDrift},
	} {
		for _, err := range []error{nil, errors.New("private diagnostic detail")} {
			t.Run(fmt.Sprintf("%s/error%v", tc.name, err != nil), func(t *testing.T) {
				stdout := &deliveryWriter{limit: -1}
				stderr := &deliveryWriter{limit: 1, err: err}
				if code := Run(tc.args, stdout, stderr); code != tc.code || stdout.calls != 0 {
					t.Fatalf("code=%d want=%d stdout calls=%d", code, tc.code, stdout.calls)
				}
			})
		}
	}
}

func TestFailedResultDeliveryKeepsGeneratedArtifacts(t *testing.T) {
	project := t.TempDir()
	writeFile(t, filepath.Join(project, "model.json"), aliasModel)
	manifest := projectJSON(out("jsonvalidator", "model.json", "generated"))
	writeFile(t, filepath.Join(project, ProjectFile), manifest)
	stdout := &deliveryWriter{limit: 0, err: errors.New("private writer detail")}
	var stderr bytes.Buffer
	args := []string{"generate", "--project", project, "--foundry", repoRoot(t), "--unpinned"}
	if code := Run(args, stdout, &stderr); code != exitFailure || !strings.HasSuffix(stderr.String(), "foundrygen: output write failed\n") {
		t.Fatalf("code=%d diagnostic=%q", code, &stderr)
	}
	if got := readFile(t, filepath.Join(project, "model.json")); got != aliasModel {
		t.Fatal("model changed")
	}
	if got := readFile(t, filepath.Join(project, ProjectFile)); got != manifest {
		t.Fatal("project file changed")
	}
	if code, so, se := runCLI(t, "check", project); code != 0 {
		t.Fatalf("generated artifacts no longer check: %d %q %q", code, so, se)
	}
}

func TestProcessClosedOutputPipes(t *testing.T) {
	args := deliveryArgs(t)
	binary := filepath.Join(t.TempDir(), "foundrygen")
	if runtime.GOOS == "windows" {
		binary += ".exe"
	}
	build := exec.CommandContext(t.Context(), "go", "build", "-o", binary, "./cmd/foundrygen")
	if result, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build entrypoint: %v %s", err, result)
	}
	for _, which := range []string{"stdout", "warning stderr"} {
		t.Run(which, func(t *testing.T) {
			reader, writer, err := os.Pipe()
			if err != nil {
				t.Fatal(err)
			}
			defer writer.Close()
			if err := reader.Close(); err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
			defer cancel()
			childArgs := append([]string{}, args...)
			if which == "warning stderr" {
				childArgs = append(childArgs, "--unpinned")
			}
			cmd := exec.CommandContext(ctx, binary, childArgs...)
			var stdout, stderr bytes.Buffer
			if which == "stdout" {
				cmd.Stdout = writer
				cmd.Stderr = &stderr
			} else {
				cmd.Stdout = &stdout
				cmd.Stderr = writer
			}
			cmd.WaitDelay = 5 * time.Second
			err = cmd.Run()
			if ctx.Err() != nil {
				t.Fatalf("bounded child timed out: %v", ctx.Err())
			}
			var exit *exec.ExitError
			if !errors.As(err, &exit) || exit.ExitCode() != exitFailure {
				t.Fatalf("exit=%v stdout=%q stderr=%q", err, &stdout, &stderr)
			}
			if which == "stdout" && stderr.String() != "foundrygen: output write failed\n" {
				t.Fatalf("diagnostic=%q", &stderr)
			}
			if which == "warning stderr" && stdout.Len() != 0 {
				t.Fatalf("warning failure delivered stdout: %q", &stdout)
			}
		})
	}
}
