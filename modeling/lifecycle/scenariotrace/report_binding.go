package scenariotrace

import (
	"reflect"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/scenarios"
)

// bindReport accepts a complete match or a prefix ending at its first mismatch.
// Native decoding checks actual envelopes and comparison flags. This pass binds
// those observations to the original suite literals, order and input flow.
func bindReport(runner scenarios.Runner, suite scenarios.Suite, raw string, r *report) error {
	d, err := runner.DecodeReport([]byte(raw))
	if err != nil || d.ModelDigest != r.Inputs.ModelDigest || d.ScenarioDigest != r.Inputs.ScenarioDigest || d.SourceDigest == nil || *d.SourceDigest != r.Inputs.SuiteSHA256 {
		return ErrInvalid
	}
	for i := range r.Cases {
		r.Cases[i].Status = "not-observed"
	}
	position, completed := 0, 0
	for i, c := range suite.Cases {
		current := c.Initial
		for j, s := range c.Steps {
			if position >= len(d.Entries) {
				return ErrInvalid
			}
			e := d.Entries[position]
			if e.CaseID != c.ID || e.StepID != s.ID || !reflect.DeepEqual(e.Input, current) || !reflect.DeepEqual(e.Command, s.Command) || !reflect.DeepEqual(e.Expected, s.Expected) {
				return ErrInvalid
			}
			p := position
			r.Cases[i].Steps[j].EntryPosition = &p
			position++
			if !e.Matched {
				if d.Matched || position != len(d.Entries) || completed != d.CasesCompleted {
					return ErrInvalid
				}
				r.Cases[i].Status = "mismatched"
				r.ReportStatus = "inspected-for-consistency"
				r.ScenarioReport = retain(raw)
				return nil
			}
			// A matched refusal retains input, even when the reported task is null.
			if e.Actual.OK {
				current = *e.Actual.Task
			}
		}
		r.Cases[i].Status = "matched"
		completed++
	}
	if !d.Matched || position != len(d.Entries) || completed != d.CasesCompleted {
		return ErrInvalid
	}
	r.ReportStatus = "inspected-for-consistency"
	r.ScenarioReport = retain(raw)
	return nil
}
