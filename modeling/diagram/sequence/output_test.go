package sequence

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"

	"github.com/spik3r/clanker-foundry/modeling/diagram/internal/svgcheck"
)

// drawn is the geometry read back from an emitted SVG, with no use of the
// layout structs.
type drawn struct {
	width, height int
	partOrder     []string
	parts         map[string]svgcheck.Rect
	partKind      map[string]string
	lifeX         map[string][2]int // x, and y1 packed below
	lifeY         map[string][2]int
	steps         []drawnStep
	frames        []drawnFrame
}

type drawnStep struct {
	id, kind, from, to, style string
	over                      []string
	box                       svgcheck.Rect
	points                    [][2]int // message path
	marker                    string
	dashed                    bool
	labelX                    int
	label                     string
}

type drawnFrame struct {
	id, kind, first, last string
	box                   svgcheck.Rect
}

func readBack(t testing.TB, svg []byte) drawn {
	t.Helper()
	els, err := svgcheck.Parse(svg)
	if err != nil {
		t.Fatalf("svg is not well-formed XML: %v", err)
	}
	d := drawn{parts: map[string]svgcheck.Rect{}, partKind: map[string]string{}, lifeX: map[string][2]int{}, lifeY: map[string][2]int{}}
	atoi := func(s string) int {
		var n int
		if _, err := fmt.Sscanf(s, "%d", &n); err != nil {
			t.Fatalf("bad integer %q", s)
		}
		return n
	}
	for i, e := range els {
		switch {
		case e.Name == "svg":
			d.width, d.height = atoi(e.Attr["width"]), atoi(e.Attr["height"])
		case e.Name == "line" && e.Attr["data-kind"] == "lifeline":
			id := e.Attr["data-id"]
			if e.Attr["x1"] != e.Attr["x2"] {
				t.Errorf("lifeline %s is not vertical", id)
			}
			d.lifeX[id] = [2]int{atoi(e.Attr["x1"]), 0}
			d.lifeY[id] = [2]int{atoi(e.Attr["y1"]), atoi(e.Attr["y2"])}
		case e.Name == "g" && strings.HasPrefix(e.Attr["class"], "participant"):
			box, err := svgcheck.Box(e.Attr["data-box"])
			if err != nil {
				t.Fatal(err)
			}
			id := e.Attr["data-id"]
			d.partOrder = append(d.partOrder, id)
			d.parts[id] = box
			d.partKind[id] = e.Attr["data-kind"]
			if e.Attr["data-kind"] == "actor" {
				// a stick figure: a circle and a path, no rect
				var circle, rect bool
				for _, c := range els[i+1:] {
					if c.Depth <= e.Depth {
						break
					}
					circle = circle || c.Name == "circle"
					rect = rect || c.Name == "rect"
				}
				if !circle || rect {
					t.Errorf("actor %s is not drawn as a figure", id)
				}
			}
		case e.Name == "g" && e.Attr["data-kind"] == "message":
			box, _ := svgcheck.Box(e.Attr["data-box"])
			s := drawnStep{id: e.Attr["data-id"], kind: "message", from: e.Attr["data-from"], to: e.Attr["data-to"], style: e.Attr["data-style"], box: box}
			for _, c := range els[i+1:] {
				if c.Depth <= e.Depth {
					break
				}
				switch c.Name {
				case "path":
					pts, err := svgcheck.Points(c.Attr["d"])
					if err != nil {
						t.Fatal(err)
					}
					s.points = pts
					s.marker = c.Attr["marker-end"]
					s.dashed = strings.Contains(c.Attr["class"], "msg-reply")
				case "text":
					s.labelX = atoi(c.Attr["x"])
					s.label = c.Text
				}
			}
			d.steps = append(d.steps, s)
		case e.Name == "g" && e.Attr["data-kind"] == "note":
			box, _ := svgcheck.Box(e.Attr["data-box"])
			d.steps = append(d.steps, drawnStep{id: e.Attr["data-id"], kind: "note", over: strings.Fields(e.Attr["data-over"]), box: box, label: e.Text})
		case e.Name == "g" && strings.HasPrefix(e.Attr["class"], "frame"):
			box, _ := svgcheck.Box(e.Attr["data-box"])
			d.frames = append(d.frames, drawnFrame{e.Attr["data-id"], e.Attr["data-kind"], e.Attr["data-first"], e.Attr["data-last"], box})
		}
	}
	return d
}

