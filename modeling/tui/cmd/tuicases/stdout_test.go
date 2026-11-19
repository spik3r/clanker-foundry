package main

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

type stdoutResultWriter struct {
	count     int
	err       error
	delivered bytes.Buffer
}

func (w *stdoutResultWriter) Write(p []byte) (int, error) {
	n := min(w.count, len(p))
	w.delivered.Write(p[:n])
	return n, w.err
}

func TestStdoutDelivery(t *testing.T) {
	for _, command := range []string{"validate", "check"} {
		t.Run(command, func(t *testing.T) {
			if command == "check" {
				skipShort(t)
			}
			args := []string{command, browserModel, browserCases}
			var complete, diagnostic bytes.Buffer
			if code := run(args, &complete, &diagnostic); code != 0 || complete.Len() == 0 || diagnostic.Len() != 0 {
				t.Fatalf("complete delivery: code=%d stdout=%q stderr=%q", code, complete.String(), diagnostic.String())
			}
			failure := errors.New("synthetic output failure")
			for _, tc := range []struct {
				name  string
				count int
				err   error
			}{
				{"error", 0, failure},
				{"short without error", 1, nil},
				{"partial with error", 1, failure},
				{"full count with error", complete.Len(), failure},
			} {
				t.Run(tc.name, func(t *testing.T) {
					w := &stdoutResultWriter{count: tc.count, err: tc.err}
					var stderr bytes.Buffer
					code := run(args, w, &stderr)
					if code != 1 || stderr.String() != "tuicases: output write failed\n" {
						t.Fatalf("code=%d stderr=%q; want exit 1 and fixed output-failure diagnostic", code, stderr.String())
					}
					if !bytes.Equal(w.delivered.Bytes(), complete.Bytes()[:tc.count]) {
						t.Fatalf("delivered prefix changed: %q", w.delivered.String())
					}
				})
			}
		})
	}
}

func TestCompiledClosedStdout(t *testing.T) {
	skipShort(t)
	binary := filepath.Join(t.TempDir(), "tuicases")
	if runtime.GOOS == "windows" {
		binary += ".exe"
	}
	ctx, cancel := context.WithTimeout(context.Background(), buildTimeout)
	defer cancel()
	build := exec.CommandContext(ctx, "go", "build", "-o", binary, ".")
	build.Env = append(os.Environ(), "GOWORK=off")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build command: %v\n%s", err, out)
	}
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer writer.Close()
	if err := reader.Close(); err != nil {
		t.Fatal(err)
	}
	rctx, rcancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer rcancel()
	cmd := exec.CommandContext(rctx, binary, "validate", browserModel, browserCases)
	cmd.Stdout = writer
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err = cmd.Run()
	var exit *exec.ExitError
	if !errors.As(err, &exit) || exit.ExitCode() != 1 || stderr.String() != "tuicases: output write failed\n" {
		t.Fatalf("closed stdout: err=%v stderr=%q; want exit 1 and fixed output-failure diagnostic", err, stderr.String())
	}
}
