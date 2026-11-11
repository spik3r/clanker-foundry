package tui

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
)

func exampleBytes(t *testing.T) []byte {
	t.Helper()
	raw, err := os.ReadFile("examples/browser.json")
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func exampleModel(t *testing.T) Model {
	t.Helper()
	m, err := Decode(exampleBytes(t))
	if err != nil {
		t.Fatalf("decode example: %v", err)
	}
	return m
}

func TestDecodeExampleAndOrdering(t *testing.T) {
	m := exampleModel(t)
	if m.ID != "CatalogBrowser" || m.Identity != "ID" || m.Summary != "Name" {
		t.Fatalf("unexpected example semantics: %+v", m)
	}
	if want := []string{"Name", "Status", "Description"}; !reflect.DeepEqual(m.Details, want) {
		t.Fatalf("details = %v, want %v", m.Details, want)
	}
	m.Details = []string{"Description", "ID", "Status", "Name"}
	raw, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	got, err := Decode(raw)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, m) {
		t.Fatalf("round trip changed model: got %+v, want %+v", got, m)
	}
}

func TestDecodeRejectsMalformedInput(t *testing.T) {
	raw := string(exampleBytes(t))
	cases := map[string]string{
		"empty": "", "whitespace": " \n\t", "array root": "[]", "string root": `"model"`, "null root": "null",
		"truncated": raw[:len(raw)/2], "trailing object": raw + `{}`, "trailing scalar": raw + `true`, "trailing junk": raw + `!`,
		"oversize": raw + strings.Repeat(" ", MaxModelBytes), "invalid UTF8": strings.Replace(raw, "Catalog", string([]byte{0xff}), 1),
		"unknown root":          strings.Replace(raw, `"title":`, `"command":"run", "title":`, 1),
		"unknown field":         strings.Replace(raw, `"label": "Identifier"`, `"label":"Identifier", "type":"string"`, 1),
		"unknown layout":        strings.Replace(raw, `"narrowWidth":`, `"direction":"horizontal", "narrowWidth":`, 1),
		"duplicate root":        strings.Replace(raw, `"title":`, `"title":"Other", "title":`, 1),
		"duplicate field":       strings.Replace(raw, `"label": "Identifier"`, `"label":"Other", "label":"Identifier"`, 1),
		"duplicate layout":      strings.Replace(raw, `"narrowWidth":`, `"narrowWidth":30, "narrowWidth":`, 1),
		"escaped duplicate key": strings.Replace(raw, `"title":`, `"ti\u0074le":"Other", "title":`, 1),
		"case variant key":      strings.Replace(raw, `"title":`, `"Title":`, 1),
		"case alias duplicate":  strings.Replace(raw, `"title":`, `"TITLE":"Other", "title":`, 1),
		"deep nesting":          strings.Repeat("[", 18) + "0" + strings.Repeat("]", 18),
	}
	for name, input := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := Decode([]byte(input)); err == nil {
				t.Fatal("accepted invalid input")
			}
		})
	}
}

func TestDecodeRequiresEveryMember(t *testing.T) {
	var root map[string]json.RawMessage
	if err := json.Unmarshal(exampleBytes(t), &root); err != nil {
		t.Fatal(err)
	}
	check := func(t *testing.T, value any) {
		t.Helper()
		raw, err := json.Marshal(value)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := Decode(raw); err == nil {
			t.Fatalf("accepted invalid model: %s", raw)
		}
	}
	for key, original := range root {
		t.Run("missing/"+key, func(t *testing.T) { delete(root, key); defer func() { root[key] = original }(); check(t, root) })
		t.Run("null/"+key, func(t *testing.T) {
			root[key] = json.RawMessage(`null`)
			defer func() { root[key] = original }()
			check(t, root)
		})
		t.Run("wrong type/"+key, func(t *testing.T) {
			root[key] = json.RawMessage(`true`)
			defer func() { root[key] = original }()
			check(t, root)
		})
	}
	for _, member := range []string{"fields", "layout"} {
		original := root[member]
		var object map[string]json.RawMessage
		if member == "fields" {
			var fields []map[string]json.RawMessage
			if err := json.Unmarshal(original, &fields); err != nil {
				t.Fatal(err)
			}
			object = fields[0]
		} else if err := json.Unmarshal(original, &object); err != nil {
			t.Fatal(err)
		}
		for key, value := range object {
			for _, kind := range []string{"missing", "null", "wrong type"} {
				t.Run(member+"/"+key+"/"+kind, func(t *testing.T) {
					defer func() { object[key] = value; root[member] = original }()
					switch kind {
					case "missing":
						delete(object, key)
					case "null":
						object[key] = json.RawMessage(`null`)
					default:
						object[key] = json.RawMessage(`false`)
					}
					var nested any = object
					if member == "fields" {
						nested = []any{object}
					}
					encoded, err := json.Marshal(nested)
					if err != nil {
						t.Fatal(err)
					}
					root[member] = encoded
					check(t, root)
				})
			}
		}
	}
}

