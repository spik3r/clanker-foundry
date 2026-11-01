// Package acceptanceauthor checks frozen declarations without deriving expected outcomes.
package acceptanceauthor

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/canonical"
	lifecycle "github.com/spik3r/clanker-foundry/modeling/lifecycle/generated"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/jsonformat"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/model"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/scenarios"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/strictjson"
)

const FormatVersion = "lifecycle-acceptance-authoring/v1"
const ReportVersion = "lifecycle-acceptance-authoring-report/v1"
const RequestVersion = "lifecycle-acceptance-authoring-request/v1"
const MaxRequestBytes = 8 << 20
const MaxManifestBytes = 64 << 10
const MaxReportBytes = (9 << 20)

var ErrInvalid = errors.New("invalid acceptance authoring input")

type caseIntent struct {
	ID     string `json:"id"`
	Intent string `json:"intent"`
}
type manifest struct {
	FormatVersion     string       `json:"formatVersion"`
	ID                string       `json:"id"`
	Author            string       `json:"author"`
	Purpose           string       `json:"purpose"`
	ModelDigest       string       `json:"modelDigest"`
	SuiteSourceSHA256 string       `json:"suiteSourceSha256"`
	Cases             []caseIntent `json:"cases"`
}
type Request struct {
	FormatVersion string `json:"formatVersion"`
	Mode          string `json:"mode"`
	Manifest      string `json:"manifest"`
	Model         string `json:"model"`
	Suite         string `json:"suite"`
}
type Report struct {
	FormatVersion       string          `json:"formatVersion"`
	Authority           string          `json:"authority"`
	ManifestSHA256      string          `json:"manifestSha256"`
	ManifestID          string          `json:"manifestId"`
	DeclaredAuthor      string          `json:"declaredAuthor"`
	DeclaredModelDigest string          `json:"declaredModelDigest"`
	CurrentModelDigest  string          `json:"currentModelDigest"`
	DeclaredSuiteSHA256 string          `json:"declaredSuiteSha256"`
	ObservedSuiteSHA256 string          `json:"observedSuiteSha256"`
	SuiteModelDigest    string          `json:"suiteModelDigest"`
	CompiledModelDigest string          `json:"compiledModelDigest"`
	ModelBinding        string          `json:"modelBinding"`
	SuiteBinding        string          `json:"suiteBinding"`
	SuiteModelBinding   string          `json:"suiteModelBinding"`
	RunnerBinding       string          `json:"runnerBinding"`
	SuiteValidation     string          `json:"suiteValidation"`
	ScenarioDigest      *string         `json:"scenarioDigest"`
	Replay              string          `json:"replay"`
	ScenarioReport      json.RawMessage `json:"scenarioReport"`
}

func hash(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }
func validHash(s string) bool {
	if len(s) != 64 {
		return false
	}
	for _, c := range []byte(s) {
		if !(c >= '0' && c <= '9' || c >= 'a' && c <= 'f') {
			return false
		}
	}
	return true
}
func nonblank(s string, n int) bool { return len(s) <= n && strings.TrimSpace(s) != "" }
func binding(a, b string) string {
	if a == b {
		return "current"
	}
	return "stale"
}

