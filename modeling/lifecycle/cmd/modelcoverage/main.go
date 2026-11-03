// modelcoverage inventories supplied assertions; it does not prove execution.
package main

import (
	"fmt"
	"io"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/compiler"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/scenariocoverage"
)

func run(args []string, in io.Reader, out, diag io.Writer) int {
	fail := func() int { _, _ = fmt.Fprintln(diag, "modelcoverage: invalid input"); return 2 }
	if len(args) != 0 || runtime.Version() != compiler.ToolchainVersion {
		return fail()
	}
	b, err := io.ReadAll(io.LimitReader(in, scenariocoverage.MaxRequestBytes+1))
	if err != nil || len(b) > scenariocoverage.MaxRequestBytes {
		return fail()
	}
	result, err := scenariocoverage.Execute(b)
	if err != nil {
		return fail()
	}
	n, err := out.Write(result)
	if err != nil || n != len(result) {
		_, _ = fmt.Fprintln(diag, "modelcoverage: output failed")
		return 3
	}
	return 0
}
func main() {
	signal.Ignore(syscall.SIGPIPE)
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}
