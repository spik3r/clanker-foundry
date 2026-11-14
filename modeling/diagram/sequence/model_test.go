package sequence

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/spik3r/clanker-foundry/modeling/diagram/contract"
)

var testGen = contract.Generator{ID: "diagramgen", Version: "1.2.1", Toolchain: "go1.27.4"}

var exampleNames = []string{"checkout-sequence", "retry-sequence"}

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
		FormatVersion: Format, ProfileVersion: Profile, ID: "Checkout", Title: "Checkout",
		Participants: []Participant{
			{"User", "Customer", KindActor},
			{"Web", "Web app", KindParticipant},
			{"Pay", "Payment service", KindParticipant},
		},
		Steps: []Step{
			{ID: "S1", Kind: StepMessage, From: "User", To: "Web", Label: "Place order", Style: StyleSync},
			{ID: "S2", Kind: StepMessage, From: "Web", To: "Pay", Label: "Charge card", Style: StyleSync},
			{ID: "S3", Kind: StepMessage, From: "Pay", To: "Web", Label: "Charged", Style: StyleReply},
			{ID: "S4", Kind: StepNote, Over: []string{"Web"}, Label: "Order is now confirmed"},
			{ID: "S5", Kind: StepMessage, From: "Web", To: "Web", Label: "Record order", Style: StyleSync},
			{ID: "S6", Kind: StepMessage, From: "Web", To: "User", Label: "Show receipt", Style: StyleAsync},
		},
		Groups: []Group{{"G2", GroupOpt, "second", "S5", "S6"}, {"G1", GroupAlt, "first", "S2", "S3"}},
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

