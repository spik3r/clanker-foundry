package verify

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	validator "github.com/spik3r/clanker-foundry/modeling/jsonvalidator/testdata/sample"
)

const frozenCorpusSHA256 = "60fb575f37adb783564fbaae7535bf7d59587127dd638c8a0a87ce083d56ef41"

func TestFrozenSyntheticAcceptance(t *testing.T) {
	data, err := os.ReadFile("../testdata/sample/cases.json")
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprintf("%x", sha256.Sum256(data)) != frozenCorpusSHA256 {
		t.Fatal("frozen expectations changed")
	}
	var cases []struct {
		ID      string `json:"id"`
		RawJSON string `json:"rawJson"`
		Valid   bool   `json:"valid"`
	}
	if err := json.Unmarshal(data, &cases); err != nil {
		t.Fatal(err)
	}
	if len(cases) != 46 {
		t.Fatal("unexpected case count")
	}
	seen := map[string]bool{}
	validCount := 0
	for _, c := range cases {
		if seen[c.ID] || c.ID == "" {
			t.Fatal("duplicate or empty case ID")
		}
		seen[c.ID] = true
		if c.Valid {
			validCount++
		}
		t.Run(c.ID, func(t *testing.T) {
			d := json.NewDecoder(strings.NewReader(c.RawJSON))
			d.UseNumber()
			var value any
			if err := d.Decode(&value); err != nil {
				t.Fatal(err)
			}
			var trailing any
			if err := d.Decode(&trailing); err != io.EOF {
				t.Fatal("fixture has trailing content")
			}
			if got := validator.Validate(value); got != c.Valid {
				t.Fatalf("got %v, want %v", got, c.Valid)
			}
		})
	}
	if validCount != 12 {
		t.Fatal("unexpected true case count")
	}
	model, err := os.ReadFile("../testdata/sample/model.json")
	if err != nil {
		t.Fatal(err)
	}
	if validator.ModelSHA256 != fmt.Sprintf("%x", sha256.Sum256(model)) {
		t.Fatal("generated Go has wrong model pin")
	}
}
