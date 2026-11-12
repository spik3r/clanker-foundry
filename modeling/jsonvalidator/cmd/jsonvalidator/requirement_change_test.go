package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

func TestGeneratedRequirementChanges(t *testing.T) {
	// These literal models and expectations were reviewed against PROFILE.md
	// before executing generated code. Each revision narrows one requirement.
	models := []struct{ name, source string }{
		{"optional", `{"profile":"json-boolean/v1","root":{"type":"object","additionalProperties":true,"properties":[{"name":"enabled","required":false,"schema":{"type":"boolean"}}]}}`},
		{"required", `{"profile":"json-boolean/v1","root":{"type":"object","additionalProperties":true,"properties":[{"name":"enabled","required":true,"schema":{"type":"boolean"}}]}}`},
		{"constant", `{"profile":"json-boolean/v1","root":{"type":"object","additionalProperties":true,"properties":[{"name":"enabled","required":true,"schema":{"type":"boolean","const":true}}]}}`},
	}
	cases := []struct {
		raw   string
		valid [3]bool
	}{
		{`{}`, [3]bool{true, false, false}},
		{`{"extra":null}`, [3]bool{true, false, false}},
		{`{"enabled":true}`, [3]bool{true, true, true}},
		{`{"enabled":false}`, [3]bool{true, true, false}},
		{`{"enabled":true,"extra":[null,1]}`, [3]bool{true, true, true}},
		{`{"enabled":false,"extra":{}}`, [3]bool{true, true, false}},
		{`{"enabled":null}`, [3]bool{false, false, false}},
		{`{"enabled":0}`, [3]bool{false, false, false}},
		{`{"enabled":"true"}`, [3]bool{false, false, false}},
		{`null`, [3]bool{false, false, false}},
		{`[]`, [3]bool{false, false, false}},
	}
	dir := t.TempDir()
	binary := filepath.Join(dir, "jsonvalidator.exe")
	build := exec.CommandContext(t.Context(), filepath.Join(runtime.GOROOT(), "bin", "go"), "build", "-o", binary, ".")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build CLI: %v\n%s", err, output)
	}
	modelPath := filepath.Join(dir, "model.json")
	names := []string{"validator.go", "validator.ts"}
	saved := map[string][][]byte{}
	generateRevision := func(t *testing.T, name, source string) [][]byte {
		t.Helper()
		model := []byte(source)
		if err := os.WriteFile(modelPath, model, 0600); err != nil {
			t.Fatal(err)
		}
		out := filepath.Join(dir, name)
		command := exec.CommandContext(t.Context(), binary, "-model", modelPath, "-out", out)
		if output, err := command.CombinedOutput(); err != nil || len(output) != 0 {
			t.Fatalf("generate %s: %v\n%s", name, err, output)
		}
		if got, err := os.ReadFile(modelPath); err != nil || !bytes.Equal(got, model) {
			t.Fatalf("generator changed model bytes: %v", err)
		}
		goSource, tsSource, err := compile(model)
		if err != nil {
			t.Fatal(err)
		}
		// Tie the CLI files to the source executed by checkGeneratedLiterals.
		compiled := [][]byte{goSource, tsSource}
		digest := fmt.Sprintf("%x", sha256.Sum256(model))
		declarations := []string{
			fmt.Sprintf("const ModelSHA256 = %q", digest),
			fmt.Sprintf("export const MODEL_SHA256 = %q;", digest),
		}
		actual := make([][]byte, len(names))
		for i, file := range names {
			actual[i], err = os.ReadFile(filepath.Join(out, file))
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(actual[i], compiled[i]) {
				t.Fatalf("CLI and executed compiler output differ: %s", file)
			}
			if !bytes.Contains(actual[i], []byte(declarations[i])) {
				t.Errorf("%s lacks exact raw-model provenance", file)
			}
		}
		for previous, files := range saved {
			for i, file := range names {
				got, err := os.ReadFile(filepath.Join(dir, previous, file))
				if err != nil || !bytes.Equal(got, files[i]) {
					t.Errorf("earlier %s/%s changed: %v", previous, file, err)
				}
			}
		}
		saved[name] = actual
		return actual
	}
	for revision, model := range models {
		if !t.Run(model.name, func(t *testing.T) {
			actual := generateRevision(t, model.name, model.source)
			if revision > 0 {
				for i, file := range names {
					if bytes.Equal(actual[i], saved[models[revision-1].name][i]) {
						t.Errorf("%s retained stale output after requirement change", file)
					}
				}
			}
			literals := make([]profileCase, len(cases))
			for i, item := range cases {
				literals[i] = profileCase{Raw: item.raw, Valid: item.valid[revision]}
			}
			checkGeneratedLiterals(t, []byte(model.source), literals)
		}) {
			return
		}
	}
	t.Run("restore-original", func(t *testing.T) {
		actual := generateRevision(t, "restored", models[0].source)
		for i, file := range names {
			if !bytes.Equal(actual[i], saved[models[0].name][i]) {
				t.Errorf("restored model did not reproduce original %s bytes", file)
			}
		}
	})
}
