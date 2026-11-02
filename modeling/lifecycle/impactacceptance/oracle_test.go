// Package impactacceptance freezes handwritten design-diff expectations before
// comparison implementation. Declared links are not execution or approval proof.
package impactacceptance

import (
	"reflect"
	"strings"
	"testing"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/canonical"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/impact"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/model"
)

// Independently hashed Python json(sort_keys=True,separators=(',',':')) UTF-8
// canonical documents, with no trailing newline; never obtained from Compare.
const baselineDigest = "aeee4ccf6214676cab8824765ae390376b4d96c7458c156ff473fa8bb60d4735"
const labelDigest = "0b169effc3a3d36404c6e67fbad92938ee6a164b0b4e5c150541def272ea2e2e"

func nonempty(ref string) model.Expr {
	return model.Expr{Kind: "op", Op: "nonempty", Args: []model.Expr{{Kind: "ref", Ref: ref}}}
}
func effects() []model.Effect {
	return []model.Effect{{Op: model.SetState}, {Op: model.IncrementRevision}}
}
func baseline() model.Model {
	return model.Model{Version: 1, ID: "M", Initial: "A", States: []model.State{{ID: "A", Label: "A"}, {ID: "B", Label: "B"}, {ID: "C", Label: "C", Terminal: true}}, Requirements: []model.Requirement{{ID: "R", Text: "Invented requirement"}}, Guards: []model.Guard{{ID: "G1", Requirement: "R", Expr: nonempty("task.outcome")}, {ID: "G2", Requirement: "R", Expr: nonempty("command.actor")}}, Transitions: []model.Transition{
		{ID: "T1", From: "A", Trigger: "advance", To: "B", Guards: []string{"G1"}, Effects: effects()},
		{ID: "T2", From: "B", Trigger: "finish", To: "C", Guards: []string{"G1"}, Effects: effects()},
		{ID: "T3", From: "A", Trigger: "skipExample", To: "C", Guards: []string{"G2"}, Effects: effects()},
	}}
}
func change(kind, id, typ string, fields ...string) impact.Change {
	return impact.Change{Kind: kind, ID: id, Type: typ, Fields: append([]string{}, fields...)}
}
func link(side, id string, causes ...string) impact.LinkedTransition {
	return impact.LinkedTransition{Side: side, ID: id, Causes: append([]string{}, causes...)}
}

type literalCase struct {
	name          string
	before, after model.Model
	changes       []impact.Change
	links         []impact.LinkedTransition
}

