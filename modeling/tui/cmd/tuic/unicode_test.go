package main

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func unicodeTree(t *testing.T, root string) map[string]string {
	t.Helper()
	files := map[string]string{}
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			b, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			files[path] = string(b)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return files
}

func TestUnicodeRefusedBeforeCLIOutput(t *testing.T) {
	raw, err := os.ReadFile(sample)
	if err != nil {
		t.Fatal(err)
	}
	for _, text := range []string{`\ud800`, `\udfff`, `\udc00\ud800`, `\ud800\u0041`} {
		t.Run(text, func(t *testing.T) {
			root := t.TempDir()
			model := filepath.Join(root, "bad.json")
			// Prefix a valid title so only the escape makes this input invalid.
			bad := bytes.Replace(raw, []byte(`"title": "`), []byte(`"title": "`+text), 1)
			if bytes.Equal(bad, raw) {
				t.Fatal("title replacement did not match")
			}
			if err := os.WriteFile(model, bad, 0o600); err != nil {
				t.Fatal(err)
			}
			existing := filepath.Join(root, "existing")
			if code, _, e := do("generate", sample, existing); code != 0 {
				t.Fatal(e)
			}
			before := unicodeTree(t, existing)
			absent := filepath.Join(root, "absent")
			for _, args := range [][]string{{"validate", model}, {"generate", model, absent}, {"check", model, absent}, {"generate", model, existing}, {"check", model, existing}} {
				code, out, stderr := do(args...)
				if code != 2 || out != "" || stderr != "tuic: unpaired surrogate escape in JSON\n" {
					t.Errorf("%v: %d %q %q", args, code, out, stderr)
				}
				if _, err := os.Lstat(absent); !os.IsNotExist(err) {
					t.Error("refused input created output")
				}
				if !reflect.DeepEqual(before, unicodeTree(t, existing)) {
					t.Error("refused input modified output")
				}
			}
		})
	}
}
