package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestInvalidRequests(t *testing.T) {
	for _, input := range []string{"{}", `{"candidate":"","candidate":"","suite":""}`, `{"unknown":true}`, "{"} {
		var out, diag bytes.Buffer
		if run([]string{"export"}, bytes.NewBufferString(input), &out, &diag) == 0 || out.Len() != 0 {
			t.Fatal("invalid export accepted")
		}
	}
	var out, diag bytes.Buffer
	if run([]string{"verify"}, bytes.NewBufferString(`{"archive":"","candidateSHA256":"","suiteSHA256":""}`), &out, &diag) == 0 || out.Len() != 0 {
		t.Fatal("invalid verify accepted")
	}
}
func TestRejectsSymlinkSource(t *testing.T) {
	root := t.TempDir()
	if e := os.WriteFile(filepath.Join(root, "go.mod"), []byte("placeholder"), 0644); e != nil {
		t.Fatal(e)
	}
	if e := os.Symlink("/tmp", filepath.Join(root, "model")); e != nil {
		t.Fatal(e)
	}
	old, e := os.Getwd()
	if e != nil {
		t.Fatal(e)
	}
	if e = os.Chdir(root); e != nil {
		t.Fatal(e)
	}
	defer os.Chdir(old)
	if _, e = readSources(); e == nil {
		t.Fatal("accepted missing/symlink inventory")
	}
}

func TestExportVerifyRoundTrip(t *testing.T) {
	old, e := os.Getwd()
	if e != nil {
		t.Fatal(e)
	}
	if e = os.Chdir("../.."); e != nil {
		t.Fatal(e)
	}
	defer os.Chdir(old)
	candidate, e := os.ReadFile("examples/lifecycle.json")
	if e != nil {
		t.Fatal(e)
	}
	suite, e := os.ReadFile("scenarios/testdata/lifecycle-v1.json")
	if e != nil {
		t.Fatal(e)
	}
	req, e := json.Marshal(map[string]string{"candidate": base64.StdEncoding.EncodeToString(candidate), "suite": base64.StdEncoding.EncodeToString(suite)})
	if e != nil {
		t.Fatal(e)
	}
	var archive, diag bytes.Buffer
	if run([]string{"export"}, bytes.NewReader(req), &archive, &diag) != 0 {
		t.Fatalf("export failed: %s", diag.String())
	}
	c := sha256.Sum256(candidate)
	s := sha256.Sum256(suite)
	req, e = json.Marshal(map[string]string{"archive": base64.StdEncoding.EncodeToString(archive.Bytes()), "candidateSHA256": hex.EncodeToString(c[:]), "suiteSHA256": hex.EncodeToString(s[:])})
	if e != nil {
		t.Fatal(e)
	}
	var out bytes.Buffer
	diag.Reset()
	if run([]string{"verify"}, bytes.NewReader(req), &out, &diag) != 0 {
		t.Fatalf("verify failed: %s", diag.String())
	}
	if run([]string{"verify"}, bytes.NewReader(req), shortWriter{}, &diag) == 0 {
		t.Fatal("accepted short write")
	}
	if !strings.Contains(out.String(), "unvalidated and unexecuted") {
		t.Fatal("missing status")
	}
}

type shortWriter struct{}

func (shortWriter) Write(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	return len(p) - 1, nil
}
func TestRejectsNonregularAndOversizedSource(t *testing.T) {
	old, e := os.Getwd()
	if e != nil {
		t.Fatal(e)
	}
	defer os.Chdir(old)
	for _, kind := range []string{"directory", "oversized"} {
		t.Run(kind, func(t *testing.T) {
			root := t.TempDir()
			p := filepath.Join(root, "go.mod")
			if kind == "directory" {
				if e := os.Mkdir(p, 0755); e != nil {
					t.Fatal(e)
				}
			} else if e := os.WriteFile(p, make([]byte, (1<<20)+1), 0644); e != nil {
				t.Fatal(e)
			}
			if e := os.Chdir(root); e != nil {
				t.Fatal(e)
			}
			defer os.Chdir(old)
			if _, e := readSources(); e == nil {
				t.Fatal("accepted invalid source")
			}
		})
	}
}

// Closed stdout must preserve run's existing exit-2 failure contract for both
// archive output and verification text. Only this trusted test binary is run.
func TestCLIClosedStdoutProcesses(t *testing.T) {
	const child = "FOUNDRY_MODELQUALIFY_CLOSED_STDOUT_CHILD"
	if mode := os.Getenv(child); mode != "" {
		os.Args = []string{"modelqualify", mode}
		main()
		return
	}
	old, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err = os.Chdir("../.."); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(old)
	root, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	candidate, err := os.ReadFile("examples/lifecycle.json")
	if err != nil {
		t.Fatal(err)
	}
	suite, err := os.ReadFile("scenarios/testdata/lifecycle-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	exportRequest, err := json.Marshal(map[string]string{"candidate": base64.StdEncoding.EncodeToString(candidate), "suite": base64.StdEncoding.EncodeToString(suite)})
	if err != nil {
		t.Fatal(err)
	}
	var archive, diag bytes.Buffer
	if code := run([]string{"export"}, bytes.NewReader(exportRequest), &archive, &diag); code != 0 {
		t.Fatal("prepare archive", code, diag.String())
	}
	c, s := sha256.Sum256(candidate), sha256.Sum256(suite)
	verifyRequest, err := json.Marshal(map[string]string{"archive": base64.StdEncoding.EncodeToString(archive.Bytes()), "candidateSHA256": hex.EncodeToString(c[:]), "suiteSHA256": hex.EncodeToString(s[:])})
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		mode    string
		request []byte
	}{{"export", exportRequest}, {"verify", verifyRequest}} {
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
			cmd.Env = append(os.Environ(), child+"="+tc.mode)
			cmd.Dir, cmd.Stdin, cmd.Stdout = root, bytes.NewReader(tc.request), writer
			var diagnostic bytes.Buffer
			cmd.Stderr = &diagnostic
			cmd.WaitDelay = time.Second
			err = cmd.Run()
			if ctx.Err() != nil {
				t.Fatal("closed stdout process deadline", ctx.Err())
			}
			var exit *exec.ExitError
			if !errors.As(err, &exit) || exit.ExitCode() != 2 || diagnostic.String() != "modelqualify: invalid request or project\n" {
				t.Fatalf("closed stdout got %v / %q; want exit2 and fixed failure diagnostic", err, diagnostic.String())
			}
		})
	}
}
