// modelscenariotrace links requirements to literal scenario cases through stdin.
package main

import (
	"errors"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/scenariotrace"
)

func failure(err error) int {
	if errors.Is(err, scenariotrace.ErrInvalid) || errors.Is(err, scenariotrace.ErrBinding) || errors.Is(err, scenariotrace.ErrLimit) {
		return 2
	}
	return 4
}

func run(args []string, in io.Reader, out, diag io.Writer) int {
	if len(args) != 0 {
		io.WriteString(diag, "modelscenariotrace accepts stdin only\n")
		return 2
	}
	raw, err := io.ReadAll(io.LimitReader(in, scenariotrace.MaxRequestBytes+1))
	if err != nil {
		io.WriteString(diag, "scenario trace input read failed\n")
		return 4
	}
	if len(raw) > scenariotrace.MaxRequestBytes {
		io.WriteString(diag, "scenario trace input rejected\n")
		return 2
	}
	report, err := scenariotrace.Check(raw)
	if err != nil {
		code := failure(err)
		if code == 2 {
			io.WriteString(diag, "scenario trace input rejected\n")
		} else {
			io.WriteString(diag, "scenario trace boundary failed\n")
		}
		return code
	}
	n, err := out.Write(report)
	if err != nil || n != len(report) {
		io.WriteString(diag, "scenario trace output failed\n")
		return 3
	}
	return 0
}

func main() {
	// Let Write report a closed pipe instead of terminating with SIGPIPE.
	signal.Ignore(syscall.SIGPIPE)
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}
