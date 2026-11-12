// Command foundrygen regenerates and checks the outputs a project declares in foundry.json.
package main

import (
	"os"

	"github.com/spik3r/clanker-foundry/modeling/foundrygen"
)

func main() {
	os.Exit(foundrygen.Run(os.Args[1:], os.Stdout, os.Stderr))
}
