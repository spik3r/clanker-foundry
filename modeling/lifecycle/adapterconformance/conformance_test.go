// Independently specified before adapterconformance logic. No reducer is invoked.
package adapterconformance_test

import (
	ac "github.com/spik3r/clanker-foundry/modeling/lifecycle/adapterconformance"
	transport "github.com/spik3r/clanker-foundry/modeling/lifecycle/generated"
	"math"
	"reflect"
	"sort"
	"strings"
	"testing"
)

type sampleState string
type sampleTask struct {
	ID         string
	State      sampleState
	Revision   uint64
	Outcome    string
	Acceptance []string
	Owner      *string
}
type sampleReview struct {
	TaskID       string
	TaskRevision uint64
	ModelDigest  string
	TaskDigest   string
	Reviewer     string
	Decision     string
}
type sampleCommand struct {
	Type             string
	Actor            string
	ExpectedRevision uint64
	Review           *sampleReview
}
type sampleResult struct {
	Mode         string
	Authority    string
	ModelDigest  string
	OK           bool
	TransitionID string
	Task         *sampleTask
	Refusals     []string
}

func copyStrings(x []string) []string {
	if x == nil {
		return nil
	}
	return append([]string{}, x...)
}
func copyOwner(x *string) *string {
	if x == nil {
		return nil
	}
	y := *x
	return &y
}
func fieldOnlyToTask(x transport.Task) sampleTask {
	return sampleTask{ID: x.ID, State: sampleState(x.State), Revision: x.Revision, Outcome: x.Outcome, Acceptance: copyStrings(x.Acceptance), Owner: copyOwner(x.Owner)}
}
func fieldOnlyFromTask(x sampleTask) transport.Task {
	return transport.Task{ID: x.ID, State: transport.State(x.State), Revision: x.Revision, Outcome: x.Outcome, Acceptance: copyStrings(x.Acceptance), Owner: copyOwner(x.Owner)}
}
func fieldOnlyToCommand(x transport.Command) sampleCommand {
	y := sampleCommand{Type: x.Type, Actor: x.Actor, ExpectedRevision: x.ExpectedRevision}
	if x.Review != nil {
		r := x.Review
		y.Review = &sampleReview{TaskID: r.TaskID, TaskRevision: r.TaskRevision, ModelDigest: r.ModelDigest, TaskDigest: r.TaskDigest, Reviewer: r.Reviewer, Decision: r.Decision}
	}
	return y
}
func fieldOnlyFromResult(x sampleResult) transport.Result {
	y := transport.Result{Mode: x.Mode, Authority: x.Authority, ModelDigest: x.ModelDigest, OK: x.OK, TransitionID: x.TransitionID, Refusals: copyStrings(x.Refusals)}
	if x.Task != nil {
		t := fieldOnlyFromTask(*x.Task)
		y.Task = &t
	}
	return y
}
func good() ac.Adapter[sampleTask, sampleCommand, sampleResult] {
	return ac.Adapter[sampleTask, sampleCommand, sampleResult]{ToTask: fieldOnlyToTask, FromTask: fieldOnlyFromTask, ToCommand: fieldOnlyToCommand, FromResult: fieldOnlyFromResult}
}
func refuse(t *testing.T, a ac.Adapter[sampleTask, sampleCommand, sampleResult]) {
	t.Helper()
	if err := ac.Check(a); err == nil {
		t.Fatal("non-conforming adapter accepted")
	}
}
func TestFieldOnlyCloneAccepted(t *testing.T) {
	if e := ac.Check(good()); e != nil {
		t.Fatal(e)
	}
}
func TestCoupledSwapCannotHideBehindRoundTrip(t *testing.T) {
	a := good()
	a.ToTask = func(x transport.Task) sampleTask {
		y := fieldOnlyToTask(x)
		y.ID, y.Outcome = y.Outcome, y.ID
		return y
	}
	a.FromTask = func(x sampleTask) transport.Task {
		y := fieldOnlyFromTask(x)
		y.ID, y.Outcome = y.Outcome, y.ID
		return y
	}
	refuse(t, a)
}

