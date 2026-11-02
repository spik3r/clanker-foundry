package views

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/model"
	"io"
	"reflect"
	"strings"
	"testing"
)

func fixture() model.Model {
	return model.Model{Version: 1, ID: "Example", Initial: "Draft", States: []model.State{{ID: "Draft", Label: "Draft"}, {ID: "Done", Label: "Done", Terminal: true}}, Requirements: []model.Requirement{{ID: "R_actor", Text: "Invented named actor requirement"}, {ID: "R_outcome", Text: "Invented outcome requirement"}}, Guards: []model.Guard{
		{ID: "G_actor", Requirement: "R_actor", Expr: model.Expr{Kind: "op", Op: "nonempty", Args: []model.Expr{{Kind: "ref", Ref: "command.actor"}}}},
		{ID: "G_outcome", Requirement: "R_outcome", Expr: model.Expr{Kind: "op", Op: "nonempty", Args: []model.Expr{{Kind: "ref", Ref: "task.outcome"}}}},
	}, Transitions: []model.Transition{{ID: "T_finish", From: "Draft", To: "Done", Trigger: "finish", Guards: []string{"G_outcome", "G_actor"}, Effects: []model.Effect{{Op: model.SetState}, {Op: model.IncrementRevision}}}}}
}
func identity() Identity {
	return Identity{strings.Repeat("a", 64), "json/v1", model.ProfileVersion, "test/v1", "go1.27.3"}
}
func render(t *testing.T, m model.Model) Artifacts {
	t.Helper()
	a, e := Render(m, identity())
	if e != nil {
		t.Fatal(e)
	}
	return a
}
func trace(t *testing.T, m model.Model, a Artifacts) []byte {
	t.Helper()
	b, e := Trace(m, identity(), a, CodeArtifacts{[]byte("code\n"), []byte("tests\n")})
	if e != nil {
		t.Fatal(e)
	}
	return b
}
func TestRepeatabilitySetsAndNonMutation(t *testing.T) {
	m := fixture()
	before, _ := json.Marshal(m)
	a := render(t, m)
	tr := trace(t, m, a)
	for i := 0; i < 5; i++ {
		b := render(t, m)
		if !reflect.DeepEqual(a, b) || !bytes.Equal(tr, trace(t, m, b)) {
			t.Fatal("nonrepeatable")
		}
	}
	after, _ := json.Marshal(m)
	if !bytes.Equal(before, after) {
		t.Fatal("mutated input")
	}
	m.States[0], m.States[1] = m.States[1], m.States[0]
	m.Requirements[0], m.Requirements[1] = m.Requirements[1], m.Requirements[0]
	m.Guards[0], m.Guards[1] = m.Guards[1], m.Guards[0]
	m.Transitions[0].Guards[0], m.Transitions[0].Guards[1] = m.Transitions[0].Guards[1], m.Transitions[0].Guards[0]
	b := render(t, m)
	if !reflect.DeepEqual(a, b) || !bytes.Equal(tr, trace(t, m, b)) {
		t.Fatal("set order affected bytes")
	}
}
func TestSemanticCompletenessAndChanges(t *testing.T) {
	m := fixture()
	a := render(t, m)
	var table struct {
		Model            model.Model
		GuardCombination string
	}
	if e := json.Unmarshal(a.DecisionTable, &table); e != nil {
		t.Fatal(e)
	}
	n, _ := model.Normalize(m)
	if !reflect.DeepEqual(n, table.Model) || !strings.Contains(table.GuardCombination, "all") {
		t.Fatal("table loses semantics")
	}
	for _, s := range []string{"T_finish", "G_actor", "G_outcome", "R_actor", "command.actor", "set_state(Done)", "increment_revision", "initial: Draft", "terminal=true"} {
		if !bytes.Contains(a.SVG, []byte(s)) {
			t.Fatalf("diagram omits %s", s)
		}
	}
	m.Guards[0].Expr.Args[0].Ref = "task.owner"
	b := render(t, m)
	if bytes.Equal(a.SVG, b.SVG) || bytes.Equal(a.DecisionTable, b.DecisionTable) {
		t.Fatal("guard change hidden")
	}
	m = fixture()
	m.Transitions[0].Effects = append([]model.Effect{{Op: model.AssignOwner}}, m.Transitions[0].Effects...)
	b = render(t, m)
	if bytes.Equal(a.SVG, b.SVG) || bytes.Equal(a.DecisionTable, b.DecisionTable) {
		t.Fatal("effect change hidden")
	}
	if bytes.Index(b.SVG, []byte("assign_owner_if_absent")) >= bytes.Index(b.SVG, []byte("set_state(Done)")) {
		t.Fatal("effects reordered")
	}
	m.Transitions[0].Effects[0], m.Transitions[0].Effects[1] = m.Transitions[0].Effects[1], m.Transitions[0].Effects[0]
	if _, e := Render(m, identity()); e == nil {
		t.Fatal("invalid effect order accepted")
	}
}
func TestBooleanSetOrder(t *testing.T) {
	m := fixture()
	x, y := m.Guards[0].Expr, m.Guards[1].Expr
	m.Guards[0].Expr = model.Expr{Kind: "op", Op: "any", Args: []model.Expr{x, y}}
	a := render(t, m)
	m.Guards[0].Expr.Args = []model.Expr{y, x}
	if !reflect.DeepEqual(a, render(t, m)) {
		t.Fatal("boolean set order")
	}
}
func TestHostileEscaping(t *testing.T) {
	hostile := "</text><script>alert(1)</script> & \" | `\n\x00\t雪"
	m := fixture()
	m.States[0].Label = hostile
	m.Requirements[0].Text = hostile
	a := render(t, m)
	d := xml.NewDecoder(bytes.NewReader(a.SVG))
	var txt strings.Builder
	for {
		tok, e := d.Token()
		if e == io.EOF {
			break
		}
		if e != nil {
			t.Fatal(e)
		}
		switch x := tok.(type) {
		case xml.StartElement:
			switch x.Name.Local {
			case "svg", "title", "rect", "path", "text":
			default:
				t.Fatal("unexpected element", x.Name.Local)
			}
			for _, at := range x.Attr {
				if strings.HasPrefix(at.Name.Local, "on") || at.Name.Local == "href" {
					t.Fatal("active attribute")
				}
			}
		case xml.CharData:
			txt.Write(x)
		}
	}
	if !strings.Contains(txt.String(), "<script>") || !strings.Contains(txt.String(), "\\x00") {
		t.Fatal("hostile text lost")
	}
	var table struct{ Model model.Model }
	if e := json.Unmarshal(a.DecisionTable, &table); e != nil {
		t.Fatal(e)
	}
	found := false
	for _, s := range table.Model.States {
		if s.ID == "Draft" {
			found = s.Label == hostile
		}
	}
	if !found {
		t.Fatal("table changed text")
	}
}
func TestFailClosed(t *testing.T) {
	mutations := map[string]func(*model.Model){
		"duplicate": func(m *model.Model) { m.States[0].ID = "G_actor" }, "effect": func(m *model.Model) { m.Transitions[0].Effects[0].Op = "shell" }, "operator": func(m *model.Model) { m.Guards[0].Expr.Op = "execute" }, "reference": func(m *model.Model) { m.Guards[0].Expr.Args[0].Ref = "private.secret" }, "kind": func(m *model.Model) { m.Guards[0].Expr.Kind = "script" }, "dangling": func(m *model.Model) { m.Transitions[0].Guards[0] = "Missing" }, "terminal": func(m *model.Model) { m.States[0].Terminal = true }, "ambiguous": func(m *model.Model) {
			x := m.Transitions[0]
			x.ID = "T_other"
			m.Transitions = append(m.Transitions, x)
		},
	}
	for name, f := range mutations {
		t.Run(name, func(t *testing.T) {
			m := fixture()
			f(&m)
			a, e := Render(m, identity())
			if e == nil || len(a.SVG) > 0 || len(a.DecisionTable) > 0 {
				t.Fatal("invalid model emitted")
			}
			if b, e := Trace(m, identity(), Artifacts{}, CodeArtifacts{}); e == nil || len(b) > 0 {
				t.Fatal("invalid trace emitted")
			}
		})
	}
	id := identity()
	id.ProfileVersion = "unknown"
	if _, e := Render(fixture(), id); e == nil {
		t.Fatal("unknown profile")
	}
	id = identity()
	id.ModelDigest = "wrong"
	if _, e := Render(fixture(), id); e == nil {
		t.Fatal("invalid digest")
	}
}
func TestTraceDigestsAndMappings(t *testing.T) {
	m := fixture()
	a := render(t, m)
	b := trace(t, m, a)
	var tr struct {
		Status    string
		Artifacts []artifactDigest
		Locations []location
	}
	if e := json.Unmarshal(b, &tr); e != nil {
		t.Fatal(e)
	}
	want := []struct {
		p string
		b []byte
	}{{"generated/decision-table.json", a.DecisionTable}, {"generated/decision.svg", a.SVG}, {"generated/lifecycle.go", []byte("code\n")}, {"generated/model_paths_test.go", []byte("tests\n")}}
	if len(tr.Artifacts) != 4 {
		t.Fatal("wrong frontier")
	}
	for i, x := range want {
		if tr.Artifacts[i].Path != x.p || tr.Artifacts[i].SHA256 != fmt.Sprintf("%x", sha256.Sum256(x.b)) {
			t.Fatal("wrong digest/path")
		}
	}
	if len(tr.Locations) != 7 {
		t.Fatal("missing locations")
	}
	req, trans := false, false
	for _, l := range tr.Locations {
		if l.ID == "R_actor" {
			req = reflect.DeepEqual(l.Guards, []string{"G_actor"}) && reflect.DeepEqual(l.Transitions, []string{"T_finish"})
		}
		if l.ID == "T_finish" {
			trans = reflect.DeepEqual(l.Requirements, []string{"R_actor", "R_outcome"}) && l.Path == "model/Example/transitions/T_finish"
		}
	}
	if !req || !trans || !strings.Contains(tr.Status, "no verified execution") {
		t.Fatal("wrong trace meaning")
	}
	a.SVG = append(a.SVG, ' ')
	if bytes.Equal(b, trace(t, m, a)) {
		t.Fatal("changed bytes invisible")
	}
	if _, e := Trace(m, identity(), Artifacts{}, CodeArtifacts{}); e == nil {
		t.Fatal("missing artifacts")
	}
}

