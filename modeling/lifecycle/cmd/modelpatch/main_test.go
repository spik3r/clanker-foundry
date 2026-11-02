package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/canonical"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/jsonformat"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/modelpatch"
	"io"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

func input(t *testing.T) []byte {
	t.Helper()
	raw, e := os.ReadFile("../../examples/lifecycle.json")
	if e != nil {
		t.Fatal(e)
	}
	m, e := jsonformat.Decode(raw)
	if e != nil {
		t.Fatal(e)
	}
	digest, e := canonical.Digest(m)
	if e != nil {
		t.Fatal(e)
	}
	return []byte(`{"formatVersion":"lifecycle-model-patch-json/v1","patchVersion":"lifecycle-model-patch/v1","profileVersion":"lifecycle/v1","before":` + string(raw) + `,"beforeDigest":"` + digest + `","operations":[]}`)
}

type failedReader struct{}

func (failedReader) Read([]byte) (int, error) { return 0, errors.New("interrupted") }

type partialReader struct{ once bool }

func (r *partialReader) Read(p []byte) (int, error) {
	if r.once {
		return 0, io.ErrUnexpectedEOF
	}
	r.once = true
	return copy(p, []byte(`{"formatVersion":`)), nil
}

type shortWriter struct {
	fail   bool
	called int
}

func (w *shortWriter) Write(p []byte) (int, error) {
	w.called++
	if w.fail {
		return len(p) / 2, io.ErrClosedPipe
	}
	return len(p) / 2, nil
}
func TestCLIActualNoopAndDeterministicStdout(t *testing.T) {
	raw := input(t)
	var out, diag bytes.Buffer
	if code := run(nil, bytes.NewReader(raw), &out, &diag); code != 0 {
		t.Fatal(code, diag.String())
	}
	expected, e := modelpatch.Execute(raw)
	if e != nil || !bytes.Equal(expected, out.Bytes()) {
		t.Fatal("CLI output differs from actual pure API")
	}
	var parsed map[string]any
	if json.Unmarshal(out.Bytes(), &parsed) != nil || parsed["mode"] != "example-only" || parsed["authority"] != "none" || parsed["beforeDigest"] != parsed["afterDigest"] {
		t.Fatal("CLI envelope")
	}
	snapshot := sha256.Sum256(out.Bytes())
	out.Reset()
	if code := run(nil, bytes.NewReader(raw), &out, &diag); code != 0 {
		t.Fatal(code)
	}
	now := sha256.Sum256(out.Bytes())
	if hex.EncodeToString(snapshot[:]) != hex.EncodeToString(now[:]) {
		t.Fatal("CLI replay differs")
	}
}
func TestCLIReadErrorsBoundsArgumentsAndShortWrites(t *testing.T) {
	for _, reader := range []io.Reader{failedReader{}, &partialReader{}, strings.NewReader("private invalid input"), strings.NewReader(strings.Repeat(" ", modelpatch.MaxRequestBytes+1))} {
		var out, diag bytes.Buffer
		if code := run(nil, reader, &out, &diag); code != 2 || out.Len() != 0 || strings.Contains(diag.String(), "private invalid input") {
			t.Fatal("rejected input leaked output/text")
		}
	}
	var out, diag bytes.Buffer
	if code := run([]string{"/tmp/output"}, bytes.NewReader(input(t)), &out, &diag); code != 2 || out.Len() != 0 {
		t.Fatal("path argument accepted")
	}
	for _, fail := range []bool{false, true} {
		w := &shortWriter{fail: fail}
		if code := run(nil, bytes.NewReader(input(t)), w, &diag); code != 3 || w.called != 1 {
			t.Fatal("short/failed write accepted", code)
		}
	}
}

// A real closed stdout pipe must use the documented output-error exit, not a
// process signal. This invokes main in a fresh bounded test process, not run with
// a fake writer; only this trusted test binary is launched.
func TestCLIClosedStdoutProcess(t *testing.T) {
	const child = "FOUNDRY_MODELPATCH_CLOSED_STDOUT_CHILD"
	if os.Getenv(child) == "1" {
		os.Args = []string{"modelpatch"}
		main()
		return
	}
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer writer.Close()
	if err = reader.Close(); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestCLIClosedStdoutProcess$")
	cmd.Env = append(os.Environ(), child+"=1")
	cmd.Stdin = bytes.NewReader(input(t))
	cmd.Stdout = writer
	var diag bytes.Buffer
	cmd.Stderr = &diag
	cmd.WaitDelay = time.Second
	err = cmd.Run()
	if ctx.Err() != nil {
		t.Fatal("closed stdout process deadline", ctx.Err())
	}
	var exit *exec.ExitError
	if !errors.As(err, &exit) || exit.ExitCode() != 3 || diag.String() != "model patch output failed\n" {
		t.Fatalf("closed stdout: got %v / %q; want exit 3 and fixed output diagnostic", err, diag.String())
	}
}
