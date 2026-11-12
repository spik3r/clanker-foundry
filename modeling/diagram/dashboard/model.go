// Package dashboard implements the dashboard/v1 profile: datasets and widgets
// (stat, bar, line, table) arranged in rows, emitted as one self-contained,
// inert HTML file. A model carries meaning and data only: no markup,
// coordinates, colours or script.
package dashboard

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
	Profile = "dashboard/v1"
)

// Widget kinds.
const (
	KindStat  = "stat"
	KindBar   = "bar"
	KindLine  = "line"
	KindTable = "table"
)

// Column types.
const (
	TypeText   = "text"
	TypeNumber = "number"
)

// Limits.
const (
	MaxDatasets     = 8
	MaxColumns      = 12
	MaxRows         = 200
	MaxChartRows    = 60
	MaxWidgets      = 24
	MaxSeries       = 4
	MaxLayoutRows   = 12
	MaxLayoutPerRow = 4
	MaxTextBytes    = 120
)

// Model is a dashboard.
type Model struct {
	FormatVersion  string     `json:"formatVersion"`
	ProfileVersion string     `json:"profileVersion"`
	ID             string     `json:"id"`
	Title          string     `json:"title"`
	Subtitle       string     `json:"subtitle,omitempty"`
	Datasets       []Dataset  `json:"datasets"`
	Widgets        []Widget   `json:"widgets"`
	Layout         [][]string `json:"layout"`
}

// Dataset is a small table of typed columns. Column and row order are
// significant.
type Dataset struct {
	ID      string   `json:"id"`
	Columns []Column `json:"columns"`
	Rows    [][]Cell `json:"rows"`
}

// Column is one typed column of a dataset.
type Column struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	Type  string `json:"type"`
}

// Cell is one value. A number is kept as canonical decimal text (Num) and as
// a fixed-point count of millionths (Micro); it never passes through float64.
type Cell struct {
	IsNum bool
	Text  string
	Num   string
	Micro int64
}

// MarshalJSON writes a text cell as a JSON string without HTML escaping and a
// number cell as its canonical decimal text.
func (c Cell) MarshalJSON() ([]byte, error) {
	if c.IsNum {
		return []byte(c.Num), nil
	}
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(c.Text); err != nil {
		return nil, err
	}
	return bytes.TrimSuffix(buf.Bytes(), []byte("\n")), nil
}

// String returns the cell as display text: the string, or the canonical number.
func (c Cell) String() string {
	if c.IsNum {
		return c.Num
	}
	return c.Text
}

// Widget is one dashboard element. Which members are used depends on Kind:
// stat uses Label, Value and Note; bar and line use Title, Dataset, Category
// and Series; table uses Title, Dataset and Columns.
type Widget struct {
	ID       string   `json:"id"`
	Kind     string   `json:"kind"`
	Label    string   `json:"label,omitempty"`
	Value    string   `json:"value,omitempty"`
	Note     string   `json:"note,omitempty"`
	Title    string   `json:"title,omitempty"`
	Dataset  string   `json:"dataset,omitempty"`
	Category string   `json:"category,omitempty"`
	Series   []string `json:"series,omitempty"`
	Columns  []string `json:"columns,omitempty"`
}

