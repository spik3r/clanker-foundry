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
	// Proposal commands and explicit change/merge execution handle cancellation. Other
	// native commands retain ordinary termination rather than swallowing signals.
	if (len(os.Args) > 1 && os.Args[1] == "proposal") ||
		(len(os.Args) == 3 && (os.Args[1] == "change" || os.Args[1] == "merge") && os.Args[2] == "execute") {
		ctx, stop = signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	}
	code := operatorcli.Run(ctx, os.Args[1:], os.Stdin, os.Stdout, os.Stderr)
	stop()
	os.Exit(code)
}
