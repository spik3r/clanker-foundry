package rehearsalinspection

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestIndependentObservationVectors(t *testing.T) {
	bytes := read(t, "testdata/observations.json")
	if hash(bytes) != "fb6433dfe6f2fa1c2e45f8390df9d081bcdab3b007fc9f95abdd9bf066f0efb8" {
		t.Fatal("frozen separate-contributor vectors changed")
	}
	var vectors struct {
		Receipt Receipt
		Vectors []struct {
			ID, Side, Expected string
			Report             json.RawMessage
		}
		RawVectors []struct{ ID, Side, ReportText, Expected string }
	}
	if e := json.Unmarshal(bytes, &vectors); e != nil {
		t.Fatal(e)
	}
	original, refs := fixture(t)
	got, e := Inspect(carrier(t, original, refs), refs)
	if e != nil || !reflect.DeepEqual(got, vectors.Receipt) {
		t.Fatalf("receipt %#v err %v", got, e)
	}
	for _, v := range vectors.Vectors {
		t.Run(v.ID, func(t *testing.T) {
			m, r := original, refs
			raw := compact(t, v.Report)
			if v.Side == "baseline" {
				m.BaselineReport = raw
				r.BaselineReportSHA256 = hash(raw)
			} else {
				m.CandidateReport = raw
				r.CandidateReportSHA256 = hash(raw)
			}
			b := carrier(t, m, r)
			got, e := Inspect(b, r)
			if v.Expected == "reject" {
				if e == nil || got != (Receipt{}) {
					t.Fatal("inconsistent report accepted")
				}
				return
			}
			if e != nil {
				t.Fatal(e)
			}
			if v.ID == "honest_early_mismatch" {
				if got.Candidate.Matched || got.Candidate.CasesCompleted != 0 || got.Candidate.StepsExecuted != 1 {
					t.Fatal("early mismatch lost")
				}
			} else if !reflect.DeepEqual(got, vectors.Receipt) {
				t.Fatal("literal receipt mismatch")
			}
		})
	}
	for _, v := range vectors.RawVectors {
		t.Run(v.ID, func(t *testing.T) {
			m, r := original, refs
			raw := []byte(v.ReportText)
			m.CandidateReport = raw
			r.CandidateReportSHA256 = hash(raw)
			if _, e := Inspect(carrier(t, m, r), r); e == nil {
				t.Fatal("invalid raw report accepted")
			}
		})
	}
}
