package dashboard

import (
	"html"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/spik3r/clanker-foundry/modeling/diagram/contract"
)

// Chart geometry, in pixels of the 640 by 320 viewBox. Marks are computed in
// hundredths of a pixel.
const (
	chartW      = 640
	chartH      = 320
	plotTop     = 16
	plotBottom  = 280
	plotRight   = 624
	leftPad     = 16 // space left of the tick labels
	tickCharW   = 8  // assumed width of one tick label character
	tickGap     = 8
	xCharW      = 8  // assumed width of one x label character
	xSlot       = 64 // minimum width given to one x label
	maxXChars   = 24
	maxBarWidth = 6000 // hundredths
)

func esc(s string) string { return html.EscapeString(s) }

func (ds Dataset) column(id string) int {
	for i, c := range ds.Columns {
		if c.ID == id {
			return i
		}
	}
	return -1
}

func (m Model) dataset(id string) Dataset {
	for _, ds := range m.Datasets {
		if ds.ID == id {
			return ds
		}
	}
	return Dataset{}
}

// Render writes dashboard.html for a validated, normalized model.
func Render(m Model, g contract.Generator) []byte {
	var b strings.Builder
	w := func(s string) { b.WriteString(s); b.WriteByte('\n') }
	byID := map[string]Widget{}
	for _, x := range m.Widgets {
		byID[x.ID] = x
	}
	w("<!doctype html>")
	w(`<html lang="en">`)
	w("<head>")
	w(`<meta charset="utf-8">`)
	w(`<meta name="viewport" content="width=device-width, initial-scale=1">`)
	w(`<meta name="generator" content="` + esc(g.ID+" "+g.Version) + `">`)
	w("<title>" + esc(m.Title) + "</title>")
	w("<style>")
	b.WriteString(css)
	w("</style>")
	w("</head>")
	w("<body>")
	w("<header>")
	w("<h1>" + esc(m.Title) + "</h1>")
	if m.Subtitle != "" {
		w(`<p class="subtitle">` + esc(m.Subtitle) + "</p>")
	}
	w("</header>")
	w("<main>")
	for _, row := range m.Layout {
		w(`<section class="row cols-` + strconv.Itoa(len(row)) + `">`)
		for _, id := range row {
			renderWidget(&b, m, byID[id])
		}
		w("</section>")
	}
	w("</main>")
	w("</body>")
	w("</html>")
	return []byte(b.String())
}

func renderWidget(b *strings.Builder, m Model, x Widget) {
	w := func(s string) { b.WriteString(s); b.WriteByte('\n') }
	w(`<article data-id="` + esc(x.ID) + `" data-kind="` + x.Kind + `">`)
	switch x.Kind {
	case KindStat:
		w(`<p class="stat-label">` + esc(x.Label) + "</p>")
		w(`<p class="stat-value">` + esc(x.Value) + "</p>")
		if x.Note != "" {
			w(`<p class="note">` + esc(x.Note) + "</p>")
		}
	case KindTable:
		w("<h2>" + esc(x.Title) + "</h2>")
		renderTable(b, m.dataset(x.Dataset), x)
	default:
		w("<h2>" + esc(x.Title) + "</h2>")
		renderChart(b, m.dataset(x.Dataset), x)
	}
	w("</article>")
}

func renderTable(b *strings.Builder, ds Dataset, x Widget) {
	w := func(s string) { b.WriteString(s); b.WriteByte('\n') }
	idx := make([]int, len(x.Columns))
	for i, id := range x.Columns {
		idx[i] = ds.column(id)
	}
	w(`<div class="scroll" role="region" tabindex="0" aria-label="` + esc(x.Title) + `">`)
	w("<table>")
	w("<thead>")
	b.WriteString("<tr>")
	for _, c := range idx {
		col := ds.Columns[c]
		b.WriteString(`<th scope="col"` + numClass(col) + ">" + esc(col.Label) + "</th>")
	}
	w("</tr>")
	w("</thead>")
	w("<tbody>")
	for _, row := range ds.Rows {
		b.WriteString("<tr>")
		for _, c := range idx {
			b.WriteString("<td" + numClass(ds.Columns[c]) + ">" + esc(row[c].String()) + "</td>")
		}
		w("</tr>")
	}
	w("</tbody>")
	w("</table>")
	w("</div>")
}

func numClass(c Column) string {
	if c.Type == TypeNumber {
		return ` class="num"`
	}
	return ""
}

// truncate keeps at most n runes, ending with an ellipsis when it cuts.
func truncate(s string, n int) string {
	if utf8.RuneCountInString(s) <= n {
		return s
	}
	r := []rune(s)
	return string(r[:n-1]) + "…"
}

// renderChart writes the HTML form of a chart: the legend list and the svg.
func renderChart(b *strings.Builder, ds Dataset, x Widget) {
	w := func(s string) { b.WriteString(s); b.WriteByte('\n') }
	g := layoutChart(ds, x)
	if len(g.legend) > 0 {
		w(`<ul class="legend">`)
		for _, it := range g.legend {
			mark := `<rect class="b` + it.pos + `" x="0" y="0" width="24" height="12"/>`
			if it.line {
				mark = `<path class="l` + it.pos + `" d="M0 6H24"/>`
			}
			w(`<li><svg class="sw" viewBox="0 0 24 12" width="24" height="12" aria-hidden="true">` + mark + `</svg>` + esc(it.label) + "</li>")
		}
		w("</ul>")
	}
	w(`<svg class="chart" viewBox="0 0 ` + strconv.Itoa(chartW) + ` ` + strconv.Itoa(chartH) + `" role="img" aria-label="` + esc(g.ariaLabel) + `">`)
	for _, e := range g.els {
		var open strings.Builder
		open.WriteString("<" + e.tag + ` class="` + e.class + `"`)
		for _, a := range e.attrs {
			open.WriteString(" " + a[0] + `="` + esc(a[1]) + `"`)
		}
		switch {
		case e.isText:
			w(open.String() + ">" + esc(e.text) + "</text>")
		case e.hasTitle:
			w(open.String() + "><title>" + esc(e.title) + "</title></" + e.tag + ">")
		default:
			w(open.String() + "/>")
		}
	}
	w("</svg>")
}
