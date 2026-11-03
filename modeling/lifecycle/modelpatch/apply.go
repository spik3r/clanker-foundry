package modelpatch

import (
	"unicode/utf8"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/canonical"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/impact"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/jsonformat"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/model"
)

func scalar(s string, max int) error {
	if len(s) > max {
		return ErrLimit
	}
	if !utf8.ValidString(s) {
		return ErrInvalid
	}
	return nil
}
func identifier(s string) error {
	if e := scalar(s, MaxIdentifierBytes); e != nil {
		return e
	}
	if !model.Identifier(s) {
		return ErrInvalid
	}
	return nil
}
func digestSyntax(s string) bool {
	if len(s) != 64 {
		return false
	}
	for _, b := range []byte(s) {
		if !(b >= '0' && b <= '9' || b >= 'a' && b <= 'f') {
			return false
		}
	}
	return true
}
func union(o Operation) error {
	if o.Action != Add && o.Action != Replace && o.Action != Remove {
		return ErrInvalid
	}
	selected := false
	switch o.Kind {
	case Model:
		selected = o.Root != nil
	case State:
		selected = o.State != nil
	case Requirement:
		selected = o.Requirement != nil
	case Guard:
		selected = o.Guard != nil
	case Transition:
		selected = o.Transition != nil
	default:
		return ErrInvalid
	}
	count := 0
	for _, yes := range []bool{o.Root != nil, o.State != nil, o.Requirement != nil, o.Guard != nil, o.Transition != nil} {
		if yes {
			count++
		}
	}
	if o.Kind == Model {
		if o.ID != "" || o.Action != Replace {
			return ErrInvalid
		}
	} else {
		if e := identifier(o.ID); e != nil {
			return e
		}
	}
	if o.Action == Remove {
		if count != 0 {
			return ErrInvalid
		}
		return nil
	}
	if !selected || count != 1 {
		return ErrInvalid
	}
	return nil
}

// Shape checks no profile blankness, guard-list uniqueness, effects semantics or
// graph existence. Those rules belong to the final existing profile validator.
func operationBase(o Operation) error {
	if e := union(o); e != nil {
		return e
	}
	if o.Action == Remove {
		return nil
	}
	switch o.Kind {
	case Model:
		if e := identifier(o.Root.ID); e != nil {
			return e
		}
		return identifier(o.Root.Initial)
	case State:
		if e := identifier(o.State.ID); e != nil {
			return e
		}
		if o.State.ID != o.ID {
			return ErrInvalid
		}
		return scalar(o.State.Label, MaxStateLabelBytes)
	case Requirement:
		if e := identifier(o.Requirement.ID); e != nil {
			return e
		}
		if o.Requirement.ID != o.ID {
			return ErrInvalid
		}
		return scalar(o.Requirement.Text, MaxRecordTextBytes)
	case Guard:
		if e := identifier(o.Guard.ID); e != nil {
			return e
		}
		if o.Guard.ID != o.ID {
			return ErrInvalid
		}
		return identifier(o.Guard.Requirement)
	case Transition:
		t := o.Transition
		if e := transitionHead(t.ID, t.From, t.Trigger, t.To, o.ID, len(t.Guards), len(t.Effects)); e != nil {
			return e
		}
		for _, id := range t.Guards {
			if e := identifier(id); e != nil {
				return e
			}
		}
		for _, effect := range t.Effects {
			if e := scalar(effect.Op, MaxEffectOpBytes); e != nil {
				return e
			}
		}
	}
	return nil
}

