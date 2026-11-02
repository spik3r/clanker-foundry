package modelpatch

import (
	"bytes"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/impact"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/jsonformat"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/model"
)

const wireBefore = `{"formatVersion":"lifecycle-json/v1","profileVersion":"lifecycle/v1","id":"M","initial":"A","states":[{"id":"A","label":"Draft","terminal":false},{"id":"B","label":"Done","terminal":true}],"requirements":[{"id":"R","text":"Invented requirement"}],"guards":[{"id":"G_old","requirement":"R","expr":{"op":"nonempty","args":[{"ref":"task.outcome"}]}}],"transitions":[{"id":"T1","from":"A","trigger":"prepare","to":"B","guards":["G_old"],"effects":[{"op":"set_state"},{"op":"increment_revision"}]}]}`
const wireState = `{"action":"replace","kind":"state","id":"A","root":null,"state":{"id":"A","label":"Edited draft","terminal":false},"requirement":null,"guard":null,"transition":null}`
const wireLiteralGuard = `{"action":"replace","kind":"guard","id":"G_old","root":null,"state":null,"requirement":null,"guard":{"id":"G_old","requirement":"R","expr":{"kind":"op","ref":"","value":"","op":"eq","args":[{"kind":"ref","ref":"task.outcome","value":"","op":"","args":[]},{"kind":"string","ref":"","value":"literal <&> 雪\n\"","op":"","args":[]}]}},"transition":null}`

