package scenariotraceimpact_test

// These expectations were stated before production existed. ParcelReview is an
// invented fixture; its supplied observations are fabricated, not execution
// evidence. Native controls validate inputs and retained text, never calculate
// the candidate, link, reason or source-change expectations below.
import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/canonical"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/jsonformat"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/scenariodiff"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/scenariotrace"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/scenariotraceimpact"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/traceimpact"
)

const (
	acceptanceRequestVersion = "lifecycle-scenario-trace-impact-request/v1"
	acceptanceReportVersion  = "lifecycle-scenario-trace-impact-report/v1"
	acceptanceZ              = "z_refuse_then_seal"
	acceptanceA              = "a_seal"
	acceptanceOrphan         = "m_orphan"
)

type acceptanceVectors struct {
	Sources map[string]string `json:"sources"`
	Reports map[string]string `json:"reports"`
}

type acceptanceReason struct {
	Code           string `json:"code"`
	Side           string `json:"side"`
	ScenarioCaseID string `json:"scenarioCaseId"`
}

type acceptanceCandidate struct {
	RequirementID string             `json:"requirementId"`
	BeforePresent bool               `json:"beforePresent"`
	AfterPresent  bool               `json:"afterPresent"`
	BeforeCaseIDs []string           `json:"beforeCaseIds"`
	AfterCaseIDs  []string           `json:"afterCaseIds"`
	Reasons       []acceptanceReason `json:"reasons"`
}

type acceptanceFixture struct {
	Name, Before, After string
	Changes             []string
	Candidates          []acceptanceCandidate
}

func acceptanceLoad(t *testing.T) acceptanceVectors {
	t.Helper()
	raw, err := os.ReadFile("../scenariotrace/testdata/vectors.json")
	if err != nil {
		t.Fatal(err)
	}
	var v acceptanceVectors
	if err := json.Unmarshal(raw, &v); err != nil {
		t.Fatal(err)
	}
	return v
}

func acceptanceJSON(t *testing.T, value any) string {
	t.Helper()
	var b bytes.Buffer
	e := json.NewEncoder(&b)
	e.SetEscapeHTML(false)
	if err := e.Encode(value); err != nil {
		t.Fatal(err)
	}
	return b.String()
}

func acceptanceObject(t *testing.T, raw string) map[string]any {
	t.Helper()
	var value map[string]any
	if err := json.Unmarshal([]byte(raw), &value); err != nil || value == nil {
		t.Fatalf("invalid fixture object: %v", err)
	}
	return value
}

