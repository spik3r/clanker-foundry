// Package changerehearsal checks a supplied model change and explicitly rehearses
// separately stated baseline and candidate expectations through native APIs.
package changerehearsal

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/acceptanceauthor"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/modelpatch"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/proposalreplay"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/scenarios"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/strictjson"
)

const (
	RequestVersion  = "lifecycle-change-rehearsal-request/v1"
	ReportVersion   = "lifecycle-change-rehearsal-report/v1"
	MaxRequestBytes = 32 << 20
	MaxReportBytes  = 64 << 20
)

var (
	ErrInvalid      = errors.New("invalid change rehearsal request")
	ErrLimit        = errors.New("change rehearsal limit exceeded")
	ErrPrecondition = errors.New("change rehearsal identity mismatch")
	ErrBoundary     = errors.New("change rehearsal native boundary failure")
)

type Request struct {
	FormatVersion           string `json:"formatVersion"`
	PatchRequest            string `json:"patchRequest"`
	BaselineProposalRequest string `json:"baselineProposalRequest"`
	CandidateSuite          string `json:"candidateSuite"`
	CandidateManifest       string `json:"candidateManifest"`
	ExpectedCandidateDigest string `json:"expectedCandidateDigest"`
}
type Side struct {
	Operation             string `json:"operation"`
	ProposalRequestSHA256 string `json:"proposalRequestSHA256"`
	ReceiptSHA256         string `json:"receiptSHA256"`
	ReceiptJSON           string `json:"receiptJSON"`
}
type Report struct {
	FormatVersion           string `json:"formatVersion"`
	Mode                    string `json:"mode"`
	Authority               string `json:"authority"`
	ReviewStatus            string `json:"reviewStatus"`
	Action                  string `json:"action"`
	Outcome                 string `json:"outcome"`
	RequestSHA256           string `json:"requestSHA256"`
	PatchRequestSHA256      string `json:"patchRequestSHA256"`
	ExpectedCandidateDigest string `json:"expectedCandidateDigest"`
	CandidateModelSHA256    string `json:"candidateModelSHA256"`
	PatchReportSHA256       string `json:"patchReportSHA256"`
	PatchReportJSON         string `json:"patchReportJSON"`
	Baseline                Side   `json:"baseline"`
	Candidate               Side   `json:"candidate"`
}
type prepared struct {
	report              Report
	baseline, candidate []byte
}

func hash(raw []byte) string { sum := sha256.Sum256(raw); return hex.EncodeToString(sum[:]) }
func validDigest(text string) bool {
	if len(text) != 64 {
		return false
	}
	for _, c := range text {
		if !(c >= 'a' && c <= 'f' || c >= '0' && c <= '9') {
			return false
		}
	}
	return true
}
func encode(value any, limit int) ([]byte, error) {
	var out bytes.Buffer
	e := json.NewEncoder(&out)
	e.SetEscapeHTML(false)
	if e.Encode(value) != nil {
		return nil, ErrBoundary
	}
	if out.Len() > limit {
		return nil, ErrLimit
	}
	return out.Bytes(), nil
}
func nativeError(err error) error {
	switch {
	case errors.Is(err, modelpatch.ErrLimit):
		return errors.Join(ErrLimit, err)
	case errors.Is(err, modelpatch.ErrPrecondition):
		return errors.Join(ErrPrecondition, err)
	case errors.Is(err, modelpatch.ErrInvalid), errors.Is(err, modelpatch.ErrConflict), errors.Is(err, proposalreplay.ErrInvalid):
		return errors.Join(ErrInvalid, err)
	default:
		return ErrBoundary
	}
}
func side(operation string, request []byte, receipt proposalreplay.Receipt) (Side, error) {
	encoded, err := proposalreplay.Encode(receipt)
	if err != nil {
		return Side{}, ErrBoundary
	}
	return Side{operation, hash(request), hash(encoded), string(encoded)}, nil
}

