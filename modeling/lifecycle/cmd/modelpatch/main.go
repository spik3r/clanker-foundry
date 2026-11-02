// modelpatch proposes a candidate via bounded stdin/stdout, without file writes.
package main

import (
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/modelpatch"
	"io"
	"os"
)

func run(args []string, in io.Reader, out, diag io.Writer) int {
	if len(args) != 0 {
		io.WriteString(diag, "modelpatch accepts stdin only\n")
		return 2
	}
	raw, e := io.ReadAll(io.LimitReader(in, modelpatch.MaxRequestBytes+1))
	if e != nil || len(raw) > modelpatch.MaxRequestBytes {
		io.WriteString(diag, "model patch input rejected\n")
		return 2
	}
	b, e := modelpatch.Execute(raw)
	if e != nil {
		io.WriteString(diag, "model patch rejected\n")
		return 2
	}
	n, e := out.Write(b)
	if e != nil || n != len(b) {
		io.WriteString(diag, "model patch output failed\n")
		return 3
	}
	return 0
}
func main() { os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr)) }
