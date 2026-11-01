// Package jsonformat is the replaceable JSON-v1 adapter; semantic IR is separate.
package jsonformat

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/model"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/strictjson"
)

const FormatVersion = "lifecycle-json/v1"

type state struct {
	ID       string `json:"id"`
	Label    string `json:"label"`
	Terminal bool   `json:"terminal"`
}
type requirement struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}
type guard struct {
	ID          string          `json:"id"`
	Requirement string          `json:"requirement"`
	Expr        json.RawMessage `json:"expr"`
}
type effect struct {
	Op string `json:"op"`
}
type transition struct {
	ID      string   `json:"id"`
	From    string   `json:"from"`
	Trigger string   `json:"trigger"`
	To      string   `json:"to"`
	Guards  []string `json:"guards"`
	Effects []effect `json:"effects"`
}
type document struct {
	FormatVersion  string        `json:"formatVersion"`
	ProfileVersion string        `json:"profileVersion"`
	ID             string        `json:"id"`
	Initial        string        `json:"initial"`
	States         []state       `json:"states"`
	Requirements   []requirement `json:"requirements"`
	Guards         []guard       `json:"guards"`
	Transitions    []transition  `json:"transitions"`
}

func expr(raw json.RawMessage, depth int) (model.Expr, error) {
	if depth > model.MaxDepth {
		return model.Expr{}, errors.New("expression depth")
	}
	var fields map[string]json.RawMessage
	if e := strictjson.Decode(raw, &fields, 1<<20, 64); e != nil {
		return model.Expr{}, e
	}
	decodeString := func(key string) (string, error) {
		var s string
		e := strictjson.Decode(fields[key], &s, 1<<20, 2)
		return s, e
	}
	if v, ok := fields["ref"]; ok && len(fields) == 1 {
		_ = v
		s, e := decodeString("ref")
		return model.Expr{Kind: "ref", Ref: s}, e
	}
	if _, ok := fields["string"]; ok && len(fields) == 1 {
		s, e := decodeString("string")
		return model.Expr{Kind: "string", Value: s}, e
	}
	if _, ok := fields["op"]; !ok || len(fields) != 2 || fields["args"] == nil {
		return model.Expr{}, errors.New("unknown expression fields")
	}
	op, e := decodeString("op")
	if e != nil {
		return model.Expr{}, e
	}
	var args []json.RawMessage
	if e = strictjson.Decode(fields["args"], &args, 1<<20, 64); e != nil {
		return model.Expr{}, e
	}
	out := model.Expr{Kind: "op", Op: op}
	for _, a := range args {
		x, e := expr(a, depth+1)
		if e != nil {
			return model.Expr{}, e
		}
		out.Args = append(out.Args, x)
	}
	return out, nil
}
func Decode(data []byte) (model.Model, error) {
	var d document
	if e := strictjson.Decode(data, &d, 1<<20, 64); e != nil {
		return model.Model{}, e
	}
	if d.FormatVersion != FormatVersion || d.ProfileVersion != model.ProfileVersion {
		return model.Model{}, errors.New("unsupported model format/profile")
	}
	m := model.Model{Version: 1, ID: d.ID, Initial: d.Initial}
	for _, s := range d.States {
		m.States = append(m.States, model.State{ID: s.ID, Label: s.Label, Terminal: s.Terminal})
	}
	for _, r := range d.Requirements {
		m.Requirements = append(m.Requirements, model.Requirement{ID: r.ID, Text: r.Text})
	}
	for _, g := range d.Guards {
		x, e := expr(g.Expr, 0)
		if e != nil {
			return model.Model{}, e
		}
		m.Guards = append(m.Guards, model.Guard{ID: g.ID, Requirement: g.Requirement, Expr: x})
	}
	for _, t := range d.Transitions {
		x := model.Transition{ID: t.ID, From: t.From, Trigger: t.Trigger, To: t.To, Guards: t.Guards}
		for _, e := range t.Effects {
			x.Effects = append(x.Effects, model.Effect{Op: e.Op})
		}
		m.Transitions = append(m.Transitions, x)
	}
	return model.Normalize(m)
}
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
func Encode(m model.Model) ([]byte, error) {
	m, e := model.Normalize(m)
	if e != nil {
		return nil, e
	}
	d := document{FormatVersion: FormatVersion, ProfileVersion: model.ProfileVersion, ID: m.ID, Initial: m.Initial}
	for _, s := range m.States {
		d.States = append(d.States, state{s.ID, s.Label, s.Terminal})
	}
	for _, r := range m.Requirements {
		d.Requirements = append(d.Requirements, requirement{r.ID, r.Text})
	}
	for _, g := range m.Guards {
		b, _ := json.Marshal(expression(g.Expr))
		d.Guards = append(d.Guards, guard{g.ID, g.Requirement, b})
	}
	for _, t := range m.Transitions {
		x := transition{ID: t.ID, From: t.From, Trigger: t.Trigger, To: t.To, Guards: t.Guards}
		for _, v := range t.Effects {
			x.Effects = append(x.Effects, effect{v.Op})
		}
		d.Transitions = append(d.Transitions, x)
	}
	var b bytes.Buffer
	enc := json.NewEncoder(&b)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	e = enc.Encode(d)
	if e != nil {
		return nil, e
	}
	if len(b.Bytes()) > 1<<20 {
		return nil, errors.New("JSON-v1 source byte limit")
	}
	if _, e = Decode(b.Bytes()); e != nil {
		return nil, e
	}
	return b.Bytes(), nil
}
