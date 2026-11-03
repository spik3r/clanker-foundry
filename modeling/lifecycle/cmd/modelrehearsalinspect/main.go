package main

import (
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/rehearsalinspection"
)

func run(args []string, in io.Reader, out, diagnostic io.Writer) int {
	if len(args) != 0 {
		_, _ = io.WriteString(diagnostic, "usage: modelrehearsalinspect < request.json\n")
		return 2
	}
	raw, err := io.ReadAll(io.LimitReader(in, rehearsalinspection.MaxRequestBytes+1))
	if err != nil {
		_, _ = io.WriteString(diagnostic, "inspection input rejected\n")
		return 2
	}
	receipt, err := rehearsalinspection.InspectRequest(raw)
	if err != nil {
		_, _ = io.WriteString(diagnostic, "inspection rejected\n")
		return 2
	}
	n, err := out.Write(receipt)
	if err != nil || n != len(receipt) {
		_, _ = io.WriteString(diagnostic, "inspection output failed\n")
		return 3
	}
	return 0
}
func main() {
	signal.Ignore(syscall.SIGPIPE)
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}
