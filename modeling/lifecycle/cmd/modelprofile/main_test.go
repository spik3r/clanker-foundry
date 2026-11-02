package main

import (
	"bytes"
	"errors"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/compiler"
	"runtime"
	"testing"
)

type poisonReader struct{ reads *int }

func (p poisonReader) Read([]byte) (int, error) {
	*p.reads++
	return 0, errors.New("stdin must not be read")
}

type errorWriter struct{}

func (errorWriter) Write([]byte) (int, error) { return 0, errors.New("private output error") }

type shortWriter struct{}

func (shortWriter) Write(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	return len(p) - 1, nil
}
func TestCLI(t *testing.T) {
	var out, diag bytes.Buffer
	reads := 0
	if run([]string{"bad"}, poisonReader{&reads}, &out, &diag) != 2 || out.Len() != 0 || diag.String() != "modelprofile: unavailable\n" {
		t.Fatal("unsupported argument")
	}
	out.Reset()
	diag.Reset()
	if runtime.Version() != compiler.ToolchainVersion {
		if run(nil, poisonReader{&reads}, &out, &diag) != 2 || out.Len() != 0 {
			t.Fatal("wrong toolchain")
		}
		t.Skip("positive CLI cases require Go 1.27.3")
	}
	if run(nil, poisonReader{&reads}, &out, &diag) != 0 || out.Len() == 0 || diag.Len() != 0 || reads != 0 {
		t.Fatal("normal output or stdin read")
	}
	diag.Reset()
	if run(nil, poisonReader{&reads}, shortWriter{}, &diag) != 3 || diag.String() != "modelprofile: output failed\n" {
		t.Fatal("short write")
	}
	diag.Reset()
	if run(nil, poisonReader{&reads}, errorWriter{}, &diag) != 3 || diag.String() != "modelprofile: output failed\n" || reads != 0 {
		t.Fatal("output error handling")
	}

}
