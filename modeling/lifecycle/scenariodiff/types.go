// Package scenariodiff compares model-bound literal fixtures, without executing
// them or treating their expectations as observed results.
package scenariodiff

import "errors"

const (
	Version         = "lifecycle-scenario-diff/v1"
	FormatVersion   = "lifecycle-scenario-diff-json/v1"
	MaxRequestBytes = 5 << 20
	MaxReportBytes  = 1 << 20
	Added           = "added"
	Removed         = "removed"
	Modified        = "modified"
)

var (
	ErrInvalid = errors.New("invalid scenario comparison input")
	ErrLimit   = errors.New("scenario comparison limit exceeded")
)

// Report describes supplied declarations only. Its change flags are not a
// judgment of behavior, compatibility, completeness or review approval.
type Report struct {
	DiffVersion       string         `json:"diffVersion"`
	ProfileVersion    string         `json:"profileVersion"`
	Mode              string         `json:"mode"`
	Authority         string         `json:"authority"`
	BeforeModelDigest string         `json:"beforeModelDigest"`
	AfterModelDigest  string         `json:"afterModelDigest"`
	BeforeSuiteDigest string         `json:"beforeSuiteDigest"`
	AfterSuiteDigest  string         `json:"afterSuiteDigest"`
	Sources           *SourceDigests `json:"sources"`
	ModelChanged      bool           `json:"modelChanged"`
	SuiteChanged      bool           `json:"suiteChanged"`
	ReviewRequired    bool           `json:"reviewRequired"`
	SuiteFields       []string       `json:"suiteFields"`
	CaseOrder         Order          `json:"caseOrder"`
	Cases             []CaseChange   `json:"cases"`
	Steps             []StepChange   `json:"steps"`
}

type SourceDigests struct {
	Request     string `json:"request"`
	BeforeModel string `json:"beforeModel"`
	BeforeSuite string `json:"beforeSuite"`
	AfterModel  string `json:"afterModel"`
	AfterSuite  string `json:"afterSuite"`
}

type Order struct {
	Before  []string `json:"before"`
	After   []string `json:"after"`
	Changed bool     `json:"changed"`
}

type CaseChange struct {
	ID          string   `json:"id"`
	Kind        string   `json:"kind"`
	Fields      []string `json:"fields"`
	BeforeSteps []string `json:"beforeSteps"`
	AfterSteps  []string `json:"afterSteps"`
}

type StepChange struct {
	CaseID string   `json:"caseId"`
	ID     string   `json:"id"`
	Kind   string   `json:"kind"`
	Fields []string `json:"fields"`
}
