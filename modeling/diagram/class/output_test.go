package class

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"

	"github.com/spik3r/clanker-foundry/modeling/diagram/internal/svgcheck"
)

type drawnClass struct {
	id, kind string
	box      svgcheck.Rect
	attrs    []string
	ops      []string
	texts    []string
}

type drawnLabel struct {
	role string
	box  svgcheck.Rect
	text string
}

type drawnRel struct {
	id, kind, from, to string
	pts                [][2]int
	dashed             bool
	ends               map[string]string // "from"/"to" -> shape
	endClass           map[string]string // "from"/"to" -> css class of the shape
	labels             []drawnLabel
}

type drawn struct {
	width, height int
	order         []string
	classes       map[string]drawnClass
	rels          []drawnRel
}

func readBack(t testing.TB, svg []byte) drawn {
	t.Helper()
	els, err := svgcheck.Parse(svg)
	if err != nil {
		t.Fatalf("svg is not well-formed XML: %v", err)
	}
	d := drawn{classes: map[string]drawnClass{}}
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
		case e.Name == "g" && e.Attr["class"] == "class-group":
			box, err := svgcheck.Box(e.Attr["data-box"])
			if err != nil {
				t.Fatal(err)
			}
			c := drawnClass{id: e.Attr["data-id"], kind: e.Attr["data-kind"], box: box}
			for _, ch := range els[i+1:] {
				if ch.Depth <= e.Depth {
					break
				}
				if ch.Name == "text" {
					c.texts = append(c.texts, ch.Text)
					switch ch.Attr["data-kind"] {
					case "attribute":
						c.attrs = append(c.attrs, ch.Text)
					case "operation":
						c.ops = append(c.ops, ch.Text)
					}
				}
			}
			d.order = append(d.order, c.id)
			d.classes[c.id] = c
		case e.Name == "g" && e.Attr["class"] == "relation-group":
			r := drawnRel{id: e.Attr["data-id"], kind: e.Attr["data-kind"], from: e.Attr["data-from"], to: e.Attr["data-to"],
				ends: map[string]string{}, endClass: map[string]string{}}
			for _, f := range strings.Fields(e.Attr["data-points"]) {
				var x, y int
				if _, err := fmt.Sscanf(f, "%d,%d", &x, &y); err != nil {
					t.Fatalf("bad point %q", f)
				}
				r.pts = append(r.pts, [2]int{x, y})
			}
			for _, ch := range els[i+1:] {
				if ch.Depth <= e.Depth {
					break
				}
				switch ch.Name {
				case "path":
					if end := ch.Attr["data-end"]; end != "" {
						r.ends[end] = ch.Attr["data-shape"]
						r.endClass[end] = ch.Attr["class"]
					} else {
						pp, err := svgcheck.Points(ch.Attr["d"])
						if err != nil {
							t.Fatal(err)
						}
						if len(pp) != len(r.pts) {
							t.Errorf("relation %s: path and data-points differ", r.id)
						}
						r.dashed = strings.Contains(ch.Attr["class"], "dashed")
					}
				case "text":
					box, err := svgcheck.Box(ch.Attr["data-box"])
					if err != nil {
						t.Fatal(err)
					}
					r.labels = append(r.labels, drawnLabel{ch.Attr["data-kind"], box, ch.Text})
				}
			}
			d.rels = append(d.rels, r)
		}
	}
	return d
}

// onBorder reports whether p lies on the top or bottom border of b, away from
// the corners, and which one.
func onBorder(p [2]int, b svgcheck.Rect) (top, bottom bool) {
	if p[0] <= b.X || p[0] >= b.X+b.W {
		return false, false
	}
	return p[1] == b.Y, p[1] == b.Y+b.H
}

