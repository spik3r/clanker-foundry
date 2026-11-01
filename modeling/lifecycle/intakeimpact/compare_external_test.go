package intakeimpact_test

// Literal changes and gaps were written by a separate assistant before the new
// comparator. Native receipt checks reuse intake, not an independent oracle for
// that existing component. No person-level independence is claimed.
import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/intake"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/intakeimpact"
)

func fixture(t *testing.T, name string) []byte {
	t.Helper()
	b, e := os.ReadFile("testdata/" + name + ".json")
	if e != nil {
		t.Fatal(e)
	}
	return b
}
func digest(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }
func object(t *testing.T, b []byte) map[string]json.RawMessage {
	t.Helper()
	var m map[string]json.RawMessage
	if e := json.Unmarshal(b, &m); e != nil {
		t.Fatal(e)
	}
	return m
}
func wire(t *testing.T, v any) map[string]json.RawMessage {
	t.Helper()
	b, e := json.Marshal(v)
	if e != nil {
		t.Fatal(e)
	}
	return object(t, b)
}
func eqJSON(t *testing.T, actual, expected []byte) {
	t.Helper()
	var a, b any
	if e := json.Unmarshal(actual, &a); e != nil {
		t.Fatal(e)
	}
	if e := json.Unmarshal(expected, &b); e != nil {
		t.Fatal(e)
	}
	if !reflect.DeepEqual(a, b) {
		t.Fatalf("got %s\nwant %s", actual, expected)
	}
}
func str(t *testing.T, m map[string]json.RawMessage, k string) string {
	t.Helper()
	var s string
	if e := json.Unmarshal(m[k], &s); e != nil {
		t.Fatal(e)
	}
	return s
}
func flag(t *testing.T, m map[string]json.RawMessage, k string) bool {
	t.Helper()
	var b bool
	if e := json.Unmarshal(m[k], &b); e != nil {
		t.Fatal(e)
	}
	return b
}
func compare(t *testing.T, b, a []byte) map[string]json.RawMessage {
	t.Helper()
	r, e := intakeimpact.Compare(b, a)
	if e != nil {
		t.Fatal(e)
	}
	return wire(t, r)
}
func encode(t *testing.T, m map[string]json.RawMessage) []byte {
	t.Helper()
	b, e := json.Marshal(m)
	if e != nil {
		t.Fatal(e)
	}
	return b
}
func replace(t *testing.T, b []byte, k string, v any) []byte {
	t.Helper()
	m := object(t, b)
	raw, e := json.Marshal(v)
	if e != nil {
		t.Fatal(e)
	}
	m[k] = raw
	return encode(t, m)
}

func TestFrozenLiteralChangesAndGaps(t *testing.T) {
	b, a := fixture(t, "before"), fixture(t, "after")
	expected := object(t, fixture(t, "expected"))
	got := compare(t, b, a)
	for _, k := range []string{"changes", "beforeGaps", "afterGaps"} {
		eqJSON(t, got[k], expected[k])
	}
	if !flag(t, got, "requestSourceChanged") || !flag(t, got, "contentChanged") {
		t.Fatal("changed fixture lost identity change")
	}
	for k, want := range map[string]string{"formatVersion": "lifecycle-intake-impact/v1", "mode": "example-only", "authority": "none", "artifactValidation": "not-run", "executionStatus": "not-run", "reviewStatus": "not-established"} {
		if str(t, got, k) != want {
			t.Fatalf("%s: %s", k, got[k])
		}
	}
}

func checkReceipt(t *testing.T, raw, receipt []byte) {
	t.Helper()
	m := object(t, receipt)
	native, e := intake.Execute(raw)
	if e != nil {
		t.Fatal(e)
	}
	if got := str(t, m, "reportJSON"); got != string(native) || !strings.HasSuffix(got, "\n") {
		t.Fatal("native receipt not preserved exactly")
	}
	if str(t, m, "requestSha256") != digest(raw) || str(t, m, "reportSha256") != digest(native) {
		t.Fatal("receipt digest mismatch")
	}
	n := object(t, native)
	// Assemble the documented fixed-order content object from exact native values.
	content := []byte(`{"formatVersion":"lifecycle-intake-content/v1"`)
	for _, k := range []string{"intent", "decisions", "questions", "bindings", "links"} {
		content = append(content, []byte(`,"`+k+`":`)...)
		content = append(content, n[k]...)
	}
	content = append(content, '}')
	if str(t, m, "contentSha256") != digest(content) {
		t.Fatalf("content digest differs from fixed-order native values: %s", content)
	}
}

func TestNativeReceiptsAndInputOwnership(t *testing.T) {
	b, a := fixture(t, "before"), fixture(t, "after")
	bc, ac := bytes.Clone(b), bytes.Clone(a)
	r := compare(t, b, a)
	checkReceipt(t, b, r["before"])
	checkReceipt(t, a, r["after"])
	if !bytes.Equal(b, bc) || !bytes.Equal(a, ac) {
		t.Fatal("inputs mutated")
	}
	first := encode(t, r)
	for i := 0; i < 10; i++ {
		if !bytes.Equal(first, encode(t, compare(t, b, a))) {
			t.Fatal("nondeterministic output")
		}
	}
	for i := range b {
		b[i] = 'x'
	}
	for i := range a {
		a[i] = 'x'
	}
	if !bytes.Equal(first, encode(t, r)) {
		t.Fatal("returned receipt aliases input")
	}
}

