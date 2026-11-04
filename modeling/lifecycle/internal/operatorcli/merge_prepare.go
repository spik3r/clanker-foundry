package operatorcli

import (
	"bytes"
	"encoding/json"
	"io"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/acceptanceauthor"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/custommergerehearsal"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/modelmerge"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/scenarios"
)

// Preparation preserves supplied artifact bytes. Native checking and the
// separately reviewed candidate identity remain the operator's explicit inputs.
func prepareMerge(args []string, _ io.Reader, out, diag io.Writer) int {
	fail := func(message string) int {
		_, _ = io.WriteString(diag, "foundry merge prepare: "+message+"\n")
		return 2
	}
	limits := map[string]int64{
		"--merge-request":      modelmerge.MaxRequestBytes,
		"--base-suite":         scenarios.MaxSourceBytes,
		"--base-manifest":      acceptanceauthor.MaxManifestBytes,
		"--left-suite":         scenarios.MaxSourceBytes,
		"--left-manifest":      acceptanceauthor.MaxManifestBytes,
		"--right-suite":        scenarios.MaxSourceBytes,
		"--right-manifest":     acceptanceauthor.MaxManifestBytes,
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
	for _, flag := range []string{
		"--merge-request", "--base-suite", "--base-manifest", "--left-suite", "--left-manifest",
		"--right-suite", "--right-manifest", "--candidate-suite", "--candidate-manifest",
	} {
		raw, err := readRegular(values[flag], limits[flag])
		if err != nil {
			return fail("input rejected")
		}
		texts[flag] = string(raw)
	}
	pair := func(role string) custommergerehearsal.Expectations {
		return custommergerehearsal.Expectations{Suite: texts["--"+role+"-suite"], Manifest: texts["--"+role+"-manifest"]}
	}
	request := custommergerehearsal.Request{
		FormatVersion:           custommergerehearsal.RequestVersion,
		MergeRequest:            texts["--merge-request"],
		Base:                    pair("base"),
		Left:                    pair("left"),
		Right:                   pair("right"),
		Candidate:               pair("candidate"),
		ExpectedCandidateDigest: digest,
	}
	var encoded bytes.Buffer
	encoder := json.NewEncoder(&encoded)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(request); err != nil || encoded.Len() > custommergerehearsal.MaxRequestBytes {
		return fail("request limit exceeded")
	}
	n, err := out.Write(encoded.Bytes())
	if err != nil || n != encoded.Len() {
		_, _ = io.WriteString(diag, "foundry merge prepare: output failed\n")
		return 3
	}
	return 0
}
