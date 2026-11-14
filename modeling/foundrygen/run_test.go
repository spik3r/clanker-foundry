package foundrygen

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"
)

// repoRoot is the Foundry checkout these tests use: this repository.
func repoRoot(t *testing.T) string {
	t.Helper()
	abs, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(abs, "modeling", "gencontract", "generators.json")); err != nil {
		t.Fatalf("not the repository root: %v", err)
	}
	return abs
}

func writeFile(t *testing.T, path, data string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func copyFile(t *testing.T, src, dst string) {
	t.Helper()
	writeFile(t, dst, readFile(t, src))
}

// scratch makes a project with the real tuic and jsonvalidator sample models.
func scratch(t *testing.T) string {
	t.Helper()
	repo := repoRoot(t)
	dir := t.TempDir()
	copyFile(t, filepath.Join(repo, "modeling/tui/examples/browser.json"), filepath.Join(dir, "design/browser.json"))
	copyFile(t, filepath.Join(repo, "modeling/jsonvalidator/testdata/sample/model.json"), filepath.Join(dir, "design/sample.json"))
	writeFile(t, filepath.Join(dir, "foundry.json"), projectJSON(
		out("tuic", "design/browser.json", "internal/browser")+","+
			out("jsonvalidator", "design/sample.json", "internal/sample")))
	return dir
}

func runCLI(t *testing.T, verb, project string, extra ...string) (int, string, string) {
	t.Helper()
	var so, se bytes.Buffer
	args := append([]string{verb, "--project", project, "--foundry", repoRoot(t), "--unpinned"}, extra...)
	code := Run(args, &so, &se)
	return code, so.String(), se.String()
}

type fileState struct {
	data    string
	modTime time.Time
}

func snapshot(t *testing.T, root string) map[string]fileState {
	t.Helper()
	m := map[string]fileState{}
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		b, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(root, p)
		m[rel] = fileState{string(b), info.ModTime()}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return m
}

func sameSnapshot(a, b map[string]fileState) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		w, ok := b[k]
		if !ok || v.data != w.data || !v.modTime.Equal(w.modTime) {
			return false
		}
	}
	return true
}

// firstGenerated returns a generated file (not the manifest) under dir.
func firstGenerated(t *testing.T, dir string) string {
	t.Helper()
	var found string
	_ = filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err == nil && !d.IsDir() && found == "" && d.Name() != "foundry-manifest.json" {
			found = p
		}
		return nil
	})
	if found == "" {
		t.Fatalf("no generated file under %s", dir)
	}
	return found
}

func flipByte(t *testing.T, path string) {
	t.Helper()
	b := []byte(readFile(t, path))
	b[len(b)-1] ^= 0x01
	if err := os.WriteFile(path, b, 0o644); err != nil {
		t.Fatal(err)
	}
}

var statusLine = regexp.MustCompile(`^(ok|generated|regenerated|unchanged) (\S+) (\S+) [0-9a-f]{64}$`)

func checkStatusLines(t *testing.T, stdout string, status string, outs ...string) {
	t.Helper()
	lines := strings.Split(strings.TrimRight(stdout, "\n"), "\n")
	if len(lines) != len(outs) {
		t.Fatalf("stdout has %d lines, want %d:\n%s", len(lines), len(outs), stdout)
	}
	for i, l := range lines {
		m := statusLine.FindStringSubmatch(l)
		if m == nil || m[1] != status || m[2] != outs[i] {
			t.Fatalf("line %d %q: want status %s for %s", i, l, status, outs[i])
		}
	}
}

