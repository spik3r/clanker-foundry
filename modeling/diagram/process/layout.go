package process

import "unicode/utf8"

// Fixed layout constants, in SVG user units. They are part of the generator
// version: changing one changes emitted bytes.
const (
	margin     = 20  // outer margin
	titleH     = 28  // height of the visible title line
	laneStripW = 140 // width of the lane label strip
	cellW      = 160 // width of a column cell
	rowH       = 176 // height of one node row inside a lane
	minSlotsW  = 72  // narrowest slot zone of a gutter
	slotPad    = 16  // space before the first and after the last routing slot
	slotStep   = 10  // distance between routing slots
	taskW      = 150
	taskH      = 64
	eventD     = 36 // start and end circle diameter
	gatewayD   = 48 // gateway diamond bounding box

	flowBudget  = 12 // characters of a flow label
	labelCharW  = 7  // assumed width of one flow label character
	labelH      = 10 // assumed height of a flow label
	labelGap    = 6  // distance from a node box to a flow label
	portsPerDir = 3  // exits on each of a gateway's top, bottom and right
	maxExits    = 3 * portsPerDir
)

// Gateway exits. Offsets are relative to the gateway centre. A flow leaving
// the top or bottom first runs vertically to a corridor line above or below
// the gateway and turns right there. Every horizontal line a flow can draw
// inside a row lies on one of these levels or on the row centre, at least 18
// units apart, so a flow label fits between two of them.
var (
	rightExitDY   = [portsPerDir]int{0, -18, 18}
	sideExitDX    = [portsPerDir]int{0, 12, -12}
	upCorridorY   = [portsPerDir]int{-54, -36, -72}
	downCorridorY = [portsPerDir]int{54, 36, 72}
)

const (
	sideRight = iota
	sideUp
	sideDown
)

// Point is a position in SVG user units.
type Point struct{ X, Y int }

// Box is an axis-aligned rectangle.
type Box struct{ X, Y, W, H int }

// LaneBox is a lane band. Box covers the label strip and the drawing area.
type LaneBox struct {
	Lane Lane
	Box  Box
}

// NodeBox is a placed node; Box is its shape's bounding box.
type NodeBox struct {
	Node   Node
	Box    Box
	CX, CY int
}

// FlowLabel is a placed flow label. Anchor is the index of the segment
// Points[Anchor]..Points[Anchor+1] the label belongs to; no other flow shares
// that segment. Box is the estimated extent of the text.
type FlowLabel struct {
	Text   string
	Box    Box
	Anchor int
}

// FlowPath is a routed flow: an orthogonal polyline from its source to the
// left side of its target. Label is nil for an unlabelled flow.
type FlowPath struct {
	Flow   Flow
	Points []Point
	Label  *FlowLabel
}

// Layout is the complete geometry of one diagram. Nodes and Flows are sorted
// by id; Lanes keep model order.
type Layout struct {
	Width, Height int
	Title         Point
	Lanes         []LaneBox
	Nodes         []NodeBox
	Flows         []FlowPath
}

func isGateway(kind string) bool { return kind == KindExclusive || kind == KindParallel }

