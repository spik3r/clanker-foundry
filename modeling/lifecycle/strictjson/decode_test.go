package strictjson

import (
	"encoding/json"
	"math"
	"strings"
	"testing"
)

type fixture struct {
	Name     string   `json:"name"`
	Count    uint64   `json:"count"`
	Items    []string `json:"items"`
	Optional *string  `json:"optional,omitempty"`
}

func TestExactFieldsNumbersAndUnicode(t *testing.T) {
	valid := []string{`{"name":"\ud83d\ude00","count":18446744073709551615,"items":[]}`, `{"name":"日本語","count":0,"items":["x"],"optional":null}`, `{"name":"\\uD800","count":1,"items":[]}`}
	for _, s := range valid {
		var v fixture
		if err := Decode([]byte(s), &v, 1<<20, 64); err != nil {
			t.Fatalf("valid %q: %v", s, err)
		}
	}
	var v fixture
	if err := Decode([]byte(valid[0]), &v, 1<<20, 64); err != nil || v.Count != math.MaxUint64 || v.Name != "😀" {
		t.Fatalf("lossy decode: %#v %v", v, err)
	}
	invalid := []string{`{"name":"x","name":"y","count":0,"items":[]}`, `{"name":"x","\u006eame":"y","count":0,"items":[]}`, `{"Name":"x","count":0,"items":[]}`, `{"name":"x","count":0,"items":[],"extra":1}`, `{"name":"x","count":0}`, `{"name":null,"count":0,"items":[]}`, `{"name":"x","count":null,"items":[]}`, `{"name":"x","count":0,"items":null}`, `{"name":"x","count":0,"items":[null]}`, `{"name":"x","count":18446744073709551616,"items":[]}`, `{"name":"x","count":-1,"items":[]}`, `{"name":"x","count":1.0,"items":[]}`, `{"name":"x","count":"1","items":[]}`, `{"name":"\ud800","count":0,"items":[]}`, `{"name":"\udc00","count":0,"items":[]}`, `{"name":"\ud800\u0041","count":0,"items":[]}`, valid[0] + ` {}`, string([]byte{'"', 0xff, '"'})}
	for _, s := range invalid {
		t.Run(s, func(t *testing.T) {
			var got fixture
			if Decode([]byte(s), &got, 1<<20, 64) == nil {
				t.Fatal("accepted invalid input")
			}
		})
	}
	var n json.Number
	if err := Decode([]byte("9007199254740993"), &n, 128, 4); err != nil || n.String() != "9007199254740993" {
		t.Fatalf("number precision: %q %v", n, err)
	}
}
func TestDecodeBoundsAndDestination(t *testing.T) {
	var out any
	for _, tc := range []struct {
		s    string
		b, d int
	}{{`[]`, 1, 4}, {`[]`, 2, 0}, {`[]`, 0, 4}, {`[[0]]`, 100, 1}, {strings.Repeat("[", 66) + "0" + strings.Repeat("]", 66), 1000, 64}, {"[" + strings.Repeat("0,", 100000) + "0]", 1 << 20, 64}} {
		if Decode([]byte(tc.s), &out, tc.b, tc.d) == nil {
			t.Fatalf("accepted bounds %+v", tc)
		}
	}
	if err := Decode([]byte(`[[0]]`), &out, 5, 2); err != nil {
		t.Fatal(err)
	}
	for _, dst := range []any{nil, fixture{}, (*fixture)(nil)} {
		if Decode([]byte(`{}`), dst, 100, 4) == nil {
			t.Fatalf("accepted destination %T", dst)
		}
	}
	var nested map[string]json.RawMessage
	if Decode([]byte(`{"raw":{"x":1,"x":2}}`), &nested, 100, 5) == nil {
		t.Fatal("RawMessage hid duplicate")
	}
}
