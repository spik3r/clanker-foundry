// modeldesignchange accepts bounded stdin only; it does not extract or execute
// archives. Suites remain unvalidated/unexecuted and their binding not inspected.
package main

import (
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/designchange"
	"io"
	"os"
)

func run(args []string, in io.Reader, out, diag io.Writer) int {
	if len(args) != 0 {
		io.WriteString(diag, "modeldesignchange accepts stdin only; suite unvalidated and unexecuted, binding not inspected\n")
		return 2
	}
	raw, e := io.ReadAll(io.LimitReader(in, designchange.MaxCarrierBytes+1))
	if e != nil || len(raw) > designchange.MaxCarrierBytes {
		io.WriteString(diag, "design-change input rejected; suite unvalidated and unexecuted, binding not inspected\n")
		return 2
	}
	b, e := designchange.Execute(raw)
	if e != nil {
		io.WriteString(diag, "design-change rejected; suite unvalidated and unexecuted, binding not inspected\n")
		return 2
	}
	n, e := out.Write(b)
	if e != nil || n != len(b) {
		io.WriteString(diag, "design-change output failed; suite unvalidated and unexecuted, binding not inspected\n")
		return 3
	}
	return 0
}
func main() { os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr)) }
