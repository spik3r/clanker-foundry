// Package acceptance contains independent expectations for the example-only demo.
package acceptance

import (
	"encoding/json"
	"math"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"

	lifecycle "github.com/spik3r/clanker-foundry/modeling/lifecycle/generated"
)

func owner(s string) *string { return &s }
func initial() lifecycle.Task {
	return lifecycle.Task{ID: "task-1", State: "Draft", Revision: 0, Outcome: "Ship outcome", Acceptance: []string{"Verified"}, Owner: owner("owner-1")}
}
func cloneTask(t lifecycle.Task) lifecycle.Task {
	c := t
	if t.Acceptance != nil {
		c.Acceptance = make([]string, len(t.Acceptance))
		copy(c.Acceptance, t.Acceptance)
	}
	if t.Owner != nil {
		c.Owner = owner(*t.Owner)
	}
	return c
}
func cloneCommand(c lifecycle.Command) lifecycle.Command {
	d := c
	if c.Review != nil {
		r := *c.Review
		d.Review = &r
	}
	return d
}
func call(t *testing.T, task lifecycle.Task, cmd lifecycle.Command) lifecycle.Result {
	t.Helper()
	before := cloneTask(task)
	cb := cloneCommand(cmd)
	r := lifecycle.Transition(task, cmd)
	if !reflect.DeepEqual(task, before) || !reflect.DeepEqual(cmd, cb) {
		t.Fatal("transition mutated input task or command")
	}
	if r.Mode != "example-only" || r.Authority != "none" || r.ModelDigest != lifecycle.ModelDigest {
		t.Fatalf("metadata: %+v", r)
	}
	return r
}
func command(task lifecycle.Task, typ, actor string) lifecycle.Command {
	return lifecycle.Command{Type: typ, Actor: actor, ExpectedRevision: task.Revision}
}
func review(t *testing.T, task lifecycle.Task, decision string) *lifecycle.Review {
	t.Helper()
	d, e := lifecycle.TaskDigest(task)
	if e != nil {
		t.Fatal(e)
	}
	return &lifecycle.Review{TaskID: task.ID, TaskRevision: task.Revision, ModelDigest: lifecycle.ModelDigest, TaskDigest: d, Reviewer: "reviewer-1", Decision: decision}
}
func succeed(t *testing.T, task lifecycle.Task, cmd lifecycle.Command, state, id string) lifecycle.Task {
	t.Helper()
	r := call(t, task, cmd)
	if !r.OK || r.Task == nil || r.TransitionID != id || len(r.Refusals) != 0 {
		t.Fatalf("success expected: %+v", r)
	}
	want := cloneTask(task)
	want.State = lifecycle.State(state)
	want.Revision++
	if cmd.Type == "start" && want.Owner == nil {
		want.Owner = owner(cmd.Actor)
	}
	if !reflect.DeepEqual(*r.Task, want) {
		t.Fatalf("unexpected result got %+v want %+v", *r.Task, want)
	}
	return *r.Task
}
func refuse(t *testing.T, task lifecycle.Task, cmd lifecycle.Command) {
	t.Helper()
	r := call(t, task, cmd)
	if r.OK || r.Task == nil || !reflect.DeepEqual(*r.Task, task) || len(r.Refusals) == 0 {
		t.Fatalf("refusal expected unchanged: %+v", r)
	}
}
func TestLifecycleAndRework(t *testing.T) {
	a := initial()
	a = succeed(t, a, command(a, "prepare", "owner-1"), "Ready", "T_prepare")
	a = succeed(t, a, command(a, "start", "owner-1"), "InProgress", "T_start")
	a = succeed(t, a, command(a, "requestReview", "owner-1"), "InReview", "T_requestReview")
	c := command(a, "requestChanges", "reviewer-1")
	c.Review = review(t, a, "changesRequested")
	a = succeed(t, a, c, "InProgress", "T_requestChanges")
	a = succeed(t, a, command(a, "requestReview", "owner-1"), "InReview", "T_requestReview")
	c = command(a, "approve", "reviewer-1")
	c.Review = review(t, a, "approved")
	a = succeed(t, a, c, "Done", "T_approve")
	refuse(t, a, command(a, "prepare", "owner-1"))
}
func TestAssignmentAndRefusals(t *testing.T) {
	a := initial()
	a.State = "Ready"
	a.Owner = nil
	succeed(t, a, command(a, "start", "new-owner"), "InProgress", "T_start")
	refuse(t, a, command(a, "start", ""))
	a.Owner = owner("owner-1")
	refuse(t, a, command(a, "start", "other"))
	a.State = "InProgress"
	a.Owner = nil
	refuse(t, a, command(a, "requestReview", "owner-1"))
	a = initial()
	refuse(t, a, command(a, "start", "owner-1"))
	c := command(a, "prepare", "owner-1")
	c.ExpectedRevision = 1
	refuse(t, a, c)
	a.Outcome = ""
	refuse(t, a, command(a, "prepare", "owner-1"))
	a = initial()
	a.Acceptance = []string{}
	refuse(t, a, command(a, "prepare", "owner-1"))
	a = initial()
	a.Revision = math.MaxUint64
	refuse(t, a, command(a, "prepare", "owner-1"))
}
func TestReviewBindingFaults(t *testing.T) {
	a := initial()
	a.State = "InReview"
	a.Revision = 3
	tests := []struct {
		name   string
		change func(*lifecycle.Task, *lifecycle.Command)
	}{
		{"wrong task", func(_ *lifecycle.Task, c *lifecycle.Command) { c.Review.TaskID = "other" }},
		{"wrong revision", func(_ *lifecycle.Task, c *lifecycle.Command) { c.Review.TaskRevision-- }},
		{"wrong model", func(_ *lifecycle.Task, c *lifecycle.Command) {
			if c.Review.ModelDigest[0] == '0' {
				c.Review.ModelDigest = "1" + c.Review.ModelDigest[1:]
			} else {
				c.Review.ModelDigest = "0" + c.Review.ModelDigest[1:]
			}
		}},
		{"wrong reviewer actor", func(_ *lifecycle.Task, c *lifecycle.Command) { c.Actor = "other" }},
		{"owner reviews", func(_ *lifecycle.Task, c *lifecycle.Command) { c.Actor = "owner-1"; c.Review.Reviewer = "owner-1" }},
		{"wrong decision", func(_ *lifecycle.Task, c *lifecycle.Command) { c.Review.Decision = "changesRequested" }},
		{"same revision changed criteria", func(a *lifecycle.Task, _ *lifecycle.Command) { a.Acceptance = []string{"Changed"} }},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			task := cloneTask(a)
			c := command(task, "approve", "reviewer-1")
			c.Review = review(t, task, "approved")
			tc.change(&task, &c)
			refuse(t, task, c)
		})
	}
}
func TestDigestVectors(t *testing.T) {
	cases := []struct {
		task lifecycle.Task
		want string
	}{
		{lifecycle.Task{ID: "task-1", State: "InReview", Revision: 3, Outcome: "Ship <&>", Acceptance: []string{"café", "✓"}, Owner: owner("owner-1")}, "e0ab0176010aa6044142a2c830c3a6d146f41f862393d3a54bdb410071babc94"},
		{lifecycle.Task{ID: "例", State: "Draft", Revision: math.MaxUint64, Outcome: "雪", Acceptance: nil, Owner: nil}, "d5137867f36f4f6f8f51dfbf3ad43924b0f937d409a10b36af4cb860acf7760f"},
	}
	for _, tc := range cases {
		got, err := lifecycle.TaskDigest(tc.task)
		if err != nil || got != tc.want {
			t.Fatalf("digest got %s err %v want %s", got, err, tc.want)
		}
	}
	bad := initial()
	bad.Outcome = string([]byte{0xff})
	if _, err := lifecycle.TaskDigest(bad); err == nil {
		t.Fatal("invalid UTF-8 accepted")
	}
	if strings.TrimSpace(lifecycle.ModelDigest) == "" {
		t.Fatal("empty model digest")
	}
}

