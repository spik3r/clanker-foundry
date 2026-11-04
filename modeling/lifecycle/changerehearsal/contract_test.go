package changerehearsal

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/canonical"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/jsonformat"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/modelpatch"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/proposalreplay"
)

// Independent static observations precede production. This file compiles with
// only existing native APIs; future entrypoints are supplied through callbacks.
type frozenRequest struct {
	FormatVersion           string `json:"formatVersion"`
	PatchRequest            string `json:"patchRequest"`
	BaselineProposalRequest string `json:"baselineProposalRequest"`
	CandidateSuite          string `json:"candidateSuite"`
	CandidateManifest       string `json:"candidateManifest"`
	ExpectedCandidateDigest string `json:"expectedCandidateDigest"`
}
type frozenVector struct {
	ID, Request, CheckOutcome, ExecuteOutcome string
	Valid                                     bool
}
type frozenData struct {
	BeforeDigest, CandidateDigest, DeclaredCandidate string
	Vectors                                          []frozenVector
}
type frozenSide struct{ Operation, ProposalRequestSHA256, ReceiptSHA256, ReceiptJSON string }
type frozenReport struct {
	FormatVersion, Mode, Authority, ReviewStatus, Action, Outcome                                                        string
	RequestSHA256, PatchRequestSHA256, ExpectedCandidateDigest, CandidateModelSHA256, PatchReportSHA256, PatchReportJSON string
	Baseline, Candidate                                                                                                  frozenSide
}

