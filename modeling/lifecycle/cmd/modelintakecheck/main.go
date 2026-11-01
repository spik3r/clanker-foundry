// modelintakecheck connects a native intake request to offline model validation.
package main

import (
	"fmt"
	"io"
	"os"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/intakemodel"
)

func run(args []string, input io.Reader, output, diagnostic io.Writer) int {
	fail := func(code int, message string) int {
		_, _ = fmt.Fprintln(diagnostic, "modelintakecheck: "+message)
		return code
	}
	if len(args) != 0 {
		return fail(2, "usage: modelintakecheck < intake-request.json")
	}
	raw, err := io.ReadAll(io.LimitReader(input, intakemodel.MaxSourceBytes+1))
	if err != nil || len(raw) > intakemodel.MaxSourceBytes {
		return fail(2, "input refused")
	}
	report, err := intakemodel.Execute(raw)
	if err != nil {
		return fail(2, "request, toolchain or artifact refused")
	}
	n, err := output.Write(report)
	if err != nil || n != len(report) {
		return fail(3, "output write failed")
	}
	return 0
}

func main() { os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr)) }
