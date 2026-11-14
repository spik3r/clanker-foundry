package process

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// svgDoc is a parsed SVG: the geometry plus every element, for safety checks.
type svgDoc struct {
	geo      geo
	elements []xml.StartElement
	titles   []string
	viewBox  [4]int
}

func attr(e xml.StartElement, name string) (string, bool) {
	for _, a := range e.Attr {
		if a.Name.Local == name {
			return a.Value, true
		}
	}
	return "", false
}

func attrInt(t *testing.T, e xml.StartElement, name string) int {
	t.Helper()
	v, ok := attr(e, name)
	if !ok {
		t.Fatalf("<%s> lacks %s", e.Name.Local, name)
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		t.Fatalf("<%s %s=%q>: %v", e.Name.Local, name, v, err)
	}
	return n
}

// parseSVG reads emitted SVG back with encoding/xml, independent of the
// generator's own layout structures.
func parseSVG(t *testing.T, data []byte) svgDoc {
	t.Helper()
	doc := svgDoc{}
	doc.geo = geo{lanes: map[string]rect{}, nodes: map[string]rect{}, kinds: map[string]string{}, flows: map[string]geoFlow{}, labels: map[string][]geoLabel{}}
	d := xml.NewDecoder(bytes.NewReader(data))
	var curNode string
	var inTitle, inFlowLabel bool
	var flowLabel string
	for {
		tok, err := d.Token()
		if err != nil {
			break
		}
		switch e := tok.(type) {
		case xml.StartElement:
			doc.elements = append(doc.elements, e.Copy())
			name := e.Name.Local
			switch {
			case name == "svg":
				vb, _ := attr(e, "viewBox")
				if _, err := fmt.Sscanf(vb, "%d %d %d %d", &doc.viewBox[0], &doc.viewBox[1], &doc.viewBox[2], &doc.viewBox[3]); err != nil {
					t.Fatalf("viewBox %q: %v", vb, err)
				}
				doc.geo.width, doc.geo.height = doc.viewBox[2], doc.viewBox[3]
			case name == "title":
				inTitle = true
			case name == "rect" && hasAttr(e, "data-lane"):
				id, _ := attr(e, "data-lane")
				if _, dup := doc.geo.lanes[id]; dup {
					t.Errorf("lane %s appears twice", id)
				}
				doc.geo.lanes[id] = rect{attrInt(t, e, "x"), attrInt(t, e, "y"), attrInt(t, e, "width"), attrInt(t, e, "height")}
				doc.geo.laneOrder = append(doc.geo.laneOrder, id)
			case name == "g" && hasAttr(e, "data-id"):
				curNode, _ = attr(e, "data-id")
				if _, dup := doc.geo.nodes[curNode]; dup {
					t.Errorf("node %s appears twice", curNode)
				}
				doc.geo.kinds[curNode], _ = attr(e, "data-kind")
				doc.geo.nodes[curNode] = rect{}
			case curNode != "" && (name == "rect" || name == "circle" || name == "polygon") && doc.geo.nodes[curNode] == (rect{}):
				switch name {
				case "rect":
					doc.geo.nodes[curNode] = rect{attrInt(t, e, "x"), attrInt(t, e, "y"), attrInt(t, e, "width"), attrInt(t, e, "height")}
				case "circle":
					cx, cy, r := attrInt(t, e, "cx"), attrInt(t, e, "cy"), attrInt(t, e, "r")
					doc.geo.nodes[curNode] = rect{cx - r, cy - r, 2 * r, 2 * r}
				case "polygon":
					pts, _ := attr(e, "points")
					x0, y0, x1, y1 := 1<<30, 1<<30, -1<<30, -1<<30
					for _, pair := range strings.Fields(pts) {
						var x, y int
						if _, err := fmt.Sscanf(pair, "%d,%d", &x, &y); err != nil {
							t.Fatalf("points %q: %v", pts, err)
						}
						x0, y0, x1, y1 = min(x0, x), min(y0, y), max(x1, x), max(y1, y)
					}
					doc.geo.nodes[curNode] = rect{x0, y0, x1 - x0, y1 - y0}
				}
			case name == "path" && hasAttr(e, "data-id"):
				id, _ := attr(e, "data-id")
				if _, dup := doc.geo.flows[id]; dup {
					t.Errorf("flow %s appears twice", id)
				}
				f := geoFlow{}
				f.from, _ = attr(e, "data-from")
				f.to, _ = attr(e, "data-to")
				dv, _ := attr(e, "d")
				fields := strings.Fields(strings.NewReplacer("M", "", "L", " ").Replace(dv))
				if len(fields)%2 != 0 {
					t.Fatalf("flow %s path %q has an odd coordinate count", id, dv)
				}
				for i := 0; i < len(fields); i += 2 {
					x, err1 := strconv.Atoi(fields[i])
					y, err2 := strconv.Atoi(fields[i+1])
					if err1 != nil || err2 != nil {
						t.Fatalf("flow %s path %q", id, dv)
					}
					f.pts = append(f.pts, Point{x, y})
				}
				doc.geo.flows[id] = f
			case name == "text" && hasAttr(e, "data-flow"):
				inFlowLabel = true
				flowLabel, _ = attr(e, "data-flow")
				var lb geoLabel
				bv, _ := attr(e, "data-box")
				if _, err := fmt.Sscanf(bv, "%d %d %d %d", &lb.box.x, &lb.box.y, &lb.box.w, &lb.box.h); err != nil {
					t.Fatalf("flow %s data-box %q: %v", flowLabel, bv, err)
				}
				lb.anchor = attrInt(t, e, "data-anchor")
				doc.geo.labels[flowLabel] = append(doc.geo.labels[flowLabel], lb)
			}
		case xml.EndElement:
			switch e.Name.Local {
			case "g":
				curNode = ""
			case "title":
				inTitle = false
			case "text":
				inFlowLabel = false
			}
		case xml.CharData:
			if inTitle {
				doc.titles = append(doc.titles, string(e))
			}
			if inFlowLabel {
				list := doc.geo.labels[flowLabel]
				list[len(list)-1].text += string(e)
			}
		}
	}
	return doc
}