// checkGeometry verifies the layout rules on an emitted drawing.
func checkGeometry(t testing.TB, m Model, svg []byte) {
	t.Helper()
	d := readBack(t, svg)
	canvas := svgcheck.Rect{X: 0, Y: 0, W: d.width, H: d.height}

	// Participants: model order, left to right, shared baseline, lifelines
	// centred under the boxes.
	if len(d.partOrder) != len(m.Participants) {
		t.Fatalf("%d participants drawn, want %d", len(d.partOrder), len(m.Participants))
	}
	var baseline, top int
	for i, p := range m.Participants {
		if d.partOrder[i] != p.ID {
			t.Fatalf("participant %d is %s, want %s", i, d.partOrder[i], p.ID)
		}
		if d.partKind[p.ID] != p.Kind {
			t.Errorf("participant %s kind %s", p.ID, d.partKind[p.ID])
		}
		b := d.parts[p.ID]
		if i == 0 {
			baseline, top = b.Y+b.H, b.Y
		} else {
			if b.Y+b.H != baseline || b.Y != top {
				t.Errorf("participant %s is not on the shared baseline", p.ID)
			}
			if prev := d.parts[m.Participants[i-1].ID]; b.X < prev.X+prev.W {
				t.Errorf("participant %s overlaps or precedes %s", p.ID, m.Participants[i-1].ID)
			}
		}
		if !canvas.Contains(b) {
			t.Errorf("participant %s outside the canvas", p.ID)
		}
		lx, ly := d.lifeX[p.ID][0], d.lifeY[p.ID]
		if lx != b.X+b.W/2 {
			t.Errorf("lifeline of %s at x=%d, box centre %d", p.ID, lx, b.X+b.W/2)
		}
		if ly[0] != baseline || ly[1] <= ly[0] || ly[1] > d.height {
			t.Errorf("lifeline of %s runs y=%d..%d, baseline %d", p.ID, ly[0], ly[1], baseline)
		}
	}
	life := func(id string) int { return d.lifeX[id][0] }

	// Steps: model order, strictly increasing y, no vertical overlap.
	if len(d.steps) != len(m.Steps) {
		t.Fatalf("%d steps drawn, want %d", len(d.steps), len(m.Steps))
	}
	byID := map[string]drawnStep{}
	for i, s := range m.Steps {
		ds := d.steps[i]
		byID[ds.id] = ds
		if ds.id != s.ID || ds.kind != s.Kind {
			t.Fatalf("step %d is %s/%s, want %s/%s", i, ds.id, ds.kind, s.ID, s.Kind)
		}
		if ds.box.Y < baseline {
			t.Errorf("step %s starts above the header baseline", s.ID)
		}
		if i > 0 {
			prev := d.steps[i-1]
			if ds.box.Y <= prev.box.Y || ds.box.Y < prev.box.Y+prev.box.H {
				t.Errorf("step %s (y=%d) does not follow step %s (y=%d..%d)", s.ID, ds.box.Y, prev.id, prev.box.Y, prev.box.Y+prev.box.H)
			}
		}
		if !canvas.Contains(ds.box) {
			t.Errorf("step %s outside the canvas: %+v in %dx%d", s.ID, ds.box, d.width, d.height)
		}
		if ds.box.Y+ds.box.H > d.lifeY[m.Participants[0].ID][1] {
			t.Errorf("step %s below the end of the lifelines", s.ID)
		}
		switch s.Kind {
		case StepMessage:
			pts := ds.points
			if ds.from != s.From || ds.to != s.To || ds.style != s.Style {
				t.Errorf("message %s attributes %s %s %s", s.ID, ds.from, ds.to, ds.style)
			}
			if s.From == s.To {
				// A loop to the right of the lifeline: out, down, back.
				if len(pts) != 4 || pts[0][0] != life(s.From) || pts[3][0] != life(s.From) ||
					pts[0][1] != pts[1][1] || pts[1][0] != pts[2][0] || pts[2][1] != pts[3][1] ||
					pts[1][0] <= pts[0][0] || pts[3][1] <= pts[0][1] {
					t.Errorf("self-message %s is not a loop right of its lifeline: %v", s.ID, pts)
				}
				if ds.labelX <= pts[1][0] {
					t.Errorf("self-message %s label is not right of the loop", s.ID)
				}
			} else {
				if len(pts) != 2 || pts[0][1] != pts[1][1] || pts[0][0] != life(s.From) || pts[1][0] != life(s.To) {
					t.Errorf("message %s does not run horizontally from lifeline %s to lifeline %s: %v", s.ID, s.From, s.To, pts)
				}
				if want := (life(s.From) + life(s.To)) / 2; ds.labelX != want {
					t.Errorf("message %s label centred at %d, want %d", s.ID, ds.labelX, want)
				}
				if ds.box.Y > pts[0][1] || ds.box.Y+ds.box.H < pts[0][1] {
					t.Errorf("message %s box does not contain its arrow", s.ID)
				}
				// The label must fit between the lifelines by the character budget.
				lo, hi := min(life(s.From), life(s.To)), max(life(s.From), life(s.To))
				if w := runes(ds.label) * charW; w > hi-lo {
					t.Errorf("message %s label of %d characters does not fit %d..%d", s.ID, runes(ds.label), lo, hi)
				}
			}
			wantMarker := "url(#head-open)"
			if s.Style == StyleSync {
				wantMarker = "url(#head-filled)"
			}
			if ds.marker != wantMarker || ds.dashed != (s.Style == StyleReply) {
				t.Errorf("message %s style %s drawn with marker %s dashed %v", s.ID, s.Style, ds.marker, ds.dashed)
			}
			if runes(ds.label) > msgBudget {
				t.Errorf("message %s label has %d characters, budget %d", s.ID, runes(ds.label), msgBudget)
			}
			if r := []rune(s.Label); len(r) > msgBudget && !strings.HasSuffix(ds.label, "…") && s.From != s.To {
				t.Errorf("message %s long label not truncated with an ellipsis: %q", s.ID, ds.label)
			}
		case StepNote:
			lo, hi := life(s.Over[0]), life(s.Over[0])
			for _, o := range s.Over {
				lo, hi = min(lo, life(o)), max(hi, life(o))
			}
			if ds.box.X >= lo || ds.box.X+ds.box.W <= hi {
				t.Errorf("note %s box %+v does not span lifelines %d..%d", s.ID, ds.box, lo, hi)
			}
		}
		// A step never overlaps a participant header.
		for _, id := range d.partOrder {
			if ds.box.Overlap(d.parts[id]) {
				t.Errorf("step %s overlaps the header of %s", s.ID, id)
			}
		}
	}
	// Notes wrap to at most three lines.
	for _, st := range readNoteLines(t, svg) {
		if st > noteLines {
			t.Errorf("a note has %d lines", st)
		}
	}

	// Frames: group order is by id; each frame encloses its steps and only its
	// steps, spans the participants involved and clears the headers.
	if len(d.frames) != len(m.Groups) {
		t.Fatalf("%d frames drawn, want %d", len(d.frames), len(m.Groups))
	}
	idx := map[string]int{}
	for i, s := range m.Steps {
		idx[s.ID] = i
	}
	for i, g := range m.Groups {
		f := d.frames[i]
		if f.id != g.ID || f.kind != g.Kind || f.first != g.First || f.last != g.Last {
			t.Fatalf("frame %d is %+v, want group %+v", i, f, g)
		}
		if !canvas.Contains(f.box) || f.box.Y < baseline {
			t.Errorf("frame %s outside the canvas or over the header", g.ID)
		}
		for j, s := range m.Steps {
			in := j >= idx[g.First] && j <= idx[g.Last]
			sb := d.steps[j].box
			switch {
			case in && !f.box.Contains(sb):
				t.Errorf("frame %s does not enclose step %s", g.ID, s.ID)
			case !in && f.box.Overlap(sb):
				t.Errorf("frame %s overlaps step %s outside its range", g.ID, s.ID)
			}
		}
		// Every participant involved lies inside the frame's x range.
		for j := idx[g.First]; j <= idx[g.Last]; j++ {
			s := m.Steps[j]
			ids := s.Over
			if s.Kind == StepMessage {
				ids = []string{s.From, s.To}
			}
			for _, p := range ids {
				if x := life(p); x < f.box.X || x > f.box.X+f.box.W {
					t.Errorf("frame %s does not span participant %s of step %s", g.ID, p, s.ID)
				}
			}
		}
		for _, o := range d.frames[i+1:] {
			if f.box.Overlap(o.box) {
				t.Errorf("frames %s and %s overlap", f.id, o.id)
			}
		}
	}
}

