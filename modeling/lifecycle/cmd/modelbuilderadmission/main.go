// modelbuilderadmission checks a fixed source frontier without building it.
package main

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/builderadmission"
)

func run(args []string, out, diag io.Writer) int {
	fail := func(code int, message string) int {
		_, _ = io.WriteString(diag, "modelbuilderadmission: "+message+"\n")
		return code
	}
	if len(args) != 5 || args[0] != "check" {
		return fail(2, "usage: check --root ABSOLUTE-EXERCISE --candidate-sha256 HEX")
	}
	flags := map[string]string{}
	for i := 1; i < len(args); i += 2 {
		if args[i] != "--root" && args[i] != "--candidate-sha256" {
			return fail(2, "invalid arguments")
		}
		if _, ok := flags[args[i]]; ok {
			return fail(2, "invalid arguments")
		}
		flags[args[i]] = args[i+1]
	}
	result, err := builderadmission.Check(flags["--root"], flags["--candidate-sha256"])
	if err != nil {
		if errors.Is(err, builderadmission.ErrInvalid) {
			return fail(2, "arguments or root rejected")
		}
		return fail(4, "admission incomplete")
	}
	var receipt struct{ Outcome string }
	if json.Unmarshal(result, &receipt) != nil {
		return fail(4, "admission incomplete")
	}
	n, err := out.Write(result)
	if err != nil || n != len(result) {
		return fail(3, "output failed")
	}
	if receipt.Outcome == "matched" {
		return 0
	}
	return 1
}
func main() { signal.Ignore(syscall.SIGPIPE); os.Exit(run(os.Args[1:], os.Stdout, os.Stderr)) }
