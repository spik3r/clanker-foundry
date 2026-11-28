// foundryreview renders inert, offline review evidence. It does not approve models.
package main

import (
	"bytes"
	_ "embed"
	"encoding/base64"
	"fmt"
	"html/template"
	"os"
	"regexp"
	"strings"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/strictjson"
)

type revision struct {
	Label  string `json:"label"`
	SHA256 string `json:"sha256"`
}
type source struct {
	Source string `json:"source"`
	SHA256 string `json:"sha256"`
}
type change struct {
	ID        string `json:"id"`
	Statement string `json:"statement"`
	Source    string `json:"source"`
}
type requirement struct {
	ID        string `json:"id"`
	Statement string `json:"statement"`
	Source    string `json:"source"`
	Status    string `json:"status"`
}
type scenario struct {
	ID             string   `json:"id"`
	Title          string   `json:"title"`
	Given          string   `json:"given"`
	When           string   `json:"when"`
	Then           string   `json:"then"`
	Expected       string   `json:"expected"`
	Observed       string   `json:"observed"`
	Status         string   `json:"status"`
	Source         string   `json:"source"`
	ChangeIDs      []string `json:"changeIds,omitempty"`
	RequirementIDs []string `json:"requirementIds,omitempty"`
}
type evidence struct {
	FormatVersion       string        `json:"formatVersion"`
	Title               string        `json:"title"`
	Baseline            revision      `json:"baseline"`
	Candidate           revision      `json:"candidate"`
	ScenarioSuite       *source       `json:"scenarioSuite,omitempty"`
	ObservationRevision *revision     `json:"observationRevision,omitempty"`
	Changes             []change      `json:"changes"`
	Requirements        []requirement `json:"requirements"`
	Scenarios           []scenario    `json:"scenarios"`
}
type page struct {
	evidence
	BeforeSVG, AfterSVG template.URL
	Native              []native
	Appearance          appearance
}
type native struct{ Title, SHA256, Body string }

const maxInput = 4 << 20

var hashPattern = regexp.MustCompile("^[0-9a-f]{64}$")
var idPattern = regexp.MustCompile("^[A-Za-z0-9_.:-]{1,120}$")

func validate(e evidence) error {
	if e.FormatVersion != "foundry-review-evidence/v1" || strings.TrimSpace(e.Title) == "" || e.Baseline.Label == "" || e.Candidate.Label == "" || !hashPattern.MatchString(e.Baseline.SHA256) || !hashPattern.MatchString(e.Candidate.SHA256) {
		return fmt.Errorf("invalid review envelope or revision identity")
	}
	changes, requirements, cases := map[string]bool{}, map[string]bool{}, map[string]bool{}
	register := func(m map[string]bool, id string) bool {
		if _, exists := m[id]; !idPattern.MatchString(id) || exists {
			return false
		}
		m[id] = true
		return true
	}
	for _, c := range e.Changes {
		if !register(changes, c.ID) || c.Statement == "" || c.Source == "" {
			return fmt.Errorf("invalid change %q", c.ID)
		}
	}
	for _, r := range e.Requirements {
		if !register(requirements, r.ID) || r.Source == "" || r.Statement == "" || (r.Status != "linked" && r.Status != "unlinked") {
			return fmt.Errorf("invalid requirement %q", r.ID)
		}
		requirements[r.ID] = r.Status == "linked"
	}
	linked := map[string]bool{}
	for _, s := range e.Scenarios {
		if !register(cases, s.ID) || s.Title == "" || s.Given == "" || s.When == "" || s.Then == "" || s.Expected == "" || s.Source == "" {
			return fmt.Errorf("invalid BDD scenario %q", s.ID)
		}
		switch s.Status {
		case "matching", "divergent":
			if s.Observed == "" {
				return fmt.Errorf("missing observation for %q", s.ID)
			}
		case "not-run", "stale":
			if s.Observed != "" {
				return fmt.Errorf("unexpected observation for %q", s.ID)
			}
		default:
			return fmt.Errorf("invalid observation status for %q", s.ID)
		}
		for _, id := range s.ChangeIDs {
			if !changes[id] {
				return fmt.Errorf("unknown change %q", id)
			}
		}
		for _, id := range s.RequirementIDs {
			if !requirements[id] {
				return fmt.Errorf("unknown or unlinked requirement %q", id)
			}
			linked[id] = true
		}
	}
	for id, want := range requirements {
		if want && !linked[id] {
			return fmt.Errorf("no declared case for requirement %q", id)
		}
	}
	return nil
}

