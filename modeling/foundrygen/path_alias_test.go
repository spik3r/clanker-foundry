package foundrygen

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

const aliasModel = `{"profile":"json-boolean/v1","root":{"type":"string"}}`

// Seed a real generated directory, including its actual manifest. The model
// later placed inside it must survive even when OUT spells the directory differently.
func ownedAliasProject(t *testing.T, real, alias string) string {
	t.Helper()
	project := t.TempDir()
	writeFile(t, filepath.Join(project, "seed.json"), aliasModel)
	writeFile(t, filepath.Join(project, ProjectFile), projectJSON(out("jsonvalidator", "seed.json", real)))
	if code, so, se := runCLI(t, "generate", project); code != 0 {
		t.Fatalf("seed output: %d %q %q", code, so, se)
	}
	a, err := os.Stat(filepath.Join(project, real))
	if err != nil {
		t.Fatal(err)
	}
	b, err := os.Stat(filepath.Join(project, alias))
	if os.IsNotExist(err) {
		t.Skip("filesystem distinguishes the fixture names")
	}
	if err != nil {
		t.Fatal(err)
	}
	if !os.SameFile(a, b) {
		t.Skip("filesystem distinguishes the fixture names")
	}
	return project
}

func requireAliasRefusal(t *testing.T, verb, project string) {
	t.Helper()
	before := snapshot(t, project)
	code, stdout, stderr := runCLI(t, verb, project)
	if code != exitRefused || stdout != "" || !strings.Contains(stderr, "aliases") {
		t.Errorf("%s: want alias refusal with empty stdout, got %d %q %q", verb, code, stdout, stderr)
	}
	if !sameSnapshot(before, snapshot(t, project)) {
		t.Error("alias refusal changed project bytes or mtimes")
	}
}

func TestExistingDirectoryAliasesRefused(t *testing.T) {
	for _, verb := range []string{"plan", "check", "generate"} {
		for _, scenario := range []string{"own model", "other model", "missing model", "equal outputs", "nested outputs"} {
			t.Run(verb+"/"+scenario, func(t *testing.T) {
				project := ownedAliasProject(t, "data", "DATA")
				writeFile(t, filepath.Join(project, "data/model.json"), aliasModel)
				outputs := out("jsonvalidator", "data/model.json", "DATA")
				switch scenario {
				case "other model":
					outputs = out("jsonvalidator", "seed.json", "DATA") + "," + out("jsonvalidator", "data/model.json", "other")
				case "missing model":
					outputs = out("jsonvalidator", "data/missing/model.json", "DATA")
				case "equal outputs":
					outputs = out("jsonvalidator", "seed.json", "data") + "," + out("jsonvalidator", "seed.json", "DATA")
				case "nested outputs":
					outputs = out("jsonvalidator", "seed.json", "data") + "," + out("jsonvalidator", "seed.json", "DATA/child")
				}
				writeFile(t, filepath.Join(project, ProjectFile), projectJSON(outputs))
				requireAliasRefusal(t, verb, project)
			})
		}
	}
}

func TestUnicodeDirectoryAliasRefused(t *testing.T) {
	for _, verb := range []string{"plan", "check", "generate"} {
		t.Run(verb, func(t *testing.T) {
			project := ownedAliasProject(t, "caf\u00e9", "cafe\u0301")
			writeFile(t, filepath.Join(project, "caf\u00e9/model.json"), aliasModel)
			writeFile(t, filepath.Join(project, ProjectFile), projectJSON(out("jsonvalidator", "caf\u00e9/model.json", "cafe\u0301")))
			requireAliasRefusal(t, verb, project)
		})
	}
}

func TestHardlinkedInputAliasesRefused(t *testing.T) {
	for _, verb := range []string{"plan", "check", "generate"} {
		for _, source := range []string{"model.json", ProjectFile} {
			t.Run(verb+"/"+source, func(t *testing.T) {
				project := t.TempDir()
				writeFile(t, filepath.Join(project, "model.json"), aliasModel)
				writeFile(t, filepath.Join(project, ProjectFile), projectJSON(out("jsonvalidator", "model.json", "output-copy")))
				if err := os.Link(filepath.Join(project, source), filepath.Join(project, "output-copy")); err != nil {
					t.Fatal(err)
				}
				requireAliasRefusal(t, verb, project)
			})
		}
	}
}

