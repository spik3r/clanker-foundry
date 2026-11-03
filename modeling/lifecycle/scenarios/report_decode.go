package scenarios

import (
	lifecycle "github.com/spik3r/clanker-foundry/modeling/lifecycle/generated"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/strictjson"
)

// Decode-only DTO keeps nullable actual refusals distinct from an empty array.
type reportActualDTO struct {
	Mode         string    `json:"mode"`
	Authority    string    `json:"authority"`
	ModelDigest  string    `json:"modelDigest"`
	OK           bool      `json:"ok"`
	TransitionID string    `json:"transitionId"`
	Task         *taskDTO  `json:"task"`
	Refusals     *[]string `json:"refusals"`
}
type reportEntryDTO struct {
	CaseID      string          `json:"caseId"`
	StepID      string          `json:"stepId"`
	Input       taskDTO         `json:"input"`
	Command     commandDTO      `json:"command"`
	Expected    expectedDTO     `json:"expected"`
	Actual      reportActualDTO `json:"actual"`
	Matched     bool            `json:"matched"`
	Differences []string        `json:"differences"`
}
type reportDTO struct {
	FormatVersion   string           `json:"formatVersion"`
	ScenarioVersion string           `json:"scenarioVersion"`
	ProfileVersion  string           `json:"profileVersion"`
	Mode            string           `json:"mode"`
	Authority       string           `json:"authority"`
	ModelDigest     string           `json:"modelDigest"`
	ScenarioDigest  string           `json:"scenarioDigest"`
	SourceDigest    *string          `json:"sourceDigest"`
	Matched         bool             `json:"matched"`
	CasesCompleted  int              `json:"casesCompleted"`
	StepsExecuted   int              `json:"stepsExecuted"`
	Entries         []reportEntryDTO `json:"entries"`
}

// DecodeReport validates the report wire form and existing report consistency.
// It does not establish that any report was produced by execution.
func DecodeReport(raw []byte) (Report, error) {
	var d reportDTO
	if strictjson.Decode(raw, &d, MaxReportBytes, MaxDepth) != nil || d.FormatVersion != ReportVersion || d.ScenarioVersion != ScenarioVersion || d.ProfileVersion != lifecycle.ProfileVersion || d.Mode != Mode || d.Authority != Authority {
		return Report{}, ErrReportLimit
	}
	r := Report{ModelDigest: d.ModelDigest, ScenarioDigest: d.ScenarioDigest, SourceDigest: d.SourceDigest, Matched: d.Matched, CasesCompleted: d.CasesCompleted, StepsExecuted: d.StepsExecuted, Entries: make([]Entry, 0, len(d.Entries))}
	if d.SourceDigest != nil {
		v := *d.SourceDigest
		r.SourceDigest = &v
	}
	for _, x := range d.Entries {
		in, e := fromTask(x.Input)
		if e != nil {
			return Report{}, ErrReportLimit
		}
		cmd, e := fromCommand(x.Command)
		if e != nil {
			return Report{}, ErrReportLimit
		}
		want := Expected{OK: x.Expected.OK, TransitionID: x.Expected.TransitionID, Refusals: copyStrings(x.Expected.Refusals)}
		if x.Expected.Task != nil {
			t, e := fromTask(*x.Expected.Task)
			if e != nil {
				return Report{}, ErrReportLimit
			}
			want.Task = &t
		}
		actual := lifecycle.Result{Mode: x.Actual.Mode, Authority: x.Actual.Authority, ModelDigest: x.Actual.ModelDigest, OK: x.Actual.OK, TransitionID: x.Actual.TransitionID}
		if x.Actual.Task != nil {
			t, e := fromTask(*x.Actual.Task)
			if e != nil {
				return Report{}, ErrReportLimit
			}
			actual.Task = &t
		}
		if x.Actual.Refusals != nil {
			actual.Refusals = copyStrings(*x.Actual.Refusals)
		}
		r.Entries = append(r.Entries, Entry{CaseID: x.CaseID, StepID: x.StepID, Input: in, Command: cmd, Expected: want, Actual: actual, Matched: x.Matched, Differences: copyStrings(x.Differences)})
	}
	if _, e := EncodeReport(r); e != nil {
		return Report{}, ErrReportLimit
	}
	return r, nil
}