// Mutate a single named leaf in a returned value. Pointer absence is deliberately
// left alone: Check must supply non-nil specimens as well as nil specimens.
func corrupt(v reflect.Value, path []string) {
	for v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return
		}
		v = v.Elem()
	}
	f := v.FieldByName(path[0])
	if len(path) > 1 {
		corrupt(f, path[1:])
		return
	}
	switch f.Kind() {
	case reflect.String:
		f.SetString(f.String() + "!corruption")
	case reflect.Uint64:
		f.SetUint(f.Uint() ^ 1)
	case reflect.Bool:
		f.SetBool(!f.Bool())
	default:
		f.Set(reflect.Zero(f.Type()))
	}
}
func TestEveryNamedFieldIsCheckedIndependently(t *testing.T) {
	tasks := [][]string{{"ID"}, {"State"}, {"Revision"}, {"Outcome"}, {"Acceptance"}, {"Owner"}}
	commands := [][]string{{"Type"}, {"Actor"}, {"ExpectedRevision"}, {"Review"}, {"Review", "TaskID"}, {"Review", "TaskRevision"}, {"Review", "ModelDigest"}, {"Review", "TaskDigest"}, {"Review", "Reviewer"}, {"Review", "Decision"}}
	results := [][]string{{"Mode"}, {"Authority"}, {"ModelDigest"}, {"OK"}, {"TransitionID"}, {"Refusals"}, {"Task"}, {"Task", "ID"}, {"Task", "State"}, {"Task", "Revision"}, {"Task", "Outcome"}, {"Task", "Acceptance"}, {"Task", "Owner"}}
	for _, p := range tasks {
		t.Run("ToTask/"+p[0], func(t *testing.T) {
			a := good()
			a.ToTask = func(x transport.Task) sampleTask {
				y := fieldOnlyToTask(x)
				corrupt(reflect.ValueOf(&y).Elem(), p)
				return y
			}
			refuse(t, a)
		})
		t.Run("FromTask/"+p[0], func(t *testing.T) {
			a := good()
			a.FromTask = func(x sampleTask) transport.Task {
				y := fieldOnlyFromTask(x)
				corrupt(reflect.ValueOf(&y).Elem(), p)
				return y
			}
			refuse(t, a)
		})
	}
	for _, p := range commands {
		t.Run("ToCommand/"+p[0]+"/"+p[len(p)-1], func(t *testing.T) {
			a := good()
			a.ToCommand = func(x transport.Command) sampleCommand {
				y := fieldOnlyToCommand(x)
				corrupt(reflect.ValueOf(&y).Elem(), p)
				return y
			}
			refuse(t, a)
		})
	}
	for _, p := range results {
		t.Run("FromResult/"+p[0]+"/"+p[len(p)-1], func(t *testing.T) {
			a := good()
			a.FromResult = func(x sampleResult) transport.Result {
				y := fieldOnlyFromResult(x)
				corrupt(reflect.ValueOf(&y).Elem(), p)
				return y
			}
			refuse(t, a)
		})
	}
}
func TestNilEmptyAndNarrowing(t *testing.T) {
	for _, direction := range []string{"ToTask", "FromTask", "FromResult"} {
		for _, nilToEmpty := range []bool{true, false} {
			t.Run(direction+map[bool]string{true: "/nil-to-empty", false: "/empty-to-nil"}[nilToEmpty], func(t *testing.T) {
				a := good()
				change := func(x []string) []string {
					if nilToEmpty && x == nil {
						return []string{}
					}
					if !nilToEmpty && x != nil && len(x) == 0 {
						return nil
					}
					return x
				}
				switch direction {
				case "ToTask":
					a.ToTask = func(x transport.Task) sampleTask {
						y := fieldOnlyToTask(x)
						y.Acceptance = change(y.Acceptance)
						return y
					}
				case "FromTask":
					a.FromTask = func(x sampleTask) transport.Task {
						y := fieldOnlyFromTask(x)
						y.Acceptance = change(y.Acceptance)
						return y
					}
				case "FromResult":
					a.FromResult = func(x sampleResult) transport.Result {
						y := fieldOnlyFromResult(x)
						y.Refusals = change(y.Refusals)
						return y
					}
				}
				refuse(t, a)
			})
		}
	}
	for _, field := range []string{"task", "command", "review", "result-task"} {
		t.Run("narrow/"+field, func(t *testing.T) {
			a := good()
			switch field {
			case "task":
				a.ToTask = func(x transport.Task) sampleTask {
					y := fieldOnlyToTask(x)
					y.Revision = uint64(uint32(y.Revision))
					return y
				}
			case "command":
				a.ToCommand = func(x transport.Command) sampleCommand {
					y := fieldOnlyToCommand(x)
					y.ExpectedRevision = uint64(uint32(y.ExpectedRevision))
					return y
				}
			case "review":
				a.ToCommand = func(x transport.Command) sampleCommand {
					y := fieldOnlyToCommand(x)
					if y.Review != nil {
						y.Review.TaskRevision = uint64(uint32(y.Review.TaskRevision))
					}
					return y
				}
			case "result-task":
				a.FromResult = func(x sampleResult) transport.Result {
					y := fieldOnlyFromResult(x)
					if y.Task != nil {
						y.Task.Revision = uint64(uint32(y.Task.Revision))
					}
					return y
				}
			}
			refuse(t, a)
		})
	}
}
func TestRefusalsAreOrderedWithMultiplicity(t *testing.T) {
	for _, dedup := range []bool{false, true} {
		t.Run(map[bool]string{true: "dedup", false: "sort"}[dedup], func(t *testing.T) {
			a := good()
			a.FromResult = func(x sampleResult) transport.Result {
				y := fieldOnlyFromResult(x)
				if !dedup {
					sort.Strings(y.Refusals)
				} else if y.Refusals != nil {
					out := []string{}
					seen := map[string]bool{}
					for _, s := range y.Refusals {
						if !seen[s] {
							out = append(out, s)
							seen[s] = true
						}
					}
					y.Refusals = out
				}
				return y
			}
			refuse(t, a)
		})
	}
}
func TestNestedAliasingRefused(t *testing.T) {
	for _, kind := range []string{"to-slice", "to-owner", "from-slice", "from-owner", "review", "refusals", "result-slice", "result-owner"} {
		t.Run(kind, func(t *testing.T) {
			a := good()
			switch kind {
			case "to-slice":
				a.ToTask = func(x transport.Task) sampleTask { y := fieldOnlyToTask(x); y.Acceptance = x.Acceptance; return y }
			case "to-owner":
				a.ToTask = func(x transport.Task) sampleTask { y := fieldOnlyToTask(x); y.Owner = x.Owner; return y }
			case "from-slice":
				a.FromTask = func(x sampleTask) transport.Task { y := fieldOnlyFromTask(x); y.Acceptance = x.Acceptance; return y }
			case "from-owner":
				a.FromTask = func(x sampleTask) transport.Task { y := fieldOnlyFromTask(x); y.Owner = x.Owner; return y }
			case "review":
				a.ToCommand = func(x transport.Command) sampleCommand {
					y := fieldOnlyToCommand(x)
					y.Review = (*sampleReview)(x.Review)
					return y
				}
			case "refusals":
				a.FromResult = func(x sampleResult) transport.Result { y := fieldOnlyFromResult(x); y.Refusals = x.Refusals; return y }
			case "result-slice":
				a.FromResult = func(x sampleResult) transport.Result {
					y := fieldOnlyFromResult(x)
					if y.Task != nil {
						y.Task.Acceptance = x.Task.Acceptance
					}
					return y
				}
			case "result-owner":
				a.FromResult = func(x sampleResult) transport.Result {
					y := fieldOnlyFromResult(x)
					if y.Task != nil {
						y.Task.Owner = x.Task.Owner
					}
					return y
				}
			}
			refuse(t, a)
		})
	}
}

