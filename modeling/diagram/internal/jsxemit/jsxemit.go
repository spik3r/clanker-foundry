// Package jsxemit holds the JavaScript and JSX emission helpers that more
// than one diagram profile uses: the string-literal escaper and a small
// indenting line writer. Model text only ever enters emitted source through
// String, so it cannot be read as code.
package jsxemit

import "strings"

// String writes s as a double-quoted JavaScript string literal. Besides the
// escapes the language needs, it writes <, > and & as \u escapes, so the
// literal holds no markup character even if the file is embedded elsewhere.
func String(s string) string {
	var b strings.Builder
	b.WriteByte('"')
	for _, r := range s {
		switch {
		case r == '"':
			b.WriteString(`\"`)
		case r == '\\':
			b.WriteString(`\\`)
		case r == '\n':
			b.WriteString(`\n`)
		case r == '\r':
			b.WriteString(`\r`)
		case r == '\t':
			b.WriteString(`\t`)
		case r < 0x20 || r == 0x7f || r == '<' || r == '>' || r == '&' || r == 0x2028 || r == 0x2029:
			b.WriteString(`\u` + strings.ToLower(hex4(int(r))))
		default:
			b.WriteRune(r)
		}
	}
	b.WriteByte('"')
	return b.String()
}

func hex4(n int) string {
	const digits = "0123456789ABCDEF"
	return string([]byte{digits[n>>12&15], digits[n>>8&15], digits[n>>4&15], digits[n&15]})
}

// Attr writes ` name={"value"}`: an attribute whose value is a JavaScript
// string literal, with a leading space.
func Attr(name, value string) string { return " " + name + "={" + String(value) + "}" }

// Writer is a small indenting line writer. Depth is the current indentation,
// in steps of two spaces.
type Writer struct {
	b     strings.Builder
	Depth int
}

// Line writes s on its own line at the current depth.
func (w *Writer) Line(s string) {
	for i := 0; i < w.Depth; i++ {
		w.b.WriteString("  ")
	}
	w.b.WriteString(s)
	w.b.WriteByte('\n')
}

// String returns everything written so far.
func (w *Writer) String() string { return w.b.String() }
