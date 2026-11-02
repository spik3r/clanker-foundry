package profilecatalog

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/jsonformat"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/model"
)

func fileHash(b []byte) string { x := sha256.Sum256(b); return hex.EncodeToString(x[:]) }
func readVector(t *testing.T, name, want string) []byte {
	t.Helper()
	b, e := os.ReadFile(filepath.Join("testdata", name))
	if e != nil || fileHash(b) != want {
		t.Fatalf("vector %s changed: %v", name, e)
	}
	return b
}
func TestDescriptorLiteralAndOwnership(t *testing.T) {
	literal := readVector(t, "catalog-vectors.json", "42d745b5307eb8f8a6985b4f4cc875d93fa1f7dcae2cf6922add38cd384c392a")
	var want Descriptor
	if e := json.Unmarshal(literal, &want); e != nil {
		t.Fatal(e)
	}
	got, e := Describe()
	if e != nil || !reflect.DeepEqual(got, want) {
		t.Fatalf("descriptor differs: %v", e)
	}
	var compact bytes.Buffer
	if e = json.Compact(&compact, literal); e != nil {
		t.Fatal(e)
	}
	expected := append(compact.Bytes(), '\n')
	encoded, e := Encode()
	if e != nil || !bytes.Equal(encoded, expected) {
		t.Fatal("exact encoded descriptor differs", e)
	}
	if len(encoded) > MaxOutputBytes {
		t.Fatal("output cap")
	}
	got.References[0].Path = "changed"
	got.SourcePins[0].SHA256 = "changed"
	got.Operators[0].AllowedOperandTypes[0] = "changed"
	got.Effects.AllowedSequences[0][0] = "changed"
	got.StructuralRules[0] = "changed"
	again, e := Describe()
	if e != nil || !reflect.DeepEqual(again, want) {
		t.Fatal("descriptor aliases later call")
	}
	var wg sync.WaitGroup
	errs := make(chan error, 16)
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			d, e := Describe()
			if e != nil || !reflect.DeepEqual(d, want) {
				errs <- fmt.Errorf("concurrent descriptor")
			}
			d.References[0].Path = "changed"
		}()
	}
	wg.Wait()
	close(errs)
	for e := range errs {
		t.Error(e)
	}
}
func matchesPin(b []byte, p SourcePin) bool {
	return len(b) == p.Size && fileHash(b) == p.SHA256 && p.IdentityKind == "expected-reviewed-source"
}
func TestReviewedSourcePins(t *testing.T) {
	d, e := Describe()
	if e != nil {
		t.Fatal(e)
	}
	for _, pin := range d.SourcePins {
		t.Run(pin.Path, func(t *testing.T) {
			b, e := os.ReadFile(filepath.Join("..", filepath.FromSlash(pin.Path)))
			if e != nil || !matchesPin(b, pin) {
				t.Fatalf("source drift: %v", e)
			}
			changed := append([]byte(nil), b...)
			changed[0] ^= 1
			if matchesPin(changed, pin) {
				t.Fatal("mutation not detected")
			}
		})
	}
}
func TestFrozenTypeOfVectors(t *testing.T) {
	raw := readVector(t, "parity-vectors.json", "789378e0e89dacd84f6a2b4f06078b6aeab895b5fd59a467d0c5af72dcfc2e94")
	var v struct {
		TypeOf []struct {
			ID    string     `json:"id"`
			Expr  model.Expr `json:"expr"`
			Valid bool       `json:"valid"`
			Type  string     `json:"type"`
		} `json:"typeOf"`
	}
	if e := json.Unmarshal(raw, &v); e != nil || len(v.TypeOf) != 23 {
		t.Fatal("parity vectors", e)
	}
	for _, c := range v.TypeOf {
		t.Run(c.ID, func(t *testing.T) {
			typ, e := model.TypeOf(c.Expr)
			if (e == nil) != c.Valid || (c.Valid && typ != c.Type) {
				t.Fatalf("type %q error %v", typ, e)
			}
		})
	}
}
func ref(s string) model.Expr                 { return model.Expr{Kind: "ref", Ref: s} }
func str(s string) model.Expr                 { return model.Expr{Kind: "string", Value: s} }
func op(s string, a ...model.Expr) model.Expr { return model.Expr{Kind: "op", Op: s, Args: a} }
func nodes(e model.Expr) int {
	n := 1
	for _, a := range e.Args {
		n += nodes(a)
	}
	return n
}
func TestExpressionBoundaries(t *testing.T) {
	leaf := op("nonempty", ref("task.id"))
	valid := leaf
	for i := 0; i < 15; i++ {
		valid = op("all", valid)
	}
	if typ, e := model.TypeOf(valid); e != nil || typ != "bool" {
		t.Fatal("depth 16 rejected", e)
	}
	if _, e := model.TypeOf(op("all", valid)); e == nil {
		t.Fatal("depth 17 accepted")
	}
	groups := []model.Expr{}
	for g := 0; g < 11; g++ {
		var children []model.Expr
		for i := 0; i < 15; i++ {
			children = append(children, op("eq", ref("task.id"), str(fmt.Sprintf("%d-%d", g, i))))
		}
		groups = append(groups, op("all", children...))
	}
	groups = append(groups, op("nonempty", ref("task.id")), op("nonempty", ref("task.outcome")))
	good := op("all", groups...)
	// The prior groups already include two nonempty nodes: 511 + 3 - 2 = 512.
	exactArgs := append(append([]model.Expr(nil), groups[:11]...), op("eq", ref("task.id"), str("extra-boundary")), op("nonempty", ref("task.id")))
	exact := op("all", exactArgs...)
	if nodes(exact) != 512 {
		t.Fatalf("bad 512 fixture: %d", nodes(exact))
	}
	if _, e := model.TypeOf(exact); e != nil {
		t.Fatal("512 nodes rejected", e)
	}
	if nodes(good) != 511 {
		t.Fatal("bad fixture")
	}
	if _, e := model.TypeOf(good); e != nil {
		t.Fatal("511 nodes rejected", e)
	}
	groups = append(groups, op("nonempty", ref("command.actor")))
	bad := op("all", groups...)
	if nodes(bad) != 513 {
		t.Fatal("bad fixture")
	}
	if _, e := model.TypeOf(bad); e == nil {
		t.Fatal("513 nodes accepted")
	}
	args := make([]model.Expr, 17)
	for i := range args {
		args[i] = op("eq", ref("task.id"), str(fmt.Sprintf("v%d", i)))
	}
	if _, e := model.TypeOf(op("all", args...)); e == nil {
		t.Fatal("17 boolean args accepted")
	}
	if _, e := model.TypeOf(str(string([]byte{0xff}))); e == nil {
		t.Fatal("invalid UTF-8 accepted")
	}
}
func baseline(t *testing.T) model.Model {
	t.Helper()
	raw, e := os.ReadFile("../examples/lifecycle.json")
	if e != nil {
		t.Fatal(e)
	}
	m, e := jsonformat.Decode(raw)
	if e != nil {
		t.Fatal(e)
	}
	return m
}
func TestModelGraphAndEffects(t *testing.T) {
	if e := model.Validate(baseline(t)); e != nil {
		t.Fatal(e)
	}
	positive := baseline(t)
	for i := range positive.Transitions {
		if len(positive.Transitions[i].Effects) == 2 {
			positive.Transitions[i].Effects = append([]model.Effect{{Op: model.AssignOwner}}, positive.Transitions[i].Effects...)
			break
		}
	}
	if e := model.Validate(positive); e != nil {
		t.Fatal("allowed assignment effect", e)
	}
	cases := map[string]func(*model.Model){
		"global duplicate": func(m *model.Model) { m.States[0].ID = m.Requirements[0].ID }, "blank state label": func(m *model.Model) { m.States[0].Label = " " }, "invalid state UTF8": func(m *model.Model) { m.States[0].Label = string([]byte{0xff}) }, "blank requirement": func(m *model.Model) { m.Requirements[0].Text = " " }, "terminal outgoing": func(m *model.Model) {
			for i := range m.States {
				if m.States[i].ID == "Draft" {
					m.States[i].Terminal = true
				}
			}
		}, "nonterminal without outgoing": func(m *model.Model) {
			for i := range m.States {
				if m.States[i].ID == "Done" {
					m.States[i].Terminal = false
				}
			}
		}, "unreachable": func(m *model.Model) {
			m.States = append(m.States, model.State{ID: "Isolated", Label: "Isolated", Terminal: true})
		}, "duplicate state trigger": func(m *model.Model) {
			x := m.Transitions[0]
			x.ID = "T_duplicate"
			m.Transitions = append(m.Transitions, x)
		}, "dangling state": func(m *model.Model) { m.Transitions[0].To = "Missing" }, "dangling guard": func(m *model.Model) { m.Transitions[0].Guards = append(m.Transitions[0].Guards, "G_missing") }, "unused requirement": func(m *model.Model) {
			m.Requirements = append(m.Requirements, model.Requirement{ID: "R_unused", Text: "Unused"})
		}, "unused guard": func(m *model.Model) {
			m.Guards = append(m.Guards, model.Guard{ID: "G_unused", Requirement: m.Requirements[0].ID, Expr: op("nonempty", ref("task.id"))})
		}, "duplicate transition guard": func(m *model.Model) {
			m.Transitions[0].Guards = append(m.Transitions[0].Guards, m.Transitions[0].Guards[0])
		}, "missing effect": func(m *model.Model) { m.Transitions[0].Effects = m.Transitions[0].Effects[:1] }, "extra effect": func(m *model.Model) {
			m.Transitions[0].Effects = append(m.Transitions[0].Effects, model.Effect{Op: "extra"})
		}, "unsupported effect": func(m *model.Model) { m.Transitions[0].Effects[0].Op = "execute" }, "reordered effects": func(m *model.Model) {
			m.Transitions[0].Effects[0], m.Transitions[0].Effects[1] = m.Transitions[0].Effects[1], m.Transitions[0].Effects[0]
		},
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			m := baseline(t)
			mutate(&m)
			if e := model.Validate(m); e == nil {
				t.Fatal("invalid model accepted")
			}
		})
	}
}
func TestNormalizationRules(t *testing.T) {
	a := baseline(t)
	b := baseline(t)
	x := op("nonempty", ref("task.id"))
	y := op("absent", ref("task.owner"))
	a.Guards[0].Expr = op("all", x, y)
	b.Guards[0].Expr = op("all", y, x)
	na, e := model.Normalize(a)
	if e != nil {
		t.Fatal(e)
	}
	nb, e := model.Normalize(b)
	if e != nil || !reflect.DeepEqual(na, nb) {
		t.Fatal("all order", e)
	}
	a = baseline(t)
	b = baseline(t)
	a.Guards[0].Expr = op("eq", ref("task.id"), str("x"))
	b.Guards[0].Expr = op("eq", str("x"), ref("task.id"))
	na, e = model.Normalize(a)
	if e != nil {
		t.Fatal(e)
	}
	nb, e = model.Normalize(b)
	if e != nil || reflect.DeepEqual(na, nb) {
		t.Fatal("comparison order changed", e)
	}
}
func TestOutputCap(t *testing.T) {
	if MaxOutputBytes != 65536 {
		t.Fatal("output cap changed")
	}
	d, e := Describe()
	if e != nil {
		t.Fatal(e)
	}
	d.Status = strings.Repeat("x", MaxOutputBytes+1)
	if b, e := encodeDescriptor(d); e == nil || b != nil {
		t.Fatal("oversized descriptor accepted")
	}
}

