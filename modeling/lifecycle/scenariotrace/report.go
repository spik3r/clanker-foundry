package scenariotrace

import (
	"encoding/json"
	"sort"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/scenarios"
)

type inputs struct {
	ModelSHA256        string `json:"modelSHA256"`
	ModelDigest        string `json:"modelDigest"`
	RequirementsSHA256 string `json:"requirementsSHA256"`
	RequirementsDigest string `json:"requirementsDigest"`
	CatalogSHA256      string `json:"catalogSHA256"`
	SuiteSHA256        string `json:"suiteSHA256"`
	ScenarioDigest     string `json:"scenarioDigest"`
	BindingsSHA256     string `json:"bindingsSHA256"`
}
type retainedText struct {
	Text   *string `json:"text"`
	SHA256 *string `json:"sha256"`
}
type requirementRef struct {
	RequirementID   string   `json:"requirementId"`
	ScenarioCaseIDs []string `json:"scenarioCaseIds"`
}
type stepObservation struct {
	StepID        string `json:"stepId"`
	EntryPosition *int   `json:"entryPosition"`
}
type caseObservation struct {
	ScenarioCaseID string            `json:"scenarioCaseId"`
	RequirementIDs []string          `json:"requirementIds"`
	Status         string            `json:"status"`
	Steps          []stepObservation `json:"steps"`
}
type report struct {
	FormatVersion           string            `json:"formatVersion"`
	Mode                    string            `json:"mode"`
	Authority               string            `json:"authority"`
	ExecutionStatus         string            `json:"executionStatus"`
	ReviewStatus            string            `json:"reviewStatus"`
	ReportStatus            string            `json:"reportStatus"`
	Inputs                  inputs            `json:"inputs"`
	TraceReport             retainedText      `json:"traceReport"`
	ScenarioReport          retainedText      `json:"scenarioReport"`
	Links                   []link            `json:"links"`
	Requirements            []requirementRef  `json:"requirements"`
	Cases                   []caseObservation `json:"cases"`
	UnlinkedRequirementIDs  []string          `json:"unlinkedRequirementIds"`
	UnlinkedScenarioCaseIDs []string          `json:"unlinkedScenarioCaseIds"`
}

func retain(text string) retainedText { h := hash(text); return retainedText{&text, &h} }
func newReport(in inputs, trace string, links []link, requirements, cases map[string][]string, suite scenarios.Suite) report {
	r := report{FormatVersion: ReportVersion, Mode: "example-only", Authority: "none", ExecutionStatus: "not-established", ReviewStatus: "not-established", ReportStatus: "not-supplied", Inputs: in, TraceReport: retain(trace), Links: links, Requirements: []requirementRef{}, Cases: []caseObservation{}, UnlinkedRequirementIDs: []string{}, UnlinkedScenarioCaseIDs: []string{}}
	for id, refs := range requirements {
		sort.Strings(refs)
		r.Requirements = append(r.Requirements, requirementRef{id, refs})
		if len(refs) == 0 {
			r.UnlinkedRequirementIDs = append(r.UnlinkedRequirementIDs, id)
		}
	}
	sort.Slice(r.Requirements, func(i, j int) bool { return r.Requirements[i].RequirementID < r.Requirements[j].RequirementID })
	for _, c := range suite.Cases {
		refs := cases[c.ID]
		sort.Strings(refs)
		obs := caseObservation{ScenarioCaseID: c.ID, RequirementIDs: refs, Status: "not-supplied", Steps: []stepObservation{}}
		for _, s := range c.Steps {
			obs.Steps = append(obs.Steps, stepObservation{StepID: s.ID})
		}
		r.Cases = append(r.Cases, obs)
		if len(refs) == 0 {
			r.UnlinkedScenarioCaseIDs = append(r.UnlinkedScenarioCaseIDs, c.ID)
		}
	}
	sort.Strings(r.UnlinkedRequirementIDs)
	sort.Strings(r.UnlinkedScenarioCaseIDs)
	return r
}
func encodeReport(r report) ([]byte, error) {
	raw, err := json.Marshal(r)
	if err != nil {
		return nil, ErrBoundary
	}
	if len(raw) >= MaxOutputBytes {
		return nil, ErrLimit
	}
	return append(raw, '\n'), nil
}
