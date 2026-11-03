package intakemodel

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"runtime"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/canonical"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/compiler"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/intake"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/jsonformat"
)

func digest(raw []byte) string {
	h := sha256.Sum256(raw)
	return hex.EncodeToString(h[:])
}

// inspect delegates complete request validation before extracting the one
// observation this bridge understands. Requirements and suite text stay opaque.
func inspect(raw []byte) (*Report, error) {
	if len(raw) > MaxSourceBytes {
		return nil, ErrLimit
	}
	if runtime.Version() != compiler.ToolchainVersion {
		return nil, ErrToolchain
	}
	native, err := intake.Execute(raw)
	if err != nil {
		return nil, ErrInvalid
	}
	var accepted struct {
		Observed struct {
			Model *string `json:"model"`
		} `json:"observed"`
	}
	// This is extraction from already accepted input, not a second intake schema.
	if json.Unmarshal(raw, &accepted) != nil {
		return nil, ErrInvalid
	}
	r := &Report{
		FormatVersion: ReportVersion, Mode: "example-only", Authority: "none",
		IntakeReportSHA256: digest(native), IntakeReportJSON: string(native),
		Schema:          SchemaReference{Path: SchemaPath, ID: SchemaID, SourceSHA256: SchemaSHA256, Validation: "not-run"},
		Model:           ModelCheck{Status: "missing"},
		ExecutionStatus: "not-run", AcceptanceStatus: "not-evaluated", ReviewStatus: "not-established",
	}
	if accepted.Observed.Model == nil {
		return r, nil
	}
	source := []byte(*accepted.Observed.Model)
	sourceHash := digest(source)
	r.Model.SourceSHA256 = &sourceHash
	m, err := jsonformat.Decode(source)
	if err != nil {
		problem := "model-validation-failed"
		r.Model.Status = "rejected"
		r.Model.Problem = &problem
		return r, nil
	}
	canonicalHash, err := canonical.Digest(m)
	if err != nil {
		return nil, ErrArtifact
	}
	normalized, err := jsonformat.Encode(m)
	if err != nil {
		return nil, ErrArtifact
	}
	r.Model.Status = "accepted"
	r.Model.CanonicalModelSHA256 = &canonicalHash
	r.Model.Normalized = &Artifact{SourceSHA256: digest(normalized), SourceJSON: string(normalized)}
	return r, nil
}

func encode(r *Report) ([]byte, error) {
	// Preserve default HTML/control escaping. Both inner JSON artifacts are
	// strings whose identities apply to decoded bytes, including their final LF.
	b, err := json.Marshal(r)
	if err != nil {
		return nil, ErrArtifact
	}
	if len(b)+1 > MaxReportBytes {
		return nil, ErrLimit
	}
	return append(b, '\n'), nil
}

// Inspect returns an owned typed report, including missing/rejected model status.
// It does not establish that declarations match, intent is correct, or tests ran.
// Raw input must not be mutated during this call; errors return a nil report.
func Inspect(raw []byte) (*Report, error) {
	r, err := inspect(raw)
	if err != nil {
		return nil, err
	}
	if _, err := encode(r); err != nil {
		return nil, err
	}
	return r, nil
}

// Execute emits a complete deterministic JSON report or nil bytes on error.
// No filesystem access, schema fetch, suite replay or generated-code execution occurs.
func Execute(raw []byte) ([]byte, error) {
	r, err := inspect(raw)
	if err != nil {
		return nil, err
	}
	return encode(r)
}
