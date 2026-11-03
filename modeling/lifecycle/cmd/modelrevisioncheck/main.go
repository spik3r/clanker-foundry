package main

import (
	"errors"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/revisioncheck"
)

func failure(err error) int {
	if errors.Is(err, revisioncheck.ErrInvalid) || errors.Is(err, revisioncheck.ErrLimit) {
		return 2
	}
	return 4
}
func run(args []string, in io.Reader, out, diag io.Writer) int {
	if len(args) != 0 {
		_, _ = io.WriteString(diag, "modelrevisioncheck: invalid request\n")
		return 2
	}
	raw, err := io.ReadAll(io.LimitReader(in, revisioncheck.MaxRequestBytes+1))
	if err != nil {
		_, _ = io.WriteString(diag, "modelrevisioncheck: input failed\n")
		return 2
	}
	report, err := revisioncheck.Check(raw)
	if err != nil {
		_, _ = io.WriteString(diag, "modelrevisioncheck: request rejected\n")
		return failure(err)
	}
	n, err := out.Write(report)
	if err != nil || n != len(report) {
		_, _ = io.WriteString(diag, "modelrevisioncheck: output failed\n")
		return 3
	}
	return 0
}
func main() {
	signal.Ignore(syscall.SIGPIPE)
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}
