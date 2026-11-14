package class

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/spik3r/clanker-foundry/modeling/diagram/contract"
)

var testGen = contract.Generator{ID: "diagramgen", Version: "1.2.2", Toolchain: "go1.27.4"}

var exampleNames = []string{"orders-class", "plugins-class"}

func readExample(t testing.TB, name string) []byte {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("..", "examples", name+".json"))
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func baseModel() Model {
	return Model{
		FormatVersion: Format, ProfileVersion: Profile, ID: "Orders", Title: "Order model",
		Classes: []Class{
			{ID: "Order", Name: "Order", Kind: KindClass,
				Attributes: []Attribute{{"id", "OrderId", Public}},
				Operations: []Operation{{"total", nil, "Money", Public}, {"add", []Parameter{{"a", "A"}, {"b", "B"}}, "", Private}}},
			{ID: "Line", Name: "OrderLine", Kind: KindClass, Attributes: []Attribute{}, Operations: []Operation{}},
			{ID: "Priced", Name: "Priced", Kind: KindInterface, Attributes: []Attribute{}, Operations: []Operation{{"price", nil, "Money", Public}}},
			{ID: "Status", Name: "Status", Kind: KindEnum, Attributes: []Attribute{{"NEW", "", Public}, {"DONE", "", Public}}, Operations: []Operation{}},
		},
		Relations: []Relation{
			{ID: "R2", Kind: Realization, From: "Line", To: "Priced"},
			{ID: "R1", Kind: Composition, From: "Order", To: "Line", FromLabel: "1", ToLabel: "1..*"},
			{ID: "R3", Kind: Directed, From: "Order", To: "Status", Label: "status"},
		},
	}
}

