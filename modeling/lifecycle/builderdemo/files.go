package builderdemo

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/builderpreflight"
)

func trimSeparators(path string) string {
	minimum := len(filepath.VolumeName(path)) + 1
	for len(path) > minimum && os.IsPathSeparator(path[len(path)-1]) {
		path = path[:len(path)-1]
	}
	return path
}
func openDirectory(path string) (*os.Root, os.FileInfo, error) {
	before, err := os.Lstat(trimSeparators(path))
	if err != nil || !before.IsDir() || before.Mode()&os.ModeSymlink != 0 {
		return nil, nil, ErrInvalid
	}
	root, err := os.OpenRoot(path)
	if err != nil {
		return nil, nil, ErrInvalid
	}
	after, err := root.Stat(".")
	if err != nil || !os.SameFile(before, after) {
		root.Close()
		return nil, nil, ErrInvalid
	}
	return root, before, nil
}
func sameDirectory(path string, before os.FileInfo) bool {
	after, err := os.Lstat(trimSeparators(path))
	return err == nil && after.IsDir() && after.Mode()&os.ModeSymlink == 0 && os.SameFile(before, after)
}
func physicalDirectory(path string, before os.FileInfo) (string, error) {
	physical, err := filepath.EvalSymlinks(path)
	if err != nil || !sameDirectory(physical, before) {
		return "", ErrInvalid
	}
	return physical, nil
}

func collect(root *os.Root, item artifact) ([]byte, error) {
	blocked := func(reason string) ([]byte, error) {
		return nil, fmt.Errorf("%w: collect %s: %s", ErrBlocked, item.Path, reason)
	}
	parts := strings.Split(item.Path, "/")
	var before os.FileInfo
	for i := range parts {
		info, err := root.Lstat(filepath.FromSlash(strings.Join(parts[:i+1], "/")))
		if err != nil {
			return blocked("metadata read failed")
		}
		if info.Mode()&os.ModeSymlink != 0 || i < len(parts)-1 && !info.IsDir() {
			return blocked("path refused")
		}
		before = info
	}
	if !before.Mode().IsRegular() || before.Size() > builderpreflight.MaxArtifactBytes {
		return blocked("file type or size refused")
	}
	file, err := root.Open(filepath.FromSlash(item.Path))
	if err != nil {
		return blocked("open failed")
	}
	defer file.Close()
	opened, err := file.Stat()
	if err != nil || !opened.Mode().IsRegular() || !os.SameFile(before, opened) || opened.Size() != before.Size() || !opened.ModTime().Equal(before.ModTime()) {
		return blocked("file changed before read")
	}
	raw, err := io.ReadAll(io.LimitReader(file, builderpreflight.MaxArtifactBytes+1))
	if err != nil {
		return blocked("byte read failed")
	}
	if len(raw) > builderpreflight.MaxArtifactBytes {
		return blocked("file exceeds byte limit")
	}
	after, err := file.Stat()
	if err != nil || !os.SameFile(opened, after) || after.Size() != int64(len(raw)) || opened.Size() != after.Size() || !opened.ModTime().Equal(after.ModTime()) {
		return blocked("file changed during read")
	}
	// A descriptor can remain valid across rename-over. Recheck the named leaf,
	// not just the opened descriptor. This is not a hostile-writer guarantee.
	named, err := root.Lstat(filepath.FromSlash(item.Path))
	if err != nil || !named.Mode().IsRegular() || !os.SameFile(after, named) || named.Size() != after.Size() || !named.ModTime().Equal(after.ModTime()) {
		return blocked("named file changed during read")
	}
	if digest(raw) != item.SHA256 {
		return blocked("bytes differ from frozen hash")
	}
	return raw, nil
}

func writeNew(root *os.Root, name string, raw []byte) error {
	path := filepath.FromSlash(name)
	if err := root.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	file, err := root.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return err
	}
	n, writeErr := file.Write(raw)
	closeErr := file.Close()
	if writeErr != nil {
		return writeErr
	}
	if n != len(raw) {
		return io.ErrShortWrite
	}
	return closeErr
}
