// modelrehearsalhandoff checks caller-supplied records; it does not execute them.
package main

import (
	"fmt"
	"io"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/compiler"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/rehearsalhandoff"
)

func run(args []string, in io.Reader, out, diag io.Writer) int {
	return runWith(args, in, out, diag, rehearsalhandoff.Pack, rehearsalhandoff.Inspect)
}
func runWith(args []string, in io.Reader, out, diag io.Writer, pack, inspect func([]byte) ([]byte, error)) int {
	fail := func() int { _, _ = fmt.Fprintln(diag, "modelrehearsalhandoff: invalid input"); return 2 }
	if len(args) != 1 || (args[0] != "pack" && args[0] != "verify") || runtime.Version() != compiler.ToolchainVersion {
		return fail()
	}
	raw, err := io.ReadAll(io.LimitReader(in, rehearsalhandoff.MaxRequestBytes+1))
	if err != nil || len(raw) > rehearsalhandoff.MaxRequestBytes {
		return fail()
	}
	var result []byte
	if args[0] == "pack" {
		result, err = pack(raw)
	} else {
		result, err = inspect(raw)
	}
	if err != nil {
		return fail()
	}
	n, err := out.Write(result)
	if err != nil || n != len(result) {
		_, _ = fmt.Fprintln(diag, "modelrehearsalhandoff: output failed")
		return 3
	}
	return 0
}
func main() {
	signal.Ignore(syscall.SIGPIPE)
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}
