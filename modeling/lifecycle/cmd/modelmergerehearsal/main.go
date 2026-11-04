// modelmergerehearsal checks supplied concurrent designs; only execute runs them.
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
	return operatorcli.RunMerge(ctx, args, in, out, diag)
}

func main() {
	signal.Ignore(syscall.SIGPIPE)
	ctx := context.Background()
	stop := func() {}
	if len(os.Args) == 2 && os.Args[1] == "execute" {
		ctx, stop = signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	}
	code := run(ctx, os.Args[1:], os.Stdin, os.Stdout, os.Stderr)
	stop()
	os.Exit(code)
}
