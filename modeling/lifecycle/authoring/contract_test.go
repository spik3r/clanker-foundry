package authoring

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/jsonformat"
)

// Expectations are handwritten in cases.json, shared with the independent
// JSON Schema runner. Schema acceptance alone does not imply model acceptance.
func TestAuthoringCasesAgainstModel(t *testing.T) {
	b, err := os.ReadFile("cases.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []struct {
		Name        string          `json:"name"`
		SchemaValid bool            `json:"schemaValid"`
		ModelValid  bool            `json:"modelValid"`
		Document    json.RawMessage `json:"document"`
	}
	if err := json.Unmarshal(b, &cases); err != nil {
		t.Fatal(err)
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			_, err := jsonformat.Decode(c.Document)
			if (err == nil) != c.ModelValid {
				t.Fatalf("model valid = %v; want %v (error: %v)", err == nil, c.ModelValid, err)
			}
			if c.ModelValid && !c.SchemaValid {
				t.Fatal("valid model must pass authoring schema")
			}
		})
	}
}
