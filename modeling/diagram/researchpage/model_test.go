package researchpage

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/spik3r/clanker-foundry/modeling/diagram/contract"
)

var testGen = contract.Generator{ID: "diagramgen", Version: "1.4.0", Toolchain: "go1.27.4"}

const good = `{
 "formatVersion": "diagram-json/v1",
 "profileVersion": "research-page/v1",
 "id": "EvidenceAtlas",
 "title": "Evidence Atlas",
 "purpose": "Compare evidence.",
 "status": "mixed",
 "class": "atlas-page",
 "bottomLine": {
  "finding": "The atlas records where evidence exists.",
  "confidence": "Exploratory",
  "limitation": "Sparse rows describe coverage.",
  "context": ["First context paragraph.", "Second context paragraph."]
 },
 "sections": [
  {"tab": "sources", "items": [
   {"kind": "heading", "title": "Sources", "description": "Where the data came from."},
   {"kind": "panel", "slot": "provenance"}]},
  {"tab": "explore", "items": [
   {"kind": "slot", "id": "coverage"},
   {"kind": "heading", "title": "Regime matrix", "description": "Empty cells mean unmeasured."},
   {"kind": "panel", "slot": "matrix"},
   {"kind": "prose", "paragraphs": ["Intro paragraph.", "Another paragraph."]},
   {"kind": "heading", "title": "No description"}]},
  {"tab": "method", "items": [
   {"kind": "methodPanel", "title": "Method and limitations", "description": "Honest notes.", "points": ["One.", "Two."]}]}
 ]
}`

