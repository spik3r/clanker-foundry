// modelprofile prints a descriptive profile; it grants no authority.
package main

import (
	"fmt"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/compiler"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/profilecatalog"
	"io"
	"os"
	"os/signal"
	"runtime"
	"syscall"
)

func run(args []string, _ io.Reader, out, diag io.Writer) int {
	if len(args) != 0 || runtime.Version() != compiler.ToolchainVersion {
		_, _ = fmt.Fprintln(diag, "modelprofile: unavailable")
		return 2
	}
	b, e := profilecatalog.Encode()
	if e != nil {
		_, _ = fmt.Fprintln(diag, "modelprofile: unavailable")
		return 2
	}
	n, e := out.Write(b)
	if e != nil || n != len(b) {
		_, _ = fmt.Fprintln(diag, "modelprofile: output failed")
		return 3
	}
	return 0
}
func main() {
	// Return the declared write-error status even when stdout is a closed pipe.
	signal.Ignore(syscall.SIGPIPE)
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}
