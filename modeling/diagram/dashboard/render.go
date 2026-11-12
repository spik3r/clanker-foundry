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

func renderChart(b *strings.Builder, ds Dataset, x Widget) {
	w := func(s string) { b.WriteString(s); b.WriteByte('\n') }
	isLine := x.Kind == KindLine
	cat := ds.column(x.Category)
	n := int64(len(ds.Rows))
	k := int64(len(x.Series))
	sc := make([]int, k)
	var dmin, dmax int64
	for i, id := range x.Series {
		sc[i] = ds.column(id)
	}
	first := true
	for _, c := range sc {
		for _, row := range ds.Rows {
			v := row[c].Micro
			if first || v < dmin {
				dmin = v
			}
			if first || v > dmax {
				dmax = v
			}
			first = false
		}
	}
	var lo, hi int64
	if isLine {
		lo, hi = LineRange(dmin, dmax)
	} else {
		lo, hi = BarRange(dmin, dmax)
	}
	scale := NiceScale(lo, hi)
	ticks := scale.Ticks()
	labels := make([]string, len(ticks))
	maxLabel := 0
	for i, t := range ticks {
		labels[i] = FormatMicro(t)
		if len(labels[i]) > maxLabel {
			maxLabel = len(labels[i])
		}
	}
	plotLeft := int64(leftPad + maxLabel*tickCharW + tickGap)
	left100, right100 := plotLeft*100, int64(plotRight)*100
	top100, bottom100 := int64(plotTop)*100, int64(plotBottom)*100
	plotW := right100 - left100
	gw := plotW / n // width of one category slot

	if k > 1 {
		w(`<ul class="legend">`)
		for i, id := range x.Series {
			label := ds.Columns[ds.column(id)].Label
			pos := strconv.Itoa(i + 1)
			if isLine {
				w(`<li><svg class="sw" viewBox="0 0 24 12" width="24" height="12" aria-hidden="true"><path class="l` + pos + `" d="M0 6H24"/></svg>` + esc(label) + "</li>")
			} else {
				w(`<li><svg class="sw" viewBox="0 0 24 12" width="24" height="12" aria-hidden="true"><rect class="b` + pos + `" x="0" y="0" width="24" height="12"/></svg>` + esc(label) + "</li>")
			}
		}
		w("</ul>")
	}
	kindName := "bar chart"
	if isLine {
		kindName = "line chart"
	}
	w(`<svg class="chart" viewBox="0 0 ` + strconv.Itoa(chartW) + ` ` + strconv.Itoa(chartH) + `" role="img" aria-label="` + esc(x.Title+", "+kindName) + `">`)
	w(`<rect class="plot" x="` + hund(left100) + `" y="` + hund(top100) + `" width="` + hund(plotW) + `" height="` + hund(bottom100-top100) + `"/>`)
	for i, t := range ticks {
		y := scale.yOf(t, top100, bottom100)
		w(`<line class="grid" data-v="` + labels[i] + `" x1="` + hund(left100) + `" y1="` + hund(y) + `" x2="` + hund(right100) + `" y2="` + hund(y) + `"/>`)
		w(`<text class="tick" x="` + hund(left100-tickGap*100) + `" y="` + hund(y+400) + `" text-anchor="end">` + labels[i] + "</text>")
	}
	if scale.Lo <= 0 && scale.Hi >= 0 {
		y := scale.yOf(0, top100, bottom100)
		w(`<line class="zero" data-v="0" x1="` + hund(left100) + `" y1="` + hund(y) + `" x2="` + hund(right100) + `" y2="` + hund(y) + `"/>`)
	}
	// X labels: show every t-th label so each gets at least xSlot pixels, and
	// cut each to the characters that fit in the space it owns.
	gwPx := gw / 100
	if gwPx < 1 {
		gwPx = 1
	}
	thin := (xSlot + gwPx - 1) / gwPx
	maxChars := int(thin*gwPx)/xCharW - 1
	if maxChars > maxXChars {
		maxChars = maxXChars
	}
	for i := int64(0); i < n; i += thin {
		cx := left100 + i*gw + gw/2
		text := truncate(ds.Rows[i][cat].Text, maxChars)
		w(`<text class="xl" x="` + hund(cx) + `" y="` + hund(int64(plotBottom+16)*100) + `" text-anchor="middle">` + esc(text) + "</text>")
	}
	y0 := scale.yOf(0, top100, bottom100)
	if !isLine {
		pad := gw / 5
		bw := (gw - pad) / k
		if bw > maxBarWidth {
			bw = maxBarWidth
		}
		inset := bw / 10
		off := (gw - bw*k) / 2
		for s := int64(0); s < k; s++ {
			label := ds.Columns[sc[s]].Label
			for r := int64(0); r < n; r++ {
				v := ds.Rows[r][sc[s]]
				yv := scale.yOf(v.Micro, top100, bottom100)
				y, h := yv, y0-yv
				if v.Micro < 0 {
					y, h = y0, yv-y0
				}
				xx := left100 + r*gw + off + s*bw + inset/2
				w(`<rect class="b` + strconv.FormatInt(s+1, 10) + `" x="` + hund(xx) + `" y="` + hund(y) + `" width="` + hund(bw-inset) + `" height="` + hund(h) + `"><title>` +
					esc(label+", "+ds.Rows[r][cat].Text+": "+v.Num) + "</title></rect>")
			}
		}
	} else {
		for s := int64(0); s < k; s++ {
			pos := strconv.FormatInt(s+1, 10)
			var d strings.Builder
			for r := int64(0); r < n; r++ {
				cmd := "L"
				if r == 0 {
					cmd = "M"
				}
				d.WriteString(cmd + hund(left100+r*gw+gw/2) + " " + hund(scale.yOf(ds.Rows[r][sc[s]].Micro, top100, bottom100)) + " ")
			}
			w(`<path class="l` + pos + `" d="` + strings.TrimSuffix(d.String(), " ") + `"/>`)
		}
		for s := int64(0); s < k; s++ {
			label := ds.Columns[sc[s]].Label
			for r := int64(0); r < n; r++ {
				v := ds.Rows[r][sc[s]]
				w(`<circle class="m` + strconv.FormatInt(s+1, 10) + `" cx="` + hund(left100+r*gw+gw/2) + `" cy="` + hund(scale.yOf(v.Micro, top100, bottom100)) + `" r="3"><title>` +
					esc(label+", "+ds.Rows[r][cat].Text+": "+v.Num) + "</title></circle>")
			}
		}
	}
	w("</svg>")
}
