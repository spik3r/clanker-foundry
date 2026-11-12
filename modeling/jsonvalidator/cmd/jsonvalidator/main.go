// jsonvalidator generates standalone Go and TypeScript boolean validators.
// It uses no third-party dependencies or execution plugins.
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
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"
)

const (
	generatorID      = "jsonvalidator"
	generatorVersion = "1.0.0"
	// toolchain names the toolchain this generator was qualified with. It is a
	// constant, not detected at run time.
	toolchain = "go1.27.3"

	contract        = "foundry-generator/v1"
	manifestVersion = "foundry-generator-manifest/v1"
	manifestName    = "foundry-manifest.json"

	exitFailure = 1
	exitRefused = 2
	exitDrift   = 3
)

const profile = "json-boolean/v1"
const maxModelBytes = 1 << 20
const maxJSONDepth = 64
const maxSchemaDepth = 32
const maxNodes = 512
const maxProperties = 128

var propertyName = regexp.MustCompile(`^[A-Za-z_$][A-Za-z0-9_$]{0,127}$`)

type property struct {
	name     string
	required bool
	schema   *node
}
type node struct {
	kind       string
	literal    *bool
	properties []property
	items      *node
	id         int
}

// Model keys must be unambiguous. This strictness is not applied to consumer
// input, whose JSON parsing remains outside generated code.
func checkJSON(data []byte) error {
	if len(data) == 0 || len(data) > maxModelBytes || !utf8.Valid(data) {
		return errors.New("invalid model bytes")
	}
	d := json.NewDecoder(bytes.NewReader(data))
	d.UseNumber()
	var walk func(int) error
	walk = func(depth int) error {
		if depth > maxJSONDepth {
			return errors.New("model JSON depth exceeded")
		}
		t, err := d.Token()
		if err != nil {
			return err
		}
		mark, container := t.(json.Delim)
		if !container {
			return nil
		}
		switch mark {
		case '{':
			seen := map[string]bool{}
			for d.More() {
				key, err := d.Token()
				if err != nil {
					return err
				}
				name, ok := key.(string)
				if !ok || seen[name] {
					return errors.New("duplicate or invalid model key")
				}
				seen[name] = true
				if err := walk(depth + 1); err != nil {
					return err
				}
			}
		case '[':
			for d.More() {
				if err := walk(depth + 1); err != nil {
					return err
				}
			}
		default:
			return errors.New("invalid JSON delimiter")
		}
		_, err = d.Token()
		return err
	}
	if err := walk(1); err != nil {
		return err
	}
	if _, err := d.Token(); err != io.EOF {
		return errors.New("trailing model content")
	}
	return nil
}

func object(raw []byte, required []string, optional ...string) (map[string]json.RawMessage, error) {
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(raw, &obj); err != nil || obj == nil {
		return nil, errors.New("expected model object")
	}
	allowed := map[string]bool{}
	for _, key := range required {
		allowed[key] = true
		if _, ok := obj[key]; !ok {
			return nil, errors.New("missing model field")
		}
	}
	for _, key := range optional {
		allowed[key] = true
	}
	for key := range obj {
		if !allowed[key] {
			return nil, errors.New("unknown model field")
		}
	}
	return obj, nil
}

func boolValue(raw []byte) (bool, error) {
	switch string(bytes.TrimSpace(raw)) {
	case "true":
		return true, nil
	case "false":
		return false, nil
	default:
		return false, errors.New("expected model boolean")
	}
}

