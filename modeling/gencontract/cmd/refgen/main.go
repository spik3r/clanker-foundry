// Command refgen is the reference Foundry generator. It implements
// foundry-generator/v1 for a toy profile, reference/v1, and is the smallest
// example of a conforming generator.
//
// A model names a Go identifier and lists some strings:
//
//	{"profile":"reference/v1","name":"Greetings","items":["hello","world"]}
//
// The order of items is not significant: the canonical form sorts them.
package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"go/format"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

const (
	generatorID      = "refgen"
	generatorVersion = "1.0.0"
	// toolchain names the toolchain this generator was qualified with. It is a
	// constant, not detected at run time.
	toolchain = "go1.27.4"
	profile   = "reference/v1"

	contract        = "foundry-generator/v1"
	manifestVersion = "foundry-generator-manifest/v1"
	manifestName    = "foundry-manifest.json"
	maxModelBytes   = 64 << 10

	exitFailure = 1
	exitRefused = 2
	exitDrift   = 3
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

// refusal is an error that maps to exit code 2: nothing was written.
type refusal struct {
	msg   string
	usage bool
}

func (r refusal) Error() string { return r.msg }

func refuse(format string, a ...any) error {
	return refusal{msg: fmt.Sprintf(format, a...)}
}

// run executes one command line and returns the process exit code. On success
// stdout is exactly the identity line; on any failure stdout is empty.
func run(args []string, stdout, stderr io.Writer) int {
	line, drift, err := dispatch(args)
	switch {
	case err != nil:
		code := exitFailure
		var r refusal
		if errors.As(err, &r) {
			code = exitRefused
			if r.usage && fault() == "usage-exit1" {
				code = exitFailure
			}
		}
		fmt.Fprintf(stderr, "refgen: %v\n", err)
		if fault() == "stdout-on-failure" {
			fmt.Fprintln(stdout, "failed")
		}
		return code
	case len(drift) > 0:
		for _, d := range drift {
			fmt.Fprintln(stderr, d)
		}
		return exitDrift
	}
	fmt.Fprint(stdout, line)
	return 0
}

func dispatch(args []string) (line string, drift []string, err error) {
	switch {
	case len(args) == 2 && args[0] == "validate":
		l, err := loadModel(args[1])
		if err != nil {
			return "", nil, err
		}
		return l.identityLine(), nil, nil
	case len(args) == 3 && args[0] == "generate":
		l, err := loadModel(args[1])
		if err != nil {
			return "", nil, err
		}
		if err := generate(l, args[2]); err != nil {
			return "", nil, err
		}
		return l.identityLine(), nil, nil
	case len(args) == 3 && args[0] == "check":
		l, err := loadModel(args[1])
		if err != nil {
			return "", nil, err
		}
		drift, err := check(l, args[2])
		if err != nil {
			return "", nil, err
		}
		if len(drift) > 0 {
			return "", drift, nil
		}
		return l.identityLine(), nil, nil
	}
	return "", nil, refusal{
		msg:   "usage: refgen validate MODEL | generate MODEL OUT | check MODEL OUT",
		usage: true,
	}
}

// model is the reference/v1 model.
type model struct {
	Profile string   `json:"profile"`
	Name    string   `json:"name"`
	Items   []string `json:"items"`
}

// loaded is a validated model together with its source bytes.
type loaded struct {
	model  model
	source []byte
}

func loadModel(path string) (loaded, error) {
	f, err := os.Open(path)
	if err != nil {
		return loaded{}, refuse("cannot open model: %v", err)
	}
	defer f.Close()
	data, err := io.ReadAll(io.LimitReader(f, maxModelBytes+1))
	if err != nil {
		return loaded{}, refuse("cannot read model: %v", err)
	}
	if len(data) > maxModelBytes {
		return loaded{}, refuse("model is larger than %d bytes", maxModelBytes)
	}
	m, err := parseModel(data)
	if err != nil {
		if fault() == "accept-invalid" {
			m = model{Profile: profile, Name: "Fallback", Items: []string{"x"}}
		} else {
			return loaded{}, refuse("invalid model: %v", err)
		}
	}
	l := loaded{model: m, source: data}
	sortItems(&l)
	return l, nil
}

// sortItems puts items in canonical order; their order in the source is not
// significant.
func sortItems(l *loaded) { sort.Strings(l.model.Items) }

// parseModel decodes strictly: one JSON object, known members only, no
// duplicates, no trailing data.
func parseModel(data []byte) (model, error) {
	var m model
	if !utf8.Valid(data) {
		return m, errors.New("not valid UTF-8")
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	tok, err := dec.Token()
	if err != nil {
		return m, fmt.Errorf("not JSON: %w", err)
	}
	if d, ok := tok.(json.Delim); !ok || d != '{' {
		return m, errors.New("model must be a JSON object")
	}
	seen := map[string]bool{}
	for dec.More() {
		kt, err := dec.Token()
		if err != nil {
			return m, err
		}
		key, _ := kt.(string)
		if seen[key] {
			return m, fmt.Errorf("duplicate member %q", key)
		}
		seen[key] = true
		switch key {
		case "profile":
			err = dec.Decode(&m.Profile)
		case "name":
			err = dec.Decode(&m.Name)
		case "items":
			err = dec.Decode(&m.Items)
		default:
			return m, fmt.Errorf("unknown member %q", key)
		}
		if err != nil {
			return m, fmt.Errorf("member %q: %w", key, err)
		}
	}
	if _, err := dec.Token(); err != nil { // closing brace
		return m, err
	}
	if _, err := dec.Token(); err != io.EOF {
		return m, errors.New("trailing data after the model object")
	}
	return m, m.validate()
}

func (m model) validate() error {
	if m.Profile != profile {
		return fmt.Errorf("profile must be %q", profile)
	}
	if !validName(m.Name) {
		return errors.New("name must be an ASCII Go exported identifier of 1 to 32 characters")
	}
	if len(m.Items) < 1 || len(m.Items) > 8 {
		return errors.New("items must hold 1 to 8 strings")
	}
	seen := map[string]bool{}
	for _, it := range m.Items {
		if it == "" || len(it) > 32 {
			return errors.New("each item must be 1 to 32 bytes")
		}
		for _, r := range it {
			if unicode.IsControl(r) {
				return errors.New("items must not contain control characters")
			}
		}
		if seen[it] {
			return fmt.Errorf("duplicate item %q", it)
		}
		seen[it] = true
	}
	return nil
}

func validName(s string) bool {
	if len(s) < 1 || len(s) > 32 || s[0] < 'A' || s[0] > 'Z' {
		return false
	}
	for i := 1; i < len(s); i++ {
		c := s[i]
		ok := c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '_'
		if !ok {
			return false
		}
	}
	return true
}

func sum(b []byte) string {
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}

// canonical is the canonical encoding: compact JSON, members in the order
// profile, name, items, items sorted, no trailing newline.
func (l loaded) canonical() []byte {
	b, _ := encodeJSON(l.model, false)
	return bytes.TrimSuffix(b, []byte("\n"))
}

func (l loaded) canonicalSHA256() string {
	if fault() == "canonical-whitespace" {
		return sum(l.source) // wrong: depends on insignificant whitespace
	}
	return sum(l.canonical())
}

func (l loaded) identityLine() string {
	return fmt.Sprintf(`{"contract":%q,"profile":%q,"sourceSHA256":%q,"canonicalSHA256":%q}`+"\n",
		contract, profile, sum(l.source), l.canonicalSHA256())
}

func encodeJSON(v any, indent bool) ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if indent {
		enc.SetIndent("", "  ")
	}
	err := enc.Encode(v)
	return buf.Bytes(), err
}

// file is one output file; path uses "/" separators.
type file struct {
	path string
	data []byte
}

type manifestFile struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
	Bytes  int    `json:"bytes"`
}

