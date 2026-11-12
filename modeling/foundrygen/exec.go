package foundrygen

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const (
	childTimeout = 60 * time.Second
	buildTimeout = 10 * time.Minute
	gitTimeout   = 60 * time.Second
	maxCapture   = 1 << 20
	contractID   = "foundry-generator/v1"
)

var hex64RE = regexp.MustCompile(`^[0-9a-f]{64}$`)

// capped keeps at most maxCapture bytes and discards the rest.
type capped struct{ buf bytes.Buffer }

func (c *capped) Write(p []byte) (int, error) {
	if room := maxCapture - c.buf.Len(); room > 0 {
		c.buf.Write(p[:min(room, len(p))])
	}
	return len(p), nil
}

type childResult struct {
	code   int
	stdout string
	stderr string
}

// runCmd runs a command to completion. A nonzero exit is reported in code;
// failing to start or exceeding the timeout is an error.
func runCmd(timeout time.Duration, dir string, env []string, name string, args ...string) (childResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = dir
	cmd.Env = env
	cmd.WaitDelay = 5 * time.Second
	var so, se capped
	cmd.Stdout, cmd.Stderr = &so, &se
	err := cmd.Run()
	res := childResult{stdout: so.buf.String(), stderr: se.buf.String()}
	if err == nil {
		return res, nil
	}
	if ctx.Err() != nil {
		return res, fmt.Errorf("%s timed out after %s", filepath.Base(name), timeout)
	}
	var ee *exec.ExitError
	if errors.As(err, &ee) {
		res.code = ee.ExitCode()
		if res.code < 0 {
			return res, fmt.Errorf("%s was killed: %v", filepath.Base(name), err)
		}
		return res, nil
	}
	return res, err
}

// Identity is the generator identity line.
type Identity struct {
	Contract        string `json:"contract"`
	Profile         string `json:"profile"`
	SourceSHA256    string `json:"sourceSHA256"`
	CanonicalSHA256 string `json:"canonicalSHA256"`
}

func parseIdentity(stdout string) (Identity, error) {
	var id Identity
	if !strings.HasSuffix(stdout, "\n") || strings.Count(stdout, "\n") != 1 {
		return id, errors.New("stdout is not exactly one line")
	}
	dec := json.NewDecoder(strings.NewReader(stdout))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&id); err != nil {
		return id, fmt.Errorf("identity line: %v", err)
	}
	if id.Contract != contractID {
		return id, fmt.Errorf("identity line: contract %q, want %q", id.Contract, contractID)
	}
	if !hex64RE.MatchString(id.SourceSHA256) || !hex64RE.MatchString(id.CanonicalSHA256) {
		return id, errors.New("identity line: bad digest")
	}
	return id, nil
}

// session holds the validated inputs of one command.
type session struct {
	root    string // absolute project root
	foundry string // absolute Foundry checkout
	proj    *Project
	reg     Registry
	binDir  string
	stderr  *bytes.Buffer
}

// say writes the lines of text to stderr, each prefixed with prefix.
func (s *session) say(prefix, text string) {
	for _, line := range strings.Split(strings.TrimRight(text, "\n"), "\n") {
		if line == "" {
			continue
		}
		if prefix != "" {
			line = prefix + ": " + line
		}
		s.stderr.WriteString(line + "\n")
	}
}

func (s *session) git(args ...string) (string, error) {
	res, err := runCmd(gitTimeout, "", os.Environ(), "git", append([]string{"-C", s.foundry}, args...)...)
	if err != nil {
		return "", fail("git %s: %v", strings.Join(args, " "), err)
	}
	if res.code != 0 {
		return "", refuse("git %s in %s failed: %s", strings.Join(args, " "), s.foundry, strings.TrimSpace(res.stderr))
	}
	return res.stdout, nil
}

// pin requires the Foundry checkout to be at the project's revision with a clean modeling tree.
func (s *session) pin() error {
	prefix, err := s.git("rev-parse", "--show-prefix")
	if err != nil {
		return err
	}
	if strings.TrimSpace(prefix) != "" {
		return refuse("%s is not the top of a git checkout", s.foundry)
	}
	head, err := s.git("rev-parse", "HEAD")
	if err != nil {
		return err
	}
	if got := strings.TrimSpace(head); got != s.proj.Revision {
		return refuse("Foundry checkout is at %s but %s pins %s", got, ProjectFile, s.proj.Revision)
	}
	status, err := s.git("status", "--porcelain", "--", "modeling")
	if err != nil {
		return err
	}
	if strings.TrimSpace(status) != "" {
		first := strings.SplitN(strings.TrimSpace(status), "\n", 2)[0]
		return refuse("Foundry checkout has uncommitted changes under modeling/ (%s)", first)
	}
	return nil
}

// build compiles every generator the project needs into the private bin dir.
func (s *session) build() error {
	env := append(os.Environ(), "GOFLAGS=-mod=readonly")
	built := map[string]bool{}
	for _, o := range s.proj.Outputs {
		if built[o.Generator] {
			continue
		}
		built[o.Generator] = true
		g := s.reg[o.Generator]
		dir := filepath.Join(s.foundry, filepath.FromSlash(g.Module))
		if fi, err := os.Stat(dir); err != nil || !fi.IsDir() {
			return fail("generator %s: module directory %s not found", g.ID, dir)
		}
		res, err := runCmd(buildTimeout, dir, env, "go", "build", "-o", filepath.Join(s.binDir, g.ID), g.Package)
		if err != nil {
			return fail("build %s: %v", g.ID, err)
		}
		if res.code != 0 {
			s.say("build "+g.ID, res.stderr)
			return fail("build %s failed", g.ID)
		}
	}
	return nil
}

// invoke runs a built generator with an empty environment.
func (s *session) invoke(o Output, verb string, args ...string) (childResult, error) {
	bin := filepath.Join(s.binDir, o.Generator)
	return runCmd(childTimeout, s.root, []string{}, bin, append([]string{verb}, args...)...)
}

func (s *session) paths(o Output) (model, out string, err error) {
	if model, err = resolveInProject(s.root, o.Model); err != nil {
		return
	}
	out, err = resolveInProject(s.root, o.Out)
	return
}

// outcome of one generator command on one output.
type outcome struct {
	code  int // exit code of foundrygen's view: 0, 1, 2 or 3 (drift)
	id    Identity
	lines string // child stderr, for reporting
	err   error  // set when code is 1 or 2
}

// run runs validate, check or generate and classifies the result.
func (s *session) run(o Output, verb, model, out string) outcome {
	args := []string{model}
	if verb != "validate" {
		args = append(args, out)
	}
	res, err := s.invoke(o, verb, args...)
	if err != nil {
		return outcome{code: exitFailure, err: fail("%s %s: %v", o.Generator, verb, err)}
	}
	switch {
	case res.code == 0:
		id, perr := parseIdentity(res.stdout)
		if perr != nil {
			return outcome{code: exitFailure, lines: res.stderr, err: fail("%s %s: %v", o.Generator, verb, perr)}
		}
		return outcome{id: id, lines: res.stderr}
	case res.code == 2:
		return outcome{code: exitRefused, lines: res.stderr, err: refuse("%s %s refused the model or output", o.Generator, verb)}
	case res.code == 3 && verb == "check":
		return outcome{code: exitDrift, lines: res.stderr}
	}
	return outcome{code: exitFailure, lines: res.stderr, err: fail("%s %s exited %d", o.Generator, verb, res.code)}
}
