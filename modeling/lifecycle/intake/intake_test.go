package intake

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
)

// These literals are stated independently of Execute. "abc" has the known
// SHA-256 below; the suite assertion deliberately binds different bytes.
const sample = `{"formatVersion":"lifecycle-intake-json/v1","intent":"Keep the original request.","decisions":[{"id":"D_z","text":"Later decision","rationale":"Reason Z"},{"id":"D_a","text":"Earlier decision","rationale":"Reason A"}],"questions":[{"id":"Q_b","text":"Still unresolved?"}],"artifacts":{"requirements":{"sha256":"ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad","author":"Requirement author"},"model":{"sha256":"ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad","author":"Model author"},"scenarioSuite":{"sha256":"ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad","author":"Scenario author"}},"dependencies":{"modelRequirements":null,"suiteRequirements":null,"suiteModel":null},"observed":{"requirements":"abc","model":null,"scenarioSuite":"changed"}}`

func TestExplicitBindingsAndIntent(t *testing.T) {
	raw := []byte(sample)
	copyRaw := append([]byte(nil), raw...)
	out, err := Execute(raw)
	if err != nil {
		t.Fatal(err)
	}
	var got report
	if err := json.Unmarshal(out, &got); err != nil {
		t.Fatal(err)
	}
	if got.FormatVersion != ReportVersion || got.Authority != "none" || got.Intent != "Keep the original request." {
		t.Fatalf("header: %+v", got)
	}
	if len(got.Decisions) != 2 || got.Decisions[0].ID != "D_a" || got.Decisions[1].ID != "D_z" || len(got.Questions) != 1 || got.Questions[0].Text != "Still unresolved?" {
		t.Fatalf("intent record: %+v", got)
	}
	expected := []struct {
		kind, status string
		observed     *string
	}{
		{"requirements", "matches-declaration", str("ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad")},
		{"model", "missing-observation", nil},
		{"scenarioSuite", "digest-mismatch", str("d67e2e944994496c8d8ec76eed0cf9f09679448d584b532bebf941852a37f5ed")},
	}
	if len(got.Bindings) != 3 {
		t.Fatal(got.Bindings)
	}
	for i, w := range expected {
		g := got.Bindings[i]
		if g.Kind != w.kind || g.Status != w.status || !same(g.ObservedSHA256, w.observed) {
			t.Fatalf("binding %d: %+v", i, g)
		}
	}
	out2, err := Execute(raw)
	if err != nil || !bytes.Equal(out, out2) || !bytes.Equal(raw, copyRaw) {
		t.Fatal("not deterministic or mutated input")
	}
}
func str(s string) *string   { return &s }
func same(a, b *string) bool { return a == nil && b == nil || a != nil && b != nil && *a == *b }
func TestMissingAndUndeclared(t *testing.T) {
	raw := `{"formatVersion":"lifecycle-intake-json/v1","intent":"Intent","decisions":[],"questions":[],"artifacts":{"requirements":null,"model":null,"scenarioSuite":null},"dependencies":{"modelRequirements":null,"suiteRequirements":null,"suiteModel":null},"observed":{"requirements":null,"model":"","scenarioSuite":null}}`
	b, e := Execute([]byte(raw))
	if e != nil {
		t.Fatal(e)
	}
	var r report
	if e = json.Unmarshal(b, &r); e != nil {
		t.Fatal(e)
	}
	if r.Bindings[0].Status != "missing-declaration" || r.Bindings[1].Status != "undeclared" || !same(r.Bindings[1].ObservedSHA256, str("e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855")) || r.Bindings[2].Status != "missing-declaration" {
		t.Fatalf("%+v", r.Bindings)
	}
}
func TestStrictRejections(t *testing.T) {
	cases := map[string]string{
		"duplicate key":      strings.Replace(sample, `"intent":"Keep`, `"intent":"extra","intent":"Keep`, 1),
		"unknown key":        strings.Replace(sample, `"intent":"Keep`, `"extra":true,"intent":"Keep`, 1),
		"missing field":      strings.Replace(sample, `"questions":[{"id":"Q_b","text":"Still unresolved?"}],`, "", 1),
		"wrong case":         strings.Replace(sample, `"formatVersion"`, `"FormatVersion"`, 1),
		"null list":          strings.Replace(sample, `"decisions":[{"id":"D_z","text":"Later decision","rationale":"Reason Z"},{"id":"D_a","text":"Earlier decision","rationale":"Reason A"}]`, `"decisions":null`, 1),
		"blank intent":       strings.Replace(sample, `Keep the original request.`, `   `, 1),
		"duplicate ID":       strings.Replace(sample, `Q_b`, `D_z`, 1),
		"wrong version":      strings.Replace(sample, `lifecycle-intake-json/v1`, `lifecycle-intake-json/v2`, 1),
		"digest uppercase":   strings.Replace(sample, `ba7816`, `BA7816`, 1),
		"blank author":       strings.Replace(sample, `Requirement author`, ` `, 1),
		"surrogate":          strings.Replace(sample, `Keep the original request.`, `\ud800`, 1),
		"trailing":           sample + ` {}`,
		"oversized artifact": strings.Replace(sample, `"requirements":"abc"`, `"requirements":"`+strings.Repeat("a", MaxArtifactBytes+1)+`"`, 1),
		"oversized request":  strings.Repeat(" ", MaxSourceBytes+1),
	}
	for name, input := range cases {
		t.Run(name, func(t *testing.T) {
			b, e := Execute([]byte(input))
			if !errors.Is(e, ErrInvalid) || b != nil {
				t.Fatalf("got %q %v", b, e)
			}
		})
	}
}
func TestExactBytesAndOrdering(t *testing.T) {
	modified := strings.Replace(sample, `"requirements":"abc"`, `"requirements":"abc\n"`, 1)
	b, e := Execute([]byte(modified))
	if e != nil {
		t.Fatal(e)
	}
	var r report
	_ = json.Unmarshal(b, &r)
	if r.Bindings[0].Status != "digest-mismatch" {
		t.Fatal(r.Bindings[0])
	}
	if r.RequestSHA256 == "" {
		t.Fatal("missing raw request identity")
	}
	if _, e = Execute([]byte(fmt.Sprintf("%s ", sample))); e != nil {
		t.Fatal(e)
	}
}

