package dashboard

import (
	"bytes"
	"encoding/xml"
	"math/rand"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

func render(t testing.TB, raw string) string {
	t.Helper()
	g, err := Profiler{}.Generate([]byte(raw), testGen)
	if err != nil {
		t.Fatal(err)
	}
	return string(g.Files[FileHTML])
}

func example(t testing.TB, name string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("..", "examples", name))
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

var exampleModels = []string{"ops-dashboard.json", "sales-dashboard.json", "minimal-dashboard.json"}

// --- a small tokenizer over the known output shape -------------------------

type tag struct {
	name  string
	close bool
	attrs map[string]string
}

var (
	tagRE   = regexp.MustCompile(`<(/?)([a-zA-Z][a-zA-Z0-9]*)((?:\s+[a-zA-Z][a-zA-Z0-9:-]*(?:="[^"<>]*")?)*)\s*(/?)>`)
	attrRE  = regexp.MustCompile(`\s+([a-zA-Z][a-zA-Z0-9:-]*)(?:="([^"]*)")?`)
	styleRE = regexp.MustCompile(`(?s)<style>.*?</style>`)
)

// tokens returns every tag outside the style element. Because all model text
// is escaped, every '<' outside <style> must begin one of these tags; any
// other '<' is reported.
func tokens(t testing.TB, out string) []tag {
	t.Helper()
	body := styleRE.ReplaceAllString(out, "<style></style>")
	var tags []tag
	rest := body
	for {
		i := strings.IndexByte(rest, '<')
		if i < 0 {
			break
		}
		if strings.HasPrefix(rest[i:], "<!doctype html>") {
			rest = rest[i+len("<!doctype html>"):]
			continue
		}
		loc := tagRE.FindStringSubmatchIndex(rest[i:])
		if loc == nil || loc[0] != 0 {
			t.Fatalf("unrecognised markup at %.60q", rest[i:])
		}
		m := tagRE.FindStringSubmatch(rest[i:])
		tg := tag{name: m[2], close: m[1] == "/", attrs: map[string]string{}}
		for _, a := range attrRE.FindAllStringSubmatch(m[3], -1) {
			if _, dup := tg.attrs[a[1]]; dup {
				t.Fatalf("duplicate attribute %s on <%s>", a[1], tg.name)
			}
			tg.attrs[a[1]] = a[2]
		}
		tags = append(tags, tg)
		rest = rest[i+loc[1]:]
	}
	return tags
}

var (
	allowedTags = map[string]bool{
		"html": true, "head": true, "meta": true, "title": true, "style": true, "body": true, "header": true,
		"h1": true, "h2": true, "p": true, "main": true, "section": true, "article": true, "ul": true, "li": true,
		"svg": true, "rect": true, "line": true, "text": true, "circle": true, "path": true,
		"div": true, "table": true, "thead": true, "tbody": true, "tr": true, "th": true, "td": true,
	}
	allowedAttrs = map[string]bool{
		"lang": true, "charset": true, "name": true, "content": true, "data-id": true, "data-kind": true, "data-v": true,
		"class": true, "viewBox": true, "role": true, "aria-label": true, "aria-hidden": true, "width": true, "height": true,
		"x": true, "y": true, "x1": true, "y1": true, "x2": true, "y2": true, "cx": true, "cy": true, "r": true, "d": true,
		"text-anchor": true, "scope": true, "tabindex": true,
	}
	forbiddenText = []string{"<script", "http", "//", "url(", "@import", "src=", "href=", "<link", "<base",
		"<iframe", "<form", "<img", "<object", "<embed", "javascript:", "expression(", "http-equiv", "srcset", "<foreignObject"}
)

func checkInert(t testing.TB, out string) {
	t.Helper()
	for _, tg := range tokens(t, out) {
		if !allowedTags[tg.name] {
			t.Errorf("tag <%s> is not allowed", tg.name)
		}
		for name := range tg.attrs {
			if !allowedAttrs[name] || strings.HasPrefix(strings.ToLower(name), "on") {
				t.Errorf("attribute %s on <%s> is not allowed", name, tg.name)
			}
		}
		if tg.name == "meta" && tg.attrs["name"] == "" && tg.attrs["charset"] == "" {
			t.Errorf("unexpected meta element %v", tg.attrs)
		}
	}
	if strings.Count(out, "<style>") != 1 || strings.Count(out, "</style>") != 1 {
		t.Errorf("want exactly one style element")
	}
	css := styleRE.FindString(out)
	for _, bad := range []string{"url(", "@import", "@font-face", "//", "http", "expression", "javascript"} {
		if strings.Contains(css, bad) {
			t.Errorf("style contains %q", bad)
		}
	}
}

func TestExamplesAreInert(t *testing.T) {
	for _, name := range exampleModels {
		t.Run(name, func(t *testing.T) {
			out := render(t, example(t, name))
			checkInert(t, out)
			for _, bad := range forbiddenText {
				if strings.Contains(out, bad) {
					t.Errorf("output contains %q", bad)
				}
			}
			if !strings.HasPrefix(out, "<!doctype html>\n<html lang=\"en\">") {
				t.Errorf("output does not start with a doctype")
			}
			if !strings.Contains(out, "prefers-color-scheme:dark") || !strings.Contains(out, "max-width:720px") {
				t.Errorf("output lacks the dark theme or the narrow-screen query")
			}
		})
	}
}

// hostile builds a model whose every text field is an attack string.
func hostile(t testing.TB) string {
	t.Helper()
	r := strings.NewReplacer(
		`"Weekly operations"`, `"<script>alert(1)</script>"`,
		`"Week 45"`, `"\"><img src=x onerror=alert(1)>"`,
		`"Orders placed"`, `"</style><b>&amp; Q&A"`,
		`"872"`, `"&lt;&#x3c;</style>"`,
		`"+4%"`, `"'\"><svg onload=alert(1)>"`,
		`"Orders by day"`, `"x\" onmouseover=\"alert(1)"`,
		`"label": "Placed"`, `"label": "<script>p</script>"`,
		`"Mon"`, `"</title><script>m</script>"`,
		`"Tue"`, `"\"><img src=x>"`,
		`"Detail"`, `"Detail"`,
	)
	return r.Replace(good)
}

func TestHostileStringsStayEscaped(t *testing.T) {
	raw := hostile(t)
	out := render(t, raw)
	checkInert(t, out)
	for _, bad := range []string{"<script", "<img", "<b>", "<svg onload", "\" onmouseover=\"", "</title><script"} {
		if strings.Contains(out, bad) {
			t.Errorf("output contains raw %q", bad)
		}
	}
	for _, want := range []string{
		"&lt;script&gt;alert(1)&lt;/script&gt;",
		"&#34;&gt;&lt;img src=x onerror=alert(1)&gt;",
		"&lt;/style&gt;&lt;b&gt;&amp;amp; Q&amp;A",
		"&amp;lt;&amp;#x3c;&lt;/style&gt;",
		"&lt;/title&gt;&lt;script&gt;m&lt;/script&gt;",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("output lacks escaped form %q", want)
		}
	}
	if strings.Count(out, "</style>") != 1 {
		t.Errorf("model text produced a style terminator")
	}
	// The hostile text also survives in the canonical model unchanged.
	g, _ := Profiler{}.Generate([]byte(raw), testGen)
	if !bytes.Contains(g.Files[FileModel], []byte(`<script>alert(1)</script>`)) {
		t.Errorf("model.json altered the title")
	}
	// Attribute values from the model (aria-label, data-id) are quoted and escaped.
	for _, tg := range tokens(t, out) {
		if tg.name == "svg" && strings.Contains(tg.attrs["aria-label"], `"`) {
			t.Errorf("aria-label holds a raw quote")
		}
	}
}

