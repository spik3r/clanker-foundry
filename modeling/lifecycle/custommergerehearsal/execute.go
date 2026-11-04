package custommergerehearsal

import (
	"context"
	"errors"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/proposalreplay"
)

type executor func(context.Context, []byte, string) (proposalreplay.Receipt, error)

// Execute requires separate caller authority on a trusted native local host.
// sourceRoot is host configuration, never request data. Cancellation and cleanup
// receipts do not prove rollback or descendant termination.
func Execute(ctx context.Context, raw []byte, sourceRoot string) ([]byte, error) {
	return executeWith(ctx, raw, sourceRoot, proposalreplay.Execute)
}
func finished(r proposalreplay.Receipt) bool {
	if r.Action != "execute" || r.Compilation != "passed" || !r.CandidateStarted || r.CandidateExitCode == nil || r.Cleanup != "removed" || r.FailureStage != nil {
		return false
	}
	switch r.Outcome {
	case "matched":
		return r.Replay == "matched" && *r.CandidateExitCode == 0
	case "mismatched":
		return r.Replay == "mismatched" && *r.CandidateExitCode == 1
	default:
		return false
	}
}
func executeWith(ctx context.Context, raw []byte, sourceRoot string, run executor) ([]byte, error) {
	p, err := prepare(raw)
	if err != nil {
		return nil, err
	}
	p.report.Action = "execute"
	if p.report.Outcome != "checked" {
		return encode(p.report, MaxReportBytes)
	}
	fail := func(cause error) ([]byte, error) {
		p.report.Outcome = "failed"
		out, e := encode(p.report, MaxReportBytes)
		return out, errors.Join(ErrBoundary, cause, e)
	}
	contextError := func() error {
		if ctx == nil {
			return ErrBoundary
		}
		return ctx.Err()
	}
	if run == nil {
		return fail(ErrBoundary)
	}
	names := [4]string{"base", "left", "right", "candidate"}
	for i, s := range p.report.stages() {
		if e := contextError(); e != nil {
			return fail(e)
		}
		receipt, nativeErr := run(ctx, p.proposals[i], sourceRoot)
		if e := retain(s, "execute", p.proposals[i], receipt); e != nil {
			return nil, errors.Join(ErrBoundary, e)
		}
		if nativeErr != nil {
			return fail(nativeErr)
		}
		if !finished(receipt) {
			return fail(ErrBoundary)
		}
		if e := contextError(); e != nil {
			return fail(e)
		}
		if receipt.Outcome == "mismatched" {
			p.report.Outcome = names[i] + "-mismatched"
			return encode(p.report, MaxReportBytes)
		}
	}
	p.report.Outcome = "matched"
	out, e := encode(p.report, MaxReportBytes)
	if e != nil {
		return nil, errors.Join(ErrBoundary, e)
	}
	return out, nil
}
