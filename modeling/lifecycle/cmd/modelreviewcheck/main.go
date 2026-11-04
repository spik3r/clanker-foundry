// modelreviewcheck produces a check-only first-design review composition.
package main

import (
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/internal/operatorcli"
	"io"
	"os"
	"os/signal"
	"syscall"
)

func run(args []string, in io.Reader, out, diag io.Writer) int {
	return operatorcli.RunReview(args, in, out, diag)
}

func main() {
	signal.Ignore(syscall.SIGPIPE)
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}
