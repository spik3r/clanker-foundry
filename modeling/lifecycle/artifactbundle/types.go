// Package artifactbundle builds bounded canonical archives. Verification checks
// consistency only: it authenticates neither payloads, authors nor executables.
package artifactbundle

import "errors"

const (
	FormatVersion        = "lifecycle-artifact-bundle/v1"
	ReviewProfile        = "lifecycle-reviewpack/v1"
	QualificationProfile = "lifecycle-qualification/v1"
	DesignChangeProfile  = "lifecycle-designchange/v1"
	ManifestName         = "manifest.json"
	MaxFiles             = 64
	MaxFileBytes         = 8 << 20
	MaxTotalBytes        = 32 << 20
	MaxArchiveBytes      = 16 << 20
	MaxManifestBytes     = 64 << 10
	MaxNameBytes         = 192
)

var (
	ErrInvalid = errors.New("invalid artifact bundle")
	ErrLimit   = errors.New("artifact bundle limit exceeded")
	ErrProfile = errors.New("unsupported artifact bundle profile")
)

// File is one payload. Names are canonical relative ASCII slash paths. Data may
// be empty. Build treats inputs as immutable; Verify returns owned sorted copies.
type File struct {
	Name string
	Data []byte
}
