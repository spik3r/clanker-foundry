// Package scenariotrace links declared requirements to actual suite cases without
// executing a reducer or establishing that a requirement is satisfied.
package scenariotrace

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"sort"

	lifecycle "github.com/spik3r/clanker-foundry/modeling/lifecycle/generated"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/model"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/scenarios"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/strictjson"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/traceability"
)

const (
	RequestVersion         = "lifecycle-scenario-trace-request/v1"
	BindingsVersion        = "lifecycle-scenario-trace-bindings/v1"
	ReportVersion          = "lifecycle-scenario-trace-report/v1"
	MaxRequestBytes        = 32 << 20
	MaxOutputBytes         = 64 << 20
	maxModelBytes          = 1 << 20
	maxRequirementsBytes   = 1 << 20
	maxCatalogBytes        = 4 << 20
	maxSuiteBytes          = 1 << 20
	maxBindingsBytes       = 512 << 10
	maxScenarioReportBytes = 8 << 20
	maxTraceBytes          = 4 << 20
	maxLinks               = 4096
)

var (
	ErrInvalid  = errors.New("invalid scenario trace input")
	ErrBinding  = errors.New("scenario trace binding mismatch")
	ErrLimit    = errors.New("scenario trace limit exceeded")
	ErrBoundary = errors.New("scenario trace boundary failed")
)

type request struct {
	FormatVersion  string  `json:"formatVersion"`
	Model          string  `json:"model"`
	Requirements   string  `json:"requirements"`
	Catalog        string  `json:"catalog"`
	Suite          string  `json:"suite"`
	Bindings       string  `json:"bindings"`
	ScenarioReport *string `json:"scenarioReport"`
}
type bindings struct {
	FormatVersion      string `json:"formatVersion"`
	ModelDigest        string `json:"modelDigest"`
	RequirementsSHA256 string `json:"requirementsSHA256"`
	CatalogSHA256      string `json:"catalogSHA256"`
	SuiteSHA256        string `json:"suiteSHA256"`
	Links              []link `json:"links"`
}
type link struct {
	RequirementID  string `json:"requirementId"`
	ScenarioCaseID string `json:"scenarioCaseId"`
}

func hash(s string) string { h := sha256.Sum256([]byte(s)); return hex.EncodeToString(h[:]) }
func digestValid(s string) bool {
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

// Check retains exact source hashes and native trace text. A supplied scenario
// report is inspected for consistency only; it may be fabricated. The function
// performs no I/O and never invokes a reducer. Errors return no partial output.
func Check(raw []byte) ([]byte, error) {
	if len(raw) > MaxRequestBytes {
		return nil, ErrLimit
	}
	var q request
	if strictjson.Decode(raw, &q, MaxRequestBytes, 4) != nil || q.FormatVersion != RequestVersion {
		return nil, ErrInvalid
	}
	if len(q.Model) > maxModelBytes || len(q.Requirements) > maxRequirementsBytes || len(q.Catalog) > maxCatalogBytes || len(q.Suite) > maxSuiteBytes || len(q.Bindings) > maxBindingsBytes || q.ScenarioReport != nil && len(*q.ScenarioReport) > maxScenarioReportBytes {
		return nil, ErrLimit
	}
	var b bindings
	if strictjson.Decode([]byte(q.Bindings), &b, maxBindingsBytes, 4) != nil || b.FormatVersion != BindingsVersion {
		return nil, ErrInvalid
	}
	if len(b.Links) > maxLinks {
		return nil, ErrLimit
	}
	for _, d := range []string{b.ModelDigest, b.RequirementsSHA256, b.CatalogSHA256, b.SuiteSHA256} {
		if !digestValid(d) {
			return nil, ErrInvalid
		}
	}
	tr, err := traceability.Build([]byte(q.Model), []byte(q.Catalog), []byte(q.Requirements))
	if err != nil {
		return nil, ErrInvalid
	}
	// This runner is a non-executing validation context, not a claim about a
	// generated executable or provider. Any accidental execution is a bug.
	runner, err := scenarios.NewRunner(scenarios.Binding{ProfileVersion: model.ProfileVersion, ModelDigest: tr.Inputs.ModelDigest}, func(lifecycle.Task, lifecycle.Command) lifecycle.Result {
		panic("scenario trace validation invoked reducer")
	})
	if err != nil {
		return nil, ErrBoundary
	}
	suite, err := runner.Decode([]byte(q.Suite))
	if err != nil {
		return nil, ErrInvalid
	}
	scenarioDigest, err := runner.Digest(suite)
	if err != nil {
		return nil, ErrBoundary
	}
	in := inputs{ModelSHA256: hash(q.Model), ModelDigest: tr.Inputs.ModelDigest, RequirementsSHA256: hash(q.Requirements), RequirementsDigest: tr.Inputs.RequirementsDigest, CatalogSHA256: hash(q.Catalog), SuiteSHA256: hash(q.Suite), ScenarioDigest: scenarioDigest, BindingsSHA256: hash(q.Bindings)}
	if b.ModelDigest != in.ModelDigest || b.RequirementsSHA256 != in.RequirementsSHA256 || b.CatalogSHA256 != in.CatalogSHA256 || b.SuiteSHA256 != in.SuiteSHA256 {
		return nil, ErrBinding
	}
	requirements := make(map[string][]string, len(tr.Requirements))
	cases := make(map[string][]string, len(suite.Cases))
	for _, r := range tr.Requirements {
		requirements[r.ID] = []string{}
	}
	for _, c := range suite.Cases {
		cases[c.ID] = []string{}
	}
	seen := make(map[link]bool, len(b.Links))
	for _, l := range b.Links {
		_, reqOK := requirements[l.RequirementID]
		_, caseOK := cases[l.ScenarioCaseID]
		if !reqOK || !caseOK || seen[l] {
			return nil, ErrInvalid
		}
		seen[l] = true
		requirements[l.RequirementID] = append(requirements[l.RequirementID], l.ScenarioCaseID)
		cases[l.ScenarioCaseID] = append(cases[l.ScenarioCaseID], l.RequirementID)
	}
	sort.Slice(b.Links, func(i, j int) bool {
		if b.Links[i].RequirementID != b.Links[j].RequirementID {
			return b.Links[i].RequirementID < b.Links[j].RequirementID
		}
		return b.Links[i].ScenarioCaseID < b.Links[j].ScenarioCaseID
	})
	tb, err := tr.JSON()
	if err != nil {
		return nil, ErrBoundary
	}
	if len(tb) > maxTraceBytes {
		return nil, ErrLimit
	}
	r := newReport(in, string(tb), b.Links, requirements, cases, suite)
	if q.ScenarioReport != nil {
		if err := bindReport(runner, suite, *q.ScenarioReport, &r); err != nil {
			return nil, err
		}
	}
	return encodeReport(r)
}
