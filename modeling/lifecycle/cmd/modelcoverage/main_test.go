package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/scenariocoverage"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/scenarios"
)

func validRequest(t *testing.T) []byte {
	t.Helper()
	m, e := os.ReadFile("../../examples/lifecycle.json")
	if e != nil {
		t.Fatal(e)
	}
	s, e := os.ReadFile("../../scenarios/testdata/lifecycle-v1.json")
	if e != nil {
		t.Fatal(e)
	}
	r, e := scenarios.Execute(s)
	if e != nil {
		t.Fatal(e)
	}
	b, e := scenarios.EncodeReport(r)
	if e != nil {
		t.Fatal(e)
	}
	req := struct {
		FormatVersion string          `json:"formatVersion"`
		Model         json.RawMessage `json:"model"`
		Report        json.RawMessage `json:"report"`
	}{scenariocoverage.RequestFormatVersion, m, b}
	out, e := json.Marshal(req)
	if e != nil {
		t.Fatal(e)
	}
	return out
}

type shortWriter struct{}

func (shortWriter) Write(p []byte) (int, error) { return len(p) / 2, nil }

type errorReader struct{}

func (errorReader) Read([]byte) (int, error) { return 0, errors.New("private input") }
func TestRunAndFailures(t *testing.T) {
	req := validRequest(t)
	var out, diag bytes.Buffer
	if code := run(nil, bytes.NewReader(req), &out, &diag); code != 0 {
		t.Fatalf("code %d: %s", code, diag.String())
	}
	first := append([]byte(nil), out.Bytes()...)
	out.Reset()
	if code := run(nil, bytes.NewReader(req), &out, &diag); code != 0 || !bytes.Equal(first, out.Bytes()) {
		t.Fatal("nondeterministic")
	}
	for _, input := range [][]byte{[]byte(`{}`), append(append([]byte(nil), req...), []byte(` {}`)...), []byte(strings.Repeat("x", scenariocoverage.MaxRequestBytes+1))} {
		out.Reset()
		diag.Reset()
		if code := run(nil, bytes.NewReader(input), &out, &diag); code != 2 || out.Len() != 0 || diag.String() != "modelcoverage: invalid input\n" {
			t.Fatalf("refusal code=%d output=%q diag=%q", code, out.String(), diag.String())
		}
	}
	out.Reset()
	diag.Reset()
	if code := run([]string{"extra"}, bytes.NewReader(req), &out, &diag); code != 2 || out.Len() != 0 {
		t.Fatal("accepted args")
	}
	diag.Reset()
	if code := run(nil, errorReader{}, &out, &diag); code != 2 || strings.Contains(diag.String(), "private") {
		t.Fatal("leaked input error")
	}
	diag.Reset()
	if code := run(nil, bytes.NewReader(req), shortWriter{}, &diag); code != 3 || diag.String() != "modelcoverage: output failed\n" {
		t.Fatal("short write not reported")
	}
}

func TestStrictRequestRefusals(t *testing.T) {
	req := validRequest(t)
	mutations := map[string][]byte{
		"duplicate wrapper":      bytes.Replace(req, []byte(`"formatVersion":`), []byte(`"formatVersion":"other","formatVersion":`), 1),
		"unknown wrapper":        bytes.Replace(req, []byte(`{"formatVersion"`), []byte(`{"unknown":1,"formatVersion"`), 1),
		"case variant":           bytes.Replace(req, []byte(`"formatVersion"`), []byte(`"FormatVersion"`), 1),
		"null model":             []byte(`{"formatVersion":"lifecycle-scenario-assertion-inventory-request/v1","model":null,"report":null}`),
		"wrong version":          bytes.Replace(req, []byte(scenariocoverage.RequestFormatVersion), []byte("wrong"), 1),
		"bad surrogate":          bytes.Replace(req, []byte(`"happy"`), []byte(`"\ud800"`), 1),
		"duplicate nested field": bytes.Replace(req, []byte(`"scenarioVersion":`), []byte(`"scenarioVersion":"bad","scenarioVersion":`), 1),
	}
	for name, raw := range mutations {
		t.Run(name, func(t *testing.T) {
			if bytes.Equal(raw, req) {
				t.Fatal("mutation did not apply")
			}
			var out, diag bytes.Buffer
			if code := run(nil, bytes.NewReader(raw), &out, &diag); code != 2 || out.Len() != 0 {
				t.Fatalf("accepted %s: %d", name, code)
			}
		})
	}
}
