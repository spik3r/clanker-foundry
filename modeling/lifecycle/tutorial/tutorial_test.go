// Package tutorial exercises the documented tools in a disposable module copy.
package tutorial

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/acceptanceauthor"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/compiler"
)

type task struct {
	State    string `json:"state"`
	Revision string `json:"revision"`
}
type actual struct {
	OK         bool     `json:"ok"`
	Transition string   `json:"transitionId"`
	Task       *task    `json:"task"`
	Refusals   []string `json:"refusals"`
}
type scenarioReport struct {
	Matched bool `json:"matched"`
	Cases   int  `json:"casesCompleted"`
	Steps   int  `json:"stepsExecuted"`
	Entries []struct {
		CaseID string `json:"caseId"`
		Actual actual `json:"actual"`
	} `json:"entries"`
}

func TestTutorialEndToEnd(t *testing.T) {
	if runtime.Version() != compiler.ToolchainVersion {
		t.Fatalf("requires %s, got %s", compiler.ToolchainVersion, runtime.Version())
	}
	originalSuite, err := os.ReadFile("suite-v2.json")
	if err != nil {
		t.Fatal(err)
	}
	temp := t.TempDir()
	root := filepath.Join(temp, "lifecycle")
	if err = os.Mkdir(root, 0700); err != nil {
		t.Fatal(err)
	}
	if err = os.CopyFS(root, os.DirFS("..")); err != nil {
		t.Fatal(err)
	}
	bin := filepath.Join(temp, "bin")
	if err = os.Mkdir(bin, 0700); err != nil {
		t.Fatal(err)
	}
	goexe := filepath.Join(runtime.GOROOT(), "bin", "go")
	env := append(os.Environ(), "GOTOOLCHAIN=local", "GOPROXY=off", "GOSUMDB=off", "GOWORK=off", "GOENV=off", "GOFLAGS=-buildvcs=false")
	run := func(exe string, args []string, input []byte, want int) []byte {
		t.Helper()
		cmd := exec.Command(exe, args...)
		cmd.Dir = root
		cmd.Env = env
		cmd.Stdin = bytes.NewReader(input)
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		err := cmd.Run()
		code := 0
		if err != nil {
			if e, ok := err.(*exec.ExitError); ok {
				code = e.ExitCode()
			} else {
				t.Fatal(err)
			}
		}
		if code != want {
			t.Fatalf("%s %v: exit %d want %d; stderr %s; stdout %s", exe, args, code, want, stderr.String(), stdout.String())
		}
		return stdout.Bytes()
	}
	build := func(name, pkg string) string {
		t.Helper()
		path := filepath.Join(bin, name)
		run(goexe, []string{"build", "-o", path, pkg}, nil, 0)
		return path
	}
	read := func(path string) []byte {
		t.Helper()
		b, e := os.ReadFile(filepath.Join(root, path))
		if e != nil {
			t.Fatal(e)
		}
		return b
	}
	pack := func(v any) []byte {
		t.Helper()
		b, e := json.Marshal(v)
		if e != nil {
			t.Fatal(e)
		}
		return b
	}
	model1 := read("examples/lifecycle.json")
	model2 := read("tutorial/model-v2.json")
	suite1 := read("tutorial/suite-v1.json")
	suite2 := read("tutorial/suite-v2.json")
	manifest1 := read("tutorial/manifest-v1.json")
	manifest2 := read("tutorial/manifest-v2.json")
	intakeBin := build("modelintake", "./cmd/modelintake")
	accept1 := build("modelacceptance-v1", "./cmd/modelacceptance")
	play1 := build("modelplay-v1", "./cmd/modelplay")
	modelc := build("modelc", "./cmd/modelc")
	run(modelc, []string{"check"}, nil, 0)
	makeIntake := func(version string, model, suite []byte) []byte {
		t.Helper()
		var d map[string]any
		if e := json.Unmarshal(read("tutorial/intake-"+version+".json"), &d); e != nil {
			t.Fatal(e)
		}
		d["observed"] = map[string]any{"requirements": string(read("tutorial/requirements-" + version + ".md")), "model": string(model), "scenarioSuite": string(suite)}
		return pack(d)
	}
	var intake struct {
		Bindings []struct{ Kind, Status string } `json:"bindings"`
		Links    []struct{ Status string }       `json:"links"`
	}
	decode := func(raw []byte, out any) {
		t.Helper()
		if e := json.Unmarshal(raw, out); e != nil {
			t.Fatal(e)
		}
	}
	decode(run(intakeBin, nil, makeIntake("v1", model1, suite1), 0), &intake)
	for _, b := range intake.Bindings {
		if b.Status != "matches-declaration" {
			t.Fatalf("initial binding: %+v", b)
		}
	}
	if len(intake.Bindings) != 3 || len(intake.Links) != 3 {
		t.Fatal("missing intake roles")
	}
	for _, l := range intake.Links {
		if l.Status != "matches-declaration" {
			t.Fatalf("initial link: %+v", l)
		}
	}
	request := func(manifest, model, suite []byte) []byte {
		return pack(acceptanceauthor.Request{FormatVersion: acceptanceauthor.RequestVersion, Mode: "replay", Manifest: string(manifest), Model: string(model), Suite: string(suite)})
	}
	checkReplay := func(raw []byte, blankOK bool) {
		t.Helper()
		var outer struct {
			Replay          string         `json:"replay"`
			SuiteValidation string         `json:"suiteValidation"`
			Scenario        scenarioReport `json:"scenarioReport"`
		}
		decode(raw, &outer)
		if outer.Replay != "matched" || outer.SuiteValidation != "valid" || !outer.Scenario.Matched || outer.Scenario.Cases != 3 || outer.Scenario.Steps != 3 || len(outer.Scenario.Entries) != 3 {
			t.Fatalf("unexpected replay: %+v", outer)
		}
		a := outer.Scenario.Entries[0].Actual
		if blankOK {
			if !a.OK || a.Transition != "T_prepare" || a.Task == nil || a.Task.State != "Ready" || a.Task.Revision != "1" || len(a.Refusals) != 0 {
				t.Fatalf("v1 blank actor: %+v", a)
			}
		} else {
			if a.OK || a.Transition != "T_prepare" || a.Task == nil || a.Task.State != "Draft" || a.Task.Revision != "0" || len(a.Refusals) != 1 || a.Refusals[0] != "G_actor" {
				t.Fatalf("v2 blank actor: %+v", a)
			}
		}
		named := outer.Scenario.Entries[1].Actual
		if !named.OK || named.Task == nil || named.Task.State != "Ready" || named.Task.Revision != "1" {
			t.Fatalf("named actor: %+v", named)
		}
		bad := outer.Scenario.Entries[2].Actual
		if bad.OK || bad.Task != nil || bad.Transition != "" || len(bad.Refusals) != 1 || bad.Refusals[0] != "INVALID_INPUT" {
			t.Fatalf("unknown command: %+v", bad)
		}
	}
	checkReplay(run(accept1, nil, request(manifest1, model1, suite1), 0), true)
	var direct scenarioReport
	decode(run(play1, nil, suite1, 0), &direct)
	if !direct.Matched || direct.Steps != 3 {
		t.Fatalf("initial direct replay: %+v", direct)
	}
	decode(run(intakeBin, nil, makeIntake("v1", model2, suite1), 0), &intake)
	if intake.Bindings[1].Kind != "model" || intake.Bindings[1].Status != "digest-mismatch" {
		t.Fatalf("old intake should detect changed model: %+v", intake.Bindings)
	}
	var stale struct {
		ModelBinding    string          `json:"modelBinding"`
		SuiteValidation string          `json:"suiteValidation"`
		Replay          string          `json:"replay"`
		Report          json.RawMessage `json:"scenarioReport"`
	}
	decode(run(accept1, nil, request(manifest1, model2, suite1), 1), &stale)
	if stale.ModelBinding != "stale" || stale.SuiteValidation != "not-checked" || stale.Replay != "not-run" || string(stale.Report) != "null" {
		t.Fatalf("stale acceptance: %+v", stale)
	}
	if err = os.WriteFile(filepath.Join(root, "examples/lifecycle.json"), model2, 0600); err != nil {
		t.Fatal(err)
	}
	run(modelc, []string{"generate"}, nil, 0)
	run(modelc, []string{"check"}, nil, 0)
	accept2 := build("modelacceptance-v2", "./cmd/modelacceptance")
	play2 := build("modelplay-v2", "./cmd/modelplay")
	run(play2, nil, suite1, 2)
	decode(run(intakeBin, nil, makeIntake("v2", model2, suite2), 0), &intake)
	for _, b := range intake.Bindings {
		if b.Status != "matches-declaration" {
			t.Fatalf("revised binding: %+v", b)
		}
	}
	for _, l := range intake.Links {
		if l.Status != "matches-declaration" {
			t.Fatalf("revised link: %+v", l)
		}
	}
	checkReplay(run(accept2, nil, request(manifest2, model2, suite2), 0), false)
	decode(run(play2, nil, suite2, 0), &direct)
	if !direct.Matched || direct.Steps != 3 {
		t.Fatalf("revised direct replay: %+v", direct)
	}
	now, err := os.ReadFile("suite-v2.json")
	if err != nil || !bytes.Equal(now, originalSuite) {
		t.Fatal("source expectation changed")
	}
}
