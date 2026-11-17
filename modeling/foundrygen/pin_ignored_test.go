package foundrygen

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPinRefusesIgnoredBuildInputs(t *testing.T) {
	for _, ignoreSource := range []string{"tracked", "local", "configured"} {
		for _, kind := range []string{"go", "embed"} {
			t.Run(ignoreSource+"/"+kind, func(t *testing.T) {
				dir, _ := gitFixture(t)
				module := filepath.Join(dir, "modeling/fake")
				writeFile(t, filepath.Join(module, "go.mod"), "module example.test/fixture\n\ngo 1.27.0\n")
				main := "package main\nimport \"fmt\"\nvar message = \"pinned\"\nfunc main() { fmt.Print(message) }\n"
				path, ignored := "modeling/fake/extra.go", "package main\nfunc init() { message = \"ignored\" }\n"
				if kind == "embed" {
					main = "package main\nimport (\"embed\"; \"fmt\")\n//go:embed assets/*\nvar assets embed.FS\nfunc main() { b, _ := assets.ReadFile(\"assets/extra.txt\"); fmt.Print(string(b)) }\n"
					writeFile(t, filepath.Join(module, "assets/base.txt"), "pinned")
					path, ignored = "modeling/fake/assets/extra.txt", "ignored"
				}
				writeFile(t, filepath.Join(module, "main.go"), main)
				rule := "/" + path + "\n"
				ignorePath := filepath.Join(dir, ".gitignore")
				switch ignoreSource {
				case "local":
					ignorePath = filepath.Join(dir, ".git/info/exclude")
				case "configured":
					ignorePath = filepath.Join(t.TempDir(), "ignore")
					pinGit(t, dir, "config", "core.excludesFile", ignorePath)
				}
				writeFile(t, ignorePath, rule)
				pinGit(t, dir, "add", "--", "modeling")
				if ignoreSource == "tracked" {
					pinGit(t, dir, "add", "--", ".gitignore")
				}
				pinGit(t, dir, "-c", "user.name=fixture", "-c", "user.email=fixture@example.invalid", "-c", "commit.gpgsign=false", "commit", "-qm", "Add generator fixture")
				head := strings.TrimSpace(pinGit(t, dir, "rev-parse", "HEAD"))
				s := &session{foundry: dir, proj: &Project{Revision: head}, stderr: &bytes.Buffer{}}
				if err := s.pin(); err != nil {
					t.Fatalf("clean pinned source refused: %v", err)
				}
				writeFile(t, filepath.Join(dir, path), ignored)
				if got := pinGit(t, dir, "status", "--porcelain", "--untracked-files=all", "--", "modeling"); got != "" {
					t.Fatalf("fixture must hide the local input from ordinary status: %q", got)
				}
				// Demonstrate that an ignored file is a real compiler input, not
				// merely unrelated clutter hidden by Git.
				s.proj.Outputs = []Output{{Generator: "fake"}}
				s.reg = Registry{"fake": {ID: "fake", Module: "modeling/fake", Package: "."}}
				s.binDir = t.TempDir()
				if err := s.build(); err != nil {
					t.Fatalf("build fixture: %v: %s", err, s.stderr)
				}
				got, err := runCmd(childTimeout, dir, []string{}, filepath.Join(s.binDir, "fake"))
				if err != nil || got.code != 0 || got.stdout != "ignored" {
					t.Fatalf("ignored input did not affect build: %+v, %v", got, err)
				}
				indexBefore, err := os.ReadFile(filepath.Join(dir, ".git/index"))
				if err != nil {
					t.Fatal(err)
				}
				if err := s.pin(); err == nil || codeOf(err) != exitRefused || !strings.Contains(err.Error(), "ignored") {
					t.Fatalf("ignored build input accepted: %v", err)
				}
				indexAfter, err := os.ReadFile(filepath.Join(dir, ".git/index"))
				if err != nil {
					t.Fatal(err)
				}
				if !bytes.Equal(indexBefore, indexAfter) || readFile(t, filepath.Join(dir, path)) != ignored || readFile(t, ignorePath) != rule || readFile(t, filepath.Join(module, "main.go")) != main {
					t.Fatal("refusal changed the index, source or ignore rules")
				}
			})
		}
	}
}

func TestPinAllowsIgnoredContentOutsideModeling(t *testing.T) {
	dir, head := gitFixture(t)
	writeFile(t, filepath.Join(dir, ".git/info/exclude"), "/outside/\n")
	writeFile(t, filepath.Join(dir, "outside/local.txt"), "local")
	if err := (&session{foundry: dir, proj: &Project{Revision: head}}).pin(); err != nil {
		t.Fatalf("ignored content outside modeling refused: %v", err)
	}
}

func TestPinRefusesIgnoredModelingDirectory(t *testing.T) {
	dir, head := gitFixture(t)
	writeFile(t, filepath.Join(dir, ".git/info/exclude"), "/modeling/local/\n")
	writeFile(t, filepath.Join(dir, "modeling/local/data.txt"), "local")
	if err := (&session{foundry: dir, proj: &Project{Revision: head}}).pin(); err == nil || codeOf(err) != exitRefused {
		t.Fatalf("ignored modeling directory accepted: %v", err)
	}
}
