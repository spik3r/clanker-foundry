package revisioncheck

import (
	"bytes"
	"encoding/json"
	"errors"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/firstdesignreview"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/scenariodiff"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/strictjson"
)

const (
	RequestVersionV2 = "lifecycle-revision-check-request/v2"
	ReportVersionV2  = "lifecycle-revision-check-report/v2"
)

type scenarioDiffResult struct {
	Status       string  `json:"status"`
	Reason       *string `json:"reason"`
	ReportSHA256 *string `json:"reportSHA256"`
	ReportJSON   *string `json:"reportJSON"`
}
type reportV2 struct {
	Report
	ScenarioDiff scenarioDiffResult `json:"scenarioDiff"`
}

// A refused native comparison is an observation, not a guess about whether its
// cause was stale bindings or malformed input. Other failures reject the result.
func scenarioDiffOutcome(raw []byte, err error) (scenarioDiffResult, error) {
	reason := ""
	switch {
	case errors.Is(err, scenariodiff.ErrInvalid):
		reason = "native-input-refused"
	case errors.Is(err, scenariodiff.ErrLimit):
		reason = "native-limit-refused"
	case err != nil:
		return scenarioDiffResult{}, ErrBoundary
	default:
		if len(raw) == 0 {
			return scenarioDiffResult{}, ErrBoundary
		}
		text, digest := string(raw), hash(raw)
		return scenarioDiffResult{Status: "checked", ReportSHA256: &digest, ReportJSON: &text}, nil
	}
	return scenarioDiffResult{Status: "not-checked", Reason: &reason}, nil
}

func scenarioTexts(s Snapshot) (string, string, error) {
	var first firstdesignreview.Request
	var proposal struct{ Model, Suite string }
	if json.Unmarshal([]byte(s.FirstDesignRequest), &first) != nil || json.Unmarshal([]byte(first.ProposalRequest), &proposal) != nil {
		return "", "", ErrBoundary
	}
	return proposal.Model, proposal.Suite, nil
}

// The retained v1 check has already accepted full strict JSON framing of each
// original model and suite (even when a manifest is stale). Insert those exact
// texts rather than marshaling RawMessage, which would compact their tokens.
func scenarioRequest(req Request) ([]byte, error) {
	bm, bs, err := scenarioTexts(req.Before)
	if err != nil {
		return nil, err
	}
	am, as, err := scenarioTexts(req.After)
	if err != nil {
		return nil, err
	}
	var out bytes.Buffer
	out.WriteString(`{"formatVersion":"` + scenariodiff.FormatVersion + `","beforeModel":`)
	out.WriteString(bm)
	out.WriteString(`,"beforeSuite":`)
	out.WriteString(bs)
	out.WriteString(`,"afterModel":`)
	out.WriteString(am)
	out.WriteString(`,"afterSuite":`)
	out.WriteString(as)
	out.WriteString("}\n")
	if out.Len() > scenariodiff.MaxRequestBytes {
		return nil, ErrLimit
	}
	return out.Bytes(), nil
}

// CheckV2 retains all four v1 native reports and adds the existing native literal
// scenario comparison. It performs no I/O or execution and never repairs input.
// Keep input bytes immutable during this call, as with Check.
func CheckV2(raw []byte) ([]byte, error) {
	if len(raw) > MaxRequestBytes {
		return nil, ErrLimit
	}
	var req Request
	if strictjson.Decode(raw, &req, MaxRequestBytes, 4) != nil || req.FormatVersion != RequestVersionV2 {
		return nil, ErrInvalid
	}
	v1 := req
	v1.FormatVersion = RequestVersion
	v1Request, err := encode(v1, MaxRequestBytes)
	if err != nil {
		return nil, err
	}
	retained, err := Check(v1Request)
	if err != nil {
		return nil, err
	}
	var original Report
	if json.Unmarshal(retained, &original) != nil {
		return nil, ErrBoundary
	}
	comparisonRequest, err := scenarioRequest(req)
	if err != nil {
		return nil, err
	}
	native, nativeErr := scenariodiff.Execute(comparisonRequest)
	comparison, err := scenarioDiffOutcome(native, nativeErr)
	if err != nil {
		return nil, err
	}
	original.FormatVersion = ReportVersionV2
	return encode(reportV2{Report: original, ScenarioDiff: comparison}, MaxReportBytes)
}
