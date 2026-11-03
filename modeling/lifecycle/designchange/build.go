package designchange

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/artifactbundle"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/canonical"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/compiler"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/jsonformat"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/model"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/modelpatch"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/qualification"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/strictjson"
)

const guide = "Example-only design-change project; authority: none.\nSupplied contents may be private. Hashes establish consistency, not authorship, authentication, human review, approval, freshness or execution.\nSuite status: unvalidated and unexecuted. Suite binding: not inspected.\nEven malformed, non-JSON or differently bound suites may be present; export does not imply candidate replay will pass. No suite or expectation is rewritten.\nqualification.zip is a separate inert offline project; this tool never extracts or executes it.\n"

func hash(b []byte) string { s := sha256.Sum256(b); return hex.EncodeToString(s[:]) }
func archiveError(e error) error {
	if e == nil {
		return nil
	}
	if errors.Is(e, artifactbundle.ErrLimit) {
		return ErrLimit
	}
	return ErrInvalid
}
func encode(v any) ([]byte, error) {
	var b bytes.Buffer
	e := json.NewEncoder(&b)
	e.SetEscapeHTML(false)
	if e.Encode(v) != nil {
		return nil, ErrInvalid
	}
	return b.Bytes(), nil
}

// sourceNames derives the complete allowed set and transport name bound from
// qualification. Exact source sizes/hashes remain the qualifier's responsibility.
func sourceNames() ([]string, map[string]bool, int) {
	names, set, max, _ := sourceNamesForRevision(qualification.SourceRevision)
	return names, set, max
}
func sourceNamesForRevision(revision string) ([]string, map[string]bool, int, error) {
	names, err := qualification.SourceNamesForRevision(revision)
	if err != nil {
		return nil, nil, 0, ErrInvalid
	}
	set := make(map[string]bool, len(names))
	max := 0
	for _, n := range names {
		set[n] = true
		if len(n) > max {
			max = len(n)
		}
	}
	return names, set, max, nil
}
func checkSourcesForRevision(s []artifactbundle.File, revision string) error {
	names, set, max, err := sourceNamesForRevision(revision)
	if err != nil {
		return err
	}
	if len(s) > len(names) {
		return ErrLimit
	}
	if len(s) != len(names) {
		return ErrInvalid
	}
	seen := make(map[string]bool, len(s))
	for _, f := range s {
		if len(f.Name) > max || !set[f.Name] || seen[f.Name] {
			return ErrInvalid
		}
		seen[f.Name] = true
	}
	total := 0
	for _, f := range s {
		if len(f.Data) > MaxSourceBytes {
			return ErrLimit
		}
		total += len(f.Data)
		if total > MaxSourcesBytes {
			return ErrLimit
		}
	}
	return nil
}
func preflightForRevision(request, suite []byte, sources []artifactbundle.File, revision string) error {
	names, _, _, err := sourceNamesForRevision(revision)
	if err != nil {
		return err
	}
	if len(request) > MaxRequestBytes || len(suite) > MaxSuiteBytes || len(sources) > len(names) {
		return ErrLimit
	}
	if len(request) == 0 || len(suite) == 0 {
		return ErrInvalid
	}
	return checkSourcesForRevision(sources, revision)
}

// acceptedRequest delegates the whole patch to the existing planner before
// extracting any identity. It does not interpret operations or suite contents.
func acceptedRequest(request []byte) ([]byte, []byte, []byte, plannerResultDTO, error) {
	result, e := modelpatch.Execute(request)
	if e != nil {
		return nil, nil, nil, plannerResultDTO{}, fmt.Errorf("change request: %w", e)
	}
	var r plannerResultDTO
	if strictjson.Decode(result, &r, modelpatch.MaxResultBytes, 64) != nil {
		return nil, nil, nil, r, ErrInvalid
	}
	// Execute already validated the exact carrier; this extraction preserves the
	// source token's spelling while excluding its surrounding carrier whitespace.
	var carrier struct {
		Before json.RawMessage `json:"before"`
	}
	if json.Unmarshal(request, &carrier) != nil || len(carrier.Before) == 0 {
		return nil, nil, nil, r, ErrInvalid
	}
	before := []byte(carrier.Before)
	if len(before) > maxModelBytes {
		return nil, nil, nil, r, ErrLimit
	}
	m, e := jsonformat.Decode(r.After)
	if e != nil {
		return nil, nil, nil, r, ErrInvalid
	}
	candidate, e := jsonformat.Encode(m)
	if e != nil {
		return nil, nil, nil, r, ErrInvalid
	}
	if len(candidate) > maxModelBytes {
		return nil, nil, nil, r, ErrLimit
	}
	d, e := canonical.Digest(m)
	if e != nil || d != r.AfterDigest {
		return nil, nil, nil, r, ErrInvalid
	}
	bm, e := jsonformat.Decode(before)
	if e != nil {
		return nil, nil, nil, r, ErrInvalid
	}
	bd, e := canonical.Digest(bm)
	if e != nil || bd != r.BeforeDigest {
		return nil, nil, nil, r, ErrInvalid
	}
	return result, before, candidate, r, nil
}

// Build returns one inert canonical archive. It never validates or executes the
// opaque suite and never inspects its binding. Inputs must not change concurrently.
func Build(request, suite []byte, sources []artifactbundle.File) ([]byte, error) {
	return buildForRevision(request, suite, sources, qualification.SourceRevision)
}

// Only native verification selects a historical policy; public builds stay current.
func buildForRevision(request, suite []byte, sources []artifactbundle.File, revision string) ([]byte, error) {
	if e := preflightForRevision(request, suite, sources, revision); e != nil {
		return nil, e
	}
	result, before, candidate, r, e := acceptedRequest(request)
	if e != nil {
		return nil, e
	}
	inner, e := qualification.BuildForRevision(revision, candidate, suite, sources)
	if e != nil {
		return nil, archiveError(e)
	}
	if len(inner) > MaxInnerBytes {
		return nil, ErrLimit
	}
	if e = qualification.Verify(inner, hash(candidate), hash(suite)); e != nil {
		return nil, archiveError(e)
	}
	meta := metadataDTO{Profile, "example-only", "none", hash(request), hash(before), r.BeforeDigest, hash(candidate), r.AfterDigest, hash(suite), hash(result), hash(inner), jsonformat.FormatVersion, model.ProfileVersion, compiler.GeneratorVersion, compiler.ToolchainVersion, revision, "unvalidated and unexecuted", "not inspected"}
	mb, e := encode(meta)
	if e != nil {
		return nil, e
	}
	if len(mb) > maxMetadataBytes || len(guide) > maxGuideBytes {
		return nil, ErrLimit
	}
	files := []artifactbundle.File{{Name: "GUIDE.txt", Data: []byte(guide)}, {Name: "before.json", Data: before}, {Name: "candidate.json", Data: candidate}, {Name: "change-request.json", Data: request}, {Name: "design-change.json", Data: mb}, {Name: "planner-result.json", Data: result}, {Name: "qualification.zip", Data: inner}}
	b, e := artifactbundle.Build(Profile, files)
	if e != nil {
		return nil, archiveError(e)
	}
	return b, nil
}
