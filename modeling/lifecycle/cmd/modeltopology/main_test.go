package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/topology"
)

func request(t *testing.T) []byte {
	t.Helper()
	m, err := os.ReadFile("../../examples/lifecycle.json")
	if err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(struct {
		FormatVersion string          `json:"formatVersion"`
		Model         json.RawMessage `json:"model"`
	}{topology.RequestFormatVersion, m})
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

type failReader struct{}

func (failReader) Read([]byte) (int, error) { return 0, errors.New("private input") }

type shortWriter struct{}

func (shortWriter) Write(p []byte) (int, error) { return len(p) / 2, nil }

type failWriter struct{}

func (failWriter) Write([]byte) (int, error) { return 0, errors.New("private output") }
func TestCLIStatusAndIO(t *testing.T) {
	req := request(t)
	var out, diag bytes.Buffer
	if code := run(nil, bytes.NewReader(req), &out, &diag); code != 0 || diag.Len() != 0 {
		t.Fatalf("code %d diag %q", code, diag.String())
	}
	first := append([]byte{}, out.Bytes()...)
	out.Reset()
	if code := run(nil, bytes.NewReader(req), &out, &diag); code != 0 || !bytes.Equal(first, out.Bytes()) {
		t.Fatal("nondeterministic CLI")
	}
	for name, b := range map[string][]byte{"invalid": []byte(`{"secret":"private"}`), "oversize": []byte(strings.Repeat(" ", topology.MaxRequestBytes+1))} {
		t.Run(name, func(t *testing.T) {
			out.Reset()
			diag.Reset()
			code := run(nil, bytes.NewReader(b), &out, &diag)
			if code != 2 || out.Len() != 0 || diag.String() != "modeltopology: invalid input\n" {
				t.Fatalf("code %d out %q diag %q", code, out.String(), diag.String())
			}
		})
	}
	out.Reset()
	diag.Reset()
	if code := run([]string{"extra"}, bytes.NewReader(req), &out, &diag); code != 2 || out.Len() != 0 {
		t.Fatal("args accepted")
	}
	diag.Reset()
	if code := run(nil, failReader{}, &out, &diag); code != 2 || strings.Contains(diag.String(), "private") {
		t.Fatal("reader error leaked")
	}
	for _, w := range []interface{ Write([]byte) (int, error) }{shortWriter{}, failWriter{}} {
		diag.Reset()
		if code := run(nil, bytes.NewReader(req), w, &diag); code != 3 || diag.String() != "modeltopology: output failed\n" {
			t.Fatal("write failure not reported")
		}
	}
}
