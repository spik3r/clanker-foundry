// modelintakediff compares upstream intent and assertions, not verified impact.
package main

import (
	"fmt"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/compiler"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/intakeimpact"
	"io"
	"os"
	"os/signal"
	"runtime"
	"syscall"
)

func run(args []string, in io.Reader, out, diagnostic io.Writer) int {
	fail := func(code int, s string) int { _, _ = fmt.Fprintln(diagnostic, "modelintakediff: "+s); return code }
	if len(args) != 0 {
		return fail(2, "usage: modelintakediff < request.json")
	}
	if runtime.Version() != compiler.ToolchainVersion {
		return fail(2, "unsupported Go toolchain")
	}
	raw, e := io.ReadAll(io.LimitReader(in, intakeimpact.MaxRequestBytes+1))
	if e != nil {
		return fail(2, "input read failed")
	}
	if len(raw) > intakeimpact.MaxRequestBytes {
		return fail(2, "invalid request or report limit")
	}
	result, e := intakeimpact.Execute(raw)
	if e != nil {
		return fail(2, "invalid request or report limit")
	}
	n, e := out.Write(result)
	if e != nil || n != len(result) {
		return fail(3, "output write failed")
	}
	return 0
}
func main() {
	signal.Ignore(syscall.SIGPIPE)
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}