func hasAttr(e xml.StartElement, name string) bool {
	_, ok := attr(e, name)
	return ok
}

func generateAll(t *testing.T, m Model) (svg, bpmn []byte, l Layout) {
	t.Helper()
	l = LayoutOf(m)
	return SVG(m, l, testGen), BPMN(m, l, testGen), l
}

func TestExampleGeometryFromSVG(t *testing.T) {
	for _, name := range exampleNames {
		t.Run(name, func(t *testing.T) {
			m := mustDecode(t, readExample(t, name))
			svg, _, _ := generateAll(t, m)
			doc := parseSVG(t, svg)
			checkGeometry(t, m, doc.geo)
			if doc.viewBox[0] != 0 || doc.viewBox[1] != 0 {
				t.Errorf("viewBox %v does not start at the origin", doc.viewBox)
			}
			for i, id := range doc.geo.laneOrder {
				if id != m.Lanes[i].ID {
					t.Errorf("lane %d is %s, model order has %s", i, id, m.Lanes[i].ID)
				}
			}
			if len(doc.titles) == 0 || doc.titles[0] != m.Title {
				t.Errorf("svg title %v, want %q", doc.titles, m.Title)
			}
		})
	}
}

func TestShapesByKind(t *testing.T) {
	m := mustDecode(t, readExample(t, "parallel"))
	svg, _, _ := generateAll(t, m)
	s := string(svg)
	for _, want := range []string{
		`data-kind="start"><circle class="shape start"`,
		`data-kind="end"><circle class="shape end"`,
		`data-kind="task"><rect class="shape"`,
		`data-kind="parallel"><polygon`,
	} {
		if !strings.Contains(s, want) {
			t.Errorf("svg lacks %s", want)
		}
	}
	x := string(mustSVG(t, "order"))
	if !regexp.MustCompile(`data-kind="exclusive"><polygon[^>]*/><path class="mark" d="M(\d+) (\d+) L(\d+) (\d+) M(\d+) (\d+) L(\d+) (\d+)"/>`).MatchString(x) {
		t.Errorf("exclusive gateway is not a diamond with a cross")
	}
}

func mustSVG(t *testing.T, name string) []byte {
	m := mustDecode(t, readExample(t, name))
	svg, _, _ := generateAll(t, m)
	return svg
}

func TestExamplesCoverRoutes(t *testing.T) {
	rework := LayoutOf(mustDecode(t, readExample(t, "rework")))
	var f7 FlowPath
	for _, f := range rework.Flows {
		if f.Flow.ID == "F7" {
			f7 = f
		}
	}
	top := rework.Lanes[0].Box.Y
	viaChannel := false
	for _, p := range f7.Points {
		if p.Y < top || p.Y > rework.Lanes[len(rework.Lanes)-1].Box.Y+rework.Lanes[len(rework.Lanes)-1].Box.H {
			viaChannel = true
		}
	}
	if !viaChannel {
		t.Errorf("rework back edge F7 does not use a channel: %v", f7.Points)
	}
	if f7.Points[len(f7.Points)-1].X >= f7.Points[0].X {
		t.Errorf("back edge F7 does not run right to left: %v", f7.Points)
	}
}