func TestCaseDistinctDirectoriesRemainAllowed(t *testing.T) {
	project := t.TempDir()
	writeFile(t, filepath.Join(project, "data/model.json"), aliasModel)
	if _, err := os.Stat(filepath.Join(project, "DATA")); err == nil {
		t.Skip("filesystem aliases the fixture names")
	} else if !os.IsNotExist(err) {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(project, ProjectFile), projectJSON(out("jsonvalidator", "data/model.json", "DATA")))
	for _, verb := range []string{"plan", "generate", "check"} {
		if code, so, se := runCLI(t, verb, project); code != 0 {
			t.Fatalf("%s: %d %q %q", verb, code, so, se)
		}
	}
	if got := readFile(t, filepath.Join(project, "data/model.json")); got != aliasModel {
		t.Fatal("distinct input changed")
	}
}

// The metadata guard also covers ordinary ancestry independent of filesystem
// case rules. These sessions exercise it directly because ParseProject already
// refuses the same spellings before the execution preflight can see them.
func TestPathIdentityAncestorGuard(t *testing.T) {
	for _, scenario := range []string{"model", "missing model", "equal outputs", "nested outputs"} {
		t.Run(scenario, func(t *testing.T) {
			project := t.TempDir()
			writeFile(t, filepath.Join(project, ProjectFile), "{}")
			writeFile(t, filepath.Join(project, "seed.json"), aliasModel)
			writeFile(t, filepath.Join(project, "owned/model.json"), aliasModel)
			outputs := []Output{{Generator: "jsonvalidator", Model: "owned/model.json", Out: "owned"}}
			switch scenario {
			case "missing model":
				outputs[0].Model = "owned/missing/model.json"
			case "equal outputs":
				outputs = []Output{{Model: "seed.json", Out: "owned"}, {Model: "seed.json", Out: "owned"}}
			case "nested outputs":
				outputs = []Output{{Model: "seed.json", Out: "owned"}, {Model: "seed.json", Out: "owned/missing/child"}}
			}
			s := &session{root: project, proj: &Project{Outputs: outputs}}
			before := snapshot(t, project)
			if err := s.checkPathAliases(); err == nil || codeOf(err) != exitRefused {
				t.Fatalf("unsafe ancestry accepted: %v", err)
			}
			if !sameSnapshot(before, snapshot(t, project)) {
				t.Fatal("metadata guard changed files")
			}
		})
	}
}

// Earlier output installation can reveal an alias that had no identity during
// plan. Refuse the later write instead of claiming both outputs are valid.
func TestNewNestedAliasRefusedAfterFirstWrite(t *testing.T) {
	project := t.TempDir()
	writeFile(t, filepath.Join(project, "design/model.json"), aliasModel)
	a, err := os.Stat(filepath.Join(project, "design"))
	if err != nil {
		t.Fatal(err)
	}
	b, err := os.Stat(filepath.Join(project, "DESIGN"))
	if os.IsNotExist(err) {
		t.Skip("filesystem distinguishes the fixture names")
	}
	if err != nil {
		t.Fatal(err)
	}
	if !os.SameFile(a, b) {
		t.Skip("filesystem distinguishes the fixture names")
	}
	writeFile(t, filepath.Join(project, ProjectFile), projectJSON(out("jsonvalidator", "design/model.json", "data")+","+out("jsonvalidator", "design/model.json", "DATA/child")))
	before := snapshot(t, project)
	if code, so, se := runCLI(t, "plan", project); code != 0 {
		t.Fatalf("plan: %d %q %q", code, so, se)
	}
	if !sameSnapshot(before, snapshot(t, project)) {
		t.Fatal("plan changed project files")
	}
	code, so, se := runCLI(t, "generate", project)
	if code != exitRefused || so != "" || !strings.Contains(se, "aliases") || !strings.Contains(se, "data: already generated before the failure") {
		t.Errorf("generate: want alias refusal with prior-result diagnostic, got %d %q %q", code, so, se)
	}
	if _, err := os.Lstat(filepath.Join(project, "data/child")); !os.IsNotExist(err) {
		t.Errorf("nested output was installed: %v", err)
	}
	after := snapshot(t, project)
	originals := map[string]fileState{}
	for name := range before {
		originals[name] = after[name]
	}
	if !sameSnapshot(before, originals) {
		t.Error("generation changed source files")
	}
	// Check the actual earlier result with the independent generator command,
	// without changing the project file to make the invalid pair disappear.
	cmd := exec.CommandContext(t.Context(), "go", "run", "./cmd/jsonvalidator", "check", filepath.Join(project, "design/model.json"), filepath.Join(project, "data"))
	cmd.Dir = filepath.Join(repoRoot(t), "modeling/jsonvalidator")
	cmd.Env = append(os.Environ(), "GOFLAGS=-mod=readonly", "GOWORK=off")
	if result, err := cmd.CombinedOutput(); err != nil {
		t.Errorf("earlier output is not valid: %v %s", err, result)
	}
}
