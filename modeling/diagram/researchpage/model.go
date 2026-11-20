// Package researchpage implements the research-page/v1 profile: the fixed
// text of a research page (title, purpose, status, bottom line, section
// headings and the method list) emitted as a SolidJS component that renders
// the host's page shell. Interactive and computed parts stay hand-written and
// are named as slots. The profile is shaped around one host's page outline; see
// RESEARCHPAGE.md.
package researchpage

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"

	"github.com/spik3r/clanker-foundry/modeling/diagram/contract"
	"github.com/spik3r/clanker-foundry/modeling/diagram/internal/modeltext"
)

// Format and Profile identify the input format and the profile.
const (
	Format  = "diagram-json/v1"
	Profile = "research-page/v1"
)

// Tabs, in the order the host's shell shows them.
const (
	TabExplore = "explore"
	TabMethod  = "method"
	TabSources = "sources"
)

// tabOrder is the canonical section order.
var tabOrder = []string{TabExplore, TabMethod, TabSources}

// Item kinds.
const (
	KindSlot        = "slot"
	KindPanel       = "panel"
	KindHeading     = "heading"
	KindMethodPanel = "methodPanel"
	KindProse       = "prose"
)

// Limits.
const (
	MaxTitleBytes    = 120
	MaxTextBytes     = 600
	MaxContext       = 6
	MaxParagraphs    = 6
	MaxPoints        = 12
	MaxItemsPerTab   = 24
	maxSlotIDLength  = 32
	maxModelIDLength = 48
)

// statuses are the values the host's ResearchBottomLine accepts, in the order
// of its STATUS_COPY table.
var statuses = []string{"supported", "mixed", "inconclusive", "experimental", "unavailable", "not-yet-run", "unknown"}

// Model is a research page.
type Model struct {
	FormatVersion  string    `json:"formatVersion"`
	ProfileVersion string    `json:"profileVersion"`
	ID             string    `json:"id"`
	Title          string    `json:"title"`
	Purpose        string    `json:"purpose"`
	Status         string    `json:"status"`
	Class          string    `json:"class,omitempty"`
	BottomLine     Bottom    `json:"bottomLine"`
	Sections       []Section `json:"sections"`
}

// Bottom is the fixed text of the bottom-line card.
type Bottom struct {
	Finding    string   `json:"finding"`
	Confidence string   `json:"confidence"`
	Limitation string   `json:"limitation"`
	Context    []string `json:"context,omitempty"`
}

// Section is the content of one tab. Item order is significant.
type Section struct {
	Tab   string `json:"tab"`
	Items []Item `json:"items"`
}

// Item is one element of a tab. Which members are used depends on Kind: slot
// uses ID; panel uses Slot; heading uses Title and optional Description;
// methodPanel uses Title, optional Description and Points; prose uses
// Paragraphs.
type Item struct {
	Kind        string   `json:"kind"`
	ID          string   `json:"id,omitempty"`
	Slot        string   `json:"slot,omitempty"`
	Title       string   `json:"title,omitempty"`
	Description string   `json:"description,omitempty"`
	Points      []string `json:"points,omitempty"`
	Paragraphs  []string `json:"paragraphs,omitempty"`
}

var (
	modelID   = regexp.MustCompile(`^[A-Z][A-Za-z0-9]{0,47}$`)
	slotID    = regexp.MustCompile(`^[a-z][A-Za-z0-9]{0,31}$`)
	classTok  = regexp.MustCompile(`^[a-z][a-z0-9-]{0,63}$`)
	protoName = map[string]bool{
		"constructor": true, "hasOwnProperty": true, "isPrototypeOf": true,
		"propertyIsEnumerable": true, "toLocaleString": true, "toString": true, "valueOf": true,
	}
)

// allowed lists the members each item kind may carry besides kind.
var allowed = map[string][]string{
	KindSlot:        {"id"},
	KindPanel:       {"slot"},
	KindHeading:     {"title", "description"},
	KindMethodPanel: {"title", "description", "points"},
	KindProse:       {"paragraphs"},
}

// schema is the strict member set; see contract.StrictScan.
var schema = contract.Schema{
	"model": {Fields: map[string]string{
		"formatVersion": "", "profileVersion": "", "id": "", "title": "", "purpose": "", "status": "", "class": "",
		"bottomLine": "bottom", "sections": "sections",
	}},
	"bottom":   {Fields: map[string]string{"finding": "", "confidence": "", "limitation": "", "context": "strings"}},
	"strings":  {Array: true},
	"sections": {Array: true, Elem: "section"},
	"section":  {Fields: map[string]string{"tab": "", "items": "items"}},
	"items":    {Array: true, Elem: "item"},
	"item": {Fields: map[string]string{
		"kind": "", "id": "", "slot": "", "title": "", "description": "", "points": "strings", "paragraphs": "strings",
	}},
}

