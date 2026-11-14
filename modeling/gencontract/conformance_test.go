package gencontract

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"testing"
)

var (
	buildDir   string
	buildMu    sync.Mutex
	buildCache = map[string]string{}
)

func TestMain(m *testing.M) {
	d, err := os.MkdirTemp("", "gencontract-bins-")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	buildDir = d
	code := m.Run()
	os.RemoveAll(d)
	os.Exit(code)
}

// goBuild builds pkg in module dir moduleDir into a binary called name.
func goBuild(t *testing.T, moduleDir, pkg, name string, tags ...string) string {
	t.Helper()
	buildMu.Lock()
	defer buildMu.Unlock()
	key := moduleDir + "|" + pkg + "|" + name + "|" + strings.Join(tags, ",")
	if bin, ok := buildCache[key]; ok {
		return bin
	}
	bin := filepath.Join(buildDir, fmt.Sprintf("%02d-%s", len(buildCache), name))
	args := []string{"build", "-o", bin}
	if len(tags) > 0 {
		args = append(args, "-tags", strings.Join(tags, ","))
	}
	args = append(args, pkg)
	cmd := exec.Command("go", args...)
	cmd.Dir = moduleDir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("go %s in %s: %v\n%s", strings.Join(args, " "), moduleDir, err, out)
	}
	buildCache[key] = bin
	return bin
}

func refgen(t *testing.T, env ...string) Generator {
	t.Helper()
	abs, err := filepath.Abs("testdata/reference/model.json")
	if err != nil {
		t.Fatal(err)
	}
	tags := []string{}
	if len(env) > 0 {
		tags = append(tags, "refgenfault")
	}
	return Generator{
		ID:      "refgen",
		Profile: "reference/v1",
		Bin:     goBuild(t, ".", "./cmd/refgen", "refgen", tags...),
		Model:   abs,
		Env:     env,
	}
}

func rules(fs []Finding) map[string]bool {
	m := map[string]bool{}
	for _, f := range fs {
		m[f.Rule] = true
	}
	return m
}

func listFindings(fs []Finding) string {
	var b strings.Builder
	for _, f := range fs {
		b.WriteString("\n  " + f.String())
	}
	return b.String()
}

func TestReferenceGeneratorConforms(t *testing.T) {
	findings, err := Check(refgen(t), t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 0 {
		t.Fatalf("unexpected findings:%s", listFindings(findings))
	}
}

func TestFaultBuildWithoutFaultConforms(t *testing.T) {
	findings, err := Check(refgen(t, "REFGEN_FAULT="), t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 0 {
		t.Fatalf("unexpected findings:%s", listFindings(findings))
	}
}

// Each injected fault must make the harness report the named rule.
func TestHarnessDetectsFaults(t *testing.T) {
	faults := []struct{ fault, rule string }{
		{"nondeterministic", "generate.deterministic"},
		{"no-header", "header"},
		{"manifest-compact", "manifest.encoding"},
		{"manifest-unsorted", "manifest.files"},
		{"manifest-missing-file", "manifest.files"},
		{"check-always-ok", "check.modified"},
		{"check-always-ok", "check.missing"},
		{"check-repairs", "check.repaired"},
		{"overwrite-out", "generate.existing-out"},
		{"accept-invalid", "refuse.invalid-model"},
		{"usage-exit1", "refuse.usage"},
		{"canonical-whitespace", "validate.whitespace"},
		{"stdout-on-failure", "refuse.usage"},
		{"stdout-on-failure", "refuse.invalid-model"},
		{"ignore-extra", "check.unexpected"},
		{"follow-output-root", "check.symlink-root"},
		{"clean-output-path", "check.path-resolution"},
	}
	byFault := map[string][]string{}
	var order []string
	for _, f := range faults {
		if _, ok := byFault[f.fault]; !ok {
			order = append(order, f.fault)
		}
		byFault[f.fault] = append(byFault[f.fault], f.rule)
	}
	for _, name := range order {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if name == "follow-output-root" || name == "clean-output-path" {
				dir := t.TempDir()
				if err := os.Symlink(dir, filepath.Join(dir, "link")); err != nil {
					t.Skipf("symlinks unavailable: %v", err)
				}
			}
			findings, err := Check(refgen(t, "REFGEN_FAULT="+name), t.TempDir())
			if err != nil {
				t.Fatal(err)
			}
			got := rules(findings)
			for _, rule := range byFault[name] {
				if !got[rule] {
					t.Errorf("fault %q: rule %q not reported; findings:%s", name, rule, listFindings(findings))
				}
			}
		})
	}
}

