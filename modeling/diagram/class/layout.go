package class

import (
	"sort"
	"strings"

	"github.com/spik3r/clanker-foundry/modeling/diagram/internal/svgx"
)

// Layout constants. All coordinates are integers. Text width is estimated from
// character counts; no font is measured, so a line of wide characters can
// still overflow its box.
const (
	margin     = 24
	charW      = 7 // estimated width of one member character
	nameCharW  = 8 // estimated width of one character of a bold class name
	lineH      = 16
	minBoxW    = 120
	maxBoxW    = 300
	boxPadX    = 8
	portSpace  = 9    // box width reserved per relation touching the box
	hGap       = 48   // between boxes of a row
	maxRowW    = 1400 // a rank wider than this is split into several rows
	chanPad    = 28   // free space above the first and below the last track
	chanEmpty  = 28   // height of a channel no relation uses
	trackPitch = 20
	trackSlack = 8
	corridor   = 12 // spacing of the vertical corridors right of the boxes
	startY     = 40
	labelH     = 12
	sweeps     = 4
)

// Rect is an axis-aligned rectangle.
type Rect struct{ X, Y, W, H int }

// Point is a vertex of a relation's polyline.
type Point struct{ X, Y int }

// ClassGeom is the drawn geometry of one class.
type ClassGeom struct {
	C     Class
	Box   Rect
	Row   int
	Stere string // "«interface»", "«enumeration»" or ""
	Name  string // truncated
	Attrs []string
	Ops   []string
	NameH int
	AttrH int
	OpsH  int
}

// TextGeom is a drawn label. Box is its estimated extent.
type TextGeom struct {
	Text   string
	X, Y   int // baseline origin
	Anchor string
	Box    Rect
}

// RelGeom is the drawn geometry of one relation. Points run from the From
// class to the To class.
type RelGeom struct {
	R         Relation
	Points    []Point
	Label     *TextGeom
	FromLabel *TextGeom
	ToLabel   *TextGeom
}

// Layout is the complete geometry of a class diagram.
type Layout struct {
	Width, Height int
	Rows          int
	Classes       []ClassGeom // sorted by id, like the model
	Relations     []RelGeom
}

func runes(s string) int { return len([]rune(s)) }

func visSymbol(v string) string {
	switch v {
	case Private:
		return "-"
	case Protected:
		return "#"
	case Package:
		return "~"
	}
	return "+"
}

func attrLine(c Class, a Attribute) string {
	if c.Kind == KindEnum {
		return a.Name
	}
	s := visSymbol(a.Visibility) + " " + a.Name
	if a.Type != "" {
		s += ": " + a.Type
	}
	return s
}

func opLine(o Operation) string {
	ps := make([]string, len(o.Parameters))
	for i, p := range o.Parameters {
		ps[i] = p.Name + ": " + p.Type
	}
	s := visSymbol(o.Visibility) + " " + o.Name + "(" + strings.Join(ps, ", ") + ")"
	if o.Returns != "" {
		s += ": " + o.Returns
	}
	return s
}

func clamp(v, lo, hi int) int { return max(lo, min(hi, v)) }

