package scenariotraceimpact

import (
	"reflect"
	"slices"
	"sort"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/scenariodiff"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/traceimpact"
)

func sourceChanges(before, after snapshot) []string {
	out := []string{}
	for _, s := range []struct{ name, before, after string }{
		{"bindings", before.Bindings, after.Bindings},
		{"catalog", before.Catalog, after.Catalog},
		{"model", before.Model, after.Model},
		{"requirements", before.Requirements, after.Requirements},
	} {
		if s.before != s.after {
			out = append(out, s.name)
		}
	}
	if !reflect.DeepEqual(before.ScenarioReport, after.ScenarioReport) {
		out = append(out, "scenarioReport")
	}
	if before.Suite != after.Suite {
		out = append(out, "suite")
	}
	return out
}

func requirements(r traceReport) map[string][]string {
	out := make(map[string][]string, len(r.Requirements))
	for _, q := range r.Requirements {
		out[q.RequirementID] = append([]string{}, q.ScenarioCaseIDs...)
		slices.Sort(out[q.RequirementID])
	}
	return out
}

func observations(r traceReport) map[string]caseObservation {
	out := make(map[string]caseObservation, len(r.Cases))
	for _, c := range r.Cases {
		out[c.ScenarioCaseID] = c
	}
	return out
}

func project(before, after traceReport, diff scenariodiff.Report, impact traceimpact.Report) []candidate {
	b, a := requirements(before), requirements(after)
	ids := map[string]bool{}
	for id := range b {
		ids[id] = true
	}
	for id := range a {
		ids[id] = true
	}
	native := map[string]bool{}
	for _, c := range impact.Candidates {
		native[c.ID] = true
	}
	changedCases := map[string]string{}
	for _, c := range diff.Cases {
		changedCases[c.ID] = "scenario-case-" + c.Kind
	}
	bo, ao := observations(before), observations(after)
	changedObservations := map[string]bool{}
	for id, left := range bo {
		if right, ok := ao[id]; ok {
			// Requirement references are intentionally absent from this projection.
			changedObservations[id] = left.Status != right.Status || !reflect.DeepEqual(left.Steps, right.Steps)
		}
	}
	out := []candidate{}
	for id := range ids {
		left, lp := b[id]
		right, rp := a[id]
		c := candidate{RequirementID: id, BeforePresent: lp, AfterPresent: rp, BeforeCaseIDs: append([]string{}, left...), AfterCaseIDs: append([]string{}, right...), Reasons: []reason{}}
		seen := map[reason]bool{}
		add := func(code, side, caseID string) { seen[reason{code, side, caseID}] = true }
		if native[id] {
			add("native-trace-impact", "both", "")
		}
		for _, caseID := range left {
			if !slices.Contains(right, caseID) {
				add("scenario-link-removed", "before", caseID)
			}
			if code, ok := changedCases[caseID]; ok {
				add(code, "before", caseID)
			}
			if changedObservations[caseID] {
				add("scenario-observation-changed", "both", caseID)
			}
		}
		for _, caseID := range right {
			if !slices.Contains(left, caseID) {
				add("scenario-link-added", "after", caseID)
			}
			if code, ok := changedCases[caseID]; ok {
				add(code, "after", caseID)
			}
			if changedObservations[caseID] {
				add("scenario-observation-changed", "both", caseID)
			}
		}
		if len(left)+len(right) > 0 {
			if diff.CaseOrder.Changed {
				add("scenario-case-order-changed", "both", "")
			}
			if len(diff.SuiteFields) > 0 {
				add("scenario-suite-fields-changed", "both", "")
			}
		}
		for r := range seen {
			c.Reasons = append(c.Reasons, r)
		}
		sort.Slice(c.Reasons, func(i, j int) bool {
			l, r := c.Reasons[i], c.Reasons[j]
			if l.Code != r.Code {
				return l.Code < r.Code
			}
			if l.Side != r.Side {
				return l.Side < r.Side
			}
			return l.ScenarioCaseID < r.ScenarioCaseID
		})
		if len(c.Reasons) > 0 {
			out = append(out, c)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].RequirementID < out[j].RequirementID })
	return out
}
