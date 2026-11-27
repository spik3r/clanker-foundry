// Package architecture implements the architecture-diagram/v1 logical system view.
package architecture

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"github.com/spik3r/clanker-foundry/modeling/diagram/contract"
	"github.com/spik3r/clanker-foundry/modeling/diagram/internal/svgx"
)

const Profile = "architecture-diagram/v1"

type model struct {
	FormatVersion  string `json:"formatVersion"`
	ProfileVersion string `json:"profileVersion"`
	ID             string `json:"id"`
	Title          string `json:"title"`
	Nodes          []node `json:"nodes"`
	Links          []link `json:"links"`
}

type node struct {
	ID    string `json:"id"`
	Kind  string `json:"kind"`
	Label string `json:"label"`
}

type link struct {
	ID    string `json:"id"`
	From  string `json:"from"`
	To    string `json:"to"`
	Label string `json:"label"`
}

var schema = contract.Schema{
	"model": {Fields: map[string]string{
		"formatVersion": "", "profileVersion": "", "id": "", "title": "",
		"nodes": "nodes", "links": "links",
	}},
	"nodes": {Array: true, Elem: "node"},
	"links": {Array: true, Elem: "link"},
	"node":  {Fields: map[string]string{"id": "", "kind": "", "label": ""}},
	"link":  {Fields: map[string]string{"id": "", "from": "", "to": "", "label": ""}},
}

// Profiler selects, validates and renders this profile without external I/O.
type Profiler struct{}

func (Profiler) ID() string { return Profile }

func (Profiler) Generate(raw []byte, _ contract.Generator) (contract.Generated, error) {
	m, err := decode(raw)
	if err != nil {
		return contract.Generated{}, contract.ModelError{Err: err}
	}
	canonical, err := svgx.Compact(m)
	if err != nil {
		return contract.Generated{}, err
	}
	pretty, err := svgx.Encode(m, true)
	if err != nil {
		return contract.Generated{}, err
	}
	return contract.Generated{
		Canonical: canonical,
		Files:     map[string][]byte{"model.json": pretty, "diagram.svg": render(m)},
	}, nil
}

func decode(raw []byte) (model, error) {
	var m model
	if err := contract.StrictScan(raw, contract.MaxModelBytes, schema); err != nil {
		return m, err
	}
	if err := json.Unmarshal(raw, &m); err != nil {
		return m, err
	}
	if m.FormatVersion != svgx.Format || m.ProfileVersion != Profile || !svgx.Identifier.MatchString(m.ID) {
		return m, fmt.Errorf("expected %s / %s and a valid model id", svgx.Format, Profile)
	}
	if err := svgx.Plain(m.Title, "title", 120); err != nil {
		return m, err
	}
	if len(m.Nodes) < 1 || len(m.Nodes) > 12 || m.Links == nil || len(m.Links) > 24 {
		return m, fmt.Errorf("expected 1..12 nodes and a required links array of 0..24 links")
	}
	ids := make(map[string]bool)
	for _, n := range m.Nodes {
		if !svgx.Identifier.MatchString(n.ID) || ids[n.ID] {
			return m, fmt.Errorf("invalid or duplicate node id %q", n.ID)
		}
		ids[n.ID] = true
		switch n.Kind {
		case "actor", "service", "store", "external":
		default:
			return m, fmt.Errorf("node %s has unsupported kind %q", n.ID, n.Kind)
		}
		if err := svgx.Plain(n.Label, "node "+n.ID+" label", 96); err != nil {
			return m, err
		}
	}
	nodeIDs := make(map[string]bool, len(ids))
	for id := range ids {
		nodeIDs[id] = true
	}
	for _, l := range m.Links {
		if !svgx.Identifier.MatchString(l.ID) || ids[l.ID] {
			return m, fmt.Errorf("invalid or duplicate link id %q", l.ID)
		}
		ids[l.ID] = true
		if l.From == l.To || !nodeIDs[l.From] || !nodeIDs[l.To] {
			return m, fmt.Errorf("link %s needs two different existing nodes", l.ID)
		}
		if err := svgx.Plain(l.Label, "link "+l.ID+" label", 96); err != nil {
			return m, err
		}
	}
	// Node order is authored visual order. Link order is not meaningful.
	slices.SortFunc(m.Links, func(a, b link) int { return strings.Compare(a.ID, b.ID) })
	return m, nil
}
