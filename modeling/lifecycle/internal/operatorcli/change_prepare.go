package operatorcli

import (
	"io"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/acceptanceauthor"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/changerehearsal"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/modelpatch"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/proposalreplay"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/scenarios"
)

// Preparation copies supplied artifact bytes, not their meaning. Native check
// remains a separate action, and the reviewed candidate identity is never inferred.
func prepareChange(args []string, _ io.Reader, out, diag io.Writer) int {
	fail := func(message string) int {
		_, _ = io.WriteString(diag, "foundry change prepare: "+message+"\n")
		return 2
	}
	limits := map[string]int64{
		"--patch-request":      modelpatch.MaxRequestBytes,
		"--baseline-proposal":  proposalreplay.MaxRequestBytes,
		"--candidate-suite":    scenarios.MaxSourceBytes,
		"--candidate-manifest": acceptanceauthor.MaxManifestBytes,
	}
	const digestFlag = "--expected-candidate-digest"
	values := make(map[string]string)
	if len(args) != 2*(len(limits)+1) {
		return fail("invalid arguments")
	}
	for i := 0; i < len(args); i += 2 {
		_, known := limits[args[i]]
		_, duplicate := values[args[i]]
		if (!known && args[i] != digestFlag) || duplicate {
			return fail("invalid arguments")
		}
		values[args[i]] = args[i+1]
	}
	digest := values[digestFlag]
	if len(digest) != 64 {
		return fail("invalid arguments")
	}
	for _, c := range digest {
		if !(c >= '0' && c <= '9' || c >= 'a' && c <= 'f') {
			return fail("invalid arguments")
		}
	}
	texts := make(map[string]string)
	for _, flag := range []string{"--patch-request", "--baseline-proposal", "--candidate-suite", "--candidate-manifest"} {
		raw, err := readRegular(values[flag], limits[flag])
		if err != nil {
			return fail("input rejected")
		}
		texts[flag] = string(raw)
	}
	request, err := encodePrepared(changerehearsal.Request{
		FormatVersion:           changerehearsal.RequestVersion,
		PatchRequest:            texts["--patch-request"],
		BaselineProposalRequest: texts["--baseline-proposal"],
		CandidateSuite:          texts["--candidate-suite"],
		CandidateManifest:       texts["--candidate-manifest"],
		ExpectedCandidateDigest: digest,
	}, changerehearsal.MaxRequestBytes)
	if err != nil {
		return fail("request limit exceeded")
	}
	n, err := out.Write(request)
	if err != nil || n != len(request) {
		_, _ = io.WriteString(diag, "foundry change prepare: output failed\n")
		return 3
	}
	return 0
}
