package foundrygen

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const usage = "usage: foundrygen check|generate|plan --project DIR --foundry CHECKOUT [--unpinned]"

// Run executes one foundrygen command line (without the program name) and
// returns the exit code. Results are buffered until the command succeeds;
// a failed delivery can leave a partial stdout prefix with exit 1.
func Run(args []string, stdout, stderr io.Writer) int {
	var out bytes.Buffer
	s := &session{stderr: &bytes.Buffer{}}
	code := s.main(args, &out)
	if s.stderr.Len() > 0 {
		n, err := stderr.Write(s.stderr.Bytes())
		if (err != nil || n != s.stderr.Len()) && code == exitOK {
			_, _ = io.WriteString(stderr, "foundrygen: diagnostic write failed\n")
			return exitFailure
		}
	}
	if code == exitOK && out.Len() > 0 {
		n, err := stdout.Write(out.Bytes())
		if err != nil || n != out.Len() {
			_, _ = io.WriteString(stderr, "foundrygen: output write failed\n")
			return exitFailure
		}
	}
	return code
}

func (s *session) usageError(format string, a ...any) int {
	s.say("foundrygen", fmt.Sprintf(format, a...))
	s.say("", usage)
	return exitRefused
}

func (s *session) main(args []string, out *bytes.Buffer) int {
	if len(args) == 0 || (args[0] != "check" && args[0] != "generate" && args[0] != "plan") {
		return s.usageError("first argument must be check, generate or plan")
	}
	verb := args[0]
	fs := flag.NewFlagSet(verb, flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	project := fs.String("project", ".", "")
	foundry := fs.String("foundry", "", "")
	unpinned := fs.Bool("unpinned", false, "")
	if err := fs.Parse(args[1:]); err != nil {
		return s.usageError("%v", err)
	}
	if fs.NArg() > 0 {
		return s.usageError("unexpected argument %q", fs.Arg(0))
	}
	if *foundry == "" {
		return s.usageError("--foundry is required")
	}
	var err error
	if s.root, err = absDir(*project); err != nil {
		return s.usageError("--project: %v", err)
	}
	if s.foundry, err = absDir(*foundry); err != nil {
		return s.usageError("--foundry: %v", err)
	}

	if s.reg, err = LoadRegistry(s.foundry); err != nil {
		return s.report(err)
	}
	if s.proj, err = LoadProject(s.root, s.reg); err != nil {
		return s.report(err)
	}
	if err := s.checkPathAliases(); err != nil {
		return s.report(err)
	}
	if *unpinned {
		s.say("foundrygen", "warning: --unpinned: Foundry revision and modeling/ state not verified")
	} else if err := s.pin(); err != nil {
		return s.report(err)
	}

	if s.binDir, err = os.MkdirTemp("", "foundrygen-bin-"); err != nil {
		return s.report(fail("temp dir: %v", err))
	}
	defer os.RemoveAll(s.binDir)
	if err := s.build(); err != nil {
		return s.report(err)
	}

	if verb == "check" {
		return s.cmdCheck(out)
	}
	if verb == "plan" {
		return s.cmdPlan(out)
	}
	return s.cmdGenerate(out)
}

func (s *session) report(err error) int {
	s.say("foundrygen", err.Error())
	return codeOf(err)
}

func absDir(p string) (string, error) {
	abs, err := filepath.Abs(p)
	if err != nil {
		return "", err
	}
	fi, err := os.Stat(abs)
	if err != nil {
		return "", err
	}
	if !fi.IsDir() {
		return "", fmt.Errorf("%s is not a directory", abs)
	}
	return abs, nil
}

// worst folds outcome codes: refused beats failure beats drift.
func worst(codes []int) int {
	has := map[int]bool{}
	for _, c := range codes {
		has[c] = true
	}
	for _, c := range []int{exitRefused, exitFailure, exitDrift} {
		if has[c] {
			return c
		}
	}
	return exitOK
}

func (s *session) cmdCheck(out *bytes.Buffer) int {
	var codes []int
	var lines bytes.Buffer
	for _, o := range s.proj.Outputs {
		model, outPath, err := s.paths(o)
		if err != nil {
			s.say(o.Out, err.Error())
			codes = append(codes, codeOf(err))
			continue
		}
		oc := s.run(o, "check", model, outPath)
		s.say(o.Out, oc.lines)
		if oc.err != nil {
			s.say(o.Out, oc.err.Error())
		}
		codes = append(codes, oc.code)
		if oc.code == exitOK {
			fmt.Fprintf(&lines, "ok %s %s %s\n", o.Out, o.Generator, oc.id.CanonicalSHA256)
		}
	}
	code := worst(codes)
	if code == exitOK {
		out.Write(lines.Bytes())
	}
	return code
}

type step struct {
	o            Output
	model, out   string
	exists       bool
	unchanged    bool
	id           Identity
	status       string
	pendingWrite bool
}

// prepareGeneration validates all inputs and classifies outputs without writing them.
// Both plan and generate use this preflight so their refusal rules stay aligned.
func (s *session) prepareGeneration() ([]*step, int) {
	steps := make([]*step, len(s.proj.Outputs))
	var codes []int

	// Phase 1: resolve paths and validate every model before anything is written.
	for i, o := range s.proj.Outputs {
		st := &step{o: o}
		steps[i] = st
		var err error
		if st.model, st.out, err = s.paths(o); err != nil {
			s.say(o.Out, err.Error())
			codes = append(codes, codeOf(err))
			continue
		}
		oc := s.run(o, "validate", st.model, "")
		s.say(o.Out, oc.lines)
		if oc.err != nil {
			s.say(o.Out, oc.err.Error())
			codes = append(codes, oc.code)
		}
		st.id = oc.id
	}
	if c := worst(codes); c != exitOK {
		return nil, c
	}

	// Phase 2: decide what each output needs, still without writing.
	for _, st := range steps {
		fi, err := os.Lstat(st.out)
		switch {
		case os.IsNotExist(err):
			st.pendingWrite = true
			st.status = "generated"
			continue
		case err != nil:
			s.say(st.o.Out, fail("%v", err).Error())
			codes = append(codes, exitFailure)
			continue
		}
		st.exists = true
		if err := ownedDir(st.out, fi); err != nil {
			s.say(st.o.Out, err.Error())
			codes = append(codes, exitRefused)
			continue
		}
		oc := s.run(st.o, "check", st.model, st.out)
		switch oc.code {
		case exitOK:
			st.unchanged, st.id, st.status = true, oc.id, "unchanged"
		case exitDrift:
			st.pendingWrite, st.status = true, "regenerated"
		default:
			s.say(st.o.Out, oc.lines)
			s.say(st.o.Out, oc.err.Error())
			codes = append(codes, oc.code)
		}
	}
	if c := worst(codes); c != exitOK {
		return nil, c
	}
	return steps, exitOK
}

// cmdPlan emits an observation, not an approval or an executable saved plan.
func (s *session) cmdPlan(out *bytes.Buffer) int {
	steps, code := s.prepareGeneration()
	if code != exitOK {
		return code
	}
	type plannedOutput struct {
		Action    string   `json:"action"`
		Generator string   `json:"generator"`
		Model     string   `json:"model"`
		Out       string   `json:"out"`
		Identity  Identity `json:"identity"`
	}
	report := struct {
		PlanVersion string          `json:"planVersion"`
		Outputs     []plannedOutput `json:"outputs"`
	}{PlanVersion: "foundry-plan/v1", Outputs: make([]plannedOutput, 0, len(steps))}
	for _, st := range steps {
		action := "unchanged"
		if st.pendingWrite {
			action = "create"
			if st.exists {
				action = "replace"
			}
		}
		report.Outputs = append(report.Outputs, plannedOutput{action, st.o.Generator, st.o.Model, st.o.Out, st.id})
	}
	enc := json.NewEncoder(out)
	enc.SetIndent("", "  ")
	if err := enc.Encode(report); err != nil {
		return s.report(fail("encode plan: %v", err))
	}
	return exitOK
}

func (s *session) cmdGenerate(out *bytes.Buffer) int {
	steps, code := s.prepareGeneration()
	if code != exitOK {
		return code
	}
	// Phase 3: write.
	for i, st := range steps {
		if !st.pendingWrite {
			continue
		}
		id, err := s.write(st)
		if err != nil {
			s.say(st.o.Out, err.Error())
			for _, done := range steps[:i] {
				if done.pendingWrite {
					s.say(done.o.Out, "already "+done.status+" before the failure")
				}
			}
			return codeOf(err)
		}
		st.id = id
	}
	for _, st := range steps {
		fmt.Fprintf(out, "%s %s %s %s\n", st.status, st.o.Out, st.o.Generator, st.id.CanonicalSHA256)
	}
	return exitOK
}

// ownedDir refuses an existing OUT unless it is a real directory holding a
// regular foundry-manifest.json.
func ownedDir(path string, fi os.FileInfo) error {
	if !fi.IsDir() {
		return refuse("%s exists and is not a directory; refusing to replace it", path)
	}
	mf, err := os.Lstat(filepath.Join(path, "foundry-manifest.json"))
	if err != nil || !mf.Mode().IsRegular() {
		return refuse("%s exists but holds no foundry-manifest.json; it was not generated, refusing to replace it", path)
	}
	return nil
}

// write generates into a new sibling directory, verifies it, then swaps it in.
func (s *session) write(st *step) (Identity, error) {
	// Earlier writes may have given formerly absent names a shared identity.
	if err := s.checkPathAliases(); err != nil {
		return Identity{}, err
	}
	parent := filepath.Dir(st.out)
	if err := os.MkdirAll(parent, 0o755); err != nil {
		return Identity{}, fail("create %s: %v", parent, err)
	}
	holder, err := os.MkdirTemp(parent, ".foundrygen-")
	if err != nil {
		return Identity{}, fail("temp dir in %s: %v", parent, err)
	}
	cleanup := func(err error) error {
		if rerr := os.RemoveAll(holder); rerr != nil {
			return fail("%v; leftover directory %s could not be removed: %v", err, holder, rerr)
		}
		return err
	}
	fresh := filepath.Join(holder, "out")
	old := filepath.Join(holder, "old")

	gen := s.run(st.o, "generate", st.model, fresh)
	s.say(st.o.Out, gen.lines)
	if gen.err != nil {
		return Identity{}, cleanup(gen.err)
	}
	chk := s.run(st.o, "check", st.model, fresh)
	if chk.code != exitOK || chk.id != gen.id {
		s.say(st.o.Out, chk.lines)
		return Identity{}, cleanup(fail("%s: freshly generated output did not pass check", st.o.Generator))
	}

	if st.exists {
		if err := os.Rename(st.out, old); err != nil {
			return Identity{}, cleanup(fail("move aside %s: %v", st.out, err))
		}
		if err := os.Rename(fresh, st.out); err != nil {
			if rerr := os.Rename(old, st.out); rerr != nil {
				return Identity{}, fail("install %s: %v; restoring the old output failed: %v; the old output is at %s", st.out, err, rerr, old)
			}
			return Identity{}, cleanup(fail("install %s: %v; the old output was restored", st.out, err))
		}
	} else {
		if _, err := os.Lstat(st.out); err == nil {
			return Identity{}, cleanup(refuse("%s appeared while generating; leaving it alone", st.out))
		}
		if err := os.Rename(fresh, st.out); err != nil {
			return Identity{}, cleanup(fail("install %s: %v", st.out, err))
		}
	}
	if err := os.RemoveAll(holder); err != nil {
		return gen.id, fail("%s was written but the temporary directory %s could not be removed: %v", st.out, holder, err)
	}
	return gen.id, nil
}