// Only the pinned, repository-owned CSS enters this trusted template slot.
// Evidence remains ordinary escaped template data.
//
//go:embed tesserae-tokens.css
var tesseraeTokens string

//go:embed tesserae-LICENSE
var tesseraeLicense string

func licensedTheme() template.CSS {
	if strings.Contains(tesseraeLicense, "*/") ||
		strings.Contains(strings.ToLower(tesseraeTokens+tesseraeLicense), "</style") {
		panic("invalid embedded Tesserae assets")
	}
	return template.CSS("/*\n" + tesseraeLicense + "*/\n" + tesseraeTokens)
}

var layout = template.Must(template.New("review").Funcs(template.FuncMap{
	"themeTokens": licensedTheme,
}).Parse(`<!doctype html>
<html lang="en" data-theme="{{.Appearance.Theme}}" data-mode="{{.Appearance.Mode}}"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width">
<meta http-equiv="Content-Security-Policy" content="default-src 'none'; img-src data:; style-src 'unsafe-inline'; base-uri 'none'; form-action 'none'">
<title>{{.Title}}</title><style>
{{themeTokens}}
body{font:1rem system-ui;max-width:78rem;margin:2rem auto;padding:0 1rem;color:var(--ts-text);background:var(--ts-bg)} section,article{padding:1rem;margin:1rem 0;background:var(--ts-surface);border:1px solid var(--ts-border);border-radius:var(--ts-radius)}
.grid{display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,22rem),1fr));gap:1rem} figure{margin:0;background:#fff;border-radius:var(--ts-radius)} img{display:block;max-width:100%;height:auto} code,pre,h3{overflow-wrap:anywhere;white-space:pre-wrap} dt{font-weight:bold} dd{margin:0 0 .7rem} summary{cursor:pointer} a{color:var(--ts-accent);text-decoration:underline}
:focus-visible{outline:2px solid var(--ts-accent);outline-offset:3px}
</style></head><body><main><h1>{{.Title}}</h1>
<p>Review evidence only. Validation and matching behaviour do not approve the intended design.</p>
<section><h2>Revisions</h2><div class="grid">
<div><h3>Baseline: {{.Baseline.Label}}</h3><code>{{.Baseline.SHA256}}</code>{{if .BeforeSVG}}<figure><img src="{{.BeforeSVG}}" alt="Baseline process projection"></figure>{{end}}</div>
<div><h3>Candidate: {{.Candidate.Label}}</h3><code>{{.Candidate.SHA256}}</code>{{if .AfterSVG}}<figure><img src="{{.AfterSVG}}" alt="Candidate process projection"></figure>{{end}}</div></div>{{if not .BeforeSVG}}<p>No diagram supplied for this review.</p>{{end}}</section>
<section><h2>What changed</h2>{{range .Changes}}<p id="change-{{.ID}}"><strong>{{.ID}}:</strong> {{.Statement}} <br><code>{{.Source}}</code></p>{{else}}<p>No change was supplied in this packet.</p>{{end}}</section>
<section><h2>Requirement trace</h2>{{range .Requirements}}<p id="req-{{.ID}}"><strong>{{.ID}} · {{.Status}}</strong>: {{.Statement}}<br><code>{{.Source}}</code></p>{{else}}<p>No requirement trace was supplied.</p>{{end}}</section>
<section><h2>Acceptance: Given / When / Then</h2>{{if .ScenarioSuite}}<p>Independent suite: <code>{{.ScenarioSuite.Source}} · {{.ScenarioSuite.SHA256}}</code></p>{{end}}{{if .ObservationRevision}}<p>Reported observation revision: <code>{{.ObservationRevision.Label}} · {{.ObservationRevision.SHA256}}</code>. Producer not authenticated.</p>{{end}}{{range .Scenarios}}<article><h3>{{.Title}} · {{.Status}}</h3><p><code>{{.ID}} · {{.Source}}</code></p>
<p>{{range .ChangeIDs}}<a href="#change-{{.}}">Change: {{.}}</a> {{end}}{{range .RequirementIDs}}<a href="#req-{{.}}">Requirement: {{.}}</a> {{else}}No requirement link declared.{{end}}</p>
<dl><dt>Given</dt><dd>{{.Given}}</dd><dt>When</dt><dd>{{.When}}</dd><dt>Then</dt><dd>{{.Then}}</dd></dl>
<div class="grid"><div><details><summary>Expected (literal)</summary><pre>{{.Expected}}</pre></details></div><div>{{if .Observed}}<details><summary>Observed · {{.Status}}</summary><pre>{{.Observed}}</pre></details>{{else}}<h4>Observed</h4><p>No bound observation: {{.Status}}.</p>{{end}}</div></div></article>{{else}}<p>No acceptance scenarios were supplied.</p>{{end}}</section>
{{if .Native}}<section><h2>Native evidence</h2>{{range .Native}}<details><summary>{{.Title}}</summary><p>SHA-256: <code>{{.SHA256}}</code></p><pre>{{.Body}}</pre></details>{{end}}</section>{{end}}
</main></body></html>`))