func acceptanceHash(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

func acceptanceSnapshot(t *testing.T, v acceptanceVectors, observation string) map[string]any {
	t.Helper()
	q := map[string]any{"formatVersion": "lifecycle-scenario-trace-request/v1", "scenarioReport": nil}
	for name, raw := range v.Sources {
		q[name] = raw
	}
	if observation != "" {
		q["scenarioReport"] = v.Reports[observation]
	}
	return q
}

// A test author changing a source must also state the matching input binding.
// This only refreshes fixture byte hashes, never an expected comparison result.
func acceptanceSetSource(t *testing.T, q map[string]any, name, raw string) {
	t.Helper()
	q[name] = raw
	if name == "requirements" || name == "catalog" || name == "suite" {
		b := acceptanceObject(t, q["bindings"].(string))
		b[name+"SHA256"] = acceptanceHash(raw)
		q["bindings"] = acceptanceJSON(t, b)
	}
}

func acceptanceEditSource(t *testing.T, q map[string]any, name string, edit func(map[string]any)) {
	t.Helper()
	source := acceptanceObject(t, q[name].(string))
	edit(source)
	acceptanceSetSource(t, q, name, acceptanceJSON(t, source))
}

func acceptanceEditLinks(t *testing.T, q map[string]any, edit func([]any) []any) {
	t.Helper()
	b := acceptanceObject(t, q["bindings"].(string))
	b["links"] = edit(b["links"].([]any))
	q["bindings"] = acceptanceJSON(t, b)
}

func acceptanceLink(requirement, scenario string) map[string]any {
	return map[string]any{"requirementId": requirement, "scenarioCaseId": scenario}
}

func acceptanceCase(suite map[string]any, index int) map[string]any {
	return suite["cases"].([]any)[index].(map[string]any)
}

func acceptanceStep(c map[string]any, index int) map[string]any {
	return c["steps"].([]any)[index].(map[string]any)
}

func acceptanceR(code, side, scenario string) acceptanceReason {
	return acceptanceReason{code, side, scenario}
}

func acceptanceUR1(reasons ...acceptanceReason) acceptanceCandidate {
	return acceptanceCandidate{"UR-001", true, true, []string{acceptanceA, acceptanceZ}, []string{acceptanceA, acceptanceZ}, reasons}
}

func acceptanceUR2(reasons ...acceptanceReason) acceptanceCandidate {
	return acceptanceCandidate{"UR-002", true, true, []string{acceptanceZ}, []string{acceptanceZ}, reasons}
}

func acceptanceFixtures(t *testing.T) []acceptanceFixture {
	t.Helper()
	v := acceptanceLoad(t)
	out := []acceptanceFixture{}
	add := func(name, observation string, edit func(map[string]any), changes []string, candidates ...acceptanceCandidate) {
		before := acceptanceSnapshot(t, v, observation)
		after := acceptanceSnapshot(t, v, observation)
		if edit != nil {
			edit(after)
		}
		out = append(out, acceptanceFixture{name, acceptanceJSON(t, before), acceptanceJSON(t, after), changes, append([]acceptanceCandidate{}, candidates...)})
	}
	native := acceptanceR("native-trace-impact", "both", "")
	order := acceptanceR("scenario-case-order-changed", "both", "")
	suiteFields := acceptanceR("scenario-suite-fields-changed", "both", "")
	obsA := acceptanceR("scenario-observation-changed", "both", acceptanceA)
	obsZ := acceptanceR("scenario-observation-changed", "both", acceptanceZ)
	modifiedAAfter := acceptanceR("scenario-case-modified", "after", acceptanceA)
	modifiedABefore := acceptanceR("scenario-case-modified", "before", acceptanceA)
	modifiedZAfter := acceptanceR("scenario-case-modified", "after", acceptanceZ)
	modifiedZBefore := acceptanceR("scenario-case-modified", "before", acceptanceZ)
	for _, observation := range []string{"", "matched", "early"} {
		add("identical_"+observation, observation, nil, []string{})
	}
	add("binding_whitespace", "", func(q map[string]any) {
		q["bindings"] = q["bindings"].(string) + " \n"
	}, []string{"bindings"})
	add("binding_order", "", func(q map[string]any) {
		acceptanceEditLinks(t, q, func(links []any) []any {
			links[0], links[2] = links[2], links[0]
			return links
		})
	}, []string{"bindings"})
	add("source_whitespace", "", func(q map[string]any) {
		acceptanceSetSource(t, q, "model", " \n"+q["model"].(string)+"\t")
		acceptanceSetSource(t, q, "requirements", q["requirements"].(string)+" \n")
	}, []string{"bindings", "model", "requirements"})
	add("catalog_whitespace_is_native_impact", "", func(q map[string]any) {
		acceptanceSetSource(t, q, "catalog", q["catalog"].(string)+" \n")
	}, []string{"bindings", "catalog"}, acceptanceUR1(native), acceptanceUR2(native))
	// A supplied matched report stays identical while requirement references in
	// the native case views change. Those references are not observation data.
	add("link_remove_and_add_without_observation_change", "matched", func(q map[string]any) {
		acceptanceEditLinks(t, q, func(links []any) []any {
			return []any{links[2], acceptanceLink("UR-003", acceptanceOrphan), links[0]}
		})
	}, []string{"bindings"},
		acceptanceCandidate{"UR-001", true, true, []string{acceptanceA, acceptanceZ}, []string{acceptanceA}, []acceptanceReason{acceptanceR("scenario-link-removed", "before", acceptanceZ)}},
		acceptanceCandidate{"UR-003", true, true, []string{}, []string{acceptanceOrphan}, []acceptanceReason{acceptanceR("scenario-link-added", "after", acceptanceOrphan)}})
	add("requirement_added_with_link", "", func(q map[string]any) {
		acceptanceEditSource(t, q, "requirements", func(d map[string]any) {
			d["requirements"] = append(d["requirements"].([]any), map[string]any{"id": "UR-004", "title": "New outcome", "statement": "Review a newly declared outcome.", "modelLinks": []any{}, "acceptanceLinks": []any{}})
		})
		acceptanceEditLinks(t, q, func(links []any) []any { return append(links, acceptanceLink("UR-004", acceptanceOrphan)) })
	}, []string{"bindings", "requirements"}, acceptanceCandidate{"UR-004", false, true, []string{}, []string{acceptanceOrphan}, []acceptanceReason{native, acceptanceR("scenario-link-added", "after", acceptanceOrphan)}})
	add("requirement_removed_with_link", "", func(q map[string]any) {
		acceptanceEditSource(t, q, "requirements", func(d map[string]any) {
			reqs := d["requirements"].([]any)
			d["requirements"] = []any{reqs[0], reqs[2]}
		})
		acceptanceEditLinks(t, q, func(links []any) []any { return links[1:] })
	}, []string{"bindings", "requirements"}, acceptanceCandidate{"UR-002", true, false, []string{acceptanceZ}, []string{}, []acceptanceReason{native, acceptanceR("scenario-link-removed", "before", acceptanceZ)}})
	add("unlinked_requirement_text", "", func(q map[string]any) {
		acceptanceEditSource(t, q, "requirements", func(d map[string]any) {
			d["requirements"].([]any)[0].(map[string]any)["statement"] = "Review the unlinked declaration again."
		})
	}, []string{"bindings", "requirements"}, acceptanceCandidate{"UR-003", true, true, []string{}, []string{}, []acceptanceReason{native}})
	add("linked_initial_literal", "", func(q map[string]any) {
		acceptanceEditSource(t, q, "suite", func(s map[string]any) {
			acceptanceCase(s, 1)["initial"].(map[string]any)["outcome"] = "Another initial literal"
		})
	}, []string{"bindings", "suite"}, acceptanceUR1(modifiedAAfter, modifiedABefore))
	add("linked_expected_literal", "", func(q map[string]any) {
		acceptanceEditSource(t, q, "suite", func(s map[string]any) {
			acceptanceStep(acceptanceCase(s, 0), 1)["expected"].(map[string]any)["task"].(map[string]any)["outcome"] = "Another expected literal"
		})
	}, []string{"bindings", "suite"}, acceptanceUR1(modifiedZAfter, modifiedZBefore), acceptanceUR2(modifiedZAfter, modifiedZBefore))
	add("unlinked_case_literal", "", func(q map[string]any) {
		acceptanceEditSource(t, q, "suite", func(s map[string]any) {
			acceptanceCase(s, 2)["initial"].(map[string]any)["outcome"] = "Unlinked change"
		})
	}, []string{"bindings", "suite"})
	add("case_order_without_reports", "", func(q map[string]any) {
		acceptanceEditSource(t, q, "suite", func(s map[string]any) {
			cases := s["cases"].([]any)
			cases[0], cases[1] = cases[1], cases[0]
		})
	}, []string{"bindings", "suite"}, acceptanceUR1(order), acceptanceUR2(order))
	add("suite_id", "", func(q map[string]any) {
		acceptanceEditSource(t, q, "suite", func(s map[string]any) { s["id"] = "parcel_cases_revised" })
	}, []string{"bindings", "suite"}, acceptanceUR1(suiteFields), acceptanceUR2(suiteFields))
	add("model_binding_with_state_label_change", "", func(q map[string]any) {
		acceptanceEditSource(t, q, "model", func(m map[string]any) { m["states"].([]any)[1].(map[string]any)["label"] = "Closed parcel" })
		m, err := jsonformat.Decode([]byte(q["model"].(string)))
		if err != nil {
			t.Fatal(err)
		}
		digest, err := canonical.Digest(m)
		if err != nil {
			t.Fatal(err)
		}
		acceptanceEditSource(t, q, "suite", func(s map[string]any) { s["modelDigest"] = digest })
		b := acceptanceObject(t, q["bindings"].(string))
		b["modelDigest"] = digest
		q["bindings"] = acceptanceJSON(t, b)
	}, []string{"bindings", "model", "suite"}, acceptanceUR1(suiteFields), acceptanceUR2(native, suiteFields))
	for _, observation := range []string{"matched", "early"} {
		add("observation_supplied_"+observation, "", func(q map[string]any) { q["scenarioReport"] = v.Reports[observation] }, []string{"scenarioReport"}, acceptanceUR1(obsA, obsZ), acceptanceUR2(obsZ))
	}
	add("observation_removed", "matched", func(q map[string]any) { q["scenarioReport"] = nil }, []string{"scenarioReport"}, acceptanceUR1(obsA, obsZ), acceptanceUR2(obsZ))
	add("matched_to_early_mismatch", "matched", func(q map[string]any) { q["scenarioReport"] = v.Reports["early"] }, []string{"scenarioReport"}, acceptanceUR1(obsA, obsZ), acceptanceUR2(obsZ))
	add("early_to_late_mismatch", "early", func(q map[string]any) { q["scenarioReport"] = v.Reports["late"] }, []string{"scenarioReport"}, acceptanceUR1(obsA, obsZ), acceptanceUR2(obsZ))
	add("different_actual_payload_same_observation", "early", func(q map[string]any) {
		r := acceptanceObject(t, q["scenarioReport"].(string))
		r["entries"].([]any)[1].(map[string]any)["actual"].(map[string]any)["task"].(map[string]any)["outcome"] = "A different fabricated mismatch"
		q["scenarioReport"] = acceptanceJSON(t, r)
	}, []string{"scenarioReport"})
	add("report_formatting_only", "matched", func(q map[string]any) { q["scenarioReport"] = " \n" + q["scenarioReport"].(string) }, []string{"scenarioReport"})
	// Even absent reports expose ordered step IDs with null positions.
	add("step_order_without_reports", "", func(q map[string]any) {
		acceptanceEditSource(t, q, "suite", func(s map[string]any) {
			steps := acceptanceCase(s, 0)["steps"].([]any)
			steps[0], steps[1] = steps[1], steps[0]
		})
	}, []string{"bindings", "suite"}, acceptanceUR1(modifiedZAfter, modifiedZBefore, obsZ), acceptanceUR2(modifiedZAfter, modifiedZBefore, obsZ))
	add("step_rename_projects_through_each_sides_links", "", func(q map[string]any) {
		acceptanceEditSource(t, q, "suite", func(s map[string]any) { acceptanceStep(acceptanceCase(s, 0), 0)["id"] = "renamed_refusal" })
		acceptanceEditLinks(t, q, func(links []any) []any { return []any{acceptanceLink("UR-003", acceptanceZ), links[0], links[2]} })
	}, []string{"bindings", "suite"},
		acceptanceCandidate{"UR-001", true, true, []string{acceptanceA, acceptanceZ}, []string{acceptanceA}, []acceptanceReason{modifiedZBefore, acceptanceR("scenario-link-removed", "before", acceptanceZ), obsZ}},
		acceptanceUR2(modifiedZAfter, modifiedZBefore, obsZ),
		acceptanceCandidate{"UR-003", true, true, []string{}, []string{acceptanceZ}, []acceptanceReason{modifiedZAfter, acceptanceR("scenario-link-added", "after", acceptanceZ), obsZ}})
	add("case_added", "", func(q map[string]any) {
		acceptanceEditSource(t, q, "suite", func(s map[string]any) {
			c := acceptanceObject(t, acceptanceJSON(t, acceptanceCase(s, 1)))
			c["id"] = "b_new"
			acceptanceStep(c, 0)["id"] = "new_seal"
			s["cases"] = append(s["cases"].([]any), c)
		})
		acceptanceEditLinks(t, q, func(links []any) []any { return append(links, acceptanceLink("UR-003", "b_new")) })
	}, []string{"bindings", "suite"}, acceptanceUR1(order), acceptanceUR2(order),
		acceptanceCandidate{"UR-003", true, true, []string{}, []string{"b_new"}, []acceptanceReason{acceptanceR("scenario-case-added", "after", "b_new"), order, acceptanceR("scenario-link-added", "after", "b_new")}})
	add("case_removed", "", func(q map[string]any) {
		acceptanceEditSource(t, q, "suite", func(s map[string]any) {
			cases := s["cases"].([]any)
			s["cases"] = []any{cases[0], cases[2]}
		})
		acceptanceEditLinks(t, q, func(links []any) []any { return links[:2] })
	}, []string{"bindings", "suite"},
		acceptanceCandidate{"UR-001", true, true, []string{acceptanceA, acceptanceZ}, []string{acceptanceZ}, []acceptanceReason{order, acceptanceR("scenario-case-removed", "before", acceptanceA), acceptanceR("scenario-link-removed", "before", acceptanceA)}}, acceptanceUR2(order))
	// Whitespace in the native request itself is distinct from embedded sources.
	q := acceptanceJSON(t, acceptanceSnapshot(t, v, ""))
	out = append(out, acceptanceFixture{"snapshot_wrapper_whitespace", q, " \n" + q + "\t", []string{}, []acceptanceCandidate{}})
	return out
}

// Token-preserving composition is used only for independent native controls.
func acceptanceScenarioRequest(t *testing.T, before, after string) string {
	t.Helper()
	b, a := acceptanceObject(t, before), acceptanceObject(t, after)
	token := func(q map[string]any, name string) string { return strings.Trim(q[name].(string), " \t\r\n") }
	return `{"formatVersion":"lifecycle-scenario-diff-json/v1","beforeModel":` + token(b, "model") + `,"beforeSuite":` + token(b, "suite") + `,"afterModel":` + token(a, "model") + `,"afterSuite":` + token(a, "suite") + "}\n"
}

func acceptanceTraceRequest(t *testing.T, before, after string) string {
	t.Helper()
	side := func(raw string) map[string]any {
		q := acceptanceObject(t, raw)
		return map[string]any{"model": q["model"], "catalog": q["catalog"], "requirements": q["requirements"]}
	}
	return acceptanceJSON(t, map[string]any{"formatVersion": "lifecycle-trace-impact-request/v1", "before": side(before), "after": side(after)})
}

func TestAcceptanceNativeFixtureControls(t *testing.T) {
	for _, tc := range acceptanceFixtures(t) {
		t.Run(tc.Name, func(t *testing.T) {
			for _, raw := range []string{tc.Before, tc.After} {
				if _, err := scenariotrace.Check([]byte(raw)); err != nil {
					t.Fatalf("invalid hand-constructed snapshot: %v", err)
				}
			}
			if _, err := scenariodiff.Execute([]byte(acceptanceScenarioRequest(t, tc.Before, tc.After))); err != nil {
				t.Fatalf("scenario comparison admission: %v", err)
			}
			if _, err := traceimpact.Execute([]byte(acceptanceTraceRequest(t, tc.Before, tc.After))); err != nil {
				t.Fatalf("trace comparison admission: %v", err)
			}
		})
	}
}

// Everything above this marker is usable by a native-only setup probe before
// production exists. Everything below exercises the new public API.

func acceptanceKeys(t *testing.T, value map[string]any, fields string) {
	t.Helper()
	names := strings.Fields(fields)
	if len(value) != len(names) {
		t.Fatalf("unexpected object keys: %v", value)
	}
	for _, name := range names {
		if _, ok := value[name]; !ok {
			t.Fatalf("missing key %q", name)
		}
	}
}

func acceptanceEqual(t *testing.T, got, want any) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %s; want %s", acceptanceJSON(t, got), acceptanceJSON(t, want))
	}
}