func request(ops, digest string) []byte {
	return []byte(`{"formatVersion":"lifecycle-model-patch-json/v1","patchVersion":"lifecycle-model-patch/v1","profileVersion":"lifecycle/v1","before":` + wireBefore + `,"beforeDigest":"` + digest + `","operations":` + ops + `}`)
}
func wireReject(t *testing.T, b []byte, want error) {
	t.Helper()
	copyRaw := append([]byte{}, b...)
	out, e := Execute(b)
	if !errors.Is(e, want) || out != nil {
		t.Fatalf("want no bytes and %v, got %s err %v", want, out, e)
	}
	if !bytes.Equal(copyRaw, b) {
		t.Fatal("wire input mutated")
	}
}
func TestStrictExecuteLiteralOutputAndReplay(t *testing.T) {
	raw := request("["+wireLiteralGuard+"]", frozenBeforeDigest)
	out, e := Execute(raw)
	if e != nil {
		t.Fatal(e)
	}
	var r struct {
		ReportVersion, ProfileVersion, Mode, Authority, BeforeDigest, AfterDigest string
		After                                                                     json.RawMessage
		Review                                                                    impact.Report
	}
	if json.Unmarshal(out, &r) != nil {
		t.Fatal("result JSON")
	}
	after, e := jsonformat.Decode(r.After)
	if e != nil {
		t.Fatal(e)
	}
	fixture := frozenFixtures()[4]
	if !reflect.DeepEqual(after, fixture.After) || r.AfterDigest != fixture.AfterDigest || r.BeforeDigest != frozenBeforeDigest || !reflect.DeepEqual(r.Review.Changes, fixture.Changes) || !reflect.DeepEqual(r.Review.LinkedTransitions, fixture.Links) {
		t.Fatal("wire output differs from frozen literal")
	}
	if r.ReportVersion != "lifecycle-model-patch-report/v1" || r.ProfileVersion != "lifecycle/v1" || r.Mode != "example-only" || r.Authority != "none" {
		t.Fatal("result envelope")
	}
	var fields map[string]json.RawMessage
	json.Unmarshal(out, &fields)
	if len(fields) != 8 || len(out) > MaxResultBytes {
		t.Fatal("unexpected fields/size")
	}
	for i := 0; i < 4; i++ {
		again, e := Execute(raw)
		if e != nil || !bytes.Equal(out, again) {
			t.Fatal("wire replay nondeterministic")
		}
	}
	var compact bytes.Buffer
	json.Compact(&compact, raw)
	plain, e := Execute(compact.Bytes())
	if e != nil || !bytes.Equal(out, plain) {
		t.Fatal("input formatting affected normalized output")
	}
	// Existing JSON-v1 may spell expression characters as JSON escapes. The
	// exact decoded value, not a particular raw spelling, is authoritative.
	if after.Guards[0].Expr.Args[1].Value != "literal <&> 雪\n\"" {
		t.Fatal("literal meaning was lost")
	}
}
func TestEveryRequiredOperationFieldAndNullableSlot(t *testing.T) {
	var op map[string]json.RawMessage
	json.Unmarshal([]byte(wireState), &op)
	for key := range op {
		copyOp := map[string]json.RawMessage{}
		for k, v := range op {
			copyOp[k] = v
		}
		delete(copyOp, key)
		b, _ := json.Marshal(copyOp)
		t.Run("missing "+key, func(t *testing.T) { wireReject(t, request("["+string(b)+"]", frozenBeforeDigest), ErrInvalid) })
	}
	malformed := []string{
		strings.Replace(wireState, `"state":{"id":"A","label":"Edited draft","terminal":false}`, `"state":null`, 1),
		strings.Replace(wireState, `"root":null`, `"root":{"id":"M","initial":"A"}`, 1),
		strings.Replace(wireState, `"label":"Edited draft"`, `"label":null`, 1),
		strings.Replace(wireState, `"terminal":false`, `"terminal":null`, 1),
		strings.Replace(wireState, `"state":`, `"State":`, 1),
		strings.Replace(wireState, `"state":`, `"extra":true,"state":`, 1),
		strings.Replace(wireState, `"state":`, `"state":null,"state":`, 1),
		strings.Replace(wireState, `"action":"replace"`, `"action":"merge"`, 1),
		strings.Replace(wireState, `"kind":"state"`, `"kind":"workflow"`, 1),
		strings.Replace(wireLiteralGuard, `"args":[]`, `"args":null`, 1),
		strings.Replace(wireLiteralGuard, `"kind":"ref"`, `"kind":"ref","kind":"ref"`, 1),
		strings.Replace(wireLiteralGuard, `"ref":"task.outcome"`, `"ref":"unsupported.field"`, 1),
		strings.Replace(wireLiteralGuard, `"value":"literal <&> 雪\n\""`, `"value":"\ud800"`, 1),
	}
	for _, op := range malformed {
		wireReject(t, request("["+op+"]", frozenBeforeDigest), ErrInvalid)
	}
}
func TestRequestShapeNullUnicodeAndLimits(t *testing.T) {
	good := request("[]", frozenBeforeDigest)
	var fields map[string]json.RawMessage
	json.Unmarshal(good, &fields)
	for key := range fields {
		copyFields := map[string]json.RawMessage{}
		for k, v := range fields {
			copyFields[k] = v
		}
		delete(copyFields, key)
		b, _ := json.Marshal(copyFields)
		wireReject(t, b, ErrInvalid)
	}
	for _, pair := range [][2]string{{`"operations":[]`, `"operations":null`}, {`"before":`, `"Before":`}, {`"before":`, `"extra":0,"before":`}, {`"before":`, `"before":null,"before":`}, {`"patchVersion":"lifecycle-model-patch/v1"`, `"patchVersion":"unknown/v1"`}, {`"profileVersion":"lifecycle/v1"`, `"profileVersion":"other/v1"`}, {`"formatVersion":"lifecycle-model-patch-json/v1"`, `"formatVersion":"other/v1"`}, {`"beforeDigest":"` + frozenBeforeDigest + `"`, `"beforeDigest":0`}} {
		wireReject(t, bytes.Replace(good, []byte(pair[0]), []byte(pair[1]), 1), ErrInvalid)
	}
	wireReject(t, append(append([]byte{}, good...), []byte(` {}`)...), ErrInvalid)
	wireReject(t, append([]byte{255}, good...), ErrInvalid)
	wireReject(t, append(bytes.Repeat([]byte(" "), MaxRequestBytes+1), good...), ErrLimit)
	wireReject(t, []byte(strings.Repeat("[", MaxJSONDepth+2)+strings.Repeat("]", MaxJSONDepth+2)), ErrLimit)
	beforeLarge := strings.Replace(wireBefore, "{", "{"+strings.Repeat(" ", MaxModelBytes), 1)
	wireReject(t, bytes.Replace(good, []byte(wireBefore), []byte(beforeLarge), 1), ErrInvalid)
}
func TestWireFinalRulesAndStalePrecedence(t *testing.T) {
	transition := `{"action":"replace","kind":"transition","id":"T1","root":null,"state":null,"requirement":null,"guard":null,"transition":{"id":"T1","from":"A","trigger":"prepare","to":"B","guards":["G_old","G_old"],"effects":[{"op":"unknown_bounded_effect"},{"op":"increment_revision"}]}}`
	wireReject(t, request("["+transition+"]", frozenStaleDigest), ErrPrecondition)
	wireReject(t, request("["+transition+"]", frozenBeforeDigest), ErrInvalid)
	badTrigger := strings.Replace(transition, `"trigger":"prepare"`, `"trigger":"bad-trigger"`, 1)
	wireReject(t, request("["+badTrigger+"]", frozenStaleDigest), ErrInvalid)
	duplicate := "[" + wireState + "," + wireState + "]"
	wireReject(t, request(duplicate, frozenStaleDigest), ErrConflict)
	blank := strings.Replace(wireState, `"label":"Edited draft"`, `"label":""`, 1)
	wireReject(t, request("["+blank+"]", frozenStaleDigest), ErrPrecondition)
	wireReject(t, request("["+blank+"]", frozenBeforeDigest), ErrInvalid)
	valid := strings.Replace(transition, `"guards":["G_old","G_old"]`, `"guards":["G_old"]`, 1)
	valid = strings.Replace(valid, `"effects":[{"op":"unknown_bounded_effect"}`, `"effects":[{"op":"assign_owner_if_absent"},{"op":"set_state"}`, 1)
	b, e := Execute(request("["+valid+"]", frozenBeforeDigest))
	if e != nil {
		t.Fatal("legal effect op longer than16 bytes rejected", e)
	}
	if !bytes.Contains(b, []byte(`"effects":[{"op":"assign_owner_if_absent"},{"op":"set_state"},{"op":"increment_revision"}]`)) {
		t.Fatal("effects reordered")
	}
}
func TestWireExpressionAndOperationBudgets(t *testing.T) {
	leaf := map[string]any{"kind": "ref", "ref": "task.outcome", "value": "", "op": "", "args": []any{}}
	var expr any = map[string]any{"kind": "op", "ref": "", "value": "", "op": "nonempty", "args": []any{leaf}}
	for i := 0; i < MaxExprDepth+1; i++ {
		expr = map[string]any{"kind": "op", "ref": "", "value": "", "op": "all", "args": []any{expr}}
	}
	var op map[string]any
	json.Unmarshal([]byte(wireLiteralGuard), &op)
	op["guard"].(map[string]any)["expr"] = expr
	b, _ := json.Marshal(op)
	wireReject(t, request("["+string(b)+"]", frozenStaleDigest), ErrLimit)
	ops := strings.Repeat(wireState+",", MaxOperations) + wireState
	wireReject(t, request("["+ops+"]", frozenBeforeDigest), ErrLimit)
	longEffect := `{"action":"replace","kind":"transition","id":"T1","root":null,"state":null,"requirement":null,"guard":null,"transition":{"id":"T1","from":"A","trigger":"prepare","to":"B","guards":["G_old"],"effects":[{"op":"` + strings.Repeat("x", 65) + `"}]}}`
	wireReject(t, request("["+longEffect+"]", frozenStaleDigest), ErrLimit)
}

