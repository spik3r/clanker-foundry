package tui

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestFilterValidation(t *testing.T) {
	nine := []string{"ID", "Name", "Description", "Status", "A", "B", "C", "D", "E"}
	for name, c := range map[string]struct {
		fields []string
		want   string
	}{
		"unknown":    {[]string{"Nope"}, "undeclared field"},
		"duplicate":  {[]string{"Name", "Name"}, "duplicate filter field"},
		"empty":      {[]string{}, "1..8"},
		"too many":   {nine, "1..8"},
		"bad member": {[]string{"name"}, "undeclared field"},
	} {
		t.Run(name, func(t *testing.T) {
			m := exampleModel(t)
			for _, id := range nine[4:] {
				m.Fields = append(m.Fields, Field{ID: id, Label: id})
			}
			m.Filter = &Filter{Fields: c.fields}
			if err := m.Validate(); err == nil || !strings.Contains(err.Error(), c.want) {
				t.Fatalf("got %v, want %q", err, c.want)
			}
		})
	}
	m := exampleModel(t)
	m.Filter = &Filter{Fields: []string{"Name", "Description", "Status", "ID"}}
	if err := m.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestFilterDecodeRefusals(t *testing.T) {
	base := string(exampleBytes(t))
	add := func(member string) []byte {
		return []byte(strings.Replace(base, "\"layout\":", member+",\"layout\":", 1))
	}
	for name, raw := range map[string][]byte{
		"null":          add(`"filter": null`),
		"no fields":     add(`"filter": {}`),
		"empty fields":  add(`"filter": {"fields": []}`),
		"unknown":       add(`"filter": {"fields": ["Name"], "mode": "x"}`),
		"duplicate key": add(`"filter": {"fields": ["Name"], "fields": ["Name"]}`),
		"not a field":   add(`"filter": {"fields": ["Zzz"]}`),
		"string":        add(`"filter": "Name"`),
	} {
		if _, err := Decode(raw); err == nil {
			t.Errorf("%s: accepted", name)
		}
	}
	if _, err := Decode(add(`"filter": {"fields": ["Name"]}`)); err != nil {
		t.Fatal(err)
	}
}

func TestFilterFieldOrderIsNotSignificant(t *testing.T) {
	a, b := exampleModel(t), exampleModel(t)
	a.Filter = &Filter{Fields: []string{"Name", "Description"}}
	b.Filter = &Filter{Fields: []string{"Description", "Name"}}
	fa, err := Compile(a)
	if err != nil {
		t.Fatal(err)
	}
	fb, err := Compile(b)
	if err != nil {
		t.Fatal(err)
	}
	if CanonicalSHA256(fa) != CanonicalSHA256(fb) {
		t.Fatal("filter field order changed the canonical digest")
	}
	for name := range fa {
		if string(fa[name]) != string(fb[name]) {
			t.Fatalf("%s differs", name)
		}
	}
	if !strings.Contains(string(fa["model.json"]), `"filter": {`) {
		t.Fatal("canonical model lacks the filter")
	}
	if a.Filter.Fields[0] != "Name" {
		t.Fatal("Compile modified the caller's filter")
	}
	// A model with a filter differs from one without.
	if CanonicalSHA256(fa) == CanonicalSHA256(mustCompile(t, exampleModel(t))) {
		t.Fatal("filter did not change the digest")
	}
}

func mustCompile(t *testing.T, m Model) map[string][]byte {
	t.Helper()
	f, err := Compile(m)
	if err != nil {
		t.Fatal(err)
	}
	return f
}

// Models without a filter keep the digests they had before the member existed.
func TestFilterlessDigestsAreUnchanged(t *testing.T) {
	for file, want := range map[string]string{
		"examples/browser.json":         "6a8370de19f321f4a8bb0737745d04d86aa14e30203259d95fe5d79b27a8ac53",
		"examples/browser-package.json": "43e31a41805dd76946df369f36d53c50df3521a937b330a54d9a3cac6a122414",
		"examples/browser-hosted.json":  "82e00bf6ad341858b7ef314250157d2d72059965c4aaddcd3107135acbde0275",
	} {
		raw, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		m, err := Decode(raw)
		if err != nil {
			t.Fatal(err)
		}
		if got := CanonicalSHA256(mustCompile(t, m)); got != want {
			t.Errorf("%s: digest %s, want %s", file, got, want)
		}
		if strings.Contains(string(mustCompile(t, m)["model.json"]), "filter") {
			t.Errorf("%s: canonical model mentions filter", file)
		}
	}
}

func TestFilterCodeOnlyDrivesFilteredModels(t *testing.T) {
	plain := mustCompile(t, exampleModel(t))
	if !strings.Contains(string(plain["state.go"]), "filterEnabled = false") {
		t.Fatal("unfiltered model must have the filter disabled")
	}
	m := exampleModel(t)
	m.Filter = &Filter{Fields: []string{"Name"}}
	f := mustCompile(t, m)
	if !strings.Contains(string(f["state.go"]), "filterEnabled = true") || !strings.Contains(string(f["records.go"]), "fold(record.Name)") {
		t.Fatal("filtered model lacks the filter")
	}
	var probe struct{ Filter struct{ Fields []string } }
	if err := json.Unmarshal(f["model.json"], &probe); err != nil || len(probe.Filter.Fields) != 1 {
		t.Fatalf("model.json: %v", err)
	}
}