func TestDeclaredDependencyLinks(t *testing.T) {
	const original = `"dependencies":{"modelRequirements":null,"suiteRequirements":null,"suiteModel":null}`
	const specified = `"dependencies":{"modelRequirements":"ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad","suiteRequirements":"e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855","suiteModel":null}`
	input := strings.Replace(sample, original, specified, 1)
	out, err := Execute([]byte(input))
	if err != nil {
		t.Fatal(err)
	}
	var r report
	if err = json.Unmarshal(out, &r); err != nil {
		t.Fatal(err)
	}
	if len(r.Links) != 3 || r.Links[0].From != "model" || r.Links[0].To != "requirements" || r.Links[0].Status != "matches-declaration" || r.Links[1].Status != "digest-mismatch" || r.Links[2].Status != "missing-assertion" {
		t.Fatalf("links: %+v", r.Links)
	}
	missing := strings.Replace(input, `"scenarioSuite":{"sha256":"ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad","author":"Scenario author"}`, `"scenarioSuite":null`, 1)
	out, err = Execute([]byte(missing))
	if err != nil {
		t.Fatal(err)
	}
	_ = json.Unmarshal(out, &r)
	if r.Links[1].Status != "missing-artifact-declaration" {
		t.Fatalf("link: %+v", r.Links[1])
	}
	invalid := strings.Replace(input, `"modelRequirements":"ba7816`, `"modelRequirements":"BA7816`, 1)
	if b, e := Execute([]byte(invalid)); !errors.Is(e, ErrInvalid) || b != nil {
		t.Fatal("accepted malformed dependency digest")
	}
}
