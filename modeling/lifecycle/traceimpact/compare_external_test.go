package traceimpact_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/traceimpact"
)

// These literals are frozen before the integration implementation. They describe
// declared-link review candidates, not execution or verified coverage.
func readFile(t *testing.T, name string) []byte {
	t.Helper()
	b, e := os.ReadFile("testdata/" + name)
	if e != nil {
		t.Fatal(e)
	}
	return b
}
func fixture(t *testing.T) (traceimpact.Snapshot, traceimpact.Snapshot, map[string]json.RawMessage) {
	t.Helper()
	before := traceimpact.Snapshot{Model: readFile(t, "before-model.json"), Catalog: readFile(t, "before-catalog.json"), Requirements: readFile(t, "before-requirements.json")}
	after := traceimpact.Snapshot{Model: readFile(t, "after-model.json"), Catalog: readFile(t, "after-catalog.json"), Requirements: readFile(t, "after-requirements.json")}
	var expected map[string]json.RawMessage
	if e := json.Unmarshal(readFile(t, "expected.json"), &expected); e != nil {
		t.Fatal(e)
	}
	return before, after, expected
}
func copySnapshot(s traceimpact.Snapshot) traceimpact.Snapshot {
	return traceimpact.Snapshot{Model: bytes.Clone(s.Model), Catalog: bytes.Clone(s.Catalog), Requirements: bytes.Clone(s.Requirements)}
}
func reportJSON(t *testing.T, r any) []byte {
	t.Helper()
	b, e := json.Marshal(r)
	if e != nil {
		t.Fatal(e)
	}
	return b
}
func sameJSON(t *testing.T, got, want []byte) {
	t.Helper()
	var a, b any
	if e := json.Unmarshal(got, &a); e != nil {
		t.Fatalf("invalid actual JSON: %v", e)
	}
	if e := json.Unmarshal(want, &b); e != nil {
		t.Fatalf("invalid expected JSON: %v", e)
	}
	if !reflect.DeepEqual(a, b) {
		var x, y bytes.Buffer
		_ = json.Indent(&x, got, "", "  ")
		_ = json.Indent(&y, want, "", "  ")
		t.Fatalf("report mismatch\ngot:\n%s\nwant:\n%s", x.String(), y.String())
	}
}
func TestFrozenChangedSnapshots(t *testing.T) {
	before, after, want := fixture(t)
	beforeCopy, afterCopy := copySnapshot(before), copySnapshot(after)
	r, e := traceimpact.Compare(before, after)
	if e != nil {
		t.Fatal(e)
	}
	if !reflect.DeepEqual(before, beforeCopy) || !reflect.DeepEqual(after, afterCopy) {
		t.Fatal("Compare mutated its inputs")
	}
	first := reportJSON(t, r)
	sameJSON(t, first, want["changed"])
	// A returned report must not retain aliases to the six source byte slices.
	for _, s := range []*traceimpact.Snapshot{&before, &after} {
		for _, v := range [][]byte{s.Model, s.Catalog, s.Requirements} {
			if len(v) > 0 {
				v[0] ^= 0xff
			}
		}
	}
	if !bytes.Equal(first, reportJSON(t, r)) {
		t.Fatal("returned report aliases source input")
	}
	again, e := traceimpact.Compare(beforeCopy, afterCopy)
	if e != nil {
		t.Fatal(e)
	}
	if !bytes.Equal(first, reportJSON(t, again)) {
		t.Fatal("repeat output changed")
	}
}
func TestNoOpAndFormattingOnlyCatalog(t *testing.T) {
	before, _, want := fixture(t)
	same, e := traceimpact.Compare(before, copySnapshot(before))
	if e != nil {
		t.Fatal(e)
	}
	sameJSON(t, reportJSON(t, same), want["noOp"])
	after := copySnapshot(before)
	after.Catalog = append(after.Catalog, '\n')
	formatted, e := traceimpact.Compare(before, after)
	if e != nil {
		t.Fatal(e)
	}
	sameJSON(t, reportJSON(t, formatted), want["catalogFormattingOnly"])
}
func reversed(values []any) {
	for i, j := 0, len(values)-1; i < j; i, j = i+1, j-1 {
		values[i], values[j] = values[j], values[i]
	}
}
func reordered(t *testing.T, s traceimpact.Snapshot) traceimpact.Snapshot {
	t.Helper()
	r := copySnapshot(s)
	var m map[string]any
	if e := json.Unmarshal(r.Model, &m); e != nil {
		t.Fatal(e)
	}
	for _, key := range []string{"states", "requirements", "guards", "transitions"} {
		reversed(m[key].([]any))
	}
	for _, v := range m["transitions"].([]any) {
		reversed(v.(map[string]any)["guards"].([]any))
	}
	var e error
	r.Model, e = json.MarshalIndent(m, "", " ")
	if e != nil {
		t.Fatal(e)
	}
	var d map[string]any
	if e = json.Unmarshal(r.Requirements, &d); e != nil {
		t.Fatal(e)
	}
	reqs := d["requirements"].([]any)
	reversed(reqs)
	for _, v := range reqs {
		req := v.(map[string]any)
		reversed(req["modelLinks"].([]any))
		reversed(req["acceptanceLinks"].([]any))
	}
	r.Requirements, e = json.MarshalIndent(d, "", " ")
	if e != nil {
		t.Fatal(e)
	}
	return r
}
func TestSemanticSetsAndSourceFormattingDoNotAffectReport(t *testing.T) {
	before, after, want := fixture(t)
	r, e := traceimpact.Compare(reordered(t, before), reordered(t, after))
	if e != nil {
		t.Fatal(e)
	}
	sameJSON(t, reportJSON(t, r), want["changed"])
}
func TestNoDeclaredCandidatesDoesNotMakeModelChangeSafe(t *testing.T) {
	before, after, _ := fixture(t)
	var req map[string]any
	if e := json.Unmarshal(before.Requirements, &req); e != nil {
		t.Fatal(e)
	}
	for _, v := range req["requirements"].([]any) {
		v.(map[string]any)["modelLinks"] = []any{}
	}
	raw, e := json.Marshal(req)
	if e != nil {
		t.Fatal(e)
	}
	before.Requirements = raw
	after.Requirements = bytes.Clone(raw)
	after.Catalog = bytes.Clone(before.Catalog)
	r, e := traceimpact.Compare(before, after)
	if e != nil {
		t.Fatal(e)
	}
	var got struct {
		ModelChanged   bool              `json:"modelChanged"`
		ReviewRequired bool              `json:"reviewRequired"`
		Candidates     []json.RawMessage `json:"candidates"`
	}
	if e = json.Unmarshal(reportJSON(t, r), &got); e != nil {
		t.Fatal(e)
	}
	if !got.ModelChanged || !got.ReviewRequired || got.Candidates == nil || len(got.Candidates) != 0 {
		t.Fatalf("empty declarations are not safety evidence: %+v", got)
	}
}
func TestCatalogUnknownContentIsConservativelyFlagged(t *testing.T) {
	before, _, _ := fixture(t)
	after := copySnapshot(before)
	after.Catalog = bytes.Replace(after.Catalog, []byte("\"version\":1"), []byte("\"version\":9"), 1)
	if bytes.Equal(after.Catalog, before.Catalog) {
		t.Fatal("catalog fixture mutation failed")
	}
	r, e := traceimpact.Compare(before, after)
	if e != nil {
		t.Fatal(e)
	}
	var got struct {
		ModelChanged        bool `json:"modelChanged"`
		RequirementsChanged bool `json:"requirementsChanged"`
		CatalogChanged      bool `json:"catalogSourceChanged"`
		ReviewRequired      bool `json:"reviewRequired"`
		Candidates          []struct {
			ID      string `json:"id"`
			Reasons []struct {
				Code   string `json:"code"`
				Side   string `json:"side"`
				Target string `json:"target"`
				Cause  string `json:"cause"`
			} `json:"reasons"`
		} `json:"candidates"`
	}
	if e = json.Unmarshal(reportJSON(t, r), &got); e != nil {
		t.Fatal(e)
	}
	if got.ModelChanged || got.RequirementsChanged || !got.CatalogChanged || !got.ReviewRequired || len(got.Candidates) != 3 {
		t.Fatalf("unexpected catalog result: %+v", got)
	}
	for i, id := range []string{"UR-001", "UR-003", "UR-004"} {
		c := got.Candidates[i]
		if c.ID != id || len(c.Reasons) != 1 || c.Reasons[0].Code != "catalog-source-changed" || c.Reasons[0].Side != "both" || c.Reasons[0].Target != "" || c.Reasons[0].Cause != "" {
			t.Fatalf("wrong conservative reason: %+v", c)
		}
	}
}
func TestInvalidSnapshotsRefuseWithoutPartialReport(t *testing.T) {
	before, after, _ := fixture(t)
	dangling := copySnapshot(after)
	var d map[string]any
	if err := json.Unmarshal(dangling.Requirements, &d); err != nil {
		t.Fatal(err)
	}
	found := false
	for _, item := range d["requirements"].([]any) {
		req := item.(map[string]any)
		if req["id"] == "UR-001" {
			req["modelLinks"].([]any)[0].(map[string]any)["id"] = "Missing"
			found = true
		}
	}
	if !found {
		t.Fatal("dangling fixture mutation failed")
	}
	var err error
	dangling.Requirements, err = json.Marshal(d)
	if err != nil {
		t.Fatal(err)
	}
	badCatalog := copySnapshot(after)
	badCatalog.Catalog = bytes.Replace(badCatalog.Catalog, []byte("\"caseId\":\"case-2\""), []byte("\"caseId\":\"missing\""), 1)
	var cases = []struct {
		name string
		b, a traceimpact.Snapshot
	}{
		{"dangling after", copySnapshot(before), dangling},
		{"dangling before", dangling, copySnapshot(after)},
		{"dangling catalog case", copySnapshot(before), badCatalog},
		{"bad model", traceimpact.Snapshot{Model: []byte("{"), Catalog: bytes.Clone(before.Catalog), Requirements: bytes.Clone(before.Requirements)}, copySnapshot(after)},
		{"bad catalog", copySnapshot(before), traceimpact.Snapshot{Model: bytes.Clone(after.Model), Catalog: []byte("{"), Requirements: bytes.Clone(after.Requirements)}},
		{"bad requirements", copySnapshot(before), traceimpact.Snapshot{Model: bytes.Clone(after.Model), Catalog: bytes.Clone(after.Catalog), Requirements: []byte("{")}},
		{"oversized model", traceimpact.Snapshot{Model: bytes.Repeat([]byte(" "), 1<<20+1), Catalog: bytes.Clone(before.Catalog), Requirements: bytes.Clone(before.Requirements)}, copySnapshot(after)},
		{"oversized requirements", traceimpact.Snapshot{Model: bytes.Clone(before.Model), Catalog: bytes.Clone(before.Catalog), Requirements: bytes.Repeat([]byte(" "), 1<<20+1)}, copySnapshot(after)},
		{"oversized catalog", traceimpact.Snapshot{Model: bytes.Clone(before.Model), Catalog: bytes.Repeat([]byte(" "), 4<<20+1), Requirements: bytes.Clone(before.Requirements)}, copySnapshot(after)},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			b, a := copySnapshot(c.b), copySnapshot(c.a)
			r, e := traceimpact.Compare(c.b, c.a)
			if e == nil || !reflect.ValueOf(r).IsZero() {
				t.Fatalf("invalid input produced a partial report: %+v, %v", r, e)
			}
			if !reflect.DeepEqual(c.b, b) || !reflect.DeepEqual(c.a, a) {
				t.Fatal("invalid input mutated")
			}
		})
	}
}
func TestRequirementsSourceOnlyFormattingIsIgnored(t *testing.T) {
	before, _, want := fixture(t)
	after := copySnapshot(before)
	after.Requirements = append([]byte(" \n"), after.Requirements...)
	after.Requirements = append(after.Requirements, []byte("\n \t")...)
	if !strings.Contains(string(after.Requirements), "\n") {
		t.Fatal("fixture edit failed")
	}
	r, e := traceimpact.Compare(before, after)
	if e != nil {
		t.Fatal(e)
	}
	sameJSON(t, reportJSON(t, r), want["noOp"])
}

