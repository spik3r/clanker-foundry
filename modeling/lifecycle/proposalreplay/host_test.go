package proposalreplay

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/artifactbundle"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/compiler"
)

func TestWritablePathsStayOutsideCheckout(t *testing.T) {
	base := t.TempDir()
	checkout := filepath.Join(base, "checkout")
	module := filepath.Join(checkout, "modeling", "lifecycle")
	temp := filepath.Join(base, "temp")
	for _, p := range []string{module, temp} {
		if err := os.MkdirAll(p, 0700); err != nil {
			t.Fatal(err)
		}
	}
	// A Git worktree marker need not be a directory.
	if err := os.WriteFile(filepath.Join(checkout, ".git"), []byte("gitdir: elsewhere\n"), 0600); err != nil {
		t.Fatal(err)
	}
	alias := filepath.Join(base, "alias")
	if err := os.Symlink(checkout, alias); err != nil {
		t.Fatal(err)
	}
	dangling := filepath.Join(base, "dangling")
	if err := os.Symlink(filepath.Join(base, "missing"), dangling); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, tmp, cache string
		valid            bool
	}{
		{"outside", temp, filepath.Join(base, "new", "cache"), true},
		{"cache-off", temp, "off", true},
		{"temp-module", module, "off", false},
		{"temp-checkout", checkout, "off", false},
		{"temp-alias", alias, "off", false},
		{"cache-checkout", temp, filepath.Join(checkout, "new", "cache"), false},
		{"cache-alias", temp, filepath.Join(alias, "new", "cache"), false},
		{"cache-relative", temp, "relative", false},
		{"temp-missing", filepath.Join(base, "absent"), "off", false},
		{"cache-dangling", temp, dangling, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("TMPDIR", tc.tmp)
			t.Setenv("GOCACHE", tc.cache)
			got, err := writableHostPaths(module)
			if (err == nil) != tc.valid || (tc.valid && got != temp) {
				t.Fatalf("%q %v", got, err)
			}
		})
	}
}

// These host fakes test failure reporting only. Actual candidate behavior is
// separately compared against the unchanged independent literal reports.
func fakeHost(t *testing.T, expected []byte) (host, *int, *string) {
	t.Helper()
	calls := 0
	dir := ""
	h := localHost()
	h.goPath = func() (string, error) { return "/trusted/go", nil }
	h.temp = func(string) (string, error) { p, e := os.MkdirTemp(t.TempDir(), "private-"); dir = p; return p, e }
	h.run = func(ctx context.Context, program string, args []string, work string, in []byte, limit int) processResult {
		calls++
		if len(args) > 0 && args[0] == "env" {
			return processResult{started: true, output: []byte(compiler.ToolchainVersion + "\n")}
		}
		if len(args) > 0 && args[0] == "build" {
			if err := os.WriteFile(filepath.Join(work, "candidate-runner"), []byte("test binary identity"), 0700); err != nil {
				t.Fatal(err)
			}
			return processResult{started: true}
		}
		return processResult{started: true, output: append([]byte(nil), expected...)}
	}
	return h, &calls, &dir
}

