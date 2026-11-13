package dashboard

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/spik3r/clanker-foundry/modeling/diagram/contract"
)

var testGen = contract.Generator{ID: "diagramgen", Version: "1.2.0", Toolchain: "go1.27.4"}

const good = `{
 "formatVersion": "diagram-json/v1",
 "profileVersion": "dashboard/v1",
 "id": "WeeklyOps",
 "title": "Weekly operations",
 "subtitle": "Week 45",
 "datasets": [
  {"id": "Orders",
   "columns": [
    {"id": "Day", "label": "Day", "type": "text"},
    {"id": "Placed", "label": "Placed", "type": "number"},
    {"id": "Shipped", "label": "Shipped", "type": "number"}],
   "rows": [["Mon", 120, 96], ["Tue", 135, 130.5], ["Wed", -4, 0]]},
  {"id": "Aux",
   "columns": [{"id": "K", "label": "Key", "type": "text"}],
   "rows": []}
 ],
 "widgets": [
  {"id": "Total", "kind": "stat", "label": "Orders placed", "value": "872", "note": "+4%"},
  {"id": "ByDay", "kind": "bar", "title": "Orders by day", "dataset": "Orders", "category": "Day", "series": ["Placed", "Shipped"]},
  {"id": "Trend", "kind": "line", "title": "Shipped per day", "dataset": "Orders", "category": "Day", "series": ["Shipped"]},
  {"id": "Detail", "kind": "table", "title": "Daily detail", "dataset": "Orders", "columns": ["Day", "Placed", "Shipped"]}
 ],
 "layout": [["Total"], ["ByDay", "Trend"], ["Detail"]]
}`

// with replaces the single occurrence of old in good; it fails the test when
// old is absent so a stale fixture cannot hide a case.
func with(t testing.TB, old, repl string) string {
	t.Helper()
	if strings.Count(good, old) != 1 {
		t.Fatalf("fixture text %q occurs %d times", old, strings.Count(good, old))
	}
	return strings.Replace(good, old, repl, 1)
}

func TestGoodModelDecodes(t *testing.T) {
	m, err := Decode([]byte(good))
	if err != nil {
		t.Fatal(err)
	}
	if len(m.Datasets) != 2 || m.Datasets[0].ID != "Aux" || m.Widgets[0].ID != "ByDay" {
		t.Errorf("datasets and widgets are not sorted by id: %+v", m)
	}
	if got := m.dataset("Orders").Rows[1][2]; got.Num != "130.5" || got.Micro != 130_500_000 {
		t.Errorf("cell %+v", got)
	}
}

func TestStrictDecoding(t *testing.T) {
	deep := strings.Repeat("[", 20) + strings.Repeat("]", 20)
	cases := []struct{ name, raw string }{
		{"empty", ""},
		{"not an object", `[]`},
		{"trailing data", good + `{}`},
		{"trailing garbage", good + ` x`},
		{"duplicate top-level key", with(t, `"id": "WeeklyOps",`, `"id": "WeeklyOps", "id": "Other",`)},
		{"duplicate widget key", with(t, `"kind": "stat",`, `"kind": "stat", "kind": "stat",`)},
		{"unknown member", with(t, `"subtitle": "Week 45",`, `"subtitle": "Week 45", "colour": "red",`)},
		{"case variant", with(t, `"title": "Weekly operations"`, `"Title": "Weekly operations"`)},
		{"case variant of nested member", with(t, `"label": "Orders placed"`, `"Label": "Orders placed"`)},
		{"unknown widget member", with(t, `"note": "+4%"`, `"note": "+4%", "width": 2`)},
		{"null member", with(t, `"subtitle": "Week 45"`, `"subtitle": null`)},
		{"null cell", with(t, `["Mon", 120, 96]`, `["Mon", null, 96]`)},
		{"null row", with(t, `["Mon", 120, 96]`, `null`)},
		{"boolean cell", with(t, `["Mon", 120, 96]`, `["Mon", true, 96]`)},
		{"object cell", with(t, `["Mon", 120, 96]`, `["Mon", {}, 96]`)},
		{"array cell", with(t, `["Mon", 120, 96]`, `["Mon", [1], 96]`)},
		{"number for string", with(t, `"title": "Weekly operations"`, `"title": 7`)},
		{"string for array", with(t, `"series": ["Placed", "Shipped"]`, `"series": "Placed"`)},
		{"object for array", with(t, `"layout": [["Total"]`, `"layout": {"a": 1}, "x": [["Total"]`)},
		{"deep nesting", with(t, `"title": "Weekly operations"`, `"title": `+deep)},
		{"invalid UTF-8", with(t, `Weekly operations`, "Weekly \xff operations")},
		{"oversize", good + strings.Repeat(" ", contract.MaxModelBytes)},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := Decode([]byte(c.raw))
			if err == nil {
				t.Fatal("accepted")
			}
			t.Log(err)
		})
	}
	pad := good + strings.Repeat(" ", contract.MaxModelBytes-len(good))
	if _, err := Decode([]byte(pad)); err != nil {
		t.Errorf("model of exactly %d bytes refused: %v", contract.MaxModelBytes, err)
	}
}

