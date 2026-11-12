package main

import (
	"strings"
	"testing"
)

func TestParseModel(t *testing.T) {
	good := `{"profile":"reference/v1","name":"Abc","items":["b","a"]}`
	bad := map[string]string{
		"empty":          ``,
		"array":          `[]`,
		"empty object":   `{}`,
		"unknown key":    `{"profile":"reference/v1","name":"Abc","items":["a"],"x":1}`,
		"duplicate key":  `{"profile":"reference/v1","name":"Abc","name":"Abd","items":["a"]}`,
		"trailing data":  good + ` {}`,
		"wrong type":     `{"profile":"reference/v1","name":"Abc","items":"a"}`,
		"lowercase":      `{"profile":"reference/v1","name":"abc","items":["a"]}`,
		"duplicate item": `{"profile":"reference/v1","name":"Abc","items":["a","a"]}`,
		"control char":   `{"profile":"reference/v1","name":"Abc","items":["a\u0001"]}`,
		"too many":       `{"profile":"reference/v1","name":"Abc","items":["1","2","3","4","5","6","7","8","9"]}`,
		"long item":      `{"profile":"reference/v1","name":"Abc","items":["` + strings.Repeat("x", 33) + `"]}`,
		"wrong profile":  `{"profile":"other/v1","name":"Abc","items":["a"]}`,
	}
	if _, err := parseModel([]byte(good)); err != nil {
		t.Fatalf("good model refused: %v", err)
	}
	for name, in := range bad {
		if _, err := parseModel([]byte(in)); err == nil {
			t.Errorf("%s: model accepted", name)
		}
	}
}

func TestItemOrderIsInsignificant(t *testing.T) {
	a := loaded{model: model{Profile: profile, Name: "Abc", Items: []string{"a", "b"}}, source: []byte("1")}
	m, err := parseModel([]byte(`{"profile":"reference/v1","name":"Abc","items":["b","a"]}`))
	if err != nil {
		t.Fatal(err)
	}
	b := loaded{model: m, source: []byte("2")}
	sortItems(&b)
	if string(a.canonical()) != string(b.canonical()) {
		t.Errorf("canonical forms differ: %s vs %s", a.canonical(), b.canonical())
	}
}
