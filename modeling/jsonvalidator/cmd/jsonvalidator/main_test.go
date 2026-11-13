package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"slices"
	"sort"
	"strings"
	"testing"
)

func wrap(root string) string { return `{"profile":"` + profile + `","root":` + root + `}` }

func TestDeterminismAndSavedOutput(t *testing.T) {
	model, err := os.ReadFile("../../testdata/sample/model.json")
	if err != nil {
		t.Fatal(err)
	}
	go1, ts1, err := compile(model)
	if err != nil {
		t.Fatal(err)
	}
	go2, ts2, err := compile(model)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(go1, go2) || !bytes.Equal(ts1, ts2) {
		t.Fatal("generation is nondeterministic")
	}
	for _, item := range []struct {
		name string
		data []byte
	}{{"validator.go", go1}, {"validator.ts", ts1}} {
		actual, err := os.ReadFile(filepath.Join("../../testdata/sample", item.name))
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(actual, item.data) {
			t.Fatalf("generated drift in %s", item.name)
		}
	}
}

func buildCLI(t *testing.T) string {
	t.Helper()
	// Build the real entry point: go run would replace the program's exit status.
	binary := filepath.Join(t.TempDir(), "jsonvalidator.exe")
	build := exec.CommandContext(t.Context(), filepath.Join(runtime.GOROOT(), "bin", "go"), "build", "-o", binary, ".")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build CLI: %v\n%s", err, output)
	}
	return binary
}

// runCLI runs the binary and returns its exit code, stdout and stderr.
func runCLI(t *testing.T, binary string, args ...string) (int, string, string) {
	t.Helper()
	command := exec.CommandContext(t.Context(), binary, args...)
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	exit := 0
	if err := command.Run(); err != nil {
		var exitError *exec.ExitError
		if !errors.As(err, &exitError) {
			t.Fatalf("run CLI: %v", err)
		}
		exit = exitError.ExitCode()
	}
	return exit, stdout.String(), stderr.String()
}