// Transport-shaped values remain legal; copying a result's task pointer is not.
func TestResultTaskPointerAliasingRefused(t *testing.T) {
	a := ac.Adapter[transport.Task, transport.Command, transport.Result]{ToTask: func(x transport.Task) transport.Task { return fieldOnlyFromTask(fieldOnlyToTask(x)) }, FromTask: func(x transport.Task) transport.Task { return fieldOnlyFromTask(fieldOnlyToTask(x)) }, ToCommand: func(x transport.Command) transport.Command {
		y := x
		if x.Review != nil {
			r := *x.Review
			y.Review = &r
		}
		return y
	}, FromResult: func(x transport.Result) transport.Result { y := x; y.Refusals = copyStrings(x.Refusals); return y }}
	if ac.Check(a) == nil {
		t.Fatal("result task pointer aliases")
	}
}
func TestNilCallbacksPreflight(t *testing.T) {
	for _, which := range []string{"ToTask", "FromTask", "ToCommand", "FromResult"} {
		t.Run(which, func(t *testing.T) {
			calls := 0
			a := good()
			a.ToTask = func(x transport.Task) sampleTask { calls++; return fieldOnlyToTask(x) }
			a.FromTask = func(x sampleTask) transport.Task { calls++; return fieldOnlyFromTask(x) }
			a.ToCommand = func(x transport.Command) sampleCommand { calls++; return fieldOnlyToCommand(x) }
			a.FromResult = func(x sampleResult) transport.Result { calls++; return fieldOnlyFromResult(x) }
			switch which {
			case "ToTask":
				a.ToTask = nil
			case "FromTask":
				a.FromTask = nil
			case "ToCommand":
				a.ToCommand = nil
			case "FromResult":
				a.FromResult = nil
			}
			refuse(t, a)
			if calls != 0 {
				t.Fatal("callback ran before nil callback preflight")
			}
		})
	}
}
func inventoryRefused[T, C, R any](t *testing.T) {
	t.Helper()
	calls := 0
	a := ac.Adapter[T, C, R]{ToTask: func(transport.Task) T { calls++; var x T; return x }, FromTask: func(T) transport.Task { calls++; return transport.Task{} }, ToCommand: func(transport.Command) C { calls++; var x C; return x }, FromResult: func(R) transport.Result { calls++; return transport.Result{} }}
	if ac.Check(a) == nil {
		t.Fatal("wrong field inventory accepted")
	}
	if calls != 0 {
		t.Fatal("callback invoked before inventory validation")
	}
}

