package scenarioexplain

const BoundFormatVersion = "lifecycle-bound-scenario-explanation/v1"

// BoundExplanation describes supplied assertions and declared model links for
// an explicit Go runner binding. It proves neither wire origin nor execution,
// compiled-code provenance, review, authority or behavioral coverage.
type BoundExplanation struct {
	FormatVersion  string  `json:"formatVersion"`
	IdentitySource string  `json:"identitySource"`
	Mode           string  `json:"mode"`
	Authority      string  `json:"authority"`
	Status         string  `json:"status"`
	ModelDigest    string  `json:"modelDigest"`
	ScenarioDigest string  `json:"scenarioDigest"`
	SourceDigest   *string `json:"sourceDigest"`
	Matched        bool    `json:"matched"`
	CasesCompleted int     `json:"casesCompleted"`
	StepsExecuted  int     `json:"stepsExecuted"`
	Entries        []Entry `json:"entries"`
}