func frozenHash(b []byte) string { d := sha256.Sum256(b); return hex.EncodeToString(d[:]) }
func frozenEncode(t *testing.T, v any) []byte {
	t.Helper()
	var b bytes.Buffer
	e := json.NewEncoder(&b)
	e.SetEscapeHTML(false)
	if e.Encode(v) != nil {
		t.Fatal("encode fixture")
	}
	return b.Bytes()
}
func frozenObservations(t *testing.T) frozenData {
	t.Helper()
	b, e := os.ReadFile("testdata/observations.json")
	if e != nil {
		t.Fatal(e)
	}
	if frozenHash(b) != "d9b30467427539c02d3355724b0d9ac13c47f46360996b4dfcc660d1a96403e3" {
		t.Fatal("independent observations changed")
	}
	var d frozenData
	if json.Unmarshal(b, &d) != nil || len(d.Vectors) != 6 {
		t.Fatal("observations")
	}
	return d
}
func frozenPrepare(t *testing.T, raw string) (frozenRequest, []byte, []byte) {
	t.Helper()
	var q frozenRequest
	if json.Unmarshal([]byte(raw), &q) != nil {
		t.Fatal("fixture request")
	}
	patch, e := modelpatch.Execute([]byte(q.PatchRequest))
	if e != nil {
		t.Fatal("native patch fixture", e)
	}
	var p struct {
		BeforeDigest, AfterDigest string
		After                     json.RawMessage
	}
	if json.Unmarshal(patch, &p) != nil {
		t.Fatal("native patch report")
	}
	candidate := frozenEncode(t, proposalreplay.Request{FormatVersion: proposalreplay.RequestVersion, Model: string(p.After), Suite: q.CandidateSuite, Manifest: q.CandidateManifest})
	return q, patch, candidate
}
func TestFrozenNativeChangeRehearsalInputs(t *testing.T) {
	d := frozenObservations(t)
	m, e := jsonformat.Decode([]byte(d.DeclaredCandidate))
	if e != nil {
		t.Fatal(e)
	}
	md, e := canonical.Digest(m)
	if e != nil || md != d.CandidateDigest {
		t.Fatal("declared candidate digest", md, e)
	}
	for _, v := range d.Vectors {
		t.Run(v.ID, func(t *testing.T) {
			q, patch, candidate := frozenPrepare(t, v.Request)
			var p struct {
				BeforeDigest, AfterDigest string
				Review                    struct {
					Changes []struct {
						Kind, ID string
						Fields   []string
					}
				}
			}
			if json.Unmarshal(patch, &p) != nil || p.BeforeDigest != d.BeforeDigest || p.AfterDigest != d.CandidateDigest {
				t.Fatal("fixed patch identity")
			}
			if len(p.Review.Changes) != 1 || p.Review.Changes[0].Kind != "guard" || p.Review.Changes[0].ID != "G_actor" || !reflect.DeepEqual(p.Review.Changes[0].Fields, []string{"expr"}) {
				t.Fatal("fixed behavioral patch review", p.Review)
			}
			if !v.Valid {
				return
			}
			br, e := proposalreplay.Check([]byte(q.BaselineProposalRequest))
			if e != nil || br.Outcome != "checked" {
				t.Fatal("baseline native preflight", e)
			}
			cr, e := proposalreplay.Check(candidate)
			if e != nil || cr.Outcome != v.CheckOutcome {
				t.Fatal("candidate native preflight", cr.Outcome, e)
			}
		})
	}
}
func frozenAssertReport(t *testing.T, v frozenVector, raw []byte, action string) frozenReport {
	t.Helper()
	q, patch, candidate := frozenPrepare(t, v.Request)
	var r frozenReport
	if json.Unmarshal(raw, &r) != nil {
		t.Fatal("bridge report")
	}
	want := v.CheckOutcome
	if action == "execute" {
		want = v.ExecuteOutcome
	}
	if r.FormatVersion != "lifecycle-change-rehearsal-report/v1" || r.Mode != "example-only" || r.Authority != "none" || r.ReviewStatus != "not-established" || r.Action != action || r.Outcome != want {
		t.Fatal("report status/authority", r.Action, r.Outcome)
	}
	var p struct{ After json.RawMessage }
	_ = json.Unmarshal(patch, &p)
	if r.RequestSHA256 != frozenHash([]byte(v.Request)) || r.PatchRequestSHA256 != frozenHash([]byte(q.PatchRequest)) || r.ExpectedCandidateDigest != q.ExpectedCandidateDigest || r.CandidateModelSHA256 != frozenHash(p.After) || r.PatchReportSHA256 != frozenHash(patch) || r.PatchReportJSON != string(patch) {
		t.Fatal("source/patch byte identity changed")
	}
	for _, s := range []struct {
		slot  frozenSide
		input []byte
	}{{r.Baseline, []byte(q.BaselineProposalRequest)}, {r.Candidate, candidate}} {
		if s.slot.ProposalRequestSHA256 != frozenHash(s.input) || s.slot.ReceiptSHA256 != frozenHash([]byte(s.slot.ReceiptJSON)) {
			t.Fatal("native receipt/request byte hash")
		}
		if s.slot.Operation == "check" {
			receipt, e := proposalreplay.Check(s.input)
			if e != nil {
				t.Fatal(e)
			}
			encoded, e := proposalreplay.Encode(receipt)
			if e != nil || s.slot.ReceiptJSON != string(encoded) {
				t.Fatal("full native check receipt bytes changed")
			}
		}
	}
	return r
}
func assertFrozenCheck(t *testing.T, check func([]byte) ([]byte, error)) {
	t.Helper()
	for _, v := range frozenObservations(t).Vectors {
		t.Run(v.ID, func(t *testing.T) {
			input := []byte(v.Request)
			saved := bytes.Clone(input)
			raw, e := check(input)
			if !bytes.Equal(input, saved) {
				t.Fatal("input mutated")
			}
			if !v.Valid {
				if e == nil || raw != nil {
					t.Fatal("identity refusal returned report")
				}
				return
			}
			if e != nil {
				t.Fatal(e)
			}
			r := frozenAssertReport(t, v, raw, "check")
			if r.Baseline.Operation != "check" || r.Candidate.Operation != "check" {
				t.Fatal("check fabricated execution")
			}
			again, e := check(input)
			if e != nil || !bytes.Equal(raw, again) {
				t.Fatal("check changed")
			}
		})
	}
	valid := []byte(frozenObservations(t).Vectors[0].Request)
	for _, bad := range [][]byte{nil, []byte("null"), []byte(`{}`), append(bytes.Clone(valid), valid...), bytes.Replace(valid, []byte(`"patchRequest":`), []byte(`"extra":null,"patchRequest":`), 1)} {
		if b, e := check(bad); e == nil || b != nil {
			t.Fatal("strict request accepted")
		}
	}
	if b, e := check(bytes.Repeat([]byte(" "), (32<<20)+1)); e == nil || b != nil {
		t.Fatal("request limit")
	}
}
func assertFrozenExecute(t *testing.T, execute func(context.Context, []byte, string) ([]byte, error), sourceRoot string) {
	t.Helper()
	for _, v := range frozenObservations(t).Vectors {
		t.Run(v.ID, func(t *testing.T) {
			raw, e := execute(context.Background(), []byte(v.Request), sourceRoot)
			if !v.Valid {
				if e == nil || raw != nil {
					t.Fatal("identity refused after execution")
				}
				return
			}
			if e != nil {
				t.Fatal(e)
			}
			r := frozenAssertReport(t, v, raw, "execute")
			if v.ExecuteOutcome == "stale" {
				if r.Baseline.Operation != "check" || r.Candidate.Operation != "check" {
					t.Fatal("stale preflight executed")
				}
				return
			}
			if r.Baseline.Operation != "execute" {
				t.Fatal("baseline not executed")
			}
			want := "execute"
			if v.ExecuteOutcome == "baseline-mismatched" {
				want = "check"
			}
			if r.Candidate.Operation != want {
				t.Fatal("candidate stage incorrect", r.Candidate.Operation)
			}
			for _, side := range []frozenSide{r.Baseline, r.Candidate} {
				if side.Operation != "execute" {
					continue
				}
				var receipt proposalreplay.Receipt
				if json.Unmarshal([]byte(side.ReceiptJSON), &receipt) != nil || receipt.Action != "execute" || !receipt.CandidateStarted || receipt.Cleanup != "removed" || receipt.Compilation != "passed" {
					t.Fatal("actual native execution receipt", receipt.Outcome, receipt.Compilation, receipt.Cleanup)
				}
				var report struct {
					Matched                       bool
					CasesCompleted, StepsExecuted int
				}
				if json.Unmarshal(receipt.CandidateReport, &report) != nil {
					t.Fatal("candidate report")
				}
				mismatch := side.ReceiptSHA256 == r.Baseline.ReceiptSHA256 && v.ExecuteOutcome == "baseline-mismatched" || side.ReceiptSHA256 == r.Candidate.ReceiptSHA256 && v.ExecuteOutcome == "candidate-mismatched"
				if report.Matched == mismatch {
					t.Fatal("literal mismatch suppressed")
				}
				if !mismatch && (report.CasesCompleted != 2 || report.StepsExecuted != 2) {
					t.Fatal("fixed scenario count")
				}
			}
		})
	}
}
