package processview_test

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/processview"
)

// Literal expected behavior was written from the process-view contract and the
// invented model, before the view implementation. It is not renderer output.
type state struct {
	ID       string `json:"id"`
	Label    string `json:"label"`
	Terminal bool   `json:"terminal"`
}
type transition struct {
	ID      string   `json:"id"`
	From    string   `json:"from"`
	Trigger string   `json:"trigger"`
	To      string   `json:"to"`
	Guards  []string `json:"guards"`
	Effects []string `json:"effects"`
}
type guard struct {
	ID          string `json:"id"`
	Requirement string `json:"requirement"`
	Expression  string `json:"expression"`
}
type projection struct {
	FormatVersion string       `json:"formatVersion"`
	Mode          string       `json:"mode"`
	Authority     string       `json:"authority"`
	ModelDigest   string       `json:"modelDigest"`
	Initial       string       `json:"initial"`
	States        []state      `json:"states"`
	Transitions   []transition `json:"transitions"`
	Guards        []guard      `json:"guards"`
}
type expectation struct {
	Projection projection `json:"projection"`
	SVGText    []string   `json:"svgText"`
}

func fixture(t *testing.T) ([]byte, expectation) {
	t.Helper()
	source, err := os.ReadFile("testdata/model.json")
	if err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile("testdata/expected.json")
	if err != nil {
		t.Fatal(err)
	}
	var want expectation
	if err = json.Unmarshal(b, &want); err != nil {
		t.Fatal(err)
	}
	return source, want
}
func TestIndependentProjectionAndSVG(t *testing.T) {
	raw, want := fixture(t)
	original := bytes.Clone(raw)
	out, err := processview.Build(raw)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(raw, original) {
		t.Fatal("Build mutated source bytes")
	}
	if len(out.JSON) == 0 || len(out.SVG) == 0 {
		t.Fatal("missing projection artifact")
	}
	var got projection
	if err = json.Unmarshal(out.JSON, &got); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	// Additional contract fields (including semantics) are intentionally allowed.
	if !reflect.DeepEqual(got, want.Projection) {
		t.Fatalf("projection mismatch\ngot:  %+v\nwant: %+v", got, want.Projection)
	}
	dec := xml.NewDecoder(bytes.NewReader(out.SVG))
	var text []string
	rootSeen := false
	for {
		tok, e := dec.Token()
		if e == io.EOF {
			break
		}
		if e != nil {
			t.Fatalf("invalid SVG XML: %v", e)
		}
		switch v := tok.(type) {
		case xml.StartElement:
			if !rootSeen {
				if v.Name.Local != "svg" {
					t.Fatalf("root %q is not svg", v.Name.Local)
				}
				rootSeen = true
			}
			if v.Name.Local == "script" || v.Name.Local == "foreignObject" {
				t.Fatalf("executable SVG element %q", v.Name.Local)
			}
		case xml.CharData:
			if s := strings.TrimSpace(string(v)); s != "" {
				text = append(text, s)
			}
		}
	}
	if !rootSeen {
		t.Fatal("missing SVG root")
	}
	joined := strings.Join(text, " ")
	for _, line := range want.SVGText {
		if !strings.Contains(joined, line) {
			t.Errorf("SVG lacks visible text %q", line)
		}
	}
	// The expression contains XML metacharacters, which must survive decoding as
	// text; unescaped markup would fail parsing or lose the exact glossary string.
	literal := "eq(command.actor, \"<&>\")"
	if !strings.Contains(joined, literal) {
		t.Fatal("escaped literal lost or altered")
	}
	if bytes.Contains(out.SVG, []byte(literal)) {
		t.Fatal("raw XML metacharacters were not escaped")
	}
}
func reverse(a []any) {
	for i, j := 0, len(a)-1; i < j; i, j = i+1, j-1 {
		a[i], a[j] = a[j], a[i]
	}
}
func TestDeterministicAcrossSemanticSetReordering(t *testing.T) {
	raw, _ := fixture(t)
	first, err := processview.Build(raw)
	if err != nil {
		t.Fatal(err)
	}
	again, err := processview.Build(raw)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err = json.Unmarshal(raw, &m); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"states", "requirements", "guards", "transitions"} {
		reverse(m[key].([]any))
	}
	for _, v := range m["transitions"].([]any) {
		reverse(v.(map[string]any)["guards"].([]any))
	}
	reordered, err := json.MarshalIndent(m, "", " ")
	if err != nil {
		t.Fatal(err)
	}
	permuted, err := processview.Build(reordered)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(first.JSON, again.JSON) || !bytes.Equal(first.SVG, again.SVG) || !bytes.Equal(first.JSON, permuted.JSON) || !bytes.Equal(first.SVG, permuted.SVG) {
		t.Fatal("output changed for repeated or reordered equivalent source")
	}
}
func TestBooleanOperandSetsDoNotChangeOutput(t *testing.T) {
	raw, _ := fixture(t)
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatal(err)
	}
	guards := m["guards"].([]any)
	for _, v := range guards {
		g := v.(map[string]any)
		if g["id"] == "Ga" {
			g["expr"] = map[string]any{"op": "all", "args": []any{map[string]any{"op": "nonempty", "args": []any{map[string]any{"ref": "task.outcome"}}}, map[string]any{"op": "nonempty", "args": []any{map[string]any{"ref": "command.actor"}}}}}
		}
	}
	a, e := json.Marshal(m)
	if e != nil {
		t.Fatal(e)
	}
	for _, v := range guards {
		g := v.(map[string]any)
		if g["id"] == "Ga" {
			reverse(g["expr"].(map[string]any)["args"].([]any))
		}
	}
	b, e := json.Marshal(m)
	if e != nil {
		t.Fatal(e)
	}
	x, e := processview.Build(a)
	if e != nil {
		t.Fatal(e)
	}
	y, e := processview.Build(b)
	if e != nil {
		t.Fatal(e)
	}
	if !bytes.Equal(x.JSON, y.JSON) || !bytes.Equal(x.SVG, y.SVG) {
		t.Fatal("boolean semantic-set reorder changed output")
	}
}
func TestInvalidInputYieldsNoArtifacts(t *testing.T) {
	raw, _ := fixture(t)
	var m map[string]any
	if e := json.Unmarshal(raw, &m); e != nil {
		t.Fatal(e)
	}
	transitions := m["transitions"].([]any)
	transitions[0].(map[string]any)["to"] = "Missing"
	dangling, e := json.Marshal(m)
	if e != nil {
		t.Fatal(e)
	}
	unknown := bytes.Replace(raw, []byte("\"formatVersion\":"), []byte("\"extra\": true, \"formatVersion\":"), 1)
	duplicate := bytes.Replace(raw, []byte("\"initial\": \"Draft\""), []byte("\"initial\": \"Draft\", \"initial\": \"Ready\""), 1)
	wrongVersion := bytes.Replace(raw, []byte("lifecycle-json/v1"), []byte("lifecycle-json/v2"), 1)
	for _, c := range []struct {
		name  string
		input []byte
	}{
		{"empty", nil}, {"truncated", []byte("{\"x\":")}, {"duplicate", duplicate}, {"unknown field", unknown}, {"wrong version", wrongVersion}, {"dangling state", dangling}, {"too large", bytes.Repeat([]byte(" "), (1<<20)+1)},
	} {
		t.Run(c.name, func(t *testing.T) {
			in := bytes.Clone(c.input)
			out, err := processview.Build(in)
			if err == nil || len(out.JSON) != 0 || len(out.SVG) != 0 {
				t.Fatalf("invalid input returned artifacts or nil error: %v", err)
			}
			if !bytes.Equal(in, c.input) {
				t.Fatal("invalid input mutated")
			}
		})
	}
}

