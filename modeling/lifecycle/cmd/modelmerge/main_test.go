package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/jsonformat"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/modelmerge"
)

func request(t *testing.T, conflict bool) []byte {
	t.Helper()
	base, err := os.ReadFile("../../examples/lifecycle.json")
	if err != nil {
		t.Fatal(err)
	}
	left, right := base, base
	if conflict {
		left = bytes.Replace(base, []byte(`"id": "TaskLifecycleExample"`), []byte(`"id": "LeftModel"`), 1)
		right = bytes.Replace(base, []byte(`"id": "TaskLifecycleExample"`), []byte(`"id": "RightModel"`), 1)
		if bytes.Equal(base, left) {
			t.Fatal("fixture replacement did not match")
		}
	}
	return []byte(`{"formatVersion":"lifecycle-model-merge-json/v1","base":` + string(base) + `,"left":` + string(left) + `,"right":` + string(right) + `}`)
}

func TestCLIReports(t *testing.T) {
	for _, conflict := range []bool{false, true} {
		raw := request(t, conflict)
		var first []byte
		for range 2 {
			var out, diag bytes.Buffer
			if code := run(nil, bytes.NewReader(raw), &out, &diag); code != 0 || diag.Len() != 0 {
				t.Fatalf("%d %s", code, diag.String())
			}
			var parsed map[string]any
			if err := json.Unmarshal(out.Bytes(), &parsed); err != nil {
				t.Fatal(err)
			}
			status := "merged"
			if conflict {
				status = "conflicted"
			}
			if parsed["status"] != status || parsed["mode"] != "example-only" || parsed["authority"] != "none" {
				t.Fatal(parsed)
			}
			if first != nil && !bytes.Equal(first, out.Bytes()) {
				t.Fatal("nondeterministic stdout")
			}
			first = append([]byte(nil), out.Bytes()...)
		}
	}
}

type badReader struct{}

func (badReader) Read([]byte) (int, error) { return 0, io.ErrUnexpectedEOF }

type noRead struct{}

func (noRead) Read([]byte) (int, error) { panic("arguments must be refused before reading") }

type shortWriter struct {
	fail  bool
	calls int
}

func (w *shortWriter) Write(b []byte) (int, error) {
	w.calls++
	if w.fail {
		return 0, io.ErrClosedPipe
	}
	return len(b) / 2, nil
}

func TestCLIRefusesAndWriteFailures(t *testing.T) {
	for _, r := range []io.Reader{badReader{}, strings.NewReader("private invalid input"), strings.NewReader(strings.Repeat(" ", modelmerge.MaxRequestBytes+1))} {
		var out, diag bytes.Buffer
		if code := run(nil, r, &out, &diag); code != 2 || out.Len() != 0 || strings.Contains(diag.String(), "private") {
			t.Fatalf("bad refusal %d %q %q", code, out.String(), diag.String())
		}
	}
	var out, diag bytes.Buffer
	if code := run([]string{"output.json"}, noRead{}, &out, &diag); code != 2 || out.Len() != 0 {
		t.Fatal(code)
	}
	for _, fail := range []bool{false, true} {
		writer := &shortWriter{fail: fail}
		if code := run(nil, bytes.NewReader(request(t, false)), writer, &diag); code != 3 || writer.calls != 1 {
			t.Fatal(code, writer.calls)
		}
	}
}

func TestCLIClosedPipe(t *testing.T) {
	const child = "FOUNDRY_MODELMERGE_CLOSED_PIPE_CHILD"
	if os.Getenv(child) == "1" {
		os.Args = []string{"modelmerge"}
		main()
		return
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
	cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestCLIClosedPipe$")
	cmd.Env = append(os.Environ(), child+"=1")
	cmd.Stdin = bytes.NewReader(request(t, false))
	cmd.Stdout = writer
	var diag bytes.Buffer
	cmd.Stderr = &diag
	cmd.WaitDelay = time.Second
	err = cmd.Run()
	if ctx.Err() != nil {
		t.Fatal(ctx.Err())
	}
	var exit *exec.ExitError
	if !errors.As(err, &exit) || exit.ExitCode() != 3 || diag.String() != "model merge output failed\n" {
		t.Fatalf("got %v / %q", err, diag.String())
	}
}

func TestCLIInvalidCombinationStillReports(t *testing.T) {
	raw, err := os.ReadFile("../../examples/lifecycle.json")
	if err != nil {
		t.Fatal(err)
	}
	left, err := jsonformat.Decode(raw)
	if err != nil {
		t.Fatal(err)
	}
	right, err := jsonformat.Decode(raw)
	if err != nil {
		t.Fatal(err)
	}
	left.ID = "CombinedCollision"
	right.Initial = "CombinedCollision"
	for i := range right.States {
		if right.States[i].ID == "Draft" {
			right.States[i].ID = "CombinedCollision"
		}
	}
	for i := range right.Transitions {
		if right.Transitions[i].From == "Draft" {
			right.Transitions[i].From = "CombinedCollision"
		}
		if right.Transitions[i].To == "Draft" {
			right.Transitions[i].To = "CombinedCollision"
		}
	}
	l, err := jsonformat.Encode(left)
	if err != nil {
		t.Fatal(err)
	}
	r, err := jsonformat.Encode(right)
	if err != nil {
		t.Fatal(err)
	}
	input := []byte(`{"formatVersion":"lifecycle-model-merge-json/v1","base":` + string(raw) + `,"left":` + string(l) + `,"right":` + string(r) + `}`)
	var out, diag bytes.Buffer
	if code := run(nil, bytes.NewReader(input), &out, &diag); code != 0 || diag.Len() != 0 {
		t.Fatal(code, diag.String())
	}
	var got map[string]any
	if err = json.Unmarshal(out.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got["status"] != "invalid-combination" || got["candidate"] != nil || got["candidateDigest"] != nil {
		t.Fatal(got)
	}
}
