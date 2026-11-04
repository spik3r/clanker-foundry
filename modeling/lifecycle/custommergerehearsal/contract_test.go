package custommergerehearsal

// These contract expectations were frozen by a separate contributor before the
// custom bridge implementation. Synthetic receipts test orchestration only.
import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/canonical"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/jsonformat"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/modelmerge"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/proposalreplay"
)

var contractRoles = []string{"base", "left", "right", "candidate"}

type contractVersion struct{ Model, ModelDigest, Suite, Manifest, MismatchSuite, MismatchManifest string }
type contractFixture struct {
	Versions                                                                                map[string]contractVersion
	MergeRequest, ReverseMergeRequest, ConflictMergeRequest, InvalidCombinationMergeRequest string
	Conflicts                                                                               []map[string]string
	ExpectedOperations                                                                      map[string][]string
}
type contractPair struct {
	Suite    string `json:"suite"`
	Manifest string `json:"manifest"`
}
type contractRequest struct {
	FormatVersion           string       `json:"formatVersion"`
	MergeRequest            string       `json:"mergeRequest"`
	Base                    contractPair `json:"base"`
	Left                    contractPair `json:"left"`
	Right                   contractPair `json:"right"`
	Candidate               contractPair `json:"candidate"`
	ExpectedCandidateDigest string       `json:"expectedCandidateDigest"`
}
type contractStage struct {
	ModelSHA256           *string `json:"modelSHA256"`
	SuiteSHA256           string  `json:"suiteSHA256"`
	ManifestSHA256        string  `json:"manifestSHA256"`
	ProposalRequestSHA256 *string `json:"proposalRequestSHA256"`
	Operation             string  `json:"operation"`
	ReceiptSHA256         *string `json:"receiptSHA256"`
	ReceiptJSON           *string `json:"receiptJSON"`
}
type contractReport struct {
	FormatVersion, Mode, Authority, ReviewStatus, Action, Outcome                                                    string
	RequestSHA256, MergeRequestSHA256, MergeRequestJSON, MergeReportSHA256, MergeReportJSON, ExpectedCandidateDigest string
	Base, Left, Right, Candidate                                                                                     contractStage
}