func acceptanceOuter(t *testing.T, before, after string) []byte {
	t.Helper()
	return []byte(acceptanceJSON(t, map[string]any{"formatVersion": acceptanceRequestVersion, "before": before, "after": after}))
}

func acceptanceCheck(t *testing.T, raw []byte) ([]byte, map[string]any) {
	t.Helper()
	original := append([]byte(nil), raw...)
	output, err := scenariotraceimpact.Check(raw)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(raw, original) {
		t.Fatal("Check mutated input")
	}
	if !bytes.HasSuffix(output, []byte{'\n'}) || bytes.HasSuffix(output, []byte{'\n', '\n'}) {
		t.Fatal("output must end with one LF")
	}
	report := acceptanceObject(t, string(output))
	acceptanceKeys(t, report, "formatVersion mode authority executionStatus reviewStatus requestSHA256 before after scenarioDiff traceImpact requestSourceChanged sourceChanges reviewRequired candidates")
	for field, want := range map[string]string{"formatVersion": acceptanceReportVersion, "mode": "example-only", "authority": "none", "executionStatus": "not-established", "reviewStatus": "not-established", "requestSHA256": acceptanceHash(string(raw))} {
		acceptanceEqual(t, report[field], want)
	}
	return output, report
}

func acceptanceRetained(t *testing.T, report map[string]any, field string) map[string]any {
	t.Helper()
	r, ok := report[field].(map[string]any)
	if !ok {
		t.Fatalf("%s is not a retained record", field)
	}
	acceptanceKeys(t, r, "requestSHA256 reportJSON reportSHA256")
	text, ok := r["reportJSON"].(string)
	if !ok {
		t.Fatalf("%s.reportJSON is not text", field)
	}
	acceptanceEqual(t, r["reportSHA256"], acceptanceHash(text))
	requestHash, ok := r["requestSHA256"].(string)
	if !ok || len(requestHash) != 64 || strings.ToLower(requestHash) != requestHash {
		t.Fatalf("%s request hash is not lowercase SHA-256", field)
	}
	if _, err := hex.DecodeString(requestHash); err != nil {
		t.Fatal(err)
	}
	return r
}

