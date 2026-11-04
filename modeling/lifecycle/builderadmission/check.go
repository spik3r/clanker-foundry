// Package builderadmission checks the fixed exercise's exact file frontier and
// byte identities. It does not parse, compile, execute or approve a candidate.
package builderadmission

import (
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"path/filepath"
	"unicode"
	"unicode/utf8"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/builderpreflight"
)

const (
	FormatVersion     = "model-builder-admission-report/v1"
	ManifestSHA256    = "e4fac14e2196dee8dde00bb7e9101b3324cfd4f2f564850a71ca21b76bac9812"
	MaxCandidateBytes = 1 << 20
	MaxEntries        = 64
	MaxReportBytes    = 1 << 20
)

var (
	ErrInvalid  = errors.New("invalid builder admission arguments or root")
	ErrBoundary = errors.New("incomplete builder admission")
	//go:embed inputs.json
	originalManifest []byte
)
var allowedNames = [...]string{"README.md", "acceptance_test.go", "candidate.go", "contract.go", "go.mod", "rubric.md"}

type entry struct {
	Name string `json:"name"`
	Kind string `json:"kind"`
}
type observation struct {
	Path           string  `json:"path"`
	Status         string  `json:"status"`
	ObservedSHA256 *string `json:"observedSHA256"`
	ObservedBytes  *int64  `json:"observedBytes"`
}
type report struct {
	FormatVersion           string      `json:"formatVersion"`
	Outcome                 string      `json:"outcome"`
	Authority               string      `json:"authority"`
	ExecutionStatus         string      `json:"executionStatus"`
	ReviewStatus            string      `json:"reviewStatus"`
	ExerciseRoot            string      `json:"exerciseRoot"`
	ExpectedCandidateSHA256 string      `json:"expectedCandidateSHA256"`
	FrozenManifestSHA256    string      `json:"frozenManifestSHA256"`
	PreflightReportSHA256   string      `json:"preflightReportSHA256"`
	PreflightReportJSON     string      `json:"preflightReportJSON"`
	FrontierComplete        bool        `json:"frontierComplete"`
	Frontier                []entry     `json:"frontier"`
	MissingNames            []string    `json:"missingNames"`
	UnexpectedNames         []string    `json:"unexpectedNames"`
	Candidate               observation `json:"candidate"`
}

func hash(b []byte) string { d := sha256.Sum256(b); return hex.EncodeToString(d[:]) }
func validPin(s string) bool {
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
func validRoot(s string) bool {
	if !filepath.IsAbs(s) || !utf8.ValidString(s) {
		return false
	}
	for _, c := range s {
		if unicode.IsControl(c) {
			return false
		}
	}
	return true
}

// Check performs only explicitly named original/candidate reads and a bounded
// direct-child filename/metadata census. Keep the exercise exclusive throughout;
// these observations are not an atomic snapshot or a later build authorization.
func Check(exerciseRoot, expectedCandidateSHA256 string) ([]byte, error) {
	if !validRoot(exerciseRoot) || !validPin(expectedCandidateSHA256) {
		return nil, ErrInvalid
	}
	if hash(originalManifest) != ManifestSHA256 {
		return nil, ErrBoundary
	}
	root, path, identity, err := openRoot(exerciseRoot)
	if err != nil {
		return nil, err
	}
	defer root.Close()
	children, infos, err := census(root)
	if err != nil {
		return nil, err
	}
	missing, unexpected := []string{}, []string{}
	known := make(map[string]bool, len(allowedNames))
	regular := true
	for _, name := range allowedNames {
		known[name] = true
		info, ok := infos[name]
		if !ok {
			missing = append(missing, name)
		} else if !info.Mode().IsRegular() {
			regular = false
		}
	}
	for _, child := range children {
		if !known[child.Name] {
			unexpected = append(unexpected, child.Name)
		}
	}
	// A case-insensitive host can resolve a missing exact spelling to an unexpected
	// child. Native preflight must not open that alias and read unlisted contents.
	for _, name := range missing {
		if name == "candidate.go" {
			continue
		}
		if _, err := root.Lstat(name); err == nil || !isMissing(err) {
			return nil, ErrBoundary
		}
	}
	native, err := builderpreflight.Check(originalManifest, exerciseRoot, ManifestSHA256)
	if err != nil {
		return nil, ErrBoundary
	}
	var status struct{ Outcome string }
	if json.Unmarshal(native, &status) != nil || (status.Outcome != "matched" && status.Outcome != "blocked") {
		return nil, ErrBoundary
	}
	candidate := inspectCandidate(root, infos["candidate.go"], expectedCandidateSHA256)
	// Recheck the same finite frontier and identities after all content reads.
	again, after, err := census(root)
	if err != nil || !sameCensus(children, infos, again, after) || !sameRoot(root, path, identity) {
		return nil, ErrBoundary
	}
	outcome := "blocked"
	if len(missing) == 0 && len(unexpected) == 0 && regular && status.Outcome == "matched" && candidate.Status == "matched" {
		outcome = "matched"
	}
	result := report{FormatVersion, outcome, "none", "not-run", "not-established", exerciseRoot, expectedCandidateSHA256, ManifestSHA256, hash(native), string(native), true, children, missing, unexpected, candidate}
	encoded, err := json.MarshalIndent(result, "", "  ")
	if err != nil || len(encoded)+1 > MaxReportBytes {
		return nil, ErrBoundary
	}
	return append(encoded, '\n'), nil
}
