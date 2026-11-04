package operatorcli

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
	"unicode/utf8"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/acceptanceauthor"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/firstdesignreview"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/intake"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/proposalreplay"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/scenarios"
)

var errPrepareLimit = errors.New("prepared request limit")

// Caller-selected paths belong to a trusted single-writer host. The leaf and
// opened-file checks reject static substitution, not hostile concurrent writes.
func readRegular(path string, limit int64) ([]byte, error) {
	leaf, err := os.Lstat(path)
	if err != nil || !leaf.Mode().IsRegular() || leaf.Size() > limit {
		return nil, errors.New("invalid input leaf")
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	opened, err := f.Stat()
	if err != nil || !opened.Mode().IsRegular() || !os.SameFile(leaf, opened) || opened.Size() > limit {
		return nil, errors.New("input leaf changed")
	}
	raw, err := io.ReadAll(io.LimitReader(f, limit+1))
	if err != nil || int64(len(raw)) > limit || !utf8.Valid(raw) {
		return nil, errors.New("invalid input bytes")
	}
	return raw, nil
}

type observations struct {
	Requirements  *string `json:"requirements"`
	Model         *string `json:"model"`
	ScenarioSuite *string `json:"scenarioSuite"`
}

// Native validation happens first, so decoding only locates the one accepted
// top-level value. Offsets preserve all original bytes outside that value,
// including whitespace, escaped keys and unrelated nested "observed" strings.
func observedRange(raw []byte) (int, int, observations, error) {
	d := json.NewDecoder(bytes.NewReader(raw))
	if _, err := d.Token(); err != nil {
		return 0, 0, observations{}, err
	}
	for d.More() {
		key, err := d.Token()
		if err != nil {
			return 0, 0, observations{}, err
		}
		var value json.RawMessage
		if err := d.Decode(&value); err != nil {
			return 0, 0, observations{}, err
		}
		if key == "observed" {
			var observed observations
			if err := json.Unmarshal(value, &observed); err != nil {
				return 0, 0, observations{}, err
			}
			end := int(d.InputOffset())
			return end - len(value), end, observed, nil
		}
	}
	return 0, 0, observations{}, errors.New("missing observations")
}

func prepare(args []string, _ io.Reader, out, diag io.Writer) int {
	fail := func(message string) int { _, _ = io.WriteString(diag, "foundry prepare: "+message+"\n"); return 2 }
	limits := map[string]int64{
		"--intake": intake.MaxSourceBytes, "--requirements": intake.MaxArtifactBytes,
		"--model": intake.MaxArtifactBytes, "--suite": scenarios.MaxSourceBytes,
		"--manifest": acceptanceauthor.MaxManifestBytes,
	}
	paths := make(map[string]string)
	if len(args) != 2*len(limits) {
		return fail("invalid arguments")
	}
	for i := 0; i < len(args); i += 2 {
		_, known := limits[args[i]]
		_, duplicate := paths[args[i]]
		if !known || duplicate {
			return fail("invalid arguments")
		}
		paths[args[i]] = args[i+1]
	}
	original, err := readRegular(paths["--intake"], limits["--intake"])
	if err != nil {
		return fail("input rejected")
	}
	if _, err := intake.Execute(original); err != nil {
		return fail("input rejected")
	}
	texts := make(map[string]string)
	for _, flag := range []string{"--requirements", "--model", "--suite", "--manifest"} {
		b, err := readRegular(paths[flag], limits[flag])
		if err != nil {
			return fail("input rejected")
		}
		texts[flag] = string(b)
	}
	start, end, old, err := observedRange(original)
	if err != nil {
		return fail("input rejected")
	}
	requirements, model, suite := texts["--requirements"], texts["--model"], texts["--suite"]
	for _, pair := range []struct {
		old  *string
		text string
	}{{old.Requirements, requirements}, {old.Model, model}, {old.ScenarioSuite, suite}} {
		if pair.old != nil && *pair.old != pair.text {
			return fail("observation conflict")
		}
	}
	replacement, err := json.Marshal(observations{&requirements, &model, &suite})
	if err != nil {
		return fail("input rejected")
	}
	if len(original)-(end-start)+len(replacement) > intake.MaxSourceBytes {
		return fail("request limit exceeded")
	}
	var next bytes.Buffer
	next.Write(original[:start])
	next.Write(replacement)
	next.Write(original[end:])
	if _, err := intake.Execute(next.Bytes()); err != nil {
		return fail("input rejected")
	}
	proposal, err := encodePrepared(proposalreplay.Request{
		FormatVersion: proposalreplay.RequestVersion, Model: model, Suite: suite, Manifest: texts["--manifest"],
	}, proposalreplay.MaxRequestBytes)
	if err != nil {
		return fail("request limit exceeded")
	}
	request, err := encodePrepared(firstdesignreview.Request{
		FormatVersion: firstdesignreview.RequestVersion, IntakeRequest: next.String(), ProposalRequest: string(proposal),
	}, firstdesignreview.MaxRequestBytes)
	if err != nil {
		return fail("request limit exceeded")
	}
	n, err := out.Write(request)
	if err != nil || n != len(request) {
		_, _ = io.WriteString(diag, "foundry prepare: output failed\n")
		return 3
	}
	return 0
}

func encodePrepared(value any, limit int) ([]byte, error) {
	var out bytes.Buffer
	if err := json.NewEncoder(&out).Encode(value); err != nil {
		return nil, err
	}
	if out.Len() > limit {
		return nil, errPrepareLimit
	}
	return out.Bytes(), nil
}
