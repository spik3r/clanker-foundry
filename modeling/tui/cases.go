package tui

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Expectation cases for record-browser/v1. See CASES.md.
const (
	CasesFormat   = "record-browser-cases/v1"
	MaxCasesBytes = 256 << 10
)

// Limits of the cases format.
const (
	maxRecordSets   = 8
	maxSetRecords   = 200
	maxCases        = 200
	maxSteps        = 64
	maxExpectTexts  = 16
	maxKeyHints     = 12
	maxValueBytes   = 4096
	maxRequirement  = 240
	maxExpectedText = 240
)

// AllowedKeys are the keys a step may press, spelled as the README documents
// them. Undocumented aliases of the fixed bindings are not allowed.
var AllowedKeys = []string{
	"tab", "shift+tab", "enter", "esc", "?", "q", "ctrl+c",
	"j", "k", "up", "down", "pgup", "pgdown", "home", "end",
}

// Cases is a decoded record-browser-cases/v1 document.
type Cases struct {
	FormatVersion  string              `json:"formatVersion"`
	ProfileVersion string              `json:"profileVersion"`
	Model          CasesModel          `json:"model"`
	Records        map[string][]Values `json:"records"`
	Cases          []Case              `json:"cases"`
}

// Values holds one record: field ID to string value.
type Values map[string]string

// CasesModel binds the cases to one model by canonical digest.
type CasesModel struct {
	CanonicalSHA256 string `json:"canonicalSHA256"`
}

// Case is one named scenario.
type Case struct {
	ID          string `json:"id"`
	Requirement string `json:"requirement"`
	Size        Size   `json:"size"`
	Records     string `json:"records"`
	Steps       []Step `json:"steps"`
	Expect      Expect `json:"expect"`
}

// Size is a terminal size in cells.
type Size struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

// Step is exactly one of Key, Resize or Replace.
type Step struct {
	Key     *string  `json:"key,omitempty"`
	Resize  *Size    `json:"resize,omitempty"`
	Replace *Replace `json:"replace,omitempty"`
}

// Replace supplies a new snapshot: Status is ready, loading or failed.
type Replace struct {
	Status  string  `json:"status"`
	Records *string `json:"records,omitempty"`
	Message *string `json:"message,omitempty"`
}

// Expect lists the observations to compare. Absent members are not checked.
type Expect struct {
	Layout   *string   `json:"layout,omitempty"`
	Focus    *string   `json:"focus,omitempty"`
	Help     *bool     `json:"help,omitempty"`
	Selected *string   `json:"selected,omitempty"`
	Status   *string   `json:"status,omitempty"`
	Visible  *[]string `json:"visible,omitempty"`
	Absent   *[]string `json:"absent,omitempty"`
	Quit     *bool     `json:"quit,omitempty"`
	// CapturesKeys and KeyHints observe CapturesKeys() and KeyHints().
	CapturesKeys *bool      `json:"capturesKeys,omitempty"`
	KeyHints     *[]KeyHint `json:"keyHints,omitempty"`
}

// KeyHint is one expected key hint: the key text and the action text.
type KeyHint struct {
	Key    string `json:"key"`
	Action string `json:"action"`
}

var (
	caseID  = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_]{0,63}$`)
	setName = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_]{0,31}$`)
	sha256x = regexp.MustCompile(`^[0-9a-f]{64}$`)
)

// DecodeCases rejects unknown, duplicate and case-variant members, null,
// trailing data, invalid UTF-8 and input over MaxCasesBytes. It checks
// everything that does not need the model; see (Cases).CheckAgainst.
func DecodeCases(raw []byte) (Cases, error) {
	var c Cases
	if len(raw) == 0 || len(raw) > MaxCasesBytes || !utf8.Valid(raw) {
		return c, fmt.Errorf("cases must be UTF-8 JSON of 1..%d bytes", MaxCasesBytes)
	}
	scan := json.NewDecoder(bytes.NewReader(raw))
	if err := checkCasesValue(scan, 0, "root"); err != nil {
		return c, err
	}
	if _, err := scan.Token(); err != io.EOF {
		return c, fmt.Errorf("cases have trailing data")
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	if err := d.Decode(&c); err != nil {
		return Cases{}, err
	}
	if err := c.validate(); err != nil {
		return Cases{}, err
	}
	return c, nil
}

// Allowed members per object context. A context without an entry (record
// sets, records) has free keys that later checks validate.
var caseMembers = map[string]string{
	"root":    "|formatVersion|profileVersion|model|records|cases|",
	"model":   "|canonicalSHA256|",
	"case":    "|id|requirement|size|records|steps|expect|",
	"size":    "|width|height|",
	"step":    "|key|resize|replace|",
	"resize":  "|width|height|",
	"replace": "|status|records|message|",
	"expect":  "|layout|focus|help|selected|status|visible|absent|quit|capturesKeys|keyHints|",
	"keyhint": "|key|action|",
}

// memberContext names the context of an object member's value.
var memberContext = map[string]map[string]string{
	"root":       {"model": "model", "records": "recordsets", "cases": "cases"},
	"recordsets": {},
	"case":       {"size": "size", "steps": "steps", "expect": "expect"},
	"step":       {"resize": "resize", "replace": "replace"},
	"expect":     {"keyHints": "keyhints"},
}

// elementContext names the context of an array's elements.
var elementContext = map[string]string{"cases": "case", "steps": "step", "recordset": "record", "keyhints": "keyhint"}

func checkCasesValue(d *json.Decoder, depth int, context string) error {
	if depth > 8 {
		return fmt.Errorf("JSON nesting exceeds 8")
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
		allowed, fixed := caseMembers[context]
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
			if fixed && !strings.Contains(allowed, "|"+k+"|") {
				return fmt.Errorf("unknown member %q in %s", k, context)
			}
			if seen[k] {
				return fmt.Errorf("duplicate key %q", k)
			}
			seen[k] = true
			child := memberContext[context][k]
			if context == "recordsets" {
				child = "recordset"
			}
			if err := checkCasesValue(d, depth+1, child); err != nil {
				return err
			}
		}
	case '[':
		for d.More() {
			if err := checkCasesValue(d, depth+1, elementContext[context]); err != nil {
				return err
			}
		}
	default:
		return fmt.Errorf("unexpected delimiter")
	}
	_, err = d.Token()
	return err
}

