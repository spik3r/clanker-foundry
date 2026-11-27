package architecture

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/spik3r/clanker-foundry/modeling/diagram/contract"
	"github.com/spik3r/clanker-foundry/modeling/diagram/internal/svgcheck"
)

const example = `{
  "formatVersion":"diagram-json/v1","profileVersion":"architecture-diagram/v1",
  "id":"Demo","title":"Order & fulfilment",
  "nodes":[{"id":"API","kind":"service","label":"Order <API>"},{"id":"DB","kind":"store","label":"Orders"}],
  "links":[{"id":"Save","from":"API","to":"DB","label":"Write & commit"},{"id":"Notify","from":"DB","to":"API","label":"Notify <caller>"}]
}`

func TestInvalidModel(t *testing.T) {
	cases := map[string]string{
		"unknown field":     strings.Replace(example, `"kind":"store"`, `"kind":"store","x":10`, 1),
		"unknown kind":      strings.Replace(example, `"kind":"store"`, `"kind":"database"`, 1),
		"duplicate field":   strings.Replace(example, `"id":"DB"`, `"id":"DB","id":"DB"`, 1),
		"duplicate node id": strings.Replace(example, `"id":"DB"`, `"id":"API"`, 1),
		"colliding link id": strings.Replace(example, `"id":"Save"`, `"id":"API"`, 1),
		"unknown endpoint":  strings.Replace(example, `"to":"DB"`, `"to":"Absent"`, 1),
		"self link":         strings.Replace(example, `"to":"DB"`, `"to":"API"`, 1),
		"blank label":       strings.Replace(example, `"Orders"`, `" "`, 1),
		"invalid Unicode":   strings.Replace(example, `"Orders"`, `"\ud800"`, 1),
		"missing links":     strings.TrimSuffix(strings.TrimSpace(example[:strings.Index(example, `"links":`)]), ",") + "}",
	}
	for name, raw := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := (Profiler{}).Generate([]byte(raw), contract.Generator{}); err == nil {
				t.Fatal("accepted invalid model")
			}
		})
	}
}

func TestCanonicalAndSVG(t *testing.T) {
	p := Profiler{}
	orig, err := p.Generate([]byte(example), contract.Generator{})
	if err != nil {
		t.Fatal(err)
	}
	var m model
	if err := json.Unmarshal([]byte(example), &m); err != nil {
		t.Fatal(err)
	}
	slices.Reverse(m.Links)
	raw, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	reordered, err := p.Generate(raw, contract.Generator{})
	if err != nil {
		t.Fatal(err)
	}
	if string(reordered.Canonical) != string(orig.Canonical) {
		t.Fatal("unordered links changed canonical bytes")
	}
	for name, data := range orig.Files {
		if string(reordered.Files[name]) != string(data) {
			t.Errorf("unordered links changed %s", name)
		}
	}
	slices.Reverse(m.Nodes)
	raw, err = json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	reordered, err = p.Generate(raw, contract.Generator{})
	if err != nil {
		t.Fatal(err)
	}
	if string(reordered.Files["diagram.svg"]) == string(orig.Files["diagram.svg"]) {
		t.Fatal("authored node order was ignored")
	}

	svg := orig.Files["diagram.svg"]
	if err := svgcheck.Inert(svg); err != nil {
		t.Fatal(err)
	}
	elems, err := svgcheck.Parse(svg)
	if err != nil {
		t.Fatal(err)
	}
	var gotLabel bool
	tracks := map[string]string{}
	for _, e := range elems {
		if e.Name == "text" && e.Text == "Order <API>" {
			gotLabel = true
		}
		if id := e.Attr["data-link"]; id != "" {
			d := e.Attr["d"]
			var startX, startY, trackX, endY, endX int
			if _, err := fmt.Sscanf(d, "M %d %d H %d V %d H %d", &startX, &startY, &trackX, &endY, &endX); err != nil {
				t.Fatal(err)
			}
			dir := 1
			if e.Attr["data-from"] == "DB" {
				dir = -1
			}
			if startX != endX || trackX <= startX || (endY-startY)*dir <= 0 {
				t.Errorf("invalid route %q", d)
			}
			tracks[id] = e.Attr["data-from"] + "->" + e.Attr["data-to"]
		}
	}
	if !gotLabel || len(tracks) != 2 || tracks["Save"] != "API->DB" || tracks["Notify"] != "DB->API" {
		t.Fatalf("wrong SVG label or directed links: gotLabel=%v, tracks=%v", gotLabel, tracks)
	}
}
