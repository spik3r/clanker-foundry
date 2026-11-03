// Package revisioncheck composes native check-only design review reports.
package revisioncheck

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/firstdesignreview"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/intakeimpact"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/strictjson"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/traceability"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/traceimpact"
)

const (
	RequestVersion  = "lifecycle-revision-check-request/v1"
	ReportVersion   = "lifecycle-revision-check-report/v1"
	MaxRequestBytes = 32 << 20
	MaxReportBytes  = 64 << 20
)

var (
	ErrInvalid  = errors.New("invalid revision check request")
	ErrLimit    = errors.New("revision check limit exceeded")
	ErrBoundary = errors.New("revision check native boundary failed")
)

// Snapshot contains original UTF-8 texts, never paths or execution selectors.
type Snapshot struct {
	FirstDesignRequest string `json:"firstDesignRequest"`
	Requirements       string `json:"requirements"`
	Catalog            string `json:"catalog"`
}
type Request struct {
	FormatVersion string   `json:"formatVersion"`
	Before        Snapshot `json:"before"`
	After         Snapshot `json:"after"`
}
type SideReport struct {
	FirstDesignRequestSHA256 string `json:"firstDesignRequestSHA256"`
	RequirementsSHA256       string `json:"requirementsSHA256"`
	CatalogSHA256            string `json:"catalogSHA256"`
	RequirementsComparison   string `json:"requirementsComparison"`
	FirstDesignReportSHA256  string `json:"firstDesignReportSHA256"`
	FirstDesignReportJSON    string `json:"firstDesignReportJSON"`
}
type Report struct {
	FormatVersion            string     `json:"formatVersion"`
	Mode                     string     `json:"mode"`
	Authority                string     `json:"authority"`
	ExecutionStatus          string     `json:"executionStatus"`
	ReviewStatus             string     `json:"reviewStatus"`
	Before                   SideReport `json:"before"`
	After                    SideReport `json:"after"`
	IntakeImpactReportSHA256 string     `json:"intakeImpactReportSHA256"`
	IntakeImpactReportJSON   string     `json:"intakeImpactReportJSON"`
	TraceImpactReportSHA256  string     `json:"traceImpactReportSHA256"`
	TraceImpactReportJSON    string     `json:"traceImpactReportJSON"`
}

func hash(b []byte) string { d := sha256.Sum256(b); return hex.EncodeToString(d[:]) }
func nativeError(err error) error {
	switch {
	case errors.Is(err, firstdesignreview.ErrLimit), errors.Is(err, intakeimpact.ErrReportLimit), errors.Is(err, traceimpact.ErrReportLimit):
		return ErrLimit
	case errors.Is(err, firstdesignreview.ErrInvalid), errors.Is(err, intakeimpact.ErrInvalid), errors.Is(err, traceimpact.ErrInvalid):
		return ErrInvalid
	default:
		return ErrBoundary
	}
}
func encode(v any, limit int) ([]byte, error) {
	var out bytes.Buffer
	e := json.NewEncoder(&out)
	e.SetEscapeHTML(false)
	if e.Encode(v) != nil {
		return nil, ErrBoundary
	}
	if out.Len() > limit {
		return nil, ErrLimit
	}
	return out.Bytes(), nil
}

type parts struct{ intake, model string }

func inspect(s Snapshot) (SideReport, parts, error) {
	if len(s.FirstDesignRequest) > firstdesignreview.MaxRequestBytes || len(s.Requirements) > traceability.MaxBytes || len(s.Catalog) > 4*traceability.MaxBytes {
		return SideReport{}, parts{}, ErrLimit
	}
	native, err := firstdesignreview.Check([]byte(s.FirstDesignRequest))
	if err != nil {
		return SideReport{}, parts{}, nativeError(err)
	}
	// Native strict checks have accepted these exact request strings. Extract
	// only for composition; do not normalize or rewrite the original artifacts.
	var f firstdesignreview.Request
	if json.Unmarshal([]byte(s.FirstDesignRequest), &f) != nil {
		return SideReport{}, parts{}, ErrBoundary
	}
	var intakeFields struct {
		Observed struct {
			Requirements *string `json:"requirements"`
		} `json:"observed"`
	}
	var proposalFields struct {
		Model string `json:"model"`
	}
	if json.Unmarshal([]byte(f.IntakeRequest), &intakeFields) != nil || json.Unmarshal([]byte(f.ProposalRequest), &proposalFields) != nil {
		return SideReport{}, parts{}, ErrBoundary
	}
	comparison := "missing"
	if intakeFields.Observed.Requirements != nil {
		comparison = "mismatch"
		if *intakeFields.Observed.Requirements == s.Requirements {
			comparison = "matches"
		}
	}
	return SideReport{hash([]byte(s.FirstDesignRequest)), hash([]byte(s.Requirements)), hash([]byte(s.Catalog)), comparison, hash(native), string(native)}, parts{f.IntakeRequest, proposalFields.Model}, nil
}

// Check preserves four complete native encoded reports and their byte hashes.
// Mismatches, questions and stale preflight statuses stay native observations,
// never an aggregate approval. Inputs must remain immutable during the call.
// This composition performs no I/O and invokes no reducer or scenario callback.
func Check(raw []byte) ([]byte, error) {
	if len(raw) > MaxRequestBytes {
		return nil, ErrLimit
	}
	var req Request
	if strictjson.Decode(raw, &req, MaxRequestBytes, 4) != nil || req.FormatVersion != RequestVersion {
		return nil, ErrInvalid
	}
	before, b, err := inspect(req.Before)
	if err != nil {
		return nil, err
	}
	after, a, err := inspect(req.After)
	if err != nil {
		return nil, err
	}
	intakeRequest, err := encode(struct {
		FormatVersion string `json:"formatVersion"`
		Before        string `json:"before"`
		After         string `json:"after"`
	}{intakeimpact.RequestVersion, b.intake, a.intake}, intakeimpact.MaxRequestBytes)
	if err != nil {
		return nil, err
	}
	intakeReport, err := intakeimpact.Execute(intakeRequest)
	if err != nil {
		return nil, nativeError(err)
	}
	type traceSide struct {
		Model        string `json:"model"`
		Requirements string `json:"requirements"`
		Catalog      string `json:"catalog"`
	}
	traceRequest, err := encode(struct {
		FormatVersion string    `json:"formatVersion"`
		Before        traceSide `json:"before"`
		After         traceSide `json:"after"`
	}{traceimpact.RequestVersion, traceSide{b.model, req.Before.Requirements, req.Before.Catalog}, traceSide{a.model, req.After.Requirements, req.After.Catalog}}, traceimpact.MaxRequestBytes)
	if err != nil {
		return nil, err
	}
	traceReport, err := traceimpact.Execute(traceRequest)
	if err != nil {
		return nil, nativeError(err)
	}
	return encode(Report{ReportVersion, "example-only", "none", "not-run", "not-established", before, after, hash(intakeReport), string(intakeReport), hash(traceReport), string(traceReport)}, MaxReportBytes)
}