// LayoutOf computes the layout of a validated, normalized model.
func LayoutOf(m Model) Layout {
	n := len(m.Classes)
	idx := map[string]int{}
	for i, c := range m.Classes {
		idx[c.ID] = i
	}
	rels := m.Relations
	from := make([]int, len(rels))
	to := make([]int, len(rels))
	deg := make([]int, n)
	for i, r := range rels {
		from[i], to[i] = idx[r.From], idx[r.To]
		deg[from[i]]++
		deg[to[i]]++
	}

	// Rank: parents above children, longest path over inheritance and
	// realization. Validation has refused cycles.
	parents := make([][]int, n)
	for i, r := range rels {
		if r.Kind == Inheritance || r.Kind == Realization {
			parents[from[i]] = append(parents[from[i]], to[i])
		}
	}
	rank := make([]int, n)
	done := make([]bool, n)
	var rankOf func(c int) int
	rankOf = func(c int) int {
		if done[c] {
			return rank[c]
		}
		r := 0
		for _, p := range parents[c] {
			r = max(r, rankOf(p)+1)
		}
		rank[c], done[c] = r, true
		return r
	}
	maxRank := 0
	for c := 0; c < n; c++ {
		maxRank = max(maxRank, rankOf(c))
	}

	// Order within a rank: a few barycenter sweeps over the relations between
	// different ranks. Ties keep the earlier order, which starts as id order.
	ranks := make([][]int, maxRank+1)
	for c := 0; c < n; c++ {
		ranks[rank[c]] = append(ranks[rank[c]], c)
	}
	nbr := make([][]int, n)
	for i := range rels {
		nbr[from[i]] = append(nbr[from[i]], to[i])
		nbr[to[i]] = append(nbr[to[i]], from[i])
	}
	pos := make([]int, n)
	setPos := func(r int) {
		for i, c := range ranks[r] {
			pos[c] = (2*i + 1) * 1000 / (2 * len(ranks[r]))
		}
	}
	for r := range ranks {
		setPos(r)
	}
	reorder := func(r int, useAbove bool) {
		bary := map[int]int{}
		for _, c := range ranks[r] {
			sum, cnt := 0, 0
			for _, o := range nbr[c] {
				if (useAbove && rank[o] < r) || (!useAbove && rank[o] > r) {
					sum += pos[o]
					cnt++
				}
			}
			if cnt == 0 {
				bary[c] = pos[c]
			} else {
				bary[c] = sum / cnt
			}
		}
		sort.SliceStable(ranks[r], func(i, j int) bool { return bary[ranks[r][i]] < bary[ranks[r][j]] })
		setPos(r)
	}
	for s := 0; s < sweeps; s++ {
		for r := 1; r <= maxRank; r++ {
			reorder(r, true)
		}
		for r := maxRank - 1; r >= 0; r-- {
			reorder(r, false)
		}
	}

	// Box sizes.
	var l Layout
	l.Classes = make([]ClassGeom, n)
	for i, c := range m.Classes {
		g := ClassGeom{C: c}
		switch c.Kind {
		case KindInterface:
			g.Stere = "«interface»"
		case KindEnum:
			g.Stere = "«enumeration»"
		}
		var attrs, ops []string
		for _, a := range c.Attributes {
			attrs = append(attrs, attrLine(c, a))
		}
		for _, o := range c.Operations {
			ops = append(ops, opLine(o))
		}
		chars := (runes(c.Name)*nameCharW + charW - 1) / charW
		chars = max(chars, runes(g.Stere))
		for _, s := range attrs {
			chars = max(chars, runes(s))
		}
		for _, s := range ops {
			chars = max(chars, runes(s))
		}
		w := clamp(chars*charW+2*boxPadX, minBoxW, maxBoxW)
		w = max(w, 2*boxPadX+portSpace*deg[i])
		budget := (w - 2*boxPadX) / charW
		g.Name = svgx.Truncate(c.Name, (w-2*boxPadX)/nameCharW)
		for _, s := range attrs {
			g.Attrs = append(g.Attrs, svgx.Truncate(s, budget))
		}
		for _, s := range ops {
			g.Ops = append(g.Ops, svgx.Truncate(s, budget))
		}
		g.NameH = 12 + lineH
		if g.Stere != "" {
			g.NameH += lineH
		}
		g.AttrH = 8 + lineH*len(attrs)
		g.OpsH = 8 + lineH*len(ops)
		g.Box = Rect{0, 0, w, g.NameH + g.AttrH + g.OpsH}
		l.Classes[i] = g
	}

	// Rows: each rank in order, split when it gets too wide. Every row of
	// rank r comes before every row of rank r+1.
	var rows [][]int
	for _, rk := range ranks {
		var cur []int
		curW := 0
		for _, c := range rk {
			w := l.Classes[c].Box.W
			if len(cur) > 0 && curW+hGap+w > maxRowW {
				rows = append(rows, cur)
				cur, curW = nil, 0
			}
			if len(cur) > 0 {
				curW += hGap
			}
			cur = append(cur, c)
			curW += w
		}
		if len(cur) > 0 {
			rows = append(rows, cur)
		}
	}
	nRows := len(rows)
	l.Rows = nRows
	rowW := make([]int, nRows)
	widest := 0
	for r, row := range rows {
		for i, c := range row {
			if i > 0 {
				rowW[r] += hGap
			}
			rowW[r] += l.Classes[c].Box.W
		}
		widest = max(widest, rowW[r])
	}
	rowOf := make([]int, n)
	rowH := make([]int, nRows)
	contentRight := 0
	for r, row := range rows {
		x := margin + (widest-rowW[r])/2
		for _, c := range row {
			g := &l.Classes[c]
			g.Row = r
			g.Box.X = x
			x += g.Box.W + hGap
			rowOf[c] = r
			rowH[r] = max(rowH[r], g.Box.H)
			contentRight = max(contentRight, g.Box.X+g.Box.W)
		}
	}

	// Routes. A relation between rows i < j leaves the bottom of the upper
	// class into the channel above row i+1 and enters the top of the lower
	// class from the channel above row j; through rows in between it runs in
	// a vertical corridor right of every box. A relation inside one row
	// arches over the row through the channel above it. Channels are the
	// box-free bands above each row.
	type port struct {
		rel, end int // end 0: From, 1: To
		box      int
		bottom   bool
		ch       int
		other    int // centre x of the class at the other end
		x        int
	}
	type route struct {
		kind   int // 0 same row, 1 adjacent rows, 2 through a corridor
		ports  [2]int
		xc     int
		leg    [2]int // leg index per channel
		legCh  [2]int
		nlegs  int
		legLoX [2]int
		legHiX [2]int
		track  [2]int
	}
	var ports []port
	routes := make([]route, len(rels))
	center := func(c int) int { return l.Classes[c].Box.X + l.Classes[c].Box.W/2 }
	for i := range rels {
		a, b := from[i], to[i]
		rt := &routes[i]
		add := func(end, box int, bottom bool, ch, other int) int {
			ports = append(ports, port{rel: i, end: end, box: box, bottom: bottom, ch: ch, other: center(other)})
			return len(ports) - 1
		}
		switch {
		case rowOf[a] == rowOf[b]:
			rt.kind = 0
			rt.ports[0] = add(0, a, false, rowOf[a], b)
			rt.ports[1] = add(1, b, false, rowOf[b], a)
		default:
			up, lo := a, b
			upEnd, loEnd := 0, 1
			if rowOf[a] > rowOf[b] {
				up, lo = b, a
				upEnd, loEnd = 1, 0
			}
			rt.kind = 1
			ch1, ch2 := rowOf[lo], rowOf[lo]
			if rowOf[lo]-rowOf[up] > 1 {
				rt.kind = 2
				ch1 = rowOf[up] + 1
			}
			p1 := add(upEnd, up, true, ch1, lo)
			p2 := add(loEnd, lo, false, ch2, up)
			// ports[0] is always the upper end, ports[1] the lower end
			rt.ports[0], rt.ports[1] = p1, p2
		}
	}

	// Port positions: along a border, ordered by the x of the class at the
	// other end, spread evenly; then moved so that no two relations leave a
	// channel at the same x.
	type side struct {
		box    int
		bottom bool
	}
	groups := map[side][]int{}
	var sides []side
	for i, p := range ports {
		s := side{p.box, p.bottom}
		if _, ok := groups[s]; !ok {
			sides = append(sides, s)
		}
		groups[s] = append(groups[s], i)
	}
	sort.Slice(sides, func(i, j int) bool {
		if sides[i].box != sides[j].box {
			return sides[i].box < sides[j].box
		}
		return !sides[i].bottom && sides[j].bottom
	})
	taken := map[[2]int]int{} // (channel, x) -> relation index + 1
	for _, s := range sides {
		g := groups[s]
		sort.SliceStable(g, func(i, j int) bool {
			a, b := ports[g[i]], ports[g[j]]
			if a.other != b.other {
				return a.other < b.other
			}
			if a.rel != b.rel {
				return a.rel < b.rel
			}
			return a.end < b.end
		})
		box := l.Classes[s.box].Box
		for k, pi := range g {
			p := &ports[pi]
			x := box.X + (k+1)*box.W/(len(g)+1)
			lo, hi := box.X+5, box.X+box.W-5
			for d := 0; ; d += 3 {
				found := false
				for _, cand := range []int{x + d, x - d} {
					if cand < lo || cand > hi {
						continue
					}
					if o := taken[[2]int{p.ch, cand}]; o == 0 || o == p.rel+1 {
						x = cand
						found = true
						break
					}
				}
				if found || d > box.W {
					break
				}
			}
			p.x = x
			taken[[2]int{p.ch, x}] = p.rel + 1
		}
	}

	// Legs: the horizontal runs of each relation, per channel, then tracks.
	type leg struct {
		rel, which int
		lo, hi     int
	}
	legs := make([][]leg, nRows)
	corr := 0
	for i := range rels {
		rt := &routes[i]
		p0, p1 := ports[rt.ports[0]], ports[rt.ports[1]]
		switch rt.kind {
		case 0:
			rt.legCh[0], rt.nlegs = p0.ch, 1
			rt.legLoX[0], rt.legHiX[0] = min(p0.x, p1.x), max(p0.x, p1.x)
		case 1:
			rt.legCh[0], rt.nlegs = p0.ch, 1
			rt.legLoX[0], rt.legHiX[0] = min(p0.x, p1.x), max(p0.x, p1.x)
		default:
			rt.xc = contentRight + chanPad + corr*corridor
			corr++
			rt.legCh[0], rt.legCh[1], rt.nlegs = p0.ch, p1.ch, 2
			rt.legLoX[0], rt.legHiX[0] = min(p0.x, rt.xc), max(p0.x, rt.xc)
			rt.legLoX[1], rt.legHiX[1] = min(p1.x, rt.xc), max(p1.x, rt.xc)
		}
		for w := 0; w < rt.nlegs; w++ {
			rt.track[w] = -1
			if rt.legLoX[w] != rt.legHiX[w] {
				legs[rt.legCh[w]] = append(legs[rt.legCh[w]], leg{i, w, rt.legLoX[w], rt.legHiX[w]})
			}
		}
	}
	tracks := make([]int, nRows)
	for ch := range legs {
		ls := legs[ch]
		sort.Slice(ls, func(i, j int) bool {
			if ls[i].lo != ls[j].lo {
				return ls[i].lo < ls[j].lo
			}
			if ls[i].hi != ls[j].hi {
				return ls[i].hi < ls[j].hi
			}
			if ls[i].rel != ls[j].rel {
				return ls[i].rel < ls[j].rel
			}
			return ls[i].which < ls[j].which
		})
		var trackLegs [][]leg
		for _, lg := range ls {
			placed := -1
			for t := range trackLegs {
				ok := true
				for _, o := range trackLegs[t] {
					if lg.lo-trackSlack <= o.hi && o.lo-trackSlack <= lg.hi {
						ok = false
						break
					}
				}
				if ok {
					placed = t
					break
				}
			}
			if placed < 0 {
				trackLegs = append(trackLegs, nil)
				placed = len(trackLegs) - 1
			}
			trackLegs[placed] = append(trackLegs[placed], lg)
			routes[lg.rel].track[lg.which] = placed
		}
		tracks[ch] = len(trackLegs)
	}
	used := make([]bool, nRows)
	for _, p := range ports {
		used[p.ch] = true
	}

	// Vertical positions.
	chTop := make([]int, nRows)
	rowTop := make([]int, nRows)
	y := startY
	for r := 0; r < nRows; r++ {
		chTop[r] = y
		h := chanEmpty
		if used[r] {
			h = 2 * chanPad
			if tracks[r] > 1 {
				h += (tracks[r] - 1) * trackPitch
			}
		}
		rowTop[r] = y + h
		y = rowTop[r] + rowH[r]
	}
	for i := range l.Classes {
		l.Classes[i].Box.Y = rowTop[l.Classes[i].Row]
	}
	bottom := y
	trackY := func(ch, t int) int { return chTop[ch] + chanPad + t*trackPitch }

	// Polylines and labels.
	boxes := make([]Rect, n)
	for i := range l.Classes {
		boxes[i] = l.Classes[i].Box
	}
	maxX, maxY := contentRight, bottom
	l.Relations = make([]RelGeom, len(rels))
	for i, r := range rels {
		rt := routes[i]
		p0, p1 := ports[rt.ports[0]], ports[rt.ports[1]]
		portY := func(p port) int {
			b := l.Classes[p.box].Box
			if p.bottom {
				return b.Y + b.H
			}
			return b.Y
		}
		var pts []Point
		switch rt.kind {
		case 0, 1:
			pts = []Point{{p0.x, portY(p0)}}
			if rt.track[0] >= 0 {
				ty := trackY(rt.legCh[0], rt.track[0])
				pts = append(pts, Point{p0.x, ty}, Point{p1.x, ty})
			}
			pts = append(pts, Point{p1.x, portY(p1)})
		default:
			t1 := trackY(rt.legCh[0], rt.track[0])
			t2 := trackY(rt.legCh[1], rt.track[1])
			pts = []Point{{p0.x, portY(p0)}, {p0.x, t1}, {rt.xc, t1}, {rt.xc, t2}, {p1.x, t2}, {p1.x, portY(p1)}}
		}
		pts = simplify(pts)
		// Ports 0 and 1 are the upper and lower ends for routes between rows;
		// the polyline must run From -> To.
		if rt.kind != 0 && ports[rt.ports[0]].end == 1 {
			for a, b := 0, len(pts)-1; a < b; a, b = a+1, b-1 {
				pts[a], pts[b] = pts[b], pts[a]
			}
		}
		g := RelGeom{R: r, Points: pts}
		for _, p := range pts {
			maxX = max(maxX, p.X)
		}
		endLabel := func(text string, p Point, bottom bool) *TextGeom {
			if text == "" {
				return nil
			}
			dir, by := -1, p.Y-6
			if bottom {
				dir, by = 1, p.Y+14
			}
			var cands []*TextGeom
			for k := 0; k < 8; k++ {
				cands = append(cands, textGeom(text, p.X+10, by+dir*k*lineH, "start"), textGeom(text, p.X-10, by+dir*k*lineH, "end"))
			}
			return pick(boxes, cands, dir)
		}
		pf, pt := p0, p1
		if p0.end == 1 {
			pf, pt = p1, p0
		}
		g.FromLabel = endLabel(r.FromLabel, pts[0], pf.bottom)
		g.ToLabel = endLabel(r.ToLabel, pts[len(pts)-1], pt.bottom)
		if r.Label != "" {
			g.Label = placeLabel(boxes, r.Label, pts)
		}
		for _, t := range []*TextGeom{g.Label, g.FromLabel, g.ToLabel} {
			if t != nil {
				maxX = max(maxX, t.Box.X+t.Box.W)
				maxY = max(maxY, t.Box.Y+t.Box.H)
			}
		}
		l.Relations[i] = g
	}
	l.Width = maxX + margin
	l.Height = maxY + margin
	return l
}