func TestStructure(t *testing.T) {
	m, err := Decode([]byte(example(t, "sales-dashboard.json")))
	if err != nil {
		t.Fatal(err)
	}
	out := render(t, example(t, "sales-dashboard.json"))
	kinds := map[string]string{}
	for _, w := range m.Widgets {
		kinds[w.ID] = w.Kind
	}
	var got []string
	var sections []int
	for _, tg := range tokens(t, out) {
		switch {
		case tg.name == "section" && !tg.close:
			sections = append(sections, 0)
			if !strings.HasPrefix(tg.attrs["class"], "row cols-") {
				t.Errorf("section class %q", tg.attrs["class"])
			}
			if n, _ := strconv.Atoi(strings.TrimPrefix(tg.attrs["class"], "row cols-")); n != len(m.Layout[len(sections)-1]) {
				t.Errorf("row %d has class %q for %d widgets", len(sections), tg.attrs["class"], len(m.Layout[len(sections)-1]))
			}
		case tg.name == "article" && !tg.close:
			got = append(got, tg.attrs["data-id"])
			sections[len(sections)-1]++
			if tg.attrs["data-kind"] != kinds[tg.attrs["data-id"]] {
				t.Errorf("article %s has data-kind %q", tg.attrs["data-id"], tg.attrs["data-kind"])
			}
		}
	}
	var want []string
	for i, row := range m.Layout {
		want = append(want, row...)
		if sections[i] != len(row) {
			t.Errorf("section %d holds %d articles, want %d", i+1, sections[i], len(row))
		}
	}
	if strings.Join(got, ",") != strings.Join(want, ",") || len(sections) != len(m.Layout) {
		t.Errorf("widgets in output order %v, want layout order %v", got, want)
	}
	for _, want := range []string{"<header>", "<h1>Annual sales review", `<p class="subtitle">`, "<main>", "<meta charset=\"utf-8\">",
		`<meta name="viewport" content="width=device-width, initial-scale=1">`, `<meta name="generator" content="diagramgen 1.2.1">`} {
		if !strings.Contains(out, want) {
			t.Errorf("output lacks %q", want)
		}
	}
	// A model without a subtitle emits no subtitle paragraph.
	if strings.Contains(render(t, example(t, "minimal-dashboard.json")), `class="subtitle"`) {
		t.Errorf("subtitle paragraph without a subtitle")
	}
}