// The models and expected record lists below are handwritten. Never replace
// these lists with generated output or calculate them from a comparison result.
func frozenCases() []literalCase {
	out := []literalCase{}
	add := func(name string, b, a model.Model, c []impact.Change, l []impact.LinkedTransition) {
		out = append(out, literalCase{name, b, a, c, l})
	}
	b, a := baseline(), baseline()
	a.States[0], a.States[2] = a.States[2], a.States[0]
	a.Transitions[0], a.Transitions[2] = a.Transitions[2], a.Transitions[0]
	a.Guards[0], a.Guards[1] = a.Guards[1], a.Guards[0]
	add("collection permutations", b, a, []impact.Change{}, []impact.LinkedTransition{})
	b, a = baseline(), baseline()
	b.Transitions[0].Guards = []string{"G1", "G2"}
	a.Transitions[0].Guards = []string{"G2", "G1"}
	add("guard set permutation", b, a, []impact.Change{}, []impact.LinkedTransition{})
	b, a = baseline(), baseline()
	x, y := nonempty("task.outcome"), nonempty("task.owner")
	b.Guards[0].Expr = model.Expr{Kind: "op", Op: "all", Args: []model.Expr{x, y}}
	a.Guards[0].Expr = model.Expr{Kind: "op", Op: "all", Args: []model.Expr{y, x}}
	add("boolean operand permutation", b, a, []impact.Change{}, []impact.LinkedTransition{})
	b, a = baseline(), baseline()
	a.States[1].Label = "Reworded B"
	add("label", b, a, []impact.Change{change("state", "B", "modified", "label")}, []impact.LinkedTransition{link("before", "T1", "state:B"), link("before", "T2", "state:B"), link("after", "T1", "state:B"), link("after", "T2", "state:B")})
	b, a = baseline(), baseline()
	a.Requirements[0].Text = "Reworded requirement"
	add("requirement text", b, a, []impact.Change{change("requirement", "R", "modified", "text")}, []impact.LinkedTransition{link("before", "T1", "requirement:R"), link("before", "T2", "requirement:R"), link("before", "T3", "requirement:R"), link("after", "T1", "requirement:R"), link("after", "T2", "requirement:R"), link("after", "T3", "requirement:R")})
	b, a = baseline(), baseline()
	a.Guards[0].Expr = nonempty("task.owner")
	add("guard expression", b, a, []impact.Change{change("guard", "G1", "modified", "expr")}, []impact.LinkedTransition{link("before", "T1", "guard:G1"), link("before", "T2", "guard:G1"), link("after", "T1", "guard:G1"), link("after", "T2", "guard:G1")})
	b, a = baseline(), baseline()
	a.Guards[0].Expr = nonempty("task.owner")
	a.Transitions[0].Guards = []string{"G2"}
	a.Transitions[1].Guards = []string{"G2"}
	a.Transitions[2].Guards = []string{"G1"}
	add("both graphs reassignment", b, a, []impact.Change{change("guard", "G1", "modified", "expr"), change("transition", "T1", "modified", "guards"), change("transition", "T2", "modified", "guards"), change("transition", "T3", "modified", "guards")}, []impact.LinkedTransition{link("before", "T1", "guard:G1", "transition:T1"), link("before", "T2", "guard:G1", "transition:T2"), link("before", "T3", "transition:T3"), link("after", "T1", "transition:T1"), link("after", "T2", "transition:T2"), link("after", "T3", "guard:G1", "transition:T3")})
	b, a = baseline(), baseline()
	a.Guards = append(a.Guards, model.Guard{ID: "G3", Requirement: "R", Expr: nonempty("task.owner")})
	a.Transitions[0].Guards = []string{"G1", "G3"}
	c := []impact.Change{change("guard", "G3", "added"), change("transition", "T1", "modified", "guards")}
	l := []impact.LinkedTransition{link("before", "T1", "transition:T1"), link("after", "T1", "guard:G3", "transition:T1")}
	add("guard added", b, a, c, l)
	add("guard removed", a, b, []impact.Change{change("guard", "G3", "removed"), change("transition", "T1", "modified", "guards")}, []impact.LinkedTransition{link("before", "T1", "guard:G3", "transition:T1"), link("after", "T1", "transition:T1")})
	b, a = baseline(), baseline()
	a.Guards = a.Guards[:1]
	a.Transitions = a.Transitions[:2]
	add("guard and transition removed", b, a, []impact.Change{change("guard", "G2", "removed"), change("transition", "T3", "removed")}, []impact.LinkedTransition{link("before", "T3", "guard:G2", "transition:T3")})
	add("guard and transition added", a, b, []impact.Change{change("guard", "G2", "added"), change("transition", "T3", "added")}, []impact.LinkedTransition{link("after", "T3", "guard:G2", "transition:T3")})
	b, a = baseline(), baseline()
	a.Requirements = append(a.Requirements, model.Requirement{ID: "R2", Text: "Second invented requirement"})
	a.Guards[0].Requirement = "R2"
	add("requirement reassignment", b, a, []impact.Change{change("guard", "G1", "modified", "requirement"), change("requirement", "R2", "added")}, []impact.LinkedTransition{link("before", "T1", "guard:G1"), link("before", "T2", "guard:G1"), link("after", "T1", "guard:G1", "requirement:R2"), link("after", "T2", "guard:G1", "requirement:R2")})
	add("requirement removal and reassignment", a, b, []impact.Change{change("guard", "G1", "modified", "requirement"), change("requirement", "R2", "removed")}, []impact.LinkedTransition{link("before", "T1", "guard:G1", "requirement:R2"), link("before", "T2", "guard:G1", "requirement:R2"), link("after", "T1", "guard:G1"), link("after", "T2", "guard:G1")})
	b, a = baseline(), baseline()
	a.Transitions[0].Effects = append([]model.Effect{{Op: model.AssignOwner}}, effects()...)
	a.Transitions[0].Trigger = "beginExample"
	add("ordered effects and field ordering", b, a, []impact.Change{change("transition", "T1", "modified", "effects", "trigger")}, []impact.LinkedTransition{link("before", "T1", "transition:T1"), link("after", "T1", "transition:T1")})
	b, a = baseline(), baseline()
	a.Transitions[2].From = "B"
	add("transition source", b, a, []impact.Change{change("transition", "T3", "modified", "from")}, []impact.LinkedTransition{link("before", "T3", "transition:T3"), link("after", "T3", "transition:T3")})
	b, a = baseline(), baseline()
	a.ID = "N"
	add("root identity", b, a, []impact.Change{change("model", "", "modified", "id")}, []impact.LinkedTransition{link("before", "T1", "model:"), link("before", "T2", "model:"), link("before", "T3", "model:"), link("after", "T1", "model:"), link("after", "T2", "model:"), link("after", "T3", "model:")})
	b, a = baseline(), baseline()
	a.Initial = "B"
	a.Transitions[1].To = "A"
	add("root initial and target", b, a, []impact.Change{change("model", "", "modified", "initial"), change("transition", "T2", "modified", "to")}, []impact.LinkedTransition{link("before", "T1", "model:"), link("before", "T2", "model:", "transition:T2"), link("before", "T3", "model:"), link("after", "T1", "model:"), link("after", "T2", "model:", "transition:T2"), link("after", "T3", "model:")})
	b, a = baseline(), baseline()
	a.States[1].Terminal = true
	a.Transitions = []model.Transition{a.Transitions[0], a.Transitions[2]}
	add("terminal and incident removal", b, a, []impact.Change{change("state", "B", "modified", "terminal"), change("transition", "T2", "removed")}, []impact.LinkedTransition{link("before", "T1", "state:B"), link("before", "T2", "state:B", "transition:T2"), link("after", "T1", "state:B")})
	b, a = baseline(), baseline()
	a.States = append(a.States, model.State{ID: "D", Label: "D", Terminal: true})
	a.Transitions = append(a.Transitions, model.Transition{ID: "T4", From: "A", Trigger: "extraExample", To: "D", Guards: []string{"G2"}, Effects: effects()})
	add("state added", b, a, []impact.Change{change("state", "D", "added"), change("transition", "T4", "added")}, []impact.LinkedTransition{link("after", "T4", "state:D", "transition:T4")})
	add("state removed", a, b, []impact.Change{change("state", "D", "removed"), change("transition", "T4", "removed")}, []impact.LinkedTransition{link("before", "T4", "state:D", "transition:T4")})
	b, a = baseline(), baseline()
	a.Guards[1].ID = "G3"
	a.Transitions[2].Guards = []string{"G3"}
	a.States = append(a.States, model.State{ID: "G2", Label: "Replacement kind", Terminal: true})
	a.Transitions = append(a.Transitions, model.Transition{ID: "T4", From: "A", Trigger: "replacementExample", To: "G2", Guards: []string{"G3"}, Effects: effects()})
	add("cross kind replacement", b, a, []impact.Change{change("guard", "G2", "removed"), change("guard", "G3", "added"), change("state", "G2", "added"), change("transition", "T3", "modified", "guards"), change("transition", "T4", "added")}, []impact.LinkedTransition{link("before", "T3", "guard:G2", "transition:T3"), link("after", "T3", "guard:G3", "transition:T3"), link("after", "T4", "guard:G3", "state:G2", "transition:T4")})

	b, a = baseline(), baseline()
	b.Requirements = append(b.Requirements, model.Requirement{ID: "R2", Text: "Second requirement old"})
	a.Requirements = []model.Requirement{{ID: "R2", Text: "Second requirement edited"}}
	b.Guards[1].Requirement = "R2"
	a.Guards[0].Requirement = "R2"
	a.Guards[1].Requirement = "R2"
	b.Transitions[0].Guards = []string{"G1", "G2"}
	a.Transitions[0].Guards = []string{"G1", "G2"}
	add("every multi-guard requirement cause", b, a, []impact.Change{change("guard", "G1", "modified", "requirement"), change("requirement", "R", "removed"), change("requirement", "R2", "modified", "text")}, []impact.LinkedTransition{
		link("before", "T1", "guard:G1", "requirement:R", "requirement:R2"), link("before", "T2", "guard:G1", "requirement:R"), link("before", "T3", "requirement:R2"),
		link("after", "T1", "guard:G1", "requirement:R2"), link("after", "T2", "guard:G1", "requirement:R2"), link("after", "T3", "requirement:R2"),
	})
	return out
}
func TestFrozenLiteralChangesAndBothGraphCauses(t *testing.T) {
	for _, tc := range frozenCases() {
		t.Run(tc.name, func(t *testing.T) {
			if e := model.Validate(tc.before); e != nil {
				t.Fatal("invalid handwritten before model", e)
			}
			if e := model.Validate(tc.after); e != nil {
				t.Fatal("invalid handwritten after model", e)
			}
			r, e := impact.Compare(tc.before, tc.after, nil)
			if e != nil {
				t.Fatal(e)
			}
			if !reflect.DeepEqual(r.Changes, tc.changes) || !reflect.DeepEqual(r.LinkedTransitions, tc.links) {
				t.Fatalf("literal mismatch\nchanges got %#v want %#v\nlinks got %#v want %#v", r.Changes, tc.changes, r.LinkedTransitions, tc.links)
			}
			bd, e := canonical.Digest(tc.before)
			if e != nil {
				t.Fatal(e)
			}
			ad, e := canonical.Digest(tc.after)
			if e != nil {
				t.Fatal(e)
			}
			if r.BeforeDigest != bd || r.AfterDigest != ad || r.Changed != (bd != ad) || r.ReviewRequired != (bd != ad) {
				t.Fatal("identity/flags differ from core canonical meaning")
			}
			if r.ReportVersion != "lifecycle-model-diff-report/v1" || r.ProfileVersion != "lifecycle/v1" || r.Mode != "example-only" || r.Authority != "none" || r.RequestDigest != nil || r.BeforeSourceDigest != nil || r.AfterSourceDigest != nil {
				t.Fatal("wrong typed envelope")
			}
			if !reflect.DeepEqual(r.Bindings, []impact.BindingResult{}) {
				t.Fatal("nil bindings did not normalize to empty array")
			}
		})
	}
}
func TestFrozenBindingPrecedenceAndOrdering(t *testing.T) {
	b, a := baseline(), baseline()
	a.States[1].Label = "Reworded B"
	for _, v := range []struct {
		m    model.Model
		want string
	}{{b, baselineDigest}, {a, labelDigest}} {
		d, e := canonical.Digest(v.m)
		if e != nil || d != v.want {
			t.Fatal("independent canonical vector mismatch")
		}
	}
	supplied := []impact.Binding{{ID: "zBefore", ModelDigest: baselineDigest}, {ID: "aAfter", ModelDigest: labelDigest}, {ID: "M", ModelDigest: strings.Repeat("0", 64)}}
	r, e := impact.Compare(b, a, supplied)
	if e != nil {
		t.Fatal(e)
	}
	want := []impact.BindingResult{{ID: "M", ModelDigest: strings.Repeat("0", 64), Status: "matches-neither"}, {ID: "aAfter", ModelDigest: labelDigest, Status: "matches-after"}, {ID: "zBefore", ModelDigest: baselineDigest, Status: "matches-before-only"}}
	if !reflect.DeepEqual(r.Bindings, want) {
		t.Fatalf("binding literals: %#v", r.Bindings)
	}
	r, e = impact.Compare(b, b, []impact.Binding{{ID: "M", ModelDigest: baselineDigest}})
	if e != nil {
		t.Fatal(e)
	}
	if !reflect.DeepEqual(r.Bindings, []impact.BindingResult{{ID: "M", ModelDigest: baselineDigest, Status: "matches-after"}}) || r.Changed || r.ReviewRequired || !reflect.DeepEqual(r.Changes, []impact.Change{}) || !reflect.DeepEqual(r.LinkedTransitions, []impact.LinkedTransition{}) {
		t.Fatal("equal identity binding precedence/empty collections")
	}
	r, e = impact.Compare(b, b, []impact.Binding{})
	if e != nil || !reflect.DeepEqual(r.Bindings, []impact.BindingResult{}) {
		t.Fatal("explicit empty bindings")
	}
}

