package process

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/spik3r/clanker-foundry/modeling/diagram/contract"
)

var testGen = contract.Generator{ID: "diagramgen", Version: "1.0.0", Toolchain: "go1.27.4"}

// exampleNames are the models under ../examples.
var exampleNames = []string{"order", "parallel", "rework"}

func readExample(t *testing.T, name string) []byte {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("..", "examples", name+".json"))
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func mustDecode(t *testing.T, raw []byte) Model {
	t.Helper()
	m, err := Decode(raw)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	return m
}

// baseModel is the order example, built in code so tests can mutate it.
func baseModel() Model {
	return Model{
		FormatVersion: Format, ProfileVersion: Profile, ID: "OrderHandling", Title: "Order handling",
		Lanes: []Lane{{"Sales", "Sales"}, {"Warehouse", "Warehouse"}},
		Nodes: []Node{
			{"Start", KindStart, "Order received", "Sales"},
			{"Check", KindTask, "Check stock", "Warehouse"},
			{"InStock", KindExclusive, "In stock?", "Warehouse"},
			{"Ship", KindTask, "Ship order", "Warehouse"},
			{"Reject", KindTask, "Notify customer", "Sales"},
			{"Done", KindEnd, "Closed", "Sales"},
		},
		Flows: []Flow{
			{"F1", "Start", "Check", ""},
			{"F2", "Check", "InStock", ""},
			{"F3", "InStock", "Ship", "yes"},
			{"F4", "InStock", "Reject", "no"},
			{"F5", "Ship", "Done", ""},
			{"F6", "Reject", "Done", ""},
		},
	}
}