func TestCLI(t *testing.T) {
	binary := buildCLI(t)
	model, err := filepath.Abs("../../testdata/sample/model.json")
	if err != nil {
		t.Fatal(err)
	}
	modelBytes, err := os.ReadFile(model)
	if err != nil {
		t.Fatal(err)
	}
	l, err := parseModel(modelBytes)
	if err != nil {
		t.Fatal(err)
	}
	wantLine := l.identityLine()
	// fail runs a command that must exit with code, print nothing on stdout and
	// name the diagnostic on stderr.
	fail := func(t *testing.T, code int, diagnostic string, args ...string) {
		t.Helper()
		exit, stdout, stderr := runCLI(t, binary, args...)
		if exit != code {
			t.Errorf("exit = %d, want %d; stderr: %s", exit, code, stderr)
		}
		if stdout != "" {
			t.Errorf("unexpected stdout: %s", stdout)
		}
		if stderr == "" || !strings.Contains(stderr, diagnostic) {
			t.Errorf("stderr = %q, want a diagnostic containing %q", stderr, diagnostic)
		}
	}
	absent := func(t *testing.T, path string) {
		t.Helper()
		if _, err := os.Lstat(path); !os.IsNotExist(err) {
			t.Errorf("%s exists: %v", path, err)
		}
	}
	t.Run("validate", func(t *testing.T) {
		exit, stdout, stderr := runCLI(t, binary, "validate", model)
		if exit != 0 || stdout != wantLine || stderr != "" {
			t.Errorf("exit %d stdout %q stderr %q, want line %q", exit, stdout, stderr, wantLine)
		}
	})
	t.Run("generate-and-check", func(t *testing.T) {
		out := filepath.Join(t.TempDir(), "generated")
		exit, stdout, stderr := runCLI(t, binary, "generate", model, out)
		if exit != 0 || stdout != wantLine || stderr != "" {
			t.Fatalf("generate: exit %d stdout %q stderr %q", exit, stdout, stderr)
		}
		entries, err := os.ReadDir(out)
		if err != nil || len(entries) != 3 {
			t.Fatalf("generated directory: entries = %v, err = %v", entries, err)
		}
		for _, name := range []string{"validator.go", "validator.ts", "foundry-manifest.json"} {
			got, err := os.ReadFile(filepath.Join(out, name))
			if err != nil {
				t.Fatal(err)
			}
			want, err := os.ReadFile(filepath.Join("../../testdata/sample", name))
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(got, want) {
				t.Errorf("CLI generated drift in %s", name)
			}
		}
		exit, stdout, stderr = runCLI(t, binary, "check", model, out)
		if exit != 0 || stdout != wantLine || stderr != "" {
			t.Errorf("check: exit %d stdout %q stderr %q", exit, stdout, stderr)
		}
	})
	t.Run("usage", func(t *testing.T) {
		out := filepath.Join(t.TempDir(), "generated")
		for name, args := range map[string][]string{
			"none":              nil,
			"old-flags":         {"-model", model, "-out", out},
			"old-flags-equals":  {"-model=" + model, "-out=" + out},
			"unknown-verb":      {"build", model, out},
			"help":              {"--help"},
			"validate-no-model": {"validate"},
			"validate-extra":    {"validate", model, out},
			"generate-no-out":   {"generate", model},
			"generate-extra":    {"generate", model, out, "extra"},
			"check-no-out":      {"check", model},
			"flag-after-verb":   {"generate", "-x", model, out},
		} {
			t.Run(name, func(t *testing.T) {
				fail(t, 2, "usage:", args...)
				absent(t, out)
			})
		}
	})
	t.Run("missing-model", func(t *testing.T) {
		dir := t.TempDir()
		input, out := filepath.Join(dir, "none.json"), filepath.Join(dir, "generated")
		fail(t, 2, "cannot open model", "validate", input)
		fail(t, 2, "cannot open model", "generate", input, out)
		fail(t, 2, "cannot open model", "check", input, out)
		absent(t, out)
	})
	t.Run("invalid-model", func(t *testing.T) {
		dir := t.TempDir()
		input, out := filepath.Join(dir, "model.json"), filepath.Join(dir, "generated")
		for _, data := range []string{"", "null", "{}", "[]", `{"profile":"json-boolean/v1","root":`} {
			if err := os.WriteFile(input, []byte(data), 0600); err != nil {
				t.Fatal(err)
			}
			fail(t, 2, "invalid model", "validate", input)
			fail(t, 2, "invalid model", "generate", input, out)
			fail(t, 2, "invalid model", "check", input, out)
			absent(t, out)
		}
	})
	t.Run("existing-output", func(t *testing.T) {
		dir := t.TempDir()
		keep := filepath.Join(dir, "keep")
		if err := os.Mkdir(keep, 0700); err != nil {
			t.Fatal(err)
		}
		file := filepath.Join(dir, "file")
		if err := os.WriteFile(file, []byte("keep file"), 0600); err != nil {
			t.Fatal(err)
		}
		link := filepath.Join(dir, "link")
		if err := os.Symlink(keep, link); err != nil {
			t.Fatal(err)
		}
		empty := filepath.Join(dir, "empty")
		if err := os.Mkdir(empty, 0700); err != nil {
			t.Fatal(err)
		}
		for _, out := range []string{keep, file, link, empty} {
			fail(t, 2, "already exists", "generate", model, out)
		}
		if got, err := os.ReadFile(file); err != nil || string(got) != "keep file" {
			t.Errorf("existing file changed: %q, %v", got, err)
		}
		for _, d := range []string{keep, empty} {
			if entries, err := os.ReadDir(d); err != nil || len(entries) != 0 {
				t.Errorf("existing directory changed: %v, %v", entries, err)
			}
		}
	})
	t.Run("missing-parent", func(t *testing.T) {
		out := filepath.Join(t.TempDir(), "no-such-parent", "generated")
		fail(t, 2, "parent", "generate", model, out)
		absent(t, filepath.Dir(out))
	})
	t.Run("drift", func(t *testing.T) {
		dir := t.TempDir()
		out := filepath.Join(dir, "generated")
		if exit, _, stderr := runCLI(t, binary, "generate", model, out); exit != 0 {
			t.Fatalf("generate: %d %s", exit, stderr)
		}
		driftLines := func(t *testing.T) []string {
			t.Helper()
			exit, stdout, stderr := runCLI(t, binary, "check", model, out)
			if exit != 3 || stdout != "" {
				t.Fatalf("check: exit %d stdout %q stderr %q", exit, stdout, stderr)
			}
			return strings.Split(strings.TrimSuffix(stderr, "\n"), "\n")
		}
		// modified
		if err := os.WriteFile(filepath.Join(out, "validator.ts"), []byte("changed"), 0600); err != nil {
			t.Fatal(err)
		}
		// missing
		if err := os.Remove(filepath.Join(out, "validator.go")); err != nil {
			t.Fatal(err)
		}
		// unexpected: a file and a directory with contents
		if err := os.WriteFile(filepath.Join(out, "extra.txt"), nil, 0600); err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(filepath.Join(out, "sub", "deeper"), 0700); err != nil {
			t.Fatal(err)
		}
		got := driftLines(t)
		sort.Strings(got)
		want := []string{"drift: missing validator.go", "drift: modified validator.ts", "drift: unexpected extra.txt", "drift: unexpected sub"}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("drift lines = %q, want %q", got, want)
		}
		// check must not repair anything.
		if _, err := os.Lstat(filepath.Join(out, "validator.go")); !os.IsNotExist(err) {
			t.Errorf("check recreated validator.go: %v", err)
		}
		if data, _ := os.ReadFile(filepath.Join(out, "validator.ts")); string(data) != "changed" {
			t.Error("check rewrote validator.ts")
		}
		// symlink in place of a file
		if err := os.Remove(filepath.Join(out, "validator.ts")); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(filepath.Join(dir, "elsewhere"), filepath.Join(out, "validator.ts")); err != nil {
			t.Fatal(err)
		}
		if got := driftLines(t); !slices.Contains(got, "drift: unexpected validator.ts") {
			t.Errorf("symlink drift lines = %q", got)
		}
		// manifest drift
		if err := os.WriteFile(filepath.Join(out, "foundry-manifest.json"), []byte("{}"), 0600); err != nil {
			t.Fatal(err)
		}
		if got := driftLines(t); !slices.Contains(got, "drift: modified foundry-manifest.json") {
			t.Errorf("manifest drift lines = %q", got)
		}
		// absent output
		gone := filepath.Join(dir, "gone")
		fail(t, 3, "drift: missing .", "check", model, gone)
		absent(t, gone)
	})
}