func TestStatAndTable(t *testing.T) {
	out := render(t, good)
	if !strings.Contains(out, `<p class="stat-label">Orders placed</p>`) || !strings.Contains(out, `<p class="stat-value">872</p>`) || !strings.Contains(out, `<p class="note">+4%</p>`) {
		t.Errorf("stat markup missing:\n%s", out)
	}
	noNote := strings.Replace(good, `, "note": "+4%"`, ``, 1)
	if strings.Contains(render(t, noNote), `class="note"`) {
		t.Errorf("note paragraph without a note")
	}
	// Table: 3 data rows by 3 columns; numbers right aligned, text not.
	if n := strings.Count(out, "<td"); n != 9 {
		t.Errorf("%d td cells, want 9", n)
	}
	if n := strings.Count(out, "<th "); n != 3 {
		t.Errorf("%d th cells, want 3", n)
	}
	for _, want := range []string{`<thead>`, `<tbody>`, `<th scope="col">Day</th>`, `<th scope="col" class="num">Placed</th>`,
		`<td>Mon</td><td class="num">120</td><td class="num">96</td>`, `<td class="num">130.5</td>`, `<td class="num">-4</td>`,
		`<div class="scroll" role="region" tabindex="0" aria-label="Daily detail">`} {
		if !strings.Contains(out, want) {
			t.Errorf("table lacks %q", want)
		}
	}
	if !strings.Contains(out, "font-variant-numeric:tabular-nums") {
		t.Errorf("no tabular numerals")
	}
	// A reordered column list is honoured; only the listed columns appear.
	re := strings.Replace(good, `"columns": ["Day", "Placed", "Shipped"]`, `"columns": ["Shipped", "Day"]`, 1)
	o2 := render(t, re)
	if !strings.Contains(o2, `<td class="num">96</td><td>Mon</td>`) || strings.Contains(o2, "<th scope=\"col\" class=\"num\">Placed</th></tr>") {
		t.Errorf("table column order or selection wrong")
	}
	// The 12-column table in the sales example has 12 headers and 6 rows.
	sales := render(t, example(t, "sales-dashboard.json"))
	i := strings.Index(sales, `data-id="WideTable"`)
	wide := sales[i:]
	wide = wide[:strings.Index(wide, "</article>")]
	if strings.Count(wide, "<th ") != 12 || strings.Count(wide, "<td") != 72 {
		t.Errorf("wide table has %d th, %d td", strings.Count(wide, "<th "), strings.Count(wide, "<td"))
	}
}

