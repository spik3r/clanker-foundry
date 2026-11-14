package tui

import (
	"os"
	"regexp"
	"slices"
	"strings"
	"testing"
)

var hintRow = regexp.MustCompile("^\\| `([a-z]+(?:\\.[a-z]+)?)` \\| (.*) \\| `(.*)` \\| `(.*)` \\|$")

// The documented hint vocabulary and the generator's list cannot drift apart:
// README.md and PROFILE.md each hold a table with exactly the generator's IDs,
// in order, with the same when/key/action text.
func TestKeyHintVocabularyIsDocumented(t *testing.T) {
	seen := map[string]bool{}
	for _, h := range KeyHintVocabulary {
		if seen[h.ID] || h.ID == "" || h.When == "" || h.Key == "" || h.Action == "" || strings.ToLower(h.ID) != h.ID {
			t.Fatalf("bad vocabulary entry %+v", h)
		}
		seen[h.ID] = true
	}
	for _, doc := range []string{"README.md", "PROFILE.md"} {
		raw, err := os.ReadFile(doc)
		if err != nil {
			t.Fatal(err)
		}
		var rows []HintSpec
		for _, line := range strings.Split(string(raw), "\n") {
			if m := hintRow.FindStringSubmatch(line); m != nil {
				rows = append(rows, HintSpec{ID: m[1], When: m[2], Key: m[3], Action: m[4]})
			}
		}
		if !slices.Equal(rows, KeyHintVocabulary) {
			t.Errorf("%s hint table\n got %+v\nwant %+v", doc, rows, KeyHintVocabulary)
		}
		for _, want := range []string{"never renamed or reused", "`KeyHint.ID`"} {
			if doc == "README.md" || want != "`KeyHint.ID`" {
				if !strings.Contains(strings.Join(strings.Fields(string(raw)), " "), want) {
					t.Errorf("%s lacks %q", doc, want)
				}
			}
		}
	}
}
