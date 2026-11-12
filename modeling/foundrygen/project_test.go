package foundrygen

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var testReg = Registry{
	"tuic":          {ID: "tuic"},
	"jsonvalidator": {ID: "jsonvalidator"},
}

const rev = "0123456789abcdef0123456789abcdef01234567"

func projectJSON(outputs string) string {
	return fmt.Sprintf(`{"projectVersion":"foundry-project/v1","foundry":{"revision":"%s"},"outputs":[%s]}`, rev, outputs)
}

func out(gen, model, o string) string {
	return fmt.Sprintf(`{"generator":%q,"model":%q,"out":%q}`, gen, model, o)
}

func TestParseProjectAccepts(t *testing.T) {
	p, err := ParseProject([]byte(projectJSON(out("tuic", "design/b.json", "internal/b")+","+out("jsonvalidator", "design/b.json", "internal/c"))), testReg)
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Outputs) != 2 || p.Outputs[0].Out != "internal/b" || p.Revision != rev {
		t.Fatalf("unexpected project %+v", p)
	}
}

func TestParseProjectRefusals(t *testing.T) {
	ok := out("tuic", "m.json", "o")
	many := make([]string, 65)
	for i := range many {
		many[i] = out("tuic", "m.json", fmt.Sprintf("o%d", i))
	}
	cases := map[string]string{
		"empty":               ``,
		"not object":          `[]`,
		"trailing data":       projectJSON(ok) + ` {}`,
		"trailing garbage":    projectJSON(ok) + `x`,
		"unknown top key":     strings.Replace(projectJSON(ok), `"outputs"`, `"extra":1,"outputs"`, 1),
		"unknown output key":  projectJSON(`{"generator":"tuic","model":"m","out":"o","x":1}`),
		"case variant key":    projectJSON(`{"Generator":"tuic","model":"m","out":"o"}`),
		"duplicate top key":   strings.Replace(projectJSON(ok), `"outputs"`, `"projectVersion":"foundry-project/v1","outputs"`, 1),
		"duplicate nested":    projectJSON(`{"generator":"tuic","generator":"tuic","model":"m","out":"o"}`),
		"missing key":         projectJSON(`{"generator":"tuic","model":"m"}`),
		"wrong version":       strings.Replace(projectJSON(ok), "foundry-project/v1", "foundry-project/v2", 1),
		"version type":        strings.Replace(projectJSON(ok), `"foundry-project/v1"`, `1`, 1),
		"short revision":      strings.Replace(projectJSON(ok), rev, rev[:39], 1),
		"upper revision":      strings.Replace(projectJSON(ok), rev, strings.ToUpper(rev), 1),
		"revision type":       strings.Replace(projectJSON(ok), `"`+rev+`"`, `null`, 1),
		"outputs not array":   strings.Replace(projectJSON(ok), `"outputs":[`+ok+`]`, `"outputs":{}`, 1),
		"no outputs":          projectJSON(``),
		"outputs null":        strings.Replace(projectJSON(ok), `"outputs":[`+ok+`]`, `"outputs":null`, 1),
		"too many outputs":    projectJSON(strings.Join(many, ",")),
		"output not object":   projectJSON(`"x"`),
		"field type":          projectJSON(`{"generator":"tuic","model":3,"out":"o"}`),
		"field null":          projectJSON(`{"generator":"tuic","model":null,"out":"o"}`),
		"unknown generator":   projectJSON(out("nope", "m", "o")),
		"empty model":         projectJSON(out("tuic", "", "o")),
		"empty out":           projectJSON(out("tuic", "m", "")),
		"absolute model":      projectJSON(out("tuic", "/etc/m", "o")),
		"absolute out":        projectJSON(out("tuic", "m", "/tmp/o")),
		"dot out":             projectJSON(out("tuic", "m", ".")),
		"dotdot model":        projectJSON(out("tuic", "../m", "o")),
		"dotdot middle":       projectJSON(out("tuic", "a/../m", "o")),
		"dot segment":         projectJSON(out("tuic", "a/./m", "o")),
		"empty segment":       projectJSON(out("tuic", "a//m", "o")),
		"trailing slash":      projectJSON(out("tuic", "m", "o/")),
		"backslash":           projectJSON(out("tuic", `a\m`, "o")),
		"duplicate out":       projectJSON(out("tuic", "m", "o") + "," + out("tuic", "n", "o")),
		"nested out":          projectJSON(out("tuic", "m", "a") + "," + out("tuic", "n", "a/b")),
		"nested out reversed": projectJSON(out("tuic", "m", "a/b") + "," + out("tuic", "n", "a")),
		"out equals model":    projectJSON(out("tuic", "m", "m")),
		"out contains model":  projectJSON(out("tuic", "d/m.json", "d")),
		"out holds other mdl": projectJSON(out("tuic", "m", "o") + "," + out("tuic", "o/n.json", "p")),
		"out is project file": projectJSON(out("tuic", "m", "foundry.json")),
	}
	for name, src := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := ParseProject([]byte(src), testReg)
			if err == nil {
				t.Fatal("accepted")
			}
			if codeOf(err) != exitRefused {
				t.Fatalf("code %d, want 2: %v", codeOf(err), err)
			}
		})
	}
}

func TestParseProjectTooLarge(t *testing.T) {
	pad := strings.Repeat(" ", maxProjectBytes)
	if _, err := ParseProject([]byte(projectJSON(out("tuic", "m", "o"))+pad), testReg); err == nil {
		t.Fatal("accepted a file over 64 KiB")
	}
}

func TestResolveRefusesSymlinks(t *testing.T) {
	root := t.TempDir()
	real := filepath.Join(root, "real")
	if err := os.Mkdir(real, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(real, filepath.Join(root, "link")); err != nil {
		t.Skip("symlinks unavailable:", err)
	}
	if err := os.WriteFile(filepath.Join(root, "file"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	for _, rel := range []string{"link", "link/x", "link/deeper/x"} {
		if _, err := resolveInProject(root, rel); err == nil || codeOf(err) != exitRefused {
			t.Errorf("%s: want refusal, got %v", rel, err)
		}
	}
	if _, err := resolveInProject(root, "file/x"); err == nil || codeOf(err) != exitRefused {
		t.Errorf("file/x: want refusal, got %v", err)
	}
	for _, rel := range []string{"real", "real/x", "missing/a/b", "file"} {
		if _, err := resolveInProject(root, rel); err != nil {
			t.Errorf("%s: %v", rel, err)
		}
	}
}
