package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/custommergerehearsal"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func fixture(t *testing.T, target int, reverse bool) []byte {
	t.Helper()
	raw, e := os.ReadFile("../../custommergerehearsal/testdata/observations.json")
	if e != nil {
		t.Fatal(e)
	}
	var f struct {
		Versions                                                map[string]struct{ Suite, Manifest, MismatchSuite, MismatchManifest, ModelDigest string }
		MergeRequest, ReverseMergeRequest, ConflictMergeRequest string
	}
	if json.Unmarshal(raw, &f) != nil {
		t.Fatal("fixture")
	}
	var pairs [4]custommergerehearsal.Expectations
	for i, role := range []string{"base", "left", "right", "candidate"} {
		v := f.Versions[role]
		pairs[i] = custommergerehearsal.Expectations{Suite: v.Suite, Manifest: v.Manifest}
		if i == target {
			pairs[i] = custommergerehearsal.Expectations{Suite: v.MismatchSuite, Manifest: v.MismatchManifest}
		}
	}
	m := f.MergeRequest
	if reverse {
		m = f.ReverseMergeRequest
		pairs[1], pairs[2] = pairs[2], pairs[1]
	}
	if target == 4 {
		m = f.ConflictMergeRequest
	}
	r := custommergerehearsal.Request{FormatVersion: custommergerehearsal.RequestVersion, MergeRequest: m, Base: pairs[0], Left: pairs[1], Right: pairs[2], Candidate: pairs[3], ExpectedCandidateDigest: f.Versions["candidate"].ModelDigest}
	b, e := json.Marshal(r)
	if e != nil {
		t.Fatal(e)
	}
	return b
}

type unreadable struct{ called bool }

func (r *unreadable) Read([]byte) (int, error) {
	r.called = true
	return 0, errors.New("PRIVATE read failure")
}

type shortWriter struct{}

func (shortWriter) Write(b []byte) (int, error) { return len(b) - 1, nil }

type failedWriter struct{}

func (failedWriter) Write([]byte) (int, error) { return 0, errors.New("PRIVATE write failure") }
func TestRunBoundaries(t *testing.T) {
	for _, args := range [][]string{nil, {"help"}, {"execute", "extra"}} {
		r := &unreadable{}
		var out, diag bytes.Buffer
		if run(context.Background(), args, r, &out, &diag) != 2 || r.called || out.Len() != 0 || diag.String() != "modelmergerehearsal: usage: modelmergerehearsal check|execute < request.json\n" {
			t.Fatal("usage before reading")
		}
	}
	raw := fixture(t, -1, false)
	want, e := custommergerehearsal.Check(raw)
	if e != nil {
		t.Fatal(e)
	}
	var out, diag bytes.Buffer
	if run(context.Background(), []string{"check"}, bytes.NewReader(raw), &out, &diag) != 0 || !bytes.Equal(out.Bytes(), want) || diag.Len() != 0 {
		t.Fatal("native check parity")
	}
	for _, input := range []io.Reader{&unreadable{}, bytes.NewBufferString("PRIVATE invalid"), io.LimitReader(zeroReader{}, custommergerehearsal.MaxRequestBytes+1)} {
		out.Reset()
		diag.Reset()
		if run(context.Background(), []string{"check"}, input, &out, &diag) != 2 || out.Len() != 0 || bytes.Contains(diag.Bytes(), []byte("PRIVATE")) {
			t.Fatal("input refusal")
		}
	}
	for _, writer := range []io.Writer{shortWriter{}, failedWriter{}} {
		diag.Reset()
		if run(context.Background(), []string{"check"}, bytes.NewReader(raw), writer, &diag) != 3 || diag.String() != "modelmergerehearsal: report delivery failed; execution may have started\n" {
			t.Fatal("delivery")
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	out.Reset()
	diag.Reset()
	if run(ctx, []string{"execute"}, bytes.NewReader(raw), &out, &diag) != 4 {
		t.Fatal("canceled invocation")
	}
	var r custommergerehearsal.Report
	if json.Unmarshal(out.Bytes(), &r) != nil || r.Base.Operation != "check" || r.Candidate.Operation != "check" {
		t.Fatal("fabricated execution")
	}
}

type zeroReader struct{}

func (zeroReader) Read(p []byte) (int, error) {
	for i := range p {
		p[i] = ' '
	}
	return len(p), nil
}
func TestActualBinaryContractAndDelivery(t *testing.T) {
	root, e := filepath.Abs("../..")
	if e != nil {
		t.Fatal(e)
	}
	binary := filepath.Join(t.TempDir(), "modelmergerehearsal")
	if runtime.GOOS == "windows" {
		binary += ".exe"
	}
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	build := exec.CommandContext(ctx, filepath.Join(runtime.GOROOT(), "bin", "go"), "build", "-o", binary, "./cmd/modelmergerehearsal")
	build.Dir = root
	build.WaitDelay = time.Second
	if b, e := build.CombinedOutput(); e != nil {
		t.Fatalf("build: %v %s", e, b)
	}
	for _, c := range []struct {
		mode         string
		target, code int
		outcome      string
	}{
		{"check", -1, 0, "checked"}, {"check", 4, 1, "conflicted"}, {"execute", -1, 0, "matched"}, {"execute", 0, 1, "base-mismatched"}, {"execute", 1, 1, "left-mismatched"}, {"execute", 2, 1, "right-mismatched"}, {"execute", 3, 1, "candidate-mismatched"},
	} {
		t.Run(c.mode+c.outcome, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 180*time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, binary, c.mode)
			cmd.Dir = root
			cmd.Stdin = bytes.NewReader(fixture(t, c.target, false))
			cmd.WaitDelay = time.Second
			var out, diag bytes.Buffer
			cmd.Stdout = &out
			cmd.Stderr = &diag
			e := cmd.Run()
			code := 0
			if e != nil {
				var ex *exec.ExitError
				if !errors.As(e, &ex) {
					t.Fatal(e)
				}
				code = ex.ExitCode()
			}
			if ctx.Err() != nil || code != c.code || diag.Len() != 0 {
				t.Fatalf("exit %d: %v %q", code, e, diag.String())
			}
			var r custommergerehearsal.Report
			if json.Unmarshal(out.Bytes(), &r) != nil || r.Outcome != c.outcome || r.Action != c.mode {
				t.Fatal("actual output")
			}
			if c.mode == "execute" && r.Base.Operation != "execute" {
				t.Fatal("missing execution")
			}
		})
	}
	for _, mode := range []string{"check", "execute"} {
		t.Run(mode+"-closed-stdout", func(t *testing.T) {
			reader, writer, e := os.Pipe()
			if e != nil {
				t.Fatal(e)
			}
			defer writer.Close()
			if reader.Close() != nil {
				t.Fatal("pipe")
			}
			ctx, cancel := context.WithTimeout(context.Background(), 180*time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, binary, mode)
			cmd.Dir = root
			cmd.Stdin = bytes.NewReader(fixture(t, -1, false))
			cmd.Stdout = writer
			cmd.WaitDelay = time.Second
			var diag bytes.Buffer
			cmd.Stderr = &diag
			e = cmd.Run()
			var ex *exec.ExitError
			if ctx.Err() != nil || !errors.As(e, &ex) || ex.ExitCode() != 3 || diag.String() != "modelmergerehearsal: report delivery failed; execution may have started\n" {
				t.Fatalf("closed stdout: %v %q", e, diag.String())
			}
		})
	}
}
