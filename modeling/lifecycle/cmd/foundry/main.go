// foundry prepares requests and exposes the existing native operator contracts.
package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/internal/operatorcli"
)

func main() {
	signal.Ignore(syscall.SIGPIPE)
	ctx := context.Background()
	stop := func() {}
	// Only the proposal CLI handles cancellation. Other native commands retain
	// their ordinary termination behavior rather than swallowing these signals.
	if len(os.Args) > 1 && os.Args[1] == "proposal" {
		ctx, stop = signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	}
	code := operatorcli.Run(ctx, os.Args[1:], os.Stdin, os.Stdout, os.Stderr)
	stop()
	os.Exit(code)
}
