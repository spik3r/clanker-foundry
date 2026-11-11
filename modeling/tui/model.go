// Package tui compiles the bounded record-browser/v1 presentation profile.
package tui

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

const Format = "tui-json/v1"
const Profile = "record-browser/v1"
const MaxModelBytes = 32768

// Model contains presentation semantics only. Record values come from the host.
type Model struct {
	FormatVersion  string   `json:"formatVersion"`
	ProfileVersion string   `json:"profileVersion"`
	ID             string   `json:"id"`
	Title          string   `json:"title"`
	Fields         []Field  `json:"fields"`
	Identity       string   `json:"identity"`
	Summary        string   `json:"summary"`
	Details        []string `json:"details"`
	ListLabel      string   `json:"listLabel"`
	DetailLabel    string   `json:"detailLabel"`
	Layout         Layout   `json:"layout"`
}
type Field struct {
	ID    string `json:"id"`
	Label string `json:"label"`
}
type Layout struct {
	NarrowWidth  int `json:"narrowWidth"`
	NarrowHeight int `json:"narrowHeight"`
	WideWidth    int `json:"wideWidth"`
	WideHeight   int `json:"wideHeight"`
	ListWidth    int `json:"listWidth"`
}

var identifier = regexp.MustCompile(`^[A-Z][A-Za-z0-9]{0,47}$`)

// Decode rejects unknown fields, duplicate object keys, nulls and trailing data.
func Decode(raw []byte) (Model, error) {
	var m Model
	if len(raw) == 0 || len(raw) > MaxModelBytes || !utf8.Valid(raw) {
		return m, fmt.Errorf("model must be UTF-8 JSON of 1..%d bytes", MaxModelBytes)
	}
	scan := json.NewDecoder(bytes.NewReader(raw))
	if err := checkValue(scan, 0, "model"); err != nil {
		return m, err
	}
	if _, err := scan.Token(); err != io.EOF {
		return m, fmt.Errorf("model has trailing data")
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	if err := d.Decode(&m); err != nil {
		return m, err
	}
	if err := m.Validate(); err != nil {
		return Model{}, err
	}
	return m, nil
}
func checkValue(d *json.Decoder, depth int, context string) error {
	if depth > 16 {
		return fmt.Errorf("JSON nesting exceeds 16")
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
	switch delim {
	case '{':
		allowed := map[string]string{
			"model":    "|formatVersion|profileVersion|id|title|fields|identity|summary|details|listLabel|detailLabel|layout|",
			"fields[]": "|id|label|",
			"layout":   "|narrowWidth|narrowHeight|wideWidth|wideHeight|listWidth|",
		}[context]
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
			if !strings.Contains(allowed, "|"+k+"|") {
				return fmt.Errorf("unknown field %q in %s", k, context)
			}
			if seen[k] {
				return fmt.Errorf("duplicate key %q", k)
			}
			seen[k] = true
			if err := checkValue(d, depth+1, k); err != nil {
				return err
			}
		}
	case '[':
		for d.More() {
			if err := checkValue(d, depth+1, context+"[]"); err != nil {
				return err
			}
		}
	default:
		return fmt.Errorf("unexpected delimiter")
	}
	_, err = d.Token()
	return err
}
func plain(s string, name string) error {
	if len(s) == 0 || len(s) > 120 || !utf8.ValidString(s) {
		return fmt.Errorf("%s must be 1..120 UTF-8 bytes", name)
	}
	for _, r := range s {
		if unicode.IsControl(r) || r == '\u2028' || r == '\u2029' {
			return fmt.Errorf("%s contains a control character", name)
		}
	}
	return nil
}

// Validate is also applied to programmatically constructed models before emission.
func (m Model) Validate() error {
	if m.FormatVersion != Format || m.ProfileVersion != Profile {
		return fmt.Errorf("unsupported format or profile")
	}
	if !identifier.MatchString(m.ID) {
		return fmt.Errorf("id must be an exported ASCII identifier")
	}
	for _, pair := range []struct{ name, value string }{{"title", m.Title}, {"listLabel", m.ListLabel}, {"detailLabel", m.DetailLabel}} {
		if err := plain(pair.value, pair.name); err != nil {
			return err
		}
	}
	if len(m.Fields) < 1 || len(m.Fields) > 16 {
		return fmt.Errorf("fields must contain 1..16 fields")
	}
	fields := map[string]bool{}
	for _, f := range m.Fields {
		if !identifier.MatchString(f.ID) || fields[f.ID] {
			return fmt.Errorf("invalid or duplicate field %q", f.ID)
		}
		fields[f.ID] = true
		if err := plain(f.Label, "field label"); err != nil {
			return err
		}
	}
	if !fields[m.Identity] || !fields[m.Summary] {
		return fmt.Errorf("identity and summary must reference declared fields")
	}
	if len(m.Details) < 1 || len(m.Details) > 16 {
		return fmt.Errorf("details must contain 1..16 fields")
	}
	seen := map[string]bool{}
	for _, id := range m.Details {
		if !fields[id] || seen[id] {
			return fmt.Errorf("invalid or duplicate detail reference %q", id)
		}
		seen[id] = true
	}
	l := m.Layout
	if l.NarrowWidth < 24 || l.NarrowWidth > 240 || l.NarrowHeight < 6 || l.NarrowHeight > 80 || l.WideWidth < l.NarrowWidth || l.WideWidth > 400 || l.WideHeight < l.NarrowHeight || l.WideHeight > 120 || l.ListWidth < 8 || l.ListWidth > l.WideWidth-9 {
		return fmt.Errorf("layout must have nested bounded breakpoints and at least 8 columns per wide pane")
	}
	return nil
}
