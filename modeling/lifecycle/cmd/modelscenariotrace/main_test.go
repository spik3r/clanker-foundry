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

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/scenariotrace"
)

func request(t *testing.T, report string) []byte {
	t.Helper()
	raw, err := os.ReadFile("../../scenariotrace/testdata/vectors.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixtures struct {
		Sources map[string]string `json:"sources"`
		Reports map[string]string `json:"reports"`
	}
	if err := json.Unmarshal(raw, &fixtures); err != nil {
		t.Fatal(err)
	}
	wrapper := map[string]any{"formatVersion": "lifecycle-scenario-trace-request/v1", "scenarioReport": nil}
	for k, v := range fixtures.Sources {
		wrapper[k] = v
	}
	if report != "" {
		value, ok := fixtures.Reports[report]
		if !ok {
			t.Fatalf("missing fixture report %q", report)
		}
		wrapper["scenarioReport"] = value
	}
	raw, err = json.Marshal(wrapper)
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
	for _, report := range []string{"", "matched", "early"} {
		t.Run(report, func(t *testing.T) {
			raw := request(t, report)
			original := bytes.Clone(raw)
			want, err := scenariotrace.Check(raw)
			if err != nil {
				t.Fatal(err)
			}
			for range 2 {
				var out countingWriter
				var diag bytes.Buffer
				if code := run(nil, bytes.NewReader(raw), &out, &diag); code != 0 || diag.Len() != 0 || out.calls != 1 || !bytes.Equal(out.Bytes(), want) {
					t.Fatalf("exit=%d writes=%d diagnostic=%q", code, out.calls, diag.String())
				}
			}
			if !bytes.Equal(raw, original) {
				t.Fatal("mutated request")
			}
		})
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
	stale := request(t, "")
	var wrapper map[string]any
	if err := json.Unmarshal(stale, &wrapper); err != nil {
		t.Fatal(err)
	}
	var bindings map[string]any
	if err := json.Unmarshal([]byte(wrapper["bindings"].(string)), &bindings); err != nil {
		t.Fatal(err)
	}
	bindings["modelDigest"] = strings.Repeat("0", 64)
	encoded, err := json.Marshal(bindings)
	if err != nil {
		t.Fatal(err)
	}
	wrapper["bindings"] = string(encoded)
	stale, err = json.Marshal(wrapper)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name string
		args []string
		in   io.Reader
		code int
		diag string
	}{
		{"arguments", []string{"private-path"}, noRead{}, 2, "modelscenariotrace accepts stdin only\n"},
		{"read", nil, badReader{}, 4, "scenario trace input read failed\n"},
		{"invalid", nil, strings.NewReader("private malformed input"), 2, "scenario trace input rejected\n"},
		{"binding", nil, bytes.NewReader(stale), 2, "scenario trace input rejected\n"},
		{"limit", nil, &spacesReader{}, 2, "scenario trace input rejected\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var out countingWriter
			var diag bytes.Buffer
			if code := run(tc.args, tc.in, &out, &diag); code != tc.code || out.calls != 0 || diag.String() != tc.diag {
				t.Fatalf("exit=%d writes=%d diagnostic=%q", code, out.calls, diag.String())
			}
			if r, ok := tc.in.(*spacesReader); ok && r.count != scenariotrace.MaxRequestBytes+1 {
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
		{scenariotrace.ErrInvalid, 2}, {scenariotrace.ErrBinding, 2}, {scenariotrace.ErrLimit, 2},
		{scenariotrace.ErrBoundary, 4}, {errors.New("private unexpected error"), 4},
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
		if code := run(nil, bytes.NewReader(request(t, "")), out, &diag); code != 3 || out.calls != 1 || diag.String() != "scenario trace output failed\n" {
			t.Fatalf("exit=%d writes=%d diagnostic=%q", code, out.calls, diag.String())
		}
	}
}

func TestCompiledCLI(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	binary := filepath.Join(t.TempDir(), "modelscenariotrace")
	build := exec.CommandContext(ctx, filepath.Join(runtime.GOROOT(), "bin", "go"), "build", "-o", binary, ".")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build: %v\n%s", err, output)
	}
	raw := request(t, "")
	want, err := scenariotrace.Check(raw)
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
		if !errors.As(err, &exit) || exit.ExitCode() != 3 || diag.String() != "scenario trace output failed\n" {
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
		if !errors.As(err, &exit) || exit.ExitCode() != 4 || out.Len() != 0 || diag.String() != "scenario trace input read failed\n" {
			t.Fatalf("run: %v stdout=%q diagnostic=%q", err, out.String(), diag.String())
		}
	})
}
