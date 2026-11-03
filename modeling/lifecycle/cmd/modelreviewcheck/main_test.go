package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func validRequest(t *testing.T) []byte {
	t.Helper()
	b, err := os.ReadFile("../../firstdesignreview/testdata/oracle.json")
	if err != nil {
		t.Fatal(err)
	}
	var f struct {
		Cases []struct {
			Request  string `json:"request"`
			Expected struct {
				Valid bool `json:"valid"`
			} `json:"expected"`
		} `json:"cases"`
	}
	if err := json.Unmarshal(b, &f); err != nil || len(f.Cases) == 0 || !f.Cases[0].Expected.Valid {
		t.Fatalf("frozen fixture: %v", err)
	}
	return []byte(f.Cases[0].Request)
}

type failingReader struct{}

func (failingReader) Read([]byte) (int, error) { return 0, errors.New("private read detail") }

type shortWriter struct{ prefix bytes.Buffer }

func (w *shortWriter) Write(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	return w.prefix.Write(p[:len(p)-1])
}

func TestRunFailuresAndPartialOutput(t *testing.T) {
	request := validRequest(t)
	for _, tc := range []struct {
		name string
		args []string
		in   io.Reader
	}{
		{"execute is not an action", []string{"execute"}, bytes.NewReader(request)},
		{"unknown flag", []string{"--help"}, bytes.NewReader(request)},
		{"bad input", nil, strings.NewReader(`{"private-secret":"do-not-echo"}`)},
		{"input read failure", nil, failingReader{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var out, diag bytes.Buffer
			if code := run(tc.args, tc.in, &out, &diag); code != 2 || out.Len() != 0 {
				t.Fatalf("code=%d stdout=%q stderr=%q", code, out.Bytes(), diag.Bytes())
			}
			if strings.Contains(diag.String(), "private") || strings.Contains(diag.String(), "do-not-echo") {
				t.Fatalf("diagnostic echoed input/error: %q", diag.String())
			}
		})
	}
	var diag bytes.Buffer
	w := &shortWriter{}
	if code := run(nil, bytes.NewReader(request), w, &diag); code != 3 || w.prefix.Len() == 0 || diag.String() != "modelreviewcheck: report delivery failed\n" {
		t.Fatalf("short write code=%d prefix=%d diag=%q", code, w.prefix.Len(), diag.String())
	}
}

// The helper process exercises the CLI entry boundary from a caller-selected
// empty directory. Marker executables detect only these named PATH lookups; they
// are not a claim about all possible operating-system effects.
func TestReviewHelper(t *testing.T) {
	if os.Getenv("REVIEW_CHECK_TEST_HELPER") == "1" {
		os.Exit(run(nil, os.Stdin, os.Stdout, os.Stderr))
	}
}

func TestWorksFromEmptyDirectoryWithoutNamedChildTools(t *testing.T) {
	request := validRequest(t)
	tmp := t.TempDir()
	cwd := filepath.Join(tmp, "cwd")
	path := filepath.Join(tmp, "path")
	if err := os.Mkdir(cwd, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(path, 0700); err != nil {
		t.Fatal(err)
	}
	marker := filepath.Join(tmp, "invoked")
	for _, name := range []string{"go", "sh", "modelplay"} {
		script := "#!/bin/sh\nprintf invoked > \"$REVIEW_CHECK_TEST_MARKER\"\nexit 99\n"
		if err := os.WriteFile(filepath.Join(path, name), []byte(script), 0700); err != nil {
			t.Fatal(err)
		}
	}
	cmd := exec.Command(os.Args[0], "-test.run=^TestReviewHelper$")
	cmd.Dir = cwd
	cmd.Stdin = bytes.NewReader(request)
	for _, e := range os.Environ() {
		if !strings.HasPrefix(e, "PATH=") && !strings.HasPrefix(e, "REVIEW_CHECK_TEST_HELPER=") && !strings.HasPrefix(e, "REVIEW_CHECK_TEST_MARKER=") {
			cmd.Env = append(cmd.Env, e)
		}
	}
	cmd.Env = append(cmd.Env, "PATH="+path, "REVIEW_CHECK_TEST_HELPER=1", "REVIEW_CHECK_TEST_MARKER="+marker)
	var out, diag bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &diag
	if err := cmd.Run(); err != nil {
		t.Fatalf("helper: %v stderr=%q", err, diag.String())
	}
	var r struct {
		ExecutionStatus string `json:"executionStatus"`
		ReviewStatus    string `json:"reviewStatus"`
		Authority       string `json:"authority"`
	}
	if err := json.Unmarshal(out.Bytes(), &r); err != nil || r.ExecutionStatus != "not-run" || r.ReviewStatus != "not-established" || r.Authority != "none" {
		t.Fatalf("report: %+v %v", r, err)
	}
	if _, err := os.Stat(marker); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("named child marker exists or cannot be inspected: %v", err)
	}
	entries, err := os.ReadDir(cwd)
	if err != nil || len(entries) != 0 {
		t.Fatalf("empty cwd changed: %v %v", entries, err)
	}
}
