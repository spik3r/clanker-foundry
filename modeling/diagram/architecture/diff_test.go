package architecture

import (
	"encoding/json"
	"reflect"
	"slices"
	"strings"
	"testing"
)

func TestDiffCanonicalEquivalenceAndOrder(t *testing.T) {
	before := []byte(example)
	m, err := decode(before)
	if err != nil {
		t.Fatal(err)
	}
	slices.Reverse(m.Links)
	raw := encodeForDiff(t, m)
	got, err := Diff(before, raw)
	if err != nil || !reflect.DeepEqual(got, DiffResult{Changes: []Change{}}) {
		t.Fatalf("equivalent diff = %+v, %v", got, err)
	}
	m.Nodes = append([]node{{ID: "User", Kind: "actor", Label: "User"}}, m.Nodes...)
	got, err = Diff(before, encodeForDiff(t, m))
	want := DiffResult{Changes: []Change{{Kind: "node-added", ID: "User"}}}
	if err != nil || !reflect.DeepEqual(got, want) {
		t.Fatalf("insertion diff = %+v, %v", got, err)
	}
	slices.Reverse(m.Nodes)
	got, err = Diff(before, encodeForDiff(t, m))
	want.NodeOrderChanged = true
	if err != nil || !reflect.DeepEqual(got, want) {
		t.Fatalf("reordered diff = %+v, %v", got, err)
	}
}

func TestDiffGraphChanges(t *testing.T) {
	before := []byte(example)
	m, err := decode(before)
	if err != nil {
		t.Fatal(err)
	}
	m.Title = "New title"
	m.Nodes[1].Kind, m.Nodes[1].Label = "external", "Remote storage"
	m.Links[0] = link{ID: "Alert", From: "DB", To: "API", Label: "Alert"}
	m.Links[1].From, m.Links[1].To = m.Links[1].To, m.Links[1].From
	raw := encodeForDiff(t, m)
	want := DiffResult{Changes: []Change{
		{Kind: "title-changed", ID: "Demo", Field: "title", Before: "Order & fulfilment", After: "New title"},
		{Kind: "node-changed", ID: "DB", Field: "kind", Before: "store", After: "external"},
		{Kind: "node-changed", ID: "DB", Field: "label", Before: "Orders", After: "Remote storage"},
		{Kind: "link-added", ID: "Alert"},
		{Kind: "link-removed", ID: "Notify"},
		{Kind: "link-changed", ID: "Save", Field: "from", Before: "API", After: "DB"},
		{Kind: "link-changed", ID: "Save", Field: "to", Before: "DB", After: "API"},
	}}
	for range 12 {
		got, err := Diff(before, raw)
		if err != nil || !reflect.DeepEqual(got, want) {
			t.Fatalf("graph diff = %+v, %v; want %+v", got, err, want)
		}
	}
}

func TestDiffRefusesInvalidAndMismatchedModels(t *testing.T) {
	for _, c := range []struct{ before, after, errorText string }{
		{"{", example, "before:"},
		{example, strings.Replace(example, `"to":"DB"`, `"to":"Missing"`, 1), "after:"},
		{example, strings.Replace(example, Profile, "process-diagram/v1", 1), "after:"},
		{example, strings.Replace(example, `"id":"Demo"`, `"id":"Other"`, 1), "model IDs differ"},
	} {
		if got, err := Diff([]byte(c.before), []byte(c.after)); err == nil || !strings.Contains(err.Error(), c.errorText) || got.Changes != nil {
			t.Errorf("diff = %+v, %v; want refusal with %q", got, err, c.errorText)
		}
	}
}

func encodeForDiff(t *testing.T, m model) []byte {
	t.Helper()
	encoded, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	return encoded
}
