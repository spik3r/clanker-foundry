package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/scenarios"
)

func source(t *testing.T) []byte {
	t.Helper()
	b, e := os.ReadFile("../../scenarios/testdata/lifecycle-v1.json")
	if e != nil {
		t.Fatal(e)
	}
	return b
}

type badReader struct{ calls int }

func (r *badReader) Read([]byte) (int, error) {
	r.calls++
	return 0, errors.New("PRIVATE_READER_DETAIL")
}

type shortWriter struct{}

func (shortWriter) Write(b []byte) (int, error) { return len(b) - 1, nil }

type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) { return 0, errors.New("PRIVATE_WRITER_DETAIL") }
func TestCLIStatusesAndNoPrivateDiagnostics(t *testing.T) {
	raw := source(t)
	var out, diag bytes.Buffer
	if n := run(nil, bytes.NewReader(raw), &out, &diag); n != 0 || diag.Len() != 0 || !json.Valid(out.Bytes()) {
		t.Fatal("valid CLI", n, diag.String())
	}
	s, e := scenarios.Decode(raw)
	if e != nil {
		t.Fatal(e)
	}
	s.Cases[0].Steps[0].Expected.Task.State = "Done"
	altered, e := scenarios.Encode(s)
	if e != nil {
		t.Fatal(e)
	}
	out.Reset()
	diag.Reset()
	if n := run(nil, bytes.NewReader(altered), &out, &diag); n != 1 || !json.Valid(out.Bytes()) || diag.Len() != 0 {
		t.Fatal("mismatch status", n)
	}
	for _, b := range [][]byte{[]byte(`{"PRIVATE_INPUT":true}`), bytes.Repeat([]byte("x"), scenarios.MaxSourceBytes+1)} {
		out.Reset()
		diag.Reset()
		if n := run(nil, bytes.NewReader(b), &out, &diag); n != 2 || out.Len() != 0 || strings.Contains(diag.String(), "PRIVATE") {
			t.Fatal("invalid CLI", n)
		}
	}
	reader := &badReader{}
	out.Reset()
	diag.Reset()
	if n := run([]string{"--out", "private-path"}, reader, &out, &diag); n != 2 || reader.calls != 0 || out.Len() != 0 || strings.Contains(diag.String(), "private-path") {
		t.Fatal("argument gate")
	}
	out.Reset()
	diag.Reset()
	if n := run(nil, reader, &out, &diag); n != 2 || out.Len() != 0 || strings.Contains(diag.String(), "PRIVATE") {
		t.Fatal("reader failure leaked")
	}
	for _, w := range []io.Writer{shortWriter{}, failingWriter{}} {
		diag.Reset()
		if n := run(nil, bytes.NewReader(raw), w, &diag); n != 3 || strings.Contains(diag.String(), "PRIVATE") {
			t.Fatal("write failure", n)
		}
	}
}
func TestSeparateProcessRepeatabilityAndNoFiles(t *testing.T) {
	dir := t.TempDir()
	bin := filepath.Join(dir, "modelplay")
	c := exec.Command(filepath.Join(runtime.GOROOT(), "bin", "go"), "build", "-o", bin, ".")
	c.Env = append(os.Environ(), "GOTOOLCHAIN=local", "GOPROXY=off", "GOSUMDB=off")
	if b, e := c.CombinedOutput(); e != nil {
		t.Fatalf("build: %v %s", e, b)
	}
	executionDir := t.TempDir()
	var first []byte
	for i := 0; i < 2; i++ {
		c := exec.Command(bin)
		c.Dir = executionDir
		c.Stdin = bytes.NewReader(source(t))
		b, e := c.Output()
		if e != nil {
			t.Fatal(e)
		}
		if i == 0 {
			first = b
		} else if !bytes.Equal(first, b) {
			t.Fatal("process nondeterminism")
		}
	}
	files, e := os.ReadDir(executionDir)
	if e != nil || len(files) != 0 {
		t.Fatal("CLI wrote files", e)
	}
}
