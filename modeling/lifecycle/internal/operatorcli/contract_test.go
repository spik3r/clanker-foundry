package operatorcli_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/firstdesignreview"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/internal/operatorcli"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/processview"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/proposalreplay"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/rehearsalinspection"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/revisioncheck"
)

// Frozen before operator production. Prepare expectations are static literals.
// Native parity is compatibility against pinned original APIs, never agreement
// between two wrappers calling the same newly extracted CLI implementation.
type corpus struct {
	Help                       string
	InspectionInput            string
	DeliberateMismatchProposal string
	NativeSourceSHA256         map[string]string
	Prepare                    []struct {
		ID, Intake, ExpectedIntake, ExpectedOutput string
		Artifacts                                  map[string]string
	}
	InvalidNative []struct {
		ID         string
		Args       []string
		Input      string
		Status     int
		Diagnostic string
	}
}

func read(t *testing.T, path string) []byte {
	t.Helper()
	b, e := os.ReadFile(path)
	if e != nil {
		t.Fatal(e)
	}
	return b
}
func vectors(t *testing.T) corpus {
	t.Helper()
	raw := read(t, "testdata/observations.json")
	sum := sha256.Sum256(raw)
	if hex.EncodeToString(sum[:]) != "8b70e7b9658d124f3ce3f5987b7d20e73c305964801aa6b3989e9dc59cd4edc3" {
		t.Fatal("independent contract corpus changed")
	}
	var c corpus
	if json.Unmarshal(raw, &c) != nil {
		t.Fatal("corpus")
	}
	return c
}

type poisonReader struct{}

func (poisonReader) Read([]byte) (int, error) {
	panic("route read stdin before admitting its argument shape")
}

type failedReader struct{}

func (failedReader) Read([]byte) (int, error) { return 0, errors.New("invented private reader detail") }

type shortWriter struct{}

