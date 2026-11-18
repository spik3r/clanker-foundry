package diagram

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/spik3r/clanker-foundry/modeling/diagram/dashboardbind"
	"github.com/spik3r/clanker-foundry/modeling/diagram/tesserae"
)

var unicodeProfiles = []struct{ file, nested string }{
	{"order", "Check stock"}, {"minimal-dashboard", "All systems normal"}, {"checkout-sequence", "Place order"}, {"orders-class", "OrderLine"},
}

func unicodeRead(t *testing.T, name string) []byte {
	t.Helper()
	raw, err := os.ReadFile("examples/" + name + ".json")
	if err != nil {
		t.Fatal(err)
	}
	return raw
}
func unicodeTitle(t *testing.T, raw []byte, text string) []byte {
	t.Helper()
	var m map[string]json.RawMessage
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatal(err)
	}
	m["title"] = json.RawMessage(`"` + text + `"`)
	b, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	return b
}
func unicodeReplace(t *testing.T, raw []byte, old, text string) []byte {
	t.Helper()
	quoted, _ := json.Marshal(old)
	if !bytes.Contains(raw, quoted) {
		t.Fatal("replacement not found")
	}
	return bytes.Replace(raw, quoted, []byte(`"`+text+`"`), 1)
}
func unicodeRefusal(t *testing.T, err error) {
	t.Helper()
	if err == nil || !strings.Contains(err.Error(), "unpaired surrogate escape in JSON") {
		t.Fatalf("Unicode refusal = %v", err)
	}
}

func TestUnicodeProfileRefusals(t *testing.T) {
	for _, p := range unicodeProfiles {
		t.Run(p.file, func(t *testing.T) {
			raw := unicodeRead(t, p.file)
			profile, err := Select(raw)
			if err != nil {
				t.Fatal(err)
			}
			for _, text := range []string{`\ud800`, `\udfff`, `\udc00\ud800`, `\ud800\u0041`, `\ud800\\udc00`} {
				for _, input := range [][]byte{unicodeTitle(t, raw, text), unicodeReplace(t, raw, p.nested, text), bytes.Replace(raw, []byte(`"title"`), []byte(`"`+text+`"`), 1)} {
					_, err := profile.Generate(input, Self)
					unicodeRefusal(t, err)
					res, err := Build(input)
					unicodeRefusal(t, err)
					if res.Identity != "" || res.Files != nil {
						t.Fatal("partial result on refusal")
					}
				}
			}
		})
	}
}

func TestUnicodeProfileValid(t *testing.T) {
	for _, p := range unicodeProfiles {
		t.Run(p.file, func(t *testing.T) {
			for _, text := range []string{`\ud83d\ude00`, `\uD800\uDC00`, `\udbff\udfff`, `\ufffd`, "�", `\\ud800`, `\\\"\ud83d\ude00\/`} {
				input := unicodeTitle(t, unicodeRead(t, p.file), text)
				var literal map[string]any
				if err := json.Unmarshal(input, &literal); err != nil {
					t.Fatal(err)
				}
				normalized, err := json.Marshal(literal)
				if err != nil {
					t.Fatal(err)
				}
				a, err := Build(input)
				if err != nil {
					t.Fatal(err)
				}
				b, err := Build(normalized)
				if err != nil {
					t.Fatal(err)
				}
				if !reflect.DeepEqual(a.Files, b.Files) {
					t.Fatal("valid escape spelling changed canonical digest or generated files")
				}
				var identity struct{ SourceSHA256 string }
				if err := json.Unmarshal([]byte(a.Identity), &identity); err != nil {
					t.Fatal(err)
				}
				sum := sha256.Sum256(input)
				if identity.SourceSHA256 != hex.EncodeToString(sum[:]) {
					t.Fatal("source digest does not cover original bytes")
				}
			}
		})
	}
}

func unicodeBinding(t *testing.T) ([]byte, []byte) {
	t.Helper()
	template := unicodeRead(t, "ops-dashboard")
	data := unicodeRead(t, "ops-report-data")
	sum := sha256.Sum256(template)
	var m map[string]json.RawMessage
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatal(err)
	}
	m["templateSHA256"], _ = json.Marshal(hex.EncodeToString(sum[:]))
	data, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	return template, data
}

func TestUnicodeBindingAndTesseraeRefusals(t *testing.T) {
	template, data := unicodeBinding(t)
	for _, text := range []string{`\ud800`, `\udfff`, `\udc00\ud800`, `\ud800\u0041`} {
		for _, input := range [][]byte{unicodeReplace(t, data, "Mon", text), unicodeReplace(t, data, "Two days, fictional example", text), bytes.Replace(data, []byte(`"note"`), []byte(`"`+text+`"`), 1)} {
			out, err := dashboardbind.Bind(template, input)
			unicodeRefusal(t, err)
			if out != nil {
				t.Fatal("partial binding result")
			}
		}
		out, err := dashboardbind.Bind(unicodeTitle(t, template, text), data)
		unicodeRefusal(t, err)
		if out != nil {
			t.Fatal("partial binding result")
		}
		raw := []byte(`{"formatVersion":"diagram-json/v1","profileVersion":"dashboard/v1","id":"Demo","title":"Valid","widgets":[{"id":"Count","kind":"stat","label":"Count","value":"42"}],"layout":[["Count"]]}`)
		for _, input := range [][]byte{unicodeTitle(t, raw, text), unicodeReplace(t, raw, "Valid", text), bytes.Replace(raw, []byte(`"label":"Count"`), []byte(`"label":"`+text+`"`), 1)} {
			out, err := tesserae.Export(input)
			unicodeRefusal(t, err)
			if out != nil {
				t.Fatal("partial Tesserae result")
			}
		}
	}
}

func TestUnicodeBindingAndTesseraeValid(t *testing.T) {
	template, data := unicodeBinding(t)
	for _, text := range []string{`\ud83d\ude00`, `\ufffd`, "�", `\\ud800`} {
		input := unicodeReplace(t, data, "Mon", text)
		out, err := dashboardbind.Bind(template, input)
		if err != nil {
			t.Fatal(err)
		}
		var literal any
		if err := json.Unmarshal(input, &literal); err != nil {
			t.Fatal(err)
		}
		normalized, _ := json.Marshal(literal)
		want, err := dashboardbind.Bind(template, normalized)
		if err != nil || !bytes.Equal(out, want) {
			t.Fatalf("binding changed text: %v", err)
		}
		raw := []byte(`{"formatVersion":"diagram-json/v1","profileVersion":"dashboard/v1","id":"Demo","title":"Valid","widgets":[{"id":"Count","kind":"stat","label":"Count","value":"42"}],"layout":[["Count"]]}`)
		input = unicodeTitle(t, raw, text)
		out, err = tesserae.Export(input)
		if err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal(input, &literal); err != nil {
			t.Fatal(err)
		}
		normalized, _ = json.Marshal(literal)
		want, err = tesserae.Export(normalized)
		if err != nil || !bytes.Equal(out, want) {
			t.Fatalf("Tesserae changed text: %v", err)
		}
	}
}

func TestUnicodeSelectKeepsHeaderRefusal(t *testing.T) {
	for _, raw := range []string{`{"profileVersion":"\ud800"}`, `{"profileVersion":"unknown"}`} {
		_, err := Build([]byte(raw))
		if err == nil || !strings.Contains(err.Error(), "unsupported profileVersion") {
			t.Fatalf("header refusal changed: %v", err)
		}
	}
}
