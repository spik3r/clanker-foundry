package proposalreplay

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/artifactbundle"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/scenarios"
)

// Authored and hashed by /root/review_current_prs before bridge logic; the
// implementer only decoded that author's frozen base64 payloads to disk.
const oracleSHA = "cdc873a15122f8ccd1bd020cf7d193f101d88e5360e3a873a1bec0317cf57804"

type oracleSet struct {
	Name, ModelFile, SuiteFile, ManifestFile                   string
	ModelSourceSHA256, SuiteSourceSHA256, ManifestSourceSHA256 string
	ModelDigest, ScenarioDigest, ReportSourceSHA256            string
	Report                                                     json.RawMessage
}

func read(t *testing.T, name string) []byte {
	t.Helper()
	b, err := os.ReadFile(name)
	if err != nil {
		t.Fatal(err)
	}
	return b
}
func literals(t *testing.T) []oracleSet {
	t.Helper()
	b := read(t, "testdata/expected.json")
	if hash(b) != oracleSHA {
		t.Fatal("independent oracle changed")
	}
	if hash(read(t, "testdata/README.md")) != "3785cfa11567ecf9f45d9fd34880e3d2ef58f60e61319714a6a96a5e7b5771d8" {
		t.Fatal("oracle provenance changed")
	}
	var x struct{ Sets []oracleSet }
	if err := json.Unmarshal(b, &x); err != nil || len(x.Sets) != 2 {
		t.Fatal("oracle metadata", err)
	}
	return x.Sets
}
func fixture(t *testing.T, x oracleSet) Request {
	t.Helper()
	m, s, a := read(t, "testdata/"+x.ModelFile), read(t, "testdata/"+x.SuiteFile), read(t, "testdata/"+x.ManifestFile)
	if hash(m) != x.ModelSourceSHA256 || hash(s) != x.SuiteSourceSHA256 || hash(a) != x.ManifestSourceSHA256 {
		t.Fatal("independent source bytes changed")
	}
	return Request{RequestVersion, string(m), string(s), string(a)}
}
func requestBytes(t *testing.T, r Request) []byte {
	t.Helper()
	b, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	return b
}
func moduleRoot(t *testing.T) string {
	t.Helper()
	p, err := filepath.Abs("..")
	if err != nil {
		t.Fatal(err)
	}
	return p
}
func sameJSON(t *testing.T, got, want []byte) {
	t.Helper()
	var g, w any
	if json.Unmarshal(got, &g) != nil || json.Unmarshal(want, &w) != nil || !reflect.DeepEqual(g, w) {
		t.Fatalf("complete JSON differs\ngot %s\nwant %s", got, want)
	}
}

func TestFrozenCheckOnly(t *testing.T) {
	for _, x := range literals(t) {
		t.Run(x.Name, func(t *testing.T) {
			q := fixture(t, x)
			r, err := Check(requestBytes(t, q))
			if err != nil || ExitCode(r, err) != 0 || r.Outcome != "checked" || r.Preflight.CurrentModelDigest != x.ModelDigest || r.Preflight.ScenarioDigest == nil || *r.Preflight.ScenarioDigest != x.ScenarioDigest || r.Preflight.Replay != "not-run" || r.Preflight.SuiteValidation != "valid" {
				t.Fatalf("check: %+v %v", r, err)
			}
			if r.CandidateStarted || r.CandidateReport != nil || r.SourceRevision != nil || r.ArchiveSHA256 != nil || r.BinarySHA256 != nil || r.Compilation != "not-requested" || r.Replay != "not-requested" || r.Cleanup != "not-created" {
				t.Fatal("check implied host work", r)
			}
			if _, err := scenarios.Decode([]byte(q.Suite)); !errors.Is(err, scenarios.ErrModelBinding) {
				t.Fatal("custom suite unexpectedly fits default runner", err)
			}
			var compact bytes.Buffer
			if json.Compact(&compact, x.Report) != nil || hash(append(compact.Bytes(), '\n')) != x.ReportSourceSHA256 {
				t.Fatal("frozen report hash")
			}
		})
	}
}