// --- chart geometry, read back from the SVG --------------------------------

type mark struct {
	class string
	attr  map[string]string
	title string
	text  string
}

type chart struct {
	label string
	marks []mark
}

var svgRE = regexp.MustCompile(`(?s)<svg class="chart".*?</svg>`)

// parseCharts parses each chart's SVG fragment with encoding/xml.
func parseCharts(t testing.TB, out string) []chart {
	t.Helper()
	var charts []chart
	for _, frag := range svgRE.FindAllString(out, -1) {
		d := xml.NewDecoder(strings.NewReader(frag))
		var c chart
		var cur *mark
		for {
			tok, err := d.Token()
			if err != nil {
				break
			}
			switch e := tok.(type) {
			case xml.StartElement:
				attr := map[string]string{}
				for _, a := range e.Attr {
					attr[a.Name.Local] = a.Value
				}
				switch e.Name.Local {
				case "svg":
					c.label = attr["aria-label"]
					if attr["role"] != "img" || attr["viewBox"] != "0 0 640 320" {
						t.Errorf("svg attributes %v", attr)
					}
				case "title":
					// the title belongs to cur
				default:
					c.marks = append(c.marks, mark{class: attr["class"], attr: attr})
					cur = &c.marks[len(c.marks)-1]
					cur.attr["_el"] = e.Name.Local
				}
			case xml.EndElement:
				if e.Name.Local != "title" {
					cur = nil
				}
			case xml.CharData:
				if cur != nil {
					s := string(e)
					if cur.attr["_el"] == "text" {
						cur.text += s
					} else {
						cur.title += s
					}
				}
			}
		}
		charts = append(charts, c)
	}
	return charts
}

func num(t testing.TB, s string) float64 {
	t.Helper()
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		t.Fatalf("%q: %v", s, err)
	}
	return f
}

func (c chart) byClass(prefix string) []mark {
	var r []mark
	for _, m := range c.marks {
		if strings.HasPrefix(m.class, prefix) && m.attr["_el"] != "" {
			r = append(r, m)
		}
	}
	return r
}