// Everything above this line is the exact frozen oracle, SHA-256
// 4497a7a693a62ac9f3d431f76660026310c12aa3274a3ae0d8bb3fd2bf753f85.
// Additional tests below do not derive or refresh those expected values.
func cloneExpr(e model.Expr) model.Expr {
	if e.Args != nil {
		a := make([]model.Expr, len(e.Args))
		for i, x := range e.Args {
			a[i] = cloneExpr(x)
		}
		e.Args = a
	}
	return e
}
func cloneModel(m model.Model) model.Model {
	m.States = append([]model.State{}, m.States...)
	m.Requirements = append([]model.Requirement{}, m.Requirements...)
	m.Guards = append([]model.Guard{}, m.Guards...)
	m.Transitions = append([]model.Transition{}, m.Transitions...)
	for i := range m.Guards {
		m.Guards[i].Expr = cloneExpr(m.Guards[i].Expr)
	}
	for i := range m.Transitions {
		m.Transitions[i].Guards = append([]string{}, m.Transitions[i].Guards...)
		m.Transitions[i].Effects = append([]model.Effect{}, m.Transitions[i].Effects...)
	}
	return m
}
func requireRejected(t *testing.T, b, a model.Model, bindings []impact.Binding) {
	t.Helper()
	r, e := impact.Compare(b, a, bindings)
	if e == nil {
		t.Fatal("invalid input accepted")
	}
	if !reflect.DeepEqual(r, impact.Report{}) {
		t.Fatal("partial typed report leaked on rejection")
	}
}
func TestInvalidTypedModelsBothSides(t *testing.T) {
	mutations := map[string]func(*model.Model){
		"version": func(m *model.Model) { m.Version = 2 }, "root collision": func(m *model.Model) { m.ID = "A" }, "duplicate": func(m *model.Model) { m.Guards[0].ID = "R" }, "dangling": func(m *model.Model) { m.Transitions[0].To = "Missing" }, "unknown reference": func(m *model.Model) { m.Guards[0].Expr.Args[0].Ref = "unknown.field" }, "unknown operator": func(m *model.Model) { m.Guards[0].Expr.Op = "execute" }, "unknown effect": func(m *model.Model) { m.Transitions[0].Effects[0].Op = "shell" }, "reordered effects": func(m *model.Model) {
			m.Transitions[0].Effects = []model.Effect{{Op: model.IncrementRevision}, {Op: model.SetState}}
		}, "unused guard": func(m *model.Model) {
			m.Guards = append(m.Guards, model.Guard{ID: "Unused", Requirement: "R", Expr: nonempty("task.owner")})
		}, "unreachable": func(m *model.Model) {
			m.States = append(m.States, model.State{ID: "UnusedState", Label: "Unused", Terminal: true})
		}, "invalid utf8": func(m *model.Model) { m.States[0].Label = string([]byte{255}) }, "ambiguous": func(m *model.Model) { x := m.Transitions[0]; x.ID = "T4"; m.Transitions = append(m.Transitions, x) },
	}
	for name, f := range mutations {
		t.Run(name, func(t *testing.T) {
			m := baseline()
			f(&m)
			requireRejected(t, m, baseline(), nil)
			requireRejected(t, baseline(), m, nil)
		})
	}
}
func TestBindingsBoundariesAndNamespaces(t *testing.T) {
	all := []impact.Binding{}
	for i := 0; i < 128; i++ {
		id := "B" + string(rune('A'+i/26)) + string(rune('A'+i%26))
		all = append(all, impact.Binding{ID: id, ModelDigest: baselineDigest})
	}
	r, e := impact.Compare(baseline(), baseline(), all)
	if e != nil || len(r.Bindings) != 128 {
		t.Fatal("128 valid bindings rejected", e)
	}
	for i, x := range r.Bindings {
		if x.ID != all[i].ID || x.Status != "matches-after" || x.ModelDigest != baselineDigest {
			t.Fatal("wrong boundary binding")
		}
	}
	requireRejected(t, baseline(), baseline(), append(all, impact.Binding{ID: "Overflow", ModelDigest: baselineDigest}))
	for _, id := range []string{"", "0bad", "bad-id", strings.Repeat("A", 49), "雪"} {
		requireRejected(t, baseline(), baseline(), []impact.Binding{{ID: id, ModelDigest: baselineDigest}})
	}
	for _, digest := range []string{"", strings.Repeat("a", 63), strings.Repeat("a", 65), strings.Repeat("A", 64), strings.Repeat("g", 64)} {
		requireRejected(t, baseline(), baseline(), []impact.Binding{{ID: "Valid", ModelDigest: digest}})
	}
	requireRejected(t, baseline(), baseline(), []impact.Binding{{ID: "Same", ModelDigest: baselineDigest}, {ID: "Same", ModelDigest: labelDigest}})
	supplied := []impact.Binding{{ID: "a", ModelDigest: baselineDigest}, {ID: "A", ModelDigest: baselineDigest}, {ID: strings.Repeat("Z", 48), ModelDigest: baselineDigest}}
	r, e = impact.Compare(baseline(), baseline(), supplied)
	if e != nil || len(r.Bindings) != 3 || r.Bindings[0].ID != "A" || r.Bindings[1].ID != strings.Repeat("Z", 48) || r.Bindings[2].ID != "a" {
		t.Fatal("case sensitivity, separate namespace or maximum ID length failed", e)
	}
	nilReport, e := impact.Compare(baseline(), baseline(), nil)
	if e != nil {
		t.Fatal(e)
	}
	emptyReport, e := impact.Compare(baseline(), baseline(), []impact.Binding{})
	if e != nil || !reflect.DeepEqual(nilReport, emptyReport) {
		t.Fatal("nil binding set differs from explicit empty")
	}
}
func TestDeepNonmutationAliasingAndRepeatedTypedReports(t *testing.T) {
	tc := frozenCases()[len(frozenCases())-1]
	b, a := tc.before, tc.after
	before, after := cloneModel(b), cloneModel(a)
	bindings := []impact.Binding{{ID: "Z", ModelDigest: baselineDigest}, {ID: "A", ModelDigest: labelDigest}}
	bindingCopy := append([]impact.Binding{}, bindings...)
	original, e := impact.Compare(b, a, bindings)
	if e != nil {
		t.Fatal(e)
	}
	for i := 0; i < 3; i++ {
		r, e := impact.Compare(b, a, bindings)
		if e != nil || !reflect.DeepEqual(r, original) {
			t.Fatal("typed report nondeterministic")
		}
	}
	if !reflect.DeepEqual(b, before) || !reflect.DeepEqual(a, after) || !reflect.DeepEqual(bindings, bindingCopy) {
		t.Fatal("caller models/bindings mutated")
	}
	first, e := impact.Compare(b, a, bindings)
	if e != nil {
		t.Fatal(e)
	}
	second, e := impact.Compare(b, a, bindings)
	if e != nil {
		t.Fatal(e)
	}
	first.Changes[0].Fields[0] = "field sentinel"
	first.LinkedTransitions[0].Causes[0] = "cause sentinel"
	first.Bindings[0].ID = "binding sentinel"
	if !reflect.DeepEqual(second, original) {
		t.Fatal("reports share output collections")
	}
	if first.LinkedTransitions[1].Causes[0] != "guard:G1" || first.LinkedTransitions[3].Causes[0] != "guard:G1" {
		t.Fatal("link records share cause arrays")
	}
	if !reflect.DeepEqual(b, before) || !reflect.DeepEqual(a, after) || !reflect.DeepEqual(bindings, bindingCopy) {
		t.Fatal("report mutation aliases caller")
	}
	b.States[0].Label = "caller state"
	a.Guards[0].Expr.Args[0].Ref = "task.owner"
	bindings[0].ID = "caller binding"
	if !reflect.DeepEqual(second, original) {
		t.Fatal("caller mutation aliases returned report")
	}
}

