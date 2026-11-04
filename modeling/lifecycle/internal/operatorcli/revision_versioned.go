package operatorcli

import (
	"io"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/revisioncheck"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/strictjson"
)

// Only foundry's revision route selects explicit request versions. The legacy
// RunRevision and standalone modelrevisioncheck remain strict v1 entrypoints.
func runRevisionVersioned(args []string, in io.Reader, out, diag io.Writer) int {
	if len(args) != 0 {
		return RunRevision(args, in, out, diag)
	}
	raw, err := io.ReadAll(io.LimitReader(in, revisioncheck.MaxRequestBytes+1))
	if err != nil {
		_, _ = io.WriteString(diag, "modelrevisioncheck: input failed\n")
		return 2
	}
	var report []byte
	var req revisioncheck.Request
	if len(raw) > revisioncheck.MaxRequestBytes {
		err = revisioncheck.ErrLimit
	} else if strictjson.Decode(raw, &req, revisioncheck.MaxRequestBytes, 4) != nil {
		err = revisioncheck.ErrInvalid
	} else {
		switch req.FormatVersion {
		case revisioncheck.RequestVersion:
			report, err = revisioncheck.Check(raw)
		case revisioncheck.RequestVersionV2:
			report, err = revisioncheck.CheckV2(raw)
		case revisioncheck.RequestVersionV3:
			report, err = revisioncheck.CheckV3(raw)
		default:
			err = revisioncheck.ErrInvalid
		}
	}
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
