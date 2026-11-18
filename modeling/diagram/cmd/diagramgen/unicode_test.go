package main

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func unicodeSnapshot(t *testing.T, root string) map[string]string {
	t.Helper()
	m := map[string]string{}
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		if d.IsDir() {
			m[rel] = "directory"
			return nil
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		m[rel] = string(b)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return m
}

func TestUnicodeCLIRefusals(t *testing.T) {
	for _, name := range []string{"order", "minimal-dashboard", "checkout-sequence", "orders-class"} {
		t.Run(name, func(t *testing.T) {
			raw, err := os.ReadFile("../../examples/" + name + ".json")
			if err != nil {
				t.Fatal(err)
			}
			for _, text := range []string{`\ud800`, `\udfff`, `\udc00\ud800`, `\ud800\u0041`} {
				root := t.TempDir()
				model := filepath.Join(root, "bad.json")
				bad := bytes.Replace(raw, []byte(`"title": "`), []byte(`"title": "`+text), 1)
				if bytes.Equal(raw, bad) {
					t.Fatal("replacement failed")
				}
				if err := os.WriteFile(model, bad, 0o600); err != nil {
					t.Fatal(err)
				}
				existing := filepath.Join(root, "existing")
				if code, _, err := do("generate", filepath.Join("../../examples", name+".json"), existing); code != 0 {
					t.Fatal(err)
				}
				before := unicodeSnapshot(t, existing)
				absent := filepath.Join(root, "absent")
				for _, args := range [][]string{{"validate", model}, {"generate", model, absent}, {"check", model, absent}, {"generate", model, existing}, {"check", model, existing}} {
					code, out, err := do(args...)
					if code != 2 || out != "" || !strings.Contains(err, "unpaired surrogate escape in JSON") {
						t.Errorf("%v: %d %q %q", args, code, out, err)
					}
					if _, err := os.Lstat(absent); !os.IsNotExist(err) {
						t.Error("refused input created output")
					}
					if !reflect.DeepEqual(before, unicodeSnapshot(t, existing)) {
						t.Error("refused input modified output")
					}
				}
			}
		})
	}
}

func TestUnicodeCLIBindAndTesseraeRefusals(t *testing.T) {
	template, err := os.ReadFile("../../examples/ops-dashboard.json")
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile("../../examples/ops-report-data.json")
	if err != nil {
		t.Fatal(err)
	}
	for _, text := range []string{`\ud800`, `\udfff`, `\udc00\ud800`, `\ud800\u0041`} {
		root := t.TempDir()
		tp := filepath.Join(root, "template.json")
		dp := filepath.Join(root, "data.json")
		write := func(p string, b []byte) {
			t.Helper()
			if err := os.WriteFile(p, b, 0o600); err != nil {
				t.Fatal(err)
			}
		}
		write(tp, template)
		write(dp, bytes.Replace(data, []byte(`"Mon"`), []byte(`"`+text+`"`), 1))
		before := unicodeSnapshot(t, root)
		code, out, err := do("bind", tp, dp)
		if code != 2 || out != "" || !strings.Contains(err, "unpaired surrogate escape in JSON") {
			t.Errorf("bind data: %d %q %q", code, out, err)
		}
		if !reflect.DeepEqual(before, unicodeSnapshot(t, root)) {
			t.Error("bind modified inputs")
		}
		bad := bytes.Replace(template, []byte(`"title": "`), []byte(`"title": "`+text), 1)
		write(tp, bad)
		write(dp, data)
		before = unicodeSnapshot(t, root)
		for _, args := range [][]string{{"bind", tp, dp}, {"tesserae", tp}} {
			code, out, err := do(args...)
			if code != 2 || out != "" || !strings.Contains(err, "unpaired surrogate escape in JSON") {
				t.Errorf("%v: %d %q %q", args, code, out, err)
			}
			if !reflect.DeepEqual(before, unicodeSnapshot(t, root)) {
				t.Error("transform modified inputs")
			}
		}
	}
}
