package acceptanceauthor

import (
	"encoding/json"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/scenarios"
)

const BoundReportVersion = "lifecycle-bound-acceptance-report/v1"

// BoundReport records an explicit Go-supplied runner binding. Its identity is a
// caller declaration, not authenticated compiled-code provenance. It deliberately
// has no compiledModelDigest field; the original Report retains that legacy field.
type BoundReport struct {
	FormatVersion       string          `json:"formatVersion"`
	Authority           string          `json:"authority"`
	IdentitySource      string          `json:"identitySource"`
	ManifestSHA256      string          `json:"manifestSha256"`
	ManifestID          string          `json:"manifestId"`
	DeclaredAuthor      string          `json:"declaredAuthor"`
	DeclaredModelDigest string          `json:"declaredModelDigest"`
	CurrentModelDigest  string          `json:"currentModelDigest"`
	DeclaredSuiteSHA256 string          `json:"declaredSuiteSha256"`
	ObservedSuiteSHA256 string          `json:"observedSuiteSha256"`
	SuiteModelDigest    string          `json:"suiteModelDigest"`
	RunnerModelDigest   string          `json:"runnerModelDigest"`
	ModelBinding        string          `json:"modelBinding"`
	SuiteBinding        string          `json:"suiteBinding"`
	SuiteModelBinding   string          `json:"suiteModelBinding"`
	RunnerBinding       string          `json:"runnerBinding"`
	SuiteValidation     string          `json:"suiteValidation"`
	ScenarioDigest      *string         `json:"scenarioDigest"`
	Replay              string          `json:"replay"`
	ScenarioReport      json.RawMessage `json:"scenarioReport"`
}

type scenarioRunner interface {
	Decode([]byte) (scenarios.Suite, error)
	Digest(scenarios.Suite) (string, error)
	Execute([]byte) (scenarios.Report, error)
	EncodeReport(scenarios.Report) ([]byte, error)
}

// These delegates retain the original package-level binding and error policy;
// they do not opt the legacy API into the new provider protocol policy.
type defaultScenarioRunner struct{}

func (defaultScenarioRunner) Decode(raw []byte) (scenarios.Suite, error) {
	return scenarios.Decode(raw)
}
func (defaultScenarioRunner) Digest(s scenarios.Suite) (string, error) { return scenarios.Digest(s) }
func (defaultScenarioRunner) Execute(raw []byte) (scenarios.Report, error) {
	return scenarios.Execute(raw)
}
func (defaultScenarioRunner) EncodeReport(r scenarios.Report) ([]byte, error) {
	return scenarios.EncodeReport(r)
}

// CheckWithRunner uses one explicit runner for validation, identity and replay.
// A zero runner refuses; no request field can select, replace or rebind it.
// Check-only and stale inputs invoke no callback. A provider protocol failure is
// returned distinctly with a zero report, but execution may already have begun.
// Trusted callbacks may panic or perform I/O; panics propagate and no effects are
// rolled back. This API does not establish purity, independence or approval.
func CheckWithRunner(manifestBytes, modelBytes, suiteBytes []byte, mode string, runner scenarios.Runner) (BoundReport, error) {
	var zero BoundReport
	b, e := runner.ModelBinding()
	if e != nil {
		return zero, e
	}
	r, e := checkWith(manifestBytes, modelBytes, suiteBytes, mode, b.ModelDigest, runner, true)
	if e != nil {
		return zero, e
	}
	return BoundReport{
		FormatVersion: BoundReportVersion, Authority: r.Authority, IdentitySource: "caller-supplied-go-binding",
		ManifestSHA256: r.ManifestSHA256, ManifestID: r.ManifestID, DeclaredAuthor: r.DeclaredAuthor,
		DeclaredModelDigest: r.DeclaredModelDigest, CurrentModelDigest: r.CurrentModelDigest,
		DeclaredSuiteSHA256: r.DeclaredSuiteSHA256, ObservedSuiteSHA256: r.ObservedSuiteSHA256,
		SuiteModelDigest: r.SuiteModelDigest, RunnerModelDigest: r.CompiledModelDigest,
		ModelBinding: r.ModelBinding, SuiteBinding: r.SuiteBinding, SuiteModelBinding: r.SuiteModelBinding,
		RunnerBinding: r.RunnerBinding, SuiteValidation: r.SuiteValidation, ScenarioDigest: r.ScenarioDigest,
		Replay: r.Replay, ScenarioReport: r.ScenarioReport,
	}, nil
}