func TestHostStageFailures(t *testing.T) {
	x := literals(t)[0]
	raw := requestBytes(t, fixture(t, x))
	root := moduleRoot(t)
	for _, stage := range []string{"sources", "project", "toolchain", "workspace", "write", "compilation", "binary", "start", "after-start", "report", "cleanup", "cleanup-no-removal"} {
		t.Run(stage, func(t *testing.T) {
			h, calls, dir := fakeHost(t, x.Report)
			original := h.run
			wantStage := stage
			wantCalls := 0
			started := false
			switch stage {
			case "sources":
				h.sources = func(string) ([]artifactbundle.File, error) { return nil, errors.New("private diagnostic") }
			case "project":
				h.sources = func(root string) ([]artifactbundle.File, error) {
					f, e := readSources(root)
					f[0].Data = append(f[0].Data, ' ')
					return f, e
				}
			case "toolchain":
				wantCalls = 1
				h.run = func(context.Context, string, []string, string, []byte, int) processResult {
					*calls++
					return processResult{started: true, output: []byte("go0.0.0\n")}
				}
			case "workspace":
				wantCalls = 1
				h.temp = func(string) (string, error) { return "", errors.New("private diagnostic") }
			case "write":
				wantStage = "workspace"
				wantCalls = 1
				makeDir := h.temp
				h.temp = func(string) (string, error) {
					d, e := makeDir("")
					if e == nil {
						e = os.WriteFile(filepath.Join(d, "go.mod"), []byte("occupied"), 0600)
					}
					return d, e
				}
			case "compilation":
				wantCalls = 2
				h.run = func(c context.Context, p string, a []string, d string, in []byte, n int) processResult {
					if len(a) > 0 && a[0] == "build" {
						*calls++
						return processResult{started: true, exit: 1}
					}
					return original(c, p, a, d, in, n)
				}
			case "binary":
				wantCalls = 2
				h.run = func(c context.Context, p string, a []string, d string, in []byte, n int) processResult {
					if len(a) > 0 && a[0] == "build" {
						*calls++
						return processResult{started: true}
					}
					return original(c, p, a, d, in, n)
				}
			case "start", "after-start":
				wantStage = "replay"
				wantCalls = 3
				started = stage == "after-start"
				h.run = func(c context.Context, p string, a []string, d string, in []byte, n int) processResult {
					if len(a) == 0 {
						*calls++
						return processResult{started: started, exit: -1, err: ErrExecution}
					}
					return original(c, p, a, d, in, n)
				}
			case "report":
				wantStage = "replay-report"
				wantCalls = 3
				started = true
				h.run = func(c context.Context, p string, a []string, d string, in []byte, n int) processResult {
					v := original(c, p, a, d, in, n)
					if len(a) == 0 {
						v.output = []byte(`{"matched":true}`)
					}
					return v
				}
			case "cleanup":
				wantCalls = 3
				started = true
				h.remove = func(d string) error { _ = os.RemoveAll(d); return errors.New("injected failure after removal") }
			case "cleanup-no-removal":
				wantStage = "cleanup"
				wantCalls = 3
				started = true
				h.remove = func(string) error { return nil }
				t.Cleanup(func() { _ = os.RemoveAll(*dir) })
			}
			r, err := execute(context.Background(), raw, root, h)
			if !errors.Is(err, ErrExecution) || ExitCode(r, err) != 4 || r.Outcome != "failed" || r.FailureStage == nil || *r.FailureStage != wantStage || *calls != wantCalls || r.CandidateStarted != started {
				t.Fatalf("stage %s calls%d receipt%+v error%v", stage, *calls, r, err)
			}
			if strings.HasPrefix(stage, "cleanup") {
				if r.Replay != "matched" || len(r.CandidateReport) == 0 || r.Cleanup != "failed" {
					t.Fatal("lost prior observed replay", r)
				}
			} else if *dir != "" && r.Cleanup != "removed" {
				t.Fatal("workspace not removed", r)
			}
			wire, encodeErr := Encode(r)
			if encodeErr != nil || bytes.Contains(wire, []byte("private diagnostic")) {
				t.Fatal("diagnostic leak", encodeErr)
			}
		})
	}
}

