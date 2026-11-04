package scenariotrace

// Frozen independently before Check existed. The fixture observations are
// fabricated literals, not execution evidence. Call runFrozenContract from the
// implementation tests with Check and a classifier returning invalid/binding/limit.
import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"

	lifecycle "github.com/spik3r/clanker-foundry/modeling/lifecycle/generated"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/scenarios"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/traceability"
)

type frozenVectors struct {
	Sources    map[string]string `json:"sources"`
	Identities map[string]string `json:"identities"`
	Reports    map[string]string `json:"reports"`
	Expected   map[string]any    `json:"expected"`
}

func frozenLoad(t *testing.T) frozenVectors {
	t.Helper()
	b, e := os.ReadFile("testdata/vectors.json")
	if e != nil {
		t.Fatal(e)
	}
	var v frozenVectors
	if e = json.Unmarshal(b, &v); e != nil {
		t.Fatal(e)
	}
	return v
}
func frozenHash(s string) string { h := sha256.Sum256([]byte(s)); return hex.EncodeToString(h[:]) }
func frozenObject(t *testing.T, s string) map[string]any {
	t.Helper()
	var x map[string]any
	if e := json.Unmarshal([]byte(s), &x); e != nil {
		t.Fatal(e)
	}
	return x
}
func frozenJSON(t *testing.T, x any) string {
	t.Helper()
	b, e := json.Marshal(x)
	if e != nil {
		t.Fatal(e)
	}
	return string(b)
}
func frozenEqual(t *testing.T, got, want any) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %s; want %s", frozenJSON(t, got), frozenJSON(t, want))
	}
}
func frozenKeys(t *testing.T, x map[string]any, names string) {
	t.Helper()
	w := strings.Fields(names)
	if len(x) != len(w) {
		t.Fatalf("unexpected keys: %v", x)
	}
	for _, k := range w {
		if _, ok := x[k]; !ok {
			t.Fatalf("missing %s", k)
		}
	}
}
func frozenRunner(t *testing.T, v frozenVectors) scenarios.Runner {
	t.Helper()
	r, e := scenarios.NewRunner(scenarios.Binding{ProfileVersion: "lifecycle/v1", ModelDigest: v.Identities["modelDigest"]}, func(lifecycle.Task, lifecycle.Command) lifecycle.Result {
		t.Fatal("validation invoked reducer")
		return lifecycle.Result{}
	})
	if e != nil {
		t.Fatal(e)
	}
	return r
}

// These controls use only pre-existing native APIs. They cannot invoke Check.
func TestFrozenNativeControls(t *testing.T) {
	v := frozenLoad(t)
	for k, s := range v.Sources {
		if frozenHash(s) != v.Identities[k+"SHA256"] {
			t.Fatalf("raw %s hash", k)
		}
	}
	tr, e := traceability.Build([]byte(v.Sources["model"]), []byte(v.Sources["catalog"]), []byte(v.Sources["requirements"]))
	if e != nil {
		t.Fatal(e)
	}
	if tr.Inputs.ModelDigest != v.Identities["modelDigest"] || tr.Inputs.RequirementsDigest != v.Identities["requirementsDigest"] {
		t.Fatal("independent canonical hash disagrees")
	}
	if tr.Requirements[0].Status != "incomplete" || !reflect.DeepEqual(tr.Requirements[0].Gaps, []string{"no-executable-acceptance"}) || tr.Requirements[0].AcceptanceLinks[0].ExecutionStatus != "pending" {
		t.Fatal("pending trace control")
	}

	c := frozenObject(t, v.Sources["catalog"])
	c["refusals"] = nil
	c["oracleOwnedPayload"] = map[string]any{"opaque": true}
	if _, e := traceability.Build([]byte(v.Sources["model"]), []byte(frozenJSON(t, c)), []byte(v.Sources["requirements"])); e != nil {
		t.Fatalf("opaque catalog control: %v", e)
	}
	r := frozenRunner(t, v)
	s, e := r.Decode([]byte(v.Sources["suite"]))
	if e != nil {
		t.Fatal(e)
	}
	d, e := r.Digest(s)
	if e != nil || d != v.Identities["scenarioDigest"] {
		t.Fatalf("scenario digest %s: %v", d, e)
	}
	for name, raw := range v.Reports {
		p, e := r.DecodeReport([]byte(raw))
		if e != nil {
			t.Fatalf("%s: %v", name, e)
		}
		if name == "late" && (p.CasesCompleted != 1 || p.StepsExecuted != 3 || p.Matched) {
			t.Fatal("late fail-fast control")
		}
	}
}

