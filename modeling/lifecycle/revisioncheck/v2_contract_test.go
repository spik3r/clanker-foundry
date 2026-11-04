package revisioncheck

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
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/scenariodiff"
)

// This harness compiles before CheckV2 exists: the future public API is supplied
// only through assertFrozenV2Contract's callback after production is added.
// Static fields/order expectations are independent; native parity preserves bytes.
type v2Observation struct {
	ID, Request, Status string
	Valid               bool
	Reason              *string
	ExpectedDiff        map[string]json.RawMessage
}

func v2SHA(raw []byte) string { s := sha256.Sum256(raw); return hex.EncodeToString(s[:]) }
func v2Observations(t *testing.T) []v2Observation {
	t.Helper()
	raw, e := os.ReadFile("testdata/v2-observations.json")
	if e != nil {
		t.Fatal(e)
	}
	if v2SHA(raw) != "274b0329ecbd039905d4042d38b5ce0e245e8cf8eed8798a4ace870026703387" {
		t.Fatal("frozen independent v2 corpus changed")
	}
	var x struct {
		V1SourceSHA256 string
		Vectors        []v2Observation
	}
	if json.Unmarshal(raw, &x) != nil {
		t.Fatal("v2 fixture")
	}
	source, e := os.ReadFile("check.go")
	if e != nil || v2SHA(source) != x.V1SourceSHA256 {
		t.Fatal("original v1 source changed")
	}
	return x.Vectors
}
func v2V1Request(t *testing.T, raw string) (Request, []byte) {
	t.Helper()
	var q Request
	if json.Unmarshal([]byte(raw), &q) != nil {
		t.Fatal("fixture wrapper")
	}
	q.FormatVersion = RequestVersion
	b, e := json.Marshal(q)
	if e != nil {
		t.Fatal(e)
	}
	return q, b
}
func v2NativeScenarioRequest(t *testing.T, q Request) []byte {
	t.Helper()
	parts := func(s Snapshot) (string, string) {
		var f firstdesignreview.Request
		var p struct{ Model, Suite string }
		if json.Unmarshal([]byte(s.FirstDesignRequest), &f) != nil || json.Unmarshal([]byte(f.ProposalRequest), &p) != nil {
			t.Fatal("fixture native strings")
		}
		return p.Model, p.Suite
	}
	bm, bs := parts(q.Before)
	am, as := parts(q.After)
	return []byte("{\"formatVersion\":\"lifecycle-scenario-diff-json/v1\",\"beforeModel\":" + bm + ",\"beforeSuite\":" + bs + ",\"afterModel\":" + am + ",\"afterSuite\":" + as + "}\n")
}
func v2AssertDeclaredDiff(t *testing.T, raw []byte, want map[string]json.RawMessage) {
	t.Helper()
	var got map[string]json.RawMessage
	if json.Unmarshal(raw, &got) != nil {
		t.Fatal("native scenario report")
	}
	for k, w := range want {
		var a, b any
		if json.Unmarshal(got[k], &a) != nil || json.Unmarshal(w, &b) != nil || !reflect.DeepEqual(a, b) {
			t.Fatalf("%s: got%s want%s", k, got[k], w)
		}
	}
}
func TestV2FrozenNativeInputsAndExpectations(t *testing.T) {
	for _, v := range v2Observations(t) {
		t.Run(v.ID, func(t *testing.T) {
			q, nativeRequest := v2V1Request(t, v.Request)
			retained, e := Check(nativeRequest)
			if !v.Valid {
				if !errors.Is(e, ErrInvalid) || retained != nil {
					t.Fatal("invalid native framing accepted")
				}
				return
			}
			if e != nil {
				t.Fatal("valid retained report fixture", e)
			}
			native, e := scenariodiff.Execute(v2NativeScenarioRequest(t, q))
			if v.Status == "not-checked" {
				if !errors.Is(e, scenariodiff.ErrInvalid) || native != nil {
					t.Fatal("native refusal fixture did not refuse")
				}
				return
			}
			if e != nil {
				t.Fatal("native comparison fixture", e)
			}
			v2AssertDeclaredDiff(t, native, v.ExpectedDiff)
		})
	}
}
func assertFrozenV2Contract(t *testing.T, check func([]byte) ([]byte, error)) {
	t.Helper()
	for _, v := range v2Observations(t) {
		t.Run(v.ID, func(t *testing.T) {
			input := []byte(v.Request)
			original := bytes.Clone(input)
			raw, e := check(input)
			if !bytes.Equal(input, original) {
				t.Fatal("v2 input mutated")
			}
			if !v.Valid {
				if !errors.Is(e, ErrInvalid) || raw != nil {
					t.Fatal("invalid request returned report")
				}
				return
			}
			if e != nil {
				t.Fatal(e)
			}
			var envelope struct {
				FormatVersion, Mode, Authority, ExecutionStatus, ReviewStatus string
				ScenarioDiff                                                  struct {
					Status                           string
					Reason, ReportSHA256, ReportJSON *string
				}
			}
			if json.Unmarshal(raw, &envelope) != nil || envelope.FormatVersion != "lifecycle-revision-check-report/v2" || envelope.Mode != "example-only" || envelope.Authority != "none" || envelope.ExecutionStatus != "not-run" || envelope.ReviewStatus != "not-established" {
				t.Fatal("v2 envelope or authority changed")
			}
			if envelope.ScenarioDiff.Status != v.Status || !reflect.DeepEqual(envelope.ScenarioDiff.Reason, v.Reason) {
				t.Fatal("native comparison status/reason", envelope.ScenarioDiff)
			}
			q, v1Request := v2V1Request(t, v.Request)
			nativeV1, e := Check(v1Request)
			if e != nil {
				t.Fatal(e)
			}
			var old, all map[string]any
			if json.Unmarshal(nativeV1, &old) != nil || json.Unmarshal(raw, &all) != nil {
				t.Fatal("reports")
			}
			delete(all, "scenarioDiff")
			all["formatVersion"] = old["formatVersion"]
			if !reflect.DeepEqual(all, old) {
				t.Fatal("v2 altered an original source observation or native report byte string/hash")
			}
			if v.Status == "not-checked" {
				if envelope.ScenarioDiff.ReportJSON != nil || envelope.ScenarioDiff.ReportSHA256 != nil {
					t.Fatal("refused comparison returned report/hash")
				}
				return
			}
			if envelope.ScenarioDiff.Reason != nil || envelope.ScenarioDiff.ReportJSON == nil || envelope.ScenarioDiff.ReportSHA256 == nil {
				t.Fatal("checked report incomplete")
			}
			native, e := scenariodiff.Execute(v2NativeScenarioRequest(t, q))
			if e != nil || string(native) != *envelope.ScenarioDiff.ReportJSON || v2SHA(native) != *envelope.ScenarioDiff.ReportSHA256 {
				t.Fatal("native scenario bytes/hash not preserved")
			}
			v2AssertDeclaredDiff(t, native, v.ExpectedDiff)
			again, e := check(input)
			if e != nil || !bytes.Equal(again, raw) {
				t.Fatal("repeated check changed")
			}
		})
	}
	valid := []byte(v2Observations(t)[0].Request)
	for _, bad := range [][]byte{nil, []byte("null"), []byte(`{}`), append(bytes.Clone(valid), valid...), bytes.Replace(valid, []byte(`"before":`), []byte(`"BEFORE":`), 1), bytes.Replace(valid, []byte(`"before":`), []byte(`"extra":null,"before":`), 1)} {
		if out, e := check(bad); e == nil || out != nil {
			t.Fatal("strict v2 wrapper accepted")
		}
	}
	if out, e := check(bytes.Repeat([]byte(" "), MaxRequestBytes+1)); !errors.Is(e, ErrLimit) || out != nil {
		t.Fatal("v2 request limit")
	}
}
