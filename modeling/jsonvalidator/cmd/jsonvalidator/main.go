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
	"unicode"
	"unicode/utf8"
)

const (
	generatorID      = "jsonvalidator"
	generatorVersion = "1.1.0"
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
const maxStringBytes = 256
const maxEnumValues = 32
const maxLengthBound = 65535

var propertyName = regexp.MustCompile(`^[A-Za-z_$][A-Za-z0-9_$]{0,127}$`)

type property struct {
	name     string
	required bool
	schema   *node
}
type node struct {
	kind       string
	literal    *bool
	constant   *string
	enum       []string
	minLength  *int
	maxLength  *int
	closed     bool
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

// modelString reads a model string that becomes part of generated code. It
// refuses non-strings, control characters, strings over 256 bytes and U+FFFD:
// an unpaired surrogate escape decodes to U+FFFD, so U+FFFD is refused rather
// than let a lone surrogate silently change meaning.
func modelString(raw []byte) (string, error) {
	raw = bytes.TrimSpace(raw)
	var s string
	if len(raw) == 0 || raw[0] != '"' || json.Unmarshal(raw, &s) != nil {
		return "", errors.New("expected model string")
	}
	if len(s) > maxStringBytes {
		return "", errors.New("model string too long")
	}
	for _, r := range s {
		if unicode.IsControl(r) || r == utf8.RuneError {
			return "", errors.New("model string has a control or replacement character")
		}
	}
	return s, nil
}

var lengthLexeme = regexp.MustCompile(`^(0|[1-9][0-9]{0,4})$`)

func lengthBound(raw []byte) (*int, error) {
	text := string(bytes.TrimSpace(raw))
	if !lengthLexeme.MatchString(text) {
		return nil, errors.New("expected model length integer")
	}
	v, _ := strconv.Atoi(text)
	if v > maxLengthBound {
		return nil, errors.New("model length out of range")
	}
	return &v, nil
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
		obj, err := object(raw, []string{"type"}, "const", "enum", "minLength", "maxLength")
		if err != nil {
			return nil, err
		}
		_, hasConst := obj["const"]
		_, hasEnum := obj["enum"]
		_, hasMin := obj["minLength"]
		_, hasMax := obj["maxLength"]
		if hasConst && hasEnum || (hasConst || hasEnum) && (hasMin || hasMax) {
			return nil, errors.New("string const, enum and length bounds conflict")
		}
		if hasConst {
			c, err := modelString(obj["const"])
			if err != nil {
				return nil, err
			}
			n.constant = &c
		}
		if hasEnum {
			var values []json.RawMessage
			if err := json.Unmarshal(obj["enum"], &values); err != nil || len(values) < 1 || len(values) > maxEnumValues {
				return nil, errors.New("enum needs 1 to 32 strings")
			}
			seen := map[string]bool{}
			for _, rawValue := range values {
				v, err := modelString(rawValue)
				if err != nil {
					return nil, err
				}
				if seen[v] {
					return nil, errors.New("duplicate enum value")
				}
				seen[v] = true
				n.enum = append(n.enum, v)
			}
			sort.Strings(n.enum)
		}
		if hasMin {
			if n.minLength, err = lengthBound(obj["minLength"]); err != nil {
				return nil, err
			}
		}
		if hasMax {
			if n.maxLength, err = lengthBound(obj["maxLength"]); err != nil {
				return nil, err
			}
		}
		if n.minLength != nil && n.maxLength != nil && *n.minLength > *n.maxLength {
			return nil, errors.New("minLength exceeds maxLength")
		}
		return n, nil
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
		if err != nil {
			return nil, err
		}
		n.closed = !allow
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

// quote writes a model string for the canonical encoding: only the double
// quote and the backslash are escaped. Control characters cannot occur.
func quote(s string) string {
	return `"` + strings.NewReplacer(`\`, `\\`, `"`, `\"`).Replace(s) + `"`
}

// canonical encodes the node as compact JSON in the fixed member order
// documented in PROFILE.md. Property names are ASCII identifiers, so
// strconv.Quote yields valid JSON.
func (n *node) canonical(b *strings.Builder) {
	fmt.Fprintf(b, `{"type":%q`, n.kind)
	switch n.kind {
	case "string":
		switch {
		case n.constant != nil:
			b.WriteString(`,"const":` + quote(*n.constant))
		case n.enum != nil:
			b.WriteString(`,"enum":[`)
			for i, v := range n.enum {
				if i > 0 {
					b.WriteByte(',')
				}
				b.WriteString(quote(v))
			}
			b.WriteByte(']')
		}
		if n.minLength != nil {
			fmt.Fprintf(b, `,"minLength":%d`, *n.minLength)
		}
		if n.maxLength != nil {
			fmt.Fprintf(b, `,"maxLength":%d`, *n.maxLength)
		}
	case "boolean":
		if n.literal != nil {
			fmt.Fprintf(b, `,"const":%t`, *n.literal)
		}
	case "array":
		b.WriteString(`,"items":`)
		n.items.canonical(b)
	case "object":
		fmt.Fprintf(b, `,"additionalProperties":%t,"properties":[`, !n.closed)
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
			emitString(n, &goSource, &tsSource)
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
			if len(n.properties) > 0 || n.closed {
				tsSource.WriteString("  const obj = value as Record<string, unknown>;\n")
			}
			if n.closed {
				emitClosed(n, &goSource, &tsSource)
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

// emitString writes the predicate body of a string node. Lengths count code
// points: Go ranges over runes and TypeScript iterates by code point, so a
// surrogate pair is one and an unpaired surrogate is one in both languages
// (Go's JSON decoder turns an unpaired surrogate escape into U+FFFD).
func emitString(n *node, goSource, tsSource *strings.Builder) {
	switch {
	case n.constant != nil:
		fmt.Fprintf(goSource, "v, ok := value.(string); return ok && v == %s\n", strconv.Quote(*n.constant))
		fmt.Fprintf(tsSource, "  return value === %s;\n", jsString(*n.constant))
	case n.enum != nil:
		var g, t []string
		for _, v := range n.enum {
			g = append(g, "v == "+strconv.Quote(v))
			t = append(t, "value === "+jsString(v))
		}
		fmt.Fprintf(goSource, "v, ok := value.(string); return ok && (%s)\n", strings.Join(g, " || "))
		fmt.Fprintf(tsSource, "  return %s;\n", strings.Join(t, " || "))
	case n.minLength != nil || n.maxLength != nil:
		goSource.WriteString("v, ok := value.(string); if !ok { return false }; n := 0; for range v { n++ }\n")
		tsSource.WriteString("  if (typeof value !== 'string') return false;\n  let n = 0;\n  for (const _ of value) n++;\n")
		var g, t []string
		if n.minLength != nil {
			g = append(g, fmt.Sprintf("n >= %d", *n.minLength))
			t = append(t, fmt.Sprintf("n >= %d", *n.minLength))
		}
		if n.maxLength != nil {
			g = append(g, fmt.Sprintf("n <= %d", *n.maxLength))
			t = append(t, fmt.Sprintf("n <= %d", *n.maxLength))
		}
		fmt.Fprintf(goSource, "return %s\n", strings.Join(g, " && "))
		fmt.Fprintf(tsSource, "  return %s;\n", strings.Join(t, " && "))
	default:
		goSource.WriteString("_, ok := value.(string); return ok\n")
		tsSource.WriteString("  return typeof value === 'string';\n")
	}
}

func jsString(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}

// emitClosed refuses any own enumerable string key that is not declared. The
// switch compares keys by value, so inherited names never count as declared.
func emitClosed(n *node, goSource, tsSource *strings.Builder) {
	if len(n.properties) == 0 {
		goSource.WriteString("if len(obj) > 0 { return false }\n")
		tsSource.WriteString("  if (Object.keys(obj).length > 0) return false;\n")
		return
	}
	var g, t []string
	for _, p := range n.properties {
		g = append(g, strconv.Quote(p.name))
		t = append(t, "      case "+jsString(p.name)+":")
	}
	fmt.Fprintf(goSource, "for key := range obj { switch key { case %s: default: return false } }\n", strings.Join(g, ", "))
	fmt.Fprintf(tsSource, "  for (const key of Object.keys(obj)) {\n    switch (key) {\n%s\n        break;\n      default:\n        return false;\n    }\n  }\n", strings.Join(t, "\n"))
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

// trimOutputRoot removes only terminal separators and dot components. Interior
// components retain filesystem meaning, including .. after a parent symlink.
func trimOutputRoot(out string) string {
	root := len(filepath.VolumeName(out))
	for root < len(out) && os.IsPathSeparator(out[root]) {
		root++
	}
	for {
		for len(out) > root && os.IsPathSeparator(out[len(out)-1]) {
			out = out[:len(out)-1]
		}
		n := len(out)
		if n > root && n >= 2 && out[n-1] == '.' && os.IsPathSeparator(out[n-2]) {
			out = out[:n-1]
			continue
		}
		return out
	}
}

// check compares out with what generate would write and returns one drift line
// per differing path. It writes nothing.
func check(l loaded, out string) ([]string, error) {
	out = trimOutputRoot(out)
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
	// Resolve the accepted directory before Join can collapse interior parent components.
	out, err = filepath.EvalSymlinks(out)
	if err != nil {
		return nil, err
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
