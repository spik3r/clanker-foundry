// Package modeltext holds the plain-text rules that profiles share for the
// text a model carries.
package modeltext

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Runes checks that s is valid UTF-8 of at most max bytes and holds no control
// character, no U+2028, U+2029, U+FFFE or U+FFFF. It accepts an empty or
// blank string.
func Runes(s, name string, max int) error {
	if len(s) > max || !utf8.ValidString(s) {
		return fmt.Errorf("%s must be at most %d UTF-8 bytes", name, max)
	}
	for _, r := range s {
		if unicode.IsControl(r) || r == '\u2028' || r == '\u2029' || r == '\ufffe' || r == '\uffff' {
			return fmt.Errorf("%s contains a control or non-XML character", name)
		}
	}
	return nil
}

// Plain checks a label, title or paragraph: 1..max bytes, not blank, with the
// character rules of Runes.
func Plain(s, name string, max int) error {
	if len(s) == 0 || len(s) > max {
		return fmt.Errorf("%s must be 1..%d UTF-8 bytes", name, max)
	}
	if err := Runes(s, name, max); err != nil {
		return err
	}
	if strings.TrimSpace(s) == "" {
		return fmt.Errorf("%s must not be blank", name)
	}
	return nil
}
