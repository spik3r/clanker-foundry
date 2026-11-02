package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/scenariodiff"
)

func input(t *testing.T, changed bool) []byte {
	t.Helper()
	m, e := os.ReadFile("../../examples/lifecycle.json")
	if e != nil {
		t.Fatal(e)
	}
	s, e := os.ReadFile("../../scenarios/testdata/lifecycle-v1.json")
	if e != nil {
		t.Fatal(e)
	}
	a := s
	if changed {
		var obj map[string]json.RawMessage
		if e = json.Unmarshal(s, &obj); e != nil {
			t.Fatal(e)
		}
		obj["id"] = []byte(`"RenamedSuite"`)
		a, e = json.Marshal(obj)
		if e != nil {
			t.Fatal(e)
		}
	}
	return []byte(`{"formatVersion":"lifecycle-scenario-diff-json/v1","beforeModel":` + string(m) + `,"beforeSuite":` + string(s) + `,"afterModel":` + string(m) + `,"afterSuite":` + string(a) + `}`)
}

func TestReportsUnchangedAndChanged(t *testing.T) {
	for _, changed := range []bool{false, true} {
		raw := input(t, changed)
		var first []byte
		for range 2 {
			var out, diag bytes.Buffer
			if code := run(nil, bytes.NewReader(raw), &out, &diag); code != 0 || diag.Len() != 0 {
				t.Fatalf("%d %s", code, diag.String())
			}
			var got scenariodiff.Report
			if e := json.Unmarshal(out.Bytes(), &got); e != nil {
				t.Fatal(e)
			}
			if got.ReviewRequired != changed || got.SuiteChanged != changed || got.ModelChanged || got.Mode != "example-only" || got.Authority != "none" {
				t.Fatal(got)
			}
			if first != nil && !bytes.Equal(first, out.Bytes()) {
				t.Fatal("nondeterministic stdout")
			}
			first = bytes.Clone(out.Bytes())
		}
	}
}

type badReader struct{}

func (badReader) Read([]byte) (int, error) { return 0, io.ErrUnexpectedEOF }

type noRead struct{}

func (noRead) Read([]byte) (int, error) { panic("arguments must reject before reading") }

type failedWriter struct {
	fail  bool
	calls int
}

func (w *failedWriter) Write(b []byte) (int, error) {
	w.calls++
	if w.fail {
		return 0, io.ErrClosedPipe
	}
	return len(b) / 2, nil
}

func TestRejectsAndWriteFailures(t *testing.T) {
	for _, r := range []io.Reader{badReader{}, strings.NewReader("private invalid input"), strings.NewReader(strings.Repeat(" ", scenariodiff.MaxRequestBytes+1))} {
		var out, diag bytes.Buffer
		if code := run(nil, r, &out, &diag); code != 2 || out.Len() != 0 || strings.Contains(diag.String(), "private") {
			t.Fatalf("%d %q %q", code, out.String(), diag.String())
		}
	}
	var out, diag bytes.Buffer
	if code := run([]string{"output.json"}, noRead{}, &out, &diag); code != 2 || out.Len() != 0 {
		t.Fatal(code)
	}
	for _, fail := range []bool{false, true} {
		w := &failedWriter{fail: fail}
		if code := run(nil, bytes.NewReader(input(t, false)), w, &diag); code != 3 || w.calls != 1 {
			t.Fatal(code, w.calls)
		}
	}
}

func TestCLIClosedPipe(t *testing.T) {
	const child = "FOUNDRY_SCENARIODIFF_CLOSED_PIPE_CHILD"
	if os.Getenv(child) == "1" {
		os.Args = []string{"modelscenariodiff"}
		main()
		return
	}
	r, w, e := os.Pipe()
	if e != nil {
		t.Fatal(e)
	}
	defer w.Close()
	if e = r.Close(); e != nil {
		t.Fatal(e)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestCLIClosedPipe$")
	cmd.Env = append(os.Environ(), child+"=1")
	cmd.Stdin = bytes.NewReader(input(t, false))
	cmd.Stdout = w
	var diag bytes.Buffer
	cmd.Stderr = &diag
	cmd.WaitDelay = time.Second
	e = cmd.Run()
	if ctx.Err() != nil {
		t.Fatal(ctx.Err())
	}
	var exit *exec.ExitError
	if !errors.As(e, &exit) || exit.ExitCode() != 3 || diag.String() != "scenario comparison output failed\n" {
		t.Fatalf("%v / %q", e, diag.String())
	}
}
