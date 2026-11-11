package builderdemo_test

// Frozen from the prospective contract and baseline sources before production
// implementation. These checks do not compile or execute a builder candidate.
import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/builderdemo"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/builderpreflight"
)

type fixedArtifact struct {
	ID     string `json:"id"`
	Role   string `json:"role"`
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
}
type fixedManifest struct {
	FormatVersion string          `json:"formatVersion"`
	Artifacts     []fixedArtifact `json:"artifacts"`
}

func read(t *testing.T, p string) []byte {
	t.Helper()
	b, e := os.ReadFile(p)
	if e != nil {
		t.Fatal(e)
	}
	return b
}
func digest(b []byte) string { s := sha256.Sum256(b); return hex.EncodeToString(s[:]) }
func decode(t *testing.T, b []byte) map[string]any {
	t.Helper()
	var v map[string]any
	if e := json.Unmarshal(b, &v); e != nil {
		t.Fatal(e)
	}
	return v
}
func receipt(t *testing.T, r builderdemo.Receipt) map[string]any {
	t.Helper()
	b, e := json.Marshal(r)
	if e != nil {
		t.Fatal(e)
	}
	return decode(t, b)
}
func inventory(t *testing.T) ([]byte, fixedManifest) {
	t.Helper()
	b := read(t, "inputs.json")
	if digest(b) != "6a87f4db69a9c43376b405aca0d1252a662a3a4cdc7496329f2b41fac2896032" {
		t.Fatal("fixed manifest bytes changed")
	}
	var m fixedManifest
	if e := json.Unmarshal(b, &m); e != nil {
		t.Fatal(e)
	}
	if m.FormatVersion != "model-builder-inputs/v1" || len(m.Artifacts) != 15 {
		t.Fatal("fixed 15-input inventory", m)
	}
	return b, m
}
func fixture(t *testing.T) string {
	t.Helper()
	_, m := inventory(t)
	root := t.TempDir()
	for _, a := range m.Artifacts {
		input := filepath.Join("../../..", a.Path)
		if a.Path == "AGENTS.md" {
			input = "testdata/AGENTS-882b5002.txt"
		}
		b := read(t, input)
		if digest(b) != a.SHA256 {
			t.Fatalf("baseline drift: %s", a.Path)
		}
		p := filepath.Join(root, a.Path)
		if e := os.MkdirAll(filepath.Dir(p), 0700); e != nil {
			t.Fatal(e)
		}
		if e := os.WriteFile(p, b, 0600); e != nil {
			t.Fatal(e)
		}
	}
	return root
}
func empty(t *testing.T, p string) {
	t.Helper()
	v, e := os.ReadDir(p)
	if e != nil || len(v) != 0 {
		t.Fatalf("unexpected writes %s: %v %v", p, v, e)
	}
}
func envelope(t *testing.T, r map[string]any, outcome string) {
	t.Helper()
	for k, w := range map[string]string{"formatVersion": "model-builder-demo-preparation/v1", "outcome": outcome, "executionStatus": "not-run", "reviewStatus": "not-established", "authority": "none"} {
		if r[k] != w {
			t.Fatalf("%s=%v want %s", k, r[k], w)
		}
	}
	for _, k := range []string{"workspace", "handoffPath", "sourceManifestSHA256", "sourceReport", "inputsReport", "exerciseReport"} {
		if _, ok := r[k]; !ok {
			t.Fatalf("missing receipt field %s", k)
		}
	}
}
func compareReport(t *testing.T, got any, manifest []byte, root string) {
	t.Helper()
	want, e := builderpreflight.Check(manifest, root, digest(manifest))
	if e != nil {
		t.Fatal(e)
	}
	if !reflect.DeepEqual(got, decode(t, want)) {
		t.Fatalf("report differs from unchanged preflight\ngot=%v\nwant=%s", got, want)
	}
}

