package operatorcli_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"testing"
	"time"
)

func compileCommands(t *testing.T) (string, map[string]string) {
	t.Helper()
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	binaries := map[string]string{}
	for _, name := range []string{"foundry", "modelreviewcheck", "modelproposal", "modelprocess", "modelrehearsalinspect", "modelrevisioncheck"} {
		ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
		bin := filepath.Join(t.TempDir(), name)
		cmd := exec.CommandContext(ctx, filepath.Join(runtime.GOROOT(), "bin", "go"), "build", "-o", bin, "./cmd/"+name)
		cmd.Dir = root
		out, err := cmd.CombinedOutput()
		cancel()
		if err != nil {
			t.Fatalf("build %s: %v %s", name, err, out)
		}
		binaries[name] = bin
	}
	return root, binaries
}
func binaryResult(t *testing.T, root, bin string, args []string, input []byte, closed bool) (int, []byte, string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, bin, args...)
	cmd.Dir = root
	cmd.Stdin = bytes.NewReader(input)
	cmd.WaitDelay = time.Second
	var out, diag bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &diag
	if closed {
		r, w, err := os.Pipe()
		if err != nil {
			t.Fatal(err)
		}
		defer w.Close()
		if err := r.Close(); err != nil {
			t.Fatal(err)
		}
		cmd.Stdout = w
	}
	err := cmd.Run()
	if ctx.Err() != nil {
		t.Fatal("binary deadline", ctx.Err())
	}
	code := 0
	if err != nil {
		var exit *exec.ExitError
		if !errors.As(err, &exit) {
			t.Fatal(err)
		}
		code = exit.ExitCode()
	}
	return code, out.Bytes(), diag.String()
}
func TestCompiledNativeEntrypointsAndFoundryKeepFullOutputsAndPipeStatuses(t *testing.T) {
	root, binaries := compileCommands(t)
	review, proposal := nativeInputs(t)
	c := vectors(t)
	var revision struct{ Vectors []struct{ Request string } }
	if json.Unmarshal(read(t, "../../revisioncheck/testdata/observations.json"), &revision) != nil {
		t.Fatal("revision fixture")
	}
	routes := []struct {
		name                    string
		nativeArgs, foundryArgs []string
		input                   []byte
		pipeDiagnostic          string
	}{
		{"modelreviewcheck", nil, []string{"review"}, review, "modelreviewcheck: report delivery failed\n"},
		{"modelproposal", []string{"check"}, []string{"proposal", "check"}, proposal, "modelproposal: receipt delivery failed; execution may have started\n"},
		{"modelprocess", []string{"json"}, []string{"process", "json"}, read(t, "../../processview/testdata/model.json"), "modelprocess: output write failed\n"},
		{"modelprocess", []string{"svg"}, []string{"process", "svg"}, read(t, "../../processview/testdata/model.json"), "modelprocess: output write failed\n"},
		{"modelrehearsalinspect", nil, []string{"inspect"}, []byte(c.InspectionInput), "inspection output failed\n"},
		{"modelrevisioncheck", nil, []string{"revisions"}, []byte(revision.Vectors[0].Request), "modelrevisioncheck: output failed\n"},
	}
	for _, v := range routes {
		t.Run(strings.Join(v.foundryArgs, "-"), func(t *testing.T) {
			oldCode, oldOut, oldDiag := binaryResult(t, root, binaries[v.name], v.nativeArgs, v.input, false)
			code, out, diag := binaryResult(t, root, binaries["foundry"], v.foundryArgs, v.input, false)
			if oldCode != 0 || code != oldCode || oldDiag != "" || diag != oldDiag || !bytes.Equal(out, oldOut) {
				t.Fatalf("entrypoint output parity old%d new%d diag%q/%q", oldCode, code, oldDiag, diag)
			}
			for _, bin := range []struct {
				path string
				args []string
			}{{binaries[v.name], v.nativeArgs}, {binaries["foundry"], v.foundryArgs}} {
				code, out, diag := binaryResult(t, root, bin.path, bin.args, v.input, true)
				if code != 3 || len(out) != 0 || diag != v.pipeDiagnostic {
					t.Fatalf("closed stdout code%d diag%q", code, diag)
				}
			}
		})
	}
	for _, v := range c.InvalidNative {
		code, out, diag := binaryResult(t, root, binaries["foundry"], v.Args, []byte(v.Input), false)
		if code != v.Status || len(out) != 0 || diag != v.Diagnostic {
			t.Fatalf("compiled refusal %s: code%d diag%q", v.ID, code, diag)
		}
	}
	// A FIFO must be refused before Open could block. The actual child timeout
	// also bounds this regression if the regular-file admission check is removed.
	fixture := c.Prepare[0]
	args, paths := prepareFiles(t, fixture.Intake, fixture.Artifacts)
	if err := os.Remove(paths["model"]); err != nil {
		t.Fatal(err)
	}
	if err := syscall.Mkfifo(paths["model"], 0600); err != nil {
		t.Fatal(err)
	}
	code, out, diag := binaryResult(t, root, binaries["foundry"], args, nil, false)
	if code != 2 || len(out) != 0 || diag != "foundry prepare: input rejected\n" {
		t.Fatal("FIFO admitted", code, diag)
	}
}
