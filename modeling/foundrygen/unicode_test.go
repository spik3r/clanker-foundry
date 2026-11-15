package foundrygen

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

var unicodeReg = Registry{"refgen": {ID: "refgen", Module: "modeling/gencontract", Package: "./cmd/refgen"}}

func unicodeProject(model, output string) []byte {
	return []byte(projectJSON(`{"generator":"refgen","model":` + model + `,"out":` + output + `}`))
}
func unicodeRegistry(module, pkg, extra string) []byte {
	return []byte(`{"registryVersion":"foundry-generator-registry/v1","generators":[{"id":"refgen","module":` + module + `,"package":` + pkg + `}]` + extra + `}`)
}
func unicodeRegistryFile(t *testing.T, root string, raw []byte) {
	t.Helper()
	writeFile(t, filepath.Join(root, "modeling", "gencontract", "generators.json"), string(raw))
}

func TestUnicodeProjectIntake(t *testing.T) {
	invalid := []string{`"design/\ud800.json"`, `"design/\udbff.json"`, `"design/\udc00.json"`, `"design/\udfff.json"`, `"design/\udc00\ud800.json"`, `"design/\ud800\ud800.json"`, `"design/\ud800\u0041.json"`, `"design/\ud800\\udc00.json"`}
	for _, bad := range []string{string([]byte{0xff}), string([]byte{0x80}), string([]byte{0xc0, 0xaf}), string([]byte{0xed, 0xa0, 0x80})} {
		invalid = append(invalid, `"design/`+bad+`.json"`)
	}
	for i, value := range invalid {
		for _, field := range []string{"model", "out"} {
			t.Run(fmt.Sprintf("invalid-%d-%s", i, field), func(t *testing.T) {
				model, output := `"design/model.json"`, `"generated"`
				if field == "model" {
					model = value
				} else {
					output = value
				}
				if _, err := ParseProject(unicodeProject(model, output), unicodeReg); err == nil || codeOf(err) != exitRefused {
					t.Fatalf("malformed Unicode admitted: %v", err)
				}
			})
		}
	}
	for _, tc := range []struct{ wire, want string }{
		{`"design/�.json"`, "design/�.json"}, {`"design/\ufffd.json"`, "design/�.json"},
		{`"design/🙂.json"`, "design/🙂.json"}, {`"design/\ud83d\ude42.json"`, "design/🙂.json"},
		{`"design/\uD800\uDC00.json"`, "design/𐀀.json"}, {`"design/\uDBFF\uDFFF.json"`, "design/\U0010ffff.json"},
		{`"design/e\u0301.json"`, "design/e\u0301.json"}, {`"design/quote\".json"`, "design/quote\".json"},
	} {
		p, err := ParseProject(unicodeProject(tc.wire, `"generated"`), unicodeReg)
		if err != nil || p.Outputs[0].Model != tc.want {
			t.Fatalf("valid Unicode changed: %q %+v %v", tc.wire, p, err)
		}
	}
	if _, err := ParseProject(unicodeProject(`"design/\\ud800.json"`, `"generated"`), unicodeReg); err == nil {
		t.Fatal("path backslash admitted")
	}
	good := string(unicodeProject(`"design/model.json"`, `"generated"`))
	for _, raw := range []string{
		strings.Replace(good, `"outputs":`, `"\u006futputs":[],"outputs":`, 1),
		strings.Replace(good, `"model":`, `"unknown":0,"model":`, 1),
		strings.Replace(good, `"model":"design/model.json"`, `"model":null`, 1),
		strings.Replace(good, `"model":"design/model.json"`, `"model":4`, 1),
		strings.Replace(good, `"model":`, `"\ud800":`, 1),
	} {
		if _, err := ParseProject([]byte(raw), unicodeReg); err == nil {
			t.Fatal("existing project schema refusal lost")
		}
	}
	escaped := strings.Replace(good, `"projectVersion"`, `"\u0070rojectVersion"`, 1)
	if _, err := ParseProject([]byte(escaped), unicodeReg); err != nil {
		t.Fatal("valid key escape rejected", err)
	}
	// Strict prefixes exercise truncated JSON and escapes without a new parser oracle.
	full := unicodeProject(`"design/\ud83d\ude42.json"`, `"generated"`)
	for n := 0; n < len(full); n++ {
		if _, err := ParseProject(full[:n], unicodeReg); err == nil {
			t.Fatalf("truncated JSON accepted at %d", n)
		}
	}
	for _, raw := range []string{good + ` {}`, strings.Replace(good, "model.json", `\u12ZZ`, 1), strings.Replace(good, "model.json", `\u123`, 1)} {
		if _, err := ParseProject([]byte(raw), unicodeReg); err == nil {
			t.Fatal("malformed syntax accepted")
		}
	}
}