// TestValidationRules mutates a valid model one rule at a time.
func TestValidationRules(t *testing.T) {
	long := strings.Repeat("x", 121)
	manyP := func(n int) []Participant {
		var ps []Participant
		for i := 0; i < n; i++ {
			ps = append(ps, Participant{"P" + string(rune('A'+i)), "p", KindParticipant})
		}
		return ps
	}
	cases := []struct {
		name string
		mut  func(m *Model)
		want string
	}{
		{"wrong format", func(m *Model) { m.FormatVersion = "diagram-json/v2" }, "unsupported"},
		{"wrong profile", func(m *Model) { m.ProfileVersion = "sequence-diagram/v2" }, "unsupported"},
		{"bad model id", func(m *Model) { m.ID = "1Bad" }, "identifier"},
		{"model id too long", func(m *Model) { m.ID = "A" + strings.Repeat("b", 48) }, "identifier"},
		{"empty title", func(m *Model) { m.Title = "" }, "title"},
		{"blank title", func(m *Model) { m.Title = "   " }, "blank"},
		{"title too long", func(m *Model) { m.Title = long }, "title"},
		{"title multibyte over 120 bytes", func(m *Model) { m.Title = strings.Repeat("é", 61) }, "title"},
		{"title with newline", func(m *Model) { m.Title = "a\nb" }, "control"},
		{"title with U+2028", func(m *Model) { m.Title = "a b" }, "control"},
		{"title with U+FFFF", func(m *Model) { m.Title = "a￿b" }, "control"},
		{"title invalid UTF-8", func(m *Model) { m.Title = "a\xffb" }, "title"},
		{"one participant", func(m *Model) { m.Participants = m.Participants[:1] }, "participants"},
		{"13 participants", func(m *Model) { m.Participants = manyP(13) }, "participants"},
		{"bad participant id", func(m *Model) { m.Participants[0].ID = "a-b" }, "identifier"},
		{"duplicate participant id", func(m *Model) { m.Participants[1].ID = "User" }, "used by"},
		{"participant id reused by a step", func(m *Model) { m.Steps[0].ID = "Web" }, "used by"},
		{"participant id reused by a group", func(m *Model) { m.Groups[0].ID = "Web" }, "used by"},
		{"empty participant label", func(m *Model) { m.Participants[0].Label = "" }, "label"},
		{"unknown participant kind", func(m *Model) { m.Participants[0].Kind = "database" }, "kind"},
		{"no steps", func(m *Model) { m.Steps = nil; m.Groups = nil }, "steps"},
		{"81 steps", func(m *Model) {
			m.Groups = nil
			m.Steps = nil
			for i := 0; i < 81; i++ {
				m.Steps = append(m.Steps, Step{ID: "S" + strings.Repeat("a", 1) + itoa(i), Kind: StepMessage, From: "User", To: "Web", Label: "m", Style: StyleSync})
			}
		}, "steps"},
		{"duplicate step id", func(m *Model) { m.Steps[1].ID = "S1" }, "used by"},
		{"bad step id", func(m *Model) { m.Steps[1].ID = "S 2" }, "identifier"},
		{"empty step label", func(m *Model) { m.Steps[0].Label = "" }, "label"},
		{"step label too long", func(m *Model) { m.Steps[0].Label = long }, "label"},
		{"unknown step kind", func(m *Model) { m.Steps[0].Kind = "fork" }, "unknown kind"},
		{"message from unknown participant", func(m *Model) { m.Steps[0].From = "Nobody" }, "unknown participant"},
		{"message to unknown participant", func(m *Model) { m.Steps[0].To = "Nobody" }, "unknown participant"},
		{"message with unknown style", func(m *Model) { m.Steps[0].Style = "loud" }, "style"},
		{"message with over", func(m *Model) { m.Steps[0].Over = []string{"Web"} }, "over"},
		{"note without over", func(m *Model) { m.Steps[3].Over = nil }, "1 or 2"},
		{"note over three", func(m *Model) { m.Steps[3].Over = []string{"User", "Web", "Pay"} }, "1 or 2"},
		{"note over the same participant twice", func(m *Model) { m.Steps[3].Over = []string{"Web", "Web"} }, "twice"},
		{"note over unknown participant", func(m *Model) { m.Steps[3].Over = []string{"Nobody"} }, "unknown participant"},
		{"note with from", func(m *Model) { m.Steps[3].From = "Web" }, "from, to or style"},
		{"note with style", func(m *Model) { m.Steps[3].Style = StyleSync }, "from, to or style"},
		{"9 groups", func(m *Model) {
			m.Groups = nil
			for i := 0; i < 9; i++ {
				m.Groups = append(m.Groups, Group{"G" + itoa(i), GroupOpt, "g", "S1", "S1"})
			}
		}, "at most"},
		{"unknown group kind", func(m *Model) { m.Groups[0].Kind = "par" }, "kind"},
		{"empty group label", func(m *Model) { m.Groups[0].Label = "" }, "label"},
		{"group with unknown first step", func(m *Model) { m.Groups[0].First = "Nope" }, "unknown step"},
		{"group with unknown last step", func(m *Model) { m.Groups[0].Last = "Nope" }, "unknown step"},
		{"group naming a participant as a step", func(m *Model) { m.Groups[0].First = "Web" }, "unknown step"},
		{"group with first after last", func(m *Model) { m.Groups[0].First, m.Groups[0].Last = "S6", "S5" }, "after last"},
		{"overlapping groups", func(m *Model) { m.Groups[1].Last = "S5" }, "overlap"},
		{"nested groups", func(m *Model) { m.Groups[0].First, m.Groups[0].Last = "S1", "S6" }, "overlap"},
		{"groups sharing a boundary step", func(m *Model) { m.Groups[1].Last = "S5" }, "overlap"},
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
}

func itoa(n int) string { return strconv.Itoa(n) }

func TestLimitsAccepted(t *testing.T) {
	m := baseModel()
	m.Groups = nil
	m.Participants = nil
	for i := 0; i < MaxParticipants; i++ {
		m.Participants = append(m.Participants, Participant{"P" + itoa(i), strings.Repeat("é", 60), KindParticipant})
	}
	m.Steps = nil
	for i := 0; i < MaxSteps; i++ {
		m.Steps = append(m.Steps, Step{ID: "M" + itoa(i), Kind: StepMessage, From: "P0", To: "P" + itoa(i%MaxParticipants), Label: strings.Repeat("x", 120), Style: StyleAsync})
	}
	for i := 0; i < MaxGroups; i++ {
		m.Groups = append(m.Groups, Group{"G" + itoa(i), GroupLoop, "g", "M" + itoa(i*2), "M" + itoa(i*2+1)})
	}
	if err := m.Validate(); err != nil {
		t.Fatalf("model at every limit rejected: %v", err)
	}
	mustDecode(t, marshal(t, m))
}

// TestStrictDecoding checks the raw JSON rules shared with the other profiles
// and the profile's own presence rules.
func TestStrictDecoding(t *testing.T) {
	good := string(marshal(t, baseModel()))
	step := `{"id":"S1","kind":"message","from":"User","to":"Web","label":"x","style":"sync"}`
	wrap := func(steps string) string {
		return `{"formatVersion":"diagram-json/v1","profileVersion":"sequence-diagram/v1","id":"A","title":"T",` +
			`"participants":[{"id":"User","label":"U","kind":"actor"},{"id":"Web","label":"W","kind":"participant"}],"steps":[` + steps + `]}`
	}
	if _, err := Decode([]byte(wrap(step))); err != nil {
		t.Fatalf("wrap(step): %v", err)
	}
	bad := map[string]string{
		"empty":                   ``,
		"not an object":           `[]`,
		"trailing data":           good + `{}`,
		"null":                    strings.Replace(good, `"Checkout"`, `null`, 1),
		"duplicate key":           strings.Replace(good, `"title":"Checkout"`, `"title":"Checkout","title":"Again"`, 1),
		"unknown member":          strings.Replace(good, `"title"`, `"colour":"red","title"`, 1),
		"case variant of member":  strings.Replace(good, `"title"`, `"Title"`, 1),
		"unknown step member":     wrap(strings.Replace(step, `"label"`, `"width":3,"label"`, 1)),
		"object for a scalar":     wrap(strings.Replace(step, `"x"`, `{}`, 1)),
		"array for a scalar":      wrap(strings.Replace(step, `"x"`, `[]`, 1)),
		"number for a string":     wrap(strings.Replace(step, `"x"`, `7`, 1)),
		"message without label":   wrap(strings.Replace(step, `"label":"x",`, ``, 1)),
		"message without style":   wrap(strings.Replace(step, `,"style":"sync"`, ``, 1)),
		"message without from":    wrap(strings.Replace(step, `"from":"User",`, ``, 1)),
		"message with empty over": wrap(strings.Replace(step, `"label"`, `"over":[],"label"`, 1)),
		"note without over":       wrap(`{"id":"N","kind":"note","label":"x"}`),
		"note with empty from":    wrap(`{"id":"N","kind":"note","over":["Web"],"from":"","label":"x"}`),
		"note with style":         wrap(`{"id":"N","kind":"note","over":["Web"],"style":"sync","label":"x"}`),
		"over holding an object":  wrap(`{"id":"N","kind":"note","over":[{}],"label":"x"}`),
		"invalid UTF-8":           strings.Replace(good, `Checkout"`, "Check\xffout\"", 1),
		"nesting too deep":        strings.Repeat(`{"a":`, 20) + `1` + strings.Repeat(`}`, 20),
		"model too large":         good + strings.Repeat(" ", contract.MaxModelBytes),
	}
	for name, raw := range bad {
		if _, err := Decode([]byte(raw)); err == nil {
			t.Errorf("%s: accepted", name)
		}
	}
	// Absent and empty groups are the same model.
	a := mustDecode(t, []byte(wrap(step)))
	b := mustDecode(t, []byte(strings.Replace(wrap(step), `"steps"`, `"groups":[],"steps"`, 1)))
	ca, _ := a.Canonical()
	cb, _ := b.Canonical()
	if string(ca) != string(cb) {
		t.Errorf("absent and empty groups differ: %s vs %s", ca, cb)
	}
}

func TestCanonicalEncoding(t *testing.T) {
	m := mustDecode(t, marshal(t, baseModel()))
	c, err := m.Canonical()
	if err != nil {
		t.Fatal(err)
	}
	want := `{"formatVersion":"diagram-json/v1","profileVersion":"sequence-diagram/v1","id":"Checkout","title":"Checkout",` +
		`"participants":[{"id":"User","label":"Customer","kind":"actor"},{"id":"Web","label":"Web app","kind":"participant"},{"id":"Pay","label":"Payment service","kind":"participant"}],` +
		`"steps":[{"id":"S1","kind":"message","from":"User","to":"Web","label":"Place order","style":"sync"},` +
		`{"id":"S2","kind":"message","from":"Web","to":"Pay","label":"Charge card","style":"sync"},` +
		`{"id":"S3","kind":"message","from":"Pay","to":"Web","label":"Charged","style":"reply"},` +
		`{"id":"S4","kind":"note","over":["Web"],"label":"Order is now confirmed"},` +
		`{"id":"S5","kind":"message","from":"Web","to":"Web","label":"Record order","style":"sync"},` +
		`{"id":"S6","kind":"message","from":"Web","to":"User","label":"Show receipt","style":"async"}],` +
		`"groups":[{"id":"G1","kind":"alt","label":"first","first":"S2","last":"S3"},{"id":"G2","kind":"opt","label":"second","first":"S5","last":"S6"}]}`
	if string(c) != want {
		t.Errorf("canonical encoding\n got %s\nwant %s", c, want)
	}
	// Characters that HTML escaping would rewrite stay as written.
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
	m := baseModel()
	want := digest(t, m)

	// Group order in the source does not matter.
	g := baseModel()
	g.Groups[0], g.Groups[1] = g.Groups[1], g.Groups[0]
	if digest(t, g) != want {
		t.Errorf("group order changed the digest")
	}
	// Neither does source formatting or member order.
	var v any
	if err := json.Unmarshal(marshal(t, m), &v); err != nil {
		t.Fatal(err)
	}
	reordered, _ := json.MarshalIndent(v, "", "\t") // sorts members
	if digest(t, mustDecode(t, reordered)) != want {
		t.Errorf("source formatting changed the digest")
	}
}

func TestDigestSensitivity(t *testing.T) {
	want := digest(t, baseModel())
	muts := map[string]func(m *Model){
		"participant order": func(m *Model) { m.Participants[1], m.Participants[2] = m.Participants[2], m.Participants[1] },
		"step order":        func(m *Model) { m.Steps[0], m.Steps[1] = m.Steps[1], m.Steps[0] },
		"over order":        func(m *Model) { m.Steps[3].Over = []string{"Web", "Pay"} },
		"label":             func(m *Model) { m.Steps[0].Label += "!" },
		"style":             func(m *Model) { m.Steps[0].Style = StyleAsync },
		"participant kind":  func(m *Model) { m.Participants[1].Kind = KindActor },
		"group range":       func(m *Model) { m.Groups[1].First = "S1" },
		"group kind":        func(m *Model) { m.Groups[1].Kind = GroupLoop },
		"drop groups":       func(m *Model) { m.Groups = nil },
		"title":             func(m *Model) { m.Title = "Other" },
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
	over := baseModel()
	over.Steps[3].Over = []string{"Web", "Pay"}
	over2 := baseModel()
	over2.Steps[3].Over = []string{"Pay", "Web"}
	if digest(t, over) == digest(t, over2) {
		t.Errorf("order inside over should be significant in the canonical form")
	}
}