func marshal(t testing.TB, m Model) []byte {
	t.Helper()
	b, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func mustDecode(t testing.TB, raw []byte) Model {
	t.Helper()
	m, err := Decode(raw)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	return m
}

func TestExamplesDecode(t *testing.T) {
	for _, n := range exampleNames {
		mustDecode(t, readExample(t, n))
	}
	mustDecode(t, marshal(t, baseModel()))
}

func TestValidationRules(t *testing.T) {
	long := strings.Repeat("x", 121)
	cases := []struct {
		name string
		mut  func(m *Model)
		want string
	}{
		{"wrong format", func(m *Model) { m.FormatVersion = "diagram-json/v2" }, "unsupported"},
		{"wrong profile", func(m *Model) { m.ProfileVersion = "class-diagram/v2" }, "unsupported"},
		{"bad model id", func(m *Model) { m.ID = "_x" }, "identifier"},
		{"empty title", func(m *Model) { m.Title = "" }, "title"},
		{"blank title", func(m *Model) { m.Title = " \t" }, "blank"},
		{"title with U+2029", func(m *Model) { m.Title = "a b" }, "control"},
		{"title multibyte over 120 bytes", func(m *Model) { m.Title = strings.Repeat("é", 61) }, "title"},
		{"no classes", func(m *Model) { m.Classes = nil; m.Relations = nil }, "classes"},
		{"41 classes", func(m *Model) {
			m.Relations = nil
			m.Classes = nil
			for i := 0; i < 41; i++ {
				m.Classes = append(m.Classes, Class{ID: "C" + strconv.Itoa(i), Name: "c", Kind: KindClass})
			}
		}, "classes"},
		{"bad class id", func(m *Model) { m.Classes[0].ID = "a b" }, "identifier"},
		{"duplicate class id", func(m *Model) { m.Classes[1].ID = "Order" }, "used by"},
		{"class id reused by relation", func(m *Model) { m.Relations[0].ID = "Order" }, "used by"},
		{"empty class name", func(m *Model) { m.Classes[0].Name = "" }, "class name"},
		{"class name too long", func(m *Model) { m.Classes[0].Name = long }, "class name"},
		{"unknown class kind", func(m *Model) { m.Classes[0].Kind = "struct" }, "unknown kind"},
		{"21 attributes", func(m *Model) {
			m.Classes[0].Attributes = nil
			for i := 0; i < 21; i++ {
				m.Classes[0].Attributes = append(m.Classes[0].Attributes, Attribute{"a" + strconv.Itoa(i), "", Public})
			}
		}, "attributes"},
		{"21 operations", func(m *Model) {
			m.Classes[0].Operations = nil
			for i := 0; i < 21; i++ {
				m.Classes[0].Operations = append(m.Classes[0].Operations, Operation{Name: "o" + strconv.Itoa(i), Visibility: Public})
			}
		}, "operations"},
		{"empty attribute name", func(m *Model) { m.Classes[0].Attributes[0].Name = "" }, "attribute name"},
		{"duplicate attribute name", func(m *Model) { m.Classes[0].Attributes = append(m.Classes[0].Attributes, Attribute{"id", "", Public}) }, "twice"},
		{"attribute type too long", func(m *Model) { m.Classes[0].Attributes[0].Type = long }, "attribute type"},
		{"attribute type with newline", func(m *Model) { m.Classes[0].Attributes[0].Type = "a\nb" }, "control"},
		{"unknown attribute visibility", func(m *Model) { m.Classes[0].Attributes[0].Visibility = "friend" }, "visibility"},
		{"empty attribute visibility", func(m *Model) { m.Classes[0].Attributes[0].Visibility = "" }, "visibility"},
		{"duplicate operation name", func(m *Model) { m.Classes[0].Operations[1].Name = "total" }, "twice"},
		{"operation returns too long", func(m *Model) { m.Classes[0].Operations[0].Returns = long }, "returns"},
		{"unknown operation visibility", func(m *Model) { m.Classes[0].Operations[0].Visibility = "x" }, "visibility"},
		{"9 parameters", func(m *Model) {
			m.Classes[0].Operations[0].Parameters = nil
			for i := 0; i < 9; i++ {
				m.Classes[0].Operations[0].Parameters = append(m.Classes[0].Operations[0].Parameters, Parameter{"p" + strconv.Itoa(i), "T"})
			}
		}, "parameters"},
		{"empty parameter name", func(m *Model) { m.Classes[0].Operations[1].Parameters[0].Name = "" }, "parameter name"},
		{"empty parameter type", func(m *Model) { m.Classes[0].Operations[1].Parameters[0].Type = "" }, "parameter type"},
		{"duplicate parameter name", func(m *Model) { m.Classes[0].Operations[1].Parameters[1].Name = "a" }, "twice"},
		{"81 relations", func(m *Model) {
			m.Relations = nil
			for i := 0; i < 81; i++ {
				m.Relations = append(m.Relations, Relation{ID: "X" + strconv.Itoa(i), Kind: Association, From: "Order", To: "Line"})
			}
		}, "at most 80"},
		{"duplicate relation id", func(m *Model) { m.Relations[1].ID = "R2" }, "used by"},
		{"unknown relation kind", func(m *Model) { m.Relations[0].Kind = "friend" }, "unknown kind"},
		{"relation from unknown class", func(m *Model) { m.Relations[0].From = "Nobody" }, "unknown class"},
		{"relation to unknown class", func(m *Model) { m.Relations[0].To = "Nobody" }, "unknown class"},
		{"self relation", func(m *Model) { m.Relations[0].To = "Line" }, "itself"},
		{"label of 25 bytes", func(m *Model) { m.Relations[0].Label = strings.Repeat("x", 25) }, "label"},
		{"fromLabel of 25 bytes", func(m *Model) { m.Relations[1].FromLabel = strings.Repeat("x", 25) }, "fromLabel"},
		{"toLabel multibyte over 24 bytes", func(m *Model) { m.Relations[1].ToLabel = strings.Repeat("é", 13) }, "toLabel"},
		{"label with control character", func(m *Model) { m.Relations[0].Label = "a\x01" }, "control"},
		{"blank label", func(m *Model) { m.Relations[0].Label = "  " }, "blank"},
		{"same pair and kind", func(m *Model) {
			m.Relations = append(m.Relations, Relation{ID: "R4", Kind: Composition, From: "Order", To: "Line"})
		}, "repeats"},
		{"same unordered pair and kind", func(m *Model) {
			m.Relations = append(m.Relations, Relation{ID: "R4", Kind: Composition, From: "Line", To: "Order"})
		}, "repeats"},
		{"inheritance cycle of two", func(m *Model) {
			m.Relations = append(m.Relations,
				Relation{ID: "I1", Kind: Inheritance, From: "Order", To: "Line"},
				Relation{ID: "I2", Kind: Realization, From: "Line", To: "Order"})
		}, "cycle"},
		{"inheritance cycle of three", func(m *Model) {
			m.Relations = []Relation{
				{ID: "I1", Kind: Inheritance, From: "Order", To: "Line"},
				{ID: "I2", Kind: Inheritance, From: "Line", To: "Priced"},
				{ID: "I3", Kind: Realization, From: "Priced", To: "Order"},
			}
		}, "cycle"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			m := baseModel()
			c.mut(&m)
			err := m.Validate()
			if err == nil {
				t.Fatalf("model accepted")
			}
			if !strings.Contains(err.Error(), c.want) {
				t.Errorf("error %q does not mention %q", err, c.want)
			}
		})
	}
	if err := baseModel().Validate(); err != nil {
		t.Fatalf("base model: %v", err)
	}
	// Different kinds between the same pair, and a cycle among other kinds,
	// are allowed.
	m := baseModel()
	m.Relations = append(m.Relations,
		Relation{ID: "R4", Kind: Association, From: "Line", To: "Order"},
		Relation{ID: "R5", Kind: Dependency, From: "Priced", To: "Line"},
		Relation{ID: "R6", Kind: Aggregation, From: "Priced", To: "Order"},
		Relation{ID: "R7", Kind: Association, From: "Status", To: "Order"})
	if err := m.Validate(); err != nil {
		t.Errorf("valid model with a non-inheritance cycle rejected: %v", err)
	}
}

