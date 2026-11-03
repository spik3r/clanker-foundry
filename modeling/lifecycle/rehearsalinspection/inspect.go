// Package rehearsalinspection inspects caller-supplied reports, not execution.
package rehearsalinspection

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"reflect"

	lifecycle "github.com/spik3r/clanker-foundry/modeling/lifecycle/generated"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/rehearsalhandoff"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/scenarios"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/strictjson"
)

const (
	FormatVersion   = "lifecycle-rehearsal-inspection/v1"
	RequestVersion  = "lifecycle-rehearsal-inspection-request/v1"
	MaxRequestBytes = rehearsalhandoff.MaxRequestBytes
)

var (
	ErrInvalid = errors.New("invalid rehearsal report inspection")
	ErrLimit   = errors.New("rehearsal report inspection limit exceeded")
)

type Observation struct {
	ModelDigest    string `json:"modelDigest"`
	ScenarioDigest string `json:"scenarioDigest"`
	SourceSHA256   string `json:"sourceSHA256"`
	Matched        bool   `json:"matched"`
	CasesCompleted int    `json:"casesCompleted"`
	StepsExecuted  int    `json:"stepsExecuted"`
}
type Receipt struct {
	FormatVersion   string      `json:"formatVersion"`
	Mode            string      `json:"mode"`
	Authority       string      `json:"authority"`
	ExecutionStatus string      `json:"executionStatus"`
	ReportStatus    string      `json:"reportStatus"`
	Baseline        Observation `json:"baseline"`
	Candidate       Observation `json:"candidate"`
}

func hash(b []byte) string { d := sha256.Sum256(b); return hex.EncodeToString(d[:]) }
func inspectReport(suiteBytes, reportBytes []byte, modelDigest string) (Observation, error) {
	// The callback exists only to supply the existing validation API's context.
	// This code uses Decode/Digest/DecodeReport and never Run or Execute.
	runner, err := scenarios.NewRunner(scenarios.Binding{ProfileVersion: lifecycle.ProfileVersion, ModelDigest: modelDigest}, func(lifecycle.Task, lifecycle.Command) lifecycle.Result { panic("inspection must not invoke reducer") })
	if err != nil {
		return Observation{}, ErrInvalid
	}
	suite, err := runner.Decode(suiteBytes)
	if err != nil {
		return Observation{}, ErrInvalid
	}
	digest, err := runner.Digest(suite)
	if err != nil {
		return Observation{}, ErrInvalid
	}
	report, err := runner.DecodeReport(reportBytes)
	if err != nil {
		return Observation{}, ErrInvalid
	}
	source := hash(suiteBytes)
	if report.ModelDigest != modelDigest || report.ScenarioDigest != digest || report.SourceDigest == nil || *report.SourceDigest != source {
		return Observation{}, ErrInvalid
	}
	index, completed := 0, 0
	for _, c := range suite.Cases {
		current := c.Initial
		for _, step := range c.Steps {
			if index >= len(report.Entries) {
				return Observation{}, ErrInvalid
			}
			e := report.Entries[index]
			if e.CaseID != c.ID || e.StepID != step.ID || !reflect.DeepEqual(e.Input, current) || !reflect.DeepEqual(e.Command, step.Command) || !reflect.DeepEqual(e.Expected, step.Expected) {
				return Observation{}, ErrInvalid
			}
			index++
			if !e.Matched {
				if report.Matched || index != len(report.Entries) || report.CasesCompleted != completed {
					return Observation{}, ErrInvalid
				}
				return observation(report, source), nil
			}
			if e.Actual.OK {
				current = *e.Actual.Task
			}
		}
		completed++
	}
	if index != len(report.Entries) || !report.Matched || report.CasesCompleted != completed {
		return Observation{}, ErrInvalid
	}
	return observation(report, source), nil
}
func observation(r scenarios.Report, source string) Observation {
	return Observation{r.ModelDigest, r.ScenarioDigest, source, r.Matched, r.CasesCompleted, r.StepsExecuted}
}

// Inspect first verifies the complete carrier, then binds each native report to
// its original suite literals. It never runs code or authenticates observations.
// Input slices must remain immutable during the call.
func Inspect(carrier []byte, refs rehearsalhandoff.References) (Receipt, error) {
	material, _, err := rehearsalhandoff.InspectMaterial(carrier, refs)
	if err != nil {
		if errors.Is(err, rehearsalhandoff.ErrLimit) {
			return Receipt{}, ErrLimit
		}
		return Receipt{}, ErrInvalid
	}
	baseline, err := inspectReport(material.BaselineSuite, material.BaselineReport, refs.BeforeModelDigest)
	if err != nil {
		return Receipt{}, err
	}
	candidate, err := inspectReport(material.CandidateSuite, material.CandidateReport, refs.CandidateModelDigest)
	if err != nil {
		return Receipt{}, err
	}
	return Receipt{FormatVersion, "example-only", "none", "not-established", "inspected-for-consistency", baseline, candidate}, nil
}

type request struct {
	FormatVersion string                      `json:"formatVersion"`
	References    rehearsalhandoff.References `json:"references"`
	BundleBase64  string                      `json:"bundleBase64"`
}

// InspectRequest accepts strict bounded JSON, never filesystem paths. The base64
// carrier and external references have the same meaning as the legacy verifier.
func InspectRequest(raw []byte) ([]byte, error) {
	if len(raw) > MaxRequestBytes {
		return nil, ErrLimit
	}
	var r request
	if strictjson.Decode(raw, &r, MaxRequestBytes, rehearsalhandoff.MaxDepth) != nil || r.FormatVersion != RequestVersion {
		return nil, ErrInvalid
	}
	if len(r.BundleBase64) > base64.StdEncoding.EncodedLen(rehearsalhandoff.MaxBundleBytes) {
		return nil, ErrLimit
	}
	if r.BundleBase64 == "" || len(r.BundleBase64)%4 != 0 {
		return nil, ErrInvalid
	}
	decodedLen := len(r.BundleBase64) / 4 * 3
	if r.BundleBase64[len(r.BundleBase64)-1] == '=' {
		decodedLen--
	}
	if len(r.BundleBase64) > 1 && r.BundleBase64[len(r.BundleBase64)-2] == '=' {
		decodedLen--
	}
	if decodedLen > rehearsalhandoff.MaxBundleBytes {
		return nil, ErrLimit
	}
	carrier, err := base64.StdEncoding.Strict().DecodeString(r.BundleBase64)
	if err != nil || base64.StdEncoding.EncodeToString(carrier) != r.BundleBase64 {
		return nil, ErrInvalid
	}
	receipt, err := Inspect(carrier, r.References)
	if err != nil {
		return nil, err
	}
	var out bytes.Buffer
	enc := json.NewEncoder(&out)
	enc.SetEscapeHTML(false)
	if enc.Encode(receipt) != nil {
		return nil, ErrInvalid
	}
	return out.Bytes(), nil
}
