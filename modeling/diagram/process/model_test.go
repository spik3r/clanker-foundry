package process

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/spik3r/clanker-foundry/modeling/diagram/contract"
)

func TestExamplesDecode(t *testing.T) {
	for _, name := range exampleNames {
		mustDecode(t, readExample(t, name))
	}
}

func TestValidateRefusals(t *testing.T) {
	long := strings.Repeat("x", 121)
	id49 := "A" + strings.Repeat("b", 48)
	cases := []struct {
		name   string
		mutate func(m *Model)
		want   string
	}{
		{"wrong format", func(m *Model) { m.FormatVersion = "diagram-json/v2" }, "unsupported"},
		{"wrong profile", func(m *Model) { m.ProfileVersion = "class-diagram/v1" }, "unsupported"},
		{"model id starts with digit", func(m *Model) { m.ID = "1Order" }, "id must be"},
		{"model id too long", func(m *Model) { m.ID = id49 }, "id must be"},
		{"empty title", func(m *Model) { m.Title = "" }, "title"},
		{"blank title", func(m *Model) { m.Title = "   " }, "title"},
		{"title too long", func(m *Model) { m.Title = long }, "title"},
		{"title control character", func(m *Model) { m.Title = "a\tb" }, "title"},
		{"title line separator", func(m *Model) { m.Title = "a b" }, "title"},
		{"title non-XML character", func(m *Model) { m.Title = "a￿b" }, "title"},
		{"title invalid UTF-8", func(m *Model) { m.Title = "a\xffb" }, "title"},
		{"no lanes", func(m *Model) { m.Lanes = nil }, "lanes"},
		{"nine lanes", func(m *Model) {
			m.Lanes = nil
			for i := 0; i < 9; i++ {
				m.Lanes = append(m.Lanes, Lane{fmt.Sprintf("L%d", i), "Lane"})
			}
		}, "lanes"},
		{"lane id invalid", func(m *Model) { m.Lanes[0].ID = "Sales Team" }, "lane id"},
		{"lane label empty", func(m *Model) { m.Lanes[0].Label = "" }, "lane label"},
		{"duplicate lane id", func(m *Model) { m.Lanes[1].ID = "Sales" }, "used by"},
		{"node id equals lane id", func(m *Model) { m.Nodes[1].ID = "Sales"; m.Flows[0].To = "Sales"; m.Flows[1].From = "Sales" }, "used by"},
		{"flow id equals node id", func(m *Model) { m.Flows[0].ID = "Start" }, "used by"},
		{"duplicate node id", func(m *Model) { m.Nodes[2].ID = "Check" }, "used by"},
		{"duplicate flow id", func(m *Model) { m.Flows[1].ID = "F1" }, "used by"},
		{"node id invalid", func(m *Model) { m.Nodes[3].ID = "_Ship" }, "node id"},
		{"flow id invalid", func(m *Model) { m.Flows[0].ID = "F-1" }, "flow id"},
		{"one node", func(m *Model) { m.Nodes = m.Nodes[:1]; m.Flows = nil }, "nodes"},
		{"no flows", func(m *Model) { m.Flows = nil }, "flows"},
		{"too many nodes", func(m *Model) {
			for i := 0; i < 65; i++ {
				m.Nodes = append(m.Nodes, Node{fmt.Sprintf("N%d", i), KindTask, "t", "Sales"})
			}
		}, "nodes must"},
		{"too many flows", func(m *Model) {
			for i := 0; i < 129; i++ {
				m.Flows = append(m.Flows, Flow{ID: fmt.Sprintf("X%d", i), From: "Check", To: "Ship"})
			}
		}, "flows must"},
		{"node label empty", func(m *Model) { m.Nodes[1].Label = "" }, "node label"},
		{"node label too long", func(m *Model) { m.Nodes[1].Label = long }, "node label"},
		{"node label newline", func(m *Model) { m.Nodes[1].Label = "a\nb" }, "node label"},
		{"flow label too long", func(m *Model) { m.Flows[2].Label = long }, "flow label"},
		{"flow label control", func(m *Model) { m.Flows[2].Label = "a\x00" }, "flow label"},
		{"unknown kind", func(m *Model) { m.Nodes[1].Kind = "subprocess" }, "unknown kind"},
		{"kind case variant", func(m *Model) { m.Nodes[1].Kind = "Task" }, "unknown kind"},
		{"unknown lane", func(m *Model) { m.Nodes[1].Lane = "Nowhere" }, "unknown lane"},
		{"no start", func(m *Model) { m.Nodes[0].Kind = KindTask }, "start node"},
		{"two starts", func(m *Model) { m.Nodes[1].Kind = KindStart }, "start node"},
		{"no end", func(m *Model) { m.Nodes[5].Kind = KindTask }, "end node"},
		{"flow from unknown node", func(m *Model) { m.Flows[0].From = "Ghost" }, "unknown node"},
		{"flow to unknown node", func(m *Model) { m.Flows[0].To = "Ghost" }, "unknown node"},
		{"self loop", func(m *Model) { m.Flows = append(m.Flows, Flow{ID: "F7", From: "Ship", To: "Ship"}) }, "self-loop"},
		{"duplicate from/to pair", func(m *Model) { m.Flows = append(m.Flows, Flow{ID: "F7", From: "Check", To: "InStock"}) }, "repeats"},
		{"start with incoming", func(m *Model) { m.Flows = append(m.Flows, Flow{ID: "F7", From: "Reject", To: "Start"}) }, "start"},
		{"start with two outgoing", func(m *Model) { m.Flows = append(m.Flows, Flow{ID: "F7", From: "Start", To: "Ship"}) }, "start"},
		{"end with outgoing", func(m *Model) { m.Flows = append(m.Flows, Flow{ID: "F7", From: "Done", To: "Ship"}) }, "end"},
		{"end without incoming", func(m *Model) {
			m.Nodes = append(m.Nodes, Node{"Lost", KindEnd, "Lost", "Sales"})
		}, "end"},
		{"task with two outgoing", func(m *Model) { m.Flows = append(m.Flows, Flow{ID: "F7", From: "Check", To: "Ship"}) }, "task"},
		{"task without outgoing", func(m *Model) { m.Flows = m.Flows[:4] }, "task"},
		{"task without incoming", func(m *Model) {
			m.Nodes = append(m.Nodes, Node{"Idle", KindTask, "Idle", "Sales"})
			m.Flows = append(m.Flows, Flow{ID: "F7", From: "Idle", To: "Done"})
		}, "task"},
		{"gateway with two flows", func(m *Model) {
			m.Nodes[2].Kind = KindParallel
			m.Flows = append(m.Flows[:3], m.Flows[4:]...) // drop F4: InStock has 2 flows
		}, "gateway"},
		{"gateway without outgoing", func(m *Model) {
			m.Nodes = append(m.Nodes, Node{"Dead", KindParallel, "Dead", "Sales"})
			m.Flows = append(m.Flows, Flow{ID: "F7", From: "InStock", To: "Dead"})
		}, "gateway"},
		{"unreachable island", func(m *Model) {
			m.Nodes = append(m.Nodes, Node{"A", KindExclusive, "A", "Sales"}, Node{"B", KindExclusive, "B", "Sales"})
			m.Flows = append(m.Flows,
				Flow{ID: "F7", From: "A", To: "B"}, Flow{ID: "F8", From: "B", To: "A"},
				Flow{ID: "F9", From: "B", To: "Done"}, Flow{ID: "F10", From: "A", To: "Done"})
			// A: in 1 (B), out 2 -> valid degrees; B: in 1, out 2. Neither is reachable from Start.
		}, "not reachable"},
		{"cannot reach an end", func(m *Model) {
			// A and B form a cycle that nothing leaves.
			m.Nodes = append(m.Nodes, Node{"A", KindExclusive, "A", "Sales"}, Node{"B", KindExclusive, "B", "Sales"})
			m.Flows = append(m.Flows,
				Flow{ID: "F7", From: "InStock", To: "A"}, Flow{ID: "F8", From: "InStock", To: "B"},
				Flow{ID: "F9", From: "A", To: "B"}, Flow{ID: "F10", From: "B", To: "A"})
		}, "cannot reach an end"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			m := baseModel()
			if err := m.Validate(); err != nil {
				t.Fatalf("base model invalid: %v", err)
			}
			c.mutate(&m)
			err := m.Validate()
			if err == nil {
				t.Fatalf("accepted")
			}
			if !strings.Contains(err.Error(), c.want) {
				t.Errorf("error %q does not mention %q", err, c.want)
			}
		})
	}
}

