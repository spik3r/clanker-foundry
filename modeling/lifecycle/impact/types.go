// Package impact compares model content and declared links for human review.
// It does not evaluate behavior, run scenarios, or grant review authority.
package impact

import "errors"

const (
	FormatVersion  = "lifecycle-model-diff-json/v1"
	ReportVersion  = "lifecycle-model-diff-report/v1"
	MaxSourceBytes = 3 << 20
	MaxReportBytes = 1 << 20
	MaxDepth       = 66
	MaxBindings    = 128
	Mode           = "example-only"
	Authority      = "none"
)

var (
	ErrInvalidInput = errors.New("invalid model review request")
	ErrReportLimit  = errors.New("model review report limit")
)

// Binding is caller-asserted metadata, not an inspected or executable scenario.
// IDs use model.Identifier and form a separate namespace from model elements.
type Binding struct {
	ID          string `json:"id"`
	ModelDigest string `json:"modelDigest"`
}

// Change is keyed by Kind and ID. Only the model root uses an empty ID.
// Added/removed elements have an empty Fields list; renames are remove+add.
type Change struct {
	Kind   string   `json:"kind"`
	ID     string   `json:"id"`
	Type   string   `json:"type"`
	Fields []string `json:"fields"`
}

// LinkedTransition lists every direct declared-link cause in one model graph.
// Its absence does not establish that a transition's behavior is unaffected.
type LinkedTransition struct {
	Side   string   `json:"side"`
	ID     string   `json:"id"`
	Causes []string `json:"causes"`
}

type BindingResult struct {
	ID          string `json:"id"`
	ModelDigest string `json:"modelDigest"`
	Status      string `json:"status"`
}

// Report is ordinary data, not authenticated evidence. Compare returns owned
// collections (empty arrays, never null); mutating a Report creates no authority.
// Compare has no source bytes and leaves all three source digests nil. Execute
// records exact request and embedded JSON-value bytes, not original file hashes.
type Report struct {
	ReportVersion      string             `json:"reportVersion"`
	ProfileVersion     string             `json:"profileVersion"`
	Mode               string             `json:"mode"`
	Authority          string             `json:"authority"`
	BeforeDigest       string             `json:"beforeDigest"`
	AfterDigest        string             `json:"afterDigest"`
	RequestDigest      *string            `json:"requestDigest"`
	BeforeSourceDigest *string            `json:"beforeSourceDigest"`
	AfterSourceDigest  *string            `json:"afterSourceDigest"`
	Changed            bool               `json:"changed"`
	ReviewRequired     bool               `json:"reviewRequired"`
	Changes            []Change           `json:"changes"`
	LinkedTransitions  []LinkedTransition `json:"linkedTransitions"`
	Bindings           []BindingResult    `json:"bindings"`
}
