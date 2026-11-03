package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/designchange"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/qualification"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func input(t *testing.T) ([]byte, designchange.Expected) {
	t.Helper()
	r, e := os.ReadFile("../../designchange/testdata/request.json")
	if e != nil {
		t.Fatal(e)
	}
	s, e := os.ReadFile("../../designchange/testdata/suite.json")
	if e != nil {
		t.Fatal(e)
	}
	sources := []map[string]string{}
	for _, n := range qualification.SourceNames() {
		b, e := os.ReadFile(filepath.Join("../..", n))
		if e != nil {
			t.Fatal(e)
		}
		sources = append(sources, map[string]string{"name": n, "dataBase64": base64.StdEncoding.EncodeToString(b)})
	}
	var expected struct {
		External designchange.Expected `json:"externalExpected"`
	}
	b, e := os.ReadFile("../../designchange/testdata/expected.json")
	if e != nil || json.Unmarshal(b, &expected) != nil {
		t.Fatal("expected")
	}
	raw, _ := json.Marshal(map[string]any{"formatVersion": designchange.FormatVersion, "operation": "build", "requestBase64": base64.StdEncoding.EncodeToString(r), "suiteBase64": base64.StdEncoding.EncodeToString(s), "sources": sources, "archiveBase64": nil, "expected": nil})
	return raw, expected.External
}

type interrupted struct{ once bool }

func (r *interrupted) Read(p []byte) (int, error) {
	if r.once {
		return 0, io.ErrUnexpectedEOF
	}
	r.once = true
	return copy(p, []byte(`{"formatVersion":`)), nil
}

type failedReader struct{}

func (failedReader) Read([]byte) (int, error) { return 0, errors.New("invented private detail") }

type shortWriter struct {
	fail  bool
	calls int
}

func (w *shortWriter) Write(p []byte) (int, error) {
	w.calls++
	if w.fail {
		return len(p) / 2, io.ErrClosedPipe
	}
	return len(p) / 2, nil
}
func TestActualCLIProjectAndReceipt(t *testing.T) {
	raw, expected := input(t)
	var out, diag bytes.Buffer
	if code := run(nil, bytes.NewReader(raw), &out, &diag); code != 0 {
		t.Fatal(code, diag.String())
	}
	archive := append([]byte{}, out.Bytes()...)
	if e := designchange.Verify(archive, expected); e != nil {
		t.Fatal(e)
	}
	out.Reset()
	if code := run(nil, bytes.NewReader(raw), &out, &diag); code != 0 || !bytes.Equal(archive, out.Bytes()) {
		t.Fatal("replay")
	}
	verify, _ := json.Marshal(map[string]any{"formatVersion": designchange.FormatVersion, "operation": "verify", "requestBase64": nil, "suiteBase64": nil, "sources": nil, "archiveBase64": base64.StdEncoding.EncodeToString(archive), "expected": expected})
	out.Reset()
	if code := run(nil, bytes.NewReader(verify), &out, &diag); code != 0 {
		t.Fatal(code)
	}
	if !strings.Contains(out.String(), `"suiteBindingStatus":"not inspected"`) || !strings.Contains(out.String(), `"suiteStatus":"unvalidated and unexecuted"`) {
		t.Fatal("status missing")
	}
}
func TestCLIRejectionsAndShortWrites(t *testing.T) {
	raw, _ := input(t)
	for _, r := range []io.Reader{failedReader{}, &interrupted{}, strings.NewReader("invented private detail"), strings.NewReader(strings.Repeat(" ", designchange.MaxCarrierBytes+1))} {
		var out, diag bytes.Buffer
		if code := run(nil, r, &out, &diag); code != 2 || out.Len() != 0 || strings.Contains(diag.String(), "invented private detail") {
			t.Fatal("read/reject leaked output")
		}
	}
	var out, diag bytes.Buffer
	if code := run([]string{"/tmp/path"}, bytes.NewReader(raw), &out, &diag); code != 2 || out.Len() != 0 {
		t.Fatal("arg accepted")
	}
	for _, fail := range []bool{false, true} {
		w := &shortWriter{fail: fail}
		if code := run(nil, bytes.NewReader(raw), w, &diag); code != 3 || w.calls != 1 {
			t.Fatal("write failure ignored")
		}
	}
}

// Exercise actual stdout file descriptor behavior, not only an injected writer.
// Both operations keep the original exit-3 output-failure contract.
func TestCLIClosedStdoutProcesses(t *testing.T) {
	const child = "FOUNDRY_MODELDESIGNCHANGE_CLOSED_STDOUT_CHILD"
	if os.Getenv(child) == "1" {
		os.Args = []string{"modeldesignchange"}
		main()
		return
	}
	build, expected := input(t)
	var archive, diag bytes.Buffer
	if code := run(nil, bytes.NewReader(build), &archive, &diag); code != 0 {
		t.Fatal("prepare archive", code, diag.String())
	}
	verify, err := json.Marshal(map[string]any{"formatVersion": designchange.FormatVersion,
		"operation": "verify", "requestBase64": nil, "suiteBase64": nil, "sources": nil,
		"archiveBase64": base64.StdEncoding.EncodeToString(archive.Bytes()), "expected": expected})
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		mode    string
		request []byte
	}{{"build", build}, {"verify", verify}} {
		t.Run(tc.mode, func(t *testing.T) {
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
			cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestCLIClosedStdoutProcesses$")
			cmd.Env = append(os.Environ(), child+"=1")
			cmd.Stdin, cmd.Stdout = bytes.NewReader(tc.request), writer
			var diagnostic bytes.Buffer
			cmd.Stderr = &diagnostic
			cmd.WaitDelay = time.Second
			err = cmd.Run()
			if ctx.Err() != nil {
				t.Fatal("closed stdout process deadline", ctx.Err())
			}
			var exit *exec.ExitError
			if !errors.As(err, &exit) || exit.ExitCode() != 3 || diagnostic.String() != "design-change output failed; suite unvalidated and unexecuted, binding not inspected\n" {
				t.Fatalf("closed stdout got %v / %q; want exit3 and fixed failure diagnostic", err, diagnostic.String())
			}
		})
	}
}
