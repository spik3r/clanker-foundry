//go:build !refgenfault

package main

// fault returns the name of the injected fault. Normal builds have none.
func fault() string { return "" }