func TestEndToEnd(t *testing.T) {
	proj := scratch(t)
	browserOut := filepath.Join(proj, "internal/browser")
	sampleOut := filepath.Join(proj, "internal/sample")

	code, so, se := runCLI(t, "check", proj)
	if code != 3 || so != "" || !strings.Contains(se, "internal/browser: drift: missing .") || !strings.Contains(se, "internal/sample: drift: missing .") {
		t.Fatalf("check before generate: code %d stdout %q stderr %q", code, so, se)
	}
	if _, err := os.Lstat(filepath.Join(proj, "internal")); err == nil {
		t.Fatal("check created internal/")
	}

	code, so, se = runCLI(t, "generate", proj)
	if code != 0 {
		t.Fatalf("generate: code %d stderr %q", code, se)
	}
	checkStatusLines(t, so, "generated", "internal/browser", "internal/sample")
	if !strings.Contains(se, "--unpinned") {
		t.Errorf("no unpinned warning: %q", se)
	}
	entries, _ := os.ReadDir(filepath.Join(proj, "internal"))
	if len(entries) != 2 {
		t.Fatalf("internal/ holds %d entries, want 2 (leftover temp dir?)", len(entries))
	}

	before := snapshot(t, proj)
	code, so, se = runCLI(t, "check", proj)
	if code != 0 {
		t.Fatalf("check: code %d stderr %q", code, se)
	}
	checkStatusLines(t, so, "ok", "internal/browser", "internal/sample")
	if !sameSnapshot(before, snapshot(t, proj)) {
		t.Fatal("check modified the project")
	}

	// Unchanged generate leaves bytes and mtimes alone.
	code, so, se = runCLI(t, "generate", proj)
	if code != 0 {
		t.Fatalf("second generate: code %d stderr %q", code, se)
	}
	checkStatusLines(t, so, "unchanged", "internal/browser", "internal/sample")
	if !sameSnapshot(before, snapshot(t, proj)) {
		t.Fatal("unchanged generate modified the project")
	}

	// One edited byte in each output: both are reported, with the output prefix.
	bf, sf := firstGenerated(t, browserOut), firstGenerated(t, sampleOut)
	flipByte(t, bf)
	flipByte(t, sf)
	relB, _ := filepath.Rel(browserOut, bf)
	relS, _ := filepath.Rel(sampleOut, sf)
	code, so, se = runCLI(t, "check", proj)
	if code != 3 || so != "" {
		t.Fatalf("tampered check: code %d stdout %q", code, so)
	}
	for _, want := range []string{
		"internal/browser: drift: modified " + filepath.ToSlash(relB),
		"internal/sample: drift: modified " + filepath.ToSlash(relS),
	} {
		if !strings.Contains(se, want) {
			t.Errorf("stderr lacks %q:\n%s", want, se)
		}
	}

	// generate repairs it.
	code, so, se = runCLI(t, "generate", proj)
	if code != 0 {
		t.Fatalf("repair generate: code %d stderr %q", code, se)
	}
	checkStatusLines(t, so, "regenerated", "internal/browser", "internal/sample")
	if code, _, se = runCLI(t, "check", proj); code != 0 {
		t.Fatalf("check after repair: %d %s", code, se)
	}

	// Model edited without regenerating.
	model := filepath.Join(proj, "design/browser.json")
	edited := strings.Replace(readFile(t, model), `"title": "Catalog"`, `"title": "Catalog changed"`, 1)
	if edited == readFile(t, model) {
		t.Fatal("model edit did not apply")
	}
	writeFile(t, model, edited)
	code, so, se = runCLI(t, "check", proj)
	if code != 3 || so != "" || !strings.Contains(se, "internal/browser: drift: modified") {
		t.Fatalf("model-edit check: code %d stdout %q stderr %q", code, so, se)
	}
	if strings.Contains(se, "internal/sample:") {
		t.Errorf("sample output wrongly reported: %s", se)
	}
	code, so, se = runCLI(t, "generate", proj)
	if code != 0 {
		t.Fatalf("generate after model edit: %d %s", code, se)
	}
	if !strings.HasPrefix(so, "regenerated internal/browser tuic ") || !strings.Contains(so, "unchanged internal/sample jsonvalidator ") {
		t.Fatalf("stdout %q", so)
	}
	if code, _, se = runCLI(t, "check", proj); code != 0 {
		t.Fatalf("final check: %d %s", code, se)
	}
}