func TestTypedWireMixedTransitionHeadErrorParity(t *testing.T) {
	rawOp := `{"action":"replace","kind":"transition","id":"T1","root":null,"state":null,"requirement":null,"guard":null,"transition":{"id":"bad-id","from":"A","trigger":"prepare","to":"B","guards":["G_old"],"effects":[{"op":"x"},{"op":"x"},{"op":"x"},{"op":"x"}]}}`
	typed := Patch{Version: 1, ProfileVersion: ProfileVersion, BeforeDigest: frozenStaleDigest, Operations: []Operation{{Action: Replace, Kind: Transition, ID: "T1", Transition: &model.Transition{ID: "bad-id", From: "A", Trigger: "prepare", To: "B", Guards: []string{"G_old"}, Effects: []model.Effect{{Op: "x"}, {Op: "x"}, {Op: "x"}, {Op: "x"}}}}}}
	assertError(t, frozenBefore(), typed, ErrInvalid)
	wireReject(t, request("["+rawOp+"]", frozenStaleDigest), ErrInvalid)
}

// Fixed mixed-error expectations remain the oracle; DTO construction here only
// transports representable invented values to exercise both public boundaries.
func TestWireFrozenMixedErrorExpectations(t *testing.T) {
	for _, f := range frozenMixedErrors() {
		if f.Name == "stale invalid UTF8" {
			continue
		}
		t.Run(f.Name, func(t *testing.T) {
			ops := []operationDTO{}
			for _, o := range f.Patch.Operations {
				ops = append(ops, operationToDTO(o))
			}
			b, e := json.Marshal(ops)
			if e != nil {
				t.Fatal(e)
			}
			wireReject(t, request(string(b), f.Patch.BeforeDigest), f.Want)
		})
	}
}
