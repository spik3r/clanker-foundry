package builderpreflight_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// The root is a host path, not an artifact path. In particular, lexical cleanup
// must not change the directory selected by an ancestor symlink followed by a parent component.
func TestRootAncestorSymlinkPreservesHostResolution(t *testing.T) {
	base := t.TempDir()
	lexical := filepath.Join(base, "lex")
	actual := filepath.Join(base, "actual")
	writeFixture(t, lexical, "chosen/input", []byte("abc"))
	writeFixture(t, actual, "chosen/input", []byte("actual root bytes"))
	if err := os.Mkdir(filepath.Join(actual, "child"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(actual, "child"), filepath.Join(lexical, "alias")); err != nil {
		t.Skipf("host does not permit symlinks: %v", err)
	}
	// Concatenate deliberately: filepath.Join would erase the path under test.
	root := lexical + string(filepath.Separator) + "alias" + string(filepath.Separator) + ".." + string(filepath.Separator) + "chosen"
	actualSHA := digest([]byte("actual root bytes"))
	for _, tc := range []struct {
		name, pin, outcome, status string
	}{
		{"lexical-pin-must-not-match", abcSHA, "blocked", "hash-mismatch"},
		{"actual-pin-must-match", actualSHA, "matched", "matched"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			raw := manifest(item("source", "requirements", "input", tc.pin))
			_, report := check(t, raw, root)
			items := assertEnvelope(t, report, digest(raw), tc.outcome, 1)
			assertArtifact(t, items[0], "source", "requirements", "input", tc.pin, tc.status, actualSHA, json.Number("17"))
		})
	}
}

func TestRootDotDotDoesNotEraseInvalidComponents(t *testing.T) {
	base := t.TempDir()
	writeFixture(t, base, "chosen/input", []byte("abc"))
	writeFixture(t, base, "regularfile", []byte("not a directory"))
	raw := manifest(item("source", "requirements", "input", abcSHA))
	for _, component := range []string{"missing", "regularfile"} {
		t.Run(component, func(t *testing.T) {
			root := base + string(filepath.Separator) + component + string(filepath.Separator) + ".." + string(filepath.Separator) + "chosen"
			assertInvalid(t, raw, root, digest(raw))
		})
	}
}

func TestRootTrailingSeparatorsAndAncestorAliases(t *testing.T) {
	base := t.TempDir()
	writeFixture(t, base, "real/chosen/input", []byte("abc"))
	raw := manifest(item("source", "requirements", "input", abcSHA))
	real := filepath.Join(base, "real", "chosen")
	separator := string(filepath.Separator)
	t.Run("real-directory-trailing-separator", func(t *testing.T) {
		_, report := check(t, raw, real+separator)
		items := assertEnvelope(t, report, digest(raw), "matched", 1)
		assertArtifact(t, items[0], "source", "requirements", "input", abcSHA, "matched", abcSHA, json.Number("3"))
	})
	alias := filepath.Join(base, "alias")
	if err := os.Symlink(filepath.Join(base, "real"), alias); err != nil {
		t.Skipf("host does not permit symlinks: %v", err)
	}
	t.Run("ancestor-symlink-without-dot-dot", func(t *testing.T) {
		_, report := check(t, raw, alias+separator+"chosen")
		items := assertEnvelope(t, report, digest(raw), "matched", 1)
		assertArtifact(t, items[0], "source", "requirements", "input", abcSHA, "matched", abcSHA, json.Number("3"))
	})
	leaf := filepath.Join(base, "leaf")
	if err := os.Symlink(real, leaf); err != nil {
		t.Fatal(err)
	}
	for _, suffix := range []string{"", separator, separator + separator} {
		t.Run("symlink-leaf-"+suffix, func(t *testing.T) {
			assertInvalid(t, raw, leaf+suffix, digest(raw))
		})
	}
}
