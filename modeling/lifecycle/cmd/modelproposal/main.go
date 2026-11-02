// modelproposal checks original proposals; only the explicit execute subcommand
// compiles and invokes a candidate on a trusted local host.
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/proposalreplay"
)

func run(ctx context.Context, args []string, in io.Reader, out, diag io.Writer) int {
	fail := func(code int, message string) int { _, _ = fmt.Fprintln(diag, "modelproposal: "+message); return code }
	if len(args) != 1 || (args[0] != "check" && args[0] != "execute") {
		return fail(2, "usage: modelproposal check|execute < request.json")
	}
	raw, err := io.ReadAll(io.LimitReader(in, proposalreplay.MaxRequestBytes+1))
	if err != nil || len(raw) > proposalreplay.MaxRequestBytes {
		return fail(2, "input read or bound failed")
	}
	var receipt proposalreplay.Receipt
	if args[0] == "check" {
		receipt, err = proposalreplay.Check(raw)
	} else {
		receipt, err = proposalreplay.Execute(ctx, raw, ".")
	}
	if errors.Is(err, proposalreplay.ErrInvalid) {
		return fail(2, "invalid proposal")
	}
	code := proposalreplay.ExitCode(receipt, err)
	encoded, encodeErr := proposalreplay.Encode(receipt)
	if encodeErr != nil {
		return fail(4, "receipt encoding failed; execution may have started")
	}
	n, writeErr := out.Write(encoded)
	if writeErr != nil || n != len(encoded) {
		return fail(3, "receipt delivery failed; execution may have started")
	}
	if code == 4 {
		_, _ = fmt.Fprintln(diag, "modelproposal: execution boundary failed; inspect receipt")
	}
	return code
}

func main() {
	signal.Ignore(syscall.SIGPIPE)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	code := run(ctx, os.Args[1:], os.Stdin, os.Stdout, os.Stderr)
	stop()
	os.Exit(code)
}
