// Package scenariotraceimpact compares declared requirement-to-case links and
// supplied observations. It establishes neither execution nor review approval.
package scenariotraceimpact

import "errors"

const (
	RequestVersion      = "lifecycle-scenario-trace-impact-request/v1"
	ReportVersion       = "lifecycle-scenario-trace-impact-report/v1"
	MaxRequestBytes     = 64 << 20
	MaxTraceReportBytes = 8 << 20
	MaxOutputBytes      = 128 << 20
)

var (
	ErrInvalid  = errors.New("invalid scenario trace impact input")
	ErrLimit    = errors.New("scenario trace impact limit exceeded")
	ErrBoundary = errors.New("scenario trace impact boundary failed")
)

type request struct {
	FormatVersion string `json:"formatVersion"`
	Before        string `json:"before"`
	After         string `json:"after"`
}

// These projections are decoded only after native validation. They cannot be
// supplied as substitutes for the original sources or native reports.
type snapshot struct {
	Model          string  `json:"model"`
	Requirements   string  `json:"requirements"`
	Catalog        string  `json:"catalog"`
	Suite          string  `json:"suite"`
	Bindings       string  `json:"bindings"`
	ScenarioReport *string `json:"scenarioReport"`
}

type traceInputs struct {
	ModelSHA256        string `json:"modelSHA256"`
	ModelDigest        string `json:"modelDigest"`
	RequirementsSHA256 string `json:"requirementsSHA256"`
	RequirementsDigest string `json:"requirementsDigest"`
	CatalogSHA256      string `json:"catalogSHA256"`
	SuiteSHA256        string `json:"suiteSHA256"`
	ScenarioDigest     string `json:"scenarioDigest"`
	BindingsSHA256     string `json:"bindingsSHA256"`
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
	Status         string            `json:"status"`
	Steps          []stepObservation `json:"steps"`
}

type traceReport struct {
	FormatVersion   string            `json:"formatVersion"`
	Mode            string            `json:"mode"`
	Authority       string            `json:"authority"`
	ExecutionStatus string            `json:"executionStatus"`
	ReviewStatus    string            `json:"reviewStatus"`
	Inputs          traceInputs       `json:"inputs"`
	Requirements    []requirementRef  `json:"requirements"`
	Cases           []caseObservation `json:"cases"`
}

type retained struct {
	RequestSHA256 string `json:"requestSHA256"`
	ReportJSON    string `json:"reportJSON"`
	ReportSHA256  string `json:"reportSHA256"`
}

type reason struct {
	Code           string `json:"code"`
	Side           string `json:"side"`
	ScenarioCaseID string `json:"scenarioCaseId"`
}

type candidate struct {
	RequirementID string   `json:"requirementId"`
	BeforePresent bool     `json:"beforePresent"`
	AfterPresent  bool     `json:"afterPresent"`
	BeforeCaseIDs []string `json:"beforeCaseIds"`
	AfterCaseIDs  []string `json:"afterCaseIds"`
	Reasons       []reason `json:"reasons"`
}

type report struct {
	FormatVersion        string      `json:"formatVersion"`
	Mode                 string      `json:"mode"`
	Authority            string      `json:"authority"`
	ExecutionStatus      string      `json:"executionStatus"`
	ReviewStatus         string      `json:"reviewStatus"`
	RequestSHA256        string      `json:"requestSHA256"`
	Before               retained    `json:"before"`
	After                retained    `json:"after"`
	ScenarioDiff         retained    `json:"scenarioDiff"`
	TraceImpact          retained    `json:"traceImpact"`
	RequestSourceChanged bool        `json:"requestSourceChanged"`
	SourceChanges        []string    `json:"sourceChanges"`
	ReviewRequired       bool        `json:"reviewRequired"`
	Candidates           []candidate `json:"candidates"`
}
