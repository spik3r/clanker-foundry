package proposalreplay

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/artifactbundle"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/compiler"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/qualification"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/scenarios"
)

const processTimeout = 180 * time.Second
const maxBinaryBytes = 64 << 20

type processResult struct {
	started bool
	output  []byte
	exit    int
	err     error
}

// Private seams permit boundary tests, never request-selected execution.
type host struct {
	sources func(string) ([]artifactbundle.File, error)
	goPath  func() (string, error)
	temp    func(string) (string, error)
	remove  func(string) error
	run     func(context.Context, string, []string, string, []byte, int) processResult
}

func localHost() host {
	return host{sources: readSources, goPath: func() (string, error) {
		p, err := exec.LookPath("go")
		if err != nil {
			return "", err
		}
		return filepath.Abs(p)
	}, temp: func(parent string) (string, error) { return os.MkdirTemp(parent, "foundry-proposal-") }, remove: os.RemoveAll, run: runProcess}
}

func execute(ctx context.Context, raw []byte, root string, h host) (r Receipt, err error) {
	p, r, err := prepare(raw, "execute")
	if err != nil || r.Outcome == "stale" {
		return r, err
	}
	fail := func(stage string) (Receipt, error) {
		r.Outcome, r.FailureStage = "failed", ptr(stage)
		return r, ErrExecution
	}
	if ctx == nil || ctx.Err() != nil {
		return fail("context")
	}
	sources, err := h.sources(root)
	if err != nil {
		return fail("sources")
	}
	archive, err := qualification.Build([]byte(p.request.Model), []byte(p.request.Suite), sources)
	if err != nil || qualification.Verify(archive, r.Sources.ModelSHA256, r.Sources.SuiteSHA256) != nil {
		return fail("project")
	}
	files, err := artifactbundle.Verify(archive, artifactbundle.QualificationProfile)
	if err != nil || !closedInventory(files) {
		return fail("project")
	}
	r.SourceRevision, r.ArchiveSHA256 = ptr(qualification.SourceRevision), ptr(hash(archive))
	parent, err := writableHostPaths(root)
	if err != nil {
		return fail("host-paths")
	}
	goPath, err := h.goPath()
	if err != nil {
		return fail("toolchain")
	}
	version := h.run(ctx, goPath, []string{"env", "GOVERSION"}, root, nil, 128)
	if version.err != nil || !version.started || version.exit != 0 || strings.TrimSpace(string(version.output)) != compiler.ToolchainVersion {
		return fail("toolchain")
	}
	if ctx.Err() != nil {
		return fail("context")
	}
	dir, err := h.temp(parent)
	if err != nil {
		return fail("workspace")
	}
	r.Cleanup = "pending"
	defer func() {
		removeErr := h.remove(dir)
		_, statErr := os.Lstat(dir)
		if removeErr != nil || !errors.Is(statErr, os.ErrNotExist) {
			r.Cleanup, r.Outcome, r.FailureStage = "failed", "failed", ptr("cleanup")
			err = ErrExecution
		} else {
			r.Cleanup = "removed"
		}
	}()
	if err := writeProject(dir, files); err != nil {
		return fail("workspace")
	}
	binary := filepath.Join(dir, "candidate-runner")
	build := h.run(ctx, goPath, []string{"build", "-p=1", "-trimpath", "-buildvcs=false", "-o", binary, "./cmd/modelplay"}, dir, nil, 4096)
	if build.err != nil || !build.started || build.exit != 0 {
		r.Compilation = "failed"
		return fail("compilation")
	}
	r.Compilation = "passed"
	binaryHash, err := hashBinary(binary)
	if err != nil {
		return fail("binary")
	}
	r.BinarySHA256 = ptr(binaryHash)
	invoked := h.run(ctx, binary, nil, dir, []byte(p.request.Suite), scenarios.MaxReportBytes)
	r.CandidateStarted = invoked.started
	if invoked.started {
		r.CandidateExitCode = ptr(invoked.exit)
	}
	if invoked.err != nil || !invoked.started {
		r.Replay = "failed-to-start"
		if invoked.started {
			r.Replay = "failed-after-start"
		}
		return fail("replay")
	}
	matched, err := captured(invoked.output, invoked.exit, p, r)
	if err != nil {
		r.Replay = "failed-after-start"
		return fail("replay-report")
	}
	r.CandidateReport = append([]byte(nil), invoked.output...)
	r.Replay, r.Outcome = "mismatched", "mismatched"
	if matched {
		r.Replay, r.Outcome = "matched", "matched"
	}
	return r, nil
}

