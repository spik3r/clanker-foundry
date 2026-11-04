// modelprocess emits an inert review projection, not an executable process.
package main

import (
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/internal/operatorcli"
	"io"
	"os"
	"os/signal"
	"syscall"
)

func run(args []string, in io.Reader, out, diag io.Writer) int {
	return operatorcli.RunProcess(args, in, out, diag)
}

func main() {
	// Return the declared write-error status even when stdout is a closed pipe.
	signal.Ignore(syscall.SIGPIPE)
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}