func TestPreparationPreservesOriginalsAndCreatesOnlyNamedFiles(t *testing.T) {
	root, parent := fixture(t), t.TempDir()
	raw, m := inventory(t)
	// A nonexistent tool path is valid: preparation may not probe or run it.
	tool := filepath.Join(t.TempDir(), "not-installed-go")
	r, e := builderdemo.Prepare(root, parent, tool)
	if e != nil {
		t.Fatal(e)
	}
	got := receipt(t, r)
	envelope(t, got, "prepared-draft")
	workspace, ok := got["workspace"].(string)
	if !ok || filepath.Dir(workspace) != parent || !strings.HasPrefix(filepath.Base(workspace), "foundry-builder-") {
		t.Fatal("workspace", got)
	}
	if got["handoffPath"] != filepath.Join(workspace, "handoff.md") || got["sourceManifestSHA256"] != digest(raw) {
		t.Fatal(got)
	}
	compareReport(t, got["sourceReport"], raw, root)
	compareReport(t, got["inputsReport"], raw, filepath.Join(workspace, "inputs"))
	if !bytes.Equal(read(t, filepath.Join(workspace, "source-manifest.json")), raw) {
		t.Fatal("source manifest bytes changed")
	}
	wantFiles := map[string]bool{"source-manifest.json": true, "source-report.json": true, "exercise-manifest.json": true, "inputs-report.json": true, "exercise-report.json": true, "handoff.md": true}
	for _, a := range m.Artifacts {
		original := read(t, filepath.Join(root, a.Path))
		if digest(original) != a.SHA256 {
			t.Fatal("source changed", a.Path)
		}
		copied := filepath.Join("inputs", a.Path)
		wantFiles[copied] = true
		if !bytes.Equal(read(t, filepath.Join(workspace, copied)), original) {
			t.Fatal("copy differs", a.Path)
		}
		if strings.HasPrefix(a.Path, "examples/model-builder-go/acceptance/") {
			p := filepath.Join("exercise", filepath.Base(a.Path))
			wantFiles[p] = true
			if !bytes.Equal(read(t, filepath.Join(workspace, p)), original) {
				t.Fatal("exercise differs", p)
			}
		}
	}
	exerciseRaw := read(t, filepath.Join(workspace, "exercise-manifest.json"))
	var em fixedManifest
	if e := json.Unmarshal(exerciseRaw, &em); e != nil {
		t.Fatal(e)
	}
	if em.FormatVersion != "model-builder-inputs/v1" || len(em.Artifacts) != 5 {
		t.Fatal("exercise inventory", em)
	}
	seen := map[string]bool{}
	for _, a := range em.Artifacts {
		if a.Path != filepath.Base(a.Path) || seen[a.Path] {
			t.Fatal("exercise path", a)
		}
		seen[a.Path] = true
		found := false
		for _, src := range m.Artifacts {
			if src.Path == "examples/model-builder-go/acceptance/"+a.Path && src.SHA256 == a.SHA256 {
				found = true
			}
		}
		if !found {
			t.Fatal("not frozen exercise input", a)
		}
	}
	compareReport(t, got["exerciseReport"], exerciseRaw, filepath.Join(workspace, "exercise"))
	for name, key := range map[string]string{"source-report.json": "sourceReport", "inputs-report.json": "inputsReport", "exercise-report.json": "exerciseReport"} {
		if !reflect.DeepEqual(decode(t, read(t, filepath.Join(workspace, name))), got[key]) {
			t.Fatal("saved report differs", name)
		}
	}
	for _, d := range []string{"inputs", "exercise", "cache", "tmp", "modcache"} {
		info, e := os.Stat(filepath.Join(workspace, d))
		if e != nil || !info.IsDir() {
			t.Fatal("missing directory", d, e)
		}
	}
	err := filepath.Walk(workspace, func(p string, info os.FileInfo, e error) error {
		if e != nil {
			return e
		}
		if info.IsDir() {
			if info.Mode().Perm()&0077 != 0 || info.Mode().Perm()&0700 != 0700 {
				t.Errorf("directory permissions %s %v", p, info.Mode())
			}
			return nil
		}
		rel, _ := filepath.Rel(workspace, p)
		if !wantFiles[rel] {
			t.Errorf("unexpected file %s", rel)
		}
		delete(wantFiles, rel)
		if !info.Mode().IsRegular() || info.Mode().Perm()&0177 != 0 {
			t.Errorf("file permissions %s %v", p, info.Mode())
		}
		return nil
	})
	if err != nil || len(wantFiles) != 0 {
		t.Fatal("file frontier", wantFiles, err)
	}
	handoff := string(read(t, filepath.Join(workspace, "handoff.md")))
	if !strings.Contains(strings.ToLower(strings.SplitN(handoff, "\n\n", 2)[0]), "initial read boundary") {
		t.Fatal("initial boundary must precede other content")
	}
	for _, a := range m.Artifacts {
		if !strings.Contains(handoff, filepath.Join(workspace, "inputs", a.Path)) || !strings.Contains(handoff, a.SHA256) {
			t.Fatal("handoff lacks pinned absolute input", a.Path)
		}
	}
	for _, s := range []string{filepath.Join(workspace, "exercise", "candidate.go"), filepath.Join(workspace, "exercise", "candidate.test"), digest(exerciseRaw), "Q_use", "go test -run", "go test -c"} {
		if !strings.Contains(handoff, s) {
			t.Errorf("missing handoff boundary %q", s)
		}
	}
	for _, s := range []string{"visual", "historical", "coordinator", "claim", "independen", "review", "authority", "initializ"} {
		if !strings.Contains(strings.ToLower(handoff), s) {
			t.Errorf("missing handoff gate %s", s)
		}
	}
	r2, e := builderdemo.Prepare(root, parent, tool)
	if e != nil {
		t.Fatal(e)
	}
	if receipt(t, r2)["workspace"] == workspace {
		t.Fatal("reused workspace")
	}
	for _, a := range m.Artifacts {
		if digest(read(t, filepath.Join(root, a.Path))) != a.SHA256 {
			t.Fatal("original changed", a.Path)
		}
	}
}

