package modelpatch

import (
	"bytes"
	"encoding/json"
	"io"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/jsonformat"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/model"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/strictjson"
)

// DTOs describe only the patch carrier, not another model semantics engine.
type exprDTO struct {
	Kind  string    `json:"kind"`
	Ref   string    `json:"ref"`
	Value string    `json:"value"`
	Op    string    `json:"op"`
	Args  []exprDTO `json:"args"`
}
type rootDTO struct {
	ID      string `json:"id"`
	Initial string `json:"initial"`
}
type stateDTO struct {
	ID       string `json:"id"`
	Label    string `json:"label"`
	Terminal bool   `json:"terminal"`
}
type requirementDTO struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}
type guardDTO struct {
	ID          string  `json:"id"`
	Requirement string  `json:"requirement"`
	Expr        exprDTO `json:"expr"`
}
type effectDTO struct {
	Op string `json:"op"`
}
type transitionDTO struct {
	ID      string      `json:"id"`
	From    string      `json:"from"`
	Trigger string      `json:"trigger"`
	To      string      `json:"to"`
	Guards  []string    `json:"guards"`
	Effects []effectDTO `json:"effects"`
}
type operationDTO struct {
	Action      Action          `json:"action"`
	Kind        Kind            `json:"kind"`
	ID          string          `json:"id"`
	Root        *rootDTO        `json:"root"`
	State       *stateDTO       `json:"state"`
	Requirement *requirementDTO `json:"requirement"`
	Guard       *guardDTO       `json:"guard"`
	Transition  *transitionDTO  `json:"transition"`
}
type patchDTO struct {
	PatchVersion   string         `json:"patchVersion"`
	ProfileVersion string         `json:"profileVersion"`
	BeforeDigest   string         `json:"beforeDigest"`
	Operations     []operationDTO `json:"operations"`
}
type requestDTO struct {
	FormatVersion  string          `json:"formatVersion"`
	PatchVersion   string          `json:"patchVersion"`
	ProfileVersion string          `json:"profileVersion"`
	Before         json.RawMessage `json:"before"`
	BeforeDigest   string          `json:"beforeDigest"`
	Operations     []operationDTO  `json:"operations"`
}

func encode(v any) ([]byte, error) {
	var b bytes.Buffer
	e := json.NewEncoder(&b)
	e.SetEscapeHTML(false)
	if e.Encode(v) != nil {
		return nil, ErrInvalid
	}
	return b.Bytes(), nil
}
func exprToDTO(e model.Expr) exprDTO {
	a := make([]exprDTO, len(e.Args))
	for i, x := range e.Args {
		a[i] = exprToDTO(x)
	}
	return exprDTO{e.Kind, e.Ref, e.Value, e.Op, a}
}
func operationToDTO(o Operation) operationDTO {
	d := operationDTO{Action: o.Action, Kind: o.Kind, ID: o.ID}
	if o.Root != nil {
		d.Root = &rootDTO{o.Root.ID, o.Root.Initial}
	}
	if o.State != nil {
		d.State = &stateDTO{o.State.ID, o.State.Label, o.State.Terminal}
	}
	if o.Requirement != nil {
		d.Requirement = &requirementDTO{o.Requirement.ID, o.Requirement.Text}
	}
	if o.Guard != nil {
		d.Guard = &guardDTO{o.Guard.ID, o.Guard.Requirement, exprToDTO(o.Guard.Expr)}
	}
	if o.Transition != nil {
		t := o.Transition
		effects := []effectDTO{}
		for _, x := range t.Effects {
			effects = append(effects, effectDTO{x.Op})
		}
		d.Transition = &transitionDTO{t.ID, t.From, t.Trigger, t.To, append([]string{}, t.Guards...), effects}
	}
	return d
}

