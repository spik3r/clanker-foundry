// Package builderdemo prepares the fixed Go exercise without dispatch or execution.
package builderdemo

import (
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/builderpreflight"
)

var (
	ErrInvalid = errors.New("invalid builder demo preparation")
	ErrBlocked = errors.New("builder demo inputs blocked")
)

// These bytes and their pin were frozen before the preparation implementation.
//
//go:embed inputs.json
var sourceManifest []byte

const sourceManifestSHA256 = "6a87f4db69a9c43376b405aca0d1252a662a3a4cdc7496329f2b41fac2896032"
const acceptancePrefix = "examples/model-builder-go/acceptance/"

type artifact struct {
	ID     string `json:"id"`
	Role   string `json:"role"`
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
}
type manifest struct {
	FormatVersion string     `json:"formatVersion"`
	Artifacts     []artifact `json:"artifacts"`
}

// Receipt separates preparation from delivery, review, authority and execution.
// Reports are complete native preflight objects; nil means that check did not run.
// A workspace remains available after an incomplete preparation or lost delivery.
type Receipt struct {
	FormatVersion        string          `json:"formatVersion"`
	Outcome              string          `json:"outcome"`
	Workspace            *string         `json:"workspace"`
	HandoffPath          *string         `json:"handoffPath"`
	SourceManifestSHA256 string          `json:"sourceManifestSHA256"`
	SourceReport         json.RawMessage `json:"sourceReport"`
	InputsReport         json.RawMessage `json:"inputsReport"`
	ExerciseReport       json.RawMessage `json:"exerciseReport"`
	ExecutionStatus      string          `json:"executionStatus"`
	ReviewStatus         string          `json:"reviewStatus"`
	Authority            string          `json:"authority"`
}

func digest(raw []byte) string {
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}
func validAbsolute(path string) bool {
	if !utf8.ValidString(path) || !filepath.IsAbs(path) {
		return false
	}
	for _, r := range path {
		if unicode.IsControl(r) {
			return false
		}
	}
	return true
}
func matched(raw []byte) bool {
	var r struct{ Outcome string }
	return json.Unmarshal(raw, &r) == nil && r.Outcome == "matched"
}

