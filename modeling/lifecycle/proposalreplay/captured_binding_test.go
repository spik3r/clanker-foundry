package proposalreplay

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
)

// These controls mutate only the independently authored, hash-checked literal
// reports. No report producer or reducer supplies the expected observations.
func capturedLiteral(t *testing.T, set int) (map[string]any, proposal, Receipt) {
	t.Helper()
	x := literals(t)[set]
	p, r, err := prepare(requestBytes(t, fixture(t, x)), "execute")
	if err != nil {
		t.Fatal(err)
	}
	var d map[string]any
	if err := json.Unmarshal(x.Report, &d); err != nil {
		t.Fatal(err)
	}
	return d, p, r
}
func capturedEntryAt(d map[string]any, n int) map[string]any {
	return d["entries"].([]any)[n].(map[string]any)
}
func capturedObject(d map[string]any, key string) map[string]any { return d[key].(map[string]any) }
func capturedWire(t *testing.T, d map[string]any) []byte {
	t.Helper()
	b, err := json.Marshal(d)
	if err != nil {
		t.Fatal(err)
	}
	return b
}
func capturedReject(t *testing.T, d map[string]any, exit int, p proposal, r Receipt) {
	t.Helper()
	if matched, err := captured(capturedWire(t, d), exit, p, r); matched || !errors.Is(err, ErrExecution) {
		t.Fatalf("accepted invalid captured report: matched=%v err=%v", matched, err)
	}
}

func TestCapturedBindingNestedShape(t *testing.T) {
	for _, field := range []string{"input", "command", "expected", "actual"} {
		for _, kind := range []string{"empty", "null", "array", "scalar", "unknown", "missing", "case", "duplicate"} {
			t.Run(field+"/"+kind, func(t *testing.T) {
				d, p, r := capturedLiteral(t, 0)
				e := capturedEntryAt(d, 0)
				obj := capturedObject(e, field)
				required := map[string]string{"input": "id", "command": "type", "expected": "ok", "actual": "mode"}[field]
				switch kind {
				case "empty":
					e[field] = map[string]any{}
				case "null":
					e[field] = nil
				case "array":
					e[field] = []any{}
				case "scalar":
					e[field] = "opaque"
				case "unknown":
					obj["unexpected"] = true
				case "missing":
					delete(obj, required)
				case "case":
					obj[strings.ToUpper(required)] = obj[required]
					delete(obj, required)
				case "duplicate":
					b := capturedWire(t, obj)
					value, err := json.Marshal(obj[required])
					if err != nil {
						t.Fatal(err)
					}
					e[field] = json.RawMessage(append([]byte(fmt.Sprintf("{%q:%s,", required, value)), b[1:]...))
				}
				capturedReject(t, d, 0, p, r)
			})
		}
	}
}

func TestCapturedBindingComparison(t *testing.T) {
	for _, field := range []string{"mode", "authority", "modelDigest", "ok", "transitionId", "task", "refusals"} {
		t.Run(field, func(t *testing.T) {
			d, p, r := capturedLiteral(t, 0)
			actual := capturedObject(capturedEntryAt(d, 0), "actual")
			switch field {
			case "ok":
				actual[field] = false
			case "task":
				capturedObject(actual, "task")["revision"] = "9"
			case "refusals":
				actual[field] = []any{"G_actor"}
			default:
				actual[field] = "changed"
			}
			capturedReject(t, d, 0, p, r)
		})
	}
	for _, differences := range [][]any{{}, {"ok"}, {"task", "ok"}, {"task", "task"}} {
		t.Run(fmt.Sprint(differences), func(t *testing.T) {
			d, p, r := capturedLiteral(t, 0)
			capturedMismatch(d, 0, 0)
			capturedEntryAt(d, 0)["differences"] = differences
			capturedReject(t, d, 1, p, r)
		})
	}
}

func TestCapturedBindingOriginalLiterals(t *testing.T) {
	for _, mutation := range []string{"command", "initial", "consistent-revision", "consistent-transition", "consistent-refusals", "empty-versus-null-owner", "acceptance-order"} {
		t.Run(mutation, func(t *testing.T) {
			d, p, r := capturedLiteral(t, 0)
			e := capturedEntryAt(d, 0)
			switch mutation {
			case "command":
				capturedObject(e, "command")["actor"] = "Other"
			case "initial":
				capturedObject(e, "input")["owner"] = ""
			case "consistent-revision":
				for _, field := range []string{"expected", "actual"} {
					capturedObject(capturedObject(e, field), "task")["revision"] = "9"
				}
				next := capturedEntryAt(d, 1)
				capturedObject(next, "input")["revision"] = "9"
				for _, field := range []string{"expected", "actual"} {
					capturedObject(capturedObject(next, field), "task")["revision"] = "9"
				}
			case "consistent-transition":
				for _, field := range []string{"expected", "actual"} {
					capturedObject(e, field)["transitionId"] = "T_other"
				}
			case "consistent-refusals":
				for _, field := range []string{"expected", "actual"} {
					capturedObject(capturedEntryAt(d, 3), field)["refusals"] = []any{"G_revision", "G_outcome", "G_actor"}
				}
			case "empty-versus-null-owner":
				for _, field := range []string{"expected", "actual"} {
					capturedObject(capturedObject(capturedEntryAt(d, 2), field), "task")["owner"] = nil
				}
			case "acceptance-order":
				for _, field := range []string{"expected", "actual"} {
					capturedObject(capturedObject(e, field), "task")["acceptance"] = []any{"Second", "First"}
				}
				next := capturedEntryAt(d, 1)
				capturedObject(next, "input")["acceptance"] = []any{"Second", "First"}
				for _, field := range []string{"expected", "actual"} {
					capturedObject(capturedObject(next, field), "task")["acceptance"] = []any{"Second", "First"}
				}
			}
			capturedReject(t, d, 0, p, r)
		})
	}
}

