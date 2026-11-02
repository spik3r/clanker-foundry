// Package modelpatch defines an explicit, digest-preconditioned proposal over
// accepted model records. It proposes candidates and declared-link review data;
// it executes no model behavior and grants no operating or review authority.
package modelpatch

import (
	"errors"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/impact"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/model"
)

const (
	FormatVersion      = "lifecycle-model-patch-json/v1"
	PatchVersion       = "lifecycle-model-patch/v1"
	ReportVersion      = "lifecycle-model-patch-report/v1"
	ProfileVersion     = model.ProfileVersion
	MaxOperations      = 64
	MaxPatchBytes      = 1 << 20
	MaxModelBytes      = 1 << 20
	MaxRequestBytes    = 3 << 20
	MaxResultBytes     = 3 << 20
	MaxJSONDepth       = 64
	MaxExprDepth       = model.MaxDepth
	MaxExprNodes       = model.MaxNodes
	MaxExprArgs        = 16
	MaxIdentifierBytes = 48
	MaxStateLabelBytes = 256
	MaxRecordTextBytes = 4096
	MaxExprRefBytes    = 128
	MaxExprAtomBytes   = 16
	MaxEffectOpBytes   = 64
	MaxGuardIDs        = 64
	MaxEffects         = 3
	Mode               = "example-only"
	Authority          = "none"
)

var (
	ErrInvalid      = errors.New("invalid model patch input or candidate")
	ErrLimit        = errors.New("model patch budget exceeded")
	ErrPrecondition = errors.New("model patch before digest mismatch")
	ErrConflict     = errors.New("model patch target conflict")
)

type Action string
type Kind string

const (
	Add         Action = "add"
	Replace     Action = "replace"
	Remove      Action = "remove"
	Model       Kind   = "model"
	State       Kind   = "state"
	Requirement Kind   = "requirement"
	Guard       Kind   = "guard"
	Transition  Kind   = "transition"
)

// Root replaces ID and Initial together. Version/profile are not patchable.
type Root struct{ ID, Initial string }

// Operation is a closed union. Non-root ID matches its complete selected record.
// Root uses empty ID and Replace only. Remove has no payload. Repeated (Kind,ID)
// targets are conflicts even when actions would cancel; no implicit rename exists.
type Operation struct {
	Action      Action
	Kind        Kind
	ID          string
	Root        *Root
	State       *model.State
	Requirement *model.Requirement
	Guard       *model.Guard
	Transition  *model.Transition
}

// Patch is format-neutral intent. Version must be1, Operations must be nonnil,
// and order is preserved. BeforeDigest binds the complete accepted before model.
// Nil leaf Expr.Args denotes an empty sequence, encoded as[] in the patch DTO.
type Patch struct {
	Version        int
	ProfileVersion string
	BeforeDigest   string
	Operations     []Operation
}

// Result is a computed proposal, not permission or proof of provenance/review.
// On error Plan returns Result{}; Execute returns no bytes. After and
// Review collections must be owned copies. No arbitrary-result encoder is exposed.
// This Go struct is not the wire DTO: After is emitted via jsonformat.Encode.
type Result struct {
	ReportVersion  string
	ProfileVersion string
	Mode           string
	Authority      string
	BeforeDigest   string
	AfterDigest    string
	After          model.Model
	Review         impact.Report
}