func acceptanceVerifyNativeRetention(t *testing.T, report map[string]any, before, after string) {
	t.Helper()
	for field, raw := range map[string]string{"before": before, "after": after} {
		native, err := scenariotrace.Check([]byte(raw))
		if err != nil {
			t.Fatal(err)
		}
		r := acceptanceRetained(t, report, field)
		acceptanceEqual(t, r["requestSHA256"], acceptanceHash(raw))
		acceptanceEqual(t, r["reportJSON"], string(native))
	}
	r := acceptanceRetained(t, report, "traceImpact")
	nativeTrace, err := traceimpact.Execute([]byte(acceptanceTraceRequest(t, before, after)))
	if err != nil {
		t.Fatal(err)
	}
	acceptanceEqual(t, r["reportJSON"], string(nativeTrace))
	r = acceptanceRetained(t, report, "scenarioDiff")
	request := acceptanceScenarioRequest(t, before, after)
	nativeScenario, err := scenariodiff.Execute([]byte(request))
	if err != nil {
		t.Fatal(err)
	}
	// Wrapper key order is not prescribed. Only that wrapper's request hash can
	// differ; every other byte must remain the exact native output encoding.
	want := strings.Replace(string(nativeScenario), `"request":"`+acceptanceHash(request)+`"`, `"request":"`+r["requestSHA256"].(string)+`"`, 1)
	acceptanceEqual(t, r["reportJSON"], want)
	sources := acceptanceObject(t, r["reportJSON"].(string))["sources"].(map[string]any)
	acceptanceEqual(t, sources["request"], r["requestSHA256"])
	for side, raw := range map[string]string{"before": before, "after": after} {
		q := acceptanceObject(t, raw)
		for source, suffix := range map[string]string{"model": "Model", "suite": "Suite"} {
			acceptanceEqual(t, sources[side+suffix], acceptanceHash(strings.Trim(q[source].(string), " \t\r\n")))
		}
	}
}

