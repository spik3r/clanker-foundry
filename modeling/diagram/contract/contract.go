// Package contract holds the plumbing every diagram profile shares: the
// Foundry generator contract v1 identity line, manifest, output writing and
// drift checking, plus a strict JSON scanner. It knows nothing about any one
// profile.
package contract

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Foundry generator contract v1 constants.
const (
	ID           = "foundry-generator/v1"
	ManifestName = "foundry-manifest.json"
	manifestTag  = "foundry-generator-manifest/v1"
	// MaxModelBytes is the largest model file any profile accepts.
	MaxModelBytes = 512 << 10
)

// Generator names the generator for the manifest and for emitted metadata.
type Generator struct {
	ID        string
	Version   string
	Toolchain string
}

// Generated is what a profile produces from one model.
type Generated struct {
	// Canonical is the canonical encoding whose SHA-256 identifies the model.
	Canonical []byte
	// Files holds every output file except the manifest, keyed by
	// slash-separated relative path.
	Files map[string][]byte
}

// Profile turns raw model bytes into output files. Generate performs no I/O
// and returns a ModelError for a model that is invalid or unsupported.
type Profile interface {
	ID() string
	Generate(raw []byte, g Generator) (Generated, error)
}

// ModelError marks a model that is unreadable, invalid or unsupported.
type ModelError struct{ Err error }

func (e ModelError) Error() string { return e.Err.Error() }
func (e ModelError) Unwrap() error { return e.Err }

// Refuse builds a ModelError from a format string.
func Refuse(format string, a ...any) error {
	return ModelError{fmt.Errorf(format, a...)}
}

// Result is everything the contract commands need for one model.
type Result struct {
	// Identity is the identity line, newline included.
	Identity string
	// Files holds every output file, the manifest included.
	Files map[string][]byte
}

// HexSum returns the lowercase hex SHA-256 of b.
func HexSum(b []byte) string {
	s := sha256.Sum256(b)
	return hex.EncodeToString(s[:])
}

// Assemble runs a profile and adds the manifest and identity line.
func Assemble(p Profile, g Generator, raw []byte) (Result, error) {
	gen, err := p.Generate(raw, g)
	if err != nil {
		return Result{}, err
	}
	if len(gen.Files) == 0 {
		return Result{}, errors.New("profile produced no files")
	}
	canonical := HexSum(gen.Canonical)
	mb, err := manifestBytes(p.ID(), g, canonical, gen.Files)
	if err != nil {
		return Result{}, err
	}
	files := make(map[string][]byte, len(gen.Files)+1)
	for name, data := range gen.Files {
		files[name] = data
	}
	files[ManifestName] = mb
	line := fmt.Sprintf(`{"contract":%q,"profile":%q,"sourceSHA256":%q,"canonicalSHA256":%q}`+"\n",
		ID, p.ID(), HexSum(raw), canonical)
	return Result{Identity: line, Files: files}, nil
}

func manifestBytes(profile string, g Generator, canonical string, files map[string][]byte) ([]byte, error) {
	type entry struct {
		Path   string `json:"path"`
		SHA256 string `json:"sha256"`
		Bytes  int    `json:"bytes"`
	}
	var m struct {
		ManifestVersion string `json:"manifestVersion"`
		Profile         string `json:"profile"`
		Generator       struct {
			ID        string `json:"id"`
			Version   string `json:"version"`
			Toolchain string `json:"toolchain"`
		} `json:"generator"`
		Model struct {
			CanonicalSHA256 string `json:"canonicalSHA256"`
		} `json:"model"`
		Files []entry `json:"files"`
	}
	m.ManifestVersion = manifestTag
	m.Profile = profile
	m.Generator.ID = g.ID
	m.Generator.Version = g.Version
	m.Generator.Toolchain = g.Toolchain
	m.Model.CanonicalSHA256 = canonical
	names := make([]string, 0, len(files))
	for name := range files {
		names = append(names, name)
	}
	sort.Strings(names)
	m.Files = make([]entry, 0, len(names))
	for _, name := range names {
		m.Files = append(m.Files, entry{name, HexSum(files[name]), len(files[name])})
	}
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(m); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// ReadModel reads a model file of at most MaxModelBytes. Any failure is a
// ModelError, so the command refuses with exit 2.
func ReadModel(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, Refuse("cannot open model: %v", err)
	}
	defer f.Close()
	raw, err := io.ReadAll(io.LimitReader(f, MaxModelBytes+1))
	if err != nil {
		return nil, Refuse("cannot read model: %v", err)
	}
	if len(raw) > MaxModelBytes {
		return nil, Refuse("model is larger than %d bytes", MaxModelBytes)
	}
	return raw, nil
}

