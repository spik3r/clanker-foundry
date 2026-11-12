package tui

import (
	"bytes"
	"encoding/json"
	"os"
	"reflect"
	"testing"
)

func sampleModel(t *testing.T) Model {
	t.Helper()
	raw, err := os.ReadFile("examples/browser.json")
	if err != nil {
		t.Fatal(err)
	}
	m, err := Decode(raw)
	if err != nil {
		t.Fatal(err)
	}
	return m
}
func TestCompilerDeterministicAndDoesNotMutateModel(t *testing.T) {
	m := sampleModel(t)
	before, _ := json.Marshal(m)
	first, err := Compile(m)
	if err != nil {
		t.Fatal(err)
	}
	second, err := Compile(m)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(first, second) {
		t.Fatal("non-deterministic emission")
	}
	after, _ := json.Marshal(m)
	if !bytes.Equal(before, after) {
		t.Fatal("compiler mutated its input")
	}
	m.Fields[0], m.Fields[3] = m.Fields[3], m.Fields[0]
	permuted, err := Compile(m)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(first, permuted) {
		t.Fatal("field declaration order changed output")
	}
}
func TestCompilerRejectsInvalidTypedModel(t *testing.T) {
	m := sampleModel(t)
	m.Identity = "Missing"
	if _, err := Compile(m); err == nil {
		t.Fatal("accepted invalid typed model")
	}
}
