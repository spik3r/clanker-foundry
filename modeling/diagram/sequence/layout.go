package sequence

import "github.com/spik3r/clanker-foundry/modeling/diagram/internal/svgx"

// Layout constants. All coordinates are integers. Text width is estimated from
// character counts (charW per character); no font is measured, so a line of
// wide characters can still overflow its shape.
const (
	margin      = 24
	headerTop   = 52  // y of the participant header band
	bandH       = 64  // height of the header band
	boxW        = 128 // width of a participant box
	firstCX     = 120 // centre of the first lifeline
	minGap      = 160 // least distance between adjacent lifelines
	charW       = 7
	lineH       = 14
	labelPad    = 24 // extra room around a message label
	msgBudget   = 40 // characters of a message label
	selfBudget  = 30 // characters of a self-message label
	loopW       = 36 // width of a self-message loop
	noteHalfW   = 80 // half width of a note over one participant
	noteOverPad = 60 // a note over two participants extends this far past each lifeline
	noteLines   = 3
	headBudget  = 16 // characters per line in a participant box
	frameMargin = 14
	frameTabH   = 16
	groupPadTop = 22 // room above the first step of a group for its tab
	groupPadBot = 10
	rowMessage  = 44
	rowSelf     = 56
	titleCharPx = 9
)

// Rect is an axis-aligned rectangle.
type Rect struct{ X, Y, W, H int }

// ParticipantBox is the drawn header of one participant.
type ParticipantBox struct {
	P   Participant
	CX  int
	Box Rect
}

// StepGeom is the drawn geometry of one step. Box encloses what is drawn for
// the step; Row is the horizontal band reserved for it.
type StepGeom struct {
	S   Step
	Box Rect
	Row Rect
	// Message: Y is the arrow's y; X1 and X2 are the lifeline x of From and To.
	Y, X1, X2 int
	// Text lines (message: one line; note: up to noteLines).
	Lines []string
	// Note: budget in characters per line.
	LabelX int
}

// FrameGeom is the drawn geometry of one group frame.
type FrameGeom struct {
	G     Group
	Box   Rect
	Tab   Rect
	Label string
}

// Layout is the complete geometry of a sequence diagram.
type Layout struct {
	Width, Height int
	Title         struct{ X, Y int }
	Participants  []ParticipantBox
	LifelineY1    int
	LifelineY2    int
	Steps         []StepGeom
	Frames        []FrameGeom
}

func runes(s string) int { return len([]rune(s)) }