// LayoutOf places a validated, normalized model. The algorithm is fixed and
// uses no randomness and no map iteration order:
//
//  1. Back edges are found by a depth-first search from the start that visits
//     outgoing flows in flow-id order. A column is the longest path from the
//     start over the remaining edges.
//  2. Each lane is a band. Nodes sharing a lane and column stack in rows
//     ordered by id. A lane is as tall as its tallest stack.
//  3. Between column cells lie gutters. A gutter starts with a label zone
//     that holds no vertical line, followed by slots. Each flow owns one
//     slot, ordered by flow id, in the two gutters its vertical segments may
//     use.
//  4. A flow leaves a start or task at the middle of its right side. It
//     leaves a gateway at the top when its target row is higher, at the
//     bottom when lower and at the right when equal, each flow at its own
//     exit. Every flow enters the middle of its target's left side.
//  5. A forward flow tries a route with one vertical segment next to its
//     source and, when it leaves rightwards, one next to its target. If a
//     route would cross a node it falls back to a channel above the first
//     lane or below the last one. A back edge always uses a channel.
//  6. A flow label sits above the flow's first horizontal run, inside the
//     label zone next to its source. That run is cut at the end of the label
//     zone so that the labelled piece is a segment no other flow shares.
func LayoutOf(m Model) Layout {
	nodes := m.Nodes
	idx := map[string]int{}
	for i, n := range nodes {
		idx[n.ID] = i
	}
	laneIdx := map[string]int{}
	for i, l := range m.Lanes {
		laneIdx[l.ID] = i
	}
	out := make([][]int, len(nodes)) // flow indexes by source, in flow-id order
	for fi, f := range m.Flows {
		out[idx[f.From]] = append(out[idx[f.From]], fi)
	}

	// 1. Back edges and columns.
	const (
		white = iota
		gray
		black
	)
	color := make([]int, len(nodes))
	back := make([]bool, len(m.Flows))
	var post []int
	var visit func(n int)
	visit = func(n int) {
		color[n] = gray
		for _, fi := range out[n] {
			t := idx[m.Flows[fi].To]
			switch color[t] {
			case white:
				visit(t)
			case gray:
				back[fi] = true
			}
		}
		color[n] = black
		post = append(post, n)
	}
	for i, n := range nodes {
		if n.Kind == KindStart {
			visit(i)
		}
	}
	col := make([]int, len(nodes))
	for i := len(post) - 1; i >= 0; i-- { // reverse postorder is topological
		n := post[i]
		for _, fi := range out[n] {
			if t := idx[m.Flows[fi].To]; !back[fi] && col[t] < col[n]+1 {
				col[t] = col[n] + 1
			}
		}
	}
	numCols := 0
	for _, c := range col {
		numCols = max(numCols, c+1)
	}

	// 2. Rows. nodes are sorted by id, so stacks are ordered by id.
	sub := make([]int, len(nodes))
	count := map[[2]int]int{}
	laneStack := make([]int, len(m.Lanes))
	for i, n := range nodes {
		key := [2]int{laneIdx[n.Lane], col[i]}
		sub[i] = count[key]
		count[key]++
		laneStack[key[0]] = max(laneStack[key[0]], count[key])
	}
	laneTop := make([]int, len(m.Lanes)+1)
	for i := range m.Lanes {
		laneTop[i+1] = laneTop[i] + laneStack[i]*rowH
	}
	areaH := laneTop[len(m.Lanes)]

	// Flow labels and exit overflow. A gateway has maxExits exits. When it
	// has more outgoing flows, the last exit is shared by every flow after
	// the first maxExits-1; such a flow carries its label beside its own
	// vertical segment instead, in a slot widened to hold it.
	labelText := make([]string, len(m.Flows))
	labelW := make([]int, len(m.Flows))
	overflow := make([]bool, len(m.Flows))
	for fi, f := range m.Flows {
		if f.Label != "" {
			labelText[fi] = truncate(f.Label, flowBudget)
			labelW[fi] = labelCharW*utf8.RuneCountInString(labelText[fi]) + 4
		}
	}
	for n := range nodes {
		if isGateway(nodes[n].Kind) && len(out[n]) > maxExits {
			for _, fi := range out[n][maxExits-1:] {
				overflow[fi] = true
			}
		}
	}

	// 3. Gutters. Gutter g lies left of column g; gutter numCols is the last.
	// A flow may place a vertical segment in gutter col(from)+1 or col(to).
	gutterOf := func(fi int) (ga, gb int) {
		f := m.Flows[fi]
		return col[idx[f.From]] + 1, col[idx[f.To]]
	}
	slots := make([]int, numCols+1) // width of the slots handed out so far
	padW := make([]int, numCols+1)  // width of the label zone
	offA := make([]int, len(m.Flows))
	offB := make([]int, len(m.Flows))
	for fi := range m.Flows {
		ga, gb := gutterOf(fi)
		offA[fi] = slots[ga]
		slots[ga] += slotStep
		switch {
		case labelW[fi] > 0 && overflow[fi]:
			slots[ga] += labelW[fi] + labelGap
		case labelW[fi] > 0:
			padW[ga] = max(padW[ga], labelW[fi]+labelGap+2)
		}
		if gb != ga {
			offB[fi] = slots[gb]
			slots[gb] += slotStep
		} else {
			offB[fi] = offA[fi]
		}
	}
	colX := make([]int, numCols)
	gutterLeft := make([]int, numCols+1)
	x := 0
	for g := 0; g <= numCols; g++ {
		gutterLeft[g] = x
		x += padW[g] + max(minSlotsW, 2*slotPad+slots[g])
		if g < numCols {
			colX[g] = x
			x += cellW
		}
	}
	bandW := x
	slotX := func(g, off int) int { return gutterLeft[g] + padW[g] + slotPad + off }

	// Nodes, in area coordinates (origin at the top left of the first lane's
	// drawing area).
	boxes := make([]NodeBox, len(nodes))
	for i, n := range nodes {
		cx := colX[col[i]] + cellW/2
		cy := laneTop[laneIdx[n.Lane]] + sub[i]*rowH + rowH/2
		w, h := nodeSize(n.Kind)
		boxes[i] = NodeBox{Node: n, Box: Box{cx - w/2, cy - h/2, w, h}, CX: cx, CY: cy}
	}

	// 4 and 5. Exits and routing, in flow-id order.
	paths := make([]FlowPath, len(m.Flows))
	used := make([][3][portsPerDir]bool, len(nodes))
	nTop, nBottom := 0, 0
	for fi, f := range m.Flows {
		si := idx[f.From]
		s, t := boxes[si], boxes[idx[f.To]]
		ga, gb := gutterOf(fi)
		tl := Point{t.Box.X, t.CY}
		xa, xb := slotX(ga, offA[fi]), slotX(gb, offB[fi])
		viaTop := s.CY+t.CY <= (areaH-s.CY)+(areaH-t.CY)

		// The exit: the points the flow starts with. It turns right at the
		// last of them.
		pre := []Point{{s.Box.X + s.Box.W, s.CY}}
		side := sideRight
		if isGateway(s.Node.Kind) {
			port := portsPerDir - 1
			if !overflow[fi] {
				switch {
				case back[fi] && viaTop, !back[fi] && t.CY < s.CY:
					side = sideUp
				case back[fi], t.CY > s.CY:
					side = sideDown
				}
				side, port = freeExit(&used[si], side, len(out[si]) > maxExits)
			}
			pre = exitPoints(s, side, port)
		}
		turn := pre[len(pre)-1]

		var pts []Point
		if !back[fi] {
			tries := []int{xa}
			if side == sideRight && !overflow[fi] {
				tries = append(tries, xb)
			}
			for _, x := range tries {
				c := simplify(append(append([]Point(nil), pre...), Point{x, turn.Y}, Point{x, tl.Y}, tl))
				if !hitsNode(c, boxes, si) {
					pts = c
					break
				}
			}
		}
		if pts == nil {
			var yc int
			if viaTop {
				yc = -(slotPad + slotStep*nTop)
				nTop++
			} else {
				yc = areaH + slotPad + slotStep*nBottom
				nBottom++
			}
			pts = simplify(append(append([]Point(nil), pre...),
				Point{xa, turn.Y}, Point{xa, yc}, Point{xb, yc}, Point{xb, tl.Y}, tl))
		}

		// 6. The label.
		var label *FlowLabel
		switch {
		case labelW[fi] == 0:
		case overflow[fi]:
			// Segment 1 is the flow's own vertical at xa. The label sits to
			// its right, just after the shared exit line.
			y := turn.Y + 2
			if pts[2].Y < pts[1].Y {
				y = turn.Y - 2 - labelH
			}
			label = &FlowLabel{labelText[fi], Box{xa + 4, y, labelW[fi], labelH}, 1}
		default:
			// The first horizontal run starts at the turn point. Cut it at
			// the end of the label zone.
			a := len(pre) - 1
			cut := Point{gutterLeft[ga] + padW[ga], turn.Y}
			pts = append(pts[:a+1], append([]Point{cut}, pts[a+1:]...)...)
			label = &FlowLabel{labelText[fi], Box{s.Box.X + s.Box.W + labelGap, turn.Y - 3 - labelH, labelW[fi], labelH}, a}
		}
		paths[fi] = FlowPath{Flow: f, Points: pts, Label: label}
	}

	// Absolute placement.
	topZone, bottomZone := 0, 0
	if nTop > 0 {
		topZone = slotPad + slotStep*nTop
	}
	if nBottom > 0 {
		bottomZone = slotPad + slotStep*nBottom
	}
	ox := margin + laneStripW
	oy := margin + titleH + topZone
	var l Layout
	l.Width = ox + bandW + margin
	l.Height = oy + areaH + bottomZone + margin
	l.Title = Point{margin, margin + 16}
	for i, ln := range m.Lanes {
		l.Lanes = append(l.Lanes, LaneBox{ln, Box{margin, oy + laneTop[i], laneStripW + bandW, laneTop[i+1] - laneTop[i]}})
	}
	for _, b := range boxes {
		b.Box.X += ox
		b.Box.Y += oy
		b.CX += ox
		b.CY += oy
		l.Nodes = append(l.Nodes, b)
	}
	for fi := range paths {
		p := &paths[fi]
		for i := range p.Points {
			p.Points[i].X += ox
			p.Points[i].Y += oy
		}
		if p.Label != nil {
			p.Label.Box.X += ox
			p.Label.Box.Y += oy
		}
	}
	l.Flows = paths
	return l
}

