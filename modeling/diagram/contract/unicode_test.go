package contract

import (
	"strings"
	"testing"
)

func TestStrictScanUnicodeRefusals(t *testing.T) {
	for _, text := range []string{`\ud800`, `\udbff`, `\udc00`, `\udfff`, `\udc00\ud800`, `\ud800x`, `\ud800\u0041`, `\ud800\ud800`, `\ud800\\udc00`, `\ud800\udc00\udfff`} {
		for _, raw := range []string{`{"a":"` + text + `"}`, `{"list":[{"k":"` + text + `"}]}`, `{"` + text + `":1}`} {
			if err := StrictScan([]byte(raw), 1<<20, testSchema); err == nil || err.Error() != "unpaired surrogate escape in JSON" {
				t.Errorf("%s: refusal = %v", raw, err)
			}
		}
	}
}

func TestStrictScanUnicodeValid(t *testing.T) {
	for _, text := range []string{`\ud83d\ude00`, `\uD800\uDC00`, `\udbff\udfff`, `\ufffd`, "�", `\\ud800`, `\\\"\ud83d\ude00\/`, `\u0061`, `\b\f\n\r\t`} {
		if err := StrictScan([]byte(`{"a":"`+text+`"}`), 1<<20, testSchema); err != nil {
			t.Errorf("%s: %v", text, err)
		}
	}
	// Escaped valid member names still resolve to their decoded names.
	if err := StrictScan([]byte(`{"\u0061":"ok"}`), 1<<20, testSchema); err != nil {
		t.Fatal(err)
	}
}

func TestStrictScanUnicodeKeepsExistingRefusals(t *testing.T) {
	for _, tc := range []struct{ raw, want string }{
		{`{"a":"\ud800\u"}`, "a: invalid escape sequence `\\u\"}` in string"},
		{`{"a":"\uxxxx"}`, "a: invalid escape sequence `\\uxxxx` in string"},
		{`{"a":"\q"}`, "a: invalid escape sequence `\\q` in string"},
		{`{} {}`, `model has trailing data`},
		{`{"a":null}`, `a: null is not supported`},
		{`{"a":1,"a":2}`, `duplicate key "a"`},
		{`{"A":1}`, `unknown field "A"`},
		{`{"a":[]}`, `a: unexpected array`},
	} {
		if err := StrictScan([]byte(tc.raw), 1<<20, testSchema); err == nil || err.Error() != tc.want {
			t.Errorf("%s: %v, want %s", tc.raw, err, tc.want)
		}
	}
	if err := StrictScan([]byte(`{"a":"\ud800"}`), 1, testSchema); err == nil || err.Error() != "model must be 1..1 bytes" {
		t.Errorf("size precedence: %v", err)
	}
	if err := StrictScan([]byte("{\"a\":\"\xff\"}"), 1<<20, testSchema); err == nil || err.Error() != "model is not valid UTF-8" {
		t.Errorf("UTF-8 refusal: %v", err)
	}
	deep := strings.Repeat(`{"nest":`, MaxDepth+1) + `{}` + strings.Repeat(`}`, MaxDepth+1)
	if err := StrictScan([]byte(deep), 1<<20, testSchema); err == nil || !strings.Contains(err.Error(), "JSON nesting exceeds 16") {
		t.Errorf("depth refusal: %v", err)
	}
}
