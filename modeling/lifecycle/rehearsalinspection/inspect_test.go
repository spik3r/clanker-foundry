package rehearsalinspection

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"
	"time"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/artifactbundle"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/designchange"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/qualification"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/rehearsalhandoff"
)

func read(t *testing.T, path string) []byte {
	t.Helper()
	b, e := os.ReadFile(path)
	if e != nil {
		t.Fatal(e)
	}
	return b
}
func compact(t *testing.T, raw json.RawMessage) []byte {
	t.Helper()
	var b bytes.Buffer
	if json.Compact(&b, raw) != nil {
		t.Fatal("invalid fixture")
	}
	return append(b.Bytes(), '\n')
}
func fixture(t *testing.T) (rehearsalhandoff.Material, rehearsalhandoff.References) {
	t.Helper()
	dir := ".."
	var x struct {
		ExternalExpected    struct{ RequestSHA256, BeforeSourceSHA256, BeforeModelDigest, CandidateSHA256, CandidateModelDigest, SuiteSHA256 string }
		BaselineSuiteSHA256 string
	}
	if e := json.Unmarshal(read(t, filepath.Join(dir, "rehearsal/testdata/expected.json")), &x); e != nil {
		t.Fatal(e)
	}
	var o struct {
		Baseline, Candidate struct {
			SHA256 string          `json:"sha256"`
			Report json.RawMessage `json:"report"`
		}
	}
	if e := json.Unmarshal(read(t, filepath.Join(dir, "rehearsalhandoff/testdata/report-identities.json")), &o); e != nil {
		t.Fatal(e)
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
	files, e := artifactbundle.Verify(archive, artifactbundle.DesignChangeProfile)
	if e != nil {
		t.Fatal(e)
	}
	var nested []byte
	for _, f := range files {
		if f.Name == "qualification.zip" {
			nested = f.Data
		}
	}
	inner, e := artifactbundle.Verify(nested, artifactbundle.QualificationProfile)
	if e != nil {
		t.Fatal(e)
	}
	var goBytes, testBytes []byte
	for _, f := range inner {
		if f.Name == "generated/lifecycle.go" {
			goBytes = f.Data
		}
		if f.Name == "generated/model_paths_test.go" {
			testBytes = f.Data
		}
	}
	refs := rehearsalhandoff.References{RequestSHA256: x.ExternalExpected.RequestSHA256, BeforeSourceSHA256: x.ExternalExpected.BeforeSourceSHA256, BeforeModelDigest: x.ExternalExpected.BeforeModelDigest, CandidateSHA256: x.ExternalExpected.CandidateSHA256, CandidateModelDigest: x.ExternalExpected.CandidateModelDigest, CandidateSuiteSHA256: x.ExternalExpected.SuiteSHA256, BaselineSuiteSHA256: x.BaselineSuiteSHA256, BaselineReportSHA256: o.Baseline.SHA256, CandidateReportSHA256: o.Candidate.SHA256}
	return rehearsalhandoff.Material{Request: req, BaselineSuite: base, CandidateSuite: candidate, Archive: archive, BaselineReport: compact(t, o.Baseline.Report), CandidateReport: compact(t, o.Candidate.Report), GeneratedGo: goBytes, GeneratedTests: testBytes}, refs
}
func carrier(t *testing.T, m rehearsalhandoff.Material, r rehearsalhandoff.References) []byte {
	t.Helper()
	b, e := rehearsalhandoff.Build(m, r)
	if e != nil {
		t.Fatal(e)
	}
	return b
}
func wire(t *testing.T, b []byte, r rehearsalhandoff.References) []byte {
	t.Helper()
	raw, e := json.Marshal(map[string]any{"formatVersion": RequestVersion, "references": r, "bundleBase64": base64.StdEncoding.EncodeToString(b)})
	if e != nil {
		t.Fatal(e)
	}
	return raw
}
func TestRequestAndOwnership(t *testing.T) {
	m, r := fixture(t)
	b := carrier(t, m, r)
	before := append([]byte{}, b...)
	receipt, e := Inspect(b, r)
	if e != nil {
		t.Fatal(e)
	}
	if !reflect.DeepEqual(b, before) {
		t.Fatal("input mutated")
	}
	raw, e := InspectRequest(wire(t, b, r))
	if e != nil {
		t.Fatal(e)
	}
	var got Receipt
	if json.Unmarshal(raw, &got) != nil || !reflect.DeepEqual(got, receipt) {
		t.Fatal("request receipt changed")
	}
	for name, raw := range map[string][]byte{"empty": nil, "unknown": []byte(`{"formatVersion":"x","extra":1}`), "duplicate": []byte(`{"formatVersion":"x","formatVersion":"x"}`), "null": []byte(`null`), "limit": bytes.Repeat([]byte(" "), MaxRequestBytes+1)} {
		t.Run(name, func(t *testing.T) {
			if out, e := InspectRequest(raw); e == nil || out != nil {
				t.Fatal("invalid request accepted")
			}
		})
	}
	for _, bad := range []string{"?", base64.StdEncoding.EncodeToString(b) + "\n"} {
		req, _ := json.Marshal(map[string]any{"formatVersion": RequestVersion, "references": r, "bundleBase64": bad})
		if _, e := InspectRequest(req); e == nil {
			t.Fatal("noncanonical base64 accepted")
		}
	}
}
func TestOpaqueLegacyCarrierStillVerifies(t *testing.T) {
	m, r := fixture(t)
	m.CandidateReport = []byte("an opaque invented report")
	r.CandidateReportSHA256 = hash(m.CandidateReport)
	b := carrier(t, m, r)
	if _, e := rehearsalhandoff.Verify(b, r); e != nil {
		t.Fatal(e)
	}
	if _, e := Inspect(b, r); e == nil {
		t.Fatal("opaque report accepted")
	}
}

// The subprocess compiles only the trusted CLI; inspection invokes no reducer.
func TestCLIRealStdoutAndClosedPipe(t *testing.T) {
	m, r := fixture(t)
	request := wire(t, carrier(t, m, r), r)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	binary := filepath.Join(t.TempDir(), "modelrehearsalinspect")
	build := exec.CommandContext(ctx, filepath.Join(runtime.GOROOT(), "bin", "go"), "build", "-o", binary, "./cmd/modelrehearsalinspect")
	build.Dir = ".."
	if out, e := build.CombinedOutput(); e != nil {
		t.Fatalf("build: %v %s", e, out)
	}
	cmd := exec.CommandContext(ctx, binary)
	cmd.Stdin = bytes.NewReader(request)
	out, e := cmd.Output()
	if e != nil {
		t.Fatal(e)
	}
	expected, e := InspectRequest(request)
	if e != nil || !bytes.Equal(out, expected) {
		t.Fatal("CLI receipt mismatch")
	}
	reader, writer, e := os.Pipe()
	if e != nil {
		t.Fatal(e)
	}
	defer writer.Close()
	if e = reader.Close(); e != nil {
		t.Fatal(e)
	}
	cmd = exec.CommandContext(ctx, binary)
	cmd.Stdin = bytes.NewReader(request)
	cmd.Stdout = writer
	cmd.WaitDelay = time.Second
	var diagnostic bytes.Buffer
	cmd.Stderr = &diagnostic
	e = cmd.Run()
	var exit *exec.ExitError
	if !errors.As(e, &exit) || exit.ExitCode() != 3 || diagnostic.String() != "inspection output failed\n" {
		t.Fatalf("closed stdout: %v %q", e, diagnostic.String())
	}
}
