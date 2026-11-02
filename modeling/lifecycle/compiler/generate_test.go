package compiler_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"testing"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/compiler"
	lifecycle "github.com/spik3r/clanker-foundry/modeling/lifecycle/generated"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/jsonformat"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/model"
)

func read(t *testing.T, p string) []byte {
	t.Helper()
	b, e := os.ReadFile(p)
	if e != nil {
		t.Fatal(e)
	}
	return b
}
func write(t *testing.T, p string, b []byte) {
	t.Helper()
	if e := os.MkdirAll(filepath.Dir(p), 0755); e != nil {
		t.Fatal(e)
	}
	if e := os.WriteFile(p, b, 0644); e != nil {
		t.Fatal(e)
	}
}
func example(t *testing.T) model.Model {
	t.Helper()
	m, e := jsonformat.Decode(read(t, "../examples/lifecycle.json"))
	if e != nil {
		t.Fatal(e)
	}
	return m
}
func goRun(t *testing.T, dir string, ok bool, args ...string) string {
	t.Helper()
	c := exec.Command(filepath.Join(runtime.GOROOT(), "bin", "go"), args...)
	c.Dir = dir
	c.Env = append(os.Environ(), "GOTOOLCHAIN=local", "GOPROXY=off", "GOSUMDB=off")
	b, e := c.CombinedOutput()
	if (e == nil) != ok {
		t.Fatalf("go %v: %v\n%s", args, e, b)
	}
	return string(b)
}
func generatedModule(t *testing.T, m model.Model) string {
	t.Helper()
	d := t.TempDir()
	code, tests, e := compiler.Generate(m)
	if e != nil {
		t.Fatal(e)
	}
	write(t, filepath.Join(d, "go.mod"), []byte("module github.com/spik3r/clanker-foundry/modeling/lifecycle\n\ngo 1.27.3\n"))
	write(t, filepath.Join(d, "generated/lifecycle.go"), code)
	write(t, filepath.Join(d, "generated/model_paths_test.go"), tests)
	return d
}
func ref(s string) model.Expr                 { return model.Expr{Kind: "ref", Ref: s} }
func lit(s string) model.Expr                 { return model.Expr{Kind: "string", Value: s} }
func op(s string, a ...model.Expr) model.Expr { return model.Expr{Kind: "op", Op: s, Args: a} }
func tiny() model.Model {
	return model.Model{Version: 1, ID: "Example", Initial: "Draft", States: []model.State{{ID: "Draft", Label: "Draft"}, {ID: "Done", Label: "Done", Terminal: true}}, Requirements: []model.Requirement{{ID: "R", Text: "A named actor."}}, Guards: []model.Guard{{ID: "G", Requirement: "R", Expr: op("eq", ref("command.actor"), lit("Alice"))}}, Transitions: []model.Transition{{ID: "T", From: "Draft", Trigger: "go", To: "Done", Guards: []string{"G"}, Effects: []model.Effect{{Op: model.SetState}, {Op: model.IncrementRevision}}}}}
}
func TestGeneratorDeterminismAndNonmutation(t *testing.T) {
	m := example(t)
	before, _ := json.Marshal(m)
	a, b, e := compiler.Generate(m)
	if e != nil {
		t.Fatal(e)
	}
	c, d, e := compiler.Generate(m)
	if e != nil || !bytes.Equal(a, c) || !bytes.Equal(b, d) {
		t.Fatal("unstable generation", e)
	}
	after, _ := json.Marshal(m)
	if !bytes.Equal(before, after) {
		t.Fatal("input changed")
	}
	m.Guards[0].Expr = op("execute", lit("arbitrary"))
	if _, _, e := compiler.Generate(m); e == nil {
		t.Fatal("unknown operator accepted")
	}
}
func TestCustomLiteralAndHonestDerivedProperties(t *testing.T) {
	d := generatedModule(t, tiny())
	write(t, filepath.Join(d, "generated/custom_test.go"), []byte(`package lifecycle
import "testing"
func TestIndependentNamedActor(t *testing.T){task:=Task{ID:"x",State:"Draft"};if Transition(task,Command{Type:"go",Actor:"Bob"}).OK{t.Fatal("Bob admitted")};r:=Transition(task,Command{Type:"go",Actor:"Alice"});if !r.OK||r.Task.State!="Done"||r.Task.Revision!=1{t.Fatalf("Alice refused: %+v",r)}}
`))
	goRun(t, d, true, "test", "-race", "./...")
}
func TestUnderscoreTerminalTestNamesAreUnambiguous(t *testing.T) {
	m := tiny()
	m.States = []model.State{{ID: "Draft", Label: "Draft"}, {ID: "A_B", Label: "A B", Terminal: true}, {ID: "A", Label: "A", Terminal: true}}
	m.Transitions[0].To = "A_B"
	m.Transitions[0].Trigger = "C"
	other := m.Transitions[0]
	other.ID = "Other"
	other.To = "A"
	other.Trigger = "B_C"
	m.Transitions = append(m.Transitions, other)
	goRun(t, generatedModule(t, m), true, "test", "./...")
}
func TestGeneratedLiteralEscaping(t *testing.T) {
	m := tiny()
	value := "Alice\"\n<>&雪\u2028"
	m.Guards[0].Expr = op("eq", ref("command.actor"), lit(value))
	d := generatedModule(t, m)
	body := fmt.Sprintf("package lifecycle\nimport \"testing\"\nfunc TestLiteral(t *testing.T){if !Transition(Task{ID:\"x\",State:\"Draft\"},Command{Type:\"go\",Actor:%q}).OK{t.Fatal(\"literal changed\")}}\n", value)
	write(t, filepath.Join(d, "generated/literal_test.go"), []byte(body))
	goRun(t, d, true, "test", "./...")
}
func TestIndependentOracleKillsAlwaysRefuse(t *testing.T) {
	d := generatedModule(t, example(t))
	write(t, filepath.Join(d, "acceptance/oracle_test.go"), read(t, "../acceptance/oracle_test.go"))
	write(t, filepath.Join(d, "examples/oracle.json"), read(t, "../examples/oracle.json"))
	p := filepath.Join(d, "generated/lifecycle.go")
	code := string(read(t, p))
	needle := "func Transition(task Task, command Command) Result {"
	if strings.Count(code, needle) != 1 {
		t.Fatal("mutation location not unique")
	}
	code = strings.Replace(code, needle, needle+"\n if len(task.ID)>0 {return refused(task,\"\",[]string{\"MUTANT\"})}\n", 1)
	write(t, p, []byte(code))
	out := goRun(t, d, false, "test", "./acceptance")
	if !strings.Contains(out, "--- FAIL:") || strings.Contains(out, "build failed") {
		t.Fatal("mutation must die behaviorally", out)
	}
}
func TestPredictedModelChange(t *testing.T) {
	m := example(t)
	for i := range m.Transitions {
		if m.Transitions[i].ID == "T_prepare" {
			m.Transitions[i].Guards = append(m.Transitions[i].Guards, "G_owner")
		}
	}
	d := generatedModule(t, m)
	write(t, filepath.Join(d, "generated/change_test.go"), []byte(`package lifecycle
import "testing"
func TestPredictedOwnerGate(t *testing.T){task:=Task{ID:"x",State:"Draft",Outcome:"Outcome",Acceptance:[]string{"Criterion"}};r:=Transition(task,Command{Type:"prepare"});if r.OK||len(r.Refusals)!=1||r.Refusals[0]!="G_owner"{t.Fatalf("predicted refusal absent: %+v",r)};owner:="Alice";task.Owner=&owner;if !Transition(task,Command{Type:"prepare"}).OK{t.Fatal("named owner should pass")}}
`))
	goRun(t, d, true, "test", "./...")
}