// readNoteLines returns the number of tspans in each note.
func readNoteLines(t testing.TB, svg []byte) []int {
	els, err := svgcheck.Parse(svg)
	if err != nil {
		t.Fatal(err)
	}
	var out []int
	for i, e := range els {
		if e.Name == "g" && e.Attr["data-kind"] == "note" {
			n := 0
			for _, c := range els[i+1:] {
				if c.Depth <= e.Depth {
					break
				}
				if c.Name == "tspan" {
					n++
				}
			}
			out = append(out, n)
		}
	}
	return out
}

func build(t testing.TB, m Model) []byte {
	t.Helper()
	if err := m.Validate(); err != nil {
		t.Fatalf("model invalid: %v", err)
	}
	n := m.Normalize()
	return SVG(n, LayoutOf(n), testGen)
}

func TestExampleGeometry(t *testing.T) {
	for _, name := range exampleNames {
		t.Run(name, func(t *testing.T) {
			m := mustDecode(t, readExample(t, name))
			svg := build(t, m)
			checkGeometry(t, m, svg)
			if err := svgcheck.Inert(svg); err != nil {
				t.Error(err)
			}
		})
	}
	t.Run("base model", func(t *testing.T) {
		m := baseModel().Normalize()
		checkGeometry(t, m, build(t, m))
	})
}

