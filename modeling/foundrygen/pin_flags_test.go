package foundrygen

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func pinGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	b, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, b)
	}
	return string(b)
}

func TestPinRefusesHiddenTrackedChanges(t *testing.T) {
	for _, flag := range []string{"--assume-unchanged", "--skip-worktree"} {
		for _, changed := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/changed=%t", flag, changed), func(t *testing.T) {
				dir, head := gitFixture(t)
				path := "modeling/gencontract/generators.json"
				pinGit(t, dir, "update-index", flag, "--", path)
				if changed {
					p := filepath.Join(dir, path)
					writeFile(t, p, strings.Replace(readFile(t, p), "p/v1", "changed/v1", 1))
				}
				beforeSource := readFile(t, filepath.Join(dir, path))
				beforeFlags := pinGit(t, dir, "ls-files", "-v", "-z")
				beforeIndex, err := os.ReadFile(filepath.Join(dir, ".git/index"))
				if err != nil {
					t.Fatal(err)
				}
				s := &session{foundry: dir, proj: &Project{Revision: head}, stderr: &bytes.Buffer{}}
				if err := s.pin(); err == nil || codeOf(err) != exitRefused || !strings.Contains(err.Error(), "index flag") {
					t.Errorf("hidden index state accepted: %v", err)
				}
				afterIndex, err := os.ReadFile(filepath.Join(dir, ".git/index"))
				if err != nil {
					t.Fatal(err)
				}
				if !bytes.Equal(beforeIndex, afterIndex) || beforeFlags != pinGit(t, dir, "ls-files", "-v", "-z") || beforeSource != readFile(t, filepath.Join(dir, path)) {
					t.Fatal("pin verification changed the index, flags or source bytes")
				}
			})
		}
	}
}

func TestPinAllowsIndexFlagsOutsideModeling(t *testing.T) {
	for _, flag := range []string{"--assume-unchanged", "--skip-worktree"} {
		t.Run(flag, func(t *testing.T) {
			dir, head := gitFixture(t)
			writeFile(t, filepath.Join(dir, "notes.txt"), "reviewed outside scope")
			pinGit(t, dir, "add", "--", "notes.txt")
			pinGit(t, dir, "update-index", flag, "--", "notes.txt")
			writeFile(t, filepath.Join(dir, "notes.txt"), "local outside scope")
			if err := (&session{foundry: dir, proj: &Project{Revision: head}}).pin(); err != nil {
				t.Fatalf("outside-modeling flag refused: %v", err)
			}
		})
	}
}

func TestTrackedVisibilityMetadata(t *testing.T) {
	for _, list := range []string{"", "H modeling/source.go\x00", "H modeling/line\nhidden.go\x00", "M modeling/conflict.go\x00"} {
		if err := checkTrackedVisibility(list); err != nil {
			t.Errorf("complete ordinary metadata refused: %v", err)
		}
	}
	for _, tag := range []string{"h", "S", "s"} {
		err := checkTrackedVisibility(tag + " modeling/source.go\x00")
		if err == nil || codeOf(err) != exitRefused {
			t.Errorf("hidden tag %q accepted: %v", tag, err)
		}
	}
	for _, list := range []string{"H modeling/source.go", "H \x00", "H\x00", "X modeling/source.go\x00", strings.Repeat("H a\x00", maxCapture/4)} {
		err := checkTrackedVisibility(list)
		if err == nil || codeOf(err) != exitFailure {
			t.Errorf("incomplete metadata of %d bytes accepted: %v", len(list), err)
		}
	}
	belowLimit := strings.Repeat("H a\x00", maxCapture/4-1)
	if err := checkTrackedVisibility(belowLimit); err != nil {
		t.Fatalf("complete metadata below capture limit refused: %v", err)
	}
}
