package foundrygen

import (
	"bytes"
	"path/filepath"
	"testing"
)

func TestBuildIgnoresAmbientWorkspace(t *testing.T) {
	for _, mode := range []string{"auto", "explicit"} {
		t.Run(mode, func(t *testing.T) {
			root, err := filepath.EvalSymlinks(t.TempDir())
			if err != nil {
				t.Fatal(err)
			}
			writeFile(t, filepath.Join(root, "modeling/generator/go.mod"), `module example.test/generator

go 1.27.0

require example.test/value v0.0.0
replace example.test/value => ../value
`)
			writeFile(t, filepath.Join(root, "modeling/generator/cmd/generator/main.go"), `package main
import (
 "fmt"
 "example.test/value"
)
func main() { fmt.Println(value.Text) }
`)
			for _, dep := range []struct{ dir, text string }{
				{"modeling/value", "pinned module"},
				{"workspace-value", "ambient workspace"},
			} {
				writeFile(t, filepath.Join(root, dep.dir, "go.mod"), "module example.test/value\n\ngo 1.27.0\n")
				writeFile(t, filepath.Join(root, dep.dir, "value.go"), "package value\nconst Text = \""+dep.text+"\"\n")
			}
			workspace := filepath.Join(root, "go.work")
			writeFile(t, workspace, `go 1.27.0
use ./modeling/generator
replace example.test/value => ./workspace-value
`)
			if mode == "auto" {
				t.Setenv("GOWORK", "auto")
			} else {
				t.Setenv("GOWORK", workspace)
			}
			s := &session{
				foundry: root,
				proj:    &Project{Outputs: []Output{{Generator: "fixture"}}},
				reg: Registry{"fixture": {
					ID: "fixture", Module: "modeling/generator", Package: "./cmd/generator",
				}},
				binDir: t.TempDir(), stderr: &bytes.Buffer{},
			}
			if err := s.build(); err != nil {
				t.Fatalf("build: %v\n%s", err, s.stderr)
			}
			got, err := runCmd(childTimeout, root, []string{}, filepath.Join(s.binDir, "fixture"))
			if err != nil || got.code != 0 || got.stdout != "pinned module\n" {
				t.Fatalf("built generator used workspace dependency: result %+v, error %v", got, err)
			}
		})
	}
}