func TestUnicodeRegistryIntake(t *testing.T) {
	invalid := []string{`"modeling/\ud800"`, `"modeling/\udc00"`, `"modeling/\ud800\udc00\udfff"`, `"modeling/\ud800\\udc00"`, `"modeling/` + string([]byte{0xff}) + `"`}
	for i, value := range invalid {
		for _, field := range []string{"module", "package", "ignored value", "ignored key"} {
			t.Run(fmt.Sprintf("invalid-%d-%s", i, field), func(t *testing.T) {
				module, pkg, extra := `"modeling/gencontract"`, `"./cmd/refgen"`, ""
				switch field {
				case "module":
					module = value
				case "package":
					pkg = strings.Replace(value, `"modeling/`, `"./cmd/`, 1)
				case "ignored value":
					extra = `,"note":` + value
				case "ignored key":
					extra = `,` + value + `:"note"`
				}
				root := t.TempDir()
				unicodeRegistryFile(t, root, unicodeRegistry(module, pkg, extra))
				if _, err := LoadRegistry(root); err == nil || codeOf(err) != exitRefused {
					t.Fatalf("malformed registry Unicode admitted: %v", err)
				}
			})
		}
	}
	for _, module := range []string{`"modeling/�"`, `"modeling/\ufffd"`, `"modeling/🙂"`, `"modeling/\ud83d\ude42"`, `"modeling/e\u0301"`, `"modeling/quote\""`} {
		root := t.TempDir()
		unicodeRegistryFile(t, root, unicodeRegistry(module, `"./cmd/refgen"`, `,"literal":"\\ud800","�":"\ufffd"`))
		if _, err := LoadRegistry(root); err != nil {
			t.Fatalf("valid registry Unicode refused: %s %v", module, err)
		}
	}
	// Unicode preflight must not import strict shape/duplicate registry policy.
	root := t.TempDir()
	raw := `{"registryVersion":"old","registryVersion":"foundry-generator-registry/v1","unknown":"valid","generators":[{"id":"refgen","module":"old","module":"modeling/gencontract","package":"./cmd/refgen","unknown":"\\ud800"}]}`
	unicodeRegistryFile(t, root, []byte(raw))
	reg, err := LoadRegistry(root)
	if err != nil || reg["refgen"].Module != "modeling/gencontract" {
		t.Fatalf("registry valid unknown/duplicate behavior changed: %+v %v", reg, err)
	}
}

