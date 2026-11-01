package impact

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"testing"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/canonical"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/model"
)

func TestOwnedOutputsAndConcurrentComparison(t *testing.T) {
	b, a := tinyModel(), tinyModel()
	a.Requirements[0].Text = "Changed"
	binding := []Binding{{"z", strings.Repeat("0", 64)}, {"a", strings.Repeat("1", 64)}}
	snapshot, _ := json.Marshal([]model.Model{b, a})
	bs, _ := json.Marshal(binding)
	r, err := Compare(b, a, binding)
	if err != nil {
		t.Fatal(err)
	}
	if r.RequestDigest != nil || r.BeforeSourceDigest != nil || r.AfterSourceDigest != nil {
		t.Fatal("invented observed source identity")
	}
	if r.Mode != Mode || r.Authority != Authority || r.ReportVersion != ReportVersion || r.ProfileVersion != model.ProfileVersion {
		t.Fatal("wrong envelope")
	}
	if !reflect.DeepEqual(r.Changes, []Change{{"requirement", "R", "modified", []string{"text"}}}) {
		t.Fatal(r.Changes)
	}
	if !reflect.DeepEqual(r.LinkedTransitions, []LinkedTransition{{"before", "T", []string{"requirement:R"}}, {"after", "T", []string{"requirement:R"}}}) {
		t.Fatal(r.LinkedTransitions)
	}
	var wg sync.WaitGroup
	for i := 0; i < 12; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 8; j++ {
				got, err := Compare(b, a, binding)
				if err != nil || !reflect.DeepEqual(got, r) {
					t.Errorf("nondeterministic comparison: %v", err)
					return
				}
			}
		}()
	}
	wg.Wait()
	r.Changes[0].Fields[0] = "mutated"
	r.LinkedTransitions[0].Causes[0] = "mutated"
	r.Bindings[0].ID = "mutated"
	if r.LinkedTransitions[1].Causes[0] != "requirement:R" {
		t.Fatal("graph outputs alias each other")
	}
	after, _ := json.Marshal([]model.Model{b, a})
	ba, _ := json.Marshal(binding)
	if !bytes.Equal(after, snapshot) || !bytes.Equal(ba, bs) {
		t.Fatal("inputs changed")
	}
	fresh, err := Compare(b, a, binding)
	if err != nil || fresh.Changes[0].Fields[0] != "text" || fresh.Bindings[0].ID != "a" {
		t.Fatal("report aliases internal state", err)
	}
}

func TestCompareRejectsUnsupportedInput(t *testing.T) {
	for name, change := range map[string]func(*model.Model){
		"version":      func(m *model.Model) { m.Version = 2 },
		"unknown ref":  func(m *model.Model) { m.Guards[0].Expr.Args[0].Ref = "host.secret" },
		"duplicate ID": func(m *model.Model) { m.States[1].ID = m.States[0].ID },
		"wrong effect order": func(m *model.Model) {
			m.Transitions[0].Effects[0], m.Transitions[0].Effects[1] = m.Transitions[0].Effects[1], m.Transitions[0].Effects[0]
		},
		"unknown effect": func(m *model.Model) { m.Transitions[0].Effects[0].Op = "execute" },
		"unreachable":    func(m *model.Model) { m.Initial = "Done" },
	} {
		t.Run(name, func(t *testing.T) {
			m := tinyModel()
			change(&m)
			for _, pair := range [][2]model.Model{{m, tinyModel()}, {tinyModel(), m}} {
				r, err := Compare(pair[0], pair[1], nil)
				if err != ErrInvalidInput || !reflect.DeepEqual(r, Report{}) {
					t.Fatal("invalid model returned report", r, err)
				}
			}
		})
	}
	r, err := Compare(tinyModel(), tinyModel(), []Binding{{"invalid-id", strings.Repeat("a", 64)}})
	if err != ErrInvalidInput || !reflect.DeepEqual(r, Report{}) {
		t.Fatal("invalid binding accepted", err)
	}
}

func TestExecuteSourceIdentityAndDeterminism(t *testing.T) {
	raw := wire(t, tinyModel(), tinyModel(), nil)
	first, err := Execute(raw)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 10; i++ {
		next, err := Execute(raw)
		if err != nil || !bytes.Equal(next, first) {
			t.Fatal("byte drift", err)
		}
	}
	var r Report
	if err := json.Unmarshal(first, &r); err != nil {
		t.Fatal(err)
	}
	d, err := decodeRequest(raw)
	if err != nil {
		t.Fatal(err)
	}
	if r.RequestDigest == nil || *r.RequestDigest != *hash(raw) || *r.BeforeSourceDigest != *hash(d.beforeSource) || *r.AfterSourceDigest != *hash(d.afterSource) {
		t.Fatal("source identity incorrect")
	}
	if r.Changed || r.ReviewRequired || r.BeforeDigest != r.AfterDigest || r.Changes == nil || r.LinkedTransitions == nil || r.Bindings == nil {
		t.Fatal("no-op report", r)
	}
	outer := append([]byte(" \n"), raw...)
	encoded, err := Execute(outer)
	if err != nil {
		t.Fatal(err)
	}
	var out Report
	json.Unmarshal(encoded, &out)
	if *r.RequestDigest == *out.RequestDigest || *r.BeforeSourceDigest != *out.BeforeSourceDigest {
		t.Fatal("outer whitespace identity")
	}
	// Preserve only value-level internal whitespace in one embedded model.
	formatted := bytes.Replace(raw, []byte(`"before":{`), []byte("\"before\":{\n  "), 1)
	encoded, err = Execute(formatted)
	if err != nil {
		t.Fatal(err)
	}
	json.Unmarshal(encoded, &out)
	if out.Changed || out.BeforeDigest != r.BeforeDigest || *out.BeforeSourceDigest == *r.BeforeSourceDigest || *out.AfterSourceDigest != *r.AfterSourceDigest {
		t.Fatal("source formatting changed normalized meaning")
	}
	// Report mutation cannot feed Execute: it accepts a complete request only.
	r.Changed = true
	forged, _ := json.Marshal(r)
	if got, err := Execute(forged); err != ErrInvalidInput || got != nil {
		t.Fatal("arbitrary report accepted", err)
	}
	if first[len(first)-1] != '\n' {
		t.Fatal("missing output newline")
	}
}

