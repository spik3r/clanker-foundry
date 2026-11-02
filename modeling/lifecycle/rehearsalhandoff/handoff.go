// Package rehearsalhandoff carries bounded caller-supplied records for review.
// It authenticates neither reports, execution, source provenance nor approval.
package rehearsalhandoff

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/artifactbundle"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/designchange"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/scenariodiff"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/strictjson"
)

const (
	FormatVersion   = "lifecycle-rehearsal-handoff/v1"
	PackVersion     = "lifecycle-rehearsal-handoff-pack/v1"
	VerifyVersion   = "lifecycle-rehearsal-handoff-verify/v1"
	ReceiptVersion  = "lifecycle-rehearsal-handoff-receipt/v1"
	ReportStatus    = "hash-matched-only-not-inspected"
	MaxBundleBytes  = 24 << 20
	MaxRequestBytes = 36 << 20
	MaxDecodedBytes = 16 << 20
	MaxDepth        = 16
)

var (
	ErrInvalid = errors.New("invalid rehearsal handoff")
	ErrLimit   = errors.New("rehearsal handoff limit exceeded")
)

// References are separately supplied caller assertions. They are never read
// from a carrier as authority and do not authenticate any content or actor.
type References struct {
	RequestSHA256         string `json:"requestSHA256"`
	BeforeSourceSHA256    string `json:"beforeSourceSHA256"`
	BeforeModelDigest     string `json:"beforeModelDigest"`
	CandidateSHA256       string `json:"candidateSHA256"`
	CandidateModelDigest  string `json:"candidateModelDigest"`
	CandidateSuiteSHA256  string `json:"candidateSuiteSHA256"`
	BaselineSuiteSHA256   string `json:"baselineSuiteSHA256"`
	BaselineReportSHA256  string `json:"baselineReportSHA256"`
	CandidateReportSHA256 string `json:"candidateReportSHA256"`
}

// Material contains no paths. Input bytes must remain immutable during a call.
type Material struct {
	Request, BaselineSuite, CandidateSuite, Archive              []byte
	BaselineReport, CandidateReport, GeneratedGo, GeneratedTests []byte
}
type wirePayload struct {
	Request         string `json:"requestBase64"`
	BaselineSuite   string `json:"baselineSuiteBase64"`
	CandidateSuite  string `json:"candidateSuiteBase64"`
	Archive         string `json:"archiveBase64"`
	BaselineReport  string `json:"baselineReportBase64"`
	CandidateReport string `json:"candidateReportBase64"`
	GeneratedGo     string `json:"generatedGoBase64"`
	GeneratedTests  string `json:"generatedTestsBase64"`
}
type bundleDTO struct {
	FormatVersion     string      `json:"formatVersion"`
	Mode              string      `json:"mode"`
	Authority         string      `json:"authority"`
	ReportBytesStatus string      `json:"reportBytesStatus"`
	Payload           wirePayload `json:"payload"`
}

// Receipt describes consistency checks on caller-selected bytes, not execution.
type Receipt struct {
	FormatVersion      string `json:"formatVersion"`
	Status             string `json:"status"`
	Authority          string `json:"authority"`
	ReportBytesStatus  string `json:"reportBytesStatus"`
	SuiteBindings      string `json:"suiteBindings"`
	ArchiveSuiteStatus string `json:"archiveSuiteStatus"`
	CarrierSHA256      string `json:"carrierSHA256"`
}

var caps = [...]int{1 << 20, 1 << 20, 1 << 20, 8 << 20, 2 << 20, 2 << 20, 1 << 20, 1 << 20}

