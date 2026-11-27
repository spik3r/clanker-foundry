package architecture

import (
	"fmt"
	"maps"
	"slices"
)

// Change describes an addition, removal or field edit identified by model ID.
type Change struct {
	Kind   string `json:"kind"`
	ID     string `json:"id"`
	Field  string `json:"field,omitempty"`
	Before string `json:"before,omitempty"`
	After  string `json:"after,omitempty"`
}

// DiffResult separates changes to the graph from layout order.
type DiffResult struct {
	Changes          []Change `json:"changes"`
	NodeOrderChanged bool     `json:"nodeOrderChanged"`
}

// Diff compares two validated versions of the same architecture model.
// Changes are ordered by entity class (title, node, link), ID and fixed field order.
func Diff(before, after []byte) (DiffResult, error) {
	a, err := decode(before)
	if err != nil {
		return DiffResult{}, fmt.Errorf("before: %w", err)
	}
	b, err := decode(after)
	if err != nil {
		return DiffResult{}, fmt.Errorf("after: %w", err)
	}
	if a.ID != b.ID {
		return DiffResult{}, fmt.Errorf("model IDs differ: %q and %q", a.ID, b.ID)
	}

	d := DiffResult{Changes: []Change{}}
	edit := func(kind, id, field, old, next string) {
		if old != next {
			d.Changes = append(d.Changes, Change{Kind: kind, ID: id, Field: field, Before: old, After: next})
		}
	}
	edit("title-changed", a.ID, "title", a.Title, b.Title)
	an, bn := keyed(a.Nodes, func(n node) string { return n.ID }), keyed(b.Nodes, func(n node) string { return n.ID })
	al, bl := keyed(a.Links, func(l link) string { return l.ID }), keyed(b.Links, func(l link) string { return l.ID })
	compare(&d, "node", an, bn, func(id string, old, next node) {
		edit("node-changed", id, "kind", old.Kind, next.Kind)
		edit("node-changed", id, "label", old.Label, next.Label)
	})
	compare(&d, "link", al, bl, func(id string, old, next link) {
		edit("link-changed", id, "from", old.From, next.From)
		edit("link-changed", id, "to", old.To, next.To)
		edit("link-changed", id, "label", old.Label, next.Label)
	})
	d.NodeOrderChanged = !slices.Equal(survivors(a.Nodes, bn), survivors(b.Nodes, an))
	return d, nil
}

func keyed[T any](values []T, id func(T) string) map[string]T {
	out := make(map[string]T, len(values))
	for _, value := range values {
		out[id(value)] = value
	}
	return out
}

func compare[T any](d *DiffResult, kind string, a, b map[string]T, edit func(string, T, T)) {
	ids := make(map[string]struct{}, len(a)+len(b))
	for id := range a {
		ids[id] = struct{}{}
	}
	for id := range b {
		ids[id] = struct{}{}
	}
	for _, id := range slices.Sorted(maps.Keys(ids)) {
		old, existed := a[id]
		next, exists := b[id]
		switch {
		case !existed:
			d.Changes = append(d.Changes, Change{Kind: kind + "-added", ID: id})
		case !exists:
			d.Changes = append(d.Changes, Change{Kind: kind + "-removed", ID: id})
		default:
			edit(id, old, next)
		}
	}
}

func survivors(order []node, other map[string]node) []string {
	var ids []string
	for _, n := range order {
		if _, ok := other[n.ID]; ok {
			ids = append(ids, n.ID)
		}
	}
	return ids
}
