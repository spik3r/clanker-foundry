package revisioncheck

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/firstdesignreview"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/intakeimpact"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/traceability"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/traceimpact"
)

func TestNativeBytesUnchangedAndNoInputMutation(t *testing.T) {
	raw := []byte(observations(t)[0].Request)
	copyRaw := append([]byte{}, raw...)
	out, e := Check(raw)
	if e != nil {
		t.Fatal(e)
	}
	if !bytes.Equal(raw, copyRaw) {
		t.Fatal("input mutated")
	}
	var req Request
	var r Report
	if json.Unmarshal(raw, &req) != nil || json.Unmarshal(out, &r) != nil {
		t.Fatal("fixtures")
	}
	var before, after firstdesignreview.Request
	if json.Unmarshal([]byte(req.Before.FirstDesignRequest), &before) != nil || json.Unmarshal([]byte(req.After.FirstDesignRequest), &after) != nil {
		t.Fatal("request")
	}
	intakeReq, _ := json.Marshal(map[string]any{"formatVersion": intakeimpact.RequestVersion, "before": before.IntakeRequest, "after": after.IntakeRequest})
	native, e := intakeimpact.Execute(intakeReq)
	if e != nil || string(native) != r.IntakeImpactReportJSON || hash(native) != r.IntakeImpactReportSHA256 {
		t.Fatal("intake native bytes changed")
	}
	var b, a struct{ Model, Suite, Manifest string }
	if json.Unmarshal([]byte(before.ProposalRequest), &b) != nil || json.Unmarshal([]byte(after.ProposalRequest), &a) != nil {
		t.Fatal("proposal")
	}
	traceReq, _ := json.Marshal(map[string]any{"formatVersion": traceimpact.RequestVersion, "before": map[string]string{"model": b.Model, "requirements": req.Before.Requirements, "catalog": req.Before.Catalog}, "after": map[string]string{"model": a.Model, "requirements": req.After.Requirements, "catalog": req.After.Catalog}})
	native, e = traceimpact.Execute(traceReq)
	if e != nil || string(native) != r.TraceImpactReportJSON || hash(native) != r.TraceImpactReportSHA256 {
		t.Fatal("trace native bytes changed")
	}
	if b.Suite != a.Suite || b.Manifest != a.Manifest {
		t.Fatal("frozen literal suite/manifest changed")
	}
	var wg sync.WaitGroup
	for range 4 {
		wg.Go(func() {
			again, e := Check(raw)
			if e != nil || !bytes.Equal(again, out) {
				t.Error("repeat check changed")
			}
		})
	}
	wg.Wait()
	req.After = req.Before
	same, _ := json.Marshal(req)
	sameOut, e := Check(same)
	if e != nil {
		t.Fatal(e)
	}
	if json.Unmarshal(sameOut, &r) != nil {
		t.Fatal("report")
	}
	var trace traceimpact.Report
	if json.Unmarshal([]byte(r.TraceImpactReportJSON), &trace) != nil || trace.ReviewRequired || trace.ModelChanged || trace.RequirementsChanged || trace.CatalogSourceChanged {
		t.Fatal("unchanged design invented change")
	}
}
func TestStrictAndLimits(t *testing.T) {
	var corpus struct{ RawRejections []struct{ Request string } }
	raw, e := os.ReadFile("testdata/observations.json")
	if e != nil || json.Unmarshal(raw, &corpus) != nil {
		t.Fatal("corpus")
	}
	invalid := [][]byte{nil, []byte("null"), []byte(`{"formatVersion":"bad"}`), []byte(`{"formatVersion":"lifecycle-revision-check-request/v1","before":null,"after":null}`)}
	for _, v := range corpus.RawRejections {
		invalid = append(invalid, []byte(v.Request))
	}
	valid := []byte(observations(t)[0].Request)
	invalid = append(invalid, append(append([]byte{}, valid...), valid...), bytes.Replace(valid, []byte(`"before":`), []byte(`"BEFORE":`), 1), bytes.Replace(valid, []byte(`"catalog":`), []byte(`"callback":"none","catalog":`), 1))
	for _, raw := range invalid {
		if out, e := Check(raw); e == nil || out != nil {
			t.Fatal("invalid request accepted")
		}
	}
	var req Request
	if json.Unmarshal(valid, &req) != nil {
		t.Fatal("fixture")
	}
	for _, modify := range []func(*Request){func(r *Request) {
		r.After.FirstDesignRequest = strings.Repeat(" ", firstdesignreview.MaxRequestBytes+1)
	}, func(r *Request) { r.After.Requirements = strings.Repeat(" ", traceability.MaxBytes+1) }, func(r *Request) { r.After.Catalog = strings.Repeat(" ", 4*traceability.MaxBytes+1) }} {
		changed := req
		modify(&changed)
		raw, _ := json.Marshal(changed)
		if out, e := Check(raw); !errors.Is(e, ErrLimit) || out != nil {
			t.Fatal("native limit lost")
		}
	}
	if out, e := Check(bytes.Repeat([]byte(" "), MaxRequestBytes+1)); !errors.Is(e, ErrLimit) || out != nil {
		t.Fatal("wrapper limit")
	}
	if _, e := encode(map[string]string{"data": "x"}, 1); !errors.Is(e, ErrLimit) {
		t.Fatal("output limit")
	}
	if _, e := encode(make(chan int), 100); !errors.Is(e, ErrBoundary) {
		t.Fatal("encoding boundary")
	}
	for _, err := range []error{firstdesignreview.ErrBoundary, errors.New("unexpected")} {
		if !errors.Is(nativeError(err), ErrBoundary) {
			t.Fatal("native boundary classification")
		}
	}
	if reflect.DeepEqual(req.Before, Snapshot{}) {
		t.Fatal("fixture empty")
	}
}