// Called only after complete bounded shape traversal; cycles cannot reach here.
func patchBytes(p Patch) ([]byte, error) {
	ops := []operationDTO{}
	for _, o := range p.Operations {
		ops = append(ops, operationToDTO(o))
	}
	b, e := encode(patchDTO{PatchVersion, p.ProfileVersion, p.BeforeDigest, ops})
	return bytes.TrimSuffix(b, []byte{'\n'}), e
}
func exprFromDTO(d exprDTO, depth int, nodes *int) (model.Expr, error) {
	*nodes++
	if depth > MaxExprDepth || *nodes > MaxExprNodes || len(d.Args) > MaxExprArgs {
		return model.Expr{}, ErrLimit
	}
	for _, x := range []struct {
		s   string
		max int
	}{{d.Kind, MaxExprAtomBytes}, {d.Ref, MaxExprRefBytes}, {d.Value, MaxRecordTextBytes}, {d.Op, MaxExprAtomBytes}} {
		if e := scalar(x.s, x.max); e != nil {
			return model.Expr{}, e
		}
	}
	e := model.Expr{Kind: d.Kind, Ref: d.Ref, Value: d.Value, Op: d.Op, Args: make([]model.Expr, len(d.Args))}
	for i, a := range d.Args {
		child, err := exprFromDTO(a, depth+1, nodes)
		if err != nil {
			return model.Expr{}, err
		}
		e.Args[i] = child
	}
	return e, nil
}
func operationFromDTO(d operationDTO, nodes *int) (Operation, error) {
	o := Operation{Action: d.Action, Kind: d.Kind, ID: d.ID}
	// Reject unknown actions/kinds and invalid payload unions before traversing
	// their record data, using the same closed union check as the typed API.
	shape := o
	if d.Root != nil {
		shape.Root = &Root{}
	}
	if d.State != nil {
		shape.State = &model.State{}
	}
	if d.Requirement != nil {
		shape.Requirement = &model.Requirement{}
	}
	if d.Guard != nil {
		shape.Guard = &model.Guard{}
	}
	if d.Transition != nil {
		shape.Transition = &model.Transition{}
	}
	if e := union(shape); e != nil {
		return Operation{}, e
	}
	if d.Root != nil {
		o.Root = &Root{d.Root.ID, d.Root.Initial}
	}
	if d.State != nil {
		o.State = &model.State{ID: d.State.ID, Label: d.State.Label, Terminal: d.State.Terminal}
	}
	if d.Requirement != nil {
		o.Requirement = &model.Requirement{ID: d.Requirement.ID, Text: d.Requirement.Text}
	}
	if d.Guard != nil {
		o.Guard = &model.Guard{ID: d.Guard.ID, Requirement: d.Guard.Requirement}
	}
	if d.Transition != nil {
		t := d.Transition
		if e := transitionHead(t.ID, t.From, t.Trigger, t.To, o.ID, len(t.Guards), len(t.Effects)); e != nil {
			return Operation{}, e
		}
		effects := make([]model.Effect, len(t.Effects))
		for i, x := range t.Effects {
			effects[i] = model.Effect{Op: x.Op}
		}
		o.Transition = &model.Transition{ID: t.ID, From: t.From, Trigger: t.Trigger, To: t.To, Guards: t.Guards, Effects: effects}
	}
	if err := operationBase(o); err != nil {
		return Operation{}, err
	}
	if d.Guard != nil {
		e, err := exprFromDTO(d.Guard.Expr, 0, nodes)
		if err != nil {
			return Operation{}, err
		}
		o.Guard.Expr = e
	}
	return o, nil
}
func encodeResult(r Result) ([]byte, error) {
	after, e := jsonformat.Encode(r.After)
	if e != nil {
		return nil, ErrInvalid
	}
	dto := struct {
		ReportVersion  string          `json:"reportVersion"`
		ProfileVersion string          `json:"profileVersion"`
		Mode           string          `json:"mode"`
		Authority      string          `json:"authority"`
		BeforeDigest   string          `json:"beforeDigest"`
		AfterDigest    string          `json:"afterDigest"`
		After          json.RawMessage `json:"after"`
		Review         any             `json:"review"`
	}{r.ReportVersion, r.ProfileVersion, r.Mode, r.Authority, r.BeforeDigest, r.AfterDigest, after, r.Review}
	b, e := encode(dto)
	if e != nil {
		return nil, e
	}
	if len(b) > MaxResultBytes {
		return nil, ErrLimit
	}
	return b, nil
}

// Structural wire budget only; strictjson remains responsible for exact keys,
// duplicates, scalar types and lossless Unicode checks.
func requestDepth(raw []byte) error {
	if !json.Valid(raw) {
		return ErrInvalid
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	depth := 0
	for {
		tok, e := d.Token()
		if e == io.EOF {
			return nil
		}
		if e != nil {
			return ErrInvalid
		}
		if delim, ok := tok.(json.Delim); ok {
			if delim == '}' || delim == ']' {
				depth--
				continue
			}
			if depth > MaxJSONDepth {
				return ErrLimit
			}
			depth++
		} else if depth > MaxJSONDepth {
			return ErrLimit
		}
	}
}

// Execute computes its own result. It accepts no caller report, paths or macros.
func Execute(raw []byte) ([]byte, error) {
	if len(raw) > MaxRequestBytes {
		return nil, ErrLimit
	}
	if e := requestDepth(raw); e != nil {
		return nil, e
	}
	var d requestDTO
	if e := strictjson.Decode(raw, &d, MaxRequestBytes, MaxJSONDepth); e != nil {
		return nil, ErrInvalid
	}
	if d.FormatVersion != FormatVersion || d.PatchVersion != PatchVersion || d.ProfileVersion != ProfileVersion {
		return nil, ErrInvalid
	}
	before, e := jsonformat.Decode(d.Before)
	if e != nil {
		return nil, ErrInvalid
	}
	if _, e := jsonformat.Encode(before); e != nil {
		return nil, ErrInvalid
	}
	if !digestSyntax(d.BeforeDigest) {
		return nil, ErrInvalid
	}
	if len(d.Operations) > MaxOperations {
		return nil, ErrLimit
	}
	p := Patch{Version: 1, ProfileVersion: d.ProfileVersion, BeforeDigest: d.BeforeDigest, Operations: []Operation{}}
	nodes := 0
	for _, x := range d.Operations {
		o, e := operationFromDTO(x, &nodes)
		if e != nil {
			return nil, e
		}
		p.Operations = append(p.Operations, o)
	}
	result, e := Plan(before, p)
	if e != nil {
		return nil, e
	}
	return encodeResult(result)
}
