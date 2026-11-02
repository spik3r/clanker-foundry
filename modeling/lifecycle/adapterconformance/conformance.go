// Package adapterconformance provides optional, finite adapter probes for tests.
// It does not run reducers or establish arbitrary adapter or source correctness.
package adapterconformance

import (
	"fmt"
	"math"
	"reflect"

	transport "github.com/spik3r/clanker-foundry/modeling/lifecycle/generated"
)

// Adapter supplies the four field conversions used by a generated-model consumer.
// Callbacks are trusted test code: Check does not sandbox or cancel them, recover
// their panics, or constrain their effects. Do not use this as runtime admission.
type Adapter[T, C, R any] struct {
	ToTask     func(transport.Task) T
	FromTask   func(T) transport.Task
	ToCommand  func(transport.Command) C
	FromResult func(R) transport.Result
}

type field struct {
	name string
	typ  reflect.Type
	kind reflect.Kind // used only when typ is nil
}

func exact(name string, typ reflect.Type) field { return field{name: name, typ: typ} }
func inventory(label string, typ reflect.Type, fields []field) error {
	if typ.Kind() != reflect.Struct || typ.NumField() != len(fields) {
		return fmt.Errorf("%s: field inventory differs", label)
	}
	for i, want := range fields {
		got := typ.Field(i)
		if got.Name != want.name || got.Anonymous || !got.IsExported() ||
			(want.typ != nil && got.Type != want.typ) ||
			(want.typ == nil && got.Type.Kind() != want.kind) {
			return fmt.Errorf("%s.%s: field inventory differs", label, want.name)
		}
	}
	return nil
}

func taskInventory(label string, typ reflect.Type) error {
	return inventory(label, typ, []field{
		exact("ID", reflect.TypeFor[string]()), {name: "State", kind: reflect.String},
		exact("Revision", reflect.TypeFor[uint64]()), exact("Outcome", reflect.TypeFor[string]()),
		exact("Acceptance", reflect.TypeFor[[]string]()), exact("Owner", reflect.TypeFor[*string]()),
	})
}
func commandInventory(label string, typ reflect.Type) error {
	if err := inventory(label, typ, []field{
		exact("Type", reflect.TypeFor[string]()), exact("Actor", reflect.TypeFor[string]()),
		exact("ExpectedRevision", reflect.TypeFor[uint64]()), {name: "Review", kind: reflect.Pointer},
	}); err != nil {
		return err
	}
	return inventory(label+".Review", typ.Field(3).Type.Elem(), []field{
		exact("TaskID", reflect.TypeFor[string]()), exact("TaskRevision", reflect.TypeFor[uint64]()),
		exact("ModelDigest", reflect.TypeFor[string]()), exact("TaskDigest", reflect.TypeFor[string]()),
		exact("Reviewer", reflect.TypeFor[string]()), exact("Decision", reflect.TypeFor[string]()),
	})
}
func resultInventory(label string, typ, task reflect.Type) error {
	return inventory(label, typ, []field{
		exact("Mode", reflect.TypeFor[string]()), exact("Authority", reflect.TypeFor[string]()),
		exact("ModelDigest", reflect.TypeFor[string]()), exact("OK", reflect.TypeFor[bool]()),
		exact("TransitionID", reflect.TypeFor[string]()), exact("Task", reflect.PointerTo(task)),
		exact("Refusals", reflect.TypeFor[[]string]()),
	})
}

// literalTask creates fresh, named values without calling either adapter. Empty
// lists have spare capacity so an alias cannot hide behind a zero length.
func literalTask(revision uint64, list, owner int) transport.Task {
	t := transport.Task{ID: "task-id", State: "invented-state", Revision: revision, Outcome: "task-outcome<&>"}
	switch list {
	case 1:
		t.Acceptance = make([]string, 0, 2)
	case 2:
		t.Acceptance = []string{"criterion-first", "criterion-second"}
	}
	if owner != 0 {
		s := ""
		if owner == 2 {
			s = "task-owner"
		}
		t.Owner = &s
	}
	return t
}

// namedTask builds the independently named custom-side fixture. It is not a
// conversion under test and never obtains expectations from the inverse adapter.
func namedTask[T any](x transport.Task) T {
	v := reflect.New(reflect.TypeFor[T]()).Elem()
	v.FieldByName("ID").SetString(x.ID)
	v.FieldByName("State").SetString(string(x.State))
	v.FieldByName("Revision").SetUint(x.Revision)
	v.FieldByName("Outcome").SetString(x.Outcome)
	v.FieldByName("Acceptance").Set(reflect.ValueOf(x.Acceptance))
	v.FieldByName("Owner").Set(reflect.ValueOf(x.Owner))
	return v.Interface().(T)
}

