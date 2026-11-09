// jsonvalidator generates standalone Go and TypeScript boolean validators.
// It uses no third-party dependencies or execution plugins.
package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"go/format"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"
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
	default:
		return nil, errors.New("unsupported schema type")
	}
	return n, nil
}

func compile(data []byte) ([]byte, []byte, error) {
	if err := checkJSON(data); err != nil {
		return nil, nil, err
	}
	obj, err := object(data, []string{"profile", "root"})
	if err != nil {
		return nil, nil, err
	}
	var version string
	if err := json.Unmarshal(obj["profile"], &version); err != nil || version != profile {
		return nil, nil, errors.New("unsupported model profile")
	}
	count := 0
	root, err := parseNode(obj["root"], 1, &count)
	if err != nil {
		return nil, nil, err
	}
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
	visit(root)
	digest := fmt.Sprintf("%x", sha256.Sum256(data))
	var goSource, tsSource strings.Builder
	fmt.Fprintf(&goSource, "// Code generated by jsonvalidator. DO NOT EDIT.\npackage validator\n\nconst ModelSHA256 = %q\n\nfunc Validate(value any) bool { return n0(value) }\n", digest)
	fmt.Fprintf(&tsSource, "// Code generated by jsonvalidator. DO NOT EDIT.\nexport const MODEL_SHA256 = %q;\n\nexport function validate(value: unknown): boolean { return n0(value); }\n", digest)
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

func generate(modelPath, output string) error {
	f, err := os.Open(modelPath)
	if err != nil {
		return err
	}
	defer f.Close()
	data, err := io.ReadAll(io.LimitReader(f, maxModelBytes+1))
	if err != nil {
		return err
	}
	goSource, tsSource, err := compile(data)
	if err != nil {
		return err
	}
	// Refuse existing destinations, including symlinks. No overwrite or rollback
	// promise: a later filesystem failure can leave a newly created partial output.
	if err := os.Mkdir(output, 0700); err != nil {
		return err
	}
	for _, item := range []struct {
		name string
		data []byte
	}{{"validator.go", goSource}, {"validator.ts", tsSource}} {
		if err := os.WriteFile(filepath.Join(output, item.name), item.data, 0600); err != nil {
			return err
		}
	}
	return nil
}

func main() {
	modelPath := flag.String("model", "", "path to the JSON boolean model")
	output := flag.String("out", "", "new, absent output directory")
	flag.Parse()
	if *modelPath == "" || *output == "" || flag.NArg() != 0 {
		fmt.Fprintln(os.Stderr, "use -model MODEL -out NEW_DIRECTORY")
		os.Exit(2)
	}
	if err := generate(*modelPath, *output); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
