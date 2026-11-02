package scenarioexplain

import (
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/model"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/scenarios"
)

const BoundFormatVersion = "lifecycle-bound-scenario-explanation/v1"

// BoundExplanation describes supplied assertions and declared model links for
// an explicit Go runner binding. It proves neither wire origin nor execution,
// compiled-code provenance, review, authority or behavioral coverage.
type BoundExplanation struct {
	FormatVersion  string  `json:"formatVersion"`
	IdentitySource string  `json:"identitySource"`
	Mode           string  `json:"mode"`
	Authority      string  `json:"authority"`
	Status         string  `json:"status"`
	ModelDigest    string  `json:"modelDigest"`
	ScenarioDigest string  `json:"scenarioDigest"`
	SourceDigest   *string `json:"sourceDigest"`
	Matched        bool    `json:"matched"`
	CasesCompleted int     `json:"casesCompleted"`
	StepsExecuted  int     `json:"stepsExecuted"`
	Entries        []Entry `json:"entries"`
}

// ExplainWithRunner projects caller assertions through an explicit model binding.
// It validates the whole report with the runner's encoder but never invokes the
// runner callback. Typed reports cannot establish wire origin or execution.
// Callers must not mutate model/report memory concurrently during the call.
func ExplainWithRunner(input model.Model, report scenarios.Report, runner scenarios.Runner) (BoundExplanation, error) {
	binding, err := runner.ModelBinding()
	if err != nil {
		return BoundExplanation{}, ErrInvalid
	}
	projection, err := explainFor(input, report, binding.ModelDigest, runner.EncodeReport)
	if err != nil {
		return BoundExplanation{}, ErrInvalid
	}
	out := BoundExplanation{
		FormatVersion:  BoundFormatVersion,
		IdentitySource: scenarios.BindingSource,
		Mode:           scenarios.Mode,
		Authority:      scenarios.Authority,
		Status:         StatusCallerAssertions,
		ModelDigest:    projection.ModelDigest,
		ScenarioDigest: projection.ScenarioDigest,
		SourceDigest:   projection.SourceDigest,
		Matched:        projection.Matched,
		CasesCompleted: projection.CasesCompleted,
		StepsExecuted:  projection.StepsExecuted,
		Entries:        projection.Entries,
	}
	// The extra bound-envelope fields must fit too; a legacy-sized projection
	// alone is insufficient. Failure returns no partial typed result.
	if _, err := encodeValue(out); err != nil {
		return BoundExplanation{}, ErrInvalid
	}
	return out, nil
}
