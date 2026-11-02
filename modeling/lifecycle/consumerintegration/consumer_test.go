package consumerintegration_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/compiler"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/jsonformat"
	"go/format"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"
)

const oracleSHA256 = "119a37c0298dea6e122c3a4f62b2cb1bddbac24e6fc91feadbc2723360e9550f"

func read(t *testing.T, p string) []byte {
	t.Helper()
	b, e := os.ReadFile(p)
	if e != nil {
		t.Fatal(e)
	}
	return b
}
func write(t *testing.T, p string, b []byte) {
	t.Helper()
	if e := os.MkdirAll(filepath.Dir(p), 0700); e != nil {
		t.Fatal(e)
	}
	if e := os.WriteFile(p, b, 0600); e != nil {
		t.Fatal(e)
	}
}
func goTest(t *testing.T, root string, args ...string) ([]byte, error) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, filepath.Join(runtime.GOROOT(), "bin", "go"), append([]string{"test", "-buildvcs=false", "-count=1"}, args...)...)
	cmd.Dir = root
	cmd.WaitDelay = 2 * time.Second
	// Keep the installed cache; prevent workspace, module and toolchain downloads.
	blocked := map[string]bool{"GOWORK": true, "GOTOOLCHAIN": true, "GOPROXY": true, "GOSUMDB": true, "GOFLAGS": true, "GOENV": true}
	for _, v := range os.Environ() {
		k, _, _ := strings.Cut(v, "=")
		if !blocked[k] {
			cmd.Env = append(cmd.Env, v)
		}
	}
	cmd.Env = append(cmd.Env, "GOWORK=off", "GOTOOLCHAIN=local", "GOPROXY=off", "GOSUMDB=off", "GOFLAGS=", "GOENV=off")
	out, e := cmd.CombinedOutput()
	if ctx.Err() != nil {
		t.Fatalf("external consumer timed out: %s", out)
	}
	return out, e
}
func TestExternalConsumerBoundModels(t *testing.T) {
	if runtime.Version() != compiler.ToolchainVersion {
		t.Fatalf("requires actual %s, have %s", compiler.ToolchainVersion, runtime.Version())
	}
	_, source, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("source location unavailable")
	}
	here := filepath.Dir(source)
	module := filepath.Dir(here)
	oracle := read(t, filepath.Join(here, "testdata", "expected.json"))
	h := sha256.Sum256(oracle)
	if hex.EncodeToString(h[:]) != oracleSHA256 {
		t.Fatal("frozen oracle bytes changed")
	}
	root := t.TempDir()
	write(t, filepath.Join(root, "go.mod"), []byte("module example.invalid/foundry-consumer-proof\n\ngo 1.27.3\n\nrequire github.com/spik3r/clanker-foundry/modeling/lifecycle v0.0.0\nreplace github.com/spik3r/clanker-foundry/modeling/lifecycle => "+strconv.Quote(module)+"\n"))
	for _, name := range []string{"requirements.md", "expected.json", "model-v1.json", "model-v2.json", "suite-v1.json", "suite-v2.json", "manifest-v1.json", "manifest-v2.json"} {
		write(t, filepath.Join(root, "fixtures", name), read(t, filepath.Join(here, "testdata", name)))
	}
	// This separately authored explanation oracle is new; all eight prior fixtures
	// retain their original bytes and independent inventory.
	projection := read(t, filepath.Join(module, "scenarioexplain", "testdata", "bound-vectors.json"))
	projectionHash := sha256.Sum256(projection)
	if hex.EncodeToString(projectionHash[:]) != "c5e011aba6dd14f5eaf2de0016de10dd1359196decaf0eb8d6edf4463fc147f9" {
		t.Fatal("bound explanation oracle bytes changed")
	}
	write(t, filepath.Join(root, "fixtures", "bound-explanation-vectors.json"), projection)
	template := string(read(t, filepath.Join(here, "testdata", "adapter.go.txt")))
	sources := map[string][]byte{}
	adapters := map[string][]byte{}
	for i := 1; i <= 2; i++ {
		version := fmt.Sprintf("v%d", i)
		ir, e := jsonformat.Decode(read(t, filepath.Join(root, "fixtures", "model-"+version+".json")))
		if e != nil {
			t.Fatal(e)
		}
		implementation, derived, e := compiler.Generate(ir)
		if e != nil {
			t.Fatal(e)
		}
		sources[version] = implementation
		write(t, filepath.Join(root, "reducer_"+version, "lifecycle.go"), implementation)
		write(t, filepath.Join(root, "reducer_"+version, "derived_test.go"), derived)
		adapter := strings.NewReplacer("__MODEL_IMPORT__", "example.invalid/foundry-consumer-proof/reducer_"+version, "__VERSION__", version, "__OTHER_VERSION__", fmt.Sprintf("v%d", 3-i)).Replace(template)
		formatted, e := format.Source([]byte(adapter))
		if e != nil {
			t.Fatal("adapter template syntax", e)
		}
		write(t, filepath.Join(root, "proof"+version, "adapter_test.go"), formatted)
		adapters[version] = formatted
	}
	if out, e := goTest(t, root, "-race", "./..."); e != nil {
		t.Fatalf("external consumer baseline failed: %v\n%s", e, out)
	}
	// Mutate only disposable generated code; compile errors do not count as kills.
	mutations := []struct{ name, old, replacement string }{
		{"always-refuse", "func Transition(task Task, command Command) Result {", "func Transition(task Task, command Command) Result {\nreturn Result{Mode: \"example-only\", Authority: \"none\", ModelDigest: ModelDigest, Refusals: []string{\"INVALID_INPUT\"}}"},
		{"removed-submit-guard", `failed = append(failed, "G_submitActor")`, `_ = "G_submitActor"`},
	}
	target := filepath.Join(root, "reducer_v2", "lifecycle.go")
	original := sources["v2"]
	for _, m := range mutations {
		t.Run(m.name, func(t *testing.T) {
			if strings.Count(string(original), m.old) != 1 {
				t.Fatal("mutation target changed; review control")
			}
			write(t, target, []byte(strings.Replace(string(original), m.old, m.replacement, 1)))
			defer write(t, target, original)
			out, e := goTest(t, root, "./proofv2", "-run", "^TestFrozenReplay$")
			if e == nil || !bytes.Contains(out, []byte("FROZEN_ORACLE_MISMATCH")) || bytes.Contains(out, []byte("build failed")) {
				t.Fatalf("mutation was not behaviorally killed: %v\n%s", e, out)
			}
		})
	}
	if out, e := goTest(t, root, "./proofv2", "-run", "^TestFrozenReplay$"); e != nil {
		t.Fatalf("restored reducer failed: %v\n%s", e, out)
	}
	// These change only disposable adapter code, not fixtures or reducers. Each
	// generated type family must actually run conformance and reject the mutant.
	type edit struct {
		old, replacement string
		count            int
	}
	adapterMutations := []struct {
		name  string
		edits []edit
	}{
		{"coupled-task-swap", []edit{{"ID: t.ID,", "ID: t.Outcome,", 2}, {"Outcome: t.Outcome,", "Outcome: t.ID,", 2}}},
		{"drop-command-actor", []edit{{"Actor: c.Actor,", "Actor: \"\",", 1}}},
		{"alias-task-slice", []edit{{"stringsCopy(t.Acceptance)", "t.Acceptance", 2}}},
		{"alias-task-owner", []edit{{"ownerCopy(t.Owner)", "t.Owner", 2}}},
		{"alias-review", []edit{{"out.Review = &custom.Review{TaskID: r.TaskID, TaskRevision: r.TaskRevision, ModelDigest: r.ModelDigest, TaskDigest: r.TaskDigest, Reviewer: r.Reviewer, Decision: r.Decision}", "_ = r; out.Review = (*custom.Review)(c.Review)", 1}}},
		{"alias-result-task-slice", []edit{{"t := fromTask(*r.Task)", "t := fromTask(*r.Task); t.Acceptance = r.Task.Acceptance", 1}}},
		{"alias-result-refusals", []edit{{"stringsCopy(r.Refusals)", "r.Refusals", 1}}},
		{"cross-field-refusals", []edit{{"out.Task = &t\n\t}\n\treturn out", "out.Task = &t\n\t}\n\tif r.Task != nil && r.Refusals != nil && len(r.Refusals) == 0 && r.Task.Acceptance != nil && len(r.Task.Acceptance) == 0 { out.Refusals = r.Task.Acceptance }\n\treturn out", 1}}},
		{"cross-kind-owner", []edit{{"out.Task = &t\n\t}\n\treturn out", "out.Task = &t\n\t}\n\tif r.Task != nil && r.Task.Owner != nil && *r.Task.Owner == \"\" && r.Task.Acceptance != nil && len(r.Task.Acceptance) == 0 && cap(r.Task.Acceptance) > 0 { out.Task.Owner = &r.Task.Acceptance[:cap(r.Task.Acceptance)][0] }\n\treturn out", 1}}},
	}
	for _, version := range []string{"v1", "v2"} {
		path := filepath.Join(root, "proof"+version, "adapter_test.go")
		original := adapters[version]
		for _, m := range adapterMutations {
			t.Run(version+"/"+m.name, func(t *testing.T) {
				mutant := string(original)
				for _, edit := range m.edits {
					if strings.Count(mutant, edit.old) != edit.count {
						t.Fatal("adapter mutation target changed; review control")
					}
					mutant = strings.ReplaceAll(mutant, edit.old, edit.replacement)
				}
				write(t, path, []byte(mutant))
				defer write(t, path, original)
				out, e := goTest(t, root, "./proof"+version, "-run", "^TestAdapterConformance$")
				if e == nil || !bytes.Contains(out, []byte("ADAPTER_CONFORMANCE_MISMATCH:")) ||
					!bytes.Contains(out, []byte("--- FAIL: TestAdapterConformance")) || bytes.Contains(out, []byte("build failed")) {
					t.Fatalf("adapter mutation was not behaviorally killed: %v\n%s", e, out)
				}
			})
		}
		if out, e := goTest(t, root, "./proof"+version, "-run", "^TestAdapterConformance$"); e != nil {
			t.Fatalf("restored %s adapter conformance failed: %v\n%s", version, e, out)
		}
	}
	// After all reducer/adapter mutations restore, actually replay and explain
	// both models again. A build-only success is not this evidence.
	for _, version := range []string{"v1", "v2"} {
		out, err := goTest(t, root, "-v", "./proof"+version, "-run", "^TestFrozenReplay$")
		if err != nil || !bytes.Contains(out, []byte("BOUND_EXPLANATION_MATCH: "+version)) {
			t.Fatalf("restored %s explanation was not executed: %v\n%s", version, err, out)
		}
	}
}
