package builderpreflight

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func openRoot(path string) (*os.Root, error) {
	// Preserve host traversal through ancestor aliases and "..". filepath.Clean
	// could select a different directory or erase a missing/non-directory parent.
	// Trim only trailing separators, so Lstat still sees a final symlink leaf.
	minimum := len(filepath.VolumeName(path)) + 1
	for len(path) > minimum && os.IsPathSeparator(path[len(path)-1]) {
		path = path[:len(path)-1]
	}
	before, err := os.Lstat(path)
	if err != nil || !before.IsDir() || before.Mode()&os.ModeSymlink != 0 {
		return nil, ErrInvalid
	}
	root, err := os.OpenRoot(path)
	if err != nil {
		return nil, ErrInvalid
	}
	after, err := root.Stat(".")
	if err != nil || !os.SameFile(before, after) {
		root.Close()
		return nil, ErrInvalid
	}
	return root, nil
}

func readStatus(err error) string {
	if os.IsNotExist(err) {
		return "missing"
	}
	return "read-failed"
}

func inspect(root *os.Root, item artifact) observation {
	result := observation{ID: item.ID, Role: item.Role, Path: item.Path, ExpectedSHA256: item.SHA256, Status: "refused"}
	parts := strings.Split(item.Path, "/")
	var leaf os.FileInfo
	for i := range parts {
		// Lstat each named component; neither expand globs nor list directories.
		info, err := root.Lstat(filepath.FromSlash(strings.Join(parts[:i+1], "/")))
		if err != nil {
			result.Status = readStatus(err)
			return result
		}
		if info.Mode()&os.ModeSymlink != 0 || i < len(parts)-1 && !info.IsDir() {
			return result
		}
		leaf = info
	}
	if !leaf.Mode().IsRegular() || leaf.Size() > MaxArtifactBytes {
		return result
	}
	file, err := root.Open(filepath.FromSlash(item.Path))
	if err != nil {
		result.Status = readStatus(err)
		return result
	}
	defer file.Close()
	opened, err := file.Stat()
	if err != nil {
		result.Status = "read-failed"
		return result
	}
	if !opened.Mode().IsRegular() || !os.SameFile(leaf, opened) || opened.Size() > MaxArtifactBytes {
		return result
	}
	hash := sha256.New()
	size, err := io.Copy(hash, io.LimitReader(file, MaxArtifactBytes+1))
	if err != nil {
		result.Status = "read-failed"
		return result
	}
	if size > MaxArtifactBytes {
		return result
	}
	// Detect ordinary edits during reading; this is not an atomic snapshot or a
	// defence against hostile concurrent writers, mounts or hardlink aliases.
	after, err := file.Stat()
	if err != nil {
		result.Status = "read-failed"
		return result
	}
	if !os.SameFile(opened, after) || after.Size() != size || opened.Size() != after.Size() || !opened.ModTime().Equal(after.ModTime()) {
		return result
	}
	observed := hex.EncodeToString(hash.Sum(nil))
	result.ObservedSHA256, result.ObservedBytes = &observed, &size
	result.Status = "hash-mismatch"
	if observed == item.SHA256 {
		result.Status = "matched"
	}
	return result
}