// Check validates the manifest and model, compares exact suite bytes and suite
// model header, and fully validates the suite only when all bindings are current.
// Replay delegates to the existing runner; it never derives or rewrites expectations.
func Check(manifestBytes, modelBytes, suiteBytes []byte, mode string) (Report, error) {
	var zero Report
	var m manifest
	if (mode != "check" && mode != "replay") || strictjson.Decode(manifestBytes, &m, MaxManifestBytes, 64) != nil || m.FormatVersion != FormatVersion || !model.Identifier(m.ID) || !nonblank(m.Author, 256) || !nonblank(m.Purpose, 4096) || !validHash(m.ModelDigest) || !validHash(m.SuiteSourceSHA256) || len(m.Cases) < 1 || len(m.Cases) > scenarios.MaxCases {
		return zero, ErrInvalid
	}
	seen := map[string]bool{}
	for _, c := range m.Cases {
		if !model.Identifier(c.ID) || seen[c.ID] || !nonblank(c.Intent, 4096) {
			return zero, ErrInvalid
		}
		seen[c.ID] = true
	}
	ir, e := jsonformat.Decode(modelBytes)
	if e != nil {
		return zero, ErrInvalid
	}
	current, e := canonical.Digest(ir)
	if e != nil {
		return zero, ErrInvalid
	}
	// Inspect only the binding header before full scenario validation. A stale
	// suite remains explicitly not checked, even if its header is well formed.
	var fields map[string]json.RawMessage
	if strictjson.Decode(suiteBytes, &fields, scenarios.MaxSourceBytes, scenarios.MaxDepth) != nil {
		return zero, ErrInvalid
	}
	var suiteModel string
	if strictjson.Decode(fields["modelDigest"], &suiteModel, 128, 2) != nil || !validHash(suiteModel) {
		return zero, ErrInvalid
	}
	observed := hash(suiteBytes)
	r := Report{FormatVersion: ReportVersion, Authority: "none", ManifestSHA256: hash(manifestBytes), ManifestID: m.ID, DeclaredAuthor: m.Author, DeclaredModelDigest: m.ModelDigest, CurrentModelDigest: current, DeclaredSuiteSHA256: m.SuiteSourceSHA256, ObservedSuiteSHA256: observed, SuiteModelDigest: suiteModel, CompiledModelDigest: lifecycle.ModelDigest, ModelBinding: binding(m.ModelDigest, current), SuiteBinding: binding(m.SuiteSourceSHA256, observed), SuiteModelBinding: binding(suiteModel, m.ModelDigest), RunnerBinding: binding(current, lifecycle.ModelDigest), SuiteValidation: "not-checked", Replay: "not-run"}
	if r.ModelBinding != "current" || r.SuiteBinding != "current" || r.SuiteModelBinding != "current" || r.RunnerBinding != "current" {
		return r, nil
	}
	suite, e := scenarios.Decode(suiteBytes)
	if e != nil {
		return zero, ErrInvalid
	}
	if len(suite.Cases) != len(m.Cases) {
		return zero, ErrInvalid
	}
	for _, c := range suite.Cases {
		if !seen[c.ID] {
			return zero, ErrInvalid
		}
	}
	sd, e := scenarios.Digest(suite)
	if e != nil {
		return zero, ErrInvalid
	}
	r.ScenarioDigest = &sd
	r.SuiteValidation = "valid"
	if mode == "replay" {
		observedReport, e := scenarios.Execute(suiteBytes)
		if e != nil {
			return zero, ErrInvalid
		}
		b, e := scenarios.EncodeReport(observedReport)
		if e != nil {
			return zero, ErrInvalid
		}
		r.ScenarioReport = json.RawMessage(bytes.TrimSpace(b))
		if observedReport.Matched {
			r.Replay = "matched"
		} else {
			r.Replay = "mismatched"
		}
	}
	return r, nil
}

// Execute receives literal UTF-8 contents as JSON strings and buffers the report.
func Execute(raw []byte) ([]byte, error) {
	var q Request
	if strictjson.Decode(raw, &q, MaxRequestBytes, 64) != nil || q.FormatVersion != RequestVersion {
		return nil, ErrInvalid
	}
	r, e := Check([]byte(q.Manifest), []byte(q.Model), []byte(q.Suite), q.Mode)
	if e != nil {
		return nil, e
	}
	var out bytes.Buffer
	enc := json.NewEncoder(&out)
	enc.SetEscapeHTML(false)
	if enc.Encode(r) != nil || out.Len() > MaxReportBytes {
		return nil, ErrInvalid
	}
	return out.Bytes(), nil
}
