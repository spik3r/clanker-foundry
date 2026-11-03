package scenarios

import (
	"encoding/json"
	lifecycle "github.com/spik3r/clanker-foundry/modeling/lifecycle/generated"
	"os"
	"testing"
)

func TestRunnerDecodeNativeReportDoesNotInvoke(t *testing.T) {
	raw, e := os.ReadFile("../rehearsalhandoff/testdata/report-identities.json")
	if e != nil {
		t.Fatal(e)
	}
	var x struct {
		Baseline, Candidate struct{ Report json.RawMessage }
	}
	if json.Unmarshal(raw, &x) != nil {
		t.Fatal("fixture")
	}
	for _, raw := range []json.RawMessage{x.Baseline.Report, x.Candidate.Report} {
		var meta struct{ ModelDigest string }
		if json.Unmarshal(raw, &meta) != nil {
			t.Fatal("report")
		}
		r, e := NewRunner(Binding{lifecycle.ProfileVersion, meta.ModelDigest}, func(lifecycle.Task, lifecycle.Command) lifecycle.Result {
			t.Fatal("callback invoked")
			return lifecycle.Result{}
		})
		if e != nil {
			t.Fatal(e)
		}
		if _, e = r.DecodeReport(raw); e != nil {
			t.Fatal(e)
		}
		r.binding.ModelDigest = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
		if _, e = r.DecodeReport(raw); e == nil {
			t.Fatal("wrong model accepted")
		}
	}
	if _, e := (Runner{}).DecodeReport(x.Baseline.Report); e == nil {
		t.Fatal("zero runner accepted")
	}
}