type extraTask struct {
	ID         string
	State      sampleState
	Revision   uint64
	Outcome    string
	Acceptance []string
	Owner      *string
	Extra      string
}
type missingTask struct {
	ID         string
	State      sampleState
	Revision   uint64
	Outcome    string
	Acceptance []string
}
type wrongTask struct {
	ID         string
	State      sampleState
	Revision   uint32
	Outcome    string
	Acceptance []string
	Owner      *string
}
type extraCommand struct {
	Type             string
	Actor            string
	ExpectedRevision uint64
	Review           *sampleReview
	Extra            string
}
type missingCommand struct {
	Type             string
	ExpectedRevision uint64
	Review           *sampleReview
}
type wrongReview struct {
	TaskID       string
	TaskRevision uint32
	ModelDigest  string
	TaskDigest   string
	Reviewer     string
	Decision     string
}
type wrongCommand struct {
	Type             string
	Actor            string
	ExpectedRevision uint64
	Review           *wrongReview
}
type extraReview struct {
	TaskID       string
	TaskRevision uint64
	ModelDigest  string
	TaskDigest   string
	Reviewer     string
	Decision     string
	Extra        bool
}
type extraReviewCommand struct {
	Type             string
	Actor            string
	ExpectedRevision uint64
	Review           *extraReview
}
type missingReview struct {
	TaskID       string
	TaskRevision uint64
	ModelDigest  string
	TaskDigest   string
	Reviewer     string
}
type missingReviewCommand struct {
	Type             string
	Actor            string
	ExpectedRevision uint64
	Review           *missingReview
}
type extraResult struct {
	Mode         string
	Authority    string
	ModelDigest  string
	OK           bool
	TransitionID string
	Task         *sampleTask
	Refusals     []string
	Extra        bool
}
type missingResult struct {
	Mode        string
	Authority   string
	ModelDigest string
	OK          bool
	Task        *sampleTask
	Refusals    []string
}
type wrongResult struct {
	Mode         string
	Authority    string
	ModelDigest  string
	OK           bool
	TransitionID string
	Task         *sampleTask
	Refusals     string
}
type wrongNestedResult struct {
	Mode         string
	Authority    string
	ModelDigest  string
	OK           bool
	TransitionID string
	Task         *extraTask
	Refusals     []string
}

