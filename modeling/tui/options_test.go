package tui

import (
	"bytes"
	"encoding/json"
	"os"
	"sort"
	"strings"
	"testing"
)

// These digests were produced by generator 1.0.0, before emit and detailText
// existed. A model without the optional members must keep them.
const (
	baselineSourceSHA256    = "ce56329a582aa3fd3f712b9344972cee1da0ba28dabb539829de166d7736d6da"
	baselineCanonicalSHA256 = "6a8370de19f321f4a8bb0737745d04d86aa14e30203259d95fe5d79b27a8ac53"
)

func withMembers(t *testing.T, members string) []byte {
	t.Helper()
	raw := bytes.TrimSpace(exampleBytes(t))
	if !bytes.HasSuffix(raw, []byte("}")) {
		t.Fatal("unexpected example")
	}
	return append(append(raw[:len(raw)-1:len(raw)-1], ',', '\n'), []byte(members+"}")...)
}

func canonicalOf(t *testing.T, raw []byte) (string, map[string][]byte) {
	t.Helper()
	r, err := Build(raw)
	if err != nil {
		t.Fatal(err)
	}
	var id struct {
		Canonical string `json:"canonicalSHA256"`
	}
	if err := json.Unmarshal([]byte(r.Identity), &id); err != nil {
		t.Fatal(err)
	}
	return id.Canonical, r.Files
}

func TestBaselineDigestIsUnchangedByOptionalMembers(t *testing.T) {
	r, err := Build(exampleBytes(t))
	if err != nil {
		t.Fatal(err)
	}
	want := `{"contract":"foundry-generator/v1","profile":"record-browser/v1","sourceSHA256":"` + baselineSourceSHA256 + `","canonicalSHA256":"` + baselineCanonicalSHA256 + `"}` + "\n"
	if r.Identity != want {
		t.Fatalf("identity changed:\n got %swant %s", r.Identity, want)
	}
	if strings.Contains(string(r.Files["model.json"]), "emit") || strings.Contains(string(r.Files["model.json"]), "detailText") {
		t.Fatal("optional members present in baseline canonical model")
	}
}

func TestExplicitDefaultsCanonicalizeToAbsence(t *testing.T) {
	for _, members := range []string{
		`"detailText": "clip"`,
		`"emit": {"mode": "module"}`,
		`"detailText": "clip", "emit": {"mode": "module"}`,
		`"emit": {"mode": "module"}, "detailText": "clip"`,
	} {
		digest, files := canonicalOf(t, withMembers(t, members))
		if digest != baselineCanonicalSHA256 {
			t.Fatalf("%s: canonical digest %s differs from baseline", members, digest)
		}
		base, err := Build(exampleBytes(t))
		if err != nil {
			t.Fatal(err)
		}
		for name, raw := range base.Files {
			if !bytes.Equal(raw, files[name]) {
				t.Fatalf("%s: %s differs from baseline output", members, name)
			}
		}
	}
}

func TestOptionalMembersAreCanonical(t *testing.T) {
	d1, f1 := canonicalOf(t, withMembers(t, `"detailText": "wrap"`))
	d2, f2 := canonicalOf(t, withMembers(t, `"emit": {"package": "mcpbrowser", "mode": "package"}`))
	d3, f3 := canonicalOf(t, withMembers(t, `"detailText": "wrap", "emit": {"mode": "package", "package": "mcpbrowser"}`))
	d4, _ := canonicalOf(t, withMembers(t, `"emit": {"mode": "package", "package": "mcpbrowser"}, "detailText": "wrap"`))
	if d1 == baselineCanonicalSHA256 || d2 == baselineCanonicalSHA256 || d1 == d2 || d3 == d1 || d3 == d2 {
		t.Fatal("optional members do not change the digest")
	}
	if d3 != d4 {
		t.Fatal("source member order changed the digest")
	}
	if !strings.Contains(string(f1["model.json"]), `"detailText": "wrap"`) || !strings.Contains(string(f2["model.json"]), `"package": "mcpbrowser"`) || !strings.Contains(string(f3["model.json"]), `"emit"`) {
		t.Fatal("members missing from canonical model")
	}
	// The wrap setting reaches the generated code.
	if !strings.Contains(string(f1["layout.go"]), "wrapDetails  = true") || !strings.Contains(string(f2["layout.go"]), "wrapDetails  = false") {
		t.Fatal("detailText not reflected in emitted code")
	}
}

