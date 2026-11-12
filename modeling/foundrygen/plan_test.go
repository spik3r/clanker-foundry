package foundrygen

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func planActions(t *testing.T, project string, want ...string) string {
	t.Helper()
	before := snapshot(t, project)
	code, stdout, stderr := runCLI(t, "plan", project)
	if code != 0 {
		t.Fatalf("plan: code %d stdout %q stderr %q", code, stdout, stderr)
	}
	if !sameSnapshot(before, snapshot(t, project)) {
		t.Fatal("plan changed project bytes or mtimes")
	}
	var report struct {
		PlanVersion string `json:"planVersion"`
		Outputs     []struct {
			Action, Generator, Model, Out string
			Identity                      Identity
		} `json:"outputs"`
	}
	if err := json.Unmarshal([]byte(stdout), &report); err != nil {
		t.Fatal(err)
	}
	if report.PlanVersion != "foundry-plan/v1" || len(report.Outputs) != len(want) {
		t.Fatalf("unexpected plan: %s", stdout)
	}
	for i, o := range report.Outputs {
		wantOut := []string{"internal/browser", "internal/sample"}[i]
		wantModel := []string{"design/browser.json", "design/sample.json"}[i]
		wantGenerator := []string{"tuic", "jsonvalidator"}[i]
		if o.Out != wantOut || o.Model != wantModel || o.Generator != wantGenerator {
			t.Fatalf("plan output order/paths changed: %+v", o)
		}
		digest := fmt.Sprintf("%x", sha256.Sum256([]byte(readFile(t, filepath.Join(project, o.Model)))))
		if o.Identity.SourceSHA256 != digest {
			t.Fatalf("wrong source digest: %+v", o)
		}
		if o.Action != want[i] || o.Model == "" || o.Out == "" || o.Generator == "" || o.Identity.Contract != contractID || !hex64RE.MatchString(o.Identity.CanonicalSHA256) || !hex64RE.MatchString(o.Identity.SourceSHA256) {
			t.Fatalf("bad output %d: %+v", i, o)
		}
	}
	return stdout
}

func TestPlanRealGenerators(t *testing.T) {
	project := scratch(t)
	first := planActions(t, project, "create", "create")
	if _, err := os.Stat(filepath.Join(project, "internal")); !os.IsNotExist(err) {
		t.Fatalf("plan created output parent: %v", err)
	}
	if again := planActions(t, project, "create", "create"); again != first {
		t.Fatal("plan is not deterministic")
	}
	if code, _, stderr := runCLI(t, "generate", project); code != 0 {
		t.Fatalf("generate: %d %s", code, stderr)
	}
	planActions(t, project, "unchanged", "unchanged")
	flipByte(t, firstGenerated(t, filepath.Join(project, "internal/browser")))
	planActions(t, project, "replace", "unchanged")
	if code, _, stderr := runCLI(t, "generate", project); code != 0 {
		t.Fatalf("repair: %d %s", code, stderr)
	}
	planActions(t, project, "unchanged", "unchanged")
}

func TestPlanRefusesWithoutOutput(t *testing.T) {
	for _, scenario := range []string{"invalid-model", "unowned-output", "symlink-output", "missing-model"} {
		t.Run(scenario, func(t *testing.T) {
			project := scratch(t)
			switch scenario {
			case "invalid-model":
				writeFile(t, filepath.Join(project, "design/sample.json"), "{}")
			case "unowned-output":
				writeFile(t, filepath.Join(project, "internal/sample/keep.txt"), "keep")
			case "symlink-output":
				if err := os.MkdirAll(filepath.Join(project, "internal"), 0755); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(t.TempDir(), filepath.Join(project, "internal/sample")); err != nil {
					t.Fatal(err)
				}
			case "missing-model":
				if err := os.Remove(filepath.Join(project, "design/sample.json")); err != nil {
					t.Fatal(err)
				}
			}
			// snapshot follows file symlinks; the output link is checked separately below.
			var before map[string]fileState
			if scenario != "symlink-output" {
				before = snapshot(t, project)
			}
			code, stdout, stderr := runCLI(t, "plan", project)
			if code == 0 || stdout != "" || stderr == "" {
				t.Fatalf("expected refusal: %d %q %q", code, stdout, stderr)
			}
			if before != nil && !sameSnapshot(before, snapshot(t, project)) {
				t.Fatal("failed plan modified project")
			}
			if _, err := os.Stat(filepath.Join(project, "internal/browser")); !os.IsNotExist(err) {
				t.Fatal("plan created earlier valid output")
			}
			if scenario == "symlink-output" {
				fi, err := os.Lstat(filepath.Join(project, "internal/sample"))
				if err != nil || fi.Mode()&os.ModeSymlink == 0 {
					t.Fatal("plan replaced symlink")
				}
			}
		})
	}
}

func TestPlanHonoursPin(t *testing.T) {
	project := scratch(t)
	var stdout, stderr strings.Builder
	code := Run([]string{"plan", "--project", project, "--foundry", repoRoot(t)}, &stdout, &stderr)
	if code != exitRefused || stdout.Len() != 0 || !strings.Contains(stderr.String(), "pins") {
		t.Fatalf("pin: %d %s %s", code, stdout.String(), stderr.String())
	}
}
