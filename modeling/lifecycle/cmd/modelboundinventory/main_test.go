package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/boundinventory"
)

func fixture(t *testing.T, index int) []byte {
	t.Helper()
	b, e := os.ReadFile("../../boundinventory/testdata/observations.json")
	if e != nil {
		t.Fatal(e)
	}
	var d struct {
		Vectors []struct{ Model, Suite, Report string }
	}
	if json.Unmarshal(b, &d) != nil {
		t.Fatal("fixture")
	}
	v := d.Vectors[index]
	var out bytes.Buffer
	enc := json.NewEncoder(&out)
	enc.SetEscapeHTML(false)
	if enc.Encode(struct {
		FormatVersion string `json:"formatVersion"`
		Model         string `json:"model"`
		Suite         string `json:"suite"`
		Report        string `json:"report"`
	}{boundinventory.RequestVersion, v.Model, v.Suite, v.Report}) != nil {
		t.Fatal("fixture encode")
	}
	return out.Bytes()
}

type unreadable struct{ called bool }

func (r *unreadable) Read([]byte) (int, error) {
	r.called = true
	return 0, errors.New("PRIVATE read error")
}

type shortWriter struct{}

func (shortWriter) Write(b []byte) (int, error) { return len(b) - 1, nil }
func TestRunBoundaries(t *testing.T) {
	r := &unreadable{}
	var out, diag bytes.Buffer
	if run([]string{"execute"}, r, &out, &diag) != 2 || r.called || out.Len() != 0 || diag.String() != "modelboundinventory: accepts stdin only\n" {
		t.Fatal("arguments must refuse before reading")
	}
	for _, index := range []int{0, 1, 2} {
		raw := fixture(t, index)
		want, err := boundinventory.Check(raw)
		if err != nil {
			t.Fatal(err)
		}
		out.Reset()
		diag.Reset()
		if run(nil, bytes.NewReader(raw), &out, &diag) != 0 || diag.Len() != 0 || !bytes.Equal(out.Bytes(), want) {
			t.Fatal("native report parity", index)
		}
	}
	out.Reset()
	diag.Reset()
	if run(nil, &unreadable{}, &out, &diag) != 4 || out.Len() != 0 || diag.String() != "modelboundinventory: input read failed\n" {
		t.Fatal("read failure")
	}
	out.Reset()
	diag.Reset()
	if run(nil, bytes.NewBufferString("PRIVATE invalid JSON"), &out, &diag) != 2 || out.Len() != 0 || diag.String() != "modelboundinventory: input rejected\n" {
		t.Fatal("input refusal")
	}
	diag.Reset()
	if run(nil, bytes.NewReader(fixture(t, 0)), shortWriter{}, &diag) != 3 || diag.String() != "modelboundinventory: output failed\n" {
		t.Fatal("short write")
	}
	for _, err := range []error{boundinventory.ErrInvalid, boundinventory.ErrBinding, boundinventory.ErrLimit} {
		if failure(err) != 2 {
			t.Fatal("input category")
		}
	}
	if failure(boundinventory.ErrBoundary) != 4 {
		t.Fatal("boundary category")
	}
}
func TestActualBinaryAndClosedStdout(t *testing.T) {
	ctx, stop := context.WithTimeout(context.Background(), 90*time.Second)
	defer stop()
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	name := "modelboundinventory"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	binary := filepath.Join(t.TempDir(), name)
	build := exec.CommandContext(ctx, filepath.Join(runtime.GOROOT(), "bin", "go"), "build", "-o", binary, "./cmd/modelboundinventory")
	build.Dir = root
	build.WaitDelay = time.Second
	if b, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build: %v %s", err, b)
	}
	for _, index := range []int{0, 1} {
		raw := fixture(t, index)
		want, err := boundinventory.Check(raw)
		if err != nil {
			t.Fatal(err)
		}
		cmd := exec.CommandContext(ctx, binary)
		cmd.Stdin = bytes.NewReader(raw)
		cmd.WaitDelay = time.Second
		var out, diag bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &diag
		if err := cmd.Run(); err != nil || diag.Len() != 0 || !bytes.Equal(out.Bytes(), want) {
			t.Fatal("actual stdout", err, diag.String())
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
	cmd := exec.CommandContext(ctx, binary)
	cmd.Stdin = bytes.NewReader(fixture(t, 0))
	cmd.Stdout = writer
	cmd.WaitDelay = time.Second
	var diag bytes.Buffer
	cmd.Stderr = &diag
	err = cmd.Run()
	var exit *exec.ExitError
	if ctx.Err() != nil || !errors.As(err, &exit) || exit.ExitCode() != 3 || diag.String() != "modelboundinventory: output failed\n" {
		t.Fatalf("closed pipe: %v %q", err, diag.String())
	}
	if run(nil, &unreadable{}, io.Discard, io.Discard) != 4 {
		t.Fatal("private read error")
	}
}
