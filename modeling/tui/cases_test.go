package tui

import (
	"errors"
	"os"
	"strings"
	"testing"
)

const baseCases = `{
  "formatVersion": "record-browser-cases/v1",
  "profileVersion": "record-browser/v1",
  "model": {"canonicalSHA256": "6a8370de19f321f4a8bb0737745d04d86aa14e30203259d95fe5d79b27a8ac53"},
  "records": {"Main": [
    {"ID": "a", "Name": "Alpha", "Description": "d", "Status": "s"},
    {"ID": "b", "Name": "Beta", "Description": "d", "Status": "s"}
  ]},
  "cases": [{
    "id": "C_one",
    "requirement": "Enter shows details.",
    "size": {"width": 100, "height": 20},
    "records": "Main",
    "steps": [{"key": "j"}, {"key": "enter"}],
    "expect": {"selected": "b", "visible": ["Name: Beta"]}
  }]
}`

func TestBaseCasesDecode(t *testing.T) {
	if _, err := DecodeCases([]byte(baseCases)); err != nil {
		t.Fatal(err)
	}
}

func TestExampleCasesDecodeAndBind(t *testing.T) {
	for _, p := range []struct{ model, cases string }{
		{"examples/browser.json", "examples/browser.cases.json"},
		{"examples/browser-package.json", "examples/browser-package.cases.json"},
		{"examples/browser-hosted.json", "examples/browser-hosted.cases.json"},
	} {
		raw, err := os.ReadFile(p.model)
		if err != nil {
			t.Fatal(err)
		}
		res, err := Build(raw)
		if err != nil {
			t.Fatal(err)
		}
		m, err := Decode(raw)
		if err != nil {
			t.Fatal(err)
		}
		craw, err := os.ReadFile(p.cases)
		if err != nil {
			t.Fatal(err)
		}
		c, err := DecodeCases(craw)
		if err != nil {
			t.Fatalf("%s: %v", p.cases, err)
		}
		if err := c.CheckAgainst(m, CanonicalSHA256(res.Files)); err != nil {
			t.Fatalf("%s: %v", p.cases, err)
		}
		if n := len(c.Cases); n < 10 || n > 15 {
			t.Fatalf("%s has %d cases", p.cases, n)
		}
	}
}

