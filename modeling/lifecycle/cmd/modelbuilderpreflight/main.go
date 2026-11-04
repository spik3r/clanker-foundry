// modelbuilderpreflight checks named input bytes without executing candidates.
package main

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/builderpreflight"
)

func run(args []string, in io.Reader, out, diag io.Writer) int {
	fail := func(message string, code int) int {
		_, _ = io.WriteString(diag, "modelbuilderpreflight: "+message+"\n")
		return code
	}
	values := map[string]string{}
	if len(args) != 4 {
		return fail("invalid arguments", 2)
	}
	for i := 0; i < len(args); i += 2 {
		if (args[i] != "--root" && args[i] != "--manifest-sha256") || values[args[i]] != "" || args[i+1] == "" {
			return fail("invalid arguments", 2)
		}
		values[args[i]] = args[i+1]
	}
	pin, root := values["--manifest-sha256"], values["--root"]
	if _, err := hex.DecodeString(pin); err != nil || len(pin) != 64 || strings.ToLower(pin) != pin || !filepath.IsAbs(root) {
		return fail("invalid arguments", 2)
	}
	raw, err := io.ReadAll(io.LimitReader(in, builderpreflight.MaxManifestBytes+1))
	if err != nil {
		return fail("input read failed", 4)
	}
	body, err := builderpreflight.Check(raw, root, pin)
	if err != nil {
		if errors.Is(err, builderpreflight.ErrInvalid) {
			return fail("input rejected", 2)
		}
		return fail("check failed", 4)
	}
	var envelope struct {
		Outcome string `json:"outcome"`
	}
	if json.Unmarshal(body, &envelope) != nil || (envelope.Outcome != "matched" && envelope.Outcome != "blocked") {
		return fail("check failed", 4)
	}
	if n, err := out.Write(body); err != nil || n != len(body) {
		return fail("output failed", 3)
	}
	if envelope.Outcome == "blocked" {
		return 1
	}
	return 0
}

func main() {
	// Keep closed-pipe delivery failures observable as exit 3.
	signal.Ignore(syscall.SIGPIPE)
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}