func TestAggregateGuardNodes(t *testing.T) {
	var groups []model.Expr
	for g := 0; g < 11; g++ {
		var children []model.Expr
		for i := 0; i < 15; i++ {
			children = append(children, op("eq", ref("task.id"), str(fmt.Sprintf("aggregate-%d-%d", g, i))))
		}
		groups = append(groups, op("all", children...))
	}
	// 1 root + 11*(1+15*3) + 3 = 510; the other guard uses 2.
	first := op("all", append(groups, op("eq", ref("task.id"), str("extra")))...)
	if nodes(first) != 510 {
		t.Fatalf("bad 512-node fixture: %d", nodes(first))
	}
	makeModel := func(expr model.Expr) model.Model {
		return model.Model{Version: 1, ID: "Example", Initial: "Start", States: []model.State{{ID: "Start", Label: "Start"}, {ID: "End", Label: "End", Terminal: true}}, Requirements: []model.Requirement{{ID: "Req", Text: "Requirement"}}, Guards: []model.Guard{{ID: "GuardA", Requirement: "Req", Expr: expr}, {ID: "GuardB", Requirement: "Req", Expr: op("nonempty", ref("task.id"))}}, Transitions: []model.Transition{{ID: "Edge", From: "Start", To: "End", Trigger: "proceed", Guards: []string{"GuardA", "GuardB"}, Effects: []model.Effect{{Op: model.SetState}, {Op: model.IncrementRevision}}}}}
	}
	if e := model.Validate(makeModel(first)); e != nil {
		t.Fatalf("512 aggregate nodes rejected: %v", e)
	}
	// 511 in GuardA plus 2 in GuardB: each guard types independently but total 513.
	over := op("all", append(groups, op("nonempty", ref("task.id")), op("nonempty", ref("task.outcome")))...)
	if nodes(over) != 511 {
		t.Fatalf("bad 513-node fixture: %d", nodes(over))
	}
	if _, e := model.TypeOf(over); e != nil {
		t.Fatal("per-expression input invalid", e)
	}
	if e := model.Validate(makeModel(over)); e == nil {
		t.Fatal("513 aggregate nodes accepted")
	}
}