func TestReturnedTaskDoesNotAliasInputs(t *testing.T) {
	original := initial()
	before := cloneTask(original)
	cmd := command(original, "prepare", "owner-1")
	cmdBefore := cloneCommand(cmd)
	r := call(t, original, cmd)
	if !r.OK || r.Task == nil || len(r.Task.Acceptance) == 0 || r.Task.Owner == nil {
		t.Fatalf("expected successful task with acceptance and owner: %+v", r)
	}
	r.Task.Acceptance[0] = "changed after return"
	*r.Task.Owner = "changed after return"
	if !reflect.DeepEqual(original, before) || !reflect.DeepEqual(cmd, cmdBefore) {
		t.Fatal("returned task aliases an input")
	}
}

func TestReviewerAndDecisionFaultsSupplemental(t *testing.T) {
	task := initial()
	task.State = "InReview"
	task.Revision = 3
	cases := []struct {
		name string
		edit func(*lifecycle.Command)
	}{
		{"blank reviewer", func(c *lifecycle.Command) { c.Review.Reviewer = "" }},
		{"reviewer differs from actor", func(c *lifecycle.Command) { c.Actor = "another" }},
		{"reviewer is owner", func(c *lifecycle.Command) { c.Actor = "owner-1"; c.Review.Reviewer = "owner-1" }},
		{"wrong request changes decision", func(c *lifecycle.Command) { c.Type = "requestChanges" }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := command(task, "approve", "reviewer-1")
			c.Review = review(t, task, "approved")
			tc.edit(&c)
			refuse(t, task, c) // helper checks native task and command immutability
		})
	}
}