func TestSVGSafety(t *testing.T) {
	m := baseModel()
	m.Title = `<b>&"x"`
	m.Lanes[0].Label = `<b>&"x"`
	m.Nodes[1].Label = `<b>&"x"` + ` 'q' <script>alert(1)</script>`
	m.Nodes[2].Label = `</text><b>`
	m.Flows[2].Label = `<b>&"x"`
	if err := m.Validate(); err != nil {
		t.Fatal(err)
	}
	m = m.Normalize()
	svg, _, _ := generateAll(t, m)
	doc := parseSVG(t, svg)
	if len(doc.titles) == 0 || doc.titles[0] != m.Title {
		t.Errorf("title round trip: %v", doc.titles)
	}
	if got := doc.geo.labels["F3"]; len(got) != 1 || got[0].text != truncate(`<b>&"x"`, flowBudget) {
		t.Errorf("flow label round trip: %v", got)
	}
	s := string(svg)
	for _, bad := range []string{"<b>", "<script", "</text><b>"} {
		if strings.Contains(s, bad) {
			t.Errorf("svg contains unescaped %q", bad)
		}
	}
	if !strings.Contains(s, `&lt;b&gt;&amp;&quot;x&quot;`) {
		t.Errorf("escaped label not found")
	}
	checkSVGInert(t, svg)
	for _, name := range exampleNames {
		checkSVGInert(t, mustSVG(t, name))
	}
}

// checkSVGInert applies the safety rules to emitted SVG.
func checkSVGInert(t *testing.T, svg []byte) {
	t.Helper()
	doc := parseSVG(t, svg)
	for _, e := range doc.elements {
		switch strings.ToLower(e.Name.Local) {
		case "script", "foreignobject", "image", "use", "iframe", "object", "embed", "a", "animate", "set", "font-face", "link":
			t.Errorf("svg contains <%s>", e.Name.Local)
		}
		for _, a := range e.Attr {
			n := strings.ToLower(a.Name.Local)
			if strings.HasPrefix(n, "on") {
				t.Errorf("<%s> has event attribute %s", e.Name.Local, a.Name.Local)
			}
			if n == "href" || (a.Name.Space != "" && a.Name.Local == "href") {
				if !strings.HasPrefix(a.Value, "#") {
					t.Errorf("<%s> has external href %q", e.Name.Local, a.Value)
				}
			}
			for _, v := range strings.Split(a.Value, "url(")[1:] {
				if !strings.HasPrefix(v, "#") {
					t.Errorf("<%s %s> references %q", e.Name.Local, a.Name.Local, a.Value)
				}
			}
			if strings.Contains(strings.ToLower(a.Value), "javascript:") {
				t.Errorf("<%s %s> contains javascript:", e.Name.Local, a.Name.Local)
			}
		}
	}
	lower := strings.ToLower(string(svg))
	for _, bad := range []string{"@import", "http://", "https://" + "x", "@font-face", "<!entity", "<!doctype"} {
		if bad == "http://" {
			// Only the SVG namespace may mention a URL.
			if strings.Count(lower, "http://") != strings.Count(lower, `xmlns="http://www.w3.org/2000/svg"`) {
				t.Errorf("svg mentions an http:// URL besides its namespace")
			}
			continue
		}
		if strings.Contains(lower, bad) {
			t.Errorf("svg contains %q", bad)
		}
	}
	if strings.Contains(string(svg), "url(") && strings.Count(string(svg), "url(#arrow)") != strings.Count(string(svg), "url(") {
		t.Errorf("svg has a url() other than #arrow")
	}
}

func TestWrap(t *testing.T) {
	cases := []struct {
		in     string
		budget int
		lines  int
		want   []string
	}{
		{"Check stock", 20, 3, []string{"Check stock"}},
		{"one two three four five six", 10, 3, []string{"one two", "three four", "five six"}},
		{"aaaaaaaaaaaaaaaaaaaaaaaaa", 10, 3, []string{"aaaaaaaaaa", "aaaaaaaaaa", "aaaaa"}},
		{"a b c d e f g h", 3, 2, []string{"a b", "c…"}},
		{"alpha beta gamma delta epsilon", 10, 2, []string{"alpha beta", "gamma…"}},
		{"héllo wörld", 20, 3, []string{"héllo wörld"}},
		{strings.Repeat("x", 200), 20, 3, []string{strings.Repeat("x", 20), strings.Repeat("x", 20), strings.Repeat("x", 19) + "…"}},
	}
	for _, c := range cases {
		got := wrap(c.in, c.budget, c.lines)
		if strings.Join(got, "|") != strings.Join(c.want, "|") {
			t.Errorf("wrap(%q, %d, %d) = %q, want %q", c.in, c.budget, c.lines, got, c.want)
		}
		if len(got) > c.lines {
			t.Errorf("wrap(%q) gave %d lines", c.in, len(got))
		}
		for _, ln := range got {
			if n := len([]rune(ln)); n > c.budget {
				t.Errorf("wrap(%q) line %q has %d characters", c.in, ln, n)
			}
		}
	}
	if truncate("abcdef", 4) != "abc…" || truncate("abcd", 4) != "abcd" {
		t.Errorf("truncate")
	}
}