func TestOptionalMembersValidation(t *testing.T) {
	bad := []string{
		`"detailText": ""`, `"detailText": "Wrap"`, `"detailText": "wrapped"`, `"detailText": null`, `"detailText": true`,
		`"emit": {}`, `"emit": null`, `"emit": "package"`, `"emit": {"mode": "Package", "package": "x"}`,
		`"emit": {"mode": "package"}`, `"emit": {"mode": "package", "package": ""}`,
		`"emit": {"mode": "module", "package": "mcpbrowser"}`, `"emit": {"mode": "module", "package": ""}`,
		`"emit": {"mode": "package", "package": "Mcp"}`, `"emit": {"mode": "package", "package": "mcp_browser"}`,
		`"emit": {"mode": "package", "package": "1abc"}`, `"emit": {"mode": "package", "package": "mcp-browser"}`,
		`"emit": {"mode": "package", "package": "` + strings.Repeat("a", 33) + `"}`,
		`"emit": {"mode": "package", "package": "é"}`, `"emit": {"mode": "package", "package": "a.b"}`,
		`"emit": {"mode": "package", "package": "a/b"}`, `"emit": {"mode": "package", "package": "a b"}`,
		`"emit": {"mode": "package", "package": "go"}`, `"emit": {"mode": "package", "package": "type"}`,
		`"emit": {"mode": "package", "package": "string"}`, `"emit": {"mode": "package", "package": "nil"}`,
		`"emit": {"mode": "package", "package": "len"}`, `"emit": {"mode": "package", "package": "main"}`,
		`"emit": {"mode": "package", "package": "mcpbrowser", "path": "x"}`,
		`"emit": {"mode": "package", "package": "mcpbrowser", "package": "other"}`,
		`"emit": {"mode": "module", "mode": "package"}`,
		`"Emit": {"mode": "module"}`, `"DetailText": "clip"`,
		`"detailText": "clip", "detailText": "clip"`,
	}
	for _, members := range bad {
		if _, err := Decode(withMembers(t, members)); err == nil {
			t.Errorf("accepted %s", members)
		}
	}
	good := []string{
		`"emit": {"mode": "package", "package": "a"}`,
		`"emit": {"mode": "package", "package": "` + strings.Repeat("a", 32) + `"}`,
		`"emit": {"mode": "package", "package": "mcp2browser"}`,
		`"detailText": "wrap"`,
	}
	for _, members := range good {
		if _, err := Decode(withMembers(t, members)); err != nil {
			t.Errorf("rejected %s: %v", members, err)
		}
	}
	// Typed models are validated the same way.
	m := exampleModel(t)
	m.Emit = &Emit{Mode: ModePackage, Package: "func"}
	if _, err := Compile(m); err == nil {
		t.Error("Compile accepted a keyword package name")
	}
	m.Emit = &Emit{}
	if err := m.Validate(); err == nil {
		t.Error("accepted empty emit mode")
	}
	m = exampleModel(t)
	m.DetailText = "other"
	if err := m.Validate(); err == nil {
		t.Error("accepted unknown detailText")
	}
}

func fileNames(files map[string][]byte) []string {
	var names []string
	for name := range files {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func TestEmitModeFileSets(t *testing.T) {
	module, err := Compile(exampleModel(t))
	if err != nil {
		t.Fatal(err)
	}
	wantModule := "browser.go,cmd/demo/main.go,foundry-manifest.json,go.mod,go.sum,keys.go,layout.go,model.json,records.go,state.go,view.go"
	if got := strings.Join(fileNames(module), ","); got != wantModule {
		t.Fatalf("module files %s", got)
	}
	if !strings.HasPrefix(string(module["go.mod"]), "module example.com/generated/recordbrowser\n") {
		t.Fatal("module path changed")
	}
	raw, err := os.ReadFile("examples/browser-package.json")
	if err != nil {
		t.Fatal(err)
	}
	r, err := Build(raw)
	if err != nil {
		t.Fatal(err)
	}
	wantPackage := "browser.go,foundry-manifest.json,keys.go,layout.go,model.json,records.go,state.go,view.go"
	if got := strings.Join(fileNames(r.Files), ","); got != wantPackage {
		t.Fatalf("package files %s", got)
	}
	for name, content := range r.Files {
		if strings.HasSuffix(name, ".go") {
			if !strings.Contains(string(content), "\npackage mcpbrowser\n") || strings.Contains(string(content), "recordbrowser") {
				t.Fatalf("%s: wrong package clause", name)
			}
			if !strings.HasPrefix(string(content), "// Code generated by tuic 1.3.0 from model ") {
				t.Fatalf("%s: header", name)
			}
		}
	}
	// The manifest lists exactly the other files.
	var mf struct {
		Files []struct{ Path string } `json:"files"`
	}
	if err := json.Unmarshal(r.Files[ManifestName], &mf); err != nil || len(mf.Files) != len(r.Files)-1 {
		t.Fatalf("manifest files: %v %d", err, len(mf.Files))
	}
}