func contractLoad(t *testing.T) contractFixture {
	t.Helper()
	raw, e := os.ReadFile("testdata/observations.json")
	if e != nil {
		t.Fatal(e)
	}
	var f contractFixture
	if e = json.Unmarshal(raw, &f); e != nil {
		t.Fatal(e)
	}
	return f
}
func contractSHA(raw []byte) string { h := sha256.Sum256(raw); return hex.EncodeToString(h[:]) }
func contractJSON(t *testing.T, v any) []byte {
	t.Helper()
	b, e := json.Marshal(v)
	if e != nil {
		t.Fatal(e)
	}
	return b
}
func contractReq(f contractFixture) contractRequest {
	pair := func(role string) contractPair { v := f.Versions[role]; return contractPair{v.Suite, v.Manifest} }
	return contractRequest{"lifecycle-custom-merge-rehearsal-request/v1", f.MergeRequest, pair("base"), pair("left"), pair("right"), pair("candidate"), f.Versions["candidate"].ModelDigest}
}
func contractPairs(r *contractRequest) []*contractPair {
	return []*contractPair{&r.Base, &r.Left, &r.Right, &r.Candidate}
}
func contractStages(r contractReport) []contractStage {
	return []contractStage{r.Base, r.Left, r.Right, r.Candidate}
}
func contractRead(t *testing.T, b []byte) contractReport {
	t.Helper()
	var r contractReport
	if e := json.Unmarshal(b, &r); e != nil {
		t.Fatalf("report: %v (%d bytes)", e, len(b))
	}
	return r
}
func contractAssertEnvelope(t *testing.T, raw, out []byte, r contractRequest) contractReport {
	t.Helper()
	got := contractRead(t, out)
	if got.FormatVersion != "lifecycle-custom-merge-rehearsal-report/v1" || got.Mode != "example-only" || got.Authority != "none" || got.ReviewStatus != "not-established" {
		t.Fatalf("incorrect fixed envelope: %+v", got)
	}
	if got.RequestSHA256 != contractSHA(raw) || got.MergeRequestSHA256 != contractSHA([]byte(r.MergeRequest)) || got.MergeRequestJSON != r.MergeRequest || got.MergeReportSHA256 != contractSHA([]byte(got.MergeReportJSON)) || got.ExpectedCandidateDigest != r.ExpectedCandidateDigest {
		t.Fatal("original envelope identity changed")
	}
	native, e := modelmerge.Execute([]byte(r.MergeRequest))
	if e != nil {
		t.Fatal(e)
	}
	// Execute returns its native encoded bytes. Compare them without re-encoding.
	if !bytes.Equal(native, []byte(got.MergeReportJSON)) {
		t.Fatal("native merge receipt bytes changed")
	}
	var source map[string]json.RawMessage
	if e = json.Unmarshal([]byte(r.MergeRequest), &source); e != nil {
		t.Fatal(e)
	}
	var merged map[string]json.RawMessage
	if e = json.Unmarshal(native, &merged); e != nil {
		t.Fatal(e)
	}
	for i, s := range contractStages(got) {
		pair := contractPairs(&r)[i]
		if s.SuiteSHA256 != contractSHA([]byte(pair.Suite)) || s.ManifestSHA256 != contractSHA([]byte(pair.Manifest)) {
			t.Fatalf("%s input text changed", contractRoles[i])
		}
		token := source[contractRoles[i]]
		if i == 3 {
			token = merged["candidate"]
		}
		if string(token) == "null" {
			if s.ModelSHA256 != nil {
				t.Fatal("nonexistent candidate hash")
			}
		} else if s.ModelSHA256 == nil || *s.ModelSHA256 != contractSHA(token) {
			t.Fatalf("%s model token changed", contractRoles[i])
		}
		if s.ReceiptJSON != nil {
			if s.ReceiptSHA256 == nil || *s.ReceiptSHA256 != contractSHA([]byte(*s.ReceiptJSON)) {
				t.Fatal("receipt byte hash")
			}
			if s.ProposalRequestSHA256 == nil || len(*s.ProposalRequestSHA256) != 64 {
				t.Fatal("missing native proposal byte identity")
			}
			if decoded, err := hex.DecodeString(*s.ProposalRequestSHA256); err != nil || len(decoded) != 32 {
				t.Fatal("invalid proposal hash")
			}
			var receipt proposalreplay.Receipt
			if err := json.Unmarshal([]byte(*s.ReceiptJSON), &receipt); err != nil {
				t.Fatal(err)
			}
			if receipt.Sources.ModelSHA256 != contractSHA(token) || receipt.Sources.SuiteSHA256 != s.SuiteSHA256 || receipt.Sources.ManifestSHA256 != s.ManifestSHA256 {
				t.Fatal("native source bytes changed")
			}
			if s.Operation == "check" {
				proposal := proposalreplay.Request{FormatVersion: "lifecycle-proposal-request/v1", Model: string(token), Suite: pair.Suite, Manifest: pair.Manifest}
				nativeCheck, err := proposalreplay.Check(contractJSON(t, proposal))
				if err != nil {
					t.Fatal(err)
				}
				encoded, err := proposalreplay.Encode(nativeCheck)
				if err != nil || string(encoded) != *s.ReceiptJSON {
					t.Fatal("native check receipt changed")
				}
			}
		}
	}
	return got
}