func TestLimitsAccepted(t *testing.T) {
	m := Model{FormatVersion: Format, ProfileVersion: Profile, ID: "Big", Title: strings.Repeat("é", 60)}
	for i := 0; i < MaxClasses; i++ {
		c := Class{ID: "C" + strconv.Itoa(i), Name: strings.Repeat("é", 60), Kind: KindClass}
		for j := 0; j < MaxMembers; j++ {
			c.Attributes = append(c.Attributes, Attribute{"a" + strconv.Itoa(j), strings.Repeat("t", 120), Private})
			var ps []Parameter
			for k := 0; k < MaxParams; k++ {
				ps = append(ps, Parameter{"p" + strconv.Itoa(k), "T"})
			}
			c.Operations = append(c.Operations, Operation{"o" + strconv.Itoa(j), ps, strings.Repeat("r", 120), Package})
		}
		m.Classes = append(m.Classes, c)
	}
	for i := 0; i < MaxRelations; i++ {
		m.Relations = append(m.Relations, Relation{ID: "R" + strconv.Itoa(i), Kind: Association, From: "C" + strconv.Itoa(i%MaxClasses), To: "C" + strconv.Itoa((i%MaxClasses+1+i/MaxClasses)%MaxClasses), Label: strings.Repeat("l", 24)})
	}
	if err := m.Validate(); err != nil {
		t.Fatalf("model at every limit rejected: %v", err)
	}
	m = mustDecode(t, marshal(t, m))
	checkGeometry(t, m, SVG(m, LayoutOf(m), testGen))
}