// Handwritten wire documents: expected report sets remain solely in the frozen
// prefix. No model/scenario encoder or comparison output authors these inputs.
const wireModel = `{"formatVersion":"lifecycle-json/v1","profileVersion":"lifecycle/v1","id":"M","initial":"A","states":[{"id":"A","label":"A","terminal":false},{"id":"B","label":"B","terminal":false},{"id":"C","label":"C","terminal":true}],"requirements":[{"id":"R","text":"Invented requirement"}],"guards":[{"id":"G1","requirement":"R","expr":{"op":"nonempty","args":[{"ref":"task.outcome"}]}},{"id":"G2","requirement":"R","expr":{"op":"nonempty","args":[{"ref":"command.actor"}]}}],"transitions":[{"id":"T1","from":"A","trigger":"advance","to":"B","guards":["G1"],"effects":[{"op":"set_state"},{"op":"increment_revision"}]},{"id":"T2","from":"B","trigger":"finish","to":"C","guards":["G1"],"effects":[{"op":"set_state"},{"op":"increment_revision"}]},{"id":"T3","from":"A","trigger":"skipExample","to":"C","guards":["G2"],"effects":[{"op":"set_state"},{"op":"increment_revision"}]}]}`

func wireRequest(b, a, bindings string) []byte {
	return []byte(`{"formatVersion":"lifecycle-model-diff-json/v1","before":` + b + `,"after":` + a + `,"bindings":` + bindings + `}`)
}
func TestMalformedWireRejectsWithoutPartialBytes(t *testing.T) {
	good := string(wireRequest(wireModel, wireModel, `[]`))
	replace := func(old, new string) []byte {
		if !strings.Contains(good, old) {
			t.Fatal("missing mutation anchor")
		}
		return []byte(strings.Replace(good, old, new, 1))
	}
	invalid := map[string][]byte{
		"bindings null":            wireRequest(wireModel, wireModel, `null`),
		"bindings missing":         []byte(strings.TrimSuffix(good, `,"bindings":[]}`) + `}`),
		"bindings unknown":         wireRequest(wireModel, wireModel, `[{"id":"X","modelDigest":"`+baselineDigest+`","extra":0}]`),
		"bindings missing field":   wireRequest(wireModel, wireModel, `[{"id":"X"}]`),
		"bindings duplicate field": wireRequest(wireModel, wireModel, `[{"id":"X","id":"Y","modelDigest":"`+baselineDigest+`"}]`),
		"model profile":            replace(`"profileVersion":"lifecycle/v1"`, `"profileVersion":"unknown/v1"`),
		"request version":          replace(`"formatVersion":"lifecycle-model-diff-json/v1"`, `"formatVersion":"unknown/v1"`),
		"case variant":             replace(`"before":`, `"Before":`),
		"unknown field":            replace(`"before":`, `"extra":true,"before":`),
		"duplicate field":          replace(`"before":`, `"before":null,"before":`),
		"null model":               wireRequest(`null`, wireModel, `[]`),
		"missing model field":      wireRequest(strings.Replace(wireModel, `"initial":"A",`, "", 1), wireModel, `[]`),
		"malformed later model":    wireRequest(wireModel, strings.Replace(wireModel, `"op":"nonempty"`, `"op":"execute"`, 1), `[]`),
		"surrogate":                wireRequest(strings.Replace(wireModel, `"label":"A"`, `"label":"\ud800"`, 1), wireModel, `[]`),
		"trailing":                 []byte(good + ` {}`),
		"invalid utf8":             append([]byte{255}, []byte(good)...),
		"source limit":             []byte(strings.Repeat(" ", impact.MaxSourceBytes+1) + good),
		"deep malformed root":      []byte(strings.Repeat("[", impact.MaxDepth+1) + strings.Repeat("]", impact.MaxDepth+1)),
	}
	// The deeply nested wrong-root request checks rejection, not isolation of the
	// independent depth gate. Existing semantic depth limits are stricter than a
	// valid request that could isolate the outer parser limit.
	hugeModel := strings.Replace(wireModel, "{", "{"+strings.Repeat(" ", 1<<20), 1)
	invalid["embedded source limit"] = wireRequest(hugeModel, wireModel, `[]`)
	for name, raw := range invalid {
		t.Run(name, func(t *testing.T) {
			copyRaw := append([]byte{}, raw...)
			out, e := impact.Execute(raw)
			if e == nil || len(out) != 0 {
				t.Fatal("invalid request emitted report bytes")
			}
			if !reflect.DeepEqual(raw, copyRaw) {
				t.Fatal("invalid input bytes mutated")
			}
		})
	}
}
func TestWireDeterminismBoundsAndCollectionShapes(t *testing.T) {
	raw := wireRequest(wireModel, wireModel, `[]`)
	before := append([]byte{}, raw...)
	out, e := impact.Execute(raw)
	if e != nil {
		t.Fatal(e)
	}
	if len(out) > impact.MaxReportBytes || len(out) == 0 {
		t.Fatal("report bound")
	}
	for i := 0; i < 3; i++ {
		again, e := impact.Execute(raw)
		if e != nil || string(out) != string(again) {
			t.Fatal("wire report nondeterministic")
		}
	}
	if !reflect.DeepEqual(raw, before) {
		t.Fatal("Execute mutated input bytes")
	}
	// Fixed field names and array representations from the frozen ABI, never null.
	for _, literal := range []string{`"reportVersion":"lifecycle-model-diff-report/v1"`, `"profileVersion":"lifecycle/v1"`, `"mode":"example-only"`, `"authority":"none"`, `"changed":false`, `"reviewRequired":false`, `"changes":[]`, `"linkedTransitions":[]`, `"bindings":[]`} {
		compact := strings.NewReplacer(" ", "", "\n", "", "\t", "").Replace(string(out))
		if !strings.Contains(compact, literal) {
			t.Fatal("missing wire contract literal", literal)
		}
	}
	// A legal changed model must likewise yield bounded complete output.
	changed := strings.Replace(wireModel, `"label":"B"`, `"label":"Reworded B"`, 1)
	out, e = impact.Execute(wireRequest(wireModel, changed, `[]`))
	if e != nil || len(out) == 0 || len(out) > impact.MaxReportBytes {
		t.Fatal("changed report bound", e)
	}
}