func TestPreflightLaunchesNothing(t *testing.T) {
	sets := literals(t)
	a, b := fixture(t, sets[0]), fixture(t, sets[1])
	staleBytes := a
	staleBytes.Suite += "\n"
	crossModel := a
	crossModel.Model = b.Model
	crossSuite := a
	crossSuite.Suite = b.Suite
	for name, q := range map[string]Request{"source-byte-drift": staleBytes, "cross-model": crossModel, "cross-suite": crossSuite} {
		t.Run(name, func(t *testing.T) {
			calls := 0
			h := host{sources: func(string) ([]artifactbundle.File, error) { calls++; return nil, ErrExecution }}
			r, err := execute(context.Background(), requestBytes(t, q), "not-a-directory", h)
			if err != nil || calls != 0 || r.Outcome != "stale" || ExitCode(r, err) != 1 || r.CandidateStarted || r.Preflight.Replay != "not-run" || r.Preflight.SuiteValidation != "not-checked" || r.Cleanup != "not-created" {
				t.Fatalf("stale launch/status: calls%d %+v %v", calls, r, err)
			}
		})
	}
	valid := requestBytes(t, a)
	invalid := [][]byte{nil, []byte("null"), append(append([]byte{}, valid...), valid...), []byte(`{"formatVersion":"lifecycle-proposal-request/v1"}`), bytes.Replace(valid, []byte(`"model":`), []byte(`"MODEL":`), 1), bytes.Replace(valid, []byte(`"model":`), []byte(`"program":"anything","model":`), 1), bytes.Replace(valid, []byte(`"model":`), []byte(`"model":"{}","model":`), 1), bytes.Repeat([]byte(" "), MaxRequestBytes+1)}
	invalid = append(invalid,
		bytes.Replace(valid, []byte(`"model":`), []byte(`"action":"execute","model":`), 1),
		bytes.Replace(valid, []byte(`"model":`), []byte(`"environment":{},"model":`), 1),
		bytes.Replace(valid, []byte(`"formatVersion":"lifecycle-proposal-request/v1"`), []byte(`"formatVersion":null`), 1),
		[]byte(`{"formatVersion":"\ud800","model":"","suite":"","manifest":""}`),
		[]byte("{\"formatVersion\":\"\xff\",\"model\":\"\",\"suite\":\"\",\"manifest\":\"\"}"),
		[]byte(`{"formatVersion":"lifecycle-proposal-request/v1","model":[[[[[0]]]]],"suite":"","manifest":""}`))
	for i, raw := range invalid {
		r, err := execute(context.Background(), raw, "", host{})
		if !errors.Is(err, ErrInvalid) || r.FormatVersion != "" || ExitCode(r, err) != 2 {
			t.Fatalf("invalid %d %+v %v", i, r, err)
		}
	}
	malformed := a
	malformed.Suite = strings.Replace(a.Suite, `"refusals":[]`, `"refusals":null`, 1)
	var manifest map[string]any
	if json.Unmarshal([]byte(a.Manifest), &manifest) != nil {
		t.Fatal("manifest")
	}
	manifest["suiteSourceSha256"] = hash([]byte(malformed.Suite))
	mb, _ := json.Marshal(manifest)
	malformed.Manifest = string(mb)
	if _, err := execute(context.Background(), requestBytes(t, malformed), "", host{}); !errors.Is(err, ErrInvalid) {
		t.Fatal("malformed whole-suite preflight", err)
	}
}

func TestChangedExpectationIsNotRepaired(t *testing.T) {
	x := literals(t)[0]
	q := fixture(t, x)
	// A separate negative input, not an amendment to any frozen fixture.
	q.Suite = strings.Replace(q.Suite, `"revision":"8"`, `"revision":"9"`, 1)
	var manifest map[string]any
	if json.Unmarshal([]byte(q.Manifest), &manifest) != nil {
		t.Fatal("manifest")
	}
	manifest["suiteSourceSha256"] = hash([]byte(q.Suite))
	mb, _ := json.Marshal(manifest)
	q.Manifest = string(mb)
	r, err := Execute(context.Background(), requestBytes(t, q), moduleRoot(t))
	if err != nil || r.Compilation != "passed" || r.Outcome != "mismatched" || !r.CandidateStarted || r.CandidateExitCode == nil || *r.CandidateExitCode != 1 || r.Cleanup != "removed" {
		t.Fatalf("changed expectation was repaired or not run: %+v %v", r, err)
	}
	var report struct {
		StepsExecuted int `json:"stepsExecuted"`
		Entries       []struct {
			Expected struct {
				Task struct {
					Revision string `json:"revision"`
				} `json:"task"`
			} `json:"expected"`
			Actual struct {
				Task struct {
					Revision string `json:"revision"`
				} `json:"task"`
			} `json:"actual"`
			Differences []string `json:"differences"`
		} `json:"entries"`
	}
	if json.Unmarshal(r.CandidateReport, &report) != nil || report.StepsExecuted != 1 || len(report.Entries) != 1 || report.Entries[0].Expected.Task.Revision != "9" || report.Entries[0].Actual.Task.Revision != "8" || !reflect.DeepEqual(report.Entries[0].Differences, []string{"task"}) {
		t.Fatalf("wrong literal mismatch: %s", r.CandidateReport)
	}
}

