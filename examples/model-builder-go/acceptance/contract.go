// Package acceptance defines the frozen, bounded FirstDesignDraft fixture API.
package acceptance

// Task is runtime-valid under the preconditions in README.md.
type Task struct {
	ID         string
	State      string
	Revision   uint64
	Outcome    string
	Acceptance []string
	Owner      *string
}

// Command supports only the known submit trigger in this fixture.
type Command struct {
	Type             string
	Actor            string
	ExpectedRevision uint64
}

// Result is a fixture-level envelope, not a claim about an unseen generated API.
// Error is empty, GUARD_FAILED, ILLEGAL_TRANSITION or REVISION_LIMIT.
// Task is always nonnil in this runtime-valid scope. No-guard results use nil.
// TransitionID is T_submit when selected, otherwise empty.
type Result struct {
	Task         *Task
	Error        string
	FailedGuards []string
	TransitionID string
	Scope        string
	Authority    string
}

// Submit must be registered by candidate.go, without changing this contract.
// Each result owns its task, acceptance storage, owner and failed-guard storage.
var Submit func(Task, Command) Result
