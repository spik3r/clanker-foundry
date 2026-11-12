package contract

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"unicode/utf8"
)

// MaxDepth is the deepest JSON nesting a model may have.
const MaxDepth = 16

// Node describes one object or array shape in a strict schema.
type Node struct {
	// Fields maps each allowed member of an object to the name of its shape.
	// An empty name means the member is a scalar.
	Fields map[string]string
	// Array marks an array whose elements have the shape named Elem. An empty
	// Elem means the elements are scalars.
	Array bool
	Elem  string
}

// Schema maps shape names to shapes. Scanning starts at the shape named "model".
type Schema map[string]Node

// StrictScan refuses what encoding/json would accept silently: invalid UTF-8,
// input over max bytes, nesting over MaxDepth, null, duplicate or unknown
// object members (case variants included), objects or arrays where the schema
// expects a scalar, and trailing data. It checks structure only; the caller
// still decodes and validates values.
func StrictScan(raw []byte, max int, s Schema) error {
	if len(raw) == 0 || len(raw) > max {
		return fmt.Errorf("model must be 1..%d bytes", max)
	}
	if !utf8.Valid(raw) {
		return fmt.Errorf("model is not valid UTF-8")
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	if err := scan(d, 0, "model", s); err != nil {
		return err
	}
	if _, err := d.Token(); err != io.EOF {
		return fmt.Errorf("model has trailing data")
	}
	return nil
}

func scan(d *json.Decoder, depth int, shape string, s Schema) error {
	if depth > MaxDepth {
		return fmt.Errorf("JSON nesting exceeds %d", MaxDepth)
	}
	t, err := d.Token()
	if err != nil {
		return err
	}
	if t == nil {
		return fmt.Errorf("null is not supported")
	}
	delim, ok := t.(json.Delim)
	if !ok {
		return nil
	}
	node := s[shape]
	switch delim {
	case '{':
		if node.Fields == nil {
			return fmt.Errorf("unexpected object")
		}
		seen := map[string]bool{}
		for d.More() {
			key, err := d.Token()
			if err != nil {
				return err
			}
			k, ok := key.(string)
			if !ok {
				return fmt.Errorf("expected object key")
			}
			child, allowed := node.Fields[k]
			if !allowed {
				return fmt.Errorf("unknown field %q", k)
			}
			if seen[k] {
				return fmt.Errorf("duplicate key %q", k)
			}
			seen[k] = true
			if err := scan(d, depth+1, child, s); err != nil {
				return fmt.Errorf("%s: %w", k, err)
			}
		}
	case '[':
		if !node.Array {
			return fmt.Errorf("unexpected array")
		}
		for d.More() {
			if err := scan(d, depth+1, node.Elem, s); err != nil {
				return err
			}
		}
	default:
		return fmt.Errorf("unexpected delimiter")
	}
	_, err = d.Token()
	return err
}
