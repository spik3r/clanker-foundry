package model

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func ref(s string) Expr           { return Expr{Kind: "ref", Ref: s} }
func op(s string, a ...Expr) Expr { return Expr{Kind: "op", Op: s, Args: a} }
func minimal() Model {
	return Model{Version: 1, ID: "Example", Initial: "Draft", States: []State{{ID: "Draft", Label: "Draft"}, {ID: "Done", Label: "Done", Terminal: true}}, Requirements: []Requirement{{ID: "R_outcome", Text: "Outcome required"}}, Guards: []Guard{{ID: "G_outcome", Requirement: "R_outcome", Expr: op("nonempty", ref("task.outcome"))}}, Transitions: []Transition{{ID: "T_finish", From: "Draft", Trigger: "finish", To: "Done", Guards: []string{"G_outcome"}, Effects: []Effect{{SetState}, {IncrementRevision}}}}}
}
func TestSemanticRefusals(t *testing.T) {
	cases := map[string]func(*Model){"version": func(m *Model) { m.Version = 2 }, "global-id": func(m *Model) { m.Guards[0].ID = m.ID }, "missing-requirement": func(m *Model) { m.Guards[0].Requirement = "missing" }, "missing-guard": func(m *Model) { m.Transitions[0].Guards = []string{"missing"} }, "duplicate-guard": func(m *Model) { m.Transitions[0].Guards = []string{"G_outcome", "G_outcome"} }, "missing-target": func(m *Model) { m.Transitions[0].To = "missing" }, "terminal-outgoing": func(m *Model) { m.States[0].Terminal = true }, "unreachable": func(m *Model) { m.States = append(m.States, State{ID: "Other", Label: "Other", Terminal: true}) }, "dead-end": func(m *Model) { m.States[1].Terminal = false }, "ambiguous-trigger": func(m *Model) { x := m.Transitions[0]; x.ID = "T_other"; m.Transitions = append(m.Transitions, x) }, "effect-order": func(m *Model) { m.Transitions[0].Effects = []Effect{{IncrementRevision}, {SetState}} }, "extra-effect": func(m *Model) { m.Transitions[0].Effects = []Effect{{SetState}, {SetState}, {IncrementRevision}} }, "unknown-effect": func(m *Model) { m.Transitions[0].Effects[0].Op = "execute" }, "label-unicode": func(m *Model) { m.States[0].Label = string([]byte{0xff}) }, "label-limit": func(m *Model) { m.States[0].Label = strings.Repeat("x", 257) }, "unused-requirement": func(m *Model) { m.Requirements = append(m.Requirements, Requirement{"R_unused", "unused"}) }, "nonboolean-root": func(m *Model) { m.Guards[0].Expr = ref("task.outcome") }}
	for name, change := range cases {
		t.Run(name, func(t *testing.T) {
			m := minimal()
			change(&m)
			if Validate(m) == nil {
				t.Fatal("accepted invalid semantics")
			}
			if _, e := Normalize(m); e == nil {
				t.Fatal("normalization accepted invalid semantics")
			}
		})
	}
	m := minimal()
	m.Transitions[0].Effects = append([]Effect{{AssignOwner}}, m.Transitions[0].Effects...)
	if e := Validate(m); e != nil {
		t.Fatal(e)
	}
}
func TestExpressionTypesAndBounds(t *testing.T) {
	for _, e := range []Expr{op("eq", ref("task.revision"), ref("command.expectedRevision")), op("absent", ref("task.owner")), op("nonempty", ref("task.acceptance"))} {
		if typ, err := TypeOf(e); err != nil || typ != "bool" {
			t.Fatalf("valid type %q %v", typ, err)
		}
	}
	invalid := []Expr{op("eq", ref("task.revision"), ref("task.outcome")), op("eq", ref("task.acceptance"), ref("task.acceptance")), op("absent", ref("task.id")), op("nonempty", ref("task.revision")), op("all"), op("any", ref("task.id")), op("eval", ref("task.id")), ref("unknown"), {Kind: "ref", Ref: "task.id", Value: "extra"}, op("all", op("nonempty", ref("task.id")), op("nonempty", ref("task.id")))}
	deep := op("nonempty", ref("task.id"))
	for range MaxDepth + 1 {
		deep = op("all", deep)
	}
	invalid = append(invalid, deep)
	for _, e := range invalid {
		if _, err := TypeOf(e); err == nil {
			t.Fatalf("accepted invalid expr %#v", e)
		}
	}
}
func TestNormalizeDeepCopyAndNestedSetDuplicates(t *testing.T) {
	m := minimal()
	a := op("nonempty", ref("task.outcome"))
	b := op("nonempty", ref("command.actor"))
	m.Guards[0].Expr = op("all", b, a)
	before := ExprKey(m.Guards[0].Expr)
	n, e := Normalize(m)
	if e != nil {
		t.Fatal(e)
	}
	if before != ExprKey(m.Guards[0].Expr) {
		t.Fatal("Normalize mutated input")
	}
	n.States[0].Label = "changed"
	n.Requirements[0].Text = "changed"
	n.Guards[0].Expr.Args[0].Args[0].Ref = "changed"
	n.Transitions[0].Guards[0] = "changed"
	n.Transitions[0].Effects[0].Op = "changed"
	if m.States[0].Label == "changed" || m.Requirements[0].Text == "changed" || ExprKey(m.Guards[0].Expr) != before || m.Transitions[0].Guards[0] != "G_outcome" || m.Transitions[0].Effects[0].Op != SetState {
		t.Fatal("normalized result aliases input")
	}
	x := minimal()
	x.Guards[0].Expr = op("any", op("all", a, b), op("all", b, a))
	if _, e := Normalize(x); e == nil {
		t.Fatal("nested reordered duplicate operands accepted")
	}
	p, _ := Normalize(m)
	q, _ := Normalize(p)
	if !reflect.DeepEqual(p, q) {
		t.Fatal("normalization not idempotent")
	}
}

func TestAggregateExpressionAndCollectionBounds(t *testing.T) {
	m := minimal()
	m.Guards = nil
	m.Transitions[0].Guards = nil
	for i := range 64 {
		id := fmt.Sprintf("G_%d", i)
		m.Guards = append(m.Guards, Guard{ID: id, Requirement: "R_outcome", Expr: op("all", op("nonempty", ref("task.id")), op("nonempty", ref("task.outcome")), op("nonempty", ref("command.actor")), op("nonempty", ref("task.owner")))})
		m.Transitions[0].Guards = append(m.Transitions[0].Guards, id)
	}
	if e := Validate(m); e == nil {
		t.Fatal("aggregate expression node budget not enforced")
	}
	for _, mutate := range []func(*Model){func(m *Model) {
		for len(m.States) < 17 {
			m.States = append(m.States, State{ID: fmt.Sprintf("S_%d", len(m.States)), Label: "extra", Terminal: true})
		}
	}, func(m *Model) { m.Requirements = make([]Requirement, 65) }, func(m *Model) { m.Guards = make([]Guard, 65) }, func(m *Model) { m.Transitions = make([]Transition, 33) }} {
		m := minimal()
		mutate(&m)
		if e := Validate(m); e == nil {
			t.Fatal("collection budget not enforced")
		}
	}
}
