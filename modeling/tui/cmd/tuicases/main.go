// tuicases checks expectation cases against a record browser generated from a
// model. It is a separate command from the generator: tuic never runs generated
// code, tuicases builds and runs it on purpose.
//
//	tuicases validate MODEL CASES
//	tuicases check MODEL CASES
//
// Exit codes: 0 all pass, 1 build/run/internal failure, 2 refused (usage, bad
// model, bad cases, stale digest, records that do not match the model), 3 at
// least one expectation failed. Stdout is empty on a nonzero exit.
package main

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"go/format"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	tui "github.com/spik3r/clanker-foundry/modeling/tui"
)

const (
	exitFailure = 1
	exitRefused = 2
	exitFailed  = 3

	buildTimeout = 5 * time.Minute
	runTimeout   = 2 * time.Minute
)

const usage = "usage: tuicases validate MODEL CASES | check MODEL CASES"

//go:embed runner.go.tmpl
var runnerTemplate string

type refusal string

func (r refusal) Error() string { return string(r) }

type loaded struct {
	model  tui.Model
	files  map[string][]byte
	digest string
	cases  tui.Cases
	raw    []byte
}

func main() { os.Exit(run(os.Args[1:], os.Stdout, os.Stderr)) }

type result struct {
	ID       string   `json:"id"`
	Failures []string `json:"failures"`
}

func run(args []string, stdout, stderr io.Writer) int {
	if len(args) != 3 || (args[0] != "validate" && args[0] != "check") {
		fmt.Fprintf(stderr, "tuicases: %s\n", usage)
		return exitRefused
	}
	l, err := load(args[1], args[2])
	if err != nil {
		fmt.Fprintf(stderr, "tuicases: %v\n", err)
		return exitRefused
	}
	if args[0] == "validate" {
		fmt.Fprintf(stdout, "ok %d cases for model %s\n", len(l.cases.Cases), l.digest)
		return 0
	}
	results, err := execute(l)
	if err != nil {
		fmt.Fprintf(stderr, "tuicases: %v\n", err)
		return exitFailure
	}
	failed := false
	for _, r := range results {
		if len(r.Failures) > 0 {
			failed = true
		}
	}
	if !failed {
		for _, r := range results {
			fmt.Fprintf(stdout, "pass %s\n", r.ID)
		}
		return 0
	}
	// Every case is reported, on stderr, in case order.
	for _, r := range results {
		if len(r.Failures) == 0 {
			fmt.Fprintf(stderr, "pass %s\n", r.ID)
		}
		for _, f := range r.Failures {
			fmt.Fprintf(stderr, "fail %s: %s\n", r.ID, f)
		}
	}
	return exitFailed
}

func readLimited(path string, limit int, what string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, refusal(fmt.Sprintf("cannot open %s: %v", what, err))
	}
	defer f.Close()
	raw, err := io.ReadAll(io.LimitReader(f, int64(limit)+1))
	if err != nil {
		return nil, refusal(fmt.Sprintf("cannot read %s: %v", what, err))
	}
	return raw, nil
}

func load(modelPath, casesPath string) (loaded, error) {
	var l loaded
	raw, err := readLimited(modelPath, tui.MaxModelBytes, "model")
	if err != nil {
		return l, err
	}
	res, err := tui.Build(raw)
	if err != nil {
		return l, fmt.Errorf("model: %w", err)
	}
	if l.model, err = tui.Decode(raw); err != nil {
		return l, fmt.Errorf("model: %w", err)
	}
	l.files, l.digest = res.Files, tui.CanonicalSHA256(res.Files)
	if l.raw, err = readLimited(casesPath, tui.MaxCasesBytes, "cases"); err != nil {
		return l, err
	}
	if l.cases, err = tui.DecodeCases(l.raw); err != nil {
		return l, fmt.Errorf("cases: %w", err)
	}
	if err := l.cases.CheckAgainst(l.model, l.digest); err != nil {
		return l, fmt.Errorf("cases: %w", err)
	}
	return l, nil
}

