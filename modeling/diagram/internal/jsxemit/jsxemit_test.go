package jsxemit

import (
	"encoding/json"
	"testing"
)

func TestStringEscapesAndRoundTrips(t *testing.T) {
	cases := map[string]string{
		"plain":          `"plain"`,
		`a"b\c`:          `"a\"b\\c"`,
		"a\nb\r\tc":      `"a\nb\r\tc"`,
		"<b>&":           `"\u003cb\u003e\u0026"`,
		"a\u2028b\u2029": `"a\u2028b\u2029"`,
		"{}`${x}`":       "\"{}`${x}`\"",
		"é\x01\x7f":      `"é\u0001\u007f"`,
	}
	for in, want := range cases {
		got := String(in)
		if got != want {
			t.Errorf("String(%q) = %s, want %s", in, got, want)
		}
		var back string
		if err := json.Unmarshal([]byte(got), &back); err != nil || back != in {
			t.Errorf("String(%q) does not round-trip: %q, %v", in, back, err)
		}
	}
}

func TestWriterAndAttr(t *testing.T) {
	w := &Writer{}
	w.Line("a")
	w.Depth++
	w.Line("b" + Attr("x", "<1>"))
	if got, want := w.String(), "a\n  b x={\"\\u003c1\\u003e\"}\n"; got != want {
		t.Errorf("got %q want %q", got, want)
	}
}
