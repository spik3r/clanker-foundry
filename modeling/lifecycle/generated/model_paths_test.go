// Code generated from the model: structural edge properties, not success witnesses.
package lifecycle

import (
	"reflect"
	"testing"
)

func TestModelIdentity(t *testing.T) {
	if ModelDigest != "948630cb7857199443b59b92d9b1ab0fadf2bf7a5bf8125f7a7e767db83958ed" {
		t.Fatal("compiled model identity mismatch")
	}
}
func TestModelPath_T_approve(t *testing.T) {
	owner := "owner"
	task := Task{ID: "derived-example", State: "InReview", Revision: 3, Outcome: "Example outcome", Acceptance: []string{"Example criterion"}, Owner: &owner}
	command := Command{Type: "approve", Actor: owner, ExpectedRevision: 3}
	d, e := TaskDigest(task)
	if e != nil {
		t.Fatal(e)
	}
	command.Actor = "reviewer"
	command.Review = &Review{TaskID: task.ID, TaskRevision: task.Revision, ModelDigest: ModelDigest, TaskDigest: d, Reviewer: "reviewer", Decision: "approved"}
	r := Transition(task, command)
	if !r.OK {
		if r.Task == nil || !reflect.DeepEqual(*r.Task, task) || len(r.Refusals) == 0 {
			t.Fatalf("invalid structural refusal: %+v", r)
		}
		return
	}
	if r.Task == nil || r.TransitionID != "T_approve" || r.Task.State != "Done" || r.Task.Revision != 4 {
		t.Fatalf("model edge effects drifted: %+v", r)
	}
}
func TestModelPath_T_prepare(t *testing.T) {
	owner := "owner"
	task := Task{ID: "derived-example", State: "Draft", Revision: 3, Outcome: "Example outcome", Acceptance: []string{"Example criterion"}, Owner: &owner}
	command := Command{Type: "prepare", Actor: owner, ExpectedRevision: 3}
	r := Transition(task, command)
	if !r.OK {
		if r.Task == nil || !reflect.DeepEqual(*r.Task, task) || len(r.Refusals) == 0 {
			t.Fatalf("invalid structural refusal: %+v", r)
		}
		return
	}
	if r.Task == nil || r.TransitionID != "T_prepare" || r.Task.State != "Ready" || r.Task.Revision != 4 {
		t.Fatalf("model edge effects drifted: %+v", r)
	}
}
func TestModelPath_T_requestChanges(t *testing.T) {
	owner := "owner"
	task := Task{ID: "derived-example", State: "InReview", Revision: 3, Outcome: "Example outcome", Acceptance: []string{"Example criterion"}, Owner: &owner}
	command := Command{Type: "requestChanges", Actor: owner, ExpectedRevision: 3}
	d, e := TaskDigest(task)
	if e != nil {
		t.Fatal(e)
	}
	command.Actor = "reviewer"
	command.Review = &Review{TaskID: task.ID, TaskRevision: task.Revision, ModelDigest: ModelDigest, TaskDigest: d, Reviewer: "reviewer", Decision: "changesRequested"}
	r := Transition(task, command)
	if !r.OK {
		if r.Task == nil || !reflect.DeepEqual(*r.Task, task) || len(r.Refusals) == 0 {
			t.Fatalf("invalid structural refusal: %+v", r)
		}
		return
	}
	if r.Task == nil || r.TransitionID != "T_requestChanges" || r.Task.State != "InProgress" || r.Task.Revision != 4 {
		t.Fatalf("model edge effects drifted: %+v", r)
	}
}
func TestModelPath_T_requestReview(t *testing.T) {
	owner := "owner"
	task := Task{ID: "derived-example", State: "InProgress", Revision: 3, Outcome: "Example outcome", Acceptance: []string{"Example criterion"}, Owner: &owner}
	command := Command{Type: "requestReview", Actor: owner, ExpectedRevision: 3}
	r := Transition(task, command)
	if !r.OK {
		if r.Task == nil || !reflect.DeepEqual(*r.Task, task) || len(r.Refusals) == 0 {
			t.Fatalf("invalid structural refusal: %+v", r)
		}
		return
	}
	if r.Task == nil || r.TransitionID != "T_requestReview" || r.Task.State != "InReview" || r.Task.Revision != 4 {
		t.Fatalf("model edge effects drifted: %+v", r)
	}
}
func TestModelPath_T_start(t *testing.T) {
	owner := "owner"
	task := Task{ID: "derived-example", State: "Ready", Revision: 3, Outcome: "Example outcome", Acceptance: []string{"Example criterion"}, Owner: &owner}
	command := Command{Type: "start", Actor: owner, ExpectedRevision: 3}
	task.Owner = nil
	r := Transition(task, command)
	if !r.OK {
		if r.Task == nil || !reflect.DeepEqual(*r.Task, task) || len(r.Refusals) == 0 {
			t.Fatalf("invalid structural refusal: %+v", r)
		}
		return
	}
	if r.Task == nil || r.TransitionID != "T_start" || r.Task.State != "InProgress" || r.Task.Revision != 4 {
		t.Fatalf("model edge effects drifted: %+v", r)
	}
	if r.Task.Owner == nil || *r.Task.Owner != command.Actor {
		t.Fatal("assignment effect missing")
	}
}
func TestModelTerminal_4_Done_7_approve(t *testing.T) {
	task := Task{ID: "terminal-example", State: "Done", Revision: 7, Outcome: "Outcome", Acceptance: []string{"Criterion"}}
	command := Command{Type: "approve", Actor: "actor", ExpectedRevision: 7}
	r := Transition(task, command)
	if r.OK || r.Task == nil || !reflect.DeepEqual(*r.Task, task) || len(r.Refusals) != 1 || r.Refusals[0] != "ILLEGAL_TRANSITION" {
		t.Fatalf("terminal edge admitted: %+v", r)
	}
}
func TestModelTerminal_4_Done_7_prepare(t *testing.T) {
	task := Task{ID: "terminal-example", State: "Done", Revision: 7, Outcome: "Outcome", Acceptance: []string{"Criterion"}}
	command := Command{Type: "prepare", Actor: "actor", ExpectedRevision: 7}
	r := Transition(task, command)
	if r.OK || r.Task == nil || !reflect.DeepEqual(*r.Task, task) || len(r.Refusals) != 1 || r.Refusals[0] != "ILLEGAL_TRANSITION" {
		t.Fatalf("terminal edge admitted: %+v", r)
	}
}
func TestModelTerminal_4_Done_14_requestChanges(t *testing.T) {
	task := Task{ID: "terminal-example", State: "Done", Revision: 7, Outcome: "Outcome", Acceptance: []string{"Criterion"}}
	command := Command{Type: "requestChanges", Actor: "actor", ExpectedRevision: 7}
	r := Transition(task, command)
	if r.OK || r.Task == nil || !reflect.DeepEqual(*r.Task, task) || len(r.Refusals) != 1 || r.Refusals[0] != "ILLEGAL_TRANSITION" {
		t.Fatalf("terminal edge admitted: %+v", r)
	}
}
func TestModelTerminal_4_Done_13_requestReview(t *testing.T) {
	task := Task{ID: "terminal-example", State: "Done", Revision: 7, Outcome: "Outcome", Acceptance: []string{"Criterion"}}
	command := Command{Type: "requestReview", Actor: "actor", ExpectedRevision: 7}
	r := Transition(task, command)
	if r.OK || r.Task == nil || !reflect.DeepEqual(*r.Task, task) || len(r.Refusals) != 1 || r.Refusals[0] != "ILLEGAL_TRANSITION" {
		t.Fatalf("terminal edge admitted: %+v", r)
	}
}
func TestModelTerminal_4_Done_5_start(t *testing.T) {
	task := Task{ID: "terminal-example", State: "Done", Revision: 7, Outcome: "Outcome", Acceptance: []string{"Criterion"}}
	command := Command{Type: "start", Actor: "actor", ExpectedRevision: 7}
	r := Transition(task, command)
	if r.OK || r.Task == nil || !reflect.DeepEqual(*r.Task, task) || len(r.Refusals) != 1 || r.Refusals[0] != "ILLEGAL_TRANSITION" {
		t.Fatalf("terminal edge admitted: %+v", r)
	}
}