type manifest struct {
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
	Files []manifestFile `json:"files"`
}

// build returns every file generate writes, the manifest last. The result
// depends only on the canonical model and the generator version.
func build(l loaded, outPath string) ([]file, error) {
	header := fmt.Sprintf("// Code generated by %s %s from model %s. DO NOT EDIT.",
		generatorID, generatorVersion, l.canonicalSHA256())
	var src strings.Builder
	if fault() != "no-header" {
		src.WriteString(header + "\n\n")
	}
	if fault() == "nondeterministic" {
		src.WriteString("// Generated into " + outPath + "\n\n")
	}
	fmt.Fprintf(&src, "package reference\n\nconst Name = %s\n\nvar Items = []string{\n", strconv.Quote(l.model.Name))
	for _, it := range l.model.Items {
		fmt.Fprintf(&src, "\t%s,\n", strconv.Quote(it))
	}
	src.WriteString("}\n")
	goSrc, err := format.Source([]byte(src.String()))
	if err != nil {
		return nil, fmt.Errorf("internal error formatting names.go: %w", err)
	}
	files := []file{
		{"data/items.txt", []byte(strings.Join(l.model.Items, "\n") + "\n")},
		{"names.go", goSrc},
	}
	sort.Slice(files, func(i, j int) bool { return files[i].path < files[j].path })

	var m manifest
	m.ManifestVersion = manifestVersion
	m.Profile = profile
	m.Generator.ID = generatorID
	m.Generator.Version = generatorVersion
	m.Generator.Toolchain = toolchain
	m.Model.CanonicalSHA256 = l.canonicalSHA256()
	for _, f := range files {
		m.Files = append(m.Files, manifestFile{Path: f.path, SHA256: sum(f.data), Bytes: len(f.data)})
	}
	switch fault() {
	case "manifest-unsorted":
		m.Files[0], m.Files[1] = m.Files[1], m.Files[0]
	case "manifest-missing-file":
		m.Files = m.Files[1:]
	}
	mb, err := encodeJSON(m, fault() != "manifest-compact")
	if err != nil {
		return nil, err
	}
	return append(files, file{manifestName, mb}), nil
}

