package tesserae

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

const fixture = `{"formatVersion":"diagram-json/v1","profileVersion":"dashboard/v1","id":"Demo","title":"Synthetic <example>","datasets":[{"id":"Rows","columns":[{"id":"Name","label":"Name","type":"text"},{"id":"Count","label":"Count","type":"number"}],"rows":[["<script>not markup</script>",1.230000],["Empty",0]]}],"widgets":[{"id":"A","kind":"stat","label":"Count","value":"-0.000000","note":"Example"},{"id":"Z","kind":"table","title":"Records","dataset":"Rows","columns":["Count","Name"]}],"layout":[["Z"],["A"]]}`

func TestMappingAndDeterminism(t *testing.T) {
	a, err := Export([]byte(fixture))
	if err != nil {
		t.Fatal(err)
	}
	b, err := Export([]byte(fixture))
	if err != nil || !bytes.Equal(a, b) {
		t.Fatalf("nondeterministic %v", err)
	}
	var got struct {
		Version    string
		Components []struct {
			Type    string
			Value   json.Number
			Rows    [][]string
			Columns []string
		}
	}
	if err := json.Unmarshal(a, &got); err != nil {
		t.Fatal(err)
	}
	if got.Version != "tesserae/v1" || len(got.Components) != 2 || got.Components[0].Type != "table" || got.Components[1].Value != "0" {
		t.Fatalf("bad mapping: %s", a)
	}
	table := got.Components[0]
	if table.Columns[0] != "Count" || table.Rows[0][0] != "1.23" || table.Rows[0][1] != "<script>not markup</script>" {
		t.Fatalf("lost table semantics: %s", a)
	}
	if !bytes.Contains(a, []byte(`"category":"Generated dashboard"`)) || bytes.Contains(a, []byte("Synthetic demonstration data")) {
		t.Fatal("unsupported provenance label")
	}
	if !bytes.HasSuffix(a, []byte("\n")) {
		t.Fatal("missing newline")
	}
}
func TestRefusals(t *testing.T) {
	for _, raw := range []string{
		strings.Replace(fixture, `"value":"-0.000000"`, `"value":"42%"`, 1),
		strings.Replace(fixture, `"value":"-0.000000"`, `"value":"1e3"`, 1),
		strings.Replace(fixture, `"value":"-0.000000"`, `"value":"unknown"`, 1),
		fixture + "{}",
		strings.Replace(fixture, `"title":"Synthetic <example>"`, `"title":"a","title":"b"`, 1),
		strings.Replace(fixture, `"kind":"stat","label":"Count","value":"-0.000000","note":"Example"`, `"kind":"bar","title":"Chart","dataset":"Rows","category":"Name","series":["Count"]`, 1),
	} {
		out, err := Export([]byte(raw))
		if err == nil || out != nil {
			t.Fatalf("expected atomic refusal, got %s %v", out, err)
		}
	}
}
func TestEmptyRows(t *testing.T) {
	raw := strings.Replace(fixture, `[["<script>not markup</script>",1.230000],["Empty",0]]`, `[]`, 1)
	out, err := Export([]byte(raw))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(out, []byte(`"rows":[]`)) {
		t.Fatalf("empty rows lost: %s", out)
	}
}
func TestTargetColumnLimit(t *testing.T) {
	var m map[string]any
	if err := json.Unmarshal([]byte(fixture), &m); err != nil {
		t.Fatal(err)
	}
	ds := m["datasets"].([]any)[0].(map[string]any)
	cols := []any{}
	selected := []string{}
	for _, id := range []string{"A", "B", "C", "D", "E", "F", "G", "H", "I"} {
		cols = append(cols, map[string]string{"id": id, "label": id, "type": "text"})
		selected = append(selected, id)
	}
	ds["columns"] = cols
	ds["rows"] = []any{}
	m["widgets"].([]any)[1].(map[string]any)["columns"] = selected
	raw, _ := json.Marshal(m)
	if _, err := Export(raw); err == nil || !strings.Contains(err.Error(), "8 selected columns") {
		t.Fatalf("want target limit: %v", err)
	}
}

func TestConsumerBlankHeadings(t *testing.T) {
	for _, field := range []string{`"title":"Synthetic <example>"`, `"label":"Count"`, `"title":"Records"`, `"label":"Name"`} {
		key := strings.Split(field, ":")[0]
		raw := strings.Replace(fixture, field, key+`:"\uFEFF"`, 1)
		out, err := Export([]byte(raw))
		if err == nil || out != nil {
			t.Fatalf("want target heading refusal for %s", field)
		}
	}
	raw := strings.Replace(fixture, `"title":"Synthetic <example>"`, `"title":"\uFEFFValid"`, 1)
	if _, err := Export([]byte(raw)); err != nil {
		t.Fatal(err)
	}
}

func TestExpandedOutputLimit(t *testing.T) {
	var m map[string]any
	if err := json.Unmarshal([]byte(fixture), &m); err != nil {
		t.Fatal(err)
	}
	ds := m["datasets"].([]any)[0].(map[string]any)
	cols := []any{}
	selected := []string{}
	row := []string{}
	for _, id := range []string{"A", "B", "C", "D", "E", "F", "G", "H"} {
		cols = append(cols, map[string]string{"id": id, "label": id, "type": "text"})
		selected = append(selected, id)
		row = append(row, strings.Repeat("x", 120))
	}
	rows := [][]string{}
	for i := 0; i < 200; i++ {
		rows = append(rows, row)
	}
	ds["columns"] = cols
	ds["rows"] = rows
	widgets := []any{}
	layout := [][]string{}
	for _, id := range []string{"One", "Two", "Three", "Four"} {
		widgets = append(widgets, map[string]any{"id": id, "kind": "table", "title": id, "dataset": "Rows", "columns": selected})
		layout = append(layout, []string{id})
	}
	m["widgets"] = widgets
	m["layout"] = layout
	raw, _ := json.Marshal(m)
	out, err := Export(raw)
	if err == nil || out != nil || !strings.Contains(err.Error(), "exceeds") {
		t.Fatalf("want output expansion refusal: %v", err)
	}
}
