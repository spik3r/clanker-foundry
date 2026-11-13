package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	tui "github.com/spik3r/clanker-foundry/modeling/tui"
)

const (
	browserModel = "../../examples/browser.json"
	browserCases = "../../examples/browser.cases.json"
	packageModel = "../../examples/browser-package.json"
	packageCases = "../../examples/browser-package.cases.json"
	hostedModel  = "../../examples/browser-hosted.json"
	hostedCases  = "../../examples/browser-hosted.cases.json"
	filterModel  = "../../examples/browser-filter.json"
	filterCases  = "../../examples/browser-filter.cases.json"
)

func invoke(args ...string) (code int, stdout, stderr string) {
	var out, errb bytes.Buffer
	code = run(args, &out, &errb)
	return code, out.String(), errb.String()
}

func skipShort(t *testing.T) {
	t.Helper()
	if testing.Short() {
		t.Skip("tuicases builds and runs generated Go")
	}
}

func write(t *testing.T, name, content string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(p, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return p
}

func readCases(t *testing.T, path string) tui.Cases {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	c, err := tui.DecodeCases(raw)
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func writeCases(t *testing.T, c tui.Cases) string {
	t.Helper()
	raw, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	return write(t, "cases.json", string(raw))
}

func str(s string) *string { return &s }
func yes(b bool) *bool     { return &b }

func digestOf(t *testing.T, model string) string {
	t.Helper()
	raw, err := os.ReadFile(model)
	if err != nil {
		t.Fatal(err)
	}
	res, err := tui.Build(raw)
	if err != nil {
		t.Fatal(err)
	}
	return tui.CanonicalSHA256(res.Files)
}

func TestExampleCasesPass(t *testing.T) {
	skipShort(t)
	for _, p := range [][2]string{{browserModel, browserCases}, {packageModel, packageCases}, {hostedModel, hostedCases}, {filterModel, filterCases}} {
		code, stdout, stderr := invoke("check", p[0], p[1])
		if code != 0 || stderr != "" {
			t.Fatalf("%s: exit %d\n%s", p[1], code, stderr)
		}
		c := readCases(t, p[1])
		var want []string
		for _, k := range c.Cases {
			want = append(want, "pass "+k.ID)
		}
		if got := strings.TrimSpace(stdout); got != strings.Join(want, "\n") {
			t.Fatalf("%s: stdout\n%s\nwant\n%s", p[1], got, strings.Join(want, "\n"))
		}
	}
}

// A wrong expectation of each kind fails with the right message, and the other
// cases of the same run are still reported.
func TestWrongExpectationsFailWithExit3(t *testing.T) {
	skipShort(t)
	base := readCases(t, browserCases)
	mk := func(id string, e tui.Expect) tui.Case {
		return tui.Case{ID: id, Requirement: "A deliberately wrong expectation.", Size: tui.Size{Width: 100, Height: 20},
			Records: "Main", Steps: []tui.Step{}, Expect: e}
	}
	two := mk("W_two", tui.Expect{Layout: str("tiny"), Focus: str("detail")})
	base.Cases = []tui.Case{
		mk("W_ok", tui.Expect{Layout: str("wide"), Focus: str("list"), Help: yes(false), Selected: str("a"), Status: str("ready"), Quit: yes(false),
			Visible: &[]string{"Name: Alpha"}, Absent: &[]string{"Name: Beta"}}),
		mk("W_layout", tui.Expect{Layout: str("narrow")}),
		mk("W_focus", tui.Expect{Focus: str("detail")}),
		mk("W_help", tui.Expect{Help: yes(true)}),
		mk("W_selected", tui.Expect{Selected: str("b")}),
		mk("W_status", tui.Expect{Status: str("failed")}),
		mk("W_visible", tui.Expect{Visible: &[]string{"Name: Alpha", "Nonexistent text"}}),
		mk("W_absent", tui.Expect{Absent: &[]string{"Name: Beta", "Name: Alpha"}}),
		mk("W_quit", tui.Expect{Quit: yes(true)}),
		two,
		mk("W_last_ok", tui.Expect{Selected: str("a")}),
	}
	code, stdout, stderr := invoke("check", browserModel, writeCases(t, base))
	if code != 3 || stdout != "" {
		t.Fatalf("exit %d, stdout %q\n%s", code, stdout, stderr)
	}
	want := []string{
		"pass W_ok",
		`fail W_layout: layout: want "narrow" got "wide"`,
		`fail W_focus: focus: want "detail" got "list"`,
		`fail W_help: help: want true got false`,
		`fail W_selected: selected: want "b" got "a"`,
		`fail W_status: status: want "failed" got "ready"`,
		`fail W_visible: visible "Nonexistent text": want shown got not shown`,
		`fail W_absent: absent "Name: Alpha": want not shown got shown`,
		`fail W_quit: quit: want true got false`,
		`fail W_two: layout: want "tiny" got "wide"`,
		`fail W_two: focus: want "detail" got "list"`,
		"pass W_last_ok",
	}
	if got := strings.Split(strings.TrimSpace(stderr), "\n"); strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Fatalf("stderr\n%s\nwant\n%s", stderr, strings.Join(want, "\n"))
	}
}

// capturesKeys and keyHints are observed in both embed modes, and a wrong
// expectation of either kind fails with its own message.
func TestHostedExpectationsReport(t *testing.T) {
	skipShort(t)
	base := readCases(t, browserCases)
	mk := func(id string, steps []tui.Step, e tui.Expect) tui.Case {
		return tui.Case{ID: id, Requirement: "An expectation about keys.", Size: tui.Size{Width: 100, Height: 20},
			Records: "Main", Steps: steps, Expect: e}
	}
	help := []tui.Step{{Key: str("?")}}
	hints := func(h ...string) *[]tui.KeyHint {
		var out []tui.KeyHint
		for i := 0; i < len(h); i += 2 {
			out = append(out, tui.KeyHint{Key: h[i], Action: h[i+1]})
		}
		return &out
	}
	base.Cases = []tui.Case{
		mk("K_ok", []tui.Step{}, tui.Expect{CapturesKeys: yes(false), KeyHints: hints("j/k", "select", "?", "help", "q", "quit")}),
		mk("K_help_ok", help, tui.Expect{CapturesKeys: yes(true), KeyHints: hints("esc/?", "close")}),
		mk("K_captures", []tui.Step{}, tui.Expect{CapturesKeys: yes(true)}),
		mk("K_order", []tui.Step{}, tui.Expect{KeyHints: hints("?", "help", "j/k", "select")}),
		mk("K_missing", []tui.Step{}, tui.Expect{KeyHints: hints("x", "nothing")}),
	}
	code, stdout, stderr := invoke("check", browserModel, writeCases(t, base))
	if code != 3 || stdout != "" {
		t.Fatalf("exit %d, stdout %q\n%s", code, stdout, stderr)
	}
	for _, want := range []string{"pass K_ok\n", "pass K_help_ok\n", "fail K_captures: capturesKeys: want true got false\n",
		`fail K_order: keyHints "j/k" "select": want shown in order got [j/k select, tab pane, enter open, ? help, q quit]`,
		`fail K_missing: keyHints "x" "nothing": want shown in order got [`} {
		if !strings.Contains(stderr, want) {
			t.Errorf("stderr lacks %q:\n%s", want, stderr)
		}
	}
}

// A model change that alters behavior, with the digest re-bound, makes old
// expectations fail; without re-binding the cases are refused as stale.
func TestModelChangeIsCaughtByCases(t *testing.T) {
	skipShort(t)
	raw, err := os.ReadFile(browserModel)
	if err != nil {
		t.Fatal(err)
	}
	for name, edit := range map[string]struct {
		replace [2]string
		failing []string
	}{
		"breakpoints": {[2]string{`"wideWidth": 80`, `"wideWidth": 120`}, []string{
			`fail C_start_first_selected: layout: want "wide" got "narrow"`,
			`fail C_start_first_selected: visible "Details"`,
		}},
		"details": {[2]string{`"details": ["Name", "Status", "Description"]`, `"details": ["Status", "Description"]`}, []string{
			`fail C_down_up_move_selection: visible "Name: Beta": want shown got not shown`,
			`fail C_start_first_selected: visible "Name: Alpha": want shown got not shown`,
		}},
	} {
		t.Run(name, func(t *testing.T) {
			changed := strings.Replace(string(raw), edit.replace[0], edit.replace[1], 1)
			if changed == string(raw) {
				t.Fatal("test setup: edit did not apply")
			}
			model := write(t, "model.json", changed)
			c := readCases(t, browserCases)
			code, stdout, stderr := invoke("check", model, browserCases)
			if code != 2 || stdout != "" || !strings.Contains(stderr, "stale") {
				t.Fatalf("stale cases: exit %d, stdout %q, stderr %q", code, stdout, stderr)
			}
			c.Model.CanonicalSHA256 = digestOf(t, model)
			code, stdout, stderr = invoke("check", model, writeCases(t, c))
			if code != 3 || stdout != "" {
				t.Fatalf("exit %d, stdout %q\n%s", code, stdout, stderr)
			}
			for _, line := range edit.failing {
				if !strings.Contains(stderr, line) {
					t.Errorf("missing %q in\n%s", line, stderr)
				}
			}
			if !strings.Contains(stderr, "pass C_quit_with_q") {
				t.Errorf("unaffected case not reported:\n%s", stderr)
			}
		})
	}
}

func TestRefusalsExit2WithEmptyStdout(t *testing.T) {
	good := readCases(t, browserCases)
	wrongField := good
	wrongField.Records = map[string][]tui.Values{"Main": {{"ID": "a", "Name": "Alpha"}}}
	extraField := good
	extraField.Records = map[string][]tui.Values{"Main": {{"ID": "a", "Name": "n", "Description": "d", "Status": "s", "Extra": "x"}}}
	dupIdentity := good
	dupIdentity.Records = map[string][]tui.Values{"Main": {{"ID": "a", "Name": "n", "Description": "d", "Status": "s"}, {"ID": "a", "Name": "n", "Description": "d", "Status": "s"}}}
	stale := good
	stale.Model.CanonicalSHA256 = strings.Repeat("0", 64)
	for name, args := range map[string][]string{
		"no arguments":         {},
		"unknown command":      {"run", browserModel, browserCases},
		"missing cases arg":    {"check", browserModel},
		"missing model file":   {"check", "/nonexistent/model.json", browserCases},
		"missing cases file":   {"validate", browserModel, "/nonexistent/cases.json"},
		"bad model":            {"validate", write(t, "m.json", `{"formatVersion":"tui-json/v1"}`), browserCases},
		"bad cases":            {"validate", browserModel, write(t, "c.json", `{"formatVersion":"record-browser-cases/v1"}`)},
		"stale digest":         {"validate", browserModel, writeCases(t, stale)},
		"missing field":        {"validate", browserModel, writeCases(t, wrongField)},
		"extra field":          {"validate", browserModel, writeCases(t, extraField)},
		"duplicate identity":   {"validate", browserModel, writeCases(t, dupIdentity)},
		"cases of other model": {"validate", packageModel, browserCases},
	} {
		code, stdout, stderr := invoke(args...)
		if code != 2 || stdout != "" || stderr == "" {
			t.Errorf("%s: exit %d, stdout %q, stderr %q", name, code, stdout, stderr)
		}
	}
	_, _, stderr := invoke("validate", packageModel, browserCases)
	if !strings.Contains(stderr, "stale") {
		t.Errorf("cases of another model are not reported as stale: %s", stderr)
	}
}

// validate needs neither go nor any execution: it works with an empty PATH.
func TestValidateNeedsNoGoToolchain(t *testing.T) {
	t.Setenv("PATH", "")
	for _, p := range [][2]string{{browserModel, browserCases}, {packageModel, packageCases}, {hostedModel, hostedCases}, {filterModel, filterCases}} {
		code, stdout, stderr := invoke("validate", p[0], p[1])
		if code != 0 || stderr != "" || !strings.HasPrefix(stdout, "ok ") {
			t.Fatalf("exit %d, stdout %q, stderr %q", code, stdout, stderr)
		}
	}
	// check without go is an internal failure, not a panic or a refusal.
	code, stdout, stderr := invoke("check", browserModel, browserCases)
	if code != 1 || stdout != "" || !strings.Contains(stderr, "go build") {
		t.Fatalf("check without go: exit %d, stdout %q, stderr %q", code, stdout, stderr)
	}
}

func TestRunnerSourceHoldsNoCaseData(t *testing.T) {
	for _, p := range [][2]string{{browserModel, browserCases}, {packageModel, packageCases}, {hostedModel, hostedCases}} {
		l, err := load(p[0], p[1])
		if err != nil {
			t.Fatal(err)
		}
		src, err := renderRunner(l, "example.com/host/internal/x")
		if err != nil {
			t.Fatal(err)
		}
		text := string(src)
		if !strings.Contains(text, `browser "example.com/host/internal/x"`) || !strings.Contains(text, `Description: v["Description"]`) {
			t.Fatalf("import path or fields missing:\n%s", text)
		}
		for _, k := range l.cases.Cases {
			for _, s := range []string{k.ID, k.Requirement} {
				if strings.Contains(text, s) {
					t.Fatalf("runner source contains case data %q", s)
				}
			}
		}
		for _, set := range l.cases.Records {
			for _, rec := range set {
				if v := rec["Description"]; len(v) > 8 && strings.Contains(text, v) {
					t.Fatalf("runner source contains record data %q", v)
				}
			}
		}
	}
}

func TestGeneratorNeverRunsGeneratedCode(t *testing.T) {
	// tuic and the root package stay free of process execution; only tuicases
	// builds and runs generated code.
	for _, dir := range []string{"../..", "../tuic"} {
		files, _ := filepath.Glob(filepath.Join(dir, "*.go"))
		for _, f := range files {
			if strings.HasSuffix(f, "_test.go") {
				continue
			}
			raw, err := os.ReadFile(f)
			if err != nil {
				t.Fatal(err)
			}
			if strings.Contains(string(raw), `"os/exec"`) {
				t.Errorf("%s imports os/exec", f)
			}
		}
	}
}

// A wrong expectation of each filter kind fails with its own message, and a
// text step is delivered one key message per code point.
func TestWrongFilterExpectationsFail(t *testing.T) {
	skipShort(t)
	base := readCases(t, filterCases)
	n3 := 3
	mk := func(id string, steps []tui.Step, e tui.Expect) tui.Case {
		return tui.Case{ID: id, Requirement: "A filter expectation.", Size: tui.Size{Width: 100, Height: 20},
			Records: "Main", Steps: steps, Expect: e}
	}
	typed := []tui.Step{{Key: str("/")}, {Text: str("ser")}}
	base.Cases = []tui.Case{
		mk("F_ok", typed, tui.Expect{Filter: str("ser"), FilterOpen: yes(true), Shown: func() *int { n := 2; return &n }()}),
		mk("F_filter", typed, tui.Expect{Filter: str("se")}),
		mk("F_open", typed, tui.Expect{FilterOpen: yes(false)}),
		mk("F_shown", typed, tui.Expect{Shown: &n3}),
		mk("F_backspace", []tui.Step{{Key: str("/")}, {Text: str("sex")}, {Key: str("backspace")}}, tui.Expect{Filter: str("sex")}),
	}
	code, stdout, stderr := invoke("check", filterModel, writeCases(t, base))
	if code != 3 || stdout != "" {
		t.Fatalf("exit %d, stdout %q\n%s", code, stdout, stderr)
	}
	for _, want := range []string{"pass F_ok\n", `fail F_filter: filter: want "se" got "ser"`, "fail F_open: filterOpen: want false got true",
		"fail F_shown: shown: want 3 got 2", `fail F_backspace: filter: want "sex" got "se"`} {
		if !strings.Contains(stderr, want) {
			t.Errorf("stderr lacks %q:\n%s", want, stderr)
		}
	}
}

// Steps, keys and expectations that need a filter are refused, at validate
// time, for a model without one.
func TestFilterCasesAreRefusedWithoutAFilter(t *testing.T) {
	base := readCases(t, browserCases)
	zero := 0
	for name, c := range map[string]tui.Case{
		"text step":  {Steps: []tui.Step{{Text: str("a")}}, Expect: tui.Expect{Selected: str("a")}},
		"slash key":  {Steps: []tui.Step{{Key: str("/")}}, Expect: tui.Expect{Selected: str("a")}},
		"backspace":  {Steps: []tui.Step{{Key: str("backspace")}}, Expect: tui.Expect{Selected: str("a")}},
		"filter":     {Steps: []tui.Step{}, Expect: tui.Expect{Filter: str("")}},
		"filterOpen": {Steps: []tui.Step{}, Expect: tui.Expect{FilterOpen: yes(false)}},
		"shown":      {Steps: []tui.Step{}, Expect: tui.Expect{Shown: &zero}},
	} {
		c.ID, c.Requirement, c.Size, c.Records = "X", "Needs a filter.", tui.Size{Width: 100, Height: 20}, "Main"
		doc := base
		doc.Cases = []tui.Case{c}
		code, stdout, stderr := invoke("validate", browserModel, writeCases(t, doc))
		if code != 2 || stdout != "" || !strings.Contains(stderr, "model with a filter") {
			t.Errorf("%s: exit %d, stdout %q, stderr %q", name, code, stdout, stderr)
		}
	}
}
