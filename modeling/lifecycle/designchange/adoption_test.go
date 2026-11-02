// Independently authored before current-main designchange adoption. The six
// identities and request/suite/expected fixtures predate this adoption; source
// pins reuse the independently reviewed current qualification closure. No target
// exporter, verifier or generated runner authored these expected values.
package designchange_test

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/artifactbundle"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/canonical"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/designchange"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/impact"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/jsonformat"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/qualification"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
)

// Legacy policy identity and eleven pins remain readable and unchanged.
const adoptionRevision = "3d6d17188cc2723cfe04649d0f8b4220a37b1afe"

const adoptionCurrentRevision = "1e88dbebb4ab13873e2a4159dc928dd1f5f73e20"

func adoptionExpected() designchange.Expected {
	return designchange.Expected{
		RequestSHA256:        "f3a14ba6adfd54832d810e26dccfc929e5fa61801990fa7c52436ec51eaacde0",
		BeforeSourceSHA256:   "95b17eb3e2ef78a9485b40076c0b085e5e542e8fa3db827ea99345abe6077e26",
		BeforeModelDigest:    "948630cb7857199443b59b92d9b1ab0fadf2bf7a5bf8125f7a7e767db83958ed",
		CandidateSHA256:      "22e6ee6f5208b124bcebffe2ac8608920a74cd34c0eeefe8ab8519b2523ea8c6",
		CandidateModelDigest: "a9ad2dfb926a24a026c8565f0d929a5b803bec43f6eb7342e196830791eccc7c",
		SuiteSHA256:          "53b5ceec6b68719243853dae5285c13d5c25df89cb203a16279f3cacd2f989f1",
	}
}

type adoptionPin struct {
	name string
	size int
	sha  string
}

var adoptionPins = []adoptionPin{
	{"go.mod", 71, "cc128fce13193d3066bdeb08cca9db627fb6c3120967c3beff254a514f2a5fac"},
	{"model/ir.go", 9385, "13f4ad2980b1ca523b07bdbc8ebeb0e0300f8df81faa8dcd1cdec5ccef007aeb"},
	{"canonical/canonical.go", 1946, "2c3cba3fd53adf8ca5a4cc2ce3d96b51457742c47b6ab4fcedba219e19001a06"},
	{"strictjson/decode.go", 4236, "a72ded4af99ffba5f5cca2931c6cace072634ee2e42eb00cd9a5988c691f1ee2"},
	{"compiler/generate.go", 12315, "e17610a27748a5d3baedf4f59c2183cb184f7e2681b5532c6f9d8d3d1eda6a38"},
	{"scenarios/types.go", 5904, "147b12bd1726fec156ba08ec5666a9a362e1563115151ae57bf9face70f86eec"},
	{"scenarios/decode.go", 7047, "e6119229899d3084820d72a777d1771037f871b14554630067b7da022cdedc15"},
	{"scenarios/run.go", 10114, "46d41c5d27d87f71f7191f18d182a99c3ca491126afb601599067692c3379dde"},
	{"cmd/modelplay/main.go", 1613, "925ec202c74aba5d50eb2bfb5dc54ea22387ccec90e969f1a7ebe1ecbd9a0c63"},
	{"scenarios/runner.go", 4005, "e85b2146d7c1d16269ff886f6c67084de5e263db8b01a7676b8c7491ed8fb59c"},
	{"scenarios/report_decode.go", 3626, "f14816a19faacf707e23b01d5d16b26be8e8c03a9f5f9af801029b89fb2ceddf"},
}

