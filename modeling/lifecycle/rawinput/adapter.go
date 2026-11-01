// Package rawinput is an offline demo adapter from raw Task and Command JSON to
// the generated lifecycle reducer. It checks transport structure only. Every
// lifecycle rule stays in the reducer, and no HTTP, file or network access exists.
package rawinput

import (
	"strconv"

	lifecycle "github.com/spik3r/clanker-foundry/modeling/lifecycle/generated"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/strictjson"
)

// Transport limits. The reducer enforces content limits (text length, criteria
// count); these only bound what the decoder reads. Task JSON can hold an outcome
// of 4096 bytes and 32 criteria of 1024 bytes, each up to six times larger when
// written as \u escapes, so 256 KiB covers every task the reducer accepts.
const (
	MaxTaskBytes    = 262144
	MaxCommandBytes = 16384
	// MaxDepth counts the top-level object as depth 0. A review object is depth
	// 1 and its strings depth 2; an acceptance array is depth 1.
	MaxDepth = 4
)

// Reason says why a raw input was rejected before the reducer ran.
type Reason string

const (
	// ReasonStrictJSON covers malformed or trailing JSON, duplicate, unknown or
	// missing keys, wrong types, invalid UTF-8 or surrogates, size and depth.
	ReasonStrictJSON Reason = "strict_json"
	// ReasonRevision means a revision string is not a canonical uint64.
	ReasonRevision Reason = "invalid_revision"
)

// InputError reports which raw input was rejected and why. It carries no input text.
type InputError struct {
	Input  string // "task" or "command"
	Reason Reason
}

func (e *InputError) Error() string {
	return "rawinput: " + e.Input + " rejected: " + string(e.Reason)
}

// Wire shapes. Every key is required. A pointer field may be JSON null; the
// strict decoder rejects null for the rest.
type wireTask struct {
	ID         string    `json:"id"`
	State      string    `json:"state"`
	Revision   string    `json:"revision"`
	Outcome    string    `json:"outcome"`
	Acceptance *[]string `json:"acceptance"`
	Owner      *string   `json:"owner"`
}

type wireReview struct {
	TaskID       string `json:"taskId"`
	TaskRevision string `json:"taskRevision"`
	ModelDigest  string `json:"modelDigest"`
	TaskDigest   string `json:"taskDigest"`
	Reviewer     string `json:"reviewer"`
	Decision     string `json:"decision"`
}

type wireCommand struct {
	Type             string      `json:"type"`
	Actor            string      `json:"actor"`
	ExpectedRevision string      `json:"expectedRevision"`
	Review           *wireReview `json:"review"`
}

// transition is the reducer. Tests replace it to prove rejected input never reaches it.
var transition = lifecycle.Transition

// Transition decodes the task, then the command, and calls the reducer once with
// the typed values. The reducer's Result is returned unchanged with a nil error.
// If either input is rejected the reducer is not called, the command is not
// examined when the task fails, and the Result has the same shape the reducer
// gives typed invalid input, with a non-nil *InputError.
func Transition(taskJSON, commandJSON []byte) (lifecycle.Result, error) {
	task, err := decodeTask(taskJSON)
	if err != nil {
		return rejected(), err
	}
	command, err := decodeCommand(commandJSON)
	if err != nil {
		return rejected(), err
	}
	return transition(task, command), nil
}

// rejected mirrors the reducer's INVALID_INPUT result without calling it.
func rejected() lifecycle.Result {
	return lifecycle.Result{Mode: "example-only", Authority: "none", ModelDigest: lifecycle.ModelDigest, Refusals: []string{"INVALID_INPUT"}}
}

func decodeTask(data []byte) (lifecycle.Task, error) {
	var w wireTask
	if strictjson.Decode(data, &w, MaxTaskBytes, MaxDepth) != nil {
		return lifecycle.Task{}, &InputError{"task", ReasonStrictJSON}
	}
	rev, ok := parseRevision(w.Revision)
	if !ok {
		return lifecycle.Task{}, &InputError{"task", ReasonRevision}
	}
	t := lifecycle.Task{ID: w.ID, State: lifecycle.State(w.State), Revision: rev, Outcome: w.Outcome}
	// null stays nil and [] stays empty; the reducer treats both as no criteria.
	if w.Acceptance != nil {
		t.Acceptance = append([]string{}, *w.Acceptance...)
	}
	// null is unassigned. A present empty string is a different, present value.
	if w.Owner != nil {
		o := *w.Owner
		t.Owner = &o
	}
	return t, nil
}

func decodeCommand(data []byte) (lifecycle.Command, error) {
	var w wireCommand
	if strictjson.Decode(data, &w, MaxCommandBytes, MaxDepth) != nil {
		return lifecycle.Command{}, &InputError{"command", ReasonStrictJSON}
	}
	expected, ok := parseRevision(w.ExpectedRevision)
	if !ok {
		return lifecycle.Command{}, &InputError{"command", ReasonRevision}
	}
	c := lifecycle.Command{Type: w.Type, Actor: w.Actor, ExpectedRevision: expected}
	if w.Review != nil {
		rr, ok := parseRevision(w.Review.TaskRevision)
		if !ok {
			return lifecycle.Command{}, &InputError{"command", ReasonRevision}
		}
		c.Review = &lifecycle.Review{
			TaskID: w.Review.TaskID, TaskRevision: rr, ModelDigest: w.Review.ModelDigest,
			TaskDigest: w.Review.TaskDigest, Reviewer: w.Review.Reviewer, Decision: w.Review.Decision,
		}
	}
	return c, nil
}

// parseRevision accepts only "0" or a decimal with no leading zero, sign, space
// or non-ASCII digit, that fits in uint64.
func parseRevision(s string) (uint64, bool) {
	if s == "" || (s[0] == '0' && len(s) > 1) {
		return 0, false
	}
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return 0, false
		}
	}
	n, err := strconv.ParseUint(s, 10, 64)
	return n, err == nil
}
