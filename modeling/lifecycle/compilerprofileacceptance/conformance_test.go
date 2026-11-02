package compilerprofileacceptance_test

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/compiler"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/jsonformat"
)

// Frozen at 4d34bdf before this harness and any new-model reducer execution.
const inventorySHA256 = "e96013d6d9d5d2bf603d3dd473075de47a5d89d4a3de633dc5b483ff767e663c"
const externalPackage = "example.com/compilerprofile/generated"

type fileIdentity struct {
	Path   string `json:"path"`
	Bytes  int    `json:"bytes"`
	SHA256 string `json:"sha256"`
}
type effectSequence struct {
	Transition  string   `json:"transition"`
	Destination string   `json:"destination"`
	Operations  []string `json:"operations"`
}
type fixture struct {
	Name        string           `json:"name"`
	ModelDigest string           `json:"modelDigest"`
	Tests       []string         `json:"tests"`
	Effects     []effectSequence `json:"effects"`
}
type inventory struct {
	Version  string         `json:"version"`
	Files    []fileIdentity `json:"files"`
	Fixtures []fixture      `json:"fixtures"`
}

func read(t *testing.T, path string) []byte {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	b, err := io.ReadAll(io.LimitReader(f, (128<<10)+1))
	if err != nil || len(b) > 128<<10 {
		t.Fatalf("bounded fixture read %s: %v", path, err)
	}
	return b
}
func digest(b []byte) string {
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}
func frozenInventory(t *testing.T) inventory {
	t.Helper()
	b := read(t, "testdata/inventory.json")
	if digest(b) != inventorySHA256 {
		t.Fatal("frozen inventory changed; requires independent refreeze, not expected-output refresh")
	}
	var in inventory
	d := json.NewDecoder(bytes.NewReader(b))
	d.DisallowUnknownFields()
	if err := d.Decode(&in); err != nil {
		t.Fatal(err)
	}
	if in.Version != "compiler-profile-acceptance/v1" || len(in.Fixtures) != 4 {
		t.Fatal("unexpected frozen inventory shape")
	}
	listed := map[string]bool{"inventory.json": true}
	for _, f := range in.Files {
		if !fs.ValidPath(f.Path) || listed[f.Path] {
			t.Fatal("invalid or repeated frozen file", f.Path)
		}
		listed[f.Path] = true
		b := read(t, filepath.Join("testdata", filepath.FromSlash(f.Path)))
		if len(b) != f.Bytes || digest(b) != f.SHA256 {
			t.Fatal("frozen fixture changed", f.Path)
		}
	}
	if err := filepath.WalkDir("testdata", func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		rel, err := filepath.Rel("testdata", path)
		if err != nil || !entry.Type().IsRegular() || !listed[filepath.ToSlash(rel)] {
			return fmt.Errorf("unlisted or nonregular fixture: %s", path)
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	return in
}
func write(t *testing.T, path string, b []byte) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, b, 0600); err != nil {
		t.Fatal(err)
	}
}

// A finite, trusted repository-test runner, not a sandbox for arbitrary models
// or code. Bound diagnostics and wall time; never count those failures as kills.
type boundedOutput struct {
	mu       sync.Mutex
	b        bytes.Buffer
	overflow bool
}

