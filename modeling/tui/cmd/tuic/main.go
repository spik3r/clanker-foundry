// tuic emits a new importable record-browser package from a reviewed JSON model.
package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"

	tui "github.com/spik3r/clanker-foundry/modeling/tui"
)

func run(args []string) error {
	if len(args) != 2 {
		return fmt.Errorf("usage: tuic MODEL.json NEW_OUTPUT_DIRECTORY")
	}
	f, err := os.Open(args[0])
	if err != nil {
		return err
	}
	defer f.Close()
	raw, err := io.ReadAll(io.LimitReader(f, tui.MaxModelBytes+1))
	if err != nil {
		return err
	}
	model, err := tui.Decode(raw)
	if err != nil {
		return err
	}
	files, err := tui.Compile(model)
	if err != nil {
		return err
	}
	// Mkdir fails on any existing path, including a symlink. Never overwrite a host app.
	if err := os.Mkdir(args[1], 0755); err != nil {
		return err
	}
	names := make([]string, 0, len(files))
	for name := range files {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		// Every path is compiler-owned; models cannot choose output paths.
		path := filepath.Join(args[1], name)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return fmt.Errorf("partial output at %s: %w", args[1], err)
		}
		out, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
		if err != nil {
			return fmt.Errorf("partial output at %s: %w", args[1], err)
		}
		_, writeErr := out.Write(files[name])
		closeErr := out.Close()
		if writeErr != nil {
			return fmt.Errorf("partial output at %s: %w", args[1], writeErr)
		}
		if closeErr != nil {
			return fmt.Errorf("partial output at %s: %w", args[1], closeErr)
		}
	}
	return nil
}
func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
