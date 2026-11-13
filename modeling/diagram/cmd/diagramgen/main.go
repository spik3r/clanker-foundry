// diagramgen turns a reviewed JSON diagram model into an SVG picture, BPMN XML
// and a canonical model, following Foundry generator contract v1.
//
//	diagramgen validate MODEL
//	diagramgen generate MODEL OUT
//	diagramgen check MODEL OUT
package main

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/spik3r/clanker-foundry/modeling/diagram"
	"github.com/spik3r/clanker-foundry/modeling/diagram/contract"
	"github.com/spik3r/clanker-foundry/modeling/diagram/dashboardbind"
	"github.com/spik3r/clanker-foundry/modeling/diagram/tesserae"
)

const (
	exitFailure = 1
	exitRefused = 2
	exitDrift   = 3
)

const usage = "usage: diagramgen validate MODEL | generate MODEL OUT | check MODEL OUT | bind TEMPLATE DATA | tesserae MODEL"

type usageError string

func (u usageError) Error() string { return string(u) }

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

// run returns the exit code. Stdout is written only on success.
func run(args []string, stdout, stderr io.Writer) int {
	line, drift, err := dispatch(args)
	switch {
	case err != nil:
		fmt.Fprintf(stderr, "diagramgen: %v\n", err)
		var u usageError
		var me contract.ModelError
		if errors.As(err, &u) || errors.As(err, &me) ||
			errors.Is(err, contract.ErrOutExists) || errors.Is(err, contract.ErrNoParent) {
			return exitRefused
		}
		return exitFailure
	case len(drift) > 0:
		for _, d := range drift {
			fmt.Fprintln(stderr, d)
		}
		return exitDrift
	}
	fmt.Fprint(stdout, line)
	return 0
}

func dispatch(args []string) (line string, drift []string, err error) {
	switch {
	case len(args) == 2 && args[0] == "tesserae":
		raw, err := contract.ReadModel(args[1])
		if err != nil {
			return "", nil, err
		}
		out, err := tesserae.Export(raw)
		if err != nil {
			return "", nil, usageError(err.Error())
		}
		return string(out), nil, nil
	case len(args) == 3 && args[0] == "bind":
		template, err := readBindingInput(args[1])
		if err != nil {
			return "", nil, err
		}
		data, err := readBindingInput(args[2])
		if err != nil {
			return "", nil, err
		}
		bound, err := dashboardbind.Bind(template, data)
		if err != nil {
			return "", nil, contract.Refuse("bind: %v", err)
		}
		return string(bound), nil, nil
	case len(args) == 2 && args[0] == "validate":
		res, err := load(args[1])
		return res.Identity, nil, err
	case len(args) == 3 && args[0] == "generate":
		res, err := load(args[1])
		if err != nil {
			return "", nil, err
		}
		if err := contract.Write(args[2], res.Files); err != nil {
			return "", nil, err
		}
		return res.Identity, nil, nil
	case len(args) == 3 && args[0] == "check":
		res, err := load(args[1])
		if err != nil {
			return "", nil, err
		}
		drift, err := contract.Drift(res.Files, args[2])
		if err != nil || len(drift) > 0 {
			return "", drift, err
		}
		return res.Identity, nil, nil
	}
	return "", nil, usageError(usage)
}

func load(path string) (contract.Result, error) {
	raw, err := contract.ReadModel(path)
	if err != nil {
		return contract.Result{}, err
	}
	return diagram.Build(raw)
}

// Binding reads only regular leaf files; do not open named pipes or devices.
// The project directory must not have a hostile concurrent writer.
func readBindingInput(path string) ([]byte, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return nil, contract.Refuse("cannot inspect binding input: %v", err)
	}
	if !info.Mode().IsRegular() {
		return nil, contract.Refuse("binding input must be a regular file")
	}
	return contract.ReadModel(path)
}