func TestBlockedSourcesHaveCompleteUnchangedReportAndNoWrites(t *testing.T) {
	for _, kind := range []string{"stale", "missing", "directory", "leaf-symlink", "parent-symlink", "oversize"} {
		t.Run(kind, func(t *testing.T) {
			root, parent := fixture(t), t.TempDir()
			raw, m := inventory(t)
			target := filepath.Join(root, m.Artifacts[0].Path)
			switch kind {
			case "stale":
				if e := os.WriteFile(target, []byte("stale"), 0600); e != nil {
					t.Fatal(e)
				}
			case "oversize":
				if e := os.WriteFile(target, bytes.Repeat([]byte("x"), (1<<20)+1), 0600); e != nil {
					t.Fatal(e)
				}
			case "missing", "directory", "leaf-symlink":
				if e := os.Remove(target); e != nil {
					t.Fatal(e)
				}
				if kind == "directory" {
					if e := os.Mkdir(target, 0700); e != nil {
						t.Fatal(e)
					}
				}
				if kind == "leaf-symlink" {
					if e := os.Symlink(filepath.Join("../../..", m.Artifacts[0].Path), target); e != nil {
						t.Fatal(e)
					}
				}
			case "parent-symlink":
				dir := filepath.Dir(target)
				if e := os.Rename(dir, dir+"-saved"); e != nil {
					t.Fatal(e)
				}
				if e := os.Symlink(dir+"-saved", dir); e != nil {
					t.Fatal(e)
				}
			}
			r, e := builderdemo.Prepare(root, parent, "/not/a/real/go")
			if !errors.Is(e, builderdemo.ErrBlocked) {
				t.Fatalf("error %v", e)
			}
			got := receipt(t, r)
			envelope(t, got, "blocked")
			if got["workspace"] != nil || got["handoffPath"] != nil || got["inputsReport"] != nil || got["exerciseReport"] != nil {
				t.Fatal("output before validation", got)
			}
			compareReport(t, got["sourceReport"], raw, root)
			empty(t, parent)
		})
	}
}

func TestInvalidPathsAndLocationsWriteNothing(t *testing.T) {
	root, parent := fixture(t), t.TempDir()
	file := filepath.Join(parent, "file")
	if e := os.WriteFile(file, nil, 0600); e != nil {
		t.Fatal(e)
	}
	alias := filepath.Join(t.TempDir(), "alias")
	if e := os.Symlink(root, alias); e != nil {
		t.Fatal(e)
	}
	parentAlias := filepath.Join(t.TempDir(), "parent")
	if e := os.Symlink(parent, parentAlias); e != nil {
		t.Fatal(e)
	}
	inside := filepath.Join(root, "output")
	if e := os.Mkdir(inside, 0700); e != nil {
		t.Fatal(e)
	}
	cases := [][3]string{{root, root, "/go"}, {root, inside, "/go"}, {alias, parent, "/go"}, {alias + "/", parent, "/go"}, {root, parentAlias, "/go"}, {root, parentAlias + "/", "/go"}, {file, parent, "/go"}, {root, file, "/go"}, {root, filepath.Join(parent, "missing"), "/go"}}
	for _, bad := range []string{"", ".", "relative", "/bad\npath", "/bad\x00path", "/bad\x7fpath", "/bad\xffpath"} {
		cases = append(cases, [3]string{bad, parent, "/go"}, [3]string{root, bad, "/go"}, [3]string{root, parent, bad})
	}
	for _, c := range cases {
		_, e := builderdemo.Prepare(c[0], c[1], c[2])
		if !errors.Is(e, builderdemo.ErrInvalid) {
			t.Fatalf("paths %q: %v", c, e)
		}
	}
	entries, e := os.ReadDir(parent)
	if e != nil || len(entries) != 1 || entries[0].Name() != "file" {
		t.Fatal("invalid arguments wrote output", entries, e)
	}
	empty(t, inside)
}