func TestCapturedBindingInputFlow(t *testing.T) {
	for _, mutation := range []string{"success-retained-old-input", "refusal-advanced", "invalid-advanced", "invalid-null"} {
		t.Run(mutation, func(t *testing.T) {
			set := 0
			if mutation == "invalid-advanced" || mutation == "invalid-null" {
				set = 1
			}
			d, p, r := capturedLiteral(t, set)
			switch mutation {
			case "success-retained-old-input":
				capturedEntryAt(d, 1)["input"] = capturedEntryAt(d, 0)["input"]
			case "refusal-advanced":
				capturedObject(capturedEntryAt(d, 4), "input")["revision"] = "13"
			case "invalid-advanced":
				capturedObject(capturedEntryAt(d, 6), "input")["revision"] = "2"
			case "invalid-null":
				capturedEntryAt(d, 6)["input"] = nil
			}
			capturedReject(t, d, 0, p, r)
		})
	}
}

// This literal mismatch changes only the last actual task's revision. Counts
// are supplied by the test's hand-counted fixture case boundaries, not a runner.
func capturedMismatch(d map[string]any, last, completed int) {
	e := capturedEntryAt(d, last)
	capturedObject(capturedObject(e, "actual"), "task")["revision"] = "99"
	e["matched"], e["differences"] = false, []any{"task"}
	d["matched"], d["casesCompleted"], d["stepsExecuted"] = false, completed, last+1
	d["entries"] = d["entries"].([]any)[:last+1]
}

func TestCapturedBindingPositiveControls(t *testing.T) {
	for _, set := range []int{0, 1} {
		t.Run(fmt.Sprintf("full-%d", set), func(t *testing.T) {
			d, p, r := capturedLiteral(t, set)
			if matched, err := captured(capturedWire(t, d), 0, p, r); !matched || err != nil {
				t.Fatalf("frozen match: %v %v", matched, err)
			}
		})
	}
	for _, tc := range []struct{ last, completed int }{{0, 0}, {1, 0}, {2, 1}, {3, 2}, {4, 2}, {5, 3}} {
		t.Run(fmt.Sprintf("mismatch-%d", tc.last), func(t *testing.T) {
			d, p, r := capturedLiteral(t, 0)
			capturedMismatch(d, tc.last, tc.completed)
			if matched, err := captured(capturedWire(t, d), 1, p, r); matched || err != nil {
				t.Fatalf("honest mismatch: %v %v", matched, err)
			}
		})
	}
}

func TestCapturedBindingPrefixControls(t *testing.T) {
	for _, mutation := range []string{"wrong-completed", "continued", "truncated-match", "false-mismatch"} {
		t.Run(mutation, func(t *testing.T) {
			d, p, r := capturedLiteral(t, 0)
			exit := 1
			switch mutation {
			case "wrong-completed":
				capturedMismatch(d, 2, 0)
			case "continued":
				all := d["entries"]
				capturedMismatch(d, 0, 0)
				d["entries"], d["stepsExecuted"] = all, 6
			case "truncated-match":
				d["entries"], d["stepsExecuted"] = d["entries"].([]any)[:1], 1
				exit = 0
			case "false-mismatch":
				e := capturedEntryAt(d, 0)
				e["matched"], e["differences"] = false, []any{"task"}
				d["entries"], d["matched"], d["casesCompleted"], d["stepsExecuted"] = d["entries"].([]any)[:1], false, 0, 1
			}
			capturedReject(t, d, exit, p, r)
		})
	}
}

// These supplemental controls were specified after implementation review. They
// document protocol tightening, not the pre-implementation red evidence above:
// truthful differences do not make an invalid actual-result envelope acceptable.
func TestCapturedBindingProtocolClassification(t *testing.T) {
	for _, field := range []string{"mode", "authority", "modelDigest", "refusals", "task", "ok"} {
		t.Run(field, func(t *testing.T) {
			d, p, r := capturedLiteral(t, 0)
			e := capturedEntryAt(d, 0)
			actual := capturedObject(e, "actual")
			switch field {
			case "mode":
				actual[field] = "production"
			case "authority":
				actual[field] = "claimed"
			case "modelDigest":
				actual[field] = strings.Repeat("0", 64)
			case "refusals", "task":
				actual[field] = nil
			case "ok":
				actual[field] = false
			}
			e["matched"], e["differences"] = false, []any{field}
			d["matched"], d["casesCompleted"], d["stepsExecuted"] = false, 0, 1
			d["entries"] = d["entries"].([]any)[:1]
			capturedReject(t, d, 1, p, r)
		})
	}
}
