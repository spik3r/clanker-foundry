package scenariodiff

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/jsonformat"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/model"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/scenarios"
)

func boundaryBytes(t *testing.T, m model.Model, s scenarios.Suite) ([]byte, []byte) {
	t.Helper()
	mb, e := jsonformat.Encode(m)
	if e != nil {
		t.Fatal(e)
	}
	r, _, e := validator(m)
	if e != nil {
		t.Fatal(e)
	}
	sb, e := r.Encode(s)
	if e != nil {
		t.Fatal(e)
	}
	return bytes.TrimSpace(mb), bytes.TrimSpace(sb)
}
func boundaryWire(bm, bs, am, as []byte) []byte {
	return []byte(`{"formatVersion":"lifecycle-scenario-diff-json/v1","beforeModel":` + string(bm) + `,"beforeSuite":` + string(bs) + `,"afterModel":` + string(am) + `,"afterSuite":` + string(as) + `}`)
}
func boundaryHash(raw []byte) string { h := sha256.Sum256(raw); return hex.EncodeToString(h[:]) }

func TestWireParityAndSourceIdentity(t *testing.T) {
	bm, bs := boundaryInputs(t)
	am, as := boundaryInputs(t)
	as.Cases[0].Steps[0].Command.Actor = "<edited>& 雪 literal \\u003c"
	bmb, bsb := boundaryBytes(t, bm, bs)
	amb, asb := boundaryBytes(t, am, as)
	raw := boundaryWire(bmb, bsb, amb, asb)
	original := bytes.Clone(raw)
	out, e := Execute(raw)
	if e != nil {
		t.Fatal(e)
	}
	var got Report
	if e = json.Unmarshal(out, &got); e != nil {
		t.Fatal(e)
	}
	if got.Sources == nil || *got.Sources != (SourceDigests{boundaryHash(raw), boundaryHash(bmb), boundaryHash(bsb), boundaryHash(amb), boundaryHash(asb)}) {
		t.Fatal("source hashes", got.Sources)
	}
	got.Sources = nil
	want, e := Compare(bm, bs, am, as)
	if e != nil || !reflect.DeepEqual(got, want) {
		t.Fatal("typed/wire differs", e)
	}
	again, e := Execute(raw)
	if e != nil || !bytes.Equal(out, again) || !bytes.Equal(original, raw) {
		t.Fatal("nondeterministic/mutating wire")
	}
	if !bytes.HasSuffix(out, []byte{'\n'}) || bytes.HasSuffix(out, []byte{'\n', '\n'}) {
		t.Fatal("terminator")
	}
	var spaced bytes.Buffer
	if e = json.Indent(&spaced, bsb, "", "  "); e != nil {
		t.Fatal(e)
	}
	varied := append([]byte(" \n"), boundaryWire(bmb, spaced.Bytes(), amb, asb)...)
	other, e := Execute(varied)
	if e != nil {
		t.Fatal(e)
	}
	var format Report
	if e = json.Unmarshal(other, &format); e != nil {
		t.Fatal(e)
	}
	if format.Sources.Request == boundaryHash(raw) || format.Sources.BeforeSuite == boundaryHash(bsb) || format.Sources.AfterSuite != boundaryHash(asb) {
		t.Fatal("formatting source identity")
	}
	format.Sources = nil
	if !reflect.DeepEqual(got, format) {
		t.Fatal("source formatting changed semantic comparison")
	}
	if bytes.Contains(out, []byte("edited")) || bytes.Contains(out, []byte("NeverExecute")) {
		t.Fatal("literal payload leaked into report")
	}
}