// mut applies fn to the decoded good model and returns the JSON text.
func mut(t testing.TB, fn func(m map[string]any)) string {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal([]byte(good), &m); err != nil {
		t.Fatal(err)
	}
	fn(m)
	b, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

type obj = map[string]any

// sec returns the section of the decoded model with the given tab.
func sec(t testing.TB, m obj, tab string) obj {
	t.Helper()
	for _, s := range m["sections"].([]any) {
		if s.(obj)["tab"] == tab {
			return s.(obj)
		}
	}
	t.Fatalf("no section %s", tab)
	return nil
}

func item(t testing.TB, m obj, tab string, i int) obj {
	return sec(t, m, tab)["items"].([]any)[i].(obj)
}

func digest(t testing.TB, raw string) string {
	t.Helper()
	m, err := Decode([]byte(raw))
	if err != nil {
		t.Fatal(err)
	}
	c, err := m.Canonical()
	if err != nil {
		t.Fatal(err)
	}
	return contract.HexSum(c)
}

func TestGoodModelDecodes(t *testing.T) {
	m, err := Decode([]byte(good))
	if err != nil {
		t.Fatal(err)
	}
	if m.Sections[0].Tab != "explore" || m.Sections[1].Tab != "method" || m.Sections[2].Tab != "sources" {
		t.Errorf("sections are not in tab order: %v %v %v", m.Sections[0].Tab, m.Sections[1].Tab, m.Sections[2].Tab)
	}
	if got := strings.Join(m.SlotIDs(), ","); got != "coverage,matrix,provenance" {
		t.Errorf("slot ids %s", got)
	}
	if m.Sections[0].Items[4].Description != "" {
		t.Errorf("a heading without description gained one")
	}
}

func TestEveryStatusIsAccepted(t *testing.T) {
	for _, s := range []string{"supported", "mixed", "inconclusive", "experimental", "unavailable", "not-yet-run", "unknown"} {
		if _, err := Decode([]byte(mut(t, func(m obj) { m["status"] = s }))); err != nil {
			t.Errorf("status %q: %v", s, err)
		}
	}
	for _, s := range []string{"", "Mixed", "MIXED", "not_yet_run", "done", " mixed", "mixed "} {
		if _, err := Decode([]byte(mut(t, func(m obj) { m["status"] = s }))); err == nil {
			t.Errorf("status %q was accepted", s)
		}
	}
}

func TestMinimalModel(t *testing.T) {
	raw := `{"formatVersion":"diagram-json/v1","profileVersion":"research-page/v1","id":"A","title":"T","purpose":"P","status":"unknown",
	"bottomLine":{"finding":"F","confidence":"C","limitation":"L"},
	"sections":[{"tab":"method","items":[{"kind":"prose","paragraphs":["x"]}]}]}`
	g, err := Profiler{}.Generate([]byte(raw), testGen)
	if err != nil {
		t.Fatal(err)
	}
	if string(g.Files[FileSlots]) != "[]\n" {
		t.Errorf("slots.json %q", g.Files[FileSlots])
	}
	jsx := string(g.Files[FileJSX])
	if strings.Contains(jsx, "renderSlot") || strings.Contains(jsx, "Panel") || strings.Contains(jsx, "SectionHeader") {
		t.Errorf("unused helper or import emitted:\n%s", jsx)
	}
	if !strings.Contains(jsx, "import { ResearchBottomLine, ResearchPageShell } from \"./researchPageParts\";") || !strings.Contains(jsx, "class={props.class}") {
		t.Errorf("unexpected module:\n%s", jsx)
	}
}

func TestStrictDecoding(t *testing.T) {
	cases := map[string]string{
		"duplicate key":          strings.Replace(good, `"id": "EvidenceAtlas",`, `"id": "EvidenceAtlas", "id": "Again",`, 1),
		"duplicate nested key":   strings.Replace(good, `"confidence": "Exploratory",`, `"confidence": "Exploratory", "confidence": "x",`, 1),
		"case variant member":    strings.Replace(good, `"title": "Evidence Atlas"`, `"Title": "Evidence Atlas"`, 1),
		"case variant nested":    strings.Replace(good, `"finding"`, `"Finding"`, 1),
		"unknown member":         strings.Replace(good, `"status"`, `"theme": "dark", "status"`, 1),
		"unknown item member":    strings.Replace(good, `{"kind": "slot", "id": "coverage"}`, `{"kind": "slot", "id": "coverage", "onClick": "x"}`, 1),
		"unknown section member": strings.Replace(good, `{"tab": "method",`, `{"tab": "method", "label": "M",`, 1),
		"unknown bottomLine key": strings.Replace(good, `"confidence"`, `"html": "x", "confidence"`, 1),
		"trailing data":          good + "{}",
		"trailing garbage":       good + " x",
		"null member":            strings.Replace(good, `"Exploratory"`, `null`, 1),
		"null class":             strings.Replace(good, `"atlas-page"`, `null`, 1),
		"null items": strings.Replace(good, `"items": [
   {"kind": "slot", "id": "coverage"},`, `"items": null, "x": [
   {"kind": "slot", "id": "coverage"},`, 1),
		"number for string":        strings.Replace(good, `"Exploratory"`, `5`, 1),
		"object for string":        strings.Replace(good, `"Exploratory"`, `{}`, 1),
		"string for array":         strings.Replace(good, `"points": ["One.", "Two."]`, `"points": "One."`, 1),
		"string for sections":      strings.Replace(good, `"sections": [`, `"sections": "x", "y": [`, 1),
		"array for object":         strings.Replace(good, `"bottomLine": {`, `"bottomLine": [{`, 1),
		"invalid UTF-8":            strings.Replace(good, `Evidence Atlas`, "Evidence \xff Atlas", 1),
		"lone surrogate escape":    strings.Replace(good, `Evidence Atlas`, `Evidence \ud800 Atlas`, 1),
		"oversize":                 good + strings.Repeat(" ", contract.MaxModelBytes),
		"empty input":              "",
		"not an object":            `[]`,
		"deep nesting":             strings.Replace(good, `"Exploratory"`, strings.Repeat("[", 20)+strings.Repeat("]", 20), 1),
		"nested string in strings": strings.Replace(good, `"First context paragraph."`, `["x"]`, 1),
	}
	for name, raw := range cases {
		if _, err := Decode([]byte(raw)); err == nil {
			t.Errorf("%s: model was accepted", name)
		}
	}
	// A broken JSON error is never a panic.
	for _, raw := range []string{`{`, `{"id":`, `nul`, "\x00"} {
		if _, err := Decode([]byte(raw)); err == nil {
			t.Errorf("%q was accepted", raw)
		}
	}
}

func repeat(s string, n int) []any {
	out := make([]any, n)
	for i := range out {
		out[i] = s
	}
	return out
}

func TestValidationRules(t *testing.T) {
	bl := func(m obj) obj { return m["bottomLine"].(obj) }
	cases := []struct {
		name string
		fn   func(m obj)
	}{
		{"wrong formatVersion", func(m obj) { m["formatVersion"] = "diagram-json/v2" }},
		{"wrong profileVersion", func(m obj) { m["profileVersion"] = "research-page/v2" }},
		{"id lower-case start", func(m obj) { m["id"] = "evidenceAtlas" }},
		{"id digit start", func(m obj) { m["id"] = "1Atlas" }},
		{"id underscore", func(m obj) { m["id"] = "Evidence_Atlas" }},
		{"id hyphen", func(m obj) { m["id"] = "Evidence-Atlas" }},
		{"id of 49 characters", func(m obj) { m["id"] = "A" + strings.Repeat("b", 48) }},
		{"id empty", func(m obj) { m["id"] = "" }},
		{"id non-ASCII", func(m obj) { m["id"] = "Évidence" }},
		{"missing title", func(m obj) { delete(m, "title") }},
		{"empty title", func(m obj) { m["title"] = "" }},
		{"blank title", func(m obj) { m["title"] = "   " }},
		{"title of 121 bytes", func(m obj) { m["title"] = strings.Repeat("x", 121) }},
		{"title of 61 two-byte characters", func(m obj) { m["title"] = strings.Repeat("é", 61) }},
		{"newline in title", func(m obj) { m["title"] = "a\nb" }},
		{"tab in title", func(m obj) { m["title"] = "a\tb" }},
		{"U+2028 in title", func(m obj) { m["title"] = "a\u2028b" }},
		{"U+2029 in title", func(m obj) { m["title"] = "a\u2029b" }},
		{"U+FFFE in title", func(m obj) { m["title"] = "a\ufffeb" }},
		{"U+FFFF in title", func(m obj) { m["title"] = "a\uffffb" }},
		{"DEL in title", func(m obj) { m["title"] = "a\u007fb" }},
		{"missing purpose", func(m obj) { delete(m, "purpose") }},
		{"empty purpose", func(m obj) { m["purpose"] = "" }},
		{"purpose of 601 bytes", func(m obj) { m["purpose"] = strings.Repeat("x", 601) }},
		{"missing status", func(m obj) { delete(m, "status") }},
		{"empty class", func(m obj) { m["class"] = "" }},
		{"class with a space", func(m obj) { m["class"] = "a b" }},
		{"class with upper case", func(m obj) { m["class"] = "Atlas" }},
		{"class starting with a digit", func(m obj) { m["class"] = "1atlas" }},
		{"class with a quote", func(m obj) { m["class"] = `a"b` }},
		{"class of 65 characters", func(m obj) { m["class"] = "a" + strings.Repeat("b", 64) }},
		{"missing bottomLine", func(m obj) { delete(m, "bottomLine") }},
		{"missing finding", func(m obj) { delete(bl(m), "finding") }},
		{"empty finding", func(m obj) { bl(m)["finding"] = "" }},
		{"finding of 601 bytes", func(m obj) { bl(m)["finding"] = strings.Repeat("x", 601) }},
		{"missing confidence", func(m obj) { delete(bl(m), "confidence") }},
		{"confidence of 121 bytes", func(m obj) { bl(m)["confidence"] = strings.Repeat("x", 121) }},
		{"missing limitation", func(m obj) { delete(bl(m), "limitation") }},
		{"limitation of 601 bytes", func(m obj) { bl(m)["limitation"] = strings.Repeat("x", 601) }},
		{"empty context", func(m obj) { bl(m)["context"] = []any{} }},
		{"7 context paragraphs", func(m obj) { bl(m)["context"] = repeat("x", 7) }},
		{"empty context paragraph", func(m obj) { bl(m)["context"] = []any{""} }},
		{"context paragraph of 601 bytes", func(m obj) { bl(m)["context"] = []any{strings.Repeat("x", 601)} }},
		{"control character in context", func(m obj) { bl(m)["context"] = []any{"a\x01b"} }},
		{"no sections", func(m obj) { m["sections"] = []any{} }},
		{"missing sections", func(m obj) { delete(m, "sections") }},
		{"4 sections", func(m obj) {
			m["sections"] = append(m["sections"].([]any), obj{"tab": "summary", "items": []any{obj{"kind": "prose", "paragraphs": []any{"x"}}}})
		}},
		{"unknown tab", func(m obj) { sec(t, m, "method")["tab"] = "summary" }},
		{"tab in another case", func(m obj) { sec(t, m, "method")["tab"] = "Method" }},
		{"duplicate tab", func(m obj) { sec(t, m, "method")["tab"] = "explore" }},
		{"section without items", func(m obj) { sec(t, m, "method")["items"] = []any{} }},
		{"25 items", func(m obj) { sec(t, m, "method")["items"] = repeat("x", 25) }},
		{"unknown kind", func(m obj) { item(t, m, "explore", 0)["kind"] = "table" }},
		{"empty kind", func(m obj) { item(t, m, "explore", 0)["kind"] = "" }},
		{"slot without id", func(m obj) { delete(item(t, m, "explore", 0), "id") }},
		{"slot with empty id", func(m obj) { item(t, m, "explore", 0)["id"] = "" }},
		{"slot with a title", func(m obj) { item(t, m, "explore", 0)["title"] = "x" }},
		{"slot with an empty title", func(m obj) { item(t, m, "explore", 0)["title"] = "" }},
		{"slot with a slot member", func(m obj) { item(t, m, "explore", 0)["slot"] = "x" }},
		{"slot id upper-case start", func(m obj) { item(t, m, "explore", 0)["id"] = "Coverage" }},
		{"slot id digit start", func(m obj) { item(t, m, "explore", 0)["id"] = "1coverage" }},
		{"slot id with hyphen", func(m obj) { item(t, m, "explore", 0)["id"] = "cov-erage" }},
		{"slot id with underscore", func(m obj) { item(t, m, "explore", 0)["id"] = "cov_erage" }},
		{"slot id of 33 characters", func(m obj) { item(t, m, "explore", 0)["id"] = "a" + strings.Repeat("b", 32) }},
		{"slot id constructor", func(m obj) { item(t, m, "explore", 0)["id"] = "constructor" }},
		{"slot id toString", func(m obj) { item(t, m, "explore", 0)["id"] = "toString" }},
		{"slot id hasOwnProperty", func(m obj) { item(t, m, "explore", 0)["id"] = "hasOwnProperty" }},
		{"slot id valueOf", func(m obj) { item(t, m, "explore", 0)["id"] = "valueOf" }},
		{"duplicate slot id across slot and panel", func(m obj) { item(t, m, "explore", 0)["id"] = "matrix" }},
		{"duplicate slot id across tabs", func(m obj) { item(t, m, "sources", 1)["slot"] = "coverage" }},
		{"panel without slot", func(m obj) { delete(item(t, m, "explore", 2), "slot") }},
		{"panel with an id", func(m obj) { item(t, m, "explore", 2)["id"] = "x" }},
		{"panel with a class", func(m obj) { item(t, m, "explore", 2)["class"] = "x" }},
		{"panel slot with a space", func(m obj) { item(t, m, "explore", 2)["slot"] = "a b" }},
		{"heading without title", func(m obj) { delete(item(t, m, "explore", 1), "title") }},
		{"heading with empty description", func(m obj) { item(t, m, "explore", 1)["description"] = "" }},
		{"heading with description of 601 bytes", func(m obj) { item(t, m, "explore", 1)["description"] = strings.Repeat("x", 601) }},
		{"heading title of 121 bytes", func(m obj) { item(t, m, "explore", 1)["title"] = strings.Repeat("x", 121) }},
		{"heading with points", func(m obj) { item(t, m, "explore", 1)["points"] = []any{"x"} }},
		{"heading with empty points", func(m obj) { item(t, m, "explore", 1)["points"] = []any{} }},
		{"methodPanel without points", func(m obj) { delete(item(t, m, "method", 0), "points") }},
		{"methodPanel with empty points", func(m obj) { item(t, m, "method", 0)["points"] = []any{} }},
		{"methodPanel with 13 points", func(m obj) { item(t, m, "method", 0)["points"] = repeat("x", 13) }},
		{"methodPanel with an empty point", func(m obj) { item(t, m, "method", 0)["points"] = []any{"a", ""} }},
		{"methodPanel with a point of 601 bytes", func(m obj) { item(t, m, "method", 0)["points"] = []any{strings.Repeat("x", 601)} }},
		{"methodPanel with a blank point", func(m obj) { item(t, m, "method", 0)["points"] = []any{"  "} }},
		{"methodPanel with a newline in a point", func(m obj) { item(t, m, "method", 0)["points"] = []any{"a\nb"} }},
		{"methodPanel without title", func(m obj) { delete(item(t, m, "method", 0), "title") }},
		{"methodPanel with paragraphs", func(m obj) { item(t, m, "method", 0)["paragraphs"] = []any{"x"} }},
		{"prose without paragraphs", func(m obj) { delete(item(t, m, "explore", 3), "paragraphs") }},
		{"prose with empty paragraphs", func(m obj) { item(t, m, "explore", 3)["paragraphs"] = []any{} }},
		{"prose with 7 paragraphs", func(m obj) { item(t, m, "explore", 3)["paragraphs"] = repeat("x", 7) }},
		{"prose with a paragraph of 601 bytes", func(m obj) { item(t, m, "explore", 3)["paragraphs"] = []any{strings.Repeat("x", 601)} }},
		{"prose with a title", func(m obj) { item(t, m, "explore", 3)["title"] = "x" }},
	}
	if len(cases) < 20 {
		t.Fatal("too few cases")
	}
	for _, c := range cases {
		if _, err := Decode([]byte(mut(t, c.fn))); err == nil {
			t.Errorf("%s: model was accepted", c.name)
		}
	}
}

func TestValidationLimitsAreInclusive(t *testing.T) {
	cases := map[string]func(m obj){
		"title of 120 bytes":      func(m obj) { m["title"] = strings.Repeat("x", 120) },
		"title of 60 two-byte":    func(m obj) { m["title"] = strings.Repeat("é", 60) },
		"purpose of 600 bytes":    func(m obj) { m["purpose"] = strings.Repeat("x", 600) },
		"confidence of 120 bytes": func(m obj) { m["bottomLine"].(obj)["confidence"] = strings.Repeat("x", 120) },
		"6 context paragraphs":    func(m obj) { m["bottomLine"].(obj)["context"] = repeat("x", 6) },
		"no context":              func(m obj) { delete(m["bottomLine"].(obj), "context") },
		"no class":                func(m obj) { delete(m, "class") },
		"class of 64 characters":  func(m obj) { m["class"] = "a" + strings.Repeat("b", 63) },
		"id of 48 characters":     func(m obj) { m["id"] = "A" + strings.Repeat("b", 47) },
		"id of one letter":        func(m obj) { m["id"] = "A" },
		"slot id of 32 chars":     func(m obj) { item(t, m, "explore", 0)["id"] = "a" + strings.Repeat("B", 31) },
		"slot id of one letter":   func(m obj) { item(t, m, "explore", 0)["id"] = "a" },
		"slot id camel case":      func(m obj) { item(t, m, "explore", 0)["id"] = "regimeMatrix2" },
		"slot id class":           func(m obj) { item(t, m, "explore", 0)["id"] = "class" },
		"12 points":               func(m obj) { item(t, m, "method", 0)["points"] = repeat("x", 12) },
		"point of 600 bytes":      func(m obj) { item(t, m, "method", 0)["points"] = []any{strings.Repeat("x", 600)} },
		"6 paragraphs":            func(m obj) { item(t, m, "explore", 3)["paragraphs"] = repeat("x", 6) },
		"24 items":                func(m obj) { sec(t, m, "method")["items"] = repeatItems(24) },
		"no description":          func(m obj) { delete(item(t, m, "method", 0), "description") },
		"non-ASCII text":          func(m obj) { m["title"] = "Évidence — 証拠 😀" },
	}
	for name, fn := range cases {
		if _, err := Decode([]byte(mut(t, fn))); err != nil {
			t.Errorf("%s: %v", name, err)
		}
	}
}

func repeatItems(n int) []any {
	out := make([]any, n)
	for i := range out {
		out[i] = obj{"kind": "prose", "paragraphs": []any{"x"}}
	}
	return out
}

func TestValidateAppliesToProgrammaticModels(t *testing.T) {
	m, err := Decode([]byte(good))
	if err != nil {
		t.Fatal(err)
	}
	if err := m.Validate(); err != nil {
		t.Fatal(err)
	}
	bad := m
	bad.Sections = append([]Section{}, m.Sections...)
	bad.Sections[0].Items = append([]Item{}, m.Sections[0].Items...)
	bad.Sections[0].Items[0] = Item{Kind: KindSlot, ID: "coverage", Title: "x"}
	if bad.Validate() == nil {
		t.Errorf("a slot item with a title was accepted")
	}
	bad.Sections[0].Items[0] = Item{Kind: KindPanel, Slot: "matrix"}
	if bad.Validate() == nil {
		t.Errorf("a duplicate slot id was accepted")
	}
	bad = m
	bad.Status = ""
	if bad.Validate() == nil {
		t.Errorf("an empty status was accepted")
	}
}

func TestCanonicalEncoding(t *testing.T) {
	m, err := Decode([]byte(good))
	if err != nil {
		t.Fatal(err)
	}
	c, err := m.Canonical()
	if err != nil {
		t.Fatal(err)
	}
	want := `{"formatVersion":"diagram-json/v1","profileVersion":"research-page/v1","id":"EvidenceAtlas","title":"Evidence Atlas","purpose":"Compare evidence.","status":"mixed","class":"atlas-page","bottomLine":{"finding":"The atlas records where evidence exists.","confidence":"Exploratory","limitation":"Sparse rows describe coverage.","context":["First context paragraph.","Second context paragraph."]},"sections":[{"tab":"explore","items":[{"kind":"slot","id":"coverage"},{"kind":"heading","title":"Regime matrix","description":"Empty cells mean unmeasured."},{"kind":"panel","slot":"matrix"},{"kind":"prose","paragraphs":["Intro paragraph.","Another paragraph."]},{"kind":"heading","title":"No description"}]},{"tab":"method","items":[{"kind":"methodPanel","title":"Method and limitations","description":"Honest notes.","points":["One.","Two."]}]},{"tab":"sources","items":[{"kind":"heading","title":"Sources","description":"Where the data came from."},{"kind":"panel","slot":"provenance"}]}]}`
	if string(c) != want {
		t.Errorf("canonical encoding\n got %s\nwant %s", c, want)
	}
	// Markup characters in text are not HTML-escaped.
	raw := mut(t, func(m obj) { m["title"] = "a<b>&c" })
	m2, _ := Decode([]byte(raw))
	c2, _ := m2.Canonical()
	if !bytes.Contains(c2, []byte(`"title":"a<b>&c"`)) {
		t.Errorf("canonical encoding escapes markup: %s", c2)
	}
	// model.json is the same model, indented, with a trailing newline.
	g, err := Profiler{}.Generate([]byte(good), testGen)
	if err != nil {
		t.Fatal(err)
	}
	pretty := g.Files[FileModel]
	if !bytes.HasSuffix(pretty, []byte("}\n")) || !bytes.HasPrefix(pretty, []byte("{\n  \"formatVersion\"")) {
		t.Errorf("model.json is not two-space indented with a trailing newline:\n%s", pretty)
	}
	var a, b any
	if json.Unmarshal(pretty, &a) != nil || json.Unmarshal(c, &b) != nil || fmt.Sprint(a) != fmt.Sprint(b) {
		t.Errorf("model.json and the canonical encoding differ")
	}
}

func TestDigestInvariance(t *testing.T) {
	want := digest(t, good)
	// JSON whitespace and member order do not change the digest.
	var m obj
	if err := json.Unmarshal([]byte(good), &m); err != nil {
		t.Fatal(err)
	}
	compact, _ := json.Marshal(m) // Go sorts map keys: a different member order
	indented, _ := json.MarshalIndent(m, "\t", "\t\t")
	for name, raw := range map[string]string{
		"compact, reordered members": string(compact),
		"indented":                   "\n\n" + string(indented) + "\n",
		"sections reordered": mut(t, func(m obj) {
			s := m["sections"].([]any)
			s[0], s[2] = s[2], s[0]
		}),
		"all section permutations": mut(t, func(m obj) {
			s := m["sections"].([]any)
			s[0], s[1], s[2] = s[1], s[2], s[0]
		}),
	} {
		if got := digest(t, raw); got != want {
			t.Errorf("%s changes the digest", name)
		}
	}
	// Section order does not change the output either.
	a := generated(t, good)
	b := generated(t, mut(t, func(m obj) { s := m["sections"].([]any); s[0], s[2] = s[2], s[0] }))
	for f, data := range a.Files {
		if !bytes.Equal(data, b.Files[f]) {
			t.Errorf("section order changes %s", f)
		}
	}
}

func TestDigestSensitivity(t *testing.T) {
	want := digest(t, good)
	edits := map[string]func(m obj){
		"id":              func(m obj) { m["id"] = "EvidenceAtlas2" },
		"title":           func(m obj) { m["title"] = "Evidence Atlas." },
		"purpose":         func(m obj) { m["purpose"] = "Compare evidence!" },
		"status":          func(m obj) { m["status"] = "supported" },
		"class":           func(m obj) { m["class"] = "atlas-page-2" },
		"class removed":   func(m obj) { delete(m, "class") },
		"finding":         func(m obj) { m["bottomLine"].(obj)["finding"] = "Changed." },
		"confidence":      func(m obj) { m["bottomLine"].(obj)["confidence"] = "Changed" },
		"limitation":      func(m obj) { m["bottomLine"].(obj)["limitation"] = "Changed." },
		"context text":    func(m obj) { m["bottomLine"].(obj)["context"].([]any)[1] = "Changed." },
		"context order":   func(m obj) { c := m["bottomLine"].(obj)["context"].([]any); c[0], c[1] = c[1], c[0] },
		"context removed": func(m obj) { delete(m["bottomLine"].(obj), "context") },
		"slot id":         func(m obj) { item(t, m, "explore", 0)["id"] = "coverage2" },
		"panel slot":      func(m obj) { item(t, m, "explore", 2)["slot"] = "matrix2" },
		"heading title":   func(m obj) { item(t, m, "explore", 1)["title"] = "Regime matrix 2" },
		"heading desc":    func(m obj) { item(t, m, "explore", 1)["description"] = "Changed." },
		"desc added":      func(m obj) { item(t, m, "explore", 4)["description"] = "Added." },
		"paragraph":       func(m obj) { item(t, m, "explore", 3)["paragraphs"].([]any)[0] = "Changed." },
		"paragraph order": func(m obj) { p := item(t, m, "explore", 3)["paragraphs"].([]any); p[0], p[1] = p[1], p[0] },
		"point":           func(m obj) { item(t, m, "method", 0)["points"].([]any)[1] = "Changed." },
		"point order":     func(m obj) { p := item(t, m, "method", 0)["points"].([]any); p[0], p[1] = p[1], p[0] },
		"item order": func(m obj) {
			it := sec(t, m, "explore")["items"].([]any)
			it[0], it[1] = it[1], it[0]
		},
		"item added": func(m obj) {
			s := sec(t, m, "method")
			s["items"] = append(s["items"].([]any), obj{"kind": "prose", "paragraphs": []any{"x"}})
		},
		"section removed": func(m obj) {
			m["sections"] = m["sections"].([]any)[:2]
		},
		"slot to panel": func(m obj) {
			it := item(t, m, "explore", 0)
			delete(it, "id")
			it["kind"] = "panel"
			it["slot"] = "coverage"
		},
	}
	seen := map[string]string{want: "the original"}
	for name, fn := range edits {
		got := digest(t, mut(t, fn))
		if prev, dup := seen[got]; dup {
			t.Errorf("%s gives the digest of %s", name, prev)
		}
		seen[got] = name
	}
}

func TestExplicitDefaultsAreNotAccepted(t *testing.T) {
	// Optional members that are present must not be empty; there is no
	// default spelling that equals an absent member.
	for name, fn := range map[string]func(m obj){
		"empty class":   func(m obj) { m["class"] = "" },
		"empty context": func(m obj) { m["bottomLine"].(obj)["context"] = []any{} },
	} {
		if _, err := Decode([]byte(mut(t, fn))); err == nil {
			t.Errorf("%s was accepted", name)
		}
	}
}

func generated(t testing.TB, raw string) contract.Generated {
	t.Helper()
	g, err := Profiler{}.Generate([]byte(raw), testGen)
	if err != nil {
		t.Fatal(err)
	}
	return g
}

func TestModelErrorsAreModelErrors(t *testing.T) {
	_, err := Profiler{}.Generate([]byte(`{}`), testGen)
	var me contract.ModelError
	if !errors.As(err, &me) {
		t.Errorf("error %v is not a ModelError", err)
	}
}