// Prepare creates a fresh private worked-example workspace. It reads only the
// fixed inventory, preserves original bytes, and records goBinary as data only.
// Callers must separately authorize the named reads and new workspace writes.
// The returned draft grants no builder permission. No process is started.
func Prepare(sourceRoot, workParent, goBinary string) (Receipt, error) {
	var result Receipt
	if !validAbsolute(sourceRoot) || !validAbsolute(workParent) || !validAbsolute(goBinary) {
		return result, ErrInvalid
	}
	var inventory manifest
	if digest(sourceManifest) != sourceManifestSHA256 || json.Unmarshal(sourceManifest, &inventory) != nil {
		return result, errors.New("fixed input inventory is invalid")
	}
	source, sourceInfo, err := openDirectory(sourceRoot)
	if err != nil {
		return result, ErrInvalid
	}
	defer source.Close()
	parent, parentInfo, err := openDirectory(workParent)
	if err != nil {
		return result, ErrInvalid
	}
	defer parent.Close()
	// Resolve actual host traversal before comparing locations. Lexical cleaning
	// of an ancestor alias followed by .. can select a different directory.
	sourcePath, err := physicalDirectory(sourceRoot, sourceInfo)
	if err != nil {
		return result, ErrInvalid
	}
	parentPath, err := physicalDirectory(workParent, parentInfo)
	if err != nil {
		return result, ErrInvalid
	}
	rel, err := filepath.Rel(sourcePath, parentPath)
	if err != nil || rel == "." || rel != ".." && !strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return result, ErrInvalid
	}
	result = Receipt{
		FormatVersion: "model-builder-demo-preparation/v1", Outcome: "incomplete",
		SourceManifestSHA256: sourceManifestSHA256, ExecutionStatus: "not-run",
		ReviewStatus: "not-established", Authority: "none",
	}
	raw, err := builderpreflight.Check(sourceManifest, sourceRoot, sourceManifestSHA256)
	if err != nil {
		return Receipt{}, ErrInvalid
	}
	result.SourceReport = raw
	if !matched(raw) {
		result.Outcome = "blocked"
		return result, fmt.Errorf("%w: source preflight did not match", ErrBlocked)
	}

	// Collect and recheck bytes before any output creation. An earlier report
	// is not a promise about bytes read later; never derive new expected hashes.
	payloads := make([][]byte, len(inventory.Artifacts))
	for i, item := range inventory.Artifacts {
		payloads[i], err = collect(source, item)
		if err != nil {
			result.Outcome = "blocked"
			return result, err
		}
	}
	if !sameDirectory(sourceRoot, sourceInfo) {
		result.Outcome = "blocked"
		return result, fmt.Errorf("%w: source root changed during collection", ErrBlocked)
	}
	if !sameDirectory(workParent, parentInfo) || !sameDirectory(parentPath, parentInfo) {
		return result, errors.New("work parent changed before creation")
	}
	workspace, err := os.MkdirTemp(parentPath, "foundry-builder-")
	if err != nil {
		return result, errors.New("workspace creation failed")
	}
	result.Workspace = &workspace
	output, err := os.OpenRoot(workspace)
	if err != nil {
		return result, errors.New("new workspace could not be opened")
	}
	defer output.Close()
	for _, dir := range []string{"inputs", "exercise", "cache", "tmp", "modcache"} {
		if err := output.Mkdir(dir, 0700); err != nil {
			return result, fmt.Errorf("workspace directory creation failed: %s", dir)
		}
	}
	exercise := manifest{FormatVersion: "model-builder-inputs/v1", Artifacts: []artifact{}}
	for i, item := range inventory.Artifacts {
		if err := writeNew(output, "inputs/"+item.Path, payloads[i]); err != nil {
			return result, fmt.Errorf("input copy failed: %s", item.Path)
		}
		if strings.HasPrefix(item.Path, acceptancePrefix) {
			item.Path = strings.TrimPrefix(item.Path, acceptancePrefix)
			exercise.Artifacts = append(exercise.Artifacts, item)
			if err := writeNew(output, "exercise/"+item.Path, payloads[i]); err != nil {
				return result, fmt.Errorf("exercise copy failed: %s", item.Path)
			}
		}
	}
	exerciseRaw, err := json.MarshalIndent(exercise, "", "  ")
	if err != nil {
		return result, errors.New("exercise inventory encoding failed")
	}
	exerciseRaw = append(exerciseRaw, '\n')
	for _, file := range []struct {
		name string
		raw  []byte
	}{{"source-manifest.json", sourceManifest}, {"source-report.json", raw}, {"exercise-manifest.json", exerciseRaw}} {
		if err := writeNew(output, file.name, file.raw); err != nil {
			return result, fmt.Errorf("evidence write failed: %s", file.name)
		}
	}
	result.InputsReport, err = builderpreflight.Check(sourceManifest, filepath.Join(workspace, "inputs"), sourceManifestSHA256)
	if err != nil {
		return result, errors.New("copied input preflight failed")
	}
	if err := writeNew(output, "inputs-report.json", result.InputsReport); err != nil {
		return result, errors.New("copied input report write failed")
	}
	result.ExerciseReport, err = builderpreflight.Check(exerciseRaw, filepath.Join(workspace, "exercise"), digest(exerciseRaw))
	if err != nil {
		return result, errors.New("exercise preflight failed")
	}
	if err := writeNew(output, "exercise-report.json", result.ExerciseReport); err != nil {
		return result, errors.New("exercise report write failed")
	}
	if !matched(result.InputsReport) || !matched(result.ExerciseReport) {
		return result, errors.New("destination bytes did not match")
	}
	draft, err := handoff(workspace, sourceRoot, goBinary, inventory, exercise, digest(exerciseRaw))
	if err != nil {
		return result, errors.New("handoff encoding failed")
	}
	if err := writeNew(output, "handoff.md", draft); err != nil {
		return result, errors.New("handoff write failed")
	}
	handoffPath := filepath.Join(workspace, "handoff.md")
	result.HandoffPath, result.Outcome = &handoffPath, "prepared-draft"
	return result, nil
}
