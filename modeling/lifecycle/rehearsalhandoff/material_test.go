package rehearsalhandoff

import (
	"bytes"
	"testing"
)

func TestInspectMaterialOwnsDecodedBytes(t *testing.T) {
	m, r := fixture(t)
	b, e := Build(m, r)
	if e != nil {
		t.Fatal(e)
	}
	before := append([]byte{}, b...)
	got, receipt, e := InspectMaterial(b, r)
	if e != nil || receipt.ReportBytesStatus != ReportStatus {
		t.Fatal(e)
	}
	got.CandidateReport[0] = 'x'
	got.CandidateSuite[0] = 'x'
	if !bytes.Equal(b, before) || bytes.Equal(got.CandidateReport, m.CandidateReport) || bytes.Equal(got.CandidateSuite, m.CandidateSuite) {
		t.Fatal("material ownership")
	}
	again, _, e := InspectMaterial(b, r)
	if e != nil || !bytes.Equal(again.CandidateReport, m.CandidateReport) {
		t.Fatal("carrier mutated")
	}
	if got, receipt, e := InspectMaterial([]byte("{}"), r); e == nil || got.Request != nil || receipt != (Receipt{}) {
		t.Fatal("partial material returned")
	}
}