// randomModel builds a valid model from seed: 2..12 participants, 1..80
// steps, notes, self-messages, long labels and non-overlapping groups.
func randomModel(seed int64) Model {
	r := rand.New(rand.NewSource(seed))
	words := []string{"order", "pay", "retry", "queue", "ok", "a much longer phrase to force truncation of the label", "x", "Verification", "ü", "<b>&"}
	text := func() string {
		var parts []string
		for i := r.Intn(8) + 1; i > 0; i-- {
			parts = append(parts, words[r.Intn(len(words))])
		}
		s := strings.Join(parts, " ")
		for len(s) > 120 {
			r := []rune(s)
			s = strings.TrimSpace(string(r[:len(r)-1]))
		}
		return s
	}
	m := Model{FormatVersion: Format, ProfileVersion: Profile, ID: "Rand", Title: text()}
	np := 2 + r.Intn(MaxParticipants-1)
	for i := 0; i < np; i++ {
		k := KindParticipant
		if r.Intn(4) == 0 {
			k = KindActor
		}
		m.Participants = append(m.Participants, Participant{fmt.Sprintf("P%d", i), text(), k})
	}
	ns := 1 + r.Intn(MaxSteps)
	for i := 0; i < ns; i++ {
		id := fmt.Sprintf("S%d", i)
		if r.Intn(5) == 0 {
			a := r.Intn(np)
			over := []string{m.Participants[a].ID}
			if r.Intn(2) == 0 {
				if b := r.Intn(np); b != a {
					over = append(over, m.Participants[b].ID)
				}
			}
			m.Steps = append(m.Steps, Step{ID: id, Kind: StepNote, Over: over, Label: text()})
			continue
		}
		a, b := r.Intn(np), r.Intn(np)
		if r.Intn(6) == 0 {
			b = a
		}
		style := []string{StyleSync, StyleAsync, StyleReply}[r.Intn(3)]
		m.Steps = append(m.Steps, Step{ID: id, Kind: StepMessage, From: m.Participants[a].ID, To: m.Participants[b].ID, Label: text(), Style: style})
	}
	for i := 0; i < ns && len(m.Groups) < MaxGroups; {
		if r.Intn(3) == 0 {
			last := i + r.Intn(min(4, ns-i))
			m.Groups = append(m.Groups, Group{fmt.Sprintf("G%d", len(m.Groups)), []string{GroupAlt, GroupOpt, GroupLoop}[r.Intn(3)], text(), m.Steps[i].ID, m.Steps[last].ID})
			i = last + 1
		} else {
			i++
		}
	}
	return m
}

func TestRandomModelGeometry(t *testing.T) {
	for seed := int64(1); seed <= 150; seed++ {
		m := randomModel(seed)
		if err := m.Validate(); err != nil {
			t.Fatalf("seed %d: generated model invalid: %v", seed, err)
		}
		m = m.Normalize()
		svg := build(t, m)
		t.Run(fmt.Sprintf("seed%d", seed), func(t *testing.T) {
			checkGeometry(t, m, svg)
			if err := svgcheck.Inert(svg); err != nil {
				t.Error(err)
			}
		})
	}
}