func TestDecodeCasesRefusals(t *testing.T) {
	sub := func(old, new string) string {
		if strings.Count(baseCases, old) != 1 {
			t.Fatalf("test setup: %q occurs %d times", old, strings.Count(baseCases, old))
		}
		return strings.Replace(baseCases, old, new, 1)
	}
	manySteps := strings.Repeat(`{"key": "j"},`, 64) + `{"key": "j"}`
	var sets []string
	for i := 0; i < 9; i++ {
		sets = append(sets, `"S`+string(rune('a'+i))+`": []`)
	}
	var recs []string
	for i := 0; i < 201; i++ {
		recs = append(recs, `{"ID": "`+strings.Repeat("x", i%7+1)+string(rune('a'+i%26))+string(rune('a'+i/26))+`"}`)
	}
	var texts []string
	for i := 0; i < 17; i++ {
		texts = append(texts, `"t"`)
	}
	var cases []string
	for i := 0; i < 201; i++ {
		cases = append(cases, `{"id":"C`+strings.Repeat("a", 1)+string(rune('a'+i%26))+string(rune('a'+i/26))+`","requirement":"r","size":{"width":80,"height":12},"records":"Main","steps":[],"expect":{"help":true}}`)
	}
	big := sub(`"Description": "d", "Status": "s"},
    {"ID": "b"`, `"Description": "`+strings.Repeat("x", 300<<10)+`", "Status": "s"},
    {"ID": "b"`)
	bad := map[string]string{
		"unknown member":                 sub(`"cases": [{`, `"extra": 1, "cases": [{`),
		"case variant member":            sub(`"cases": [{`, `"Cases": [], "cases": [{`),
		"case variant in case":           sub(`"id": "C_one"`, `"ID": "C_one"`),
		"unknown in case":                sub(`"records": "Main",`, `"records": "Main", "note": "x",`),
		"unknown in expect":              sub(`"selected": "b"`, `"selection": "b"`),
		"unknown in size":                sub(`"width": 100, "height": 20`, `"width": 100, "height": 20, "depth": 1`),
		"duplicate key":                  sub(`"id": "C_one",`, `"id": "C_one", "id": "C_two",`),
		"duplicate record key":           sub(`{"ID": "a", "Name": "Alpha"`, `{"ID": "a", "ID": "z", "Name": "Alpha"`),
		"duplicate record set":           sub(`"records": {"Main": [`, `"records": {"Main": [], "Main": [`),
		"null":                           sub(`"selected": "b"`, `"selected": null`),
		"null case list":                 sub(`"steps": [{"key": "j"}, {"key": "enter"}]`, `"steps": null`),
		"trailing data":                  baseCases + ` {}`,
		"trailing garbage":               baseCases + "x",
		"invalid UTF-8":                  sub(`"Alpha"`, "\"Al\xffpha\""),
		"over 256 KiB":                   big,
		"empty input":                    "",
		"wrong formatVersion":            sub(`record-browser-cases/v1`, `record-browser-cases/v2`),
		"wrong profileVersion":           sub(`"profileVersion": "record-browser/v1"`, `"profileVersion": "record-browser/v2"`),
		"short digest":                   sub(`6a8370de19f321f4a8bb0737745d04d86aa14e30203259d95fe5d79b27a8ac53`, `6a8370de`),
		"uppercase digest":               sub(`6a8370de19f321f4a8bb0737745d04d86aa14e30203259d95fe5d79b27a8ac53`, `6A8370DE19F321F4A8BB0737745D04D86AA14E30203259D95FE5D79B27A8AC53`),
		"no record sets":                 baseCases[:strings.Index(baseCases, `"records"`)] + `"records": {}, ` + baseCases[strings.Index(baseCases, `"cases"`):],
		"nine record sets":               sub(`"records": {"Main": [`, `"records": {`+strings.Join(sets, ",")+`, "Main": [`),
		"bad set name":                   sub(`"records": {"Main": [`, `"records": {"1Main": [`),
		"201 records":                    sub(`"records": {"Main": [`, `"records": {"Main": [`+strings.Join(recs, ",")+`,`),
		"record value not string":        sub(`"Name": "Alpha"`, `"Name": 5`),
		"no cases":                       baseCases[:strings.Index(baseCases, `"cases"`)] + `"cases": []}`,
		"201 cases":                      sub(`"cases": [{`, `"cases": [`+strings.Join(cases, ",")+`,{`),
		"duplicate case id":              sub(`"cases": [{`, `"cases": [{"id":"C_one","requirement":"r","size":{"width":80,"height":12},"records":"Main","steps":[],"expect":{"help":true}},{`),
		"case id with space":             sub(`"C_one"`, `"C one"`),
		"case id starts digit":           sub(`"C_one"`, `"1_one"`),
		"case id of 65":                  sub(`"C_one"`, `"C`+strings.Repeat("a", 64)+`"`),
		"empty requirement":              sub(`"Enter shows details."`, `""`),
		"requirement of 241":             sub(`"Enter shows details."`, `"`+strings.Repeat("x", 241)+`"`),
		"requirement newline":            sub(`"Enter shows details."`, `"a\nb"`),
		"size zero":                      sub(`"width": 100`, `"width": 0`),
		"size too large":                 sub(`"height": 20`, `"height": 201`),
		"size as float":                  sub(`"width": 100`, `"width": 100.5`),
		"unknown record set":             sub(`"records": "Main",`, `"records": "Other",`),
		"steps missing":                  sub(`"steps": [{"key": "j"}, {"key": "enter"}],`, ``),
		"65 steps":                       sub(`[{"key": "j"}, {"key": "enter"}]`, `[`+manySteps+`]`),
		"empty step":                     sub(`{"key": "j"}`, `{}`),
		"step with two members":          sub(`{"key": "j"}`, `{"key": "j", "resize": {"width": 30, "height": 8}}`),
		"unknown key":                    sub(`{"key": "j"}`, `{"key": "x"}`),
		"undocumented alias key":         sub(`{"key": "j"}`, `{"key": "G"}`),
		"key case variant":               sub(`{"key": "j"}`, `{"key": "Enter"}`),
		"empty key":                      sub(`{"key": "j"}`, `{"key": ""}`),
		"resize zero":                    sub(`{"key": "j"}`, `{"resize": {"width": 0, "height": 5}}`),
		"resize missing height":          sub(`{"key": "j"}`, `{"resize": {"width": 30}}`),
		"replace bad status":             sub(`{"key": "j"}`, `{"replace": {"status": "done"}}`),
		"replace missing status":         sub(`{"key": "j"}`, `{"replace": {}}`),
		"replace loading records":        sub(`{"key": "j"}`, `{"replace": {"status": "loading", "records": "Main"}}`),
		"replace unknown records":        sub(`{"key": "j"}`, `{"replace": {"status": "ready", "records": "Nope"}}`),
		"replace empty message":          sub(`{"key": "j"}`, `{"replace": {"status": "failed", "message": ""}}`),
		"replace message control":        sub(`{"key": "j"}`, `{"replace": {"status": "failed", "message": "a\u0007b"}}`),
		"expect empty":                   sub(`{"selected": "b", "visible": ["Name: Beta"]}`, `{}`),
		"expect bad layout":              sub(`"selected": "b"`, `"layout": "huge"`),
		"expect bad focus":               sub(`"selected": "b"`, `"focus": "both"`),
		"expect bad status":              sub(`"selected": "b"`, `"status": "idle"`),
		"expect help not bool":           sub(`"selected": "b"`, `"help": "yes"`),
		"expect visible empty":           sub(`"visible": ["Name: Beta"]`, `"visible": []`),
		"expect 17 visible":              sub(`"visible": ["Name: Beta"]`, `"visible": [`+strings.Join(texts, ",")+`]`),
		"expect visible empty str":       sub(`"visible": ["Name: Beta"]`, `"visible": [""]`),
		"expect absent control":          sub(`"visible": ["Name: Beta"]`, `"absent": ["a\tb"]`),
		"expect visible 241 bytes":       sub(`"visible": ["Name: Beta"]`, `"visible": ["`+strings.Repeat("x", 241)+`"]`),
		"expect capturesKeys not bool":   sub(`"selected": "b"`, `"capturesKeys": "yes"`),
		"expect capturesKeys null":       sub(`"selected": "b"`, `"capturesKeys": null`),
		"expect keyHints empty":          sub(`"selected": "b"`, `"keyHints": []`),
		"expect 13 keyHints":             sub(`"selected": "b"`, `"keyHints": [`+strings.TrimSuffix(strings.Repeat(`{"key":"?","action":"help"},`, 13), ",")+`]`),
		"expect keyHints not objects":    sub(`"selected": "b"`, `"keyHints": ["?"]`),
		"expect keyHints missing action": sub(`"selected": "b"`, `"keyHints": [{"key":"?"}]`),
		"expect keyHints empty key":      sub(`"selected": "b"`, `"keyHints": [{"key":"","action":"help"}]`),
		"expect keyHints empty action":   sub(`"selected": "b"`, `"keyHints": [{"key":"?","action":""}]`),
		"expect keyHints unknown member": sub(`"selected": "b"`, `"keyHints": [{"key":"?","action":"help","x":1}]`),
		"expect keyHints control":        sub(`"selected": "b"`, `"keyHints": [{"key":"?","action":"a\tb"}]`),
		"expect keyHints duplicate key":  sub(`"selected": "b"`, `"keyHints": [{"key":"?","key":"q","action":"help"}]`),
		"expect keyHints case variant":   sub(`"selected": "b"`, `"KeyHints": [{"key":"?","action":"help"}]`),
		"expect keyHints 241 bytes":      sub(`"selected": "b"`, `"keyHints": [{"key":"?","action":"`+strings.Repeat("x", 241)+`"}]`),
		"deep nesting":                   sub(`"expect": {`, `"expect": {"visible": [[[[[[[[[[1]]]]]]]]]], `),
	}
	for name, text := range bad {
		if _, err := DecodeCases([]byte(text)); err == nil {
			t.Errorf("%s: accepted", name)
		} else if testing.Verbose() {
			t.Logf("%s: %v", name, err)
		}
	}
	if len(bad) < 50 {
		t.Fatalf("only %d refusals", len(bad))
	}
}

