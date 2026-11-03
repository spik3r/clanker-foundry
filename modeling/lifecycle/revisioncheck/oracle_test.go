package revisioncheck

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/firstdesignreview"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/intakemodel"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/proposalreplay"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/traceimpact"
)

type vector struct {
	ID, Request string
	Expected    map[string]json.RawMessage
}

func observations(t *testing.T) []vector {
	t.Helper()
	raw, e := os.ReadFile("testdata/observations.json")
	if e != nil {
		t.Fatal(e)
	}
	if hash(raw) != "7987c38be6221b2ae1a44b45cfd0ab81fbe7dcc70dbab7ff3f9ed18fd4caa669" {
		t.Fatal("frozen separate-contributor vectors changed")
	}
	var x struct{ Vectors []vector }
	if json.Unmarshal(raw, &x) != nil {
		t.Fatal("vectors")
	}
	return x.Vectors
}
func get[T any](t *testing.T, raw json.RawMessage) T {
	t.Helper()
	var v T
	if json.Unmarshal(raw, &v) != nil {
		t.Fatal("invalid observation")
	}
	return v
}
func assertExpected(t *testing.T, w map[string]json.RawMessage, key string, got any) {
	t.Helper()
	raw, e := json.Marshal(got)
	if e != nil {
		t.Fatal(e)
	}
	var actual, expected any
	if json.Unmarshal(raw, &actual) != nil || json.Unmarshal(w[key], &expected) != nil || !reflect.DeepEqual(actual, expected) {
		t.Fatalf("%s got %s want %s", key, raw, w[key])
	}
}
func TestIndependentCustomRevisionVectors(t *testing.T) {
	for _, v := range observations(t) {
		t.Run(v.ID, func(t *testing.T) {
			result, e := Check([]byte(v.Request))
			if !get[bool](t, v.Expected["valid"]) {
				if e == nil || result != nil {
					t.Fatal("invalid request yielded report")
				}
				return
			}
			if e != nil {
				t.Fatal(e)
			}
			var r Report
			if json.Unmarshal(result, &r) != nil {
				t.Fatal("report")
			}
			for key, got := range map[string]any{"formatVersion": r.FormatVersion, "mode": r.Mode, "authority": r.Authority, "executionStatus": r.ExecutionStatus, "reviewStatus": r.ReviewStatus, "beforeRequirementsComparison": r.Before.RequirementsComparison, "afterRequirementsComparison": r.After.RequirementsComparison} {
				assertExpected(t, v.Expected, key, got)
			}
			type native struct {
				ModelComparison, SuiteComparison, ProposalReportJSON string
				IntakeReport                                         intakemodel.Report
			}
			var before, after native
			if json.Unmarshal([]byte(r.Before.FirstDesignReportJSON), &before) != nil || json.Unmarshal([]byte(r.After.FirstDesignReportJSON), &after) != nil {
				t.Fatal("native first design")
			}
			for key, got := range map[string]any{"beforeModelComparison": before.ModelComparison, "afterModelComparison": after.ModelComparison, "beforeSuiteComparison": before.SuiteComparison, "afterSuiteComparison": after.SuiteComparison} {
				assertExpected(t, v.Expected, key, got)
			}
			var bp, ap proposalreplay.Receipt
			if json.Unmarshal([]byte(before.ProposalReportJSON), &bp) != nil || json.Unmarshal([]byte(after.ProposalReportJSON), &ap) != nil {
				t.Fatal("native proposal")
			}
			for key, got := range map[string]any{"beforeProposalOutcome": bp.Outcome, "afterProposalOutcome": ap.Outcome, "beforeSuiteValidation": bp.Preflight.SuiteValidation, "afterSuiteValidation": ap.Preflight.SuiteValidation, "beforeModelDigest": bp.Preflight.CurrentModelDigest, "afterModelDigest": ap.Preflight.CurrentModelDigest} {
				assertExpected(t, v.Expected, key, got)
			}
			if bp.CandidateStarted || ap.CandidateStarted || bp.Preflight.Replay != "not-run" || ap.Preflight.Replay != "not-run" {
				t.Fatal("check claimed execution")
			}
			for _, side := range []native{before, after} {
				var intake struct{ Questions []struct{ ID string } }
				if json.Unmarshal([]byte(side.IntakeReport.IntakeReportJSON), &intake) != nil {
					t.Fatal("native intake")
				}
				ids := []string{}
				for _, q := range intake.Questions {
					ids = append(ids, q.ID)
				}
				assertExpected(t, v.Expected, "questionIDs", ids)
			}
			var trace traceimpact.Report
			if json.Unmarshal([]byte(r.TraceImpactReportJSON), &trace) != nil {
				t.Fatal("trace")
			}
			assertExpected(t, v.Expected, "traceModelChanged", trace.ModelChanged)
			ids := []string{}
			for _, c := range trace.Candidates {
				ids = append(ids, c.ID)
			}
			assertExpected(t, v.Expected, "traceCandidateRequirementIDs", ids)
			required := get[string](t, v.Expected["traceGapRequired"])
			for _, gaps := range [][]traceimpact.Gap{trace.BeforeGaps, trace.AfterGaps} {
				found := false
				for _, g := range gaps {
					for _, code := range g.Codes {
						found = found || code == required
					}
				}
				if !found {
					t.Fatal("declared coverage gap lost")
				}
			}
			if _, exists := v.Expected["traceCatalogChanged"]; exists {
				assertExpected(t, v.Expected, "traceCatalogChanged", trace.CatalogSourceChanged)
				required := get[string](t, v.Expected["traceCatalogReasonRequired"])
				found := false
				for _, c := range trace.Candidates {
					for _, reason := range c.Reasons {
						found = found || reason.Code == required
					}
				}
				if !found {
					t.Fatal("catalog source reason lost")
				}
			}
			var req Request
			if json.Unmarshal([]byte(v.Request), &req) != nil {
				t.Fatal("request")
			}
			type hashes struct {
				FirstDesignRequestSHA256 string `json:"firstDesignRequestSHA256"`
				RequirementsSHA256       string `json:"requirementsSHA256"`
				CatalogSHA256            string `json:"catalogSHA256"`
			}
			assertExpected(t, v.Expected, "sideSourceHashes", map[string]hashes{"before": {r.Before.FirstDesignRequestSHA256, r.Before.RequirementsSHA256, r.Before.CatalogSHA256}, "after": {r.After.FirstDesignRequestSHA256, r.After.RequirementsSHA256, r.After.CatalogSHA256}})
			// Marshal typed hashes with the declared field spelling for expected JSON.
			for _, pair := range []struct {
				Snapshot Snapshot
				Side     SideReport
			}{{req.Before, r.Before}, {req.After, r.After}} {
				native, e := firstdesignreview.Check([]byte(pair.Snapshot.FirstDesignRequest))
				if e != nil || string(native) != pair.Side.FirstDesignReportJSON || hash(native) != pair.Side.FirstDesignReportSHA256 {
					t.Fatal("native first-design bytes changed")
				}
			}
		})
	}
}
