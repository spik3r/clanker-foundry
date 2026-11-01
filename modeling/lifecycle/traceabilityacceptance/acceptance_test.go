// Package traceabilityacceptance checks the requirements trace against
// expectations written by hand before the implementation existed. The
// expectations come from the requirement statements, the model and the oracle
// case catalog, not from tool output. Passing here shows the declared links are
// reported as frozen; it does not show that any requirement is correct.
package traceabilityacceptance

import (
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"sort"
	"testing"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/traceability"
)

type issue struct {
	Code        string `json:"code"`
	Requirement string `json:"requirement"`
	Target      string `json:"target"`
}
type status struct {
	Status string   `json:"status"`
	Gaps   []string `json:"gaps"`
}
type expectations struct {
	Example struct {
		Summary                   map[string]int      `json:"summary"`
		Requirements              map[string]status   `json:"requirements"`
		ModelTrace                map[string][]string `json:"modelTrace"`
		ModelRequirementsViaGuard map[string][]string `json:"modelRequirementsViaGuard"`
		AcceptanceTrace           map[string][]string `json:"acceptanceTrace"`
		PendingAcceptanceCases    []string            `json:"pendingAcceptanceCases"`
		UnlinkedAcceptanceCases   []string            `json:"unlinkedAcceptanceCases"`
		UntracedModelElements     []string            `json:"untracedModelElements"`
	} `json:"example"`
	RemovedGuardLink struct {
		Remove struct {
			Requirement, Kind, ID string
		} `json:"remove"`
		UntracedModelElements []string          `json:"untracedModelElements"`
		RequirementStatus     map[string]status `json:"requirementStatus"`
		Summary               map[string]int    `json:"summary"`
	} `json:"removedGuardLink"`
	ClearedRequirementLinks struct {
		Clear                   string            `json:"clear"`
		RequirementStatus       map[string]status `json:"requirementStatus"`
		UnlinkedAcceptanceCases []string          `json:"unlinkedAcceptanceCases"`
		UntracedModelElements   []string          `json:"untracedModelElements"`
	} `json:"clearedRequirementLinks"`
	Invalid []struct {
		Name   string  `json:"name"`
		Input  string  `json:"input"`
		Issues []issue `json:"issues"`
	} `json:"invalid"`
}

