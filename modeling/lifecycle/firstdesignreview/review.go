// Package firstdesignreview composes existing check-only review reports.
// It does not establish review, identity, execution or deployment authority.
package firstdesignreview

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/intakemodel"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/proposalreplay"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/strictjson"
)

const (
	RequestVersion  = "lifecycle-first-design-review-request/v1"
	ReportVersion   = "lifecycle-first-design-review-report/v1"
	MaxRequestBytes = 8 << 20
	MaxReportBytes  = 32 << 20
)

var (
	ErrInvalid  = errors.New("invalid first-design review request")
	ErrLimit    = errors.New("first-design review limit exceeded")
	ErrBoundary = errors.New("first-design review boundary failed")
)

// Request contains complete native request texts. No path, command, callback or
// execution mode is supplied by the caller.
type Request struct {
	FormatVersion   string `json:"formatVersion"`
	IntakeRequest   string `json:"intakeRequest"`
	ProposalRequest string `json:"proposalRequest"`
}

type report struct {
	FormatVersion         string              `json:"formatVersion"`
	Mode                  string              `json:"mode"`
	Authority             string              `json:"authority"`
	ExecutionStatus       string              `json:"executionStatus"`
	ReviewStatus          string              `json:"reviewStatus"`
	IntakeRequestSHA256   string              `json:"intakeRequestSHA256"`
	ProposalRequestSHA256 string              `json:"proposalRequestSHA256"`
	ModelComparison       string              `json:"modelComparison"`
	SuiteComparison       string              `json:"suiteComparison"`
	IntakeReport          *intakemodel.Report `json:"intakeReport"`
	ProposalReportSHA256  string              `json:"proposalReportSHA256"`
	ProposalReportJSON    string              `json:"proposalReportJSON"`
}

func digest(raw []byte) string {
	s := sha256.Sum256(raw)
	return hex.EncodeToString(s[:])
}

func compare(observed *string, proposal string) string {
	if observed == nil {
		return "missing"
	}
	if *observed == proposal {
		return "matches"
	}
	return "mismatch"
}

// Check returns one complete report, or nil bytes on failure. It preserves the
// intake report as a complete typed object, including its exact nested native
// intake JSON, and preserves the proposal report's native encoded bytes as a
// string. Input slices must not be changed concurrently. Neither native check
// calls a candidate reducer; this function performs no filesystem or process I/O.
func Check(raw []byte) ([]byte, error) {
	if len(raw) > MaxRequestBytes {
		return nil, ErrLimit
	}
	var req Request
	if strictjson.Decode(raw, &req, MaxRequestBytes, 4) != nil || req.FormatVersion != RequestVersion {
		return nil, ErrInvalid
	}
	intakeRaw, proposalRaw := []byte(req.IntakeRequest), []byte(req.ProposalRequest)
	intakeReport, err := intakemodel.Inspect(intakeRaw)
	if err != nil {
		switch {
		case errors.Is(err, intakemodel.ErrInvalid):
			return nil, ErrInvalid
		case errors.Is(err, intakemodel.ErrLimit):
			return nil, ErrLimit
		default:
			return nil, ErrBoundary
		}
	}
	proposalReport, err := proposalreplay.Check(proposalRaw)
	if err != nil {
		if errors.Is(err, proposalreplay.ErrInvalid) {
			return nil, ErrInvalid
		}
		return nil, ErrBoundary
	}
	proposalJSON, err := proposalreplay.Encode(proposalReport)
	if err != nil {
		return nil, ErrBoundary
	}

	// These extractions happen only after the corresponding native strict
	// decoder accepted the complete request. They do not validate or normalize
	// artifact text and cannot turn stale declarations into current ones.
	var intakeFields struct {
		Observed struct {
			Model         *string `json:"model"`
			ScenarioSuite *string `json:"scenarioSuite"`
		} `json:"observed"`
	}
	var proposalFields proposalreplay.Request
	if json.Unmarshal(intakeRaw, &intakeFields) != nil || json.Unmarshal(proposalRaw, &proposalFields) != nil {
		return nil, ErrBoundary
	}
	r := report{
		FormatVersion: ReportVersion, Mode: "example-only", Authority: "none",
		ExecutionStatus: "not-run", ReviewStatus: "not-established",
		IntakeRequestSHA256: digest(intakeRaw), ProposalRequestSHA256: digest(proposalRaw),
		ModelComparison: compare(intakeFields.Observed.Model, proposalFields.Model),
		SuiteComparison: compare(intakeFields.Observed.ScenarioSuite, proposalFields.Suite),
		IntakeReport:    intakeReport, ProposalReportSHA256: digest(proposalJSON),
		ProposalReportJSON: string(proposalJSON),
	}
	return encode(r)
}

func encode(r report) ([]byte, error) {
	var out bytes.Buffer
	e := json.NewEncoder(&out)
	e.SetEscapeHTML(false)
	if e.Encode(r) != nil {
		return nil, ErrBoundary
	}
	if out.Len() > MaxReportBytes {
		return nil, ErrLimit
	}
	return out.Bytes(), nil
}