// checkChart verifies one chart against the dataset it draws.
func checkChart(t *testing.T, c chart, w Widget, ds Dataset) {
	t.Helper()
	plots := c.byClass("plot")
	if len(plots) != 1 {
		t.Fatalf("%s: %d plot rects", w.ID, len(plots))
	}
	px, py := num(t, plots[0].attr["x"]), num(t, plots[0].attr["y"])
	pr, pb := px+num(t, plots[0].attr["width"]), py+num(t, plots[0].attr["height"])
	if px < 0 || py < 0 || pr > 640 || pb > 320 {
		t.Errorf("%s: plot %v outside the viewBox", w.ID, plots[0].attr)
	}
	// values to draw
	var vals []float64
	for _, s := range w.Series {
		col := ds.column(s)
		for _, row := range ds.Rows {
			vals = append(vals, float64(row[col].Micro)/1e6)
		}
	}
	min, max := vals[0], vals[0]
	for _, v := range vals {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}
	// ticks
	grid := c.byClass("grid")
	if len(grid) < 4 || len(grid) > 6 {
		t.Errorf("%s: %d gridlines, want 4..6", w.ID, len(grid))
	}
	var tv, ty []float64
	zeroY := -1.0
	for _, g := range grid {
		v, y := num(t, g.attr["data-v"]), num(t, g.attr["y1"])
		tv, ty = append(tv, v), append(ty, y)
		if y < py-0.001 || y > pb+0.001 {
			t.Errorf("%s: gridline %v outside the plot", w.ID, g.attr)
		}
		if v == 0 {
			zeroY = y
		}
	}
	for i := 1; i < len(tv); i++ {
		if tv[i] <= tv[i-1] || ty[i] >= ty[i-1] {
			t.Errorf("%s: ticks not increasing: %v at %v", w.ID, tv, ty)
		}
	}
	step := tv[1] - tv[0]
	for i := 1; i < len(tv); i++ {
		if d := tv[i] - tv[i-1]; d < step-1e-9 || d > step+1e-9 {
			t.Errorf("%s: uneven ticks %v", w.ID, tv)
		}
	}
	if !isNice(step) {
		t.Errorf("%s: step %v is not 1, 2 or 5 times a power of ten", w.ID, step)
	}
	if tv[0] > min+1e-9 || tv[len(tv)-1] < max-1e-9 {
		t.Errorf("%s: ticks %v do not cover data %v..%v", w.ID, tv, min, max)
	}
	if w.Kind == KindBar && (tv[0] > 0 || tv[len(tv)-1] < 0) {
		t.Errorf("%s: bar axis %v lacks zero", w.ID, tv)
	}
	if len(ty) > 0 && (ty[0] < pb-0.01 || ty[0] > pb+0.01 || ty[len(ty)-1] < py-0.01 || ty[len(ty)-1] > py+0.01) {
		t.Errorf("%s: axis ends %v..%v do not span the plot %v..%v", w.ID, ty[0], ty[len(ty)-1], pb, py)
	}
	// tick labels match the gridlines
	ticks := c.byClass("tick")
	if len(ticks) != len(grid) {
		t.Fatalf("%s: %d tick labels for %d gridlines", w.ID, len(ticks), len(grid))
	}
	for i, tk := range ticks {
		if num(t, tk.text) != tv[i] {
			t.Errorf("%s: tick label %q for value %v", w.ID, tk.text, tv[i])
		}
	}
	// zero baseline
	zeros := c.byClass("zero")
	if zeroY >= 0 {
		if len(zeros) != 1 || num(t, zeros[0].attr["y1"]) != zeroY {
			t.Errorf("%s: zero line %v does not sit at gridline y %v", w.ID, zeros, zeroY)
		}
	} else if len(zeros) != 0 {
		t.Errorf("%s: zero line drawn although 0 is outside the axis", w.ID)
	}
	yOfValue := func(v float64) float64 {
		return pb - (v-tv[0])/(tv[len(tv)-1]-tv[0])*(pb-py)
	}
	rows := len(ds.Rows)
	switch w.Kind {
	case KindBar:
		bars := c.byClass("b")
		if len(bars) != rows*len(w.Series) {
			t.Fatalf("%s: %d bars, want %d", w.ID, len(bars), rows*len(w.Series))
		}
		i := 0
		for s, sid := range w.Series {
			col := ds.column(sid)
			for r := 0; r < rows; r++ {
				b := bars[i]
				i++
				if b.class != "b"+strconv.Itoa(s+1) {
					t.Errorf("%s: bar %d has class %q, want series colour %d", w.ID, i, b.class, s+1)
				}
				x, y, bw, bh := num(t, b.attr["x"]), num(t, b.attr["y"]), num(t, b.attr["width"]), num(t, b.attr["height"])
				if x < px-0.001 || x+bw > pr+0.001 || y < py-0.001 || y+bh > pb+0.001 || bw <= 0 || bh < 0 {
					t.Errorf("%s: bar %v outside the plot %v,%v..%v,%v", w.ID, b.attr, px, py, pr, pb)
				}
				v := float64(ds.Rows[r][col].Micro) / 1e6
				if v >= 0 && (y+bh < zeroY-0.011 || y+bh > zeroY+0.011) {
					t.Errorf("%s: positive bar %v does not end on the baseline %v", w.ID, b.attr, zeroY)
				}
				if v < 0 && (y < zeroY-0.011 || y > zeroY+0.011) {
					t.Errorf("%s: negative bar %v does not start on the baseline %v", w.ID, b.attr, zeroY)
				}
				top := y
				if v < 0 {
					top = y + bh
				}
				if d := top - yOfValue(v); d < -0.02 || d > 0.02 {
					t.Errorf("%s: bar for %v ends at y %v, expected about %v", w.ID, v, top, yOfValue(v))
				}
				want := ds.Columns[col].Label + ", " + ds.Rows[r][ds.column(w.Category)].Text + ": " + ds.Rows[r][col].Num
				if b.title != want {
					t.Errorf("%s: bar title %q, want %q", w.ID, b.title, want)
				}
			}
		}
	case KindLine:
		pts := c.byClass("m")
		if len(pts) != rows*len(w.Series) {
			t.Fatalf("%s: %d points, want %d", w.ID, len(pts), rows*len(w.Series))
		}
		for _, p := range pts {
			x, y := num(t, p.attr["cx"]), num(t, p.attr["cy"])
			if x < px || x > pr || y < py-0.001 || y > pb+0.001 {
				t.Errorf("%s: point %v outside the plot", w.ID, p.attr)
			}
			if p.title == "" {
				t.Errorf("%s: point without a title", w.ID)
			}
		}
		paths := 0
		for _, m := range c.marks {
			if m.attr["_el"] == "path" {
				paths++
				if !strings.HasPrefix(m.attr["d"], "M") || strings.Count(m.attr["d"], "L") != rows-1 {
					t.Errorf("%s: path %q", w.ID, m.attr["d"])
				}
			}
		}
		if paths != len(w.Series) {
			t.Errorf("%s: %d paths for %d series", w.ID, paths, len(w.Series))
		}
	}
	// x labels never leave the viewBox and none is longer than the budget
	for _, l := range c.byClass("xl") {
		if x := num(t, l.attr["x"]); x < px || x > pr {
			t.Errorf("%s: x label at %v outside the plot", w.ID, x)
		}
		if n := len([]rune(l.text)); n > maxXChars {
			t.Errorf("%s: x label %q has %d characters", w.ID, l.text, n)
		}
	}
}