func TestTransitionSetOrderAndCycle(t *testing.T) {
	m := fixture()
	m.Transitions = append(m.Transitions, model.Transition{ID: "T_retry", From: "Draft", To: "Draft", Trigger: "retry", Guards: []string{"G_actor"}, Effects: []model.Effect{{Op: model.SetState}, {Op: model.IncrementRevision}}})
	a := render(t, m)
	tr := trace(t, m, a)
	m.Transitions[0], m.Transitions[1] = m.Transitions[1], m.Transitions[0]
	b := render(t, m)
	if !reflect.DeepEqual(a, b) || !bytes.Equal(tr, trace(t, m, b)) {
		t.Fatal("transition set order changed output")
	}
	if !bytes.Contains(a.SVG, []byte("T_retry")) {
		t.Fatal("cycle omitted")
	}
}

func TestMaximumTransitionLabelUsesCompactFont(t *testing.T) {
	m := fixture()
	m.Transitions[0].ID = "T" + strings.Repeat("x", 47)
	m.Transitions[0].Trigger = "E" + strings.Repeat("y", 47)
	a := render(t, m)
	d := xml.NewDecoder(bytes.NewReader(a.SVG))
	compact := false
	for {
		tok, e := d.Token()
		if e == io.EOF {
			break
		}
		if e != nil {
			t.Fatal(e)
		}
		if el, ok := tok.(xml.StartElement); ok && el.Name.Local == "text" {
			x, font := "", ""
			for _, at := range el.Attr {
				if at.Name.Local == "x" {
					x = at.Value
				}
				if at.Name.Local == "font-size" {
					font = at.Value
				}
			}
			if x == "380" {
				compact = font == "7"
			}
		}
	}
	if !compact {
		t.Fatal("maximum transition label lacks compact font")
	}
}
