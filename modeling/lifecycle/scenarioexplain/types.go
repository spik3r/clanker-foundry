// Package scenarioexplain proposes a pure explanation of supplied assertions.
// These types do not establish execution, review, authority or coverage.
package scenarioexplain

const (
	FormatVersion          = "lifecycle-scenario-explanation/v1"
	RequestFormatVersion   = "lifecycle-scenario-explain-request/v1"
	StatusCallerAssertions = "caller-supplied-assertions"
	MaxRequestBytes        = 10 << 20
	MaxRequestDepth        = 66
	MaxModelBytes          = 1 << 20
	MaxModelDepth          = 64
	MaxReportDepth         = 64
	MaxOutputBytes         = 8 << 20
)

// Explanation contains supplied report assertions and declared model links.
// It must own its nested slices and expressions; values may contain private IDs
// or model text and expressions. Task, actor, owner, review and acceptance
// fields are not copied, but equivalent sensitive text can occur in literals.
type Explanation struct {
	FormatVersion  string  `json:"formatVersion"`
	Status         string  `json:"status"`
	ModelDigest    string  `json:"modelDigest"`
	ScenarioDigest string  `json:"scenarioDigest"`
	SourceDigest   *string `json:"sourceDigest"`
	Matched        bool    `json:"matched"`
	CasesCompleted int     `json:"casesCompleted"`
	StepsExecuted  int     `json:"stepsExecuted"`
	Entries        []Entry `json:"entries"`
}

// Entry preserves literal expected values and mismatch fields without asserting
// that any actual result occurred. Expected reasons are not guard definitions.
type Entry struct {
	CaseID               string      `json:"caseId"`
	StepID               string      `json:"stepId"`
	Matched              bool        `json:"matched"`
	Differences          []string    `json:"differences"`
	ExpectedTransitionID string      `json:"expectedTransitionId"`
	ExpectedRefusals     []string    `json:"expectedRefusals"`
	ActualOK             bool        `json:"actualOk"`
	ActualTransitionID   string      `json:"actualTransitionId"`
	ActualRefusals       []string    `json:"actualRefusals"`
	RuntimeReason        string      `json:"runtimeReason"`
	Guards               []GuardLink `json:"guards"`
}

// GuardLink records only the declaration named in a supplied actual refusal.
// It does not assert guard truth or execution.
type GuardLink struct {
	GuardID         string     `json:"guardId"`
	RequirementID   string     `json:"requirementId"`
	RequirementText string     `json:"requirementText"`
	Expression      Expression `json:"expression"`
	TransitionID    string     `json:"transitionId"`
}

// Expression is a projection of the declared IR, not an evaluator. Args is
// always a non-nil array, including for leaves; all other fields are strings.
type Expression struct {
	Kind  string       `json:"kind"`
	Ref   string       `json:"ref"`
	Value string       `json:"value"`
	Op    string       `json:"op"`
	Args  []Expression `json:"args"`
}
