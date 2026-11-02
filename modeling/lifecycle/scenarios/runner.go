package scenarios

import (
	"errors"

	lifecycle "github.com/spik3r/clanker-foundry/modeling/lifecycle/generated"
)

const BoundReportVersion = "lifecycle-bound-scenario-report/v1"
const BindingSource = "caller-supplied-go-binding"

var ErrInvalidRunner = errors.New("invalid bound scenario runner")

// ErrProviderProtocol means a callback was invoked and returned an invalid
// envelope or an unbounded/structurally invalid result. No later call is made.
// It does not mean not-run, rollback, or absence of callback side effects.
var ErrProviderProtocol = errors.New("bound scenario provider protocol failure after invocation")

// Binding is a caller's declaration, not authenticated compiled-code provenance.
// Compose it in reviewed Go code from the generated reducer's constants, never
// by substituting a request's model identity for those constants.
type Binding struct {
	ProfileVersion string
	ModelDigest    string
}

// Runner snapshots configuration, not callback behavior. The callback is trusted
// application code: it can block, allocate, mutate its closure or perform I/O.
// Panics propagate. Runner does not sandbox, cancel or roll back callback work.
// Existing generated types serve only as the lifecycle/v1 transport here.
type Runner struct {
	binding Binding
	apply   transition
}

func defaultBinding() Binding {
	return Binding{lifecycle.ProfileVersion, lifecycle.ModelDigest}
}

// NewRunner does not execute the callback or authenticate its claimed identity.
func NewRunner(binding Binding, apply func(lifecycle.Task, lifecycle.Command) lifecycle.Result) (Runner, error) {
	r := Runner{binding: binding, apply: apply}
	if _, err := r.ModelBinding(); err != nil {
		return Runner{}, err
	}
	return r, nil
}

// ModelBinding returns an owned value. A zero runner is never a default runner.
func (r Runner) ModelBinding() (Binding, error) {
	if r.apply == nil || r.binding.ProfileVersion != lifecycle.ProfileVersion || !hexDigest(r.binding.ModelDigest) {
		return Binding{}, ErrInvalidRunner
	}
	return r.binding, nil
}

func (r Runner) Validate(s Suite) error {
	if _, err := r.ModelBinding(); err != nil {
		return err
	}
	return validateFor(s, r.binding)
}

func (r Runner) Decode(raw []byte) (Suite, error) {
	if _, err := r.ModelBinding(); err != nil {
		return Suite{}, err
	}
	return decodeFor(raw, r.binding)
}

func (r Runner) Encode(s Suite) ([]byte, error) {
	if _, err := r.ModelBinding(); err != nil {
		return nil, err
	}
	return encodeFor(s, r.binding)
}

func (r Runner) Digest(s Suite) (string, error) {
	if _, err := r.ModelBinding(); err != nil {
		return "", err
	}
	return digestFor(s, r.binding)
}

func (r Runner) Run(s Suite) (Report, error) {
	if _, err := r.ModelBinding(); err != nil {
		return Report{}, err
	}
	return runFor(s, r.apply, r.binding, true)
}

func (r Runner) Execute(raw []byte) (Report, error) {
	if _, err := r.ModelBinding(); err != nil {
		return Report{}, err
	}
	return executeFor(raw, r.apply, r.binding, true)
}

func (r Runner) EncodeReport(report Report) ([]byte, error) {
	if _, err := r.ModelBinding(); err != nil {
		return nil, err
	}
	return encodeReportFor(report, r.binding, true)
}

// Inspect bounded structure before cloning callback-controlled slices or tasks.
// This bounds our processing only, not allocation or other work in the callback.
func providerResultValid(result lifecycle.Result, binding Binding) bool {
	if result.Mode != Mode || result.Authority != Authority || result.ModelDigest != binding.ModelDigest {
		return false
	}
	b := budget{limit: MaxReportBytes}
	return b.expected(Expected{result.OK, result.TransitionID, result.Task, result.Refusals})
}

func copySuite(s Suite) Suite {
	cases := make([]Case, len(s.Cases))
	for i, c := range s.Cases {
		steps := make([]Step, len(c.Steps))
		for j, step := range c.Steps {
			steps[j] = Step{step.ID, copyCommand(step.Command), copyExpected(step.Expected)}
		}
		cases[i] = Case{c.ID, copyTask(c.Initial), steps}
	}
	s.Cases = cases
	return s
}
