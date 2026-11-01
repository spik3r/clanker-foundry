// modelintake reports declared and observed identities; it grants no authority.
package main

import (
	"fmt"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/compiler"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/intake"
	"io"
	"os"
	"runtime"
)

func run(args []string, in io.Reader, out, diagnostic io.Writer) int {
	fail := func(code int, s string) int { _, _ = fmt.Fprintln(diagnostic, "modelintake: "+s); return code }
	if len(args) != 0 {
		return fail(2, "usage: modelintake < request.json")
	}
	if runtime.Version() != compiler.ToolchainVersion {
		return fail(2, "unsupported Go toolchain")
	}
	b, e := io.ReadAll(io.LimitReader(in, intake.MaxSourceBytes+1))
	if e != nil {
		return fail(2, "input read failed")
	}
	if len(b) > intake.MaxSourceBytes {
		return fail(2, "invalid intake request")
	}
	r, e := intake.Execute(b)
	if e != nil {
		return fail(2, "invalid intake request")
	}
	n, e := out.Write(r)
	if e != nil || n != len(r) {
		return fail(3, "output write failed")
	}
	return 0
}
func main() { os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr)) }