// segCrossesInterior reports whether the axis-aligned segment a-b passes
// through the open interior of r.
func segCrossesInterior(a, b [2]int, r svgcheck.Rect) bool {
	x1, x2 := min(a[0], b[0]), max(a[0], b[0])
	y1, y2 := min(a[1], b[1]), max(a[1], b[1])
	return x2 > r.X && x1 < r.X+r.W && y2 > r.Y && y1 < r.Y+r.H
}

// collinearOverlap reports whether two axis-aligned segments lie on one line
// and share more than a point.
func collinearOverlap(a1, a2, b1, b2 [2]int) bool {
	switch {
	case a1[0] == a2[0] && b1[0] == b2[0] && a1[0] == b1[0]:
		return min(max(a1[1], a2[1]), max(b1[1], b2[1])) > max(min(a1[1], a2[1]), min(b1[1], b2[1]))
	case a1[1] == a2[1] && b1[1] == b2[1] && a1[1] == b1[1]:
		return min(max(a1[0], a2[0]), max(b1[0], b2[0])) > max(min(a1[0], a2[0]), min(b1[0], b2[0]))
	}
	return false
}

func checkGeometry(t testing.TB, m Model, svg []byte) {
	t.Helper()
	d := readBack(t, svg)
	canvas := svgcheck.Rect{X: 0, Y: 0, W: d.width, H: d.height}
	if len(d.order) != len(m.Classes) || len(d.rels) != len(m.Relations) {
		t.Fatalf("%d classes and %d relations drawn, want %d and %d", len(d.order), len(d.rels), len(m.Classes), len(m.Relations))
	}
	// Classes in id order, inside the canvas, no two overlapping.
	for i, c := range m.Classes {
		dc := d.classes[c.ID]
		if d.order[i] != c.ID || dc.kind != c.Kind {
			t.Fatalf("class %d is %s/%s, want %s/%s", i, d.order[i], dc.kind, c.ID, c.Kind)
		}
		if !canvas.Contains(dc.box) {
			t.Errorf("class %s outside the canvas", c.ID)
		}
		if len(dc.attrs) != len(c.Attributes) || len(dc.ops) != len(c.Operations) {
			t.Errorf("class %s draws %d attributes and %d operations, want %d and %d", c.ID, len(dc.attrs), len(dc.ops), len(c.Attributes), len(c.Operations))
		}
		for _, line := range append(append([]string{}, dc.attrs...), dc.ops...) {
			if runes(line)*charW+2*boxPadX > dc.box.W {
				t.Errorf("class %s line %q (%d characters) does not fit width %d", c.ID, line, runes(line), dc.box.W)
			}
		}
		for _, o := range m.Classes[:i] {
			if dc.box.Overlap(d.classes[o.ID].box) {
				t.Errorf("classes %s and %s overlap", c.ID, o.ID)
			}
		}
		switch c.Kind {
		case KindInterface:
			if !contains(dc.texts, "«interface»") {
				t.Errorf("interface %s lacks its stereotype", c.ID)
			}
		case KindEnum:
			if !contains(dc.texts, "«enumeration»") {
				t.Errorf("enum %s lacks its stereotype", c.ID)
			}
		}
	}
	var boxes []svgcheck.Rect
	for _, id := range d.order {
		boxes = append(boxes, d.classes[id].box)
	}

	type seg struct {
		rel    string
		a, b   [2]int
		serial int
	}
	var segs []seg
	for i, r := range m.Relations {
		dr := d.rels[i]
		if dr.id != r.ID || dr.kind != r.Kind || dr.from != r.From || dr.to != r.To {
			t.Fatalf("relation %d is %+v, want %+v", i, dr, r)
		}
		from, to := d.classes[r.From].box, d.classes[r.To].box
		first, last := dr.pts[0], dr.pts[len(dr.pts)-1]
		if top, bottom := onBorder(first, from); !top && !bottom {
			t.Errorf("relation %s starts at %v, not on the border of %s %+v", r.ID, first, r.From, from)
		}
		if top, bottom := onBorder(last, to); !top && !bottom {
			t.Errorf("relation %s ends at %v, not on the border of %s %+v", r.ID, last, r.To, to)
		}
		// The line leaves each end away from its box.
		if len(dr.pts) < 2 {
			t.Fatalf("relation %s has %d points", r.ID, len(dr.pts))
		}
		if top, _ := onBorder(first, from); (top && dr.pts[1][1] >= first[1]) || (!top && dr.pts[1][1] <= first[1]) {
			t.Errorf("relation %s leaves %s into the box", r.ID, r.From)
		}
		if top, _ := onBorder(last, to); (top && dr.pts[len(dr.pts)-2][1] >= last[1]) || (!top && dr.pts[len(dr.pts)-2][1] <= last[1]) {
			t.Errorf("relation %s enters %s from inside the box", r.ID, r.To)
		}
		for j := 0; j+1 < len(dr.pts); j++ {
			a, b := dr.pts[j], dr.pts[j+1]
			if a[0] != b[0] && a[1] != b[1] {
				t.Errorf("relation %s segment %v-%v is not axis-aligned", r.ID, a, b)
			}
			if a == b {
				t.Errorf("relation %s has a zero-length segment", r.ID)
			}
			for k, bx := range boxes {
				if segCrossesInterior(a, b, bx) {
					t.Errorf("relation %s segment %v-%v passes through class %s", r.ID, a, b, m.Classes[k].ID)
				}
			}
			for _, p := range [][2]int{a, b} {
				if p[0] < 0 || p[1] < 0 || p[0] > d.width || p[1] > d.height {
					t.Errorf("relation %s point %v outside the canvas", r.ID, p)
				}
			}
			segs = append(segs, seg{r.ID, a, b, j})
		}
		// Line style and end shapes follow the kind.
		wantDashed := r.Kind == Realization || r.Kind == Dependency
		if dr.dashed != wantDashed {
			t.Errorf("relation %s dashed=%v", r.ID, dr.dashed)
		}
		atFrom, atTo := ends(r.Kind)
		wantShape := func(e string) string {
			switch e {
			case "triangle":
				return "triangle"
			case "arrow":
				return "arrow"
			case "":
				return ""
			}
			return "diamond"
		}
		if dr.ends["from"] != wantShape(atFrom) || dr.ends["to"] != wantShape(atTo) {
			t.Errorf("relation %s (%s) end shapes %v", r.ID, r.Kind, dr.ends)
		}
		switch r.Kind {
		case Inheritance, Realization:
			if dr.endClass["to"] != "hollow" {
				t.Errorf("relation %s: triangle is not hollow", r.ID)
			}
			// The parent is drawn above the child.
			if !(to.Y+to.H <= from.Y) {
				t.Errorf("%s %s: parent %s is not above child %s", r.Kind, r.ID, r.To, r.From)
			}
		case Composition:
			if dr.endClass["from"] != "solid" {
				t.Errorf("relation %s: composition diamond is not filled", r.ID)
			}
		case Aggregation:
			if dr.endClass["from"] != "hollow" {
				t.Errorf("relation %s: aggregation diamond is not hollow", r.ID)
			}
		}
		// Labels exist as given and stay clear of every class box.
		want := map[string]string{"label": r.Label, "fromLabel": r.FromLabel, "toLabel": r.ToLabel}
		got := map[string]string{}
		for _, lb := range dr.labels {
			got[lb.role] = lb.text
			if !canvas.Contains(lb.box) {
				t.Errorf("relation %s %s outside the canvas: %+v", r.ID, lb.role, lb.box)
			}
			for k, bx := range boxes {
				if lb.box.Overlap(bx) {
					t.Errorf("relation %s %s %q overlaps class %s", r.ID, lb.role, lb.text, m.Classes[k].ID)
				}
			}
		}
		for role, w := range want {
			if got[role] != w {
				t.Errorf("relation %s %s drawn as %q, want %q", r.ID, role, got[role], w)
			}
		}
	}
	// No two relations share a stretch of line.
	for i := range segs {
		for j := i + 1; j < len(segs); j++ {
			if segs[i].rel != segs[j].rel && collinearOverlap(segs[i].a, segs[i].b, segs[j].a, segs[j].b) {
				t.Errorf("relations %s and %s run along the same line: %v-%v and %v-%v", segs[i].rel, segs[j].rel, segs[i].a, segs[i].b, segs[j].a, segs[j].b)
			}
		}
	}
}

