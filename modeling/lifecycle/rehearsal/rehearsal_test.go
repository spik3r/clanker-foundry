// Package rehearsal tests the fixed, explicitly local model-change walkthrough.
package rehearsal

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/canonical"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/compiler"
	lifecycle "github.com/spik3r/clanker-foundry/modeling/lifecycle/generated"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/jsonformat"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/modelpatch"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/scenarios"
)

func moduleRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot locate module")
	}
	return filepath.Dir(filepath.Dir(file))
}
func file(t *testing.T, path string) []byte {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return b
}
func checksum(b []byte) string { return fmt.Sprintf("%x", sha256.Sum256(b)) }
func snapshot(t *testing.T, root string) map[string]string {
	t.Helper()
	out := map[string]string{}
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !d.Type().IsRegular() {
			return fmt.Errorf("nonregular module input")
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		out[rel] = fmt.Sprintf("%o:%s", info.Mode().Perm(), checksum(b))
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return out
}
func cleanEnv(extra map[string]string) []string {
	out := []string{}
	for _, v := range os.Environ() {
		key, _, _ := strings.Cut(v, "=")
		if _, ok := extra[key]; !ok {
			out = append(out, v)
		}
	}
	for k, v := range extra {
		out = append(out, k+"="+v)
	}
	return out
}
func run(t *testing.T, dir string, env []string, stdin []byte, name string, args ...string) ([]byte, []byte, error) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = dir
	cmd.Env = env
	cmd.Stdin = bytes.NewReader(stdin)
	var out, diag bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &diag
	err := cmd.Run()
	if ctx.Err() != nil {
		t.Fatalf("command timed out: %s", name)
	}
	return out.Bytes(), diag.Bytes(), err
}
func environment(tmp string) []string {
	return cleanEnv(map[string]string{"TMPDIR": tmp, "GOTOOLCHAIN": "local", "GOPROXY": "off", "GOSUMDB": "off", "GOWORK": "off", "GOENV": "off", "GOFLAGS": "-buildvcs=false"})
}

func TestDocumentedRehearsalTwiceWithoutCheckoutMutation(t *testing.T) {
	if runtime.Version() != compiler.ToolchainVersion {
		t.Fatalf("requires %s, got %s", compiler.ToolchainVersion, runtime.Version())
	}
	root := moduleRoot(t)
	before := snapshot(t, root)
	temp := t.TempDir()
	poison := t.TempDir()
	marker := filepath.Join(t.TempDir(), "unexpected-python-startup")
	startup := "import os\nopen(os.environ['REHEARSAL_POISON_MARKER'], 'w').write('loaded')\nraise SystemExit(91)\n"
	if err := os.WriteFile(filepath.Join(poison, "sitecustomize.py"), []byte(startup), 0600); err != nil {
		t.Fatal(err)
	}
	env := cleanEnv(map[string]string{"TMPDIR": temp, "PYTHONPATH": poison, "REHEARSAL_POISON_MARKER": marker, "GOTOOLCHAIN": "local", "GOPROXY": "off", "GOSUMDB": "off", "GOWORK": "off", "GOENV": "off", "GOFLAGS": "-buildvcs=false"})
	script := filepath.Join(root, "rehearsal", "run.sh")
	oracle := file(t, filepath.Join(root, "rehearsal/testdata/expected.json"))
	if checksum(oracle) != "393b2080e0180581fff49857b08a5f9f86314447ee58211e7acf88b67bc0206f" {
		t.Fatal("frozen oracle bytes changed")
	}
	var expected struct {
		ExternalExpected struct {
			BeforeModelDigest    string
			CandidateModelDigest string
		}
		Baseline           struct{ Steps int }
		Candidate          struct{ Steps int }
		SuiteStatus        string
		SuiteBindingStatus string
	}
	if err := json.Unmarshal(oracle, &expected); err != nil {
		t.Fatal(err)
	}
	var first []byte
	for i := 0; i < 2; i++ {
		out, diag, err := run(t, root, env, nil, "bash", script)
		if err != nil || len(diag) != 0 {
			t.Fatalf("rehearsal %d failed: %v stderr=%q", i, err, diag)
		}
		if i == 0 {
			first = append([]byte{}, out...)
		} else if !bytes.Equal(first, out) {
			t.Fatal("rehearsal output not deterministic")
		}
	}
	var got map[string]any
	if err := json.Unmarshal(first, &got); err != nil {
		t.Fatal(err)
	}
	want := map[string]any{"formatVersion": "lifecycle-design-rehearsal-result/v1", "mode": "example-only", "authority": "none", "beforeModelDigest": expected.ExternalExpected.BeforeModelDigest, "candidateModelDigest": expected.ExternalExpected.CandidateModelDigest, "baselineMatched": true, "baselineSteps": float64(expected.Baseline.Steps), "candidateMatched": true, "candidateSteps": float64(expected.Candidate.Steps), "candidateBindingRefusedByBaseline": true, "baselineBindingRefusedByCandidate": true, "modelReviewRequired": true, "suiteReviewRequired": true, "topologyMode": "structural-only", "archiveConsistent": true, "archiveSuiteStatus": expected.SuiteStatus, "archiveBindingStatus": expected.SuiteBindingStatus}
	validHash := func(v any) bool {
		s, ok := v.(string)
		if !ok || len(s) != 64 || strings.ToLower(s) != s {
			return false
		}
		_, err := hex.DecodeString(s)
		return err == nil
	}
	if !validHash(got["archiveSHA256"]) {
		t.Fatal("archive hash missing or invalid")
	}
	hashes, ok := got["generatedArtifactSHA256"].(map[string]any)
	if !ok || len(hashes) != 5 {
		t.Fatal("generated artifact inventory missing")
	}
	for _, name := range []string{"generated/lifecycle.go", "generated/model_paths_test.go", "generated/decision.svg", "generated/decision-table.json", "generated/trace.json"} {
		if !validHash(hashes[name]) {
			t.Fatalf("invalid generated artifact hash: %s", name)
		}
	}
	delete(got, "archiveSHA256")
	delete(got, "generatedArtifactSHA256")
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("summary mismatch: got %#v want %#v", got, want)
	}
	if _, err := os.Stat(marker); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("ambient Python startup code ran")
	}
	after := snapshot(t, root)
	if !reflect.DeepEqual(before, after) {
		t.Fatal("checkout module changed")
	}
	entries, err := os.ReadDir(temp)
	if err != nil || len(entries) != 0 {
		t.Fatalf("private temp not cleaned: %v %v", entries, err)
	}
	out, _, err := run(t, root, env, nil, "bash", script, "unexpected")
	var exit *exec.ExitError
	if !errors.As(err, &exit) || exit.ExitCode() != 2 || len(out) != 0 {
		t.Fatal("arguments not refused without stdout")
	}
}