func plainText(s, name string, max int) error {
	if len(s) == 0 || len(s) > max || !utf8.ValidString(s) {
		return fmt.Errorf("%s must be 1..%d UTF-8 bytes", name, max)
	}
	for _, r := range s {
		if unicode.IsControl(r) || r == ' ' || r == ' ' {
			return fmt.Errorf("%s contains a control character", name)
		}
	}
	return nil
}

func validSize(s Size, name string) error {
	if s.Width < 1 || s.Width > 400 || s.Height < 1 || s.Height > 200 {
		return fmt.Errorf("%s must be 1..400 wide and 1..200 high", name)
	}
	return nil
}

func isAllowedKey(k string) bool {
	for _, a := range AllowedKeys {
		if a == k {
			return true
		}
	}
	return false
}

func oneOf(s string, options ...string) bool {
	for _, o := range options {
		if s == o {
			return true
		}
	}
	return false
}

func (c Cases) validate() error {
	if c.FormatVersion != CasesFormat {
		return fmt.Errorf("formatVersion must be %q", CasesFormat)
	}
	if c.ProfileVersion != Profile {
		return fmt.Errorf("profileVersion must be %q", Profile)
	}
	if !sha256x.MatchString(c.Model.CanonicalSHA256) {
		return fmt.Errorf("model.canonicalSHA256 must be 64 lowercase hex digits")
	}
	if len(c.Records) < 1 || len(c.Records) > maxRecordSets {
		return fmt.Errorf("records must hold 1..%d record sets", maxRecordSets)
	}
	for name, set := range c.Records {
		if !setName.MatchString(name) {
			return fmt.Errorf("record set name %q must match %s", name, setName)
		}
		if set == nil || len(set) > maxSetRecords {
			return fmt.Errorf("record set %q must hold 0..%d records", name, maxSetRecords)
		}
	}
	if len(c.Cases) < 1 || len(c.Cases) > maxCases {
		return fmt.Errorf("cases must hold 1..%d cases", maxCases)
	}
	ids := map[string]bool{}
	for _, k := range c.Cases {
		if !caseID.MatchString(k.ID) || ids[k.ID] {
			return fmt.Errorf("invalid or duplicate case id %q", k.ID)
		}
		ids[k.ID] = true
		if err := k.validate(c.Records); err != nil {
			return fmt.Errorf("case %s: %w", k.ID, err)
		}
	}
	return nil
}

func (k Case) validate(sets map[string][]Values) error {
	if err := plainText(k.Requirement, "requirement", maxRequirement); err != nil {
		return err
	}
	if err := validSize(k.Size, "size"); err != nil {
		return err
	}
	if _, ok := sets[k.Records]; !ok {
		return fmt.Errorf("records names unknown record set %q", k.Records)
	}
	if k.Steps == nil || len(k.Steps) > maxSteps {
		return fmt.Errorf("steps must be a list of 0..%d steps", maxSteps)
	}
	for i, s := range k.Steps {
		if err := s.validate(sets); err != nil {
			return fmt.Errorf("step %d: %w", i+1, err)
		}
	}
	return k.Expect.validate()
}