func marshal(t *testing.T, m Model) []byte {
	t.Helper()
	b, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

// rect is an axis-aligned rectangle used by the geometry checks.
type rect struct{ x, y, w, h int }

// geo is the geometry of a drawing, whether taken from a Layout or parsed
// back out of emitted SVG.
type geo struct {
	width, height int
	lanes         map[string]rect
	laneOrder     []string
	nodes         map[string]rect
	kinds         map[string]string
	flows         map[string]geoFlow
	labels        map[string][]geoLabel // by flow id
}

type geoLabel struct {
	box    rect
	anchor int
	text   string
}

type geoFlow struct {
	from, to string
	pts      []Point
}

func geoOfLayout(l Layout) geo {
	g := geo{width: l.Width, height: l.Height, lanes: map[string]rect{}, nodes: map[string]rect{}, kinds: map[string]string{}, flows: map[string]geoFlow{}, labels: map[string][]geoLabel{}}
	for _, ln := range l.Lanes {
		g.lanes[ln.Lane.ID] = rect{ln.Box.X, ln.Box.Y, ln.Box.W, ln.Box.H}
		g.laneOrder = append(g.laneOrder, ln.Lane.ID)
	}
	for _, n := range l.Nodes {
		g.nodes[n.Node.ID] = rect{n.Box.X, n.Box.Y, n.Box.W, n.Box.H}
		g.kinds[n.Node.ID] = n.Node.Kind
	}
	for _, f := range l.Flows {
		g.flows[f.Flow.ID] = geoFlow{f.Flow.From, f.Flow.To, f.Points}
		if f.Label != nil {
			bx := f.Label.Box
			g.labels[f.Flow.ID] = append(g.labels[f.Flow.ID], geoLabel{rect{bx.X, bx.Y, bx.W, bx.H}, f.Label.Anchor, f.Label.Text})
		}
	}
	return g
}

// checkGeometry verifies the layout promises against model m.
func checkGeometry(t *testing.T, m Model, g geo) {
	t.Helper()
	if len(g.nodes) != len(m.Nodes) || len(g.flows) != len(m.Flows) || len(g.lanes) != len(m.Lanes) {
		t.Fatalf("drawing has %d nodes, %d flows, %d lanes; model has %d, %d, %d",
			len(g.nodes), len(g.flows), len(g.lanes), len(m.Nodes), len(m.Flows), len(m.Lanes))
	}
	// Lanes keep model order, are stacked without gaps and share x extents.
	prevBottom := -1
	for i, ln := range m.Lanes {
		r, ok := g.lanes[ln.ID]
		if !ok {
			t.Fatalf("lane %s missing", ln.ID)
		}
		if i > 0 && r.y != prevBottom {
			t.Errorf("lane %s starts at %d, previous ended at %d", ln.ID, r.y, prevBottom)
		}
		prevBottom = r.y + r.h
	}
	// Nodes inside their lane band and inside the viewBox; no overlap.
	ids := make([]string, 0, len(m.Nodes))
	for _, n := range m.Nodes {
		ids = append(ids, n.ID)
		b, ok := g.nodes[n.ID]
		if !ok {
			t.Fatalf("node %s missing", n.ID)
		}
		if g.kinds[n.ID] != n.Kind {
			t.Errorf("node %s kind %q, want %q", n.ID, g.kinds[n.ID], n.Kind)
		}
		lane := g.lanes[n.Lane]
		if b.x < lane.x || b.y < lane.y || b.x+b.w > lane.x+lane.w || b.y+b.h > lane.y+lane.h {
			t.Errorf("node %s %v lies outside lane %s %v", n.ID, b, n.Lane, lane)
		}
	}
	for i, a := range ids {
		for _, b := range ids[i+1:] {
			ra, rb := g.nodes[a], g.nodes[b]
			if ra.x < rb.x+rb.w && rb.x < ra.x+ra.w && ra.y < rb.y+rb.h && rb.y < ra.y+ra.h {
				t.Errorf("nodes %s %v and %s %v overlap", a, ra, b, rb)
			}
		}
	}
	areaTop, areaBottom := g.lanes[m.Lanes[0].ID].y, prevBottom
	type seg struct {
		flow       string
		a, b       Point
		horizontal bool
	}
	var segs []seg
	for _, f := range m.Flows {
		gf, ok := g.flows[f.ID]
		if !ok {
			t.Fatalf("flow %s missing", f.ID)
		}
		if gf.from != f.From || gf.to != f.To {
			t.Errorf("flow %s runs %s->%s, want %s->%s", f.ID, gf.from, gf.to, f.From, f.To)
		}
		if len(gf.pts) < 2 {
			t.Errorf("flow %s has %d points", f.ID, len(gf.pts))
			continue
		}
		s, e := g.nodes[f.From], g.nodes[f.To]
		first, last := gf.pts[0], gf.pts[len(gf.pts)-1]
		if isGateway(g.kinds[f.From]) {
			// A gateway exit lies on the diamond itself.
			cx, cy, r := s.x+s.w/2, s.y+s.h/2, s.w/2
			if d := max(first.X-cx, cx-first.X) + max(first.Y-cy, cy-first.Y); d != r || first.X < cx-r/2 {
				t.Errorf("flow %s starts at %v, not on the top, right or bottom of diamond %s %v", f.ID, first, f.From, s)
			}
		} else if first.X != s.x+s.w || first.Y != s.y+s.h/2 {
			t.Errorf("flow %s starts at %v, not at the middle of the right side of %s %v", f.ID, first, f.From, s)
		}
		if last.X != e.x || last.Y != e.y+e.h/2 {
			t.Errorf("flow %s ends at %v, not at the middle of the left side of %s %v", f.ID, last, f.To, e)
		}
		if n := len(gf.pts); gf.pts[n-2].Y != last.Y || gf.pts[n-2].X >= last.X {
			t.Errorf("flow %s does not enter %s from the left: %v", f.ID, f.To, gf.pts)
		}
		for i, p := range gf.pts {
			if p.X < 0 || p.Y < 0 || p.X > g.width || p.Y > g.height {
				t.Errorf("flow %s point %v lies outside the %dx%d canvas", f.ID, p, g.width, g.height)
			}
			if i == 0 {
				continue
			}
			q := gf.pts[i-1]
			if p.X != q.X && p.Y != q.Y {
				t.Errorf("flow %s segment %v-%v is not axis aligned", f.ID, q, p)
			}
			if p == q {
				t.Errorf("flow %s has a zero-length segment at %v", f.ID, p)
			}
			segs = append(segs, seg{f.ID, q, p, p.Y == q.Y})
			x0, x1 := min(p.X, q.X), max(p.X, q.X)
			y0, y1 := min(p.Y, q.Y), max(p.Y, q.Y)
			for _, id := range ids {
				n := g.nodes[id]
				if i == 1 && id == f.From && isGateway(g.kinds[id]) {
					continue // the exit lies on the diamond, inside its box
				}
				if x0 < n.x+n.w && x1 > n.x && y0 < n.y+n.h && y1 > n.y {
					t.Errorf("flow %s segment %v-%v passes through the interior of node %s %v", f.ID, q, p, id, n)
				}
			}
		}
	}
	// Segments of different flows never coincide: not vertical ones, and not
	// horizontal ones in a channel outside the lanes. (Flows meeting at one
	// node legitimately share the stub at that node.)
	for i, a := range segs {
		for _, b := range segs[i+1:] {
			if a.flow == b.flow || a.horizontal != b.horizontal {
				continue
			}
			if a.horizontal {
				if a.a.Y != b.a.Y || (a.a.Y >= areaTop && a.a.Y <= areaBottom) {
					continue
				}
				if min(a.a.X, a.b.X) < max(b.a.X, b.b.X) && min(b.a.X, b.b.X) < max(a.a.X, a.b.X) {
					t.Errorf("channel segments of %s and %s coincide at y=%d", a.flow, b.flow, a.a.Y)
				}
				continue
			}
			if a.a.X == b.a.X && min(a.a.Y, a.b.Y) < max(b.a.Y, b.b.Y) && min(b.a.Y, b.b.Y) < max(a.a.Y, a.b.Y) {
				t.Errorf("vertical segments of %s and %s coincide at x=%d", a.flow, b.flow, a.a.X)
			}
		}
	}
	// Flows leaving one gateway never run along each other, as long as the
	// gateway has an exit for each (at most maxExits outgoing flows).
	from, outDeg := map[string]string{}, map[string]int{}
	for _, f := range m.Flows {
		from[f.ID] = f.From
		outDeg[f.From]++
	}
	for i, a := range segs {
		for _, b := range segs[i+1:] {
			if a.flow == b.flow || from[a.flow] != from[b.flow] || a.horizontal != b.horizontal || outDeg[from[a.flow]] > maxExits {
				continue
			}
			if a.horizontal && a.a.Y == b.a.Y && min(a.a.X, a.b.X) < max(b.a.X, b.b.X) && min(b.a.X, b.b.X) < max(a.a.X, a.b.X) {
				t.Errorf("flows %s and %s leave %s along the same line y=%d", a.flow, b.flow, from[a.flow], a.a.Y)
			}
		}
	}
	checkLabels(t, m, g, func(a, b Point) []string {
		var sharers []string
		for _, o := range segs {
			if o.horizontal != (a.Y == b.Y) {
				continue
			}
			if o.horizontal && o.a.Y == a.Y && min(o.a.X, o.b.X) < max(a.X, b.X) && min(a.X, b.X) < max(o.a.X, o.b.X) ||
				!o.horizontal && o.a.X == a.X && min(o.a.Y, o.b.Y) < max(a.Y, b.Y) && min(a.Y, b.Y) < max(o.a.Y, o.b.Y) {
				sharers = append(sharers, o.flow)
			}
		}
		return sharers
	}, func(r rect) string {
		for _, o := range segs {
			if min(o.a.X, o.b.X) < r.x+r.w && max(o.a.X, o.b.X) > r.x && min(o.a.Y, o.b.Y) < r.y+r.h && max(o.a.Y, o.b.Y) > r.y {
				return o.flow
			}
		}
		return ""
	})
}

func intersects(a, b rect) bool {
	return a.x < b.x+b.w && b.x < a.x+a.w && a.y < b.y+b.h && b.y < a.y+a.h
}

// checkLabels verifies the flow label promises. overlapping returns the flows
// with a segment collinear with and overlapping a-b; crossing returns a flow
// with a segment inside r.
func checkLabels(t *testing.T, m Model, g geo, overlapping func(a, b Point) []string, crossing func(r rect) string) {
	t.Helper()
	type placed struct {
		flow string
		box  rect
	}
	var all []placed
	for _, f := range m.Flows {
		labels := g.labels[f.ID]
		if f.Label == "" {
			if len(labels) != 0 {
				t.Errorf("unlabelled flow %s has %d labels", f.ID, len(labels))
			}
			continue
		}
		if len(labels) != 1 {
			t.Errorf("flow %s has %d labels, want 1", f.ID, len(labels))
			continue
		}
		lb := labels[0]
		if want := truncate(f.Label, flowBudget); lb.text != want {
			t.Errorf("flow %s label %q, want %q", f.ID, lb.text, want)
		}
		if want := labelCharW*len([]rune(lb.text)) + 4; lb.box.w != want || lb.box.h != labelH {
			t.Errorf("flow %s label box %v, want %dx%d", f.ID, lb.box, want, labelH)
		}
		pts := g.flows[f.ID].pts
		if lb.anchor < 0 || lb.anchor+1 >= len(pts) {
			t.Errorf("flow %s label anchor %d is not a segment of its %d points", f.ID, lb.anchor, len(pts))
			continue
		}
		a, b := pts[lb.anchor], pts[lb.anchor+1]
		for _, other := range overlapping(a, b) {
			if other != f.ID {
				t.Errorf("flow %s label is anchored to segment %v-%v, which flow %s shares", f.ID, a, b, other)
			}
		}
		// Beside the anchor: just above a horizontal segment, just right of
		// a vertical one, and within its length.
		if a.Y == b.Y {
			gap := a.Y - (lb.box.y + lb.box.h)
			if gap < 0 || gap > 4 || lb.box.x < min(a.X, b.X) || lb.box.x+lb.box.w > max(a.X, b.X) {
				t.Errorf("flow %s label %v is not just above its segment %v-%v", f.ID, lb.box, a, b)
			}
		} else {
			gap := lb.box.x - a.X
			if gap < 0 || gap > 6 || lb.box.y < min(a.Y, b.Y) || lb.box.y+lb.box.h > max(a.Y, b.Y) {
				t.Errorf("flow %s label %v is not just right of its segment %v-%v", f.ID, lb.box, a, b)
			}
		}
		if lb.box.x < 0 || lb.box.y < 0 || lb.box.x+lb.box.w > g.width || lb.box.y+lb.box.h > g.height {
			t.Errorf("flow %s label %v lies outside the canvas", f.ID, lb.box)
		}
		for _, n := range m.Nodes {
			if intersects(lb.box, g.nodes[n.ID]) {
				t.Errorf("flow %s label %v overlaps node %s %v", f.ID, lb.box, n.ID, g.nodes[n.ID])
			}
		}
		if other := crossing(lb.box); other != "" {
			t.Errorf("flow %s label %v lies on a segment of flow %s", f.ID, lb.box, other)
		}
		for _, p := range all {
			if intersects(lb.box, p.box) {
				t.Errorf("labels of flows %s %v and %s %v overlap", f.ID, lb.box, p.flow, p.box)
			}
		}
		all = append(all, placed{f.ID, lb.box})
	}
}
