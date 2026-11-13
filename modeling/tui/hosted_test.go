package tui

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Digests of the two examples that existed before generator 1.3.0. A model
// without the new members must keep them.
const (
	packageExampleCanonicalSHA256 = "43e31a41805dd76946df369f36d53c50df3521a937b330a54d9a3cac6a122414"
	hostedExampleCanonicalSHA256  = "82e00bf6ad341858b7ef314250157d2d72059965c4aaddcd3107135acbde0275"
)

func TestExistingExampleDigestsAreUnchanged(t *testing.T) {
	for file, want := range map[string]string{
		"examples/browser.json":         baselineCanonicalSHA256,
		"examples/browser-package.json": packageExampleCanonicalSHA256,
		"examples/browser-hosted.json":  hostedExampleCanonicalSHA256,
	} {
		raw, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		got, files := canonicalOf(t, raw)
		if got != want {
			t.Fatalf("%s: canonical digest %s, want %s", file, got, want)
		}
		if file != "examples/browser-hosted.json" && (strings.Contains(string(files["model.json"]), "embed") || strings.Contains(string(files["model.json"]), "detailSpacing")) {
			t.Fatalf("%s: new members present in the canonical model", file)
		}
	}
}

func TestExplicitEmbedAndSpacingDefaultsCanonicalizeToAbsence(t *testing.T) {
	for _, members := range []string{
		`"embed": "standalone"`,
		`"detailSpacing": "compact"`,
		`"embed": "standalone", "detailSpacing": "compact"`,
		`"detailSpacing": "compact", "detailText": "clip", "embed": "standalone", "emit": {"mode": "module"}`,
	} {
		digest, files := canonicalOf(t, withMembers(t, members))
		if digest != baselineCanonicalSHA256 {
			t.Fatalf("%s: digest %s differs from baseline", members, digest)
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
	// The same holds in package mode with wrap.
	pkg, err := os.ReadFile("examples/browser-package.json")
	if err != nil {
		t.Fatal(err)
	}
	raw := bytes.TrimSpace(pkg)
	raw = append(raw[:len(raw)-1:len(raw)-1], []byte(`, "embed": "standalone", "detailSpacing": "compact"}`)...)
	if digest, _ := canonicalOf(t, raw); digest != packageExampleCanonicalSHA256 {
		t.Fatalf("package example with explicit defaults: %s", digest)
	}
}

func TestEmbedAndSpacingAreCanonical(t *testing.T) {
	d1, f1 := canonicalOf(t, withMembers(t, `"embed": "hosted"`))
	d2, f2 := canonicalOf(t, withMembers(t, `"detailSpacing": "spaced"`))
	d3, f3 := canonicalOf(t, withMembers(t, `"detailSpacing": "spaced", "embed": "hosted"`))
	d4, _ := canonicalOf(t, withMembers(t, `"embed": "hosted", "detailSpacing": "spaced"`))
	if d1 == baselineCanonicalSHA256 || d2 == baselineCanonicalSHA256 || d1 == d2 || d3 == d1 || d3 == d2 {
		t.Fatal("new members do not change the digest")
	}
	if d3 != d4 {
		t.Fatal("source member order changed the digest")
	}
	if !strings.Contains(string(f1["model.json"]), `"embed": "hosted"`) || !strings.Contains(string(f2["model.json"]), `"detailSpacing": "spaced"`) {
		t.Fatal("members missing from the canonical model")
	}
	// Canonical order: emit, detailText, embed, detailSpacing.
	full := string(f3["model.json"])
	if strings.Index(full, `"embed"`) > strings.Index(full, `"detailSpacing"`) {
		t.Fatal("embed must precede detailSpacing in the canonical encoding")
	}
	all, _ := canonicalOf(t, withMembers(t, `"detailSpacing": "spaced", "embed": "hosted", "detailText": "wrap", "emit": {"mode": "package", "package": "page"}`))
	raw, err := os.ReadFile("examples/browser-hosted.json")
	if err != nil {
		t.Fatal(err)
	}
	if all == hostedExampleCanonicalSHA256 {
		t.Fatal("different model, same digest")
	}
	if _, files := canonicalOf(t, raw); strings.Index(string(files["model.json"]), `"emit"`) > strings.Index(string(files["model.json"]), `"detailText"`) ||
		strings.Index(string(files["model.json"]), `"detailText"`) > strings.Index(string(files["model.json"]), `"embed"`) {
		t.Fatal("canonical order is emit, detailText, embed, detailSpacing")
	}
	// The members reach the generated code.
	if !strings.Contains(string(f2["layout.go"]), "spaced       = true") || !strings.Contains(string(f1["layout.go"]), "spaced       = false") {
		t.Fatalf("detailSpacing not reflected in emitted code:\n%s", f2["layout.go"])
	}
	if !strings.Contains(string(f1["browser.go"]), "This browser is hosted") || !strings.Contains(string(f1["browser.go"]), "Forward tea.WindowSizeMsg with the size of the body area") {
		t.Fatal("hosting contract missing from the package documentation")
	}
	if strings.Contains(string(f2["browser.go"]), "This browser is hosted") || strings.Contains(string(f2["browser.go"]), "return m,nil}\n[[[") {
		t.Fatal("standalone browser carries hosted documentation")
	}
	if !strings.Contains(string(f2["browser.go"]), "return m, tea.Quit") || strings.Contains(string(f1["browser.go"]), "return m, tea.Quit") {
		t.Fatal("tea.Quit must appear exactly in standalone output")
	}
	for _, name := range []string{"cmd/demo/main.go"} {
		if !strings.Contains(string(f1[name]), "CapturesKeys") || strings.Contains(string(f2[name]), "CapturesKeys") {
			t.Fatal("only the hosted demo is a host")
		}
	}
}

func TestEmbedAndSpacingValidation(t *testing.T) {
	bad := []string{
		`"embed": ""`, `"embed": "Hosted"`, `"embed": "host"`, `"embed": null`, `"embed": true`, `"embed": {"mode": "hosted"}`,
		`"detailSpacing": ""`, `"detailSpacing": "Spaced"`, `"detailSpacing": "wide"`, `"detailSpacing": null`, `"detailSpacing": 1`,
		`"embed": "hosted", "embed": "hosted"`, `"detailSpacing": "spaced", "detailSpacing": "spaced"`,
		`"Embed": "hosted"`, `"DetailSpacing": "spaced"`,
	}
	for _, members := range bad {
		if _, err := Decode(withMembers(t, members)); err == nil {
			t.Errorf("accepted %s", members)
		}
	}
	for _, members := range []string{`"embed": "hosted"`, `"embed": "standalone"`, `"detailSpacing": "spaced"`, `"detailSpacing": "compact"`,
		`"embed": "hosted", "detailSpacing": "spaced", "detailText": "wrap", "emit": {"mode": "package", "package": "page"}`,
		`"embed": "hosted", "emit": {"mode": "module"}`} {
		if _, err := Decode(withMembers(t, members)); err != nil {
			t.Errorf("rejected %s: %v", members, err)
		}
	}
	m := exampleModel(t)
	m.Embed = "other"
	if err := m.Validate(); err == nil {
		t.Error("accepted unknown embed")
	}
	m = exampleModel(t)
	m.DetailSpacing = "other"
	if _, err := Compile(m); err == nil {
		t.Error("Compile accepted unknown detailSpacing")
	}
}

func TestHostedFileSetsAndVersion(t *testing.T) {
	if GeneratorVersion != "1.3.1" {
		t.Fatalf("generator version %s", GeneratorVersion)
	}
	raw, err := os.ReadFile("examples/browser-hosted.json")
	if err != nil {
		t.Fatal(err)
	}
	r, err := Build(raw)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := strings.Join(fileNames(r.Files), ","), "browser.go,foundry-manifest.json,keys.go,layout.go,model.json,records.go,state.go,view.go"; got != want {
		t.Fatalf("hosted package files %s", got)
	}
	m := exampleModel(t)
	m.Embed = EmbedHosted
	files, err := Compile(m)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := files["cmd/demo/main.go"]; !ok || files["go.mod"] == nil {
		t.Fatal("hosted module mode keeps its demo")
	}
}

// A hosted browser, in package mode, inside a host that has its own top bar,
// footer and page keys; the standalone browser keeps its behaviour there too.
func TestHostedBrowserInsideApplicationShell(t *testing.T) {
	if testing.Short() {
		t.Skip("generated Go acceptance invokes compiler")
	}
	module, err := Compile(sampleModel(t))
	if err != nil {
		t.Fatal(err)
	}
	gomod := strings.Replace(string(module["go.mod"]), "module example.com/generated/recordbrowser", "module example.com/host", 1)
	gomod = strings.Replace(gomod, "\ngo 1.25.5\n", "\ngo 1.26.0\n", 1)

	pkg := func(name string, edit func(*Model)) Model {
		m := sampleModel(t)
		m.Emit = &Emit{Mode: ModePackage, Package: name}
		edit(&m)
		return m
	}
	models := map[string]Model{
		"hostedbrowser": pkg("hostedbrowser", func(m *Model) { m.Embed, m.DetailSpacing, m.DetailText = EmbedHosted, SpacingSpaced, TextWrap }),
		"clipbrowser": pkg("clipbrowser", func(m *Model) {
			// Four details: at the minimum body a clipped, spaced record scrolls.
			m.Embed, m.DetailSpacing, m.Details = EmbedHosted, SpacingSpaced, []string{"Name", "Status", "Description", "ID"}
		}),
		"plainbrowser": pkg("plainbrowser", func(m *Model) {}),
	}
	read := func(name string) []byte {
		raw, err := os.ReadFile(filepath.Join("acceptance", name))
		if err != nil {
			t.Fatal(err)
		}
		return raw
	}
	host := t.TempDir()
	writeFiles(t, host, map[string][]byte{
		"go.mod":        []byte(gomod),
		"go.sum":        module["go.sum"],
		"main.go":       read("hosted_shell_main.go.txt"),
		"shell_test.go": read("hosted_shell_test.go.txt"),
	})
	for name, m := range models {
		files, err := Compile(m)
		if err != nil {
			t.Fatal(err)
		}
		out := filepath.Join(host, "internal", name)
		writeFiles(t, out, files)
		if m.Hosted() {
			raw := bytes.Replace(read("hosted_test.go.txt"), []byte("package recordbrowser\n"), []byte("package "+name+"\n"), 1)
			writeFiles(t, out, map[string][]byte{"hosted_test.go": raw})
		}
	}
	goIn(t, host, nil, "vet", "./...")
	out := goIn(t, host, nil, "test", "-race", "-count=1", "-p=1", "-v", "./...")
	for _, want := range []string{
		"--- PASS: TestHostQuitsButBrowserDoesNot", "--- PASS: TestDigitsSwitchPagesUnlessBrowserCapturesKeys",
		"--- PASS: TestBodyIsRenderedAtTheReducedSizeAndFitsTheShell", "--- PASS: TestWithColorsInsideTheShell",
		"--- PASS: TestStandaloneBrowserIsUnchanged", "--- PASS: TestHostedBodyFillsExactlyTheGivenSize",
		"--- PASS: TestHostedNeverQuits", "--- PASS: TestSpacedDetailRowsAndScrolling", "--- PASS: TestSpacedWrapRowCapCountsBlankRows",
		"--- PASS: TestWithColorsHosted",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %q in\n%s", want, out)
		}
	}
	// The clip-mode hosted package must run the spaced test in clip form, not skip it.
	if strings.Contains(out, "--- SKIP: TestSpacedDetailRowsAndScrolling") {
		t.Fatalf("spaced scrolling test skipped:\n%s", out)
	}
}

// Module-mode output of a hosted model, with its reference host as the demo,
// builds and vets.
func TestHostedModuleModeDemoBuilds(t *testing.T) {
	if testing.Short() {
		t.Skip("generated Go acceptance invokes compiler")
	}
	m := sampleModel(t)
	m.Embed, m.DetailSpacing = EmbedHosted, SpacingSpaced
	files, err := Compile(m)
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	writeFiles(t, dir, files)
	goIn(t, dir, nil, "vet", "./...")
	goIn(t, dir, nil, "build", "-o", filepath.Join(t.TempDir(), "demo"), "./cmd/demo")
}