func TestLongTaskLabelIsTruncated(t *testing.T) {
	m := mustDecode(t, readExample(t, "rework"))
	doc := string(mustSVG(t, "rework"))
	if !strings.Contains(doc, "…") {
		t.Errorf("long task label was not truncated")
	}
	if got := wrap(m.Nodes[nodeIndex(m, "Review")].Label, taskBudget, taskLines); len(got) != 3 || !strings.HasSuffix(got[2], "…") {
		t.Errorf("review label wraps to %q", got)
	}
}

func nodeIndex(m Model, id string) int {
	for i, n := range m.Nodes {
		if n.ID == id {
			return i
		}
	}
	return -1
}

func TestDeterministic(t *testing.T) {
	for _, name := range exampleNames {
		m := mustDecode(t, readExample(t, name))
		a1, b1, _ := generateAll(t, m)
		a2, b2, _ := generateAll(t, mustDecode(t, readExample(t, name)))
		if !bytes.Equal(a1, a2) || !bytes.Equal(b1, b2) {
			t.Errorf("%s: output differs between runs", name)
		}
	}
}

// bpmnDoc collects the elements of an emitted BPMN file by local name.
type bpmnDoc struct {
	byName map[string][]xml.StartElement
	text   map[string][]string // element local name -> text children in order
}

func parseBPMN(t *testing.T, data []byte) bpmnDoc {
	t.Helper()
	d := xml.NewDecoder(bytes.NewReader(data))
	doc := bpmnDoc{byName: map[string][]xml.StartElement{}, text: map[string][]string{}}
	var stack []string
	for {
		tok, err := d.Token()
		if err != nil {
			if err.Error() != "EOF" {
				t.Fatalf("BPMN is not well-formed: %v", err)
			}
			break
		}
		switch e := tok.(type) {
		case xml.StartElement:
			doc.byName[e.Name.Local] = append(doc.byName[e.Name.Local], e.Copy())
			stack = append(stack, e.Name.Local)
		case xml.EndElement:
			stack = stack[:len(stack)-1]
		case xml.CharData:
			if s := strings.TrimSpace(string(e)); s != "" && len(stack) > 0 {
				doc.text[stack[len(stack)-1]] = append(doc.text[stack[len(stack)-1]], s)
			}
		}
	}
	if len(stack) != 0 {
		t.Fatalf("unbalanced elements: %v", stack)
	}
	return doc
}

