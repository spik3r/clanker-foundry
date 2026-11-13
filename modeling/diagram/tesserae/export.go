// Package tesserae maps a deliberately bounded dashboard subset to the external
// Tesserae presentation contract. It contains no HTML, CSS or renderer code.
package tesserae

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"unicode"

	"github.com/spik3r/clanker-foundry/modeling/diagram/contract"
	"github.com/spik3r/clanker-foundry/modeling/diagram/dashboard"
)

type page struct {
	Version    string      `json:"version"`
	Kind       string      `json:"kind"`
	Title      string      `json:"title"`
	Category   string      `json:"category"`
	Footer     string      `json:"footer"`
	Subtitle   string      `json:"subtitle,omitempty"`
	Components []component `json:"components"`
}
type component struct {
	Type        string       `json:"type"`
	Title       string       `json:"title"`
	Value       *json.Number `json:"value,omitempty"`
	Description string       `json:"description,omitempty"`
	Columns     []string     `json:"columns,omitempty"`
	Rows        *[][]string  `json:"rows,omitempty"`
}

// Export validates dashboard input, then emits tesserae/v1 JSON. Only numeric
// stats and tables with at most eight selected columns are supported. Layout
// reading order is retained, but source grid row groupings are not represented.
func Export(raw []byte) ([]byte, error) {
	m, err := dashboard.Decode(raw)
	if err != nil {
		return nil, err
	}
	if err := heading(m.Title); err != nil {
		return nil, err
	}
	widgets := map[string]dashboard.Widget{}
	datasets := map[string]dashboard.Dataset{}
	for _, w := range m.Widgets {
		widgets[w.ID] = w
	}
	for _, d := range m.Datasets {
		datasets[d.ID] = d
	}
	out := page{Version: "tesserae/v1", Kind: "dashboard", Title: m.Title, Subtitle: m.Subtitle, Category: "Generated dashboard", Footer: "Generated from dashboard/v1 data.", Components: []component{}}
	for _, row := range m.Layout {
		for _, id := range row {
			w := widgets[id]
			var c component
			switch w.Kind {
			case dashboard.KindStat:
				n, _, err := dashboard.CanonicalNumber(w.Value)
				if err != nil {
					return nil, fmt.Errorf("widget %q: Tesserae metric requires an unformatted decimal value: %w", id, err)
				}
				value := json.Number(n)
				c = component{Type: "metric", Title: w.Label, Value: &value, Description: w.Note}
			case dashboard.KindTable:
				if len(w.Columns) > 8 {
					return nil, fmt.Errorf("widget %q: Tesserae tables support at most 8 selected columns", id)
				}
				ds := datasets[w.Dataset]
				indexes := map[string]int{}
				for i, col := range ds.Columns {
					indexes[col.ID] = i
				}
				c = component{Type: "table", Title: w.Title, Columns: []string{}}
				rows := [][]string{}
				for _, col := range w.Columns {
					c.Columns = append(c.Columns, ds.Columns[indexes[col]].Label)
				}
				for _, r := range ds.Rows {
					cells := []string{}
					for _, col := range w.Columns {
						cells = append(cells, r[indexes[col]].String())
					}
					rows = append(rows, cells)
				}
				c.Rows = &rows
			default:
				return nil, fmt.Errorf("widget %q: kind %q is not supported by the Tesserae export", id, w.Kind)
			}
			if err := heading(c.Title); err != nil {
				return nil, err
			}
			for _, label := range c.Columns {
				if err := heading(label); err != nil {
					return nil, err
				}
			}
			out.Components = append(out.Components, c)
		}
	}
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(out); err != nil {
		return nil, err
	}
	if buf.Len() > contract.MaxModelBytes {
		return nil, fmt.Errorf("Tesserae export exceeds %d bytes", contract.MaxModelBytes)
	}
	return buf.Bytes(), nil
}

// JavaScript trim also removes U+FEFF, which Go's TrimSpace does not.
func heading(s string) error {
	if strings.TrimFunc(s, func(r rune) bool { return unicode.IsSpace(r) || r == '\uFEFF' }) == "" {
		return fmt.Errorf("Tesserae heading must not be blank")
	}
	return nil
}
