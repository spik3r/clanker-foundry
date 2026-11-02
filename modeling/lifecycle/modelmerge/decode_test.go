package modelmerge

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/jsonformat"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/model"
)

func modelBytes(t *testing.T, m model.Model) []byte {
	t.Helper()
	raw, err := jsonformat.Encode(m)
	if err != nil {
		t.Fatal(err)
	}
	return bytes.TrimSpace(raw)
}
func wire(base, left, right []byte) []byte {
	return []byte(`{"formatVersion":"lifecycle-model-merge-json/v1","base":` + string(base) + `,"left":` + string(left) + `,"right":` + string(right) + `}`)
}
func digestBytes(b []byte) string { sum := sha256.Sum256(b); return hex.EncodeToString(sum[:]) }

type wireResult struct {
	MergeVersion    string          `json:"mergeVersion"`
	ProfileVersion  string          `json:"profileVersion"`
	Mode            string          `json:"mode"`
	Authority       string          `json:"authority"`
	BaseDigest      string          `json:"baseDigest"`
	LeftDigest      string          `json:"leftDigest"`
	RightDigest     string          `json:"rightDigest"`
	Sources         SourceDigests   `json:"sources"`
	Status          string          `json:"status"`
	CandidateDigest *string         `json:"candidateDigest"`
	Candidate       json.RawMessage `json:"candidate"`
	Conflicts       []Conflict      `json:"conflicts"`
}

func TestWireFrozenExpectations(t *testing.T) {
	for _, v := range oracleVectors() {
		t.Run(v.name, func(t *testing.T) {
			b, l, r := modelBytes(t, v.base), modelBytes(t, v.left), modelBytes(t, v.right)
			raw := wire(b, l, r)
			before := append([]byte(nil), raw...)
			out, err := Execute(raw)
			if err != nil {
				t.Fatal(err)
			}
			var got wireResult
			if err = json.Unmarshal(out, &got); err != nil {
				t.Fatal(err)
			}
			if got.Status != v.status || !reflect.DeepEqual(got.Conflicts, v.conflicts) || got.MergeVersion != Version || got.ProfileVersion != "lifecycle/v1" || got.Mode != "example-only" || got.Authority != "none" {
				t.Fatalf("wire envelope %+v", got)
			}
			if got.Sources != (SourceDigests{digestBytes(raw), digestBytes(b), digestBytes(l), digestBytes(r)}) {
				t.Fatal("source hashes", got.Sources)
			}
			if v.candidate == nil {
				if string(got.Candidate) != "null" || got.CandidateDigest != nil {
					t.Fatal("partial candidate exposed")
				}
			} else {
				candidate, e := jsonformat.Decode(got.Candidate)
				if e != nil || !reflect.DeepEqual(candidate, *v.candidate) || got.CandidateDigest == nil || *got.CandidateDigest != oracleCandidateDigest(v.name) {
					t.Fatal("candidate differs from frozen expectation", e)
				}
			}
			again, err := Execute(raw)
			if err != nil || !bytes.Equal(out, again) || !bytes.Equal(raw, before) {
				t.Fatal("nondeterministic or mutating execution")
			}
			if len(out) > MaxResultBytes || out[len(out)-1] != '\n' || bytes.Contains(out, []byte{'\n', '\n'}) {
				t.Fatal("wire encoding bounds/terminator")
			}
		})
	}
}