func TestAcceptsBoundaryValues(t *testing.T) {
	m := baseModel()
	m.ID = "A" + strings.Repeat("b", 47)
	m.Title = strings.Repeat("é", 60) // 120 bytes
	m.Flows[2].Label = strings.Repeat("y", 120)
	if err := m.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestDecodeRefusals(t *testing.T) {
	good := string(marshal(t, baseModel()))
	with := func(old, repl string) string {
		if !strings.Contains(good, old) {
			t.Fatalf("fixture lacks %q", old)
		}
		return strings.Replace(good, old, repl, 1)
	}
	deep := strings.Repeat("[", 20) + strings.Repeat("]", 20)
	cases := []struct {
		name string
		raw  string
	}{
		{"empty", ""},
		{"whitespace only", " \n"},
		{"not JSON", "model"},
		{"array", "[]"},
		{"scalar", "7"},
		{"truncated", good[:len(good)/2]},
		{"trailing object", good + good},
		{"trailing garbage", good + " x"},
		{"duplicate member", with(`"id":"OrderHandling"`, `"id":"OrderHandling","id":"Other"`)},
		{"duplicate nested member", with(`"label":"Sales"`, `"label":"Sales","label":"Sales"`)},
		{"unknown member", with(`"id":"OrderHandling"`, `"id":"OrderHandling","colour":"red"`)},
		{"unknown nested member", with(`"label":"Sales"`, `"label":"Sales","x":10`)},
		{"case-variant member", with(`"title"`, `"Title"`)},
		{"case-variant nested member", with(`"kind":"start"`, `"Kind":"start"`)},
		{"null model member", with(`"title":"Order handling"`, `"title":null`)},
		{"null element", with(`"lanes":[`, `"lanes":[null,`)},
		{"null flow label", with(`"label":"yes"`, `"label":null`)},
		{"empty flow label", with(`"label":"yes"`, `"label":""`)},
		{"number for string", with(`"title":"Order handling"`, `"title":7`)},
		{"object for string", with(`"title":"Order handling"`, `"title":{}`)},
		{"array for string", with(`"title":"Order handling"`, `"title":[]`)},
		{"string for array", with(`"lanes":[`, `"lanes":"x","x":[`)},
		{"deep array", with(`"title":"Order handling"`, `"title":`+deep)},
		{"invalid UTF-8", with(`Order handling`, "Order \xff handling")},
		{"oversize", good + strings.Repeat(" ", contract.MaxModelBytes)},
		{"missing members", `{"formatVersion":"diagram-json/v1","profileVersion":"process-diagram/v1"}`},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if _, err := Decode([]byte(c.raw)); err == nil {
				t.Fatal("accepted")
			}
		})
	}
	if _, err := Decode([]byte(good)); err != nil {
		t.Fatalf("good fixture refused: %v", err)
	}
	// A model that is exactly at the size limit with trailing whitespace is fine.
	pad := good + strings.Repeat(" ", contract.MaxModelBytes-len(good))
	if _, err := Decode([]byte(pad)); err != nil {
		t.Errorf("model of exactly %d bytes refused: %v", contract.MaxModelBytes, err)
	}
}

