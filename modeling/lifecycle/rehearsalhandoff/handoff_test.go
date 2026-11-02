package rehearsalhandoff

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"testing"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/artifactbundle"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/designchange"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/qualification"
)

func read(t *testing.T, p string) []byte {
	t.Helper()
	b, e := os.ReadFile(p)
	if e != nil {
		t.Fatal(e)
	}
	return b
}
func digestBytes(b []byte) string { return fmt.Sprintf("%x", sha256.Sum256(b)) }
func root(t *testing.T) string {
	t.Helper()
	_, f, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("no source path")
	}
	return filepath.Dir(filepath.Dir(f))
}
func compactReport(t *testing.T, raw json.RawMessage) []byte {
	t.Helper()
	var b bytes.Buffer
	if e := json.Compact(&b, raw); e != nil {
		t.Fatal(e)
	}
	return append(b.Bytes(), '\n')
}
func fixture(t *testing.T) (Material, References) {
	t.Helper()
	dir := root(t)
	expectedBytes := read(t, filepath.Join(dir, "rehearsal/testdata/expected.json"))
	var x struct {
		ExternalExpected    struct{ RequestSHA256, BeforeSourceSHA256, BeforeModelDigest, CandidateSHA256, CandidateModelDigest, SuiteSHA256 string }
		BaselineSuiteSHA256 string
	}
	if e := json.Unmarshal(expectedBytes, &x); e != nil {
		t.Fatal(e)
	}
	oracleBytes := read(t, filepath.Join(dir, "rehearsalhandoff/testdata/report-identities.json"))
	if digestBytes(oracleBytes) != "dcc235323c66de66aa9cba3fa805260ea1008b71a4c383960cf5f8eb13646a61" {
		t.Fatal("frozen report oracle changed")
	}
	var o struct {
		Baseline struct {
			SHA256 string          `json:"sha256"`
			Report json.RawMessage `json:"report"`
		}
		Candidate struct {
			SHA256 string          `json:"sha256"`
			Report json.RawMessage `json:"report"`
		}
	}
	if e := json.Unmarshal(oracleBytes, &o); e != nil {
		t.Fatal(e)
	}
	br, cr := compactReport(t, o.Baseline.Report), compactReport(t, o.Candidate.Report)
	if digestBytes(br) != o.Baseline.SHA256 || digestBytes(cr) != o.Candidate.SHA256 {
		t.Fatal("independently frozen report bytes mismatch")
	}
	req := read(t, filepath.Join(dir, "designchange/testdata/request.json"))
	base := read(t, filepath.Join(dir, "rehearsal/testdata/baseline-suite.json"))
	candidate := read(t, filepath.Join(dir, "designchange/testdata/suite.json"))
	sources := []artifactbundle.File{}
	for _, n := range qualification.SourceNames() {
		sources = append(sources, artifactbundle.File{Name: n, Data: read(t, filepath.Join(dir, n))})
	}
	archive, e := designchange.Build(req, candidate, sources)
	if e != nil {
		t.Fatal(e)
	}
	outer, e := artifactbundle.Verify(archive, artifactbundle.DesignChangeProfile)
	if e != nil {
		t.Fatal(e)
	}
	inner, e := artifactbundle.Verify(fileMap(outer)["qualification.zip"], artifactbundle.QualificationProfile)
	if e != nil {
		t.Fatal(e)
	}
	im := fileMap(inner)
	refs := References{x.ExternalExpected.RequestSHA256, x.ExternalExpected.BeforeSourceSHA256, x.ExternalExpected.BeforeModelDigest, x.ExternalExpected.CandidateSHA256, x.ExternalExpected.CandidateModelDigest, x.ExternalExpected.SuiteSHA256, x.BaselineSuiteSHA256, o.Baseline.SHA256, o.Candidate.SHA256}
	return Material{req, base, candidate, archive, br, cr, im["generated/lifecycle.go"], im["generated/model_paths_test.go"]}, refs
}
func copyMaterial(m Material) Material {
	p := parts(m)
	q := make([][]byte, len(p))
	for i, b := range p {
		q[i] = append([]byte{}, b...)
	}
	return Material{q[0], q[1], q[2], q[3], q[4], q[5], q[6], q[7]}
}
func TestRoundTripDeterminismAndOwnership(t *testing.T) {
	m, r := fixture(t)
	before := copyMaterial(m)
	raw, e := Build(m, r)
	if e != nil {
		t.Fatal(e)
	}
	again, e := Build(m, r)
	if e != nil || !bytes.Equal(raw, again) {
		t.Fatal("nondeterministic build", e)
	}
	receipt, e := Verify(raw, r)
	if e != nil {
		t.Fatal(e)
	}
	if receipt.Status != "caller-supplied-records" || receipt.Authority != "none" || receipt.ReportBytesStatus != ReportStatus || receipt.CarrierSHA256 != digestBytes(raw) {
		t.Fatal(receipt)
	}
	if !reflect.DeepEqual(m, before) {
		t.Fatal("input mutated")
	}
	m.Request[0] ^= 1
	if _, e = Verify(raw, r); e != nil {
		t.Fatal("output aliased input", e)
	}
	var wg sync.WaitGroup
	errs := make(chan error, 4)
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			out, e := Build(before, r)
			if e != nil || !bytes.Equal(out, raw) {
				errs <- fmt.Errorf("concurrent mismatch: %v", e)
			}
		}()
	}
	wg.Wait()
	close(errs)
	for e := range errs {
		t.Error(e)
	}
}
func TestTamperAndOpaqueReportBoundary(t *testing.T) {
	m, r := fixture(t)
	cases := map[string]func(*Material, *References){
		"request": func(x *Material, _ *References) { x.Request[0] ^= 1 }, "baseline suite": func(x *Material, _ *References) { x.BaselineSuite = append(x.BaselineSuite, ' ') }, "candidate suite": func(x *Material, _ *References) { x.CandidateSuite = append(x.CandidateSuite, ' ') },
		"archive": func(x *Material, _ *References) { x.Archive[len(x.Archive)/2] ^= 1 }, "baseline report": func(x *Material, _ *References) { x.BaselineReport[0] ^= 1 }, "candidate report": func(x *Material, _ *References) { x.CandidateReport[0] ^= 1 },
		"generated go": func(x *Material, _ *References) { x.GeneratedGo = append(x.GeneratedGo, ' ') }, "generated tests": func(x *Material, _ *References) { x.GeneratedTests = append(x.GeneratedTests, ' ') }, "external ref": func(_ *Material, y *References) { y.CandidateSHA256 = strings.Repeat("0", 64) }, "invalid ref": func(_ *Material, y *References) { y.RequestSHA256 = "BAD" },
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			x, y := copyMaterial(m), r
			mutate(&x, &y)
			if out, e := Build(x, y); e == nil || out != nil {
				t.Fatal("accepted tamper")
			}
		})
	}
	stale := copyMaterial(m)
	stale.BaselineSuite = bytes.Replace(stale.BaselineSuite, []byte(r.BeforeModelDigest), []byte(r.CandidateModelDigest), 1)
	if bytes.Equal(stale.BaselineSuite, m.BaselineSuite) {
		t.Fatal("stale mutation missed")
	}
	changed := r
	changed.BaselineSuiteSHA256 = digestBytes(stale.BaselineSuite)
	if out, e := Build(stale, changed); e == nil || out != nil {
		t.Fatal("accepted stale binding with updated byte reference")
	}
	// Reports are intentionally opaque: caller-selected bytes plus matching
	// caller reference pass. This must never be called execution validation.
	fake := copyMaterial(m)
	fake.CandidateReport = []byte("invented, not a report")
	fakeRefs := r
	fakeRefs.CandidateReportSHA256 = digestBytes(fake.CandidateReport)
	raw, e := Build(fake, fakeRefs)
	if e != nil {
		t.Fatal(e)
	}
	receipt, e := Verify(raw, fakeRefs)
	if e != nil || receipt.ReportBytesStatus != ReportStatus {
		t.Fatal("opaque boundary changed", e)
	}
}
func TestStrictCanonicalAndLimits(t *testing.T) {
	m, r := fixture(t)
	raw, e := Build(m, r)
	if e != nil {
		t.Fatal(e)
	}
	cases := map[string][]byte{
		"trailing":  append(append([]byte{}, raw...), []byte(" ")...),
		"unknown":   bytes.Replace(raw, []byte(`{"formatVersion"`), []byte(`{"unknown":1,"formatVersion"`), 1),
		"duplicate": bytes.Replace(raw, []byte(`"formatVersion":`), []byte(`"formatVersion":"x","formatVersion":`), 1),
		"case":      bytes.Replace(raw, []byte(`"formatVersion"`), []byte(`"FormatVersion"`), 1),
		"null":      []byte(`null`), "invalid utf8": append(append([]byte{}, raw...), 0xff),
		"depth": []byte(strings.Repeat("[", MaxDepth+1) + strings.Repeat("]", MaxDepth+1)),
	}
	for name, b := range cases {
		t.Run(name, func(t *testing.T) {
			if bytes.Equal(b, raw) {
				t.Fatal("mutation missed")
			}
			if receipt, e := Verify(b, r); e == nil || !reflect.DeepEqual(receipt, Receipt{}) {
				t.Fatal("accepted invalid carrier")
			}
		})
	}
	p := wire(m)
	p.Request = "YR=="
	if _, e := decodePayload(p); !errors.Is(e, ErrInvalid) {
		t.Fatalf("noncanonical pad bits: %v", e)
	}
	p = wire(m)
	p.Request = "YQ==\n"
	if _, e := decodePayload(p); !errors.Is(e, ErrInvalid) {
		t.Fatalf("newline: %v", e)
	}
	p = wire(m)
	p.Request = strings.Repeat("A", 4*((caps[0]+2)/3)+4)
	if _, e := decodePayload(p); !errors.Is(e, ErrLimit) {
		t.Fatalf("encoded limit: %v", e)
	}
	large := copyMaterial(m)
	large.GeneratedGo = make([]byte, caps[6]+1)
	if out, e := Build(large, r); !errors.Is(e, ErrLimit) || out != nil {
		t.Fatalf("typed limit: %v", e)
	}
	if _, e := Verify(make([]byte, MaxBundleBytes+1), r); !errors.Is(e, ErrLimit) {
		t.Fatalf("bundle limit: %v", e)
	}
}

