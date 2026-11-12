// Package process implements the process-diagram/v1 profile: lanes, events,
// tasks, exclusive and parallel gateways and sequence flows, drawn as SVG and
// exported as a BPMN-style XML subset.
package process

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/spik3r/clanker-foundry/modeling/diagram/contract"
)

// Format and Profile identify the input format and the profile.
const (
	Format  = "diagram-json/v1"
	Profile = "process-diagram/v1"
)

// Node kinds.
const (
	KindStart     = "start"
	KindEnd       = "end"
	KindTask      = "task"
	KindExclusive = "exclusive"
	KindParallel  = "parallel"
)

// Limits.
const (
	MaxLanes = 8
	MaxNodes = 64
	MaxFlows = 128
)

// Model is a process diagram. It carries semantics only: no coordinates,
// sizes, colours or markup.
type Model struct {
	FormatVersion  string `json:"formatVersion"`
	ProfileVersion string `json:"profileVersion"`
	ID             string `json:"id"`
	Title          string `json:"title"`
	Lanes          []Lane `json:"lanes"`
	Nodes          []Node `json:"nodes"`
	Flows          []Flow `json:"flows"`
}

// Lane is a labelled horizontal band. Lane order is the visual order.
type Lane struct {
	ID    string `json:"id"`
	Label string `json:"label"`
}

// Node is a start or end event, a task or a gateway placed in one lane.
type Node struct {
	ID    string `json:"id"`
	Kind  string `json:"kind"`
	Label string `json:"label"`
	Lane  string `json:"lane"`
}

// Flow is a sequence flow. Label is optional.
type Flow struct {
	ID    string `json:"id"`
	From  string `json:"from"`
	To    string `json:"to"`
	Label string `json:"label,omitempty"`
}