func isNice(step float64) bool {
	for step >= 10 {
		step /= 10
	}
	for step < 1 {
		step *= 10
		step = float64(int64(step*1e9+0.5)) / 1e9
	}
	for _, m := range []float64{1, 2, 5} {
		if step > m-1e-6 && step < m+1e-6 {
			return true
		}
	}
	return false
}

func TestChartGeometry(t *testing.T) {
	models := append(append([]string{}, example(t, "ops-dashboard.json"), example(t, "sales-dashboard.json")), good, longCategoryModel(), flatModel(), tinyModel())
	for _, raw := range models {
		m, err := Decode([]byte(raw))
		if err != nil {
			t.Fatal(err)
		}
		charts := parseCharts(t, render(t, raw))
		n := 0
		byID := map[string]Widget{}
		for _, w := range m.Widgets {
			byID[w.ID] = w
		}
		var inOrder []Widget
		for _, row := range m.Layout {
			for _, id := range row {
				inOrder = append(inOrder, byID[id])
			}
		}
		for _, w := range inOrder {
			if w.Kind != KindBar && w.Kind != KindLine {
				continue
			}
			if n >= len(charts) {
				t.Fatalf("fewer charts than chart widgets")
			}
			checkChart(t, charts[n], w, m.dataset(w.Dataset))
			if !strings.HasPrefix(charts[n].label, w.Title+", ") {
				t.Errorf("aria-label %q does not start with the title %q", charts[n].label, w.Title)
			}
			n++
		}
		if n != len(charts) {
			t.Errorf("%d charts in output, %d chart widgets", len(charts), n)
		}
	}
}

// chartModel builds a one-chart model from rows of "cat",v1,v2.
func chartModel(kind, rows string, series string) string {
	return `{"formatVersion":"diagram-json/v1","profileVersion":"dashboard/v1","id":"C","title":"T",
"datasets":[{"id":"D","columns":[{"id":"K","label":"K","type":"text"},{"id":"A","label":"A","type":"number"},{"id":"B","label":"B","type":"number"}],"rows":[` + rows + `]}],
"widgets":[{"id":"W","kind":"` + kind + `","title":"Chart","dataset":"D","category":"K","series":[` + series + `]}],"layout":[["W"]]}`
}