func TestCanonicalDigestInvariance(t *testing.T) {
	digest := func(t *testing.T, model string) string {
		t.Helper()
		l, err := parseModel([]byte(model))
		if err != nil {
			t.Fatal(err)
		}
		return l.canonicalSHA256()
	}
	base := `{"profile":"json-boolean/v1","root":{"type":"object","additionalProperties":true,"properties":[{"name":"b","required":true,"schema":{"type":"boolean","const":true}},{"name":"a","required":false,"schema":{"type":"array","items":{"type":"string"}}}]}}`
	want := digest(t, base)
	same := map[string]string{
		"whitespace":          "\n{ \"profile\" : \"json-boolean/v1\" ,\n\t\"root\" : { \"type\" : \"object\", \"additionalProperties\" : true, \"properties\" : [ { \"name\" : \"b\", \"required\" : true, \"schema\" : { \"type\" : \"boolean\", \"const\" : true } },\r\n { \"name\" : \"a\", \"required\" : false, \"schema\" : { \"type\" : \"array\", \"items\" : { \"type\" : \"string\" } } } ] } }\n\n",
		"top-level-key-order": `{"root":{"type":"object","additionalProperties":true,"properties":[{"name":"b","required":true,"schema":{"type":"boolean","const":true}},{"name":"a","required":false,"schema":{"type":"array","items":{"type":"string"}}}]},"profile":"json-boolean/v1"}`,
		"member-key-order":    `{"profile":"json-boolean/v1","root":{"properties":[{"schema":{"const":true,"type":"boolean"},"required":true,"name":"b"},{"schema":{"items":{"type":"string"},"type":"array"},"name":"a","required":false}],"additionalProperties":true,"type":"object"}}`,
		"property-order":      `{"profile":"json-boolean/v1","root":{"type":"object","additionalProperties":true,"properties":[{"name":"a","required":false,"schema":{"type":"array","items":{"type":"string"}}},{"name":"b","required":true,"schema":{"type":"boolean","const":true}}]}}`,
		"escaped-name":        `{"profile":"json-boolean/v1","root":{"type":"object","additionalProperties":true,"properties":[{"name":"b","required":true,"schema":{"type":"boolean","const":true}},{"name":"a","required":false,"schema":{"type":"array","items":{"type":"string"}}}]}}`,
	}
	for name, model := range same {
		if got := digest(t, model); got != want {
			t.Errorf("%s changed the canonical digest", name)
		}
	}
	different := map[string]string{
		"required":      strings.Replace(base, `"required":true`, `"required":false`, 1),
		"const":         strings.Replace(base, `"const":true`, `"const":false`, 1),
		"const-absent":  strings.Replace(base, `,"const":true`, ``, 1),
		"name":          strings.Replace(base, `"name":"b"`, `"name":"c"`, 1),
		"type":          strings.Replace(base, `"items":{"type":"string"}`, `"items":{"type":"boolean"}`, 1),
		"property-drop": strings.Replace(base, `,{"name":"a","required":false,"schema":{"type":"array","items":{"type":"string"}}}`, ``, 1),
	}
	for name, model := range different {
		if model == base {
			t.Fatalf("%s: mutation did not apply", name)
		}
		if got := digest(t, model); got == want {
			t.Errorf("%s did not change the canonical digest", name)
		}
	}
}

