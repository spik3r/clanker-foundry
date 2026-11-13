package sequence

import (
	"github.com/spik3r/clanker-foundry/modeling/diagram/contract"
	"github.com/spik3r/clanker-foundry/modeling/diagram/internal/svgx"
)

// Output file names.
const (
	FileSVG   = "diagram.svg"
	FileModel = "model.json"
)

// Profiler implements contract.Profile for sequence-diagram/v1.
type Profiler struct{}

// ID returns the profile identifier.
func (Profiler) ID() string { return Profile }

// Generate decodes and validates raw, then renders every output file. It
// performs no I/O. Emitted bytes depend only on the canonical model and g.
func (Profiler) Generate(raw []byte, g contract.Generator) (contract.Generated, error) {
	m, err := Decode(raw)
	if err != nil {
		return contract.Generated{}, contract.ModelError{Err: err}
	}
	canonical, err := m.Canonical()
	if err != nil {
		return contract.Generated{}, err
	}
	pretty, err := svgx.Encode(m, true)
	if err != nil {
		return contract.Generated{}, err
	}
	return contract.Generated{
		Canonical: canonical,
		Files: map[string][]byte{
			FileSVG:   SVG(m, LayoutOf(m), g),
			FileModel: pretty,
		},
	}, nil
}