func (b *boundedOutput) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.b.Len()+len(p) > 8<<20 {
		b.overflow = true
		return 0, fmt.Errorf("subprocess output limit")
	}
	return b.b.Write(p)
}
func goCommand(t *testing.T, dir string, args ...string) ([]byte, error) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, filepath.Join(runtime.GOROOT(), "bin", "go"), args...)
	cmd.Dir, cmd.WaitDelay = dir, 2*time.Second
	blocked := map[string]bool{"GOWORK": true, "GOTOOLCHAIN": true, "GOPROXY": true, "GOSUMDB": true, "GOFLAGS": true, "GOENV": true, "GOOS": true, "GOARCH": true}
	for _, entry := range os.Environ() {
		key, _, _ := strings.Cut(entry, "=")
		if !blocked[key] {
			cmd.Env = append(cmd.Env, entry)
		}
	}
	cmd.Env = append(cmd.Env, "GOWORK=off", "GOTOOLCHAIN=local", "GOPROXY=off", "GOSUMDB=off", "GOFLAGS=", "GOENV=off", "GOOS="+runtime.GOOS, "GOARCH="+runtime.GOARCH)
	var out boundedOutput
	cmd.Stdout, cmd.Stderr = &out, &out
	err := cmd.Run()
	if ctx.Err() != nil || out.overflow {
		t.Fatalf("subprocess limit, not oracle evidence: timeout=%v outputLimit=%v", ctx.Err(), out.overflow)
	}
	return out.b.Bytes(), err
}

type testEvent struct {
	Action, Package, Test, Output string
}

