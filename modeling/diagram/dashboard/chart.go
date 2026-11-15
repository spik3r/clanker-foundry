package dashboard

import "strconv"

// chartGeom is the computed geometry of one bar or line chart: the legend and
// the svg children in drawing order. Both output targets serialize the same
// value, so the arithmetic exists once. Class names here are the HTML names;
// the Solid target adds its prefix.
type chartGeom struct {
	legend    []legendItem
	ariaLabel string
	els       []svgEl
}

type legendItem struct {
	label string
	line  bool
	pos   string // "1".."4", the series position
}

// svgEl is one svg child. attrs follow class, in output order. A text element
// carries text; a data mark carries a title; any other element has neither.
type svgEl struct {
	tag      string
	class    string
	attrs    [][2]string
	text     string
	isText   bool
	title    string
	hasTitle bool
}

func el(tag, class string, attrs ...string) svgEl {
	e := svgEl{tag: tag, class: class}
	for i := 0; i+1 < len(attrs); i += 2 {
		e.attrs = append(e.attrs, [2]string{attrs[i], attrs[i+1]})
	}
	return e
}

// layoutChart computes every mark of a validated chart widget over ds.
func layoutChart(ds Dataset, x Widget) chartGeom {
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

	var g chartGeom
	if k > 1 {
		for i, id := range x.Series {
			g.legend = append(g.legend, legendItem{label: ds.Columns[ds.column(id)].Label, line: isLine, pos: strconv.Itoa(i + 1)})
		}
	}
	kindName := "bar chart"
	if isLine {
		kindName = "line chart"
	}
	g.ariaLabel = x.Title + ", " + kindName
	add := func(e svgEl) { g.els = append(g.els, e) }
	add(el("rect", "plot", "x", hund(left100), "y", hund(top100), "width", hund(plotW), "height", hund(bottom100-top100)))
	for i, t := range ticks {
		y := scale.yOf(t, top100, bottom100)
		add(el("line", "grid", "data-v", labels[i], "x1", hund(left100), "y1", hund(y), "x2", hund(right100), "y2", hund(y)))
		tick := el("text", "tick", "x", hund(left100-tickGap*100), "y", hund(y+400), "text-anchor", "end")
		tick.isText, tick.text = true, labels[i]
		add(tick)
	}
	if scale.Lo <= 0 && scale.Hi >= 0 {
		y := scale.yOf(0, top100, bottom100)
		add(el("line", "zero", "data-v", "0", "x1", hund(left100), "y1", hund(y), "x2", hund(right100), "y2", hund(y)))
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
		xl := el("text", "xl", "x", hund(cx), "y", hund(int64(plotBottom+16)*100), "text-anchor", "middle")
		xl.isText, xl.text = true, truncate(ds.Rows[i][cat].Text, maxChars)
		add(xl)
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
				b := el("rect", "b"+strconv.FormatInt(s+1, 10), "x", hund(xx), "y", hund(y), "width", hund(bw-inset), "height", hund(h))
				b.hasTitle, b.title = true, label+", "+ds.Rows[r][cat].Text+": "+v.Num
				add(b)
			}
		}
		return g
	}
	for s := int64(0); s < k; s++ {
		var d []byte
		for r := int64(0); r < n; r++ {
			cmd := "L"
			if r == 0 {
				cmd = "M"
			}
			d = append(d, (cmd + hund(left100+r*gw+gw/2) + " " + hund(scale.yOf(ds.Rows[r][sc[s]].Micro, top100, bottom100)) + " ")...)
		}
		add(el("path", "l"+strconv.FormatInt(s+1, 10), "d", string(d[:len(d)-1])))
	}
	for s := int64(0); s < k; s++ {
		label := ds.Columns[sc[s]].Label
		for r := int64(0); r < n; r++ {
			v := ds.Rows[r][sc[s]]
			c := el("circle", "m"+strconv.FormatInt(s+1, 10), "cx", hund(left100+r*gw+gw/2), "cy", hund(scale.yOf(v.Micro, top100, bottom100)), "r", "3")
			c.hasTitle, c.title = true, label+", "+ds.Rows[r][cat].Text+": "+v.Num
			add(c)
		}
	}
	return g
}
