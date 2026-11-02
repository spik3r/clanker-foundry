package scenarios

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"math"
	"reflect"
	"strings"
	"testing"
)

func TestFixtureRoundtripAndCanonicalIdentity(t *testing.T) {
	s := minimalSuite()
	b, e := Encode(s)
	if e != nil {
		t.Fatal(e)
	}
	got, e := Decode(b)
	if e != nil || !reflect.DeepEqual(got, s) {
		t.Fatal("roundtrip", e)
	}
	d, e := Digest(s)
	if e != nil {
		t.Fatal(e)
	}
	// The expected canonical shape is a handwritten literal, not toDTO output.
	literal := `{"scenarioVersion":"lifecycle-scenarios/v1","profileVersion":"lifecycle/v1","modelDigest":"` + s.ModelDigest + `","id":"Example","cases":[{"id":"Happy","initial":{"id":"task-1","state":"Draft","revision":"0","outcome":"Ship outcome","acceptance":["Verified"],"owner":null},"steps":[{"id":"Prepare","command":{"type":"prepare","actor":"","expectedRevision":"0","review":null},"expected":{"ok":true,"transitionId":"T_prepare","task":{"id":"task-1","state":"Ready","revision":"1","outcome":"Ship outcome","acceptance":["Verified"],"owner":null},"refusals":[]}}]}]}`
	h := sha256.Sum256([]byte(literal))
	if d != hex.EncodeToString(h[:]) {
		t.Fatal("canonical field/identity drift")
	}
	var expanded bytes.Buffer
	if e := json.Indent(&expanded, b, "", "  "); e != nil {
		t.Fatal(e)
	}
	other, e := Decode(expanded.Bytes())
	if e != nil {
		t.Fatal(e)
	}
	d2, _ := Digest(other)
	if d != d2 {
		t.Fatal("source formatting affected semantics")
	}
	s.Cases[0].Steps[0].Expected.Task.Outcome = "Different expectation"
	d2, _ = Digest(s)
	if d == d2 {
		t.Fatal("expectation excluded from digest")
	}
}
func TestNilEmptyAndDecimalRoundtrip(t *testing.T) {
	for _, empty := range []bool{false, true} {
		s := minimalSuite()
		s.Cases[0].Initial.Acceptance = nil
		if empty {
			s.Cases[0].Initial.Acceptance = []string{}
		}
		owner := ""
		s.Cases[0].Initial.Owner = &owner
		s.Cases[0].Initial.Revision = math.MaxUint64
		s.Cases[0].Steps[0].Command.ExpectedRevision = math.MaxUint64
		b, e := Encode(s)
		if e != nil {
			t.Fatal(e)
		}
		if !bytes.Contains(b, []byte(`"revision":"18446744073709551615"`)) {
			t.Fatal("revision representation")
		}
		got, e := Decode(b)
		if e != nil || !reflect.DeepEqual(got, s) {
			t.Fatal("representation lost", e)
		}
	}
	a := minimalSuite()
	a.Cases[0].Initial.Acceptance = nil
	b := minimalSuite()
	b.Cases[0].Initial.Acceptance = []string{}
	da, _ := Digest(a)
	db, _ := Digest(b)
	if da == db {
		t.Fatal("scenario representation identity lost nil/empty distinction")
	}
}
func TestStrictFixtureJSONRefusals(t *testing.T) {
	raw, e := Encode(minimalSuite())
	if e != nil {
		t.Fatal(e)
	}

	var obj map[string]any
	if e := json.Unmarshal(raw, &obj); e != nil {
		t.Fatal(e)
	}
	obj["cases"].([]any)[0].(map[string]any)["initial"] = nil
	nullInitial, e := json.Marshal(obj)
	if e != nil {
		t.Fatal(e)
	}
	cases := map[string][]byte{
		"extra":            bytes.Replace(raw, []byte(`"id":"Example"`), []byte(`"id":"Example","extra":1`), 1),
		"case variant":     bytes.Replace(raw, []byte(`"id":"Example"`), []byte(`"Id":"Example"`), 1),
		"duplicate nested": bytes.Replace(raw, []byte(`"actor":""`), []byte(`"actor":"","\u0061ctor":"other"`), 1),
		"missing owner":    bytes.Replace(raw, []byte(`,"owner":null`), nil, 1),
		"null required":    nullInitial,
		"nil reasons":      bytes.Replace(raw, []byte(`"refusals":[]`), []byte(`"refusals":null`), 1),
		"numeric revision": bytes.Replace(raw, []byte(`"revision":"0"`), []byte(`"revision":0`), 1),
		"trailing":         append(append([]byte(nil), raw...), []byte(`{}`)...),
		"invalid UTF8":     append([]byte{255}, raw...),
		"surrogate":        bytes.Replace(raw, []byte(`Ship outcome`), []byte(`\ud800`), 1),
		"format":           bytes.Replace(raw, []byte(FormatVersion), []byte("unknown/v1"), 1),
		"scenario version": bytes.Replace(raw, []byte(ScenarioVersion), []byte("unknown/v1"), 1),
		"oversize":         bytes.Repeat([]byte(" "), MaxSourceBytes+1),
		"depth":            []byte(strings.Repeat("[", 65) + "0" + strings.Repeat("]", 65)),
	}
	for _, n := range []string{"00", "01", "+1", "-1", " 1", "1 ", "1.0", "1e2", "١", "18446744073709551616", ""} {
		cases["decimal "+n] = bytes.Replace(raw, []byte(`"revision":"0"`), []byte(`"revision":"`+n+`"`), 1)
	}
	for name, b := range cases {
		t.Run(name, func(t *testing.T) {
			if _, e := Decode(b); e == nil {
				t.Fatal("malformed fixture accepted")
			}
		})
	}
}
