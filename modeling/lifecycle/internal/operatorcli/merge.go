package operatorcli

import (
	"context"
	"io"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/custommergerehearsal"
)

// RunMerge preserves the standalone custom merge rehearsal command contract.
func RunMerge(ctx context.Context, args []string, in io.Reader, out, diag io.Writer) int {
	fail := func(code int, message string) int {
		_, _ = io.WriteString(diag, "modelmergerehearsal: "+message+"\n")
		return code
	}
	if len(args) != 1 || args[0] != "check" && args[0] != "execute" {
		return fail(2, "usage: modelmergerehearsal check|execute < request.json")
	}
	raw, err := io.ReadAll(io.LimitReader(in, custommergerehearsal.MaxRequestBytes+1))
	if err != nil {
		return fail(2, "input read failed")
	}
	var report []byte
	if args[0] == "check" {
		report, err = custommergerehearsal.Check(raw)
	} else {
		report, err = custommergerehearsal.Execute(ctx, raw, ".")
	}
	code := custommergerehearsal.ExitCode(report, err)
	if len(report) == 0 {
		if code == 2 {
			return fail(2, "request rejected")
		}
		return fail(4, "boundary failure; execution may have started")
	}
	n, e := out.Write(report)
	if e != nil || n != len(report) {
		return fail(3, "report delivery failed; execution may have started")
	}
	if code == 4 {
		_, _ = io.WriteString(diag, "modelmergerehearsal: execution boundary failed; inspect receipt\n")
	}
	return code
}
