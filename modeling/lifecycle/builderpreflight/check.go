// Package builderpreflight compares named input bytes without executing them.
package builderpreflight

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/strictjson"
)

const (
	MaxManifestBytes = 64 << 10
	MaxArtifactBytes = 1 << 20
	MaxArtifacts     = 64
)

// ErrInvalid rejects the request without producing an artifact report.
var ErrInvalid = errors.New("invalid builder input preflight")

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

type observation struct {
	ID             string  `json:"id"`
	Role           string  `json:"role"`
	Path           string  `json:"path"`
	ExpectedSHA256 string  `json:"expectedSHA256"`
	Status         string  `json:"status"`
	ObservedSHA256 *string `json:"observedSHA256"`
	ObservedBytes  *int64  `json:"observedBytes"`
}

type report struct {
	FormatVersion   string        `json:"formatVersion"`
	ManifestSHA256  string        `json:"manifestSHA256"`
	Outcome         string        `json:"outcome"`
	ExecutionStatus string        `json:"executionStatus"`
	ReviewStatus    string        `json:"reviewStatus"`
	Authority       string        `json:"authority"`
	Artifacts       []observation `json:"artifacts"`
}

func validDigest(value string) bool {
	if len(value) != 64 {
		return false
	}
	for _, c := range value {
		if !(c >= '0' && c <= '9' || c >= 'a' && c <= 'f') {
			return false
		}
	}
	return true
}

func validID(value string) bool {
	if len(value) < 1 || len(value) > 64 {
		return false
	}
	for _, c := range value {
		if !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '_' || c == '-') {
			return false
		}
	}
	return true
}

func validRole(value string) bool {
	switch value {
	case "requirements", "model", "interface", "acceptance", "harness", "rubric", "packet", "instructions", "claim":
		return true
	}
	return false
}

func validPath(value string) bool {
	if len(value) < 1 || len(value) > 1024 || strings.ContainsAny(value, "\\:") {
		return false
	}
	for _, c := range value {
		if unicode.IsControl(c) {
			return false
		}
	}
	for _, component := range strings.Split(value, "/") {
		if component == "" || component == "." || component == ".." {
			return false
		}
	}
	return true
}

// Check compares a fully validated, separately pinned manifest with only its
// named files beneath an explicit root. It performs no discovery, writes,
// semantic interpretation, subprocess calls or candidate execution. Callers
// must authorize the root and listed reads independently of manifest contents.
func Check(raw []byte, rootPath, manifestSHA256 string) ([]byte, error) {
	if len(raw) > MaxManifestBytes || !validDigest(manifestSHA256) {
		return nil, ErrInvalid
	}
	actual := sha256.Sum256(raw)
	if hex.EncodeToString(actual[:]) != manifestSHA256 {
		return nil, ErrInvalid
	}
	var source manifest
	if strictjson.Decode(raw, &source, MaxManifestBytes, 8) != nil || source.FormatVersion != "model-builder-inputs/v1" || len(source.Artifacts) == 0 || len(source.Artifacts) > MaxArtifacts {
		return nil, ErrInvalid
	}
	ids, paths := map[string]bool{}, map[string]bool{}
	for _, item := range source.Artifacts {
		if !validID(item.ID) || !validRole(item.Role) || !validPath(item.Path) || !validDigest(item.SHA256) || ids[item.ID] || paths[item.Path] {
			return nil, ErrInvalid
		}
		ids[item.ID], paths[item.Path] = true, true
	}
	// No filesystem operation may precede full manifest validation.
	if !filepath.IsAbs(rootPath) {
		return nil, ErrInvalid
	}
	root, err := openRoot(filepath.Clean(rootPath))
	if err != nil {
		return nil, ErrInvalid
	}
	defer root.Close()
	result := report{
		FormatVersion: "model-builder-input-report/v1", ManifestSHA256: manifestSHA256,
		Outcome: "matched", ExecutionStatus: "not-run", ReviewStatus: "not-established", Authority: "none",
		Artifacts: make([]observation, 0, len(source.Artifacts)),
	}
	for _, item := range source.Artifacts {
		observed := inspect(root, item)
		if observed.Status != "matched" {
			result.Outcome = "blocked"
		}
		result.Artifacts = append(result.Artifacts, observed)
	}
	encoded, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(encoded, '\n'), nil
}