func TestStrictCarrierAndEmbeddedRejections(t *testing.T) {
	m, s := boundaryInputs(t)
	mb, sb := boundaryBytes(t, m, s)
	good := boundaryWire(mb, sb, mb, sb)
	var object map[string]json.RawMessage
	if e := json.Unmarshal(good, &object); e != nil {
		t.Fatal(e)
	}
	bad := [][]byte{nil, []byte(`null`), []byte(`[]`), append(bytes.Clone(good), []byte(` {}`)...), bytes.Replace(good, []byte(FormatVersion), []byte("unknown"), 1), append([]byte(`{"formatVersion":"duplicate",`), good[1:]...)}
	for _, key := range []string{"formatVersion", "beforeModel", "beforeSuite", "afterModel", "afterSuite"} {
		for _, mode := range []string{"missing", "null", "case"} {
			copy := map[string]json.RawMessage{}
			for k, v := range object {
				copy[k] = v
			}
			switch mode {
			case "missing":
				delete(copy, key)
			case "null":
				copy[key] = []byte(`null`)
			case "case":
				delete(copy, key)
				copy[strings.ToUpper(key)] = object[key]
			}
			raw, e := json.Marshal(copy)
			if e != nil {
				t.Fatal(e)
			}
			bad = append(bad, raw)
		}
	}
	bad = append(bad,
		bytes.Replace(good, []byte(`"beforeModel":`), []byte(`"extra":0,"beforeModel":`), 1),
		boundaryWire(mb, bytes.Replace(sb, []byte(`"actor":`), []byte(`"actor":"duplicate","actor":`), 1), mb, sb),
		boundaryWire(mb, bytes.Replace(sb, []byte(`"actor":`), []byte(`"ACTOR":`), 1), mb, sb),
		boundaryWire(mb, bytes.Replace(sb, []byte(`"expectedRevision":"18446744073709551615"`), []byte(`"expectedRevision":"18446744073709551616"`), 1), mb, sb),
		boundaryWire(mb, bytes.Replace(sb, []byte(`"NeverExecute"`), []byte(`"\ud800"`), 1), mb, sb),
		boundaryWire(mb, bytes.Replace(sb, []byte(`"NeverExecute"`), []byte{'"', 0xff, '"'}, 1), mb, sb),
		boundaryWire(mb, bytes.Replace(sb, []byte(s.ModelDigest), []byte(strings.Repeat("0", 64)), 1), mb, sb),
		boundaryWire([]byte(strings.Repeat("[", 67)+strings.Repeat("]", 67)), sb, mb, sb),
		boundaryWire(append(append([]byte("{"), bytes.Repeat([]byte{' '}, 1<<20)...), mb[1:]...), sb, mb, sb),
		boundaryWire(mb, append(append([]byte("{"), bytes.Repeat([]byte{' '}, 1<<20)...), sb[1:]...), mb, sb),
	)
	for i, raw := range bad {
		if bytes.Equal(raw, good) {
			t.Fatalf("rejection fixture %d unchanged", i)
		}
		out, e := Execute(raw)
		if e != ErrInvalid || out != nil {
			t.Errorf("%d got %d bytes / %v", i, len(out), e)
		}
	}
	if out, e := Execute(make([]byte, MaxRequestBytes+1)); e != ErrLimit || out != nil {
		t.Fatal("request limit", e)
	}
	padded := append(bytes.Clone(good), bytes.Repeat([]byte{' '}, MaxRequestBytes-len(good))...)
	if _, e := Execute(padded); e != nil {
		t.Fatal("exact wrapper limit", e)
	}
}

func FuzzExecuteNeverReturnsPartialOutput(f *testing.F) {
	f.Add([]byte(`{}`))
	f.Add([]byte(`{"formatVersion":"lifecycle-scenario-diff-json/v1"}`))
	f.Add([]byte("\xff"))
	f.Fuzz(func(t *testing.T, raw []byte) {
		if len(raw) > 2048 {
			t.Skip()
		}
		out, e := Execute(raw)
		if e != nil && out != nil {
			t.Fatal("partial output")
		}
		if e == nil && !json.Valid(out) {
			t.Fatal("invalid output")
		}
	})
}