func contains(ss []string, s string) bool {
	for _, x := range ss {
		if x == s {
			return true
		}
	}
	return false
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

// randomModel builds a valid model from seed. Inheritance and realization
// only point from a later class to an earlier one, so they cannot cycle.
func randomModel(seed int64) Model {
	r := rand.New(rand.NewSource(seed))
	words := []string{"order", "Money", "a rather long member description that goes past the width budget of the box", "id", "<b>&", "ü"}
	text := func(max int) string {
		s := words[r.Intn(len(words))]
		for len(s) > max {
			rr := []rune(s)
			s = strings.TrimSpace(string(rr[:len(rr)-1]))
		}
		return s
	}
	m := Model{FormatVersion: Format, ProfileVersion: Profile, ID: "Rand", Title: text(120)}
	n := 1 + r.Intn(24)
	vis := []string{Public, Private, Protected, Package}
	for i := 0; i < n; i++ {
		c := Class{ID: fmt.Sprintf("C%02d", i), Name: text(120), Kind: []string{KindClass, KindClass, KindInterface, KindEnum}[r.Intn(4)]}
		for j := r.Intn(5); j > 0; j-- {
			a := Attribute{Name: fmt.Sprintf("a%d", j), Visibility: vis[r.Intn(4)]}
			if r.Intn(2) == 0 {
				a.Type = text(120)
			}
			c.Attributes = append(c.Attributes, a)
		}
		for j := r.Intn(5); j > 0; j-- {
			o := Operation{Name: fmt.Sprintf("o%d", j), Visibility: vis[r.Intn(4)]}
			for k := r.Intn(4); k > 0; k-- {
				o.Parameters = append(o.Parameters, Parameter{fmt.Sprintf("p%d", k), text(30)})
			}
			if r.Intn(2) == 0 {
				o.Returns = text(120)
			}
			c.Operations = append(c.Operations, o)
		}
		m.Classes = append(m.Classes, c)
	}
	kinds := []string{Inheritance, Realization, Association, Directed, Aggregation, Composition, Dependency}
	if n >= 2 {
		type key struct{ a, b, k string }
		seen := map[key]bool{}
		nr := r.Intn(min(MaxRelations, 3*n) + 1)
		for i := 0; i < nr; i++ {
			a, b := r.Intn(n), r.Intn(n)
			if a == b {
				continue
			}
			k := kinds[r.Intn(len(kinds))]
			if (k == Inheritance || k == Realization) && a < b {
				a, b = b, a // child has the higher index
			}
			f, to := m.Classes[a].ID, m.Classes[b].ID
			u, v := f, to
			if u > v {
				u, v = v, u
			}
			if seen[key{u, v, k}] {
				continue
			}
			seen[key{u, v, k}] = true
			rel := Relation{ID: fmt.Sprintf("R%02d", i), Kind: k, From: f, To: to}
			if r.Intn(3) == 0 {
				rel.Label = text(24)
			}
			if r.Intn(3) == 0 {
				rel.FromLabel = text(24)
			}
			if r.Intn(3) == 0 {
				rel.ToLabel = text(24)
			}
			m.Relations = append(m.Relations, rel)
		}
	}
	return m
}

func TestRandomModelGeometry(t *testing.T) {
	for seed := int64(1); seed <= 200; seed++ {
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
	for seed := int64(1); seed <= 30; seed++ {
		if string(build(t, randomModel(seed))) != string(build(t, randomModel(seed))) {
			t.Errorf("seed %d: output differs between runs", seed)
		}
	}
}

func TestSourceOrderDoesNotChangeOutput(t *testing.T) {
	a := baseModel()
	b := baseModel()
	b.Classes[0], b.Classes[3] = b.Classes[3], b.Classes[0]
	b.Relations[0], b.Relations[2] = b.Relations[2], b.Relations[0]
	if string(build(t, a)) != string(build(t, b)) {
		t.Errorf("class or relation order in the source changed the picture")
	}
}

func TestOutputFiles(t *testing.T) {
	gen, err := Profiler{}.Generate(readExample(t, "orders-class"), testGen)
	if err != nil {
		t.Fatal(err)
	}
	if len(gen.Files) != 2 || gen.Files[FileSVG] == nil || gen.Files[FileModel] == nil {
		t.Fatalf("files %d", len(gen.Files))
	}
	mj := string(gen.Files[FileModel])
	if !strings.HasSuffix(mj, "}\n") || !strings.Contains(mj, "\n  \"id\"") {
		t.Errorf("model.json is not indented with a trailing newline")
	}
	svg := string(gen.Files[FileSVG])
	for _, want := range []string{`viewBox="0 0 `, "<title>Order model</title>", "generated by diagramgen 1.2.2", "«interface»", "«enumeration»",
		`data-kind="composition"`, `data-kind="realization"`, `data-kind="aggregation"`, `data-kind="inheritance"`} {
		if !strings.Contains(svg, want) {
			t.Errorf("svg lacks %s", want)
		}
	}
}

// TestRanking checks that parents sit above children over a deep chain and a
// diamond, independent of ids.
func TestRanking(t *testing.T) {
	mk := func(id string) Class { return Class{ID: id, Name: id, Kind: KindClass} }
	m := Model{FormatVersion: Format, ProfileVersion: Profile, ID: "R", Title: "t",
		Classes: []Class{mk("A"), mk("B"), mk("C"), mk("D"), mk("E")},
		Relations: []Relation{
			{ID: "r1", Kind: Inheritance, From: "A", To: "B"},
			{ID: "r2", Kind: Inheritance, From: "B", To: "C"},
			{ID: "r3", Kind: Realization, From: "A", To: "D"},
			{ID: "r4", Kind: Inheritance, From: "D", To: "C"},
			{ID: "r5", Kind: Association, From: "E", To: "A"},
		}}
	m = m.Normalize()
	l := LayoutOf(m)
	row := map[string]int{}
	for _, c := range l.Classes {
		row[c.C.ID] = c.Row
	}
	if !(row["C"] < row["B"] && row["B"] < row["A"] && row["C"] < row["D"] && row["D"] < row["A"]) {
		t.Errorf("rows %v", row)
	}
	checkGeometry(t, m, SVG(m, l, testGen))
}

// TestWideRankIsSplit checks that a rank with many classes becomes several
// rows and that the parent rows still come first.
func TestWideRankIsSplit(t *testing.T) {
	m := Model{FormatVersion: Format, ProfileVersion: Profile, ID: "W", Title: "wide"}
	m.Classes = append(m.Classes, Class{ID: "Root", Name: "Root", Kind: KindInterface})
	for i := 0; i < 30; i++ {
		id := fmt.Sprintf("K%02d", i)
		m.Classes = append(m.Classes, Class{ID: id, Name: "Implementation " + id, Kind: KindClass})
		m.Relations = append(m.Relations, Relation{ID: "R" + id, Kind: Realization, From: id, To: "Root"})
	}
	m = m.Normalize()
	l := LayoutOf(m)
	if l.Rows < 3 {
		t.Errorf("30 wide classes drawn in %d rows", l.Rows)
	}
	if l.Width > maxRowW+400 {
		t.Errorf("diagram is %d wide", l.Width)
	}
	checkGeometry(t, m, SVG(m, l, testGen))
}

func TestHostileText(t *testing.T) {
	hostile := []string{
		`<script>alert(1)</script>`,
		`"><svg onload=alert(1)>`,
		`</text><foreignObject><body onload=x></foreignObject>`,
		`a & b &amp; &#x3C;`,
		`' onmouseover='x`,
		`javascript:alert(1) http://evil.example/x`,
		`url(http://evil.example/a) @import "x"`,
		`]]> <![CDATA[ <!DOCTYPE x [<!ENTITY e "x">]>`,
		"emoji 😀 and é and 中文",
	}
	for _, h := range hostile {
		m := baseModel()
		m.Title = h
		h24 := h
		for len(h24) > 24 {
			r := []rune(h24)
			h24 = string(r[:len(r)-1])
		}
		for i := range m.Classes {
			m.Classes[i].Name = h
			for j := range m.Classes[i].Attributes {
				m.Classes[i].Attributes[j].Name += h
				m.Classes[i].Attributes[j].Type = h
			}
			for j := range m.Classes[i].Operations {
				m.Classes[i].Operations[j].Name += h
				m.Classes[i].Operations[j].Returns = h
				for k := range m.Classes[i].Operations[j].Parameters {
					m.Classes[i].Operations[j].Parameters[k].Name += h
					m.Classes[i].Operations[j].Parameters[k].Type = h
				}
			}
		}
		for i := range m.Relations {
			m.Relations[i].Label, m.Relations[i].FromLabel, m.Relations[i].ToLabel = h24, h24, h24
		}
		for i := range m.Classes {
			for j := range m.Classes[i].Attributes {
				if len(m.Classes[i].Attributes[j].Name) > 120 {
					m.Classes[i].Attributes[j].Name = fmt.Sprintf("a%d", j)
				}
			}
			for j := range m.Classes[i].Operations {
				if len(m.Classes[i].Operations[j].Name) > 120 {
					m.Classes[i].Operations[j].Name = fmt.Sprintf("o%d", j)
				}
				for k := range m.Classes[i].Operations[j].Parameters {
					if len(m.Classes[i].Operations[j].Parameters[k].Name) > 120 {
						m.Classes[i].Operations[j].Parameters[k].Name = fmt.Sprintf("p%d", k)
					}
				}
			}
		}
		m = m.Normalize()
		if err := m.Validate(); err != nil {
			t.Fatalf("%q: %v", h, err)
		}
		svg := build(t, m)
		if err := svgcheck.Inert(svg); err != nil {
			t.Errorf("%q: %v", h, err)
		}
		checkGeometry(t, m, svg)
		bad := baseModel()
		bad.Classes[0].ID = `A"><script>`
		if bad.Validate() == nil {
			t.Errorf("hostile id accepted")
		}
	}
	// The text survives escaping.
	m := baseModel()
	m.Classes[0].Name = `a<b>&"c'`
	m = m.Normalize()
	d := readBack(t, build(t, m))
	if !contains(d.classes["Order"].texts, `a<b>&"c'`) {
		t.Errorf("class name read back as %v", d.classes["Order"].texts)
	}
}

func TestMemberLines(t *testing.T) {
	m := baseModel().Normalize()
	d := readBack(t, build(t, m))
	c := d.classes["Order"]
	if len(c.attrs) != 1 || c.attrs[0] != "+ id: OrderId" {
		t.Errorf("attributes %q", c.attrs)
	}
	if len(c.ops) != 2 || c.ops[0] != "+ total(): Money" || c.ops[1] != "- add(a: A, b: B)" {
		t.Errorf("operations %q", c.ops)
	}
	e := d.classes["Status"]
	if len(e.attrs) != 2 || e.attrs[0] != "NEW" || e.attrs[1] != "DONE" {
		t.Errorf("enum literals %q", e.attrs)
	}
	for sym, vis := range map[string]string{"+": Public, "-": Private, "#": Protected, "~": Package} {
		if visSymbol(vis) != sym {
			t.Errorf("symbol of %s is %s", vis, visSymbol(vis))
		}
	}
	// Long lines are truncated with an ellipsis; the width is clamped.
	m.Classes[1].Attributes[0].Type = strings.Repeat("T", 100) // Classes[1] is Order once sorted
	c = readBack(t, build(t, m)).classes["Order"]
	if !strings.HasSuffix(c.attrs[0], "…") || c.box.W != maxBoxW {
		t.Errorf("long attribute %q in a box of width %d", c.attrs[0], c.box.W)
	}
}

// A tall peer blocks every old bounded end-label candidate beside the short
// parent. The label must remain visible rather than being painted under it.
func TestRelationLabelsClearTallNeighbors(t *testing.T) {
	for _, height := range []int{8, MaxMembers} {
		for _, label := range []string{"Parent reference", strings.Repeat("L", MaxEndLabel)} {
			for _, from := range []bool{false, true} {
				for _, mirror := range []bool{false, true} {
					name := fmt.Sprintf("height%d/length%d/from%v/mirror%v", height, len(label), from, mirror)
					t.Run(name, func(t *testing.T) {
						parent, peer := "Parent", "Tall"
						if mirror {
							parent, peer = "Tall", "Parent"
						}
						m := Model{FormatVersion: Format, ProfileVersion: Profile, ID: "Example", Title: "Class label overlap",
							Classes:   []Class{{ID: parent, Name: "Parent", Kind: KindClass}, {ID: peer, Name: "Tall", Kind: KindClass}, {ID: "Child", Name: "Child", Kind: KindClass}},
							Relations: []Relation{{ID: "R", Kind: Inheritance, From: "Child", To: parent, ToLabel: label}}}
						for i := 0; i < height; i++ {
							m.Classes[1].Attributes = append(m.Classes[1].Attributes, Attribute{Name: fmt.Sprintf("a%d", i), Visibility: Public})
						}
						if from {
							m.Relations[0].ToLabel = ""
							m.Relations = append(m.Relations, Relation{ID: "S", Kind: Association, From: parent, To: "Child", FromLabel: label})
						}
						m = m.Normalize()
						if err := m.Validate(); err != nil {
							t.Fatal(err)
						}
						checkGeometry(t, m, build(t, m))
					})
				}
			}
		}
	}
}

func TestRelationCenterLabelsClearTallNeighbors(t *testing.T) {
	for _, reverse := range []bool{false, true} {
		t.Run(fmt.Sprintf("reverse%v", reverse), func(t *testing.T) {
			m := Model{FormatVersion: Format, ProfileVersion: Profile, ID: "Center", Title: "Center label overlap",
				Classes:   []Class{{ID: "Parent", Name: "Parent", Kind: KindClass}, {ID: "Tall", Name: "Tall", Kind: KindClass}, {ID: "Child", Name: "Child", Kind: KindClass}, {ID: "OtherChild", Name: "Other child", Kind: KindClass}},
				Relations: []Relation{{ID: "R", Kind: Inheritance, From: "Child", To: "Parent"}, {ID: "S", Kind: Inheritance, From: "OtherChild", To: "Tall"}, {ID: "T", Kind: Association, From: "Child", To: "Parent", Label: strings.Repeat("L", MaxEndLabel)}}}
			for i := 0; i < MaxMembers; i++ {
				m.Classes[1].Attributes = append(m.Classes[1].Attributes, Attribute{Name: fmt.Sprintf("a%d", i), Visibility: Public})
			}
			if reverse {
				m.Relations[2].From, m.Relations[2].To = m.Relations[2].To, m.Relations[2].From
			}
			m = m.Normalize()
			if err := m.Validate(); err != nil {
				t.Fatal(err)
			}
			checkGeometry(t, m, build(t, m))
		})
	}
}
