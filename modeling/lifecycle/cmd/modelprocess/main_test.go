package main

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/processview"
	"io"
	"os"
	"strings"
	"testing"
)

type brokenReader struct{}

func (brokenReader) Read([]byte) (int, error) { return 0, errors.New("private I/O details") }

type shortWriter struct{}

func (shortWriter) Write(p []byte) (int, error) { return len(p) / 2, nil }
func TestRun(t *testing.T) {
	source, e := os.ReadFile("../../processview/testdata/model.json")
	if e != nil {
		t.Fatal(e)
	}
	for _, c := range []struct {
		name string
		args []string
		in   io.Reader
		out  io.Writer
		want int
	}{
		{"json", []string{"json"}, bytes.NewReader(source), &bytes.Buffer{}, 0},
		{"svg", []string{"svg"}, bytes.NewReader(source), &bytes.Buffer{}, 0},
		{"missing argument", nil, bytes.NewReader(source), &bytes.Buffer{}, 2},
		{"extra argument", []string{"svg", "file"}, bytes.NewReader(source), &bytes.Buffer{}, 2},
		{"invalid model", []string{"svg"}, strings.NewReader("{"), &bytes.Buffer{}, 2},
		{"oversized", []string{"json"}, io.LimitReader(strings.NewReader(strings.Repeat(" ", processview.MaxSourceBytes+1)), processview.MaxSourceBytes+1), &bytes.Buffer{}, 2},
		{"read failure", []string{"json"}, brokenReader{}, &bytes.Buffer{}, 2},
		{"short write", []string{"svg"}, bytes.NewReader(source), shortWriter{}, 3},
	} {
		t.Run(c.name, func(t *testing.T) {
			var diag bytes.Buffer
			status := run(c.args, c.in, c.out, &diag)
			if status != c.want {
				t.Fatalf("status %d want %d; %s", status, c.want, diag.String())
			}
			if strings.Contains(diag.String(), "private I/O details") {
				t.Fatal("leaked input error")
			}
			if c.want == 2 {
				if b, ok := c.out.(*bytes.Buffer); ok && b.Len() != 0 {
					t.Fatal("failure produced stdout")
				}
			}
			if c.want == 0 {
				b := c.out.(*bytes.Buffer).Bytes()
				if c.name == "json" && !json.Valid(b) {
					t.Fatal("invalid JSON")
				}
				if c.name == "svg" {
					dec := xml.NewDecoder(bytes.NewReader(b))
					for {
						_, e := dec.Token()
						if e == io.EOF {
							break
						}
						if e != nil {
							t.Fatal(e)
						}
					}
				}
			}
		})
	}
}
