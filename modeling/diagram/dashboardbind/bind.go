// Package dashboardbind replaces only data in a reviewed dashboard/v1 model.
// It has no renderer, expressions, file access or network access.
package dashboardbind

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/spik3r/clanker-foundry/modeling/diagram/contract"
	"github.com/spik3r/clanker-foundry/modeling/diagram/dashboard"
)

const Version = "dashboard-data/v1"

var schema = contract.Schema{
	"model":    {Fields: map[string]string{"dataVersion": "", "templateSHA256": "", "datasets": "datasets", "stats": "stats"}},
	"datasets": {Array: true, Elem: "dataset"},
	"dataset":  {Fields: map[string]string{"id": "", "rows": "rows"}},
	"rows":     {Array: true, Elem: "row"},
	"row":      {Array: true},
	"stats":    {Array: true, Elem: "stat"},
	"stat":     {Fields: map[string]string{"id": "", "value": "", "note": ""}},
}

type dataset struct {
	ID   string               `json:"id"`
	Rows *[][]json.RawMessage `json:"rows"`
}
type stat struct {
	ID    string  `json:"id"`
	Value string  `json:"value"`
	Note  *string `json:"note"`
}
type payload struct {
	DataVersion    string     `json:"dataVersion"`
	TemplateSHA256 string     `json:"templateSHA256"`
	Datasets       *[]dataset `json:"datasets"`
	Stats          *[]stat    `json:"stats"`
}

// Bind returns a canonical dashboard model, not HTML. Every dataset and stat
// must be supplied exactly once, including empty arrays when neither exists.
// The exact template bytes are pinned, including whitespace. No default data
// survives: omitted stat notes are removed, not inherited from the template.
func Bind(template, data []byte) ([]byte, error) {
	base, err := dashboard.Decode(template)
	if err != nil {
		return nil, fmt.Errorf("template: %w", err)
	}
	if err := contract.StrictScan(data, contract.MaxModelBytes, schema); err != nil {
		return nil, fmt.Errorf("data: %w", err)
	}
	var p payload
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, fmt.Errorf("data: %w", err)
	}
	if p.DataVersion != Version {
		return nil, fmt.Errorf("unsupported dataVersion")
	}
	sum := sha256.Sum256(template)
	if p.TemplateSHA256 != hex.EncodeToString(sum[:]) {
		return nil, fmt.Errorf("templateSHA256 does not match exact template bytes")
	}
	if p.Datasets == nil || p.Stats == nil {
		return nil, fmt.Errorf("datasets and stats arrays are required")
	}
	if len(*p.Datasets) != len(base.Datasets) {
		return nil, fmt.Errorf("data must replace every dataset exactly once")
	}
	statCount := 0
	for _, w := range base.Widgets {
		if w.Kind == dashboard.KindStat {
			statCount++
		}
	}
	if len(*p.Stats) != statCount {
		return nil, fmt.Errorf("data must replace every stat exactly once")
	}
	dsByID := map[string]dataset{}
	for _, ds := range *p.Datasets {
		if _, exists := dsByID[ds.ID]; exists {
			return nil, fmt.Errorf("duplicate dataset id")
		}
		if ds.Rows == nil {
			return nil, fmt.Errorf("dataset rows are required")
		}
		dsByID[ds.ID] = ds
	}
	statsByID := map[string]stat{}
	for _, s := range *p.Stats {
		if _, exists := statsByID[s.ID]; exists {
			return nil, fmt.Errorf("duplicate stat id")
		}
		statsByID[s.ID] = s
	}
	// Keep number tokens as RawMessage until dashboard.Decode applies the existing
	// decimal grammar and bounds. Never round-trip report values through float64.
	canonical, err := base.Canonical()
	if err != nil {
		return nil, err
	}
	var model map[string]json.RawMessage
	if err := json.Unmarshal(canonical, &model); err != nil {
		return nil, err
	}
	datasets := make([]map[string]json.RawMessage, 0, len(base.Datasets))
	for _, ds := range base.Datasets {
		replacement, ok := dsByID[ds.ID]
		if !ok {
			return nil, fmt.Errorf("data is missing template dataset %q", ds.ID)
		}
		id, _ := json.Marshal(ds.ID)
		columns, _ := json.Marshal(ds.Columns)
		rows, err := json.Marshal(*replacement.Rows)
		if err != nil {
			return nil, err
		}
		datasets = append(datasets, map[string]json.RawMessage{"id": id, "columns": columns, "rows": rows})
	}
	model["datasets"], err = json.Marshal(datasets)
	if err != nil {
		return nil, err
	}
	var widgets []map[string]json.RawMessage
	if err := json.Unmarshal(model["widgets"], &widgets); err != nil {
		return nil, err
	}
	for i, w := range base.Widgets {
		if w.Kind != dashboard.KindStat {
			continue
		}
		s, ok := statsByID[w.ID]
		if !ok {
			return nil, fmt.Errorf("data is missing template stat %q", w.ID)
		}
		widgets[i]["value"], _ = json.Marshal(s.Value)
		delete(widgets[i], "note")
		if s.Note != nil {
			widgets[i]["note"], _ = json.Marshal(*s.Note)
		}
	}
	model["widgets"], err = json.Marshal(widgets)
	if err != nil {
		return nil, err
	}
	raw, err := json.Marshal(model)
	if err != nil {
		return nil, err
	}
	bound, err := dashboard.Decode(raw)
	if err != nil {
		return nil, fmt.Errorf("bound dashboard: %w", err)
	}
	result, err := bound.Canonical()
	if err != nil {
		return nil, err
	}
	var pretty bytes.Buffer
	if err := json.Indent(&pretty, result, "", "  "); err != nil {
		return nil, err
	}
	pretty.WriteByte('\n')
	return pretty.Bytes(), nil
}