func parseNode(raw []byte, depth int, count *int) (*node, error) {
	*count++
	if depth > maxSchemaDepth || *count > maxNodes {
		return nil, errors.New("model schema limit exceeded")
	}
	var header map[string]json.RawMessage
	if err := json.Unmarshal(raw, &header); err != nil || header == nil {
		return nil, errors.New("expected schema object")
	}
	var kind string
	if err := json.Unmarshal(header["type"], &kind); err != nil || kind == "" {
		return nil, errors.New("missing or invalid schema type")
	}
	n := &node{kind: kind}
	switch kind {
	case "string":
		_, err := object(raw, []string{"type"})
		return n, err
	case "boolean":
		obj, err := object(raw, []string{"type"}, "const")
		if err != nil {
			return nil, err
		}
		if value, ok := obj["const"]; ok {
			literal, err := boolValue(value)
			if err != nil {
				return nil, err
			}
			n.literal = &literal
		}
	case "array":
		obj, err := object(raw, []string{"type", "items"})
		if err != nil {
			return nil, err
		}
		n.items, err = parseNode(obj["items"], depth+1, count)
		if err != nil {
			return nil, err
		}
	case "object":
		obj, err := object(raw, []string{"type", "additionalProperties", "properties"})
		if err != nil {
			return nil, err
		}
		allow, err := boolValue(obj["additionalProperties"])
		if err != nil || !allow {
			return nil, errors.New("only permissive extra properties are supported")
		}
		var properties []json.RawMessage
		if err := json.Unmarshal(obj["properties"], &properties); err != nil || properties == nil || len(properties) > maxProperties {
			return nil, errors.New("invalid model properties")
		}
		seen := map[string]bool{}
		for _, rawProperty := range properties {
			fields, err := object(rawProperty, []string{"name", "required", "schema"})
			if err != nil {
				return nil, err
			}
			var p property
			if err := json.Unmarshal(fields["name"], &p.name); err != nil || !propertyName.MatchString(p.name) || seen[p.name] {
				return nil, errors.New("invalid or duplicate property name")
			}
			seen[p.name] = true
			p.required, err = boolValue(fields["required"])
			if err != nil {
				return nil, err
			}
			p.schema, err = parseNode(fields["schema"], depth+1, count)
			if err != nil {
				return nil, err
			}
			n.properties = append(n.properties, p)
		}
		// Property order is not significant: the predicate is a conjunction.
		// Canonical order is byte order of the name.
		sort.Slice(n.properties, func(i, j int) bool { return n.properties[i].name < n.properties[j].name })
	default:
		return nil, errors.New("unsupported schema type")
	}
	return n, nil
}

// canonical encodes the node as compact JSON in the fixed member order
// documented in PROFILE.md. Property names are ASCII identifiers, so
// strconv.Quote yields valid JSON.
func (n *node) canonical(b *strings.Builder) {
	fmt.Fprintf(b, `{"type":%q`, n.kind)
	switch n.kind {
	case "boolean":
		if n.literal != nil {
			fmt.Fprintf(b, `,"const":%t`, *n.literal)
		}
	case "array":
		b.WriteString(`,"items":`)
		n.items.canonical(b)
	case "object":
		b.WriteString(`,"additionalProperties":true,"properties":[`)
		for i, p := range n.properties {
			if i > 0 {
				b.WriteByte(',')
			}
			fmt.Fprintf(b, `{"name":%s,"required":%t,"schema":`, strconv.Quote(p.name), p.required)
			p.schema.canonical(b)
			b.WriteByte('}')
		}
		b.WriteString("]")
	}
	b.WriteByte('}')
}

// loaded is a validated model: the source bytes, the parsed tree and the
// canonical encoding.
type loaded struct {
	source    []byte
	root      *node
	canonical []byte
}

func sum(b []byte) string {
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}

func (l loaded) canonicalSHA256() string { return sum(l.canonical) }

func (l loaded) identityLine() string {
	return fmt.Sprintf(`{"contract":%q,"profile":%q,"sourceSHA256":%q,"canonicalSHA256":%q}`+"\n",
		contract, profile, sum(l.source), l.canonicalSHA256())
}