func TestReorderedPropertiesEmitIdenticalBytes(t *testing.T) {
	// Nested objects reorder too, so node numbering must follow canonical order.
	a := `{"profile":"json-boolean/v1","root":{"type":"object","additionalProperties":true,"properties":[{"name":"z","required":true,"schema":{"type":"object","additionalProperties":true,"properties":[{"name":"q","required":true,"schema":{"type":"string"}},{"name":"p","required":false,"schema":{"type":"boolean"}}]}},{"name":"m","required":false,"schema":{"type":"array","items":{"type":"string"}}}]}}`
	b := `{ "root" : {"properties":[{"schema":{"items":{"type":"string"},"type":"array"},"name":"m","required":false},{"name":"z","required":true,"schema":{"type":"object","properties":[{"name":"p","required":false,"schema":{"type":"boolean"}},{"name":"q","required":true,"schema":{"type":"string"}}],"additionalProperties":true}}],"type":"object","additionalProperties":true}, "profile":"json-boolean/v1"}
`
	dir := t.TempDir()
	var outs []string
	for i, model := range []string{a, b} {
		path := filepath.Join(dir, fmt.Sprintf("model%d.json", i))
		if err := os.WriteFile(path, []byte(model), 0600); err != nil {
			t.Fatal(err)
		}
		out := filepath.Join(dir, fmt.Sprintf("out%d", i))
		if err := generate(path, out); err != nil {
			t.Fatal(err)
		}
		outs = append(outs, out)
	}
	for _, name := range []string{"validator.go", "validator.ts", "foundry-manifest.json"} {
		x, err1 := os.ReadFile(filepath.Join(outs[0], name))
		y, err2 := os.ReadFile(filepath.Join(outs[1], name))
		if err1 != nil || err2 != nil || !bytes.Equal(x, y) {
			t.Errorf("%s differs between reordered models (%v, %v)", name, err1, err2)
		}
	}
}

func TestSupportedShapes(t *testing.T) {
	for _, root := range []string{
		`{"type":"string"}`, `{"type":"boolean"}`,
		`{"type":"boolean","const":true}`, `{"type":"boolean","const":false}`,
		`{"type":"array","items":{"type":"string"}}`,
		`{"type":"object","additionalProperties":true,"properties":[]}`,
		`{"type":"object","additionalProperties":true,"properties":[{"name":"__proto__","required":false,"schema":{"type":"string"}}]}`,
	} {
		if _, _, err := compile([]byte(wrap(root))); err != nil {
			t.Errorf("supported shape rejected: %s: %v", root, err)
		}
	}
}

func TestUnsupportedModelsRefuse(t *testing.T) {
	cases := map[string]string{
		"empty": "", "null": "null", "array": "[]", "missing-profile": `{"root":{"type":"string"}}`,
		"wrong-version":         `{"profile":"json-boolean/v2","root":{"type":"string"}}`,
		"duplicate-key":         `{"profile":"` + profile + `","profile":"` + profile + `","root":{"type":"string"}}`,
		"escaped-duplicate-key": `{"profile":"` + profile + `","profi\u006ce":"` + profile + `","root":{"type":"string"}}`,
		"trailing":              wrap(`{"type":"string"}`) + ` {}`,
		"unknown-root-field":    `{"profile":"` + profile + `","root":{"type":"string"},"ignored":true}`,
		"null-node":             wrap(`null`), "unknown-type": wrap(`{"type":"number"}`),
		"duplicate-node-type":     wrap(`{"type":"string","type":"boolean"}`),
		"nullable":                wrap(`{"type":"string","nullable":true}`),
		"coercion":                wrap(`{"type":"string","coerce":true}`),
		"default":                 wrap(`{"type":"string","default":"x"}`),
		"regex":                   wrap(`{"type":"string","pattern":".*"}`),
		"const-and-enum":          wrap(`{"type":"string","const":"a","enum":["a"]}`),
		"const-and-min":           wrap(`{"type":"string","const":"a","minLength":1}`),
		"enum-and-max":            wrap(`{"type":"string","enum":["a"],"maxLength":3}`),
		"const-number":            wrap(`{"type":"string","const":1}`),
		"const-null":              wrap(`{"type":"string","const":null}`),
		"const-control":           wrap(`{"type":"string","const":"a\u0001"}`),
		"const-newline":           wrap(`{"type":"string","const":"a\n"}`),
		"const-del":               wrap(`{"type":"string","const":"a\u007f"}`),
		"const-c1":                wrap(`{"type":"string","const":"a\u0085"}`),
		"const-lone-surrogate":    wrap(`{"type":"string","const":"\ud800"}`),
		"const-replacement-char":  wrap(`{"type":"string","const":"\ufffd"}`),
		"const-257-bytes":         wrap(`{"type":"string","const":"` + strings.Repeat("a", 257) + `"}`),
		"enum-empty":              wrap(`{"type":"string","enum":[]}`),
		"enum-null":               wrap(`{"type":"string","enum":null}`),
		"enum-not-array":          wrap(`{"type":"string","enum":"a"}`),
		"enum-number":             wrap(`{"type":"string","enum":["a",1]}`),
		"enum-duplicate":          wrap(`{"type":"string","enum":["a","b","a"]}`),
		"enum-33-values":          wrap(`{"type":"string","enum":[` + enumValues(33) + `]}`),
		"min-negative":            wrap(`{"type":"string","minLength":-1}`),
		"min-float":               wrap(`{"type":"string","minLength":1.0}`),
		"min-exponent":            wrap(`{"type":"string","minLength":1e2}`),
		"min-string":              wrap(`{"type":"string","minLength":"1"}`),
		"min-null":                wrap(`{"type":"string","minLength":null}`),
		"min-leading-zero":        wrap(`{"type":"string","minLength":01}`),
		"max-65536":               wrap(`{"type":"string","maxLength":65536}`),
		"min-above-max":           wrap(`{"type":"string","minLength":3,"maxLength":2}`),
		"closed-null":             wrap(`{"type":"object","additionalProperties":"false","properties":[]}`),
		"null-const":              wrap(`{"type":"boolean","const":null}`),
		"numeric-const":           wrap(`{"type":"boolean","const":1}`),
		"missing-items":           wrap(`{"type":"array"}`),
		"null-items":              wrap(`{"type":"array","items":null}`),
		"missing-extra-policy":    wrap(`{"type":"object","properties":[]}`),
		"null-extra-policy":       wrap(`{"type":"object","additionalProperties":null,"properties":[]}`),
		"null-properties":         wrap(`{"type":"object","additionalProperties":true,"properties":null}`),
		"missing-required":        wrap(`{"type":"object","additionalProperties":true,"properties":[{"name":"x","schema":{"type":"string"}}]}`),
		"null-required":           wrap(`{"type":"object","additionalProperties":true,"properties":[{"name":"x","required":null,"schema":{"type":"string"}}]}`),
		"property-code":           wrap(`{"type":"object","additionalProperties":true,"properties":[{"name":"x\";process.exit()","required":true,"schema":{"type":"string"}}]}`),
		"surrogate-property-name": wrap(`{"type":"object","additionalProperties":true,"properties":[{"name":"\ud800","required":true,"schema":{"type":"string"}}]}`),
		"empty-property-name":     wrap(`{"type":"object","additionalProperties":true,"properties":[{"name":"","required":true,"schema":{"type":"string"}}]}`),
		"duplicate-property":      wrap(`{"type":"object","additionalProperties":true,"properties":[{"name":"x","required":true,"schema":{"type":"string"}},{"name":"x","required":false,"schema":{"type":"string"}}]}`),
	}
	deep := `{"type":"string"}`
	for range 65 {
		deep = `{"type":"array","items":` + deep + `}`
	}
	cases["deep"] = wrap(deep)
	cases["byte-limit"] = strings.Repeat(" ", maxModelBytes) + wrap(`{"type":"string"}`)
	cases["invalid-utf8"] = string([]byte{0xff})
	for name, data := range cases {
		t.Run(name, func(t *testing.T) {
			g, s, err := compile([]byte(data))
			if err == nil || g != nil || s != nil {
				t.Fatal("invalid model produced output")
			}
		})
	}
}