func TestCapturedExitIdentityAndPrefix(t *testing.T) {
	x := literals(t)[0]
	p, r, err := prepare(requestBytes(t, fixture(t, x)), "execute")
	if err != nil {
		t.Fatal(err)
	}
	for _, field := range []string{"modelDigest", "scenarioDigest", "sourceDigest", "formatVersion", "authority"} {
		var d map[string]any
		if json.Unmarshal(x.Report, &d) != nil {
			t.Fatal("oracle")
		}
		d[field] = "wrong"
		b, _ := json.Marshal(d)
		if _, err := captured(b, 0, p, r); err == nil {
			t.Fatal("accepted wrong", field)
		}
	}
	if _, err := captured(x.Report, 1, p, r); err == nil {
		t.Fatal("exit/report disagreement")
	}
	var d map[string]any
	_ = json.Unmarshal(x.Report, &d)
	entries := d["entries"].([]any)
	entries[0].(map[string]any)["stepId"] = "wrong"
	bad, _ := json.Marshal(d)
	if _, err := captured(bad, 0, p, r); err == nil {
		t.Fatal("wrong replay prefix")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	got, e := execute(ctx, requestBytes(t, fixture(t, x)), "", host{})
	if e == nil || got.CandidateStarted || got.FailureStage == nil || *got.FailureStage != "context" {
		t.Fatal("cancelled pre-invocation", got, e)
	}
}

func TestStaticSourceAndClosedWorkspace(t *testing.T) {
	files, err := readSources(moduleRoot(t))
	if err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	for _, f := range files {
		p := filepath.Join(root, f.Name)
		if err := os.MkdirAll(filepath.Dir(p), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, f.Data, 0600); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Remove(filepath.Join(root, "go.mod")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(moduleRoot(t), "go.mod"), filepath.Join(root, "go.mod")); err != nil {
		t.Fatal(err)
	}
	if _, err := readSources(root); err == nil {
		t.Fatal("source symlink accepted")
	}
	if err := writeProject(t.TempDir(), []artifactbundle.File{{Name: "../outside", Data: []byte("x")}}); err == nil {
		t.Fatal("open inventory accepted")
	}
	if _, err := hashBinary(filepath.Join(root, "go.mod")); err == nil {
		t.Fatal("symlink hashed as binary")
	}
}

func TestProcessEnvironment(t *testing.T) {
	for k, v := range map[string]string{"GOFLAGS": "-toolexec=untrusted", "GOTOOLCHAIN": "auto", "GOENV": "unexpected", "GOWORK": "unexpected", "GOPROXY": "https://invalid.invalid", "GOROOT": "unexpected", "GOOS": "unexpected", "LD_PRELOAD": "unexpected", "GODEBUG": "unexpected"} {
		t.Setenv(k, v)
	}
	e := strings.Join(processEnvironment(), "\n")
	for _, bad := range []string{"unexpected", "untrusted", "https://invalid.invalid", "GOTOOLCHAIN=auto"} {
		if strings.Contains(e, bad) {
			t.Fatal("ambient execution setting survived", bad)
		}
	}
	for _, want := range []string{"GOTOOLCHAIN=local", "GOPROXY=off", "GOSUMDB=off", "CGO_ENABLED=0", "GOENV=off", "GOWORK=off", "GOFLAGS="} {
		if !strings.Contains(e, want) {
			t.Fatal("missing pinned setting", want)
		}
	}
}

func TestProcessHelper(t *testing.T) {
	if len(os.Args) < 2 {
		return
	}
	switch os.Args[len(os.Args)-1] {
	case "proposal-helper-output":
		fmt.Print(strings.Repeat("x", 1024))
		os.Exit(0)
	case "proposal-helper-exit":
		os.Exit(1)
	case "proposal-helper-wait":
		time.Sleep(10 * time.Second)
		os.Exit(0)
	}
}
func TestRealProcessFailureBoundaries(t *testing.T) {
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	r := runProcess(context.Background(), exe, []string{"-test.run=^TestProcessHelper$", "--", "proposal-helper-exit"}, t.TempDir(), nil, 128)
	if !r.started || r.exit != 1 || r.err != nil {
		t.Fatal("ordinary nonzero exit", r)
	}
	r = runProcess(context.Background(), exe, []string{"-test.run=^TestProcessHelper$", "--", "proposal-helper-output"}, t.TempDir(), nil, 64)
	if !r.started || r.err == nil || len(r.output) > 64 {
		t.Fatal("bounded output", r)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	r = runProcess(ctx, exe, []string{"-test.run=^TestProcessHelper$", "--", "proposal-helper-wait"}, t.TempDir(), nil, 128)
	if !r.started || r.err == nil {
		t.Fatal("started timeout misreported", r)
	}
	r = runProcess(context.Background(), filepath.Join(t.TempDir(), "missing"), nil, "", nil, 128)
	if r.started || r.err == nil {
		t.Fatal("start failure misreported", r)
	}
}