func TestAcceptanceHandStatedChanges(t *testing.T) {
	for _, tc := range acceptanceFixtures(t) {
		t.Run(tc.Name, func(t *testing.T) {
			_, report := acceptanceCheck(t, acceptanceOuter(t, tc.Before, tc.After))
			acceptanceEqual(t, report["requestSourceChanged"], tc.Before != tc.After)
			acceptanceEqual(t, report["reviewRequired"], tc.Before != tc.After)
			var changes []string
			if err := json.Unmarshal([]byte(acceptanceJSON(t, report["sourceChanges"])), &changes); err != nil {
				t.Fatal(err)
			}
			acceptanceEqual(t, changes, tc.Changes)
			var candidates []acceptanceCandidate
			if err := json.Unmarshal([]byte(acceptanceJSON(t, report["candidates"])), &candidates); err != nil {
				t.Fatal(err)
			}
			acceptanceEqual(t, candidates, tc.Candidates)
			for _, item := range report["candidates"].([]any) {
				c := item.(map[string]any)
				acceptanceKeys(t, c, "requirementId beforePresent afterPresent beforeCaseIds afterCaseIds reasons")
				for _, item := range c["reasons"].([]any) {
					acceptanceKeys(t, item.(map[string]any), "code side scenarioCaseId")
				}
			}
			acceptanceVerifyNativeRetention(t, report, tc.Before, tc.After)
		})
	}
}