func TestNodeAndPropertyLimits(t *testing.T) {
	for _, size := range []int{129, 513} {
		props := make([]any, size)
		for i := range props {
			props[i] = map[string]any{"name": "p" + strings.Repeat("x", i+1), "required": true, "schema": map[string]any{"type": "string"}}
		}
		data, _ := json.Marshal(map[string]any{"profile": profile, "root": map[string]any{"type": "object", "additionalProperties": true, "properties": props}})
		if _, _, err := compile(data); err == nil {
			t.Fatal("oversized properties accepted")
		}
	}
	// 128 declared fields, each containing three array nodes plus a string: 513 nodes.
	props := make([]any, 128)
	for i := range props {
		var child any = map[string]any{"type": "string"}
		for range 3 {
			child = map[string]any{"type": "array", "items": child}
		}
		props[i] = map[string]any{"name": "p" + string(rune('A'+i/26)) + string(rune('A'+i%26)), "required": true, "schema": child}
	}
	data, _ := json.Marshal(map[string]any{"profile": profile, "root": map[string]any{"type": "object", "additionalProperties": true, "properties": props}})
	if _, _, err := compile(data); err == nil {
		t.Fatal("oversized node count accepted")
	}
}

func TestOutputNeverClobbersExistingDestination(t *testing.T) {
	dir := t.TempDir()
	model := filepath.Join(dir, "model.json")
	if err := os.WriteFile(model, []byte(wrap(`{"type":"string"}`)), 0600); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(dir, "existing")
	if err := os.Mkdir(out, 0700); err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(out, "validator.go")
	if err := os.WriteFile(file, []byte("keep"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := generate(model, out); err == nil {
		t.Fatal("existing output accepted")
	}
	link := filepath.Join(dir, "link")
	if err := os.Symlink(out, link); err != nil {
		t.Fatal(err)
	}
	if err := generate(model, link); err == nil {
		t.Fatal("symlink output accepted")
	}
	got, _ := os.ReadFile(file)
	if string(got) != "keep" {
		t.Fatal("existing source changed")
	}
	if err := os.WriteFile(model, []byte(`null`), 0600); err != nil {
		t.Fatal(err)
	}
	invalidOut := filepath.Join(dir, "invalid-output")
	if err := generate(model, invalidOut); err == nil {
		t.Fatal("invalid model generated")
	}
	if _, err := os.Stat(invalidOut); !os.IsNotExist(err) {
		t.Fatal("invalid model created output")
	}
}

func TestCompilerExactBoundaries(t *testing.T) {
	stringNode := map[string]any{"type": "string"}
	array := func(child any) any { return map[string]any{"type": "array", "items": child} }
	object := func(properties []any) any {
		return map[string]any{"type": "object", "additionalProperties": true, "properties": properties}
	}
	property := func(name string, child any) any {
		return map[string]any{"name": name, "required": true, "schema": child}
	}
	encode := func(root any) []byte {
		data, err := json.Marshal(map[string]any{"profile": profile, "root": root})
		if err != nil {
			t.Fatal(err)
		}
		return data
	}
	properties := func(size, shortFirst int) any {
		props := make([]any, size)
		for i := range props {
			var child any = stringNode
			wraps := 3
			if i == 0 {
				wraps -= shortFirst
			}
			for j := 0; j < wraps; j++ {
				child = array(child)
			}
			props[i] = property(fmt.Sprintf("p%03d", i), child)
		}
		return object(props)
	}
	depthModel := func(arrays, objects int) []byte {
		var child any = stringNode
		for range arrays {
			child = array(child)
		}
		for range objects {
			child = object([]any{property("x", child)})
		}
		return encode(child)
	}
	countDepth := func(data []byte) int {
		var value any
		if err := json.Unmarshal(data, &value); err != nil {
			t.Fatal(err)
		}
		var visit func(any, int) int
		visit = func(v any, depth int) int {
			result := depth
			switch value := v.(type) {
			case map[string]any:
				for _, child := range value {
					if n := visit(child, depth+1); n > result {
						result = n
					}
				}
			case []any:
				for _, child := range value {
					if n := visit(child, depth+1); n > result {
						result = n
					}
				}
			}
			return result
		}
		// The document root is depth one, matching the documented intake limit.
		return visit(value, 1)
	}
	json64, json65 := depthModel(1, 20), depthModel(2, 20)
	if countDepth(json64) != 64 || countDepth(json65) != 65 {
		t.Fatal("structural depth fixture is not on the stated boundary")
	}
	schema32, schema33 := depthModel(31, 0), depthModel(32, 0)
	for _, data := range [][]byte{schema32, schema33} {
		if err := checkJSON(data); err != nil {
			t.Fatalf("schema limit fixture hits JSON limit first: %v", err)
		}
	}
	simple := encode(stringNode)
	bytesAtLimit := append(bytes.Repeat([]byte(" "), 1048576-len(simple)), simple...)
	cases := []struct {
		name       string
		model      []byte
		accepted   bool
		diagnostic string
	}{
		{"name-128", encode(object([]any{property(strings.Repeat("x", 128), stringNode)})), true, ""},
		{"name-129", encode(object([]any{property(strings.Repeat("x", 129), stringNode)})), false, "invalid or duplicate property name"},
		{"properties-128", encode(func() any {
			p := make([]any, 128)
			for i := range p {
				p[i] = property(fmt.Sprintf("p%03d", i), stringNode)
			}
			return object(p)
		}()), true, ""},
		{"properties-129", encode(func() any {
			p := make([]any, 129)
			for i := range p {
				p[i] = property(fmt.Sprintf("p%03d", i), stringNode)
			}
			return object(p)
		}()), false, "invalid model properties"},
		{"schema-depth-32", schema32, true, ""},
		{"schema-depth-33", schema33, false, "model schema limit exceeded"},
		{"nodes-512", encode(properties(128, 1)), true, ""},
		{"nodes-513", encode(properties(128, 0)), false, "model schema limit exceeded"},
		{"json-depth-64", json64, true, ""},
		{"json-depth-65", json65, false, "model JSON depth exceeded"},
		{"bytes-1048576", bytesAtLimit, true, ""},
		{"bytes-1048577", append(append([]byte{}, bytesAtLimit...), ' '), false, "invalid model bytes"},
	}
	// Execute only these two maximum-size outputs. Keep the literal expectations
	// separate from the schema builders rather than interpreting the model again.
	nestedValue := func(arrays int, leaf string) string {
		return strings.Repeat("[", arrays) + leaf + strings.Repeat("]", arrays)
	}
	wideValue := func(first, last string) string {
		var fields []string
		for i := range 128 {
			value := `[[["x"]]]`
			if i == 0 {
				value = first
			} else if i == 127 {
				value = last
			}
			if value != "" {
				fields = append(fields, fmt.Sprintf(`"p%03d":%s`, i, value))
			}
		}
		return "{" + strings.Join(fields, ",") + "}"
	}
	t.Run("nodes-512-execution", func(t *testing.T) {
		checkGeneratedLiterals(t, encode(properties(128, 1)), []profileCase{
			{wideValue(`[["x"]]`, `[[["x"]]]`), true},
			{wideValue(`[]`, `[]`), true},
			{wideValue("", `[[["x"]]]`), false},
			{wideValue(`[["x"]]`, ""), false},
			{wideValue(`[[false]]`, `[[["x"]]]`), false},
			{wideValue(`[["x"]]`, `[[[false]]]`), false},
			{wideValue(`[["x"]]`, `[[["x",false]]]`), false},
			{`[]`, false},
		})
	})
	t.Run("schema-depth-32-execution", func(t *testing.T) {
		checkGeneratedLiterals(t, schema32, []profileCase{
			{nestedValue(31, `"x"`), true},
			{`[]`, true},
			{nestedValue(30, `[]`), true},
			{nestedValue(31, `false`), false},
			{nestedValue(30, `["x",false]`), false},
			{nestedValue(30, `"x"`), false},
			{nestedValue(32, `"x"`), false},
			{`null`, false},
		})
	})
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			dir := t.TempDir()
			model := filepath.Join(dir, "model.json")
			out := filepath.Join(dir, "generated")
			if err := os.WriteFile(model, c.model, 0600); err != nil {
				t.Fatal(err)
			}
			err := generate(model, out)
			if c.accepted {
				if err != nil {
					t.Fatal(err)
				}
				for _, name := range []string{"validator.go", "validator.ts"} {
					data, err := os.ReadFile(filepath.Join(out, name))
					if err != nil || len(data) == 0 {
						t.Fatalf("missing emitted output %s: %v", name, err)
					}
				}
			} else {
				if err == nil || !strings.Contains(err.Error(), c.diagnostic) {
					t.Fatalf("wrong refusal: %v", err)
				}
				if _, err := os.Stat(out); !os.IsNotExist(err) {
					t.Fatal("refused model created output directory")
				}
			}
		})
	}
}