func TestMaximumDeclaredLinksAndBindingBudget(t *testing.T) {
	longID := func(kind string, i int) string { return fmt.Sprintf("%s%02d", kind, i) + strings.Repeat("x", 44) }
	m := model.Model{Version: 1, ID: "Maximum", Initial: longID("S", 0)}
	for i := 0; i < 16; i++ {
		m.States = append(m.States, model.State{ID: longID("S", i), Label: "State", Terminal: i == 15})
	}
	ids := []string{}
	for i := 0; i < 64; i++ {
		rid, gid := longID("R", i), longID("G", i)
		m.Requirements = append(m.Requirements, model.Requirement{ID: rid, Text: "Rule"})
		m.Guards = append(m.Guards, model.Guard{ID: gid, Requirement: rid, Expr: model.Expr{Kind: "op", Op: "nonempty", Args: []model.Expr{{Kind: "ref", Ref: "task.outcome"}}}})
		ids = append(ids, gid)
	}
	for i := 0; i < 32; i++ {
		from, to := 0, 15
		if i < 15 {
			from, to = i, i+1
		}
		m.Transitions = append(m.Transitions, model.Transition{ID: longID("T", i), From: longID("S", from), To: longID("S", to), Trigger: fmt.Sprintf("event%d", i), Guards: append([]string{}, ids...), Effects: []model.Effect{{Op: model.SetState}, {Op: model.IncrementRevision}}})
	}
	a, err := model.Normalize(m)
	if err != nil {
		t.Fatal(err)
	}
	for i := range a.Requirements {
		a.Requirements[i].Text = "Changed rule"
		a.Guards[i].Expr.Args[0].Ref = "command.actor"
	}
	for i := range a.States {
		a.States[i].Label = "Changed state"
	}
	bindings := []Binding{}
	for i := 0; i < MaxBindings; i++ {
		bindings = append(bindings, Binding{fmt.Sprintf("Scenario%d", i), strings.Repeat("0", 64)})
	}
	r, err := Compare(m, a, bindings)
	if err != nil {
		t.Fatal(err)
	}
	if len(r.LinkedTransitions) != 64 || len(r.LinkedTransitions[0].Causes) != 130 || len(r.Bindings) != MaxBindings {
		t.Fatal("maximum graph not fully represented")
	}
	raw := wire(t, m, a, bindings)
	out, err := Execute(raw)
	if err != nil || len(out) > MaxReportBytes {
		t.Fatal("valid maximum graph exceeds report budget", len(out), err)
	}
	bd, _ := canonical.Digest(m)
	ad, _ := canonical.Digest(a)
	if r.BeforeDigest != bd || r.AfterDigest != ad {
		t.Fatal("wrong model identity")
	}
}

// Mutate only disposable source copies. Independent literal expectations are
// copied unchanged; a compiler/setup error is not credited as a killed mutant.
func TestIndependentOracleKillsOmissionMutants(t *testing.T) {
	for name, needle := range map[string]string{
		"requirement propagation": `add("requirement", guards[id].Requirement)`,
		"removed side links":      `r.LinkedTransitions = append(r.LinkedTransitions, links("before", b, changed)...)`,
	} {
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			copyFile := func(src, dst string) {
				t.Helper()
				raw, err := os.ReadFile(src)
				if err != nil {
					t.Fatal(err)
				}
				if err := os.MkdirAll(filepath.Dir(dst), 0700); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(dst, raw, 0600); err != nil {
					t.Fatal(err)
				}
			}
			copyFile("../go.mod", filepath.Join(dir, "go.mod"))
			for _, pkg := range []string{"model", "canonical", "strictjson", "jsonformat", "impact", "impactacceptance"} {
				entries, err := os.ReadDir("../" + pkg)
				if err != nil {
					t.Fatal(err)
				}
				for _, entry := range entries {
					name := entry.Name()
					if entry.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") && pkg != "impactacceptance" {
						continue
					}
					copyFile(filepath.Join("..", pkg, name), filepath.Join(dir, pkg, name))
				}
			}
			path := filepath.Join(dir, "impact", "compare.go")
			raw, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if strings.Count(string(raw), needle) != 1 {
				t.Fatal("mutation target drift")
			}
			mutated := strings.Replace(string(raw), needle, "// deliberately omitted in scratch mutant", 1)
			if err := os.WriteFile(path, []byte(mutated), 0600); err != nil {
				t.Fatal(err)
			}
			cmd := exec.Command(filepath.Join(runtime.GOROOT(), "bin", "go"), "test", "-count=1", "./impactacceptance")
			cmd.Dir = dir
			cmd.Env = append(os.Environ(), "GOTOOLCHAIN=local", "GOPROXY=off", "GOSUMDB=off")
			out, err := cmd.CombinedOutput()
			if err == nil || !bytes.Contains(out, []byte("--- FAIL:")) || bytes.Contains(out, []byte("[build failed]")) {
				t.Fatalf("mutant not killed behaviorally: %v\n%s", err, out)
			}
		})
	}
}
