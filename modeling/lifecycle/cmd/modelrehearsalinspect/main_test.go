package main

import (
	"bytes"
	"errors"
	"io"
	"testing"
)

type failedReader struct{}

func (failedReader) Read([]byte) (int, error) { return 0, errors.New("read failure") }
func TestInputRefusals(t *testing.T) {
	for _, args := range [][]string{nil, {"extra"}} {
		var out, diagnostic bytes.Buffer
		if got := run(args, bytes.NewBufferString("{}"), &out, &diagnostic); got != 2 || out.Len() != 0 || diagnostic.Len() == 0 {
			t.Fatal("invalid request")
		}
	}
	if got := run(nil, failedReader{}, io.Discard, io.Discard); got != 2 {
		t.Fatal("reader error")
	}
}