func TestUnicodeConfigurationSizeBoundaries(t *testing.T) {
	good := unicodeProject(`"design/model.json"`, `"generated"`)
	exact := append(append([]byte(nil), good...), bytes.Repeat([]byte(" "), maxProjectBytes-len(good))...)
	if _, err := ParseProject(exact, unicodeReg); err != nil {
		t.Fatal("exact project bound refused", err)
	}
	if _, err := ParseProject(append(exact, ' '), unicodeReg); err == nil || !strings.Contains(err.Error(), "larger than") {
		t.Fatal("project bound changed", err)
	}
	raw := unicodeRegistry(`"modeling/gencontract"`, `"./cmd/refgen"`, "")
	exact = append(append([]byte(nil), raw...), bytes.Repeat([]byte(" "), maxRegistryBytes-len(raw))...)
	root := t.TempDir()
	unicodeRegistryFile(t, root, exact)
	if _, err := LoadRegistry(root); err != nil {
		t.Fatal("exact registry bound refused", err)
	}
	unicodeRegistryFile(t, root, append(exact, ' '))
	if _, err := LoadRegistry(root); err == nil || !strings.Contains(err.Error(), "larger than") {
		t.Fatal("registry bound changed", err)
	}
}

func TestUnicodeRunRefusesBeforeBuild(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("executable Go trap requires a POSIX host")
	}
	realGo, err := exec.LookPath("go")
	if err != nil {
		t.Fatal(err)
	}
	bin := t.TempDir()
	marker := filepath.Join(bin, "build-called")
	writeFile(t, filepath.Join(bin, "go"), "#!/bin/sh\nif [ \"${1:-}\" = build ]; then printf called >> \"$UNICODE_GO_MARKER\"; exit 97; fi\nexec \"$UNICODE_REAL_GO\" \"$@\"\n")
	if err := os.Chmod(filepath.Join(bin, "go"), 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("UNICODE_GO_MARKER", marker)
	t.Setenv("UNICODE_REAL_GO", realGo)
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	model := readFile(t, filepath.Join(repoRoot(t), "modeling/gencontract/testdata/reference/model.json"))
	fixture := func() string {
		p := t.TempDir()
		writeFile(t, filepath.Join(p, "design", "�.json"), model)
		writeFile(t, filepath.Join(p, "generated", "keep"), "prior output\n")
		return p
	}
	invoke := func(verb, project, foundry string) (int, string) {
		var so, se bytes.Buffer
		c := Run([]string{verb, "--project", project, "--foundry", foundry, "--unpinned"}, &so, &se)
		return c, so.String()
	}
	// The positive control proves the forwarding trap can observe a compiler build.
	armed := fixture()
	writeFile(t, filepath.Join(armed, ProjectFile), string(unicodeProject(`"design/�.json"`, `"generated/new"`)))
	if code, _ := invoke("plan", armed, repoRoot(t)); code != exitFailure {
		t.Fatalf("build trap not reached: %d", code)
	}
	if _, err := os.Stat(marker); err != nil {
		t.Fatal("build trap did not record positive control")
	}
	if err := os.Remove(marker); err != nil {
		t.Fatal(err)
	}
	for _, verb := range []string{"plan", "check", "generate"} {
		for _, kind := range []string{"model surrogate", "model bytes", "out surrogate", "registry module", "registry ignored value"} {
			t.Run(verb+"/"+kind, func(t *testing.T) {
				project := fixture()
				foundry := repoRoot(t)
				wireModel, wireOut := `"design/�.json"`, `"generated/new"`
				switch kind {
				case "model surrogate":
					wireModel = `"design/\ud800.json"`
				case "model bytes":
					wireModel = `"design/` + string([]byte{0xff}) + `.json"`
				case "out surrogate":
					wireOut = `"generated/\ud800"`
				case "registry module", "registry ignored value":
					foundry = t.TempDir()
					module, extra := `"modeling/�"`, ""
					if kind == "registry module" {
						module = `"modeling/\ud800"`
					} else {
						extra = `,"note":"\ud800"`
					}
					unicodeRegistryFile(t, foundry, unicodeRegistry(module, `"./cmd/refgen"`, extra))
					if err := os.MkdirAll(filepath.Join(foundry, "modeling", "�"), 0700); err != nil {
						t.Fatal(err)
					}
				}
				writeFile(t, filepath.Join(project, ProjectFile), string(unicodeProject(wireModel, wireOut)))
				before := unicodeSnapshot(t, project)
				var source map[string]unicodeSaved
				if foundry != repoRoot(t) {
					source = unicodeSnapshot(t, foundry)
				}
				code, stdout := invoke(verb, project, foundry)
				if code != exitRefused || stdout != "" {
					t.Errorf("malformed configuration reached later work: code%d stdout%q", code, stdout)
				}
				if _, err := os.Lstat(marker); !os.IsNotExist(err) {
					t.Errorf("compiler build reached: %v", err)
					_ = os.Remove(marker)
				}
				if !unicodeSame(before, unicodeSnapshot(t, project)) {
					t.Fatal("project/output bytes or metadata changed")
				}
				if source != nil && !unicodeSame(source, unicodeSnapshot(t, foundry)) {
					t.Fatal("registry/source bytes or metadata changed")
				}
			})
		}
	}
}