func TestIndependentFrozenModelAndNativePreflight(t *testing.T) {
	f := contractLoad(t)
	for _, role := range contractRoles {
		v := f.Versions[role]
		m, e := jsonformat.Decode([]byte(v.Model))
		if e != nil {
			t.Fatal(e)
		}
		d, e := canonical.Digest(m)
		if e != nil || d != v.ModelDigest {
			t.Fatalf("%s independently frozen digest: %s %v", role, d, e)
		}
		req := proposalreplay.Request{FormatVersion: "lifecycle-proposal-request/v1", Model: v.Model, Suite: v.Suite, Manifest: v.Manifest}
		receipt, e := proposalreplay.Check(contractJSON(t, req))
		if e != nil || receipt.Outcome != "checked" || receipt.CandidateStarted {
			t.Fatalf("%s native-only preflight: %v %+v", role, e, receipt)
		}
	}
}
func TestIndependentCheckAndOrientation(t *testing.T) {
	f := contractLoad(t)
	for _, reverse := range []bool{false, true} {
		r := contractReq(f)
		if reverse {
			r.MergeRequest = f.ReverseMergeRequest
			r.Left, r.Right = r.Right, r.Left
		}
		raw := contractJSON(t, r)
		before := append([]byte(nil), raw...)
		out, e := Check(raw)
		if e != nil {
			t.Fatal(e)
		}
		got := contractAssertEnvelope(t, raw, out, r)
		if got.Outcome != "checked" || got.Action != "check" || ExitCode(out, e) != 0 {
			t.Fatal("not a check-only receipt")
		}
		var merge struct {
			Candidate       json.RawMessage
			CandidateDigest string
		}
		if e = json.Unmarshal([]byte(got.MergeReportJSON), &merge); e != nil {
			t.Fatal(e)
		}
		expected, e := jsonformat.Decode([]byte(f.Versions["candidate"].Model))
		if e != nil {
			t.Fatal(e)
		}
		actual, e := jsonformat.Decode(merge.Candidate)
		if e != nil || !reflect.DeepEqual(actual, expected) || merge.CandidateDigest != f.Versions["candidate"].ModelDigest {
			t.Fatal("candidate differs from independent complete model")
		}
		for _, s := range contractStages(got) {
			if s.Operation != "check" || s.ReceiptJSON == nil {
				t.Fatal("missing preflight")
			}
			var n proposalreplay.Receipt
			if e = json.Unmarshal([]byte(*s.ReceiptJSON), &n); e != nil || n.Outcome != "checked" || n.CandidateStarted || n.Replay != "not-run" {
				t.Fatal("check executed or lost native preflight")
			}
		}
		if !bytes.Equal(raw, before) {
			t.Fatal("input mutated")
		}
	}
}
func TestIndependentSemanticRefusals(t *testing.T) {
	f := contractLoad(t)
	for _, c := range []struct{ name, merge string }{{"conflicted", f.ConflictMergeRequest}, {"invalid-combination", f.InvalidCombinationMergeRequest}} {
		t.Run(c.name, func(t *testing.T) {
			r := contractReq(f)
			r.MergeRequest = c.merge
			// Nonempty but invalid native suites prove semantic merge refusal precedes preflight.
			for _, p := range contractPairs(&r) {
				p.Suite = "unvalidated-suite"
				p.Manifest = "unvalidated-manifest"
			}
			raw := contractJSON(t, r)
			calls := 0
			out, e := executeWith(context.Background(), raw, "unused", func(context.Context, []byte, string) (proposalreplay.Receipt, error) {
				calls++
				panic("forbidden execution")
			})
			if e != nil || calls != 0 || ExitCode(out, e) != 1 {
				t.Fatalf("semantic refusal: %v, %d", e, calls)
			}
			got := contractAssertEnvelope(t, raw, out, r)
			if got.Outcome != c.name {
				t.Fatal(got.Outcome)
			}
			for _, s := range contractStages(got) {
				if s.Operation != "not-attempted" || s.ReceiptJSON != nil || s.ReceiptSHA256 != nil || s.ProposalRequestSHA256 != nil {
					t.Fatal("fabricated preflight")
				}
			}
			var mr struct {
				Candidate       json.RawMessage
				CandidateDigest *string
				Conflicts       []map[string]string
			}
			if e = json.Unmarshal([]byte(got.MergeReportJSON), &mr); e != nil {
				t.Fatal(e)
			}
			if string(mr.Candidate) != "null" || mr.CandidateDigest != nil {
				t.Fatal("candidate leaked")
			}
			if c.name == "conflicted" && !reflect.DeepEqual(mr.Conflicts, f.Conflicts) {
				t.Fatalf("conflict list: %+v", mr.Conflicts)
			}
		})
	}
}
func TestIndependentStaleAndIdentitySuppressAll(t *testing.T) {
	f := contractLoad(t)
	for i := 0; i < 4; i++ {
		r := contractReq(f)
		contractPairs(&r)[i].Suite += " "
		raw := contractJSON(t, r)
		calls := 0
		out, e := executeWith(context.Background(), raw, "unused", func(context.Context, []byte, string) (proposalreplay.Receipt, error) {
			calls++
			panic("stale execution")
		})
		if e != nil || calls != 0 || contractRead(t, out).Outcome != "stale" || ExitCode(out, e) != 1 {
			t.Fatalf("stale %d: %v calls=%d", i, e, calls)
		}
		for _, s := range contractStages(contractRead(t, out)) {
			if s.Operation != "check" {
				t.Fatal("stale discarded preflight")
			}
		}
	}
	r := contractReq(f)
	r.ExpectedCandidateDigest = strings.Repeat("0", 64)
	calls := 0
	out, e := executeWith(context.Background(), contractJSON(t, r), "unused", func(context.Context, []byte, string) (proposalreplay.Receipt, error) {
		calls++
		panic("identity execution")
	})
	if e == nil || len(out) != 0 || calls != 0 || ExitCode(out, e) != 2 {
		t.Fatal("candidate assertion not refused before execution")
	}
}