func longCategoryModel() string {
	var rows []string
	for i := 0; i < 60; i++ {
		rows = append(rows, `["A fairly long category label `+strconv.Itoa(i)+`",`+strconv.Itoa(i*37%101-40)+`.25,`+strconv.Itoa(i)+`]`)
	}
	return chartModel("bar", strings.Join(rows, ","), `"A","B"`)
}

func flatModel() string {
	return chartModel("line", `["a",7,0],["b",7,0],["c",7,0]`, `"A"`)
}

func tinyModel() string {
	return chartModel("bar", `["only",0.000001,-0.000002]`, `"A","B"`)
}

func TestXLabelsAreThinnedAndTruncated(t *testing.T) {
	charts := parseCharts(t, render(t, longCategoryModel()))
	labels := charts[0].byClass("xl")
	if len(labels) >= 60 || len(labels) < 5 {
		t.Errorf("%d x labels for 60 categories; expected them thinned", len(labels))
	}
	for _, l := range labels {
		if !strings.HasSuffix(l.text, "…") && len([]rune(l.text)) > 0 && l.text != "" {
			t.Errorf("label %q was not truncated", l.text)
		}
	}
	// Few categories with short names keep every label untouched.
	c2 := parseCharts(t, render(t, good))
	if got := len(c2[0].byClass("xl")); got != 3 {
		t.Errorf("%d labels for 3 categories", got)
	}
	// Every category is still readable from the mark titles.
	if n := len(charts[0].byClass("b")); n != 120 {
		t.Errorf("%d bars", n)
	}
	if !strings.Contains(charts[0].byClass("b")[0].title, "A fairly long category label 0") {
		t.Errorf("mark title lacks the full category: %q", charts[0].byClass("b")[0].title)
	}
}

func TestLineAxisRule(t *testing.T) {
	cases := []struct {
		dmin, dmax int64
		lo, hi     int64
	}{
		{96e6, 147e6, 80e6, 160e6}, // min above half of max: axis does not start at 0
		{40e6, 80e6, 0, 80e6},      // min at half of max: includes zero
		{0, 10e6, 0, 10e6},
		{-5e6, 5e6, -5e6, 10e6}, // step 5 leaves 2 intervals; one step is added
		{7e6, 7e6, 0, 8e6},
		{0, 0, 0, 1e6},
		{-9e6, -8e6, -9e6, -8e6},
	}
	for _, c := range cases {
		lo, hi := LineRange(c.dmin, c.dmax)
		s := NiceScale(lo, hi)
		if s.Lo != c.lo || s.Hi != c.hi {
			t.Errorf("line %d..%d: axis %d..%d, want %d..%d", c.dmin, c.dmax, s.Lo, s.Hi, c.lo, c.hi)
		}
	}
}

func TestNiceScaleProperties(t *testing.T) {
	rng := rand.New(rand.NewSource(1))
	check := func(lo0, hi0 int64) {
		s := NiceScale(lo0, hi0)
		n := s.Intervals()
		if n < 3 || n > 5 {
			t.Fatalf("%d..%d: %d intervals (axis %+v)", lo0, hi0, n, s)
		}
		if s.Lo > lo0 || s.Hi < hi0 || s.Lo%s.Step != 0 || s.Hi%s.Step != 0 {
			t.Fatalf("%d..%d: axis %+v does not cover or is not aligned", lo0, hi0, s)
		}
		if !isNice(float64(s.Step)) {
			t.Fatalf("%d..%d: step %d not nice", lo0, hi0, s.Step)
		}
		if got := len(s.Ticks()); got < 4 || got > 6 {
			t.Fatalf("%d..%d: %d ticks", lo0, hi0, got)
		}
	}
	for i := 0; i < 20000; i++ {
		mag := int64(1)
		for k := rng.Intn(18); k > 0; k-- {
			mag *= 10
		}
		a := rng.Int63n(mag) - mag/3
		b := a + 1 + rng.Int63n(mag)
		check(a, b)
	}
	for _, r := range [][2]int64{{0, 1}, {0, 2}, {0, 3}, {-1, 0}, {0, 1e6}, {0, 5e6}, {0, 999_999_999_999_000_000}, {-999_999_999_999_000_000, 999_999_999_999_000_000}} {
		check(r[0], r[1])
	}
	if s := NiceScale(0, 120e6); s.Hi != 150e6 || s.Step != 50e6 {
		// 0..120 with step 20 gives 6 intervals; step 50 gives 3.
		t.Errorf("0..120: %+v", s)
	}
	if s := NiceScale(0, 151e6); s.Hi != 200e6 || s.Step != 50e6 {
		t.Errorf("0..151: %+v", s)
	}
}