func enumValues(n int) string {
	var parts []string
	for i := range n {
		parts = append(parts, fmt.Sprintf("%q", fmt.Sprintf("v%d", i)))
	}
	return strings.Join(parts, ",")
}

func TestNewConstructsBoundaries(t *testing.T) {
	for _, root := range []string{
		`{"type":"string","const":""}`,
		`{"type":"string","const":"` + strings.Repeat("a", 256) + `"}`,
		`{"type":"string","const":"caf\u00e9\ud83d\ude00"}`,
		`{"type":"string","enum":["a"]}`,
		`{"type":"string","enum":[` + enumValues(32) + `]}`,
		`{"type":"string","minLength":0}`,
		`{"type":"string","maxLength":65535}`,
		`{"type":"string","minLength":2,"maxLength":2}`,
		`{"type":"object","additionalProperties":false,"properties":[]}`,
	} {
		if _, _, err := compile([]byte(wrap(root))); err != nil {
			t.Errorf("supported shape rejected: %.80s: %v", root, err)
		}
	}
}

// The models below used only constructs that existed in 1.0.0. Their canonical
// digests must not move.
func TestExistingDigestsUnchanged(t *testing.T) {
	sample, err := os.ReadFile("../../testdata/sample/model.json")
	if err != nil {
		t.Fatal(err)
	}
	object := func(prop string) string {
		return `{"profile":"json-boolean/v1","root":{"type":"object","additionalProperties":true,"properties":[` + prop + `]}}`
	}
	for name, item := range map[string]struct{ model, want string }{
		"sample":               {string(sample), "01bea4ac3c1699b3b2381c9d1339fc2cd8ce9b2fbb86b80e093b3834fc254563"},
		"requirement-optional": {object(`{"name":"enabled","required":false,"schema":{"type":"boolean"}}`), "2d686ed86c427d4c0b8118e495c42545abd8469fb5a9bed8948839cc023df6c3"},
		"requirement-required": {object(`{"name":"enabled","required":true,"schema":{"type":"boolean"}}`), "b57974dd597ddbf8aff73d0cc26d1190fec715fb01839509f24059f909e9c8da"},
		"requirement-constant": {object(`{"name":"enabled","required":true,"schema":{"type":"boolean","const":true}}`), "3e518d8b9f85b2b52f8c027acccb4f360791f60fa6bf92e14ba6c18e3c71f887"},
	} {
		l, err := parseModel([]byte(item.model))
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if got := l.canonicalSHA256(); got != item.want {
			t.Errorf("%s digest moved: %s", name, got)
		}
	}
}

