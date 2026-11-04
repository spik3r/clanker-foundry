package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const cliABC = "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad"

func cliDigest(raw []byte) string {
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}

func cliFixture(t *testing.T) (string, []byte, []string) {
	t.Helper()
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "input"), []byte("abc"), 0600); err != nil {
		t.Fatal(err)
	}
	raw := []byte(`{"formatVersion":"model-builder-inputs/v1","artifacts":[{"id":"source","role":"requirements","path":"input","sha256":"` + cliABC + `"}]}`)
	return root, raw, []string{"--root", root, "--manifest-sha256", cliDigest(raw)}
}

func cliReport(t *testing.T, raw []byte, pin, outcome, status string) {
	t.Helper()
	if len(raw) == 0 || raw[len(raw)-1] != '\n' {
		t.Fatal("missing newline-terminated JSON report")
	}
	var got map[string]any
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	if err := decoder.Decode(&got); err != nil {
		t.Fatal(err)
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		t.Fatal("trailing stdout", err)
	}
	for key, value := range map[string]any{
		"formatVersion": "model-builder-input-report/v1", "manifestSHA256": pin,
		"outcome": outcome, "executionStatus": "not-run", "reviewStatus": "not-established", "authority": "none",
	} {
		if got[key] != value {
			t.Fatalf("%s = %#v, want %#v", key, got[key], value)
		}
	}
	artifacts, ok := got["artifacts"].([]any)
	if !ok || len(artifacts) != 1 {
		t.Fatalf("artifacts: %#v", got["artifacts"])
	}
	artifact, ok := artifacts[0].(map[string]any)
	if !ok || artifact["id"] != "source" || artifact["role"] != "requirements" || artifact["path"] != "input" || artifact["expectedSHA256"] != cliABC || artifact["status"] != status {
		t.Fatalf("artifact: %#v", artifacts[0])
	}
	if status == "matched" && (artifact["observedSHA256"] != cliABC || artifact["observedBytes"] != json.Number("3")) {
		t.Fatal("matched byte observation", artifact)
	}
	if status == "missing" && (artifact["observedSHA256"] != nil || artifact["observedBytes"] != nil) {
		t.Fatal("missing file has an observation", artifact)
	}
}

func TestCLIExactArgumentsAndExitCodes(t *testing.T) {
	root, raw, args := cliFixture(t)
	var first []byte
	for _, flags := range [][]string{args, {args[2], args[3], args[0], args[1]}} {
		var out, diag bytes.Buffer
		if code := run(flags, bytes.NewReader(raw), &out, &diag); code != 0 || diag.Len() != 0 {
			t.Fatalf("match: exit=%d diag=%q", code, diag.String())
		}
		cliReport(t, out.Bytes(), cliDigest(raw), "matched", "matched")
		if first == nil {
			first = bytes.Clone(out.Bytes())
		} else if !bytes.Equal(first, out.Bytes()) {
			t.Fatal("argument pair order changed report bytes")
		}
	}
	if err := os.WriteFile(filepath.Join(root, "input"), []byte("different"), 0600); err != nil {
		t.Fatal(err)
	}
	var out, diag bytes.Buffer
	if code := run(args, bytes.NewReader(raw), &out, &diag); code != 1 {
		t.Fatalf("hash mismatch exit=%d diag=%q", code, diag.String())
	}
	cliReport(t, out.Bytes(), cliDigest(raw), "blocked", "hash-mismatch")
	if err := os.Remove(filepath.Join(root, "input")); err != nil {
		t.Fatal(err)
	}
	out.Reset()
	diag.Reset()
	if code := run(args, bytes.NewReader(raw), &out, &diag); code != 1 {
		t.Fatalf("missing input exit=%d diag=%q", code, diag.String())
	}
	cliReport(t, out.Bytes(), cliDigest(raw), "blocked", "missing")
}

func TestCLIInvalidArgumentsAndRoots(t *testing.T) {
	root, raw, args := cliFixture(t)
	pin := args[3]
	cases := [][]string{
		nil, {}, {"--help"}, {"--version"}, {"--root", root}, {"--manifest-sha256", pin},
		{"--root", root, "--manifest-sha256"},
		{"--root", root, "--root", root},
		{"--manifest-sha256", pin, "--manifest-sha256", pin},
		{"--root", root, "--manifest-sha256", pin, "--root", root},
		{"--root", root, "--manifest-sha256", pin, "extra"},
		{"--root=" + root, "--manifest-sha256=" + pin},
		{"root", root, "manifest-sha256", pin},
		{"--unknown", root, "--manifest-sha256", pin},
		{"--root", "", "--manifest-sha256", pin},
		{"--root", ".", "--manifest-sha256", pin},
		{"--root", "PRIVATE_RELATIVE_PATH", "--manifest-sha256", pin},
		{"--root", filepath.Join(root, "PRIVATE_MISSING_ROOT"), "--manifest-sha256", pin},
		{"--root", filepath.Join(root, "input"), "--manifest-sha256", pin},
		{"--root", root, "--manifest-sha256", ""},
		{"--root", root, "--manifest-sha256", strings.ToUpper(pin)},
		{"--root", root, "--manifest-sha256", strings.Repeat("0", 63)},
		{"--root", root, "--manifest-sha256", strings.Repeat("g", 64)},
	}
	for _, flags := range cases {
		var out, diag bytes.Buffer
		code := run(flags, bytes.NewReader(raw), &out, &diag)
		if code != 2 || out.Len() != 0 || diag.Len() == 0 {
			t.Fatalf("args=%q: exit=%d out=%q diag=%q", flags, code, out.String(), diag.String())
		}
		if strings.Contains(diag.String(), root) || strings.Contains(diag.String(), "PRIVATE_") {
			t.Fatalf("private argument leaked: %q", diag.String())
		}
	}
}

