// Package custommergerehearsal checks caller-supplied concurrent designs and
// explicitly rehearses their independently supplied expectations through native APIs.
package custommergerehearsal

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/acceptanceauthor"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/modelmerge"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/proposalreplay"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/scenarios"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/strictjson"
)

const (
	RequestVersion  = "lifecycle-custom-merge-rehearsal-request/v1"
	ReportVersion   = "lifecycle-custom-merge-rehearsal-report/v1"
	MaxRequestBytes = 32 << 20
	MaxReportBytes  = 128 << 20
)

var (
	ErrInvalid      = errors.New("invalid custom merge rehearsal request")
	ErrLimit        = errors.New("custom merge rehearsal limit exceeded")
	ErrPrecondition = errors.New("custom merge rehearsal identity mismatch")
	ErrBoundary     = errors.New("custom merge rehearsal native boundary failure")
)

type Expectations struct {
	Suite    string `json:"suite"`
	Manifest string `json:"manifest"`
}
type Request struct {
	FormatVersion           string       `json:"formatVersion"`
	MergeRequest            string       `json:"mergeRequest"`
	Base                    Expectations `json:"base"`
	Left                    Expectations `json:"left"`
	Right                   Expectations `json:"right"`
	Candidate               Expectations `json:"candidate"`
	ExpectedCandidateDigest string       `json:"expectedCandidateDigest"`
}
type Stage struct {
	ModelSHA256           *string `json:"modelSHA256"`
	SuiteSHA256           string  `json:"suiteSHA256"`
	ManifestSHA256        string  `json:"manifestSHA256"`
	ProposalRequestSHA256 *string `json:"proposalRequestSHA256"`
	Operation             string  `json:"operation"`
	ReceiptSHA256         *string `json:"receiptSHA256"`
	ReceiptJSON           *string `json:"receiptJSON"`
}
type Report struct {
	FormatVersion           string `json:"formatVersion"`
	Mode                    string `json:"mode"`
	Authority               string `json:"authority"`
	ReviewStatus            string `json:"reviewStatus"`
	Action                  string `json:"action"`
	Outcome                 string `json:"outcome"`
	RequestSHA256           string `json:"requestSHA256"`
	MergeRequestSHA256      string `json:"mergeRequestSHA256"`
	MergeRequestJSON        string `json:"mergeRequestJSON"`
	MergeReportSHA256       string `json:"mergeReportSHA256"`
	MergeReportJSON         string `json:"mergeReportJSON"`
	ExpectedCandidateDigest string `json:"expectedCandidateDigest"`
	Base                    Stage  `json:"base"`
	Left                    Stage  `json:"left"`
	Right                   Stage  `json:"right"`
	Candidate               Stage  `json:"candidate"`
}

func (r *Report) stages() [4]*Stage { return [4]*Stage{&r.Base, &r.Left, &r.Right, &r.Candidate} }

type prepared struct {
	report    Report
	proposals [4][]byte
}

func hash(raw []byte) string { s := sha256.Sum256(raw); return hex.EncodeToString(s[:]) }
func ptr(s string) *string   { return &s }
func validDigest(s string) bool {
	if len(s) != 64 {
		return false
	}
	for _, c := range s {
		if !(c >= '0' && c <= '9' || c >= 'a' && c <= 'f') {
			return false
		}
	}
	return true
}
func encode(v any, limit int) ([]byte, error) {
	var b bytes.Buffer
	e := json.NewEncoder(&b)
	e.SetEscapeHTML(false)
	if e.Encode(v) != nil {
		return nil, ErrBoundary
	}
	if b.Len() > limit {
		return nil, ErrLimit
	}
	return b.Bytes(), nil
}
func nativeError(err error) error {
	switch {
	case errors.Is(err, modelmerge.ErrLimit):
		return errors.Join(ErrLimit, err)
	case errors.Is(err, modelmerge.ErrInvalid), errors.Is(err, proposalreplay.ErrInvalid):
		return errors.Join(ErrInvalid, err)
	default:
		return errors.Join(ErrBoundary, err)
	}
}
func retain(s *Stage, operation string, request []byte, r proposalreplay.Receipt) error {
	raw, err := proposalreplay.Encode(r)
	if err != nil {
		return errors.Join(ErrBoundary, err)
	}
	s.Operation = operation
	s.ProposalRequestSHA256 = ptr(hash(request))
	s.ReceiptSHA256 = ptr(hash(raw))
	s.ReceiptJSON = ptr(string(raw))
	return nil
}

