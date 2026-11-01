// Package scenarioacceptance checks independently frozen invented expectations.
// It never calls the reducer to author or refresh an expected result.
package scenarioacceptance

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"reflect"
	"strings"
	"testing"

	lifecycle "github.com/spik3r/clanker-foundry/modeling/lifecycle/generated"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/scenarios"
)

const fixtureHash = "ef199aa93c9e70887f3da8e72aa4a79dcaf130c4cec72d27b240d6780a3e4b26"
const boundModel = "948630cb7857199443b59b92d9b1ab0fadf2bf7a5bf8125f7a7e767db83958ed"

func fixture(t *testing.T) []byte {
	t.Helper()
	b, e := os.ReadFile("../scenarios/testdata/lifecycle-v1.json")
	if e != nil {
		t.Fatal(e)
	}
	if fmt.Sprintf("%x", sha256.Sum256(b)) != fixtureHash {
		t.Fatal("independent frozen fixture changed; never refresh from observed results")
	}
	return b
}
func decoded(t *testing.T) scenarios.Suite {
	t.Helper()
	s, e := scenarios.Decode(fixture(t))
	if e != nil {
		t.Fatal(e)
	}
	return s
}
func snapshot(t *testing.T, v any) []byte {
	t.Helper()
	b, e := json.Marshal(v)
	if e != nil {
		t.Fatal(e)
	}
	return b
}
func str(s string) *string { return &s }
func task(state string, rev uint64, acceptance []string, owner *string) lifecycle.Task {
	return lifecycle.Task{ID: "task-1", State: lifecycle.State(state), Revision: rev, Outcome: "Ship <&>", Acceptance: acceptance, Owner: owner}
}
func assertResult(t *testing.T, got lifecycle.Result, want scenarios.Expected) {
	t.Helper()
	expected := lifecycle.Result{Mode: "example-only", Authority: "none", ModelDigest: boundModel, OK: want.OK, TransitionID: want.TransitionID, Task: want.Task, Refusals: want.Refusals}
	if !reflect.DeepEqual(got, expected) {
		t.Fatalf("complete result mismatch\ngot %#v\nwant %#v", got, expected)
	}
}
func assertRejected(t *testing.T, s scenarios.Suite) {
	t.Helper()
	r, e := scenarios.Run(s)
	if e == nil {
		t.Fatal("invalid whole suite accepted")
	}
	if r.StepsExecuted != 0 || r.CasesCompleted != 0 || len(r.Entries) != 0 {
		t.Fatal("invalid suite produced execution entries")
	}
}
func TestFrozenLiteralSuiteAndRetention(t *testing.T) {
	s := decoded(t)
	before := snapshot(t, s)
	r, e := scenarios.Run(s)
	if e != nil {
		t.Fatal(e)
	}
	if !r.Matched || r.CasesCompleted != 4 || r.StepsExecuted != 13 || len(r.Entries) != 13 || r.SourceDigest != nil || r.ModelDigest != boundModel {
		t.Fatalf("wrong report summary: %+v", r)
	}
	// These inputs are handwritten, including retention after INVALID_INPUT.
	inputs := []struct {
		state      string
		rev        uint64
		owner      *string
		acceptance []string
	}{
		{"Draft", 0, nil, []string{"café", "✓"}}, {"Ready", 1, nil, []string{"café", "✓"}}, {"InProgress", 2, str("owner-1"), []string{"café", "✓"}}, {"InReview", 3, str("owner-1"), []string{"café", "✓"}},
		{"Ready", 1, nil, []string{"café", "✓"}}, {"Ready", 1, nil, []string{"café", "✓"}}, {"Ready", 1, nil, []string{"café", "✓"}},
		{"InReview", 3, str("owner-1"), []string{"café", "✓"}}, {"InProgress", 4, str("owner-1"), []string{"café", "✓"}}, {"InReview", 5, str("owner-1"), []string{"café", "✓"}}, {"InReview", 5, str("owner-1"), []string{"café", "✓"}}, {"InReview", 5, str("owner-1"), []string{"café", "✓"}},
		{"InReview", 3, str("owner-1"), []string{"Changed"}},
	}
	n := 0
	for _, c := range s.Cases {
		for _, step := range c.Steps {
			entry := r.Entries[n]
			x := inputs[n]
			if entry.CaseID != c.ID || entry.StepID != step.ID || !entry.Matched || len(entry.Differences) != 0 {
				t.Fatalf("wrong ordered entry %d", n)
			}
			if !reflect.DeepEqual(entry.Input, task(x.state, x.rev, x.acceptance, x.owner)) || !reflect.DeepEqual(entry.Command, step.Command) || !reflect.DeepEqual(entry.Expected, step.Expected) {
				t.Fatalf("wrong input/command/expectation %d", n)
			}
			assertResult(t, entry.Actual, step.Expected)
			n++
		}
	}
	if !bytes.Equal(before, snapshot(t, s)) {
		t.Fatal("Run mutated caller suite")
	}
}
func TestFirstMismatchStopsBeforeSentinels(t *testing.T) {
	s := decoded(t)
	s.Cases[0].Steps[0].Expected.Task.State = "Done" // deliberate wrong literal; do not repair it
	before := snapshot(t, s)
	r, e := scenarios.Run(s)
	if e != nil {
		t.Fatal(e)
	}
	if r.Matched || r.StepsExecuted != 1 || r.CasesCompleted != 0 || len(r.Entries) != 1 || r.Entries[0].Matched || len(r.Entries[0].Differences) == 0 {
		t.Fatalf("did not stop at first mismatch: %+v", r)
	}
	// The expected success on the first step is deliberately false; all later
	// steps and cases are sentinels which must have no report entries.
	assertResult(t, r.Entries[0].Actual, scenarios.Expected{OK: true, TransitionID: "T_prepare", Task: func() *lifecycle.Task { x := task("Ready", 1, []string{"café", "✓"}, nil); return &x }(), Refusals: []string{}})
	if !bytes.Equal(before, snapshot(t, s)) {
		t.Fatal("mismatch mutated expectations")
	}
}
func TestTypedWholeSuitePreflight(t *testing.T) {
	cases := map[string]func(*scenarios.Suite){
		"wrong model":              func(s *scenarios.Suite) { s.ModelDigest = strings.Repeat("0", 64) },
		"wrong profile":            func(s *scenarios.Suite) { s.ProfileVersion = "other/v1" },
		"later malformed expected": func(s *scenarios.Suite) { s.Cases[3].Steps[0].Expected.Refusals = nil },
		"later duplicate id":       func(s *scenarios.Suite) { s.Cases[3].Steps[0].ID = s.Cases[0].ID },
		"later expected success missing task": func(s *scenarios.Suite) {
			s.Cases[3].Steps[0].Expected = scenarios.Expected{OK: true, TransitionID: "T_approve", Refusals: []string{}}
		},
		"later too many acceptance": func(s *scenarios.Suite) { s.Cases[3].Initial.Acceptance = make([]string, 65) },
	}
	for name, f := range cases {
		t.Run(name, func(t *testing.T) {
			s := decoded(t)
			f(&s)
			before := snapshot(t, s)
			assertRejected(t, s)
			if !bytes.Equal(before, snapshot(t, s)) {
				t.Fatal("invalid suite mutated")
			}
		})
	}
}
func TestStrictJSONAndDecimalRejection(t *testing.T) {
	raw := fixture(t)
	replace := func(old, new string) []byte {
		if !bytes.Contains(raw, []byte(old)) {
			t.Fatal("mutation anchor missing", old)
		}
		return bytes.Replace(raw, []byte(old), []byte(new), 1)
	}
	invalid := map[string][]byte{
		"duplicate":     replace(`"formatVersion":`, `"id":"duplicate", "formatVersion":`),
		"unknown":       replace(`"formatVersion":`, `"extra":true,"formatVersion":`),
		"missing":       replace(`"expectedRevision": "0",`, ``),
		"null required": replace(`"type": "prepare"`, `"type": null`),
		"null refusals": replace(`"refusals": []`, `"refusals": null`),
		"trailing":      append(append([]byte{}, raw...), []byte(` {}`)...),
		"invalid utf8":  append([]byte{0xff}, raw...),
		"surrogate":     replace(`"outcome": "Ship <&>"`, `"outcome": "\ud800"`),
		"wrong binding": replace(boundModel, strings.Repeat("0", 64)),
		"wrong format":  replace(`lifecycle-scenarios-json/v1`, `lifecycle-scenarios-json/v2`),
		"source bound":  bytes.Repeat([]byte(" "), (1<<20)+1),
	}
	for _, v := range []string{`0`, `"00"`, `"01"`, `"-1"`, `"+1"`, `" 1"`, `"1 "`, `"1.0"`, `"1e0"`, `"18446744073709551616"`, `null`} {
		invalid["decimal "+v] = replace(`"revision": "0"`, `"revision": `+v)
	}
	// A defect in the final case must reject before the first otherwise valid case.
	marker := []byte(`"id": "old_content_review",`)
	if !bytes.Contains(raw, marker) {
		t.Fatal("late mutation anchor missing")
	}
	invalid["late unknown"] = bytes.Replace(raw, marker, []byte(`"extra":true,"id":"old_content_review",`), 1)
	for name, b := range invalid {
		t.Run(name, func(t *testing.T) {
			if _, e := scenarios.Decode(b); e == nil {
				t.Fatal("malformed source decoded")
			}
			r, e := scenarios.Execute(b)
			if e == nil || r.StepsExecuted != 0 || len(r.Entries) != 0 {
				t.Fatal("malformed source produced execution report")
			}
		})
	}
}
func oneCase(id string, initial lifecycle.Task, c lifecycle.Command, w scenarios.Expected) scenarios.Case {
	return scenarios.Case{ID: id, Initial: initial, Steps: []scenarios.Step{{ID: id + "Step", Command: c, Expected: w}}}
}
func TestNilEmptyAndOwnerDistinctions(t *testing.T) {
	s := scenarios.Suite{Version: 1, ID: "Distinctions", ProfileVersion: "lifecycle/v1", ModelDigest: boundModel}
	for _, x := range []struct {
		id string
		a  []string
	}{{"NilAcceptance", nil}, {"EmptyAcceptance", []string{}}} {
		a := task("Draft", 0, x.a, nil)
		want := a
		s.Cases = append(s.Cases, oneCase(x.id, a, lifecycle.Command{Type: "prepare", ExpectedRevision: 0}, scenarios.Expected{TransitionID: "T_prepare", Task: &want, Refusals: []string{"G_acceptance"}}))
	}
	a := task("Ready", 1, []string{"café", "✓"}, str(""))
	want := a
	s.Cases = append(s.Cases, oneCase("PresentEmptyOwner", a, lifecycle.Command{Type: "start", Actor: "owner-1", ExpectedRevision: 1}, scenarios.Expected{TransitionID: "T_start", Task: &want, Refusals: []string{"G_ownerCompatible"}}))
	a = task("Draft", 0, []string{"café", "✓"}, str(""))
	want = task("Ready", 1, []string{"café", "✓"}, str(""))
	s.Cases = append(s.Cases, oneCase("PreserveEmptyOwner", a, lifecycle.Command{Type: "prepare", ExpectedRevision: 0}, scenarios.Expected{OK: true, TransitionID: "T_prepare", Task: &want, Refusals: []string{}}))
	encoded, e := scenarios.Encode(s)
	if e != nil {
		t.Fatal(e)
	}
	round, e := scenarios.Decode(encoded)
	if e != nil || !reflect.DeepEqual(s, round) {
		t.Fatal("wire roundtrip lost nil/empty distinctions", e)
	}
	r, e := scenarios.Run(s)
	if e != nil || !r.Matched || len(r.Entries) != 4 {
		t.Fatal("distinction suite failed", e)
	}
	for i, c := range s.Cases {
		assertResult(t, r.Entries[i].Actual, c.Steps[0].Expected)
	}
}
func TestUint64EdgeAndGuardPrecedence(t *testing.T) {
	a := task("Draft", math.MaxUint64-1, []string{"café", "✓"}, nil)
	ready := task("Ready", math.MaxUint64, []string{"café", "✓"}, nil)
	s := scenarios.Suite{Version: 1, ID: "RevisionBoundary", ProfileVersion: "lifecycle/v1", ModelDigest: boundModel, Cases: []scenarios.Case{{ID: "NearMaximum", Initial: a, Steps: []scenarios.Step{
		{ID: "LastIncrement", Command: lifecycle.Command{Type: "prepare", ExpectedRevision: math.MaxUint64 - 1}, Expected: scenarios.Expected{OK: true, TransitionID: "T_prepare", Task: &ready, Refusals: []string{}}},
		{ID: "LimitRefusal", Command: lifecycle.Command{Type: "start", Actor: "owner-1", ExpectedRevision: math.MaxUint64}, Expected: scenarios.Expected{TransitionID: "T_start", Task: &ready, Refusals: []string{"REVISION_LIMIT"}}},
		{ID: "GuardBeforeLimit", Command: lifecycle.Command{Type: "start", Actor: "owner-1", ExpectedRevision: math.MaxUint64 - 1}, Expected: scenarios.Expected{TransitionID: "T_start", Task: &ready, Refusals: []string{"G_revision"}}},
	}}}}
	raw, e := scenarios.Encode(s)
	if e != nil {
		t.Fatal(e)
	}
	if !bytes.Contains(raw, []byte(`"revision":"18446744073709551615"`)) {
		t.Fatal("maximum revision lost decimal-string encoding")
	}
	r, e := scenarios.Execute(raw)
	if e != nil || !r.Matched || r.StepsExecuted != 3 {
		t.Fatal("uint64 suite failed", e)
	}
	for i, x := range s.Cases[0].Steps {
		assertResult(t, r.Entries[i].Actual, x.Expected)
	}
}
func TestReportDeterminismSourceAndDeepCopies(t *testing.T) {
	raw := fixture(t)
	s := decoded(t)
	before := snapshot(t, s)
	r, e := scenarios.Run(s)
	if e != nil {
		t.Fatal(e)
	}
	b, e := scenarios.EncodeReport(r)
	if e != nil {
		t.Fatal(e)
	}
	for i := 0; i < 3; i++ {
		again, e := scenarios.Run(s)
		if e != nil {
			t.Fatal(e)
		}
		bb, e := scenarios.EncodeReport(again)
		if e != nil || !bytes.Equal(b, bb) {
			t.Fatal("report bytes nondeterministic")
		}
	}
	observed, e := scenarios.Execute(raw)
	if e != nil {
		t.Fatal(e)
	}
	source := fmt.Sprintf("%x", sha256.Sum256(raw))
	if observed.SourceDigest == nil || *observed.SourceDigest != source || observed.ScenarioDigest != r.ScenarioDigest {
		t.Fatal("source/semantic identity confused")
	}
	formatted := append([]byte(" \n"), raw...)
	other, e := scenarios.Execute(formatted)
	if e != nil || other.SourceDigest == nil || *other.SourceDigest == source || other.ScenarioDigest != r.ScenarioDigest {
		t.Fatal("format-only change affected semantic identity", e)
	}
	// Deliberate sentinels mutate every exposed pointer/slice family. They must
	// not alter the supplied suite, another entry, or a subsequent execution.
	r.Entries[0].Input.Acceptance[0] = "input sentinel"
	if r.Entries[0].Expected.Task.Acceptance[0] != "café" || r.Entries[0].Actual.Task.Acceptance[0] != "café" {
		t.Fatal("input aliases expected/actual within entry")
	}
	r.Entries[0].Expected.Task.Acceptance[0] = "expected sentinel"
	if r.Entries[0].Actual.Task.Acceptance[0] != "café" || r.Entries[0].Input.Acceptance[0] != "input sentinel" {
		t.Fatal("expected aliases input/actual within entry")
	}
	r.Entries[0].Actual.Task.Acceptance[0] = "actual sentinel"
	if r.Entries[0].Expected.Task.Acceptance[0] != "expected sentinel" || r.Entries[0].Input.Acceptance[0] != "input sentinel" {
		t.Fatal("actual aliases input/expected within entry")
	}
	*r.Entries[2].Actual.Task.Owner = "owner sentinel"
	if *r.Entries[2].Input.Owner != "owner-1" || *r.Entries[2].Expected.Task.Owner != "owner-1" {
		t.Fatal("actual owner aliases sibling input/expectation")
	}
	r.Entries[3].Command.Review.Reviewer = "review sentinel"
	r.Entries[4].Actual.Refusals[0] = "refusal sentinel"
	if r.Entries[4].Expected.Refusals[0] != "G_actor" {
		t.Fatal("actual refusals alias expected refusals")
	}
	if !bytes.Equal(before, snapshot(t, s)) {
		t.Fatal("report aliases caller suite")
	}
	if r.Entries[1].Input.Acceptance[0] != "café" || r.Entries[3].Input.Owner == nil || *r.Entries[3].Input.Owner != "owner-1" {
		t.Fatal("report entries alias one another")
	}
	again, e := scenarios.Run(s)
	if e != nil {
		t.Fatal(e)
	}
	bb, e := scenarios.EncodeReport(again)
	if e != nil || !bytes.Equal(b, bb) {
		t.Fatal("report mutation contaminated replay")
	}
	// Caller mutations after Run must not change a previously returned report.
	stable := snapshot(t, again)
	s.Cases[0].Initial.Acceptance[0] = "caller sentinel"
	s.Cases[0].Steps[3].Command.Review.Reviewer = "caller review"
	s.Cases[0].Steps[0].Expected.Task.Acceptance[0] = "caller expected"
	if !bytes.Equal(stable, snapshot(t, again)) {
		t.Fatal("report retains caller aliases")
	}
}

