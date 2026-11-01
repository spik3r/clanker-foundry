package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/intakemodel"
)

func requests(t *testing.T) []string {
	t.Helper()
	b, err := os.ReadFile("../../intakemodel/testdata/expected.json")
	if err != nil {
		t.Fatal(err)
	}
	var f struct {
		Cases []struct {
			RequestJSON string `json:"requestJSON"`
		} `json:"cases"`
	}
	if err = json.Unmarshal(b, &f); err != nil {
		t.Fatal(err)
	}
	var out []string
	for _, c := range f.Cases {
		out = append(out, c.RequestJSON)
	}
	return out
}

type badReader struct{}

func (badReader) Read([]byte) (int, error) { return 0, errors.New("private read details") }

type badWriter struct{ short bool }

func (w badWriter) Write(p []byte) (int, error) {
	if w.short {
		return len(p) - 1, nil
	}
	return 1, errors.New("private write details")
}

func TestCLIEmitsCompleteReportsIncludingRejectedAndMissing(t *testing.T) {
	for _, raw := range requests(t) {
		var out, diag bytes.Buffer
		if code := run(nil, strings.NewReader(raw), &out, &diag); code != 0 || diag.Len() != 0 {
			t.Fatal(code, diag.String())
		}
		want, err := intakemodel.Execute([]byte(raw))
		if err != nil || !bytes.Equal(out.Bytes(), want) {
			t.Fatal("CLI report mismatch", err)
		}
	}
}
func TestCLIRefusalsAndOutputFailures(t *testing.T) {
	for _, tc := range []struct {
		args   []string
		reader io.Reader
	}{
		{[]string{"validate"}, strings.NewReader("private")}, {nil, badReader{}}, {nil, strings.NewReader("")}, {nil, strings.NewReader(`{"private":"data"}`)}, {nil, strings.NewReader(strings.Repeat(" ", intakemodel.MaxSourceBytes+1))},
	} {
		var out, diag bytes.Buffer
		if code := run(tc.args, tc.reader, &out, &diag); code != 2 || out.Len() != 0 || diag.Len() == 0 {
			t.Fatal(code, out.String(), diag.String())
		}
		if strings.Contains(diag.String(), "private") {
			t.Fatal("caller text leaked")
		}
	}
	for _, short := range []bool{false, true} {
		var diag bytes.Buffer
		if code := run(nil, strings.NewReader(requests(t)[0]), badWriter{short}, &diag); code != 3 || diag.String() != "modelintakecheck: output write failed\n" {
			t.Fatal(code, diag.String())
		}
	}
}

func TestCLIClosedOutputPipe(t *testing.T) {
	const helper = "FOUNDRY_INTAKECHECK_CLOSED_PIPE_TEST"
	if os.Getenv(helper) == "1" {
		os.Args = os.Args[:1]
		main()
		return
	}
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := reader.Close(); err != nil {
		t.Fatal(err)
	}
	defer writer.Close()
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(executable, "-test.run=^TestCLIClosedOutputPipe$")
	cmd.Env = append(os.Environ(), helper+"=1")
	cmd.Stdin = strings.NewReader(requests(t)[0])
	cmd.Stdout = writer
	var diagnostic bytes.Buffer
	cmd.Stderr = &diagnostic
	err = cmd.Run()
	var exit *exec.ExitError
	if !errors.As(err, &exit) || exit.ExitCode() != 3 || diagnostic.String() != "modelintakecheck: output write failed\n" {
		t.Fatalf("closed pipe: error=%v diagnostic=%q", err, diagnostic.String())
	}
}