func TestNewConstructCanonicalForms(t *testing.T) {
	canon := func(root string) string {
		l, err := parseModel([]byte(wrap(root)))
		if err != nil {
			t.Fatal(err)
		}
		return string(l.canonical)
	}
	for _, c := range []struct{ root, want string }{
		{`{"maxLength":5,"type":"string","minLength":1}`, `{"type":"string","minLength":1,"maxLength":5}`},
		{`{"type":"string","enum":["b","a","\u00e9","B"]}`, "{\"type\":\"string\",\"enum\":[\"B\",\"a\",\"b\",\"\u00e9\"]}"},
		{`{"type":"string","const":"a\"b\\c"}`, `{"type":"string","const":"a\"b\\c"}`},
	} {
		want := `{"profile":"json-boolean/v1","root":` + c.want + `}`
		if got := canon(c.root); got != want {
			t.Errorf("canonical form\n got %s\nwant %s", got, want)
		}
	}
	one, constant := canon(`{"type":"string","enum":["a"]}`), canon(`{"type":"string","const":"a"}`)
	if one == constant {
		t.Error("one-value enum and const share a canonical form")
	}
	closed := canon(`{"type":"object","additionalProperties":false,"properties":[]}`)
	open := canon(`{"type":"object","additionalProperties":true,"properties":[]}`)
	if closed == open || !strings.Contains(closed, `"additionalProperties":false`) {
		t.Error("closed and open objects must differ in the canonical form")
	}
	a := canon(`{"type":"string","enum":["z","y"]}`)
	b := canon(`{"type":"string","enum":["y","z"]}`)
	if a != b {
		t.Error("enum order changed the canonical form")
	}
}

