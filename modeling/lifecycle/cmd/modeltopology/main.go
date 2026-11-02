// modeltopology describes declared state-graph structure, not liveness.
package main

import (
	"fmt"
	"io"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/compiler"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/topology"
)

func run(args []string, in io.Reader, out, diag io.Writer) int {
	fail := func() int { _, _ = fmt.Fprintln(diag, "modeltopology: invalid input"); return 2 }
	if len(args) != 0 || runtime.Version() != compiler.ToolchainVersion {
		return fail()
	}
	b, err := io.ReadAll(io.LimitReader(in, topology.MaxRequestBytes+1))
	if err != nil || len(b) > topology.MaxRequestBytes {
		return fail()
	}
	result, err := topology.Execute(b)
	if err != nil {
		return fail()
	}
	n, err := out.Write(result)
	if err != nil || n != len(result) {
		_, _ = fmt.Fprintln(diag, "modeltopology: output failed")
		return 3
	}
	return 0
}
func main() {
	signal.Ignore(syscall.SIGPIPE)
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}
