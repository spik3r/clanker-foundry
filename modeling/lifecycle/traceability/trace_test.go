package traceability

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"testing"
)

func inputs(t *testing.T) (m, c, r []byte) {
	t.Helper()
	read := func(p string) []byte {
		b, err := os.ReadFile(p)
		if err != nil {
			t.Fatal(err)
		}
		return b
	}
	return read("../examples/lifecycle.json"), read("../examples/oracle.json"), read("requirements.json")
}

func renderAll(t *testing.T, m, c, r []byte) [3][]byte {
	t.Helper()
	rep, err := Build(m, c, r)
	if err != nil {
		t.Fatal(err)
	}
	j, err := rep.JSON()
	if err != nil {
		t.Fatal(err)
	}
	return [3][]byte{j, rep.Markdown(), rep.SVG()}
}

// shuffle reorders requirements and both link lists and re-encodes compactly.
func shuffle(t *testing.T, src []byte, seed int64) []byte {
	t.Helper()
	var doc map[string]any
	if err := json.Unmarshal(src, &doc); err != nil {
		t.Fatal(err)
	}
	rng := rand.New(rand.NewSource(seed))
	reqs := doc["requirements"].([]any)
	rng.Shuffle(len(reqs), func(i, j int) { reqs[i], reqs[j] = reqs[j], reqs[i] })
	for _, x := range reqs {
		q := x.(map[string]any)
		for _, k := range []string{"modelLinks", "acceptanceLinks"} {
			l := q[k].([]any)
			rng.Shuffle(len(l), func(i, j int) { l[i], l[j] = l[j], l[i] })
		}
	}
	out, err := json.Marshal(doc)
	if err != nil {
		t.Fatal(err)
	}
	return out
}

func same(a, b [3][]byte) bool {
	return bytes.Equal(a[0], b[0]) && bytes.Equal(a[1], b[1]) && bytes.Equal(a[2], b[2])
}

func TestDeterministicAcrossRunsAndInputOrder(t *testing.T) {
	m, c, r := inputs(t)
	want := renderAll(t, m, c, r)
	for i := 0; i < 3; i++ {
		if got := renderAll(t, m, c, r); !same(got, want) {
			t.Fatal("repeated run differs")
		}
	}
	for seed := int64(1); seed <= 20; seed++ {
		s := shuffle(t, r, seed)
		if bytes.Equal(s, r) {
			t.Fatal("shuffle produced identical bytes")
		}
		if got := renderAll(t, m, c, s); !same(got, want) {
			t.Fatalf("seed %d: output depends on input order", seed)
		}
	}
}

func TestCommittedViewsMatchGeneration(t *testing.T) {
	m, c, r := inputs(t)
	got := renderAll(t, m, c, r)
	for i, name := range []string{"trace.json", "trace.md", "trace.svg"} {
		b, err := os.ReadFile("view/" + name)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(b, got[i]) {
			t.Fatalf("view/%s is stale; run go run ./cmd/modeltrace generate", name)
		}
	}
}

func TestContentChangeChangesDigestNotModel(t *testing.T) {
	m, c, r := inputs(t)
	a, _ := Build(m, c, r)
	edited := bytes.Replace(r, []byte("Done is final"), []byte("Done is final and closed"), 1)
	b, err := Build(m, c, edited)
	if err != nil {
		t.Fatal(err)
	}
	if a.Inputs.RequirementsDigest == b.Inputs.RequirementsDigest {
		t.Fatal("requirements digest ignored a title change")
	}
	if a.Inputs.ModelDigest != b.Inputs.ModelDigest || a.Inputs.CatalogSHA256 != b.Inputs.CatalogSHA256 {
		t.Fatal("model or catalog identity changed")
	}
}

func issues(t *testing.T, err error) []Issue {
	t.Helper()
	var inv *InvalidError
	if !errors.As(err, &inv) {
		t.Fatalf("want InvalidError, got %v", err)
	}
	return inv.Issues
}

