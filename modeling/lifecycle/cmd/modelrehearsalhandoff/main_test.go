package main

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/rehearsalhandoff"
)

type failReader struct{}

func (failReader) Read([]byte) (int, error) { return 0, errors.New("private input") }

type shortWriter struct{}

func (shortWriter) Write(b []byte) (int, error) { return len(b) / 2, nil }

type failWriter struct{}

func (failWriter) Write([]byte) (int, error) { return 0, errors.New("private output") }
func TestCLIControlFlowAndIO(t *testing.T) {
	calls := 0
	handler := func(b []byte) ([]byte, error) {
		calls++
		if string(b) != "fixed" {
			t.Fatal("input changed")
		}
		return []byte("complete\n"), nil
	}
	var out, diag bytes.Buffer
	for _, mode := range []string{"pack", "verify"} {
		out.Reset()
		diag.Reset()
		if code := runWith([]string{mode}, strings.NewReader("fixed"), &out, &diag, handler, handler); code != 0 || out.String() != "complete\n" || diag.Len() != 0 {
			t.Fatal("successful route failed")
		}
	}
	if calls != 2 {
		t.Fatal("wrong route call count")
	}
	for _, args := range [][]string{nil, {"bad"}, {"pack", "extra"}} {
		out.Reset()
		diag.Reset()
		if code := runWith(args, strings.NewReader("fixed"), &out, &diag, handler, handler); code != 2 || out.Len() != 0 || diag.String() != "modelrehearsalhandoff: invalid input\n" {
			t.Fatal("bad args accepted")
		}
	}
	out.Reset()
	diag.Reset()
	if code := runWith([]string{"pack"}, failReader{}, &out, &diag, handler, handler); code != 2 || out.Len() != 0 || strings.Contains(diag.String(), "private") {
		t.Fatal("read error leaked")
	}
	for _, w := range []interface{ Write([]byte) (int, error) }{shortWriter{}, failWriter{}} {
		diag.Reset()
		if code := runWith([]string{"verify"}, strings.NewReader("fixed"), w, &diag, handler, handler); code != 3 || diag.String() != "modelrehearsalhandoff: output failed\n" {
			t.Fatal("write failure not handled")
		}
	}
	bad := func([]byte) ([]byte, error) { return nil, errors.New("private cause") }
	out.Reset()
	diag.Reset()
	if code := runWith([]string{"pack"}, strings.NewReader("fixed"), &out, &diag, bad, bad); code != 2 || out.Len() != 0 || strings.Contains(diag.String(), "private") {
		t.Fatal("handler failure leaked")
	}
	out.Reset()
	diag.Reset()
	if code := run([]string{"verify"}, strings.NewReader(`{}`), &out, &diag); code != 2 || out.Len() != 0 {
		t.Fatal("real parser accepted empty request")
	}
	out.Reset()
	diag.Reset()
	if code := runWith([]string{"pack"}, strings.NewReader(strings.Repeat(" ", rehearsalhandoff.MaxRequestBytes+1)), &out, &diag, handler, handler); code != 2 || out.Len() != 0 {
		t.Fatal("oversize accepted")
	}
}