func TestUnicodeValidFilenameRun(t *testing.T) {
	model := readFile(t, filepath.Join(repoRoot(t), "modeling/gencontract/testdata/reference/model.json"))
	for _, tc := range []struct{ wire, path string }{
		{`"design/�.json"`, "design/�.json"}, {`"design/\ufffd.json"`, "design/�.json"},
		{`"design/🙂.json"`, "design/🙂.json"}, {`"design/\ud83d\ude42.json"`, "design/🙂.json"},
	} {
		t.Run(tc.wire, func(t *testing.T) {
			root := t.TempDir()
			writeFile(t, filepath.Join(root, filepath.FromSlash(tc.path)), model)
			writeFile(t, filepath.Join(root, ProjectFile), string(unicodeProject(tc.wire, `"generated"`)))
			before := unicodeSnapshot(t, filepath.Join(root, "design"))
			code, stdout, stderr := runCLI(t, "plan", root)
			if code != exitOK {
				t.Fatalf("valid Unicode plan refused: %d %s", code, stderr)
			}
			var plan struct {
				Outputs []struct {
					Model string `json:"model"`
				} `json:"outputs"`
			}
			if err := json.Unmarshal([]byte(stdout), &plan); err != nil || len(plan.Outputs) != 1 || plan.Outputs[0].Model != tc.path {
				t.Fatalf("valid filename changed: %s %v", stdout, err)
			}
			for _, verb := range []string{"generate", "check"} {
				if code, _, stderr := runCLI(t, verb, root); code != exitOK {
					t.Fatalf("valid Unicode %s refused: %d %s", verb, code, stderr)
				}
			}
			if !unicodeSame(before, unicodeSnapshot(t, filepath.Join(root, "design"))) {
				t.Fatal("valid model bytes or metadata changed")
			}
		})
	}
}

type unicodeSaved struct {
	info os.FileInfo
	data []byte
}

func unicodeSnapshot(t *testing.T, root string) map[string]unicodeSaved {
	t.Helper()
	m := map[string]unicodeSaved{}
	err := filepath.WalkDir(root, func(p string, d os.DirEntry, e error) error {
		if e != nil {
			return e
		}
		info, e := os.Lstat(p)
		if e != nil {
			return e
		}
		var data []byte
		if info.Mode().IsRegular() {
			data, e = os.ReadFile(p)
			if e != nil {
				return e
			}
		} else if !info.IsDir() {
			return fmt.Errorf("unexpected fixture file type")
		}
		rel, e := filepath.Rel(root, p)
		if e != nil {
			return e
		}
		m[rel] = unicodeSaved{info, data}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return m
}
func unicodeSame(a, b map[string]unicodeSaved) bool {
	if len(a) != len(b) {
		return false
	}
	for p, x := range a {
		y, ok := b[p]
		if !ok || !os.SameFile(x.info, y.info) || x.info.Mode() != y.info.Mode() || !x.info.ModTime().Equal(y.info.ModTime()) || !bytes.Equal(x.data, y.data) {
			return false
		}
	}
	return true
}
