package contract

import (
	"strings"
	"testing"
)

var testSchema = Schema{
	"model": {Fields: map[string]string{"a": "", "list": "list", "nest": "model"}},
	"list":  {Array: true, Elem: "item"},
	"item":  {Fields: map[string]string{"k": ""}},
}

func TestStrictScan(t *testing.T) {
	deep := strings.Repeat(`{"nest":`, MaxDepth+1) + `{}` + strings.Repeat(`}`, MaxDepth+1)
	ok := strings.Repeat(`{"nest":`, MaxDepth) + `{}` + strings.Repeat(`}`, MaxDepth)
	cases := []struct {
		name, raw string
		ok        bool
	}{
		{"object", `{"a":"x","list":[{"k":1}]}`, true},
		{"depth at limit", ok, true},
		{"depth over limit", deep, false},
		{"duplicate", `{"a":1,"a":2}`, false},
		{"unknown", `{"b":1}`, false},
		{"case variant", `{"A":1}`, false},
		{"null", `{"a":null}`, false},
		{"array for object element", `{"list":[[]]}`, false},
		{"array for object", `{"nest":[]}`, false},
		{"object for array", `{"list":{}}`, false},
		{"trailing", `{} {}`, false},
		{"trailing text", `{} x`, false},
		{"invalid UTF-8", "{\"a\":\"\xff\"}", false},
		{"empty", ``, false},
	}
	for _, c := range cases {
		err := StrictScan([]byte(c.raw), 1<<20, testSchema)
		if (err == nil) != c.ok {
			t.Errorf("%s: err = %v, want ok=%v", c.name, err, c.ok)
		}
	}
	if err := StrictScan([]byte(`{"a":"xxxxxxxx"}`), 8, testSchema); err == nil {
		t.Errorf("size limit not enforced")
	}
}
