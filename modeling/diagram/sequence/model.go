// Package sequence implements the sequence-diagram/v1 profile: participants
// and actors with lifelines, ordered messages and notes, and optional
// non-nested alt, opt and loop frames, drawn as an inert SVG picture. The
// picture is UML-style; SEQUENCE.md lists exactly what is drawn.
package sequence

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/spik3r/clanker-foundry/modeling/diagram/contract"
	"github.com/spik3r/clanker-foundry/modeling/diagram/internal/svgx"
)

// Format and Profile identify the input format and the profile.
const (
	Format  = svgx.Format
	Profile = "sequence-diagram/v1"
)

// Participant kinds.
const (
	KindActor       = "actor"
	KindParticipant = "participant"
)

// Step kinds.
const (
	StepMessage = "message"
	StepNote    = "note"
)

// Message styles.
const (
	StyleSync  = "sync"
	StyleAsync = "async"
	StyleReply = "reply"
)

// Group kinds.
const (
	GroupAlt  = "alt"
	GroupOpt  = "opt"
	GroupLoop = "loop"
)

// Limits.
const (
	MinParticipants = 2
	MaxParticipants = 12
	MaxSteps        = 80
	MaxGroups       = 8
	maxText         = 120
)

// Model is a sequence diagram. It carries meaning only: no coordinates,
// sizes, colours or markup.
type Model struct {
	FormatVersion  string        `json:"formatVersion"`
	ProfileVersion string        `json:"profileVersion"`
	ID             string        `json:"id"`
	Title          string        `json:"title"`
	Participants   []Participant `json:"participants"`
	Steps          []Step        `json:"steps"`
	Groups         []Group       `json:"groups,omitempty"`
}

// Participant is a lifeline. Order is the left-to-right order.
type Participant struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	Kind  string `json:"kind"`
}

// Step is a message or a note. Order is the top-to-bottom order. A message
// uses From, To and Style; a note uses Over. Label is common.
type Step struct {
	ID    string   `json:"id"`
	Kind  string   `json:"kind"`
	From  string   `json:"from,omitempty"`
	To    string   `json:"to,omitempty"`
	Over  []string `json:"over,omitempty"`
	Label string   `json:"label"`
	Style string   `json:"style,omitempty"`
}

// Group is a frame around a contiguous range of steps.
type Group struct {
	ID    string `json:"id"`
	Kind  string `json:"kind"`
	Label string `json:"label"`
	First string `json:"first"`
	Last  string `json:"last"`
}

// schema is the strict member set; see contract.StrictScan.
var schema = contract.Schema{
	"model": {Fields: map[string]string{
		"formatVersion": "", "profileVersion": "", "id": "", "title": "",
		"participants": "participants", "steps": "steps", "groups": "groups",
	}},
	"participants": {Array: true, Elem: "participant"},
	"steps":        {Array: true, Elem: "step"},
	"groups":       {Array: true, Elem: "group"},
	"over":         {Array: true},
	"participant":  {Fields: map[string]string{"id": "", "label": "", "kind": ""}},
	"step": {Fields: map[string]string{
		"id": "", "kind": "", "from": "", "to": "", "over": "over", "label": "", "style": "",
	}},
	"group": {Fields: map[string]string{"id": "", "kind": "", "label": "", "first": "", "last": ""}},
}

