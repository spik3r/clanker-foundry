// Package topology reports properties of the declared state graph only. It
// does not evaluate guards or establish executable reachability or liveness.
package topology

import (
	"bytes"
	"encoding/json"
	"errors"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/canonical"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/jsonformat"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/model"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/strictjson"
)

const (
	FormatVersion        = "lifecycle-structural-topology/v1"
	RequestFormatVersion = "lifecycle-structural-topology-request/v1"
	MaxRequestBytes      = 2 << 20
	MaxModelBytes        = 1 << 20
	MaxOutputBytes       = 1 << 20
	MaxDepth             = 66
)

var ErrInvalid = errors.New("invalid structural topology input")

type State struct {
	ID              string   `json:"id"`
	HasTerminalPath bool     `json:"hasTerminalPath"`
	TerminalID      string   `json:"terminalId"`
	TransitionIDs   []string `json:"transitionIds"`
}
type Cycle struct {
	States              []string `json:"states"`
	InternalTransitions []string `json:"internalTransitions"`
	OutgoingTransitions []string `json:"outgoingTransitions"`
}
type Report struct {
	FormatVersion string  `json:"formatVersion"`
	Mode          string  `json:"mode"`
	Authority     string  `json:"authority"`
	ModelID       string  `json:"modelId"`
	ModelDigest   string  `json:"modelDigest"`
	Initial       string  `json:"initial"`
	States        []State `json:"states"`
	Cycles        []Cycle `json:"cycles"`
}
type node struct {
	id   string
	path []string
}

// Analyze validates and normalizes a complete model before deriving any output.
// Paths are graph declarations only. The returned slices are independently owned;
// concurrent mutation of the input is unsupported.
func Analyze(input model.Model) (Report, error) {
	m, err := model.Normalize(input)
	if err != nil {
		return Report{}, ErrInvalid
	}
	digest, err := canonical.Digest(m)
	if err != nil {
		return Report{}, ErrInvalid
	}
	index := make(map[string]int, len(m.States))
	terminal := make(map[string]bool, len(m.States))
	for i, s := range m.States {
		index[s.ID] = i
		terminal[s.ID] = s.Terminal
	}
	// Normalize sorts transition IDs, so each outgoing adjacency list inherits
	// that order. BFS first arrival is shortest and then lexicographically first.
	edges := make(map[string][]model.Transition, len(m.States))
	n := len(m.States)
	reach := make([][]bool, n)
	for i := range reach {
		reach[i] = make([]bool, n)
		reach[i][i] = true
	}
	for _, t := range m.Transitions {
		edges[t.From] = append(edges[t.From], t)
		reach[index[t.From]][index[t.To]] = true
	}
	for k := 0; k < n; k++ {
		for i := 0; i < n; i++ {
			for j := 0; j < n; j++ {
				reach[i][j] = reach[i][j] || (reach[i][k] && reach[k][j])
			}
		}
	}
	out := Report{FormatVersion: FormatVersion, Mode: "structural-only", Authority: "none", ModelID: m.ID, ModelDigest: digest, Initial: m.Initial, States: []State{}, Cycles: []Cycle{}}
	for _, s := range m.States {
		item := State{ID: s.ID, TransitionIDs: []string{}}
		seen := map[string]bool{s.ID: true}
		queue := []node{{id: s.ID, path: []string{}}}
		for len(queue) > 0 {
			current := queue[0]
			queue = queue[1:]
			if terminal[current.id] {
				item.HasTerminalPath = true
				item.TerminalID = current.id
				item.TransitionIDs = append([]string{}, current.path...)
				break
			}
			for _, edge := range edges[current.id] {
				if !seen[edge.To] {
					seen[edge.To] = true
					path := append(append([]string{}, current.path...), edge.ID)
					queue = append(queue, node{id: edge.To, path: path})
				}
			}
		}
		out.States = append(out.States, item)
	}
	used := make([]bool, n)
	for i := 0; i < n; i++ {
		if used[i] {
			continue
		}
		component := []string{}
		members := map[string]bool{}
		for j := 0; j < n; j++ {
			if !used[j] && reach[i][j] && reach[j][i] {
				used[j] = true
				component = append(component, m.States[j].ID)
				members[m.States[j].ID] = true
			}
		}
		internal := []string{}
		outgoing := []string{}
		selfLoop := false
		for _, t := range m.Transitions {
			if members[t.From] {
				if members[t.To] {
					internal = append(internal, t.ID)
					if t.From == t.To {
						selfLoop = true
					}
				} else {
					outgoing = append(outgoing, t.ID)
				}
			}
		}
		if len(component) > 1 || selfLoop {
			out.Cycles = append(out.Cycles, Cycle{States: component, InternalTransitions: internal, OutgoingTransitions: outgoing})
		}
	}
	if _, err := encode(out); err != nil {
		return Report{}, ErrInvalid
	}
	return out, nil
}
func encode(v Report) ([]byte, error) {
	var b bytes.Buffer
	e := json.NewEncoder(&b)
	e.SetEscapeHTML(false)
	if e.Encode(v) != nil || b.Len() > MaxOutputBytes {
		return nil, ErrInvalid
	}
	return b.Bytes(), nil
}

type request struct {
	FormatVersion string          `json:"formatVersion"`
	Model         json.RawMessage `json:"model"`
}

// Execute decodes a strict wrapper and returns complete owned JSON bytes.
func Execute(raw []byte) ([]byte, error) {
	var req request
	if strictjson.Decode(raw, &req, MaxRequestBytes, MaxDepth) != nil || req.FormatVersion != RequestFormatVersion || len(req.Model) > MaxModelBytes {
		return nil, ErrInvalid
	}
	m, err := jsonformat.Decode(req.Model)
	if err != nil {
		return nil, ErrInvalid
	}
	report, err := Analyze(m)
	if err != nil {
		return nil, ErrInvalid
	}
	return encode(report)
}
