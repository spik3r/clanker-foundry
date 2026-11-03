package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/revisioncheck"
)

func fixture(t *testing.T) []byte {
	t.Helper()
	b, e := os.ReadFile("../../revisioncheck/testdata/observations.json")
	if e != nil {
		t.Fatal(e)
	}
	var v struct{ Vectors []struct{ Request string } }
	if json.Unmarshal(b, &v) != nil {
		t.Fatal("fixture")
	}
	return []byte(v.Vectors[0].Request)
}

type readerError struct{}

func (readerError) Read([]byte) (int, error) { return 0, errors.New("private read error") }

type shortWriter struct{}

func (shortWriter) Write(b []byte) (int, error) { return len(b) - 1, nil }
func TestRunBoundaries(t *testing.T) {
	raw := fixture(t)
	var out, diag bytes.Buffer
	if code := run(nil, bytes.NewReader(raw), &out, &diag); code != 0 || diag.Len() != 0 {
		t.Fatal("valid request")
	}
	expected, e := revisioncheck.Check(raw)
	if e != nil || !bytes.Equal(out.Bytes(), expected) {
		t.Fatal("complete native receipt")
	}
	for _, err := range []error{revisioncheck.ErrInvalid, revisioncheck.ErrLimit} {
		if failure(err) != 2 {
			t.Fatal("input exit")
		}
	}
	if failure(revisioncheck.ErrBoundary) != 4 {
		t.Fatal("boundary exit")
	}
	if run(nil, bytes.NewReader(raw), shortWriter{}, io.Discard) != 3 {
		t.Fatal("short write")
	}
	if run(nil, readerError{}, io.Discard, io.Discard) != 2 {
		t.Fatal("read failure")
	}
	for _, args := range [][]string{nil, {"execute"}} {
		out.Reset()
		diag.Reset()
		if run(args, bytes.NewBufferString("PRIVATE INPUT"), &out, &diag) != 2 || out.Len() != 0 || bytes.Contains(diag.Bytes(), []byte("PRIVATE INPUT")) {
			t.Fatal("invalid request disclosure")
		}
	}
}
func TestActualClosedStdout(t *testing.T) {
	const child = "FOUNDRY_REVISION_CHECK_CLOSED_STDOUT"
	if os.Getenv(child) == "1" {
		os.Args = []string{"modelrevisioncheck"}
		main()
		return
	}
	raw := fixture(t)
	reader, writer, e := os.Pipe()
	if e != nil {
		t.Fatal(e)
	}
	defer writer.Close()
	if reader.Close() != nil {
		t.Fatal("pipe")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestActualClosedStdout$")
	cmd.Env = append(os.Environ(), child+"=1")
	cmd.Stdin = bytes.NewReader(raw)
	cmd.Stdout = writer
	cmd.WaitDelay = time.Second
	var diag bytes.Buffer
	cmd.Stderr = &diag
	e = cmd.Run()
	var exit *exec.ExitError
	if ctx.Err() != nil || !errors.As(e, &exit) || exit.ExitCode() != 3 || diag.String() != "modelrevisioncheck: output failed\n" {
		t.Fatalf("closed stdout: %v %q", e, diag.String())
	}
}
