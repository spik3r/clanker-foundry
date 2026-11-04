// modelboundinventory inspects supplied assertions; it never runs a candidate.
package main

import (
	"errors"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/boundinventory"
)

func failure(err error) int {
	if errors.Is(err, boundinventory.ErrInvalid) || errors.Is(err, boundinventory.ErrBinding) || errors.Is(err, boundinventory.ErrLimit) {
		return 2
	}
	return 4
}
func run(args []string, in io.Reader, out, diag io.Writer) int {
	fail := func(code int, message string) int {
		_, _ = io.WriteString(diag, "modelboundinventory: "+message+"\n")
		return code
	}
	if len(args) != 0 {
		return fail(2, "accepts stdin only")
	}
	raw, err := io.ReadAll(io.LimitReader(in, boundinventory.MaxRequestBytes+1))
	if err != nil {
		return fail(4, "input read failed")
	}
	result, err := boundinventory.Check(raw)
	if err != nil {
		code := failure(err)
		if code == 2 {
			return fail(2, "input rejected")
		}
		return fail(4, "boundary failed")
	}
	n, err := out.Write(result)
	if err != nil || n != len(result) {
		return fail(3, "output failed")
	}
	return 0
}
func main() {
	signal.Ignore(syscall.SIGPIPE)
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}
