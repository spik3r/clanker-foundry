// Package foundrygen reads a project's foundry.json and drives Foundry
// generators (contract foundry-generator/v1) to regenerate or check its outputs.
package foundrygen

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Exit codes.
const (
	exitOK      = 0
	exitFailure = 1
	exitRefused = 2
	exitDrift   = 3
)

const (
	// ProjectVersion is the only accepted projectVersion value.
	ProjectVersion = "foundry-project/v1"
	// ProjectFile is the project file name in the project root.
	ProjectFile = "foundry.json"
	// RegistryVersion is the only accepted registryVersion value.
	RegistryVersion = "foundry-generator-registry/v1"

	maxProjectBytes  = 64 << 10
	maxRegistryBytes = 1 << 20
	maxOutputs       = 64
)

var (
	revisionRE = regexp.MustCompile(`^[0-9a-f]{40}$`)
	idRE       = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*$`)
)

// codedError carries the exit code a failure maps to.
type codedError struct {
	code int
	msg  string
}

func (e *codedError) Error() string { return e.msg }

func refuse(format string, a ...any) error {
	return &codedError{exitRefused, fmt.Sprintf(format, a...)}
}

func fail(format string, a ...any) error {
	return &codedError{exitFailure, fmt.Sprintf(format, a...)}
}

func codeOf(err error) int {
	var c *codedError
	if errors.As(err, &c) {
		return c.code
	}
	return exitFailure
}

// Output is one declared generator output.
type Output struct {
	Generator string
	Model     string // slash path relative to the project root
	Out       string // slash path relative to the project root
}

// Project is a decoded foundry.json.
type Project struct {
	Revision string
	Outputs  []Output
}

// GeneratorEntry is one registry row the pipeline needs.
type GeneratorEntry struct {
	ID      string `json:"id"`
	Module  string `json:"module"`
	Package string `json:"package"`
}

// Registry maps generator ids to their entries.
type Registry map[string]GeneratorEntry

// checkJSONShape refuses duplicate object keys at any depth and trailing data.
func checkJSONShape(data []byte) error {
	dec := json.NewDecoder(bytes.NewReader(data))
	if err := walkJSON(dec); err != nil {
		return err
	}
	if _, err := dec.Token(); err != io.EOF {
		return errors.New("trailing data after the JSON value")
	}
	return nil
}

func walkJSON(dec *json.Decoder) error {
	tok, err := dec.Token()
	if err != nil {
		return err
	}
	d, ok := tok.(json.Delim)
	if !ok {
		return nil
	}
	switch d {
	case '{':
		seen := map[string]bool{}
		for dec.More() {
			k, err := dec.Token()
			if err != nil {
				return err
			}
			key, _ := k.(string)
			if seen[key] {
				return fmt.Errorf("duplicate key %q", key)
			}
			seen[key] = true
			if err := walkJSON(dec); err != nil {
				return err
			}
		}
	case '[':
		for dec.More() {
			if err := walkJSON(dec); err != nil {
				return err
			}
		}
	}
	_, err = dec.Token() // closing delimiter
	return err
}

// object decodes raw as a JSON object with exactly the given keys.
func object(raw json.RawMessage, what string, keys ...string) (map[string]json.RawMessage, error) {
	var m map[string]json.RawMessage
	if err := json.Unmarshal(raw, &m); err != nil || m == nil {
		return nil, fmt.Errorf("%s must be an object", what)
	}
	allowed := map[string]bool{}
	for _, k := range keys {
		allowed[k] = true
		if _, ok := m[k]; !ok {
			return nil, fmt.Errorf("%s: missing key %q", what, k)
		}
	}
	for k := range m {
		if !allowed[k] {
			return nil, fmt.Errorf("%s: unknown key %q", what, k)
		}
	}
	return m, nil
}

func str(raw json.RawMessage, what string) (string, error) {
	var s string
	if bytes.Equal(bytes.TrimSpace(raw), []byte("null")) || json.Unmarshal(raw, &s) != nil {
		return "", fmt.Errorf("%s must be a string", what)
	}
	return s, nil
}

// ValidRelPath checks a relative slash path inside a project.
func ValidRelPath(p string) error {
	if p == "" {
		return errors.New("empty path")
	}
	if strings.Contains(p, `\`) {
		return errors.New("backslash in path")
	}
	if strings.HasPrefix(p, "/") {
		return errors.New("absolute path")
	}
	for _, r := range p {
		if r < 0x20 || r == 0x7f {
			return errors.New("control character in path")
		}
	}
	for _, seg := range strings.Split(p, "/") {
		switch seg {
		case "":
			return errors.New("empty path segment")
		case ".", "..":
			return fmt.Errorf("%q path segment", seg)
		}
	}
	return nil
}

func within(p, dir string) bool { return p == dir || strings.HasPrefix(p, dir+"/") }

// ParseProject decodes and validates project file bytes against a registry.
func ParseProject(data []byte, reg Registry) (*Project, error) {
	bad := func(format string, a ...any) error {
		return refuse("%s: %s", ProjectFile, fmt.Sprintf(format, a...))
	}
	if len(data) > maxProjectBytes {
		return nil, bad("larger than %d bytes", maxProjectBytes)
	}
	if err := checkJSONShape(data); err != nil {
		return nil, bad("%v", err)
	}
	top, err := object(data, "project", "projectVersion", "foundry", "outputs")
	if err != nil {
		return nil, bad("%v", err)
	}
	ver, err := str(top["projectVersion"], "projectVersion")
	if err != nil {
		return nil, bad("%v", err)
	}
	if ver != ProjectVersion {
		return nil, bad("projectVersion must be %q", ProjectVersion)
	}
	f, err := object(top["foundry"], "foundry", "revision")
	if err != nil {
		return nil, bad("%v", err)
	}
	rev, err := str(f["revision"], "foundry.revision")
	if err != nil {
		return nil, bad("%v", err)
	}
	if !revisionRE.MatchString(rev) {
		return nil, bad("foundry.revision must be 40 lowercase hex digits")
	}
	var rawOuts []json.RawMessage
	if err := json.Unmarshal(top["outputs"], &rawOuts); err != nil || rawOuts == nil {
		return nil, bad("outputs must be an array")
	}
	if len(rawOuts) < 1 || len(rawOuts) > maxOutputs {
		return nil, bad("outputs must hold 1 to %d entries, has %d", maxOutputs, len(rawOuts))
	}
	p := &Project{Revision: rev}
	for i, ro := range rawOuts {
		what := fmt.Sprintf("outputs[%d]", i)
		m, err := object(ro, what, "generator", "model", "out")
		if err != nil {
			return nil, bad("%v", err)
		}
		var o Output
		for _, fld := range []struct {
			key string
			dst *string
		}{{"generator", &o.Generator}, {"model", &o.Model}, {"out", &o.Out}} {
			if *fld.dst, err = str(m[fld.key], what+"."+fld.key); err != nil {
				return nil, bad("%v", err)
			}
		}
		if _, ok := reg[o.Generator]; !ok {
			return nil, bad("%s.generator %q is not in the Foundry registry", what, o.Generator)
		}
		if err := ValidRelPath(o.Model); err != nil {
			return nil, bad("%s.model %q: %v", what, o.Model, err)
		}
		if err := ValidRelPath(o.Out); err != nil {
			return nil, bad("%s.out %q: %v", what, o.Out, err)
		}
		p.Outputs = append(p.Outputs, o)
	}
	for i, a := range p.Outputs {
		if within(ProjectFile, a.Out) {
			return nil, bad("outputs[%d].out %q contains %s", i, a.Out, ProjectFile)
		}
		for j, b := range p.Outputs {
			if within(b.Model, a.Out) {
				return nil, bad("outputs[%d].out %q contains outputs[%d].model %q", i, a.Out, j, b.Model)
			}
			if j > i && (within(a.Out, b.Out) || within(b.Out, a.Out)) {
				return nil, bad("outputs[%d].out %q and outputs[%d].out %q are equal or nested", i, a.Out, j, b.Out)
			}
		}
	}
	return p, nil
}

// LoadRegistry reads the Foundry checkout's generator registry.
func LoadRegistry(foundry string) (Registry, error) {
	path := filepath.Join(foundry, "modeling", "gencontract", "generators.json")
	data, err := readLimited(path, maxRegistryBytes)
	if err != nil {
		return nil, refuse("registry: %v", err)
	}
	var doc struct {
		RegistryVersion string           `json:"registryVersion"`
		Generators      []GeneratorEntry `json:"generators"`
	}
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, refuse("registry %s: %v", path, err)
	}
	if doc.RegistryVersion != RegistryVersion {
		return nil, refuse("registry %s: registryVersion must be %q", path, RegistryVersion)
	}
	reg := Registry{}
	for _, g := range doc.Generators {
		if !idRE.MatchString(g.ID) {
			return nil, refuse("registry %s: bad generator id %q", path, g.ID)
		}
		if _, dup := reg[g.ID]; dup {
			return nil, refuse("registry %s: duplicate generator id %q", path, g.ID)
		}
		if err := ValidRelPath(g.Module); err != nil {
			return nil, refuse("registry %s: %s module %q: %v", path, g.ID, g.Module, err)
		}
		if !strings.HasPrefix(g.Package, "./") || strings.ContainsAny(g.Package, " \t\n\\") {
			return nil, refuse("registry %s: %s package %q must be a ./ relative package path", path, g.ID, g.Package)
		}
		reg[g.ID] = g
	}
	return reg, nil
}

func readLimited(path string, limit int64) ([]byte, error) {
	fi, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	if !fi.Mode().IsRegular() {
		return nil, fmt.Errorf("%s is not a regular file", path)
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	data, err := io.ReadAll(io.LimitReader(f, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > limit {
		return nil, fmt.Errorf("%s is larger than %d bytes", path, limit)
	}
	return data, nil
}

// LoadProject reads root/foundry.json.
func LoadProject(root string, reg Registry) (*Project, error) {
	path := filepath.Join(root, ProjectFile)
	if fi, err := os.Lstat(path); err != nil {
		return nil, refuse("%s: %v", ProjectFile, err)
	} else if !fi.Mode().IsRegular() {
		return nil, refuse("%s is not a regular file", path)
	}
	data, err := readLimited(path, maxProjectBytes)
	if err != nil {
		return nil, refuse("%v", err)
	}
	return ParseProject(data, reg)
}

// resolveInProject joins rel to root and refuses when any existing component
// is a symlink or a non-directory in a non-final position.
func resolveInProject(root, rel string) (string, error) {
	cur := root
	segs := strings.Split(rel, "/")
	for i, seg := range segs {
		cur = filepath.Join(cur, seg)
		fi, err := os.Lstat(cur)
		if errors.Is(err, os.ErrNotExist) {
			break
		}
		if err != nil {
			return "", fail("%s: %v", rel, err)
		}
		if fi.Mode()&os.ModeSymlink != 0 {
			return "", refuse("%s: %q is a symlink", rel, strings.Join(segs[:i+1], "/"))
		}
		if i < len(segs)-1 && !fi.IsDir() {
			return "", refuse("%s: %q is not a directory", rel, strings.Join(segs[:i+1], "/"))
		}
	}
	return filepath.Join(root, filepath.FromSlash(rel)), nil
}