// freeExit hands out a gateway exit: the first free one on the wanted side,
// else on the right, top or bottom. When the gateway has more flows than
// exits, the last right exit is kept back for the flows that must share one.
func freeExit(used *[3][portsPerDir]bool, want int, keepShared bool) (side, port int) {
	for _, s := range []int{want, sideRight, sideUp, sideDown} {
		for p := 0; p < portsPerDir; p++ {
			if used[s][p] || keepShared && s == sideRight && p == portsPerDir-1 {
				continue
			}
			used[s][p] = true
			return s, p
		}
	}
	return sideRight, portsPerDir - 1 // unreachable: at most maxExits flows ask
}

// exitPoints returns the points a flow starts with when it leaves gateway s
// at the given exit. The first point lies on the diamond.
func exitPoints(s NodeBox, side, port int) []Point {
	r := s.Box.W / 2
	abs := func(n int) int { return max(n, -n) }
	switch side {
	case sideUp:
		dx := sideExitDX[port]
		return []Point{{s.CX + dx, s.CY - r + abs(dx)}, {s.CX + dx, s.CY + upCorridorY[port]}}
	case sideDown:
		dx := sideExitDX[port]
		return []Point{{s.CX + dx, s.CY + r - abs(dx)}, {s.CX + dx, s.CY + downCorridorY[port]}}
	}
	dy := rightExitDY[port]
	return []Point{{s.CX + r - abs(dy), s.CY + dy}}
}

