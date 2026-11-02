package jsonformat_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/jsonformat"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/model"
	"os"
	"reflect"
	"strings"
	"testing"
)

func example(t *testing.T) (model.Model, []byte) {
	t.Helper()
	b, e := os.ReadFile("../examples/lifecycle.json")
	if e != nil {
		t.Fatal(e)
	}
	m, e := jsonformat.Decode(b)
	if e != nil {
		t.Fatal(e)
	}
	return m, b
}
func TestRoundtripAndStrictFields(t *testing.T) {
	m, b := example(t)
	out, e := jsonformat.Encode(m)
	if e != nil {
		t.Fatal(e)
	}
	n, e := jsonformat.Decode(out)
	if e != nil || !reflect.DeepEqual(m, n) {
		t.Fatalf("roundtrip: %v", e)
	}
	for _, key := range []string{"formatVersion", "profileVersion", "id", "initial", "states", "requirements", "guards", "transitions"} {
		for _, null := range []bool{false, true} {
			var d map[string]any
			json.Unmarshal(b, &d)
			if null {
				d[key] = nil
			} else {
				delete(d, key)
			}
			v, _ := json.Marshal(d)
			if _, e := jsonformat.Decode(v); e == nil {
				t.Fatalf("accepted missing/null %s", key)
			}
		}
	}
	for _, v := range [][]byte{append(append([]byte(nil), b...), []byte(` {}`)...), bytes.Replace(b, []byte(`"formatVersion"`), []byte(`"FormatVersion"`), 1), bytes.Replace(b, []byte(`"lifecycle-json/v1"`), []byte(`"lifecycle-json/v2"`), 1), append([]byte(`{"extra":1,`), bytes.TrimSpace(b)[1:]...)} {
		if _, e := jsonformat.Decode(v); e == nil {
			t.Fatal("invalid source accepted")
		}
	}
}
func TestMaximumEscapedLiteralRoundtrip(t *testing.T) {
	m, _ := example(t)
	m.Guards[0].Expr = model.Expr{Kind: "op", Op: "eq", Args: []model.Expr{{Kind: "ref", Ref: "task.outcome"}, {Kind: "string", Value: strings.Repeat("\x01", 4096)}}}
	b, e := jsonformat.Encode(m)
	if e != nil {
		t.Fatal(e)
	}
	if !bytes.Contains(b, []byte(`\u0001`)) {
		t.Fatal("missing escaping")
	}
	n, e := jsonformat.Decode(b)
	want, _ := model.Normalize(m)
	if e != nil || !reflect.DeepEqual(n, want) {
		t.Fatalf("literal roundtrip: %v", e)
	}
}
func TestUnrepresentableIRRefusesEncoding(t *testing.T) {
	m := model.Model{Version: 1, ID: "Large", Initial: "Draft", States: []model.State{{ID: "Draft", Label: "Draft"}, {ID: "Done", Label: "Done", Terminal: true}}}
	tr := model.Transition{ID: "T_finish", From: "Draft", To: "Done", Trigger: "finish", Effects: []model.Effect{{Op: model.SetState}, {Op: model.IncrementRevision}}}
	for i := range 64 {
		r, g := fmt.Sprintf("R_%d", i), fmt.Sprintf("G_%d", i)
		m.Requirements = append(m.Requirements, model.Requirement{ID: r, Text: strings.Repeat("\x01", 4096)})
		m.Guards = append(m.Guards, model.Guard{ID: g, Requirement: r, Expr: model.Expr{Kind: "op", Op: "nonempty", Args: []model.Expr{{Kind: "ref", Ref: "task.outcome"}}}})
		tr.Guards = append(tr.Guards, g)
	}
	m.Transitions = []model.Transition{tr}
	if e := model.Validate(m); e != nil {
		t.Fatal(e)
	}
	if b, e := jsonformat.Encode(m); e == nil || len(b) != 0 {
		t.Fatalf("unrepresentable model emitted: %d %v", len(b), e)
	}
}
