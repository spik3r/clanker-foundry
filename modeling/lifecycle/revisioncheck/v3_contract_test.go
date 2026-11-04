package revisioncheck

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"sync"
	"testing"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/firstdesignreview"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/impact"
)

// This harness compiles before CheckV3 exists. The future public entrypoint is
// supplied only as a callback. Hand-stated full changes and declared links are
// separate from native byte parity, which cannot establish behavior or approval.
type v3Observation struct {
	ID, Request, ScenarioStatus                    string
	Valid                                          bool
	ScenarioReason                                 *string
	EmbeddedModelSourcesEqual, FullModelTextsEqual bool
	ExpectedModelDiff                              map[string]json.RawMessage
}

func v3SHA(raw []byte) string {
	d := sha256.Sum256(raw)
	return hex.EncodeToString(d[:])
}

func v3Observations(t *testing.T) []v3Observation {
	t.Helper()
	raw, err := os.ReadFile("testdata/v3-observations.json")
	if err != nil || v3SHA(raw) != "e0c892b4c65a295eba8ffb68257399317f2352a770b7b531fdb9391446df9098" {
		t.Fatal("frozen v3 corpus changed", err)
	}
	var corpus struct {
		PreservedFileSHA256 map[string]string
		Vectors             []v3Observation
	}
	if json.Unmarshal(raw, &corpus) != nil || len(corpus.Vectors) != 18 {
		t.Fatal("v3 corpus framing")
	}
	for path, want := range corpus.PreservedFileSHA256 {
		source, err := os.ReadFile(path)
		if err != nil || v3SHA(source) != want {
			t.Fatalf("original source or fixture changed: %s", path)
		}
	}
	return corpus.Vectors
}

func v3RequestAsV2(t *testing.T, raw string) (Request, []byte) {
	t.Helper()
	var req Request
	if json.Unmarshal([]byte(raw), &req) != nil {
		t.Fatal("fixture wrapper")
	}
	req.FormatVersion = "lifecycle-revision-check-request/v2"
	out, err := json.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}
	return req, out
}

func v3ModelTexts(t *testing.T, req Request) (string, string) {
	t.Helper()
	model := func(side Snapshot) string {
		var first firstdesignreview.Request
		var proposal struct{ Model string }
		if json.Unmarshal([]byte(side.FirstDesignRequest), &first) != nil || json.Unmarshal([]byte(first.ProposalRequest), &proposal) != nil {
			t.Fatal("fixture native strings")
		}
		return proposal.Model
	}
	return model(req.Before), model(req.After)
}

func v3NativeModelRequest(t *testing.T, req Request) []byte {
	t.Helper()
	before, after := v3ModelTexts(t, req)
	// Preserve original embedded JSON tokens, including their internal spacing.
	// Empty bindings are deliberate: no scenario assertions may be inferred.
	return []byte("{\"formatVersion\":\"lifecycle-model-diff-json/v1\",\"before\":" + before + ",\"after\":" + after + ",\"bindings\":[]}\n")
}

func v3AssertModelDiff(t *testing.T, raw []byte, req Request, observation v3Observation) {
	t.Helper()
	var fields map[string]json.RawMessage
	if json.Unmarshal(raw, &fields) != nil {
		t.Fatal("native model report")
	}
	for key, want := range observation.ExpectedModelDiff {
		var a, b any
		if json.Unmarshal(fields[key], &a) != nil || json.Unmarshal(want, &b) != nil || !reflect.DeepEqual(a, b) {
			t.Fatalf("hand-stated %s: got %s want %s", key, fields[key], want)
		}
	}
	var report impact.Report
	if json.Unmarshal(raw, &report) != nil {
		t.Fatal("model report framing")
	}
	before, after := v3ModelTexts(t, req)
	if (before == after) != observation.FullModelTextsEqual || (bytes.Equal(bytes.TrimSpace([]byte(before)), bytes.TrimSpace([]byte(after)))) != observation.EmbeddedModelSourcesEqual {
		t.Fatal("fixture source relation differs from the declared case")
	}
	if report.RequestDigest == nil || *report.RequestDigest != v3SHA(v3NativeModelRequest(t, req)) || report.BeforeSourceDigest == nil || *report.BeforeSourceDigest != v3SHA(bytes.TrimSpace([]byte(before))) || report.AfterSourceDigest == nil || *report.AfterSourceDigest != v3SHA(bytes.TrimSpace([]byte(after))) {
		t.Fatal("native source hashes must identify original embedded JSON values")
	}
	if (report.BeforeDigest != report.AfterDigest) != report.Changed || (*report.BeforeSourceDigest == *report.AfterSourceDigest) != observation.EmbeddedModelSourcesEqual {
		t.Fatal("canonical and source identities conflated")
	}
	if len(raw) == 0 || raw[len(raw)-1] != '\n' {
		t.Fatal("native model report missing final newline")
	}
}

