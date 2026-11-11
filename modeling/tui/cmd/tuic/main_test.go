package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGenerateNewDirectoryAndRefuseOverwrite(t *testing.T) {
	root := t.TempDir()
	out := filepath.Join(root, "output")
	if err := run([]string{"../../examples/browser.json", out}); err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(filepath.Join(out, "records.go"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(out, "cmd/demo/main.go")); err != nil {
		t.Fatal(err)
	}
	if err := run([]string{"../../examples/browser.json", out}); err == nil {
		t.Fatal("overwrote existing output")
	}
	after, _ := os.ReadFile(filepath.Join(out, "records.go"))
	if string(before) != string(after) {
		t.Fatal("changed existing output")
	}
}
func TestInvalidModelLeavesNoOutput(t *testing.T) {
	root := t.TempDir()
	input := filepath.Join(root, "invalid.json")
	out := filepath.Join(root, "output")
	if err := os.WriteFile(input, []byte(`{"formatVersion":"wrong"}`), 0600); err != nil {
		t.Fatal(err)
	}
	if err := run([]string{input, out}); err == nil {
		t.Fatal("accepted invalid model")
	}
	if _, err := os.Lstat(out); !os.IsNotExist(err) {
		t.Fatal("created output for invalid input")
	}
}
func TestOutputSymlinkIsNotFollowed(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "target")
	out := filepath.Join(root, "out")
	if err := os.Mkdir(target, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, out); err != nil {
		t.Skip(err)
	}
	if err := run([]string{"../../examples/browser.json", out}); err == nil {
		t.Fatal("followed output symlink")
	}
	entries, _ := os.ReadDir(target)
	if len(entries) != 0 {
		t.Fatal("changed symlink target")
	}
}