func read(t *testing.T, path string) []byte {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func load(t *testing.T) (expectations, []byte, []byte, []byte) {
	t.Helper()
	var e expectations
	if err := json.Unmarshal(read(t, "expected.json"), &e); err != nil {
		t.Fatal(err)
	}
	return e, read(t, "../examples/lifecycle.json"), read(t, "../examples/oracle.json"), read(t, "../traceability/requirements.json")
}

func build(t *testing.T, m, c, r []byte) traceability.Report {
	t.Helper()
	rep, err := traceability.Build(m, c, r)
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	return rep
}

func key(kind, id string) string { return kind + ":" + id }

func untraced(rep traceability.Report) []string {
	out := []string{}
	for _, el := range rep.ModelElements {
		if len(el.TracedBy) == 0 && len(el.TracedViaGuard) == 0 {
			out = append(out, key(el.Kind, el.ID))
		}
	}
	sort.Strings(out)
	return out
}

func unlinked(rep traceability.Report) []string {
	out := []string{}
	for _, c := range rep.AcceptanceCases {
		if len(c.TracedBy) == 0 {
			out = append(out, c.CaseID)
		}
	}
	sort.Strings(out)
	return out
}

func statuses(rep traceability.Report) map[string]status {
	out := map[string]status{}
	for _, r := range rep.Requirements {
		gaps := append([]string{}, r.Gaps...)
		out[r.ID] = status{r.Status, gaps}
	}
	return out
}

func summary(rep traceability.Report) map[string]int {
	s := rep.Summary
	return map[string]int{
		"requirements": s.Requirements, "linked": s.Linked, "incomplete": s.Incomplete,
		"modelElements": s.ModelElements, "untracedModelElements": s.UntracedModelElements,
		"acceptanceCases": s.AcceptanceCases, "pendingAcceptanceCases": s.PendingAcceptanceCases,
		"unlinkedAcceptanceCases": s.UnlinkedAcceptanceCases,
	}
}

func TestExampleTrace(t *testing.T) {
	e, m, c, r := load(t)
	rep := build(t, m, c, r)
	if rep.Mode != "example-only" || rep.Authority != "none" {
		t.Fatalf("metadata: %q %q", rep.Mode, rep.Authority)
	}
	if got := summary(rep); !reflect.DeepEqual(got, e.Example.Summary) {
		t.Fatalf("summary\n got %v\nwant %v", got, e.Example.Summary)
	}
	if got := statuses(rep); !reflect.DeepEqual(got, e.Example.Requirements) {
		t.Fatalf("requirement status\n got %v\nwant %v", got, e.Example.Requirements)
	}
	model := map[string][]string{}
	via := map[string][]string{}
	for _, el := range rep.ModelElements {
		if len(el.TracedBy) > 0 {
			model[key(el.Kind, el.ID)] = el.TracedBy
		}
		if el.Kind == "modelRequirement" {
			via[el.ID] = el.TracedViaGuard
		} else if el.TracedViaGuard != nil {
			t.Fatalf("%s carries via-guard trace", key(el.Kind, el.ID))
		}
	}
	if !reflect.DeepEqual(model, e.Example.ModelTrace) {
		t.Fatalf("model trace\n got %v\nwant %v", model, e.Example.ModelTrace)
	}
	if !reflect.DeepEqual(via, e.Example.ModelRequirementsViaGuard) {
		t.Fatalf("via guard\n got %v\nwant %v", via, e.Example.ModelRequirementsViaGuard)
	}
	cases := map[string][]string{}
	pending := []string{}
	for _, cs := range rep.AcceptanceCases {
		cases[cs.CaseID] = cs.TracedBy
		if cs.ExecutionStatus == "pending" {
			pending = append(pending, cs.CaseID)
		} else if cs.ExecutionStatus != "executable" {
			t.Fatalf("%s status %q", cs.CaseID, cs.ExecutionStatus)
		}
	}
	if !reflect.DeepEqual(cases, e.Example.AcceptanceTrace) {
		t.Fatalf("acceptance trace\n got %v\nwant %v", cases, e.Example.AcceptanceTrace)
	}
	sort.Strings(pending)
	if !reflect.DeepEqual(pending, e.Example.PendingAcceptanceCases) {
		t.Fatalf("pending %v", pending)
	}
	if got := unlinked(rep); !reflect.DeepEqual(got, e.Example.UnlinkedAcceptanceCases) {
		t.Fatalf("unlinked %v", got)
	}
	if got := untraced(rep); !reflect.DeepEqual(got, e.Example.UntracedModelElements) {
		t.Fatalf("untraced %v", got)
	}
}

func edit(t *testing.T, src []byte, f func(reqs []map[string]any)) []byte {
	t.Helper()
	var doc map[string]any
	if err := json.Unmarshal(src, &doc); err != nil {
		t.Fatal(err)
	}
	raw := doc["requirements"].([]any)
	reqs := make([]map[string]any, len(raw))
	for i, x := range raw {
		reqs[i] = x.(map[string]any)
	}
	f(reqs)
	out, err := json.Marshal(doc)
	if err != nil {
		t.Fatal(err)
	}
	return out
}

// A deliberately removed link must surface as an untraced model element while
// the requirement itself stays valid and linked.
func TestRemovedGuardLinkIsReportedNotRejected(t *testing.T) {
	e, m, c, r := load(t)
	x := e.RemovedGuardLink
	removed := 0
	src := edit(t, r, func(reqs []map[string]any) {
		for _, q := range reqs {
			if q["id"] != x.Remove.Requirement {
				continue
			}
			keep := []any{}
			for _, l := range q["modelLinks"].([]any) {
				lm := l.(map[string]any)
				if lm["kind"] == x.Remove.Kind && lm["id"] == x.Remove.ID {
					removed++
					continue
				}
				keep = append(keep, l)
			}
			q["modelLinks"] = keep
		}
	})
	if removed != 1 {
		t.Fatalf("fixture removed %d links", removed)
	}
	rep := build(t, m, c, src)
	if got := untraced(rep); !reflect.DeepEqual(got, x.UntracedModelElements) {
		t.Fatalf("untraced %v", got)
	}
	st := statuses(rep)
	for id, want := range x.RequirementStatus {
		if !reflect.DeepEqual(st[id], want) {
			t.Fatalf("%s: %v", id, st[id])
		}
	}
	if rep.Summary.UntracedModelElements != x.Summary["untracedModelElements"] {
		t.Fatalf("summary %+v", rep.Summary)
	}
}

func TestClearedRequirementIsIncompleteNotInvalid(t *testing.T) {
	e, m, c, r := load(t)
	x := e.ClearedRequirementLinks
	src := edit(t, r, func(reqs []map[string]any) {
		for _, q := range reqs {
			if q["id"] == x.Clear {
				q["modelLinks"] = []any{}
				q["acceptanceLinks"] = []any{}
			}
		}
	})
	rep := build(t, m, c, src)
	st := statuses(rep)
	for id, want := range x.RequirementStatus {
		if !reflect.DeepEqual(st[id], want) {
			t.Fatalf("%s: %v", id, st[id])
		}
	}
	if got := unlinked(rep); !reflect.DeepEqual(got, x.UnlinkedAcceptanceCases) {
		t.Fatalf("unlinked %v", got)
	}
	if got := untraced(rep); !reflect.DeepEqual(got, x.UntracedModelElements) {
		t.Fatalf("untraced %v", got)
	}
}

func TestInvalidInputs(t *testing.T) {
	e, m, c, _ := load(t)
	if len(e.Invalid) < 20 {
		t.Fatalf("only %d frozen cases", len(e.Invalid))
	}
	for _, tc := range e.Invalid {
		t.Run(tc.Name, func(t *testing.T) {
			rep, err := traceability.Build(m, c, []byte(tc.Input))
			if len(tc.Issues) == 0 {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if rep.Summary.Requirements == 0 {
					t.Fatal("empty report")
				}
				return
			}
			var inv *traceability.InvalidError
			if !errors.As(err, &inv) {
				t.Fatalf("want InvalidError, got %v", err)
			}
			got := []issue{}
			for _, i := range inv.Issues {
				got = append(got, issue{i.Code, i.Requirement, i.Target})
			}
			if !reflect.DeepEqual(got, tc.Issues) {
				t.Fatalf("issues\n got %v\nwant %v", got, tc.Issues)
			}
			if !reflect.DeepEqual(rep, traceability.Report{}) {
				t.Fatal("invalid input produced a report")
			}
		})
	}
}

// Invalid UTF-8 cannot be stored as a JSON string in expected.json.
func TestInvalidUTF8(t *testing.T) {
	_, m, c, r := load(t)
	bad := append([]byte{}, r...)
	bad[len(bad)/2] = 0xff
	_, err := traceability.Build(m, c, bad)
	var inv *traceability.InvalidError
	if !errors.As(err, &inv) || len(inv.Issues) != 1 || inv.Issues[0].Code != "malformed-json" {
		t.Fatalf("got %v", err)
	}
}
