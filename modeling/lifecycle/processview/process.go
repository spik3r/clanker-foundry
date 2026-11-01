// Package processview projects a validated example model for human review.
// It does not execute a process or assert BPMN conformance.
package processview

import (
	"bytes"
	"encoding/json"
	"errors"
	"strconv"
	"strings"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/canonical"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/jsonformat"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/model"
)

const FormatVersion = "lifecycle-process-view/v1"
const MaxSourceBytes = 1 << 20
const MaxOutputBytes = 16 << 20

var ErrInvalid = errors.New("invalid or oversized process view input")

// Outputs contains independent projections of one accepted model. A failure
// returns neither output; both are bounded and owned by the caller.
type Outputs struct{ JSON, SVG []byte }
type state struct {
	ID       string `json:"id"`
	Label    string `json:"label"`
	Terminal bool   `json:"terminal"`
}
type transition struct {
	ID      string   `json:"id"`
	From    string   `json:"from"`
	Trigger string   `json:"trigger"`
	To      string   `json:"to"`
	Guards  []string `json:"guards"`
	Effects []string `json:"effects"`
}
type guard struct {
	ID          string `json:"id"`
	Requirement string `json:"requirement"`
	Expression  string `json:"expression"`
}
type semantics struct {
	Scope         string `json:"scope"`
	Guards        string `json:"guards"`
	RevisionLimit string `json:"revisionLimit"`
	Effects       string `json:"effects"`
	Unsupported   string `json:"unsupported"`
}
type document struct {
	FormatVersion string       `json:"formatVersion"`
	Mode          string       `json:"mode"`
	Authority     string       `json:"authority"`
	ModelDigest   string       `json:"modelDigest"`
	Initial       string       `json:"initial"`
	States        []state      `json:"states"`
	Transitions   []transition `json:"transitions"`
	Guards        []guard      `json:"guards"`
	Semantics     semantics    `json:"semantics"`
}

func expression(e model.Expr) string {
	switch e.Kind {
	case "ref":
		return e.Ref
	case "string":
		return strconv.QuoteToASCII(e.Value)
	case "op":
		args := make([]string, len(e.Args))
		for i, a := range e.Args {
			args[i] = expression(a)
		}
		return e.Op + "(" + strings.Join(args, ", ") + ")"
	default:
		panic("expression was not validated")
	}
}

// Build accepts the existing strict JSON model, normalizes its semantic sets and
// recomputes its canonical identity. It never relies on a caller-supplied digest.
func Build(raw []byte) (Outputs, error) {
	if len(raw) > MaxSourceBytes {
		return Outputs{}, ErrInvalid
	}
	m, e := jsonformat.Decode(raw)
	if e != nil {
		return Outputs{}, ErrInvalid
	}
	digest, e := canonical.Digest(m)
	if e != nil {
		return Outputs{}, ErrInvalid
	}
	d := document{FormatVersion: FormatVersion, Mode: "example-only", Authority: "none", ModelDigest: digest, Initial: m.Initial, States: make([]state, 0, len(m.States)), Transitions: make([]transition, 0, len(m.Transitions)), Guards: make([]guard, 0, len(m.Guards)), Semantics: semantics{
		Scope:         "Each lane starts after valid input and a matching state/trigger; invalid input and illegal transitions are outside these lanes.",
		Guards:        "All declared guards must pass; failure returns the unchanged task and sorted failed guard IDs.",
		RevisionLimit: "After guards pass, maximum revision refuses without changing the task.",
		Effects:       "Otherwise the declared effects apply atomically in their listed order.",
		Unsupported:   "Not BPMN: no pools, roles, parallelism, timers, events, subprocesses, scheduling, execution or approval is represented.",
	}}
	for _, s := range m.States {
		d.States = append(d.States, state{s.ID, s.Label, s.Terminal})
	}
	for _, t := range m.Transitions {
		v := transition{ID: t.ID, From: t.From, Trigger: t.Trigger, To: t.To, Guards: append([]string{}, t.Guards...), Effects: make([]string, 0, len(t.Effects))}
		for _, e := range t.Effects {
			v.Effects = append(v.Effects, e.Op)
		}
		d.Transitions = append(d.Transitions, v)
	}
	for _, g := range m.Guards {
		d.Guards = append(d.Guards, guard{g.ID, g.Requirement, expression(g.Expr)})
	}
	var b bytes.Buffer
	enc := json.NewEncoder(&b)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if e = enc.Encode(d); e != nil || b.Len() > MaxOutputBytes {
		return Outputs{}, ErrInvalid
	}
	svg, e := renderSVG(d)
	if e != nil || len(svg) > MaxOutputBytes {
		return Outputs{}, ErrInvalid
	}
	return Outputs{JSON: b.Bytes(), SVG: svg}, nil
}