func textGeom(text string, x, baseline int, anchor string) *TextGeom {
	w := runes(text) * charW
	bx := x
	switch anchor {
	case "middle":
		bx = x - w/2
	case "end":
		bx = x - w
	}
	return &TextGeom{Text: text, X: x, Y: baseline, Anchor: anchor, Box: Rect{bx, baseline - labelH + 2, w, labelH}}
}

// pick preserves the first clear candidate. If none fits, class edges give a
// finite set of alternative baselines. End labels keep moving away from their
// class; centre labels beside vertical segments may move in either direction.
// direction is -1 above, +1 below, or 0 for either side of the baseline.
func pick(boxes []Rect, cands []*TextGeom, direction int) *TextGeom {
	clear := func(c *TextGeom) bool {
		if c.Box.X < margin || c.Box.Y < margin {
			return false
		}
		for _, b := range boxes {
			if overlap(c.Box, b) {
				return false
			}
		}
		return true
	}
	for _, c := range cands {
		if clear(c) {
			return c
		}
	}

	hasUnshiftedCandidate := false
	for _, c := range cands {
		hasUnshiftedCandidate = hasUnshiftedCandidate || c.Box.X >= margin
	}
	var best *TextGeom
	bestDistance := 0
	for _, c := range cands {
		// Keep end labels beside their port. Do not shift a left-anchored
		// candidate across its line when the right side is available.
		if hasUnshiftedCandidate && c.Box.X < margin {
			continue
		}
		baselines := []int{max(c.Y, margin+labelH-2)}
		for _, b := range boxes {
			// A text box spans baseline-labelH+2 through baseline+2.
			baselines = append(baselines, b.Y-2, b.Y+b.H+labelH-2)
		}
		for _, baseline := range baselines {
			if (direction > 0 && baseline < cands[0].Y) || (direction < 0 && baseline > cands[0].Y) {
				continue
			}
			n := *c
			n.X += max(0, margin-n.Box.X)
			n.Box.X = max(n.Box.X, margin)
			n.Y = baseline
			n.Box.Y = baseline - labelH + 2
			if !clear(&n) {
				continue
			}
			distance := abs(n.Box.X-cands[0].Box.X) + abs(n.Box.Y-cands[0].Box.Y)
			if best == nil || distance < bestDistance {
				best, bestDistance = &n, distance
			}
		}
	}
	return best
}