func TestBPMN(t *testing.T) {
	element := map[string]string{KindStart: "startEvent", KindEnd: "endEvent", KindTask: "task", KindExclusive: "exclusiveGateway", KindParallel: "parallelGateway"}
	for _, name := range exampleNames {
		t.Run(name, func(t *testing.T) {
			m := mustDecode(t, readExample(t, name))
			_, bpmn, l := generateAll(t, m)
			doc := parseBPMN(t, bpmn)
			if got := len(doc.byName["definitions"]); got != 1 {
				t.Fatalf("%d definitions", got)
			}
			defs := doc.byName["definitions"][0]
			if v, _ := attr(defs, "targetNamespace"); v != BPMNNamespace {
				t.Errorf("targetNamespace %q", v)
			}
			if len(doc.byName["collaboration"]) != 1 || len(doc.byName["participant"]) != 1 ||
				len(doc.byName["process"]) != 1 || len(doc.byName["laneSet"]) != 1 {
				t.Errorf("expected one collaboration, participant, process and laneSet")
			}
			if len(doc.byName["lane"]) != len(m.Lanes) {
				t.Errorf("%d lanes, want %d", len(doc.byName["lane"]), len(m.Lanes))
			}
			ids := map[string]bool{}
			for _, list := range doc.byName {
				for _, e := range list {
					if id, ok := attr(e, "id"); ok {
						if ids[id] {
							t.Errorf("duplicate id %q", id)
						}
						ids[id] = true
					}
				}
			}
			nodeKinds := map[string]string{}
			for _, n := range m.Nodes {
				nodeKinds[n.ID] = n.Kind
				list := doc.byName[element[n.Kind]]
				found := 0
				for _, e := range list {
					if id, _ := attr(e, "id"); id == n.ID {
						found++
						if label, _ := attr(e, "name"); label != n.Label {
							t.Errorf("node %s name %q, want %q", n.ID, label, n.Label)
						}
					}
				}
				if found != 1 {
					t.Errorf("node %s appears %d times as <%s>", n.ID, found, element[n.Kind])
				}
			}
			total := 0
			for _, el := range []string{"startEvent", "endEvent", "task", "exclusiveGateway", "parallelGateway"} {
				total += len(doc.byName[el])
			}
			if total != len(m.Nodes) {
				t.Errorf("%d flow nodes, want %d", total, len(m.Nodes))
			}
			if len(doc.byName["sequenceFlow"]) != len(m.Flows) {
				t.Fatalf("%d sequenceFlows, want %d", len(doc.byName["sequenceFlow"]), len(m.Flows))
			}
			for i, e := range doc.byName["sequenceFlow"] {
				f := m.Flows[i]
				id, _ := attr(e, "id")
				src, _ := attr(e, "sourceRef")
				dst, _ := attr(e, "targetRef")
				if id != f.ID || src != f.From || dst != f.To {
					t.Errorf("sequenceFlow %d is %s %s->%s, want %s %s->%s", i, id, src, dst, f.ID, f.From, f.To)
				}
				if nodeKinds[src] == "" || nodeKinds[dst] == "" {
					t.Errorf("sequenceFlow %s refers to a missing node", id)
				}
			}
			inc, out := map[string]int{}, map[string]int{}
			for _, f := range m.Flows {
				out[f.From]++
				inc[f.To]++
			}
			if len(doc.byName["incoming"]) != len(m.Flows) || len(doc.byName["outgoing"]) != len(m.Flows) {
				t.Errorf("incoming/outgoing counts %d/%d, want %d", len(doc.byName["incoming"]), len(doc.byName["outgoing"]), len(m.Flows))
			}
			for _, s := range append(append([]string(nil), doc.text["incoming"]...), doc.text["outgoing"]...) {
				found := false
				for _, f := range m.Flows {
					found = found || f.ID == s
				}
				if !found {
					t.Errorf("incoming/outgoing names unknown flow %q", s)
				}
			}
			refs := 0
			for _, s := range doc.text["flowNodeRef"] {
				if nodeKinds[s] == "" {
					t.Errorf("flowNodeRef %q is not a node", s)
				}
				refs++
			}
			if refs != len(m.Nodes) {
				t.Errorf("%d flowNodeRefs, want %d", refs, len(m.Nodes))
			}

			// Diagram interchange.
			shapes := map[string]int{}
			for _, e := range doc.byName["BPMNShape"] {
				ref, _ := attr(e, "bpmnElement")
				shapes[ref]++
			}
			for _, n := range m.Nodes {
				if shapes[n.ID] != 1 {
					t.Errorf("node %s has %d shapes", n.ID, shapes[n.ID])
				}
			}
			for _, ln := range m.Lanes {
				if shapes[ln.ID] != 1 {
					t.Errorf("lane %s has %d shapes", ln.ID, shapes[ln.ID])
				}
			}
			if len(doc.byName["BPMNShape"]) != len(m.Nodes)+len(m.Lanes)+1 {
				t.Errorf("%d shapes, want nodes + lanes + the participant", len(doc.byName["BPMNShape"]))
			}
			edges := map[string]int{}
			for _, e := range doc.byName["BPMNEdge"] {
				ref, _ := attr(e, "bpmnElement")
				edges[ref]++
			}
			for _, f := range m.Flows {
				if edges[f.ID] != 1 {
					t.Errorf("flow %s has %d edges", f.ID, edges[f.ID])
				}
			}
			if len(doc.byName["BPMNEdge"]) != len(m.Flows) {
				t.Errorf("%d edges, want %d", len(doc.byName["BPMNEdge"]), len(m.Flows))
			}
			waypoints := 0
			for _, f := range l.Flows {
				waypoints += len(f.Points)
			}
			if len(doc.byName["waypoint"]) != waypoints {
				t.Errorf("%d waypoints, want %d", len(doc.byName["waypoint"]), waypoints)
			}
			bounds := doc.byName["Bounds"]
			if len(bounds) != len(m.Nodes)+len(m.Lanes)+1 {
				t.Errorf("%d bounds", len(bounds))
			}
			// Bounds repeat the SVG layout.
			svgGeo := parseSVG(t, SVG(m, l, testGen)).geo
			for i, e := range doc.byName["BPMNShape"] {
				ref, _ := attr(e, "bpmnElement")
				want, ok := svgGeo.nodes[ref]
				if !ok {
					want, ok = svgGeo.lanes[ref]
				}
				if !ok {
					continue
				}
				got := rect{attrInt(t, bounds[i], "x"), attrInt(t, bounds[i], "y"), attrInt(t, bounds[i], "width"), attrInt(t, bounds[i], "height")}
				if got != want {
					t.Errorf("shape %s bounds %v differ from svg %v", ref, got, want)
				}
			}
		})
	}
}

