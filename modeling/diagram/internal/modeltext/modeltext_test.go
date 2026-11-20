package modeltext

import (
	"strings"
	"testing"
)

func TestPlain(t *testing.T) {
	for _, ok := range []string{"x", "é", strings.Repeat("x", 10), "a b", "😀"} {
		if err := Plain(ok, "t", 10); err != nil {
			t.Errorf("%q: %v", ok, err)
		}
	}
	for _, bad := range []string{"", " ", "\u00a0", strings.Repeat("x", 11), "a\nb", "a\u2028b", "a\u2029b", "a\ufffeb", "a\uffffb", "a\x7fb", "a\xffb", "a\u0085b"} {
		if err := Plain(bad, "t", 10); err == nil {
			t.Errorf("%q was accepted", bad)
		}
	}
	if err := Runes("", "t", 10); err != nil {
		t.Errorf("Runes refuses an empty string: %v", err)
	}
	if err := Runes("   ", "t", 10); err != nil {
		t.Errorf("Runes refuses a blank string: %v", err)
	}
}