func literalCommand(revision uint64, review int) transport.Command {
	c := transport.Command{Type: "command-trigger", Actor: "command-actor", ExpectedRevision: revision}
	if review != 0 {
		// Distinct revisions expose a swap that two equal MaxUint64 values hide.
		r := uint64(math.MaxUint64)
		if revision == math.MaxUint64 {
			r = 0
		}
		c.Review = &transport.Review{TaskID: "review-task", TaskRevision: r, ModelDigest: "actual-review-model", TaskDigest: "review-content", Reviewer: "review-actor", Decision: "review-decision"}
		if review == 2 {
			c.Review = &transport.Review{}
		}
	}
	return c
}
func namedCommand[C any](x transport.Command) C {
	v := reflect.New(reflect.TypeFor[C]()).Elem()
	v.FieldByName("Type").SetString(x.Type)
	v.FieldByName("Actor").SetString(x.Actor)
	v.FieldByName("ExpectedRevision").SetUint(x.ExpectedRevision)
	if x.Review != nil {
		r := reflect.New(v.FieldByName("Review").Type().Elem())
		r.Elem().FieldByName("TaskID").SetString(x.Review.TaskID)
		r.Elem().FieldByName("TaskRevision").SetUint(x.Review.TaskRevision)
		r.Elem().FieldByName("ModelDigest").SetString(x.Review.ModelDigest)
		r.Elem().FieldByName("TaskDigest").SetString(x.Review.TaskDigest)
		r.Elem().FieldByName("Reviewer").SetString(x.Review.Reviewer)
		r.Elem().FieldByName("Decision").SetString(x.Review.Decision)
		v.FieldByName("Review").Set(r)
	}
	return v.Interface().(C)
}
func literalResult(ok bool, task *transport.Task, refusals int) transport.Result {
	r := transport.Result{Mode: "actual-mode", Authority: "actual-authority", ModelDigest: "actual-model-digest", OK: ok, TransitionID: "actual-transition", Task: task}
	switch refusals {
	case 1:
		r.Refusals = make([]string, 0, 2)
	case 2:
		r.Refusals = []string{"z-refusal", "a-refusal", "z-refusal"}
	}
	return r
}
func namedResult[T, R any](x transport.Result) R {
	v := reflect.New(reflect.TypeFor[R]()).Elem()
	v.FieldByName("Mode").SetString(x.Mode)
	v.FieldByName("Authority").SetString(x.Authority)
	v.FieldByName("ModelDigest").SetString(x.ModelDigest)
	v.FieldByName("OK").SetBool(x.OK)
	v.FieldByName("TransitionID").SetString(x.TransitionID)
	v.FieldByName("Refusals").Set(reflect.ValueOf(x.Refusals))
	if x.Task != nil {
		t := namedTask[T](*x.Task)
		v.FieldByName("Task").Set(reflect.ValueOf(&t))
	}
	return v.Interface().(R)
}

// detached checks mutable storage in both directions after value equality. It
// mutates only the newly constructed probes and restores their original values.
// Walking the declared, finite shape also covers nested task/review pointers.
func detached(a, b reflect.Value) bool {
	switch a.Kind() {
	case reflect.Struct:
		for i := 0; i < a.NumField(); i++ {
			if !detached(a.Field(i), b.Field(i)) {
				return false
			}
		}
	case reflect.Pointer:
		if !a.IsNil() && !b.IsNil() {
			if a.Pointer() == b.Pointer() {
				return false
			}
			return detached(a.Elem(), b.Elem())
		}
	case reflect.Slice:
		if a.Cap() != 0 && b.Cap() != 0 {
			x, y := a.Slice(0, a.Cap()), b.Slice(0, b.Cap())
			// Include overlapping empty subslices with different starting
			// offsets; equal length and first-element checks alone miss them.
			xp, yp, size := x.Pointer(), y.Pointer(), x.Type().Elem().Size()
			if (xp <= yp && yp-xp < uintptr(x.Cap())*size) ||
				(yp <= xp && xp-yp < uintptr(y.Cap())*size) {
				return false
			}
			// Probes have distinct first/second entries. First-element mutation
			// also detects shared empty capacity without requiring equal caps.
			x0, y0 := x.Index(0).String(), y.Index(0).String()
			x.Index(0).SetString("mutated-source")
			forward := y.Index(0).String() == y0
			x.Index(0).SetString(x0)
			y.Index(0).SetString("mutated-result")
			backward := x.Index(0).String() == x0
			y.Index(0).SetString(y0)
			return forward && backward
		}
	}
	return true
}