func (shortWriter) Write(p []byte) (int, error) { return len(p) / 2, nil }
func prepareFiles(t *testing.T, intake string, artifacts map[string]string) ([]string, map[string]string) {
	t.Helper()
	d := t.TempDir()
	values := map[string]string{"intake": intake}
	for k, v := range artifacts {
		values[k] = v
	}
	paths := map[string]string{}
	args := []string{"prepare"}
	for _, k := range []string{"intake", "requirements", "model", "suite", "manifest"} {
		p := filepath.Join(d, k+".json")
		if e := os.WriteFile(p, []byte(values[k]), 0600); e != nil {
			t.Fatal(e)
		}
		paths[k] = p
		args = append(args, "--"+k, p)
	}
	return args, paths
}
func run(args []string, input io.Reader) (int, []byte, string) {
	var out, diag bytes.Buffer
	code := operatorcli.Run(context.Background(), args, input, &out, &diag)
	return code, out.Bytes(), diag.String()
}
func TestFrozenPreparePreservesExactArtifactAndUnchangedIntakeBytes(t *testing.T) {
	for _, v := range vectors(t).Prepare {
		t.Run(v.ID, func(t *testing.T) {
			args, paths := prepareFiles(t, v.Intake, v.Artifacts)
			before := map[string][]byte{}
			for k, p := range paths {
				before[k] = read(t, p)
			}
			code, out, diag := run(args, poisonReader{})
			if code != 0 || diag != "" || string(out) != v.ExpectedOutput {
				t.Fatalf("prepare code %d diag %q\ngot %s\nwant %s", code, diag, out, v.ExpectedOutput)
			}
			var wrapper firstdesignreview.Request
			if json.Unmarshal(out, &wrapper) != nil || wrapper.IntakeRequest != v.ExpectedIntake {
				t.Fatal("intake source changed outside observed value")
			}
			var p proposalreplay.Request
			if json.Unmarshal([]byte(wrapper.ProposalRequest), &p) != nil || p.Model != v.Artifacts["model"] || p.Suite != v.Artifacts["suite"] || p.Manifest != v.Artifacts["manifest"] {
				t.Fatal("artifact bytes changed")
			}
			for k, p := range paths {
				if !bytes.Equal(read(t, p), before[k]) {
					t.Fatal("input file changed", k)
				}
			}
			reversed := []string{"prepare"}
			for i := len(args) - 2; i >= 1; i -= 2 {
				reversed = append(reversed, args[i:i+2]...)
			}
			code, again, diag := run(reversed, poisonReader{})
			if code != 0 || diag != "" || !bytes.Equal(again, out) {
				t.Fatal("flag order changed output")
			}
		})
	}
}
func TestPrepareRefusalsAreFixedAndReturnNoPartialOutput(t *testing.T) {
	c := vectors(t)
	v := c.Prepare[0]
	for _, role := range []string{"requirements", "model", "scenarioSuite"} {
		t.Run("conflict-"+role, func(t *testing.T) {
			var intake map[string]any
			_ = json.Unmarshal([]byte(v.Intake), &intake)
			intake["observed"].(map[string]any)[role] = "invented conflicting bytes"
			b, _ := json.Marshal(intake)
			args, _ := prepareFiles(t, string(b), v.Artifacts)
			code, out, diag := run(args, poisonReader{})
			if code != 2 || len(out) != 0 || diag != "foundry prepare: observation conflict\n" {
				t.Fatal("conflict classification", code, diag)
			}
		})
	}
	for _, bad := range []string{"{}", "null", "{", "{\"formatVersion\":\"x\",\"formatVersion\":\"x\"}"} {
		args, _ := prepareFiles(t, bad, v.Artifacts)
		code, out, diag := run(args, poisonReader{})
		if code != 2 || len(out) != 0 || diag != "foundry prepare: input rejected\n" {
			t.Fatal("bad original intake", code, diag)
		}
	}
	for _, kind := range []string{"symlink", "directory", "invalid-utf8", "oversize"} {
		t.Run(kind, func(t *testing.T) {
			args, paths := prepareFiles(t, v.Intake, v.Artifacts)
			p := paths["model"]
			if e := os.Remove(p); e != nil {
				t.Fatal(e)
			}
			switch kind {
			case "symlink":
				if e := os.Symlink(paths["suite"], p); e != nil {
					t.Fatal(e)
				}
			case "directory":
				if e := os.Mkdir(p, 0700); e != nil {
					t.Fatal(e)
				}
			case "invalid-utf8":
				if e := os.WriteFile(p, []byte{0xff}, 0600); e != nil {
					t.Fatal(e)
				}
			case "oversize":
				if e := os.WriteFile(p, bytes.Repeat([]byte("x"), (1<<20)+1), 0600); e != nil {
					t.Fatal(e)
				}
			}
			code, out, diag := run(args, poisonReader{})
			if code != 2 || len(out) != 0 || diag != "foundry prepare: input rejected\n" {
				t.Fatal("leaf refusal", code, diag)
			}
		})
	}
	args, _ := prepareFiles(t, v.Intake, v.Artifacts)
	for _, bad := range [][]string{{"prepare"}, append(append([]string{}, args...), "--model", args[6]), append(append([]string{}, args...), "--extra", "x"), args[:len(args)-1]} {
		code, out, diag := run(bad, poisonReader{})
		if code != 2 || len(out) != 0 || diag != "foundry prepare: invalid arguments\n" {
			t.Fatal("argument refusal", code, diag)
		}
	}
	var diag bytes.Buffer
	if code := operatorcli.Run(context.Background(), args, poisonReader{}, shortWriter{}, &diag); code != 3 || diag.String() != "foundry prepare: output failed\n" {
		t.Fatal("short prepare output", code, diag.String())
	}
}
func TestHelpAndUnknownRoutesDoNotReadInput(t *testing.T) {
	c := vectors(t)
	for _, args := range [][]string{nil, {"help"}, {"--help"}} {
		code, out, diag := run(args, poisonReader{})
		// The frozen v1 route lines stay intact; the new independent change
		// contract checks the complete extended help and execution warning.
		legacyRoutes := strings.SplitN(c.Help, "\n\n", 2)[0]
		if code != 0 || !strings.HasPrefix(string(out), legacyRoutes+"\n") || diag != "" {
			t.Fatal("help", code, diag)
		}
	}
	for _, args := range [][]string{{"help", "extra"}, {"--help", "extra"}, {"execute"}, {"unknown"}} {
		code, out, diag := run(args, poisonReader{})
		if code != 2 || len(out) != 0 || diag != "foundry: unknown command; use foundry help\n" {
			t.Fatal("unknown route", code, diag)
		}
	}
	var diag bytes.Buffer
	if code := operatorcli.Run(context.Background(), []string{"help"}, poisonReader{}, shortWriter{}, &diag); code != 3 || diag.String() != "foundry: output failed\n" {
		t.Fatal("help short write", code, diag.String())
	}
}
func TestFrozenNativeArgumentAndMalformedInputDiagnostics(t *testing.T) {
	for _, v := range vectors(t).InvalidNative {
		t.Run(v.ID, func(t *testing.T) {
			code, out, diag := run(v.Args, strings.NewReader(v.Input))
			if code != v.Status || len(out) != 0 || diag != v.Diagnostic {
				t.Fatalf("code %d diag %q want %d %q", code, diag, v.Status, v.Diagnostic)
			}
		})
	}
}
func nativeInputs(t *testing.T) ([]byte, []byte) {
	t.Helper()
	p := filepath.Join("..", "..", "..", "..", "skills", "model-driven-design", "examples")
	text := func(n string) string { return string(read(t, filepath.Join(p, n))) }
	var intake map[string]any
	if json.Unmarshal([]byte(text("intake.json")), &intake) != nil {
		t.Fatal("intake fixture")
	}
	intake["observed"] = map[string]string{"requirements": text("brief.md"), "model": text("model.json"), "scenarioSuite": text("suite.json")}
	i, _ := json.Marshal(intake)
	proposal, _ := json.Marshal(proposalreplay.Request{FormatVersion: proposalreplay.RequestVersion, Model: text("model.json"), Suite: text("suite.json"), Manifest: text("manifest.json")})
	review, _ := json.Marshal(firstdesignreview.Request{FormatVersion: firstdesignreview.RequestVersion, IntakeRequest: string(i), ProposalRequest: string(proposal)})
	return review, proposal
}
func TestNativeAPIsRemainPinnedAndFullNonexecutingOutputsMatch(t *testing.T) {
	c := vectors(t)
	for path, want := range c.NativeSourceSHA256 {
		b := read(t, filepath.Join("..", "..", filepath.FromSlash(path)))
		sum := sha256.Sum256(b)
		if hex.EncodeToString(sum[:]) != want {
			t.Fatal("native API source changed", path)
		}
	}
	review, proposal := nativeInputs(t)
	model := read(t, "../../processview/testdata/model.json")
	var revisions struct{ Vectors []struct{ Request string } }
	if json.Unmarshal(read(t, "../../revisioncheck/testdata/observations.json"), &revisions) != nil {
		t.Fatal("revision fixture")
	}
	cases := []struct {
		args     []string
		input    []byte
		expected func() ([]byte, error)
	}{
		{[]string{"review"}, review, func() ([]byte, error) { return firstdesignreview.Check(review) }},
		{[]string{"proposal", "check"}, proposal, func() ([]byte, error) {
			r, e := proposalreplay.Check(proposal)
			if e != nil {
				return nil, e
			}
			return proposalreplay.Encode(r)
		}},
		{[]string{"process", "json"}, model, func() ([]byte, error) { o, e := processview.Build(model); return o.JSON, e }},
		{[]string{"process", "svg"}, model, func() ([]byte, error) { o, e := processview.Build(model); return o.SVG, e }},
		{[]string{"inspect"}, []byte(c.InspectionInput), func() ([]byte, error) { return rehearsalinspection.InspectRequest([]byte(c.InspectionInput)) }},
		{[]string{"revisions"}, []byte(revisions.Vectors[0].Request), func() ([]byte, error) { return revisioncheck.Check([]byte(revisions.Vectors[0].Request)) }},
	}
	for _, v := range cases {
		t.Run(strings.Join(v.args, "-"), func(t *testing.T) {
			expected, e := v.expected()
			if e != nil {
				t.Fatal("native baseline", e)
			}
			code, out, diag := run(v.args, bytes.NewReader(v.input))
			if code != 0 || diag != "" || !bytes.Equal(out, expected) {
				t.Fatalf("native output changed code%d diag%q", code, diag)
			}
		})
	}
	// The changed model retains its native stale observation, not a new
	// aggregate readiness decision. Decode the two explicit JSON-text layers.
	code, out, diag := run([]string{"revisions"}, strings.NewReader(revisions.Vectors[0].Request))
	var revision revisioncheck.Report
	var after struct {
		ProposalReportJSON string `json:"proposalReportJSON"`
	}
	var preflight proposalreplay.Receipt
	if code != 0 || diag != "" || json.Unmarshal(out, &revision) != nil || json.Unmarshal([]byte(revision.After.FirstDesignReportJSON), &after) != nil || json.Unmarshal([]byte(after.ProposalReportJSON), &preflight) != nil || preflight.Outcome != "stale" || preflight.CandidateStarted {
		t.Fatal("native stale result lost")
	}
}
func TestExplicitProposalExecutionAndCancelledContextKeepNativeMeanings(t *testing.T) {
	_, proposal := nativeInputs(t)
	c := vectors(t)
	// The existing trusted execution bridge requires its module-root source path.
	t.Chdir("../..")
	for _, v := range []struct {
		name    string
		input   []byte
		status  int
		outcome string
	}{{"matched", proposal, 0, "matched"}, {"mismatched", []byte(c.DeliberateMismatchProposal), 1, "mismatched"}} {
		t.Run(v.name, func(t *testing.T) {
			native, e := proposalreplay.Execute(context.Background(), v.input, ".")
			expected, encErr := proposalreplay.Encode(native)
			if e != nil || encErr != nil || native.Outcome != v.outcome {
				t.Fatal("native execution baseline", e, encErr, native.Outcome)
			}
			var out, diag bytes.Buffer
			code := operatorcli.Run(context.Background(), []string{"proposal", "execute"}, bytes.NewReader(v.input), &out, &diag)
			if code != v.status || diag.Len() != 0 || !bytes.Equal(out.Bytes(), expected) {
				t.Fatalf("explicit execute parity code%d diag%q", code, diag.String())
			}
		})
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	var out, diag bytes.Buffer
	code := operatorcli.Run(ctx, []string{"proposal", "execute"}, bytes.NewReader(proposal), &out, &diag)
	var r proposalreplay.Receipt
	if json.Unmarshal(out.Bytes(), &r) != nil || code != 4 || r.FailureStage == nil || *r.FailureStage != "context" || r.CandidateStarted || r.Compilation != "not-started" || r.Replay != "not-started" || diag.String() != "modelproposal: execution boundary failed; inspect receipt\n" {
		t.Fatal("cancelled execution classification", code, diag.String())
	}
}
