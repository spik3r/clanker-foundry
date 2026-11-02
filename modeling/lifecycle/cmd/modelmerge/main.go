// modelmerge proposes a model merge through bounded stdin/stdout only.
package main

import (
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/modelmerge"
)

func run(args []string, in io.Reader, out, diag io.Writer) int {
	if len(args) != 0 {
		io.WriteString(diag, "modelmerge accepts stdin only\n")
		return 2
	}
	raw, err := io.ReadAll(io.LimitReader(in, modelmerge.MaxRequestBytes+1))
	if err != nil || len(raw) > modelmerge.MaxRequestBytes {
		io.WriteString(diag, "model merge input rejected\n")
		return 2
	}
	result, err := modelmerge.Execute(raw)
	if err != nil {
		io.WriteString(diag, "model merge rejected\n")
		return 2
	}
	n, err := out.Write(result)
	if err != nil || n != len(result) {
		io.WriteString(diag, "model merge output failed\n")
		return 3
	}
	return 0
}

func main() {
	// A real closed stdout pipe must reach Write's error path, not exit on SIGPIPE.
	signal.Ignore(syscall.SIGPIPE)
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}
