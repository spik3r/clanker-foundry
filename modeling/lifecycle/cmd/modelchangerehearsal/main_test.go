package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/changerehearsal"
)

func fixture(t *testing.T, index int) []byte {
	t.Helper()
	b, err := os.ReadFile("../../changerehearsal/testdata/observations.json")
	if err != nil {
		t.Fatal(err)
	}
	var data struct{ Vectors []struct{ Request string } }
	if json.Unmarshal(b, &data) != nil {
		t.Fatal("fixture")
	}
	return []byte(data.Vectors[index].Request)
}

type unreadable struct{ called bool }

func (r *unreadable) Read([]byte) (int, error) {
	r.called = true
	return 0, errors.New("PRIVATE read failure")
}

type shortWriter struct{}

func (shortWriter) Write(b []byte) (int, error) { return len(b) - 1, nil }
func TestRunBoundaries(t *testing.T) {
	for _, args := range [][]string{nil, {"help"}, {"execute", "extra"}} {
		r := &unreadable{}
		var out, diag bytes.Buffer
		if run(context.Background(), args, r, &out, &diag) != 2 || r.called || out.Len() != 0 || diag.String() != "modelchangerehearsal: usage: modelchangerehearsal check|execute < request.json\n" {
			t.Fatal("argument refusal must precede input")
		}
	}
	raw := fixture(t, 0)
	var out, diag bytes.Buffer
	want, err := changerehearsal.Check(raw)
	if err != nil {
		t.Fatal(err)
	}
	if run(context.Background(), []string{"check"}, bytes.NewReader(raw), &out, &diag) != 0 || !bytes.Equal(out.Bytes(), want) || diag.Len() != 0 {
		t.Fatal("native check parity")
	}
	out.Reset()
	diag.Reset()
	if run(context.Background(), []string{"check"}, bytes.NewReader(fixture(t, 3)), &out, &diag) != 1 || out.Len() == 0 || diag.Len() != 0 {
		t.Fatal("stale delivery")
	}
	for _, input := range []io.Reader{&unreadable{}, bytes.NewBufferString("PRIVATE invalid input")} {
		out.Reset()
		diag.Reset()
		if run(context.Background(), []string{"check"}, input, &out, &diag) != 2 || out.Len() != 0 || bytes.Contains(diag.Bytes(), []byte("PRIVATE")) {
			t.Fatal("input refusal disclosure")
		}
	}
	diag.Reset()
	if run(context.Background(), []string{"check"}, bytes.NewReader(raw), shortWriter{}, &diag) != 3 || diag.String() != "modelchangerehearsal: report delivery failed; execution may have started\n" {
		t.Fatal("short delivery")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	out.Reset()
	diag.Reset()
	if run(ctx, []string{"execute"}, bytes.NewReader(raw), &out, &diag) != 4 || diag.String() != "modelchangerehearsal: execution boundary failed; inspect receipt\n" {
		t.Fatal("cancel delivery")
	}
	var report changerehearsal.Report
	if json.Unmarshal(out.Bytes(), &report) != nil || report.Baseline.Operation != "check" || report.Candidate.Operation != "check" {
		t.Fatal("cancel must not fabricate execution")
	}
}
func TestActualBinaryOutputsAndClosedStdout(t *testing.T) {
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(t.TempDir(), "modelchangerehearsal")
	buildCtx, stopBuild := context.WithTimeout(context.Background(), 90*time.Second)
	defer stopBuild()
	build := exec.CommandContext(buildCtx, filepath.Join(runtime.GOROOT(), "bin", "go"), "build", "-o", binary, "./cmd/modelchangerehearsal")
	build.WaitDelay = time.Second
	build.Dir = root
	if b, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build: %v %s", err, b)
	}
	for _, test := range []struct {
		mode        string
		index, code int
		outcome     string
	}{
		{"check", 0, 0, "checked"}, {"execute", 0, 0, "matched"}, {"execute", 1, 1, "baseline-mismatched"}, {"execute", 2, 1, "candidate-mismatched"},
	} {
		t.Run(test.mode+test.outcome, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, binary, test.mode)
			cmd.Dir = root
			cmd.Stdin = bytes.NewReader(fixture(t, test.index))
			cmd.WaitDelay = time.Second
			var out, diag bytes.Buffer
			cmd.Stdout = &out
			cmd.Stderr = &diag
			err := cmd.Run()
			code := 0
			if err != nil {
				var exit *exec.ExitError
				if !errors.As(err, &exit) {
					t.Fatal(err)
				}
				code = exit.ExitCode()
			}
			if ctx.Err() != nil || code != test.code || diag.Len() != 0 {
				t.Fatalf("exit %d: %v %q", code, err, diag.String())
			}
			var report changerehearsal.Report
			if json.Unmarshal(out.Bytes(), &report) != nil || report.Outcome != test.outcome {
				t.Fatal("receipt outcome")
			}
			if test.mode == "execute" && report.Baseline.Operation != "execute" {
				t.Fatal("actual baseline omitted")
			}
			if test.outcome == "baseline-mismatched" && report.Candidate.Operation != "check" {
				t.Fatal("candidate was not suppressed")
			}
		})
	}
	for _, mode := range []string{"check", "execute"} {
		t.Run(mode+"-closed-stdout", func(t *testing.T) {
			reader, writer, err := os.Pipe()
			if err != nil {
				t.Fatal(err)
			}
			defer writer.Close()
			if reader.Close() != nil {
				t.Fatal("pipe")
			}
			ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, binary, mode)
			cmd.Dir = root
			cmd.Stdin = bytes.NewReader(fixture(t, 0))
			cmd.Stdout = writer
			cmd.WaitDelay = time.Second
			var diag bytes.Buffer
			cmd.Stderr = &diag
			err = cmd.Run()
			var exit *exec.ExitError
			if ctx.Err() != nil || !errors.As(err, &exit) || exit.ExitCode() != 3 || diag.String() != "modelchangerehearsal: report delivery failed; execution may have started\n" {
				t.Fatalf("closed stdout: %v %q", err, diag.String())
			}
		})
	}
}
