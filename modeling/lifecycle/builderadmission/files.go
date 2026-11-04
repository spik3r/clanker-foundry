package builderadmission

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"sort"
	"unicode/utf8"
)

func isMissing(err error) bool { return os.IsNotExist(err) }
func openRoot(path string) (*os.Root, string, os.FileInfo, error) {
	minimum := len(filepath.VolumeName(path)) + 1
	for len(path) > minimum && os.IsPathSeparator(path[len(path)-1]) {
		path = path[:len(path)-1]
	}
	before, err := os.Lstat(path)
	if err != nil || !before.IsDir() || before.Mode()&os.ModeSymlink != 0 {
		return nil, "", nil, ErrInvalid
	}
	root, err := os.OpenRoot(path)
	if err != nil {
		return nil, "", nil, ErrInvalid
	}
	after, err := root.Stat(".")
	if err != nil || !os.SameFile(before, after) {
		root.Close()
		return nil, "", nil, ErrBoundary
	}
	return root, path, after, nil
}
func sameRoot(root *os.Root, path string, before os.FileInfo) bool {
	opened, e := root.Stat(".")
	named, n := os.Lstat(path)
	return e == nil && n == nil && named.IsDir() && named.Mode()&os.ModeSymlink == 0 && os.SameFile(before, opened) && os.SameFile(opened, named)
}
func kind(info os.FileInfo) string {
	switch {
	case info.Mode()&os.ModeSymlink != 0:
		return "symlink"
	case info.IsDir():
		return "directory"
	case info.Mode().IsRegular():
		return "regular"
	default:
		return "special"
	}
}
func census(root *os.Root) ([]entry, map[string]os.FileInfo, error) {
	dir, err := root.Open(".")
	if err != nil {
		return nil, nil, ErrBoundary
	}
	defer dir.Close()
	var names []os.DirEntry
	for {
		batch, readErr := dir.ReadDir(MaxEntries - len(names) + 1)
		names = append(names, batch...)
		if len(names) > MaxEntries {
			return nil, nil, ErrBoundary
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil || len(batch) == 0 {
			return nil, nil, ErrBoundary
		}
	}
	out := make([]entry, 0, len(names))
	infos := make(map[string]os.FileInfo, len(names))
	for _, name := range names {
		if !utf8.ValidString(name.Name()) {
			return nil, nil, ErrBoundary
		}
		info, err := root.Lstat(name.Name())
		if err != nil {
			return nil, nil, ErrBoundary
		}
		out = append(out, entry{name.Name(), kind(info)})
		infos[name.Name()] = info
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, infos, nil
}
func sameCensus(a []entry, before map[string]os.FileInfo, b []entry, after map[string]os.FileInfo) bool {
	if len(a) != len(b) {
		return false
	}
	for i, x := range a {
		if x != b[i] {
			return false
		}
		p, q := before[x.Name], after[x.Name]
		if !os.SameFile(p, q) || p.Mode() != q.Mode() || p.Size() != q.Size() || !p.ModTime().Equal(q.ModTime()) {
			return false
		}
	}
	return true
}
func inspectCandidate(root *os.Root, leaf os.FileInfo, pin string) observation {
	out := observation{Path: "candidate.go", Status: "missing"}
	if leaf == nil {
		return out
	}
	out.Status = "refused"
	if !leaf.Mode().IsRegular() || leaf.Size() > MaxCandidateBytes {
		return out
	}
	file, err := root.Open("candidate.go")
	if err != nil {
		out.Status = "read-failed"
		return out
	}
	defer file.Close()
	opened, err := file.Stat()
	if err != nil {
		out.Status = "read-failed"
		return out
	}
	if !opened.Mode().IsRegular() || opened.Size() > MaxCandidateBytes || !os.SameFile(leaf, opened) {
		return out
	}
	digest := sha256.New()
	size, err := io.Copy(digest, io.LimitReader(file, MaxCandidateBytes+1))
	if err != nil {
		out.Status = "read-failed"
		return out
	}
	if size > MaxCandidateBytes {
		return out
	}
	after, err := file.Stat()
	named, namedErr := root.Lstat("candidate.go")
	if err != nil || namedErr != nil {
		out.Status = "read-failed"
		return out
	}
	if !named.Mode().IsRegular() || !os.SameFile(opened, after) || !os.SameFile(after, named) || opened.Size() != size || after.Size() != size || named.Size() != size || !opened.ModTime().Equal(after.ModTime()) || !after.ModTime().Equal(named.ModTime()) {
		return out
	}
	observed := hex.EncodeToString(digest.Sum(nil))
	out.ObservedSHA256 = &observed
	out.ObservedBytes = &size
	out.Status = "hash-mismatch"
	if observed == pin {
		out.Status = "matched"
	}
	return out
}
