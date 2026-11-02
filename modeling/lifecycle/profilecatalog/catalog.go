// Package profilecatalog describes the reviewed lifecycle language profile.
package profilecatalog

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/compiler"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/jsonformat"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/model"
)

const FormatVersion = "lifecycle-profile-catalog/v1"
const ProjectionVersion = "lifecycle-v1-source-projection/v1"
const MaxOutputBytes = 65536

type Descriptor struct {
	FormatVersion         string                      `json:"formatVersion"`
	ProjectionVersion     string                      `json:"projectionVersion"`
	Status                string                      `json:"status"`
	Authority             string                      `json:"authority"`
	Identities            Identities                  `json:"identities"`
	SourcePins            []SourcePin                 `json:"sourcePins"`
	References            []model.ReferenceCapability `json:"references"`
	ExpressionForms       []ExpressionForm            `json:"expressionForms"`
	Operators             []Operator                  `json:"operators"`
	Effects               Effects                     `json:"effects"`
	Limits                Limits                      `json:"limits"`
	StructuralRules       []string                    `json:"structuralRules"`
	Limitations           []string                    `json:"limitations"`
	SourceIdentityMeaning string                      `json:"sourceIdentityMeaning"`
}
type Identities struct {
	Module              string `json:"module"`
	TypedIRVersion      int    `json:"typedIRVersion"`
	ProfileVersion      string `json:"profileVersion"`
	SourceFormatVersion string `json:"sourceFormatVersion"`
	GeneratorVersion    string `json:"generatorVersion"`
	ToolchainVersion    string `json:"toolchainVersion"`
}
type SourcePin struct {
	Path         string `json:"path"`
	Size         int    `json:"size"`
	SHA256       string `json:"sha256"`
	IdentityKind string `json:"identityKind"`
}
type ExpressionForm struct {
	Kind    string `json:"kind"`
	Meaning string `json:"meaning"`
}
type Policy string

const (
	DirectOptionalReference Policy = "direct-optional-reference"
	DistinctBooleanSet      Policy = "distinct-boolean-set"
	SameTypeScalarPair      Policy = "same-type-scalar-pair"
	NonemptyValue           Policy = "nonempty-value"
)

type Operator struct {
	Name                          string   `json:"name"`
	MinArity                      int      `json:"minArity"`
	MaxArity                      int      `json:"maxArity"`
	Arguments                     string   `json:"arguments"`
	Result                        string   `json:"result"`
	Policy                        Policy   `json:"policy"`
	AllowedOperandTypes           []string `json:"allowedOperandTypes"`
	DirectReferenceRequired       bool     `json:"directReferenceRequired"`
	OptionalReferenceRequired     bool     `json:"optionalReferenceRequired"`
	SameOperandTypeRequired       bool     `json:"sameOperandTypeRequired"`
	DistinctExprKeyRequired       bool     `json:"distinctExprKeyRequired"`
	NormalizationDuplicateRecheck bool     `json:"normalizationDuplicateRecheck"`
}
type Effects struct {
	AllowedSequences [][]string `json:"allowedSequences"`
	Semantics        string     `json:"semantics"`
}
type Limits struct {
	IdentifierBytesMin         int    `json:"identifierBytesMin"`
	IdentifierBytesMax         int    `json:"identifierBytesMax"`
	IdentifierGrammar          string `json:"identifierGrammar"`
	StatesMin                  int    `json:"statesMin"`
	StatesMax                  int    `json:"statesMax"`
	RequirementsMin            int    `json:"requirementsMin"`
	RequirementsMax            int    `json:"requirementsMax"`
	GuardsMin                  int    `json:"guardsMin"`
	GuardsMax                  int    `json:"guardsMax"`
	TransitionsMin             int    `json:"transitionsMin"`
	TransitionsMax             int    `json:"transitionsMax"`
	TransitionGuardsMin        int    `json:"transitionGuardsMin"`
	TransitionGuardsMax        int    `json:"transitionGuardsMax"`
	StateLabelBytesMax         int    `json:"stateLabelBytesMax"`
	RequirementTextBytesMax    int    `json:"requirementTextBytesMax"`
	StringLiteralBytesMax      int    `json:"stringLiteralBytesMax"`
	ExpressionMaxDepthRootZero int    `json:"expressionMaxDepthRootZero"`
	TypeOfMaxNodes             int    `json:"typeOfMaxNodes"`
	WholeModelMaxGuardNodes    int    `json:"wholeModelMaxGuardNodes"`
	BooleanArgumentsMin        int    `json:"booleanArgumentsMin"`
	BooleanArgumentsMax        int    `json:"booleanArgumentsMax"`
	JSONSourceBytesMax         int    `json:"jsonSourceBytesMax"`
	JSONSourceMaxDepth         int    `json:"jsonSourceMaxDepth"`
	CatalogOutputBytesMax      int    `json:"catalogOutputBytesMax"`
	StrictJSONMaxScannedValues int    `json:"strictJSONMaxScannedValues"`
}