func events(t *testing.T, output []byte) []testEvent {
	t.Helper()
	var result []testEvent
	scanner := bufio.NewScanner(bytes.NewReader(output))
	scanner.Buffer(make([]byte, 4096), 1<<20)
	for scanner.Scan() {
		var e testEvent
		if err := json.Unmarshal(scanner.Bytes(), &e); err != nil {
			t.Fatalf("invalid go test event (not behavioral evidence): %v\n%s", err, output)
		}
		if e.Package != externalPackage {
			t.Fatalf("unexpected tested package %q", e.Package)
		}
		result = append(result, e)
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	return result
}
func positive(t *testing.T, root string, f fixture) {
	t.Helper()
	out, err := goCommand(t, root, "test", "-json", "-race", "-count=1", "-buildvcs=false", "./generated")
	if err != nil {
		t.Fatalf("frozen %s oracle failed: %v\n%s", f.Name, err, out)
	}
	if err := positiveEvents(events(t, out), f.Tests); err != nil {
		t.Fatalf("independent test inventory mismatch for %s: %v", f.Name, err)
	}
	t.Logf("%s: all %d frozen test/subtest events passed", f.Name, len(f.Tests))
}
func positiveEvents(observed []testEvent, names []string) error {
	ran, passed := map[string]int{}, map[string]int{}
	packagePass := 0
	for _, e := range observed {
		if e.Action == "skip" || e.Action == "fail" {
			return fmt.Errorf("unexpected non-pass event: %+v", e)
		}
		if e.Action == "pass" && e.Test == "" {
			packagePass++
		}
		if strings.HasPrefix(e.Test, "TestProfile") {
			if e.Action == "run" {
				ran[e.Test]++
			}
			if e.Action == "pass" {
				passed[e.Test]++
			}
		}
	}
	want := map[string]int{}
	for _, name := range names {
		if want[name] != 0 {
			return fmt.Errorf("duplicate frozen test: %s", name)
		}
		want[name] = 1
	}
	if packagePass != 1 || !reflect.DeepEqual(ran, want) || !reflect.DeepEqual(passed, want) {
		return fmt.Errorf("ran=%v passed=%v want=%v packagePass=%d", ran, passed, want, packagePass)
	}
	return nil
}

func TestPositiveEventInventory(t *testing.T) {
	names := []string{"TestProfileExample", "TestProfileExample/case"}
	valid := []testEvent{{Action: "run", Test: names[0]}, {Action: "run", Test: names[1]}, {Action: "pass", Test: names[1]}, {Action: "pass", Test: names[0]}, {Action: "pass"}}
	if err := positiveEvents(valid, names); err != nil {
		t.Fatal(err)
	}
	invalid := map[string][]testEvent{
		"empty":            {},
		"missing_leaf":     {valid[0], valid[3], valid[4]},
		"missing_run":      {valid[0], valid[2], valid[3], valid[4]},
		"missing_package":  valid[:4],
		"duplicate_pass":   append(slices.Clone(valid), valid[2]),
		"unexpected_case":  append(slices.Clone(valid), testEvent{Action: "run", Test: "TestProfileOther"}),
		"skipped_derived":  append(slices.Clone(valid), testEvent{Action: "skip", Test: "TestModelIdentity"}),
		"failed_unrelated": append(slices.Clone(valid), testEvent{Action: "fail", Test: "TestOther"}),
	}
	for name, observed := range invalid {
		t.Run(name, func(t *testing.T) {
			if positiveEvents(observed, names) == nil {
				t.Fatal("accepted incomplete or failing event inventory")
			}
		})
	}
	if positiveEvents(valid, append(slices.Clone(names), names[0])) == nil {
		t.Fatal("accepted duplicate declared test")
	}
}

func printNode(node any) string {
	var b bytes.Buffer
	if err := format.Node(&b, token.NewFileSet(), node); err != nil {
		return "invalid AST: " + err.Error()
	}
	return b.String()
}
func effectOrderError(code []byte, f fixture) error {
	file, err := parser.ParseFile(token.NewFileSet(), "generated.go", code, 0)
	if err != nil {
		return err
	}
	want := map[string][]string{}
	for _, sequence := range f.Effects {
		var body strings.Builder
		for _, op := range sequence.Operations {
			switch op {
			case "assign_owner_if_absent":
				body.WriteString("if next.Owner == nil { v := command.Actor; next.Owner = &v };\n")
			case "set_state":
				fmt.Fprintf(&body, "next.State = %s;\n", strconv.Quote(sequence.Destination))
			case "increment_revision":
				body.WriteString("next.Revision++;\n")
			default:
				return fmt.Errorf("unknown frozen effect %s", op)
			}
		}
		expected, err := parser.ParseFile(token.NewFileSet(), "expected.go", "package expected; func effect(){"+body.String()+"}", 0)
		if err != nil {
			return err
		}
		for _, statement := range expected.Decls[0].(*ast.FuncDecl).Body.List {
			want[sequence.Transition] = append(want[sequence.Transition], printNode(statement))
		}
	}
	got := map[string][]string{}
	switches := 0
	for _, declaration := range file.Decls {
		fn, ok := declaration.(*ast.FuncDecl)
		if !ok || fn.Name.Name != "Transition" {
			continue
		}
		for _, statement := range fn.Body.List {
			sw, ok := statement.(*ast.SwitchStmt)
			if !ok {
				continue
			}
			id, ok := sw.Tag.(*ast.Ident)
			if !ok || id.Name != "id" {
				continue
			}
			switches++
			for _, clause := range sw.Body.List {
				c := clause.(*ast.CaseClause)
				if len(c.List) != 1 {
					return fmt.Errorf("unexpected effect case")
				}
				literal, ok := c.List[0].(*ast.BasicLit)
				if !ok || literal.Kind != token.STRING {
					return fmt.Errorf("nonliteral effect case")
				}
				key, err := strconv.Unquote(literal.Value)
				if err != nil || got[key] != nil {
					return fmt.Errorf("invalid or repeated effect case")
				}
				for _, s := range c.Body {
					got[key] = append(got[key], printNode(s))
				}
			}
		}
	}
	if switches != 1 || !reflect.DeepEqual(got, want) {
		return fmt.Errorf("ordered effect statements differ: got=%v want=%v", got, want)
	}
	return nil
}

type mutation struct {
	Name, Fixture, Function, Old, New, Witness string
	Count                                      int
}

var controls = []mutation{
	{"uint-ne-to-eq", "integers", "notEqualUint", "a.value != b.value", "a.value == b.value", "TestProfileIntegerBoundaries/above_float_precision", 1},
	{"uint-float-narrowing", "integers", "notEqualUint", "a.value != b.value", "float64(a.value) != float64(b.value)", "TestProfileIntegerBoundaries/above_float_precision", 1},
	{"uint32-narrowing", "integers", "notEqualUint", "a.value != b.value", "uint32(a.value) != uint32(b.value)", "TestProfileIntegerBoundaries/above_uint32", 1},
	{"optional-uint-eq-presence", "integers", "equalUint", "a.present && b.present && ", "", "TestProfileIntegerOptionalOperands/reviewEqual_absent", 1},
	{"optional-uint-ne-presence", "integers", "notEqualUint", "a.present && b.present && a.value != b.value", "!a.present || !b.present || a.value != b.value", "TestProfileIntegerOptionalOperands/reviewDifferent_absent", 1},
	{"overflow-before-guards", "integers", "Transition", "if len(failed) > 0 {", "if task.Revision == math.MaxUint64 { return refused(task, id, []string{\"REVISION_LIMIT\"}) }; if len(failed) > 0 {", "TestProfileIntegerBoundaries/guard_precedes_overflow", 1},
	{"optional-string-eq-presence", "presence", "equalString", "a.present && b.present && ", "", "TestProfileOptionalStrings/ownerEmpty_nil", 1},
	{"unicode-trim-removed", "presence", "nonempty", "strings.TrimSpace(a.value)", "a.value", "TestProfileOptionalStrings/ownerText_unicode_space", 1},
	{"all-to-any", "presence", "guard_G_allSixteen", " && ", " || ", "TestProfileBooleanComposition/all_reject_operand_00", 15},
	{"any-to-all", "presence", "guard_G_anySixteen", " || ", " && ", "TestProfileBooleanComposition/any_accept_operand_00", 15},
	{"assignment-hardcoded-start", "dispatch", "Transition", "if next.Owner == nil {", "if next.Owner == nil && command.Type == \"start\" {", "TestProfileDispatch/assign_nonstart", 1},
	{"refusal-order-reversed", "dispatch", "refused", "sort.Strings(reasons)", "sort.Sort(sort.Reverse(sort.StringSlice(reasons)))", "TestProfileDispatch/guard_order_before_overflow", 1},
	{"shallow-task-copy", "dispatch", "cloneTask", "c.Acceptance = append([]string{}, t.Acceptance...)", "c.Acceptance = t.Acceptance", "TestProfileDispatch/assign_nonstart", 1},
	{"reference-misbound", "textdigest", "refString", "return stringValue{t.ID, true}", "return stringValue{t.Outcome, true}", "TestProfileReferenceBindings/id", 1},
	{"digest-revision-number", "textdigest", "TaskDigest", "strconv.FormatUint(t.Revision, 10)", "t.Revision", "TestProfileTaskDigests/base", 1},
}

func mutate(t *testing.T, code []byte, m mutation) []byte {
	t.Helper()
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "generated.go", code, 0)
	if err != nil {
		t.Fatal(err)
	}
	for _, declaration := range file.Decls {
		fn, ok := declaration.(*ast.FuncDecl)
		if !ok || fn.Name.Name != m.Function {
			continue
		}
		start, end := fset.Position(fn.Pos()).Offset, fset.Position(fn.End()).Offset
		part := string(code[start:end])
		if strings.Count(part, m.Old) != m.Count {
			t.Fatal("mutation target changed; review control", m.Name)
		}
		changed := string(code[:start]) + strings.ReplaceAll(part, m.Old, m.New) + string(code[end:])
		out, err := format.Source([]byte(changed))
		if err != nil {
			t.Fatal("invalid mutation syntax, not a kill", m.Name, err)
		}
		return out
	}
	t.Fatal("mutation function not found", m.Function)
	return nil
}
func compile(t *testing.T, root string) {
	t.Helper()
	if out, err := goCommand(t, root, "test", "-c", "-buildvcs=false", "-o", filepath.Join(root, "compiled.test"), "./generated"); err != nil {
		t.Fatalf("mutation does not compile; not behavioral evidence: %v\n%s", err, out)
	}
}
func killed(t *testing.T, root string, m mutation) {
	t.Helper()
	parts := strings.Split(m.Witness, "/")
	for i, p := range parts {
		parts[i] = "^" + regexp.QuoteMeta(p) + "$"
	}
	out, err := goCommand(t, root, "test", "-json", "-count=1", "-buildvcs=false", "-run", strings.Join(parts, "/"), "./generated")
	var exit *exec.ExitError
	if !errors.As(err, &exit) || exit.ExitCode() != 1 {
		t.Fatalf("mutation did not produce ordinary test failure: %s: %v\n%s", m.Name, err, out)
	}
	failed, ran, packageFailed, marked := 0, 0, 0, false
	parent := strings.Split(m.Witness, "/")[0]
	for _, e := range events(t, out) {
		if e.Action == "skip" {
			t.Fatal("mutation witness skipped")
		}
		if e.Action == "fail" && e.Test != "" && e.Test != parent && e.Test != m.Witness {
			t.Fatal("unrelated failed test does not count as kill", e.Test)
		}
		if e.Action == "fail" && e.Test == "" {
			packageFailed++
		}
		if e.Test == m.Witness {
			if e.Action == "fail" {
				failed++
			}
			if e.Action == "run" {
				ran++
			}
			marked = marked || strings.Contains(e.Output, "PROFILE_ORACLE_MISMATCH")
		}
	}
	if failed != 1 || ran != 1 || packageFailed != 1 || !marked {
		t.Fatalf("mutation was not killed by intended named oracle: %s\n%s", m.Name, out)
	}
	t.Logf("%s compiled and failed %s", m.Name, m.Witness)
}

