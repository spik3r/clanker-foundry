// Package diagram selects a diagram profile by the model's profileVersion and
// builds its output under Foundry generator contract v1. Each profile lives in
// its own package and implements contract.Profile; add a new one to profiles.
package diagram

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/spik3r/clanker-foundry/modeling/diagram/contract"
	"github.com/spik3r/clanker-foundry/modeling/diagram/process"
)

// Generator identity. Toolchain is the Go release this generator was
// qualified with; it is a constant, not detected at run time. Any change that
// alters emitted bytes for any model needs a new Version.
const (
	GeneratorID      = "diagramgen"
	GeneratorVersion = "1.0.0"
	Toolchain        = "go1.27.4"
)

// Self is the generator identity passed to every profile.
var Self = contract.Generator{ID: GeneratorID, Version: GeneratorVersion, Toolchain: Toolchain}

// profiles lists the supported profiles by profileVersion.
var profiles = map[string]contract.Profile{
	process.Profile: process.Profiler{},
}

// Supported returns the supported profileVersion values, sorted.
func Supported() []string {
	var names []string
	for name := range profiles {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// Select returns the profile a model asks for. Strict decoding is the
// profile's job; this only reads profileVersion.
func Select(raw []byte) (contract.Profile, error) {
	if len(raw) > contract.MaxModelBytes {
		return nil, contract.Refuse("model is larger than %d bytes", contract.MaxModelBytes)
	}
	var head struct {
		ProfileVersion string `json:"profileVersion"`
	}
	if err := json.Unmarshal(raw, &head); err != nil {
		return nil, contract.Refuse("not a JSON object: %v", err)
	}
	p, ok := profiles[head.ProfileVersion]
	if !ok {
		return nil, contract.Refuse("unsupported profileVersion %q; supported: %s",
			head.ProfileVersion, strings.Join(Supported(), ", "))
	}
	return p, nil
}

// Build selects the profile and computes the complete output for raw. It
// performs no I/O.
func Build(raw []byte) (contract.Result, error) {
	p, err := Select(raw)
	if err != nil {
		return contract.Result{}, err
	}
	res, err := contract.Assemble(p, Self, raw)
	if err != nil {
		return contract.Result{}, fmt.Errorf("%s: %w", p.ID(), err)
	}
	return res, nil
}