func TestValidation(t *testing.T) {
	long := strings.Repeat("x", 121)
	cols := func(n int) string {
		var cs, row []string
		for i := 0; i < n; i++ {
			cs = append(cs, fmt.Sprintf(`{"id":"C%d","label":"C","type":"number"}`, i))
			row = append(row, "1")
		}
		return `{"id":"Wide","columns":[` + strings.Join(cs, ",") + `],"rows":[[` + strings.Join(row, ",") + `]]}`
	}
	rows := func(n int) string {
		r := make([]string, n)
		for i := range r {
			r[i] = `["a",1]`
		}
		return `{"id":"Big","columns":[{"id":"K","label":"K","type":"text"},{"id":"V","label":"V","type":"number"}],"rows":[` + strings.Join(r, ",") + `]}`
	}
	nDatasets := func(n int) []string {
		var s []string
		for i := 0; i < n; i++ {
			s = append(s, fmt.Sprintf(`{"id":"D%d","columns":[{"id":"A","label":"A","type":"text"}]}`, i))
		}
		return s
	}
	// datasets swaps the whole datasets array for a different one.
	datasets := func(sets ...string) string {
		i := strings.Index(good, `"datasets": [`)
		j := strings.Index(good, `"widgets":`)
		return good[:i] + `"datasets": [` + strings.Join(sets, ",") + `],` + "\n " + good[j:]
	}
	chartOn := func(set string) string {
		s := datasets(set)
		return strings.Replace(s, `"dataset": "Orders", "category": "Day", "series": ["Placed", "Shipped"]`, `"dataset": "Big", "category": "K", "series": ["V"]`, 1)
	}
	cases := []struct{ name, raw string }{
		{"wrong formatVersion", with(t, `diagram-json/v1`, `diagram-json/v2`)},
		{"wrong profileVersion", with(t, `dashboard/v1`, `dashboard/v2`)},
		{"model id starts with digit", with(t, `"WeeklyOps"`, `"1Ops"`)},
		{"model id with hyphen", with(t, `"WeeklyOps"`, `"Weekly-Ops"`)},
		{"model id of 49 characters", with(t, `"WeeklyOps"`, `"A`+strings.Repeat("b", 48)+`"`)},
		{"empty title", with(t, `"Weekly operations"`, `""`)},
		{"blank title", with(t, `"Weekly operations"`, `"   "`)},
		{"title of 121 bytes", with(t, `"Weekly operations"`, `"`+long+`"`)},
		{"title with newline", with(t, `"Weekly operations"`, `"a\nb"`)},
		{"title with U+2028", with(t, `"Weekly operations"`, `"a b"`)},
		{"title with U+FFFF", with(t, `"Weekly operations"`, `"a￿b"`)},
		{"empty subtitle", with(t, `"Week 45"`, `""`)},
		{"subtitle with control", with(t, `"Week 45"`, `"a\tb"`)},
		{"9 datasets", datasets(nDatasets(9)...)},
		{"duplicate dataset id", datasets(`{"id":"Orders","columns":[{"id":"A","label":"A","type":"text"}]}`, `{"id":"Orders","columns":[{"id":"A","label":"A","type":"text"}]}`)},
		{"dataset id invalid", datasets(`{"id":"1x","columns":[{"id":"A","label":"A","type":"text"}]}`)},
		{"no columns", datasets(`{"id":"Orders","columns":[]}`)},
		{"13 columns", datasets(cols(13))},
		{"duplicate column id", datasets(`{"id":"Orders","columns":[{"id":"A","label":"A","type":"text"},{"id":"A","label":"B","type":"text"}]}`)},
		{"column id invalid", datasets(`{"id":"Orders","columns":[{"id":"a-b","label":"A","type":"text"}]}`)},
		{"column label empty", datasets(`{"id":"Orders","columns":[{"id":"A","label":"","type":"text"}]}`)},
		{"unknown column type", datasets(`{"id":"Orders","columns":[{"id":"A","label":"A","type":"date"}]}`)},
		{"201 rows", chartOn(rows(201))},
		{"row too short", with(t, `["Mon", 120, 96]`, `["Mon", 120]`)},
		{"row too long", with(t, `["Mon", 120, 96]`, `["Mon", 120, 96, 1]`)},
		{"text cell of 121 bytes", with(t, `"Mon"`, `"`+long+`"`)},
		{"text cell with control", with(t, `"Mon"`, `"a\u0001b"`)},
		{"number in text column", with(t, `["Mon", 120, 96]`, `[5, 120, 96]`)},
		{"string in number column", with(t, `["Mon", 120, 96]`, `["Mon", "120", 96]`)},
		{"exponent", with(t, `120,`, `1.2e2,`)},
		{"exponent upper case", with(t, `120,`, `1E2,`)},
		{"leading plus", with(t, `120,`, `+120,`)},
		{"magnitude 1e12", with(t, `120,`, `1000000000000,`)},
		{"negative magnitude 1e12", with(t, `120,`, `-1000000000000,`)},
		{"7 fractional digits", with(t, `120,`, `0.0000001,`)},
		{"16 significant digits", with(t, `120,`, `123456789012.3456,`)},
		{"no widgets", `{"formatVersion":"diagram-json/v1","profileVersion":"dashboard/v1","id":"A","title":"T","widgets":[],"layout":[]}`},
		{"duplicate widget id", with(t, `"id": "Trend"`, `"id": "ByDay"`)},
		{"widget id invalid", with(t, `"id": "Trend"`, `"id": "Tr end"`)},
		{"unknown widget kind", with(t, `"kind": "line"`, `"kind": "pie"`)},
		{"stat without value", with(t, `"value": "872", `, ``)},
		{"stat with empty note", with(t, `"note": "+4%"`, `"note": ""`)},
		{"stat with dataset", with(t, `"note": "+4%"`, `"dataset": "Orders"`)},
		{"stat with empty extra member", with(t, `"note": "+4%"`, `"title": ""`)},
		{"stat value too long", with(t, `"872"`, `"`+long+`"`)},
		{"bar without title", with(t, `"title": "Orders by day", `, ``)},
		{"bar with columns", with(t, `"series": ["Placed", "Shipped"]`, `"series": ["Placed"], "columns": ["Day"]`)},
		{"bar with unknown dataset", with(t, `"dataset": "Orders", "category": "Day", "series": ["Placed", "Shipped"]`, `"dataset": "Nope", "category": "Day", "series": ["Placed"]`)},
		{"bar with unknown category", with(t, `"category": "Day", "series": ["Placed", "Shipped"]`, `"category": "Nope", "series": ["Placed"]`)},
		{"bar with number category", with(t, `"category": "Day", "series": ["Placed", "Shipped"]`, `"category": "Placed", "series": ["Shipped"]`)},
		{"bar with text series", with(t, `"series": ["Placed", "Shipped"]`, `"series": ["Day"]`)},
		{"bar with unknown series", with(t, `"series": ["Placed", "Shipped"]`, `"series": ["Nope"]`)},
		{"bar with duplicate series", with(t, `"series": ["Placed", "Shipped"]`, `"series": ["Placed", "Placed"]`)},
		{"bar with no series", with(t, `"series": ["Placed", "Shipped"]`, `"series": []`)},
		{"bar with 5 series", `{"formatVersion":"diagram-json/v1","profileVersion":"dashboard/v1","id":"A","title":"T","datasets":[{"id":"D","columns":[{"id":"K","label":"K","type":"text"},{"id":"A","label":"A","type":"number"},{"id":"B","label":"B","type":"number"},{"id":"C","label":"C","type":"number"},{"id":"D","label":"D","type":"number"},{"id":"E","label":"E","type":"number"}],"rows":[["a",1,2,3,4,5]]}],"widgets":[{"id":"W","kind":"bar","title":"T","dataset":"D","category":"K","series":["A","B","C","D","E"]}],"layout":[["W"]]}`},
		{"line over an empty dataset", with(t, `"rows": [["Mon", 120, 96], ["Tue", 135, 130.5], ["Wed", -4, 0]]`, `"rows": []`)},
		{"chart over 61 rows", chartOn(rows(61))},
		{"table with no columns", with(t, `"columns": ["Day", "Placed", "Shipped"]`, `"columns": []`)},
		{"table with unknown column", with(t, `"columns": ["Day", "Placed", "Shipped"]`, `"columns": ["Day", "Nope"]`)},
		{"table with duplicate column", with(t, `"columns": ["Day", "Placed", "Shipped"]`, `"columns": ["Day", "Day"]`)},
		{"table with series", with(t, `"columns": ["Day", "Placed", "Shipped"]`, `"columns": ["Day"], "series": ["Placed"]`)},
		{"empty layout", with(t, `"layout": [["Total"], ["ByDay", "Trend"], ["Detail"]]`, `"layout": []`)},
		{"13 layout rows", with(t, `"layout": [["Total"], ["ByDay", "Trend"], ["Detail"]]`, `"layout": [["Total"],["ByDay"],["Trend"],["Detail"],[],[],[],[],[],[],[],[],[]]`)},
		{"empty layout row", with(t, `["Total"], ["ByDay"`, `["Total"], [], ["ByDay"`)},
		{"5 widgets in a row", with(t, `"layout": [["Total"], ["ByDay", "Trend"], ["Detail"]]`, `"layout": [["Total","ByDay","Trend","Detail","Total"]]`)},
		{"layout names unknown widget", with(t, `["Detail"]]`, `["Nope"]]`)},
		{"layout repeats a widget", with(t, `["Detail"]]`, `["Detail"], ["Total"]]`)},
		{"layout omits a widget", with(t, `["ByDay", "Trend"], ["Detail"]`, `["ByDay", "Trend"]`)},
	}
	// Twelve layout rows and twelve widgets per row limits are accepted.
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := Decode([]byte(c.raw))
			if err == nil {
				t.Fatal("accepted")
			}
			t.Log(err)
		})
	}
}