func TestStrictDecoding(t *testing.T) {
	good := string(marshal(t, baseModel()))
	wrap := func(classes, rels string) string {
		return `{"formatVersion":"diagram-json/v1","profileVersion":"class-diagram/v1","id":"A","title":"T","classes":[` + classes + `],"relations":[` + rels + `]}`
	}
	cls := `{"id":"A","name":"A","kind":"class","attributes":[{"name":"x","type":"int","visibility":"public"}],"operations":[{"name":"f","parameters":[{"name":"p","type":"T"}],"returns":"R","visibility":"private"}]}`
	cls2 := `{"id":"B","name":"B","kind":"class"}`
	rel := `{"id":"R","kind":"association","from":"A","to":"B","label":"x"}`
	if _, err := Decode([]byte(wrap(cls+","+cls2, rel))); err != nil {
		t.Fatalf("wrap: %v", err)
	}
	bad := map[string]string{
		"empty":                        ``,
		"not an object":                `[]`,
		"trailing data":                good + ` 1`,
		"null":                         strings.Replace(good, `"Orders"`, `null`, 1),
		"duplicate key":                strings.Replace(good, `"title":"Order model"`, `"title":"a","title":"b"`, 1),
		"unknown member":               strings.Replace(good, `"title"`, `"x":1,"title"`, 1),
		"case variant of member":       strings.Replace(good, `"classes"`, `"Classes"`, 1),
		"unknown class member":         wrap(strings.Replace(cls, `"kind"`, `"colour":"red","kind"`, 1)+","+cls2, rel),
		"unknown attribute member":     wrap(strings.Replace(cls, `"type"`, `"static":true,"type"`, 1)+","+cls2, rel),
		"unknown parameter member":     wrap(strings.Replace(cls, `"type":"T"`, `"type":"T","default":"1"`, 1)+","+cls2, rel),
		"object for a scalar":          wrap(strings.Replace(cls, `"name":"A"`, `"name":{}`, 1)+","+cls2, rel),
		"array for a scalar":           wrap(cls+","+cls2, strings.Replace(rel, `"x"`, `["x"]`, 1)),
		"number for a string":          wrap(cls+","+cls2, strings.Replace(rel, `"x"`, `5`, 1)),
		"empty optional type":          wrap(strings.Replace(cls, `"type":"int"`, `"type":""`, 1)+","+cls2, rel),
		"empty optional returns":       wrap(strings.Replace(cls, `"returns":"R"`, `"returns":""`, 1)+","+cls2, rel),
		"empty optional label":         wrap(cls+","+cls2, strings.Replace(rel, `"x"`, `""`, 1)),
		"empty optional fromLabel":     wrap(cls+","+cls2, strings.Replace(rel, `"label"`, `"fromLabel":"","label"`, 1)),
		"empty visibility":             wrap(strings.Replace(cls, `"public"`, `""`, 1)+","+cls2, rel),
		"missing parameter type":       wrap(strings.Replace(cls, `,"type":"T"`, ``, 1)+","+cls2, rel),
		"invalid UTF-8":                strings.Replace(good, `Order model`, "Order \xffmodel", 1),
		"nesting too deep":             strings.Repeat(`[`, 20) + strings.Repeat(`]`, 20),
		"model too large":              good + strings.Repeat(" ", contract.MaxModelBytes),
		"attributes as an object":      wrap(strings.Replace(cls, `"attributes":[`, `"attributes":{"a":[`, 1)+","+cls2, rel),
		"relations member as a scalar": strings.Replace(wrap(cls+","+cls2, rel), `"relations":[`+rel+`]`, `"relations":7`, 1),
	}
	for name, raw := range bad {
		if _, err := Decode([]byte(raw)); err == nil {
			t.Errorf("%s: accepted", name)
		}
	}
	// Absent lists and visibility are filled in; the canonical form is the
	// same as the spelled-out model.
	a := mustDecode(t, []byte(wrap(`{"id":"A","name":"A","kind":"class"}`, ``)))
	b := mustDecode(t, []byte(wrap(`{"id":"A","name":"A","kind":"class","attributes":[],"operations":[]}`, ``)))
	c := mustDecode(t, []byte(strings.Replace(wrap(`{"id":"A","name":"A","kind":"class","attributes":[{"name":"x"}],"operations":[]}`, ``), `,"relations":[]`, ``, 1)))
	ca, _ := a.Canonical()
	cb, _ := b.Canonical()
	if string(ca) != string(cb) {
		t.Errorf("absent and empty lists differ: %s vs %s", ca, cb)
	}
	cc, _ := c.Canonical()
	if !strings.Contains(string(cc), `{"name":"x","visibility":"public"}`) {
		t.Errorf("default visibility not written: %s", cc)
	}
}

