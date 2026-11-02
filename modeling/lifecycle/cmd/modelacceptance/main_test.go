package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/acceptanceauthor"
	"io"
	"os"
	"strings"
	"testing"
)

type failedReader struct{}

func (failedReader) Read([]byte) (int, error) { return 0, errors.New("private input error") }

type shortWriter struct{}

func (shortWriter) Write(b []byte) (int, error) { return len(b) / 2, nil }
func request(t *testing.T, stale bool) []byte {
	t.Helper()
	read := func(p string) string {
		b, e := os.ReadFile(p)
		if e != nil {
			t.Fatal(e)
		}
		return string(b)
	}
	m := read("../../acceptanceauthor/testdata/manifest.json")
	model := read("../../examples/lifecycle.json")
	s := read("../../acceptanceauthor/testdata/suite.json")
	if stale {
		s += "\n"
	}
	b, e := json.Marshal(acceptanceauthor.Request{FormatVersion: acceptanceauthor.RequestVersion, Mode: "replay", Manifest: m, Model: model, Suite: s})
	if e != nil {
		t.Fatal(e)
	}
	return b
}
func TestRun(t *testing.T) {
	for _, c := range []struct {
		name string
		args []string
		in   io.Reader
		out  io.Writer
		want int
	}{
		{"matched", nil, bytes.NewReader(request(t, false)), &bytes.Buffer{}, 0},
		{"stale", nil, bytes.NewReader(request(t, true)), &bytes.Buffer{}, 1},
		{"args", []string{"file"}, strings.NewReader(""), &bytes.Buffer{}, 2},
		{"bad", nil, strings.NewReader("{"), &bytes.Buffer{}, 2},
		{"read", nil, failedReader{}, &bytes.Buffer{}, 2},
		{"short", nil, bytes.NewReader(request(t, false)), shortWriter{}, 3},
	} {
		t.Run(c.name, func(t *testing.T) {
			var diag bytes.Buffer
			got := run(c.args, c.in, c.out, &diag)
			if got != c.want {
				t.Fatalf("exit %d want %d: %s", got, c.want, diag.String())
			}
			if strings.Contains(diag.String(), "private input error") {
				t.Fatal("leaked error")
			}
			if c.want == 2 {
				if b, ok := c.out.(*bytes.Buffer); ok && b.Len() != 0 {
					t.Fatal("error wrote stdout")
				}
			}
			if c.want == 1 {
				if b := c.out.(*bytes.Buffer); !bytes.Contains(b.Bytes(), []byte(`"replay":"not-run"`)) {
					t.Fatal("stale replay was not suppressed")
				}
			}
		})
	}
}