func TestBPMNEscapesAndIsDeterministic(t *testing.T) {
	m := baseModel()
	m.Title = `<b>&"x"`
	m.Nodes[1].Label = `a < b & "c" > 'd'`
	m = m.Normalize()
	_, bpmn, _ := generateAll(t, m)
	doc := parseBPMN(t, bpmn)
	var got string
	for _, e := range doc.byName["task"] {
		if id, _ := attr(e, "id"); id == "Check" {
			got, _ = attr(e, "name")
		}
	}
	if got != m.Nodes[nodeIndex(m, "Check")].Label {
		t.Errorf("task name round trip: %q", got)
	}
	s := string(bpmn)
	if strings.Contains(s, "<b>") || strings.Contains(strings.ToLower(s), "time") && strings.Contains(s, "20") && false {
		t.Errorf("unescaped markup in BPMN")
	}
	_, again, _ := generateAll(t, m)
	if !bytes.Equal(bpmn, again) {
		t.Errorf("BPMN differs between runs")
	}
}

// randomModel builds a pseudo-random but deterministic process. It returns
// false when the draw does not satisfy the profile.
func randomModel(r *rand.Rand, seed int) (Model, bool) {
	k := 3 + r.Intn(18)
	lanes := 1 + r.Intn(4)
	m := Model{FormatVersion: Format, ProfileVersion: Profile, ID: "Random", Title: fmt.Sprintf("Random %d", seed)}
	for i := 0; i < lanes; i++ {
		m.Lanes = append(m.Lanes, Lane{fmt.Sprintf("L%d", i), fmt.Sprintf("Lane %d", i)})
	}
	lane := func() string { return m.Lanes[r.Intn(lanes)].ID }
	m.Nodes = append(m.Nodes, Node{"S", KindStart, "Start", lane()})
	for i := 0; i < k; i++ {
		kind := []string{KindExclusive, KindParallel, KindTask}[r.Intn(3)]
		m.Nodes = append(m.Nodes, Node{fmt.Sprintf("G%02d", i), kind, fmt.Sprintf("Step %d with a rather long label", i), lane()})
	}
	m.Nodes = append(m.Nodes, Node{"E1", KindEnd, "End one", lane()}, Node{"E2", KindEnd, "End two", lane()})
	seen := map[[2]string]bool{}
	add := func(from, to string) {
		if from == to || seen[[2]string{from, to}] {
			return
		}
		seen[[2]string{from, to}] = true
		m.Flows = append(m.Flows, Flow{ID: fmt.Sprintf("F%03d", len(m.Flows)), From: from, To: to, Label: fmt.Sprintf("l%d", len(m.Flows))})
	}
	g := func(i int) string { return fmt.Sprintf("G%02d", i) }
	add("S", g(0))
	for i := 0; i < k; i++ {
		if i > 0 {
			add(g(r.Intn(i)), g(i))
		}
		for j := r.Intn(3); j >= 0; j-- {
			if i+1 < k && r.Intn(4) != 0 {
				add(g(i), g(i+1+r.Intn(k-i-1)))
			} else {
				add(g(i), []string{"E1", "E2"}[r.Intn(2)])
			}
		}
		if i > 1 && r.Intn(4) == 0 {
			add(g(i), g(r.Intn(i))) // rework loop
		}
	}
	add(g(k-1), "E1")
	// Tasks need exactly one outgoing flow: make every task a gateway if not.
	out := map[string]int{}
	for _, f := range m.Flows {
		out[f.From]++
	}
	for i := range m.Nodes {
		if m.Nodes[i].Kind == KindTask && out[m.Nodes[i].ID] != 1 {
			m.Nodes[i].Kind = KindExclusive
		}
	}
	if m.Validate() != nil {
		return Model{}, false
	}
	return m.Normalize(), true
}

