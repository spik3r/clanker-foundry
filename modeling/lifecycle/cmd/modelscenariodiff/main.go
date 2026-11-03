// modelscenariodiff compares literal fixtures through bounded stdin/stdout only.
package main

import (
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/scenariodiff"
)

func run(args []string, in io.Reader, out, diag io.Writer) int {
	if len(args) != 0 {
		io.WriteString(diag, "modelscenariodiff accepts stdin only\n")
		return 2
	}
	raw, err := io.ReadAll(io.LimitReader(in, scenariodiff.MaxRequestBytes+1))
	if err != nil || len(raw) > scenariodiff.MaxRequestBytes {
		io.WriteString(diag, "scenario comparison input rejected\n")
		return 2
	}
	report, err := scenariodiff.Execute(raw)
	if err != nil {
		io.WriteString(diag, "scenario comparison rejected\n")
		return 2
	}
	n, err := out.Write(report)
	if err != nil || n != len(report) {
		io.WriteString(diag, "scenario comparison output failed\n")
		return 3
	}
	return 0
}

func main() {
	// Ensure closed stdout pipes reach Write's failure path rather than SIGPIPE.
	signal.Ignore(syscall.SIGPIPE)
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}
