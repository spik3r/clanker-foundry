package dashboardbind_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/spik3r/clanker-foundry/modeling/diagram"
	"github.com/spik3r/clanker-foundry/modeling/diagram/contract"
	"github.com/spik3r/clanker-foundry/modeling/diagram/dashboard"
	"github.com/spik3r/clanker-foundry/modeling/diagram/dashboardbind"
)

const template = `{"formatVersion":"diagram-json/v1","profileVersion":"dashboard/v1","id":"Weekly","title":"Fixed title","subtitle":"Fixed subtitle","datasets":[{"id":"Work","columns":[{"id":"Day","label":"Day","type":"text"},{"id":"Count","label":"Count","type":"number"}],"rows":[["OLD",999]]}],"widgets":[{"id":"Total","kind":"stat","label":"Fixed label","value":"OLD","note":"OLD note"},{"id":"Chart","kind":"bar","title":"Fixed chart","dataset":"Work","category":"Day","series":["Count"]},{"id":"Table","kind":"table","title":"Fixed table","dataset":"Work","columns":["Day","Count"]}],"layout":[["Total","Chart"],["Table"]]}`

func sum(raw string) string { h := sha256.Sum256([]byte(raw)); return hex.EncodeToString(h[:]) }
func data(tpl string) map[string]any {
	return map[string]any{"dataVersion": dashboardbind.Version, "templateSHA256": sum(tpl), "datasets": []any{map[string]any{"id": "Work", "rows": [][]any{{"Mon", json.Number("1.5000000")}, {"Tue", json.Number("-0.0")}}}}, "stats": []any{map[string]any{"id": "Total", "value": "1.5"}}}
}
func encoded(t *testing.T, v any) []byte {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return b
}
func bind(t *testing.T, tpl string, p any) []byte {
	t.Helper()
	b, err := dashboardbind.Bind([]byte(tpl), encoded(t, p))
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func TestBindPreservesStructureAndReplacesAllData(t *testing.T) {
	raw := bind(t, template, data(template))
	m, err := dashboard.Decode(raw)
	if err != nil {
		t.Fatal(err)
	}
	if m.Title != "Fixed title" || m.Subtitle != "Fixed subtitle" || len(m.Layout) != 2 || m.Layout[0][1] != "Chart" {
		t.Fatalf("structure changed: %+v", m)
	}
	if m.Datasets[0].Columns[1].Label != "Count" || m.Datasets[0].Rows[0][0].Text != "Mon" || m.Datasets[0].Rows[0][1].Num != "1.5" || m.Datasets[0].Rows[1][1].Num != "0" {
		t.Fatalf("wrong replacement: %+v", m.Datasets)
	}
	for _, w := range m.Widgets {
		if w.Kind == dashboard.KindStat && (w.Label != "Fixed label" || w.Value != "1.5" || w.Note != "") {
			t.Fatalf("stat: %+v", w)
		}
	}
	if bytes.Contains(raw, []byte("OLD")) {
		t.Fatal("stale template data retained")
	}
	res, err := diagram.Build(raw)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(res.Files["dashboard.html"], []byte("Mon")) {
		t.Fatal("existing renderer did not receive data")
	}
}

func TestStrictDataRefusals(t *testing.T) {
	tests := map[string]func(map[string]any){
		"version":                 func(p map[string]any) { p["dataVersion"] = "other" },
		"pin":                     func(p map[string]any) { p["templateSHA256"] = strings.Repeat("0", 64) },
		"uppercase pin":           func(p map[string]any) { p["templateSHA256"] = strings.ToUpper(sum(template)) },
		"unknown layout override": func(p map[string]any) { p["layout"] = []any{} },
		"case alias":              func(p map[string]any) { p["Stats"] = p["stats"] },
		"missing datasets":        func(p map[string]any) { delete(p, "datasets") },
		"missing stats":           func(p map[string]any) { delete(p, "stats") },
		"null datasets":           func(p map[string]any) { p["datasets"] = nil },
		"missing dataset":         func(p map[string]any) { p["datasets"] = []any{} },
		"missing stat":            func(p map[string]any) { p["stats"] = []any{} },
		"extra dataset": func(p map[string]any) {
			p["datasets"] = append(p["datasets"].([]any), map[string]any{"id": "Other", "rows": []any{}})
		},
		"unknown dataset":  func(p map[string]any) { p["datasets"].([]any)[0].(map[string]any)["id"] = "Other" },
		"wrong stat id":    func(p map[string]any) { p["stats"].([]any)[0].(map[string]any)["id"] = "Chart" },
		"columns override": func(p map[string]any) { p["datasets"].([]any)[0].(map[string]any)["columns"] = []any{} },
		"rows missing":     func(p map[string]any) { delete(p["datasets"].([]any)[0].(map[string]any), "rows") },
		"rows null":        func(p map[string]any) { p["datasets"].([]any)[0].(map[string]any)["rows"] = nil },
		"row length":       func(p map[string]any) { p["datasets"].([]any)[0].(map[string]any)["rows"] = [][]any{{"x"}} },
		"boolean cell":     func(p map[string]any) { p["datasets"].([]any)[0].(map[string]any)["rows"] = [][]any{{"x", true}} },
		"string number":    func(p map[string]any) { p["datasets"].([]any)[0].(map[string]any)["rows"] = [][]any{{"x", "10"}} },
		"null cell":        func(p map[string]any) { p["datasets"].([]any)[0].(map[string]any)["rows"] = [][]any{{"x", nil}} },
		"exponent": func(p map[string]any) {
			p["datasets"].([]any)[0].(map[string]any)["rows"] = [][]any{{"x", json.Number("1e3")}}
		},
		"large number": func(p map[string]any) {
			p["datasets"].([]any)[0].(map[string]any)["rows"] = [][]any{{"x", json.Number("1000000000000")}}
		},
		"chart empty":        func(p map[string]any) { p["datasets"].([]any)[0].(map[string]any)["rows"] = []any{} },
		"stat value missing": func(p map[string]any) { delete(p["stats"].([]any)[0].(map[string]any), "value") },
		"stat value number":  func(p map[string]any) { p["stats"].([]any)[0].(map[string]any)["value"] = 1 },
		"stat note empty":    func(p map[string]any) { p["stats"].([]any)[0].(map[string]any)["note"] = "" },
		"control char":       func(p map[string]any) { p["stats"].([]any)[0].(map[string]any)["value"] = "a\nb" },
	}
	for name, mut := range tests {
		t.Run(name, func(t *testing.T) {
			p := data(template)
			mut(p)
			b, err := dashboardbind.Bind([]byte(template), encoded(t, p))
			if err == nil || len(b) != 0 {
				t.Fatalf("accepted: %s, %v", b, err)
			}
		})
	}
}

func TestDuplicateIDsAndKeys(t *testing.T) {
	tpl := strings.Replace(template, `"datasets":[`, `"datasets":[{"id":"Other","columns":[{"id":"C","label":"C","type":"text"}],"rows":[]},`, 1)
	p := data(tpl)
	p["datasets"] = append(p["datasets"].([]any), p["datasets"].([]any)[0])
	if _, err := dashboardbind.Bind([]byte(tpl), encoded(t, p)); err == nil {
		t.Fatal("duplicate dataset id accepted")
	}
	tpl = strings.Replace(template, `"widgets":[`, `"widgets":[{"id":"OtherStat","kind":"stat","label":"Other","value":"old"},`, 1)
	p = data(tpl)
	p["stats"] = append(p["stats"].([]any), p["stats"].([]any)[0])
	if _, err := dashboardbind.Bind([]byte(tpl), encoded(t, p)); err == nil {
		t.Fatal("duplicate stat id accepted")
	}
	raw := encoded(t, data(template))
	raw = bytes.Replace(raw, []byte(`"dataVersion":`), []byte(`"dataVersion":"dashboard-data/v1","dataVersion":`), 1)
	if _, err := dashboardbind.Bind([]byte(template), raw); err == nil {
		t.Fatal("duplicate key accepted")
	}
}

func TestDeterminismPinAndStrictInput(t *testing.T) {
	raw := encoded(t, data(template))
	want := bind(t, template, data(template))
	var pretty bytes.Buffer
	json.Indent(&pretty, raw, "", "   ")
	got, err := dashboardbind.Bind([]byte(template), pretty.Bytes())
	if err != nil || !bytes.Equal(got, want) {
		t.Fatal("data whitespace changed result", err)
	}
	for _, invalid := range [][]byte{append(raw, '{', '}'), append([]byte{0xff}, raw...), bytes.Repeat([]byte(" "), contract.MaxModelBytes+1), []byte("null")} {
		if _, err := dashboardbind.Bind([]byte(template), invalid); err == nil {
			t.Fatal("invalid data accepted")
		}
	}
	if _, err := dashboardbind.Bind([]byte(template+"\n"), raw); err == nil {
		t.Fatal("stale exact-source pin accepted")
	}
	if _, err := dashboardbind.Bind([]byte(strings.Replace(template, "dashboard/v1", "process-diagram/v1", 1)), raw); err == nil {
		t.Fatal("wrong profile accepted")
	}
	p := data(template)
	p["datasets"].([]any)[0].(map[string]any)["rows"] = [][]any{{"Mon", json.Number("1.5")}, {"Tue", 0}}
	if !bytes.Equal(bind(t, template, p), want) {
		t.Fatal("equivalent decimals differ")
	}
}

func TestReportTextIsEscapedByExistingRenderer(t *testing.T) {
	p := data(template)
	p["stats"].([]any)[0].(map[string]any)["value"] = `<script>alert(1)</script>`
	p["stats"].([]any)[0].(map[string]any)["note"] = `<img src=x onerror=alert(1)>`
	bound := bind(t, template, p)
	res, err := diagram.Build(bound)
	if err != nil {
		t.Fatal(err)
	}
	html := string(res.Files["dashboard.html"])
	if strings.Contains(html, "<script>") || strings.Contains(html, "<img ") || !strings.Contains(html, "&lt;script&gt;") {
		t.Fatal("unsafe output")
	}
}

func TestStatOnlyAndTableEmpty(t *testing.T) {
	tpl := `{"formatVersion":"diagram-json/v1","profileVersion":"dashboard/v1","id":"Stats","title":"Title","widgets":[{"id":"A","kind":"stat","label":"Value","value":"old"}],"layout":[["A"]]}`
	p := map[string]any{"dataVersion": dashboardbind.Version, "templateSHA256": sum(tpl), "datasets": []any{}, "stats": []any{map[string]any{"id": "A", "value": "new"}}}
	bind(t, tpl, p)
	tpl = strings.Replace(template, `{"id":"Chart","kind":"bar","title":"Fixed chart","dataset":"Work","category":"Day","series":["Count"]},`, "", 1)
	tpl = strings.Replace(tpl, `["Total","Chart"]`, `["Total"]`, 1)
	p = data(tpl)
	p["datasets"].([]any)[0].(map[string]any)["rows"] = []any{}
	out := bind(t, tpl, p)
	if bytes.Contains(out, []byte("OLD")) {
		t.Fatal("old row retained")
	}
}

func TestRepositoryExample(t *testing.T) {
	tpl, err := os.ReadFile("../examples/ops-dashboard.json")
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile("../examples/ops-report-data.json")
	if err != nil {
		t.Fatal(err)
	}
	out, err := dashboardbind.Bind(tpl, raw)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := diagram.Build(out); err != nil {
		t.Fatal(err)
	}
}

func TestTwoReportsReuseThemeAndStructure(t *testing.T) {
	tpl, err := os.ReadFile("../examples/ops-dashboard.json")
	if err != nil {
		t.Fatal(err)
	}
	var models []dashboard.Model
	var pages []string
	for _, name := range []string{"ops-report-data.json", "ops-report-data-two.json"} {
		raw, err := os.ReadFile("../examples/" + name)
		if err != nil {
			t.Fatal(err)
		}
		bound, err := dashboardbind.Bind(tpl, raw)
		if err != nil {
			t.Fatal(err)
		}
		m, err := dashboard.Decode(bound)
		if err != nil {
			t.Fatal(err)
		}
		models = append(models, m)
		res, err := diagram.Build(bound)
		if err != nil {
			t.Fatal(err)
		}
		pages = append(pages, string(res.Files["dashboard.html"]))
	}
	if len(models[0].Datasets[0].Rows) != 2 || len(models[1].Datasets[0].Rows) != 3 {
		t.Fatal("example replacement counts differ")
	}
	if models[0].Datasets[0].Rows[0][0].Text != "Mon" || models[1].Datasets[0].Rows[0][0].Text != "Wed" {
		t.Fatal("examples not independently bound")
	}
	if !bytes.Equal(encoded(t, models[0].Layout), encoded(t, models[1].Layout)) || !bytes.Equal(encoded(t, models[0].Datasets[0].Columns), encoded(t, models[1].Datasets[0].Columns)) {
		t.Fatal("structure changed")
	}
	style := func(page string) string {
		_, rest, ok := strings.Cut(page, "<style>")
		if !ok {
			t.Fatal("missing style")
		}
		css, _, ok := strings.Cut(rest, "</style>")
		if !ok {
			t.Fatal("missing closing style")
		}
		return css
	}
	if style(pages[0]) != style(pages[1]) || !strings.Contains(style(pages[0]), "prefers-color-scheme") {
		t.Fatal("shared theme changed")
	}
	if pages[0] == pages[1] {
		t.Fatal("distinct reports unexpectedly identical")
	}
	for _, page := range pages {
		for _, kind := range []string{`data-kind="stat"`, `data-kind="bar"`, `data-kind="line"`, `data-kind="table"`} {
			if !strings.Contains(page, kind) {
				t.Fatalf("missing shared component %s", kind)
			}
		}
	}
}

func TestBoundariesDoNotSilentlyTruncateData(t *testing.T) {
	for _, n := range []int{60, 61} {
		p := data(template)
		rows := make([][]any, n)
		for i := range rows {
			rows[i] = []any{"Day", 1}
		}
		p["datasets"].([]any)[0].(map[string]any)["rows"] = rows
		out, err := dashboardbind.Bind([]byte(template), encoded(t, p))
		if n == 60 {
			if err != nil {
				t.Fatal(err)
			}
			m, _ := dashboard.Decode(out)
			if len(m.Datasets[0].Rows) != 60 {
				t.Fatal("rows truncated")
			}
		} else if err == nil {
			t.Fatal("chart row limit ignored")
		}
	}
	p := data(template)
	p["stats"].([]any)[0].(map[string]any)["value"] = strings.Repeat("x", 121)
	if _, err := dashboardbind.Bind([]byte(template), encoded(t, p)); err == nil {
		t.Fatal("text bound ignored")
	}
}
