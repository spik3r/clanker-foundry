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
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/proposalreplay"
)

func request(t *testing.T, set string) []byte {
	t.Helper()
	r := proposalreplay.Request{FormatVersion: proposalreplay.RequestVersion}
	for _, v := range []struct {
		name   string
		target *string
	}{{"model", &r.Model}, {"suite", &r.Suite}, {"manifest", &r.Manifest}} {
		b, e := os.ReadFile("../../proposalreplay/testdata/" + v.name + "-" + set + ".json")
		if e != nil {
			t.Fatal(e)
		}
		*v.target = string(b)
	}
	b, e := json.Marshal(r)
	if e != nil {
		t.Fatal(e)
	}
	return b
}

type badReader struct{}

func (badReader) Read([]byte) (int, error) { return 0, errors.New("private read text") }

type badWriter struct{ short bool }

func (w badWriter) Write(b []byte) (int, error) {
	if w.short {
		return len(b) - 1, nil
	}
	return 0, errors.New("private write text")
}

func TestCommandBoundaries(t *testing.T) {
	raw := request(t, "a")
	ctx := context.Background()
	for _, args := range [][]string{nil, {"replay"}, {"execute", "extra"}} {
		var out, diag bytes.Buffer
		if run(ctx, args, bytes.NewReader(raw), &out, &diag) != 2 || out.Len() != 0 {
			t.Fatal("arguments", args)
		}
	}
	for _, in := range []io.Reader{badReader{}, strings.NewReader("{}"), strings.NewReader(strings.Repeat(" ", proposalreplay.MaxRequestBytes+1))} {
		var out, diag bytes.Buffer
		if run(ctx, []string{"check"}, in, &out, &diag) != 2 || out.Len() != 0 || strings.Contains(diag.String(), "private") {
			t.Fatal("input boundary")
		}
	}
	for _, out := range []io.Writer{badWriter{}, badWriter{short: true}} {
		var diag bytes.Buffer
		if run(ctx, []string{"check"}, bytes.NewReader(raw), out, &diag) != 3 || strings.Contains(diag.String(), "private") {
			t.Fatal("output boundary")
		}
	}
	var r proposalreplay.Request
	_ = json.Unmarshal(raw, &r)
	r.Suite += "\n"
	stale, _ := json.Marshal(r)
	var out, diag bytes.Buffer
	if run(ctx, []string{"execute"}, bytes.NewReader(stale), &out, &diag) != 1 {
		t.Fatal("stale should refuse before reading this non-module cwd", diag.String())
	}
	var receipt proposalreplay.Receipt
	if json.Unmarshal(out.Bytes(), &receipt) != nil || receipt.CandidateStarted || receipt.Preflight.Replay != "not-run" {
		t.Fatal("stale receipt")
	}
}

func TestExecutableCustomModelsAndClosedPipe(t *testing.T) {
	work := t.TempDir()
	binary := filepath.Join(work, "modelproposal")
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Minute)
	defer cancel()
	build := exec.CommandContext(ctx, "go", "build", "-p=1", "-o", binary, ".")
	if out, e := build.CombinedOutput(); e != nil {
		t.Fatalf("build: %v %s", e, out)
	}
	root, e := filepath.Abs("../..")
	if e != nil {
		t.Fatal(e)
	}
	var oracle struct {
		Sets []struct {
			Name   string
			Report json.RawMessage
		}
	}
	b, e := os.ReadFile("../../proposalreplay/testdata/expected.json")
	if e != nil || json.Unmarshal(b, &oracle) != nil {
		t.Fatal("oracle", e)
	}
	for _, x := range oracle.Sets {
		cmd := exec.CommandContext(ctx, binary, "execute")
		cmd.Dir = root
		cmd.Stdin = bytes.NewReader(request(t, x.Name))
		var out, diag bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &diag
		if e := cmd.Run(); e != nil {
			t.Fatalf("execute %s: %v %s %s", x.Name, e, diag.String(), out.String())
		}
		var r proposalreplay.Receipt
		if json.Unmarshal(out.Bytes(), &r) != nil || r.Outcome != "matched" || !r.CandidateStarted || r.Cleanup != "removed" {
			t.Fatal("actual command receipt", out.String())
		}
		var got, want any
		if json.Unmarshal(r.CandidateReport, &got) != nil || json.Unmarshal(x.Report, &want) != nil || !reflect.DeepEqual(got, want) {
			t.Fatal("actual command full oracle mismatch", x.Name)
		}
	}
	// A prebuilt check command needs neither module source nor writable temp space.
	cmd := exec.CommandContext(ctx, binary, "check")
	cmd.Dir = work
	cmd.Env = append(os.Environ(), "TMPDIR="+filepath.Join(work, "missing"))
	cmd.Stdin = bytes.NewReader(request(t, "a"))
	if out, e := cmd.CombinedOutput(); e != nil {
		t.Fatalf("pure command check: %v %s", e, out)
	}
	r, w, e := os.Pipe()
	if e != nil {
		t.Fatal(e)
	}
	_ = r.Close()
	defer w.Close()
	cmd = exec.CommandContext(ctx, binary, "check")
	cmd.Dir = work
	cmd.Stdin = bytes.NewReader(request(t, "a"))
	cmd.Stdout = w
	e = cmd.Run()
	var exit *exec.ExitError
	if !errors.As(e, &exit) || exit.ExitCode() != 3 {
		t.Fatalf("closed stdout: %v", e)
	}
}
