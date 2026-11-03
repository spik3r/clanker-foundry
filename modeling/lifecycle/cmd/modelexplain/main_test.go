package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/compiler"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/scenarioexplain"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"testing"
	"time"
)

type shortWriter struct{}

func (shortWriter) Write(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	return len(p) - 1, nil
}

type failingReader struct{}

func (failingReader) Read([]byte) (int, error) { return 0, errors.New("PRIVATE_READER_ERROR") }
func TestCLI(t *testing.T) {
	if runtime.Version() != compiler.ToolchainVersion {
		var out, diag bytes.Buffer
		if run(nil, strings.NewReader("anything"), &out, &diag) != 2 || out.Len() != 0 || diag.String() != "modelexplain: invalid input\n" {
			t.Fatal("wrong toolchain not rejected")
		}
		t.Skip("remaining CLI cases require Go 1.27.3")
	}
	model, e := os.ReadFile("../../examples/lifecycle.json")
	if e != nil {
		t.Fatal(e)
	}
	raw, e := os.ReadFile("../../scenarioexplain/testdata/vectors.json")
	if e != nil {
		t.Fatal(e)
	}
	var v struct {
		Cases []struct {
			Report json.RawMessage `json:"report"`
		} `json:"cases"`
	}
	if json.Unmarshal(raw, &v) != nil || len(v.Cases) == 0 {
		t.Fatal("vectors")
	}
	req := []byte(`{"formatVersion":"lifecycle-scenario-explain-request/v1","model":` + string(model) + `,"report":` + string(v.Cases[0].Report) + `}`)
	var out, diag bytes.Buffer
	if run(nil, bytes.NewReader(req), &out, &diag) != 0 || out.Len() == 0 || diag.Len() != 0 {
		t.Fatal("valid request failed")
	}
	diag.Reset()
	if run(nil, bytes.NewReader(req), shortWriter{}, &diag) != 3 || diag.String() != "modelexplain: output failed\n" {
		t.Fatal("short write not reported")
	}
	for _, input := range []string{"{", strings.Repeat("x", scenarioexplain.MaxRequestBytes+1)} {
		out.Reset()
		diag.Reset()
		if run(nil, strings.NewReader(input), &out, &diag) != 2 || out.Len() != 0 || diag.String() != "modelexplain: invalid input\n" {
			t.Fatal("invalid input behavior")
		}
	}
	out.Reset()
	diag.Reset()
	if run(nil, failingReader{}, &out, &diag) != 2 || out.Len() != 0 || strings.Contains(diag.String(), "PRIVATE_READER_ERROR") {
		t.Fatal("reader error leaked")
	}
	out.Reset()
	diag.Reset()
	if run([]string{"unexpected"}, bytes.NewReader(req), &out, &diag) != 2 || out.Len() != 0 {
		t.Fatal("argument accepted")
	}
}

// Exercise main with an actual closed stdout pipe, not only a fake writer.
// This launches only the trusted test binary with a bounded wait.
func TestCLIClosedStdoutProcess(t *testing.T) {
	const child = "FOUNDRY_MODELEXPLAIN_CLOSED_STDOUT_CHILD"
	if os.Getenv(child) == "1" {
		os.Args = []string{"modelexplain"}
		main()
		return
	}
	model, err := os.ReadFile("../../examples/lifecycle.json")
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile("../../scenarioexplain/testdata/vectors.json")
	if err != nil {
		t.Fatal(err)
	}
	var vectors struct {
		Cases []struct {
			Report json.RawMessage `json:"report"`
		} `json:"cases"`
	}
	if err = json.Unmarshal(raw, &vectors); err != nil || len(vectors.Cases) == 0 {
		t.Fatal("frozen vectors", err)
	}
	request := []byte(`{"formatVersion":"lifecycle-scenario-explain-request/v1","model":` + string(model) + `,"report":` + string(vectors.Cases[0].Report) + `}`)
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
	cmd.Stdin = bytes.NewReader(request)
	cmd.Stdout = writer
	var diag bytes.Buffer
	cmd.Stderr = &diag
	cmd.WaitDelay = time.Second
	err = cmd.Run()
	if ctx.Err() != nil {
		t.Fatal("closed stdout process deadline", ctx.Err())
	}
	var exit *exec.ExitError
	if !errors.As(err, &exit) || exit.ExitCode() != 3 || diag.String() != "modelexplain: output failed\n" {
		t.Fatalf("closed stdout: got %v / %q; want exit 3 and fixed output diagnostic", err, diag.String())
	}
}
