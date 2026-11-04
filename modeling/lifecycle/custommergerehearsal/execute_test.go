package custommergerehearsal

import (
	"context"
	"errors"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/proposalreplay"
	"path/filepath"
	"testing"
)

func TestNoExecutorAndNilContext(t *testing.T) {
	raw := contractJSON(t, contractReq(contractLoad(t)))
	for _, nilContext := range []bool{false, true} {
		ctx := context.Background()
		var run executor
		if nilContext {
			ctx = nil
			run = func(context.Context, []byte, string) (proposalreplay.Receipt, error) {
				t.Fatal("nil context invoked executor")
				return proposalreplay.Receipt{}, nil
			}
		}
		out, e := executeWith(ctx, raw, "unused", run)
		if !errors.Is(e, ErrBoundary) || ExitCode(out, e) != 4 || contractRead(t, out).Outcome != "failed" {
			t.Fatal("missing context/executor boundary")
		}
		for _, s := range contractStages(contractRead(t, out)) {
			if s.Operation != "check" {
				t.Fatal("fabricated execution")
			}
		}
	}
}
func TestInconsistentReceiptsSuppressLaterStages(t *testing.T) {
	raw := contractJSON(t, contractReq(contractLoad(t)))
	mutations := []func(*proposalreplay.Receipt){
		func(r *proposalreplay.Receipt) { r.Action = "check" },
		func(r *proposalreplay.Receipt) { r.Compilation = "failed" },
		func(r *proposalreplay.Receipt) { r.CandidateStarted = false },
		func(r *proposalreplay.Receipt) { r.CandidateExitCode = nil },
		func(r *proposalreplay.Receipt) { n := 7; r.CandidateExitCode = &n },
		func(r *proposalreplay.Receipt) { r.Replay = "mismatched" },
		func(r *proposalreplay.Receipt) { r.Outcome = "stale" },
		func(r *proposalreplay.Receipt) { s := "cleanup"; r.FailureStage = &s },
	}
	for _, mutate := range mutations {
		calls := 0
		out, e := executeWith(context.Background(), raw, "unused", func(_ context.Context, p []byte, _ string) (proposalreplay.Receipt, error) {
			calls++
			r := contractSynthetic(t, p)
			mutate(&r)
			return r, nil
		})
		if calls != 1 || !errors.Is(e, ErrBoundary) || ExitCode(out, e) != 4 {
			t.Fatal("inconsistent native receipt continued")
		}
		if contractRead(t, out).Left.Operation != "check" {
			t.Fatal("later stage not suppressed")
		}
	}
}
func TestUnencodableNativeReceiptIsBoundaryNotInputRefusal(t *testing.T) {
	raw := contractJSON(t, contractReq(contractLoad(t)))
	calls := 0
	out, e := executeWith(context.Background(), raw, "unused", func(_ context.Context, p []byte, _ string) (proposalreplay.Receipt, error) {
		calls++
		r := contractSynthetic(t, p)
		r.CandidateReport = []byte("invalid JSON")
		return r, nil
	})
	if calls != 1 || out != nil || !errors.Is(e, ErrBoundary) || ExitCode(out, e) != 4 {
		t.Fatal("post-call error hid possible execution")
	}
	if ExitCode(nil, errors.Join(ErrBoundary, ErrLimit)) != 4 {
		t.Fatal("post-call limit became initial refusal")
	}
}
func TestActualReverseOrientation(t *testing.T) {
	f := contractLoad(t)
	r := contractReq(f)
	r.MergeRequest = f.ReverseMergeRequest
	r.Left, r.Right = r.Right, r.Left
	raw := contractJSON(t, r)
	root, e := filepath.Abs("..")
	if e != nil {
		t.Fatal(e)
	}
	out, e := Execute(context.Background(), raw, root)
	if e != nil || ExitCode(out, e) != 0 {
		t.Fatal("reverse native execution", e)
	}
	got := contractAssertEnvelope(t, raw, out, r)
	if got.Outcome != "matched" {
		t.Fatal("reverse mismatch")
	}
	for _, s := range contractStages(got) {
		if s.Operation != "execute" {
			t.Fatal("missing reverse execution")
		}
	}
}
