package mergerehearsal

import (
	"bytes"
	"context"
	"crypto/sha256"
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

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/compiler"
	transport "github.com/spik3r/clanker-foundry/modeling/lifecycle/generated"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/scenarios"
)

func root(t *testing.T) string {
	t.Helper()
	_, source, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("source location unavailable")
	}
	return filepath.Dir(filepath.Dir(source))
}

func read(t *testing.T, path string) []byte {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func digest(b []byte) string { return fmt.Sprintf("%x", sha256.Sum256(b)) }

func snapshot(t *testing.T, dir string) map[string]string {
	t.Helper()
	result := map[string]string{}
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("nonregular source: %s", path)
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		result[rel] = fmt.Sprintf("%o:%s", info.Mode().Perm(), digest(read(t, path)))
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func environment(temp string, extra map[string]string) []string {
	replacements := map[string]string{
		"PATH":   filepath.Join(runtime.GOROOT(), "bin") + string(os.PathListSeparator) + os.Getenv("PATH"),
		"TMPDIR": temp, "GOTMPDIR": temp, "GOTOOLCHAIN": "local", "GOPROXY": "off", "GOSUMDB": "off",
		"GOWORK": "off", "GOENV": "off", "GOFLAGS": "-buildvcs=false", "GOMAXPROCS": "2",
	}
	for k, v := range extra {
		replacements[k] = v
	}
	var env []string
	for _, value := range os.Environ() {
		k, _, _ := strings.Cut(value, "=")
		if _, exists := replacements[k]; !exists {
			env = append(env, value)
		}
	}
	for k, v := range replacements {
		env = append(env, k+"="+v)
	}
	return env
}

func command(t *testing.T, dir string, env []string, name string, args ...string) ([]byte, []byte, error) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir, cmd.Env = dir, env
	cmd.WaitDelay = 2 * time.Second
	var out, diag bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &diag
	err := cmd.Run()
	if ctx.Err() != nil {
		t.Fatalf("command deadline: %s", diag.Bytes())
	}
	return out.Bytes(), diag.Bytes(), err
}

func TestDocumentedDemoTwice(t *testing.T) {
	if runtime.Version() != compiler.ToolchainVersion {
		t.Fatalf("requires %s", compiler.ToolchainVersion)
	}
	module := root(t)
	before := snapshot(t, module)
	temp := t.TempDir()
	poison := t.TempDir()
	marker := filepath.Join(poison, "loaded")
	if err := os.WriteFile(filepath.Join(poison, "sitecustomize.py"), []byte("open("+fmt.Sprintf("%q", marker)+",'w').write('loaded')\nraise SystemExit(92)\n"), 0600); err != nil {
		t.Fatal(err)
	}
	env := environment(temp, map[string]string{"PYTHONPATH": poison, "GOFLAGS": "-invalid-ambient-flag"})
	script := filepath.Join(module, "mergerehearsal", "run.sh")
	var prior []byte
	for i := 0; i < 2; i++ {
		out, diag, err := command(t, module, env, "bash", script)
		if err != nil {
			t.Fatalf("actual demo failed: %v %s", err, diag)
		}
		if i > 0 && !bytes.Equal(out, prior) {
			t.Fatal("complete receipts differ")
		}
		prior = out
		var result struct {
			FormatVersion, Mode, Authority, Cleanup, Inputs string
			GenerationRuns, ReplayRuns                      int
			Merges                                          []struct{ Status, Generation, Replay string }
			Revisions                                       map[string]struct{ Cases, Steps int }
			Controls                                        map[string]string
		}
		if err := json.Unmarshal(out, &result); err != nil {
			t.Fatal(err)
		}
		if result.FormatVersion != "lifecycle-merge-rehearsal-result/v1" || result.Mode != "example-only" || result.Authority != "none" || result.Cleanup != "completed" || result.Inputs != "unchanged" || result.GenerationRuns != 6 || result.ReplayRuns != 6 || len(result.Merges) != 6 || len(result.Revisions) != 4 || len(result.Controls) != 4 {
			t.Fatalf("bad receipt: %s", out)
		}
		for _, m := range result.Merges {
			if m.Status == "merged" {
				if m.Generation != "observed" || m.Replay != "observed" {
					t.Fatal("merged replay absent")
				}
			} else if m.Generation != "not-run" || m.Replay != "not-run" {
				t.Fatal("conflict continued")
			}
		}
		for _, r := range result.Revisions {
			if r.Cases != 7 || r.Steps != 9 {
				t.Fatal("incomplete frozen replay")
			}
		}
		entries, err := os.ReadDir(temp)
		if err != nil || len(entries) != 0 {
			t.Fatalf("temporary residue: %v %v", entries, err)
		}
	}
	if !reflect.DeepEqual(before, snapshot(t, module)) {
		t.Fatal("checkout bytes or modes changed")
	}
	if _, err := os.Stat(marker); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("ambient Python startup ran")
	}
	out, diag, err := command(t, module, env, "bash", script, "unexpected")
	var exit *exec.ExitError
	if !errors.As(err, &exit) || exit.ExitCode() != 2 || len(out) != 0 || string(diag) != "merge rehearsal: no arguments accepted\n" {
		t.Fatalf("argument refusal: %q %q %v", out, diag, err)
	}
}