// Shared head validation lets the wire adapter bound effect allocation without
// changing error precedence relative to the typed entrypoint.
func transitionHead(id, from, trigger, to, target string, guards, effects int) error {
	if e := identifier(id); e != nil {
		return e
	}
	if id != target {
		return ErrInvalid
	}
	for _, s := range []string{from, trigger, to} {
		if e := identifier(s); e != nil {
			return e
		}
	}
	if guards > MaxGuardIDs || effects > MaxEffects {
		return ErrLimit
	}
	return nil
}
func exprBounds(e model.Expr, depth int, nodes *int) error {
	*nodes++
	if depth > MaxExprDepth || *nodes > MaxExprNodes || len(e.Args) > MaxExprArgs {
		return ErrLimit
	}
	for _, x := range []struct {
		s   string
		max int
	}{{e.Kind, MaxExprAtomBytes}, {e.Ref, MaxExprRefBytes}, {e.Value, MaxRecordTextBytes}, {e.Op, MaxExprAtomBytes}} {
		if err := scalar(x.s, x.max); err != nil {
			return err
		}
	}
	for _, a := range e.Args {
		if err := exprBounds(a, depth+1, nodes); err != nil {
			return err
		}
	}
	return nil
}
func preflight(p Patch) error {
	if p.Version != 1 || p.ProfileVersion != ProfileVersion || !digestSyntax(p.BeforeDigest) || p.Operations == nil {
		return ErrInvalid
	}
	if len(p.Operations) > MaxOperations {
		return ErrLimit
	}
	nodes := 0
	// Complete shape/resource bounding precedes recursive DTO copying or TypeOf.
	for _, o := range p.Operations {
		if e := operationBase(o); e != nil {
			return e
		}
		if o.Guard != nil {
			if e := exprBounds(o.Guard.Expr, 0, &nodes); e != nil {
				return e
			}
		}
	}
	for _, o := range p.Operations {
		if o.Guard != nil {
			typ, e := model.TypeOf(o.Guard.Expr)
			if e != nil || typ != "bool" {
				return ErrInvalid
			}
		}
	}
	b, e := patchBytes(p)
	if e != nil {
		return e
	}
	if len(b) > MaxPatchBytes {
		return ErrLimit
	}
	seen := map[struct {
		k  Kind
		id string
	}]bool{}
	for _, o := range p.Operations {
		key := struct {
			k  Kind
			id string
		}{o.Kind, o.ID}
		if seen[key] {
			return ErrConflict
		}
		seen[key] = true
	}
	return nil
}
func copyExpr(e model.Expr) model.Expr {
	if e.Args != nil {
		a := make([]model.Expr, len(e.Args))
		for i, x := range e.Args {
			a[i] = copyExpr(x)
		}
		e.Args = a
	}
	return e
}
func edit[T any](items []T, o Operation, value T, id func(T) string) ([]T, error) {
	index := -1
	for i, x := range items {
		if id(x) == o.ID {
			index = i
			break
		}
	}
	if o.Action == Add {
		if index >= 0 {
			return nil, ErrConflict
		}
		return append(items, value), nil
	}
	if index < 0 {
		return nil, ErrConflict
	}
	if o.Action == Replace {
		items[index] = value
		return items, nil
	}
	return append(items[:index], items[index+1:]...), nil
}
func apply(m model.Model, o Operation) (model.Model, error) {
	var e error
	switch o.Kind {
	case Model:
		m.ID = o.Root.ID
		m.Initial = o.Root.Initial
	case State:
		v := model.State{}
		if o.State != nil {
			v = *o.State
		}
		m.States, e = edit(m.States, o, v, func(x model.State) string { return x.ID })
	case Requirement:
		v := model.Requirement{}
		if o.Requirement != nil {
			v = *o.Requirement
		}
		m.Requirements, e = edit(m.Requirements, o, v, func(x model.Requirement) string { return x.ID })
	case Guard:
		v := model.Guard{}
		if o.Guard != nil {
			v = *o.Guard
			v.Expr = copyExpr(v.Expr)
		}
		m.Guards, e = edit(m.Guards, o, v, func(x model.Guard) string { return x.ID })
	case Transition:
		v := model.Transition{}
		if o.Transition != nil {
			v = *o.Transition
			v.Guards = append([]string{}, v.Guards...)
			v.Effects = append([]model.Effect{}, v.Effects...)
		}
		m.Transitions, e = edit(m.Transitions, o, v, func(x model.Transition) string { return x.ID })
	}
	return m, e
}

// Plan proposes a normalized owned candidate. It executes no model behavior and
// gives no permission, freshness/authentication or scenario/review binding claim.
func Plan(before model.Model, p Patch) (Result, error) {
	working, e := model.Normalize(before)
	if e != nil {
		return Result{}, ErrInvalid
	}
	if _, e = jsonformat.Encode(working); e != nil {
		return Result{}, ErrInvalid
	}
	if e = preflight(p); e != nil {
		return Result{}, e
	}
	beforeDigest, e := canonical.Digest(working)
	if e != nil {
		return Result{}, ErrInvalid
	}
	if beforeDigest != p.BeforeDigest {
		return Result{}, ErrPrecondition
	}
	for _, o := range p.Operations {
		working, e = apply(working, o)
		if e != nil {
			return Result{}, e
		}
	}
	after, e := model.Normalize(working)
	if e != nil {
		return Result{}, ErrInvalid
	}
	if _, e = jsonformat.Encode(after); e != nil {
		return Result{}, ErrInvalid
	}
	afterDigest, e := canonical.Digest(after)
	if e != nil {
		return Result{}, ErrInvalid
	}
	review, e := impact.Compare(before, after, nil)
	if e != nil {
		return Result{}, ErrInvalid
	}
	result := Result{ReportVersion: ReportVersion, ProfileVersion: ProfileVersion, Mode: Mode, Authority: Authority, BeforeDigest: beforeDigest, AfterDigest: afterDigest, After: after, Review: review}
	if _, e = encodeResult(result); e != nil {
		return Result{}, e
	}
	return result, nil
}
