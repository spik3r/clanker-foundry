// Package qualification builds example-only offline candidate projects.
package qualification

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"runtime"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/artifactbundle"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/canonical"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/compiler"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/jsonformat"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/model"
)

var ErrInvalid = errors.New("invalid qualification project")

const maxInput = 1 << 20
const maxSourceName = len("scenarios/report_decode.go")
const SourceRevision = "3d6d17188cc2723cfe04649d0f8b4220a37b1afe"
const guide = "Example-only candidate project; no review or execution authority.\nThe literal suite has not been validated or executed. Use a trusted single-writer host.\nOffline with Go 1.27.3:\n  GOTOOLCHAIN=local GOWORK=off GOENV=off GOFLAGS= GOPROXY=off GOSUMDB=off go test -buildvcs=false ./generated\n  GOTOOLCHAIN=local GOWORK=off GOENV=off GOFLAGS= GOPROXY=off GOSUMDB=off go run -buildvcs=false ./cmd/modelplay < suite.json\nProject identity comes from the verified manifest and supplied hashes, not ambient VCS stamping.\nDo not change the suite or model binding to force a pass.\n"

type pin struct {
	name string
	size int
	hash string
}

var pins = []pin{
	{"go.mod", 71, "cc128fce13193d3066bdeb08cca9db627fb6c3120967c3beff254a514f2a5fac"},
	{"model/ir.go", 9385, "13f4ad2980b1ca523b07bdbc8ebeb0e0300f8df81faa8dcd1cdec5ccef007aeb"},
	{"canonical/canonical.go", 1946, "2c3cba3fd53adf8ca5a4cc2ce3d96b51457742c47b6ab4fcedba219e19001a06"},
	{"strictjson/decode.go", 4236, "a72ded4af99ffba5f5cca2931c6cace072634ee2e42eb00cd9a5988c691f1ee2"},
	{"compiler/generate.go", 12315, "e17610a27748a5d3baedf4f59c2183cb184f7e2681b5532c6f9d8d3d1eda6a38"},
	{"scenarios/types.go", 5904, "147b12bd1726fec156ba08ec5666a9a362e1563115151ae57bf9face70f86eec"},
	{"scenarios/decode.go", 7047, "e6119229899d3084820d72a777d1771037f871b14554630067b7da022cdedc15"},
	{"scenarios/run.go", 10114, "46d41c5d27d87f71f7191f18d182a99c3ca491126afb601599067692c3379dde"},
	{"scenarios/runner.go", 4005, "e85b2146d7c1d16269ff886f6c67084de5e263db8b01a7676b8c7491ed8fb59c"},
	{"scenarios/report_decode.go", 3626, "f14816a19faacf707e23b01d5d16b26be8e8c03a9f5f9af801029b89fb2ceddf"},
	{"cmd/modelplay/main.go", 1613, "925ec202c74aba5d50eb2bfb5dc54ea22387ccec90e969f1a7ebe1ecbd9a0c63"},
}

// SourceNames returns a fresh list of the only accepted source paths.
func SourceNames() []string {
	out := make([]string, len(pins))
	for i, p := range pins {
		out[i] = p.name
	}
	return out
}
func hash(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }

type identity struct {
	Format         string `json:"format"`
	Profile        string `json:"profile"`
	Generator      string `json:"generator"`
	Toolchain      string `json:"toolchain"`
	ModelDigest    string `json:"modelDigest"`
	SourceRevision string `json:"sourceRevision"`
	SuiteStatus    string `json:"suiteStatus"`
}

// Build accepts only the pinned source inventory. The suite is copied unvalidated.
// Callers must not mutate candidate, suite, or source bytes during this call.
func Build(candidate, suite []byte, sources []artifactbundle.File) ([]byte, error) {
	if runtime.Version() != compiler.ToolchainVersion {
		return nil, ErrInvalid
	}
	if len(candidate) == 0 || len(candidate) > maxInput || len(suite) == 0 || len(suite) > maxInput || len(sources) != len(pins) {
		return nil, ErrInvalid
	}
	byName := make(map[string][]byte, len(sources))
	for _, f := range sources {
		// Bound and match the caller-controlled name before using it as a map key.
		if len(f.Name) > maxSourceName {
			return nil, ErrInvalid
		}
		var matched *pin
		for i := range pins {
			if f.Name == pins[i].name {
				matched = &pins[i]
				break
			}
		}
		if matched == nil || len(f.Data) != matched.size {
			return nil, ErrInvalid
		}
		if _, ok := byName[f.Name]; ok {
			return nil, ErrInvalid
		}
		if hash(f.Data) != matched.hash {
			return nil, ErrInvalid
		}
		byName[f.Name] = f.Data
	}
	files := make([]artifactbundle.File, 0, len(pins)+6)
	for _, p := range pins {
		b, ok := byName[p.name]
		if !ok || len(b) != p.size || hash(b) != p.hash {
			return nil, ErrInvalid
		}
		files = append(files, artifactbundle.File{Name: p.name, Data: b})
	}
	m, e := jsonformat.Decode(candidate)
	if e != nil {
		return nil, ErrInvalid
	}
	digest, e := canonical.Digest(m)
	if e != nil {
		return nil, ErrInvalid
	}
	code, tests, e := compiler.Generate(m)
	if e != nil {
		return nil, ErrInvalid
	}
	id := identity{jsonformat.FormatVersion, model.ProfileVersion, compiler.GeneratorVersion, compiler.ToolchainVersion, digest, SourceRevision, "unvalidated and unexecuted"}
	meta, e := json.Marshal(id)
	if e != nil {
		return nil, ErrInvalid
	}
	meta = append(meta, '\n')
	files = append(files, artifactbundle.File{Name: "generated/lifecycle.go", Data: code}, artifactbundle.File{Name: "generated/model_paths_test.go", Data: tests}, artifactbundle.File{Name: "candidate.json", Data: candidate}, artifactbundle.File{Name: "suite.json", Data: suite}, artifactbundle.File{Name: "qualification.json", Data: meta}, artifactbundle.File{Name: "GUIDE.txt", Data: []byte(guide)})
	result, e := artifactbundle.Build(artifactbundle.QualificationProfile, files)
	if e != nil {
		return nil, fmt.Errorf("qualification bundle: %w", e)
	}
	return result, nil
}

// Verify requires the complete exact domain inventory and regenerates artifacts.
// Callers must not mutate raw archive bytes during this call.
func Verify(raw []byte, expectedCandidateSHA256, expectedSuiteSHA256 string) error {
	files, e := artifactbundle.Verify(raw, artifactbundle.QualificationProfile)
	if e != nil {
		return ErrInvalid
	}
	byName := make(map[string][]byte, len(files))
	for _, f := range files {
		byName[f.Name] = f.Data
	}
	sources := make([]artifactbundle.File, 0, len(pins))
	for _, p := range pins {
		b, ok := byName[p.name]
		if !ok {
			return ErrInvalid
		}
		sources = append(sources, artifactbundle.File{Name: p.name, Data: b})
	}
	candidate, ok := byName["candidate.json"]
	if !ok {
		return ErrInvalid
	}
	suite, ok := byName["suite.json"]
	if !ok {
		return ErrInvalid
	}
	if hash(candidate) != expectedCandidateSHA256 || hash(suite) != expectedSuiteSHA256 {
		return ErrInvalid
	}
	rebuilt, e := Build(candidate, suite, sources)
	if e != nil || !bytes.Equal(raw, rebuilt) {
		return ErrInvalid
	}
	return nil
}
