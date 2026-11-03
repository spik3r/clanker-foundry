package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/traceimpact"
	"io"
	"os"
	"strings"
	"testing"
)

type brokenReader struct{}

func (brokenReader) Read([]byte) (int, error) { return 0, errors.New("private input error") }

type shortWriter struct{}

func (shortWriter) Write(p []byte) (int, error) { return len(p) / 2, nil }
func example(t *testing.T) []byte {
	t.Helper()
	read := func(name string) string {
		b, e := os.ReadFile("../../traceimpact/testdata/" + name)
		if e != nil {
			t.Fatal(e)
		}
		return string(b)
	}
	side := func(prefix string) map[string]string {
		return map[string]string{"model": read(prefix + "-model.json"), "catalog": read(prefix + "-catalog.json"), "requirements": read(prefix + "-requirements.json")}
	}
	b, e := json.Marshal(map[string]any{"formatVersion": traceimpact.RequestVersion, "before": side("before"), "after": side("after")})
	if e != nil {
		t.Fatal(e)
	}
	return b
}
func TestRun(t *testing.T) {
	input := example(t)
	for _, c := range []struct {
		name string
		args []string
		in   io.Reader
		out  io.Writer
		want int
	}{
		{"changed is valid", nil, bytes.NewReader(input), &bytes.Buffer{}, 0},
		{"arguments", []string{"file"}, bytes.NewReader(input), &bytes.Buffer{}, 2},
		{"malformed", nil, strings.NewReader("{"), &bytes.Buffer{}, 2},
		{"oversized", nil, strings.NewReader(strings.Repeat(" ", traceimpact.MaxRequestBytes+1)), &bytes.Buffer{}, 2},
		{"read failure", nil, brokenReader{}, &bytes.Buffer{}, 2},
		{"short write", nil, bytes.NewReader(input), shortWriter{}, 3},
	} {
		t.Run(c.name, func(t *testing.T) {
			var diag bytes.Buffer
			status := run(c.args, c.in, c.out, &diag)
			if status != c.want {
				t.Fatalf("status %d want %d: %s", status, c.want, diag.String())
			}
			if strings.Contains(diag.String(), "private input error") {
				t.Fatal("leaked error")
			}
			if c.want == 2 {
				if b, ok := c.out.(*bytes.Buffer); ok && b.Len() != 0 {
					t.Fatal("invalid input wrote stdout")
				}
			}
			if c.want == 0 {
				b := c.out.(*bytes.Buffer).Bytes()
				var r struct {
					Review bool `json:"reviewRequired"`
				}
				if e := json.Unmarshal(b, &r); e != nil || !r.Review {
					t.Fatalf("valid changed report: %v %+v", e, r)
				}
			}
		})
	}
}
func TestRequestStrictnessAndDeterminism(t *testing.T) {
	raw := example(t)
	a, e := traceimpact.Execute(raw)
	if e != nil {
		t.Fatal(e)
	}
	b, e := traceimpact.Execute(raw)
	if e != nil || !bytes.Equal(a, b) {
		t.Fatal("non-deterministic output")
	}
	for name, input := range map[string][]byte{
		"trailing":        append(bytes.Clone(raw), []byte(` {}`)...),
		"unknown":         bytes.Replace(raw, []byte(`"formatVersion":`), []byte(`"unknown":1,"formatVersion":`), 1),
		"duplicate":       bytes.Replace(raw, []byte(`"formatVersion":`), []byte(`"formatVersion":"duplicate","formatVersion":`), 1),
		"case":            bytes.Replace(raw, []byte(`"formatVersion":`), []byte(`"FormatVersion":`), 1),
		"invalid unicode": bytes.Replace(raw, []byte(`"formatVersion":`), []byte(`"\ud800":1,"formatVersion":`), 1),
	} {
		t.Run(name, func(t *testing.T) {
			out, e := traceimpact.Execute(input)
			if e == nil || out != nil {
				t.Fatal("malformed request accepted")
			}
		})
	}
}
