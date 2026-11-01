// Package model defines format-independent semantics for the isolated example.
package model

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"unicode/utf8"
)

const ProfileVersion = "lifecycle/v1"
const MaxDepth = 16
const MaxNodes = 512

type Model struct {
	Version      int
	ID           string
	Initial      string
	States       []State
	Requirements []Requirement
	Guards       []Guard
	Transitions  []Transition
}
type State struct {
	ID, Label string
	Terminal  bool
}
type Requirement struct{ ID, Text string }
type Guard struct {
	ID, Requirement string
	Expr            Expr
}
type Transition struct {
	ID, From, Trigger, To string
	Guards                []string
	Effects               []Effect
}
type Effect struct{ Op string }
type Expr struct {
	Kind, Ref, Value, Op string
	Args                 []Expr
}

const (
	AssignOwner       = "assign_owner_if_absent"
	SetState          = "set_state"
	IncrementRevision = "increment_revision"
)

type refType struct {
	typ      string
	optional bool
}

var references = map[string]refType{
	"task.id": {"string", false}, "task.revision": {"uint64", false},
	"task.outcome": {"string", false}, "task.acceptance": {"strings", false},
	"task.owner": {"string", true}, "task.digest": {"string", false},
	"command.actor": {"string", false}, "command.expectedRevision": {"uint64", false},
	"review.taskId": {"string", true}, "review.taskRevision": {"uint64", true},
	"review.modelDigest": {"string", true}, "review.taskDigest": {"string", true},
	"review.reviewer": {"string", true}, "review.decision": {"string", true},
	"model.digest": {"string", false},
}

func Identifier(s string) bool {
	if len(s) < 1 || len(s) > 48 {
		return false
	}
	for i, c := range []byte(s) {
		if !(c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'z' || i > 0 && (c >= '0' && c <= '9' || c == '_')) {
			return false
		}
	}
	return true
}
func text(s string, limit int, nonblank bool) bool {
	return len(s) <= limit && utf8.ValidString(s) && (!nonblank || strings.TrimSpace(s) != "")
}
func invalid(s string) error { return errors.New("invalid model: " + s) }

// ExprKey is a format-independent structural ordering key, not a digest.
func ExprKey(e Expr) string {
	atom := func(s string) string { return fmt.Sprintf("%d:%s", len(s), s) }
	s := atom(e.Kind) + atom(e.Ref) + atom(e.Value) + atom(e.Op)
	for _, a := range e.Args {
		s += atom(ExprKey(a))
	}
	return s
}
func exprType(e Expr, depth int, count *int) (string, error) {
	*count++
	if depth > MaxDepth || *count > MaxNodes {
		return "", invalid("expression bounds")
	}
	switch e.Kind {
	case "ref":
		if e.Value != "" || e.Op != "" || len(e.Args) != 0 {
			return "", invalid("mixed reference")
		}
		r, ok := references[e.Ref]
		if !ok {
			return "", invalid("unknown reference")
		}
		return r.typ, nil
	case "string":
		if e.Ref != "" || e.Op != "" || len(e.Args) != 0 || !text(e.Value, 4096, false) {
			return "", invalid("string literal")
		}
		return "string", nil
	case "op":
		if e.Ref != "" || e.Value != "" {
			return "", invalid("mixed operator")
		}
		arity := len(e.Args)
		switch e.Op {
		case "eq", "ne":
			if arity != 2 {
				return "", invalid("comparison arity")
			}
		case "nonempty", "absent":
			if arity != 1 {
				return "", invalid("unary arity")
			}
		case "all", "any":
			if arity < 1 || arity > 16 {
				return "", invalid("boolean arity")
			}
		default:
			return "", invalid("unknown operator")
		}
		types := make([]string, arity)
		for i, a := range e.Args {
			t, err := exprType(a, depth+1, count)
			if err != nil {
				return "", err
			}
			types[i] = t
		}
		switch e.Op {
		case "eq", "ne":
			if types[0] != types[1] || (types[0] != "string" && types[0] != "uint64") {
				return "", invalid("comparison type")
			}
		case "nonempty":
			if types[0] != "string" && types[0] != "strings" {
				return "", invalid("nonempty type")
			}
		case "absent":
			a := e.Args[0]
			r, ok := references[a.Ref]
			if a.Kind != "ref" || !ok || !r.optional {
				return "", invalid("absent requires optional reference")
			}
		case "all", "any":
			seen := map[string]bool{}
			for i, a := range e.Args {
				if types[i] != "bool" {
					return "", invalid("boolean operand")
				}
				k := ExprKey(a)
				if seen[k] {
					return "", invalid("duplicate boolean operand")
				}
				seen[k] = true
			}
		}
		return "bool", nil
	default:
		return "", invalid("unknown expression kind")
	}
}