type v3ScenarioEnvelope struct {
	Status                           string
	Reason, ReportSHA256, ReportJSON *string
}

func v3AssertScenario(t *testing.T, raw []byte, observation v3Observation) {
	t.Helper()
	var report struct{ ScenarioDiff v3ScenarioEnvelope }
	if json.Unmarshal(raw, &report) != nil || report.ScenarioDiff.Status != observation.ScenarioStatus || !reflect.DeepEqual(report.ScenarioDiff.Reason, observation.ScenarioReason) {
		t.Fatal("scenario comparison status/reason changed")
	}
	comparison := report.ScenarioDiff
	if comparison.Status == "not-checked" {
		if comparison.ReportJSON != nil || comparison.ReportSHA256 != nil {
			t.Fatal("refused scenario comparison returned report/hash")
		}
	} else if comparison.Reason != nil || comparison.ReportJSON == nil || comparison.ReportSHA256 == nil || *comparison.ReportSHA256 != v3SHA([]byte(*comparison.ReportJSON)) {
		t.Fatal("checked scenario report/hash incomplete")
	}
}

func v3AssertNoExecutionAndOriginalModelHash(t *testing.T, raw []byte, req Request) {
	t.Helper()
	var report Report
	if json.Unmarshal(raw, &report) != nil || report.Mode != "example-only" || report.Authority != "none" || report.ExecutionStatus != "not-run" || report.ReviewStatus != "not-established" {
		t.Fatal("revision authority or execution changed")
	}
	before, after := v3ModelTexts(t, req)
	for i, side := range []SideReport{report.Before, report.After} {
		var first struct{ ProposalReportJSON string }
		var proposal struct {
			Action, Compilation, Replay, Cleanup string
			CandidateStarted                     bool
			Sources                              struct {
				ModelSHA256 string `json:"modelSha256"`
			}
			Preflight struct{ Replay string }
		}
		if json.Unmarshal([]byte(side.FirstDesignReportJSON), &first) != nil || json.Unmarshal([]byte(first.ProposalReportJSON), &proposal) != nil {
			t.Fatal("retained proposal report")
		}
		if proposal.Sources.ModelSHA256 != v3SHA([]byte([]string{before, after}[i])) || proposal.Action != "check" || proposal.CandidateStarted || proposal.Compilation != "not-requested" || proposal.Replay != "not-requested" || proposal.Preflight.Replay != "not-run" || proposal.Cleanup != "not-created" {
			t.Fatal("original full model identity or non-execution evidence changed")
		}
	}
}

// Run only check-only natives to validate the independently declared inputs.
// This runs before production v3 exists; it never runs a candidate or scenario.
func TestV3FrozenNativeInputsAndExpectations(t *testing.T) {
	for _, observation := range v3Observations(t) {
		t.Run(observation.ID, func(t *testing.T) {
			req, oldRequest := v3RequestAsV2(t, observation.Request)
			retained, err := CheckV2(oldRequest)
			if !observation.Valid {
				if !errors.Is(err, ErrInvalid) || retained != nil {
					t.Fatal("malformed original input accepted", err)
				}
				return
			}
			if err != nil {
				t.Fatal("valid v2 input refused", err)
			}
			v3AssertScenario(t, retained, observation)
			v3AssertNoExecutionAndOriginalModelHash(t, retained, req)
			native, err := impact.Execute(v3NativeModelRequest(t, req))
			if err != nil {
				t.Fatal("valid native model comparison refused", err)
			}
			v3AssertModelDiff(t, native, req, observation)
		})
	}
}