// Synthetic native receipts below establish only ordering, suppression and honest
// retention. They are never used as the observed behavior oracle.
func contractSynthetic(t *testing.T, raw []byte) proposalreplay.Receipt {
	t.Helper()
	r, e := proposalreplay.Check(raw)
	if e != nil {
		t.Fatal(e)
	}
	zero := 0
	r.Action = "execute"
	r.Outcome = "matched"
	r.Compilation = "passed"
	r.Replay = "matched"
	r.CandidateStarted = true
	r.CandidateExitCode = &zero
	r.Cleanup = "removed"
	return r
}
func TestIndependentBoundarySuppressionAndPartialReceipts(t *testing.T) {
	f := contractLoad(t)
	r := contractReq(f)
	raw := contractJSON(t, r)
	for _, kind := range []string{"error", "cleanup", "cancel", "mismatch"} {
		for target := 0; target < 4; target++ {
			t.Run(kind+contractRoles[target], func(t *testing.T) {
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				calls := 0
				var last proposalreplay.Receipt
				var captured [][]byte
				checked, err := Check(raw)
				if err != nil {
					t.Fatal(err)
				}
				checkedStages := contractStages(contractRead(t, checked))
				var originalModels map[string]json.RawMessage
				if err := json.Unmarshal([]byte(r.MergeRequest), &originalModels); err != nil {
					t.Fatal(err)
				}
				nativeMerge, err := modelmerge.Execute([]byte(r.MergeRequest))
				if err != nil {
					t.Fatal(err)
				}
				var mergeTokens map[string]json.RawMessage
				if err := json.Unmarshal(nativeMerge, &mergeTokens); err != nil {
					t.Fatal(err)
				}
				originalModels["candidate"] = mergeTokens["candidate"]
				out, e := executeWith(ctx, raw, "sentinel-source", func(c context.Context, p []byte, root string) (proposalreplay.Receipt, error) {
					if c != ctx || root != "sentinel-source" {
						t.Fatal("trusted configuration changed")
					}
					idx := calls
					calls++
					captured = append(captured, append([]byte(nil), p...))
					var request proposalreplay.Request
					if err := json.Unmarshal(p, &request); err != nil {
						t.Fatal(err)
					}
					pair := contractPairs(&r)[idx]
					if request.FormatVersion != "lifecycle-proposal-request/v1" || request.Model != string(originalModels[contractRoles[idx]]) || request.Suite != pair.Suite || request.Manifest != pair.Manifest {
						t.Fatal("executor inputs changed")
					}
					n := contractSynthetic(t, p)
					if idx == target {
						switch kind {
						case "error":
							n.Outcome = "failed"
							stage := "replay"
							n.FailureStage = &stage
							n.Replay = "failed-after-start"
							last = n
							return n, errors.New("synthetic boundary error")
						case "cleanup":
							n.Cleanup = "failed"
						case "cancel":
							cancel()
						case "mismatch":
							one := 1
							n.Outcome = "mismatched"
							n.Replay = "mismatched"
							n.CandidateExitCode = &one
						}
					}
					last = n
					return n, nil
				})
				if calls != target+1 {
					t.Fatalf("calls=%d target=%d", calls, target)
				}
				got := contractRead(t, out)
				want := "failed"
				code := 4
				if kind == "mismatch" {
					want = contractRoles[target] + "-mismatched"
					code = 1
					if e != nil {
						t.Fatal(e)
					}
				} else if e == nil {
					t.Fatal("failure lacks error")
				}
				if got.Outcome != want || ExitCode(out, e) != code {
					t.Fatalf("outcome=%s exit=%d err=%v", got.Outcome, ExitCode(out, e), e)
				}
				encoded, ee := proposalreplay.Encode(last)
				if ee != nil {
					t.Fatal(ee)
				}
				for i, s := range contractStages(got) {
					op := "check"
					if i <= target {
						op = "execute"
						if s.ProposalRequestSHA256 == nil || *s.ProposalRequestSHA256 != contractSHA(captured[i]) || checkedStages[i].ProposalRequestSHA256 == nil || *s.ProposalRequestSHA256 != *checkedStages[i].ProposalRequestSHA256 {
							t.Fatal("captured proposal byte identity changed")
						}
					}
					if s.Operation != op || s.ReceiptJSON == nil {
						t.Fatal("reached stage mislabeled")
					}
					if i == target && !bytes.Equal(encoded, []byte(*s.ReceiptJSON)) {
						t.Fatal("partial native receipt changed")
					}
				}
			})
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	calls := 0
	out, e := executeWith(ctx, raw, "unused", func(context.Context, []byte, string) (proposalreplay.Receipt, error) {
		calls++
		panic("canceled execution")
	})
	if e == nil || calls != 0 || contractRead(t, out).Outcome != "failed" || ExitCode(out, e) != 4 {
		t.Fatal("pre-call cancellation")
	}
}
func TestIndependentActualFourStageReplayAndMismatches(t *testing.T) {
	f := contractLoad(t)
	root, e := filepath.Abs("..")
	if e != nil {
		t.Fatal(e)
	}
	for target := -1; target < 4; target++ {
		name := "matched"
		if target >= 0 {
			name = contractRoles[target] + "-mismatched"
		}
		t.Run(name, func(t *testing.T) {
			r := contractReq(f)
			if target >= 0 {
				v := f.Versions[contractRoles[target]]
				*contractPairs(&r)[target] = contractPair{v.MismatchSuite, v.MismatchManifest}
			}
			raw := contractJSON(t, r)
			original := append([]byte(nil), raw...)
			out, e := Execute(context.Background(), raw, root)
			if e != nil {
				t.Fatal(e)
			}
			got := contractAssertEnvelope(t, raw, out, r)
			if got.Outcome != name || got.Action != "execute" {
				t.Fatalf("outcome=%s", got.Outcome)
			}
			wantExit := 0
			if target >= 0 {
				wantExit = 1
			}
			if ExitCode(out, e) != wantExit {
				t.Fatal("exit")
			}
			for i, s := range contractStages(got) {
				wantOp := "execute"
				if target >= 0 {
					wantOp = f.ExpectedOperations[name][i]
				}
				if s.Operation != wantOp || s.ReceiptJSON == nil {
					t.Fatalf("%s operation", contractRoles[i])
				}
				var n proposalreplay.Receipt
				if e = json.Unmarshal([]byte(*s.ReceiptJSON), &n); e != nil {
					t.Fatal(e)
				}
				if wantOp == "check" {
					if n.CandidateStarted || n.Replay != "not-run" {
						t.Fatal("suppressed stage ran")
					}
					continue
				}
				if !n.CandidateStarted || n.Compilation != "passed" || n.Cleanup != "removed" || n.FailureStage != nil {
					t.Fatalf("stage %s incomplete", contractRoles[i])
				}
				if i == target {
					if n.Outcome != "mismatched" || n.Replay != "mismatched" || n.CandidateExitCode == nil || *n.CandidateExitCode != 1 {
						t.Fatal("actual mismatch missing")
					}
					continue
				}
				if n.Outcome != "matched" || n.Replay != "matched" || n.CandidateExitCode == nil || *n.CandidateExitCode != 0 {
					t.Fatal("actual match missing")
				}
				var report struct {
					Entries []struct {
						Expected map[string]any
						Actual   map[string]any
						Matched  bool
					}
					StepsExecuted int
				}
				if e = json.Unmarshal(n.CandidateReport, &report); e != nil {
					t.Fatal(e)
				}
				if len(report.Entries) != 4 || report.StepsExecuted != 4 {
					t.Fatal("did not observe all independent witnesses")
				}
				var suite struct {
					Cases []struct {
						Steps []struct{ Expected map[string]any }
					}
				}
				if e = json.Unmarshal([]byte(f.Versions[contractRoles[i]].Suite), &suite); e != nil {
					t.Fatal(e)
				}
				for j, entry := range report.Entries {
					literal := suite.Cases[j].Steps[0].Expected
					if !entry.Matched || !reflect.DeepEqual(entry.Expected, literal) {
						t.Fatal("literal expectation changed")
					}
					for key, value := range literal {
						if !reflect.DeepEqual(entry.Actual[key], value) {
							t.Fatalf("actual %s differs from frozen literal", key)
						}
					}
				}
			}
			if !bytes.Equal(raw, original) {
				t.Fatal("input changed")
			}
		})
	}
}
func TestIndependentStrictOuterAndConcurrentCheck(t *testing.T) {
	f := contractLoad(t)
	r := contractReq(f)
	valid := contractJSON(t, r)
	bad := [][]byte{append(append([]byte(nil), valid...), []byte(" {}")...), []byte(`{"formatVersion":"lifecycle-custom-merge-rehearsal-request/v1","formatVersion":"duplicate"}`), []byte(`{"formatVersion":"\ud800"}`), {0xff}, bytes.Repeat([]byte(" "), (32<<20)+1)}
	var fields map[string]json.RawMessage
	if e := json.Unmarshal(valid, &fields); e != nil {
		t.Fatal(e)
	}
	for key := range fields {
		copyFields := map[string]json.RawMessage{}
		for k, v := range fields {
			copyFields[k] = v
		}
		delete(copyFields, key)
		bad = append(bad, contractJSON(t, copyFields))
		copyFields[key] = json.RawMessage("null")
		bad = append(bad, contractJSON(t, copyFields))
	}
	fields["unknown"] = json.RawMessage("true")
	bad = append(bad, contractJSON(t, fields))
	for i, b := range bad {
		out, e := Check(b)
		if e == nil || len(out) != 0 || ExitCode(out, e) != 2 {
			t.Fatalf("invalid input %d accepted", i)
		}
	}
	reference, e := Check(valid)
	if e != nil {
		t.Fatal(e)
	}
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			out, e := Check(valid)
			if e != nil || !bytes.Equal(out, reference) {
				t.Error("concurrent Check changed")
			}
		}()
	}
	wg.Wait()
}