func TestCanonicalEncoding(t *testing.T) {
	m := mustDecode(t, marshal(t, baseModel()))
	c, err := m.Canonical()
	if err != nil {
		t.Fatal(err)
	}
	want := `{"formatVersion":"diagram-json/v1","profileVersion":"class-diagram/v1","id":"Orders","title":"Order model","classes":[` +
		`{"id":"Line","name":"OrderLine","kind":"class","attributes":[],"operations":[]},` +
		`{"id":"Order","name":"Order","kind":"class","attributes":[{"name":"id","type":"OrderId","visibility":"public"}],` +
		`"operations":[{"name":"total","returns":"Money","visibility":"public"},{"name":"add","parameters":[{"name":"a","type":"A"},{"name":"b","type":"B"}],"visibility":"private"}]},` +
		`{"id":"Priced","name":"Priced","kind":"interface","attributes":[],"operations":[{"name":"price","returns":"Money","visibility":"public"}]},` +
		`{"id":"Status","name":"Status","kind":"enum","attributes":[{"name":"NEW","visibility":"public"},{"name":"DONE","visibility":"public"}],"operations":[]}],` +
		`"relations":[{"id":"R1","kind":"composition","from":"Order","to":"Line","fromLabel":"1","toLabel":"1..*"},` +
		`{"id":"R2","kind":"realization","from":"Line","to":"Priced"},` +
		`{"id":"R3","kind":"directed","from":"Order","to":"Status","label":"status"}]}`
	if string(c) != want {
		t.Errorf("canonical encoding\n got %s\nwant %s", c, want)
	}
	m.Title = "a<b>&c"
	c, _ = m.Canonical()
	if !strings.Contains(string(c), `"title":"a<b>&c"`) {
		t.Errorf("canonical encoding escaped HTML characters: %s", c)
	}
}

func digest(t *testing.T, m Model) string {
	t.Helper()
	c, err := m.Canonical()
	if err != nil {
		t.Fatal(err)
	}
	return contract.HexSum(c)
}

func TestDigestInvariance(t *testing.T) {
	want := digest(t, baseModel())
	// Class and relation order in the source do not matter.
	m := baseModel()
	m.Classes[0], m.Classes[3] = m.Classes[3], m.Classes[0]
	m.Relations[0], m.Relations[2] = m.Relations[2], m.Relations[0]
	if digest(t, m) != want {
		t.Errorf("class or relation order changed the digest")
	}
	// Source formatting and member order do not matter either.
	var v any
	if err := json.Unmarshal(marshal(t, baseModel()), &v); err != nil {
		t.Fatal(err)
	}
	reformatted, _ := json.MarshalIndent(v, "", "  ")
	if digest(t, mustDecode(t, reformatted)) != want {
		t.Errorf("source formatting changed the digest")
	}
}

func TestDigestSensitivity(t *testing.T) {
	want := digest(t, baseModel())
	muts := map[string]func(m *Model){
		"attribute order": func(m *Model) {
			m.Classes[3].Attributes[0], m.Classes[3].Attributes[1] = m.Classes[3].Attributes[1], m.Classes[3].Attributes[0]
		},
		"parameter order": func(m *Model) {
			p := m.Classes[0].Operations[1].Parameters
			p[0], p[1] = p[1], p[0]
		},
		"operation order": func(m *Model) {
			o := m.Classes[0].Operations
			o[0], o[1] = o[1], o[0]
		},
		"visibility":    func(m *Model) { m.Classes[0].Attributes[0].Visibility = Private },
		"attr type":     func(m *Model) { m.Classes[0].Attributes[0].Type = "Other" },
		"class kind":    func(m *Model) { m.Classes[1].Kind = KindInterface },
		"relation kind": func(m *Model) { m.Relations[0].Kind = Aggregation },
		"direction":     func(m *Model) { m.Relations[1].From, m.Relations[1].To = "Line", "Order" },
		"toLabel":       func(m *Model) { m.Relations[1].ToLabel = "2" },
		"label":         func(m *Model) { m.Relations[2].Label = "state" },
		"title":         func(m *Model) { m.Title = "Other" },
		"drop relation": func(m *Model) { m.Relations = m.Relations[:2] },
		"returns":       func(m *Model) { m.Classes[0].Operations[1].Returns = "X" },
	}
	seen := map[string]string{}
	for name, mut := range muts {
		m := baseModel()
		mut(&m)
		if err := m.Validate(); err != nil {
			t.Fatalf("%s: mutated model invalid: %v", name, err)
		}
		d := digest(t, m)
		if d == want {
			t.Errorf("%s did not change the digest", name)
		}
		if prev, dup := seen[d]; dup {
			t.Errorf("%s and %s share a digest", name, prev)
		}
		seen[d] = name
	}
}