// Full wire report goldens bind independently authored exact request/value bytes.
// The SHA-256 literals were computed with Python hashlib before Execute existed.
func TestExactWireGoldenReports(t *testing.T) {
	compact := func(b []byte) string { return strings.NewReplacer(" ", "", "\n", "", "\t", "").Replace(string(b)) }
	unchanged := `{"reportVersion":"lifecycle-model-diff-report/v1","profileVersion":"lifecycle/v1","mode":"example-only","authority":"none","beforeDigest":"aeee4ccf6214676cab8824765ae390376b4d96c7458c156ff473fa8bb60d4735","afterDigest":"aeee4ccf6214676cab8824765ae390376b4d96c7458c156ff473fa8bb60d4735","requestDigest":"623f9c1f56afaba032fa4cd0262b27bd9b27e29e3cd166e3019113140378ae95","beforeSourceDigest":"f92248cad49e3f94aae846570e855594019f2a1a63878696e45dd5a28ca0bd6e","afterSourceDigest":"f92248cad49e3f94aae846570e855594019f2a1a63878696e45dd5a28ca0bd6e","changed":false,"reviewRequired":false,"changes":[],"linkedTransitions":[],"bindings":[]}`
	changed := `{"reportVersion":"lifecycle-model-diff-report/v1","profileVersion":"lifecycle/v1","mode":"example-only","authority":"none","beforeDigest":"aeee4ccf6214676cab8824765ae390376b4d96c7458c156ff473fa8bb60d4735","afterDigest":"0b169effc3a3d36404c6e67fbad92938ee6a164b0b4e5c150541def272ea2e2e","requestDigest":"62ea09351ee44d978aa8b77a1412c12dc34e586928c81e3bc82868efafb08d28","beforeSourceDigest":"f92248cad49e3f94aae846570e855594019f2a1a63878696e45dd5a28ca0bd6e","afterSourceDigest":"3d7910ae4b9807ec067fcd80688eeead46015c78e5f2f1ed6e405d49a5642374","changed":true,"reviewRequired":true,"changes":[{"kind":"state","id":"B","type":"modified","fields":["label"]}],"linkedTransitions":[{"side":"before","id":"T1","causes":["state:B"]},{"side":"before","id":"T2","causes":["state:B"]},{"side":"after","id":"T1","causes":["state:B"]},{"side":"after","id":"T2","causes":["state:B"]}],"bindings":[]}`
	for _, tc := range []struct{ name, after, want string }{{"unchanged", wireModel, unchanged}, {"label", strings.Replace(wireModel, `"label":"B"`, `"label":"Reworded B"`, 1), changed}} {
		t.Run(tc.name, func(t *testing.T) {
			b, e := impact.Execute(wireRequest(wireModel, tc.after, `[]`))
			if e != nil {
				t.Fatal(e)
			}
			if compact(b) != tc.want {
				t.Fatalf("wire golden differs: %s", b)
			}
		})
	}
}