func TestHostTraversalIsNotLexicallyCleaned(t *testing.T) {
	root, parent := fixture(t), t.TempDir()
	base := t.TempDir()
	deep := filepath.Join(base, "real", "deep")
	if e := os.MkdirAll(deep, 0700); e != nil {
		t.Fatal(e)
	}
	alias := filepath.Join(base, "alias")
	if e := os.Symlink(deep, alias); e != nil {
		t.Fatal(e)
	}
	// alias/../source resolves in real/, not in base/.
	if e := os.Rename(root, filepath.Join(base, "real", "source")); e != nil {
		t.Fatal(e)
	}
	r, e := builderdemo.Prepare(alias+"/../source", parent, "/record-only-go")
	if e != nil {
		t.Fatal("host traversal must resolve actual source", e)
	}
	envelope(t, receipt(t, r), "prepared-draft")
	for _, bad := range []string{base + "/absent/../real/source", filepath.Join(base, "real", "source", "AGENTS.md") + "/../"} {
		_, e := builderdemo.Prepare(bad, parent, "/go")
		if !errors.Is(e, builderdemo.ErrInvalid) {
			t.Fatal("erased invalid traversal", bad, e)
		}
	}
}

func TestHandoffCommandIsDataWithLiteralHostPaths(t *testing.T) {
	root := fixture(t)
	parent := filepath.Join(t.TempDir(), "spaces ' quotes $HOME ; $(touch SENTINEL)")
	if e := os.Mkdir(parent, 0700); e != nil {
		t.Fatal(e)
	}
	tool := filepath.Join(t.TempDir(), "go ' ; $(touch TOOL_SENTINEL)")
	r, e := builderdemo.Prepare(root, parent, tool)
	if e != nil {
		t.Fatal(e)
	}
	got := receipt(t, r)
	ws := got["workspace"].(string)
	raw := read(t, got["handoffPath"].(string))
	var command map[string]any
	// Locate a JSON object without imposing a Markdown fence or prose layout.
	for i, b := range raw {
		if b != '{' {
			continue
		}
		var v map[string]any
		if json.NewDecoder(bytes.NewReader(raw[i:])).Decode(&v) == nil && v["program"] != nil {
			command = v
			break
		}
	}
	if command == nil {
		t.Fatal("missing command JSON")
	}
	if command["program"] != tool || command["directory"] != filepath.Join(ws, "exercise") {
		t.Fatal("command paths", command)
	}
	if !reflect.DeepEqual(command["args"], []any{"test", "-c", "-o", filepath.Join(ws, "exercise", "candidate.test")}) {
		t.Fatal("not compile-only argv", command)
	}
	env, ok := command["environment"].(map[string]any)
	if !ok {
		t.Fatal("environment must be a JSON object")
	}
	for k, v := range map[string]string{"GOTOOLCHAIN": "local", "GOPROXY": "off", "GOSUMDB": "off", "GOWORK": "off", "GOENV": "off", "GOFLAGS": "", "GOCACHE": filepath.Join(ws, "cache"), "GOTMPDIR": filepath.Join(ws, "tmp"), "TMPDIR": filepath.Join(ws, "tmp"), "GOMODCACHE": filepath.Join(ws, "modcache")} {
		if env[k] != v {
			t.Errorf("%s=%v want %q", k, env[k], v)
		}
	}
	for _, p := range []string{filepath.Join(ws, "exercise", "candidate.go"), filepath.Join(ws, "exercise", "candidate.test"), filepath.Join(parent, "SENTINEL")} {
		if _, e := os.Lstat(p); !os.IsNotExist(e) {
			t.Fatal("candidate or injection output exists", p, e)
		}
	}
}

