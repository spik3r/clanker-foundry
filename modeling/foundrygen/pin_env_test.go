package foundrygen

import (
	"bytes"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// cloneFixture makes a second checkout at the fixture's committed state.
func cloneFixture(t *testing.T, src string) string {
	t.Helper()
	dst := t.TempDir()
	cmd := exec.Command("git", "clone", "-q", "--no-hardlinks", src, dst)
	if b, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("clone fixture: %v\n%s", err, b)
	}
	return dst
}

// Git's repository-location variables override -C. An inherited GIT_DIR and
// GIT_WORK_TREE must not let a dirty selected checkout pass the pin.
func TestPinIgnoresInheritedGitLocation(t *testing.T) {
	selected, head := gitFixture(t)
	clean := cloneFixture(t, selected)
	writeFile(t, filepath.Join(selected, "modeling/extra.go"), "package extra")

	t.Setenv("GIT_DIR", filepath.Join(clean, ".git"))
	t.Setenv("GIT_WORK_TREE", clean)

	s := &session{foundry: selected, proj: &Project{Revision: head}, stderr: &bytes.Buffer{}}
	err := s.pin()
	if err == nil || codeOf(err) != exitRefused || !strings.Contains(err.Error(), "uncommitted") {
		t.Fatalf("pin followed inherited Git location: %v", err)
	}
}

func TestGitEnvDropsRepositoryLocation(t *testing.T) {
	env := []string{
		"PATH=/bin",
		"GIT_DIR=/elsewhere",
		"git_work_tree=/elsewhere",
		"GIT_INDEX_FILE=/elsewhere/index",
		"GIT_OBJECT_DIRECTORY=/elsewhere/objects",
		"GIT_ALTERNATE_OBJECT_DIRECTORIES=/elsewhere/objects",
		"GIT_COMMON_DIR=/elsewhere",
		"GIT_NAMESPACE=elsewhere",
		"GIT_PREFIX=sub/",
		"GIT_SUPER_PREFIX=super/",
		"GIT_AUTHOR_NAME=kept",
	}
	want := []string{"PATH=/bin", "GIT_AUTHOR_NAME=kept"}
	if got := gitEnv(env); strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Fatalf("gitEnv = %q, want %q", got, want)
	}
}
