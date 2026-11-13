package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestTesseraeCLI(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "model.json")
	raw := `{"formatVersion":"diagram-json/v1","profileVersion":"dashboard/v1","id":"Demo","title":"Demo","datasets":[],"widgets":[{"id":"Count","kind":"stat","label":"Count","value":"0"}],"layout":[["Count"]]}`
	if err := os.WriteFile(path, []byte(raw), 0600); err != nil {
		t.Fatal(err)
	}
	var out, stderr bytes.Buffer
	if code := run([]string{"tesserae", path}, &out, &stderr); code != 0 {
		t.Fatalf("%d: %s", code, &stderr)
	}
	if !bytes.Contains(out.Bytes(), []byte(`"value":0`)) {
		t.Fatal(out.String())
	}
	entries, _ := os.ReadDir(dir)
	if len(entries) != 1 {
		t.Fatal("export wrote files")
	}
	if err := os.WriteFile(path, []byte(`{}`), 0600); err != nil {
		t.Fatal(err)
	}
	out.Reset()
	stderr.Reset()
	if code := run([]string{"tesserae", path}, &out, &stderr); code != exitRefused || out.Len() != 0 {
		t.Fatalf("%d %s %s", code, &out, &stderr)
	}
}