// ErrOutExists reports that the output path is already taken.
var ErrOutExists = errors.New("output path already exists")

// ErrNoParent reports that the parent of the output path is missing.
var ErrNoParent = errors.New("parent of output path does not exist")

// Write creates out and writes files, manifest last. It refuses an existing
// path (file, directory or symlink) and a missing parent, writing nothing.
func Write(out string, files map[string][]byte) error {
	if _, err := os.Lstat(out); err == nil {
		return ErrOutExists
	} else if !os.IsNotExist(err) {
		return err
	}
	if err := os.Mkdir(out, 0o755); err != nil {
		if os.IsExist(err) {
			return ErrOutExists
		}
		if os.IsNotExist(err) {
			return ErrNoParent
		}
		return err
	}
	names := make([]string, 0, len(files))
	for name := range files {
		if name != ManifestName {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	names = append(names, ManifestName)
	for _, name := range names {
		// Every path is generator-owned; models cannot choose output paths.
		p := filepath.Join(out, filepath.FromSlash(name))
		err := os.MkdirAll(filepath.Dir(p), 0o755)
		if err == nil {
			err = os.WriteFile(p, files[name], 0o644)
		}
		if err != nil {
			return fmt.Errorf("partial output at %s: %w", out, err)
		}
	}
	return nil
}

// Drift compares out with files and returns one "drift: KIND PATH" line per
// differing path, sorted. It never writes.
func Drift(files map[string][]byte, out string) ([]string, error) {
	fi, err := os.Lstat(out)
	if os.IsNotExist(err) {
		return []string{"drift: missing ."}, nil
	}
	if err != nil {
		return nil, err
	}
	if !fi.IsDir() { // a file, or a symlink even to a directory
		return []string{"drift: unexpected ."}, nil
	}
	dirs := map[string]bool{"": true}
	children := map[string]map[string]bool{}
	for name := range files {
		parts := strings.Split(name, "/")
		for i := range parts {
			parent := strings.Join(parts[:i], "/")
			if children[parent] == nil {
				children[parent] = map[string]bool{}
			}
			children[parent][parts[i]] = true
			if i < len(parts)-1 {
				dirs[strings.Join(parts[:i+1], "/")] = true
			}
		}
	}
	var lines []string
	var walk func(rel string) error
	walk = func(rel string) error {
		entries, err := os.ReadDir(filepath.Join(out, filepath.FromSlash(rel)))
		if err != nil {
			return err
		}
		seen := map[string]bool{}
		for _, e := range entries {
			seen[e.Name()] = true
			p := e.Name()
			if rel != "" {
				p = rel + "/" + p
			}
			want, isFile := files[p]
			switch {
			case isFile:
				if !e.Type().IsRegular() {
					lines = append(lines, "drift: unexpected "+p)
					continue
				}
				got, err := os.ReadFile(filepath.Join(out, filepath.FromSlash(p)))
				if err != nil {
					return err
				}
				if !bytes.Equal(got, want) {
					lines = append(lines, "drift: modified "+p)
				}
			case dirs[p]:
				if !e.IsDir() {
					lines = append(lines, "drift: unexpected "+p)
					continue
				}
				if err := walk(p); err != nil {
					return err
				}
			default:
				lines = append(lines, "drift: unexpected "+p)
			}
		}
		for name := range children[rel] {
			if !seen[name] {
				p := name
				if rel != "" {
					p = rel + "/" + name
				}
				lines = append(lines, "drift: missing "+p)
			}
		}
		return nil
	}
	if err := walk(""); err != nil {
		return nil, err
	}
	sort.Strings(lines)
	return lines, nil
}