func TestReformatChangesOnlyWhitespace(t *testing.T) {
	for _, in := range []string{
		`{"a":1.50,"b":[1,2]}`,
		"{\n  \"a\": 1.50,\n  \"b\": [1, 2]\n}\n",
	} {
		out, err := reformat([]byte(in))
		if err != nil {
			t.Fatal(err)
		}
		if bytes.Equal(out, []byte(in)) {
			t.Errorf("reformat(%q) did not change the bytes", in)
		}
		var a, b bytes.Buffer
		if err := json.Compact(&a, []byte(in)); err != nil {
			t.Fatal(err)
		}
		if err := json.Compact(&b, out); err != nil {
			t.Fatal(err)
		}
		if a.String() != b.String() {
			t.Errorf("reformat(%q) changed more than whitespace: %q", in, out)
		}
	}
}

func TestHarnessProblemsAreErrors(t *testing.T) {
	g := refgen(t)
	g.Model = filepath.Join(t.TempDir(), "absent.json")
	if _, err := Check(g, t.TempDir()); err == nil {
		t.Error("missing sample model: want an error")
	}
	g = refgen(t)
	g.Bin = filepath.Join(t.TempDir(), "no-such-binary")
	if _, err := Check(g, t.TempDir()); err == nil {
		t.Error("missing binary: want an error")
	}
}

type registry struct {
	RegistryVersion string `json:"registryVersion"`
	Generators      []struct {
		ID      string `json:"id"`
		Profile string `json:"profile"`
		Module  string `json:"module"`
		Package string `json:"package"`
		Model   string `json:"model"`
	} `json:"generators"`
}

func TestRegisteredGeneratorsConform(t *testing.T) {
	raw, err := os.ReadFile("generators.json")
	if err != nil {
		t.Fatal(err)
	}
	var reg registry
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&reg); err != nil {
		t.Fatalf("generators.json: %v", err)
	}
	if dec.More() {
		t.Fatal("generators.json: data after the registry object")
	}
	if reg.RegistryVersion != "foundry-generator-registry/v1" {
		t.Fatalf("registryVersion is %q", reg.RegistryVersion)
	}
	if len(reg.Generators) == 0 {
		t.Fatal("registry lists no generators")
	}
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	idRe := regexp.MustCompile(`^[a-z0-9-]+$`)
	seen := map[string]bool{}
	for _, e := range reg.Generators {
		t.Run(e.ID, func(t *testing.T) {
			if !idRe.MatchString(e.ID) || seen[e.ID] {
				t.Fatalf("id %q is invalid or duplicated", e.ID)
			}
			seen[e.ID] = true
			if e.Profile == "" || e.Module == "" || e.Package == "" || e.Model == "" {
				t.Fatalf("incomplete entry: %+v", e)
			}
			model := filepath.Join(root, filepath.FromSlash(e.Model))
			if _, err := os.Stat(model); err != nil {
				t.Fatal(err)
			}
			bin := goBuild(t, filepath.Join(root, filepath.FromSlash(e.Module)), e.Package, e.ID)
			findings, err := Check(Generator{ID: e.ID, Profile: e.Profile, Bin: bin, Model: model}, t.TempDir())
			if err != nil {
				t.Fatal(err)
			}
			if len(findings) != 0 {
				t.Fatalf("%s does not conform:%s", e.ID, listFindings(findings))
			}
		})
	}
}