func TestReceiptEncodingBoundary(t *testing.T) {
	r, err := Check(requestBytes(t, fixture(t, literals(t)[0])))
	if err != nil {
		t.Fatal(err)
	}
	r.CandidateReport = json.RawMessage(`{"text":"<tag>& literal \\u003c"}`)
	b, err := Encode(r)
	if err != nil || !bytes.Contains(b, []byte(`<tag>& literal \\u003c`)) || bytes.Contains(b, []byte(`\u003ctag`)) {
		t.Fatalf("encoding %s %v", b, err)
	}
	r.CandidateReport = json.RawMessage(`not-json`)
	if _, err := Encode(r); !errors.Is(err, ErrExecution) {
		t.Fatal("invalid raw report", err)
	}
	r.CandidateReport = nil
	r.HostPlatform = strings.Repeat("x", MaxReceiptBytes)
	if _, err := Encode(r); !errors.Is(err, ErrExecution) {
		t.Fatal("unbounded receipt", err)
	}
}

func snapshot(t *testing.T, root string) map[string]string {
	t.Helper()
	out := map[string]string{}
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return errors.New("unexpected source type")
		}
		rel, _ := filepath.Rel(root, path)
		out[rel] = info.Mode().String() + " " + hash(read(t, path))
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return out
}

func TestActualCandidatesAndRepeatability(t *testing.T) {
	root := moduleRoot(t)
	before := snapshot(t, root)
	var first []byte
	sets := literals(t)
	for _, x := range append(sets, sets[0]) {
		r, err := Execute(context.Background(), requestBytes(t, fixture(t, x)), root)
		if err != nil || ExitCode(r, err) != 0 || r.Outcome != "matched" || r.Replay != "matched" || !r.CandidateStarted || r.CandidateExitCode == nil || *r.CandidateExitCode != 0 || r.Cleanup != "removed" || r.Compilation != "passed" || r.BinarySHA256 == nil || r.ArchiveSHA256 == nil {
			t.Fatalf("actual %s %+v %v", x.Name, r, err)
		}
		if hash(r.CandidateReport) != x.ReportSourceSHA256 {
			t.Fatalf("independently frozen raw report %s: got %s want %s", x.Name, hash(r.CandidateReport), x.ReportSourceSHA256)
		}
		sameJSON(t, r.CandidateReport, x.Report)
		wire, err := Encode(r)
		if err != nil {
			t.Fatal(err)
		}
		if x.Name == "a" {
			if first == nil {
				first = wire
			} else if !bytes.Equal(first, wire) {
				t.Fatal("receipt differs across identical invocations")
			}
		}
	}
	if !reflect.DeepEqual(before, snapshot(t, root)) {
		t.Fatal("module source bytes/modes changed")
	}
}

func TestCompilingBehavioralMutants(t *testing.T) {
	sets := literals(t)
	for _, tc := range []struct {
		name     string
		set      int
		old, new string
	}{
		{"owner-assignment", 0, `if next.Owner == nil {`, `if false && next.Owner == nil {`},
		{"curator-guard", 1, `if !guard_G_curator(task, command, currentDigest) {`, `if false {`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			h := localHost()
			realRun := h.run
			patched := false
			h.run = func(ctx context.Context, program string, args []string, dir string, in []byte, limit int) processResult {
				if len(args) > 0 && args[0] == "build" {
					path := filepath.Join(dir, "generated/lifecycle.go")
					b := read(t, path)
					if !bytes.Contains(b, []byte(tc.old)) {
						t.Fatalf("mutant target absent: %s", tc.old)
					}
					b = bytes.Replace(b, []byte(tc.old), []byte(tc.new), 1)
					if err := os.WriteFile(path, b, 0600); err != nil {
						t.Fatal(err)
					}
					patched = true
				}
				return realRun(ctx, program, args, dir, in, limit)
			}
			r, err := execute(context.Background(), requestBytes(t, fixture(t, sets[tc.set])), moduleRoot(t), h)
			if !patched || err != nil || r.Compilation != "passed" || !r.CandidateStarted || r.Outcome != "mismatched" || r.Replay != "mismatched" || r.CandidateExitCode == nil || *r.CandidateExitCode != 1 || ExitCode(r, err) != 1 || r.Cleanup != "removed" {
				t.Fatalf("not a compiling behavioral kill: %+v %v", r, err)
			}
		})
	}
}