func TestHostedExpectationsDecode(t *testing.T) {
	good := strings.Replace(baseCases, `"selected": "b"`, `"capturesKeys": false, "keyHints": [{"key":"j/k","action":"select"},{"key":"?","action":"help"}]`, 1)
	c, err := DecodeCases([]byte(good))
	if err != nil {
		t.Fatal(err)
	}
	e := c.Cases[0].Expect
	if e.CapturesKeys == nil || *e.CapturesKeys || e.KeyHints == nil || len(*e.KeyHints) != 2 || (*e.KeyHints)[1].Action != "help" {
		t.Fatalf("expectations not decoded: %+v", e)
	}
	only := strings.Replace(baseCases, `"selected": "b", "visible": ["Name: Beta"]`, `"capturesKeys": true`, 1)
	if _, err := DecodeCases([]byte(only)); err != nil {
		t.Fatalf("capturesKeys alone is an expectation: %v", err)
	}
	twelve := strings.TrimSuffix(strings.Repeat(`{"key":"?","action":"help"},`, 12), ",")
	if _, err := DecodeCases([]byte(strings.Replace(baseCases, `"selected": "b"`, `"keyHints": [`+twelve+`]`, 1))); err != nil {
		t.Fatalf("12 hints are allowed: %v", err)
	}
}