func TestAcceptanceLiteralTokensAndWrapperFormatting(t *testing.T) {
	v := acceptanceLoad(t)
	q := acceptanceSnapshot(t, v, "")
	// Preserve internal whitespace and an escaped spelling, not just decoded
	// string values. A RawMessage compactor would change these token hashes.
	for _, field := range []string{"model", "suite"} {
		raw := strings.ReplaceAll(q[field].(string), ":", " : ")
		raw = strings.ReplaceAll(raw, ",", ",\n\t")
		if field == "model" {
			raw = strings.Replace(raw, "ParcelReview", `\u0050arcelReview`, 1)
		} else {
			raw = strings.ReplaceAll(raw, "café", `caf\u00e9`)
		}
		acceptanceSetSource(t, q, field, " \r\n"+raw+"\t ")
	}
	snapshot := acceptanceJSON(t, q)
	raw := acceptanceOuter(t, snapshot, snapshot)
	firstBytes, first := acceptanceCheck(t, raw)
	acceptanceVerifyNativeRetention(t, first, snapshot, snapshot)
	acceptanceEqual(t, first["requestSourceChanged"], false)
	acceptanceEqual(t, first["reviewRequired"], false)
	acceptanceEqual(t, first["sourceChanges"], []any{})
	acceptanceEqual(t, first["candidates"], []any{})
	_, second := acceptanceCheck(t, append([]byte(" \n"), raw...))
	if first["requestSHA256"] == second["requestSHA256"] {
		t.Fatal("outer formatting must change its own hash")
	}
	delete(first, "requestSHA256")
	delete(second, "requestSHA256")
	acceptanceEqual(t, first, second)
	// Returned buffers are owned, and repeated calls cannot depend on mutations
	// made to another call's output.
	again, _ := acceptanceCheck(t, raw)
	if !bytes.Equal(firstBytes, again) {
		t.Fatal("identical input produced different output")
	}
	firstBytes[0] = '!'
	third, _ := acceptanceCheck(t, raw)
	if !bytes.Equal(again, third) {
		t.Fatal("returned buffers share state")
	}
}

