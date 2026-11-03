// modelreviewcheck produces a check-only first-design review composition.
package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/firstdesignreview"
)

func run(args []string, in io.Reader, out, diag io.Writer) int {
	fail := func(code int, msg string) int {
		_, _ = fmt.Fprintln(diag, "modelreviewcheck: "+msg)
		return code
	}
	if len(args) != 0 {
		return fail(2, "no arguments accepted")
	}
	raw, err := io.ReadAll(io.LimitReader(in, firstdesignreview.MaxRequestBytes+1))
	if err != nil || len(raw) > firstdesignreview.MaxRequestBytes {
		return fail(2, "input read or bound failed")
	}
	result, err := firstdesignreview.Check(raw)
	if err != nil {
		if errors.Is(err, firstdesignreview.ErrInvalid) || errors.Is(err, firstdesignreview.ErrLimit) {
			return fail(2, "invalid request or limit")
		}
		return fail(4, "review boundary failed")
	}
	n, err := out.Write(result)
	if err != nil || n != len(result) {
		return fail(3, "report delivery failed")
	}
	return 0
}

func main() {
	signal.Ignore(syscall.SIGPIPE)
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}
