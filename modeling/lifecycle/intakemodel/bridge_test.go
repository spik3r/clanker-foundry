package intakemodel

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"strings"
	"sync"
	"testing"
)

type vector struct {
	Name        string `json:"name"`
	RequestJSON string `json:"requestJSON"`
	Expected    struct {
		IntakeReportJSON   string     `json:"intakeReportJSON"`
		IntakeReportSHA256 string     `json:"intakeReportSHA256"`
		Model              ModelCheck `json:"model"`
	} `json:"expected"`
}
type oracle struct {
	Invariants Report          `json:"expectedReportInvariants"`
	Schema     SchemaReference `json:"schema"`
	Cases      []vector        `json:"cases"`
}

func vectors(t *testing.T) oracle {
	t.Helper()
	b, err := os.ReadFile("testdata/expected.json")
	if err != nil {
		t.Fatal(err)
	}
	if digest(b) != "1a951b981246cbc395e9e439df381b80a9909df9e1d7bd7e3d9f00da515386b6" {
		t.Fatal("independent preimplementation vectors changed; never refresh from bridge output")
	}
	var v oracle
	if err = json.Unmarshal(b, &v); err != nil {
		t.Fatal(err)
	}
	if len(v.Cases) != 6 {
		t.Fatal("incomplete independent fixture")
	}
	return v
}

func reportFrom(t *testing.T, raw []byte) Report {
	t.Helper()
	var r Report
	if err := json.Unmarshal(raw, &r); err != nil {
		t.Fatal(err)
	}
	return r
}