func parseModel(data []byte) (loaded, error) {
	if err := checkJSON(data); err != nil {
		return loaded{}, err
	}
	obj, err := object(data, []string{"profile", "root"})
	if err != nil {
		return loaded{}, err
	}
	var version string
	if err := json.Unmarshal(obj["profile"], &version); err != nil || version != profile {
		return loaded{}, errors.New("unsupported model profile")
	}
	count := 0
	root, err := parseNode(obj["root"], 1, &count)
	if err != nil {
		return loaded{}, err
	}
	var b strings.Builder
	fmt.Fprintf(&b, `{"profile":%q,"root":`, profile)
	root.canonical(&b)
	b.WriteByte('}')
	return loaded{source: data, root: root, canonical: []byte(b.String())}, nil
}

// refusal is an error that maps to exit code 2: nothing was written.
type refusal struct{ msg string }

func (r refusal) Error() string { return r.msg }

func refuse(format string, a ...any) error {
	return refusal{msg: fmt.Sprintf(format, a...)}
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
	l, err := parseModel(data)
	if err != nil {
		return loaded{}, refuse("invalid model: %v", err)
	}
	return l, nil
}

// emit returns the Go and TypeScript sources. They depend only on the
// canonical model: properties are already in canonical order.
func (l loaded) emit() ([]byte, []byte, error) {
	var nodes []*node
	var visit func(*node)
	visit = func(n *node) {
		n.id = len(nodes)
		nodes = append(nodes, n)
		for _, p := range n.properties {
			visit(p.schema)
		}
		if n.items != nil {
			visit(n.items)
		}
	}
	visit(l.root)
	digest := l.canonicalSHA256()
	header := fmt.Sprintf("// Code generated by %s %s from model %s. DO NOT EDIT.\n", generatorID, generatorVersion, digest)
	var goSource, tsSource strings.Builder
	fmt.Fprintf(&goSource, "%s\npackage validator\n\nconst ModelSHA256 = %q\n\nfunc Validate(value any) bool { return n0(value) }\n", header, digest)
	fmt.Fprintf(&tsSource, "%s\nexport const MODEL_SHA256 = %q;\n\nexport function validate(value: unknown): boolean { return n0(value); }\n", header, digest)
	for _, n := range nodes {
		fmt.Fprintf(&goSource, "\nfunc n%d(value any) bool {\n", n.id)
		fmt.Fprintf(&tsSource, "\nfunction n%d(value: unknown): boolean {\n", n.id)
		switch n.kind {
		case "string":
			goSource.WriteString("_, ok := value.(string); return ok\n")
			tsSource.WriteString("  return typeof value === 'string';\n")
		case "boolean":
			if n.literal == nil {
				goSource.WriteString("_, ok := value.(bool); return ok\n")
				tsSource.WriteString("  return typeof value === 'boolean';\n")
			} else {
				fmt.Fprintf(&goSource, "v, ok := value.(bool); return ok && v == %t\n", *n.literal)
				fmt.Fprintf(&tsSource, "  return value === %t;\n", *n.literal)
			}
		case "array":
			fmt.Fprintf(&goSource, "items, ok := value.([]any); if !ok || items == nil { return false }; for _, item := range items { if !n%d(item) { return false } }; return true\n", n.items.id)
			fmt.Fprintf(&tsSource, "  return Array.isArray(value) && value.every(n%d);\n", n.items.id)
		case "object":
			goSource.WriteString("obj, ok := value.(map[string]any); if !ok || obj == nil { return false }\n")
			tsSource.WriteString("  if (value === null || typeof value !== 'object' || Array.isArray(value)) return false;\n")
			if len(n.properties) > 0 {
				tsSource.WriteString("  const obj = value as Record<string, unknown>;\n")
			}
			for _, p := range n.properties {
				key, _ := json.Marshal(p.name)
				if p.required {
					fmt.Fprintf(&goSource, "if !n%d(obj[%s]) { return false }\n", p.schema.id, strconv.Quote(p.name))
					fmt.Fprintf(&tsSource, "  if (!Object.prototype.hasOwnProperty.call(obj, %s) || !n%d(obj[%s])) return false;\n", key, p.schema.id, key)
				} else {
					fmt.Fprintf(&goSource, "if child, exists := obj[%s]; exists && !n%d(child) { return false }\n", strconv.Quote(p.name), p.schema.id)
					fmt.Fprintf(&tsSource, "  if (Object.prototype.hasOwnProperty.call(obj, %s) && !n%d(obj[%s])) return false;\n", key, p.schema.id, key)
				}
			}
			goSource.WriteString("return true\n")
			tsSource.WriteString("  return true;\n")
		}
		goSource.WriteString("}\n")
		tsSource.WriteString("}\n")
	}
	formatted, err := format.Source([]byte(goSource.String()))
	if err != nil {
		return nil, nil, err
	}
	return formatted, []byte(tsSource.String()), nil
}