func TestGenerateRefusesUnownedOut(t *testing.T) {
	proj := scratch(t)
	writeFile(t, filepath.Join(proj, "internal/browser/hand-written.go"), "package x\n")
	before := snapshot(t, proj)
	code, so, se := runCLI(t, "generate", proj)
	if code != 2 || so != "" || !strings.Contains(se, "foundry-manifest.json") {
		t.Fatalf("code %d stdout %q stderr %q", code, so, se)
	}
	if !sameSnapshot(before, snapshot(t, proj)) {
		t.Fatal("project changed")
	}
	if _, err := os.Lstat(filepath.Join(proj, "internal/sample")); err == nil {
		t.Fatal("second output was written despite the refusal")
	}

	// A regular file at OUT is also refused.
	proj = scratch(t)
	writeFile(t, filepath.Join(proj, "internal/browser"), "x")
	if code, _, _ = runCLI(t, "generate", proj); code != 2 {
		t.Fatalf("file at OUT: code %d", code)
	}
	// So is an empty directory.
	proj = scratch(t)
	if err := os.MkdirAll(filepath.Join(proj, "internal/browser"), 0o755); err != nil {
		t.Fatal(err)
	}
	if code, _, _ = runCLI(t, "generate", proj); code != 2 {
		t.Fatalf("empty dir at OUT: code %d", code)
	}
}

func TestGenerateInvalidModelWritesNothing(t *testing.T) {
	proj := scratch(t)
	writeFile(t, filepath.Join(proj, "design/sample.json"), "{}")
	before := snapshot(t, proj)
	code, so, se := runCLI(t, "generate", proj)
	if code != 2 || so != "" || !strings.Contains(se, "internal/sample") {
		t.Fatalf("code %d stdout %q stderr %q", code, so, se)
	}
	if !sameSnapshot(before, snapshot(t, proj)) {
		t.Fatal("project changed")
	}
	if _, err := os.Lstat(filepath.Join(proj, "internal")); err == nil {
		t.Fatal("internal/ was created")
	}
	if code, so, _ = runCLI(t, "check", proj); code != 2 || so != "" {
		t.Fatalf("check with bad model: code %d", code)
	}
}

func TestSymlinkedPathsRefused(t *testing.T) {
	proj := scratch(t)
	elsewhere := t.TempDir()
	if err := os.Symlink(elsewhere, filepath.Join(proj, "internal")); err != nil {
		t.Skip("symlinks unavailable:", err)
	}
	for _, verb := range []string{"check", "generate"} {
		if code, so, _ := runCLI(t, verb, proj); code != 2 || so != "" {
			t.Errorf("%s through symlinked parent: code %d", verb, code)
		}
	}
	if entries, _ := os.ReadDir(elsewhere); len(entries) != 0 {
		t.Fatal("wrote through the symlink")
	}
}

func TestUsage(t *testing.T) {
	repo := repoRoot(t)
	proj := scratch(t)
	for _, args := range [][]string{
		nil,
		{"frobnicate"},
		{"check"},
		{"check", "--foundry", repo, "--project", proj, "extra"},
		{"check", "--foundry", repo, "--project", proj, "--nope"},
		{"check", "--foundry", filepath.Join(repo, "missing"), "--project", proj},
		{"generate", "--foundry", repo, "--project", filepath.Join(proj, "missing")},
	} {
		var so, se bytes.Buffer
		if code := Run(args, &so, &se); code != 2 || so.Len() != 0 || se.Len() == 0 {
			t.Errorf("%v: code %d stdout %q stderr %q", args, code, so.String(), se.String())
		}
	}
}

func TestPinRequiredWithoutUnpinned(t *testing.T) {
	proj := scratch(t)
	var so, se bytes.Buffer
	code := Run([]string{"check", "--project", proj, "--foundry", repoRoot(t)}, &so, &se)
	// The scratch revision is not this repository's HEAD.
	if code != 2 || so.Len() != 0 || !strings.Contains(se.String(), "pins "+rev) {
		t.Fatalf("code %d stderr %q", code, se.String())
	}
}

func gitFixture(t *testing.T) (dir, head string) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git unavailable")
	}
	dir = t.TempDir()
	git := func(args ...string) string {
		t.Helper()
		cmd := exec.Command("git", append([]string{"-C", dir, "-c", "user.name=t", "-c", "user.email=t@example.invalid", "-c", "commit.gpgsign=false"}, args...)...)
		b, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, b)
		}
		return strings.TrimSpace(string(b))
	}
	git("init", "-q")
	writeFile(t, filepath.Join(dir, "modeling/gencontract/generators.json"),
		`{"registryVersion":"foundry-generator-registry/v1","generators":[{"id":"fake","profile":"p/v1","module":"modeling/fake","package":"./cmd/fake","model":"m.json"}]}`)
	git("add", "-A")
	git("commit", "-q", "-m", "fixture")
	return dir, git("rev-parse", "HEAD")
}

