package main

import (
	"bytes"
	"errors"
	"testing"
)

type stdoutResultWriter struct {
	count     int
	err       error
	delivered bytes.Buffer
}

func (w *stdoutResultWriter) Write(p []byte) (int, error) {
	n := min(w.count, len(p))
	w.delivered.Write(p[:n])
	return n, w.err
}

func TestStdoutDelivery(t *testing.T) {
	args := []string{"validate", "../../testdata/reference/model.json"}
	var complete, diagnostic bytes.Buffer
	if code := run(args, &complete, &diagnostic); code != 0 || complete.Len() == 0 || diagnostic.Len() != 0 {
		t.Fatalf("complete delivery: code=%d stdout=%q stderr=%q", code, complete.String(), diagnostic.String())
	}
	failure := errors.New("synthetic output failure")
	for _, tc := range []struct {
		name  string
		count int
		err   error
	}{
		{"error", 0, failure},
		{"short without error", 1, nil},
		{"partial with error", 1, failure},
		{"full count with error", complete.Len(), failure},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := &stdoutResultWriter{count: tc.count, err: tc.err}
			var stderr bytes.Buffer
			code := run(args, w, &stderr)
			if code != 1 || stderr.String() != "refgen: output write failed\n" {
				t.Fatalf("code=%d stderr=%q; want exit 1 and fixed output-failure diagnostic", code, stderr.String())
			}
			if !bytes.Equal(w.delivered.Bytes(), complete.Bytes()[:tc.count]) {
				t.Fatalf("delivered prefix changed: %q", w.delivered.String())
			}
		})
	}
}
