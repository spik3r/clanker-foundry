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
	template := string(read(t, filepath.Join(here, "testdata", "adapter.go.txt")))
	sources := map[string][]byte{}
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
}