func TestDigestLineAndParagraphSeparators(t *testing.T) {
	task := lifecycle.Task{ID: "task-1", State: "Draft", Revision: 0,
		Outcome: "line\u2028paragraph\u2029", Acceptance: nil, Owner: nil}
	// SHA-256 of the specified canonical UTF-8 bytes, with literal JSON escapes
	// \\u2028 and \\u2029, independently computed outside the function under test.
	const want = "baee757788bc60219e2655f29f061b479b888ae964e79cefffe4117e56fe6f79"
	got, err := lifecycle.TaskDigest(task)
	if err != nil || got != want {
		t.Fatalf("digest got %q, err %v, want %q", got, err, want)
	}
}

// Pending agreed contract, intentionally not assertions:
// * Whether acceptance entries containing only whitespace invalidate a task.
// * Whether a non-nil pointer to an empty/blank owner is invalid input, a failed
//   actor guard, or an assigned value; Start assigns only when Owner is nil.
// * Exact blank/whitespace definition and whether invalid actor is input or guard failure.
// * Unknown state/event: rejection representation and invalid-input precedence.
// * Duplicate/unknown JSON fields: typed Go ABI has no raw JSON adapter yet.
// The base oracle covers nil owner assignment, blank actor, wrong assigned owner,
// success/refusal input snapshots, and the independent digest vectors.

// TestIsolatedGuardFailures uses the reviewed closed-profile guard semantics.
func TestIsolatedGuardFailures(t *testing.T) {
	cases := []struct {
		name    string
		task    lifecycle.Task
		command func(lifecycle.Task) lifecycle.Command
		guard   string
	}{
		{"present empty owner and actor", func() lifecycle.Task { a := initial(); a.State = "InProgress"; a.Owner = owner(""); return a }(),
			func(a lifecycle.Task) lifecycle.Command { return command(a, "requestReview", "") }, "G_owner"},
		{"different named request actor", func() lifecycle.Task { a := initial(); a.State = "InProgress"; return a }(),
			func(a lifecycle.Task) lifecycle.Command { return command(a, "requestReview", "other") }, "G_ownerActor"},
		{"present empty reviewer and actor", func() lifecycle.Task { a := initial(); a.State = "InReview"; a.Revision = 3; return a }(),
			func(a lifecycle.Task) lifecycle.Command {
				c := command(a, "approve", "")
				c.Review = review(t, a, "approved")
				c.Review.Reviewer = ""
				return c
			}, "G_reviewer"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := tc.command(tc.task)
			r := call(t, tc.task, c)
			if r.OK || r.Task == nil || !reflect.DeepEqual(*r.Task, tc.task) || !reflect.DeepEqual(r.Refusals, []string{tc.guard}) {
				t.Fatalf("expected only %s with unchanged task, got %+v", tc.guard, r)
			}
		})
	}
}