func modify(t *testing.T, raw string, edit func(map[string]any)) []byte {
	t.Helper()
	var d map[string]any
	if err := json.Unmarshal([]byte(raw), &d); err != nil {
		t.Fatal(err)
	}
	edit(d)
	b, err := json.Marshal(d)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func refuse(t *testing.T, raw []byte, want error) {
	t.Helper()
	b, err := Execute(raw)
	if b != nil || !errors.Is(err, want) {
		t.Fatalf("Execute got %d bytes / %v, want nil / %v", len(b), err, want)
	}
	r, err := Inspect(raw)
	if r != nil || !errors.Is(err, want) {
		t.Fatalf("Inspect got %v / %v, want nil / %v", r, err, want)
	}
}

func TestIndependentPreimplementationVectors(t *testing.T) {
	f := vectors(t)
	for _, v := range f.Cases {
		t.Run(v.Name, func(t *testing.T) {
			want := f.Invariants
			want.Schema = f.Schema
			want.IntakeReportJSON = v.Expected.IntakeReportJSON
			want.IntakeReportSHA256 = v.Expected.IntakeReportSHA256
			want.Model = v.Expected.Model
			raw := []byte(v.RequestJSON)
			before := bytes.Clone(raw)
			got, err := Inspect(raw)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(*got, want) {
				t.Fatalf("independent report mismatch\ngot %+v\nwant %+v", *got, want)
			}
			b, err := Execute(raw)
			if err != nil || !reflect.DeepEqual(reportFrom(t, b), want) {
				t.Fatal("encoded independent report mismatch", err)
			}
			if !bytes.Equal(raw, before) {
				t.Fatal("request mutated")
			}
			if !strings.HasSuffix(got.IntakeReportJSON, "\n") || digest([]byte(got.IntakeReportJSON)) != got.IntakeReportSHA256 {
				t.Fatal("native report byte identity lost")
			}
			if got.Model.Normalized != nil && (digest([]byte(got.Model.Normalized.SourceJSON)) != got.Model.Normalized.SourceSHA256 || !strings.HasSuffix(got.Model.Normalized.SourceJSON, "\n")) {
				t.Fatal("normalized artifact identity lost")
			}
		})
	}
}

func TestModelValidityDoesNotPromoteIntakeClaims(t *testing.T) {
	f := vectors(t)
	for _, i := range []int{1, 5} {
		r, err := Inspect([]byte(f.Cases[i].RequestJSON))
		if err != nil {
			t.Fatal(err)
		}
		if r.Model.Status != "accepted" || r.Authority != "none" || r.ExecutionStatus != "not-run" || r.AcceptanceStatus != "not-evaluated" || r.ReviewStatus != "not-established" || r.Schema.Validation != "not-run" {
			t.Fatal("model acceptance promoted unrelated assurance")
		}
		if r.IntakeReportJSON != f.Cases[i].Expected.IntakeReportJSON {
			t.Fatal("declarations/questions were rewritten")
		}
	}
	r, err := Inspect([]byte(f.Cases[2].RequestJSON))
	if err != nil {
		t.Fatal(err)
	}
	if r.Model.Status != "rejected" || !strings.Contains(r.IntakeReportJSON, `"status":"matches-declaration"`) {
		t.Fatal("hash match treated as model validation")
	}
}

func TestSourceCanonicalAndNormalizedIdentityAreSeparate(t *testing.T) {
	f := vectors(t)
	a, err := Inspect([]byte(f.Cases[0].RequestJSON))
	if err != nil {
		t.Fatal(err)
	}
	b, err := Inspect([]byte(f.Cases[5].RequestJSON))
	if err != nil {
		t.Fatal(err)
	}
	if *a.Model.SourceSHA256 == *b.Model.SourceSHA256 || *a.Model.CanonicalModelSHA256 != *b.Model.CanonicalModelSHA256 || !reflect.DeepEqual(a.Model.Normalized, b.Model.Normalized) {
		t.Fatal("CRLF original bytes were normalized or canonical identity changed")
	}
	if *a.Model.SourceSHA256 == a.Model.Normalized.SourceSHA256 {
		t.Fatal("fixture should distinguish original and normalized source")
	}
	white, err := Inspect(append([]byte(f.Cases[0].RequestJSON), '\n'))
	if err != nil {
		t.Fatal(err)
	}
	if white.IntakeReportSHA256 == a.IntakeReportSHA256 || !reflect.DeepEqual(white.Model, a.Model) {
		t.Fatal("request identity confused with model identity")
	}
}

func TestOnlyObservedModelIsInterpreted(t *testing.T) {
	f := vectors(t)
	raw := modify(t, f.Cases[0].RequestJSON, func(d map[string]any) {
		obs := d["observed"].(map[string]any)
		obs["requirements"] = "\x00 not a requirements file <script>"
		obs["scenarioSuite"] = "not JSON and no model header"
	})
	r, err := Inspect(raw)
	if err != nil || r.Model.Status != "accepted" {
		t.Fatal("bridge interpreted opaque peer artifacts", err)
	}
	// Empty observed text is present and invalid, not a missing observation.
	raw = modify(t, f.Cases[0].RequestJSON, func(d map[string]any) { d["observed"].(map[string]any)["model"] = "" })
	r, err = Inspect(raw)
	if err != nil {
		t.Fatal(err)
	}
	if r.Model.Status != "rejected" || r.Model.SourceSHA256 == nil || *r.Model.SourceSHA256 != digest(nil) || r.Model.Normalized != nil || r.Model.CanonicalModelSHA256 != nil || r.Model.Problem == nil || *r.Model.Problem != "model-validation-failed" {
		t.Fatal("empty observation misclassified")
	}
}

func TestNativeIntakeValidationPrecedesModelExtraction(t *testing.T) {
	f := vectors(t)
	valid := []byte(f.Cases[0].RequestJSON)
	for _, raw := range [][]byte{
		append(bytes.Clone(valid), []byte(" {}")...), append([]byte("\xef\xbb\xbf"), valid...),
		bytes.Replace(valid, []byte(`"intent":`), []byte(`"Intent":`), 1),
		bytes.Replace(valid, []byte(`"intent":`), []byte(`"intent":"duplicate","intent":`), 1),
		bytes.Replace(valid, []byte(`"intent":`), []byte(`"extra":true,"intent":`), 1),
	} {
		refuse(t, raw, ErrInvalid)
	}
	for _, key := range []string{"intent", "decisions", "questions", "artifacts", "dependencies", "observed"} {
		refuse(t, modify(t, f.Cases[0].RequestJSON, func(d map[string]any) { delete(d, key) }), ErrInvalid)
		refuse(t, modify(t, f.Cases[0].RequestJSON, func(d map[string]any) { d[key] = nil }), ErrInvalid)
	}
	for _, edit := range []func(map[string]any){
		func(d map[string]any) { d["artifacts"].(map[string]any)["model"].(map[string]any)["sha256"] = "bad" },
		func(d map[string]any) {
			d["observed"].(map[string]any)["model"] = map[string]any{"invalid": "not a string"}
		},
		func(d map[string]any) { d["observed"].(map[string]any)["extra"] = "field" },
		func(d map[string]any) { d["observed"].(map[string]any)["model"] = strings.Repeat("a", (1<<20)+1) },
	} {
		refuse(t, modify(t, f.Cases[0].RequestJSON, edit), ErrInvalid)
	}
	refuse(t, []byte("\xff"), ErrInvalid)
	refuse(t, make([]byte, MaxSourceBytes+1), ErrLimit)
}

func TestModelBoundaryUsesStrictExistingAdapter(t *testing.T) {
	f := vectors(t)
	var base map[string]any
	_ = json.Unmarshal([]byte(f.Cases[0].RequestJSON), &base)
	modelSource := base["observed"].(map[string]any)["model"].(string)
	for _, bad := range []string{
		modelSource + " {}", "\ufeff" + modelSource,
		strings.Replace(modelSource, `"id":`, `"id":"duplicate","id":`, 1),
		strings.Replace(modelSource, `"initial":`, `"Initial":`, 1),
		`{"formatVersion":"lifecycle-json/v1","id":"\ud800"}`,
		strings.Replace(modelSource, `"initial": "Draft"`, `"initial": "Unknown"`, 1),
	} {
		raw := modify(t, f.Cases[0].RequestJSON, func(d map[string]any) { d["observed"].(map[string]any)["model"] = bad })
		r, err := Inspect(raw)
		if err != nil {
			t.Fatal(err)
		}
		if r.Model.Status != "rejected" || r.Model.Normalized != nil || r.Model.CanonicalModelSHA256 != nil || *r.Model.SourceSHA256 != digest([]byte(bad)) {
			t.Fatal("strict model adapter was bypassed")
		}
	}
}

func TestMergedAuthoringCorpusSeparatesShapeFromSemantics(t *testing.T) {
	f := vectors(t)
	b, err := os.ReadFile("../authoring/cases.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []struct {
		Name        string
		SchemaValid bool
		ModelValid  bool
		Document    json.RawMessage
	}
	if err = json.Unmarshal(b, &cases); err != nil {
		t.Fatal(err)
	}
	separation := 0
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			raw := modify(t, f.Cases[0].RequestJSON, func(d map[string]any) { d["observed"].(map[string]any)["model"] = string(c.Document) })
			r, err := Inspect(raw)
			if err != nil {
				t.Fatal(err)
			}
			if (r.Model.Status == "accepted") != c.ModelValid || r.Schema.Validation != "not-run" {
				t.Fatal("schema shape confused with model semantics")
			}
			if c.SchemaValid && !c.ModelValid {
				separation++
			}
		})
	}
	if separation == 0 {
		t.Fatal("corpus no longer exercises schema-valid/model-invalid distinction")
	}
}

