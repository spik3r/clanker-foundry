package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/acceptanceauthor"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/firstdesignreview"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/processview"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/proposalreplay"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/revisioncheck"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/scenariodiff"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/scenarios"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/strictjson"
)

func bounded(path string, max int) ([]byte, error) {
	f, e := os.Open(path)
	if e != nil {
		return nil, e
	}
	defer f.Close()
	st, e := f.Stat()
	if e != nil {
		return nil, e
	}
	if !st.Mode().IsRegular() {
		return nil, fmt.Errorf("not a regular file: %s", path)
	}
	b, e := io.ReadAll(io.LimitReader(f, int64(max)+1))
	if len(b) > max {
		return nil, fmt.Errorf("input exceeds limit: %s", path)
	}
	return b, e
}
func hash(b []byte) string { return fmt.Sprintf("%x", sha256.Sum256(b)) }
func toJSON(v any) string {
	b, e := json.Marshal(v)
	if e != nil {
		panic(e)
	}
	return string(b)
}

type demo struct {
	revisioncheck.Snapshot
	revision
	Model, Suite, Manifest string
	SuitePath              string
}

func loadDemo(version int) (demo, error) {
	var d demo
	d.Label = fmt.Sprintf("Lifecycle revision %d", version)
	modelPath := "examples/lifecycle.json"
	if version == 2 {
		modelPath = "tutorial/model-v2.json"
	}
	d.SuitePath = fmt.Sprintf("tutorial/suite-v%d.json", version)
	paths := []string{modelPath, d.SuitePath, fmt.Sprintf("tutorial/manifest-v%d.json", version), fmt.Sprintf("revisionreview/testdata/requirements-v%d.json", version), "examples/oracle.json", fmt.Sprintf("revisionreview/testdata/intake-v%d.json", version)}
	fields := []*string{&d.Model, &d.Suite, &d.Manifest, &d.Requirements, &d.Catalog}
	for i, dst := range fields {
		b, e := bounded(paths[i], maxInput)
		if e != nil {
			return d, e
		}
		*dst = string(b)
	}
	b, e := bounded(paths[5], maxInput)
	if e != nil {
		return d, e
	}
	var intake map[string]json.RawMessage
	if e = json.Unmarshal(b, &intake); e != nil {
		return d, e
	}
	intake["observed"] = json.RawMessage(toJSON(map[string]string{"requirements": d.Requirements, "model": d.Model, "scenarioSuite": d.Suite}))
	d.FirstDesignRequest = toJSON(firstdesignreview.Request{
		FormatVersion: firstdesignreview.RequestVersion, IntakeRequest: toJSON(intake),
		ProposalRequest: toJSON(proposalreplay.Request{FormatVersion: proposalreplay.RequestVersion, Model: d.Model, Suite: d.Suite, Manifest: d.Manifest}),
	})
	d.SHA256 = hash([]byte(d.Model))
	return d, nil
}

type stepLiteral struct {
	ID                string
	Command, Expected json.RawMessage
}
type literal struct {
	ModelDigest string
	Cases       []struct {
		ID      string
		Initial json.RawMessage
		Steps   []stepLiteral
	}
}
type replayReport struct {
	FormatVersion   string  `json:"formatVersion"`
	ScenarioVersion string  `json:"scenarioVersion"`
	ProfileVersion  string  `json:"profileVersion"`
	Mode            string  `json:"mode"`
	Authority       string  `json:"authority"`
	ModelDigest     string  `json:"modelDigest"`
	ScenarioDigest  string  `json:"scenarioDigest"`
	SourceDigest    *string `json:"sourceDigest"`
	Matched         bool    `json:"matched"`
	CasesCompleted  int     `json:"casesCompleted"`
	StepsExecuted   int     `json:"stepsExecuted"`
	Entries         []struct {
		CaseID   string          `json:"caseId"`
		StepID   string          `json:"stepId"`
		Input    json.RawMessage `json:"input"`
		Command  json.RawMessage `json:"command"`
		Expected json.RawMessage `json:"expected"`
		Actual   struct {
			Mode         string          `json:"mode"`
			Authority    string          `json:"authority"`
			ModelDigest  string          `json:"modelDigest"`
			OK           bool            `json:"ok"`
			TransitionID string          `json:"transitionId"`
			Task         json.RawMessage `json:"task"`
			Refusals     []string        `json:"refusals"`
		} `json:"actual"`
		Matched     bool     `json:"matched"`
		Differences []string `json:"differences"`
	} `json:"entries"`
}