// LayoutOf computes the layout of a validated, normalized model.
func LayoutOf(m Model) Layout {
	var l Layout
	n := len(m.Participants)
	col := map[string]int{}
	for i, p := range m.Participants {
		col[p.ID] = i
	}

	// Column gaps: widened for message labels that need more room.
	gap := make([]int, n-1)
	for i := range gap {
		gap[i] = minGap
	}
	span := func(a, b int) int {
		s := 0
		for i := a; i < b; i++ {
			s += gap[i]
		}
		return s
	}
	widen := func(a, b, need int) {
		if b <= a {
			return
		}
		d := need - span(a, b)
		if d <= 0 {
			return
		}
		k := b - a
		for i := a; i < b; i++ {
			gap[i] += d / k
			if i-a < d%k {
				gap[i]++
			}
		}
	}
	for _, s := range m.Steps {
		if s.Kind != StepMessage {
			continue
		}
		a, b := col[s.From], col[s.To]
		if a == b {
			if a < n-1 {
				w := runes(svgx.Truncate(s.Label, selfBudget))*charW + loopW + 8 + labelPad/2
				widen(a, a+1, w)
			}
			continue
		}
		if a > b {
			a, b = b, a
		}
		widen(a, b, runes(svgx.Truncate(s.Label, msgBudget))*charW+labelPad)
	}
	cx := make([]int, n)
	cx[0] = firstCX
	for i := 1; i < n; i++ {
		cx[i] = cx[i-1] + gap[i-1]
	}
	for i, p := range m.Participants {
		l.Participants = append(l.Participants, ParticipantBox{
			P: p, CX: cx[i], Box: Rect{cx[i] - boxW/2, headerTop, boxW, bandH},
		})
	}

	// Which steps start or end a group.
	first, last := map[string]bool{}, map[string]bool{}
	for _, g := range m.Groups {
		first[g.First], last[g.Last] = true, true
	}

	y := headerTop + bandH + 16
	maxRight := cx[n-1] + boxW/2
	for _, s := range m.Steps {
		pt, pb := 0, 0
		if first[s.ID] {
			pt = groupPadTop
		}
		if last[s.ID] {
			pb = groupPadBot
		}
		g := StepGeom{S: s}
		var base int
		switch s.Kind {
		case StepMessage:
			a, b := col[s.From], col[s.To]
			g.X1, g.X2 = cx[a], cx[b]
			if a == b {
				base = rowSelf
				y1 := y + pt + 24
				g.Y = y1
				txt := svgx.Truncate(s.Label, selfBudget)
				g.Lines = []string{txt}
				g.LabelX = cx[a] + loopW + 8
				right := g.LabelX + runes(txt)*charW
				g.Box = Rect{cx[a], y1 - 4, right - cx[a], 16 + 8}
			} else {
				base = rowMessage
				g.Y = y + pt + 32
				txt := svgx.Truncate(s.Label, msgBudget)
				g.Lines = []string{txt}
				g.LabelX = (cx[a] + cx[b]) / 2
				lo, hi := min(cx[a], cx[b]), max(cx[a], cx[b])
				g.Box = Rect{lo, g.Y - 18, hi - lo, 22}
			}
		default:
			a, b := col[s.Over[0]], col[s.Over[0]]
			if len(s.Over) == 2 {
				b = col[s.Over[1]]
			}
			if a > b {
				a, b = b, a
			}
			var x1, x2 int
			if a == b {
				x1, x2 = cx[a]-noteHalfW, cx[a]+noteHalfW
			} else {
				x1, x2 = cx[a]-noteOverPad, cx[b]+noteOverPad
			}
			g.Lines = svgx.Wrap(s.Label, (x2-x1-16)/charW, noteLines)
			h := len(g.Lines)*lineH + 10
			base = h + 16
			g.Box = Rect{x1, y + pt + 8, x2 - x1, h}
		}
		h := pt + base + pb
		g.Row = Rect{0, y, 0, h} // x extent is filled in below
		l.Steps = append(l.Steps, g)
		maxRight = max(maxRight, g.Box.X+g.Box.W)
		y += h
	}
	bottom := y + 8
	l.LifelineY1 = headerTop + bandH
	l.LifelineY2 = bottom

	// Frames enclose the steps of a group.
	idx := map[string]int{}
	for i, s := range m.Steps {
		idx[s.ID] = i
	}
	for _, g := range m.Groups {
		f, la := idx[g.First], idx[g.Last]
		x1, x2 := l.Steps[f].Box.X, l.Steps[f].Box.X+l.Steps[f].Box.W
		for i := f; i <= la; i++ {
			x1 = min(x1, l.Steps[i].Box.X)
			x2 = max(x2, l.Steps[i].Box.X+l.Steps[i].Box.W)
		}
		x1 -= frameMargin
		x2 += frameMargin
		label := svgx.Truncate(g.Label, msgBudget)
		tabW := runes(g.Kind)*charW + 16
		x2 = max(x2, x1+tabW+6+(runes(label)+2)*charW+8)
		top := l.Steps[f].Row.Y
		bot := l.Steps[la].Row.Y + l.Steps[la].Row.H
		l.Frames = append(l.Frames, FrameGeom{
			G: g, Box: Rect{x1, top, x2 - x1, bot - top}, Tab: Rect{x1, top, tabW, frameTabH}, Label: label,
		})
		maxRight = max(maxRight, x2)
	}

	l.Width = maxRight + margin
	l.Height = bottom + margin
	for i := range l.Steps {
		l.Steps[i].Row.X, l.Steps[i].Row.W = 0, l.Width
	}
	l.Title.X, l.Title.Y = margin, 30
	return l
}