func runFrozenContract(t *testing.T, check func([]byte) ([]byte, error), classify func(error) string) {
	v := frozenLoad(t)
	request := func(report any) map[string]any {
		q := map[string]any{"formatVersion": "lifecycle-scenario-trace-request/v1", "scenarioReport": report}
		for k, s := range v.Sources {
			q[k] = s
		}
		return q
	}
	call := func(t *testing.T, q map[string]any) map[string]any {
		t.Helper()
		raw := []byte(frozenJSON(t, q))
		before := append([]byte(nil), raw...)
		b, e := check(raw)
		if e != nil {
			t.Fatal(e)
		}
		if !bytes.Equal(raw, before) {
			t.Fatal("input mutated")
		}
		b2, e := check(raw)
		if e != nil || !bytes.Equal(b, b2) {
			t.Fatal("not byte deterministic")
		}
		return frozenObject(t, string(b))
	}
	verify := func(t *testing.T, q map[string]any, out map[string]any, name string) {
		t.Helper()
		frozenKeys(t, out, "formatVersion mode authority executionStatus reviewStatus reportStatus inputs traceReport scenarioReport links requirements cases unlinkedRequirementIds unlinkedScenarioCaseIds")
		for k, w := range map[string]any{"formatVersion": "lifecycle-scenario-trace-report/v1", "mode": "example-only", "authority": "none", "executionStatus": "not-established", "reviewStatus": "not-established"} {
			frozenEqual(t, out[k], w)
		}
		inputs := out["inputs"].(map[string]any)
		frozenKeys(t, inputs, "modelSHA256 modelDigest requirementsSHA256 requirementsDigest catalogSHA256 suiteSHA256 scenarioDigest bindingsSHA256")
		for _, k := range []string{"model", "requirements", "catalog", "suite", "bindings"} {
			frozenEqual(t, inputs[k+"SHA256"], frozenHash(q[k].(string)))
		}
		for _, k := range []string{"modelDigest", "requirementsDigest", "scenarioDigest"} {
			frozenEqual(t, inputs[k], v.Identities[k])
		}
		tr, e := traceability.Build([]byte(q["model"].(string)), []byte(q["catalog"].(string)), []byte(q["requirements"].(string)))
		if e != nil {
			t.Fatal(e)
		}
		tb, e := tr.JSON()
		if e != nil {
			t.Fatal(e)
		}
		frozenEqual(t, out["traceReport"], map[string]any{"text": string(tb), "sha256": frozenHash(string(tb))})
		rp := map[string]any{"text": nil, "sha256": nil}
		rs := "not-supplied"
		if q["scenarioReport"] != nil {
			rp["text"] = q["scenarioReport"]
			rp["sha256"] = frozenHash(q["scenarioReport"].(string))
			rs = "inspected-for-consistency"
		}
		frozenEqual(t, out["scenarioReport"], rp)
		frozenEqual(t, out["reportStatus"], rs)
		for k, w := range v.Expected {
			frozenEqual(t, out[k], w)
		}
		cases := out["cases"].([]any)
		if len(cases) != 3 {
			t.Fatal("case count")
		}
		ids := []string{"z_refuse_then_seal", "a_seal", "m_orphan"}
		steps := [][]string{{"refuse_actor", "seal_after_refusal"}, {"seal_direct"}, {"orphan_seal"}}
		reqs := [][]string{{"UR-001", "UR-002"}, {"UR-001"}, {}}
		statuses := map[string][]string{"none": {"not-supplied", "not-supplied", "not-supplied"}, "matched": {"matched", "matched", "matched"}, "early": {"mismatched", "not-observed", "not-observed"}, "late": {"matched", "mismatched", "not-observed"}}[name]
		count := map[string]int{"none": 0, "matched": 4, "early": 2, "late": 3}[name]
		position := 0
		for i, x := range cases {
			c := x.(map[string]any)
			frozenKeys(t, c, "scenarioCaseId requirementIds status steps")
			frozenEqual(t, c["scenarioCaseId"], ids[i])
			frozenEqual(t, c["requirementIds"], frozenObject(t, `{"v":`+frozenJSON(t, reqs[i])+`}`)["v"])
			frozenEqual(t, c["status"], statuses[i])
			ss := c["steps"].([]any)
			if len(ss) != len(steps[i]) {
				t.Fatal("step count")
			}
			for j, s := range ss {
				st := s.(map[string]any)
				frozenKeys(t, st, "stepId entryPosition")
				frozenEqual(t, st["stepId"], steps[i][j])
				var p any
				if position < count {
					p = float64(position)
				}
				frozenEqual(t, st["entryPosition"], p)
				position++
			}
		}
	}
	for _, name := range []string{"none", "matched", "early", "late"} {
		t.Run(name, func(t *testing.T) {
			var report any
			if name != "none" {
				report = v.Reports[name]
			}
			q := request(report)
			verify(t, q, call(t, q), name)
		})
	}
	t.Run("source_whitespace", func(t *testing.T) {
		q := request(nil)
		b := frozenObject(t, q["bindings"].(string))
		for _, k := range []string{"model", "requirements", "catalog", "suite"} {
			q[k] = q[k].(string) + " \n"
			if k != "model" {
				b[k+"SHA256"] = frozenHash(q[k].(string))
			}
		}
		q["bindings"] = frozenJSON(t, b)
		verify(t, q, call(t, q), "none")
	})
	t.Run("mapping_order", func(t *testing.T) {
		q := request(nil)
		b := frozenObject(t, q["bindings"].(string))
		ls := b["links"].([]any)
		ls[0], ls[2] = ls[2], ls[0]
		q["bindings"] = frozenJSON(t, b)
		verify(t, q, call(t, q), "none")
	})
	t.Run("empty_links", func(t *testing.T) {
		q := request(nil)
		b := frozenObject(t, q["bindings"].(string))
		b["links"] = []any{}
		q["bindings"] = frozenJSON(t, b)
		o := call(t, q)
		frozenEqual(t, o["links"], []any{})
		frozenEqual(t, o["unlinkedRequirementIds"], []any{"UR-001", "UR-002", "UR-003"})
		frozenEqual(t, o["unlinkedScenarioCaseIds"], []any{"a_seal", "m_orphan", "z_refuse_then_seal"})
	})

	t.Run("native_catalog_opaque_fields", func(t *testing.T) {
		q := request(nil)
		c := frozenObject(t, q["catalog"].(string))
		c["refusals"] = nil
		c["oracleOwnedPayload"] = map[string]any{"opaque": true}
		q["catalog"] = frozenJSON(t, c)
		b := frozenObject(t, q["bindings"].(string))
		b["catalogSHA256"] = frozenHash(q["catalog"].(string))
		q["bindings"] = frozenJSON(t, b)
		verify(t, q, call(t, q), "none")
	})
	refuse := func(t *testing.T, raw string, want string) {
		t.Helper()
		b, e := check([]byte(raw))
		if e == nil || len(b) != 0 {
			t.Fatal("refusal returned result")
		}
		if classify(e) != want {
			t.Fatalf("class %s; want %s", classify(e), want)
		}
		if strings.Contains(e.Error(), "parcel-7") || strings.Contains(e.Error(), "invented-secret") {
			t.Fatal("input echoed")
		}
	}

	for _, field := range []string{"formatVersion", "modelDigest", "requirementsSHA256", "catalogSHA256", "suiteSHA256", "links"} {
		for _, kind := range []string{"missing", "null", "case", "duplicate"} {
			t.Run("bindings_"+field+"_"+kind, func(t *testing.T) {
				q := request(nil)
				b := frozenObject(t, q["bindings"].(string))
				raw := ""
				switch kind {
				case "missing":
					delete(b, field)
				case "null":
					b[field] = nil
				case "case":
					b[strings.ToUpper(field)] = b[field]
					delete(b, field)
				case "duplicate":
					raw = frozenJSON(t, b)
					raw = raw[:len(raw)-1] + `,"` + field + `":null}`
				}
				if raw == "" {
					raw = frozenJSON(t, b)
				}
				q["bindings"] = raw
				refuse(t, frozenJSON(t, q), "invalid")
			})
		}
	}
	for _, kind := range []string{"uppercase_digest", "unknown", "trailing", "link_unknown", "link_null", "link_missing", "wrong_version"} {
		t.Run("bindings_"+kind, func(t *testing.T) {
			q := request(nil)
			b := frozenObject(t, q["bindings"].(string))
			switch kind {
			case "uppercase_digest":
				b["modelDigest"] = strings.ToUpper(b["modelDigest"].(string))
			case "unknown":
				b["extra"] = true
			case "link_unknown":
				b["links"].([]any)[0].(map[string]any)["extra"] = true
			case "link_null":
				b["links"].([]any)[0] = nil
			case "link_missing":
				delete(b["links"].([]any)[0].(map[string]any), "requirementId")
			case "wrong_version":
				b["formatVersion"] = "v2"
			}
			q["bindings"] = frozenJSON(t, b)
			if kind == "trailing" {
				q["bindings"] = q["bindings"].(string) + " {}"
			}
			refuse(t, frozenJSON(t, q), "invalid")
		})
	}
	for field, limit := range map[string]int{"model": 1 << 20, "requirements": 1 << 20, "catalog": 4 << 20, "suite": 1 << 20, "bindings": 512 << 10, "scenarioReport": 8 << 20} {
		t.Run("limit_"+field, func(t *testing.T) {
			q := request(nil)
			q[field] = strings.Repeat(" ", limit+1)
			refuse(t, frozenJSON(t, q), "limit")
		})
	}
	t.Run("limit_wrapper", func(t *testing.T) { refuse(t, strings.Repeat(" ", (32<<20)+1), "limit") })
	for _, key := range []string{"modelDigest", "requirementsSHA256", "catalogSHA256", "suiteSHA256"} {
		t.Run("stale_"+key, func(t *testing.T) {
			q := request(nil)
			b := frozenObject(t, q["bindings"].(string))
			b[key] = strings.Repeat("0", 64)
			q["bindings"] = frozenJSON(t, b)
			refuse(t, frozenJSON(t, q), "binding")
		})
	}
	for _, kind := range []string{"duplicate", "unknown_requirement", "unknown_case"} {
		t.Run(kind, func(t *testing.T) {
			q := request(nil)
			b := frozenObject(t, q["bindings"].(string))
			ls := b["links"].([]any)
			switch kind {
			case "duplicate":
				ls = append(ls, ls[0])
			case "unknown_requirement":
				ls[0].(map[string]any)["requirementId"] = "UR-999"
			case "unknown_case":
				ls[0].(map[string]any)["scenarioCaseId"] = "catalog-live"
			}
			b["links"] = ls
			q["bindings"] = frozenJSON(t, b)
			refuse(t, frozenJSON(t, q), "invalid")
		})
	}

	for _, field := range []string{"model", "requirements", "catalog", "suite"} {
		t.Run("native_invalid_"+field, func(t *testing.T) {
			q := request(nil)
			q[field] = "{}"
			b := frozenObject(t, q["bindings"].(string))
			if field != "model" {
				b[field+"SHA256"] = frozenHash("{}")
			}
			q["bindings"] = frozenJSON(t, b)
			refuse(t, frozenJSON(t, q), "invalid")
		})
	}
	t.Run("limit_mapping_count", func(t *testing.T) {
		q := request(nil)
		b := frozenObject(t, q["bindings"].(string))
		ls := make([]any, 4097)
		for i := range ls {
			ls[i] = map[string]any{"requirementId": "UR-001", "scenarioCaseId": "a_seal"}
		}
		b["links"] = ls
		q["bindings"] = frozenJSON(t, b)
		refuse(t, frozenJSON(t, q), "limit")
	})
	for _, field := range []string{"formatVersion", "model", "requirements", "catalog", "suite", "bindings", "scenarioReport"} {
		for _, kind := range []string{"missing", "null", "case", "duplicate"} {
			if field == "scenarioReport" && kind == "null" {
				continue
			}
			t.Run("wrapper_"+field+"_"+kind, func(t *testing.T) {
				q := request(nil)
				var raw string
				switch kind {
				case "missing":
					delete(q, field)
				case "null":
					q[field] = nil
				case "case":
					q[strings.ToUpper(field)] = q[field]
					delete(q, field)
				case "duplicate":
					raw = frozenJSON(t, q)
					raw = raw[:len(raw)-1] + `,"` + field + `":null}`
				}
				if raw == "" {
					raw = frozenJSON(t, q)
				}
				refuse(t, raw, "invalid")
			})
		}
	}
	for _, kind := range []string{"unknown", "trailing", "unicode", "version"} {
		t.Run("strict_"+kind, func(t *testing.T) {
			q := request(nil)
			raw := ""
			switch kind {
			case "unknown":
				q["invented-secret"] = true
			case "version":
				q["formatVersion"] = "v2"
			case "trailing":
				raw = frozenJSON(t, q) + " {}"
			case "unicode":
				raw = strings.Replace(frozenJSON(t, q), "lifecycle-scenario-trace-request/v1", `\ud800`, 1)
			}
			if raw == "" {
				raw = frozenJSON(t, q)
			}
			refuse(t, raw, "invalid")
		})
	}
	for _, kind := range []string{"model", "scenario", "source", "null_source", "expected", "command", "initial", "chain", "omitted", "extra", "reordered", "counts", "after_mismatch"} {
		t.Run("report_"+kind, func(t *testing.T) {
			r := frozenObject(t, v.Reports["matched"])
			es := r["entries"].([]any)
			first := es[0].(map[string]any)
			want := "invalid"
			switch kind {
			case "model":
				r["modelDigest"] = strings.Repeat("0", 64)
			case "scenario":
				r["scenarioDigest"] = strings.Repeat("0", 64)
				want = "invalid"
			case "source":
				r["sourceDigest"] = strings.Repeat("0", 64)
				want = "invalid"
			case "null_source":
				r["sourceDigest"] = nil
				want = "invalid"
			case "expected":
				first["expected"].(map[string]any)["refusals"] = []any{"OTHER"}
				first["actual"].(map[string]any)["refusals"] = []any{"OTHER"}
				want = "invalid"
			case "command":
				first["command"].(map[string]any)["actor"] = "invented-secret"
				want = "invalid"
			case "initial":
				es[2].(map[string]any)["input"].(map[string]any)["outcome"] = "invented-secret"
				want = "invalid"
			case "chain":
				es[1].(map[string]any)["input"].(map[string]any)["revision"] = "1"
			case "omitted":
				es = es[:3]
				r["stepsExecuted"] = 3
				r["casesCompleted"] = 2
				want = "invalid"
			case "extra":
				x := frozenObject(t, frozenJSON(t, es[3]))
				x["caseId"] = "invented_extra"
				x["stepId"] = "extra_step"
				es = append(es, x)
				r["stepsExecuted"] = 5
				r["casesCompleted"] = 4
				want = "invalid"
			case "reordered":
				es[2], es[3] = es[3], es[2]
				want = "invalid"
			case "counts":
				r["casesCompleted"] = 2
			case "after_mismatch":
				r = frozenObject(t, v.Reports["early"])
				es = r["entries"].([]any)
				es = append(es, frozenObject(t, v.Reports["matched"])["entries"].([]any)[2])
				r["stepsExecuted"] = 3
			}
			r["entries"] = es
			q := request(frozenJSON(t, r))
			refuse(t, frozenJSON(t, q), want)
		})
	}
}