// No public reducer-injection/call-observer API exists in the verified ABI.
// Report counters/entries above establish observable stopping/preflight behavior;
// actual zero-extra-call assertions require the lead's package-internal spy tests.

// This vector hashes independently authored compact fixture bytes in the
// documented canonical order, excluding formatVersion and trailing newline.
// Python hashlib/json produced this literal before any runner execution.
const semanticFixtureDigest = "c70c0d5f3a56e706624f3ff8662f41c873ef118b4ab282f8be3eb780c384db9d"

func TestCompleteReportWireAgainstFrozenLiterals(t *testing.T) {
	raw := fixture(t)
	var authored map[string]any
	if e := json.Unmarshal(raw, &authored); e != nil {
		t.Fatal(e)
	}
	// Build only the expected report DTO from frozen authored literals. No actual
	// result or implementation serializer contributes any expected field.
	entries := []any{}
	for _, cv := range authored["cases"].([]any) {
		c := cv.(map[string]any)
		current := c["initial"]
		for _, sv := range c["steps"].([]any) {
			step := sv.(map[string]any)
			expected := step["expected"].(map[string]any)
			actual := map[string]any{"mode": "example-only", "authority": "none", "modelDigest": boundModel, "ok": expected["ok"], "transitionId": expected["transitionId"], "task": expected["task"], "refusals": expected["refusals"]}
			entries = append(entries, map[string]any{"caseId": c["id"], "stepId": step["id"], "input": current, "command": step["command"], "expected": expected, "actual": actual, "matched": true, "differences": []any{}})
			if expected["ok"].(bool) {
				current = expected["task"]
			}
		}
	}
	for _, fromSource := range []bool{false, true} {
		t.Run(fmt.Sprint(fromSource), func(t *testing.T) {
			var r scenarios.Report
			var e error
			var source any
			if fromSource {
				r, e = scenarios.Execute(raw)
				source = fixtureHash
			} else {
				r, e = scenarios.Run(decoded(t))
			}
			if e != nil {
				t.Fatal(e)
			}
			if r.ScenarioDigest != semanticFixtureDigest {
				t.Fatal("semantic digest differs from independent fixed vector")
			}
			b, e := scenarios.EncodeReport(r)
			if e != nil {
				t.Fatal(e)
			}
			var got map[string]any
			if e = json.Unmarshal(b, &got); e != nil {
				t.Fatal(e)
			}
			want := map[string]any{"formatVersion": "lifecycle-scenario-report/v1", "scenarioVersion": "lifecycle-scenarios/v1", "profileVersion": "lifecycle/v1", "mode": "example-only", "authority": "none", "modelDigest": boundModel, "scenarioDigest": semanticFixtureDigest, "sourceDigest": source, "matched": true, "casesCompleted": float64(4), "stepsExecuted": float64(13), "entries": entries}
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("report wire differs from complete authored DTO\ngot %s\nwant %s", b, snapshot(t, want))
			}
		})
	}
}