func hash(b []byte) string { x := sha256.Sum256(b); return hex.EncodeToString(x[:]) }
func digest(s string) bool {
	if len(s) != 64 {
		return false
	}
	for _, c := range s {
		if !(c >= '0' && c <= '9' || c >= 'a' && c <= 'f') {
			return false
		}
	}
	return true
}
func validRefs(r References) bool {
	for _, s := range []string{r.RequestSHA256, r.BeforeSourceSHA256, r.BeforeModelDigest, r.CandidateSHA256, r.CandidateModelDigest, r.CandidateSuiteSHA256, r.BaselineSuiteSHA256, r.BaselineReportSHA256, r.CandidateReportSHA256} {
		if !digest(s) {
			return false
		}
	}
	return true
}
func parts(m Material) [][]byte {
	return [][]byte{m.Request, m.BaselineSuite, m.CandidateSuite, m.Archive, m.BaselineReport, m.CandidateReport, m.GeneratedGo, m.GeneratedTests}
}
func checkSizes(m Material) error {
	total := 0
	for i, b := range parts(m) {
		if len(b) > caps[i] {
			return ErrLimit
		}
		if len(b) == 0 {
			return ErrInvalid
		}
		total += len(b)
		if total > MaxDecodedBytes {
			return ErrLimit
		}
	}
	return nil
}
func wire(m Material) wirePayload {
	p := parts(m)
	return wirePayload{base64.StdEncoding.EncodeToString(p[0]), base64.StdEncoding.EncodeToString(p[1]), base64.StdEncoding.EncodeToString(p[2]), base64.StdEncoding.EncodeToString(p[3]), base64.StdEncoding.EncodeToString(p[4]), base64.StdEncoding.EncodeToString(p[5]), base64.StdEncoding.EncodeToString(p[6]), base64.StdEncoding.EncodeToString(p[7])}
}
func stringsOf(p wirePayload) []string {
	return []string{p.Request, p.BaselineSuite, p.CandidateSuite, p.Archive, p.BaselineReport, p.CandidateReport, p.GeneratedGo, p.GeneratedTests}
}
func decodePayload(p wirePayload) (Material, error) {
	values := stringsOf(p)
	total := 0
	// Enforce all encoded and predicted decoded budgets before bulk allocation.
	for i, s := range values {
		if s == "" {
			return Material{}, ErrInvalid
		}
		if len(s) > 4*((caps[i]+2)/3) {
			return Material{}, ErrLimit
		}
		if len(s)%4 != 0 {
			return Material{}, ErrInvalid
		}
		pad := 0
		if s[len(s)-1] == '=' {
			pad++
		}
		if len(s) > 1 && s[len(s)-2] == '=' {
			pad++
		}
		n := len(s)/4*3 - pad
		if n > caps[i] {
			return Material{}, ErrLimit
		}
		total += n
		if total > MaxDecodedBytes {
			return Material{}, ErrLimit
		}
	}
	out := make([][]byte, len(values))
	for i, s := range values {
		b, e := base64.StdEncoding.Strict().DecodeString(s)
		if e != nil || base64.StdEncoding.EncodeToString(b) != s {
			return Material{}, ErrInvalid
		}
		out[i] = b
	}
	return Material{out[0], out[1], out[2], out[3], out[4], out[5], out[6], out[7]}, nil
}
func upstream(e error) error {
	if errors.Is(e, designchange.ErrLimit) || errors.Is(e, artifactbundle.ErrLimit) || errors.Is(e, scenariodiff.ErrLimit) {
		return ErrLimit
	}
	return ErrInvalid
}
func fileMap(files []artifactbundle.File) map[string][]byte {
	m := make(map[string][]byte, len(files))
	for _, f := range files {
		m[f.Name] = f.Data
	}
	return m
}
func encode(v any) ([]byte, error) {
	var b bytes.Buffer
	e := json.NewEncoder(&b)
	e.SetEscapeHTML(false)
	if e.Encode(v) != nil {
		return nil, ErrInvalid
	}
	return b.Bytes(), nil
}

func validate(m Material, r References) error {
	if e := checkSizes(m); e != nil {
		return e
	}
	if !validRefs(r) {
		return ErrInvalid
	}
	if hash(m.Request) != r.RequestSHA256 || hash(m.BaselineSuite) != r.BaselineSuiteSHA256 || hash(m.CandidateSuite) != r.CandidateSuiteSHA256 || hash(m.BaselineReport) != r.BaselineReportSHA256 || hash(m.CandidateReport) != r.CandidateReportSHA256 {
		return ErrInvalid
	}
	expected := designchange.Expected{RequestSHA256: r.RequestSHA256, BeforeSourceSHA256: r.BeforeSourceSHA256, BeforeModelDigest: r.BeforeModelDigest, CandidateSHA256: r.CandidateSHA256, CandidateModelDigest: r.CandidateModelDigest, SuiteSHA256: r.CandidateSuiteSHA256}
	if e := designchange.Verify(m.Archive, expected); e != nil {
		return upstream(e)
	}
	outerFiles, e := artifactbundle.Verify(m.Archive, artifactbundle.DesignChangeProfile)
	if e != nil {
		return upstream(e)
	}
	outer := fileMap(outerFiles)
	innerFiles, e := artifactbundle.Verify(outer["qualification.zip"], artifactbundle.QualificationProfile)
	if e != nil {
		return upstream(e)
	}
	inner := fileMap(innerFiles)
	if !bytes.Equal(m.Request, outer["change-request.json"]) || !bytes.Equal(m.CandidateSuite, inner["suite.json"]) || !bytes.Equal(m.GeneratedGo, inner["generated/lifecycle.go"]) || !bytes.Equal(m.GeneratedTests, inner["generated/model_paths_test.go"]) {
		return ErrInvalid
	}
	// Reuse the existing declaration-comparison preflight; it validates each
	// suite against its actual verified model without invoking a reducer.
	request := struct {
		FormatVersion string          `json:"formatVersion"`
		BeforeModel   json.RawMessage `json:"beforeModel"`
		BeforeSuite   json.RawMessage `json:"beforeSuite"`
		AfterModel    json.RawMessage `json:"afterModel"`
		AfterSuite    json.RawMessage `json:"afterSuite"`
	}{scenariodiff.FormatVersion, outer["before.json"], m.BaselineSuite, outer["candidate.json"], m.CandidateSuite}
	b, e := encode(request)
	if e != nil {
		return e
	}
	if _, e = scenariodiff.Execute(b); e != nil {
		return upstream(e)
	}
	return nil
}

