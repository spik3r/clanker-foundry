package main

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"errors"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/processview"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

type brokenReader struct{}

func (brokenReader) Read([]byte) (int, error) { return 0, errors.New("private I/O details") }

type shortWriter struct{}

func (shortWriter) Write(p []byte) (int, error) { return len(p) / 2, nil }
func TestRun(t *testing.T) {
	source, e := os.ReadFile("../../processview/testdata/model.json")
	if e != nil {
		t.Fatal(e)
	}
	for _, c := range []struct {
		name string
		args []string
		in   io.Reader
		out  io.Writer
		want int
	}{
		{"json", []string{"json"}, bytes.NewReader(source), &bytes.Buffer{}, 0},
		{"svg", []string{"svg"}, bytes.NewReader(source), &bytes.Buffer{}, 0},
		{"missing argument", nil, bytes.NewReader(source), &bytes.Buffer{}, 2},
		{"extra argument", []string{"svg", "file"}, bytes.NewReader(source), &bytes.Buffer{}, 2},
		{"invalid model", []string{"svg"}, strings.NewReader("{"), &bytes.Buffer{}, 2},
		{"oversized", []string{"json"}, io.LimitReader(strings.NewReader(strings.Repeat(" ", processview.MaxSourceBytes+1)), processview.MaxSourceBytes+1), &bytes.Buffer{}, 2},
		{"read failure", []string{"json"}, brokenReader{}, &bytes.Buffer{}, 2},
		{"short write", []string{"svg"}, bytes.NewReader(source), shortWriter{}, 3},
	} {
		t.Run(c.name, func(t *testing.T) {
			var diag bytes.Buffer
			status := run(c.args, c.in, c.out, &diag)
			if status != c.want {
				t.Fatalf("status %d want %d; %s", status, c.want, diag.String())
			}
			if strings.Contains(diag.String(), "private I/O details") {
				t.Fatal("leaked input error")
			}
			if c.want == 2 {
				if b, ok := c.out.(*bytes.Buffer); ok && b.Len() != 0 {
					t.Fatal("failure produced stdout")
				}
			}
			if c.want == 0 {
				b := c.out.(*bytes.Buffer).Bytes()
				if c.name == "json" && !json.Valid(b) {
					t.Fatal("invalid JSON")
				}
				if c.name == "svg" {
					dec := xml.NewDecoder(bytes.NewReader(b))
					for {
						_, e := dec.Token()
						if e == io.EOF {
							break
						}
						if e != nil {
							t.Fatal(e)
						}
					}
				}
			}
		})
	}
}

// Build the real CLI so process-level SIGPIPE behavior reaches main.
func TestCLIProcessOutput(t *testing.T) {
	source, err := os.ReadFile("../../processview/testdata/model.json")
	if err != nil {
		t.Fatal(err)
	}
	expected, err := processview.Build(source)
	if err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(t.TempDir(), "modelprocess")
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	build := exec.CommandContext(ctx, filepath.Join(runtime.GOROOT(), "bin", "go"), "build", "-o", binary, ".")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build: %v: %s", err, output)
	}
	for _, mode := range []string{"json", "svg"} {
		t.Run(mode, func(t *testing.T) {
			t.Run("complete", func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
				defer cancel()
				cmd := exec.CommandContext(ctx, binary, mode)
				cmd.Stdin = bytes.NewReader(source)
				var stdout, diagnostic bytes.Buffer
				cmd.Stdout, cmd.Stderr = &stdout, &diagnostic
				cmd.WaitDelay = time.Second
				if err := cmd.Run(); err != nil {
					t.Fatalf("complete output: %v / %q", err, diagnostic.String())
				}
				want := expected.JSON
				if mode == "svg" {
					want = expected.SVG
				}
				if !bytes.Equal(stdout.Bytes(), want) || diagnostic.Len() != 0 {
					t.Fatal("incomplete output or unexpected diagnostic")
				}
			})
			t.Run("closed_pipe", func(t *testing.T) {
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
				cmd := exec.CommandContext(ctx, binary, mode)
				cmd.Stdin = bytes.NewReader(source)
				cmd.Stdout = writer
				var diagnostic bytes.Buffer
				cmd.Stderr = &diagnostic
				cmd.WaitDelay = time.Second
				err = cmd.Run()
				if ctx.Err() != nil {
					t.Fatal("closed stdout deadline", ctx.Err())
				}
				var exit *exec.ExitError
				if !errors.As(err, &exit) || exit.ExitCode() != 3 || diagnostic.String() != "modelprocess: output write failed\n" {
					t.Fatalf("closed stdout: %v / %q; want exit 3 with fixed diagnostic", err, diagnostic.String())
				}
			})
		})
	}
}