func TestDeterminism(t *testing.T) {
	for _, name := range exampleNames {
		raw := readExample(t, name)
		a, err := Profiler{}.Generate(raw, testGen)
		if err != nil {
			t.Fatal(err)
		}
		b, _ := Profiler{}.Generate(raw, testGen)
		for f, data := range a.Files {
			if string(data) != string(b.Files[f]) {
				t.Errorf("%s: %s differs between runs", name, f)
			}
		}
		if string(a.Canonical) != string(b.Canonical) {
			t.Errorf("%s: canonical form differs between runs", name)
		}
	}
	for seed := int64(1); seed <= 20; seed++ {
		a, b := build(t, randomModel(seed)), build(t, randomModel(seed))
		if string(a) != string(b) {
			t.Errorf("seed %d: output differs between runs", seed)
		}
	}
}

func TestOutputFiles(t *testing.T) {
	gen, err := Profiler{}.Generate(readExample(t, "checkout-sequence"), testGen)
	if err != nil {
		t.Fatal(err)
	}
	if len(gen.Files) != 2 || gen.Files[FileSVG] == nil || gen.Files[FileModel] == nil {
		t.Fatalf("files %v", len(gen.Files))
	}
	mj := string(gen.Files[FileModel])
	if !strings.HasSuffix(mj, "}\n") || !strings.Contains(mj, "\n  \"id\"") {
		t.Errorf("model.json is not indented with a trailing newline")
	}
	svg := string(gen.Files[FileSVG])
	for _, want := range []string{`viewBox="0 0 `, "<title>Checkout</title>", "generated by diagramgen 1.2.1", `data-kind="alt"`, `data-kind="actor"`, `data-kind="note"`, `data-kind="lifeline"`} {
		if !strings.Contains(svg, want) {
			t.Errorf("svg lacks %s", want)
		}
	}
}

// TestHostileText puts markup, entities and script lookalikes in every text
// field and checks that the picture stays inert and well formed.
func TestHostileText(t *testing.T) {
	hostile := []string{
		`<script>alert(1)</script>`,
		`"><svg onload=alert(1)>`,
		`</text><foreignObject><body onload=x></foreignObject>`,
		`a & b &amp; &#x3C;`,
		`' onmouseover='x`,
		`javascript:alert(1) http://evil.example/x //evil.example`,
		`url(http://evil.example/a) @import "x"`,
		`]]> <![CDATA[ <!DOCTYPE x [<!ENTITY e "x">]>`,
		`<use href="http://evil.example/#a"/>`,
		"emoji 😀 and é and 中文",
	}
	for _, h := range hostile {
		m := baseModel()
		m.Title = h
		for i := range m.Participants {
			m.Participants[i].Label = h
		}
		for i := range m.Steps {
			m.Steps[i].Label = h
		}
		for i := range m.Groups {
			m.Groups[i].Label = h
		}
		m = m.Normalize()
		svg := build(t, m)
		if err := svgcheck.Inert(svg); err != nil {
			t.Errorf("%q: %v", h, err)
		}
		checkGeometry(t, m, svg)
		// Ids that look hostile are refused before drawing.
		bad := baseModel()
		bad.Participants[0].ID = `A"><script>`
		if bad.Validate() == nil {
			t.Errorf("hostile id accepted")
		}
	}
	// The text survives escaping: it comes back out of the parsed document.
	m := baseModel()
	m.Steps[0].Label = `a<b>&"c'`
	m = m.Normalize()
	d := readBack(t, build(t, m))
	if d.steps[0].label != `a<b>&"c'` {
		t.Errorf("message label read back as %q", d.steps[0].label)
	}
}

// TestWidening checks that a long label between adjacent columns widens the
// gap, and a short one does not.
func TestWidening(t *testing.T) {
	m := baseModel()
	m.Groups = nil
	short := LayoutOf(m.Normalize())
	m.Steps[0].Label = strings.Repeat("w", 39)
	long := LayoutOf(m.Normalize())
	gap := func(l Layout) int { return l.Participants[1].CX - l.Participants[0].CX }
	if gap(short) != minGap || gap(long) <= minGap {
		t.Errorf("gaps short=%d long=%d", gap(short), gap(long))
	}
	if gap(long) < 39*charW {
		t.Errorf("gap %d cannot hold the label", gap(long))
	}
}

func TestSelfMessageOnLastColumn(t *testing.T) {
	m := baseModel()
	m.Groups = nil
	m.Steps = []Step{{ID: "A", Kind: StepMessage, From: "Pay", To: "Pay", Label: strings.Repeat("z", 60), Style: StyleAsync}}
	m = m.Normalize()
	checkGeometry(t, m, build(t, m))
}