// Describe returns owned descriptive data; it does not inspect source files.
func Describe() (Descriptor, error) {
	if !compatibleVersions(model.ProfileVersion, jsonformat.FormatVersion, compiler.GeneratorVersion, compiler.ToolchainVersion) {
		return Descriptor{}, errors.New("profile catalog unavailable")
	}
	d := Descriptor{
		FormatVersion:     FormatVersion,
		ProjectionVersion: ProjectionVersion,
		Status:            "descriptive-projection",
		Authority:         "none",
		Identities: Identities{
			Module:              "github.com/spik3r/clanker-foundry/modeling/lifecycle",
			TypedIRVersion:      1,
			ProfileVersion:      model.ProfileVersion,
			SourceFormatVersion: jsonformat.FormatVersion,
			GeneratorVersion:    compiler.GeneratorVersion,
			ToolchainVersion:    compiler.ToolchainVersion,
		},
		SourcePins: []SourcePin{
			SourcePin{
				Path:         "model/ir.go",
				Size:         9385,
				SHA256:       "13f4ad2980b1ca523b07bdbc8ebeb0e0300f8df81faa8dcd1cdec5ccef007aeb",
				IdentityKind: "expected-reviewed-source",
			},
			SourcePin{
				Path:         "jsonformat/json.go",
				Size:         5162,
				SHA256:       "fe4eab95f978c6bf714171b553d2c7647322c20442d886a19d147b892899c9f3",
				IdentityKind: "expected-reviewed-source",
			},
			SourcePin{
				Path:         "strictjson/decode.go",
				Size:         4236,
				SHA256:       "a72ded4af99ffba5f5cca2931c6cace072634ee2e42eb00cd9a5988c691f1ee2",
				IdentityKind: "expected-reviewed-source",
			},
			SourcePin{
				Path:         "compiler/generate.go",
				Size:         12315,
				SHA256:       "e17610a27748a5d3baedf4f59c2183cb184f7e2681b5532c6f9d8d3d1eda6a38",
				IdentityKind: "expected-reviewed-source",
			},
		},
		References: model.ReferenceCapabilities(),
		ExpressionForms: []ExpressionForm{
			ExpressionForm{
				Kind:    "ref",
				Meaning: "one declared typed reference; other scalar fields and args empty",
			},
			ExpressionForm{
				Kind:    "string",
				Meaning: "UTF-8 string literal up to 4096 bytes; blank allowed; other scalar fields and args empty",
			},
			ExpressionForm{
				Kind:    "op",
				Meaning: "one supported operator with typed arguments; ref and value fields empty",
			},
		},
		Operators: []Operator{
			Operator{
				Name:      "absent",
				MinArity:  1,
				MaxArity:  1,
				Arguments: "one direct optional reference",
				Result:    "bool",
				Policy:    "direct-optional-reference",
				AllowedOperandTypes: []string{
					"string",
					"uint64",
				},
				DirectReferenceRequired:       true,
				OptionalReferenceRequired:     true,
				SameOperandTypeRequired:       false,
				DistinctExprKeyRequired:       false,
				NormalizationDuplicateRecheck: false,
			},
			Operator{
				Name:      "all",
				MinArity:  1,
				MaxArity:  16,
				Arguments: "boolean expressions with distinct ExprKey; duplicates rechecked after normalization",
				Result:    "bool",
				Policy:    "distinct-boolean-set",
				AllowedOperandTypes: []string{
					"bool",
				},
				DirectReferenceRequired:       false,
				OptionalReferenceRequired:     false,
				SameOperandTypeRequired:       false,
				DistinctExprKeyRequired:       true,
				NormalizationDuplicateRecheck: true,
			},
			Operator{
				Name:      "any",
				MinArity:  1,
				MaxArity:  16,
				Arguments: "boolean expressions with distinct ExprKey; duplicates rechecked after normalization",
				Result:    "bool",
				Policy:    "distinct-boolean-set",
				AllowedOperandTypes: []string{
					"bool",
				},
				DirectReferenceRequired:       false,
				OptionalReferenceRequired:     false,
				SameOperandTypeRequired:       false,
				DistinctExprKeyRequired:       true,
				NormalizationDuplicateRecheck: true,
			},
			Operator{
				Name:      "eq",
				MinArity:  2,
				MaxArity:  2,
				Arguments: "same-type string/string or uint64/uint64",
				Result:    "bool",
				Policy:    "same-type-scalar-pair",
				AllowedOperandTypes: []string{
					"string",
					"uint64",
				},
				DirectReferenceRequired:       false,
				OptionalReferenceRequired:     false,
				SameOperandTypeRequired:       true,
				DistinctExprKeyRequired:       false,
				NormalizationDuplicateRecheck: false,
			},
			Operator{
				Name:      "ne",
				MinArity:  2,
				MaxArity:  2,
				Arguments: "same-type string/string or uint64/uint64",
				Result:    "bool",
				Policy:    "same-type-scalar-pair",
				AllowedOperandTypes: []string{
					"string",
					"uint64",
				},
				DirectReferenceRequired:       false,
				OptionalReferenceRequired:     false,
				SameOperandTypeRequired:       true,
				DistinctExprKeyRequired:       false,
				NormalizationDuplicateRecheck: false,
			},
			Operator{
				Name:      "nonempty",
				MinArity:  1,
				MaxArity:  1,
				Arguments: "one string or strings expression",
				Result:    "bool",
				Policy:    "nonempty-value",
				AllowedOperandTypes: []string{
					"string",
					"strings",
				},
				DirectReferenceRequired:       false,
				OptionalReferenceRequired:     false,
				SameOperandTypeRequired:       false,
				DistinctExprKeyRequired:       false,
				NormalizationDuplicateRecheck: false,
			},
		},
		Effects: Effects{
			AllowedSequences: [][]string{
				{model.SetState, model.IncrementRevision},
				{model.AssignOwner, model.SetState, model.IncrementRevision},
			},
			Semantics: "ordered; set_state uses the transition target; no arbitrary effect op",
		},
		Limits: Limits{
			IdentifierBytesMin:         1,
			IdentifierBytesMax:         48,
			IdentifierGrammar:          "ASCII letter followed by ASCII letters, digits or underscores",
			StatesMin:                  2,
			StatesMax:                  16,
			RequirementsMin:            1,
			RequirementsMax:            64,
			GuardsMin:                  1,
			GuardsMax:                  64,
			TransitionsMin:             1,
			TransitionsMax:             32,
			TransitionGuardsMin:        1,
			TransitionGuardsMax:        64,
			StateLabelBytesMax:         256,
			RequirementTextBytesMax:    4096,
			StringLiteralBytesMax:      4096,
			ExpressionMaxDepthRootZero: model.MaxDepth,
			TypeOfMaxNodes:             model.MaxNodes,
			WholeModelMaxGuardNodes:    model.MaxNodes,
			BooleanArgumentsMin:        1,
			BooleanArgumentsMax:        16,
			JSONSourceBytesMax:         1048576,
			JSONSourceMaxDepth:         64,
			CatalogOutputBytesMax:      MaxOutputBytes,
			StrictJSONMaxScannedValues: 100000,
		},
		StructuralRules: []string{
			"Model version is 1; IDs are globally unique across model, states, requirements, guards and transitions.",
			"A guard references a requirement and has a boolean root expression; all guards and requirements are used.",
			"Transition endpoints and guards exist; transition guard IDs are unique; state/trigger pairs are unique.",
			"At least one terminal state exists; terminals have no outgoing transitions, nonterminals have an outgoing transition, and all states are reachable from the initial state.",
			"State labels and requirement text are nonblank valid UTF-8; string literals may be blank.",
			"Normalization sorts ID collections, transition guards, and all/any operands; effect order and eq/ne operand order remain significant.",
		},
		Limitations: []string{
			"Operator and limit entries are a reviewed source-pinned descriptive projection, not a reflective authoritative registry or evaluator.",
			"Source hashes are consistency assertions, not authenticated binary provenance or proof of human review.",
			"TypeOf and this catalog do not establish model validity, guard satisfiability, execution, review, authorization or production policy.",
			"The available source adapter is JSON-v1 only; no JSON Schema, second format, plugin system or UML/BPMN/SysML conformance is claimed.",
			"Example lifecycle guards and decisions are model content, not universal profile policy.",
		},
		SourceIdentityMeaning: "Expected reviewed source identities; the pure API and CLI do not read or observe source bytes at runtime. Test-only drift checks do not authenticate binaries, hosts, or review.",
	}
	return d, nil
}

// Encode builds its own descriptor and emits bounded deterministic JSON plus LF.
func Encode() ([]byte, error) {
	d, e := Describe()
	if e != nil {
		return nil, e
	}
	return encodeDescriptor(d)
}

// encodeDescriptor is private; callers cannot encode an arbitrary descriptor.
func encodeDescriptor(d Descriptor) ([]byte, error) {
	var b bytes.Buffer
	enc := json.NewEncoder(&b)
	enc.SetEscapeHTML(false)
	if enc.Encode(d) != nil || b.Len() > MaxOutputBytes {
		return nil, errors.New("profile catalog unavailable")
	}
	return b.Bytes(), nil
}

// compatibleVersions prevents a reviewed projection from being labeled as a new profile.
func compatibleVersions(profile, format, generator, toolchain string) bool {
	return profile == "lifecycle/v1" && format == "lifecycle-json/v1" && generator == "lifecycle-go/v1" && toolchain == "go1.27.3"
}