func TestPinCheck(t *testing.T) {
	dir, head := gitFixture(t)
	newSession := func(revision string) *session {
		return &session{foundry: dir, proj: &Project{Revision: revision}, stderr: &bytes.Buffer{}}
	}

	if err := newSession(head).pin(); err != nil {
		t.Fatalf("matching revision refused: %v", err)
	}

	err := newSession(strings.Repeat("0", 40)).pin()
	if err == nil || codeOf(err) != exitRefused || !strings.Contains(err.Error(), head) {
		t.Fatalf("mismatch: %v", err)
	}

	// Changes outside modeling/ do not matter; changes inside do.
	writeFile(t, filepath.Join(dir, "notes.txt"), "x")
	if err := newSession(head).pin(); err != nil {
		t.Fatalf("untracked file outside modeling/ refused: %v", err)
	}
	writeFile(t, filepath.Join(dir, "modeling/new.go"), "package x")
	err = newSession(head).pin()
	if err == nil || codeOf(err) != exitRefused || !strings.Contains(err.Error(), "uncommitted") {
		t.Fatalf("untracked file in modeling/: %v", err)
	}
	if err := os.Remove(filepath.Join(dir, "modeling/new.go")); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(dir, "modeling/gencontract/generators.json"), "{}")
	if err := newSession(head).pin(); err == nil || codeOf(err) != exitRefused {
		t.Fatalf("modified tracked file in modeling/: %v", err)
	}
}

func TestPinRefusesUntrackedFilesWhenGitHidesThem(t *testing.T) {
	for _, path := range []string{"modeling/extra.go", "modeling/new-package/extra.go"} {
		t.Run(path, func(t *testing.T) {
			dir, head := gitFixture(t)
			cmd := exec.Command("git", "-C", dir, "config", "status.showUntrackedFiles", "no")
			if b, err := cmd.CombinedOutput(); err != nil {
				t.Fatalf("configure Git: %v: %s", err, b)
			}
			s := &session{foundry: dir, proj: &Project{Revision: head}, stderr: &bytes.Buffer{}}
			writeFile(t, filepath.Join(dir, "notes.txt"), "outside modeling")
			if err := s.pin(); err != nil {
				t.Fatalf("clean modeling tree refused: %v", err)
			}
			writeFile(t, filepath.Join(dir, path), "package extra")
			err := s.pin()
			if err == nil || codeOf(err) != exitRefused || !strings.Contains(err.Error(), "uncommitted") {
				t.Fatalf("hidden untracked source accepted: %v", err)
			}
		})
	}
}

func TestRunPinMismatchOnFixture(t *testing.T) {
	dir, head := gitFixture(t)
	proj := t.TempDir()
	writeFile(t, filepath.Join(proj, "foundry.json"), fmt.Sprintf(
		`{"projectVersion":"foundry-project/v1","foundry":{"revision":"%s"},"outputs":[%s]}`,
		strings.Repeat("a", 40), out("fake", "m.json", "o")))
	var so, se bytes.Buffer
	code := Run([]string{"check", "--project", proj, "--foundry", dir}, &so, &se)
	if code != 2 || so.Len() != 0 || !strings.Contains(se.String(), head) {
		t.Fatalf("code %d stderr %q", code, se.String())
	}
}

func TestParseIdentity(t *testing.T) {
	h := strings.Repeat("a", 64)
	good := fmt.Sprintf(`{"contract":"foundry-generator/v1","profile":"p/v1","sourceSHA256":"%s","canonicalSHA256":"%s"}`+"\n", h, h)
	if _, err := parseIdentity(good); err != nil {
		t.Fatal(err)
	}
	for name, s := range map[string]string{
		"no newline": strings.TrimSuffix(good, "\n"),
		"two lines":  good + good,
		"bad hash":   strings.Replace(good, h, "zz", 1),
		"contract":   strings.Replace(good, "foundry-generator/v1", "x", 1),
		"extra key":  strings.Replace(good, `"profile"`, `"extra":"1","profile"`, 1),
	} {
		if _, err := parseIdentity(s); err == nil {
			t.Errorf("%s: accepted", name)
		}
	}
}
