// Package scenarios executes bounded, invented lifecycle fixtures. Its reports
// compare supplied expectations; they authenticate neither actors nor code.
package scenarios

import (
	"errors"
	"strings"
	"unicode/utf8"

	lifecycle "github.com/spik3r/clanker-foundry/modeling/lifecycle/generated"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/model"
)

const (
	FormatVersion   = "lifecycle-scenarios-json/v1"
	ScenarioVersion = "lifecycle-scenarios/v1"
	ReportVersion   = "lifecycle-scenario-report/v1"
	MaxSourceBytes  = 1 << 20
	MaxReportBytes  = 8 << 20
	MaxDepth        = 64
	MaxCases        = 16
	MaxStepsPerCase = 32
	MaxSteps        = 128
	MaxTextBytes    = 8192
	MaxItems        = 64
	Mode            = "example-only"
	Authority       = "none"
)

var (
	ErrInvalidScenario = errors.New("invalid scenario fixture")
	ErrModelBinding    = errors.New("unsupported scenario model binding")
	ErrReportLimit     = errors.New("invalid or oversized scenario report")
)

// Suite is the format-neutral scenario profile. Case and step order is semantic.
// Version is 1; a zero value is not an implicit default.
type Suite struct {
	Version                         int
	ID, ProfileVersion, ModelDigest string
	Cases                           []Case
}
type Case struct {
	ID      string
	Initial lifecycle.Task
	Steps   []Step
}
type Step struct {
	ID       string
	Command  lifecycle.Command
	Expected Expected
}
type Expected struct {
	OK           bool
	TransitionID string
	Task         *lifecycle.Task
	Refusals     []string
}

// Report owns its task/command/expectation copies. SourceDigest is nil for typed
// Run, because typed input has no observed original source bytes.
type Report struct {
	ModelDigest, ScenarioDigest   string
	SourceDigest                  *string
	Matched                       bool
	CasesCompleted, StepsExecuted int
	Entries                       []Entry
}
type Entry struct {
	CaseID, StepID string
	Input          lifecycle.Task
	Command        lifecycle.Command
	Expected       Expected
	Actual         lifecycle.Result
	Matched        bool
	Differences    []string
}

type budget struct{ used, limit int }

func (b *budget) text(s string, limit int) bool {
	if len(s) > limit || !utf8.ValidString(s) {
		return false
	}
	b.used += len(s)
	return b.used <= b.limit
}
func (b *budget) task(t lifecycle.Task) bool {
	if !b.text(t.ID, MaxTextBytes) || !b.text(string(t.State), MaxTextBytes) || !b.text(t.Outcome, MaxTextBytes) || len(t.Acceptance) > MaxItems {
		return false
	}
	if t.Owner != nil && !b.text(*t.Owner, MaxTextBytes) {
		return false
	}
	for _, s := range t.Acceptance {
		if !b.text(s, MaxTextBytes) {
			return false
		}
	}
	return true
}
func (b *budget) command(c lifecycle.Command) bool {
	if !b.text(c.Type, MaxTextBytes) || !b.text(c.Actor, MaxTextBytes) {
		return false
	}
	if c.Review != nil {
		r := c.Review
		for _, s := range []string{r.TaskID, r.ModelDigest, r.TaskDigest, r.Reviewer, r.Decision} {
			if !b.text(s, MaxTextBytes) {
				return false
			}
		}
	}
	return true
}
func (b *budget) reasons(v []string) bool {
	if len(v) > MaxItems {
		return false
	}
	for _, s := range v {
		if !b.text(s, 256) {
			return false
		}
	}
	return true
}
func (b *budget) expected(e Expected) bool {
	if !b.text(e.TransitionID, MaxTextBytes) || e.Refusals == nil || !b.reasons(e.Refusals) {
		return false
	}
	if e.OK && (e.Task == nil || strings.TrimSpace(e.TransitionID) == "" || len(e.Refusals) != 0) {
		return false
	}
	if !e.OK && len(e.Refusals) == 0 {
		return false
	}
	return e.Task == nil || b.task(*e.Task)
}
func hexDigest(s string) bool {
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

// Validate checks the whole suite before execution. It deliberately does not
// enforce the reducer's business-input limits: fixtures may expect INVALID_INPUT.
func Validate(s Suite) error { return validateFor(s, defaultBinding()) }

func validateFor(s Suite, binding Binding) error {
	if s.Version != 1 || len(s.Cases) < 1 || len(s.Cases) > MaxCases || !hexDigest(s.ModelDigest) {
		return ErrInvalidScenario
	}
	b := budget{limit: MaxSourceBytes}
	ids := map[string]bool{}
	id := func(v string) bool {
		if !model.Identifier(v) || ids[v] || !b.text(v, 48) {
			return false
		}
		ids[v] = true
		return true
	}
	if !id(s.ID) || !b.text(s.ProfileVersion, MaxTextBytes) {
		return ErrInvalidScenario
	}
	total := 0
	for _, c := range s.Cases {
		if !id(c.ID) || len(c.Steps) < 1 || len(c.Steps) > MaxStepsPerCase || !b.task(c.Initial) {
			return ErrInvalidScenario
		}
		total += len(c.Steps)
		if total > MaxSteps {
			return ErrInvalidScenario
		}
		for _, step := range c.Steps {
			if !id(step.ID) || !b.command(step.Command) || !b.expected(step.Expected) {
				return ErrInvalidScenario
			}
		}
	}
	// A typed fixture must be representable by the first format adapter too.
	raw, e := encodeJSON(toDTO(s))
	if e != nil || len(raw) > MaxSourceBytes {
		return ErrInvalidScenario
	}
	if s.ProfileVersion != binding.ProfileVersion || s.ModelDigest != binding.ModelDigest {
		return ErrModelBinding
	}
	return nil
}

func copyStrings(v []string) []string {
	if v == nil {
		return nil
	}
	return append([]string{}, v...)
}
func copyTask(t lifecycle.Task) lifecycle.Task {
	t.Acceptance = copyStrings(t.Acceptance)
	if t.Owner != nil {
		v := *t.Owner
		t.Owner = &v
	}
	return t
}
func copyCommand(c lifecycle.Command) lifecycle.Command {
	if c.Review != nil {
		r := *c.Review
		c.Review = &r
	}
	return c
}
func copyExpected(e Expected) Expected {
	e.Refusals = copyStrings(e.Refusals)
	if e.Task != nil {
		t := copyTask(*e.Task)
		e.Task = &t
	}
	return e
}
func copyResult(r lifecycle.Result) lifecycle.Result {
	r.Refusals = copyStrings(r.Refusals)
	if r.Task != nil {
		t := copyTask(*r.Task)
		r.Task = &t
	}
	return r
}