// TestCatalogExercisedIDs checks the full case objects and shared task/defaults
// against handwritten literals before executing independent typed assertions.
func TestCatalogExercisedIDs(t *testing.T) {
	data, err := os.ReadFile("../examples/oracle.json")
	if err != nil {
		t.Fatal(err)
	}
	var catalog struct {
		ReviewTemplate json.RawMessage   `json:"reviewTemplate"`
		BaseTask       json.RawMessage   `json:"baseTask"`
		Defaults       json.RawMessage   `json:"defaults"`
		Sequences      []json.RawMessage `json:"sequences"`
		Refusals       []json.RawMessage `json:"refusals"`
		Special        []json.RawMessage `json:"special"`
	}
	if err := json.Unmarshal(data, &catalog); err != nil {
		t.Fatal(err)
	}
	sameJSON := func(actual json.RawMessage, literal string) bool {
		var a, b interface{}
		if json.Unmarshal(actual, &a) != nil || json.Unmarshal([]byte(literal), &b) != nil {
			return false
		}
		return reflect.DeepEqual(a, b)
	}
	if !sameJSON(catalog.ReviewTemplate, `{"taskId":"task-1","taskRevision":"$CURRENT_REVISION","modelDigest":"$MODEL_DIGEST","taskDigest":"$CURRENT_TASK_DIGEST","reviewer":"reviewer-1","decision":"approved"}`) {
		t.Fatal("review template differs from independent literal")
	}
	if !sameJSON(catalog.BaseTask, `{"id":"task-1","state":"Draft","revision":"0","outcome":"Ship outcome","acceptance":["Verified"],"owner":"owner-1"}`) || !sameJSON(catalog.Defaults, `{"actor":"owner-1","expectedRevision":"$CURRENT_REVISION"}`) {
		t.Fatal("catalog task/defaults differ from independent literal")
	}
	expected := map[string]string{
		"stale-command":        `{"name":"stale command revision","state":"Draft","type":"prepare","command":{"expectedRevision":"0"},"revision":"1","guards":["G_revision"],"caseId":"stale-command","executionStatus":"executable"}`,
		"assigned-wrong-actor": `{"name":"wrong owner actor","state":"Ready","type":"start","command":{"actor":"other"},"guards":["G_ownerCompatible"],"caseId":"assigned-wrong-actor","executionStatus":"executable"}`,
		"start-assigns":        `{"name":"start assigns named actor on unassigned task","state":"Ready","task":{"owner":null},"type":"start","ok":true,"stateAfter":"InProgress","ownerAfter":"owner-1","revisionDelta":1,"caseId":"start-assigns","executionStatus":"executable"}`,
		"maximum-revision":     `{"name":"maximum revision refusal","caseId":"maximum-revision","executionStatus":"executable","revision":"18446744073709551615","expect":"refused after guards, task unchanged"}`,
	}
	guardOnly := func(t *testing.T, a lifecycle.Task, c lifecycle.Command, id string) {
		t.Helper()
		r := call(t, a, c)
		if r.OK || r.Task == nil || !reflect.DeepEqual(*r.Task, a) || !reflect.DeepEqual(r.Refusals, []string{id}) {
			t.Fatalf("expected only %s, unchanged task: %+v", id, r)
		}
	}
	registered := map[string]func(*testing.T){
		"stale-command": func(t *testing.T) {
			a := initial()
			a.Revision = 1
			c := command(a, "prepare", "owner-1")
			c.ExpectedRevision = 0
			guardOnly(t, a, c, "G_revision")
		},
		"assigned-wrong-actor": func(t *testing.T) {
			a := initial()
			a.State = "Ready"
			guardOnly(t, a, command(a, "start", "other"), "G_ownerCompatible")
		},
		"start-assigns": func(t *testing.T) {
			a := initial()
			a.State = "Ready"
			a.Owner = nil
			succeed(t, a, command(a, "start", "owner-1"), "InProgress", "T_start")
		},
		"maximum-revision": func(t *testing.T) {
			n, e := strconv.ParseUint("18446744073709551615", 10, 64)
			if e != nil {
				t.Fatal(e)
			}
			a := initial()
			a.Revision = n
			refuse(t, a, command(a, "prepare", "owner-1"))
		},
	}
	expected["sequences-01"] = `{"name":"happy path","steps":[{"type":"prepare","expectState":"Ready","expectTransition":"T_prepare"},{"type":"start","expectState":"InProgress","expectTransition":"T_start"},{"type":"requestReview","expectState":"InReview","expectTransition":"T_requestReview"},{"type":"approve","actor":"reviewer-1","review":{},"expectState":"Done","expectTransition":"T_approve"}],"caseId":"sequences-01","executionStatus":"executable"}`
	expected["sequences-02"] = `{"name":"rework and approval","steps":[{"type":"prepare","expectState":"Ready","expectTransition":"T_prepare"},{"type":"start","expectState":"InProgress","expectTransition":"T_start"},{"type":"requestReview","expectState":"InReview","expectTransition":"T_requestReview"},{"type":"requestChanges","actor":"reviewer-1","review":{"decision":"changesRequested"},"expectState":"InProgress","expectTransition":"T_requestChanges"},{"type":"requestReview","expectState":"InReview","expectTransition":"T_requestReview"},{"type":"approve","actor":"reviewer-1","review":{},"expectState":"Done","expectTransition":"T_approve"}],"caseId":"sequences-02","executionStatus":"executable"}`
	expected["stale-command"] = `{"name":"stale command revision","state":"Draft","type":"prepare","command":{"expectedRevision":"0"},"revision":"1","guards":["G_revision"],"caseId":"stale-command","executionStatus":"executable"}`
	expected["refusals-02"] = `{"name":"blank outcome","state":"Draft","type":"prepare","task":{"outcome":" \t\n"},"guards":["G_outcome"],"caseId":"refusals-02","executionStatus":"executable"}`
	expected["refusals-03"] = `{"name":"empty criteria","state":"Draft","type":"prepare","task":{"acceptance":[]},"guards":["G_acceptance"],"caseId":"refusals-03","executionStatus":"executable"}`
	expected["assigned-wrong-actor"] = `{"name":"wrong owner actor","state":"Ready","type":"start","command":{"actor":"other"},"guards":["G_ownerCompatible"],"caseId":"assigned-wrong-actor","executionStatus":"executable"}`
	expected["refusals-05"] = `{"name":"request review owner actor","state":"InProgress","type":"requestReview","command":{"actor":"other"},"guards":["G_ownerActor"],"caseId":"refusals-05","executionStatus":"executable"}`
	expected["refusals-06"] = `{"name":"blank reviewer and actor","state":"InReview","type":"approve","command":{"actor":"","review":{"reviewer":""}},"guards":["G_reviewer"],"caseId":"refusals-06","executionStatus":"executable"}`
	expected["refusals-07"] = `{"name":"wrong task","state":"InReview","type":"approve","review":{"taskId":"other"},"guards":["G_reviewTask"],"caseId":"refusals-07","executionStatus":"executable"}`
	expected["refusals-08"] = `{"name":"stale review","state":"InReview","type":"approve","revision":"2","review":{"taskRevision":"1"},"guards":["G_reviewRevision"],"caseId":"refusals-08","executionStatus":"executable"}`
	expected["refusals-09"] = `{"name":"wrong digest","state":"InReview","type":"approve","review":{"modelDigest":"$OTHER_MODEL_DIGEST"},"guards":["G_reviewModel"],"caseId":"refusals-09","executionStatus":"executable"}`
	expected["refusals-10"] = `{"name":"review actor mismatch","state":"InReview","type":"approve","command":{"actor":"other"},"guards":["G_reviewActor"],"caseId":"refusals-10","executionStatus":"executable"}`
	expected["refusals-11"] = `{"name":"owner reviews own task","state":"InReview","type":"approve","command":{"actor":"owner-1"},"review":{"reviewer":"owner-1"},"guards":["G_independent"],"caseId":"refusals-11","executionStatus":"executable"}`
	expected["refusals-12"] = `{"name":"wrong approval decision","state":"InReview","type":"approve","review":{"decision":"changesRequested"},"guards":["G_approved"],"caseId":"refusals-12","executionStatus":"executable"}`
	expected["refusals-13"] = `{"name":"wrong rework decision","state":"InReview","type":"requestChanges","review":{"decision":"approved"},"guards":["G_changesRequested"],"caseId":"refusals-13","executionStatus":"executable"}`
	expected["refusals-14"] = `{"name":"skip lifecycle","state":"Draft","type":"start","code":"ILLEGAL_TRANSITION","caseId":"refusals-14","executionStatus":"executable"}`
	expected["refusals-15"] = `{"name":"done is terminal","state":"Done","type":"prepare","code":"ILLEGAL_TRANSITION","caseId":"refusals-15","executionStatus":"executable"}`
	expected["refusals-16"] = `{"name":"blank actor cannot start unassigned task","state":"Ready","type":"start","task":{"owner":null},"command":{"actor":""},"guards":["G_actor"],"caseId":"refusals-16","executionStatus":"executable"}`
	expected["refusals-17"] = `{"name":"missing owner cannot submit review","state":"InProgress","type":"requestReview","task":{"owner":null},"guardsContain":["G_owner","G_ownerActor"],"caseId":"refusals-17","executionStatus":"executable"}`
	expected["special-01"] = `{"name":"unassigned draft may be prepared","task":{"owner":null},"type":"prepare","ok":true,"state":"Ready","caseId":"special-01","executionStatus":"executable"}`
	expected["special-02"] = `{"name":"multiple failed guards sorted","state":"Draft","task":{"outcome":"","acceptance":[]},"type":"prepare","guards":["G_acceptance","G_outcome"],"caseId":"special-02","executionStatus":"executable"}`
	expected["start-assigns"] = `{"name":"start assigns named actor on unassigned task","state":"Ready","task":{"owner":null},"type":"start","ok":true,"stateAfter":"InProgress","ownerAfter":"owner-1","revisionDelta":1,"caseId":"start-assigns","executionStatus":"executable"}`
	expected["maximum-revision"] = `{"name":"maximum revision refusal","caseId":"maximum-revision","executionStatus":"executable","revision":"18446744073709551615","expect":"refused after guards, task unchanged"}`
	registered["sequences-01"] = func(t *testing.T) {
		a := initial()
		a = succeed(t, a, command(a, "prepare", "owner-1"), "Ready", "T_prepare")
		a = succeed(t, a, command(a, "start", "owner-1"), "InProgress", "T_start")
		a = succeed(t, a, command(a, "requestReview", "owner-1"), "InReview", "T_requestReview")
		c := command(a, "approve", "reviewer-1")
		c.Review = review(t, a, "approved")
		succeed(t, a, c, "Done", "T_approve")
	}
	registered["sequences-02"] = func(t *testing.T) { TestLifecycleAndRework(t) }
	registered["refusals-02"] = func(t *testing.T) {
		a := initial()
		a.Outcome = " \t\n"
		guardOnly(t, a, command(a, "prepare", "owner-1"), "G_outcome")
	}
	registered["refusals-03"] = func(t *testing.T) {
		a := initial()
		a.Acceptance = []string{}
		guardOnly(t, a, command(a, "prepare", "owner-1"), "G_acceptance")
	}
	registered["refusals-05"] = func(t *testing.T) {
		a := initial()
		a.State = "InProgress"
		guardOnly(t, a, command(a, "requestReview", "other"), "G_ownerActor")
	}
	registered["refusals-06"] = func(t *testing.T) {
		a := initial()
		a.State = "InReview"
		c := command(a, "approve", "")
		c.Review = review(t, a, "approved")
		c.Review.Reviewer = ""
		guardOnly(t, a, c, "G_reviewer")
	}
	registered["refusals-07"] = func(t *testing.T) {
		a := initial()
		a.State = "InReview"
		c := command(a, "approve", "reviewer-1")
		c.Review = review(t, a, "approved")
		c.Review.TaskID = "other"
		guardOnly(t, a, c, "G_reviewTask")
	}
	registered["refusals-08"] = func(t *testing.T) {
		a := initial()
		a.State = "InReview"
		a.Revision = 2
		c := command(a, "approve", "reviewer-1")
		c.Review = review(t, a, "approved")
		c.Review.TaskRevision = 1
		guardOnly(t, a, c, "G_reviewRevision")
	}
	registered["refusals-09"] = func(t *testing.T) {
		a := initial()
		a.State = "InReview"
		c := command(a, "approve", "reviewer-1")
		c.Review = review(t, a, "approved")
		d := c.Review.ModelDigest
		if len(d) != 64 {
			t.Fatal("expected 64 hexadecimal digest")
		}
		if d[0] == '0' {
			c.Review.ModelDigest = "1" + d[1:]
		} else {
			c.Review.ModelDigest = "0" + d[1:]
		}
		guardOnly(t, a, c, "G_reviewModel")
	}
	registered["refusals-10"] = func(t *testing.T) {
		a := initial()
		a.State = "InReview"
		c := command(a, "approve", "other")
		c.Review = review(t, a, "approved")
		guardOnly(t, a, c, "G_reviewActor")
	}
	registered["refusals-11"] = func(t *testing.T) {
		a := initial()
		a.State = "InReview"
		c := command(a, "approve", "owner-1")
		c.Review = review(t, a, "approved")
		c.Review.Reviewer = "owner-1"
		guardOnly(t, a, c, "G_independent")
	}
	registered["refusals-12"] = func(t *testing.T) {
		a := initial()
		a.State = "InReview"
		c := command(a, "approve", "reviewer-1")
		c.Review = review(t, a, "changesRequested")
		guardOnly(t, a, c, "G_approved")
	}
	registered["refusals-13"] = func(t *testing.T) {
		a := initial()
		a.State = "InReview"
		c := command(a, "requestChanges", "reviewer-1")
		c.Review = review(t, a, "approved")
		guardOnly(t, a, c, "G_changesRequested")
	}
	registered["refusals-14"] = func(t *testing.T) {
		a := initial()
		r := call(t, a, command(a, "start", "owner-1"))
		if r.OK || r.Task == nil || !reflect.DeepEqual(*r.Task, a) || !reflect.DeepEqual(r.Refusals, []string{"ILLEGAL_TRANSITION"}) {
			t.Fatalf("illegal pair: %+v", r)
		}
	}
	registered["refusals-15"] = func(t *testing.T) {
		a := initial()
		a.State = "Done"
		r := call(t, a, command(a, "prepare", "owner-1"))
		if r.OK || r.Task == nil || !reflect.DeepEqual(*r.Task, a) || !reflect.DeepEqual(r.Refusals, []string{"ILLEGAL_TRANSITION"}) {
			t.Fatalf("terminal state: %+v", r)
		}
	}
	registered["refusals-16"] = func(t *testing.T) {
		a := initial()
		a.State = "Ready"
		a.Owner = nil
		guardOnly(t, a, command(a, "start", ""), "G_actor")
	}
	registered["refusals-17"] = func(t *testing.T) {
		a := initial()
		a.State = "InProgress"
		a.Owner = nil
		r := call(t, a, command(a, "requestReview", "owner-1"))
		if r.OK || r.Task == nil || !reflect.DeepEqual(*r.Task, a) || !reflect.DeepEqual(r.Refusals, []string{"G_owner", "G_ownerActor"}) {
			t.Fatalf("missing owner: %+v", r)
		}
	}
	registered["special-01"] = func(t *testing.T) {
		a := initial()
		a.Owner = nil
		succeed(t, a, command(a, "prepare", "owner-1"), "Ready", "T_prepare")
	}
	registered["special-02"] = func(t *testing.T) {
		a := initial()
		a.Outcome = ""
		a.Acceptance = []string{}
		r := call(t, a, command(a, "prepare", "owner-1"))
		if r.OK || r.Task == nil || !reflect.DeepEqual(*r.Task, a) || !reflect.DeepEqual(r.Refusals, []string{"G_acceptance", "G_outcome"}) {
			t.Fatalf("guard sorting: %+v", r)
		}
	}

	seen := map[string]bool{}
	exercised := map[string]bool{}
	all := append(append(catalog.Sequences, catalog.Refusals...), catalog.Special...)
	if len(all) == 0 {
		t.Fatal("empty catalog")
	}
	for _, raw := range all {
		var c struct {
			CaseID          string `json:"caseId"`
			ExecutionStatus string `json:"executionStatus"`
			PendingReason   string `json:"pendingReason"`
		}
		if err := json.Unmarshal(raw, &c); err != nil {
			t.Fatal(err)
		}
		if c.CaseID == "" || seen[c.CaseID] {
			t.Fatalf("missing or duplicate ID %q", c.CaseID)
		}
		seen[c.CaseID] = true
		switch c.ExecutionStatus {
		case "pending":
			if c.PendingReason == "" {
				t.Fatalf("pending %s lacks reason", c.CaseID)
			}
			if _, ok := registered[c.CaseID]; ok {
				t.Fatalf("registered case %s pending", c.CaseID)
			}
			t.Logf("PENDING %s: %s", c.CaseID, c.PendingReason)
		case "executable":
			fn, ok := registered[c.CaseID]
			if !ok {
				t.Fatalf("no handler for %s", c.CaseID)
			}
			literal, ok := expected[c.CaseID]
			if !ok || !sameJSON(raw, literal) {
				t.Fatalf("catalog case %s differs from independent literal", c.CaseID)
			}
			if t.Run(c.CaseID, fn) {
				exercised[c.CaseID] = true
			}
		default:
			t.Fatalf("invalid status %q for %s", c.ExecutionStatus, c.CaseID)
		}
	}
	if len(exercised) == 0 || len(exercised) != len(registered) || len(expected) != len(registered) {
		t.Fatalf("exercised %d of %d registered", len(exercised), len(registered))
	}
	for id := range registered {
		if !exercised[id] {
			t.Errorf("case %s not exercised successfully", id)
		}
	}
}
