package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/intakeimpact"
)

type brokenReader struct{}

func (brokenReader) Read([]byte) (int, error) { return 0, errors.New("private read detail") }

type brokenWriter struct{}

func (brokenWriter) Write([]byte) (int, error) { return 0, errors.New("private write detail") }

type shortWriter struct{}

func (shortWriter) Write(b []byte) (int, error) { return len(b) / 2, nil }

type spaces struct{}

func (spaces) Read(b []byte) (int, error) {
	for i := range b {
		b[i] = ' '
	}
	return len(b), nil
}
func request(t *testing.T) []byte {
	t.Helper()
	read := func(name string) string {
		b, e := os.ReadFile("../../intakeimpact/testdata/" + name + ".json")
		if e != nil {
			t.Fatal(e)
		}
		return string(b)
	}
	b, e := json.Marshal(map[string]string{"formatVersion": intakeimpact.RequestVersion, "before": read("before"), "after": read("after")})
	if e != nil {
		t.Fatal(e)
	}
	return b
}
func TestRun(t *testing.T) {
	raw := request(t)
	for _, tc := range []struct {
		name string
		args []string
		in   io.Reader
		out  io.Writer
		want int
	}{
		{"changed valid", nil, bytes.NewReader(raw), &bytes.Buffer{}, 0},
		{"arguments", []string{"file"}, bytes.NewReader(raw), &bytes.Buffer{}, 2},
		{"malformed", nil, strings.NewReader("{"), &bytes.Buffer{}, 2},
		{"oversized", nil, io.LimitReader(spaces{}, intakeimpact.MaxRequestBytes+1), &bytes.Buffer{}, 2},
		{"read failure", nil, brokenReader{}, &bytes.Buffer{}, 2},
		{"short write", nil, bytes.NewReader(raw), shortWriter{}, 3},
		{"failed write", nil, bytes.NewReader(raw), brokenWriter{}, 3},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var diag bytes.Buffer
			if status := run(tc.args, tc.in, tc.out, &diag); status != tc.want {
				t.Fatalf("exit %d want %d: %s", status, tc.want, diag.String())
			}
			if strings.Contains(diag.String(), "private") {
				t.Fatal("private diagnostic leaked")
			}
			if tc.want == 2 {
				if b, ok := tc.out.(*bytes.Buffer); ok && b.Len() != 0 {
					t.Fatal("refusal emitted output")
				}
			}
			if tc.want == 0 {
				var r intakeimpact.Report
				if e := json.Unmarshal(tc.out.(*bytes.Buffer).Bytes(), &r); e != nil || !r.ContentChanged || len(r.AfterGaps) == 0 {
					t.Fatalf("invalid changed report: %v %+v", e, r)
				}
				if diag.Len() != 0 {
					t.Fatal("valid report emitted diagnostic")
				}
			}
		})
	}
}
func TestSubprocessEntry(t *testing.T) {
	if os.Getenv("FOUNDRY_INTAKE_DIFF_TEST_CHILD") != "1" {
		return
	}
	os.Args = []string{"modelintakediff"}
	main()
}
func TestClosedStdoutReturnsThree(t *testing.T) {
	binary, e := os.Executable()
	if e != nil {
		t.Fatal(e)
	}
	r, w, e := os.Pipe()
	if e != nil {
		t.Fatal(e)
	}
	defer w.Close()
	if e = r.Close(); e != nil {
		t.Fatal(e)
	}
	cmd := exec.Command(binary, "-test.run=^TestSubprocessEntry$")
	cmd.Env = append(os.Environ(), "FOUNDRY_INTAKE_DIFF_TEST_CHILD=1")
	cmd.Stdin = bytes.NewReader(request(t))
	cmd.Stdout = w
	var diagnostic bytes.Buffer
	cmd.Stderr = &diagnostic
	e = cmd.Run()
	var exit *exec.ExitError
	if !errors.As(e, &exit) || exit.ExitCode() != 3 {
		t.Fatalf("closed pipe exit: %v; %s", e, diagnostic.String())
	}
	if diagnostic.String() != "modelintakediff: output write failed\n" {
		t.Fatalf("diagnostic: %q", diagnostic.String())
	}
}
