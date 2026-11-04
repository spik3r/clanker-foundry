package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/builderadmission"
)

func fixture(t *testing.T) (string, string) {
	t.Helper()
	root := t.TempDir()
	for _, name := range []string{"README.md", "go.mod", "contract.go", "acceptance_test.go", "rubric.md"} {
		b, err := os.ReadFile(filepath.Join("../../../../examples/model-builder-go/acceptance", name))
		if err != nil {
			t.Fatal(err)
		}
		if os.WriteFile(filepath.Join(root, name), b, 0600) != nil {
			t.Fatal("copy fixture")
		}
	}
	text := []byte("opaque candidate bytes, not Go code")
	if os.WriteFile(filepath.Join(root, "candidate.go"), text, 0600) != nil {
		t.Fatal("candidate fixture")
	}
	sum := sha256.Sum256(text)
	return root, hex.EncodeToString(sum[:])
}

type shortWriter struct{}

func (shortWriter) Write(b []byte) (int, error) { return len(b) - 1, nil }
func TestRunBoundaries(t *testing.T) {
	root, pin := fixture(t)
	args := []string{"check", "--root", root, "--candidate-sha256", pin}
	var out, diag bytes.Buffer
	want, err := builderadmission.Check(root, pin)
	if err != nil {
		t.Fatal(err)
	}
	if run(args, &out, &diag) != 0 || diag.Len() != 0 || !bytes.Equal(out.Bytes(), want) {
		t.Fatal("complete native admission parity")
	}
	out.Reset()
	diag.Reset()
	reordered := []string{"check", "--candidate-sha256", pin, "--root", root}
	if run(reordered, &out, &diag) != 0 || !bytes.Equal(out.Bytes(), want) {
		t.Fatal("flag order")
	}
	for _, bad := range [][]string{nil, {"execute"}, {"check", "--root", root, "--root", root}, {"check", "--root", root, "--unknown", pin}, {"check", "--root", root, "--candidate-sha256", "private invalid pin"}} {
		out.Reset()
		diag.Reset()
		if run(bad, &out, &diag) != 2 || out.Len() != 0 || bytes.Contains(diag.Bytes(), []byte("private invalid")) {
			t.Fatal("argument refusal")
		}
	}
	if run(args, shortWriter{}, io.Discard) != 3 {
		t.Fatal("short write")
	}
	if os.WriteFile(filepath.Join(root, "unknown.go"), nil, 0600) != nil {
		t.Fatal("extra")
	}
	out.Reset()
	diag.Reset()
	if run(args, &out, &diag) != 1 || diag.Len() != 0 || out.Len() == 0 {
		t.Fatal("blocked report delivery")
	}
}
func TestActualCLIAndClosedStdout(t *testing.T) {
	root, pin := fixture(t)
	ctx, stop := context.WithTimeout(context.Background(), 90*time.Second)
	defer stop()
	name := "modelbuilderadmission"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	binary := filepath.Join(t.TempDir(), name)
	module, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	build := exec.CommandContext(ctx, filepath.Join(runtime.GOROOT(), "bin", "go"), "build", "-o", binary, "./cmd/modelbuilderadmission")
	build.Dir = module
	build.WaitDelay = time.Second
	if b, err := build.CombinedOutput(); err != nil {
		t.Fatalf("tool build: %v %s", err, b)
	}
	args := []string{"check", "--root", root, "--candidate-sha256", pin}
	for _, blocked := range []bool{false, true} {
		if blocked {
			if os.WriteFile(filepath.Join(root, "extra.go"), []byte("inert extra"), 0600) != nil {
				t.Fatal("extra")
			}
		}
		want, e := builderadmission.Check(root, pin)
		if e != nil {
			t.Fatal(e)
		}
		cmd := exec.CommandContext(ctx, binary, args...)
		cmd.WaitDelay = time.Second
		var out, diag bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &diag
		e = cmd.Run()
		if !blocked && e != nil {
			t.Fatal(e)
		}
		if blocked {
			var exit *exec.ExitError
			if !errors.As(e, &exit) || exit.ExitCode() != 1 {
				t.Fatal("blocked exit", e)
			}
		}
		if diag.Len() != 0 || !bytes.Equal(out.Bytes(), want) {
			t.Fatal("actual full report", diag.String())
		}
	}
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer writer.Close()
	if reader.Close() != nil {
		t.Fatal("pipe")
	}
	cmd := exec.CommandContext(ctx, binary, args...)
	cmd.WaitDelay = time.Second
	cmd.Stdout = writer
	var diag bytes.Buffer
	cmd.Stderr = &diag
	err = cmd.Run()
	var exit *exec.ExitError
	if ctx.Err() != nil || !errors.As(err, &exit) || exit.ExitCode() != 3 || diag.String() != "modelbuilderadmission: output failed\n" {
		t.Fatalf("closed pipe: %v %q", err, diag.String())
	}
}
