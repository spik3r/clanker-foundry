package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

type profileCase struct {
	Raw   string `json:"raw"`
	Valid bool   `json:"valid"`
}

// Execute the emitted files, not a second interpretation of the input model.
func checkGeneratedLiterals(t *testing.T, model []byte, cases []profileCase) {
	t.Helper()
	goSource, tsSource, err := compile(model)
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	for _, file := range []struct {
		Name string
		Data []byte
	}{
		{"validator.go", goSource}, {"validator.ts", tsSource},
		{"go.mod", []byte("module example.test/profile\n\ngo 1.27.3\n")},
	} {
		if err := os.WriteFile(filepath.Join(dir, file.Name), file.Data, 0600); err != nil {
			t.Fatal(err)
		}
	}
	var testSource strings.Builder
	testSource.WriteString("package validator\nimport (\"encoding/json\";\"testing\")\nfunc TestLiterals(t *testing.T) {\n")
	for i, c := range cases {
		fmt.Fprintf(&testSource, "{var v any; if err:=json.Unmarshal([]byte(%q),&v);err!=nil {t.Fatal(err)}; if got:=Validate(v);got!=%t {t.Errorf(\"case %d: got %%v\",got)}}\n", c.Raw, c.Valid, i)
	}
	testSource.WriteString("}\n")
	if err := os.WriteFile(filepath.Join(dir, "validator_test.go"), []byte(testSource.String()), 0600); err != nil {
		t.Fatal(err)
	}
	command := exec.Command(filepath.Join(runtime.GOROOT(), "bin", "go"), "test", "-count=1", ".")
	command.Dir = dir
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("emitted Go failed: %v\n%s", err, output)
	}
	literals, _ := json.Marshal(cases)
	js := `import assert from 'node:assert/strict';
import {readFileSync} from 'node:fs';
import {stripTypeScriptTypes} from 'node:module';
const code=stripTypeScriptTypes(readFileSync(new URL('./validator.ts',import.meta.url),'utf8'),{mode:'transform'});
const module=await import('data:text/javascript;base64,'+Buffer.from(code).toString('base64'));
for(const c of ` + string(literals) + `) assert.equal(module.validate(JSON.parse(c.raw)),c.valid,c.raw);
`
	if err := os.WriteFile(filepath.Join(dir, "check.mjs"), []byte(js), 0600); err != nil {
		t.Fatal(err)
	}
	command = exec.Command("node", filepath.Join(dir, "check.mjs"))
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("emitted TypeScript failed: %v\n%s", err, output)
	}
}

func TestGeneratedRootShapes(t *testing.T) {
	for _, item := range []struct {
		Name, Root string
		Cases      []profileCase
	}{
		{"string", `{"type":"string"}`, []profileCase{{`""`, true}, {`"x"`, true}, {`null`, false}, {`true`, false}, {`1`, false}, {`[]`, false}, {`{}`, false}}},
		{"boolean", `{"type":"boolean"}`, []profileCase{{`true`, true}, {`false`, true}, {`null`, false}, {`"false"`, false}, {`0`, false}, {`[]`, false}, {`{}`, false}}},
		{"constant-true", `{"type":"boolean","const":true}`, []profileCase{{`true`, true}, {`false`, false}, {`1`, false}, {`"true"`, false}, {`null`, false}}},
		{"constant-false", `{"type":"boolean","const":false}`, []profileCase{{`false`, true}, {`true`, false}, {`0`, false}, {`"false"`, false}, {`null`, false}}},
		{"string-array", `{"type":"array","items":{"type":"string"}}`, []profileCase{{`[]`, true}, {`["","x"]`, true}, {`["x",1]`, false}, {`[null]`, false}, {`{}`, false}, {`null`, false}}},
		{"boolean-array", `{"type":"array","items":{"type":"boolean"}}`, []profileCase{{`[]`, true}, {`[true,false]`, true}, {`[true,null]`, false}, {`[0]`, false}, {`"false"`, false}}},
		{"empty-object", `{"type":"object","additionalProperties":true,"properties":[]}`, []profileCase{{`{}`, true}, {`{"arbitrary":[1,null,"x"]}`, true}, {`[]`, false}, {`null`, false}, {`true`, false}}},
	} {
		t.Run(item.Name, func(t *testing.T) { checkGeneratedLiterals(t, []byte(wrap(item.Root)), item.Cases) })
	}
}

// Presence rules must not mistake inherited JavaScript names for JSON fields.
func TestGeneratedOwnPropertyPresence(t *testing.T) {
	for _, name := range []string{"__proto__", "constructor", "toString"} {
		for _, scenario := range []struct {
			label            string
			required, object bool
		}{
			{"optional-string", false, false}, {"required-string", true, false}, {"required-object", true, true},
		} {
			t.Run(name+"/"+scenario.label, func(t *testing.T) {
				var schema any = map[string]any{"type": "string"}
				if scenario.object {
					schema = map[string]any{"type": "object", "additionalProperties": true, "properties": []any{}}
				}
				model, _ := json.Marshal(map[string]any{"profile": profile, "root": map[string]any{
					"type": "object", "additionalProperties": true, "properties": []any{map[string]any{"name": name, "required": scenario.required, "schema": schema}},
				}})
				quoted, _ := json.Marshal(name)
				cases := []profileCase{
					{`{}`, !scenario.required}, {`{` + string(quoted) + `:"value"}`, !scenario.object},
					{`{` + string(quoted) + `:null}`, false}, {`{` + string(quoted) + `:{}}`, scenario.object},
				}
				checkGeneratedLiterals(t, model, cases)
			})
		}
	}
}

// Independent review literals cover object items, nested arrays and later failures.
func TestGeneratedNestedContainers(t *testing.T) {
	model := []byte(`{"profile":"json-boolean/v1","root":{"type":"array","items":{"type":"object","additionalProperties":true,"properties":[{"name":"flag","required":true,"schema":{"type":"boolean","const":true}},{"name":"groups","required":false,"schema":{"type":"array","items":{"type":"array","items":{"type":"string"}}}}]}}}`)
	cases := []profileCase{
		{`[]`, true},
		{`[{"flag":true}]`, true},
		{`[{"flag":true,"groups":[[],["","x"]],"extra":null}]`, true},
		{`[{"flag":true,"groups":[]},{"flag":true}]`, true},
		{`[{"flag":false}]`, false},
		{`[{}]`, false},
		{`[null]`, false},
		{`[{"flag":true,"groups":[["x",false]]}]`, false},
		{`[{"flag":true,"groups":[null]}]`, false},
		{`[{"flag":true,"groups":null}]`, false},
		{`[{"flag":true},{"flag":"true"}]`, false},
		{`{}`, false},
	}
	checkGeneratedLiterals(t, model, cases)
}