// Decode scans raw strictly, decodes it, validates it and returns the
// normalized model (groups sorted by id; participant and step order kept).
func Decode(raw []byte) (Model, error) {
	if err := contract.StrictScan(raw, contract.MaxModelBytes, schema); err != nil {
		return Model{}, err
	}
	// Pointers tell an absent member from an empty one, so that a message
	// cannot carry an empty "over" and a note cannot carry an empty "from".
	var in struct {
		FormatVersion  string        `json:"formatVersion"`
		ProfileVersion string        `json:"profileVersion"`
		ID             string        `json:"id"`
		Title          string        `json:"title"`
		Participants   []Participant `json:"participants"`
		Steps          []struct {
			ID    string    `json:"id"`
			Kind  string    `json:"kind"`
			From  *string   `json:"from"`
			To    *string   `json:"to"`
			Over  *[]string `json:"over"`
			Label *string   `json:"label"`
			Style *string   `json:"style"`
		} `json:"steps"`
		Groups []Group `json:"groups"`
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	if err := d.Decode(&in); err != nil {
		return Model{}, err
	}
	m := Model{
		FormatVersion: in.FormatVersion, ProfileVersion: in.ProfileVersion,
		ID: in.ID, Title: in.Title, Participants: in.Participants, Groups: in.Groups,
	}
	for _, s := range in.Steps {
		st := Step{ID: s.ID, Kind: s.Kind}
		if s.Label == nil {
			return Model{}, fmt.Errorf("step %q: label is required", s.ID)
		}
		st.Label = *s.Label
		switch s.Kind {
		case StepMessage:
			if s.Over != nil {
				return Model{}, fmt.Errorf("step %q: a message must not have over", s.ID)
			}
			if s.From == nil || s.To == nil || s.Style == nil {
				return Model{}, fmt.Errorf("step %q: a message needs from, to and style", s.ID)
			}
			st.From, st.To, st.Style = *s.From, *s.To, *s.Style
		case StepNote:
			if s.From != nil || s.To != nil || s.Style != nil {
				return Model{}, fmt.Errorf("step %q: a note must not have from, to or style", s.ID)
			}
			if s.Over == nil {
				return Model{}, fmt.Errorf("step %q: a note needs over", s.ID)
			}
			st.Over = *s.Over
		default:
			return Model{}, fmt.Errorf("step %q: unknown kind %q", s.ID, s.Kind)
		}
		m.Steps = append(m.Steps, st)
	}
	if err := m.Validate(); err != nil {
		return Model{}, err
	}
	return m.Normalize(), nil
}

// Normalize returns a copy with groups sorted by id. Participant order and
// step order are significant and kept.
func (m Model) Normalize() Model {
	n := m
	n.Participants = append([]Participant(nil), m.Participants...)
	n.Steps = make([]Step, len(m.Steps))
	for i, s := range m.Steps {
		s.Over = append([]string(nil), s.Over...)
		if len(s.Over) == 0 {
			s.Over = nil
		}
		n.Steps[i] = s
	}
	n.Groups = append([]Group(nil), m.Groups...)
	sort.Slice(n.Groups, func(i, j int) bool { return n.Groups[i].ID < n.Groups[j].ID })
	if len(n.Groups) == 0 {
		n.Groups = nil
	}
	return n
}

// Canonical is the canonical encoding: compact JSON, fixed member order,
// groups sorted by id, participant and step order kept, no trailing newline,
// no HTML escaping. Call it on a validated model.
func (m Model) Canonical() ([]byte, error) { return svgx.Compact(m.Normalize()) }

// Validate applies every profile rule. It is also applied to programmatically
// constructed models before emission. Go validation is authoritative.
func (m Model) Validate() error {
	if m.FormatVersion != Format || m.ProfileVersion != Profile {
		return fmt.Errorf("unsupported format or profile")
	}
	if !svgx.Identifier.MatchString(m.ID) {
		return fmt.Errorf("id must be an ASCII identifier")
	}
	if err := svgx.Plain(m.Title, "title", maxText); err != nil {
		return err
	}
	if len(m.Participants) < MinParticipants || len(m.Participants) > MaxParticipants {
		return fmt.Errorf("participants must contain %d..%d participants", MinParticipants, MaxParticipants)
	}
	if len(m.Steps) < 1 || len(m.Steps) > MaxSteps {
		return fmt.Errorf("steps must contain 1..%d steps", MaxSteps)
	}
	if len(m.Groups) > MaxGroups {
		return fmt.Errorf("groups must contain at most %d groups", MaxGroups)
	}
	ids := map[string]string{}
	claim := func(kind, id string) error {
		if !svgx.Identifier.MatchString(id) {
			return fmt.Errorf("%s id %q must be an ASCII identifier", kind, id)
		}
		if prev, dup := ids[id]; dup {
			return fmt.Errorf("id %q is used by a %s and a %s", id, prev, kind)
		}
		ids[id] = kind
		return nil
	}
	known := map[string]bool{}
	for _, p := range m.Participants {
		if err := claim("participant", p.ID); err != nil {
			return err
		}
		if err := svgx.Plain(p.Label, "participant label", maxText); err != nil {
			return err
		}
		if p.Kind != KindActor && p.Kind != KindParticipant {
			return fmt.Errorf("participant %q: unknown kind %q", p.ID, p.Kind)
		}
		known[p.ID] = true
	}
	stepIndex := map[string]int{}
	for i, s := range m.Steps {
		if err := claim("step", s.ID); err != nil {
			return err
		}
		if err := svgx.Plain(s.Label, "step label", maxText); err != nil {
			return err
		}
		switch s.Kind {
		case StepMessage:
			if !known[s.From] || !known[s.To] {
				return fmt.Errorf("message %q names an unknown participant", s.ID)
			}
			if s.Style != StyleSync && s.Style != StyleAsync && s.Style != StyleReply {
				return fmt.Errorf("message %q: unknown style %q", s.ID, s.Style)
			}
			if len(s.Over) != 0 {
				return fmt.Errorf("message %q must not have over", s.ID)
			}
		case StepNote:
			if len(s.Over) < 1 || len(s.Over) > 2 {
				return fmt.Errorf("note %q: over must name 1 or 2 participants", s.ID)
			}
			seen := map[string]bool{}
			for _, o := range s.Over {
				if !known[o] {
					return fmt.Errorf("note %q names an unknown participant %q", s.ID, o)
				}
				if seen[o] {
					return fmt.Errorf("note %q names participant %q twice", s.ID, o)
				}
				seen[o] = true
			}
			if s.From != "" || s.To != "" || s.Style != "" {
				return fmt.Errorf("note %q must not have from, to or style", s.ID)
			}
		default:
			return fmt.Errorf("step %q: unknown kind %q", s.ID, s.Kind)
		}
		stepIndex[s.ID] = i
	}
	type span struct {
		id          string
		first, last int
	}
	var spans []span
	for _, g := range m.Groups {
		if err := claim("group", g.ID); err != nil {
			return err
		}
		if g.Kind != GroupAlt && g.Kind != GroupOpt && g.Kind != GroupLoop {
			return fmt.Errorf("group %q: unknown kind %q", g.ID, g.Kind)
		}
		if err := svgx.Plain(g.Label, "group label", maxText); err != nil {
			return err
		}
		f, okF := stepIndex[g.First]
		l, okL := stepIndex[g.Last]
		if !okF || !okL {
			return fmt.Errorf("group %q names an unknown step", g.ID)
		}
		if f > l {
			return fmt.Errorf("group %q: first step %q comes after last step %q", g.ID, g.First, g.Last)
		}
		spans = append(spans, span{g.ID, f, l})
	}
	sort.Slice(spans, func(i, j int) bool { return spans[i].first < spans[j].first })
	for i := 1; i < len(spans); i++ {
		if spans[i].first <= spans[i-1].last {
			return fmt.Errorf("groups %q and %q overlap", spans[i-1].id, spans[i].id)
		}
	}
	return nil
}
