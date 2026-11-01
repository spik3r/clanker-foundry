// Package intakemodel connects intake observations to the existing model validator.
// It does not verify intent, author identity, acceptance, review or execution.
package intakemodel

import (
	"errors"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/intake"
)

const (
	ReportVersion  = "lifecycle-intake-model-report/v1"
	MaxSourceBytes = intake.MaxSourceBytes
	MaxReportBytes = 16 << 20
	SchemaPath     = "authoring/lifecycle-v1.schema.json"
	SchemaID       = "https://clanker-foundry.invalid/schemas/lifecycle-json-v1.schema.json"
	SchemaSHA256   = "dc1f75f1ce24b71cb034c0b8b3af62a70d0bfa297e7bbc7b5c5337eb7336a5b7"
)

var (
	ErrInvalid   = errors.New("invalid model intake request")
	ErrLimit     = errors.New("model intake limit exceeded")
	ErrToolchain = errors.New("model intake requires pinned Go toolchain")
	ErrArtifact  = errors.New("accepted model artifact encoding failed")
)

// Artifact carries exact UTF-8 JSON bytes as a string, including the final LF.
// SourceSHA256 hashes that decoded string, not its escaped outer representation.
type Artifact struct {
	SourceSHA256 string `json:"sourceSHA256"`
	SourceJSON   string `json:"sourceJSON"`
}

// ModelCheck describes validation of observed bytes, not a declared hash match.
// Accepted models can still have mismatched declarations or unresolved questions.
type ModelCheck struct {
	Status               string    `json:"status"`
	SourceSHA256         *string   `json:"sourceSHA256"`
	CanonicalModelSHA256 *string   `json:"canonicalModelSHA256"`
	Normalized           *Artifact `json:"normalized"`
	Problem              *string   `json:"problem"`
}

// SchemaReference identifies a pinned authoring aid. It is never fetched or run.
type SchemaReference struct {
	Path         string `json:"path"`
	ID           string `json:"id"`
	SourceSHA256 string `json:"sourceSHA256"`
	Validation   string `json:"validation"`
}

type Report struct {
	FormatVersion      string          `json:"formatVersion"`
	Mode               string          `json:"mode"`
	Authority          string          `json:"authority"`
	IntakeReportSHA256 string          `json:"intakeReportSHA256"`
	IntakeReportJSON   string          `json:"intakeReportJSON"`
	Schema             SchemaReference `json:"schema"`
	Model              ModelCheck      `json:"model"`
	ExecutionStatus    string          `json:"executionStatus"`
	AcceptanceStatus   string          `json:"acceptanceStatus"`
	ReviewStatus       string          `json:"reviewStatus"`
}
