package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
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

func TestCLI(t *testing.T) {
	// Build the real entry point: go run would replace the program's exit status.
	binary := filepath.Join(t.TempDir(), "jsonvalidator.exe")
	build := exec.CommandContext(t.Context(), filepath.Join(runtime.GOROOT(), "bin", "go"), "build", "-o", binary, ".")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build CLI: %v\n%s", err, output)
	}
	model, err := filepath.Abs("../../testdata/sample/model.json")
	if err != nil {
		t.Fatal(err)
	}
	run := func(t *testing.T, args []string, wantExit int, wantError string) {
		t.Helper()
		command := exec.CommandContext(t.Context(), binary, args...)
		var stdout, stderr bytes.Buffer
		command.Stdout, command.Stderr = &stdout, &stderr
		err := command.Run()
		exit := 0
		if err != nil {
			var exitError *exec.ExitError
			if !errors.As(err, &exitError) {
				t.Fatalf("run CLI: %v", err)
			}
			exit = exitError.ExitCode()
		}
		if exit != wantExit {
			t.Errorf("exit = %d, want %d; stderr: %s", exit, wantExit, &stderr)
		}
		if stdout.Len() != 0 {
			t.Errorf("unexpected stdout: %s", &stdout)
		}
		if wantExit == 0 && stderr.Len() != 0 {
			t.Errorf("unexpected stderr: %s", &stderr)
		} else if wantExit != 0 && (stderr.Len() == 0 || !strings.Contains(stderr.String(), wantError)) {
			t.Errorf("stderr = %q, want a diagnostic containing %q", stderr.String(), wantError)
		}
	}
	for _, form := range []string{"flags", "reversed-equals", "empty-terminator"} {
		t.Run(form, func(t *testing.T) {
			out := filepath.Join(t.TempDir(), "generated")
			args := []string{"-model", model, "-out", out}
			if form == "reversed-equals" {
				args = []string{"-out=" + out, "-model=" + model}
			} else if form == "empty-terminator" {
				args = append(args, "--")
			}
			run(t, args, 0, "")
			entries, err := os.ReadDir(out)
			if err != nil || len(entries) != 2 {
				t.Fatalf("generated directory: entries = %v, err = %v", entries, err)
			}
			for _, name := range []string{"validator.go", "validator.ts"} {
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
		})
	}
	for _, name := range []string{
		"missing-both", "missing-model", "missing-out", "empty-model", "empty-out",
		"unknown-flag", "missing-flag-value", "leading-positional", "between-flags",
		"trailing-positional", "after-terminator", "positionals-only",
	} {
		t.Run(name, func(t *testing.T) {
			out := filepath.Join(t.TempDir(), "generated")
			args := []string{"-model", model, "-out", out}
			diagnostic := "use -model MODEL -out NEW_DIRECTORY"
			switch name {
			case "missing-both":
				args = nil
			case "missing-model":
				args = []string{"-out", out}
			case "missing-out":
				args = []string{"-model", model}
			case "empty-model":
				args = []string{"-model=", "-out", out}
			case "empty-out":
				args = []string{"-model", model, "-out="}
			case "unknown-flag":
				args = append(args, "-unknown")
				diagnostic = "flag provided but not defined: -unknown"
			case "missing-flag-value":
				args = []string{"-model", model, "-out"}
				diagnostic = "flag needs an argument: -out"
			case "leading-positional":
				args = append([]string{"extra"}, args...)
			case "between-flags":
				args = []string{"-model", model, "extra", "-out", out}
			case "trailing-positional":
				args = append(args, "extra")
			case "after-terminator":
				args = append(args, "--", "extra")
			case "positionals-only":
				args = []string{model, out}
			}
			run(t, args, 2, diagnostic)
			if _, err := os.Lstat(out); !os.IsNotExist(err) {
				t.Errorf("usage error created output: %v", err)
			}
		})
	}
	for _, name := range []string{"missing-model-file", "invalid-model"} {
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			input, out := filepath.Join(dir, "model.json"), filepath.Join(dir, "generated")
			if name == "invalid-model" {
				if err := os.WriteFile(input, []byte("null"), 0600); err != nil {
					t.Fatal(err)
				}
			}
			run(t, []string{"-model", input, "-out", out}, 1, "")
			if _, err := os.Lstat(out); !os.IsNotExist(err) {
				t.Errorf("input error created output: %v", err)
			}
		})
	}
	t.Run("existing-output", func(t *testing.T) {
		out := t.TempDir()
		for _, name := range []string{"validator.go", "validator.ts", "keep.txt"} {
			if err := os.WriteFile(filepath.Join(out, name), []byte("keep "+name), 0600); err != nil {
				t.Fatal(err)
			}
		}
		run(t, []string{"-model", model, "-out", out}, 1, out)
		entries, err := os.ReadDir(out)
		if err != nil || len(entries) != 3 {
			t.Fatalf("existing directory changed: entries = %v, err = %v", entries, err)
		}
		for _, name := range []string{"validator.go", "validator.ts", "keep.txt"} {
			got, err := os.ReadFile(filepath.Join(out, name))
			if err != nil || string(got) != "keep "+name {
				t.Errorf("existing %s changed: data = %q, err = %v", name, got, err)
			}
		}
	})
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
		"null-const":              wrap(`{"type":"boolean","const":null}`),
		"numeric-const":           wrap(`{"type":"boolean","const":1}`),
		"missing-items":           wrap(`{"type":"array"}`),
		"null-items":              wrap(`{"type":"array","items":null}`),
		"missing-extra-policy":    wrap(`{"type":"object","properties":[]}`),
		"reject-extras":           wrap(`{"type":"object","additionalProperties":false,"properties":[]}`),
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