// A bounded, high-content model must remain complete and XML-safe, including
// long expression literals. The expected values are not derived from the view.
func TestLargeGuardGlossaryIsComplete(t *testing.T) {
	raw, _ := fixture(t)
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatal(err)
	}
	requirements := make([]any, 0, 64)
	guards := make([]any, 0, 64)
	guardIDs := make([]any, 0, 64)
	literal := strings.Repeat("<&", 1500)
	for i := 0; i < 64; i++ {
		rid := fmt.Sprintf("R%02d", i)
		gid := fmt.Sprintf("G%02d", i)
		requirements = append(requirements, map[string]any{"id": rid, "text": "Example requirement"})
		guards = append(guards, map[string]any{"id": gid, "requirement": rid, "expr": map[string]any{"op": "eq", "args": []any{map[string]any{"ref": "command.actor"}, map[string]any{"string": literal}}}})
		guardIDs = append(guardIDs, gid)
	}
	m["requirements"], m["guards"] = requirements, guards
	transitions := m["transitions"].([]any)
	transitions[0].(map[string]any)["guards"] = guardIDs
	transitions[1].(map[string]any)["guards"] = []any{"G00"}
	var input bytes.Buffer
	enc := json.NewEncoder(&input)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(m); err != nil {
		t.Fatal(err)
	}
	raw = input.Bytes()
	out, err := processview.Build(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(out.JSON) > processview.MaxOutputBytes || len(out.SVG) > processview.MaxOutputBytes {
		t.Fatal("output exceeds advertised bound")
	}
	var got projection
	if err = json.Unmarshal(out.JSON, &got); err != nil {
		t.Fatal(err)
	}
	if len(got.Guards) != 64 {
		t.Fatalf("got %d guards", len(got.Guards))
	}
	for _, g := range got.Guards {
		if g.Expression != `eq(command.actor, "`+literal+`")` {
			t.Fatalf("guard %s was truncated", g.ID)
		}
	}
	d := xml.NewDecoder(bytes.NewReader(out.SVG))
	for {
		_, err := d.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("invalid large SVG: %v", err)
		}
	}
}

