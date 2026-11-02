package designchange

import (
	"bytes"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/artifactbundle"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/canonical"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/jsonformat"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/qualification"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/strictjson"
)

// This DTO is only a bounded schema for routing. Native qualification.Verify
// must validate every field and rebuilt byte before the selected policy is trusted.
type qualificationIdentity struct {
	Format         string `json:"format"`
	Profile        string `json:"profile"`
	Generator      string `json:"generator"`
	Toolchain      string `json:"toolchain"`
	ModelDigest    string `json:"modelDigest"`
	SourceRevision string `json:"sourceRevision"`
	SuiteStatus    string `json:"suiteStatus"`
}

func digest(s string) bool {
	if len(s) != 64 {
		return false
	}
	for _, c := range s {
		if !(c >= '0' && c <= '9' || c >= 'a' && c <= 'f') {
			return false
		}
	}
	return true
}
func validExpected(e Expected) bool {
	for _, s := range []string{e.RequestSHA256, e.BeforeSourceSHA256, e.BeforeModelDigest, e.CandidateSHA256, e.CandidateModelDigest, e.SuiteSHA256} {
		if !digest(s) {
			return false
		}
	}
	return true
}
func inventory(files []artifactbundle.File, names []string) (map[string][]byte, error) {
	if len(files) != len(names) {
		return nil, ErrInvalid
	}
	allowed := make(map[string]bool, len(names))
	for _, n := range names {
		allowed[n] = true
	}
	out := make(map[string][]byte, len(files))
	for _, f := range files {
		if !allowed[f.Name] {
			return nil, ErrInvalid
		}
		if _, ok := out[f.Name]; ok {
			return nil, ErrInvalid
		}
		out[f.Name] = f.Data
	}
	return out, nil
}
func semantic(b []byte) (string, error) {
	m, e := jsonformat.Decode(b)
	if e != nil {
		return "", ErrInvalid
	}
	d, e := canonical.Digest(m)
	if e != nil {
		return "", ErrInvalid
	}
	return d, nil
}

// Verify compares actual payload and semantic identities against external caller
// expectations, verifies the pinned inner project and rebuilds exact archive bytes.
// It establishes consistency only; suites stay unvalidated/unexecuted/not inspected.
func Verify(raw []byte, expected Expected) error {
	if !validExpected(expected) {
		return ErrInvalid
	}
	if len(raw) > MaxArchiveBytes {
		return ErrLimit
	}
	files, e := artifactbundle.Verify(raw, Profile)
	if e != nil {
		return archiveError(e)
	}
	outer, e := inventory(files, []string{"GUIDE.txt", "before.json", "candidate.json", "change-request.json", "design-change.json", "planner-result.json", "qualification.zip"})
	if e != nil {
		return e
	}
	if len(outer["qualification.zip"]) > MaxInnerBytes || len(outer["change-request.json"]) > MaxRequestBytes || len(outer["before.json"]) > maxModelBytes || len(outer["candidate.json"]) > maxModelBytes || len(outer["planner-result.json"]) > MaxRequestBytes || len(outer["design-change.json"]) > maxMetadataBytes || len(outer["GUIDE.txt"]) > maxGuideBytes {
		return ErrLimit
	}
	innerFiles, e := artifactbundle.Verify(outer["qualification.zip"], artifactbundle.QualificationProfile)
	if e != nil {
		return archiveError(e)
	}
	var metadata qualificationIdentity
	var metadataBytes []byte
	for _, f := range innerFiles {
		if f.Name == "qualification.json" {
			metadataBytes = f.Data
		}
	}
	if strictjson.Decode(metadataBytes, &metadata, 4<<10, 4) != nil {
		return ErrInvalid
	}
	names, _, _, e := sourceNamesForRevision(metadata.SourceRevision)
	if e != nil {
		return e
	}
	innerNames := append(append([]string{}, names...), "GUIDE.txt", "candidate.json", "suite.json", "qualification.json", "generated/lifecycle.go", "generated/model_paths_test.go")
	inner, e := inventory(innerFiles, innerNames)
	if e != nil {
		return e
	}
	if len(inner["candidate.json"]) > maxModelBytes || len(inner["suite.json"]) > MaxSuiteBytes {
		return ErrLimit
	}
	sources := make([]artifactbundle.File, 0, len(names))
	for _, n := range names {
		sources = append(sources, artifactbundle.File{Name: n, Data: inner[n]})
	}
	if e = checkSourcesForRevision(sources, metadata.SourceRevision); e != nil {
		return e
	}
	if hash(outer["change-request.json"]) != expected.RequestSHA256 || hash(outer["before.json"]) != expected.BeforeSourceSHA256 || hash(inner["candidate.json"]) != expected.CandidateSHA256 || hash(inner["suite.json"]) != expected.SuiteSHA256 || !bytes.Equal(outer["candidate.json"], inner["candidate.json"]) {
		return ErrInvalid
	}
	bd, e := semantic(outer["before.json"])
	if e != nil || bd != expected.BeforeModelDigest {
		return ErrInvalid
	}
	ad, e := semantic(inner["candidate.json"])
	if e != nil || ad != expected.CandidateModelDigest {
		return ErrInvalid
	}
	if e = qualification.Verify(outer["qualification.zip"], expected.CandidateSHA256, expected.SuiteSHA256); e != nil {
		return archiveError(e)
	}
	rebuilt, e := buildForRevision(outer["change-request.json"], inner["suite.json"], sources, metadata.SourceRevision)
	if e != nil {
		return e
	}
	if !bytes.Equal(raw, rebuilt) {
		return ErrInvalid
	}
	return nil
}