func TestStaleBindingsRefuseBeforeCallback(t *testing.T) {
	root := moduleRoot(t)
	candidate := file(t, filepath.Join(root, "designchange/testdata/suite.json"))
	baseline := file(t, filepath.Join(root, "rehearsal/testdata/baseline-suite.json"))
	calls := 0
	apply := func(lifecycle.Task, lifecycle.Command) lifecycle.Result {
		calls++
		panic("stale suite invoked callback")
	}
	before, err := scenarios.NewRunner(scenarios.Binding{ProfileVersion: lifecycle.ProfileVersion, ModelDigest: lifecycle.ModelDigest}, apply)
	if err != nil {
		t.Fatal(err)
	}
	after, err := scenarios.NewRunner(scenarios.Binding{ProfileVersion: lifecycle.ProfileVersion, ModelDigest: "a9ad2dfb926a24a026c8565f0d929a5b803bec43f6eb7342e196830791eccc7c"}, apply)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = before.Execute(candidate); !errors.Is(err, scenarios.ErrModelBinding) {
		t.Fatalf("candidate on baseline: %v", err)
	}
	if _, err = after.Execute(baseline); !errors.Is(err, scenarios.ErrModelBinding) {
		t.Fatalf("baseline on candidate: %v", err)
	}
	if calls != 0 {
		t.Fatal("stale bindings invoked callback")
	}
}

