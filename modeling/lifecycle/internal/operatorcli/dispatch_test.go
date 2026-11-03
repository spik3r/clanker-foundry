package operatorcli

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/firstdesignreview"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/proposalreplay"
)

type unreadableInput struct{}

func (unreadableInput) Read([]byte) (int, error) { panic("unexpected input read") }

func TestDispatchKeepsExecutionCapabilityExplicit(t *testing.T) {
	for _, args := range [][]string{
		nil, {"help"}, {"--help"}, {"unknown"}, {"help", "extra"},
		{"prepare"}, {"review"}, {"proposal", "check"}, {"process", "json"},
		{"process", "svg"}, {"inspect"}, {"revisions"},
		{"proposal"}, {"proposal", "bogus"}, {"proposal", "execute", "extra"},
	} {
		t.Run(strings.Join(args, "-"), func(t *testing.T) {
			calls := 0
			harmless := func(context.Context, []string, io.Reader, io.Writer, io.Writer) int { return 0 }
			execute := func(context.Context, []string, io.Reader, io.Writer, io.Writer) int { calls++; return 99 }
			c := capabilities{prepare: harmless, review: harmless, check: harmless, execute: execute, process: harmless, inspect: harmless, revisions: harmless}
			runWith(context.Background(), args, unreadableInput{}, io.Discard, io.Discard, c)
			if calls != 0 {
				t.Fatal("non-executing route invoked execution capability")
			}
		})
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	calls := 0
	c := capabilities{execute: func(got context.Context, args []string, _ io.Reader, _ io.Writer, _ io.Writer) int {
		calls++
		if got != ctx || got.Err() != context.Canceled || len(args) != 0 {
			t.Fatal("execution context/arguments changed")
		}
		return 73
	}}
	if code := runWith(ctx, []string{"proposal", "execute"}, unreadableInput{}, io.Discard, io.Discard, c); code != 73 || calls != 1 {
		t.Fatal("explicit execution dispatch", code, calls)
	}
}

func prepareFixture(t *testing.T) (string, map[string]string) {
	t.Helper()
	raw, err := os.ReadFile("testdata/observations.json")
	if err != nil {
		t.Fatal(err)
	}
	var c struct {
		Prepare []struct {
			Intake    string
			Artifacts map[string]string
		}
	}
	if json.Unmarshal(raw, &c) != nil || len(c.Prepare) == 0 {
		t.Fatal("independent fixture")
	}
	return c.Prepare[0].Intake, c.Prepare[0].Artifacts
}
func writePrepareFiles(t *testing.T, intake string, artifacts map[string]string) ([]string, map[string]string) {
	t.Helper()
	d := t.TempDir()
	paths := map[string]string{}
	args := []string{"prepare"}
	for _, role := range []string{"intake", "requirements", "model", "suite", "manifest"} {
		text := intake
		if role != "intake" {
			text = artifacts[role]
		}
		p := filepath.Join(d, role)
		if err := os.WriteFile(p, []byte(text), 0600); err != nil {
			t.Fatal(err)
		}
		paths[role] = p
		args = append(args, "--"+role, p)
	}
	return args, paths
}
func TestPrepareFileAndEncodedRequestBounds(t *testing.T) {
	original, artifacts := prepareFixture(t)
	for role, limit := range map[string]int{"intake": 4 << 20, "requirements": 1 << 20, "model": 1 << 20, "suite": 1 << 20, "manifest": 64 << 10} {
		t.Run(role, func(t *testing.T) {
			args, paths := writePrepareFiles(t, original, artifacts)
			if err := os.WriteFile(paths[role], bytes.Repeat([]byte("x"), limit+1), 0600); err != nil {
				t.Fatal(err)
			}
			var out, diag bytes.Buffer
			if code := Run(context.Background(), args, unreadableInput{}, &out, &diag); code != 2 || out.Len() != 0 || diag.String() != "foundry prepare: input rejected\n" {
				t.Fatal("source bound", code, diag.String())
			}
		})
	}
	for _, kind := range []string{"expanded-intake", "expanded-wrapper"} {
		t.Run(kind, func(t *testing.T) {
			text := original
			inputs := map[string]string{}
			for k, v := range artifacts {
				inputs[k] = v
			}
			if kind == "expanded-intake" {
				text += strings.Repeat(" ", (4<<20)-len(text))
			} else {
				inputs["requirements"] = ""
				inputs["model"] = strings.Repeat("\x01", 320000)
				inputs["suite"] = strings.Repeat("\x01", 320000)
				// Measure this independently assembled fixture before testing the
				// implementation: nested requests fit, only the outer one exceeds.
				var intakeFields map[string]json.RawMessage
				if json.Unmarshal([]byte(original), &intakeFields) != nil {
					t.Fatal("fixture intake")
				}
				observed, _ := json.Marshal(map[string]string{"requirements": inputs["requirements"], "model": inputs["model"], "scenarioSuite": inputs["suite"]})
				intakeFields["observed"] = observed
				inner, _ := json.Marshal(intakeFields)
				proposal, _ := json.Marshal(proposalreplay.Request{FormatVersion: proposalreplay.RequestVersion, Model: inputs["model"], Suite: inputs["suite"], Manifest: inputs["manifest"]})
				outer, _ := json.Marshal(firstdesignreview.Request{FormatVersion: firstdesignreview.RequestVersion, IntakeRequest: string(inner), ProposalRequest: string(proposal)})
				if len(inner) > 4<<20 || len(proposal) > proposalreplay.MaxRequestBytes || len(outer) <= firstdesignreview.MaxRequestBytes {
					t.Fatalf("wrong fixture sizes: intake=%d proposal=%d outer=%d", len(inner), len(proposal), len(outer))
				}
			}
			args, _ := writePrepareFiles(t, text, inputs)
			var out, diag bytes.Buffer
			if code := Run(context.Background(), args, unreadableInput{}, &out, &diag); code != 2 || out.Len() != 0 || diag.String() != "foundry prepare: request limit exceeded\n" {
				t.Fatal("encoded bound", code, diag.String())
			}
		})
	}
}
func TestPrepareAcceptsParentAliasButLeavesInputsUntouched(t *testing.T) {
	original, artifacts := prepareFixture(t)
	args, paths := writePrepareFiles(t, original, artifacts)
	var expected, diag bytes.Buffer
	if Run(context.Background(), args, unreadableInput{}, &expected, &diag) != 0 {
		t.Fatal(diag.String())
	}
	alias := filepath.Join(t.TempDir(), "input-alias")
	if err := os.Symlink(filepath.Dir(paths["intake"]), alias); err != nil {
		t.Fatal(err)
	}
	for i := 2; i < len(args); i += 2 {
		args[i] = filepath.Join(alias, filepath.Base(args[i]))
	}
	var out bytes.Buffer
	diag.Reset()
	if code := Run(context.Background(), args, unreadableInput{}, &out, &diag); code != 0 || !bytes.Equal(out.Bytes(), expected.Bytes()) || diag.Len() != 0 {
		t.Fatal("parent alias", code, diag.String())
	}
}