func TestValidateRejectsInvalidSemantics(t *testing.T) {
	cases := map[string]func(*Model){
		"format": func(m *Model) { m.FormatVersion = "tui-json/v2" }, "profile": func(m *Model) { m.ProfileVersion = "other/v1" },
		"missing identity": func(m *Model) { m.Identity = "Missing" }, "missing summary": func(m *Model) { m.Summary = "Missing" },
		"missing detail": func(m *Model) { m.Details = []string{"Missing"} }, "duplicate details": func(m *Model) { m.Details = []string{"Name", "Name"} },
		"empty details": func(m *Model) { m.Details = nil }, "excess details": func(m *Model) { m.Details = make([]string, 17) },
		"empty fields": func(m *Model) { m.Fields = nil }, "duplicate fields": func(m *Model) { m.Fields = append(m.Fields, m.Fields[0]) },
		"excess fields": func(m *Model) {
			for len(m.Fields) < 17 {
				m.Fields = append(m.Fields, Field{ID: fmt.Sprintf("Field%d", len(m.Fields)), Label: "Field"})
			}
		},
	}
	for _, id := range []string{"", "private", "type", "A_B", "A-B", "Å", "1ID", "ID; panic(1)", "ID\nfunc Inject(){}", strings.Repeat("A", 49)} {
		cases["model id/"+id] = func(m *Model) { m.ID = id }
		cases["field id/"+id] = func(m *Model) { m.Fields[0].ID = id }
	}
	for _, value := range []string{"", strings.Repeat("x", 121), strings.Repeat("é", 61), "a\x00b", "a\x1bb", "a\nb", "a\tb", "a\u0085b", "a\u2028b", "a\u2029b", string([]byte{0xff})} {
		cases["title/"+value] = func(m *Model) { m.Title = value }
		cases["list label/"+value] = func(m *Model) { m.ListLabel = value }
		cases["detail label/"+value] = func(m *Model) { m.DetailLabel = value }
		cases["field label/"+value] = func(m *Model) { m.Fields[0].Label = value }
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			m := exampleModel(t)
			mutate(&m)
			if err := m.Validate(); err == nil {
				t.Fatal("accepted invalid model")
			}
			raw, err := json.Marshal(m)
			if err != nil {
				t.Fatal(err)
			}
			// encoding/json replaces invalid UTF-8; test that case through Validate only.
			if strings.Contains(name, string([]byte{0xff})) {
				return
			}
			if _, err := Decode(raw); err == nil {
				t.Fatal("Decode bypassed semantic validation")
			}
		})
	}
}

func TestValidateLayoutBounds(t *testing.T) {
	cases := map[string]Layout{
		"narrow width low": {23, 6, 80, 12, 28}, "narrow width high": {241, 6, 300, 12, 28},
		"narrow height low": {24, 5, 80, 12, 28}, "narrow height high": {24, 81, 80, 100, 28},
		"wide width low": {40, 6, 39, 12, 8}, "wide width high": {24, 6, 401, 12, 28},
		"wide height low": {24, 12, 80, 11, 28}, "wide height high": {24, 6, 80, 121, 28},
		"list width low": {24, 6, 80, 12, 7}, "detail width low": {24, 6, 80, 12, 72},
	}
	for name, layout := range cases {
		t.Run(name, func(t *testing.T) {
			m := exampleModel(t)
			m.Layout = layout
			if err := m.Validate(); err == nil {
				t.Fatal("accepted invalid layout")
			}
		})
	}
	for _, layout := range []Layout{{24, 6, 24, 6, 8}, {240, 80, 400, 120, 391}, {24, 6, 80, 12, 71}} {
		m := exampleModel(t)
		m.Layout = layout
		if err := m.Validate(); err != nil {
			t.Fatalf("rejected boundary %+v: %v", layout, err)
		}
	}
}

func TestDecodeAllowsBoundedUnicodeAndWhitespace(t *testing.T) {
	m := exampleModel(t)
	m.Title = strings.Repeat("é", 60)
	m.Fields[0].Label = `Quote " and slash \`
	m.ID = strings.Repeat("A", 48)
	raw, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	raw = append(raw, []byte(strings.Repeat(" ", MaxModelBytes-len(raw)))...)
	if _, err := Decode(raw); err != nil {
		t.Fatalf("rejected valid boundary: %v", err)
	}
}

func TestDecodeRejectsInvalidArrayMembersAndNumbers(t *testing.T) {
	raw := string(exampleBytes(t))
	for name, input := range map[string]string{
		"null field":       strings.Replace(raw, `{"id": "ID", "label": "Identifier"}`, `null`, 1),
		"scalar field":     strings.Replace(raw, `{"id": "ID", "label": "Identifier"}`, `17`, 1),
		"null detail":      strings.Replace(raw, `["Name", "Status", "Description"]`, `["Name", null]`, 1),
		"numeric detail":   strings.Replace(raw, `["Name", "Status", "Description"]`, `["Name", 17]`, 1),
		"fractional width": strings.Replace(raw, `"narrowWidth": 24`, `"narrowWidth": 24.5`, 1),
		"overflow width":   strings.Replace(raw, `"narrowWidth": 24`, `"narrowWidth": 999999999999999999999999999`, 1),
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := Decode([]byte(input)); err == nil {
				t.Fatal("accepted invalid input")
			}
		})
	}
}

func TestDecodeAllowsMaximumFieldCount(t *testing.T) {
	m := exampleModel(t)
	for len(m.Fields) < 16 {
		m.Fields = append(m.Fields, Field{ID: fmt.Sprintf("Field%d", len(m.Fields)), Label: "Field"})
	}
	m.Details = nil
	for _, field := range m.Fields {
		m.Details = append(m.Details, field.ID)
	}
	raw, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	got, err := Decode(raw)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(m, got) {
		t.Fatal("maximum-sized model changed during decoding")
	}
}
