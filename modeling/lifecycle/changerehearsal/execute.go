package changerehearsal

import (
	"context"
	"errors"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/proposalreplay"
)

type executor func(context.Context, []byte, string) (proposalreplay.Receipt, error)

// Execute requires separate caller authority on the native trusted local host.
// sourceRoot is host configuration, never request data. Public callers cannot
// supply executors. Cancellation or failure does not prove child rollback.
func Execute(ctx context.Context, raw []byte, sourceRoot string) ([]byte, error) {
	return executeWith(ctx, raw, sourceRoot, proposalreplay.Execute)
}

func finished(receipt proposalreplay.Receipt) bool {
	if receipt.Action != "execute" || receipt.Compilation != "passed" || !receipt.CandidateStarted || receipt.CandidateExitCode == nil || receipt.Cleanup != "removed" || receipt.FailureStage != nil {
		return false
	}
	switch receipt.Outcome {
	case "matched":
		return receipt.Replay == "matched" && *receipt.CandidateExitCode == 0
	case "mismatched":
		return receipt.Replay == "mismatched" && *receipt.CandidateExitCode == 1
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
	if p.report.Outcome == "stale" {
		return encode(p.report, MaxReportBytes)
	}
	fail := func(cause error) ([]byte, error) {
		p.report.Outcome = "failed"
		encoded, err := encode(p.report, MaxReportBytes)
		if err != nil {
			return nil, errors.Join(ErrBoundary, err)
		}
		return encoded, errors.Join(ErrBoundary, cause)
	}
	contextError := func() error {
		if ctx == nil {
			return ErrBoundary
		}
		return ctx.Err()
	}
	if err := contextError(); err != nil {
		return fail(err)
	}
	if run == nil {
		return fail(ErrBoundary)
	}
	baseline, nativeErr := run(ctx, p.baseline, sourceRoot)
	p.report.Baseline, err = side("execute", p.baseline, baseline)
	if err != nil {
		return nil, ErrBoundary
	}
	if nativeErr != nil {
		return fail(nativeErr)
	}
	if !finished(baseline) {
		return fail(ErrBoundary)
	}
	if baseline.Outcome == "mismatched" {
		p.report.Outcome = "baseline-mismatched"
		return encode(p.report, MaxReportBytes)
	}
	if err := contextError(); err != nil {
		return fail(err)
	}
	candidate, nativeErr := run(ctx, p.candidate, sourceRoot)
	p.report.Candidate, err = side("execute", p.candidate, candidate)
	if err != nil {
		return nil, ErrBoundary
	}
	if nativeErr != nil {
		return fail(nativeErr)
	}
	if !finished(candidate) {
		return fail(ErrBoundary)
	}
	p.report.Outcome = "matched"
	if candidate.Outcome == "mismatched" {
		p.report.Outcome = "candidate-mismatched"
	}
	return encode(p.report, MaxReportBytes)
}