func TestHundredths(t *testing.T) {
	for in, want := range map[int64]string{0: "0.00", 5: "0.05", 100: "1.00", 12345: "123.45", -250: "-2.50", 64000: "640.00"} {
		if got := hund(in); got != want {
			t.Errorf("hund(%d) = %q, want %q", in, got, want)
		}
	}
	s := NiceScale(0, 10e6)
	if y := s.yOf(5e6, 1600, 28000); y != 14800 {
		t.Errorf("yOf midpoint = %d, want 14800", y)
	}
	if y := s.yOf(0, 1600, 28000); y != 28000 {
		t.Errorf("yOf(lo) = %d", y)
	}
	if y := s.yOf(10e6, 1600, 28000); y != 1600 {
		t.Errorf("yOf(hi) = %d", y)
	}
}

func TestDeterministic(t *testing.T) {
	for _, name := range exampleModels {
		raw := example(t, name)
		a, err := Profiler{}.Generate([]byte(raw), testGen)
		if err != nil {
			t.Fatal(err)
		}
		b, _ := Profiler{}.Generate([]byte(raw), testGen)
		for f, data := range a.Files {
			if !bytes.Equal(data, b.Files[f]) {
				t.Errorf("%s: %s differs between runs", name, f)
			}
		}
		if !bytes.Equal(a.Canonical, b.Canonical) {
			t.Errorf("%s: canonical encoding differs between runs", name)
		}
	}
	// Coordinates always have exactly two decimals.
	out := render(t, example(t, "sales-dashboard.json"))
	coord := regexp.MustCompile(`\s(?:x|y|x1|y1|x2|y2|cx|cy|width|height)="(-?[0-9.]+)"`)
	for _, m := range coord.FindAllStringSubmatch(out, -1) {
		if i := strings.IndexByte(m[1], '.'); i >= 0 && len(m[1])-i-1 != 2 {
			t.Errorf("coordinate %q is not fixed to two decimals", m[1])
		}
	}
	if strings.Contains(out, "NaN") || strings.Contains(out, "Inf") || strings.Contains(out, "e+") {
		t.Errorf("non-finite or exponent number in output")
	}
}

func TestSeriesColoursAndLineDashes(t *testing.T) {
	out := render(t, example(t, "sales-dashboard.json"))
	for _, want := range []string{".l2{stroke:var(--c2);stroke-dasharray:7 4}", ".l3{stroke:var(--c3);stroke-dasharray:2 3}", ".l4{stroke:var(--c4);stroke-dasharray:9 3 2 3}"} {
		if !strings.Contains(out, want) {
			t.Errorf("css lacks %q", want)
		}
	}
	// A multi-series chart has a legend with text labels; a single series has none.
	if !strings.Contains(out, `<ul class="legend">`) {
		t.Errorf("no legend on a multi-series chart")
	}
	single := render(t, chartModel("bar", `["a",1,2]`, `"A"`))
	if strings.Contains(single, "legend") && strings.Contains(single, `<ul class="legend">`) {
		t.Errorf("legend on a single-series chart")
	}
	// Every light and dark custom property exists once on :root.
	if strings.Count(out, "--c1:") != 2 || strings.Count(out, ":root{") != 2 {
		t.Errorf("palette not defined once per theme")
	}
}