func overlap(a, b Rect) bool {
	return a.X < b.X+b.W && b.X < a.X+a.W && a.Y < b.Y+b.H && b.Y < a.Y+a.H
}

// placeLabel puts a relation label above its longest horizontal segment, or
// beside its longest vertical segment when it has no horizontal one.
func placeLabel(boxes []Rect, text string, pts []Point) *TextGeom {
	best, bestLen := -1, 0
	for i := 0; i+1 < len(pts); i++ {
		if pts[i].Y == pts[i+1].Y {
			if d := abs(pts[i+1].X - pts[i].X); d > bestLen {
				best, bestLen = i, d
			}
		}
	}
	if best >= 0 {
		a, b := pts[best], pts[best+1]
		return pick(boxes, []*TextGeom{textGeom(text, (a.X+b.X)/2, a.Y-5, "middle")}, -1)
	}
	for i := 0; i+1 < len(pts); i++ {
		if d := abs(pts[i+1].Y - pts[i].Y); d > bestLen {
			best, bestLen = i, d
		}
	}
	a, b := pts[best], pts[best+1]
	mid := (a.Y+b.Y)/2 + 4
	var cands []*TextGeom
	for k := 0; k < 5; k++ {
		for _, dy := range []int{k * lineH, -k * lineH} {
			cands = append(cands, textGeom(text, a.X+8, mid+dy, "start"), textGeom(text, a.X-8, mid+dy, "end"))
		}
	}
	return pick(boxes, cands, 0)
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

// simplify drops repeated points and middle points of straight runs.
func simplify(pts []Point) []Point {
	var out []Point
	for _, p := range pts {
		if len(out) > 0 && out[len(out)-1] == p {
			continue
		}
		for len(out) >= 2 {
			a, b := out[len(out)-2], out[len(out)-1]
			if (a.X == b.X && b.X == p.X) || (a.Y == b.Y && b.Y == p.Y) {
				out = out[:len(out)-1]
				continue
			}
			break
		}
		out = append(out, p)
	}
	return out
}