func TestWirePackAndInspectBoundaries(t *testing.T) {
	m, refs := fixture(t)
	pack, e := encode(packRequest{PackVersion, refs, wire(m)})
	if e != nil {
		t.Fatal(e)
	}
	carrier, e := Pack(pack)
	if e != nil {
		t.Fatal(e)
	}
	vr, e := encode(verifyRequest{VerifyVersion, refs, base64.StdEncoding.EncodeToString(carrier)})
	if e != nil {
		t.Fatal(e)
	}
	raw, e := Inspect(vr)
	if e != nil {
		t.Fatal(e)
	}
	var receipt Receipt
	if e = json.Unmarshal(raw, &receipt); e != nil || receipt.Status != "caller-supplied-records" || receipt.ReportBytesStatus != ReportStatus {
		t.Fatal("bad wire receipt", e, receipt)
	}
	badPack := map[string][]byte{
		"missing":   []byte(`{"formatVersion":"lifecycle-rehearsal-handoff-pack/v1"}`),
		"duplicate": bytes.Replace(pack, []byte(`"formatVersion":`), []byte(`"formatVersion":"x","formatVersion":`), 1),
		"unknown":   bytes.Replace(pack, []byte(`{"formatVersion"`), []byte(`{"extra":1,"formatVersion"`), 1),
		"case":      bytes.Replace(pack, []byte(`"formatVersion"`), []byte(`"FormatVersion"`), 1),
		"trailing":  append(append([]byte{}, pack...), []byte(` {}`)...),
		"null":      []byte(`null`),
		"surrogate": bytes.Replace(pack, []byte(PackVersion), []byte(`\ud800`), 1),
	}
	for name, b := range badPack {
		t.Run(name, func(t *testing.T) {
			if bytes.Equal(b, pack) {
				t.Fatal("mutation missed")
			}
			if out, e := Pack(b); e == nil || out != nil {
				t.Fatal("accepted malformed pack")
			}
		})
	}
	badVerify := verifyRequest{VerifyVersion, refs, "YR=="}
	b, e := encode(badVerify)
	if e != nil {
		t.Fatal(e)
	}
	if out, e := Inspect(b); e == nil || out != nil {
		t.Fatal("accepted noncanonical bundle base64")
	}
	badVerify = verifyRequest{VerifyVersion, refs, ""}
	b, e = encode(badVerify)
	if e != nil {
		t.Fatal(e)
	}
	if out, e := Inspect(b); e == nil || out != nil {
		t.Fatal("accepted empty bundle")
	}
	invalidRefs := refs
	invalidRefs.RequestSHA256 = "bad"
	badVerify = verifyRequest{VerifyVersion, invalidRefs, base64.StdEncoding.EncodeToString(carrier)}
	b, e = encode(badVerify)
	if e != nil {
		t.Fatal(e)
	}
	if out, e := Inspect(b); e == nil || out != nil {
		t.Fatal("accepted invalid references")
	}
}