func assertFrozenV3Contract(t *testing.T, check func([]byte) ([]byte, error)) {
	t.Helper()
	for _, observation := range v3Observations(t) {
		t.Run(observation.ID, func(t *testing.T) {
			input := []byte(observation.Request)
			original := bytes.Clone(input)
			raw, err := check(input)
			if !bytes.Equal(input, original) {
				t.Fatal("v3 input mutated")
			}
			if !observation.Valid {
				if !errors.Is(err, ErrInvalid) || raw != nil {
					t.Fatal("invalid request returned a partial report", err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			var envelope struct {
				FormatVersion                              string
				ModelDiffReportJSON, ModelDiffReportSHA256 *string
			}
			if json.Unmarshal(raw, &envelope) != nil || envelope.FormatVersion != "lifecycle-revision-check-report/v3" || envelope.ModelDiffReportJSON == nil || envelope.ModelDiffReportSHA256 == nil {
				t.Fatal("missing v3 model report envelope")
			}
			req, oldRequest := v3RequestAsV2(t, observation.Request)
			retained, err := CheckV2(oldRequest)
			if err != nil {
				t.Fatal("retained v2", err)
			}
			var old, all map[string]any
			if json.Unmarshal(retained, &old) != nil || json.Unmarshal(raw, &all) != nil {
				t.Fatal("revision report framing")
			}
			delete(all, "modelDiffReportJSON")
			delete(all, "modelDiffReportSHA256")
			all["formatVersion"] = old["formatVersion"]
			if !reflect.DeepEqual(all, old) {
				t.Fatal("v3 changed v2 observations, native report bytes/hashes or added an undeclared decision")
			}
			v3AssertScenario(t, raw, observation)
			v3AssertNoExecutionAndOriginalModelHash(t, raw, req)
			native, err := impact.Execute(v3NativeModelRequest(t, req))
			if err != nil || string(native) != *envelope.ModelDiffReportJSON || v3SHA(native) != *envelope.ModelDiffReportSHA256 {
				t.Fatal("full native model report bytes/hash not retained", err)
			}
			v3AssertModelDiff(t, []byte(*envelope.ModelDiffReportJSON), req, observation)
			again, err := check(input)
			if err != nil || !bytes.Equal(again, raw) {
				t.Fatal("repeated v3 check changed", err)
			}
		})
	}
	valid := []byte(v3Observations(t)[0].Request)
	badInputs := [][]byte{
		nil, []byte("null"), []byte("{}"), []byte("[]"), {0xff},
		append(bytes.Clone(valid), valid...),
		bytes.Replace(valid, []byte(`"before":`), []byte(`"BEFORE":`), 1),
		bytes.Replace(valid, []byte(`"before":`), []byte(`"extra":null,"before":`), 1),
		bytes.Replace(valid, []byte(`"before":`), []byte(`"formatVersion":"lifecycle-revision-check-request/v3","before":`), 1),
	}
	for _, version := range []string{"lifecycle-revision-check-request/v1", "lifecycle-revision-check-request/v2", "lifecycle-revision-check-request/v99"} {
		badInputs = append(badInputs, bytes.Replace(valid, []byte("lifecycle-revision-check-request/v3"), []byte(version), 1))
	}
	for _, key := range []string{"formatVersion", "before", "after"} {
		var fields map[string]json.RawMessage
		if json.Unmarshal(valid, &fields) != nil {
			t.Fatal("valid wrapper")
		}
		fields[key] = json.RawMessage("null")
		null, _ := json.Marshal(fields)
		badInputs = append(badInputs, null)
		delete(fields, key)
		missing, _ := json.Marshal(fields)
		badInputs = append(badInputs, missing)
	}
	for _, bad := range badInputs {
		if out, err := check(bad); !errors.Is(err, ErrInvalid) || out != nil {
			t.Fatal("strict v3 wrapper accepted or returned partial bytes", err)
		}
	}
	if out, err := check(bytes.Repeat([]byte(" "), MaxRequestBytes+1)); !errors.Is(err, ErrLimit) || out != nil {
		t.Fatal("v3 request limit did not refuse completely", err)
	}
	original := bytes.Clone(valid)
	expected, err := check(valid)
	if err != nil {
		t.Fatal(err)
	}
	var group sync.WaitGroup
	for i := 0; i < 8; i++ {
		group.Add(1)
		go func() {
			defer group.Done()
			out, err := check(valid)
			if err != nil || !bytes.Equal(out, expected) {
				t.Errorf("concurrent v3 result changed: %v", err)
			}
		}()
	}
	group.Wait()
	if !bytes.Equal(valid, original) {
		t.Fatal("concurrent v3 calls mutated shared input")
	}
}
