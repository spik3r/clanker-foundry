package tui

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Foundry generator contract v1 constants.
const (
	ContractID       = "foundry-generator/v1"
	GeneratorID      = "tuic"
	GeneratorVersion = "1.2.0"
	// Toolchain is the Go release this generator was qualified with. It is a
	// constant, not detected at run time.
	Toolchain    = "go1.27.4"
	ManifestName = "foundry-manifest.json"
	manifestTag  = "foundry-generator-manifest/v1"
)

// ModelError marks a model that is unreadable, invalid or unsupported.
type ModelError struct{ Err error }

func (e ModelError) Error() string { return e.Err.Error() }
func (e ModelError) Unwrap() error { return e.Err }

// Result is everything the contract commands need for one model.
type Result struct {
	// Identity is the identity line, newline included.
	Identity string
	// Files holds every output file, the manifest included, keyed by
	// slash-separated relative path.
	Files map[string][]byte
}

func hexSum(b []byte) string {
	s := sha256.Sum256(b)
	return hex.EncodeToString(s[:])
}

// Build decodes and validates raw model bytes and computes the complete output.
// It performs no I/O. A bad model yields a ModelError.
func Build(raw []byte) (Result, error) {
	m, err := Decode(raw)
	if err != nil {
		return Result{}, ModelError{err}
	}
	if err := m.Validate(); err != nil {
		return Result{}, ModelError{err}
	}
	files, err := Compile(m)
	if err != nil {
		return Result{}, err
	}
	var manifest struct {
		Model struct {
			CanonicalSHA256 string `json:"canonicalSHA256"`
		} `json:"model"`
	}
	if err := json.Unmarshal(files[ManifestName], &manifest); err != nil {
		return Result{}, err
	}
	line := fmt.Sprintf(`{"contract":%q,"profile":%q,"sourceSHA256":%q,"canonicalSHA256":%q}`+"\n",
		ContractID, Profile, hexSum(raw), manifest.Model.CanonicalSHA256)
	return Result{Identity: line, Files: files}, nil
}

// manifestBytes encodes the manifest over files, which must not hold it yet.
func manifestBytes(canonicalSHA256 string, files map[string][]byte) ([]byte, error) {
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
	m.Profile = Profile
	m.Generator.ID = GeneratorID
	m.Generator.Version = GeneratorVersion
	m.Generator.Toolchain = Toolchain
	m.Model.CanonicalSHA256 = canonicalSHA256
	names := make([]string, 0, len(files))
	for name := range files {
		names = append(names, name)
	}
	sort.Strings(names)
	m.Files = make([]entry, 0, len(names))
	for _, name := range names {
		m.Files = append(m.Files, entry{name, hexSum(files[name]), len(files[name])})
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
		// Every path is compiler-owned; models cannot choose output paths.
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