func nodeSize(kind string) (w, h int) {
	switch kind {
	case KindStart, KindEnd:
		return eventD, eventD
	case KindTask:
		return taskW, taskH
	}
	return gatewayD, gatewayD
}

// simplify drops repeated points and points in the middle of a straight run.
func simplify(in []Point) []Point {
	var p []Point
	for _, q := range in {
		if len(p) > 0 && p[len(p)-1] == q {
			continue
		}
		p = append(p, q)
	}
	for i := 1; i < len(p)-1; {
		a, b, c := p[i-1], p[i], p[i+1]
		if (a.X == b.X && b.X == c.X) || (a.Y == b.Y && b.Y == c.Y) {
			p = append(p[:i], p[i+1:]...)
			continue
		}
		i++
	}
	return p
}

// hitsNode reports whether any segment crosses the interior of any node box.
// The first segment may start inside the box of its source, node src: a
// gateway exit lies on the diamond, inside the bounding box.
func hitsNode(p []Point, boxes []NodeBox, src int) bool {
	for i := 1; i < len(p); i++ {
		a, b := p[i-1], p[i]
		x0, x1 := min(a.X, b.X), max(a.X, b.X)
		y0, y1 := min(a.Y, b.Y), max(a.Y, b.Y)
		for ni, n := range boxes {
			if i == 1 && ni == src {
				continue
			}
			bx := n.Box
			if x0 < bx.X+bx.W && x1 > bx.X && y0 < bx.Y+bx.H && y1 > bx.Y {
				return true
			}
		}
	}
	return false
}