func TestCanonicalInvariance(t *testing.T) {
	digest := func(raw []byte) string {
		m := mustDecode(t, raw)
		c, err := m.Canonical()
		if err != nil {
			t.Fatal(err)
		}
		return contract.HexSum(c)
	}
	base := digest(readExample(t, "order"))

	// Whitespace and member order.
	m := baseModel()
	spaced := strings.NewReplacer(`{`, "{\n  ", `,`, ",\n  ", `}`, "\n}").Replace(string(marshal(t, m)))
	if got := digest([]byte(spaced)); got != base {
		t.Errorf("whitespace changed the digest")
	}
	reordered := `{"flows":[{"to":"Check","from":"Start","id":"F1"},{"id":"F2","from":"Check","to":"InStock"},{"id":"F3","from":"InStock","to":"Ship","label":"yes"},{"id":"F4","from":"InStock","to":"Reject","label":"no"},{"id":"F5","from":"Ship","to":"Done"},{"id":"F6","from":"Reject","to":"Done"}],` +
		`"nodes":[{"lane":"Sales","label":"Closed","kind":"end","id":"Done"},{"id":"Reject","kind":"task","label":"Notify customer","lane":"Sales"},{"id":"Ship","kind":"task","label":"Ship order","lane":"Warehouse"},{"id":"InStock","kind":"exclusive","label":"In stock?","lane":"Warehouse"},{"id":"Check","kind":"task","label":"Check stock","lane":"Warehouse"},{"id":"Start","kind":"start","label":"Order received","lane":"Sales"}],` +
		`"lanes":[{"label":"Sales","id":"Sales"},{"label":"Warehouse","id":"Warehouse"}],"title":"Order handling","id":"OrderHandling","profileVersion":"process-diagram/v1","formatVersion":"diagram-json/v1"}`
	if got := digest([]byte(reordered)); got != base {
		t.Errorf("member, node or flow order changed the digest")
	}
	rev := baseModel()
	for i, j := 0, len(rev.Flows)-1; i < j; i, j = i+1, j-1 {
		rev.Flows[i], rev.Flows[j] = rev.Flows[j], rev.Flows[i]
	}
	for i, j := 0, len(rev.Nodes)-1; i < j; i, j = i+1, j-1 {
		rev.Nodes[i], rev.Nodes[j] = rev.Nodes[j], rev.Nodes[i]
	}
	if got := digest(marshal(t, rev)); got != base {
		t.Errorf("reversed node and flow order changed the digest")
	}

	// Sensitivity.
	swapped := baseModel()
	swapped.Lanes[0], swapped.Lanes[1] = swapped.Lanes[1], swapped.Lanes[0]
	if got := digest(marshal(t, swapped)); got == base {
		t.Errorf("lane order did not change the digest")
	}
	relabelled := baseModel()
	relabelled.Nodes[1].Label = "Check the stock"
	if got := digest(marshal(t, relabelled)); got == base {
		t.Errorf("a label change did not change the digest")
	}
	titled := baseModel()
	titled.Title = "Order handling v2"
	if got := digest(marshal(t, titled)); got == base {
		t.Errorf("a title change did not change the digest")
	}
}

func TestCanonicalEncoding(t *testing.T) {
	m := baseModel()
	m.Title = `a<b>&"c"`
	c, err := m.Canonical()
	if err != nil {
		t.Fatal(err)
	}
	if bytes.HasSuffix(c, []byte("\n")) || bytes.Contains(c, []byte(`": `)) || bytes.Contains(c, []byte(`", `)) {
		t.Errorf("canonical encoding is not compact: %s", c)
	}
	if !bytes.HasPrefix(c, []byte(`{"formatVersion":"diagram-json/v1","profileVersion":"process-diagram/v1","id":"OrderHandling","title":"a<b>&\"c\"","lanes":[`)) {
		t.Errorf("unexpected canonical prefix: %.140s", c)
	}
	if i, j := bytes.Index(c, []byte(`"id":"Check"`)), bytes.Index(c, []byte(`"id":"Start"`)); i < 0 || j < 0 || i > j {
		t.Errorf("nodes are not sorted by id")
	}
}