const header = `{"formatVersion":"lifecycle-requirements-json/v1","modelId":"TaskLifecycleExample","acceptanceCatalog":"lifecycle-semantic-oracle-v2","requirements":[%s]}`

func req(id, statement, model, cases string) string {
	return fmt.Sprintf(`{"id":%q,"title":"T","statement":%q,"modelLinks":[%s],"acceptanceLinks":[%s]}`, id, statement, model, cases)
}

func TestBounds(t *testing.T) {
	m, c, _ := inputs(t)
	links := strings.Repeat(`{"kind":"state","id":"Done"},`, MaxLinks+1)
	_, err := Build(m, c, []byte(fmt.Sprintf(header, req("UR-1", "S", strings.TrimSuffix(links, ","), ""))))
	if got := issues(t, err); len(got) != 1 || got[0] != (Issue{"too-many-links", "UR-1", ""}) {
		t.Fatalf("got %v", got)
	}
	many := []string{}
	for i := 0; i <= MaxRequirements; i++ {
		many = append(many, req(fmt.Sprintf("UR-%d", i), fmt.Sprintf("S%d", i), "", ""))
	}
	_, err = Build(m, c, []byte(fmt.Sprintf(header, strings.Join(many, ","))))
	if got := issues(t, err); len(got) != 1 || got[0].Code != "too-many-requirements" {
		t.Fatalf("got %v", got)
	}
	long := strings.Repeat("x", MaxStatement+1)
	_, err = Build(m, c, []byte(fmt.Sprintf(header, req("UR-1", long, "", ""))))
	if got := issues(t, err); len(got) != 1 || got[0].Code != "invalid-statement" {
		t.Fatalf("got %v", got)
	}
	_, err = Build(m, c, []byte(fmt.Sprintf(header, req("UR-1", "line\nbreak", "", ""))))
	if got := issues(t, err); len(got) != 1 || got[0].Code != "invalid-statement" {
		t.Fatalf("control character: %v", got)
	}
	big := append([]byte(fmt.Sprintf(header, req("UR-1", "S", "", ""))), bytes.Repeat([]byte(" "), MaxBytes)...)
	if got := issues(t, func() error { _, e := Build(m, c, big); return e }()); got[0].Code != "malformed-json" {
		t.Fatalf("oversize: %v", got)
	}
}

func TestRequirementIDs(t *testing.T) {
	for id, ok := range map[string]bool{"UR-001": true, "SR-1": true, "ABCDEFGH-123456": true, "ur-1": false, "UR1": false,
		"UR-": false, "-1": false, "UR-1a": false, "ABCDEFGHI-1": false, "UR-1234567": false, "UR-1-2": false, "ÜR-1": false} {
		if RequirementID(id) != ok {
			t.Errorf("%q: want %v", id, ok)
		}
	}
}

func TestAmbiguousStatementsAcrossThree(t *testing.T) {
	m, c, _ := inputs(t)
	in := fmt.Sprintf(header, strings.Join([]string{req("UR-3", "same thing", "", ""), req("UR-1", "Same  Thing", "", ""), req("UR-2", " same thing ", "", "")}, ","))
	_, err := Build(m, c, []byte(in))
	want := []Issue{{"ambiguous-statement", "UR-2", "UR-1"}, {"ambiguous-statement", "UR-3", "UR-1"}}
	if got := issues(t, err); fmt.Sprint(got) != fmt.Sprint(want) {
		t.Fatalf("got %v", got)
	}
}

