package clioutputacceptance_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/compiler"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/impact"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/intake"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/traceimpact"
)

// These tests run real entrypoints: run-function writer doubles cannot expose
// the Go runtime's handling of SIGPIPE on stdout.
func TestReviewCommandOutput(t *testing.T) {
	if runtime.Version() != compiler.ToolchainVersion {
		t.Fatalf("toolchain %s, want %s", runtime.Version(), compiler.ToolchainVersion)
	}
	model := readFixture(t, "../examples/lifecycle.json")
	diffRequest := marshalRequest(t, map[string]any{
		"formatVersion": impact.FormatVersion,
		"before":        json.RawMessage(model), "after": json.RawMessage(model),
		"bindings": []impact.Binding{},
	})
	traceSide := func(prefix string) map[string]string {
		return map[string]string{
			"model":        string(readFixture(t, "../traceimpact/testdata/"+prefix+"-model.json")),
			"catalog":      string(readFixture(t, "../traceimpact/testdata/"+prefix+"-catalog.json")),
			"requirements": string(readFixture(t, "../traceimpact/testdata/"+prefix+"-requirements.json")),
		}
	}
	traceRequest := marshalRequest(t, map[string]any{
		"formatVersion": traceimpact.RequestVersion,
		"before":        traceSide("before"), "after": traceSide("after"),
	})
	const intakeRequest = `{"formatVersion":"lifecycle-intake-json/v1","intent":"intent","decisions":[],"questions":[],"artifacts":{"requirements":null,"model":null,"scenarioSuite":null},"dependencies":{"modelRequirements":null,"suiteRequirements":null,"suiteModel":null},"observed":{"requirements":null,"model":null,"scenarioSuite":null}}`
	for _, tc := range []struct {
		name              string
		request           []byte
		execute           func([]byte) ([]byte, error)
		invalidDiagnostic string
	}{
		{"modeldiff", diffRequest, impact.Execute, "modeldiff: invalid model review request\n"},
		{"modelintake", []byte(intakeRequest), intake.Execute, "modelintake: invalid intake request\n"},
		{"modeltraceimpact", traceRequest, traceimpact.Execute, "modeltraceimpact: invalid request or report limit\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			want, err := tc.execute(tc.request)
			if err != nil || len(want) == 0 {
				t.Fatalf("native report: %v, bytes %d", err, len(want))
			}
			// Windows resolves executable extensions even for absolute paths.
			// The suffix is also harmless when running these binaries on Unix.
			bin := filepath.Join(t.TempDir(), tc.name+".exe")
			buildCtx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
			defer cancel()
			build := exec.CommandContext(buildCtx, filepath.Join(runtime.GOROOT(), "bin", "go"), "build", "-o", bin, "./cmd/"+tc.name)
			build.Dir = ".."
			build.Env = append(os.Environ(), "GOTOOLCHAIN=local", "GOPROXY=off", "GOSUMDB=off", "GOWORK=off", "GOENV=off", "GOFLAGS=")
			build.WaitDelay = 5 * time.Second
			if output, err := build.CombinedOutput(); err != nil {
				t.Fatalf("build: %v (context: %v)\n%s", err, buildCtx.Err(), output)
			}
			t.Run("complete_stdout", func(t *testing.T) {
				var stdout, stderr bytes.Buffer
				cmd := command(t, bin, tc.request)
				cmd.Stdout, cmd.Stderr = &stdout, &stderr
				if err := cmd.Run(); err != nil {
					t.Fatalf("run: %v; stderr %q", err, stderr.String())
				}
				if !bytes.Equal(stdout.Bytes(), want) || stderr.Len() != 0 {
					t.Fatalf("stdout differs from native API: got %d bytes, want %d; stderr %q", stdout.Len(), len(want), stderr.String())
				}
			})
			t.Run("malformed_input", func(t *testing.T) {
				var stdout, stderr bytes.Buffer
				cmd := command(t, bin, []byte("{"))
				cmd.Stdout, cmd.Stderr = &stdout, &stderr
				err := cmd.Run()
				if code := exitCode(err); code != 2 || stdout.Len() != 0 || stderr.String() != tc.invalidDiagnostic {
					t.Fatalf("exit %d (%v), want 2; stdout %q; stderr %q, want %q", code, err, stdout.String(), stderr.String(), tc.invalidDiagnostic)
				}
			})
			t.Run("closed_stdout_pipe", func(t *testing.T) {
				reader, writer, err := os.Pipe()
				if err != nil {
					t.Fatal(err)
				}
				defer writer.Close()
				// Close before launching, so no timing race can leave a reader.
				if err := reader.Close(); err != nil {
					t.Fatal(err)
				}
				var stderr bytes.Buffer
				cmd := command(t, bin, tc.request)
				// An *os.File is inherited as fd 1, not copied by os/exec.
				cmd.Stdout, cmd.Stderr = writer, &stderr
				err = cmd.Run()
				wantDiagnostic := tc.name + ": output write failed\n"
				if code := exitCode(err); code != 3 || stderr.String() != wantDiagnostic {
					t.Fatalf("exit %d (%v), want 3; stderr %q, want %q", code, err, stderr.String(), wantDiagnostic)
				}
			})
		})
	}
}

func command(t *testing.T, bin string, input []byte) *exec.Cmd {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	t.Cleanup(cancel)
	cmd := exec.CommandContext(ctx, bin)
	cmd.Stdin = bytes.NewReader(input)
	cmd.WaitDelay = time.Second
	return cmd
}

func exitCode(err error) int {
	if err == nil {
		return 0
	}
	var exit *exec.ExitError
	if errors.As(err, &exit) {
		return exit.ExitCode()
	}
	return -2
}

func readFixture(t *testing.T, path string) []byte {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func marshalRequest(t *testing.T, value any) []byte {
	t.Helper()
	b, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return b
}