func TestLimitsAreAccepted(t *testing.T) {
	var cs, row []string
	for i := 0; i < 12; i++ {
		cs = append(cs, fmt.Sprintf(`{"id":"C%d","label":"C","type":"number"}`, i))
		row = append(row, "-0.000001")
	}
	rows := make([]string, 200)
	for i := range rows {
		rows[i] = "[" + strings.Join(row, ",") + "]"
	}
	var sets []string
	for i := 0; i < 8; i++ {
		sets = append(sets, fmt.Sprintf(`{"id":"D%d","columns":[%s],"rows":[%s]}`, i, strings.Join(cs, ","), strings.Join(rows, ",")))
	}
	var ws, lay []string
	for i := 0; i < 24; i++ {
		ws = append(ws, fmt.Sprintf(`{"id":"W%d","kind":"stat","label":"l","value":"v"}`, i))
	}
	for i := 0; i < 12; i++ {
		lay = append(lay, fmt.Sprintf(`["W%d","W%d"]`, 2*i, 2*i+1))
	}
	raw := `{"formatVersion":"diagram-json/v1","profileVersion":"dashboard/v1","id":"Big","title":"T","datasets":[` +
		strings.Join(sets, ",") + `],"widgets":[` + strings.Join(ws, ",") + `],"layout":[` + strings.Join(lay, ",") + `]}`
	if len(raw) > contract.MaxModelBytes {
		t.Fatalf("fixture is %d bytes, over the limit", len(raw))
	}
	if _, err := Decode([]byte(raw)); err != nil {
		t.Fatalf("maximal model refused: %v", err)
	}
	// 60 rows is the chart limit; 200 rows are fine for a table.
	if _, err := (Profiler{}).Generate([]byte(raw), testGen); err != nil {
		t.Fatal(err)
	}
	long := strings.Repeat("é", 60) // 120 bytes
	if _, err := Decode([]byte(with(t, `"Weekly operations"`, `"`+long+`"`))); err != nil {
		t.Errorf("120-byte title refused: %v", err)
	}
	if _, err := Decode([]byte(with(t, `"Mon"`, `""`))); err != nil {
		t.Errorf("empty text cell refused: %v", err)
	}
	if _, err := Decode([]byte(with(t, `"Mon"`, `"`+strings.Repeat("x", 120)+`"`))); err != nil {
		t.Errorf("120-byte text cell refused: %v", err)
	}
}