func svg(data []byte) template.URL {
	// SVG runs only in the browser's inert image context, not as document markup.
	return template.URL("data:image/svg+xml;base64," + base64.StdEncoding.EncodeToString(data))
}

type appearance struct{ Theme, Mode string }

func render(p page) ([]byte, error) {
	if p.Appearance.Theme == "" {
		p.Appearance = appearance{"slate", "auto"}
	}
	if err := validate(p.evidence); err != nil {
		return nil, err
	}
	var b bytes.Buffer
	if err := layout.Execute(&b, p); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}
func run(args []string) ([]byte, error) {
	a := appearance{"slate", "auto"}
	seen := make(map[string]bool)
	for len(args) > 0 && (args[0] == "--theme" || args[0] == "--mode") {
		flag := args[0]
		if len(args) < 2 || seen[flag] {
			return nil, fmt.Errorf("missing or repeated %s", flag)
		}
		seen[flag] = true
		value := args[1]
		switch flag {
		case "--theme":
			if value != "slate" && value != "sand" {
				return nil, fmt.Errorf("theme must be slate or sand")
			}
			a.Theme = value
		case "--mode":
			if value != "auto" && value != "light" && value != "dark" {
				return nil, fmt.Errorf("mode must be auto, light or dark")
			}
			a.Mode = value
		}
		args = args[2:]
	}
	if len(args) > 0 && args[0] == "--tutorial" {
		return tutorial(args[1:], a)
	}
	if len(args) != 1 {
		return nil, fmt.Errorf("usage: foundryreview [--theme slate|sand] [--mode auto|light|dark] (EVIDENCE_JSON | --tutorial [BASELINE_ACCEPTANCE CANDIDATE_ACCEPTANCE])")
	}
	p := page{Appearance: a}
	raw, err := bounded(args[0], maxInput)
	if err != nil {
		return nil, err
	}
	if err := strictjson.Decode(raw, &p.evidence, maxInput, 16); err != nil {
		return nil, err
	}
	if p.ScenarioSuite != nil && (p.ScenarioSuite.Source == "" || !hashPattern.MatchString(p.ScenarioSuite.SHA256)) {
		return nil, fmt.Errorf("invalid suite binding")
	}
	if p.ObservationRevision != nil && *p.ObservationRevision != p.Candidate {
		return nil, fmt.Errorf("observation revision differs from candidate")
	}
	for _, c := range p.Scenarios {
		if (c.Status == "matching" || c.Status == "divergent") && (p.ObservationRevision == nil || p.ScenarioSuite == nil) {
			return nil, fmt.Errorf("executed evidence lacks revision or suite binding")
		}
	}
	return render(p)
}
func main() {
	b, err := run(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, "foundryreview:", err)
		os.Exit(2)
	}
	if _, err = os.Stdout.Write(b); err != nil {
		fmt.Fprintln(os.Stderr, "foundryreview:", err)
		os.Exit(3)
	}
}