func TestNoOpRetainsEveryGap(t *testing.T) {
	b := fixture(t, "before")
	r := compare(t, b, b)
	if flag(t, r, "contentChanged") || flag(t, r, "requestSourceChanged") {
		t.Fatal("no-op changed")
	}
	eqJSON(t, r["changes"], []byte(`[]`))
	want := object(t, fixture(t, "expected"))["beforeGaps"]
	eqJSON(t, r["beforeGaps"], want)
	eqJSON(t, r["afterGaps"], want)
}

func TestFormattingAndRecordReorderingDoNotChangeContent(t *testing.T) {
	b := fixture(t, "before")
	m := object(t, b)
	for _, k := range []string{"decisions", "questions"} {
		var values []json.RawMessage
		if e := json.Unmarshal(m[k], &values); e != nil {
			t.Fatal(e)
		}
		for i, j := 0, len(values)-1; i < j; i, j = i+1, j-1 {
			values[i], values[j] = values[j], values[i]
		}
		m[k], _ = json.Marshal(values)
	}
	a := encode(t, m)
	r := compare(t, b, a)
	if !flag(t, r, "requestSourceChanged") || flag(t, r, "contentChanged") {
		t.Fatal("source and content identities conflated")
	}
	eqJSON(t, r["changes"], []byte(`[]`))
	checkReceipt(t, b, r["before"])
	checkReceipt(t, a, r["after"])
}

func TestNullAndEmptyObservationRemainDistinct(t *testing.T) {
	b := fixture(t, "before")
	m := object(t, b)
	o := object(t, m["observed"])
	o["model"] = json.RawMessage(`""`)
	m["observed"] = encode(t, o)
	r := compare(t, b, encode(t, m))
	eqJSON(t, r["changes"], []byte(`[{"kind":"artifact","id":"model","change":"modified","fields":["observedSha256","status"]}]`))
	if !flag(t, r, "contentChanged") {
		t.Fatal("empty observation collapsed into missing")
	}
}

func TestRawArtifactWhitespaceChangesContentWithoutRebinding(t *testing.T) {
	b := fixture(t, "before")
	m := object(t, b)
	o := object(t, m["observed"])
	o["requirements"] = json.RawMessage(`"abc\n"`)
	m["observed"] = encode(t, o)
	r := compare(t, b, encode(t, m))
	eqJSON(t, r["changes"], []byte(`[{"kind":"artifact","id":"requirements","change":"modified","fields":["observedSha256","status"]}]`))
	if !flag(t, r, "contentChanged") {
		t.Fatal("opaque bytes normalized")
	}
}

func TestInvalidIntakeHasNoPartialReport(t *testing.T) {
	valid := fixture(t, "before")
	m := object(t, valid)
	delete(m, "intent")
	invalid := [][]byte{nil, []byte(`{}`), append(bytes.Clone(valid), []byte(`{}`)...), encode(t, m), replace(t, valid, "formatVersion", "bad"), replace(t, valid, "intent", nil), replace(t, valid, "intent", "")}
	invalid = append(invalid, bytes.Repeat([]byte{' '}, intake.MaxSourceBytes+1))
	for i, b := range invalid {
		for _, side := range []bool{false, true} {
			left, right := b, valid
			if side {
				left, right = valid, b
			}
			r, e := intakeimpact.Compare(left, right)
			if e == nil {
				t.Fatalf("case %d side %v accepted", i, side)
			}
			if !reflect.DeepEqual(r, intakeimpact.Report{}) {
				t.Fatalf("case %d leaked partial report", i)
			}
		}
	}
}

func outer(t *testing.T, b, a []byte) []byte {
	t.Helper()
	raw, e := json.Marshal(map[string]any{"formatVersion": "lifecycle-intake-impact-request/v1", "before": string(b), "after": string(a)})
	if e != nil {
		t.Fatal(e)
	}
	return raw
}
func TestStrictOuterTransport(t *testing.T) {
	b, a := fixture(t, "before"), fixture(t, "after")
	raw := outer(t, b, a)
	out, e := intakeimpact.Execute(raw)
	if e != nil {
		t.Fatal(e)
	}
	eqJSON(t, out, encode(t, compare(t, b, a)))
	invalid := [][]byte{nil, []byte(`{}`), append(bytes.Clone(raw), []byte(`{}`)...), replace(t, raw, "formatVersion", "bad"), replace(t, raw, "before", nil), replace(t, raw, "before", map[string]any{}), replace(t, raw, "extra", true), []byte(`{"formatVersion":"lifecycle-intake-impact-request/v1","before":"x","before":"x","after":"x"}`), []byte(`{"formatVersion":"lifecycle-intake-impact-request/v1","Before":"x","after":"x"}`), []byte(`{"formatVersion":"lifecycle-intake-impact-request/v1","before":"\ud800","after":"x"}`)}
	m := object(t, raw)
	delete(m, "after")
	invalid = append(invalid, encode(t, m))
	for i, v := range invalid {
		out, e := intakeimpact.Execute(v)
		if e == nil || len(out) != 0 {
			t.Fatalf("transport case %d: bytes=%d error=%v", i, len(out), e)
		}
	}
}
