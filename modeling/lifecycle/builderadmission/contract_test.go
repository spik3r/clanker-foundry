package builderadmission

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/builderpreflight"
)

// This harness compiles before Check exists. All candidate contents are inert
// source bytes. Neither native proof nor future callback invokes a compiler.
type frozenEntry struct {
	Name string `json:"name"`
	Kind string `json:"kind"`
}
type frozenCase struct {
	ID, Candidate, Status, Outcome, Remove, ChangeOriginal, CandidateKind, ExpectedPin string
	WrongPin                                                                           bool
	Extras                                                                             []frozenEntry
}
type frozenObservations struct {
	ManifestSHA256 string
	AllowedNames   []string
	Cases          []frozenCase
}

func frozenSHA(b []byte) string { d := sha256.Sum256(b); return hex.EncodeToString(d[:]) }
func frozenData(t *testing.T) (frozenObservations, []byte) {
	t.Helper()
	b, e := os.ReadFile("testdata/observations.json")
	if e != nil {
		t.Fatal(e)
	}
	if frozenSHA(b) != "29bf0e0edda998fc987f1cec4fca08f039efa7ae74f3548ab2751657be2308a1" {
		t.Fatal("independent observations changed")
	}
	var d frozenObservations
	if json.Unmarshal(b, &d) != nil {
		t.Fatal("observations")
	}
	m, e := os.ReadFile("inputs.json")
	if e != nil || frozenSHA(m) != d.ManifestSHA256 {
		t.Fatal("independent manifest changed")
	}
	return d, m
}
func frozenRoot(t *testing.T) string {
	t.Helper()
	root := filepath.Join(t.TempDir(), "exercise")
	if os.Mkdir(root, 0700) != nil {
		t.Fatal("mkdir")
	}
	for _, name := range []string{"README.md", "go.mod", "contract.go", "acceptance_test.go", "rubric.md"} {
		b, e := os.ReadFile(filepath.Join("../../../examples/model-builder-go/acceptance", name))
		if e != nil {
			t.Fatal(e)
		}
		if os.WriteFile(filepath.Join(root, name), b, 0600) != nil {
			t.Fatal("copy")
		}
	}
	return root
}
func TestIndependentFiveOriginalPins(t *testing.T) {
	d, m := frozenData(t)
	root := frozenRoot(t)
	b, e := builderpreflight.Check(m, root, d.ManifestSHA256)
	if e != nil {
		t.Fatal(e)
	}
	var r struct {
		Outcome   string
		Artifacts []any
	}
	if json.Unmarshal(b, &r) != nil || r.Outcome != "matched" || len(r.Artifacts) != 5 {
		t.Fatal("native original-pin proof")
	}
	if _, e := os.Stat(filepath.Join(root, "candidate.go")); !os.IsNotExist(e) {
		t.Fatal("proof must not create a candidate")
	}
}
func frozenCall(t *testing.T, check func(string, string) ([]byte, error), root, pin string) ([]byte, error) {
	t.Helper()
	type result struct {
		b []byte
		e error
	}
	done := make(chan result, 1)
	go func() { b, e := check(root, pin); done <- result{b, e} }()
	select {
	case r := <-done:
		return r.b, r.e
	case <-time.After(2 * time.Second):
		t.Fatal("closed census/read must not block on a special leaf")
		return nil, nil
	}
}
func frozenExpectedFrontier(d frozenObservations, c frozenCase) []frozenEntry {
	var out []frozenEntry
	for _, n := range d.AllowedNames {
		if n != c.Remove {
			k := "regular"
			if n == "candidate.go" && c.CandidateKind != "" {
				k = c.CandidateKind
			}
			out = append(out, frozenEntry{n, k})
		}
	}
	out = append(out, c.Extras...)
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}
func assertFrozenAdmission(t *testing.T, check func(string, string) ([]byte, error)) {
	t.Helper()
	d, m := frozenData(t)
	for _, c := range d.Cases {
		t.Run(c.ID, func(t *testing.T) {
			root := frozenRoot(t)
			candidate := filepath.Join(root, "candidate.go")
			if c.CandidateKind == "symlink" {
				outside := filepath.Join(filepath.Dir(root), "outside.txt")
				if os.WriteFile(outside, []byte(c.Candidate), 0600) != nil || os.Symlink(outside, candidate) != nil {
					t.Fatal("candidate link fixture")
				}
			} else if c.Remove != "candidate.go" {
				if os.WriteFile(candidate, []byte(c.Candidate), 0600) != nil {
					t.Fatal("candidate fixture")
				}
			}
			for _, x := range c.Extras {
				p := filepath.Join(root, x.Name)
				switch x.Kind {
				case "directory":
					if os.Mkdir(p, 0700) != nil {
						t.Fatal("extra dir")
					}
					if os.WriteFile(filepath.Join(p, "untouched"), []byte("do not descend"), 0000) != nil {
						t.Fatal("nested fixture")
					}
				case "symlink":
					if os.Symlink("candidate.go", p) != nil {
						t.Fatal("extra link")
					}
				default:
					content := []byte("inert unexpected input")
					if x.Name == "unreviewed.go" {
						content = []byte("package acceptance\nfunc init(){panic(\"unreviewed initializer must never run\")}\n")
					}
					if os.WriteFile(p, content, 0600) != nil {
						t.Fatal("extra fixture")
					}
				}
			}
			if c.Remove != "" && c.Remove != "candidate.go" {
				if os.Remove(filepath.Join(root, c.Remove)) != nil {
					t.Fatal("remove fixture")
				}
			}
			if c.ChangeOriginal != "" {
				p := filepath.Join(root, c.ChangeOriginal)
				b, _ := os.ReadFile(p)
				if os.WriteFile(p, append(b, '\n'), 0600) != nil {
					t.Fatal("change fixture")
				}
			}
			native, e := builderpreflight.Check(m, root, d.ManifestSHA256)
			if e != nil {
				t.Fatal(e)
			}
			before, e := os.ReadDir(root)
			if e != nil {
				t.Fatal(e)
			}
			raw, e := frozenCall(t, check, root, c.ExpectedPin)
			if e != nil {
				t.Fatal(e)
			}
			var got struct {
				FormatVersion, Outcome, Authority, ExecutionStatus, ReviewStatus, ExerciseRoot, ExpectedCandidateSHA256, FrozenManifestSHA256, PreflightReportSHA256, PreflightReportJSON string
				FrontierComplete                                                                                                                                                          bool
				Frontier                                                                                                                                                                  []frozenEntry
				MissingNames, UnexpectedNames                                                                                                                                             []string
				Candidate                                                                                                                                                                 struct {
					Path, Status   string
					ObservedSHA256 *string
					ObservedBytes  *int64
				}
			}
			if json.Unmarshal(raw, &got) != nil || got.FormatVersion != "model-builder-admission-report/v1" || got.Outcome != c.Outcome || got.Authority != "none" || got.ExecutionStatus != "not-run" || got.ReviewStatus != "not-established" || !got.FrontierComplete {
				t.Fatal("admission result", string(raw))
			}
			if got.ExerciseRoot != root || got.ExpectedCandidateSHA256 != c.ExpectedPin || got.FrozenManifestSHA256 != d.ManifestSHA256 || got.PreflightReportJSON != string(native) || got.PreflightReportSHA256 != frozenSHA(native) {
				t.Fatal("original source/native full report bytes changed")
			}
			if !reflect.DeepEqual(got.Frontier, frozenExpectedFrontier(d, c)) {
				t.Fatal("complete sorted frontier", got.Frontier)
			}
			missing := []string{}
			if c.Remove != "" {
				missing = append(missing, c.Remove)
			}
			unexpected := []string{}
			for _, x := range c.Extras {
				unexpected = append(unexpected, x.Name)
			}
			sort.Strings(unexpected)
			if !reflect.DeepEqual(got.MissingNames, missing) || !reflect.DeepEqual(got.UnexpectedNames, unexpected) {
				t.Fatal("missing/unexpected names", got.MissingNames, got.UnexpectedNames)
			}
			if got.Candidate.Path != "candidate.go" || got.Candidate.Status != c.Status {
				t.Fatal("candidate observation", got.Candidate.Status)
			}
			if c.Status == "matched" || c.Status == "hash-mismatch" {
				h := frozenSHA([]byte(c.Candidate))
				n := int64(len([]byte(c.Candidate)))
				if got.Candidate.ObservedSHA256 == nil || *got.Candidate.ObservedSHA256 != h || got.Candidate.ObservedBytes == nil || *got.Candidate.ObservedBytes != n {
					t.Fatal("candidate observed identity")
				}
			} else if got.Candidate.ObservedSHA256 != nil || got.Candidate.ObservedBytes != nil {
				t.Fatal("unread candidate has fabricated identity")
			}
			after, e := os.ReadDir(root)
			if e != nil || len(after) != len(before) {
				t.Fatal("checker changed files")
			}
			again, e := check(root, c.ExpectedPin)
			if e != nil || !bytes.Equal(raw, again) {
				t.Fatal("check changed inputs or output")
			}
		})
	}
	t.Run("malformed-arguments", func(t *testing.T) {
		root := frozenRoot(t)
		for _, pin := range []string{"", strings.Repeat("A", 64), strings.Repeat("a", 63), strings.Repeat("a", 65)} {
			if out, e := check(root, pin); e == nil || out != nil {
				t.Fatal("invalid candidate pin")
			}
		}
		for _, p := range []string{"relative", "", root + "\n", string([]byte{0xff})} {
			if out, e := check(p, strings.Repeat("a", 64)); e == nil || out != nil {
				t.Fatal("invalid root")
			}
		}
	})
	t.Run("root-symlink-even-with-trailing-separator", func(t *testing.T) {
		root := frozenRoot(t)
		alias := filepath.Join(filepath.Dir(root), "alias")
		if os.Symlink(root, alias) != nil {
			t.Fatal("root link")
		}
		for _, p := range []string{alias, alias + string(os.PathSeparator)} {
			if out, e := check(p, strings.Repeat("a", 64)); e == nil || out != nil {
				t.Fatal("root leaf alias accepted")
			}
		}
	})
	t.Run("ancestor-alias-dotdot-uses-host-traversal", func(t *testing.T) {
		parent := t.TempDir()
		target := filepath.Join(parent, "real", "child")
		if os.MkdirAll(target, 0700) != nil {
			t.Fatal("target")
		}
		alias := filepath.Join(parent, "link")
		if os.Symlink(target, alias) != nil {
			t.Fatal("alias")
		}
		for _, name := range []string{"README.md", "go.mod", "contract.go", "acceptance_test.go", "rubric.md"} {
			b, e := os.ReadFile(filepath.Join("../../../examples/model-builder-go/acceptance", name))
			if e != nil || os.WriteFile(filepath.Join(parent, "real", name), b, 0600) != nil {
				t.Fatal("original")
			}
		}
		text := []byte("host traversal opaque candidate")
		if os.WriteFile(filepath.Join(parent, "real", "candidate.go"), text, 0600) != nil {
			t.Fatal("candidate")
		}
		spelling := alias + string(os.PathSeparator) + ".."
		out, e := check(spelling, frozenSHA(text))
		if e != nil {
			t.Fatal(e)
		}
		// The physical target contains child/, so its census is blocked. Its five
		// originals and candidate still match; lexical cleaning selects parent,
		// where those inputs are absent. This distinguishes actual traversal.
		var r struct {
			Outcome, ExerciseRoot, PreflightReportJSON string
			Candidate                                  struct{ Status string }
		}
		if json.Unmarshal(out, &r) != nil || r.Outcome != "blocked" || r.ExerciseRoot != spelling || r.Candidate.Status != "matched" {
			t.Fatal("path was lexically cleaned", string(out))
		}
		var native struct{ Outcome string }
		if json.Unmarshal([]byte(r.PreflightReportJSON), &native) != nil || native.Outcome != "matched" {
			t.Fatal("actual traversal lost the pinned originals")
		}
	})
	t.Run("oversized-candidate-refused", func(t *testing.T) {
		root := frozenRoot(t)
		if os.WriteFile(filepath.Join(root, "candidate.go"), bytes.Repeat([]byte("x"), (1<<20)+1), 0600) != nil {
			t.Fatal("large candidate")
		}
		out, e := check(root, strings.Repeat("a", 64))
		if e != nil {
			t.Fatal(e)
		}
		var r struct {
			Outcome   string
			Candidate struct {
				Status         string
				ObservedSHA256 *string
			}
		}
		if json.Unmarshal(out, &r) != nil || r.Outcome != "blocked" || r.Candidate.Status != "refused" || r.Candidate.ObservedSHA256 != nil {
			t.Fatal("oversize observation")
		}
	})
	t.Run("incomplete-census-limit", func(t *testing.T) {
		root := frozenRoot(t)
		for n := 0; n < 65; n++ {
			name := filepath.Join(root, strings.Repeat("x", n+1))
			if os.WriteFile(name, nil, 0600) != nil {
				t.Fatal("extra")
			}
		}
		if out, e := check(root, strings.Repeat("a", 64)); e == nil || out != nil {
			t.Fatal("oversize census claimed complete")
		}
	})
	t.Run("fifo-leaves-never-opened", func(t *testing.T) {
		if _, e := exec.LookPath("mkfifo"); e != nil {
			t.Skip("host lacks mkfifo fixture utility")
		}
		for _, name := range []string{"unexpected.pipe", "candidate.go"} {
			root := frozenRoot(t)
			if b, e := exec.Command("mkfifo", filepath.Join(root, name)).CombinedOutput(); e != nil {
				t.Fatalf("fifo fixture %v %s", e, b)
			}
			raw, e := frozenCall(t, check, root, strings.Repeat("a", 64))
			if e != nil {
				t.Fatal(e)
			}
			var r struct {
				Outcome  string
				Frontier []frozenEntry
			}
			if json.Unmarshal(raw, &r) != nil || r.Outcome != "blocked" {
				t.Fatal("special leaf")
			}
			found := false
			for _, x := range r.Frontier {
				if x.Name == name {
					found = x.Kind == "special"
				}
			}
			if !found {
				t.Fatal("special kind missing")
			}
		}
	})
}