var identifier = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_]{0,47}$`)

// schema is the strict member set; see contract.StrictScan.
var schema = contract.Schema{
	"model": {Fields: map[string]string{
		"formatVersion": "", "profileVersion": "", "id": "", "title": "",
		"lanes": "lanes", "nodes": "nodes", "flows": "flows",
	}},
	"lanes": {Array: true, Elem: "lane"},
	"nodes": {Array: true, Elem: "node"},
	"flows": {Array: true, Elem: "flow"},
	"lane":  {Fields: map[string]string{"id": "", "label": ""}},
	"node":  {Fields: map[string]string{"id": "", "kind": "", "label": "", "lane": ""}},
	"flow":  {Fields: map[string]string{"id": "", "from": "", "to": "", "label": ""}},
}

// Decode scans raw strictly, decodes it, validates it and returns the
// normalized model (nodes and flows sorted by id; lane order kept).
func Decode(raw []byte) (Model, error) {
	if err := contract.StrictScan(raw, contract.MaxModelBytes, schema); err != nil {
		return Model{}, err
	}
	// A flow label is optional, but when present it must not be empty. The
	// pointer tells absent from empty.
	var in struct {
		FormatVersion  string `json:"formatVersion"`
		ProfileVersion string `json:"profileVersion"`
		ID             string `json:"id"`
		Title          string `json:"title"`
		Lanes          []Lane `json:"lanes"`
		Nodes          []Node `json:"nodes"`
		Flows          []struct {
			ID    string  `json:"id"`
			From  string  `json:"from"`
			To    string  `json:"to"`
			Label *string `json:"label"`
		} `json:"flows"`
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	if err := d.Decode(&in); err != nil {
		return Model{}, err
	}
	m := Model{
		FormatVersion: in.FormatVersion, ProfileVersion: in.ProfileVersion,
		ID: in.ID, Title: in.Title, Lanes: in.Lanes, Nodes: in.Nodes,
	}
	for _, f := range in.Flows {
		fl := Flow{ID: f.ID, From: f.From, To: f.To}
		if f.Label != nil {
			if *f.Label == "" {
				return Model{}, fmt.Errorf("flow %q: label must not be empty when present", f.ID)
			}
			fl.Label = *f.Label
		}
		m.Flows = append(m.Flows, fl)
	}
	if err := m.Validate(); err != nil {
		return Model{}, err
	}
	return m.Normalize(), nil
}

// Normalize returns a copy with nodes and flows sorted by id. Lane order is
// significant and kept.
func (m Model) Normalize() Model {
	n := m
	n.Lanes = append([]Lane(nil), m.Lanes...)
	n.Nodes = append([]Node(nil), m.Nodes...)
	n.Flows = append([]Flow(nil), m.Flows...)
	sort.Slice(n.Nodes, func(i, j int) bool { return n.Nodes[i].ID < n.Nodes[j].ID })
	sort.Slice(n.Flows, func(i, j int) bool { return n.Flows[i].ID < n.Flows[j].ID })
	return n
}

// Canonical is the canonical encoding: compact JSON, fixed member order, nodes
// and flows sorted by id, lane order kept, no trailing newline, no HTML
// escaping. Call it on a validated model.
func (m Model) Canonical() ([]byte, error) {
	b, err := encode(m.Normalize(), false)
	if err != nil {
		return nil, err
	}
	return bytes.TrimSuffix(b, []byte("\n")), nil
}

func encode(v any, indent bool) ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if indent {
		enc.SetIndent("", "  ")
	}
	if err := enc.Encode(v); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func plain(s, name string) error {
	if len(s) == 0 || len(s) > 120 || !utf8.ValidString(s) {
		return fmt.Errorf("%s must be 1..120 UTF-8 bytes", name)
	}
	if strings.TrimSpace(s) == "" {
		return fmt.Errorf("%s must not be blank", name)
	}
	for _, r := range s {
		// U+FFFE and U+FFFF are not XML characters.
		if unicode.IsControl(r) || r == ' ' || r == ' ' || r == '￾' || r == '￿' {
			return fmt.Errorf("%s contains a control or non-XML character", name)
		}
	}
	return nil
}

// Validate applies every profile rule. It is also applied to programmatically
// constructed models before emission. Go validation is authoritative.
func (m Model) Validate() error {
	if m.FormatVersion != Format || m.ProfileVersion != Profile {
		return fmt.Errorf("unsupported format or profile")
	}
	if !identifier.MatchString(m.ID) {
		return fmt.Errorf("id must be an ASCII identifier")
	}
	if err := plain(m.Title, "title"); err != nil {
		return err
	}
	if len(m.Lanes) < 1 || len(m.Lanes) > MaxLanes {
		return fmt.Errorf("lanes must contain 1..%d lanes", MaxLanes)
	}
	if len(m.Nodes) < 2 || len(m.Nodes) > MaxNodes {
		return fmt.Errorf("nodes must contain 2..%d nodes", MaxNodes)
	}
	if len(m.Flows) < 1 || len(m.Flows) > MaxFlows {
		return fmt.Errorf("flows must contain 1..%d flows", MaxFlows)
	}
	ids := map[string]string{}
	claim := func(kind, id string) error {
		if !identifier.MatchString(id) {
			return fmt.Errorf("%s id %q must be an ASCII identifier", kind, id)
		}
		if prev, dup := ids[id]; dup {
			return fmt.Errorf("id %q is used by a %s and a %s", id, prev, kind)
		}
		ids[id] = kind
		return nil
	}
	lanes := map[string]bool{}
	for _, l := range m.Lanes {
		if err := claim("lane", l.ID); err != nil {
			return err
		}
		if err := plain(l.Label, "lane label"); err != nil {
			return err
		}
		lanes[l.ID] = true
	}
	kinds := map[string]string{}
	var start string
	starts, ends := 0, 0
	for _, n := range m.Nodes {
		if err := claim("node", n.ID); err != nil {
			return err
		}
		if err := plain(n.Label, "node label"); err != nil {
			return err
		}
		switch n.Kind {
		case KindStart:
			starts++
			start = n.ID
		case KindEnd:
			ends++
		case KindTask, KindExclusive, KindParallel:
		default:
			return fmt.Errorf("node %q: unknown kind %q", n.ID, n.Kind)
		}
		if !lanes[n.Lane] {
			return fmt.Errorf("node %q names unknown lane %q", n.ID, n.Lane)
		}
		kinds[n.ID] = n.Kind
	}
	if starts != 1 {
		return fmt.Errorf("exactly one start node is required, found %d", starts)
	}
	if ends < 1 {
		return fmt.Errorf("at least one end node is required")
	}
	in, out := map[string]int{}, map[string]int{}
	pairs := map[[2]string]bool{}
	for _, f := range m.Flows {
		if err := claim("flow", f.ID); err != nil {
			return err
		}
		if f.Label != "" {
			if err := plain(f.Label, "flow label"); err != nil {
				return err
			}
		}
		if kinds[f.From] == "" || kinds[f.To] == "" {
			return fmt.Errorf("flow %q names an unknown node", f.ID)
		}
		if f.From == f.To {
			return fmt.Errorf("flow %q is a self-loop", f.ID)
		}
		pair := [2]string{f.From, f.To}
		if pairs[pair] {
			return fmt.Errorf("flow %q repeats the flow from %q to %q", f.ID, f.From, f.To)
		}
		pairs[pair] = true
		out[f.From]++
		in[f.To]++
	}
	for _, n := range m.Nodes {
		i, o := in[n.ID], out[n.ID]
		switch n.Kind {
		case KindStart:
			if i != 0 || o != 1 {
				return fmt.Errorf("start %q needs no incoming and exactly one outgoing flow", n.ID)
			}
		case KindEnd:
			if o != 0 || i < 1 {
				return fmt.Errorf("end %q needs no outgoing and at least one incoming flow", n.ID)
			}
		case KindTask:
			if i < 1 || o != 1 {
				return fmt.Errorf("task %q needs at least one incoming and exactly one outgoing flow", n.ID)
			}
		default:
			if i < 1 || o < 1 || i+o < 3 {
				return fmt.Errorf("gateway %q needs at least one incoming, one outgoing and three flows in total", n.ID)
			}
		}
	}
	succ, pred := map[string][]string{}, map[string][]string{}
	for _, f := range m.Flows {
		succ[f.From] = append(succ[f.From], f.To)
		pred[f.To] = append(pred[f.To], f.From)
	}
	fromStart := reach(start, succ)
	for _, n := range m.Nodes {
		if !fromStart[n.ID] {
			return fmt.Errorf("node %q is not reachable from the start", n.ID)
		}
	}
	toEnd := map[string]bool{}
	for _, n := range m.Nodes {
		if n.Kind == KindEnd {
			for id := range reach(n.ID, pred) {
				toEnd[id] = true
			}
		}
	}
	for _, n := range m.Nodes {
		if !toEnd[n.ID] {
			return fmt.Errorf("node %q cannot reach an end", n.ID)
		}
	}
	return nil
}

func reach(from string, next map[string][]string) map[string]bool {
	seen := map[string]bool{from: true}
	stack := []string{from}
	for len(stack) > 0 {
		n := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		for _, t := range next[n] {
			if !seen[t] {
				seen[t] = true
				stack = append(stack, t)
			}
		}
	}
	return seen
}