func TestUnwritableParentIsIncompleteNotPrepared(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root bypasses permission bits; no false permission-failure claim")
	}
	root, parent := fixture(t), t.TempDir()
	if e := os.Chmod(parent, 0500); e != nil {
		t.Fatal(e)
	}
	defer os.Chmod(parent, 0700)
	r, e := builderdemo.Prepare(root, parent, "/record-only-go")
	if e == nil || errors.Is(e, builderdemo.ErrInvalid) || errors.Is(e, builderdemo.ErrBlocked) {
		t.Fatalf("creation failure must be incomplete: %v", e)
	}
	got := receipt(t, r)
	envelope(t, got, "incomplete")
	if got["workspace"] != nil || got["handoffPath"] != nil || got["inputsReport"] != nil || got["exerciseReport"] != nil {
		t.Fatal("invented output", got)
	}
	raw, _ := inventory(t)
	compareReport(t, got["sourceReport"], raw, root)
	empty(t, parent)
}

func TestPhysicalParentOverlapAndAncestorAlias(t *testing.T) {
	root, parent := fixture(t), t.TempDir()
	base := t.TempDir()
	alias := filepath.Join(base, "source-alias")
	if e := os.Symlink(root, alias); e != nil {
		t.Fatal(e)
	}
	child := filepath.Join(root, "output")
	if e := os.Mkdir(child, 0700); e != nil {
		t.Fatal(e)
	}
	// Ancestor aliases may be traversed, but they do not hide source overlap.
	for _, p := range []string{alias + "/output", alias + "/output/.."} {
		_, e := builderdemo.Prepare(root, p, "/go")
		if !errors.Is(e, builderdemo.ErrInvalid) {
			t.Fatalf("physical overlap %q: %v", p, e)
		}
	}
	empty(t, child)
	realChild := filepath.Join(parent, "child")
	if e := os.Mkdir(realChild, 0700); e != nil {
		t.Fatal(e)
	}
	outAlias := filepath.Join(base, "output-alias")
	if e := os.Symlink(realChild, outAlias); e != nil {
		t.Fatal(e)
	}
	r, e := builderdemo.Prepare(root, outAlias+"/..", "/go")
	if e != nil {
		t.Fatal("valid host parent traversal", e)
	}
	got := receipt(t, r)
	ws := got["workspace"].(string)
	info, e := os.Stat(filepath.Dir(ws))
	if e != nil {
		t.Fatal(e)
	}
	want, e := os.Stat(parent)
	if e != nil || !os.SameFile(info, want) {
		t.Fatal("workspace created under wrong host target", ws, e)
	}
}

// This supplements, rather than rewrites, the original frozen acceptance.
func TestHistoricalInstructionsReplayAndChangedInstructionsRefuse(t *testing.T) {
	historical := read(t, "testdata/AGENTS-882b5002.txt")
	if digest(historical) != "4574ee25fe89b6658ae237f3f50648ef1e542c487a2c23d7ab2833696f9455eb" {
		t.Fatal("historical instruction snapshot changed")
	}
	current := read(t, "../../../AGENTS.md")
	mutated := append(append([]byte{}, historical...), []byte("\nfixture-only mutation\n")...)
	cases := []struct {
		name    string
		input   []byte
		allowed bool
	}{
		{"historical", historical, true},
		// Eligibility follows the existing byte pin, not whether a source is
		// called current. Restoring the exact old bytes would be eligible too.
		{"current", current, bytes.Equal(current, historical)},
		{"mutated", mutated, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root, parent := fixture(t), t.TempDir()
			target := filepath.Join(root, "AGENTS.md")
			if !bytes.Equal(read(t, target), historical) {
				t.Fatal("fixture did not replay historical instructions")
			}
			if e := os.WriteFile(target, tc.input, 0600); e != nil {
				t.Fatal(e)
			}
			r, e := builderdemo.Prepare(root, parent, "/not/a/real/go")
			got := receipt(t, r)
			if tc.allowed {
				if e != nil {
					t.Fatal(e)
				}
				envelope(t, got, "prepared-draft")
				workspace, ok := got["workspace"].(string)
				if !ok || !bytes.Equal(read(t, filepath.Join(workspace, "inputs", "AGENTS.md")), historical) {
					t.Fatal("prepared instructions differ from the historical snapshot")
				}
			} else {
				if !errors.Is(e, builderdemo.ErrBlocked) {
					t.Fatalf("changed instructions were not blocked: %v", e)
				}
				envelope(t, got, "blocked")
				if got["workspace"] != nil || got["handoffPath"] != nil || got["inputsReport"] != nil || got["exerciseReport"] != nil {
					t.Fatal("output before source validation", got)
				}
				empty(t, parent)
			}
			raw, _ := inventory(t)
			compareReport(t, got["sourceReport"], raw, root)
			if !bytes.Equal(read(t, target), tc.input) {
				t.Fatal("preparation rewrote the instruction input")
			}
		})
	}
}