func TestInventoryPrecedesEveryCallback(t *testing.T) {
	t.Run("extra-task", func(t *testing.T) { inventoryRefused[extraTask, sampleCommand, sampleResult](t) })
	t.Run("missing-task", func(t *testing.T) { inventoryRefused[missingTask, sampleCommand, sampleResult](t) })
	t.Run("wrong-task", func(t *testing.T) { inventoryRefused[wrongTask, sampleCommand, sampleResult](t) })
	t.Run("extra-command", func(t *testing.T) { inventoryRefused[sampleTask, extraCommand, sampleResult](t) })
	t.Run("missing-command", func(t *testing.T) { inventoryRefused[sampleTask, missingCommand, sampleResult](t) })
	t.Run("wrong-review", func(t *testing.T) { inventoryRefused[sampleTask, wrongCommand, sampleResult](t) })
	t.Run("extra-review", func(t *testing.T) { inventoryRefused[sampleTask, extraReviewCommand, sampleResult](t) })
	t.Run("missing-review", func(t *testing.T) { inventoryRefused[sampleTask, missingReviewCommand, sampleResult](t) })
	t.Run("extra-result", func(t *testing.T) { inventoryRefused[sampleTask, sampleCommand, extraResult](t) })
	t.Run("missing-result", func(t *testing.T) { inventoryRefused[sampleTask, sampleCommand, missingResult](t) })
	t.Run("wrong-result", func(t *testing.T) { inventoryRefused[sampleTask, sampleCommand, wrongResult](t) })
	t.Run("nested-result-task", func(t *testing.T) { inventoryRefused[sampleTask, sampleCommand, wrongNestedResult](t) })
	t.Run("non-struct", func(t *testing.T) { inventoryRefused[int, sampleCommand, sampleResult](t) })
}
func TestSpecimenCoverageWithoutPrivateLiteralCoupling(t *testing.T) {
	seen := map[string]bool{}
	markList := func(prefix string, x []string) {
		switch {
		case x == nil:
			seen[prefix+"nil"] = true
		case len(x) == 0:
			seen[prefix+"empty"] = true
		default:
			seen[prefix+"nonempty"] = true
		}
	}
	markUint := func(prefix string, x uint64) {
		if x == 0 {
			seen[prefix+"zero"] = true
		}
		if x == math.MaxUint64 {
			seen[prefix+"max"] = true
		}
	}
	markTask := func(prefix string, x transport.Task) {
		markList(prefix+"acceptance/", x.Acceptance)
		markUint(prefix+"revision/", x.Revision)
		if x.ID != "" && x.Outcome != "" && x.ID != x.Outcome {
			seen[prefix+"distinct-id-outcome"] = true
		}
		if x.Owner == nil {
			seen[prefix+"owner/nil"] = true
		} else if *x.Owner == "" {
			seen[prefix+"owner/empty"] = true
		} else {
			seen[prefix+"owner/nonempty"] = true
		}
	}
	a := good()
	a.ToTask = func(x transport.Task) sampleTask { markTask("to/", x); return fieldOnlyToTask(x) }
	a.FromTask = func(x sampleTask) transport.Task {
		markTask("from/", fieldOnlyFromTask(x))
		return fieldOnlyFromTask(x)
	}
	a.ToCommand = func(x transport.Command) sampleCommand {
		markUint("command/revision/", x.ExpectedRevision)
		if x.Review == nil {
			seen["command/review/nil"] = true
		} else {
			seen["command/review/nonempty"] = true
			markUint("review/revision/", x.Review.TaskRevision)
			if x.ExpectedRevision != x.Review.TaskRevision {
				seen["command/distinct-revisions"] = true
			}
		}
		return fieldOnlyToCommand(x)
	}
	a.FromResult = func(x sampleResult) transport.Result {
		markList("result/refusals/", x.Refusals)
		if x.Task == nil {
			seen["result/task/nil"] = true
		} else {
			seen["result/task/nonempty"] = true
			markTask("result/", fieldOnlyFromTask(*x.Task))
		}
		if x.OK {
			seen["result/ok/true"] = true
		} else {
			seen["result/ok/false"] = true
		}
		return fieldOnlyFromResult(x)
	}
	if e := ac.Check(a); e != nil {
		t.Fatal(e)
	}
	want := []string{"command/revision/zero", "command/revision/max", "command/review/nil", "command/review/nonempty", "review/revision/zero", "review/revision/max", "command/distinct-revisions", "result/task/nil", "result/task/nonempty", "result/ok/true", "result/ok/false", "result/refusals/nil", "result/refusals/empty", "result/refusals/nonempty"}
	for _, prefix := range []string{"to/", "from/", "result/"} {
		for _, suffix := range []string{"acceptance/nil", "acceptance/empty", "acceptance/nonempty", "revision/zero", "revision/max", "distinct-id-outcome", "owner/nil", "owner/empty", "owner/nonempty"} {
			want = append(want, prefix+suffix)
		}
	}
	for _, key := range want {
		if !seen[key] {
			t.Error("missing independent probe class:", key)
		}
	}
}