// compile returns the emitted sources for a model.
func compile(data []byte) ([]byte, []byte, error) {
	l, err := parseModel(data)
	if err != nil {
		return nil, nil, err
	}
	return l.emit()
}

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

// build returns every file generate writes, the manifest last.
func build(l loaded) ([]file, error) {
	goSource, tsSource, err := l.emit()
	if err != nil {
		return nil, fmt.Errorf("internal error: %w", err)
	}
	files := []file{{"validator.go", goSource}, {"validator.ts", tsSource}}
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
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(m); err != nil {
		return nil, err
	}
	return append(files, file{manifestName, buf.Bytes()}), nil
}

// generate writes a new output directory. Existing destinations, including
// symlinks, are refused. A later filesystem failure can leave a partial
// directory; there is no rollback.
func generate(modelPath, out string) error {
	l, err := loadModel(modelPath)
	if err != nil {
		return err
	}
	return write(l, out)
}

func write(l loaded, out string) error {
	files, err := build(l)
	if err != nil {
		return err
	}
	if _, err := os.Lstat(out); err == nil {
		return refuse("%s already exists", out)
	} else if !os.IsNotExist(err) {
		return err
	}
	if err := os.Mkdir(out, 0o755); err != nil {
		if os.IsNotExist(err) {
			return refuse("parent of %s does not exist", out)
		}
		if os.IsExist(err) {
			return refuse("%s already exists", out)
		}
		return err
	}
	for _, f := range files {
		if err := os.WriteFile(filepath.Join(out, f.path), f.data, 0o644); err != nil {
			return fmt.Errorf("output %s is partial: %w", out, err)
		}
	}
	return nil
}

// check compares out with what generate would write and returns one drift line
// per differing path. It writes nothing.
func check(l loaded, out string) ([]string, error) {
	files, err := build(l)
	if err != nil {
		return nil, err
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
	entries, err := os.ReadDir(out)
	if err != nil {
		return nil, err
	}
	present := map[string]os.DirEntry{}
	for _, e := range entries {
		present[e.Name()] = e
	}
	expected := map[string]bool{}
	var lines []string
	for _, f := range files {
		expected[f.path] = true
		e, ok := present[f.path]
		switch {
		case !ok:
			lines = append(lines, "drift: missing "+f.path)
		case !e.Type().IsRegular():
			lines = append(lines, "drift: unexpected "+f.path)
		default:
			got, err := os.ReadFile(filepath.Join(out, f.path))
			if err != nil {
				return nil, err
			}
			if !bytes.Equal(got, f.data) {
				lines = append(lines, "drift: modified "+f.path)
			}
		}
	}
	for name := range present {
		if !expected[name] {
			lines = append(lines, "drift: unexpected "+name)
		}
	}
	sort.Strings(lines)
	return lines, nil
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

// run executes one command line and returns the exit code. On success stdout
// is exactly the identity line; on any failure stdout is empty.
func run(args []string, stdout, stderr io.Writer) int {
	line, drift, err := dispatch(args)
	switch {
	case err != nil:
		code := exitFailure
		var r refusal
		if errors.As(err, &r) {
			code = exitRefused
		}
		fmt.Fprintf(stderr, "jsonvalidator: %v\n", err)
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

const usage = "usage: jsonvalidator validate MODEL | generate MODEL OUT | check MODEL OUT"

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
		if err := write(l, args[2]); err != nil {
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
	return "", nil, refuse("%s", usage)
}
