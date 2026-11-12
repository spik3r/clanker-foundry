package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestBindThenGenerateCheck(t *testing.T) {
	template := filepath.Join("..", "..", "examples", "ops-dashboard.json")
	data := filepath.Join("..", "..", "examples", "ops-report-data.json")
	templateBefore, err := os.ReadFile(template)
	if err != nil {
		t.Fatal(err)
	}
	dataBefore, err := os.ReadFile(data)
	if err != nil {
		t.Fatal(err)
	}
	code, out, errs := do("bind", template, data)
	if code != 0 || out == "" || errs != "" {
		t.Fatalf("bind: %d %s", code, errs)
	}
	dir := t.TempDir()
	model := filepath.Join(dir, "model.json")
	rendered := filepath.Join(dir, "rendered")
	if err := os.WriteFile(model, []byte(out), 0600); err != nil {
		t.Fatal(err)
	}
	if code, _, errs := do("generate", model, rendered); code != 0 {
		t.Fatalf("generate: %d %s", code, errs)
	}
	if code, _, errs := do("check", model, rendered); code != 0 {
		t.Fatalf("check: %d %s", code, errs)
	}
	after, _ := os.ReadFile(template)
	if !bytes.Equal(after, templateBefore) {
		t.Fatal("template mutated")
	}
	after, _ = os.ReadFile(data)
	if !bytes.Equal(after, dataBefore) {
		t.Fatal("data mutated")
	}
	if err := os.WriteFile(filepath.Join(rendered, "dashboard.html"), []byte("changed"), 0600); err != nil {
		t.Fatal(err)
	}
	if code, stdout, _ := do("check", model, rendered); code != 3 || stdout != "" {
		t.Fatal("drift not detected")
	}
}

func TestBindRefusesBeforeOutput(t *testing.T) {
	template := filepath.Join("..", "..", "examples", "ops-dashboard.json")
	data := filepath.Join("..", "..", "examples", "ops-report-data.json")
	dir := t.TempDir()
	bad := filepath.Join(dir, "bad.json")
	if err := os.WriteFile(bad, []byte(`{"dataVersion":"wrong"}`), 0600); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{{"bind"}, {"bind", template}, {"bind", template, data, "extra"}, {"bind", template, bad}, {"bind", dir, data}, {"bind", template, dir}, {"bind", filepath.Join(dir, "missing"), data}, {"bind", template, filepath.Join(dir, "missing")}} {
		if code, out, errs := do(args...); code != 2 || out != "" || errs == "" {
			t.Fatalf("%v: %d %q %q", args, code, out, errs)
		}
	}
	target, err := filepath.Abs(data)
	if err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(dir, "link.json")
	if err := os.Symlink(target, link); err != nil {
		t.Skip(err)
	}
	if code, out, _ := do("bind", template, link); code != 2 || out != "" {
		t.Fatal("symlink accepted")
	}
}