func TestFrozenCompilerProfile(t *testing.T) {
	if runtime.Version() != compiler.ToolchainVersion {
		t.Fatalf("requires actual %s, have %s", compiler.ToolchainVersion, runtime.Version())
	}
	in := frozenInventory(t)
	for _, f := range in.Fixtures {
		t.Run(f.Name, func(t *testing.T) {
			ir, err := jsonformat.Decode(read(t, filepath.Join("testdata", f.Name, "model.json")))
			if err != nil {
				t.Fatal(err)
			}
			code, derived, err := compiler.Generate(ir)
			if err != nil {
				t.Fatal(err)
			}
			if err := effectOrderError(code, f); err != nil {
				t.Fatal(err)
			}
			root := t.TempDir()
			write(t, filepath.Join(root, "go.mod"), []byte("module example.com/compilerprofile\n\ngo 1.27.3\n"))
			target := filepath.Join(root, "generated", "lifecycle.go")
			write(t, target, code)
			write(t, filepath.Join(root, "generated", "derived_test.go"), derived)
			write(t, filepath.Join(root, "generated", "oracle_test.go"), read(t, filepath.Join("testdata", f.Name, "oracle_test.go.txt")))
			positive(t, root, f)
			for _, m := range controls {
				if m.Fixture != f.Name {
					continue
				}
				t.Run(m.Name, func(t *testing.T) {
					if !slices.Contains(f.Tests, m.Witness) {
						t.Fatal("mutation names unfrozen witness", m.Witness)
					}
					write(t, target, mutate(t, code, m))
					defer write(t, target, code)
					compile(t, root)
					killed(t, root, m)
				})
			}
			if f.Name == "dispatch" {
				t.Run("structural-effect-order", func(t *testing.T) {
					m := mutation{Name: "effect-order", Function: "Transition", Old: "next.State = \"ReviewBay\"\n\t\tnext.Revision++", New: "next.Revision++\n\t\tnext.State = \"ReviewBay\"", Count: 2}
					changed := mutate(t, code, m)
					write(t, target, changed)
					defer write(t, target, code)
					compile(t, root)
					if effectOrderError(changed, f) == nil {
						t.Fatal("structural checker missed swapped effect order")
					}
					// These effects commute. Passing behavior is not an order proof.
					positive(t, root, f)
				})
			}
			if !bytes.Equal(read(t, target), code) {
				t.Fatal("generated code was not restored")
			}
			positive(t, root, f)
		})
	}
}
