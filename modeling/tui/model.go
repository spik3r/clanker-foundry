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
	// Emit and DetailText are optional. Their zero values are the defaults and
	// are omitted from the canonical encoding.
	Emit       *Emit  `json:"emit,omitempty"`
	DetailText string `json:"detailText,omitempty"`
	// Embed and DetailSpacing were added in generator 1.3.0 and follow the same
	// rule: the default is the same as absence.
	Embed         string `json:"embed,omitempty"`
	DetailSpacing string `json:"detailSpacing,omitempty"`
}

// Emit selects what the compiler writes: a standalone module (default) or only
// the files of an importable package with the given name.
type Emit struct {
	Mode    string `json:"mode"`
	Package string `json:"package,omitempty"`
}

// Emit modes and detail text modes.
const (
	ModeModule  = "module"
	ModePackage = "package"
	TextClip    = "clip"
	TextWrap    = "wrap"
	// EmbedStandalone is the default: the browser owns the whole terminal.
	// EmbedHosted draws only a body for a host's page and never quits.
	EmbedStandalone = "standalone"
	EmbedHosted     = "hosted"
	// SpacingCompact is the default: detail fields on consecutive rows.
	// SpacingSpaced puts one blank row between consecutive fields.
	SpacingCompact = "compact"
	SpacingSpaced  = "spaced"
	// DefaultPackage is the package name of module-mode output.
	DefaultPackage = "recordbrowser"
)

// PackageName is the name in the package clause of every emitted Go file.
func (m Model) PackageName() string {
	if m.Emit != nil && m.Emit.Mode == ModePackage {
		return m.Emit.Package
	}
	return DefaultPackage
}

// PackageMode reports whether only an importable package is emitted.
func (m Model) PackageMode() bool { return m.Emit != nil && m.Emit.Mode == ModePackage }

// Wrap reports whether the detail pane wraps values instead of clipping them.
func (m Model) Wrap() bool { return m.DetailText == TextWrap }

// Hosted reports whether the browser is drawn as a body inside a host's page.
func (m Model) Hosted() bool { return m.Embed == EmbedHosted }

// Spaced reports whether one blank row separates consecutive detail fields.
func (m Model) Spaced() bool { return m.DetailSpacing == SpacingSpaced }

// canonical returns m with explicit defaults replaced by absence. It does not
// modify the pointed-to Emit value.
func (m Model) canonical() Model {
	if m.Emit != nil && m.Emit.Mode == ModeModule {
		m.Emit = nil
	}
	if m.DetailText == TextClip {
		m.DetailText = ""
	}
	if m.Embed == EmbedStandalone {
		m.Embed = ""
	}
	if m.DetailSpacing == SpacingCompact {
		m.DetailSpacing = ""
	}
	return m
}

var packageName = regexp.MustCompile(`^[a-z][a-z0-9]{0,31}$`)

// reservedPackageNames holds Go keywords, predeclared identifiers and names
// that cannot be imported as a package (main).
var reservedPackageNames = map[string]bool{}

func init() {
	for _, w := range strings.Fields(`break default func interface select case defer go map struct chan else goto
package switch const fallthrough if range type continue for import return var
any bool byte comparable complex64 complex128 error float32 float64 int int8 int16
int32 int64 rune string uint uint8 uint16 uint32 uint64 uintptr true false iota nil
append cap clear close complex copy delete imag len make max min new panic print
println real recover main init`) {
		reservedPackageNames[w] = true
	}
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
	// Present-but-empty optional members are refused, not read as defaults.
	var probe struct {
		DetailText    *string `json:"detailText"`
		Embed         *string `json:"embed"`
		DetailSpacing *string `json:"detailSpacing"`
		Emit          *struct {
			Package *string `json:"package"`
		} `json:"emit"`
	}
	if err := json.Unmarshal(raw, &probe); err != nil {
		return Model{}, err
	}
	if probe.DetailText != nil && m.DetailText == "" {
		return Model{}, fmt.Errorf("detailText must be %q or %q", TextClip, TextWrap)
	}
	if probe.Embed != nil && m.Embed == "" {
		return Model{}, fmt.Errorf("embed must be %q or %q", EmbedStandalone, EmbedHosted)
	}
	if probe.DetailSpacing != nil && m.DetailSpacing == "" {
		return Model{}, fmt.Errorf("detailSpacing must be %q or %q", SpacingCompact, SpacingSpaced)
	}
	if probe.Emit != nil && probe.Emit.Package != nil && m.Emit.Mode != ModePackage {
		return Model{}, fmt.Errorf("emit.package is only allowed with mode %q", ModePackage)
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
			"model":    "|formatVersion|profileVersion|id|title|fields|identity|summary|details|listLabel|detailLabel|layout|emit|detailText|embed|detailSpacing|",
			"emit":     "|mode|package|",
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
	if m.DetailText != "" && m.DetailText != TextClip && m.DetailText != TextWrap {
		return fmt.Errorf("detailText must be %q or %q", TextClip, TextWrap)
	}
	if m.Embed != "" && m.Embed != EmbedStandalone && m.Embed != EmbedHosted {
		return fmt.Errorf("embed must be %q or %q", EmbedStandalone, EmbedHosted)
	}
	if m.DetailSpacing != "" && m.DetailSpacing != SpacingCompact && m.DetailSpacing != SpacingSpaced {
		return fmt.Errorf("detailSpacing must be %q or %q", SpacingCompact, SpacingSpaced)
	}
	if e := m.Emit; e != nil {
		switch e.Mode {
		case ModeModule:
			if e.Package != "" {
				return fmt.Errorf("emit.package is only allowed with mode %q", ModePackage)
			}
		case ModePackage:
			if !packageName.MatchString(e.Package) || reservedPackageNames[e.Package] {
				return fmt.Errorf("emit.package must be a lowercase Go package name of 1..32 letters and digits, not a keyword, predeclared identifier or main")
			}
		default:
			return fmt.Errorf("emit.mode must be %q or %q", ModeModule, ModePackage)
		}
	}
	l := m.Layout
	if l.NarrowWidth < 24 || l.NarrowWidth > 240 || l.NarrowHeight < 6 || l.NarrowHeight > 80 || l.WideWidth < l.NarrowWidth || l.WideWidth > 400 || l.WideHeight < l.NarrowHeight || l.WideHeight > 120 || l.ListWidth < 8 || l.ListWidth > l.WideWidth-9 {
		return fmt.Errorf("layout must have nested bounded breakpoints and at least 8 columns per wide pane")
	}
	return nil
}