func acceptanceRefuse(t *testing.T, raw []byte, want error) {
	t.Helper()
	original := append([]byte(nil), raw...)
	out, err := scenariotraceimpact.Check(raw)
	if err != want || out != nil {
		t.Fatalf("got output length %d, error %v; want nil output and %v", len(out), err, want)
	}
	if !bytes.Equal(original, raw) {
		t.Fatal("refusal mutated input")
	}
	if strings.Contains(err.Error(), "invented-secret") || strings.Contains(err.Error(), "parcel-7") {
		t.Fatal("diagnostic echoed input")
	}
}

func TestAcceptanceStrictRefusals(t *testing.T) {
	v := acceptanceLoad(t)
	snapshot := acceptanceJSON(t, acceptanceSnapshot(t, v, ""))
	for _, field := range []string{"formatVersion", "before", "after"} {
		for _, kind := range []string{"missing", "null", "case", "duplicate", "wrong_type"} {
			t.Run(field+"_"+kind, func(t *testing.T) {
				q := map[string]any{"formatVersion": acceptanceRequestVersion, "before": snapshot, "after": snapshot}
				switch kind {
				case "missing":
					delete(q, field)
				case "null":
					q[field] = nil
				case "case":
					q[strings.ToUpper(field)] = q[field]
					delete(q, field)
				case "wrong_type":
					q[field] = map[string]any{}
				}
				raw := strings.TrimSpace(acceptanceJSON(t, q))
				if kind == "duplicate" {
					raw = raw[:len(raw)-1] + `,"` + field + `":null}`
				}
				acceptanceRefuse(t, []byte(raw), scenariotraceimpact.ErrInvalid)
			})
		}
	}
	raw := string(acceptanceOuter(t, snapshot, snapshot))
	for name, bad := range map[string]string{
		"unknown":      strings.TrimSpace(raw)[:len(strings.TrimSpace(raw))-1] + `,"invented-secret":true}`,
		"version":      strings.Replace(raw, acceptanceRequestVersion, "v2", 1),
		"unicode":      strings.Replace(raw, acceptanceRequestVersion, `\ud800`, 1),
		"invalid_utf8": strings.Replace(raw, acceptanceRequestVersion, string([]byte{0xff}), 1),
		"trailing":     raw + "{}",
		"empty":        "",
		"null":         "null",
		"array":        "[]",
	} {
		t.Run(name, func(t *testing.T) { acceptanceRefuse(t, []byte(bad), scenariotraceimpact.ErrInvalid) })
	}
	for _, side := range []string{"before", "after"} {
		for _, name := range []string{"modelDigest", "requirementsSHA256", "catalogSHA256", "suiteSHA256"} {
			t.Run(side+"_stale_"+name, func(t *testing.T) {
				q := acceptanceSnapshot(t, v, "")
				b := acceptanceObject(t, q["bindings"].(string))
				b[name] = strings.Repeat("0", 64)
				q["bindings"] = acceptanceJSON(t, b)
				outer := map[string]any{"formatVersion": acceptanceRequestVersion, "before": snapshot, "after": snapshot}
				outer[side] = acceptanceJSON(t, q)
				acceptanceRefuse(t, []byte(acceptanceJSON(t, outer)), scenariotraceimpact.ErrInvalid)
			})
		}
	}
	for _, kind := range []string{"unknown_case", "duplicate_link", "stale_source", "native_source", "native_wrapper", "observation_binding", "rebound_expectation"} {
		t.Run(kind, func(t *testing.T) {
			q := acceptanceSnapshot(t, v, "matched")
			switch kind {
			case "unknown_case":
				acceptanceEditLinks(t, q, func(links []any) []any { return append(links, acceptanceLink("UR-003", "catalog-live")) })
			case "duplicate_link":
				acceptanceEditLinks(t, q, func(links []any) []any { return append(links, links[0]) })
			case "stale_source":
				q["suite"] = q["suite"].(string) + " "
			case "native_source":
				acceptanceSetSource(t, q, "requirements", "{}")
			case "native_wrapper":
				delete(q, "scenarioReport")
			case "observation_binding":
				r := acceptanceObject(t, q["scenarioReport"].(string))
				r["sourceDigest"] = strings.Repeat("0", 64)
				q["scenarioReport"] = acceptanceJSON(t, r)
			case "rebound_expectation":
				r := acceptanceObject(t, v.Reports["late"])
				entry := r["entries"].([]any)[1].(map[string]any)
				for _, field := range []string{"expected", "actual"} {
					entry[field].(map[string]any)["task"].(map[string]any)["outcome"] = "invented-secret"
				}
				q["scenarioReport"] = acceptanceJSON(t, r)
			}
			acceptanceRefuse(t, acceptanceOuter(t, snapshot, acceptanceJSON(t, q)), scenariotraceimpact.ErrInvalid)
		})
	}
}

