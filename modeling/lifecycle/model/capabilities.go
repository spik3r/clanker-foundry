package model

import "sort"

// ReferenceCapability describes a declared reference's type and optionality.
type ReferenceCapability struct {
	Path     string `json:"path"`
	Type     string `json:"type"`
	Optional bool   `json:"optional"`
}

// ReferenceCapabilities returns sorted owned scalar copies of the live registry.
func ReferenceCapabilities() []ReferenceCapability {
	out := make([]ReferenceCapability, 0, len(references))
	for path, ref := range references {
		out = append(out, ReferenceCapability{Path: path, Type: ref.typ, Optional: ref.optional})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Path < out[j].Path })
	return out
}
