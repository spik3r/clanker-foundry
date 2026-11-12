//go:build refgenfault

package main

import "os"

// fault returns the fault named by REFGEN_FAULT. It exists only in binaries
// built with -tags refgenfault, which the harness self-tests use.
func fault() string { return os.Getenv("REFGEN_FAULT") }