func TestEveryFrozenReference(t *testing.T) {
	raw := readVector(t, "catalog-vectors.json", "42d745b5307eb8f8a6985b4f4cc875d93fa1f7dcae2cf6922add38cd384c392a")
	var v struct {
		References []model.ReferenceCapability `json:"references"`
	}
	if e := json.Unmarshal(raw, &v); e != nil || len(v.References) != 15 {
		t.Fatal("reference vectors", e)
	}
	for _, r := range v.References {
		t.Run(r.Path, func(t *testing.T) {
			typ, e := model.TypeOf(ref(r.Path))
			if e != nil || typ != r.Type {
				t.Fatalf("type %q: %v", typ, e)
			}
			_, e = model.TypeOf(op("absent", ref(r.Path)))
			if (e == nil) != r.Optional {
				t.Fatalf("optional=%v error=%v", r.Optional, e)
			}
		})
	}
}
func TestVersionCompatibility(t *testing.T) {
	good := []string{"lifecycle/v1", "lifecycle-json/v1", "lifecycle-go/v1", "go1.27.3"}
	if !compatibleVersions(good[0], good[1], good[2], good[3]) {
		t.Fatal("reviewed versions refused")
	}
	for i := range good {
		v := append([]string(nil), good...)
		v[i] = "different"
		if compatibleVersions(v[0], v[1], v[2], v[3]) {
			t.Fatalf("version %d accepted", i)
		}
	}
}