func TestGeneratedClosedObjects(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		checkGeneratedLiterals(t, []byte(wrap(`{"type":"object","additionalProperties":false,"properties":[]}`)), []profileCase{
			{`{}`, true}, {`{"a":1}`, false}, {`{"__proto__":{}}`, false}, {`{"constructor":1}`, false}, {`[]`, false}, {`null`, false},
		})
	})
	for _, name := range []string{"__proto__", "constructor", "toString"} {
		t.Run("declared-"+name, func(t *testing.T) {
			quoted, _ := json.Marshal(name)
			model := wrap(`{"type":"object","additionalProperties":false,"properties":[{"name":` + string(quoted) + `,"required":false,"schema":{"type":"string"}}]}`)
			checkGeneratedLiterals(t, []byte(model), []profileCase{
				{`{}`, true}, {`{` + string(quoted) + `:"x"}`, true}, {`{` + string(quoted) + `:1}`, false},
				{`{"other":"x"}`, false}, {`{` + string(quoted) + `:"x","other":"x"}`, false},
			})
		})
		t.Run("undeclared-"+name, func(t *testing.T) {
			quoted, _ := json.Marshal(name)
			model := wrap(`{"type":"object","additionalProperties":false,"properties":[{"name":"a","required":false,"schema":{"type":"string"}}]}`)
			checkGeneratedLiterals(t, []byte(model), []profileCase{
				{`{"a":"x"}`, true}, {`{` + string(quoted) + `:"x"}`, false}, {`{` + string(quoted) + `:null}`, false}, {`{"a":"x",` + string(quoted) + `:{}}`, false},
			})
		})
	}
	t.Run("nested", func(t *testing.T) {
		model := wrap(`{"type":"array","items":{"type":"object","additionalProperties":false,"properties":[{"name":"a","required":true,"schema":{"type":"object","additionalProperties":true,"properties":[]}}]}}`)
		checkGeneratedLiterals(t, []byte(model), []profileCase{
			{`[{"a":{"free":1}}]`, true}, {`[{"a":{},"b":1}]`, false}, {`[{"a":{}},{"a":{},"z":0}]`, false}, {`[]`, true},
		})
	})
}

func TestGeneratedStringConstructs(t *testing.T) {
	checkGeneratedLiterals(t, []byte(wrap(`{"type":"string","const":"a\u00e9\ud83d\ude00"}`)), []profileCase{
		{`"a\u00e9\ud83d\ude00"`, true}, {`"aé😀"`, true}, {`"ae\u0301\ud83d\ude00"`, false}, {`"a\u00e9\ud83d"`, false}, {`"a\u00e9"`, false}, {`""`, false}, {`1`, false}, {`null`, false},
	})
	checkGeneratedLiterals(t, []byte(wrap(`{"type":"string","enum":["a","B","\u00e9"]}`)), []profileCase{
		{`"a"`, true}, {`"B"`, true}, {`"é"`, true}, {`"b"`, false}, {`"A"`, false}, {`"e\u0301"`, false}, {`"a "`, false}, {`""`, false}, {`["a"]`, false}, {`true`, false},
	})
	checkGeneratedLiterals(t, []byte(wrap(`{"type":"string","minLength":2,"maxLength":3}`)), []profileCase{
		{`""`, false}, {`"a"`, false}, {`"ab"`, true}, {`"abc"`, true}, {`"abcd"`, false},
		{`"\ud83d\ude00"`, false}, {`"\ud83d\ude00\ud83d\ude00"`, true}, {`"\ud83d\ude00\ud83d\ude00\ud83d\ude00\ud83d\ude00"`, false},
		{`"\ud800"`, false}, {`"\ud800\ud800"`, true}, {`"\ud800\ud800\ud800\ud800"`, false}, {`"\ude00\ud83d"`, true},
		{`"e\u0301"`, true}, {`1`, false}, {`null`, false},
	})
}