func TestAllowedKeysAreTheDocumentedKeys(t *testing.T) {
	want := "tab shift+tab enter esc ? q ctrl+c j k up down pgup pgdown home end"
	if got := strings.Join(AllowedKeys, " "); got != want {
		t.Fatalf("allowed keys %q, want %q", got, want)
	}
	for _, k := range AllowedKeys {
		text := strings.Replace(baseCases, `{"key": "j"}`, `{"key": "`+k+`"}`, 1)
		if _, err := DecodeCases([]byte(text)); err != nil {
			t.Errorf("key %q refused: %v", k, err)
		}
	}
}

func TestCheckAgainstModel(t *testing.T) {
	raw, err := os.ReadFile("examples/browser.json")
	if err != nil {
		t.Fatal(err)
	}
	res, err := Build(raw)
	if err != nil {
		t.Fatal(err)
	}
	m, _ := Decode(raw)
	digest := CanonicalSHA256(res.Files)
	if digest != "6a8370de19f321f4a8bb0737745d04d86aa14e30203259d95fe5d79b27a8ac53" {
		t.Fatalf("digest %s", digest)
	}
	cases := func(old, new string) Cases {
		c, err := DecodeCases([]byte(strings.Replace(baseCases, old, new, 1)))
		if err != nil {
			t.Fatal(err)
		}
		return c
	}
	if err := cases("x", "x").CheckAgainst(m, digest); err != nil {
		t.Fatal(err)
	}
	var stale StaleError
	err = cases("x", "x").CheckAgainst(m, strings.Repeat("0", 64))
	if !errors.As(err, &stale) || !strings.Contains(err.Error(), "stale") {
		t.Fatalf("digest mismatch: %v", err)
	}
	for name, c := range map[string]Cases{
		"missing field":      cases(`"Name": "Alpha", `, ``),
		"extra field":        cases(`"Name": "Alpha", `, `"Name": "Alpha", "Extra": "x", `),
		"case variant field": cases(`"Name": "Alpha"`, `"name": "Alpha"`),
		"duplicate identity": cases(`"ID": "b"`, `"ID": "a"`),
		"empty identity":     cases(`"ID": "b"`, `"ID": ""`),
		"value over 4096":    cases(`"Name": "Beta"`, `"Name": "`+strings.Repeat("x", 4097)+`"`),
	} {
		err := c.CheckAgainst(m, digest)
		if err == nil || errors.As(err, &stale) {
			t.Errorf("%s: %v", name, err)
		}
	}
}