var outputs = []string{"decision-table.json", "decision.svg", "lifecycle.go", "model_paths_test.go", "trace.json"}

type stamp struct {
	Bytes string
	Time  int64
	Mode  os.FileMode
}

func snapshot(t *testing.T, dir string) map[string]stamp {
	t.Helper()
	r := map[string]stamp{}
	e := filepath.WalkDir(dir, func(p string, x os.DirEntry, e error) error {
		if e != nil {
			return e
		}
		if x.IsDir() {
			return nil
		}
		s, e := os.Lstat(p)
		if e != nil {
			return e
		}
		rel, _ := filepath.Rel(dir, p)
		var b []byte
		if s.Mode()&os.ModeSymlink != 0 {
			v, e := os.Readlink(p)
			if e != nil {
				return e
			}
			b = []byte(v)
		} else {
			b, e = os.ReadFile(p)
			if e != nil {
				return e
			}
		}
		r[rel] = stamp{string(b), s.ModTime().UnixNano(), s.Mode()}
		return nil
	})
	if e != nil {
		t.Fatal(e)
	}
	return r
}
func TestCLIReproducibilityAndWriteFrontier(t *testing.T) {
	bin := filepath.Join(t.TempDir(), "modelc")
	root, e := filepath.Abs("..")
	if e != nil {
		t.Fatal(e)
	}
	goRun(t, root, true, "build", "-o", bin, "./cmd/modelc")
	run := func(d string, ok bool, args ...string) {
		t.Helper()
		c := exec.Command(bin, args...)
		c.Dir = d
		b, e := c.CombinedOutput()
		if (e == nil) != ok {
			t.Fatalf("modelc %v: %v\n%s", args, e, b)
		}
	}
	fresh := func() string {
		d := t.TempDir()
		write(t, filepath.Join(d, "go.mod"), read(t, "../go.mod"))
		write(t, filepath.Join(d, "examples/lifecycle.json"), read(t, "../examples/lifecycle.json"))
		return d
	}
	unchanged := func(d string, before map[string]stamp) {
		t.Helper()
		if !reflect.DeepEqual(before, snapshot(t, d)) {
			t.Fatal("refused/check operation wrote files")
		}
	}
	t.Run("missing_no_write", func(t *testing.T) {
		d := fresh()
		before := snapshot(t, d)
		run(d, true, "validate")
		run(d, false, "check")
		unchanged(d, before)
	})
	t.Run("repeat_and_semantic_reorder", func(t *testing.T) {
		d := fresh()
		run(d, true, "generate")
		before := snapshot(t, d)
		run(d, true, "check")
		run(d, true, "generate")
		unchanged(d, before)
		m := example(t)
		for i, j := 0, len(m.States)-1; i < j; i, j = i+1, j-1 {
			m.States[i], m.States[j] = m.States[j], m.States[i]
		}
		b, e := jsonformat.Encode(m)
		if e != nil {
			t.Fatal(e)
		}
		write(t, filepath.Join(d, "examples/lifecycle.json"), b)
		before = snapshot(t, d)
		run(d, true, "check")
		run(d, true, "generate")
		unchanged(d, before)
	})
	for _, name := range outputs {
		t.Run("modified_"+name, func(t *testing.T) {
			d := fresh()
			run(d, true, "generate")
			p := filepath.Join(d, "generated", name)
			write(t, p, append(read(t, p), ' '))
			before := snapshot(t, d)
			run(d, false, "check")
			run(d, false, "generate")
			unchanged(d, before)
		})
		t.Run("missing_"+name, func(t *testing.T) {
			d := fresh()
			run(d, true, "generate")
			if e := os.Remove(filepath.Join(d, "generated", name)); e != nil {
				t.Fatal(e)
			}
			before := snapshot(t, d)
			run(d, false, "check")
			run(d, false, "generate")
			unchanged(d, before)
		})
	}
	for _, name := range []string{"old.go", "old.svg", "large-old.svg", ".lifecycle-stale.tmp"} {
		t.Run("stale_"+name, func(t *testing.T) {
			d := fresh()
			run(d, true, "generate")
			content := []byte("unowned Go collision")
			if strings.HasSuffix(name, ".svg") {
				content = read(t, filepath.Join(d, "generated/decision.svg"))
				if name == "large-old.svg" {
					content = append(content, bytes.Repeat([]byte(" "), 1<<20)...)
				}
			}
			write(t, filepath.Join(d, "generated", name), content)
			before := snapshot(t, d)
			run(d, false, "check")
			run(d, false, "generate")
			unchanged(d, before)
		})
	}
	t.Run("symlink_and_collision", func(t *testing.T) {
		d := fresh()
		sentinel := filepath.Join(t.TempDir(), "sentinel")
		write(t, sentinel, []byte("unchanged"))
		if e := os.Mkdir(filepath.Join(d, "generated"), 0755); e != nil {
			t.Fatal(e)
		}
		if e := os.Symlink(sentinel, filepath.Join(d, "generated/lifecycle.go")); e != nil {
			t.Fatal(e)
		}
		before := snapshot(t, d)
		run(d, false, "generate")
		run(d, false, "check")
		unchanged(d, before)
		if string(read(t, sentinel)) != "unchanged" {
			t.Fatal("symlink target changed")
		}
		if e := os.Remove(filepath.Join(d, "generated/lifecycle.go")); e != nil {
			t.Fatal(e)
		}
		write(t, filepath.Join(d, "generated/lifecycle.go"), []byte("unowned collision"))
		before = snapshot(t, d)
		run(d, false, "generate")
		unchanged(d, before)
	})
	t.Run("predicted_change_all_views_and_unowned_preservation", func(t *testing.T) {
		d := fresh()
		run(d, true, "generate")
		write(t, filepath.Join(d, "generated/notes.txt"), []byte("keep"))
		before := snapshot(t, d)
		m := example(t)
		for i := range m.Transitions {
			if m.Transitions[i].ID == "T_prepare" {
				m.Transitions[i].Guards = append(m.Transitions[i].Guards, "G_owner")
			}
		}
		b, e := jsonformat.Encode(m)
		if e != nil {
			t.Fatal(e)
		}
		write(t, filepath.Join(d, "examples/lifecycle.json"), b)
		run(d, false, "check")
		run(d, true, "generate")
		run(d, true, "check")
		after := snapshot(t, d)
		for _, n := range outputs {
			if before["generated/"+n].Bytes == after["generated/"+n].Bytes {
				t.Fatal("model change did not propagate", n)
			}
		}
		if before["generated/notes.txt"] != after["generated/notes.txt"] {
			t.Fatal("unowned file changed")
		}
	})
	t.Run("arguments_and_wrong_directory", func(t *testing.T) {
		d := fresh()
		before := snapshot(t, d)
		run(d, false, "generate", "--out", "elsewhere")
		unchanged(d, before)
		run(t.TempDir(), false, "generate")
	})
}

