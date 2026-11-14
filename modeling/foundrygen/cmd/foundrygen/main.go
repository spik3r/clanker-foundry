// Command foundrygen regenerates and checks the outputs a project declares in foundry.json.
package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/spik3r/clanker-foundry/modeling/foundrygen"
)

func main() {
	// Let closed output pipes reach ordinary delivery-failure handling.
	signal.Ignore(syscall.SIGPIPE)
	os.Exit(foundrygen.Run(os.Args[1:], os.Stdout, os.Stderr))
}
