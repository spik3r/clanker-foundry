// Package operatorcli shares native CLI contracts and bounded request assembly.
package operatorcli

import (
	"context"
	"io"
)

const helpText = "Usage: foundry prepare --intake FILE --requirements FILE --model FILE --suite FILE --manifest FILE\n" +
	"       foundry review\n" +
	"       foundry proposal check|execute\n" +
	"       foundry process json|svg\n" +
	"       foundry inspect\n" +
	"       foundry revisions\n\n" +
	"All reports use the native tool contracts. Only proposal execute runs a candidate.\n"

type runner func(context.Context, []string, io.Reader, io.Writer, io.Writer) int

type capabilities struct {
	prepare, review, check, execute, process, inspect, revisions runner
}

// Run preserves each native route's reports and status meanings. Only the
// explicit proposal execute route receives the candidate-execution capability.
func Run(ctx context.Context, args []string, in io.Reader, out, diag io.Writer) int {
	withoutContext := func(f func([]string, io.Reader, io.Writer, io.Writer) int) runner {
		return func(_ context.Context, args []string, in io.Reader, out, diag io.Writer) int {
			return f(args, in, out, diag)
		}
	}
	proposal := func(action string) runner {
		return func(ctx context.Context, args []string, in io.Reader, out, diag io.Writer) int {
			return RunProposal(ctx, append([]string{action}, args...), in, out, diag)
		}
	}
	return runWith(ctx, args, in, out, diag, capabilities{
		prepare: withoutContext(prepare), review: withoutContext(RunReview),
		check: proposal("check"), execute: proposal("execute"),
		process: withoutContext(RunProcess), inspect: withoutContext(RunInspection), revisions: withoutContext(RunRevision),
	})
}

// Capabilities are internal test seams, not configurable commands or plugins.
func runWith(ctx context.Context, args []string, in io.Reader, out, diag io.Writer, c capabilities) int {
	if len(args) == 0 || (len(args) == 1 && (args[0] == "help" || args[0] == "--help")) {
		n, err := io.WriteString(out, helpText)
		if err != nil || n != len(helpText) {
			_, _ = io.WriteString(diag, "foundry: output failed\n")
			return 3
		}
		return 0
	}
	switch args[0] {
	case "prepare":
		return c.prepare(ctx, args[1:], in, out, diag)
	case "review":
		return c.review(ctx, args[1:], in, out, diag)
	case "proposal":
		if len(args) == 2 {
			switch args[1] {
			case "check":
				return c.check(ctx, nil, in, out, diag)
			case "execute":
				return c.execute(ctx, nil, in, out, diag)
			}
		}
		// Native argument refusal happens before input or execution is admitted.
		return RunProposal(ctx, args[1:], in, out, diag)
	case "process":
		return c.process(ctx, args[1:], in, out, diag)
	case "inspect":
		return c.inspect(ctx, args[1:], in, out, diag)
	case "revisions":
		return c.revisions(ctx, args[1:], in, out, diag)
	default:
		_, _ = io.WriteString(diag, "foundry: unknown command; use foundry help\n")
		return 2
	}
}