// Shared immutable inputs can be evaluated concurrently; each result owns its data.
func TestConcurrentPureTransitions(t *testing.T) {
	task := lifecycle.Task{ID: "parallel", State: "Draft", Outcome: "Outcome", Acceptance: []string{"Criterion"}}
	cmd := lifecycle.Command{Type: "prepare"}
	var wg sync.WaitGroup
	for range 16 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range 100 {
				r := lifecycle.Transition(task, cmd)
				if !r.OK || r.Task == nil || r.Task.Revision != 1 {
					t.Errorf("transition: %+v", r)
					return
				}
				r.Task.Acceptance[0] = "isolated"
			}
		}()
	}
	wg.Wait()
	if task.Revision != 0 || task.Acceptance[0] != "Criterion" {
		t.Fatal("input mutated")
	}
}
func BenchmarkExampleTransition(b *testing.B) {
	task := lifecycle.Task{ID: "bench", State: "Draft", Outcome: "Outcome", Acceptance: []string{"Criterion"}}
	cmd := lifecycle.Command{Type: "prepare"}
	b.ReportAllocs()
	for b.Loop() {
		r := lifecycle.Transition(task, cmd)
		if !r.OK {
			b.Fatal(r.Refusals)
		}
	}
}
func BenchmarkCompileExample(b *testing.B) {
	raw, e := os.ReadFile("../examples/lifecycle.json")
	if e != nil {
		b.Fatal(e)
	}
	m, e := jsonformat.Decode(raw)
	if e != nil {
		b.Fatal(e)
	}
	b.ReportAllocs()
	for b.Loop() {
		if _, _, e := compiler.Generate(m); e != nil {
			b.Fatal(e)
		}
	}
}

