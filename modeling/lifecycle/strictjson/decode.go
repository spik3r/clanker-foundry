// Package strictjson supplies bounded, lossless JSON decoding, not model policy.
package strictjson

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"reflect"
	"strconv"
	"strings"
	"unicode/utf8"
)

func bad() error { return errors.New("invalid strict JSON") }
func unicodeEscapes(b []byte) error {
	for i := 0; i < len(b); i++ {
		if b[i] != '"' {
			continue
		}
		i++
		for i < len(b) && b[i] != '"' {
			if b[i] != '\\' {
				i++
				continue
			}
			i++
			if b[i] != 'u' {
				i++
				continue
			}
			n, e := strconv.ParseUint(string(b[i+1:i+5]), 16, 16)
			if e != nil {
				return bad()
			}
			i += 5
			if n >= 0xd800 && n <= 0xdbff {
				if i+6 > len(b) || b[i] != '\\' || b[i+1] != 'u' {
					return bad()
				}
				low, e := strconv.ParseUint(string(b[i+2:i+6]), 16, 16)
				if e != nil || low < 0xdc00 || low > 0xdfff {
					return bad()
				}
				i += 6
			} else if n >= 0xdc00 && n <= 0xdfff {
				return bad()
			}
		}
	}
	return nil
}
func scan(d *json.Decoder, depth, maxDepth int, n *int) error {
	*n++
	if depth > maxDepth || *n > 100000 {
		return bad()
	}
	tok, e := d.Token()
	if e != nil {
		return bad()
	}
	delim, ok := tok.(json.Delim)
	if !ok {
		return nil
	}
	switch delim {
	case '{':
		keys := map[string]bool{}
		for d.More() {
			t, e := d.Token()
			k, ok := t.(string)
			if e != nil || !ok || keys[k] {
				return bad()
			}
			keys[k] = true
			if e = scan(d, depth+1, maxDepth, n); e != nil {
				return e
			}
		}
		t, e := d.Token()
		if e != nil || t != json.Delim('}') {
			return bad()
		}
	case '[':
		for d.More() {
			if e := scan(d, depth+1, maxDepth, n); e != nil {
				return e
			}
		}
		t, e := d.Token()
		if e != nil || t != json.Delim(']') {
			return bad()
		}
	default:
		return bad()
	}
	return nil
}

var rawType = reflect.TypeOf(json.RawMessage{})

func exact(v any, t reflect.Type) error {
	if t == rawType {
		return nil
	}
	if t.Kind() == reflect.Pointer {
		if v == nil {
			return nil
		}
		return exact(v, t.Elem())
	}
	if v == nil {
		return bad()
	}
	switch t.Kind() {
	case reflect.Struct:
		obj, ok := v.(map[string]any)
		if !ok {
			return bad()
		}
		allowed := map[string]reflect.Type{}
		required := map[string]bool{}
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			tag := strings.Split(f.Tag.Get("json"), ",")
			name := tag[0]
			if name == "-" {
				continue
			}
			if name == "" {
				name = f.Name
			}
			allowed[name] = f.Type
			required[name] = true
			for _, o := range tag[1:] {
				if o == "omitempty" {
					required[name] = false
				}
			}
		}
		for k, x := range obj {
			ft, ok := allowed[k]
			if !ok {
				return bad()
			}
			if e := exact(x, ft); e != nil {
				return e
			}
		}
		for k, need := range required {
			if _, ok := obj[k]; need && !ok {
				return bad()
			}
		}
	case reflect.Slice, reflect.Array:
		a, ok := v.([]any)
		if !ok {
			return bad()
		}
		for _, x := range a {
			if e := exact(x, t.Elem()); e != nil {
				return e
			}
		}
	case reflect.Map:
		obj, ok := v.(map[string]any)
		if !ok || t.Key().Kind() != reflect.String {
			return bad()
		}
		for _, x := range obj {
			if e := exact(x, t.Elem()); e != nil {
				return e
			}
		}
	}
	return nil
}

// Decode rejects duplicates, invalid Unicode, unknown/case-folded struct keys,
// absent required fields, null non-pointer fields, trailing data and bounds.
func Decode(data []byte, out any, maxBytes, maxDepth int) error {
	if maxBytes < 1 || maxDepth < 1 || len(data) > maxBytes || !utf8.Valid(data) || !json.Valid(data) {
		return bad()
	}
	if e := unicodeEscapes(data); e != nil {
		return e
	}
	d := json.NewDecoder(bytes.NewReader(data))
	d.UseNumber()
	n := 0
	if e := scan(d, 0, maxDepth, &n); e != nil {
		return e
	}
	if _, e := d.Token(); e != io.EOF {
		return bad()
	}
	d = json.NewDecoder(bytes.NewReader(data))
	d.UseNumber()
	var raw any
	if e := d.Decode(&raw); e != nil {
		return bad()
	}
	t := reflect.TypeOf(out)
	if t == nil || t.Kind() != reflect.Pointer || reflect.ValueOf(out).IsNil() {
		return bad()
	}
	if e := exact(raw, t.Elem()); e != nil {
		return e
	}
	d = json.NewDecoder(bytes.NewReader(data))
	d.UseNumber()
	d.DisallowUnknownFields()
	if e := d.Decode(out); e != nil {
		return bad()
	}
	return nil
}