func TestCanonicalNumber(t *testing.T) {
	good := []struct {
		in, text string
		micro    int64
	}{
		{"0", "0", 0},
		{"-0", "0", 0},
		{"-0.0", "0", 0},
		{"0.000000", "0", 0},
		{"1.50", "1.5", 1_500_000},
		{"1.5", "1.5", 1_500_000},
		{"-1.500000", "-1.5", -1_500_000},
		{"100", "100", 100_000_000},
		{"100.0", "100", 100_000_000},
		{"0.000001", "0.000001", 1},
		{"-0.000001", "-0.000001", -1},
		{"0.5", "0.5", 500_000},
		{"999999999999", "999999999999", 999_999_999_999_000_000},
		{"-999999999999", "-999999999999", -999_999_999_999_000_000},
		{"999999999.999999", "999999999.999999", 999_999_999_999_999},
		{"123456789012", "123456789012", 123_456_789_012_000_000},
		{"12345678.123456", "12345678.123456", 12_345_678_123_456},
		{"1.1234560", "1.123456", 1_123_456},
	}
	for _, c := range good {
		text, micro, err := CanonicalNumber(c.in)
		if err != nil || text != c.text || micro != c.micro {
			t.Errorf("%q: got %q %d %v, want %q %d", c.in, text, micro, err, c.text, c.micro)
		}
		if FormatMicro(micro) != text {
			t.Errorf("%q: FormatMicro(%d) = %q, want %q", c.in, micro, FormatMicro(micro), text)
		}
	}
	bad := []string{"", "1e3", "1E3", "1e-3", "+1", ".5", "5.", "01", "-", "--1", "1.2.3", "0x10", "NaN", "Infinity",
		"1000000000000", "-1000000000000", "0.0000001", "1234567890123.5", "1234567.1234567",
		"1234567890.123456", "9999999999999999"}
	for _, in := range bad {
		if text, _, err := CanonicalNumber(in); err == nil {
			t.Errorf("%q accepted as %q", in, text)
		}
	}
}

