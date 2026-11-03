// modelproposal checks original proposals; only the explicit execute subcommand
// compiles and invokes a candidate on a trusted local host.
package main

import (
	"context"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/internal/operatorcli"
	"io"
	"os"
	"os/signal"
	"syscall"
)

func run(ctx context.Context, args []string, in io.Reader, out, diag io.Writer) int {
	return operatorcli.RunProposal(ctx, args, in, out, diag)
}

func main() {
	signal.Ignore(syscall.SIGPIPE)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	code := run(ctx, os.Args[1:], os.Stdin, os.Stdout, os.Stderr)
	stop()
	os.Exit(code)
}
