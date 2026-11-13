// Package class implements the class-diagram/v1 profile: classes, interfaces
// and enums with attributes and operations, and seven kinds of relation, drawn
// by a layered layout as an inert SVG picture. The picture is UML-style;
// CLASS.md lists exactly what is drawn.
package class

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
	Profile = "class-diagram/v1"
)

// Class kinds.
const (
	KindClass     = "class"
	KindInterface = "interface"
	KindEnum      = "enum"
)

// Visibilities.
const (
	Public    = "public"
	Private   = "private"
	Protected = "protected"
	Package   = "package"
)

// Relation kinds.
const (
	Inheritance = "inheritance"
	Realization = "realization"
	Association = "association"
	Directed    = "directed"
	Aggregation = "aggregation"
	Composition = "composition"
	Dependency  = "dependency"
)

// Limits.
const (
	MinClasses   = 1
	MaxClasses   = 40
	MaxMembers   = 20
	MaxParams    = 8
	MaxRelations = 80
	MaxEndLabel  = 24
	maxText      = 120
)

// Model is a class diagram. It carries meaning only: no coordinates, sizes,
// colours or markup.
type Model struct {
	FormatVersion  string     `json:"formatVersion"`
	ProfileVersion string     `json:"profileVersion"`
	ID             string     `json:"id"`
	Title          string     `json:"title"`
	Classes        []Class    `json:"classes"`
	Relations      []Relation `json:"relations"`
}

// Class is a class, interface or enum. For an enum the attributes are the
// literals.
type Class struct {
	ID         string      `json:"id"`
	Name       string      `json:"name"`
	Kind       string      `json:"kind"`
	Attributes []Attribute `json:"attributes"`
	Operations []Operation `json:"operations"`
}

// Attribute is a field. Type may be empty.
type Attribute struct {
	Name       string `json:"name"`
	Type       string `json:"type,omitempty"`
	Visibility string `json:"visibility"`
}

// Operation is a method. Returns may be empty.
type Operation struct {
	Name       string      `json:"name"`
	Parameters []Parameter `json:"parameters,omitempty"`
	Returns    string      `json:"returns,omitempty"`
	Visibility string      `json:"visibility"`
}

// Parameter is one operation parameter.
type Parameter struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// Relation connects two different classes. The labels are optional.
type Relation struct {
	ID        string `json:"id"`
	Kind      string `json:"kind"`
	From      string `json:"from"`
	To        string `json:"to"`
	Label     string `json:"label,omitempty"`
	FromLabel string `json:"fromLabel,omitempty"`
	ToLabel   string `json:"toLabel,omitempty"`
}

// schema is the strict member set; see contract.StrictScan.
var schema = contract.Schema{
	"model": {Fields: map[string]string{
		"formatVersion": "", "profileVersion": "", "id": "", "title": "",
		"classes": "classes", "relations": "relations",
	}},
	"classes":    {Array: true, Elem: "class"},
	"relations":  {Array: true, Elem: "relation"},
	"attributes": {Array: true, Elem: "attribute"},
	"operations": {Array: true, Elem: "operation"},
	"parameters": {Array: true, Elem: "parameter"},
	"class": {Fields: map[string]string{
		"id": "", "name": "", "kind": "", "attributes": "attributes", "operations": "operations",
	}},
	"attribute": {Fields: map[string]string{"name": "", "type": "", "visibility": ""}},
	"operation": {Fields: map[string]string{"name": "", "parameters": "parameters", "returns": "", "visibility": ""}},
	"parameter": {Fields: map[string]string{"name": "", "type": ""}},
	"relation": {Fields: map[string]string{
		"id": "", "kind": "", "from": "", "to": "", "label": "", "fromLabel": "", "toLabel": "",
	}},
}