func adoptionSHA(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }
func adoptionRead(t *testing.T, path, sha string, size int) []byte {
	t.Helper()
	b, e := os.ReadFile(path)
	if e != nil {
		t.Fatal(e)
	}
	if len(b) != size || adoptionSHA(b) != sha {
		t.Fatal("frozen source identity drift", path)
	}
	return b
}
func adoptionSources(t *testing.T) []artifactbundle.File {
	t.Helper()
	want := []string{}
	out := []artifactbundle.File{}
	for _, p := range adoptionCurrentPins() {
		want = append(want, p.name)
		out = append(out, artifactbundle.File{Name: p.name, Data: adoptionRead(t, filepath.Join("..", p.name), p.sha, p.size)})
	}
	got := qualification.SourceNames()
	sort.Strings(got)
	sort.Strings(want)
	if !reflect.DeepEqual(got, want) || qualification.SourceRevision != adoptionCurrentRevision {
		t.Fatal("current qualifier closure changed")
	}
	return out
}
func adoptionInputs(t *testing.T) ([]byte, []byte) {
	t.Helper()
	w := adoptionExpected()
	r := adoptionRead(t, "testdata/request.json", w.RequestSHA256, 9295)
	s := adoptionRead(t, "testdata/suite.json", w.SuiteSHA256, 6656)
	adoptionRead(t, "testdata/expected.json", "efc7cc5088e1a79210303fe410af6c74831ceab9a77d1f4075efc4fb338c47a3", 1417)
	return r, s
}
func adoptionUnpack(t *testing.T, raw []byte, profile string, names []string) map[string][]byte {
	t.Helper()
	files, e := artifactbundle.Verify(raw, profile)
	if e != nil {
		t.Fatal(e)
	}
	got := []string{}
	out := map[string][]byte{}
	for _, f := range files {
		got = append(got, f.Name)
		out[f.Name] = f.Data
	}
	want := append([]string{}, names...)
	sort.Strings(want)
	sort.Strings(got)
	if !reflect.DeepEqual(got, want) {
		t.Fatal("complete payload inventory mismatch", got, want)
	}
	z, e := zip.NewReader(bytes.NewReader(raw), int64(len(raw)))
	if e != nil || len(z.File) != len(names)+1 {
		t.Fatal("archive must add exactly one manifest", e)
	}
	return out
}
func adoptionOuterNames() []string {
	return []string{"GUIDE.txt", "before.json", "candidate.json", "change-request.json", "design-change.json", "planner-result.json", "qualification.zip"}
}
func adoptionInnerNames() []string {
	n := []string{"GUIDE.txt", "candidate.json", "suite.json", "qualification.json", "generated/lifecycle.go", "generated/model_paths_test.go"}
	for _, p := range adoptionCurrentPins() {
		n = append(n, p.name)
	}
	return n
}
func adoptionJSON(t *testing.T, b []byte) map[string]any {
	t.Helper()
	var m map[string]any
	if e := json.Unmarshal(b, &m); e != nil {
		t.Fatal(e)
	}
	return m
}
func adoptionCheckMetadata(t *testing.T, outer, inner map[string][]byte, w designchange.Expected) {
	t.Helper()
	// Only the two hashes of dynamically generated artifacts are consistency
	// assertions computed from their bytes. They are not independent golden output
	// digests, provenance, or a substitute for the six frozen external identities.
	want := map[string]any{"formatVersion": "lifecycle-designchange/v1", "mode": "example-only", "authority": "none", "requestSHA256": w.RequestSHA256, "beforeSourceSHA256": w.BeforeSourceSHA256, "beforeModelDigest": w.BeforeModelDigest, "candidateSHA256": w.CandidateSHA256, "candidateModelDigest": w.CandidateModelDigest, "suiteSHA256": w.SuiteSHA256, "plannerResultSHA256": adoptionSHA(outer["planner-result.json"]), "qualificationSHA256": adoptionSHA(outer["qualification.zip"]), "modelFormatVersion": "lifecycle-json/v1", "profileVersion": "lifecycle/v1", "generatorVersion": "lifecycle-go/v1", "toolchainVersion": "go1.27.3", "qualificationSourceRevision": adoptionCurrentRevision, "suiteStatus": "unvalidated and unexecuted", "suiteBindingStatus": "not inspected"}
	if !reflect.DeepEqual(adoptionJSON(t, outer["design-change.json"]), want) {
		t.Fatal("outer metadata/status/authority drift")
	}
	innerWant := map[string]any{"format": "lifecycle-json/v1", "profile": "lifecycle/v1", "generator": "lifecycle-go/v1", "toolchain": "go1.27.3", "modelDigest": w.CandidateModelDigest, "sourceRevision": adoptionCurrentRevision, "suiteStatus": "unvalidated and unexecuted"}
	if !reflect.DeepEqual(adoptionJSON(t, inner["qualification.json"]), innerWant) {
		t.Fatal("nested metadata or suite status drift")
	}
}
func adoptionModelDigest(t *testing.T, raw []byte, want string) {
	t.Helper()
	m, e := jsonformat.Decode(raw)
	if e != nil {
		t.Fatal(e)
	}
	d, e := canonical.Digest(m)
	if e != nil || d != want {
		t.Fatal("semantic identity drift", d, e)
	}
}
func TestAdoptionCurrentNestedClosureAndFrozenIdentities(t *testing.T) {
	request, suite := adoptionInputs(t)
	sources := adoptionSources(t)
	w := adoptionExpected()
	beforeRequest, beforeSuite := bytes.Clone(request), bytes.Clone(suite)
	saved := []artifactbundle.File{}
	for _, f := range sources {
		saved = append(saved, artifactbundle.File{Name: f.Name, Data: bytes.Clone(f.Data)})
	}
	raw, e := designchange.Build(request, suite, sources)
	if e != nil {
		t.Fatal(e)
	}
	again, e := designchange.Build(request, suite, sources)
	if e != nil || !bytes.Equal(raw, again) {
		t.Fatal("nondeterministic export", e)
	}
	if e = designchange.Verify(raw, w); e != nil {
		t.Fatal(e)
	}
	outer := adoptionUnpack(t, raw, "lifecycle-designchange/v1", adoptionOuterNames())
	inner := adoptionUnpack(t, outer["qualification.zip"], "lifecycle-qualification/v1", adoptionInnerNames())
	adoptionCheckMetadata(t, outer, inner, w)
	var envelope struct {
		Before json.RawMessage `json:"before"`
	}
	if e = json.Unmarshal(request, &envelope); e != nil {
		t.Fatal(e)
	}
	if !bytes.Equal(outer["change-request.json"], request) || !bytes.Equal(outer["before.json"], envelope.Before) || !bytes.Equal(inner["suite.json"], suite) || !bytes.Equal(outer["candidate.json"], inner["candidate.json"]) {
		t.Fatal("request, before token, opaque suite or nested candidate bytes changed")
	}
	if adoptionSHA(outer["before.json"]) != w.BeforeSourceSHA256 || adoptionSHA(outer["candidate.json"]) != w.CandidateSHA256 || adoptionSHA(inner["suite.json"]) != w.SuiteSHA256 {
		t.Fatal("frozen raw identity mismatch")
	}
	adoptionModelDigest(t, outer["before.json"], w.BeforeModelDigest)
	adoptionModelDigest(t, outer["candidate.json"], w.CandidateModelDigest)
	// The old separately formatted candidate is NOT the normalized archive source.
	if adoptionSHA(outer["candidate.json"]) == "9674eb213352232f354c2c338709baf6a3230a599ef89845e1cfac5525d4c6e1" {
		t.Fatal("normalized candidate confused with old formatted source")
	}
	for _, p := range adoptionCurrentPins() {
		if len(inner[p.name]) != p.size || adoptionSHA(inner[p.name]) != p.sha {
			t.Fatal("inner source pin differs", p.name)
		}
	}
	native, e := qualification.Build(outer["candidate.json"], suite, sources)
	if e != nil || !bytes.Equal(native, outer["qualification.zip"]) {
		t.Fatal("nested project differs from native current qualification export", e)
	}
	if e = qualification.Verify(outer["qualification.zip"], w.CandidateSHA256, w.SuiteSHA256); e != nil {
		t.Fatal("native qualifier rejected nested project", e)
	}
	var planner struct {
		ReportVersion  string
		ProfileVersion string
		Mode           string
		Authority      string
		BeforeDigest   string
		AfterDigest    string
		After          json.RawMessage
		Review         impact.Report
	}
	if e = json.Unmarshal(outer["planner-result.json"], &planner); e != nil {
		t.Fatal(e)
	}
	wantReview := impact.Report{ReportVersion: "lifecycle-model-diff-report/v1", ProfileVersion: "lifecycle/v1", Mode: "example-only", Authority: "none", BeforeDigest: w.BeforeModelDigest, AfterDigest: w.CandidateModelDigest, Changed: true, ReviewRequired: true, Changes: []impact.Change{{Kind: "transition", ID: "T_prepare", Type: "modified", Fields: []string{"guards"}}}, LinkedTransitions: []impact.LinkedTransition{{Side: "before", ID: "T_prepare", Causes: []string{"transition:T_prepare"}}, {Side: "after", ID: "T_prepare", Causes: []string{"transition:T_prepare"}}}, Bindings: []impact.BindingResult{}}
	if planner.ReportVersion != "lifecycle-model-patch-report/v1" || planner.ProfileVersion != "lifecycle/v1" || planner.Mode != "example-only" || planner.Authority != "none" || planner.BeforeDigest != w.BeforeModelDigest || planner.AfterDigest != w.CandidateModelDigest || !reflect.DeepEqual(planner.Review, wantReview) {
		t.Fatal("literal declared impact or planner assurance differs")
	}
	adoptionModelDigest(t, planner.After, w.CandidateModelDigest)
	if !bytes.Equal(request, beforeRequest) || !bytes.Equal(suite, beforeSuite) || !reflect.DeepEqual(sources, saved) {
		t.Fatal("caller inputs mutated")
	}
	for i := 0; i < 6; i++ {
		wrong := w
		fields := []*string{&wrong.RequestSHA256, &wrong.BeforeSourceSHA256, &wrong.BeforeModelDigest, &wrong.CandidateSHA256, &wrong.CandidateModelDigest, &wrong.SuiteSHA256}
		*fields[i] = strings.Repeat("0", 64)
		if designchange.Verify(raw, wrong) == nil {
			t.Fatal("ignored independent external identity", i)
		}
	}
}
func TestAdoptionOpaqueSuiteNeverAcquiresValidationOrBinding(t *testing.T) {
	request, _ := adoptionInputs(t)
	sources := adoptionSources(t)
	for _, literal := range []string{"{ malformed suite\n", `{"modelDigest":"wrong binding","cases":null}`} {
		suite := []byte(literal)
		raw, e := designchange.Build(request, suite, sources)
		if e != nil {
			t.Fatal(e)
		}
		w := adoptionExpected()
		w.SuiteSHA256 = adoptionSHA(suite)
		if e = designchange.Verify(raw, w); e != nil {
			t.Fatal(e)
		}
		outer := adoptionUnpack(t, raw, designchange.Profile, adoptionOuterNames())
		inner := adoptionUnpack(t, outer["qualification.zip"], artifactbundle.QualificationProfile, adoptionInnerNames())
		if !bytes.Equal(inner["suite.json"], suite) {
			t.Fatal("opaque suite rewritten")
		}
		adoptionCheckMetadata(t, outer, inner, w)
	}
}
func adoptionRepack(t *testing.T, profile string, m map[string][]byte) []byte {
	t.Helper()
	files := []artifactbundle.File{}
	for n, b := range m {
		files = append(files, artifactbundle.File{Name: n, Data: b})
	}
	out, e := artifactbundle.Build(profile, files)
	if e != nil {
		t.Fatal(e)
	}
	return out
}
func TestAdoptionRejectsMissingTamperedAndOldNestedClosure(t *testing.T) {
	request, suite := adoptionInputs(t)
	sources := adoptionSources(t)
	w := adoptionExpected()
	for _, name := range []string{"scenarios/runner.go", "scenarios/report_decode.go"} {
		missing := []artifactbundle.File{}
		tampered := []artifactbundle.File{}
		for _, f := range sources {
			if f.Name != name {
				missing = append(missing, f)
			}
			g := artifactbundle.File{Name: f.Name, Data: bytes.Clone(f.Data)}
			if g.Name == name {
				g.Data[0] ^= 1
			}
			tampered = append(tampered, g)
		}
		for _, src := range [][]artifactbundle.File{missing, tampered} {
			out, e := designchange.Build(request, suite, src)
			if e == nil || out != nil {
				t.Fatal("missing or stale current dependency accepted", name)
			}
		}
	}
	original, e := designchange.Build(request, suite, sources)
	if e != nil {
		t.Fatal(e)
	}
	for _, mutation := range []string{"missing-runner", "missing-report-decoder", "old-nine-source-inventory", "stale-revision", "tampered-runner", "tampered-report-decoder"} {
		outer := adoptionUnpack(t, original, designchange.Profile, adoptionOuterNames())
		inner := adoptionUnpack(t, outer["qualification.zip"], artifactbundle.QualificationProfile, adoptionInnerNames())
		switch mutation {
		case "missing-runner":
			delete(inner, "scenarios/runner.go")
		case "missing-report-decoder":
			delete(inner, "scenarios/report_decode.go")
		case "old-nine-source-inventory":
			delete(inner, "model/capabilities.go")
			delete(inner, "scenarios/runner.go")
			delete(inner, "scenarios/report_decode.go")
		case "stale-revision":
			meta := adoptionJSON(t, inner["qualification.json"])
			meta["sourceRevision"] = "3312a090afac526ddeda392dbd31ed49fa0e04ca"
			inner["qualification.json"], e = json.Marshal(meta)
			if e != nil {
				t.Fatal(e)
			}
			inner["qualification.json"] = append(inner["qualification.json"], '\n')
		case "tampered-runner":
			inner["scenarios/runner.go"][0] ^= 1
		case "tampered-report-decoder":
			inner["scenarios/report_decode.go"][0] ^= 1
		}
		nested := adoptionRepack(t, artifactbundle.QualificationProfile, inner)
		// Repack correct generic manifests and repair ONLY the outer declared nested
		// archive hash. Domain/source trust must still come from the native qualifier.
		outer["qualification.zip"] = nested
		metadata := adoptionJSON(t, outer["design-change.json"])
		metadata["qualificationSHA256"] = adoptionSHA(nested)
		outer["design-change.json"], e = json.Marshal(metadata)
		if e != nil {
			t.Fatal(e)
		}
		outer["design-change.json"] = append(outer["design-change.json"], '\n')
		changed := adoptionRepack(t, designchange.Profile, outer)
		if qualification.Verify(nested, w.CandidateSHA256, w.SuiteSHA256) == nil || designchange.Verify(changed, w) == nil {
			t.Fatal("stale or incomplete nested domain accepted", mutation)
		}
	}
}

// New default export uses the actual helper-freeze commit; the legacy public
// archive remains tested independently by profilecatalog's compatibility oracle.
func adoptionCurrentPins() []adoptionPin {
	out := append([]adoptionPin{}, adoptionPins...)
	return append(out, adoptionPin{"model/capabilities.go", 654, "e78f4f9c102f8d894e50130dd0bfb8030e0c7c325b01eb2452e9839c02c27e4e"})
}