func TestEveryStalePairHasZeroCallbacks(t *testing.T) {
	module := root(t)
	var expected struct{ ModelDigests map[string]string }
	if err := json.Unmarshal(read(t, filepath.Join(module, "mergerehearsal/testdata/expected.json")), &expected); err != nil {
		t.Fatal(err)
	}
	for _, binding := range []string{"base", "left", "right", "merged"} {
		calls := 0
		runner, err := scenarios.NewRunner(scenarios.Binding{ProfileVersion: "lifecycle/v1", ModelDigest: expected.ModelDigests[binding]}, func(transport.Task, transport.Command) transport.Result { calls++; panic("stale callback") })
		if err != nil {
			t.Fatal(err)
		}
		for _, suite := range []string{"base", "left", "right", "merged"} {
			if suite == binding {
				continue
			}
			raw := read(t, filepath.Join(module, "mergerehearsal/testdata/suite-"+suite+".json"))
			before := append([]byte(nil), raw...)
			if _, err = runner.Execute(raw); !errors.Is(err, scenarios.ErrModelBinding) {
				t.Fatalf("%s on %s: %v", suite, binding, err)
			}
			if !bytes.Equal(raw, before) || calls != 0 {
				t.Fatal("stale input mutated or invoked callback")
			}
		}
	}
}

// Execute the real Python control flow with explicit host-side test doubles.
// These are not environment-selected product modes or arbitrary command hooks.
func TestFailureAndShortCircuitControls(t *testing.T) {
	module := root(t)
	code := `import json,os,pathlib,subprocess,sys,tempfile
path=pathlib.Path(sys.argv[1]); ns={'__name__':'rehearsal_test','__file__':str(path)}
exec(compile(path.read_bytes(),str(path),'exec'),ns)
source=path.parent.parent; expected=json.loads((source/'mergerehearsal/testdata/expected.json').read_bytes())
W,F=ns['Workflow'],ns['Failure']
w=W(source,pathlib.Path(tempfile.gettempdir()),'go',dict(os.environ),expected)
w.replay_model=lambda *a: (_ for _ in ()).throw(AssertionError('forbidden generation/replay'))
for case in expected['merges']:
    if case['expected']['status']=='merged': continue
    calls=[]
    def invoke(stage,args,data=None,code=0):
        calls.append(args)
        assert pathlib.Path(args[0]).name=='modelmerge'
        return json.dumps(case['expected']).encode()
    w.invoke=invoke
    result=w.merge_case(case)
    assert len(calls)==1 and result['generation']==result['replay']=='not-run'
    assert (w.generated,w.replayed)==(0,0)
# Do not accept a matching status with a corrupted complete report.
case=expected['merges'][2]; changed=dict(case['expected'],conflicts=[])
w.invoke=lambda *a,**kw:json.dumps(changed).encode()
try: w.merge_case(case); raise AssertionError('accepted altered full report')
except F: pass
# Cleanup must actually remove the temporary directory and then fail. No success
# byte may escape main even if the workflow already returned a success object.
removed=[]
class Cleanup:
    def __init__(self,**kw): self.ctx=tempfile.TemporaryDirectory(**kw)
    def __enter__(self): self.name=self.ctx.__enter__(); return self.name
    def __exit__(self,*args):
        self.ctx.__exit__(*args); removed.append(not pathlib.Path(self.name).exists())
        raise OSError('private cleanup detail')
class NoWork:
    def __init__(self,*args): pass
    def run(self): return {'forbidden':'success'}
perform=ns['perform']
ns['perform']=lambda src,env:perform(src,env,temporary_directory=Cleanup,workflow=NoWork)
sys.argv=[str(path)]
assert ns['main']()==1 and removed==[True]
# Ordinary tool failure, launch failure, and timeout produce fixed stages.
w=W(source,pathlib.Path(tempfile.gettempdir()),'go',dict(os.environ),expected)
original=subprocess.run
for failure in [OSError('private launch detail'),subprocess.TimeoutExpired('private command',1)]:
    def failed(*a,**kw): raise failure
    subprocess.run=failed
    try: w.invoke('fixed stage',['go']); raise AssertionError('failure swallowed')
    except F as exc: assert str(exc)=='fixed stage'
subprocess.run=lambda *a,**kw:subprocess.CompletedProcess([],99,b'private output',b'private diag')
try: w.invoke('fixed stage',['go']); raise AssertionError('bad exit accepted')
except F as exc: assert str(exc)=='fixed stage'
subprocess.run=original
`
	temp := t.TempDir()
	out, diag, err := command(t, module, environment(temp, nil), "python3", "-I", "-c", code, filepath.Join(module, "mergerehearsal/run.py"))
	if err != nil || len(out) != 0 || string(diag) != "merge rehearsal: local operation failed\n" {
		t.Fatalf("failure controls: %v stdout=%q stderr=%q", err, out, diag)
	}
	entries, err := os.ReadDir(temp)
	if err != nil || len(entries) != 0 {
		t.Fatalf("cleanup test residue: %v %v", entries, err)
	}
}