// Decode scans raw strictly, decodes it, validates it and returns the
// normalized model (classes and relations sorted by id; member and parameter
// order kept; absent optional lists and visibilities filled in).
func Decode(raw []byte) (Model, error) {
	if err := contract.StrictScan(raw, contract.MaxModelBytes, schema); err != nil {
		return Model{}, err
	}
	// Pointers tell an absent optional member from an empty one: an optional
	// text member that is present must not be empty.
	var in struct {
		FormatVersion  string `json:"formatVersion"`
		ProfileVersion string `json:"profileVersion"`
		ID             string `json:"id"`
		Title          string `json:"title"`
		Classes        []struct {
			ID         string `json:"id"`
			Name       string `json:"name"`
			Kind       string `json:"kind"`
			Attributes []struct {
				Name       string  `json:"name"`
				Type       *string `json:"type"`
				Visibility *string `json:"visibility"`
			} `json:"attributes"`
			Operations []struct {
				Name       string      `json:"name"`
				Parameters []Parameter `json:"parameters"`
				Returns    *string     `json:"returns"`
				Visibility *string     `json:"visibility"`
			} `json:"operations"`
		} `json:"classes"`
		Relations []struct {
			ID        string  `json:"id"`
			Kind      string  `json:"kind"`
			From      string  `json:"from"`
			To        string  `json:"to"`
			Label     *string `json:"label"`
			FromLabel *string `json:"fromLabel"`
			ToLabel   *string `json:"toLabel"`
		} `json:"relations"`
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	if err := d.Decode(&in); err != nil {
		return Model{}, err
	}
	m := Model{FormatVersion: in.FormatVersion, ProfileVersion: in.ProfileVersion, ID: in.ID, Title: in.Title}
	opt := func(p *string, what string) (string, error) {
		if p == nil {
			return "", nil
		}
		if *p == "" {
			return "", fmt.Errorf("%s must not be empty when present", what)
		}
		return *p, nil
	}
	vis := func(p *string) string {
		if p == nil {
			return Public
		}
		return *p // an empty or unknown value is refused by Validate
	}
	for _, c := range in.Classes {
		k := Class{ID: c.ID, Name: c.Name, Kind: c.Kind}
		for _, a := range c.Attributes {
			t, err := opt(a.Type, fmt.Sprintf("class %q attribute %q type", c.ID, a.Name))
			if err != nil {
				return Model{}, err
			}
			k.Attributes = append(k.Attributes, Attribute{Name: a.Name, Type: t, Visibility: vis(a.Visibility)})
		}
		for _, o := range c.Operations {
			r, err := opt(o.Returns, fmt.Sprintf("class %q operation %q returns", c.ID, o.Name))
			if err != nil {
				return Model{}, err
			}
			k.Operations = append(k.Operations, Operation{Name: o.Name, Parameters: o.Parameters, Returns: r, Visibility: vis(o.Visibility)})
		}
		m.Classes = append(m.Classes, k)
	}
	for _, r := range in.Relations {
		rel := Relation{ID: r.ID, Kind: r.Kind, From: r.From, To: r.To}
		var err error
		if rel.Label, err = opt(r.Label, fmt.Sprintf("relation %q label", r.ID)); err != nil {
			return Model{}, err
		}
		if rel.FromLabel, err = opt(r.FromLabel, fmt.Sprintf("relation %q fromLabel", r.ID)); err != nil {
			return Model{}, err
		}
		if rel.ToLabel, err = opt(r.ToLabel, fmt.Sprintf("relation %q toLabel", r.ID)); err != nil {
			return Model{}, err
		}
		m.Relations = append(m.Relations, rel)
	}
	if err := m.Validate(); err != nil {
		return Model{}, err
	}
	return m.Normalize(), nil
}

// Normalize returns a copy with classes and relations sorted by id. Member and
// parameter order is significant and kept. Absent lists become empty ones so
// the canonical form always writes them.
func (m Model) Normalize() Model {
	n := m
	n.Classes = make([]Class, len(m.Classes))
	for i, c := range m.Classes {
		c.Attributes = append([]Attribute{}, c.Attributes...)
		c.Operations = make([]Operation, len(c.Operations))
		for j, o := range m.Classes[i].Operations {
			if len(o.Parameters) == 0 {
				o.Parameters = nil
			} else {
				o.Parameters = append([]Parameter(nil), o.Parameters...)
			}
			c.Operations[j] = o
		}
		n.Classes[i] = c
	}
	n.Relations = append([]Relation{}, m.Relations...)
	sort.Slice(n.Classes, func(i, j int) bool { return n.Classes[i].ID < n.Classes[j].ID })
	sort.Slice(n.Relations, func(i, j int) bool { return n.Relations[i].ID < n.Relations[j].ID })
	return n
}

// Canonical is the canonical encoding: compact JSON, fixed member order,
// classes and relations sorted by id, member and parameter order kept, no
// trailing newline, no HTML escaping. Call it on a validated model.
func (m Model) Canonical() ([]byte, error) { return svgx.Compact(m.Normalize()) }

func validVisibility(v string) bool {
	return v == Public || v == Private || v == Protected || v == Package
}

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
	if len(m.Classes) < MinClasses || len(m.Classes) > MaxClasses {
		return fmt.Errorf("classes must contain %d..%d classes", MinClasses, MaxClasses)
	}
	if len(m.Relations) > MaxRelations {
		return fmt.Errorf("relations must contain at most %d relations", MaxRelations)
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
	for _, c := range m.Classes {
		if err := claim("class", c.ID); err != nil {
			return err
		}
		if err := svgx.Plain(c.Name, "class name", maxText); err != nil {
			return err
		}
		switch c.Kind {
		case KindClass, KindInterface, KindEnum:
		default:
			return fmt.Errorf("class %q: unknown kind %q", c.ID, c.Kind)
		}
		if len(c.Attributes) > MaxMembers {
			return fmt.Errorf("class %q: at most %d attributes", c.ID, MaxMembers)
		}
		if len(c.Operations) > MaxMembers {
			return fmt.Errorf("class %q: at most %d operations", c.ID, MaxMembers)
		}
		names := map[string]bool{}
		for _, a := range c.Attributes {
			if err := svgx.Plain(a.Name, "attribute name", maxText); err != nil {
				return fmt.Errorf("class %q: %w", c.ID, err)
			}
			if names[a.Name] {
				return fmt.Errorf("class %q: attribute %q appears twice", c.ID, a.Name)
			}
			names[a.Name] = true
			if a.Type != "" {
				if err := svgx.Plain(a.Type, "attribute type", maxText); err != nil {
					return fmt.Errorf("class %q attribute %q: %w", c.ID, a.Name, err)
				}
			}
			if !validVisibility(a.Visibility) {
				return fmt.Errorf("class %q attribute %q: unknown visibility %q", c.ID, a.Name, a.Visibility)
			}
		}
		names = map[string]bool{}
		for _, o := range c.Operations {
			if err := svgx.Plain(o.Name, "operation name", maxText); err != nil {
				return fmt.Errorf("class %q: %w", c.ID, err)
			}
			if names[o.Name] {
				return fmt.Errorf("class %q: operation %q appears twice", c.ID, o.Name)
			}
			names[o.Name] = true
			if o.Returns != "" {
				if err := svgx.Plain(o.Returns, "operation returns", maxText); err != nil {
					return fmt.Errorf("class %q operation %q: %w", c.ID, o.Name, err)
				}
			}
			if !validVisibility(o.Visibility) {
				return fmt.Errorf("class %q operation %q: unknown visibility %q", c.ID, o.Name, o.Visibility)
			}
			if len(o.Parameters) > MaxParams {
				return fmt.Errorf("class %q operation %q: at most %d parameters", c.ID, o.Name, MaxParams)
			}
			pn := map[string]bool{}
			for _, p := range o.Parameters {
				if err := svgx.Plain(p.Name, "parameter name", maxText); err != nil {
					return fmt.Errorf("class %q operation %q: %w", c.ID, o.Name, err)
				}
				if err := svgx.Plain(p.Type, "parameter type", maxText); err != nil {
					return fmt.Errorf("class %q operation %q parameter %q: %w", c.ID, o.Name, p.Name, err)
				}
				if pn[p.Name] {
					return fmt.Errorf("class %q operation %q: parameter %q appears twice", c.ID, o.Name, p.Name)
				}
				pn[p.Name] = true
			}
		}
		known[c.ID] = true
	}
	type key struct{ a, b, kind string }
	pairs := map[key]bool{}
	parents := map[string][]string{} // child -> parents over inheritance and realization
	for _, r := range m.Relations {
		if err := claim("relation", r.ID); err != nil {
			return err
		}
		switch r.Kind {
		case Inheritance, Realization, Association, Directed, Aggregation, Composition, Dependency:
		default:
			return fmt.Errorf("relation %q: unknown kind %q", r.ID, r.Kind)
		}
		if !known[r.From] || !known[r.To] {
			return fmt.Errorf("relation %q names an unknown class", r.ID)
		}
		if r.From == r.To {
			return fmt.Errorf("relation %q relates class %q to itself", r.ID, r.From)
		}
		for _, l := range []struct{ v, name string }{{r.Label, "label"}, {r.FromLabel, "fromLabel"}, {r.ToLabel, "toLabel"}} {
			if l.v != "" {
				if err := svgx.Plain(l.v, fmt.Sprintf("relation %q %s", r.ID, l.name), MaxEndLabel); err != nil {
					return err
				}
			}
		}
		a, b := r.From, r.To
		if a > b {
			a, b = b, a
		}
		k := key{a, b, r.Kind}
		if pairs[k] {
			return fmt.Errorf("relation %q repeats a %s relation between %q and %q", r.ID, r.Kind, a, b)
		}
		pairs[k] = true
		if r.Kind == Inheritance || r.Kind == Realization {
			parents[r.From] = append(parents[r.From], r.To)
		}
	}
	// Inheritance and realization must form no cycle: the parent is drawn
	// above the child.
	state := map[string]int{} // 1 visiting, 2 done
	var visit func(c string) string
	visit = func(c string) string {
		state[c] = 1
		for _, p := range parents[c] {
			if state[p] == 1 {
				return p
			}
			if state[p] == 0 {
				if bad := visit(p); bad != "" {
					return bad
				}
			}
		}
		state[c] = 2
		return ""
	}
	order := make([]string, 0, len(m.Classes))
	for _, c := range m.Classes {
		order = append(order, c.ID)
	}
	sort.Strings(order)
	for _, c := range order {
		if state[c] == 0 {
			if bad := visit(c); bad != "" {
				return fmt.Errorf("inheritance and realization form a cycle through class %q", bad)
			}
		}
	}
	return nil
}