// layoutProject writes the generated browser and the fixed runner under dir and
// returns the package directory of the runner, relative to dir, and the import
// path of the generated package.
func layoutProject(dir string, l loaded) (runnerDir, importPath string, err error) {
	write := func(files map[string][]byte) error {
		for name, content := range files {
			p := filepath.Join(dir, filepath.FromSlash(name))
			if err := os.MkdirAll(filepath.Dir(p), 0o700); err != nil {
				return err
			}
			if err := os.WriteFile(p, content, 0o600); err != nil {
				return err
			}
		}
		return nil
	}
	if !l.model.PackageMode() {
		first, _, _ := strings.Cut(string(l.files["go.mod"]), "\n")
		importPath = strings.TrimSpace(strings.TrimPrefix(first, "module "))
		return "cmd/tuicasesrun", importPath, write(l.files)
	}
	// Package mode: a private host module with the dependency requirements of
	// the module-mode output of the same model.
	base := l.model
	base.Emit = nil
	module, err := tui.Compile(base)
	if err != nil {
		return "", "", err
	}
	const host = "example.com/host"
	first, rest, _ := strings.Cut(string(module["go.mod"]), "\n")
	if !strings.HasPrefix(first, "module ") {
		return "", "", errors.New("unexpected go.mod of module-mode output")
	}
	if err := write(map[string][]byte{"go.mod": []byte("module " + host + "\n" + rest), "go.sum": module["go.sum"]}); err != nil {
		return "", "", err
	}
	sub := map[string][]byte{}
	pkg := l.model.PackageName()
	for name, content := range l.files {
		sub["internal/"+pkg+"/"+name] = content
	}
	return ".", host + "/internal/" + pkg, write(sub)
}

func renderRunner(l loaded, importPath string) ([]byte, error) {
	tmpl, err := template.New("runner").Delims("[[[", "]]]").Option("missingkey=error").Parse(runnerTemplate)
	if err != nil {
		return nil, err
	}
	var out bytes.Buffer
	data := struct {
		Import string
		Fields []tui.Field
	}{importPath, l.model.Fields}
	if err := tmpl.Execute(&out, data); err != nil {
		return nil, err
	}
	return format.Source(out.Bytes())
}

func execute(l loaded) ([]result, error) {
	dir, err := os.MkdirTemp("", "tuicases-")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(dir)
	project := filepath.Join(dir, "project")
	if err := os.Mkdir(project, 0o700); err != nil {
		return nil, err
	}
	runnerDir, importPath, err := layoutProject(project, l)
	if err != nil {
		return nil, err
	}
	src, err := renderRunner(l, importPath)
	if err != nil {
		return nil, err
	}
	target := filepath.Join(project, filepath.FromSlash(runnerDir))
	if err := os.MkdirAll(target, 0o700); err != nil {
		return nil, err
	}
	if err := os.WriteFile(filepath.Join(target, "main.go"), src, 0o600); err != nil {
		return nil, err
	}
	casesFile := filepath.Join(dir, "cases.json")
	if err := os.WriteFile(casesFile, l.raw, 0o600); err != nil {
		return nil, err
	}

	binary := filepath.Join(dir, "runner")
	ctx, cancel := context.WithTimeout(context.Background(), buildTimeout)
	defer cancel()
	build := exec.CommandContext(ctx, "go", "build", "-mod=readonly", "-o", binary, "./"+filepath.ToSlash(runnerDir))
	build.Dir = project
	build.Env = append(os.Environ(), "GOWORK=off")
	if out, err := build.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("go build failed: %v\n%s", err, bytes.TrimSpace(out))
	}

	rctx, rcancel := context.WithTimeout(context.Background(), runTimeout)
	defer rcancel()
	cmd := exec.CommandContext(rctx, binary, casesFile)
	cmd.Dir = dir
	cmd.Env = []string{}
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("runner failed: %v\n%s", err, bytes.TrimSpace(stderr.Bytes()))
	}
	var results []result
	dec := json.NewDecoder(&stdout)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&results); err != nil {
		return nil, fmt.Errorf("runner output: %w", err)
	}
	if len(results) != len(l.cases.Cases) {
		return nil, fmt.Errorf("runner reported %d cases, want %d", len(results), len(l.cases.Cases))
	}
	for i, r := range results {
		if r.ID != l.cases.Cases[i].ID {
			return nil, fmt.Errorf("runner reported case %q, want %q", r.ID, l.cases.Cases[i].ID)
		}
	}
	return results, nil
}
