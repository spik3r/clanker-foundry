package foundrygen

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Both destinations exist, so cleaning can silently select a lexical decoy.
func rootPathFixture(t *testing.T) (base, physical, decoy, raw string) {
	t.Helper()
	base = t.TempDir()
	physical = filepath.Join(base, "physical space 用户", "chosen")
	decoy = filepath.Join(base, "chosen")
	child := filepath.Join(filepath.Dir(physical), "child")
	for _, dir := range []string{physical, decoy, child} {
		if err := os.MkdirAll(dir, 0700); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Symlink(child, filepath.Join(base, "alias")); err != nil {
		t.Skipf("symlink fixture unavailable: %v", err)
	}
	sep := string(os.PathSeparator)
	raw = base + sep + "alias" + sep + ".." + sep + "chosen"
	return
}
func requireRootIdentity(t *testing.T, spelling, physical string) {
	t.Helper()
	got, err := absDir(spelling)
	if err != nil {
		t.Fatal(err)
	}
	if !filepath.IsAbs(got) {
		t.Fatalf("root is not absolute: %q", got)
	}
	wantInfo, err := os.Stat(physical)
	if err != nil {
		t.Fatal(err)
	}
	gotInfo, err := os.Stat(got)
	if err != nil {
		t.Fatal(err)
	}
	if !os.SameFile(wantInfo, gotInfo) {
		t.Fatalf("root selected another directory: %q -> %q, want physical %q", spelling, got, physical)
	}
}
func TestRootPathPhysicalIdentity(t *testing.T) {
	base, physical, _, raw := rootPathFixture(t)
	sep := string(os.PathSeparator)
	for _, spelling := range []string{raw, raw + sep, raw + sep + ".", physical} {
		requireRootIdentity(t, spelling, physical)
	}
	if err := os.Symlink(physical, filepath.Join(base, "terminal")); err != nil {
		t.Fatal(err)
	}
	requireRootIdentity(t, filepath.Join(base, "terminal"), physical)
	if err := os.Symlink(filepath.Dir(physical), filepath.Join(base, "ancestor")); err != nil {
		t.Fatal(err)
	}
	requireRootIdentity(t, filepath.Join(base, "ancestor", "chosen"), physical)
	if err := os.Symlink(filepath.Join("physical space 用户", "child"), filepath.Join(base, "relative-alias")); err != nil {
		t.Fatal(err)
	}
	t.Chdir(base)
	requireRootIdentity(t, "alias"+sep+".."+sep+"chosen", physical)
	requireRootIdentity(t, "relative-alias"+sep+".."+sep+"chosen", physical)
	t.Chdir(physical)
	for _, spelling := range []string{"", ".", "." + sep} {
		requireRootIdentity(t, spelling, physical)
	}
}
func TestRootPathRefusesUnresolvedComponents(t *testing.T) {
	base, _, _, _ := rootPathFixture(t)
	writeFile(t, filepath.Join(base, "file"), "owned regular file")
	if err := os.Symlink("missing-target", filepath.Join(base, "dangling")); err != nil {
		t.Fatal(err)
	}
	sep := string(os.PathSeparator)
	for _, spelling := range []string{filepath.Join(base, "missing"), filepath.Join(base, "file"), base + sep + "missing" + sep + ".." + sep + "chosen", base + sep + "file" + sep + ".." + sep + "chosen", filepath.Join(base, "dangling")} {
		if got, err := absDir(spelling); err == nil {
			t.Fatalf("invalid root became %q: %q", got, spelling)
		}
	}
}
func rootMarkerRun(t *testing.T, verb, project, foundry, marker string) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	code := Run([]string{verb, "--project", project, "--foundry", foundry, "--unpinned"}, &stdout, &stderr)
	if code != exitRefused || stdout.Len() != 0 || !strings.Contains(stderr.String(), marker) || strings.Contains(stderr.String(), "DECOY_ROOT") {
		t.Fatalf("%s selected wrong intake: code=%d stdout=%q stderr=%q", verb, code, stdout.String(), stderr.String())
	}
}
func TestRootPathCLISelectsPhysicalProject(t *testing.T) {
	_, physical, decoy, raw := rootPathFixture(t)
	document := func(marker string) string {
		return `{"projectVersion":"foundry-project/v1","foundry":{},"outputs":[],"` + marker + `":true}`
	}
	writeFile(t, filepath.Join(physical, ProjectFile), document("ACTUAL_ROOT"))
	writeFile(t, filepath.Join(decoy, ProjectFile), document("DECOY_ROOT"))
	beforePhysical, beforeDecoy := snapshot(t, physical), snapshot(t, decoy)
	foundry := repoRoot(t)
	t.Setenv("PATH", t.TempDir()) // Marker refusals must precede generator builds.
	for _, verb := range []string{"plan", "check", "generate"} {
		rootMarkerRun(t, verb, raw, foundry, "ACTUAL_ROOT")
	}
	if !sameSnapshot(beforePhysical, snapshot(t, physical)) || !sameSnapshot(beforeDecoy, snapshot(t, decoy)) {
		t.Fatal("intake refusal changed fixture bytes or mtimes")
	}
	t.Chdir(physical)
	for _, args := range [][]string{{"plan", "--foundry", foundry, "--unpinned"}, {"plan", "--project", "", "--foundry", foundry, "--unpinned"}} {
		var out, diagnostic bytes.Buffer
		if code := Run(args, &out, &diagnostic); code != exitRefused || out.Len() != 0 || !strings.Contains(diagnostic.String(), "ACTUAL_ROOT") {
			t.Fatalf("current-directory intake changed: %d %q %q", code, out.String(), diagnostic.String())
		}
	}
}
func TestRootPathCLISelectsPhysicalFoundry(t *testing.T) {
	_, physical, decoy, raw := rootPathFixture(t)
	// Invalid IDs distinguish registries before any project or generator is used.
	document := func(marker string) string {
		return `{"registryVersion":"foundry-generator-registry/v1","generators":[{"id":"` + marker + `!","module":"unused","package":"./unused"}]}`
	}
	for dir, marker := range map[string]string{physical: "ACTUAL_ROOT", decoy: "DECOY_ROOT"} {
		writeFile(t, filepath.Join(dir, "modeling", "gencontract", "generators.json"), document(marker))
	}
	beforePhysical, beforeDecoy := snapshot(t, physical), snapshot(t, decoy)
	project := t.TempDir()
	t.Setenv("PATH", t.TempDir())
	for _, verb := range []string{"plan", "check", "generate"} {
		rootMarkerRun(t, verb, project, raw, "ACTUAL_ROOT")
	}
	if !sameSnapshot(beforePhysical, snapshot(t, physical)) || !sameSnapshot(beforeDecoy, snapshot(t, decoy)) {
		t.Fatal("registry refusal changed fixture bytes or mtimes")
	}
}
func TestRootPathRealGeneratorIntegration(t *testing.T) {
	base, physical, decoy, raw := rootPathFixture(t)
	foundry := repoRoot(t)
	if err := os.Symlink(filepath.Join(foundry, "modeling"), filepath.Join(base, "checkout-child")); err != nil {
		t.Fatal(err)
	}
	foundryAlias := base + string(os.PathSeparator) + "checkout-child" + string(os.PathSeparator) + ".."
	writeFile(t, filepath.Join(physical, "model.json"), aliasModel)
	writeFile(t, filepath.Join(physical, ProjectFile), projectJSON(out("jsonvalidator", "model.json", "internal/generated")))
	writeFile(t, filepath.Join(decoy, ProjectFile), "DECOY_ROOT")
	beforeDecoy := snapshot(t, decoy)
	if code, so, se := runCLI(t, "generate", physical); code != exitOK {
		t.Fatalf("direct generation: %d %q %q", code, so, se)
	}
	before := snapshot(t, physical)
	for _, verb := range []string{"plan", "check", "generate"} {
		directCode, directOut, directErr := runCLI(t, verb, physical)
		aliasCode, aliasOut, aliasErr := runCLI(t, verb, raw)
		if directCode != exitOK || aliasCode != exitOK || directOut != aliasOut || directErr != aliasErr {
			t.Fatalf("%s direct/alias differ: %d/%d %q/%q %q/%q", verb, directCode, aliasCode, directOut, aliasOut, directErr, aliasErr)
		}
		bothCode, bothOut, bothErr := runCLI(t, verb, raw, "--foundry", foundryAlias)
		if bothCode != exitOK || directOut != bothOut || directErr != bothErr {
			t.Fatalf("%s aliased Foundry differs: %d %q %q", verb, bothCode, bothOut, bothErr)
		}
		if !sameSnapshot(before, snapshot(t, physical)) || !sameSnapshot(beforeDecoy, snapshot(t, decoy)) {
			t.Fatal("unchanged alias command rewrote project or decoy")
		}
	}
	// Allowed root aliases do not relax internal model-path symlink restrictions.
	if err := os.Symlink("model.json", filepath.Join(physical, "linked-model.json")); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(physical, ProjectFile), projectJSON(out("jsonvalidator", "linked-model.json", "internal/generated")))
	for _, verb := range []string{"plan", "check", "generate"} {
		code, stdout, stderr := runCLI(t, verb, raw)
		if code != exitRefused || stdout != "" || !strings.Contains(stderr, "symlink") {
			t.Fatalf("internal model symlink policy changed: %d %q %q", code, stdout, stderr)
		}
	}
}