func TestEveryKindResolves(t *testing.T) {
	m, c, _ := inputs(t)
	in := fmt.Sprintf(header, req("UR-1", "S", `{"kind":"state","id":"Draft"},{"kind":"transition","id":"T_start"},{"kind":"guard","id":"G_owner"},{"kind":"modelRequirement","id":"R_owner"}`, `{"caseId":"refusals-05"}`))
	rep, err := Build(m, c, []byte(in))
	if err != nil {
		t.Fatal(err)
	}
	got := []string{}
	for _, l := range rep.Requirements[0].ModelLinks {
		got = append(got, l.Kind+":"+l.ID)
	}
	if strings.Join(got, " ") != "state:Draft transition:T_start guard:G_owner modelRequirement:R_owner" {
		t.Fatalf("got %v", got)
	}
	for _, e := range rep.ModelElements {
		if e.ID == "R_owner" && (fmt.Sprint(e.TracedBy) != "[UR-1]" || fmt.Sprint(e.TracedViaGuard) != "[UR-1]") {
			t.Fatalf("R_owner: %+v", e)
		}
	}
}

func TestModelAndCatalogErrorsAreNotRequirementIssues(t *testing.T) {
	m, c, r := inputs(t)
	var inv *InvalidError
	if _, err := Build([]byte(`{}`), c, r); err == nil || errors.As(err, &inv) {
		t.Fatalf("bad model: %v", err)
	}
	for name, cat := range map[string]string{
		"not json":     `{`,
		"wrong schema": `{"schema":"other","sequences":[],"refusals":[],"special":[]}`,
		"empty":        `{"schema":"lifecycle-semantic-oracle-v2","sequences":[],"refusals":[],"special":[]}`,
		"duplicate":    `{"schema":"lifecycle-semantic-oracle-v2","sequences":[{"caseId":"a","name":"x","executionStatus":"executable"}],"refusals":[{"caseId":"a","name":"y","executionStatus":"executable"}],"special":[]}`,
		"no caseId":    `{"schema":"lifecycle-semantic-oracle-v2","sequences":[{"name":"x","executionStatus":"executable"}],"refusals":[],"special":[]}`,
		"bad status":   `{"schema":"lifecycle-semantic-oracle-v2","sequences":[{"caseId":"a","name":"x","executionStatus":"done"}],"refusals":[],"special":[]}`,
		"bad caseId":   `{"schema":"lifecycle-semantic-oracle-v2","sequences":[{"caseId":"A B","name":"x","executionStatus":"executable"}],"refusals":[],"special":[]}`,
	} {
		if _, err := Build(m, []byte(cat), r); err == nil || errors.As(err, &inv) {
			t.Errorf("%s: %v", name, err)
		}
	}
}

func TestRenderingEscapesText(t *testing.T) {
	m, c, _ := inputs(t)
	in := fmt.Sprintf(header, req("UR-1", `a | b <script>alert(1)</script> & "q"`, `{"kind":"state","id":"Done"}`, `{"caseId":"refusals-15"}`))
	in = strings.Replace(in, `"title":"T"`, `"title":"<img src=x onerror=1> ünïcödé title that is long enough to be truncated"`, 1)
	rep, err := Build(m, c, []byte(in))
	if err != nil {
		t.Fatal(err)
	}
	md, svg := string(rep.Markdown()), string(rep.SVG())
	for _, s := range []string{md, svg} {
		if strings.Contains(s, "<script") || strings.Contains(s, "<img") {
			t.Fatal("unescaped markup")
		}
	}
	if !strings.Contains(md, `a \| b &lt;script&gt;`) {
		t.Fatal("markdown cell not escaped")
	}
	for _, bad := range []string{"href", "xlink", "<foreignObject", "onload", "<script", "<image"} {
		if strings.Contains(svg, bad) {
			t.Fatalf("svg contains %s", bad)
		}
	}
	if !strings.Contains(svg, "ünïcödé") || !strings.Contains(svg, "…") {
		t.Fatal("svg label not truncated on a rune boundary")
	}
}

func TestInvalidReportsNothingAndValidHasNoNulls(t *testing.T) {
	m, c, r := inputs(t)
	rep, err := Build(m, c, r)
	if err != nil {
		t.Fatal(err)
	}
	j, _ := rep.JSON()
	if bytes.Contains(j, []byte("null")) {
		t.Fatal("report contains null")
	}
	if !bytes.HasSuffix(j, []byte("}\n")) {
		t.Fatal("missing trailing newline")
	}
}
