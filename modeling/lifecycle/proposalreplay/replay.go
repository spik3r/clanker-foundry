// Package proposalreplay connects frozen proposals to explicitly requested local
// candidate execution. It grants no review, identity or deployment authority.
package proposalreplay

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"runtime"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/acceptanceauthor"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/canonical"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/compiler"
	lifecycle "github.com/spik3r/clanker-foundry/modeling/lifecycle/generated"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/jsonformat"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/model"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/scenarios"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/strictjson"
)

const RequestVersion = "lifecycle-proposal-request/v1"
const ReceiptVersion = "lifecycle-proposal-replay/v1"
const MaxRequestBytes = 16 << 20
const MaxReceiptBytes = 9 << 20

var ErrInvalid = errors.New("invalid proposal request")
var ErrExecution = errors.New("proposal execution boundary failed")

// Request contains original file text, not paths or executable configuration.
type Request struct {
	FormatVersion string `json:"formatVersion"`
	Model         string `json:"model"`
	Suite         string `json:"suite"`
	Manifest      string `json:"manifest"`
}

type Sources struct {
	ModelSHA256    string `json:"modelSha256"`
	SuiteSHA256    string `json:"suiteSha256"`
	ManifestSHA256 string `json:"manifestSha256"`
}

// Receipt separates validation-only claims from a process actually started by
// this invocation. Hashes and a successful outcome do not authenticate the host.
type Receipt struct {
	FormatVersion     string                       `json:"formatVersion"`
	Mode              string                       `json:"mode"`
	Authority         string                       `json:"authority"`
	Action            string                       `json:"action"`
	Outcome           string                       `json:"outcome"`
	Sources           Sources                      `json:"sources"`
	ValidationContext string                       `json:"validationContext"`
	Preflight         acceptanceauthor.BoundReport `json:"preflight"`
	SourceRevision    *string                      `json:"sourceRevision"`
	ArchiveSHA256     *string                      `json:"archiveSha256"`
	BinarySHA256      *string                      `json:"binarySha256"`
	HostPlatform      string                       `json:"hostPlatform"`
	Compilation       string                       `json:"compilation"`
	Replay            string                       `json:"replay"`
	CandidateStarted  bool                         `json:"candidateStarted"`
	CandidateExitCode *int                         `json:"candidateExitCode"`
	CandidateReport   json.RawMessage              `json:"candidateReport"`
	Cleanup           string                       `json:"cleanup"`
	FailureStage      *string                      `json:"failureStage"`
}

type proposal struct {
	request Request
	suite   scenarios.Suite
}

func hash(b []byte) string {
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}
func ptr[T any](v T) *T { return &v }

func prepare(raw []byte, action string) (proposal, Receipt, error) {
	var p proposal
	var r Receipt
	if runtime.Version() != compiler.ToolchainVersion || strictjson.Decode(raw, &p.request, MaxRequestBytes, 4) != nil || p.request.FormatVersion != RequestVersion || len(p.request.Model) > 1<<20 || len(p.request.Suite) > scenarios.MaxSourceBytes || len(p.request.Manifest) > acceptanceauthor.MaxManifestBytes {
		return p, r, ErrInvalid
	}
	m, err := jsonformat.Decode([]byte(p.request.Model))
	if err != nil {
		return p, r, ErrInvalid
	}
	digest, err := canonical.Digest(m)
	if err != nil {
		return p, r, ErrInvalid
	}
	// This private callback is a validation-only context, never a candidate
	// composition. No Run/Execute call may use it.
	validation, err := scenarios.NewRunner(scenarios.Binding{ProfileVersion: model.ProfileVersion, ModelDigest: digest}, func(lifecycle.Task, lifecycle.Command) lifecycle.Result { panic("proposal validation context invoked") })
	if err != nil {
		return p, r, ErrInvalid
	}
	checked, err := acceptanceauthor.CheckWithRunner([]byte(p.request.Manifest), []byte(p.request.Model), []byte(p.request.Suite), "check", validation)
	if err != nil {
		return p, r, ErrInvalid
	}
	r = Receipt{FormatVersion: ReceiptVersion, Mode: "example-only", Authority: "none", Action: action, Outcome: "stale", Sources: Sources{hash([]byte(p.request.Model)), hash([]byte(p.request.Suite)), hash([]byte(p.request.Manifest))}, ValidationContext: "non-executing-model-context", Preflight: checked, HostPlatform: runtime.GOOS + "/" + runtime.GOARCH, Compilation: "not-started", Replay: "not-started", Cleanup: "not-created"}
	if action == "check" {
		r.Compilation, r.Replay = "not-requested", "not-requested"
	}
	if checked.ModelBinding != "current" || checked.SuiteBinding != "current" || checked.SuiteModelBinding != "current" || checked.RunnerBinding != "current" {
		return p, r, nil
	}
	if checked.SuiteValidation != "valid" || checked.ScenarioDigest == nil || checked.Replay != "not-run" {
		return proposal{}, Receipt{}, ErrInvalid
	}
	p.suite, err = validation.Decode([]byte(p.request.Suite))
	if err != nil {
		return proposal{}, Receipt{}, ErrInvalid
	}
	r.Outcome = "checked"
	return p, r, nil
}

