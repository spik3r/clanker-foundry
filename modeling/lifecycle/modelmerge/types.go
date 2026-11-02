// Package modelmerge proposes a three-way merge of reviewed lifecycle models.
// A conflict-free candidate is not approval, execution or behavioral equivalence.
package modelmerge

import (
	"errors"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/model"
)

const (
	Version            = "lifecycle-model-merge/v1"
	FormatVersion      = "lifecycle-model-merge-json/v1"
	MaxRequestBytes    = 4 << 20
	MaxResultBytes     = 2 << 20
	Merged             = "merged"
	Conflicted         = "conflicted"
	InvalidCombination = "invalid-combination"
)

var (
	ErrInvalid = errors.New("invalid model merge input")
	ErrLimit   = errors.New("model merge resource limit")
)

// Conflict describes one atomic merge unit, not a proposed resolution.
// ID is empty for root fields; Field is empty for named records.
type Conflict struct {
	Kind   string `json:"kind"`
	ID     string `json:"id"`
	Field  string `json:"field"`
	Reason string `json:"reason"`
}

// SourceDigests identify observed bytes, not authenticated provenance.
type SourceDigests struct {
	Request string `json:"request"`
	Base    string `json:"base"`
	Left    string `json:"left"`
	Right   string `json:"right"`
}

// Result owns its candidate and conflicts. Only Merged has a candidate and
// candidate digest. Source digests are nil for typed Merge calls.
type Result struct {
	MergeVersion    string
	ProfileVersion  string
	Mode            string
	Authority       string
	BaseDigest      string
	LeftDigest      string
	RightDigest     string
	Sources         *SourceDigests
	Status          string
	CandidateDigest *string
	Candidate       *model.Model
	Conflicts       []Conflict
}
