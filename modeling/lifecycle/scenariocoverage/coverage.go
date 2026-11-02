// Package scenariocoverage inventories caller-supplied report assertions. It
// neither executes a reducer nor establishes coverage, correctness or authority.
package scenariocoverage

import (
	"bytes"
	"encoding/json"
	"errors"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/jsonformat"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/model"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/scenarioexplain"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/scenarios"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/strictjson"
)

const (
	FormatVersion        = "lifecycle-scenario-assertion-inventory/v1"
	RequestFormatVersion = "lifecycle-scenario-assertion-inventory-request/v1"
	MaxRequestBytes      = scenarioexplain.MaxRequestBytes
	MaxOutputBytes       = 4 << 20
)

var ErrInvalid = errors.New("invalid scenario assertion inventory input")

type Reference struct {
	CaseID string `json:"caseId"`
	StepID string `json:"stepId"`
}
type Guard struct {
	ID       string      `json:"id"`
	Refusals []Reference `json:"refusals"`
}
type Transition struct {
	ID            string      `json:"id"`
	Successes     []Reference `json:"successes"`
	RevisionLimit []Reference `json:"revisionLimit"`
	Guards        []Guard     `json:"guards"`
}
type Inventory struct {
	FormatVersion     string       `json:"formatVersion"`
	Status            string       `json:"status"`
	ModelDigest       string       `json:"modelDigest"`
	ScenarioDigest    string       `json:"scenarioDigest"`
	SourceDigest      *string      `json:"sourceDigest"`
	InputMatched      bool         `json:"inputMatched"`
	CasesCompleted    int          `json:"casesCompleted"`
	StepsExecuted     int          `json:"stepsExecuted"`
	Transitions       []Transition `json:"transitions"`
	InvalidInput      []Reference  `json:"invalidInput"`
	IllegalTransition []Reference  `json:"illegalTransition"`
	Excluded          []Reference  `json:"excluded"`
}

// Analyze validates the entire supplied report and model through the existing
// explanation boundary before creating an inventory. It never calls a reducer.
// The result owns its slices; callers must not mutate inputs concurrently.
func Analyze(input model.Model, report scenarios.Report) (Inventory, error) {
	explanation, err := scenarioexplain.Explain(input, report)
	if err != nil {
		return Inventory{}, ErrInvalid
	}
	m, err := model.Normalize(input)
	if err != nil {
		return Inventory{}, ErrInvalid
	}
	out := Inventory{FormatVersion: FormatVersion, Status: scenarioexplain.StatusCallerAssertions,
		ModelDigest: explanation.ModelDigest, ScenarioDigest: explanation.ScenarioDigest,
		InputMatched: explanation.Matched, CasesCompleted: explanation.CasesCompleted,
		StepsExecuted: explanation.StepsExecuted, Transitions: []Transition{},
		InvalidInput: []Reference{}, IllegalTransition: []Reference{}, Excluded: []Reference{}}
	if explanation.SourceDigest != nil {
		d := *explanation.SourceDigest
		out.SourceDigest = &d
	}
	transitionIndex := make(map[string]int, len(m.Transitions))
	guardIndex := make(map[string]map[string]int, len(m.Transitions))
	for _, t := range m.Transitions {
		transitionIndex[t.ID] = len(out.Transitions)
		guards := make([]Guard, 0, len(t.Guards))
		guardIndex[t.ID] = make(map[string]int, len(t.Guards))
		for _, id := range t.Guards {
			guardIndex[t.ID][id] = len(guards)
			guards = append(guards, Guard{ID: id, Refusals: []Reference{}})
		}
		out.Transitions = append(out.Transitions, Transition{ID: t.ID, Successes: []Reference{}, RevisionLimit: []Reference{}, Guards: guards})
	}
	for _, entry := range explanation.Entries {
		ref := Reference{CaseID: entry.CaseID, StepID: entry.StepID}
		if !entry.Matched {
			out.Excluded = append(out.Excluded, ref)
			continue
		}
		switch {
		case entry.ActualOK:
			i, ok := transitionIndex[entry.ActualTransitionID]
			if !ok {
				return Inventory{}, ErrInvalid
			}
			out.Transitions[i].Successes = append(out.Transitions[i].Successes, ref)
		case entry.RuntimeReason == "INVALID_INPUT":
			out.InvalidInput = append(out.InvalidInput, ref)
		case entry.RuntimeReason == "ILLEGAL_TRANSITION":
			out.IllegalTransition = append(out.IllegalTransition, ref)
		case entry.RuntimeReason == "REVISION_LIMIT":
			i, ok := transitionIndex[entry.ActualTransitionID]
			if !ok {
				return Inventory{}, ErrInvalid
			}
			out.Transitions[i].RevisionLimit = append(out.Transitions[i].RevisionLimit, ref)
		default:
			i, ok := transitionIndex[entry.ActualTransitionID]
			if !ok || entry.RuntimeReason != "" || len(entry.Guards) == 0 {
				return Inventory{}, ErrInvalid
			}
			for _, g := range entry.Guards {
				j, ok := guardIndex[entry.ActualTransitionID][g.GuardID]
				if !ok {
					return Inventory{}, ErrInvalid
				}
				out.Transitions[i].Guards[j].Refusals = append(out.Transitions[i].Guards[j].Refusals, ref)
			}
		}
	}
	if _, err := encode(out); err != nil {
		return Inventory{}, ErrInvalid
	}
	return out, nil
}

func encode(v Inventory) ([]byte, error) {
	var b bytes.Buffer
	e := json.NewEncoder(&b)
	e.SetEscapeHTML(false)
	if e.Encode(v) != nil || b.Len() > MaxOutputBytes {
		return nil, ErrInvalid
	}
	return b.Bytes(), nil
}

type request struct {
	FormatVersion string          `json:"formatVersion"`
	Model         json.RawMessage `json:"model"`
	Report        json.RawMessage `json:"report"`
}

// Execute accepts the bounded strict wrapper and returns complete owned JSON.
func Execute(raw []byte) ([]byte, error) {
	var req request
	if strictjson.Decode(raw, &req, MaxRequestBytes, scenarioexplain.MaxRequestDepth) != nil || req.FormatVersion != RequestFormatVersion || len(req.Model) > scenarioexplain.MaxModelBytes || len(req.Report) > scenarios.MaxReportBytes {
		return nil, ErrInvalid
	}
	m, err := jsonformat.Decode(req.Model)
	if err != nil {
		return nil, ErrInvalid
	}
	r, err := scenarios.DecodeReport(req.Report)
	if err != nil {
		return nil, ErrInvalid
	}
	x, err := Analyze(m, r)
	if err != nil {
		return nil, ErrInvalid
	}
	return encode(x)
}