// Build checks independently supplied references and returns canonical carrier
// bytes. Reports are opaque hash-matched records, not validated executions.
func Build(m Material, r References) ([]byte, error) {
	if e := validate(m, r); e != nil {
		return nil, e
	}
	raw, e := encode(bundleDTO{FormatVersion, "example-only", "none", ReportStatus, wire(m)})
	if e != nil {
		return nil, e
	}
	if len(raw) > MaxBundleBytes {
		return nil, ErrLimit
	}
	return raw, nil
}

// Verify checks the entire carrier and separately supplied references. No report
// content is parsed. The returned receipt does not authenticate any assertion.
func Verify(raw []byte, r References) (Receipt, error) {
	if len(raw) > MaxBundleBytes {
		return Receipt{}, ErrLimit
	}
	var d bundleDTO
	if strictjson.Decode(raw, &d, MaxBundleBytes, MaxDepth) != nil || d.FormatVersion != FormatVersion || d.Mode != "example-only" || d.Authority != "none" || d.ReportBytesStatus != ReportStatus {
		return Receipt{}, ErrInvalid
	}
	if !validRefs(r) {
		return Receipt{}, ErrInvalid
	}
	// Reject noncanonical carriers before nested archive processing.
	encoded, e := encode(d)
	if e != nil || !bytes.Equal(raw, encoded) {
		return Receipt{}, ErrInvalid
	}
	m, e := decodePayload(d.Payload)
	if e != nil {
		return Receipt{}, e
	}
	canonical, e := Build(m, r)
	if e != nil {
		return Receipt{}, e
	}
	if !bytes.Equal(raw, canonical) {
		return Receipt{}, ErrInvalid
	}
	return Receipt{ReceiptVersion, "caller-supplied-records", "none", ReportStatus, "consistent-with-verified-models", "unvalidated and unexecuted; binding not inspected by archive", hash(raw)}, nil
}

type packRequest struct {
	FormatVersion string      `json:"formatVersion"`
	References    References  `json:"references"`
	Payload       wirePayload `json:"payload"`
}
type verifyRequest struct {
	FormatVersion string     `json:"formatVersion"`
	References    References `json:"references"`
	Bundle        string     `json:"bundleBase64"`
}

// Pack decodes strict bounded JSON request and builds one canonical carrier.
func Pack(raw []byte) ([]byte, error) {
	if len(raw) > MaxRequestBytes {
		return nil, ErrLimit
	}
	var d packRequest
	if strictjson.Decode(raw, &d, MaxRequestBytes, MaxDepth) != nil || d.FormatVersion != PackVersion {
		return nil, ErrInvalid
	}
	if !validRefs(d.References) {
		return nil, ErrInvalid
	}
	m, e := decodePayload(d.Payload)
	if e != nil {
		return nil, e
	}
	return Build(m, d.References)
}

// Inspect decodes strict bounded JSON request, verifies the carrier, and emits
// one canonical receipt. External references remain caller-supplied assertions.
func Inspect(raw []byte) ([]byte, error) {
	if len(raw) > MaxRequestBytes {
		return nil, ErrLimit
	}
	var d verifyRequest
	if strictjson.Decode(raw, &d, MaxRequestBytes, MaxDepth) != nil || d.FormatVersion != VerifyVersion {
		return nil, ErrInvalid
	}
	if !validRefs(d.References) || d.Bundle == "" {
		return nil, ErrInvalid
	}
	if len(d.Bundle) > 4*((MaxBundleBytes+2)/3) {
		return nil, ErrLimit
	}
	if len(d.Bundle)%4 != 0 {
		return nil, ErrInvalid
	}
	pad := 0
	if d.Bundle[len(d.Bundle)-1] == '=' {
		pad++
	}
	if len(d.Bundle) > 1 && d.Bundle[len(d.Bundle)-2] == '=' {
		pad++
	}
	if len(d.Bundle)/4*3-pad > MaxBundleBytes {
		return nil, ErrLimit
	}
	b, e := base64.StdEncoding.Strict().DecodeString(d.Bundle)
	if e != nil || base64.StdEncoding.EncodeToString(b) != d.Bundle {
		return nil, ErrInvalid
	}
	if len(b) > MaxBundleBytes {
		return nil, ErrLimit
	}
	receipt, e := Verify(b, d.References)
	if e != nil {
		return nil, e
	}
	return encode(receipt)
}