func TestFrozenCandidateKillsCompiledOwnerGuardMutant(t *testing.T) {
	root := moduleRoot(t)
	request := file(t, filepath.Join(root, "designchange/testdata/request.json"))
	raw, err := modelpatch.Execute(request)
	if err != nil {
		t.Fatal(err)
	}
	var result struct {
		After       json.RawMessage `json:"after"`
		AfterDigest string          `json:"afterDigest"`
	}
	if err = json.Unmarshal(raw, &result); err != nil {
		t.Fatal(err)
	}
	if result.AfterDigest != "a9ad2dfb926a24a026c8565f0d929a5b803bec43f6eb7342e196830791eccc7c" {
		t.Fatal("unexpected candidate identity")
	}
	m, err := jsonformat.Decode(result.After)
	if err != nil {
		t.Fatal(err)
	}
	digest, err := canonical.Digest(m)
	if err != nil || digest != result.AfterDigest {
		t.Fatal("candidate digest mismatch")
	}
	generated, _, err := compiler.Generate(m)
	if err != nil {
		t.Fatal(err)
	}
	old := []byte(`return nonempty(refString("task.owner", task, command, currentDigest))`)
	if bytes.Count(generated, old) != 1 {
		t.Fatal("owner-guard mutation target changed")
	}
	mutant := bytes.Replace(generated, old, []byte(`return true`), 1)
	dir := t.TempDir()
	paths := []string{"go.mod", "model/ir.go", "model/capabilities.go", "canonical/canonical.go", "strictjson/decode.go", "compiler/generate.go", "scenarios/types.go", "scenarios/decode.go", "scenarios/run.go", "scenarios/runner.go", "scenarios/report_decode.go", "cmd/modelplay/main.go"}
	for _, name := range paths {
		target := filepath.Join(dir, name)
		if err = os.MkdirAll(filepath.Dir(target), 0700); err != nil {
			t.Fatal(err)
		}
		if err = os.WriteFile(target, file(t, filepath.Join(root, name)), 0600); err != nil {
			t.Fatal(err)
		}
	}
	generatedPath := filepath.Join(dir, "generated/lifecycle.go")
	if err = os.MkdirAll(filepath.Dir(generatedPath), 0700); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(generatedPath, mutant, 0600); err != nil {
		t.Fatal(err)
	}
	temp := t.TempDir()
	env := environment(temp)
	binary := filepath.Join(dir, "modelplay-mutant")
	goexe := filepath.Join(runtime.GOROOT(), "bin", "go")
	if out, diag, e := run(t, dir, env, nil, goexe, "build", "-o", binary, "./cmd/modelplay"); e != nil {
		t.Fatalf("mutant did not compile: %v %s %s", e, out, diag)
	}
	suite := file(t, filepath.Join(root, "designchange/testdata/suite.json"))
	out, _, err := run(t, dir, env, suite, binary)
	var exit *exec.ExitError
	if !errors.As(err, &exit) || exit.ExitCode() != 1 {
		t.Fatalf("compiled mutant not rejected: %v", err)
	}
	var report struct {
		Matched       bool `json:"matched"`
		StepsExecuted int  `json:"stepsExecuted"`
		Entries       []struct {
			Actual struct {
				OK           bool   `json:"ok"`
				TransitionID string `json:"transitionId"`
			} `json:"actual"`
		} `json:"entries"`
	}
	if err = json.Unmarshal(out, &report); err != nil {
		t.Fatal(err)
	}
	if report.Matched || report.StepsExecuted != 1 || len(report.Entries) != 1 || !report.Entries[0].Actual.OK || report.Entries[0].Actual.TransitionID != "T_prepare" {
		t.Fatal("wrong mutant failure")
	}
}

func TestCleanupFailureNeverPublishesSuccess(t *testing.T) {
	root := moduleRoot(t)
	temp := t.TempDir()
	marker := filepath.Join(t.TempDir(), "cleanup-reached")
	script := file(t, filepath.Join(root, "rehearsal/run.sh"))
	begin := []byte("exec python3 -I - \"$module\" <<'PY'\n")
	end := []byte("\nPY\n")
	i := bytes.Index(script, begin)
	j := bytes.LastIndex(script, end)
	if i < 0 || j < i+len(begin) {
		t.Fatal("Python program boundary changed")
	}
	program := append([]byte{}, script[i+len(begin):j]...)
	anchor := []byte("    with tempfile.TemporaryDirectory(prefix='foundry-design-rehearsal-') as tmp:\n")
	if bytes.Count(program, anchor) != 1 {
		t.Fatal("cleanup injection anchor changed")
	}
	markerLiteral, err := json.Marshal(marker)
	if err != nil {
		t.Fatal(err)
	}
	// Fault-inject a test-only copy of the exact program, after real cleanup.
	// Production receives no injection hook and Python isolated mode ignores
	// ambient PYTHONPATH and user site customization.
	injection := fmt.Sprintf(`    _original_cleanup = tempfile.TemporaryDirectory.cleanup
    def _fail_after_cleanup(self):
        _original_cleanup(self)
        pathlib.Path(%s).write_text('cleaned', encoding='ascii')
        raise OSError('private injected cleanup failure')
    tempfile.TemporaryDirectory.cleanup = _fail_after_cleanup
`, markerLiteral)
	program = bytes.Replace(program, anchor, append([]byte(injection), anchor...), 1)
	out, diag, err := run(t, root, environment(temp), program, "python3", "-I", "-", root)
	var exit *exec.ExitError
	if !errors.As(err, &exit) || exit.ExitCode() != 1 || len(out) != 0 || string(diag) != "rehearsal: local operation failed\n" {
		t.Fatalf("cleanup failure: err=%v stdout=%q stderr=%q", err, out, diag)
	}
	if string(file(t, marker)) != "cleaned" {
		t.Fatal("cleanup fault not reached")
	}
	entries, err := os.ReadDir(temp)
	if err != nil || len(entries) != 0 {
		t.Fatalf("cleanup not completed: %v %v", entries, err)
	}
}