func readSources(root string) ([]artifactbundle.File, error) {
	if root == "" {
		return nil, ErrExecution
	}
	st, err := os.Lstat(root)
	if err != nil || !st.IsDir() || st.Mode()&os.ModeSymlink != 0 {
		return nil, ErrExecution
	}
	var files []artifactbundle.File
	for _, name := range qualification.SourceNames() {
		path := root
		for _, part := range strings.Split(name, "/") {
			path = filepath.Join(path, part)
			st, err = os.Lstat(path)
			if err != nil || st.Mode()&os.ModeSymlink != 0 {
				return nil, ErrExecution
			}
		}
		if !st.Mode().IsRegular() || st.Size() > 1<<20 {
			return nil, ErrExecution
		}
		f, err := os.Open(path)
		if err != nil {
			return nil, ErrExecution
		}
		b, readErr := io.ReadAll(io.LimitReader(f, (1<<20)+1))
		closeErr := f.Close()
		if readErr != nil || closeErr != nil || len(b) > 1<<20 {
			return nil, ErrExecution
		}
		files = append(files, artifactbundle.File{Name: name, Data: b})
	}
	return files, nil
}

func closedInventory(files []artifactbundle.File) bool {
	names := append(qualification.SourceNames(), "generated/lifecycle.go", "generated/model_paths_test.go", "candidate.json", "suite.json", "qualification.json", "GUIDE.txt")
	if len(files) != len(names) {
		return false
	}
	remaining := map[string]bool{}
	for _, name := range names {
		remaining[name] = true
	}
	for _, f := range files {
		if !remaining[f.Name] {
			return false
		}
		delete(remaining, f.Name)
	}
	return len(remaining) == 0
}

func writeProject(dir string, files []artifactbundle.File) error {
	if !closedInventory(files) {
		return ErrExecution
	}
	for _, f := range files {
		path := filepath.Join(dir, filepath.FromSlash(f.Name))
		if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			return err
		}
		out, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
		if err != nil {
			return err
		}
		n, writeErr := out.Write(f.Data)
		closeErr := out.Close()
		if writeErr != nil || closeErr != nil || n != len(f.Data) {
			return ErrExecution
		}
	}
	return nil
}