func compact(t *testing.T, raw string) string {
	t.Helper()
	var b bytes.Buffer
	if err := json.Compact(&b, []byte(raw)); err != nil {
		t.Fatal(err)
	}
	return b.String()
}

func indent(t *testing.T, raw string) string {
	t.Helper()
	var b bytes.Buffer
	if err := json.Indent(&b, []byte(raw), "\t", "    "); err != nil {
		t.Fatal(err)
	}
	return b.String()
}

// digest returns the canonical digest of a model text.
func digest(t *testing.T, raw string) string {
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

func TestDigestInvariance(t *testing.T) {
	want := digest(t, good)
	same := map[string]string{
		"compact whitespace":     compact(t, good) + "\n",
		"indented whitespace":    indent(t, compact(t, good)),
		"number 130.5 as 130.50": with(t, `130.5`, `130.50`),
		"number 120 as 120.0":    with(t, `120,`, `120.0,`),
		"negative zero":          with(t, `["Wed", -4, 0]`, `["Wed", -4, -0]`),
		"top-level member order": `{"title": "Weekly operations", "formatVersion": "diagram-json/v1", ` + strings.TrimPrefix(
			with(t, ` "title": "Weekly operations",`, ``), "{\n \"formatVersion\": \"diagram-json/v1\",")[0:],
	}
	// Swap the two datasets and two of the widgets textually.
	a := strings.Index(good, `{"id": "Orders",`)
	b := strings.Index(good, `{"id": "Aux",`)
	e := strings.Index(good, "\n ],\n \"widgets\"")
	same["datasets reordered"] = good[:a] + good[b:e] + ",\n  " + strings.TrimSuffix(strings.TrimSpace(good[a:b]), ",") + good[e:]
	same["widgets reordered"] = with(t, `{"id": "Total", "kind": "stat", "label": "Orders placed", "value": "872", "note": "+4%"},`, ``)
	same["widgets reordered"] = strings.Replace(same["widgets reordered"], `{"id": "Detail", "kind": "table"`, `{"id": "Total", "kind": "stat", "label": "Orders placed", "value": "872", "note": "+4%"},
  {"id": "Detail", "kind": "table"`, 1)
	same["member order inside widget"] = with(t, `{"id": "Total", "kind": "stat", "label": "Orders placed", "value": "872", "note": "+4%"}`,
		`{"note": "+4%", "value": "872", "label": "Orders placed", "kind": "stat", "id": "Total"}`)
	for name, raw := range same {
		if got := digest(t, raw); got != want {
			t.Errorf("%s changed the digest", name)
		}
	}
	// Source bytes differ, so the identity line's sourceSHA256 differs, but the
	// generated files do not.
	r1, err := Profiler{}.Generate([]byte(good), testGen)
	if err != nil {
		t.Fatal(err)
	}
	r2, err := Profiler{}.Generate([]byte(with(t, `130.5`, `130.50`)), testGen)
	if err != nil {
		t.Fatal(err)
	}
	for name, data := range r1.Files {
		if !bytes.Equal(data, r2.Files[name]) {
			t.Errorf("%s differs between 130.5 and 130.50", name)
		}
	}

	differ := map[string]string{
		"model id":        with(t, `"WeeklyOps"`, `"WeeklyOps2"`),
		"title":           with(t, `"Weekly operations"`, `"Weekly operation"`),
		"subtitle":        with(t, `"Week 45"`, `"Week 46"`),
		"subtitle absent": with(t, ` "subtitle": "Week 45",`, ``),
		"cell value":      with(t, `130.5`, `130.6`),
		"text cell":       with(t, `"Tue"`, `"Tuw"`),
		"row order":       with(t, `["Mon", 120, 96], ["Tue", 135, 130.5]`, `["Tue", 135, 130.5], ["Mon", 120, 96]`),
		"series order":    with(t, `"series": ["Placed", "Shipped"]`, `"series": ["Shipped", "Placed"]`),
		"table columns":   with(t, `"columns": ["Day", "Placed", "Shipped"]`, `"columns": ["Day", "Shipped", "Placed"]`),
		"layout order":    with(t, `["ByDay", "Trend"]`, `["Trend", "ByDay"]`),
		"layout rows":     with(t, `["Total"], ["ByDay", "Trend"]`, `["ByDay", "Trend"], ["Total"]`),
		"widget kind":     with(t, `"kind": "line"`, `"kind": "bar"`),
		"stat note":       with(t, `"+4%"`, `"+5%"`),
		"column label":    with(t, `"label": "Placed"`, `"label": "Placed "`),
		"column type":     with(t, `{"id": "K", "label": "Key", "type": "text"}`, `{"id": "K", "label": "Key", "type": "number"}`),
	}
	for name, raw := range differ {
		if got := digest(t, raw); got == want {
			t.Errorf("%s did not change the digest", name)
		}
	}
	// Column order is significant: swap two columns and their cells.
	swapped := strings.NewReplacer(
		`{"id": "Placed", "label": "Placed", "type": "number"},
    {"id": "Shipped", "label": "Shipped", "type": "number"}`,
		`{"id": "Shipped", "label": "Shipped", "type": "number"},
    {"id": "Placed", "label": "Placed", "type": "number"}`,
		`["Mon", 120, 96], ["Tue", 135, 130.5], ["Wed", -4, 0]`, `["Mon", 96, 120], ["Tue", 130.5, 135], ["Wed", 0, -4]`).Replace(good)
	if digest(t, swapped) == want {
		t.Errorf("column order did not change the digest")
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
	want := `{"formatVersion":"diagram-json/v1","profileVersion":"dashboard/v1","id":"WeeklyOps","title":"Weekly operations","subtitle":"Week 45",` +
		`"datasets":[{"id":"Aux","columns":[{"id":"K","label":"Key","type":"text"}],"rows":[]},` +
		`{"id":"Orders","columns":[{"id":"Day","label":"Day","type":"text"},{"id":"Placed","label":"Placed","type":"number"},{"id":"Shipped","label":"Shipped","type":"number"}],` +
		`"rows":[["Mon",120,96],["Tue",135,130.5],["Wed",-4,0]]}],` +
		`"widgets":[{"id":"ByDay","kind":"bar","title":"Orders by day","dataset":"Orders","category":"Day","series":["Placed","Shipped"]},` +
		`{"id":"Detail","kind":"table","title":"Daily detail","dataset":"Orders","columns":["Day","Placed","Shipped"]},` +
		`{"id":"Total","kind":"stat","label":"Orders placed","value":"872","note":"+4%"},` +
		`{"id":"Trend","kind":"line","title":"Shipped per day","dataset":"Orders","category":"Day","series":["Shipped"]}],` +
		`"layout":[["Total"],["ByDay","Trend"],["Detail"]]}`
	if string(c) != want {
		t.Errorf("canonical encoding\n got %s\nwant %s", c, want)
	}
	// Characters HTML would escape stay as written.
	h := with(t, `"Week 45"`, `"a <b> & c"`)
	hm, err := Decode([]byte(h))
	if err != nil {
		t.Fatal(err)
	}
	hc, _ := hm.Canonical()
	if !bytes.Contains(hc, []byte(`"a <b> & c"`)) {
		t.Errorf("canonical encoding escaped HTML characters: %s", hc)
	}
}

func TestModelJSONIsIndentedWithNewline(t *testing.T) {
	g, err := Profiler{}.Generate([]byte(good), testGen)
	if err != nil {
		t.Fatal(err)
	}
	mj := g.Files[FileModel]
	if !bytes.HasSuffix(mj, []byte("}\n")) || !bytes.HasPrefix(mj, []byte("{\n  \"formatVersion\"")) {
		t.Errorf("model.json is not indented with a trailing newline:\n%.80s", mj)
	}
	// model.json decodes to the same model.
	if digest(t, string(mj)) != digest(t, good) {
		t.Errorf("model.json does not round trip")
	}
}
