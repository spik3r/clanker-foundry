package builderdemo_test

// These corrective regressions supplement the unchanged original acceptance.
import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/builderdemo"
)

func boundaryFixture(t *testing.T) string {
	t.Helper()
	root := filepath.Join(t.TempDir(), "SourceRoot")
	if err := os.Rename(fixture(t), root); err != nil {
		t.Fatal(err)
	}
	return root
}

func boundaryTree(t *testing.T, root string) map[string]string {
	t.Helper()
	result := map[string]string{}
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		if entry.IsDir() {
			result[rel] = "directory"
		} else {
			result[rel] = digest(read(t, path))
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func TestCaseAliasesCannotPlaceWorkspaceInsideSource(t *testing.T) {
	cases := []struct {
		name, child              string
		sourceAlias, parentAlias bool
	}{
		{"root-alias", "", false, true},
		{"descendant-alias", "Output", false, true},
		{"deep-descendant-alias", filepath.Join("Output", "Nested"), false, true},
		{"source-alias", "", true, false},
		{"source-alias-descendant", "Output", true, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			root := boundaryFixture(t)
			alias := filepath.Join(filepath.Dir(root), "sourceroot")
			aliasInfo, err := os.Stat(alias)
			if os.IsNotExist(err) {
				t.Skip("fixture filesystem distinguishes the two case spellings")
			}
			rootInfo, rootErr := os.Stat(root)
			if err != nil || rootErr != nil || !os.SameFile(rootInfo, aliasInfo) {
				t.Fatalf("case-alias fixture is not the same directory: %v %v", err, rootErr)
			}
			if err := os.MkdirAll(filepath.Join(root, "Output", "Nested"), 0700); err != nil {
				t.Fatal(err)
			}
			source, parent := root, root
			if c.sourceAlias {
				source = alias
			}
			if c.parentAlias {
				parent = alias
			}
			parent = filepath.Join(parent, c.child)
			before := boundaryTree(t, root)
			r, err := builderdemo.Prepare(source, parent, "/record-only-go")
			if !errors.Is(err, builderdemo.ErrInvalid) || !reflect.DeepEqual(r, builderdemo.Receipt{}) {
				t.Errorf("overlapping case alias: error=%v outcome=%q workspace=%v", err, r.Outcome, r.Workspace)
			}
			if after := boundaryTree(t, root); !reflect.DeepEqual(before, after) {
				t.Errorf("invalid location changed the isolated source tree: before=%d entries after=%d", len(before), len(after))
			}
		})
	}
}

func TestDistinctCaseSensitiveSiblingRemainsOutsideSource(t *testing.T) {
	root := boundaryFixture(t)
	parent := filepath.Join(filepath.Dir(root), "sourceroot")
	if err := os.Mkdir(parent, 0700); err != nil {
		rootInfo, rootErr := os.Stat(root)
		parentInfo, parentErr := os.Stat(parent)
		if os.IsExist(err) && rootErr == nil && parentErr == nil && os.SameFile(rootInfo, parentInfo) {
			t.Skip("fixture filesystem aliases the two case spellings")
		}
		t.Fatal(err)
	}
	before := boundaryTree(t, root)
	r, err := builderdemo.Prepare(root, parent, "/record-only-go")
	if err != nil || r.Outcome != "prepared-draft" || r.Workspace == nil {
		t.Fatalf("distinct outside directory was rejected: %v %q", err, r.Outcome)
	}
	want, err := os.Stat(parent)
	if err != nil {
		t.Fatal(err)
	}
	got, err := os.Stat(filepath.Dir(*r.Workspace))
	if err != nil || !os.SameFile(want, got) {
		t.Fatalf("workspace was not created under the distinct parent: %v", err)
	}
	if !reflect.DeepEqual(before, boundaryTree(t, root)) {
		t.Fatal("preparation changed source bytes or paths")
	}
}
