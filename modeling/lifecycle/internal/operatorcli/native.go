package operatorcli

import (
	"context"
	"errors"
	"fmt"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/compiler"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/firstdesignreview"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/processview"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/proposalreplay"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/rehearsalinspection"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/revisioncheck"
	"io"
	"runtime"
)

// RunReview preserves the standalone command contract.
func RunReview(args []string, in io.Reader, out, diag io.Writer) int {
	fail := func(code int, msg string) int {
		_, _ = fmt.Fprintln(diag, "modelreviewcheck: "+msg)
		return code
	}
	if len(args) != 0 {
		return fail(2, "no arguments accepted")
	}
	raw, err := io.ReadAll(io.LimitReader(in, firstdesignreview.MaxRequestBytes+1))
	if err != nil || len(raw) > firstdesignreview.MaxRequestBytes {
		return fail(2, "input read or bound failed")
	}
	result, err := firstdesignreview.Check(raw)
	if err != nil {
		if errors.Is(err, firstdesignreview.ErrInvalid) || errors.Is(err, firstdesignreview.ErrLimit) {
			return fail(2, "invalid request or limit")
		}
		return fail(4, "review boundary failed")
	}
	n, err := out.Write(result)
	if err != nil || n != len(result) {
		return fail(3, "report delivery failed")
	}
	return 0
}

// RunProposal preserves the standalone command contract.
func RunProposal(ctx context.Context, args []string, in io.Reader, out, diag io.Writer) int {
	fail := func(code int, message string) int { _, _ = fmt.Fprintln(diag, "modelproposal: "+message); return code }
	if len(args) != 1 || (args[0] != "check" && args[0] != "execute") {
		return fail(2, "usage: modelproposal check|execute < request.json")
	}
	raw, err := io.ReadAll(io.LimitReader(in, proposalreplay.MaxRequestBytes+1))
	if err != nil || len(raw) > proposalreplay.MaxRequestBytes {
		return fail(2, "input read or bound failed")
	}
	var receipt proposalreplay.Receipt
	if args[0] == "check" {
		receipt, err = proposalreplay.Check(raw)
	} else {
		receipt, err = proposalreplay.Execute(ctx, raw, ".")
	}
	if errors.Is(err, proposalreplay.ErrInvalid) {
		return fail(2, "invalid proposal")
	}
	code := proposalreplay.ExitCode(receipt, err)
	encoded, encodeErr := proposalreplay.Encode(receipt)
	if encodeErr != nil {
		return fail(4, "receipt encoding failed; execution may have started")
	}
	n, writeErr := out.Write(encoded)
	if writeErr != nil || n != len(encoded) {
		return fail(3, "receipt delivery failed; execution may have started")
	}
	if code == 4 {
		_, _ = fmt.Fprintln(diag, "modelproposal: execution boundary failed; inspect receipt")
	}
	return code
}

// RunProcess preserves the standalone command contract.
func RunProcess(args []string, in io.Reader, out, diagnostic io.Writer) int {
	fail := func(code int, message string) int {
		_, _ = fmt.Fprintln(diagnostic, "modelprocess: "+message)
		return code
	}
	if len(args) != 1 || (args[0] != "json" && args[0] != "svg") {
		return fail(2, "usage: modelprocess json|svg < model.json")
	}
	if runtime.Version() != compiler.ToolchainVersion {
		return fail(2, "unsupported Go toolchain")
	}
	raw, e := io.ReadAll(io.LimitReader(in, processview.MaxSourceBytes+1))
	if e != nil {
		return fail(2, "input read failed")
	}
	if len(raw) > processview.MaxSourceBytes {
		return fail(2, "invalid model or output limit")
	}
	result, e := processview.Build(raw)
	if e != nil {
		return fail(2, "invalid model or output limit")
	}
	b := result.JSON
	if args[0] == "svg" {
		b = result.SVG
	}
	n, e := out.Write(b)
	if e != nil || n != len(b) {
		return fail(3, "output write failed")
	}
	return 0
}

// RunInspection preserves the standalone command contract.
func RunInspection(args []string, in io.Reader, out, diagnostic io.Writer) int {
	if len(args) != 0 {
		_, _ = io.WriteString(diagnostic, "usage: modelrehearsalinspect < request.json\n")
		return 2
	}
	raw, err := io.ReadAll(io.LimitReader(in, rehearsalinspection.MaxRequestBytes+1))
	if err != nil {
		_, _ = io.WriteString(diagnostic, "inspection input rejected\n")
		return 2
	}
	receipt, err := rehearsalinspection.InspectRequest(raw)
	if err != nil {
		_, _ = io.WriteString(diagnostic, "inspection rejected\n")
		return 2
	}
	n, err := out.Write(receipt)
	if err != nil || n != len(receipt) {
		_, _ = io.WriteString(diagnostic, "inspection output failed\n")
		return 3
	}
	return 0
}

// RevisionFailure preserves the standalone revision CLI error classifier.
func RevisionFailure(err error) int {
	if errors.Is(err, revisioncheck.ErrInvalid) || errors.Is(err, revisioncheck.ErrLimit) {
		return 2
	}
	return 4
}

// RunRevision preserves the standalone command contract.
func RunRevision(args []string, in io.Reader, out, diag io.Writer) int {
	if len(args) != 0 {
		_, _ = io.WriteString(diag, "modelrevisioncheck: invalid request\n")
		return 2
	}
	raw, err := io.ReadAll(io.LimitReader(in, revisioncheck.MaxRequestBytes+1))
	if err != nil {
		_, _ = io.WriteString(diag, "modelrevisioncheck: input failed\n")
		return 2
	}
	report, err := revisioncheck.Check(raw)
	if err != nil {
		_, _ = io.WriteString(diag, "modelrevisioncheck: request rejected\n")
		return RevisionFailure(err)
	}
	n, err := out.Write(report)
	if err != nil || n != len(report) {
		_, _ = io.WriteString(diag, "modelrevisioncheck: output failed\n")
		return 3
	}
	return 0
}
