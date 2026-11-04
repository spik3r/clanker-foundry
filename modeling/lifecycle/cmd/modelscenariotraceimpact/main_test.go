package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/scenariotraceimpact"
)

func request(t *testing.T) []byte {
	t.Helper()
	raw, err := os.ReadFile("../../scenariotrace/testdata/vectors.json")
	if err != nil {
		t.Fatal(err)
	}
	var v struct {
		Sources map[string]string `json:"sources"`
	}
	if err := json.Unmarshal(raw, &v); err != nil {
		t.Fatal(err)
	}
	source := map[string]any{"formatVersion": "lifecycle-scenario-trace-request/v1", "scenarioReport": nil}
	for k, value := range v.Sources {
		source[k] = value
	}
	snapshot, err := json.Marshal(source)
	if err != nil {
		t.Fatal(err)
	}
	raw, err = json.Marshal(map[string]string{"formatVersion": scenariotraceimpact.RequestVersion, "before": string(snapshot), "after": string(snapshot)})
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

type countingWriter struct {
	bytes.Buffer
	calls int
}

func (w *countingWriter) Write(p []byte) (int, error) { w.calls++; return w.Buffer.Write(p) }

func TestRunCompleteOutput(t *testing.T) {
	raw := request(t)
	want, err := scenariotraceimpact.Check(raw)
	if err != nil {
		t.Fatal(err)
	}
	var out countingWriter
	var diag bytes.Buffer
	if code := run(nil, bytes.NewReader(raw), &out, &diag); code != 0 || out.calls != 1 || diag.Len() != 0 || !bytes.Equal(out.Bytes(), want) {
		t.Fatalf("exit=%d writes=%d diagnostic=%q", code, out.calls, diag.String())
	}
}

type noRead struct{}

func (noRead) Read([]byte) (int, error) { panic("arguments must reject before reading") }

type badReader struct{}

func (badReader) Read(p []byte) (int, error) {
	return copy(p, "private input"), errors.New("private reader error")
}

type spacesReader struct{ count int }

func (r *spacesReader) Read(p []byte) (int, error) {
	for i := range p {
		p[i] = ' '
	}
	r.count += len(p)
	return len(p), nil
}

func TestRunRefusals(t *testing.T) {
	for _, tc := range []struct {
		name string
		args []string
		in   io.Reader
		diag string
	}{
		{"arguments", []string{"private-path"}, noRead{}, "modelscenariotraceimpact accepts stdin only\n"},
		{"read", nil, badReader{}, "scenario trace impact input read failed\n"},
		{"invalid", nil, strings.NewReader("private malformed input"), "scenario trace impact input rejected\n"},
		{"limit", nil, &spacesReader{}, "scenario trace impact input rejected\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var out countingWriter
			var diag bytes.Buffer
			if code := run(tc.args, tc.in, &out, &diag); code != 2 || out.calls != 0 || diag.String() != tc.diag {
				t.Fatalf("exit=%d writes=%d diagnostic=%q", code, out.calls, diag.String())
			}
			if r, ok := tc.in.(*spacesReader); ok && r.count != scenariotraceimpact.MaxRequestBytes+1 {
				t.Fatalf("read %d bytes", r.count)
			}
		})
	}
}

func TestFailureClassification(t *testing.T) {
	for _, tc := range []struct {
		err  error
		code int
	}{
		{scenariotraceimpact.ErrInvalid, 2}, {scenariotraceimpact.ErrLimit, 2},
		{scenariotraceimpact.ErrBoundary, 4}, {errors.New("private unexpected error"), 4},
	} {
		for _, err := range []error{tc.err, fmt.Errorf("private context: %w", tc.err)} {
			if got := failure(err); got != tc.code {
				t.Fatalf("exit=%d want=%d", got, tc.code)
			}
		}
	}
}

type failedWriter struct {
	calls int
	fail  bool
}

func (w *failedWriter) Write(p []byte) (int, error) {
	w.calls++
	if w.fail {
		return 0, errors.New("private write error")
	}
	return len(p) / 2, nil
}

func TestRunDeliveryFailures(t *testing.T) {
	for _, fail := range []bool{false, true} {
		out := &failedWriter{fail: fail}
		var diag bytes.Buffer
		if code := run(nil, bytes.NewReader(request(t)), out, &diag); code != 3 || out.calls != 1 || diag.String() != "scenario trace impact output failed\n" {
			t.Fatalf("exit=%d writes=%d diagnostic=%q", code, out.calls, diag.String())
		}
	}
}

func TestCompiledCLI(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	binary := filepath.Join(t.TempDir(), "modelscenariotraceimpact.exe")
	build := exec.CommandContext(ctx, filepath.Join(runtime.GOROOT(), "bin", "go"), "build", "-o", binary, ".")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build: %v\n%s", err, output)
	}
	raw := request(t)
	want, err := scenariotraceimpact.Check(raw)
	if err != nil {
		t.Fatal(err)
	}
	t.Run("complete", func(t *testing.T) {
		cmd := exec.CommandContext(ctx, binary)
		cmd.Stdin = bytes.NewReader(raw)
		var out, diag bytes.Buffer
		cmd.Stdout, cmd.Stderr = &out, &diag
		if err := cmd.Run(); err != nil || !bytes.Equal(out.Bytes(), want) || diag.Len() != 0 {
			t.Fatalf("run: %v diagnostic=%q", err, diag.String())
		}
	})
	t.Run("closed-pipe", func(t *testing.T) {
		r, w, err := os.Pipe()
		if err != nil {
			t.Fatal(err)
		}
		defer w.Close()
		if err := r.Close(); err != nil {
			t.Fatal(err)
		}
		cmd := exec.CommandContext(ctx, binary)
		cmd.Stdin, cmd.Stdout = bytes.NewReader(raw), w
		var diag bytes.Buffer
		cmd.Stderr = &diag
		err = cmd.Run()
		var exit *exec.ExitError
		if !errors.As(err, &exit) || exit.ExitCode() != 3 || diag.String() != "scenario trace impact output failed\n" {
			t.Fatalf("run: %v diagnostic=%q", err, diag.String())
		}
	})
	t.Run("read-failure", func(t *testing.T) {
		dir, err := os.Open(".")
		if err != nil {
			t.Fatal(err)
		}
		defer dir.Close()
		cmd := exec.CommandContext(ctx, binary)
		cmd.Stdin = dir
		var out, diag bytes.Buffer
		cmd.Stdout, cmd.Stderr = &out, &diag
		err = cmd.Run()
		var exit *exec.ExitError
		if !errors.As(err, &exit) || exit.ExitCode() != 2 || out.Len() != 0 || diag.String() != "scenario trace impact input read failed\n" {
			t.Fatalf("run: %v diagnostic=%q", err, diag.String())
		}
	})
}