func TestSourceIdentityAndUnicodeOutput(t *testing.T) {
	m := oracleBase()
	m.Requirements[0].Text = "<review> 雪 \u2028"
	m.Guards[0].Expr = model.Expr{Kind: "op", Op: "eq", Args: []model.Expr{
		{Kind: "ref", Ref: "task.outcome"},
		{Kind: "string", Value: "<guard>& 雪 literal \\u003c"},
	}}
	original := modelBytes(t, m)
	var compact bytes.Buffer
	if err := json.Compact(&compact, original); err != nil {
		t.Fatal(err)
	}
	first := wire(original, original, original)
	second := append([]byte(" \n"), wire(compact.Bytes(), original, original)...)
	one, err := Execute(first)
	if err != nil {
		t.Fatal(err)
	}
	two, err := Execute(second)
	if err != nil {
		t.Fatal(err)
	}
	var a, b wireResult
	json.Unmarshal(one, &a)
	json.Unmarshal(two, &b)
	if a.BaseDigest != b.BaseDigest || a.LeftDigest != b.LeftDigest || a.RightDigest != b.RightDigest || *a.CandidateDigest != *b.CandidateDigest || a.Status != b.Status {
		t.Fatal("formatting changed merge semantics")
	}
	if a.Sources.Request == b.Sources.Request || a.Sources.Base == b.Sources.Base || a.Sources.Left != b.Sources.Left || a.Sources.Right != b.Sources.Right {
		t.Fatal("source hash distinction")
	}
	if !bytes.Contains(one, []byte("<review> 雪")) || !bytes.Contains(one, []byte("<guard>& 雪")) || bytes.Contains(one, []byte(`\u003cguard`)) || !bytes.Contains(one, []byte(`\u2028`)) {
		t.Fatal("unexpected Unicode/HTML encoding")
	}
	decoded, decodeErr := jsonformat.Decode(a.Candidate)
	if decodeErr != nil || !reflect.DeepEqual(decoded, m) {
		t.Fatal("candidate string literal changed while removing HTML escapes", decodeErr)
	}
}

func TestStrictWireRefusals(t *testing.T) {
	m := modelBytes(t, oracleBase())
	good := wire(m, m, m)
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(good, &obj); err != nil {
		t.Fatal(err)
	}
	invalid := [][]byte{nil, []byte(`null`), []byte(`[]`), append(append([]byte(nil), good...), []byte(` {}`)...), bytes.Replace(good, []byte(FormatVersion), []byte("unknown"), 1), append([]byte("{"), append([]byte(`"formatVersion":"duplicate",`), good[1:]...)...)}
	for _, key := range []string{"formatVersion", "base", "left", "right"} {
		for _, mode := range []string{"missing", "null", "case"} {
			copy := map[string]json.RawMessage{}
			for k, v := range obj {
				copy[k] = v
			}
			switch mode {
			case "missing":
				delete(copy, key)
			case "null":
				copy[key] = []byte(`null`)
			case "case":
				delete(copy, key)
				copy[strings.ToUpper(key)] = obj[key]
			}
			raw, e := json.Marshal(copy)
			if e != nil {
				t.Fatal(e)
			}
			invalid = append(invalid, raw)
		}
	}
	invalid = append(invalid,
		bytes.Replace(good, []byte(`"formatVersion":`), []byte(`"extra":0,"formatVersion":`), 1),
		bytes.Replace(good, []byte(`"label": "Draft"`), []byte(`"label": "\ud800"`), 1),
		bytes.Replace(good, []byte(`"label": "Draft"`), []byte{'"', 'l', 'a', 'b', 'e', 'l', '"', ':', '"', 0xff, '"'}, 1),
		bytes.Replace(good, []byte(`"terminal": false`), []byte(`"terminal": false,"terminal": false`), 1),
		wire([]byte(strings.Repeat("[", 67)+strings.Repeat("]", 67)), m, m),
		wire(append(append([]byte("{"), bytes.Repeat([]byte{' '}, 1<<20)...), m[1:]...), m, m),
	)
	for i, raw := range invalid {
		out, err := Execute(raw)
		if !errors.Is(err, ErrInvalid) || out != nil {
			t.Errorf("case %d: %d bytes / %v", i, len(out), err)
		}
	}
	if out, err := Execute(make([]byte, MaxRequestBytes+1)); !errors.Is(err, ErrLimit) || out != nil {
		t.Fatal("request limit", err)
	}
	atLimit := append(append([]byte(nil), good...), bytes.Repeat([]byte{' '}, MaxRequestBytes-len(good))...)
	if _, err := Execute(atLimit); err != nil {
		t.Fatal("exact request byte limit rejected", err)
	}
}
