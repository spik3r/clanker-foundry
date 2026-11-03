package firstdesignreview_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"testing"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/firstdesignreview"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/intakemodel"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/proposalreplay"
)

// Expectations were frozen by a separate author at 20dfd89, before production
// implementation. Native API calls below test continuity, never derive oracle
// expectations. No candidate execution API is used.
type oracleExpectation struct {
	Valid              bool     `json:"valid"`
	ModelComparison    string   `json:"modelComparison"`
	SuiteComparison    string   `json:"suiteComparison"`
	IntakeModelStatus  string   `json:"intakeModelStatus"`
	IntakeModelBinding string   `json:"intakeModelBinding"`
	IntakeSuiteBinding string   `json:"intakeSuiteBinding"`
	ProposalOutcome    string   `json:"proposalOutcome"`
	SuiteValidation    string   `json:"suiteValidation"`
	PreflightReplay    string   `json:"preflightReplay"`
	QuestionIDs        []string `json:"questionIDs"`
}

type oracleReport struct {
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

func oracleHash(raw []byte) string {
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}

func oracleDecode(t *testing.T, raw []byte, dst any) {
	t.Helper()
	if err := json.Unmarshal(raw, dst); err != nil {
		t.Fatalf("decode: %v", err)
	}
}

func TestFrozenIndependentOracle(t *testing.T) {
	raw, err := os.ReadFile("testdata/oracle.json")
	if err != nil {
		t.Fatal(err)
	}
	if oracleHash(raw) != "b6c5705d00349d62532f079f4770448662e756e604a0f5b7d48d8ed62caffa0a" {
		t.Fatal("independently frozen oracle bytes changed")
	}
	var fixture struct {
		Cases []struct {
			ID       string            `json:"id"`
			Request  string            `json:"request"`
			Expected oracleExpectation `json:"expected"`
		} `json:"cases"`
	}
	oracleDecode(t, raw, &fixture)
	if len(fixture.Cases) != 33 {
		t.Fatalf("case count = %d, want 33", len(fixture.Cases))
	}
	seen := map[string]bool{}
	for _, c := range fixture.Cases {
		if c.ID == "" || seen[c.ID] {
			t.Fatalf("empty/duplicate case ID %q", c.ID)
		}
		seen[c.ID] = true
		t.Run(c.ID, func(t *testing.T) {
			input := []byte(c.Request)
			before := bytes.Clone(input)
			got, err := firstdesignreview.Check(input)
			if !bytes.Equal(input, before) {
				t.Fatal("input mutated")
			}
			again, againErr := firstdesignreview.Check(input)
			if !bytes.Equal(input, before) {
				t.Fatal("repeat check mutated input")
			}
			if !bytes.Equal(got, again) || !errors.Is(againErr, err) {
				t.Fatal("check is not repeatable")
			}
			if !c.Expected.Valid {
				if err == nil || got != nil || again != nil {
					t.Fatalf("invalid case produced report or nil error: %v", err)
				}
				if !errors.Is(err, firstdesignreview.ErrInvalid) {
					t.Fatalf("invalid request error = %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("valid report case refused: %v", err)
			}
			var r oracleReport
			oracleDecode(t, got, &r)
			if r.IntakeReport == nil {
				t.Fatal("missing intake report")
			}
			var p proposalreplay.Receipt
			oracleDecode(t, []byte(r.ProposalReportJSON), &p)
			var intake struct {
				Questions []struct {
					ID string `json:"id"`
				} `json:"questions"`
				Bindings []struct {
					Kind   string `json:"kind"`
					Status string `json:"status"`
				} `json:"bindings"`
			}
			oracleDecode(t, []byte(r.IntakeReport.IntakeReportJSON), &intake)
			bindings := map[string]string{}
			for _, b := range intake.Bindings {
				if _, exists := bindings[b.Kind]; exists {
					t.Fatalf("duplicate binding %q", b.Kind)
				}
				bindings[b.Kind] = b.Status
			}
			ids := make([]string, 0, len(intake.Questions))
			for _, q := range intake.Questions {
				ids = append(ids, q.ID)
			}
			observed := oracleExpectation{
				Valid: true, ModelComparison: r.ModelComparison, SuiteComparison: r.SuiteComparison,
				IntakeModelStatus: r.IntakeReport.Model.Status, IntakeModelBinding: bindings["model"],
				IntakeSuiteBinding: bindings["scenarioSuite"], ProposalOutcome: p.Outcome,
				SuiteValidation: p.Preflight.SuiteValidation, PreflightReplay: p.Preflight.Replay, QuestionIDs: ids,
			}
			if !reflect.DeepEqual(observed, c.Expected) {
				t.Fatalf("frozen observation mismatch\ngot:  %+v\nwant: %+v", observed, c.Expected)
			}
			if r.FormatVersion != "lifecycle-first-design-review-report/v1" || r.Mode != "example-only" || r.Authority != "none" || r.ExecutionStatus != "not-run" || r.ReviewStatus != "not-established" {
				t.Fatal("combined no-authority envelope changed")
			}
			if p.CandidateStarted || p.Action != "check" || p.Compilation != "not-requested" || p.Replay != "not-requested" || p.CandidateExitCode != nil || p.ArchiveSHA256 != nil || p.BinarySHA256 != nil || p.SourceRevision != nil || string(p.CandidateReport) != "null" || p.Cleanup != "not-created" {
				t.Fatal("proposal report claims execution artifacts or changed check-only state")
			}
			// A separate continuity property: equality with the entire native typed
			// intake report and the exact native proposal encoding, including final LF.
			t.Run("native_continuity", func(t *testing.T) {
				var request firstdesignreview.Request
				oracleDecode(t, input, &request)
				nativeIntake, err := intakemodel.Inspect([]byte(request.IntakeRequest))
				if err != nil {
					t.Fatalf("native intake: %v", err)
				}
				if !reflect.DeepEqual(r.IntakeReport, nativeIntake) {
					t.Fatal("complete typed intake report not preserved")
				}
				nativeProposal, err := proposalreplay.Check([]byte(request.ProposalRequest))
				if err != nil {
					t.Fatalf("native proposal: %v", err)
				}
				nativeBytes, err := proposalreplay.Encode(nativeProposal)
				if err != nil {
					t.Fatal(err)
				}
				if !bytes.Equal([]byte(r.ProposalReportJSON), nativeBytes) || r.ProposalReportSHA256 != oracleHash(nativeBytes) {
					t.Fatal("exact proposal report bytes/hash not preserved")
				}
				if r.IntakeRequestSHA256 != oracleHash([]byte(request.IntakeRequest)) || r.ProposalRequestSHA256 != oracleHash([]byte(request.ProposalRequest)) {
					t.Fatal("original request byte hash mismatch")
				}
				if r.IntakeReport.IntakeReportSHA256 != oracleHash([]byte(r.IntakeReport.IntakeReportJSON)) {
					t.Fatal("native intake text hash mismatch")
				}
			})
		})
	}
}
