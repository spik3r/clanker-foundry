// Package canonical identifies accepted IR independently from source formatting.
package canonical

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/model"
)

func expression(e model.Expr) any {
	switch e.Kind {
	case "ref":
		return map[string]any{"ref": e.Ref}
	case "string":
		return map[string]any{"string": e.Value}
	default:
		a := make([]any, len(e.Args))
		for i, x := range e.Args {
			a[i] = expression(x)
		}
		return map[string]any{"op": e.Op, "args": a}
	}
}
func Bytes(m model.Model) ([]byte, error) {
	m, e := model.Normalize(m)
	if e != nil {
		return nil, e
	}
	states := []any{}
	reqs := []any{}
	guards := []any{}
	transitions := []any{}
	for _, s := range m.States {
		states = append(states, map[string]any{"id": s.ID, "label": s.Label, "terminal": s.Terminal})
	}
	for _, r := range m.Requirements {
		reqs = append(reqs, map[string]any{"id": r.ID, "text": r.Text})
	}
	for _, g := range m.Guards {
		guards = append(guards, map[string]any{"id": g.ID, "requirement": g.Requirement, "expr": expression(g.Expr)})
	}
	for _, t := range m.Transitions {
		effects := []any{}
		for _, v := range t.Effects {
			effects = append(effects, map[string]any{"op": v.Op})
		}
		transitions = append(transitions, map[string]any{"id": t.ID, "from": t.From, "trigger": t.Trigger, "to": t.To, "guards": t.Guards, "effects": effects})
	}
	doc := map[string]any{"profileVersion": model.ProfileVersion, "id": m.ID, "initial": m.Initial, "states": states, "requirements": reqs, "guards": guards, "transitions": transitions}
	var b bytes.Buffer
	enc := json.NewEncoder(&b)
	enc.SetEscapeHTML(false)
	if e = enc.Encode(doc); e != nil {
		return nil, e
	}
	return bytes.TrimSuffix(b.Bytes(), []byte{'\n'}), nil
}
func Digest(m model.Model) (string, error) {
	b, e := Bytes(m)
	if e != nil {
		return "", e
	}
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:]), nil
}
