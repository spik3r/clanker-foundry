// Package svgx holds the small text and encoding helpers shared by the UML-style
// diagram profiles (sequence-diagram/v1 and class-diagram/v1): identifier and
// plain-text rules, XML escaping, character-budget wrapping and canonical JSON
// encoding. It does no layout and knows nothing about any one profile.
package svgx

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Format is the input format shared by every diagram profile.
const Format = "diagram-json/v1"

// Identifier matches ids: an ASCII letter, then letters, digits or underscores,
// 1..48 characters in all.
var Identifier = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_]{0,47}$`)

// Plain checks a plain-text value: 1..max UTF-8 bytes, not blank, no control
// character, no U+2028, U+2029, U+FFFE or U+FFFF. These are the same character
// rules the process profile applies to labels.
func Plain(s, name string, max int) error {
	if len(s) == 0 || len(s) > max || !utf8.ValidString(s) {
		return fmt.Errorf("%s must be 1..%d UTF-8 bytes", name, max)
	}
	if strings.TrimSpace(s) == "" {
		return fmt.Errorf("%s must not be blank", name)
	}
	for _, r := range s {
		if unicode.IsControl(r) || r == ' ' || r == ' ' || r == '￾' || r == '￿' {
			return fmt.Errorf("%s contains a control or non-XML character", name)
		}
	}
	return nil
}

// Esc escapes text for XML character data and attribute values.
func Esc(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch r {
		case '&':
			b.WriteString("&amp;")
		case '<':
			b.WriteString("&lt;")
		case '>':
			b.WriteString("&gt;")
		case '"':
			b.WriteString("&quot;")
		case '\'':
			b.WriteString("&apos;")
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}

// Truncate shortens s to at most n characters (runes), ending with an
// ellipsis when it cuts.
func Truncate(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	if n < 1 {
		return ""
	}
	return string(r[:n-1]) + "…"
}

// Wrap breaks s into at most maxLines lines of at most budget characters,
// filling lines greedily word by word. A word longer than a line is cut into
// line-sized pieces. Text beyond maxLines is dropped and the last line ends
// with an ellipsis.
func Wrap(s string, budget, maxLines int) []string {
	var lines []string
	cur := ""
	curLen := 0
	flush := func() {
		if cur != "" {
			lines = append(lines, cur)
			cur, curLen = "", 0
		}
	}
	for _, w := range strings.Fields(s) {
		r := []rune(w)
		for len(r) > budget {
			flush()
			lines = append(lines, string(r[:budget]))
			r = r[budget:]
		}
		switch {
		case cur == "":
			cur, curLen = string(r), len(r)
		case curLen+1+len(r) <= budget:
			cur += " " + string(r)
			curLen += 1 + len(r)
		default:
			flush()
			cur, curLen = string(r), len(r)
		}
	}
	flush()
	if len(lines) > maxLines {
		lines = lines[:maxLines]
		last := []rune(lines[maxLines-1])
		if len(last) >= budget {
			last = last[:budget-1]
		}
		lines[maxLines-1] = strings.TrimRight(string(last), " ") + "…"
	}
	return lines
}

// Itoa formats an integer.
func Itoa(n int) string { return strconv.Itoa(n) }

// Encode marshals v as JSON without HTML escaping, with a trailing newline,
// indented when indent is true.
func Encode(v any, indent bool) ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if indent {
		enc.SetIndent("", "  ")
	}
	if err := enc.Encode(v); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Compact is Encode without indentation and without the trailing newline: the
// canonical form whose digest identifies a model.
func Compact(v any) ([]byte, error) {
	b, err := Encode(v, false)
	if err != nil {
		return nil, err
	}
	return bytes.TrimSuffix(b, []byte("\n")), nil
}
