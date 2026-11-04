// modelbuilderdemo prepares the fixed Go handoff without dispatch or execution.
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/builderdemo"
)

func run(args []string, out, diag io.Writer) int {
	fail := func(message string, code int) int {
		_, _ = fmt.Fprintln(diag, "modelbuilderdemo: "+message)
		return code
	}
	if len(args) != 7 || args[0] != "prepare" {
		return fail("invalid arguments", 2)
	}
	values := map[string]string{}
	for i := 1; i < len(args); i += 2 {
		flag := args[i]
		if (flag != "--root" && flag != "--work-parent" && flag != "--go") || values[flag] != "" || args[i+1] == "" {
			return fail("invalid arguments", 2)
		}
		values[flag] = args[i+1]
	}
	receipt, err := builderdemo.Prepare(values["--root"], values["--work-parent"], values["--go"])
	if errors.Is(err, builderdemo.ErrInvalid) {
		return fail("invalid arguments or locations", 2)
	}
	recovery := ""
	if receipt.Workspace != nil {
		recovery = fmt.Sprintf("; retained workspace %q", *receipt.Workspace)
	}
	if receipt.FormatVersion == "" {
		return fail("preparation failed"+recovery, 4)
	}
	raw, encodeErr := json.MarshalIndent(receipt, "", "  ")
	if encodeErr != nil {
		return fail("receipt encoding failed"+recovery, 4)
	}
	raw = append(raw, '\n')
	if n, writeErr := out.Write(raw); writeErr != nil || n != len(raw) {
		return fail("output failed"+recovery, 3)
	}
	if errors.Is(err, builderdemo.ErrBlocked) {
		return fail(err.Error(), 1)
	}
	if err != nil {
		return fail(err.Error()+recovery, 4)
	}
	return 0
}

func main() {
	// Delivery failure must be exit 3, not a SIGPIPE success-looking interruption.
	signal.Ignore(syscall.SIGPIPE)
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}