// Validate applies the same bounded semantic rules to every source format.
func Validate(m Model) error {
	if m.Version != 1 || !Identifier(m.ID) || !Identifier(m.Initial) {
		return invalid("identity/version")
	}
	if len(m.States) < 2 || len(m.States) > 16 || len(m.Guards) < 1 || len(m.Guards) > 64 || len(m.Requirements) < 1 || len(m.Requirements) > 64 || len(m.Transitions) < 1 || len(m.Transitions) > 32 {
		return invalid("collection bounds")
	}
	ids := map[string]bool{}
	add := func(id string) bool {
		if !Identifier(id) || ids[id] {
			return false
		}
		ids[id] = true
		return true
	}
	states := map[string]State{}
	requirements := map[string]bool{}
	guards := map[string]bool{}
	for _, s := range m.States {
		if !add(s.ID) || !text(s.Label, 256, true) {
			return invalid("state")
		}
		states[s.ID] = s
	}
	if _, ok := states[m.Initial]; !ok {
		return invalid("initial state")
	}
	for _, r := range m.Requirements {
		if !add(r.ID) || !text(r.Text, 4096, true) {
			return invalid("requirement")
		}
		requirements[r.ID] = true
	}
	count := 0
	for _, g := range m.Guards {
		if !add(g.ID) || !requirements[g.Requirement] {
			return invalid("guard identity/requirement")
		}
		typ, e := exprType(g.Expr, 0, &count)
		if e != nil {
			return e
		}
		if typ != "bool" {
			return invalid("guard not boolean")
		}
		guards[g.ID] = true
	}
	pairs := map[string]bool{}
	out := map[string]int{}
	edges := map[string][]string{}
	usedGuards := map[string]bool{}
	usedRequirements := map[string]bool{}
	for _, t := range m.Transitions {
		from, fok := states[t.From]
		_, tok := states[t.To]
		if !add(t.ID) || !fok || !tok || from.Terminal || !Identifier(t.Trigger) {
			return invalid("transition identity/state")
		}
		pair := t.From + "\x00" + t.Trigger
		if pairs[pair] {
			return invalid("ambiguous state/trigger")
		}
		pairs[pair] = true
		if len(t.Guards) < 1 || len(t.Guards) > 64 {
			return invalid("transition guards")
		}
		seen := map[string]bool{}
		for _, g := range t.Guards {
			if !guards[g] || seen[g] {
				return invalid("transition guard reference")
			}
			seen[g] = true
			usedGuards[g] = true
		}
		effects := t.Effects
		if len(effects) == 3 {
			if effects[0].Op != AssignOwner {
				return invalid("assignment effect/order")
			}
			effects = effects[1:]
		}
		if len(effects) != 2 || effects[0].Op != SetState || effects[1].Op != IncrementRevision {
			return invalid("effects/order")
		}
		out[t.From]++
		edges[t.From] = append(edges[t.From], t.To)
	}
	terminals := 0
	for _, s := range m.States {
		if s.Terminal {
			terminals++
		} else if out[s.ID] == 0 {
			return invalid("nonterminal dead end")
		}
	}
	if terminals == 0 {
		return invalid("missing terminal")
	}
	reached := map[string]bool{m.Initial: true}
	queue := []string{m.Initial}
	for len(queue) > 0 {
		n := queue[0]
		queue = queue[1:]
		for _, to := range edges[n] {
			if !reached[to] {
				reached[to] = true
				queue = append(queue, to)
			}
		}
	}
	if len(reached) != len(states) {
		return invalid("unreachable state")
	}
	for _, g := range m.Guards {
		if !usedGuards[g.ID] {
			return invalid("unused guard")
		}
		usedRequirements[g.Requirement] = true
	}
	for _, r := range m.Requirements {
		if !usedRequirements[r.ID] {
			return invalid("unused requirement")
		}
	}
	return nil
}
func normalizeExpr(e Expr) Expr {
	e.Args = append([]Expr(nil), e.Args...)
	for i := range e.Args {
		e.Args[i] = normalizeExpr(e.Args[i])
	}
	if e.Kind == "op" && (e.Op == "all" || e.Op == "any") {
		sort.Slice(e.Args, func(i, j int) bool { return ExprKey(e.Args[i]) < ExprKey(e.Args[j]) })
	}
	return e
}

// Normalize copies input; only semantic sets are reordered. Effects stay ordered.
func Normalize(m Model) (Model, error) {
	if err := Validate(m); err != nil {
		return Model{}, err
	}
	m.States = append([]State(nil), m.States...)
	m.Requirements = append([]Requirement(nil), m.Requirements...)
	m.Guards = append([]Guard(nil), m.Guards...)
	m.Transitions = append([]Transition(nil), m.Transitions...)
	for i := range m.Guards {
		m.Guards[i].Expr = normalizeExpr(m.Guards[i].Expr)
	}
	for i := range m.Transitions {
		t := &m.Transitions[i]
		t.Guards = append([]string(nil), t.Guards...)
		sort.Strings(t.Guards)
		t.Effects = append([]Effect(nil), t.Effects...)
	}
	sort.Slice(m.States, func(i, j int) bool { return m.States[i].ID < m.States[j].ID })
	sort.Slice(m.Requirements, func(i, j int) bool { return m.Requirements[i].ID < m.Requirements[j].ID })
	sort.Slice(m.Guards, func(i, j int) bool { return m.Guards[i].ID < m.Guards[j].ID })
	sort.Slice(m.Transitions, func(i, j int) bool { return m.Transitions[i].ID < m.Transitions[j].ID })
	// Recheck normalized duplicate boolean operands, including nested set order.
	if err := Validate(m); err != nil {
		return Model{}, err
	}
	return m, nil
}