func TestCatalogCaseCountLimit(t *testing.T) {
	before, _, _ := fixture(t)
	makeCatalog := func(count int) []byte {
		entries := make([]any, 0, count)
		for i := 0; i < count; i++ {
			entries = append(entries, map[string]any{"caseId": fmt.Sprintf("case-%d", i+1), "name": "Example", "executionStatus": "pending"})
		}
		b, e := json.Marshal(map[string]any{"schema": "lifecycle-semantic-oracle-v2", "sequences": entries, "refusals": []any{}, "special": []any{}})
		if e != nil {
			t.Fatal(e)
		}
		return b
	}
	before.Catalog = makeCatalog(traceimpact.MaxCatalogCases)
	if _, e := traceimpact.Compare(before, before); e != nil {
		t.Fatalf("limit should be accepted: %v", e)
	}
	before.Catalog = makeCatalog(traceimpact.MaxCatalogCases + 1)
	got, e := traceimpact.Compare(before, before)
	if e == nil || !reflect.ValueOf(got).IsZero() {
		t.Fatal("oversized case count did not fail without partial report")
	}
}

func TestReasonBudgetRefusesWithoutPartialReport(t *testing.T) {
	guards := make([]any, 0, 64)
	transitions := make([]any, 0, 32)
	ids := make([]any, 0, 64)
	for i := 0; i < 64; i++ {
		id := fmt.Sprintf("G%02d", i)
		ids = append(ids, id)
		guards = append(guards, map[string]any{"id": id, "requirement": "Rrule", "expr": map[string]any{"op": "nonempty", "args": []any{map[string]any{"ref": "command.actor"}}}})
	}
	links := make([]any, 0, 32)
	for i := 0; i < 32; i++ {
		id := fmt.Sprintf("T%02d", i)
		transitions = append(transitions, map[string]any{"id": id, "from": "Draft", "trigger": fmt.Sprintf("cmd%02d", i), "to": "Done", "guards": ids, "effects": []any{map[string]any{"op": "set_state"}, map[string]any{"op": "increment_revision"}}})
		links = append(links, map[string]any{"kind": "transition", "id": id})
	}
	doc := map[string]any{"formatVersion": "lifecycle-json/v1", "profileVersion": "lifecycle/v1", "id": "Flow", "initial": "Draft", "states": []any{map[string]any{"id": "Draft", "label": "Draft", "terminal": false}, map[string]any{"id": "Done", "label": "Done", "terminal": true}}, "requirements": []any{map[string]any{"id": "Rrule", "text": "Rule"}}, "guards": guards, "transitions": transitions}
	beforeModel, e := json.Marshal(doc)
	if e != nil {
		t.Fatal(e)
	}
	for _, item := range guards {
		item.(map[string]any)["expr"].(map[string]any)["args"].([]any)[0] = map[string]any{"ref": "task.outcome"}
	}
	afterModel, e := json.Marshal(doc)
	if e != nil {
		t.Fatal(e)
	}
	requirements := make([]any, 0, 256)
	for i := 0; i < 256; i++ {
		requirements = append(requirements, map[string]any{"id": fmt.Sprintf("UR-%03d", i), "title": fmt.Sprintf("Title %d", i), "statement": fmt.Sprintf("Unique intent %d", i), "modelLinks": links, "acceptanceLinks": []any{}})
	}
	req, e := json.Marshal(map[string]any{"formatVersion": "lifecycle-requirements-json/v1", "modelId": "Flow", "acceptanceCatalog": "lifecycle-semantic-oracle-v2", "requirements": requirements})
	if e != nil {
		t.Fatal(e)
	}
	cat := []byte(`{"schema":"lifecycle-semantic-oracle-v2","sequences":[{"caseId":"case-1","name":"Example","executionStatus":"pending"}],"refusals":[],"special":[]}`)
	before := traceimpact.Snapshot{Model: beforeModel, Catalog: cat, Requirements: req}
	after := traceimpact.Snapshot{Model: afterModel, Catalog: cat, Requirements: req}
	got, e := traceimpact.Compare(before, after)
	if !errors.Is(e, traceimpact.ErrReportLimit) || !reflect.ValueOf(got).IsZero() {
		t.Fatalf("reason budget produced partial result or wrong error: %v", e)
	}
}
