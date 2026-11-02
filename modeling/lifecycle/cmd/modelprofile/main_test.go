package main

import (
	"bytes"
	"context"
	"errors"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/compiler"
	"os"
	"os/exec"
	"runtime"
	"testing"
	"time"
)

type poisonReader struct{ reads *int }

func (p poisonReader) Read([]byte) (int, error) {
	*p.reads++
	return 0, errors.New("stdin must not be read")
}

type errorWriter struct{}

func (errorWriter) Write([]byte) (int, error) { return 0, errors.New("private output error") }

type shortWriter struct{}

func (shortWriter) Write(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	return len(p) - 1, nil
}
func TestCLI(t *testing.T) {
	var out, diag bytes.Buffer
	reads := 0
	if run([]string{"bad"}, poisonReader{&reads}, &out, &diag) != 2 || out.Len() != 0 || diag.String() != "modelprofile: unavailable\n" {
		t.Fatal("unsupported argument")
	}
	out.Reset()
	diag.Reset()
	if runtime.Version() != compiler.ToolchainVersion {
		if run(nil, poisonReader{&reads}, &out, &diag) != 2 || out.Len() != 0 {
			t.Fatal("wrong toolchain")
		}
		t.Skip("positive CLI cases require Go 1.27.3")
	}
	if run(nil, poisonReader{&reads}, &out, &diag) != 0 || out.Len() == 0 || diag.Len() != 0 || reads != 0 {
		t.Fatal("normal output or stdin read")
	}
	diag.Reset()
	if run(nil, poisonReader{&reads}, shortWriter{}, &diag) != 3 || diag.String() != "modelprofile: output failed\n" {
		t.Fatal("short write")
	}
	diag.Reset()
	if run(nil, poisonReader{&reads}, errorWriter{}, &diag) != 3 || diag.String() != "modelprofile: output failed\n" || reads != 0 {
		t.Fatal("output error handling")
	}

}

// This executes main with a real closed stdout descriptor, not a fake writer.
func TestCLIClosedStdoutProcess(t *testing.T) {
	const child = "FOUNDRY_MODELPROFILE_CLOSED_STDOUT_CHILD"
	if os.Getenv(child) == "1" {
		os.Args = []string{"modelprofile"}
		main()
		return
	}
	if runtime.Version() != compiler.ToolchainVersion {
		t.Skip("positive CLI case requires pinned Go toolchain")
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
	cmd.Stdout = writer
	var diagnostic bytes.Buffer
	cmd.Stderr = &diagnostic
	cmd.WaitDelay = time.Second
	err = cmd.Run()
	if ctx.Err() != nil {
		t.Fatal("closed stdout deadline", ctx.Err())
	}
	var exit *exec.ExitError
	if !errors.As(err, &exit) || exit.ExitCode() != 3 || diagnostic.String() != "modelprofile: output failed\n" {
		t.Fatalf("closed stdout got %v / %q; want exit3 with fixed output diagnostic", err, diagnostic.String())
	}
}