func TestGeometryOnRandomModels(t *testing.T) {
	valid := 0
	backEdges := 0
	for seed := 0; seed < 600; seed++ {
		r := rand.New(rand.NewSource(int64(seed)))
		m, ok := randomModel(r, seed)
		if !ok {
			continue
		}
		valid++
		l := LayoutOf(m)
		g := geoOfLayout(l)
		checkGeometry(t, m, g)
		doc := parseSVG(t, SVG(m, l, testGen))
		checkGeometry(t, m, doc.geo)
		checkBPMNGeometry(t, m, BPMN(m, l, testGen))
		for _, f := range l.Flows {
			if f.Points[len(f.Points)-1].X < f.Points[0].X {
				backEdges++
			}
		}
		if t.Failed() {
			t.Fatalf("seed %d failed", seed)
		}
	}
	if valid < 100 {
		t.Fatalf("only %d of 600 random models were valid", valid)
	}
	if backEdges == 0 {
		t.Errorf("no random model exercised a back edge")
	}
	t.Logf("%d valid random models, %d right-to-left flows", valid, backEdges)
}

// exitSide reports where a flow leaves its gateway source.
func exitSide(l Layout, id string) string {
	for _, f := range l.Flows {
		if f.Flow.ID != id {
			continue
		}
		var s NodeBox
		for _, n := range l.Nodes {
			if n.Node.ID == f.Flow.From {
				s = n
			}
		}
		switch p := f.Points[0]; {
		case p == Point{s.CX, s.Box.Y}:
			return "top"
		case p == Point{s.CX, s.Box.Y + s.Box.H}:
			return "bottom"
		case p == Point{s.Box.X + s.Box.W, s.CY}:
			return "right"
		}
		return "offset"
	}
	return "missing"
}

func TestGatewayExitsByDirection(t *testing.T) {
	cases := []struct{ example, flow, want string }{
		{"order", "F3", "right"},      // Ship is in the gateway's row
		{"order", "F4", "top"},        // Reject is in the lane above
		{"rework", "F4", "bottom"},    // Publish is two rows below
		{"rework", "F5", "top"},       // Revise is in the lane above
		{"rework", "F6", "offset"},    // Escalate is below too: second bottom exit
		{"parallel", "F03", "bottom"}, // Pick is below
		{"parallel", "F04", "offset"}, // Invoice is below: second bottom exit
		{"parallel", "F05", "right"},  // Notify is in the same row
		{"parallel", "F10", "bottom"}, // Dispatch is below the join
	}
	for _, c := range cases {
		l := LayoutOf(mustDecode(t, readExample(t, c.example)))
		if got := exitSide(l, c.flow); got != c.want {
			t.Errorf("%s %s leaves at %s, want %s", c.example, c.flow, got, c.want)
		}
	}
	// The second bottom exit starts below the centre, on the diamond.
	l := LayoutOf(mustDecode(t, readExample(t, "rework")))
	for _, f := range l.Flows {
		if f.Flow.ID == "F6" {
			var s NodeBox
			for _, n := range l.Nodes {
				if n.Node.ID == "Approved" {
					s = n
				}
			}
			if f.Points[0].Y <= s.CY || f.Points[0].X == s.CX {
				t.Errorf("F6 starts at %v, want an offset point on the lower half of %v", f.Points[0], s.Box)
			}
		}
	}
}

// fanOut builds a gateway with n labelled outgoing flows, one per task, that
// all join again.
func fanOut(n int) Model {
	m := Model{FormatVersion: Format, ProfileVersion: Profile, ID: "Fan", Title: "Fan out",
		Lanes: []Lane{{"A", "A"}, {"B", "B"}, {"C", "C"}}}
	m.Nodes = []Node{{"S", KindStart, "Start", "B"}, {"G", KindExclusive, "Which?", "B"},
		{"J", KindExclusive, "Joined", "B"}, {"E", KindEnd, "End", "B"}}
	m.Flows = []Flow{{"A1", "S", "G", "go"}, {"Z1", "J", "E", "done"}}
	for i := 0; i < n; i++ {
		id := fmt.Sprintf("T%02d", i)
		m.Nodes = append(m.Nodes, Node{id, KindTask, "Task " + id, m.Lanes[i%3].ID})
		m.Flows = append(m.Flows,
			Flow{fmt.Sprintf("F%02d", i), "G", id, fmt.Sprintf("option %d", i)},
			Flow{fmt.Sprintf("M%02d", i), id, "J", fmt.Sprintf("merge %d", i)})
	}
	return m.Normalize()
}

func TestLabelsOnWideFanOut(t *testing.T) {
	for _, n := range []int{2, 3, 8, 9, 10, 14, 30} {
		t.Run(fmt.Sprint(n), func(t *testing.T) {
			m := fanOut(n)
			if err := m.Validate(); err != nil {
				t.Fatal(err)
			}
			l := LayoutOf(m)
			checkGeometry(t, m, geoOfLayout(l))
			checkGeometry(t, m, parseSVG(t, SVG(m, l, testGen)).geo)
			parseBPMN(t, BPMN(m, l, testGen))
		})
	}
}

