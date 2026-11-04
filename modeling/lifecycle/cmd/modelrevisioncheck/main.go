package main

import (
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/internal/operatorcli"
	"io"
	"os"
	"os/signal"
	"syscall"
)

func failure(err error) int { return operatorcli.RevisionFailure(err) }

func run(args []string, in io.Reader, out, diag io.Writer) int {
	return operatorcli.RunRevision(args, in, out, diag)
}

func main() {
	signal.Ignore(syscall.SIGPIPE)
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}