func TestAcceptanceInputLimits(t *testing.T) {
	v := acceptanceLoad(t)
	snapshot := acceptanceJSON(t, acceptanceSnapshot(t, v, ""))
	t.Run("outer", func(t *testing.T) {
		acceptanceRefuse(t, []byte(strings.Repeat(" ", (64<<20)+1)), scenariotraceimpact.ErrLimit)
	})
	t.Run("original_snapshot", func(t *testing.T) {
		acceptanceRefuse(t, acceptanceOuter(t, snapshot, strings.Repeat(" ", (32<<20)+1)), scenariotraceimpact.ErrLimit)
	})
	t.Run("native_source", func(t *testing.T) {
		q := acceptanceSnapshot(t, v, "")
		q["suite"] = strings.Repeat(" ", (1<<20)+1)
		acceptanceRefuse(t, acceptanceOuter(t, snapshot, acceptanceJSON(t, q)), scenariotraceimpact.ErrLimit)
	})
	t.Run("native_link_count", func(t *testing.T) {
		q := acceptanceSnapshot(t, v, "")
		acceptanceEditLinks(t, q, func([]any) []any {
			links := make([]any, 4097)
			for i := range links {
				links[i] = acceptanceLink("UR-001", acceptanceA)
			}
			return links
		})
		acceptanceRefuse(t, acceptanceOuter(t, snapshot, acceptanceJSON(t, q)), scenariotraceimpact.ErrLimit)
	})
}