func TestSVGRepresentationsOfControlAndUnicode(t *testing.T) {
	raw, _ := fixture(t)
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatal(err)
	}
	states := m["states"].([]any)
	literal := "Control\x00 line\n Ω <script>"
	states[0].(map[string]any)["label"] = literal
	raw, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	out, err := processview.Build(raw)
	if err != nil {
		t.Fatal(err)
	}
	var got projection
	if err = json.Unmarshal(out.JSON, &got); err != nil {
		t.Fatal(err)
	}
	found := false
	for _, s := range got.States {
		if s.Label == literal {
			found = true
		}
	}
	if !found {
		t.Fatal("JSON did not preserve state label")
	}
	d := xml.NewDecoder(bytes.NewReader(out.SVG))
	var text strings.Builder
	for {
		tok, e := d.Token()
		if e == io.EOF {
			break
		}
		if e != nil {
			t.Fatalf("invalid SVG XML: %v", e)
		}
		if x, ok := tok.(xml.StartElement); ok && x.Name.Local == "script" {
			t.Fatal("injected SVG script")
		}
		if x, ok := tok.(xml.CharData); ok {
			text.Write(x)
		}
	}
	for _, part := range []string{`\x00`, `\n`, `\u03a9`, `<script>`} {
		if !strings.Contains(text.String(), part) {
			t.Fatalf("missing escaped display text %q", part)
		}
	}
}

func TestDistinctEscapedLabelsAndModelID(t *testing.T) {
	raw, _ := fixture(t)
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatal(err)
	}
	states := m["states"].([]any)
	for _, pair := range []struct{ actual, literal, encoded, escaped string }{
		{"Ω", `\u03a9`, `\u03a9`, `\\u03a9`},
		{"a\nb", `a\nb`, `a\nb`, `a\\nb`},
	} {
		states[0].(map[string]any)["label"] = pair.actual
		states[2].(map[string]any)["label"] = pair.literal
		input, err := json.Marshal(m)
		if err != nil {
			t.Fatal(err)
		}
		out, err := processview.Build(input)
		if err != nil {
			t.Fatal(err)
		}
		var id struct {
			ModelID string `json:"modelId"`
		}
		if err = json.Unmarshal(out.JSON, &id); err != nil || id.ModelID != "Flow" {
			t.Fatalf("model id: %+v %v", id, err)
		}
		dec := xml.NewDecoder(bytes.NewReader(out.SVG))
		var lines []string
		for {
			token, e := dec.Token()
			if e == io.EOF {
				break
			}
			if e != nil {
				t.Fatal(e)
			}
			if v, ok := token.(xml.CharData); ok {
				lines = append(lines, string(v))
			}
		}
		visible := strings.Join(lines, " ")
		for _, want := range []string{"Model ID: Flow", "label: " + pair.encoded, "label: " + pair.escaped} {
			if !strings.Contains(visible, want) {
				t.Fatalf("missing distinct SVG text %q", want)
			}
		}
	}
}

func TestOversizedCanvasIsRefusedWithoutArtifacts(t *testing.T) {
	raw, _ := fixture(t)
	var m map[string]any
	if e := json.Unmarshal(raw, &m); e != nil {
		t.Fatal(e)
	}
	reqs := make([]any, 0, 64)
	guards := make([]any, 0, 64)
	ids := make([]any, 0, 64)
	for i := 0; i < 64; i++ {
		rid := fmt.Sprintf("R%02d", i)
		gid := fmt.Sprintf("G%02d", i)
		reqs = append(reqs, map[string]any{"id": rid, "text": "Requirement"})
		guards = append(guards, map[string]any{"id": gid, "requirement": rid, "expr": map[string]any{"op": "eq", "args": []any{map[string]any{"ref": "command.actor"}, map[string]any{"string": strings.Repeat("Ω", 2000)}}}})
		ids = append(ids, gid)
	}
	m["requirements"], m["guards"] = reqs, guards
	ts := m["transitions"].([]any)
	ts[0].(map[string]any)["guards"] = ids
	ts[1].(map[string]any)["guards"] = []any{"G00"}
	var b bytes.Buffer
	enc := json.NewEncoder(&b)
	enc.SetEscapeHTML(false)
	if e := enc.Encode(m); e != nil {
		t.Fatal(e)
	}
	out, e := processview.Build(b.Bytes())
	if e == nil || len(out.JSON) != 0 || len(out.SVG) != 0 {
		t.Fatalf("oversized canvas must refuse both outputs: %v", e)
	}
}