// Check performs no filesystem access, subprocess invocation or candidate work.
// A current runnerBinding in Preflight belongs only to the labeled validation
// context; it is not an assertion about any compiled executable.
func Check(raw []byte) (Receipt, error) {
	_, r, err := prepare(raw, "check")
	return r, err
}

// Execute requires separate caller authorization on a trusted single-writer
// host. sourceRoot is trusted host configuration, not request data. Stale and
// malformed preflight launches no process. Later failure may follow invocation;
// inspect CandidateStarted and Replay rather than inferring rollback from error.
func Execute(ctx context.Context, raw []byte, sourceRoot string) (Receipt, error) {
	return execute(ctx, raw, sourceRoot, localHost())
}

func Encode(r Receipt) ([]byte, error) {
	var out bytes.Buffer
	e := json.NewEncoder(&out)
	e.SetEscapeHTML(false)
	if err := e.Encode(r); err != nil || out.Len() > MaxReceiptBytes {
		return nil, ErrExecution
	}
	return out.Bytes(), nil
}

// ExitCode describes the requested check/replay boundary, not design approval.
func ExitCode(r Receipt, err error) int {
	if errors.Is(err, ErrInvalid) {
		return 2
	}
	if err != nil || r.FailureStage != nil {
		return 4
	}
	if r.Outcome == "stale" || r.Outcome == "mismatched" {
		return 1
	}
	if r.Outcome == "checked" || r.Outcome == "matched" {
		return 0
	}
	return 4
}

// The child is the closed, pinned modelplay, not an arbitrary report provider.
// These checks bind its captured envelope and ordered prefix to this invocation;
// they are not a second lifecycle evaluator or a portable report verifier.
type capturedReport struct {
	FormatVersion   string          `json:"formatVersion"`
	ScenarioVersion string          `json:"scenarioVersion"`
	ProfileVersion  string          `json:"profileVersion"`
	Mode            string          `json:"mode"`
	Authority       string          `json:"authority"`
	ModelDigest     string          `json:"modelDigest"`
	ScenarioDigest  string          `json:"scenarioDigest"`
	SourceDigest    string          `json:"sourceDigest"`
	Matched         bool            `json:"matched"`
	CasesCompleted  int             `json:"casesCompleted"`
	StepsExecuted   int             `json:"stepsExecuted"`
	Entries         []capturedEntry `json:"entries"`
}
type capturedEntry struct {
	CaseID      string          `json:"caseId"`
	StepID      string          `json:"stepId"`
	Input       json.RawMessage `json:"input"`
	Command     json.RawMessage `json:"command"`
	Expected    json.RawMessage `json:"expected"`
	Actual      json.RawMessage `json:"actual"`
	Matched     bool            `json:"matched"`
	Differences []string        `json:"differences"`
}

func captured(raw []byte, exit int, p proposal, r Receipt) (bool, error) {
	var d capturedReport
	if strictjson.Decode(raw, &d, scenarios.MaxReportBytes, scenarios.MaxDepth) != nil || (exit != 0 && exit != 1) || d.FormatVersion != scenarios.ReportVersion || d.ScenarioVersion != scenarios.ScenarioVersion || d.ProfileVersion != model.ProfileVersion || d.Mode != scenarios.Mode || d.Authority != scenarios.Authority || d.ModelDigest != r.Preflight.CurrentModelDigest || r.Preflight.ScenarioDigest == nil || d.ScenarioDigest != *r.Preflight.ScenarioDigest || d.SourceDigest != r.Sources.SuiteSHA256 || d.Matched != (exit == 0) || d.StepsExecuted != len(d.Entries) || len(d.Entries) == 0 {
		return false, ErrExecution
	}
	position := 0
	completed := 0
	for _, c := range p.suite.Cases {
		for _, s := range c.Steps {
			if position >= len(d.Entries) {
				return false, ErrExecution
			}
			e := d.Entries[position]
			if e.CaseID != c.ID || e.StepID != s.ID || e.Matched != (len(e.Differences) == 0) {
				return false, ErrExecution
			}
			position++
			if !e.Matched {
				if d.Matched || position != len(d.Entries) || completed != d.CasesCompleted {
					return false, ErrExecution
				}
				return false, nil
			}
		}
		completed++
	}
	if !d.Matched || position != len(d.Entries) || completed != d.CasesCompleted {
		return false, ErrExecution
	}
	return true, nil
}