// Decode scans raw strictly, decodes it, validates it and returns the
// normalized model (sections in tab order).
func Decode(raw []byte) (Model, error) {
	if err := contract.StrictScan(raw, contract.MaxModelBytes, schema); err != nil {
		return Model{}, err
	}
	var in struct {
		FormatVersion  string  `json:"formatVersion"`
		ProfileVersion string  `json:"profileVersion"`
		ID             string  `json:"id"`
		Title          string  `json:"title"`
		Purpose        string  `json:"purpose"`
		Status         string  `json:"status"`
		Class          *string `json:"class"`
		BottomLine     struct {
			Finding    string    `json:"finding"`
			Confidence string    `json:"confidence"`
			Limitation string    `json:"limitation"`
			Context    *[]string `json:"context"`
		} `json:"bottomLine"`
		Sections []struct {
			Tab   string `json:"tab"`
			Items []struct {
				Kind        string    `json:"kind"`
				ID          *string   `json:"id"`
				Slot        *string   `json:"slot"`
				Title       *string   `json:"title"`
				Description *string   `json:"description"`
				Points      *[]string `json:"points"`
				Paragraphs  *[]string `json:"paragraphs"`
			} `json:"items"`
		} `json:"sections"`
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	if err := d.Decode(&in); err != nil {
		return Model{}, err
	}
	m := Model{
		FormatVersion: in.FormatVersion, ProfileVersion: in.ProfileVersion,
		ID: in.ID, Title: in.Title, Purpose: in.Purpose, Status: in.Status,
		BottomLine: Bottom{Finding: in.BottomLine.Finding, Confidence: in.BottomLine.Confidence, Limitation: in.BottomLine.Limitation},
		Sections:   []Section{},
	}
	if in.Class != nil {
		if *in.Class == "" {
			return Model{}, fmt.Errorf("class must not be empty when present")
		}
		m.Class = *in.Class
	}
	if c := in.BottomLine.Context; c != nil {
		if len(*c) == 0 {
			return Model{}, fmt.Errorf("context must not be empty when present")
		}
		m.BottomLine.Context = *c
	}
	for _, sec := range in.Sections {
		out := Section{Tab: sec.Tab, Items: []Item{}}
		for i, it := range sec.Items {
			x := Item{Kind: it.Kind}
			// Presence matters: a member the kind does not use is refused even
			// when it is empty, and an optional member that is present must
			// not be empty.
			present := map[string]bool{}
			str := func(name string, p *string, dst *string) {
				if p != nil {
					present[name] = true
					*dst = *p
				}
			}
			list := func(name string, p *[]string, dst *[]string) {
				if p != nil {
					present[name] = true
					*dst = *p
				}
			}
			str("id", it.ID, &x.ID)
			str("slot", it.Slot, &x.Slot)
			str("title", it.Title, &x.Title)
			str("description", it.Description, &x.Description)
			list("points", it.Points, &x.Points)
			list("paragraphs", it.Paragraphs, &x.Paragraphs)
			use, ok := allowed[x.Kind]
			if !ok {
				return Model{}, fmt.Errorf("section %q item %d: unknown kind %q", sec.Tab, i+1, x.Kind)
			}
			okName := map[string]bool{}
			for _, n := range use {
				okName[n] = true
			}
			var extra []string
			for name := range present {
				if !okName[name] {
					extra = append(extra, name)
				}
			}
			if len(extra) > 0 {
				sort.Strings(extra)
				return Model{}, fmt.Errorf("section %q item %d: a %s item must not have %s", sec.Tab, i+1, x.Kind, joinNames(extra))
			}
			if present["description"] && x.Description == "" {
				return Model{}, fmt.Errorf("section %q item %d: description must not be empty when present", sec.Tab, i+1)
			}
			out.Items = append(out.Items, x)
		}
		m.Sections = append(m.Sections, out)
	}
	if err := m.Validate(); err != nil {
		return Model{}, err
	}
	return m.Normalize(), nil
}

func joinNames(names []string) string {
	s := ""
	for i, n := range names {
		if i > 0 {
			s += ", "
		}
		s += n
	}
	return s
}

// Normalize returns a copy with sections in tab order. Item order, point
// order and paragraph order are significant and kept.
func (m Model) Normalize() Model {
	n := m
	n.Sections = append([]Section{}, m.Sections...)
	rank := map[string]int{}
	for i, t := range tabOrder {
		rank[t] = i
	}
	sort.SliceStable(n.Sections, func(i, j int) bool { return rank[n.Sections[i].Tab] < rank[n.Sections[j].Tab] })
	return n
}

// Canonical is the canonical encoding: compact JSON, members in the struct
// order, sections in tab order, no trailing newline, no HTML escaping. Call
// it on a validated model.
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

// SlotIDs returns the slot ids of a normalized model in output order: tab
// order, then item order.
func (m Model) SlotIDs() []string {
	ids := []string{}
	for _, sec := range m.Sections {
		for _, it := range sec.Items {
			switch it.Kind {
			case KindSlot:
				ids = append(ids, it.ID)
			case KindPanel:
				ids = append(ids, it.Slot)
			}
		}
	}
	return ids
}

// Validate applies every profile rule. It is also applied to programmatically
// constructed models. Go validation is authoritative.
func (m Model) Validate() error {
	if m.FormatVersion != Format || m.ProfileVersion != Profile {
		return fmt.Errorf("unsupported format or profile")
	}
	if !modelID.MatchString(m.ID) {
		return fmt.Errorf("id must be an ASCII identifier of 1..%d characters starting with an upper-case letter", maxModelIDLength)
	}
	if err := modeltext.Plain(m.Title, "title", MaxTitleBytes); err != nil {
		return err
	}
	if err := modeltext.Plain(m.Purpose, "purpose", MaxTextBytes); err != nil {
		return err
	}
	if !contains(statuses, m.Status) {
		return fmt.Errorf("status must be one of %s", joinNames(statuses))
	}
	if m.Class != "" && !classTok.MatchString(m.Class) {
		return fmt.Errorf("class must be one CSS class token: lower-case letters, digits and hyphens, starting with a letter, at most 64 characters")
	}
	b := m.BottomLine
	if err := modeltext.Plain(b.Finding, "finding", MaxTextBytes); err != nil {
		return err
	}
	if err := modeltext.Plain(b.Confidence, "confidence", MaxTitleBytes); err != nil {
		return err
	}
	if err := modeltext.Plain(b.Limitation, "limitation", MaxTextBytes); err != nil {
		return err
	}
	if err := paragraphs(b.Context, "context", 0, MaxContext); err != nil {
		return err
	}
	if len(m.Sections) < 1 || len(m.Sections) > len(tabOrder) {
		return fmt.Errorf("sections must contain 1..%d sections", len(tabOrder))
	}
	tabs := map[string]bool{}
	slots := map[string]bool{}
	for _, sec := range m.Sections {
		if !contains(tabOrder, sec.Tab) {
			return fmt.Errorf("tab must be one of %s", joinNames(tabOrder))
		}
		if tabs[sec.Tab] {
			return fmt.Errorf("tab %q is used twice", sec.Tab)
		}
		tabs[sec.Tab] = true
		if len(sec.Items) < 1 || len(sec.Items) > MaxItemsPerTab {
			return fmt.Errorf("tab %q: items must contain 1..%d items", sec.Tab, MaxItemsPerTab)
		}
		for i, it := range sec.Items {
			if err := validateItem(it, slots); err != nil {
				return fmt.Errorf("tab %q item %d: %w", sec.Tab, i+1, err)
			}
		}
	}
	return nil
}

func contains(list []string, s string) bool {
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
}

// paragraphs checks a list of plain paragraphs of min..max entries.
func paragraphs(list []string, name string, min, max int) error {
	if len(list) < min || len(list) > max {
		return fmt.Errorf("%s must hold %d..%d entries", name, min, max)
	}
	for i, p := range list {
		if err := modeltext.Plain(p, fmt.Sprintf("%s %d", name, i+1), MaxTextBytes); err != nil {
			return err
		}
	}
	return nil
}

func validateItem(it Item, slots map[string]bool) error {
	has := map[string]bool{
		"id": it.ID != "", "slot": it.Slot != "", "title": it.Title != "", "description": it.Description != "",
		"points": len(it.Points) > 0, "paragraphs": len(it.Paragraphs) > 0,
	}
	use, ok := allowed[it.Kind]
	if !ok {
		return fmt.Errorf("unknown kind %q", it.Kind)
	}
	okName := map[string]bool{}
	for _, n := range use {
		okName[n] = true
	}
	var extra []string
	for name, present := range has {
		if present && !okName[name] {
			extra = append(extra, name)
		}
	}
	if len(extra) > 0 {
		sort.Strings(extra)
		return fmt.Errorf("a %s item must not have %s", it.Kind, joinNames(extra))
	}
	claim := func(id string) error {
		switch {
		case !slotID.MatchString(id):
			return fmt.Errorf("slot id %q must match [a-z][A-Za-z0-9]{0,%d}", id, maxSlotIDLength-1)
		case protoName[id]:
			return fmt.Errorf("slot id %q is an inherited object property name", id)
		case slots[id]:
			return fmt.Errorf("slot id %q is used twice", id)
		}
		slots[id] = true
		return nil
	}
	heading := func() error {
		if err := modeltext.Plain(it.Title, "title", MaxTitleBytes); err != nil {
			return err
		}
		if it.Description != "" {
			return modeltext.Plain(it.Description, "description", MaxTextBytes)
		}
		return nil
	}
	switch it.Kind {
	case KindSlot:
		return claim(it.ID)
	case KindPanel:
		return claim(it.Slot)
	case KindHeading:
		return heading()
	case KindMethodPanel:
		if err := heading(); err != nil {
			return err
		}
		return paragraphs(it.Points, "points", 1, MaxPoints)
	default: // KindProse
		return paragraphs(it.Paragraphs, "paragraphs", 1, MaxParagraphs)
	}
}
