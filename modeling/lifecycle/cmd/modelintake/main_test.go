package main

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"
)

const valid = `{"formatVersion":"lifecycle-intake-json/v1","intent":"intent","decisions":[],"questions":[],"artifacts":{"requirements":null,"model":null,"scenarioSuite":null},"dependencies":{"modelRequirements":null,"suiteRequirements":null,"suiteModel":null},"observed":{"requirements":null,"model":null,"scenarioSuite":null}}`

type failedReader struct{}

func (failedReader) Read([]byte) (int, error) { return 0, errors.New("private error") }

type shortWriter struct{}

func (shortWriter) Write(b []byte) (int, error) { return len(b) / 2, nil }
func TestRun(t *testing.T) {
	for _, c := range []struct {
		name   string
		args   []string
		input  io.Reader
		output io.Writer
		want   int
	}{
		{"ok", nil, strings.NewReader(valid), &bytes.Buffer{}, 0},
		{"args", []string{"file"}, strings.NewReader(valid), &bytes.Buffer{}, 2},
		{"bad json", nil, strings.NewReader("{"), &bytes.Buffer{}, 2},
		{"read failure", nil, failedReader{}, &bytes.Buffer{}, 2},
		{"short write", nil, strings.NewReader(valid), shortWriter{}, 3},
	} {
		t.Run(c.name, func(t *testing.T) {
			var diag bytes.Buffer
			got := run(c.args, c.input, c.output, &diag)
			if got != c.want {
				t.Fatalf("exit %d want %d", got, c.want)
			}
			if strings.Contains(diag.String(), "private error") {
				t.Fatal("leaked error")
			}
			if c.want == 2 {
				if o, ok := c.output.(*bytes.Buffer); ok && o.Len() != 0 {
					t.Fatal("error wrote stdout")
				}
			}
		})
	}
}