// bpmnGeometry reads the diagram interchange of emitted BPMN: shape bounds by
// bpmnElement and edge waypoints by bpmnElement.
func bpmnGeometry(t *testing.T, data []byte) (shapes map[string]rect, edges map[string][]Point) {
	t.Helper()
	shapes, edges = map[string]rect{}, map[string][]Point{}
	d := xml.NewDecoder(bytes.NewReader(data))
	cur := ""
	for {
		tok, err := d.Token()
		if err != nil {
			break
		}
		e, ok := tok.(xml.StartElement)
		if !ok {
			continue
		}
		switch e.Name.Local {
		case "BPMNShape", "BPMNEdge":
			cur, _ = attr(e, "bpmnElement")
		case "Bounds":
			shapes[cur] = rect{attrInt(t, e, "x"), attrInt(t, e, "y"), attrInt(t, e, "width"), attrInt(t, e, "height")}
		case "waypoint":
			edges[cur] = append(edges[cur], Point{attrInt(t, e, "x"), attrInt(t, e, "y")})
		}
	}
	return shapes, edges
}

func inside(in, out rect) bool {
	return in.x >= out.x && in.y >= out.y && in.x+in.w <= out.x+out.w && in.y+in.h <= out.y+out.h
}

// checkBPMNGeometry verifies the pool, lane, node and edge promises of the
// emitted BPMN against model m.
func checkBPMNGeometry(t *testing.T, m Model, bpmn []byte) {
	t.Helper()
	shapes, edges := bpmnGeometry(t, bpmn)
	pool, ok := shapes["Participant."+m.ID]
	if !ok {
		t.Fatalf("no participant shape")
	}
	for _, ln := range m.Lanes {
		lr, ok := shapes[ln.ID]
		if !ok {
			t.Fatalf("lane %s has no shape", ln.ID)
		}
		if lr.x != pool.x+poolStripW {
			t.Errorf("lane %s x %d, want %d (pool x %d + %d)", ln.ID, lr.x, pool.x+poolStripW, pool.x, poolStripW)
		}
		if !inside(lr, pool) {
			t.Errorf("lane %s %v is not inside participant %v", ln.ID, lr, pool)
		}
	}
	for _, n := range m.Nodes {
		nr, ok := shapes[n.ID]
		if !ok {
			t.Fatalf("node %s has no shape", n.ID)
		}
		if !inside(nr, shapes[n.Lane]) {
			t.Errorf("node %s %v is not inside lane %s %v", n.ID, nr, n.Lane, shapes[n.Lane])
		}
	}
	type seg struct {
		flow, to   string
		a, b       Point
		horizontal bool
	}
	var segs []seg
	for _, f := range m.Flows {
		pts := edges[f.ID]
		if len(pts) < 2 {
			t.Fatalf("flow %s has %d waypoints", f.ID, len(pts))
		}
		for i, p := range pts {
			if p.X < pool.x || p.X > pool.x+pool.w || p.Y < pool.y || p.Y > pool.y+pool.h {
				t.Errorf("flow %s waypoint %v is outside participant %v", f.ID, p, pool)
			}
			if i > 0 {
				segs = append(segs, seg{f.ID, f.To, pts[i-1], p, p.Y == pts[i-1].Y})
			}
		}
	}
	out := map[string]int{}
	src := map[string]string{}
	for _, f := range m.Flows {
		out[f.From]++
		src[f.ID] = f.From
	}
	for i, a := range segs {
		for _, b := range segs[i+1:] {
			if a.to == b.to || a.horizontal != b.horizontal || src[a.flow] == src[b.flow] && out[src[a.flow]] > maxExits {
				continue
			}
			if a.horizontal && a.a.Y == b.a.Y && min(a.a.X, a.b.X) < max(b.a.X, b.b.X) && min(b.a.X, b.b.X) < max(a.a.X, a.b.X) ||
				!a.horizontal && a.a.X == b.a.X && min(a.a.Y, a.b.Y) < max(b.a.Y, b.b.Y) && min(b.a.Y, b.b.Y) < max(a.a.Y, a.b.Y) {
				t.Errorf("flows %s and %s into different targets share a BPMN segment %v-%v / %v-%v", a.flow, b.flow, a.a, a.b, b.a, b.b)
			}
		}
	}
}

func TestBPMNPoolLanesAndEdgesOnExamples(t *testing.T) {
	for _, name := range exampleNames {
		t.Run(name, func(t *testing.T) {
			m := mustDecode(t, readExample(t, name))
			_, bpmn, _ := generateAll(t, m)
			checkBPMNGeometry(t, m, bpmn)
		})
	}
}
