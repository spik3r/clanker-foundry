// tuic turns a reviewed JSON model into a record-browser package, following
// Foundry generator contract v1.
//
//	tuic validate MODEL
//	tuic generate MODEL OUT
//	tuic check MODEL OUT
package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	tui "github.com/spik3r/clanker-foundry/modeling/tui"
)

const (
	exitFailure = 1
	exitRefused = 2
	exitDrift   = 3
)

const usage = "usage: tuic validate MODEL | generate MODEL OUT | check MODEL OUT"

type refusal string

func (r refusal) Error() string { return string(r) }

func main() {
	// A closed stdout pipe must reach the contract I/O failure path.
	signal.Ignore(syscall.SIGPIPE)
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

// run writes stdout only after dispatch succeeds. A delivery failure may leave
// a prefix; the exit code reports whether the full result was delivered.
func run(args []string, stdout, stderr io.Writer) int {
	line, drift, err := dispatch(args)
	switch {
	case err != nil:
		fmt.Fprintf(stderr, "tuic: %v\n", err)
		var r refusal
		var me tui.ModelError
		if errors.As(err, &r) || errors.As(err, &me) ||
			errors.Is(err, tui.ErrOutExists) || errors.Is(err, tui.ErrNoParent) {
			return exitRefused
		}
		return exitFailure
	case len(drift) > 0:
		for _, d := range drift {
			fmt.Fprintln(stderr, d)
		}
		return exitDrift
	}
	n, err := io.WriteString(stdout, line)
	if err != nil || n != len(line) {
		fmt.Fprintln(stderr, "tuic: output write failed")
		return exitFailure
	}
	return 0
}

func dispatch(args []string) (line string, drift []string, err error) {
	switch {
	case len(args) == 2 && args[0] == "validate":
		res, err := load(args[1])
		return res.Identity, nil, err
	case len(args) == 3 && args[0] == "generate":
		res, err := load(args[1])
		if err != nil {
			return "", nil, err
		}
		if err := tui.Write(args[2], res.Files); err != nil {
			return "", nil, err
		}
		return res.Identity, nil, nil
	case len(args) == 3 && args[0] == "check":
		res, err := load(args[1])
		if err != nil {
			return "", nil, err
		}
		drift, err := tui.Drift(res.Files, args[2])
		if err != nil || len(drift) > 0 {
			return "", drift, err
		}
		return res.Identity, nil, nil
	}
	return "", nil, refusal(usage)
}

func load(path string) (tui.Result, error) {
	f, err := os.Open(path)
	if err != nil {
		return tui.Result{}, refusal("cannot open model: " + err.Error())
	}
	defer f.Close()
	raw, err := io.ReadAll(io.LimitReader(f, tui.MaxModelBytes+1))
	if err != nil {
		return tui.Result{}, refusal("cannot read model: " + err.Error())
	}
	return tui.Build(raw)
}