var identifier = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_]{0,47}$`)

// schema is the strict member set; see contract.StrictScan.
var schema = contract.Schema{
	"model": {Fields: map[string]string{
		"formatVersion": "", "profileVersion": "", "id": "", "title": "", "subtitle": "",
		"datasets": "datasets", "widgets": "widgets", "layout": "layout",
	}},
	"datasets":  {Array: true, Elem: "dataset"},
	"dataset":   {Fields: map[string]string{"id": "", "columns": "dsColumns", "rows": "rows"}},
	"dsColumns": {Array: true, Elem: "column"},
	"column":    {Fields: map[string]string{"id": "", "label": "", "type": ""}},
	"rows":      {Array: true, Elem: "row"},
	"row":       {Array: true},
	"widgets":   {Array: true, Elem: "widget"},
	"widget": {Fields: map[string]string{
		"id": "", "kind": "", "label": "", "value": "", "note": "", "title": "",
		"dataset": "", "category": "", "series": "strings", "columns": "strings",
	}},
	"strings":   {Array: true},
	"layout":    {Array: true, Elem: "layoutRow"},
	"layoutRow": {Array: true},
}

// allowed lists the members each widget kind may carry besides id and kind.
var allowed = map[string][]string{
	KindStat:  {"label", "value", "note"},
	KindBar:   {"title", "dataset", "category", "series"},
	KindLine:  {"title", "dataset", "category", "series"},
	KindTable: {"title", "dataset", "columns"},
}

// Decode scans raw strictly, decodes it, validates it and returns the
// normalized model (datasets and widgets sorted by id).
func Decode(raw []byte) (Model, error) {
	if err := contract.StrictScan(raw, contract.MaxModelBytes, schema); err != nil {
		return Model{}, err
	}
	var in struct {
		FormatVersion  string  `json:"formatVersion"`
		ProfileVersion string  `json:"profileVersion"`
		ID             string  `json:"id"`
		Title          string  `json:"title"`
		Subtitle       *string `json:"subtitle"`
		Datasets       []struct {
			ID      string              `json:"id"`
			Columns []Column            `json:"columns"`
			Rows    [][]json.RawMessage `json:"rows"`
		} `json:"datasets"`
		Widgets []map[string]json.RawMessage `json:"widgets"`
		Layout  [][]string                   `json:"layout"`
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	if err := d.Decode(&in); err != nil {
		return Model{}, err
	}
	m := Model{
		FormatVersion: in.FormatVersion, ProfileVersion: in.ProfileVersion,
		ID: in.ID, Title: in.Title, Layout: in.Layout,
		Datasets: []Dataset{}, Widgets: []Widget{},
	}
	if in.Subtitle != nil {
		if *in.Subtitle == "" {
			return Model{}, fmt.Errorf("subtitle must not be empty when present")
		}
		m.Subtitle = *in.Subtitle
	}
	for _, ds := range in.Datasets {
		out := Dataset{ID: ds.ID, Columns: ds.Columns, Rows: [][]Cell{}}
		for r, row := range ds.Rows {
			cells := make([]Cell, 0, len(row))
			for c, tok := range row {
				cell, err := decodeCell(tok)
				if err != nil {
					return Model{}, fmt.Errorf("dataset %q row %d column %d: %w", ds.ID, r+1, c+1, err)
				}
				cells = append(cells, cell)
			}
			out.Rows = append(out.Rows, cells)
		}
		m.Datasets = append(m.Datasets, out)
	}
	for i, wm := range in.Widgets {
		w, err := decodeWidget(wm)
		if err != nil {
			return Model{}, fmt.Errorf("widget %d: %w", i+1, err)
		}
		m.Widgets = append(m.Widgets, w)
	}
	if err := m.Validate(); err != nil {
		return Model{}, err
	}
	return m.Normalize(), nil
}

// decodeWidget reads one widget. Presence matters: a member the widget's kind
// does not use is refused even when it is empty, and an optional member that
// is present must not be empty.
func decodeWidget(wm map[string]json.RawMessage) (Widget, error) {
	var w Widget
	str := func(name string, dst *string) error {
		raw, ok := wm[name]
		if !ok {
			return nil
		}
		return json.Unmarshal(raw, dst)
	}
	list := func(name string, dst *[]string) error {
		raw, ok := wm[name]
		if !ok {
			return nil
		}
		return json.Unmarshal(raw, dst)
	}
	for _, f := range []struct {
		name string
		dst  *string
	}{{"id", &w.ID}, {"kind", &w.Kind}, {"label", &w.Label}, {"value", &w.Value}, {"note", &w.Note},
		{"title", &w.Title}, {"dataset", &w.Dataset}, {"category", &w.Category}} {
		if err := str(f.name, f.dst); err != nil {
			return Widget{}, fmt.Errorf("%s: %w", f.name, err)
		}
	}
	if err := list("series", &w.Series); err != nil {
		return Widget{}, fmt.Errorf("series: %w", err)
	}
	if err := list("columns", &w.Columns); err != nil {
		return Widget{}, fmt.Errorf("columns: %w", err)
	}
	use, ok := allowed[w.Kind]
	if !ok {
		return Widget{}, fmt.Errorf("unknown kind %q", w.Kind)
	}
	okName := map[string]bool{"id": true, "kind": true}
	for _, n := range use {
		okName[n] = true
	}
	var extra []string
	for name := range wm {
		if !okName[name] {
			extra = append(extra, name)
		}
	}
	if len(extra) > 0 {
		sort.Strings(extra)
		return Widget{}, fmt.Errorf("a %s widget must not have %s", w.Kind, strings.Join(extra, ", "))
	}
	if _, present := wm["note"]; present && w.Note == "" {
		return Widget{}, fmt.Errorf("note must not be empty when present")
	}
	return w, nil
}

// decodeCell reads one cell token: a JSON string or a number written without
// exponent.
func decodeCell(tok json.RawMessage) (Cell, error) {
	if len(tok) > 0 && tok[0] == '"' {
		var s string
		if err := json.Unmarshal(tok, &s); err != nil {
			return Cell{}, err
		}
		return Cell{Text: s}, nil
	}
	switch string(tok) {
	case "true", "false":
		return Cell{}, fmt.Errorf("a cell must be a string or a number")
	}
	num, micro, err := CanonicalNumber(string(tok))
	if err != nil {
		return Cell{}, err
	}
	return Cell{IsNum: true, Num: num, Micro: micro}, nil
}

var numberToken = regexp.MustCompile(`^(-?)(0|[1-9][0-9]*)(?:\.([0-9]+))?$`)

// CanonicalNumber checks a JSON number token and returns its canonical decimal
// text and its value in millionths. The token is read as decimal text, never as
// a float. Rules: no exponent; trailing fractional zeros are dropped; at most 6
// fractional digits remain; at most 15 significant digits remain (digits after
// leading zeros, counting integer digits and the remaining fractional digits);
// the magnitude is below 1e12; negative zero becomes 0.
func CanonicalNumber(tok string) (string, int64, error) {
	if strings.ContainsAny(tok, "eE") {
		return "", 0, fmt.Errorf("number %q must be written without an exponent", tok)
	}
	g := numberToken.FindStringSubmatch(tok)
	if g == nil {
		return "", 0, fmt.Errorf("%q is not a JSON number", tok)
	}
	neg, intPart, frac := g[1] == "-", g[2], strings.TrimRight(g[3], "0")
	if len(intPart) > 12 {
		return "", 0, fmt.Errorf("number %q is not below 1e12 in magnitude", tok)
	}
	if len(frac) > 6 {
		return "", 0, fmt.Errorf("number %q has more than 6 fractional digits", tok)
	}
	sig := len(strings.TrimLeft(intPart+frac, "0"))
	if sig > 15 {
		return "", 0, fmt.Errorf("number %q has more than 15 significant digits", tok)
	}
	var micro int64
	for _, r := range intPart + (frac + "000000")[:6] {
		micro = micro*10 + int64(r-'0')
	}
	if micro == 0 {
		return "0", 0, nil
	}
	text := intPart
	if frac != "" {
		text += "." + frac
	}
	if neg {
		return "-" + text, -micro, nil
	}
	return text, micro, nil
}

// FormatMicro writes millionths as canonical decimal text.
func FormatMicro(v int64) string {
	neg := v < 0
	if neg {
		v = -v
	}
	s := fmt.Sprintf("%d", v/1_000_000)
	if f := v % 1_000_000; f != 0 {
		s += "." + strings.TrimRight(fmt.Sprintf("%06d", f), "0")
	}
	if neg {
		s = "-" + s
	}
	return s
}

// Normalize returns a copy with datasets and widgets sorted by id. Column,
// row, series, table-column and layout order are significant and kept.
func (m Model) Normalize() Model {
	n := m
	n.Datasets = append([]Dataset{}, m.Datasets...)
	n.Widgets = append([]Widget{}, m.Widgets...)
	sort.Slice(n.Datasets, func(i, j int) bool { return n.Datasets[i].ID < n.Datasets[j].ID })
	sort.Slice(n.Widgets, func(i, j int) bool { return n.Widgets[i].ID < n.Widgets[j].ID })
	return n
}

// Canonical is the canonical encoding: compact JSON, members in the struct
// order, datasets and widgets sorted by id, numbers in canonical decimal text,
// no trailing newline, no HTML escaping. Call it on a validated model.
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

func plainRunes(s, name string) error {
	if len(s) > MaxTextBytes || !utf8.ValidString(s) {
		return fmt.Errorf("%s must be at most %d UTF-8 bytes", name, MaxTextBytes)
	}
	for _, r := range s {
		if unicode.IsControl(r) || r == ' ' || r == ' ' || r == '￾' || r == '￿' {
			return fmt.Errorf("%s contains a control or non-XML character", name)
		}
	}
	return nil
}

// plain checks a label or title: 1..120 bytes, not blank.
func plain(s, name string) error {
	if len(s) == 0 || len(s) > MaxTextBytes {
		return fmt.Errorf("%s must be 1..%d UTF-8 bytes", name, MaxTextBytes)
	}
	if err := plainRunes(s, name); err != nil {
		return err
	}
	if strings.TrimSpace(s) == "" {
		return fmt.Errorf("%s must not be blank", name)
	}
	return nil
}

// Validate applies every profile rule. It is also applied to programmatically
// constructed models. Go validation is authoritative.
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
	if m.Subtitle != "" {
		if err := plain(m.Subtitle, "subtitle"); err != nil {
			return err
		}
	}
	if len(m.Datasets) > MaxDatasets {
		return fmt.Errorf("datasets must contain 0..%d datasets", MaxDatasets)
	}
	sets := map[string]*Dataset{}
	for i := range m.Datasets {
		ds := &m.Datasets[i]
		if err := validateDataset(ds); err != nil {
			return err
		}
		if sets[ds.ID] != nil {
			return fmt.Errorf("dataset id %q is used twice", ds.ID)
		}
		sets[ds.ID] = ds
	}
	if len(m.Widgets) < 1 || len(m.Widgets) > MaxWidgets {
		return fmt.Errorf("widgets must contain 1..%d widgets", MaxWidgets)
	}
	ids := map[string]bool{}
	for _, w := range m.Widgets {
		if !identifier.MatchString(w.ID) {
			return fmt.Errorf("widget id %q must be an ASCII identifier", w.ID)
		}
		if ids[w.ID] {
			return fmt.Errorf("widget id %q is used twice", w.ID)
		}
		ids[w.ID] = true
		if err := validateWidget(w, sets); err != nil {
			return fmt.Errorf("widget %q: %w", w.ID, err)
		}
	}
	return validateLayout(m.Layout, ids)
}

func validateDataset(ds *Dataset) error {
	if !identifier.MatchString(ds.ID) {
		return fmt.Errorf("dataset id %q must be an ASCII identifier", ds.ID)
	}
	if len(ds.Columns) < 1 || len(ds.Columns) > MaxColumns {
		return fmt.Errorf("dataset %q: columns must contain 1..%d columns", ds.ID, MaxColumns)
	}
	seen := map[string]bool{}
	for _, c := range ds.Columns {
		if !identifier.MatchString(c.ID) {
			return fmt.Errorf("dataset %q: column id %q must be an ASCII identifier", ds.ID, c.ID)
		}
		if seen[c.ID] {
			return fmt.Errorf("dataset %q: column id %q is used twice", ds.ID, c.ID)
		}
		seen[c.ID] = true
		if err := plain(c.Label, "column label"); err != nil {
			return fmt.Errorf("dataset %q: %w", ds.ID, err)
		}
		if c.Type != TypeText && c.Type != TypeNumber {
			return fmt.Errorf("dataset %q: column %q has unknown type %q", ds.ID, c.ID, c.Type)
		}
	}
	if len(ds.Rows) > MaxRows {
		return fmt.Errorf("dataset %q: rows must contain 0..%d rows", ds.ID, MaxRows)
	}
	for r, row := range ds.Rows {
		if len(row) != len(ds.Columns) {
			return fmt.Errorf("dataset %q row %d: has %d cells, want %d", ds.ID, r+1, len(row), len(ds.Columns))
		}
		for c, cell := range row {
			col := ds.Columns[c]
			where := fmt.Sprintf("dataset %q row %d column %q", ds.ID, r+1, col.ID)
			if (col.Type == TypeNumber) != cell.IsNum {
				return fmt.Errorf("%s: cell does not match column type %s", where, col.Type)
			}
			if cell.IsNum {
				num, micro, err := CanonicalNumber(cell.Num)
				if err != nil || num != cell.Num || micro != cell.Micro {
					return fmt.Errorf("%s: number is not canonical", where)
				}
			} else if err := plainRunes(cell.Text, "text cell"); err != nil {
				return fmt.Errorf("%s: %w", where, err)
			}
		}
	}
	return nil
}

func validateWidget(w Widget, sets map[string]*Dataset) error {
	has := map[string]bool{
		"label": w.Label != "", "value": w.Value != "", "note": w.Note != "",
		"title": w.Title != "", "dataset": w.Dataset != "", "category": w.Category != "",
		"series": len(w.Series) > 0, "columns": len(w.Columns) > 0,
	}
	use, ok := allowed[w.Kind]
	if !ok {
		return fmt.Errorf("unknown kind %q", w.Kind)
	}
	okName := map[string]bool{}
	for _, n := range use {
		okName[n] = true
	}
	for name, present := range has {
		if present && !okName[name] {
			return fmt.Errorf("a %s widget must not have %s", w.Kind, name)
		}
	}
	if w.Kind == KindStat {
		if err := plain(w.Label, "label"); err != nil {
			return err
		}
		if err := plain(w.Value, "value"); err != nil {
			return err
		}
		if w.Note != "" {
			return plain(w.Note, "note")
		}
		return nil
	}
	if err := plain(w.Title, "title"); err != nil {
		return err
	}
	ds := sets[w.Dataset]
	if ds == nil {
		return fmt.Errorf("dataset %q does not exist", w.Dataset)
	}
	cols := map[string]string{}
	for _, c := range ds.Columns {
		cols[c.ID] = c.Type
	}
	list := func(name string, ids []string, min, max int, want string) error {
		if len(ids) < min || len(ids) > max {
			return fmt.Errorf("%s must name %d..%d columns", name, min, max)
		}
		seen := map[string]bool{}
		for _, id := range ids {
			t, ok := cols[id]
			switch {
			case !ok:
				return fmt.Errorf("%s names unknown column %q of dataset %q", name, id, ds.ID)
			case seen[id]:
				return fmt.Errorf("%s names column %q twice", name, id)
			case want != "" && t != want:
				return fmt.Errorf("%s column %q must be of type %s", name, id, want)
			}
			seen[id] = true
		}
		return nil
	}
	if w.Kind == KindTable {
		return list("columns", w.Columns, 1, MaxColumns, "")
	}
	if len(ds.Rows) < 1 || len(ds.Rows) > MaxChartRows {
		return fmt.Errorf("dataset %q must have 1..%d rows for a chart", ds.ID, MaxChartRows)
	}
	if err := list("category", []string{w.Category}, 1, 1, TypeText); err != nil {
		return err
	}
	return list("series", w.Series, 1, MaxSeries, TypeNumber)
}

func validateLayout(layout [][]string, widgets map[string]bool) error {
	if len(layout) < 1 || len(layout) > MaxLayoutRows {
		return fmt.Errorf("layout must contain 1..%d rows", MaxLayoutRows)
	}
	placed := map[string]bool{}
	for i, row := range layout {
		if len(row) < 1 || len(row) > MaxLayoutPerRow {
			return fmt.Errorf("layout row %d must hold 1..%d widgets", i+1, MaxLayoutPerRow)
		}
		for _, id := range row {
			if !widgets[id] {
				return fmt.Errorf("layout row %d names unknown widget %q", i+1, id)
			}
			if placed[id] {
				return fmt.Errorf("layout places widget %q twice", id)
			}
			placed[id] = true
		}
	}
	if len(placed) != len(widgets) {
		return fmt.Errorf("layout must place every widget exactly once")
	}
	return nil
}
