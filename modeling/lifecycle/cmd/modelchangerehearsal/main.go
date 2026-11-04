// modelchangerehearsal checks supplied changes; only execute invokes candidates.
package main

import (
	"context"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/changerehearsal"
)

func run(ctx context.Context, args []string, in io.Reader, out, diag io.Writer) int {
	fail := func(code int, message string) int {
		_, _ = io.WriteString(diag, "modelchangerehearsal: "+message+"\n")
		return code
	}
	if len(args) != 1 || (args[0] != "check" && args[0] != "execute") {
		return fail(2, "usage: modelchangerehearsal check|execute < request.json")
	}
	raw, err := io.ReadAll(io.LimitReader(in, changerehearsal.MaxRequestBytes+1))
	if err != nil {
		return fail(2, "input read failed")
	}
	var report []byte
	if args[0] == "check" {
		report, err = changerehearsal.Check(raw)
	} else {
		report, err = changerehearsal.Execute(ctx, raw, ".")
	}
	code := changerehearsal.ExitCode(report, err)
	if len(report) == 0 {
		if code == 2 {
			return fail(2, "request rejected")
		}
		return fail(4, "boundary failure; execution may have started")
	}
	n, writeErr := out.Write(report)
	if writeErr != nil || n != len(report) {
		return fail(3, "report delivery failed; execution may have started")
	}
	if code == 4 {
		_, _ = io.WriteString(diag, "modelchangerehearsal: execution boundary failed; inspect receipt\n")
	}
	return code
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
