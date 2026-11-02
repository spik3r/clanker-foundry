// modeltrace validates the example requirements trace and writes or checks its
// report. It is offline and takes no input or output paths: it reads
// examples/lifecycle.json, examples/oracle.json and traceability/requirements.json
// and owns only traceability/view/trace.{json,md,svg}.
//
// Exit codes: 0 success; 1 check found missing or stale views; 2 invalid
// requirements, unreadable inputs, a view over the output limit or bad usage;
// 3 valid but incomplete coverage with validate -complete.
package main

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/traceability"
)

const (
	modelPath        = "examples/lifecycle.json"
	catalogPath      = "examples/oracle.json"
	requirementsPath = "traceability/requirements.json"
	viewDir          = "traceability/view"
	// readLimit bounds every file read, including views read back by check.
	// Rendering refuses a view over this size, so generate never writes a
	// view that check would reject.
	readLimit = 4 << 20
)

var outputs = []string{"trace.json", "trace.md", "trace.svg"}

func main() { os.Exit(run(os.Args[1:], ".", os.Stdout, os.Stderr)) }

func usage(stderr io.Writer) int {
	fmt.Fprintln(stderr, "usage: modeltrace validate [-complete] | generate | check")
	return 2
}

func regular(r *os.Root, path string) ([]byte, error) {
	s, err := r.Lstat(path)
	if err != nil {
		return nil, err
	}
	if !s.Mode().IsRegular() || s.Size() > readLimit {
		return nil, errors.New(path + ": not a bounded regular file")
	}
	b, err := r.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if len(b) > readLimit {
		return nil, errors.New(path + ": read limit")
	}
	return b, nil
}

func build(r *os.Root) (traceability.Report, error) {
	var in [3][]byte
	for i, p := range []string{modelPath, catalogPath, requirementsPath} {
		b, err := regular(r, p)
		if err != nil {
			return traceability.Report{}, err
		}
		in[i] = b
	}
	return traceability.Build(in[0], in[1], in[2])
}

func render(rep traceability.Report) (map[string][]byte, error) {
	j, err := rep.JSON()
	if err != nil {
		return nil, err
	}
	out := map[string][]byte{"trace.json": j, "trace.md": rep.Markdown(), "trace.svg": rep.SVG()}
	for _, name := range outputs {
		if len(out[name]) > readLimit {
			return nil, fmt.Errorf("%s/%s would be %d bytes, over the %d-byte output limit", viewDir, name, len(out[name]), readLimit)
		}
	}
	return out, nil
}

func run(args []string, dir string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		return usage(stderr)
	}
	complete := false
	switch {
	case args[0] == "validate" && len(args) == 2 && args[1] == "-complete":
		complete = true
	case len(args) != 1:
		return usage(stderr)
	case args[0] != "validate" && args[0] != "generate" && args[0] != "check":
		return usage(stderr)
	}
	root, err := os.OpenRoot(dir)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}
	defer root.Close()
	rep, err := build(root)
	if err != nil {
		var inv *traceability.InvalidError
		if errors.As(err, &inv) {
			fmt.Fprintf(stderr, "invalid requirements (%d issues):\n", len(inv.Issues))
			for _, i := range inv.Issues {
				fmt.Fprintln(stderr, "  "+i.String())
			}
		} else {
			fmt.Fprintln(stderr, err)
		}
		return 2
	}
	want, err := render(rep)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}
	switch args[0] {
	case "validate":
		s := rep.Summary
		fmt.Fprintf(stdout, "valid: %d requirements, %d linked, %d incomplete; %d untraced model elements; %d unlinked acceptance cases\n",
			s.Requirements, s.Linked, s.Incomplete, s.UntracedModelElements, s.UnlinkedAcceptanceCases)
		for _, q := range rep.Requirements {
			for _, g := range q.Gaps {
				fmt.Fprintf(stdout, "gap %s %s\n", q.ID, g)
			}
		}
		for _, e := range rep.ModelElements {
			if e.Untraced() {
				fmt.Fprintf(stdout, "untraced %s:%s\n", e.Kind, e.ID)
			}
		}
		for _, c := range rep.AcceptanceCases {
			if len(c.TracedBy) == 0 {
				fmt.Fprintf(stdout, "unlinked case %s\n", c.CaseID)
			}
		}
		if complete && (s.Incomplete > 0 || s.UntracedModelElements > 0 || s.UnlinkedAcceptanceCases > 0) {
			return 3
		}
		return 0
	case "check":
		stale := 0
		for _, name := range outputs {
			got, err := regular(root, viewDir+"/"+name)
			if err != nil || !bytes.Equal(got, want[name]) {
				fmt.Fprintf(stderr, "stale or missing: %s/%s\n", viewDir, name)
				stale++
			}
		}
		if stale > 0 {
			return 1
		}
		fmt.Fprintln(stdout, "views up to date")
		return 0
	default: // generate
		if err := write(root, want); err != nil {
			fmt.Fprintln(stderr, err)
			return 2
		}
		fmt.Fprintf(stdout, "wrote %s/{trace.json,trace.md,trace.svg}\n", viewDir)
		return 0
	}
}

// write refuses a symlinked or non-directory view directory and non-regular
// existing outputs. It then creates every temporary file exclusively under a
// random name, so a planted file or symlink is never opened or followed, and
// renames them over the outputs. On error it removes only the temporaries it
// created. It is not atomic across the three files.
func write(root *os.Root, files map[string][]byte) error {
	if s, err := root.Lstat(viewDir); err == nil {
		if !s.IsDir() {
			return errors.New(viewDir + ": not a directory")
		}
	} else if errors.Is(err, os.ErrNotExist) {
		if err := root.Mkdir(viewDir, 0o755); err != nil {
			return err
		}
	} else {
		return err
	}
	for _, name := range outputs {
		p := viewDir + "/" + name
		if s, err := root.Lstat(p); err == nil && !s.Mode().IsRegular() {
			return errors.New(p + ": not a regular file")
		} else if err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
	}
	created := map[string]string{}
	cleanup := func() {
		for _, tmp := range created {
			root.Remove(tmp)
		}
	}
	for _, name := range outputs {
		var r [8]byte
		if _, err := rand.Read(r[:]); err != nil {
			cleanup()
			return err
		}
		tmp := viewDir + "/." + name + "." + hex.EncodeToString(r[:]) + ".tmp"
		f, err := root.OpenFile(tmp, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
		if err != nil {
			cleanup()
			return err
		}
		created[name] = tmp
		_, werr := f.Write(files[name])
		cerr := f.Close()
		if werr != nil || cerr != nil {
			cleanup()
			return errors.Join(werr, cerr)
		}
	}
	for _, name := range outputs {
		if err := root.Rename(created[name], viewDir+"/"+name); err != nil {
			cleanup()
			return err
		}
		delete(created, name)
	}
	return nil
}
