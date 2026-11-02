// Package scenarioexplain explains supplied report assertions and declared model links.
package scenarioexplain

import (
	"bytes"
	"encoding/json"
	"errors"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/canonical"
	lifecycle "github.com/spik3r/clanker-foundry/modeling/lifecycle/generated"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/jsonformat"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/model"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/scenarios"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/strictjson"
)

var ErrInvalid = errors.New("invalid scenario explanation input")

func cloneStrings(x []string) []string {
	if x == nil {
		return nil
	}
	return append([]string{}, x...)
}
func expression(e model.Expr) Expression {
	x := Expression{Kind: e.Kind, Ref: e.Ref, Value: e.Value, Op: e.Op, Args: make([]Expression, 0, len(e.Args))}
	for _, a := range e.Args {
		x.Args = append(x.Args, expression(a))
	}
	return x
}
func fixed(id string) bool {
	return id == "INVALID_INPUT" || id == "ILLEGAL_TRANSITION" || id == "REVISION_LIMIT"
}
func encode(x Explanation) ([]byte, error) {
	var b bytes.Buffer
	e := json.NewEncoder(&b)
	e.SetEscapeHTML(false)
	if e.Encode(x) != nil || b.Len() > MaxOutputBytes {
		return nil, ErrInvalid
	}
	return b.Bytes(), nil
}

// encodedSize uses the same escaping policy as the final encoder. Each model
// expression and each report entry is independently bounded by the input profile.
func encodedSize(v any) (int, error) {
	var b bytes.Buffer
	enc := json.NewEncoder(&b)
	enc.SetEscapeHTML(false)
	if enc.Encode(v) != nil || b.Len() > MaxOutputBytes {
		return 0, ErrInvalid
	}
	return b.Len(), nil
}

// Explain reports caller-supplied assertions and declared links, not execution.
func Explain(input model.Model, report scenarios.Report) (Explanation, error) {
	m, e := model.Normalize(input)
	if e != nil {
		return Explanation{}, ErrInvalid
	}
	for _, g := range m.Guards {
		if fixed(g.ID) {
			return Explanation{}, ErrInvalid
		}
	}
	digest, e := canonical.Digest(m)
	if e != nil || digest != lifecycle.ModelDigest || report.ModelDigest != digest {
		return Explanation{}, ErrInvalid
	}
	if _, e = scenarios.EncodeReport(report); e != nil {
		return Explanation{}, ErrInvalid
	}
	guards := make(map[string]model.Guard, len(m.Guards))
	requirements := make(map[string]string, len(m.Requirements))
	transitions := make(map[string]model.Transition, len(m.Transitions))
	for _, g := range m.Guards {
		guards[g.ID] = g
	}
	for _, r := range m.Requirements {
		requirements[r.ID] = r.Text
	}
	for _, t := range m.Transitions {
		transitions[t.ID] = t
	}
	out := Explanation{FormatVersion: FormatVersion, Status: StatusCallerAssertions, ModelDigest: digest, ScenarioDigest: report.ScenarioDigest, Matched: report.Matched, CasesCompleted: report.CasesCompleted, StepsExecuted: report.StepsExecuted, Entries: make([]Entry, 0, len(report.Entries))}
	if report.SourceDigest != nil {
		x := *report.SourceDigest
		out.SourceDigest = &x
	}
	budget, err := encodedSize(out)
	if err != nil {
		return Explanation{}, ErrInvalid
	}
	// Empty entries array already occupies two bytes; a first encoded entry
	// replaces no comma, while each later entry adds one. The helper includes LF.
	for _, entry := range report.Entries {
		actual := entry.Actual
		if actual.Mode != scenarios.Mode || actual.Authority != scenarios.Authority || actual.ModelDigest != digest {
			return Explanation{}, ErrInvalid
		}
		t, known := transitions[actual.TransitionID]
		compatible := known && t.From == string(entry.Input.State) && t.Trigger == entry.Command.Type
		item := Entry{CaseID: entry.CaseID, StepID: entry.StepID, Matched: entry.Matched, Differences: cloneStrings(entry.Differences), ExpectedTransitionID: entry.Expected.TransitionID, ExpectedRefusals: cloneStrings(entry.Expected.Refusals), ActualOK: actual.OK, ActualTransitionID: actual.TransitionID, ActualRefusals: cloneStrings(actual.Refusals), Guards: []GuardLink{}}
		if actual.OK {
			if !compatible || len(actual.Refusals) != 0 {
				return Explanation{}, ErrInvalid
			}
		} else {
			if len(actual.Refusals) == 0 {
				return Explanation{}, ErrInvalid
			}
			if fixed(actual.Refusals[0]) {
				if len(actual.Refusals) != 1 {
					return Explanation{}, ErrInvalid
				}
				item.RuntimeReason = actual.Refusals[0]
				switch item.RuntimeReason {
				case "INVALID_INPUT":
					if actual.TransitionID != "" || actual.Task != nil {
						return Explanation{}, ErrInvalid
					}
				case "ILLEGAL_TRANSITION":
					if actual.TransitionID != "" {
						return Explanation{}, ErrInvalid
					}
				case "REVISION_LIMIT":
					if !compatible {
						return Explanation{}, ErrInvalid
					}
				}
			} else {
				if !compatible {
					return Explanation{}, ErrInvalid
				}
				allowed := make(map[string]bool, len(t.Guards))
				for _, id := range t.Guards {
					allowed[id] = true
				}
				guardBytes := 0
				for i, id := range actual.Refusals {
					g, ok := guards[id]
					if !ok || !allowed[id] || fixed(id) || (i > 0 && actual.Refusals[i-1] >= id) {
						return Explanation{}, ErrInvalid
					}
					link := GuardLink{GuardID: id, RequirementID: g.Requirement, RequirementText: requirements[g.Requirement], Expression: expression(g.Expr), TransitionID: t.ID}
					n, err := encodedSize(link)
					if err != nil {
						return Explanation{}, ErrInvalid
					}
					guardBytes += n
					if budget+guardBytes > MaxOutputBytes {
						return Explanation{}, ErrInvalid
					}
					item.Guards = append(item.Guards, link)
				}
			}
		}
		n, err := encodedSize(item)
		if err != nil {
			return Explanation{}, ErrInvalid
		}
		if len(out.Entries) == 0 {
			budget += n - 1
		} else {
			budget += n
		}
		if budget > MaxOutputBytes {
			return Explanation{}, ErrInvalid
		}
		out.Entries = append(out.Entries, item)
	}
	if _, e = encode(out); e != nil {
		return Explanation{}, ErrInvalid
	}
	return out, nil
}

type request struct {
	FormatVersion string          `json:"formatVersion"`
	Model         json.RawMessage `json:"model"`
	Report        json.RawMessage `json:"report"`
}

// Execute decodes one strict request and returns a fresh complete JSON explanation.
func Execute(raw []byte) ([]byte, error) {
	var req request
	if strictjson.Decode(raw, &req, MaxRequestBytes, MaxRequestDepth) != nil || req.FormatVersion != RequestFormatVersion || len(req.Model) > MaxModelBytes || len(req.Report) > scenarios.MaxReportBytes {
		return nil, ErrInvalid
	}
	m, e := jsonformat.Decode(req.Model)
	if e != nil {
		return nil, ErrInvalid
	}
	r, e := scenarios.DecodeReport(req.Report)
	if e != nil {
		return nil, ErrInvalid
	}
	x, e := Explain(m, r)
	if e != nil {
		return nil, ErrInvalid
	}
	return encode(x)
}