func drop(v reflect.Value, path []string) {
	for v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return
		}
		v = v.Elem()
	}
	f := v.FieldByName(path[0])
	if len(path) > 1 {
		drop(f, path[1:])
		return
	}
	f.Set(reflect.Zero(f.Type()))
}
func TestIndividualScalarFieldDrops(t *testing.T) {
	fields := map[string][]string{
		"ToTask":     {"ID", "State", "Revision", "Outcome"},
		"FromTask":   {"ID", "State", "Revision", "Outcome"},
		"ToCommand":  {"Type", "Actor", "ExpectedRevision", "Review.TaskID", "Review.TaskRevision", "Review.ModelDigest", "Review.TaskDigest", "Review.Reviewer", "Review.Decision"},
		"FromResult": {"Mode", "Authority", "ModelDigest", "OK", "TransitionID", "Task.ID", "Task.State", "Task.Revision", "Task.Outcome"},
	}
	for direction, names := range fields {
		for _, name := range names {
			t.Run(direction+"/"+name, func(t *testing.T) {
				a := good()
				p := strings.Split(name, ".")
				switch direction {
				case "ToTask":
					a.ToTask = func(x transport.Task) sampleTask {
						y := fieldOnlyToTask(x)
						drop(reflect.ValueOf(&y).Elem(), p)
						return y
					}
				case "FromTask":
					a.FromTask = func(x sampleTask) transport.Task {
						y := fieldOnlyFromTask(x)
						drop(reflect.ValueOf(&y).Elem(), p)
						return y
					}
				case "ToCommand":
					a.ToCommand = func(x transport.Command) sampleCommand {
						y := fieldOnlyToCommand(x)
						drop(reflect.ValueOf(&y).Elem(), p)
						return y
					}
				case "FromResult":
					a.FromResult = func(x sampleResult) transport.Result {
						y := fieldOnlyFromResult(x)
						drop(reflect.ValueOf(&y).Elem(), p)
						return y
					}
				}
				refuse(t, a)
			})
		}
	}
}