// Every mutant must compile and fail the independent behavioral oracle, not just
// disagree with compiler-derived snapshots.
func TestIndependentOracleKillsEachDemoGuardBypass(t *testing.T) {
	m := example(t)
	code, _, e := compiler.Generate(m)
	if e != nil {
		t.Fatal(e)
	}
	for _, g := range m.Guards {
		t.Run(g.ID, func(t *testing.T) {
			d := generatedModule(t, m)
			re := regexp.MustCompile(`(?s)func guard_` + regexp.QuoteMeta(g.ID) + `\(task Task, command Command, currentDigest string\) bool \{.*?\n\}`)
			if len(re.FindAll(code, -1)) != 1 {
				t.Fatal("mutation target not unique")
			}
			mutant := re.ReplaceAll(code, []byte("func guard_"+g.ID+"(task Task, command Command, currentDigest string) bool {return true}"))
			write(t, filepath.Join(d, "generated/lifecycle.go"), mutant)
			write(t, filepath.Join(d, "acceptance/oracle_test.go"), read(t, "../acceptance/oracle_test.go"))
			write(t, filepath.Join(d, "examples/oracle.json"), read(t, "../examples/oracle.json"))
			out := goRun(t, d, false, "test", "./acceptance")
			if !strings.Contains(out, "--- FAIL:") || strings.Contains(out, "build failed") {
				t.Fatal("mutant must fail behaviorally", out)
			}
		})
	}
}