func generate(l loaded, out string) error {
	files, err := build(l, out)
	if err != nil {
		return err
	}
	if fault() != "overwrite-out" {
		if _, err := os.Lstat(out); err == nil {
			return refuse("%s already exists", out)
		} else if !os.IsNotExist(err) {
			return err
		}
	}
	if err := os.Mkdir(out, 0o755); err != nil {
		if os.IsNotExist(err) {
			return refuse("parent of %s does not exist", out)
		}
		if os.IsExist(err) {
			if fault() != "overwrite-out" {
				return refuse("%s already exists", out)
			}
		} else {
			return err
		}
	}
	for _, f := range files {
		if err := writeFile(out, f); err != nil {
			return fmt.Errorf("output %s is partial: %w", out, err)
		}
	}
	return nil
}

func writeFile(out string, f file) error {
	p := filepath.Join(out, filepath.FromSlash(f.path))
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	return os.WriteFile(p, f.data, 0o644)
}

// check compares OUT with what generate would write and returns one drift line
// per differing path. It writes nothing.
func check(l loaded, out string) ([]string, error) {
	files, err := build(l, out)
	if err != nil {
		return nil, err
	}
	if fault() == "check-always-ok" {
		return nil, nil
	}
	fi, err := os.Lstat(out)
	if os.IsNotExist(err) {
		return []string{"drift: missing ."}, nil
	}
	if err != nil {
		return nil, err
	}
	if !fi.IsDir() { // includes a symlink to a directory
		return []string{"drift: unexpected ."}, nil
	}

	expected := map[string]file{}
	dirs := map[string]bool{"": true}
	children := map[string]map[string]bool{}
	for _, f := range files {
		expected[f.path] = f
		parts := strings.Split(f.path, "/")
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

	type diff struct{ kind, path string }
	var diffs []diff
	add := func(kind, path string) {
		if kind == "unexpected" && fault() == "ignore-extra" {
			return
		}
		diffs = append(diffs, diff{kind, path})
	}
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
				p = rel + "/" + e.Name()
			}
			switch {
			case expected[p].path != "":
				if !e.Type().IsRegular() {
					add("unexpected", p)
					continue
				}
				got, err := os.ReadFile(filepath.Join(out, filepath.FromSlash(p)))
				if err != nil {
					return err
				}
				if !bytes.Equal(got, expected[p].data) {
					add("modified", p)
					if fault() == "check-repairs" {
						if err := writeFile(out, expected[p]); err != nil {
							return err
						}
					}
				}
			case dirs[p]:
				if !e.IsDir() {
					add("unexpected", p)
					continue
				}
				if err := walk(p); err != nil {
					return err
				}
			default:
				add("unexpected", p)
			}
		}
		for name := range children[rel] {
			if seen[name] {
				continue
			}
			p := name
			if rel != "" {
				p = rel + "/" + name
			}
			add("missing", p)
			if f, ok := expected[p]; ok && fault() == "check-repairs" {
				if err := writeFile(out, f); err != nil {
					return err
				}
			}
		}
		return nil
	}
	if err := walk(""); err != nil {
		return nil, err
	}
	sort.Slice(diffs, func(i, j int) bool {
		if diffs[i].path != diffs[j].path {
			return diffs[i].path < diffs[j].path
		}
		return diffs[i].kind < diffs[j].kind
	})
	var lines []string
	for _, d := range diffs {
		lines = append(lines, "drift: "+d.kind+" "+d.path)
	}
	return lines, nil
}
