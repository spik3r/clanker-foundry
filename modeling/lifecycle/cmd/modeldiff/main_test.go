package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/impact"
)

func request(t *testing.T, changed bool) []byte {
	t.Helper()
	before, err := os.ReadFile("../../examples/lifecycle.json")
	if err != nil {
		t.Fatal(err)
	}
	after := append([]byte{}, before...)
	if changed {
		after = bytes.Replace(after, []byte(`"label": "Draft"`), []byte(`"label": "Draft revised"`), 1)
		if bytes.Equal(before, after) {
			t.Fatal("sample edit did not apply")
		}
	}
	raw, err := json.Marshal(struct {
		FormatVersion string           `json:"formatVersion"`
		Before        json.RawMessage  `json:"before"`
		After         json.RawMessage  `json:"after"`
		Bindings      []impact.Binding `json:"bindings"`
	}{impact.FormatVersion, before, after, []impact.Binding{{ID: "Asserted", ModelDigest: strings.Repeat("0", 64)}}})
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

type failReader struct{}

func (failReader) Read([]byte) (int, error) { return 0, errors.New("private input error") }

type failWriter struct{ short bool }

func (w failWriter) Write(p []byte) (int, error) {
	if w.short {
		return len(p) - 1, nil
	}
	return 0, errors.New("private output error")
}

type countReader struct{ n int }

func (r *countReader) Read(p []byte) (int, error) {
	for i := range p {
		p[i] = ' '
	}
	r.n += len(p)
	return len(p), nil
}

func TestCLIStatusAndIO(t *testing.T) {
	for _, changed := range []bool{false, true} {
		t.Run(map[bool]string{false: "equal", true: "changed"}[changed], func(t *testing.T) {
			var out, diag bytes.Buffer
			if code := run(nil, bytes.NewReader(request(t, changed)), &out, &diag); code != 0 || diag.Len() != 0 {
				t.Fatalf("code%d diagnostic%q", code, diag.String())
			}
			var r impact.Report
			if err := json.Unmarshal(out.Bytes(), &r); err != nil {
				t.Fatal(err)
			}
			if r.Changed != changed || r.ReviewRequired != changed || r.Bindings[0].Status != "matches-neither" {
				t.Fatal("comparison treated as approval or binding failure", r)
			}
		})
	}
	for name, input := range map[string]io.Reader{"empty": strings.NewReader(""), "malformed": strings.NewReader(`{"private":"secret"}`), "read failure": failReader{}, "oversize": strings.NewReader(strings.Repeat(" ", impact.MaxSourceBytes+1))} {
		t.Run(name, func(t *testing.T) {
			var out, diag bytes.Buffer
			if code := run(nil, input, &out, &diag); code != 2 || out.Len() != 0 || strings.Contains(diag.String(), "secret") || strings.Contains(diag.String(), "private") {
				t.Fatalf("code%d stdout%d diagnostic%q", code, out.Len(), diag.String())
			}
		})
	}
	var out, diag bytes.Buffer
	if code := run([]string{"--output=private"}, bytes.NewReader(request(t, false)), &out, &diag); code != 2 || out.Len() != 0 || strings.Contains(diag.String(), "private") {
		t.Fatal("arbitrary args accepted or leaked")
	}
	for _, short := range []bool{false, true} {
		diag.Reset()
		if code := run(nil, bytes.NewReader(request(t, false)), failWriter{short}, &diag); code != 3 || diag.String() != "modeldiff: output write failed\n" {
			t.Fatal("write failure not explicit", code, diag.String())
		}
	}
	r := &countReader{}
	out.Reset()
	diag.Reset()
	if code := run(nil, r, &out, &diag); code != 2 || out.Len() != 0 || r.n != impact.MaxSourceBytes+1 {
		t.Fatal("input was not bounded", code, r.n)
	}
}

func TestSeparateProcessRepeatAndNoWrites(t *testing.T) {
	tmp := t.TempDir()
	bin := filepath.Join(tmp, "modeldiff")
	build := exec.Command(filepath.Join(runtime.GOROOT(), "bin", "go"), "build", "-o", bin, ".")
	build.Env = append(os.Environ(), "GOTOOLCHAIN=local", "GOPROXY=off", "GOSUMDB=off")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build: %v\n%s", err, out)
	}
	cwd := filepath.Join(tmp, "empty")
	if err := os.Mkdir(cwd, 0700); err != nil {
		t.Fatal(err)
	}
	raw := request(t, true)
	var previous []byte
	for i := 0; i < 2; i++ {
		cmd := exec.Command(bin)
		cmd.Dir = cwd
		cmd.Stdin = bytes.NewReader(raw)
		var diag bytes.Buffer
		cmd.Stderr = &diag
		out, err := cmd.Output()
		if err != nil || diag.Len() != 0 {
			t.Fatal("process failed", err, diag.String())
		}
		if i > 0 && !bytes.Equal(out, previous) {
			t.Fatal("process byte drift")
		}
		previous = out
	}
	entries, err := os.ReadDir(cwd)
	if err != nil || len(entries) != 0 {
		t.Fatal("CLI wrote to work directory", entries, err)
	}
}
