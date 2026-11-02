// Package profilecatalog describes the reviewed lifecycle language profile.
package profilecatalog

import "github.com/spik3r/clanker-foundry/modeling/lifecycle/model"

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