// Preparation receives no executor. Native checks establish consistency only;
// they cannot authorize a later rehearsal or repair supplied expectations.
func prepare(raw []byte) (prepared, error) {
	if len(raw) > MaxRequestBytes {
		return prepared{}, ErrLimit
	}
	var q Request
	if strictjson.Decode(raw, &q, MaxRequestBytes, 4) != nil || q.FormatVersion != RequestVersion || !validDigest(q.ExpectedCandidateDigest) {
		return prepared{}, ErrInvalid
	}
	if len(q.PatchRequest) > modelpatch.MaxRequestBytes || len(q.BaselineProposalRequest) > proposalreplay.MaxRequestBytes || len(q.CandidateSuite) > scenarios.MaxSourceBytes || len(q.CandidateManifest) > acceptanceauthor.MaxManifestBytes {
		return prepared{}, ErrLimit
	}
	patch, err := modelpatch.Execute([]byte(q.PatchRequest))
	if err != nil {
		return prepared{}, nativeError(err)
	}
	// Decode only the trusted native result after successful patch validation.
	// RawMessage retains the exact produced candidate token rather than recoding it.
	var plan struct {
		BeforeDigest, AfterDigest string
		After                     json.RawMessage
	}
	if json.Unmarshal(patch, &plan) != nil || len(plan.After) == 0 || !validDigest(plan.BeforeDigest) || !validDigest(plan.AfterDigest) {
		return prepared{}, ErrBoundary
	}
	if q.ExpectedCandidateDigest != plan.AfterDigest {
		return prepared{}, ErrPrecondition
	}
	baseline := []byte(q.BaselineProposalRequest)
	before, err := proposalreplay.Check(baseline)
	if err != nil {
		return prepared{}, nativeError(err)
	}
	if before.Preflight.CurrentModelDigest != plan.BeforeDigest {
		return prepared{}, ErrPrecondition
	}
	candidate, err := encode(proposalreplay.Request{FormatVersion: proposalreplay.RequestVersion, Model: string(plan.After), Suite: q.CandidateSuite, Manifest: q.CandidateManifest}, proposalreplay.MaxRequestBytes)
	if err != nil {
		return prepared{}, err
	}
	after, err := proposalreplay.Check(candidate)
	if err != nil {
		return prepared{}, nativeError(err)
	}
	if after.Preflight.CurrentModelDigest != plan.AfterDigest {
		return prepared{}, ErrBoundary
	}
	outcome := "checked"
	for _, receipt := range []proposalreplay.Receipt{before, after} {
		switch receipt.Outcome {
		case "checked":
		case "stale":
			outcome = "stale"
		default:
			return prepared{}, ErrBoundary
		}
	}
	b, err := side("check", baseline, before)
	if err != nil {
		return prepared{}, err
	}
	c, err := side("check", candidate, after)
	if err != nil {
		return prepared{}, err
	}
	report := Report{ReportVersion, "example-only", "none", "not-established", "check", outcome, hash(raw), hash([]byte(q.PatchRequest)), q.ExpectedCandidateDigest, hash(plan.After), hash(patch), string(patch), b, c}
	return prepared{report, baseline, candidate}, nil
}

// Check performs no filesystem, subprocess, network or scenario execution.
// Caller inputs must remain immutable during this and Execute's calls.
func Check(raw []byte) ([]byte, error) {
	p, err := prepare(raw)
	if err != nil {
		return nil, err
	}
	return encode(p.report, MaxReportBytes)
}

// ExitCode describes delivery/rehearsal outcomes, never human approval.
func ExitCode(raw []byte, err error) int {
	if err != nil {
		if len(raw) != 0 {
			return 4
		}
		if errors.Is(err, ErrInvalid) || errors.Is(err, ErrLimit) || errors.Is(err, ErrPrecondition) {
			return 2
		}
		return 4
	}
	var r Report
	if json.Unmarshal(raw, &r) != nil || r.FormatVersion != ReportVersion {
		return 4
	}
	switch r.Outcome {
	case "checked", "matched":
		return 0
	case "stale", "baseline-mismatched", "candidate-mismatched":
		return 1
	default:
		return 4
	}
}