func TestCLIRejectsManifestWithoutReport(t *testing.T) {
	root, valid, _ := cliFixture(t)
	cases := [][]byte{nil, []byte("PRIVATE_MANIFEST_CONTENT"), []byte(`{"PRIVATE_MANIFEST_CONTENT":true}`), append(bytes.Clone(valid), []byte(` {}`)...), append(bytes.Clone(valid), bytes.Repeat([]byte(" "), (64<<10)-len(valid)+1)...)}
	for _, raw := range cases {
		var out, diag bytes.Buffer
		code := run([]string{"--root", root, "--manifest-sha256", cliDigest(raw)}, bytes.NewReader(raw), &out, &diag)
		if code != 2 || out.Len() != 0 || diag.Len() == 0 || strings.Contains(diag.String(), "PRIVATE_") || strings.Contains(diag.String(), root) {
			t.Fatalf("invalid manifest: exit=%d out=%q diag=%q", code, out.String(), diag.String())
		}
	}
	var out, diag bytes.Buffer
	if code := run([]string{"--root", root, "--manifest-sha256", strings.Repeat("0", 64)}, bytes.NewReader(valid), &out, &diag); code != 2 || out.Len() != 0 {
		t.Fatalf("stale manifest pin: exit=%d out=%q", code, out.String())
	}
}

type cliFailedReader struct{ data []byte }

func (r *cliFailedReader) Read(p []byte) (int, error) {
	n := copy(p, r.data)
	r.data = r.data[n:]
	return n, errors.New("PRIVATE_INPUT_IO_DETAILS")
}

func TestCLIInputIOFailuresAreExitFour(t *testing.T) {
	_, valid, args := cliFixture(t)
	for _, partial := range [][]byte{nil, []byte(`{"formatVersion":`), valid} {
		var out, diag bytes.Buffer
		code := run(args, &cliFailedReader{bytes.Clone(partial)}, &out, &diag)
		if code != 4 || out.Len() != 0 || diag.Len() == 0 || strings.Contains(diag.String(), "PRIVATE_") {
			t.Fatalf("input I/O: exit=%d out=%q diag=%q", code, out.String(), diag.String())
		}
	}
}

type cliFailedWriter struct{ mode string }

func (w cliFailedWriter) Write(p []byte) (int, error) {
	switch w.mode {
	case "zero":
		return 0, nil
	case "short":
		return len(p) - 1, nil
	case "partial-error":
		return len(p) / 2, errors.New("PRIVATE_OUTPUT_IO_DETAILS")
	default:
		return len(p), errors.New("PRIVATE_OUTPUT_IO_DETAILS")
	}
}

func TestCLIOutputErrorsOverrideMatchedAndBlockedExit(t *testing.T) {
	root, raw, args := cliFixture(t)
	for _, state := range []string{"matched", "blocked"} {
		if state == "blocked" {
			if err := os.Remove(filepath.Join(root, "input")); err != nil {
				t.Fatal(err)
			}
		}
		for _, mode := range []string{"zero", "short", "partial-error", "complete-error"} {
			var diag bytes.Buffer
			code := run(args, bytes.NewReader(raw), cliFailedWriter{mode}, &diag)
			if code != 3 || diag.Len() == 0 || strings.Contains(diag.String(), "PRIVATE_") || strings.Contains(diag.String(), root) {
				t.Fatalf("%s/%s: exit=%d diag=%q", state, mode, code, diag.String())
			}
		}
	}
}

func TestCLIClosedStdoutProcess(t *testing.T) {
	const child = "FOUNDRY_BUILDER_PREFLIGHT_CLOSED_STDOUT_CHILD"
	if os.Getenv(child) == "1" {
		os.Args = []string{"modelbuilderpreflight", "--root", os.Getenv("FOUNDRY_BUILDER_PREFLIGHT_ROOT"), "--manifest-sha256", os.Getenv("FOUNDRY_BUILDER_PREFLIGHT_PIN")}
		main()
		return
	}
	root, raw, _ := cliFixture(t)
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer writer.Close()
	if err := reader.Close(); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestCLIClosedStdoutProcess$")
	cmd.Env = append(os.Environ(), child+"=1", "FOUNDRY_BUILDER_PREFLIGHT_ROOT="+root, "FOUNDRY_BUILDER_PREFLIGHT_PIN="+cliDigest(raw))
	cmd.Stdin = bytes.NewReader(raw)
	cmd.Stdout = writer
	var diag bytes.Buffer
	cmd.Stderr = &diag
	cmd.WaitDelay = time.Second
	err = cmd.Run()
	if ctx.Err() != nil {
		t.Fatal("closed-pipe child exceeded its deadline", ctx.Err())
	}
	var exit *exec.ExitError
	if !errors.As(err, &exit) || exit.ExitCode() != 3 || diag.Len() == 0 || strings.Contains(diag.String(), root) {
		t.Fatalf("closed pipe: error=%v diag=%q; want exit 3 without signal termination", err, diag.String())
	}
}
