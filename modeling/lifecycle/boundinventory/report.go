package boundinventory

import (
	"bytes"
	"encoding/json"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/model"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/scenarioexplain"
)

type inputs struct {
	ModelSHA256    string `json:"modelSHA256"`
	SuiteSHA256    string `json:"suiteSHA256"`
	ReportSHA256   string `json:"reportSHA256"`
	ModelDigest    string `json:"modelDigest"`
	ScenarioDigest string `json:"scenarioDigest"`
}
type reference struct {
	CaseID        string `json:"caseId"`
	StepID        string `json:"stepId"`
	EntryPosition int    `json:"entryPosition"`
}
type guard struct {
	ID       string      `json:"id"`
	Refusals []reference `json:"refusals"`
}
type transition struct {
	ID            string      `json:"id"`
	Successes     []reference `json:"successes"`
	RevisionLimit []reference `json:"revisionLimit"`
	Guards        []guard     `json:"guards"`
}
type report struct {
	FormatVersion     string       `json:"formatVersion"`
	Mode              string       `json:"mode"`
	Authority         string       `json:"authority"`
	ExecutionStatus   string       `json:"executionStatus"`
	ReviewStatus      string       `json:"reviewStatus"`
	Status            string       `json:"status"`
	Inputs            inputs       `json:"inputs"`
	InputMatched      bool         `json:"inputMatched"`
	CasesCompleted    int          `json:"casesCompleted"`
	StepsExecuted     int          `json:"stepsExecuted"`
	Transitions       []transition `json:"transitions"`
	InvalidInput      []reference  `json:"invalidInput"`
	IllegalTransition []reference  `json:"illegalTransition"`
	Excluded          []reference  `json:"excluded"`
}

func inventory(m model.Model, explanation scenarioexplain.BoundExplanation, in inputs) (report, error) {
	out := report{FormatVersion: ReportVersion, Mode: "example-only", Authority: "none", ExecutionStatus: "not-established", ReviewStatus: "not-established", Status: scenarioexplain.StatusCallerAssertions, Inputs: in, InputMatched: explanation.Matched, CasesCompleted: explanation.CasesCompleted, StepsExecuted: explanation.StepsExecuted, Transitions: []transition{}, InvalidInput: []reference{}, IllegalTransition: []reference{}, Excluded: []reference{}}
	transitions := make(map[string]int, len(m.Transitions))
	guards := make(map[string]map[string]int, len(m.Transitions))
	// Native decoding normalizes transition IDs and associated guard IDs.
	for _, t := range m.Transitions {
		transitions[t.ID] = len(out.Transitions)
		gs := make([]guard, 0, len(t.Guards))
		guards[t.ID] = make(map[string]int, len(t.Guards))
		for _, id := range t.Guards {
			guards[t.ID][id] = len(gs)
			gs = append(gs, guard{id, []reference{}})
		}
		out.Transitions = append(out.Transitions, transition{t.ID, []reference{}, []reference{}, gs})
	}
	for position, e := range explanation.Entries {
		ref := reference{e.CaseID, e.StepID, position}
		if !e.Matched {
			out.Excluded = append(out.Excluded, ref)
			continue
		}
		switch {
		case e.ActualOK:
			i, ok := transitions[e.ActualTransitionID]
			if !ok {
				return report{}, ErrBoundary
			}
			out.Transitions[i].Successes = append(out.Transitions[i].Successes, ref)
		case e.RuntimeReason == "INVALID_INPUT":
			out.InvalidInput = append(out.InvalidInput, ref)
		case e.RuntimeReason == "ILLEGAL_TRANSITION":
			out.IllegalTransition = append(out.IllegalTransition, ref)
		case e.RuntimeReason == "REVISION_LIMIT":
			i, ok := transitions[e.ActualTransitionID]
			if !ok {
				return report{}, ErrBoundary
			}
			out.Transitions[i].RevisionLimit = append(out.Transitions[i].RevisionLimit, ref)
		default:
			i, ok := transitions[e.ActualTransitionID]
			if !ok || e.RuntimeReason != "" || len(e.Guards) == 0 {
				return report{}, ErrBoundary
			}
			for _, g := range e.Guards {
				j, ok := guards[e.ActualTransitionID][g.GuardID]
				if !ok {
					return report{}, ErrBoundary
				}
				out.Transitions[i].Guards[j].Refusals = append(out.Transitions[i].Guards[j].Refusals, ref)
			}
		}
	}
	return out, nil
}
func encode(value report) ([]byte, error) {
	var b bytes.Buffer
	e := json.NewEncoder(&b)
	e.SetEscapeHTML(false)
	if e.Encode(value) != nil {
		return nil, ErrBoundary
	}
	if b.Len() > MaxOutputBytes {
		return nil, ErrLimit
	}
	return b.Bytes(), nil
}
