package changerehearsal

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"path/filepath"
	"testing"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/proposalreplay"
)

func TestExecuteMatchesIndependentFrozenContract(t *testing.T) {
	root, err := filepath.Abs("..")
	if err != nil {
		t.Fatal(err)
	}
	assertFrozenExecute(t, Execute, root)
}
func simulatedFinished(t *testing.T, raw []byte, matched bool) proposalreplay.Receipt {
	t.Helper()
	r, err := proposalreplay.Check(raw)
	if err != nil {
		t.Fatal(err)
	}
	r.Action = "execute"
	r.Outcome = "matched"
	r.Replay = "matched"
	code := 0
	if !matched {
		r.Outcome = "mismatched"
		r.Replay = "mismatched"
		code = 1
	}
	r.Compilation = "passed"
	r.CandidateStarted = true
	r.CandidateExitCode = &code
	r.Cleanup = "removed"
	return r
}
func TestExecutionSuppressionAndActualPartialReceipts(t *testing.T) {
	d := frozenObservations(t)
	valid := []byte(d.Vectors[0].Request)
	for _, v := range d.Vectors[3:] {
		calls := 0
		out, err := executeWith(context.Background(), []byte(v.Request), "owned-root", func(context.Context, []byte, string) (proposalreplay.Receipt, error) {
			calls++
			return proposalreplay.Receipt{}, errors.New("must not execute")
		})
		if calls != 0 {
			t.Fatal("preflight/identity refusal invoked executor", v.ID)
		}
		if v.Valid {
			if err != nil || ExitCode(out, err) != 1 {
				t.Fatal("stale result", err)
			}
		} else if err == nil || out != nil {
			t.Fatal("identity result")
		}
	}
	for _, kind := range []string{"cancel-before", "cancel-between", "baseline-mismatch", "baseline-failure", "baseline-cleanup", "candidate-failure"} {
		t.Run(kind, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			if kind == "cancel-before" {
				cancel()
			}
			calls := 0
			var returned []byte
			run := func(got context.Context, request []byte, root string) (proposalreplay.Receipt, error) {
				calls++
				if got != ctx || root != "owned-root" {
					t.Fatal("host configuration/context changed")
				}
				r := simulatedFinished(t, request, true)
				if kind == "cancel-between" && calls == 1 {
					cancel()
				}
				var err error
				if kind == "baseline-mismatch" && calls == 1 {
					r = simulatedFinished(t, request, false)
				}
				if kind == "baseline-failure" && calls == 1 || kind == "candidate-failure" && calls == 2 {
					r.Outcome = "failed"
					stage := "compilation"
					r.FailureStage = &stage
					r.Compilation = "failed"
					r.Replay = "not-started"
					r.CandidateStarted = false
					r.CandidateExitCode = nil
					err = proposalreplay.ErrExecution
				}
				if kind == "baseline-cleanup" && calls == 1 {
					r.Cleanup = "failed"
				}
				returned, _ = proposalreplay.Encode(r)
				return r, err
			}
			raw, err := executeWith(ctx, valid, "owned-root", run)
			wantCalls := 1
			if kind == "cancel-before" {
				wantCalls = 0
			}
			if kind == "candidate-failure" {
				wantCalls = 2
			}
			if calls != wantCalls {
				t.Fatal("stage sequence", calls, wantCalls)
			}
			var report Report
			if json.Unmarshal(raw, &report) != nil {
				t.Fatal("partial report missing", err)
			}
			if kind == "baseline-mismatch" {
				if err != nil || report.Outcome != "baseline-mismatched" || ExitCode(raw, err) != 1 {
					t.Fatal("mismatch changed", err)
				}
			} else if !errors.Is(err, ErrBoundary) || report.Outcome != "failed" || ExitCode(raw, err) != 4 {
				t.Fatal("boundary failure changed", err)
			}
			if kind == "cancel-before" {
				if report.Baseline.Operation != "check" {
					t.Fatal("fabricated baseline invocation")
				}
			} else {
				actual := report.Baseline
				if kind == "candidate-failure" {
					actual = report.Candidate
				}
				if actual.Operation != "execute" || actual.ReceiptJSON != string(returned) || actual.ReceiptSHA256 != hash(returned) {
					t.Fatal("actual native partial receipt changed")
				}
			}
			if kind != "candidate-failure" && report.Candidate.Operation != "check" {
				t.Fatal("suppressed candidate fabricated invocation")
			}
		})
	}
}
func TestExecutorReceivesExactPreparedRequests(t *testing.T) {
	raw := []byte(frozenObservations(t).Vectors[0].Request)
	saved := bytes.Clone(raw)
	p, err := prepare(raw)
	if err != nil {
		t.Fatal(err)
	}
	calls := 0
	out, err := executeWith(context.Background(), raw, "root", func(_ context.Context, input []byte, _ string) (proposalreplay.Receipt, error) {
		want := p.baseline
		if calls == 1 {
			want = p.candidate
		}
		calls++
		if !bytes.Equal(input, want) {
			t.Fatal("native proposal bytes changed")
		}
		return simulatedFinished(t, input, true), nil
	})
	if err != nil || calls != 2 || ExitCode(out, err) != 0 || !bytes.Equal(raw, saved) {
		t.Fatal("matched sequence", calls, err)
	}
}
