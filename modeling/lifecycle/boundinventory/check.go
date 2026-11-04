// Package boundinventory inventories supplied assertions for an exact custom
// model and suite. It executes no reducer and establishes no behavior coverage.
package boundinventory

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"reflect"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/canonical"
	g "github.com/spik3r/clanker-foundry/modeling/lifecycle/generated"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/jsonformat"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/model"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/scenarioexplain"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/scenarios"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/strictjson"
)

const (
	RequestVersion  = "lifecycle-bound-assertion-inventory-request/v1"
	ReportVersion   = "lifecycle-bound-assertion-inventory/v1"
	MaxRequestBytes = 16 << 20
	MaxOutputBytes  = 8 << 20
	maxModelBytes   = 1 << 20
)

var (
	ErrInvalid  = errors.New("invalid bound assertion inventory input")
	ErrBinding  = errors.New("bound assertion inventory identity mismatch")
	ErrLimit    = errors.New("bound assertion inventory limit exceeded")
	ErrBoundary = errors.New("bound assertion inventory boundary failure")
)

type request struct {
	FormatVersion string `json:"formatVersion"`
	Model         string `json:"model"`
	Suite         string `json:"suite"`
	Report        string `json:"report"`
}

func hash(text string) string { d := sha256.Sum256([]byte(text)); return hex.EncodeToString(d[:]) }

// Check validates original source identities, all native result declarations and
// exact suite literals before mapping any assertions. No callback is invoked.
// Caller input must remain immutable during the call; returned bytes are owned.
func Check(raw []byte) ([]byte, error) {
	if len(raw) > MaxRequestBytes {
		return nil, ErrLimit
	}
	var q request
	if strictjson.Decode(raw, &q, MaxRequestBytes, 4) != nil || q.FormatVersion != RequestVersion {
		return nil, ErrInvalid
	}
	if len(q.Model) > maxModelBytes || len(q.Suite) > scenarios.MaxSourceBytes || len(q.Report) > scenarios.MaxReportBytes {
		return nil, ErrLimit
	}
	m, err := jsonformat.Decode([]byte(q.Model))
	if err != nil {
		return nil, ErrInvalid
	}
	md, err := canonical.Digest(m)
	if err != nil {
		return nil, ErrBoundary
	}
	// The runner is a validation context, never an executable identity claim.
	runner, err := scenarios.NewRunner(scenarios.Binding{ProfileVersion: model.ProfileVersion, ModelDigest: md}, func(g.Task, g.Command) g.Result { panic("bound inventory validation invoked reducer") })
	if err != nil {
		return nil, ErrBoundary
	}
	suite, err := runner.Decode([]byte(q.Suite))
	if err != nil {
		return nil, ErrInvalid
	}
	sd, err := runner.Digest(suite)
	if err != nil {
		return nil, ErrBoundary
	}
	report, err := runner.DecodeReport([]byte(q.Report))
	if err != nil {
		return nil, ErrInvalid
	}
	explanation, err := scenarioexplain.ExplainWithRunner(m, report, runner)
	if err != nil {
		return nil, ErrInvalid
	}
	in := inputs{hash(q.Model), hash(q.Suite), hash(q.Report), md, sd}
	if report.SourceDigest == nil || *report.SourceDigest != in.SuiteSHA256 || report.ScenarioDigest != sd {
		return nil, ErrBinding
	}
	if err := bindSuite(suite, report); err != nil {
		return nil, err
	}
	out, err := inventory(m, explanation, in)
	if err != nil {
		return nil, err
	}
	return encode(out)
}

// Native decoders validate envelopes and comparisons. This additional pass binds
// that report to original suite order/literals, including an early mismatch.
func bindSuite(suite scenarios.Suite, report scenarios.Report) error {
	position, completed := 0, 0
	for _, c := range suite.Cases {
		current := c.Initial
		for _, step := range c.Steps {
			if position >= len(report.Entries) {
				return ErrInvalid
			}
			e := report.Entries[position]
			if e.CaseID != c.ID || e.StepID != step.ID || !reflect.DeepEqual(e.Input, current) || !reflect.DeepEqual(e.Command, step.Command) || !reflect.DeepEqual(e.Expected, step.Expected) {
				return ErrInvalid
			}
			position++
			if !e.Matched {
				if report.Matched || position != len(report.Entries) || completed != report.CasesCompleted {
					return ErrInvalid
				}
				return nil
			}
			// Matching refusals retain input, even when actual task is null.
			if e.Actual.OK {
				current = *e.Actual.Task
			}
		}
		completed++
	}
	if !report.Matched || position != len(report.Entries) || completed != report.CasesCompleted {
		return ErrInvalid
	}
	return nil
}