func unchanged(label string, got, want any) error {
	if !reflect.DeepEqual(got, want) {
		return fmt.Errorf("%s: named fields or nil/empty shape differ", label)
	}
	return nil
}
func owned(label string, input, output any) error {
	if !detached(reflect.ValueOf(input), reflect.ValueOf(output)) {
		return fmt.Errorf("%s: mutable storage aliases", label)
	}
	return nil
}

// Check runs fixed one-way field and ownership probes, not scenario acceptance.
// It returns the first error after complete shape/nil-callback preflight. A nil
// error says only that these finite probes passed, not that the adapter is right
// for arbitrary inputs or that any claimed model identity has trusted provenance.
func Check[T, C, R any](a Adapter[T, C, R]) error {
	for _, err := range []error{
		taskInventory("transport.Task", reflect.TypeFor[transport.Task]()),
		commandInventory("transport.Command", reflect.TypeFor[transport.Command]()),
		resultInventory("transport.Result", reflect.TypeFor[transport.Result](), reflect.TypeFor[transport.Task]()),
		taskInventory("Task", reflect.TypeFor[T]()), commandInventory("Command", reflect.TypeFor[C]()),
		resultInventory("Result", reflect.TypeFor[R](), reflect.TypeFor[T]()),
	} {
		if err != nil {
			return err
		}
	}
	if a.ToTask == nil || a.FromTask == nil || a.ToCommand == nil || a.FromResult == nil {
		return fmt.Errorf("adapter callback is nil")
	}
	for _, revision := range []uint64{0, math.MaxUint64} {
		for list := 0; list < 3; list++ {
			for owner := 0; owner < 3; owner++ {
				input := literalTask(revision, list, owner)
				got := a.ToTask(input)
				if err := unchanged("ToTask input", input, literalTask(revision, list, owner)); err != nil {
					return err
				}
				if err := unchanged("ToTask", got, namedTask[T](literalTask(revision, list, owner))); err != nil {
					return err
				}
				if err := owned("ToTask", input, got); err != nil {
					return err
				}

				source := namedTask[T](literalTask(revision, list, owner))
				out := a.FromTask(source)
				if err := unchanged("FromTask input", source, namedTask[T](literalTask(revision, list, owner))); err != nil {
					return err
				}
				if err := unchanged("FromTask", out, literalTask(revision, list, owner)); err != nil {
					return err
				}
				if err := owned("FromTask", source, out); err != nil {
					return err
				}

				for _, ok := range []bool{false, true} {
					for refusals := 0; refusals < 3; refusals++ {
						makeResult := func() transport.Result {
							t := literalTask(revision, list, owner)
							return literalResult(ok, &t, refusals)
						}
						if err := checkResult[T](a.FromResult, makeResult); err != nil {
							return err
						}
					}
				}
			}
		}
		for review := 0; review < 3; review++ {
			input := literalCommand(revision, review)
			got := a.ToCommand(input)
			if err := unchanged("ToCommand input", input, literalCommand(revision, review)); err != nil {
				return err
			}
			if err := unchanged("ToCommand", got, namedCommand[C](literalCommand(revision, review))); err != nil {
				return err
			}
			if err := owned("ToCommand", input, got); err != nil {
				return err
			}
		}
	}
	for _, ok := range []bool{false, true} {
		for refusals := 0; refusals < 3; refusals++ {
			if err := checkResult[T](a.FromResult, func() transport.Result { return literalResult(ok, nil, refusals) }); err != nil {
				return err
			}
		}
	}
	return nil
}

func checkResult[T, R any](from func(R) transport.Result, literal func() transport.Result) error {
	input := namedResult[T, R](literal())
	got := from(input)
	if err := unchanged("FromResult input", input, namedResult[T, R](literal())); err != nil {
		return err
	}
	if err := unchanged("FromResult", got, literal()); err != nil {
		return err
	}
	return owned("FromResult", input, got)
}