func equivalent(a, b json.RawMessage) bool {
	decode := func(raw json.RawMessage) (any, error) {
		var v any
		d := json.NewDecoder(strings.NewReader(string(raw)))
		d.UseNumber()
		err := d.Decode(&v)
		return v, err
	}
	x, e := decode(a)
	y, f := decode(b)
	return e == nil && f == nil && reflect.DeepEqual(x, y)
}

type observed struct{ body, status string }

func observations(d demo, s literal, suiteDigest, path string) (map[string]observed, error) {
	seen := map[string]observed{}
	if path == "" || path == "-" {
		return seen, nil
	}
	raw, err := bounded(path, maxInput)
	if err != nil {
		return nil, err
	}
	var a acceptanceauthor.Report
	if err = strictjson.Decode(raw, &a, maxInput, 48); err != nil {
		return nil, err
	}
	if a.FormatVersion != acceptanceauthor.ReportVersion || a.Authority != "none" {
		return nil, fmt.Errorf("unknown acceptance receipt")
	}
	sh, mh := hash([]byte(d.Suite)), hash([]byte(d.Manifest))
	if a.ManifestSHA256 != mh || a.DeclaredSuiteSHA256 != sh || a.ObservedSuiteSHA256 != sh || a.DeclaredModelDigest != s.ModelDigest || a.CurrentModelDigest != s.ModelDigest || a.SuiteModelDigest != s.ModelDigest || a.CompiledModelDigest != s.ModelDigest || a.RunnerBinding != "current" || a.SuiteModelBinding != "current" || a.ModelBinding != "current" || a.SuiteBinding != "current" {
		for _, c := range s.Cases {
			for _, step := range c.Steps {
				seen[c.ID+"/"+step.ID] = observed{status: "stale"}
			}
		}
		return seen, nil
	}
	if a.Replay == "not-run" {
		return seen, nil
	}
	if a.SuiteValidation != "valid" || (a.Replay != "matched" && a.Replay != "mismatched") {
		return nil, fmt.Errorf("invalid replay state")
	}
	var r replayReport
	if err = strictjson.Decode(a.ScenarioReport, &r, maxInput, 48); err != nil {
		return nil, err
	}
	if r.FormatVersion != scenarios.ReportVersion || r.ScenarioVersion != scenarios.ScenarioVersion || r.ProfileVersion != "lifecycle/v1" || r.Mode != "example-only" || r.Authority != "none" || r.ModelDigest != s.ModelDigest || r.SourceDigest == nil || *r.SourceDigest != sh || a.ScenarioDigest == nil || !hashPattern.MatchString(suiteDigest) || *a.ScenarioDigest != suiteDigest || r.ScenarioDigest != suiteDigest || r.CasesCompleted != len(s.Cases) || r.StepsExecuted != len(r.Entries) {
		return nil, fmt.Errorf("unbound scenario report")
	}
	i := 0
	all := true
	for _, c := range s.Cases {
		for stepIndex, step := range c.Steps {
			if i >= len(r.Entries) {
				return nil, fmt.Errorf("incomplete replay")
			}
			n := r.Entries[i]
			i++
			if n.CaseID != c.ID || n.StepID != step.ID || !equivalent(n.Command, step.Command) || !equivalent(n.Expected, step.Expected) || (stepIndex == 0 && !equivalent(n.Input, c.Initial)) || n.Actual.ModelDigest != s.ModelDigest || n.Actual.Mode != "example-only" || n.Actual.Authority != "none" {
				return nil, fmt.Errorf("unbound replay case %s/%s", c.ID, step.ID)
			}
			body := toJSON(map[string]any{"ok": n.Actual.OK, "transitionId": n.Actual.TransitionID, "task": n.Actual.Task, "refusals": n.Actual.Refusals})
			matched := equivalent(json.RawMessage(body), step.Expected)
			if matched != n.Matched {
				return nil, fmt.Errorf("contradictory replay comparison")
			}
			status := "matching"
			if !matched {
				status = "divergent"
				all = false
			}
			seen[c.ID+"/"+step.ID] = observed{pretty(json.RawMessage(body)), status}
		}
	}
	if i != len(r.Entries) || all != r.Matched || all != (a.Replay == "matched") {
		return nil, fmt.Errorf("inconsistent replay summary")
	}
	return seen, nil
}
func pretty(raw json.RawMessage) string {
	b, e := json.MarshalIndent(raw, "", "  ")
	if e != nil {
		return string(raw)
	}
	return string(b)
}
func narrative(input, command, expected json.RawMessage) (string, string, string, error) {
	var state struct{ State, Revision string }
	var trigger struct{ Type, Actor string }
	var outcome struct {
		OK       bool
		Refusals []string
		Task     *struct{ State, Revision string }
	}
	if json.Unmarshal(input, &state) != nil || json.Unmarshal(command, &trigger) != nil || json.Unmarshal(expected, &outcome) != nil || state.State == "" || trigger.Type == "" {
		return "", "", "", fmt.Errorf("cannot describe fixture")
	}
	given := fmt.Sprintf("A task is %s at revision %s.", state.State, state.Revision)
	when := fmt.Sprintf("Actor %q sends %s.", trigger.Actor, trigger.Type)
	if trigger.Actor == "" {
		when = fmt.Sprintf("A blank actor sends %s.", trigger.Type)
	}
	then := "The command is accepted."
	if !outcome.OK {
		then = "The command is refused: " + strings.Join(outcome.Refusals, ", ") + "."
	}
	if outcome.Task != nil {
		if outcome.Task.State == state.State && outcome.Task.Revision == state.Revision {
			then += " The task state and revision stay unchanged."
		} else {
			then += fmt.Sprintf(" The task is %s at revision %s.", outcome.Task.State, outcome.Task.Revision)
		}
	}
	return given, when, then, nil
}
func tutorial(args []string) ([]byte, error) {
	if len(args) != 0 && len(args) != 2 {
		return nil, fmt.Errorf("usage: foundryreview --tutorial [BASELINE_ACCEPTANCE CANDIDATE_ACCEPTANCE]")
	}
	before, e := loadDemo(1)
	if e != nil {
		return nil, e
	}
	after, e := loadDemo(2)
	if e != nil {
		return nil, e
	}
	if before.SHA256 != "a60916ca68771b28ee48712288decad7eb741d6a18e1cbc95b2a196eb10d12d8" || after.SHA256 != "6c6820e6d8544136a5629aa2977aa5319d853a63a5f1363b86766d8204f73499" || hash([]byte(before.Requirements)) != "5b8161cab626cc0b9bd0d56ca44b18158364b5b2724ba0dfaacd134bc90e8cc7" || hash([]byte(after.Requirements)) != "f096c373f5a3c8e8e591b5b39587683f849460cae6ccddbd951482efca7044c3" {
		return nil, fmt.Errorf("tutorial narrative is pinned to the two original model and trace snapshots")
	}
	raw, e := revisioncheck.CheckV3([]byte(toJSON(revisioncheck.Request{FormatVersion: revisioncheck.RequestVersionV3, Before: before.Snapshot, After: after.Snapshot})))
	if e != nil {
		return nil, e
	}
	var report struct {
		revisioncheck.Report
		ModelDiffReportJSON, ModelDiffReportSHA256 string
		ScenarioDiff                               struct{ Status, ReportJSON, ReportSHA256 string }
	}
	if e = json.Unmarshal(raw, &report); e != nil {
		return nil, e
	}
	// Native RawMessage decoding strips framing whitespace, not internal tokens.
	// Receipt/source binding below still uses exact original file bytes.
	tokenSHA := func(text string) string { return hash([]byte(strings.Trim(text, " \t\r\n"))) }
	var suiteDiff scenariodiff.Report
	if report.ScenarioDiff.Status != "checked" || json.Unmarshal([]byte(report.ScenarioDiff.ReportJSON), &suiteDiff) != nil ||
		suiteDiff.Sources == nil || suiteDiff.Sources.BeforeModel != tokenSHA(before.Model) || suiteDiff.Sources.AfterModel != tokenSHA(after.Model) ||
		suiteDiff.Sources.BeforeSuite != tokenSHA(before.Suite) || suiteDiff.Sources.AfterSuite != tokenSHA(after.Suite) {
		return nil, fmt.Errorf("native suite diff is unbound")
	}
	suiteDigests := []string{suiteDiff.BeforeSuiteDigest, suiteDiff.AfterSuiteDigest}
	var p page
	p.FormatVersion, p.Title = "foundry-review-evidence/v1", "Lifecycle model review"
	p.Baseline, p.Candidate = before.revision, after.revision
	p.Changes = []change{{"T_prepare", "Candidate requires G_actor to prepare a draft.", "tutorial/model-v2.json · T_prepare.guards"}}
	p.Requirements = []requirement{{"UR-101", "The transition and guard are linked in the trace. No acceptance case is linked; Q_case_link remains open.", "revisionreview/testdata/requirements-v2.json", "unlinked"}}
	p.Native = []native{
		{"Native v3 revision check (validation only; replay not run)", hash(raw), string(raw)},
		{"Full model diff", report.ModelDiffReportSHA256, report.ModelDiffReportJSON},
		{"Literal suite diff: " + report.ScenarioDiff.Status, report.ScenarioDiff.ReportSHA256, report.ScenarioDiff.ReportJSON},
		{"Requirement trace impact", report.TraceImpactReportSHA256, report.TraceImpactReportJSON},
	}
	for i, d := range []demo{before, after} {
		view, e := processview.Build([]byte(d.Model))
		if e != nil {
			return nil, e
		}
		if i == 0 {
			p.BeforeSVG = svg(view.SVG)
		} else {
			p.AfterSVG = svg(view.SVG)
		}
		var s literal
		if e = json.Unmarshal([]byte(d.Suite), &s); e != nil {
			return nil, e
		}
		path := ""
		if len(args) > 0 {
			path = args[i]
		}
		o, e := observations(d, s, suiteDigests[i], path)
		if e != nil {
			return nil, e
		}
		for _, c := range s.Cases {
			for _, step := range c.Steps {
				key := c.ID + "/" + step.ID
				v := o[key]
				if v.status == "" {
					v.status = "not-run"
				}
				links := []string(nil)
				if c.ID == "blank_actor" || c.ID == "named_actor" {
					links = []string{"T_prepare"}
				}
				given, when, then, e := narrative(c.Initial, step.Command, step.Expected)
				if e != nil {
					return nil, e
				}
				if len(c.Steps) != 1 {
					return nil, fmt.Errorf("tutorial narrative requires its original one-step cases")
				}
				p.Scenarios = append(p.Scenarios, scenario{ID: fmt.Sprintf("v%d.%s.%s", i+1, c.ID, step.ID), Title: d.Label + " / " + strings.ReplaceAll(c.ID, "_", " "), Given: given, When: when, Then: then, Expected: pretty(step.Expected), Observed: v.body, Status: v.status, Source: d.SuitePath + " · " + key, ChangeIDs: links})
			}
		}
	}
	return render(p)
}