func TestAdditionalUnorderedSetsAndExplicitEmptyOutput(t *testing.T) {
	tc := frozenCases()[len(frozenCases())-1]
	b := tc.before
	a := cloneModel(b)
	a.Requirements[0], a.Requirements[1] = a.Requirements[1], a.Requirements[0]
	a.Transitions[0].Guards[0], a.Transitions[0].Guards[1] = a.Transitions[0].Guards[1], a.Transitions[0].Guards[0]
	for _, op := range []string{"all", "any"} {
		x, y := nonempty("task.outcome"), nonempty("task.owner")
		b.Guards[0].Expr = model.Expr{Kind: "op", Op: op, Args: []model.Expr{x, y}}
		a.Guards[0].Expr = model.Expr{Kind: "op", Op: op, Args: []model.Expr{y, x}}
		r, e := impact.Compare(b, a, []impact.Binding{})
		if e != nil {
			t.Fatal(e)
		}
		if r.Changed || r.ReviewRequired || !reflect.DeepEqual(r.Changes, []impact.Change{}) || !reflect.DeepEqual(r.LinkedTransitions, []impact.LinkedTransition{}) || !reflect.DeepEqual(r.Bindings, []impact.BindingResult{}) {
			t.Fatal("semantic sets changed normalized report")
		}
	}
}
