package revisioncheck

import (
	"bytes"
	"encoding/json"
	"errors"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/impact"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/strictjson"
)

const (
	RequestVersionV3 = "lifecycle-revision-check-request/v3"
	ReportVersionV3  = "lifecycle-revision-check-report/v3"
)

type modelDiffResult struct {
	ModelDiffReportSHA256 string `json:"modelDiffReportSHA256"`
	ModelDiffReportJSON   string `json:"modelDiffReportJSON"`
}
type reportV3 struct {
	reportV2
	modelDiffResult
}

func modelDiffOutcome(raw []byte, err error) (modelDiffResult, error) {
	switch {
	case errors.Is(err, impact.ErrInvalidInput):
		return modelDiffResult{}, ErrInvalid
	case errors.Is(err, impact.ErrReportLimit):
		return modelDiffResult{}, ErrLimit
	case err != nil || len(raw) == 0:
		return modelDiffResult{}, ErrBoundary
	default:
		return modelDiffResult{hash(raw), string(raw)}, nil
	}
}

// The retained v2 checks have already accepted each original model's strict
// framing. Insert its original text, preserving internal token whitespace.
// No scenario binding assertions are inferred from suites or trace catalogs.
func modelDiffRequest(req Request) ([]byte, error) {
	before, _, err := scenarioTexts(req.Before)
	if err != nil {
		return nil, err
	}
	after, _, err := scenarioTexts(req.After)
	if err != nil {
		return nil, err
	}
	var out bytes.Buffer
	out.WriteString(`{"formatVersion":"` + impact.FormatVersion + `","before":`)
	out.WriteString(before)
	out.WriteString(`,"after":`)
	out.WriteString(after)
	out.WriteString(",\"bindings\":[]}\n")
	if out.Len() > impact.MaxSourceBytes {
		return nil, ErrLimit
	}
	return out.Bytes(), nil
}

// CheckV3 retains v2's native observations and adds a complete native model diff.
// It performs no I/O or execution and never repairs or accepts a design. Keep
// input bytes immutable during this call, as with Check and CheckV2.
func CheckV3(raw []byte) ([]byte, error) {
	if len(raw) > MaxRequestBytes {
		return nil, ErrLimit
	}
	var req Request
	if strictjson.Decode(raw, &req, MaxRequestBytes, 4) != nil || req.FormatVersion != RequestVersionV3 {
		return nil, ErrInvalid
	}
	v2 := req
	v2.FormatVersion = RequestVersionV2
	v2Request, err := encode(v2, MaxRequestBytes)
	if err != nil {
		return nil, err
	}
	retained, err := CheckV2(v2Request)
	if err != nil {
		return nil, err
	}
	var original reportV2
	if json.Unmarshal(retained, &original) != nil {
		return nil, ErrBoundary
	}
	request, err := modelDiffRequest(req)
	if err != nil {
		return nil, err
	}
	native, nativeErr := impact.Execute(request)
	comparison, err := modelDiffOutcome(native, nativeErr)
	if err != nil {
		return nil, err
	}
	original.FormatVersion = ReportVersionV3
	return encode(reportV3{original, comparison}, MaxReportBytes)
}