// Preparation has no executor capability. Native checks establish consistency,
// not authorization, reviewed intent, or authenticated execution evidence.
func prepare(raw []byte) (prepared, error) {
	var p prepared
	if len(raw) > MaxRequestBytes {
		return p, ErrLimit
	}
	var q Request
	if strictjson.Decode(raw, &q, MaxRequestBytes, 5) != nil || q.FormatVersion != RequestVersion || !validDigest(q.ExpectedCandidateDigest) {
		return p, ErrInvalid
	}
	pairs := [4]Expectations{q.Base, q.Left, q.Right, q.Candidate}
	if len(q.MergeRequest) > modelmerge.MaxRequestBytes {
		return p, ErrLimit
	}
	for _, pair := range pairs {
		if pair.Suite == "" || pair.Manifest == "" {
			return p, ErrInvalid
		}
		if len(pair.Suite) > scenarios.MaxSourceBytes || len(pair.Manifest) > acceptanceauthor.MaxManifestBytes {
			return p, ErrLimit
		}
	}
	native, err := modelmerge.Execute([]byte(q.MergeRequest))
	if err != nil {
		return p, nativeError(err)
	}
	// Extract tokens only after the native merger has validated the request.
	var source struct{ Base, Left, Right json.RawMessage }
	var result struct {
		BaseDigest, LeftDigest, RightDigest, Status string
		CandidateDigest                             *string
		Candidate                                   json.RawMessage
	}
	if json.Unmarshal([]byte(q.MergeRequest), &source) != nil || json.Unmarshal(native, &result) != nil {
		return p, ErrBoundary
	}
	tokens := [4]json.RawMessage{source.Base, source.Left, source.Right, nil}
	digests := [4]string{result.BaseDigest, result.LeftDigest, result.RightDigest, ""}
	for _, d := range digests[:3] {
		if !validDigest(d) {
			return p, ErrBoundary
		}
	}
	switch result.Status {
	case modelmerge.Merged:
		if result.CandidateDigest == nil || !validDigest(*result.CandidateDigest) || len(result.Candidate) == 0 || bytes.Equal(result.Candidate, []byte("null")) {
			return p, ErrBoundary
		}
		if q.ExpectedCandidateDigest != *result.CandidateDigest {
			return p, ErrPrecondition
		}
		tokens[3] = result.Candidate
		digests[3] = *result.CandidateDigest
	case modelmerge.Conflicted, modelmerge.InvalidCombination:
		if result.CandidateDigest != nil || !bytes.Equal(result.Candidate, []byte("null")) {
			return p, ErrBoundary
		}
	default:
		return p, ErrBoundary
	}
	p.report = Report{FormatVersion: ReportVersion, Mode: "example-only", Authority: "none", ReviewStatus: "not-established", Action: "check", Outcome: result.Status, RequestSHA256: hash(raw), MergeRequestSHA256: hash([]byte(q.MergeRequest)), MergeRequestJSON: q.MergeRequest, MergeReportSHA256: hash(native), MergeReportJSON: string(native), ExpectedCandidateDigest: q.ExpectedCandidateDigest}
	stages := p.report.stages()
	for i, s := range stages {
		s.Operation = "not-attempted"
		s.SuiteSHA256 = hash([]byte(pairs[i].Suite))
		s.ManifestSHA256 = hash([]byte(pairs[i].Manifest))
		if len(tokens[i]) > 1<<20 {
			return prepared{}, ErrLimit
		}
		if tokens[i] != nil {
			s.ModelSHA256 = ptr(hash(tokens[i]))
		}
	}
	if result.Status != modelmerge.Merged {
		return p, nil
	}
	p.report.Outcome = "checked"
	for i, s := range stages {
		p.proposals[i], err = encode(proposalreplay.Request{FormatVersion: proposalreplay.RequestVersion, Model: string(tokens[i]), Suite: pairs[i].Suite, Manifest: pairs[i].Manifest}, proposalreplay.MaxRequestBytes)
		if err != nil {
			return prepared{}, err
		}
		receipt, e := proposalreplay.Check(p.proposals[i])
		if e != nil {
			return prepared{}, nativeError(e)
		}
		if receipt.Preflight.CurrentModelDigest != digests[i] {
			return prepared{}, ErrPrecondition
		}
		switch receipt.Outcome {
		case "checked":
		case "stale":
			p.report.Outcome = "stale"
		default:
			return prepared{}, ErrBoundary
		}
		if err = retain(s, "check", p.proposals[i], receipt); err != nil {
			return prepared{}, err
		}
	}
	// Reserve the complete maximum native receipt envelope before side effects.
	// Valid native JSON needs at most twice its byte length when embedded as text.
	skeleton := p.report
	for _, s := range skeleton.stages() {
		s.ReceiptSHA256 = ptr(string(bytes.Repeat([]byte("0"), 64)))
		s.ReceiptJSON = nil
	}
	envelope, e := encode(skeleton, MaxReportBytes)
	if e != nil {
		return prepared{}, e
	}
	if len(envelope)+4*(2*proposalreplay.MaxReceiptBytes+2) > MaxReportBytes {
		return prepared{}, ErrLimit
	}
	return p, nil
}

// Check performs no filesystem, subprocess, network or scenario execution.
// Caller inputs must remain immutable during Check and Execute.
func Check(raw []byte) ([]byte, error) {
	p, e := prepare(raw)
	if e != nil {
		return nil, e
	}
	return encode(p.report, MaxReportBytes)
}

// ExitCode describes observed rehearsal outcomes, never human approval.
func ExitCode(raw []byte, err error) int {
	if err != nil {
		if errors.Is(err, ErrBoundary) {
			return 4
		}
		if len(raw) == 0 && (errors.Is(err, ErrInvalid) || errors.Is(err, ErrLimit) || errors.Is(err, ErrPrecondition)) {
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
	case "conflicted", "invalid-combination", "stale", "base-mismatched", "left-mismatched", "right-mismatched", "candidate-mismatched":
		return 1
	default:
		return 4
	}
}