func TestPinnedSchemaReferenceIsNotFetchedOrExecuted(t *testing.T) {
	b, err := os.ReadFile("../" + SchemaPath)
	if err != nil {
		t.Fatal(err)
	}
	if digest(b) != SchemaSHA256 {
		t.Fatal("authoring schema changed; deliberately review reference pin instead of silently relabeling")
	}
	var schema struct {
		ID string `json:"$id"`
	}
	if err = json.Unmarshal(b, &schema); err != nil || schema.ID != SchemaID {
		t.Fatal("schema ID drift", err)
	}
	f := vectors(t)
	r, err := Inspect([]byte(f.Cases[0].RequestJSON))
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(r.Schema, f.Schema) || r.Schema.Validation != "not-run" {
		t.Fatal("incorrect schema assurance")
	}
}

func TestPrivateTextEscapingAndOutputBounds(t *testing.T) {
	f := vectors(t)
	raw := modify(t, f.Cases[0].RequestJSON, func(d map[string]any) { d["intent"] = "private <script>& \x00\nintent" })
	b, err := Execute(raw)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(b, []byte("<script>")) || bytes.Contains(b, []byte{0}) {
		t.Fatal("unescaped private text")
	}
	r := reportFrom(t, b)
	var native map[string]any
	if err = json.Unmarshal([]byte(r.IntakeReportJSON), &native); err != nil {
		t.Fatal(err)
	}
	if native["intent"] != "private <script>& \x00\nintent" || digest([]byte(r.IntakeReportJSON)) != r.IntakeReportSHA256 {
		t.Fatal("escaping changed native artifact")
	}
	tooLarge := &Report{IntakeReportJSON: strings.Repeat("\x00", MaxReportBytes/6+1)}
	b, err = encode(tooLarge)
	if b != nil || !errors.Is(err, ErrLimit) {
		t.Fatal("escaped report ceiling not enforced")
	}
}

func TestOwnershipAndConcurrentDeterminism(t *testing.T) {
	f := vectors(t)
	raw := []byte(f.Cases[0].RequestJSON)
	want, err := Execute(raw)
	if err != nil {
		t.Fatal(err)
	}
	r, err := Inspect(raw)
	if err != nil {
		t.Fatal(err)
	}
	r.IntakeReportJSON = "changed"
	*r.Model.SourceSHA256 = "changed"
	r.Model.Normalized.SourceJSON = "changed"
	var wg sync.WaitGroup
	for i := 0; i < 12; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			b, err := Execute(raw)
			if err != nil || !bytes.Equal(b, want) {
				t.Error("report alias or nondeterminism", err)
			}
		}()
	}
	wg.Wait()
	got, err := Inspect(raw)
	if err != nil || got.Model.Normalized.SourceJSON == "changed" || *got.Model.SourceSHA256 == "changed" {
		t.Fatal("owned result aliases later call")
	}
	before := got.IntakeReportJSON
	for i := range raw {
		raw[i] = 'x'
	}
	if got.IntakeReportJSON != before {
		t.Fatal("report aliases raw input")
	}
}