func (s Step) validate(sets map[string][]Values) error {
	n := 0
	if s.Key != nil {
		n++
		if !isAllowedKey(*s.Key) {
			return fmt.Errorf("key %q is not one of the profile's keys", *s.Key)
		}
	}
	if s.Resize != nil {
		n++
		if err := validSize(*s.Resize, "resize"); err != nil {
			return err
		}
	}
	if s.Replace != nil {
		n++
		r := s.Replace
		if !oneOf(r.Status, "ready", "loading", "failed") {
			return fmt.Errorf("replace.status must be ready, loading or failed")
		}
		if r.Records != nil {
			if r.Status != "ready" {
				return fmt.Errorf("replace.records is only allowed with status ready")
			}
			if _, ok := sets[*r.Records]; !ok {
				return fmt.Errorf("replace.records names unknown record set %q", *r.Records)
			}
		}
		if r.Message != nil {
			if err := plainText(*r.Message, "replace.message", maxRequirement); err != nil {
				return err
			}
		}
	}
	if n != 1 {
		return fmt.Errorf("a step must have exactly one of key, resize and replace")
	}
	return nil
}

func (e Expect) validate() error {
	n := 0
	if e.Layout != nil {
		n++
		if !oneOf(*e.Layout, "wide", "narrow", "tiny") {
			return fmt.Errorf("expect.layout must be wide, narrow or tiny")
		}
	}
	if e.Focus != nil {
		n++
		if !oneOf(*e.Focus, "list", "detail") {
			return fmt.Errorf("expect.focus must be list or detail")
		}
	}
	if e.Help != nil {
		n++
	}
	if e.Selected != nil {
		n++
		if len(*e.Selected) > maxValueBytes || !utf8.ValidString(*e.Selected) {
			return fmt.Errorf("expect.selected must be at most %d UTF-8 bytes", maxValueBytes)
		}
	}
	if e.Status != nil {
		n++
		if !oneOf(*e.Status, "ready", "loading", "failed") {
			return fmt.Errorf("expect.status must be ready, loading or failed")
		}
	}
	if e.Quit != nil {
		n++
	}
	if e.CapturesKeys != nil {
		n++
	}
	if e.KeyHints != nil {
		n++
		if len(*e.KeyHints) < 1 || len(*e.KeyHints) > maxKeyHints {
			return fmt.Errorf("expect.keyHints must hold 1..%d hints", maxKeyHints)
		}
		for _, h := range *e.KeyHints {
			if err := plainText(h.Key, "expect.keyHints key", maxExpectedText); err != nil {
				return err
			}
			if err := plainText(h.Action, "expect.keyHints action", maxExpectedText); err != nil {
				return err
			}
		}
	}
	for _, p := range []struct {
		name string
		list *[]string
	}{{"visible", e.Visible}, {"absent", e.Absent}} {
		if p.list == nil {
			continue
		}
		n++
		if len(*p.list) < 1 || len(*p.list) > maxExpectTexts {
			return fmt.Errorf("expect.%s must hold 1..%d strings", p.name, maxExpectTexts)
		}
		for _, s := range *p.list {
			if err := plainText(s, "expect."+p.name+" string", maxExpectedText); err != nil {
				return err
			}
		}
	}
	if n == 0 {
		return fmt.Errorf("expect must state at least one expectation")
	}
	return nil
}

// CanonicalSHA256 returns the canonical digest of the model whose compiled
// files are given.
func CanonicalSHA256(files map[string][]byte) string {
	return hexSum(bytes.TrimSuffix(files["model.json"], []byte("\n")))
}

// StaleError marks cases written for a different model.
type StaleError struct{ Cases, Model string }

func (e StaleError) Error() string {
	return fmt.Sprintf("cases are stale: they were written for model %s but the model is %s; "+
		"review every case against the changed model and set model.canonicalSHA256 only after that review", e.Cases, e.Model)
}

// CheckAgainst binds the cases to the model by digest and checks that every
// record has exactly the model's fields, as strings, with a unique nonempty
// identity. A digest mismatch is a StaleError.
func (c Cases) CheckAgainst(m Model, digest string) error {
	if c.Model.CanonicalSHA256 != digest {
		return StaleError{Cases: c.Model.CanonicalSHA256, Model: digest}
	}
	want := make([]string, len(m.Fields))
	for i, f := range m.Fields {
		want[i] = f.ID
	}
	sort.Strings(want)
	names := make([]string, 0, len(c.Records))
	for name := range c.Records {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		identities := map[string]bool{}
		for i, rec := range c.Records[name] {
			got := make([]string, 0, len(rec))
			for k, v := range rec {
				got = append(got, k)
				if len(v) > maxValueBytes || !utf8.ValidString(v) {
					return fmt.Errorf("records.%s[%d].%s must be at most %d UTF-8 bytes", name, i, k, maxValueBytes)
				}
			}
			sort.Strings(got)
			if strings.Join(got, ",") != strings.Join(want, ",") {
				return fmt.Errorf("records.%s[%d] must have exactly the fields %s, has %s",
					name, i, strings.Join(want, ","), strings.Join(got, ","))
			}
			id := rec[m.Identity]
			if id == "" || identities[id] {
				return fmt.Errorf("records.%s[%d]: identity %s must be unique and nonempty", name, i, m.Identity)
			}
			identities[id] = true
		}
	}
	return nil
}