func hashBinary(path string) (string, error) {
	st, err := os.Lstat(path)
	if err != nil || !st.Mode().IsRegular() || st.Size() <= 0 || st.Size() > maxBinaryBytes {
		return "", ErrExecution
	}
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	h := sha256.New()
	n, readErr := io.Copy(h, io.LimitReader(f, maxBinaryBytes+1))
	closeErr := f.Close()
	if readErr != nil || closeErr != nil || n != st.Size() || n > maxBinaryBytes {
		return "", ErrExecution
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func cachePath() (string, error) {
	if p := os.Getenv("GOCACHE"); p != "" {
		return p, nil
	}
	p, err := os.UserCacheDir()
	return filepath.Join(p, "go-build"), err
}

// Resolve an existing ancestor before accepting a not-yet-created cache path.
// Dangling links and other filesystem errors refuse rather than looking lexical.
func resolvedPath(path string) (string, error) {
	p, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	var missing []string
	for {
		_, err = os.Lstat(p)
		if err == nil {
			break
		}
		if !errors.Is(err, os.ErrNotExist) || filepath.Dir(p) == p {
			return "", ErrExecution
		}
		missing = append(missing, filepath.Base(p))
		p = filepath.Dir(p)
	}
	p, err = filepath.EvalSymlinks(p)
	if err != nil {
		return "", err
	}
	for i := len(missing) - 1; i >= 0; i-- {
		p = filepath.Join(p, missing[i])
	}
	return p, nil
}

func writableHostPaths(root string) (string, error) {
	root, err := filepath.EvalSymlinks(root)
	if err != nil {
		return "", err
	}
	root, err = filepath.Abs(root)
	if err != nil {
		return "", err
	}
	checkout := root
	for p := root; ; p = filepath.Dir(p) {
		if _, e := os.Lstat(filepath.Join(p, ".git")); e == nil {
			checkout = p
			break
		} else if !errors.Is(e, os.ErrNotExist) {
			return "", e
		}
		if filepath.Dir(p) == p {
			break
		}
	}
	parent, err := resolvedPath(os.TempDir())
	if err != nil {
		return "", err
	}
	st, err := os.Stat(parent)
	if err != nil || !st.IsDir() {
		return "", ErrExecution
	}
	cache, err := cachePath()
	if err != nil {
		return "", err
	}
	paths := []string{parent}
	if cache != "off" {
		if !filepath.IsAbs(cache) {
			return "", ErrExecution
		}
		cache, err = resolvedPath(cache)
		if err != nil {
			return "", err
		}
		paths = append(paths, cache)
	}
	for _, path := range paths {
		rel, e := filepath.Rel(checkout, path)
		if e != nil || (rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))) {
			return "", ErrExecution
		}
	}
	return parent, nil
}

func processEnvironment() []string {
	var env []string
	for _, key := range []string{"PATH", "HOME", "LANG", "LC_ALL"} {
		if value, ok := os.LookupEnv(key); ok {
			env = append(env, key+"="+value)
		}
	}
	cache, err := cachePath()
	if err != nil {
		cache = "off"
	}
	parent, err := filepath.Abs(os.TempDir())
	if err != nil {
		parent = os.TempDir()
	}
	env = append(env, "GOCACHE="+cache, "TMPDIR="+parent)
	return append(env, "GOTOOLCHAIN=local", "GOPROXY=off", "GOSUMDB=off", "GOWORK=off", "GOENV=off", "GOFLAGS=", "CGO_ENABLED=0", "GOMAXPROCS=2")
}

type boundedOutput struct {
	buffer   bytes.Buffer // Do not promote Buffer.ReadFrom: io.Copy could bypass Write.
	limit    int
	overflow bool
	cancel   context.CancelFunc
}

func (b *boundedOutput) Write(p []byte) (int, error) {
	if len(p) > b.limit-b.buffer.Len() {
		b.overflow = true
		b.cancel()
		return 0, ErrExecution
	}
	return b.buffer.Write(p)
}

func runProcess(ctx context.Context, program string, args []string, dir string, input []byte, limit int) processResult {
	ctx, cancel := context.WithTimeout(ctx, processTimeout)
	defer cancel()
	result := processResult{exit: -1}
	output := &boundedOutput{limit: limit, cancel: cancel}
	cmd := exec.CommandContext(ctx, program, args...)
	cmd.Dir, cmd.Env = dir, processEnvironment()
	cmd.Stdin, cmd.Stdout, cmd.Stderr = bytes.NewReader(input), output, io.Discard
	cmd.WaitDelay = 2 * time.Second
	if err := cmd.Start(); err != nil {
		result.err = err
		return result
	}
	result.started = true
	err := cmd.Wait()
	if cmd.ProcessState != nil {
		result.exit = cmd.ProcessState.ExitCode()
	}
	result.output = append([]byte(nil), output.buffer.Bytes()...)
	var exitErr *exec.ExitError
	if ctx.Err() != nil || output.overflow || result.exit < 0 || (err != nil && !errors.As(err, &exitErr)) {
		result.err = ErrExecution
	}
	return result
}
