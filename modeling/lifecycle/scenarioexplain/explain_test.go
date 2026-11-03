package scenarioexplain

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/jsonformat"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/model"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/scenarios"
)

type vectors struct {
	Cases []struct {
		ID          string          `json:"id"`
		Report      json.RawMessage `json:"report"`
		ModelChange *struct {
			Operation string `json:"operation"`
			From      string `json:"from"`
			To        string `json:"to"`
			ID        string `json:"id"`
			Value     string `json:"value"`
		} `json:"modelChange"`
		Expected struct {
			Result      string          `json:"result"`
			Explanation json.RawMessage `json:"explanation"`
		} `json:"expected"`
	} `json:"cases"`
	Raw []struct {
		ID         string `json:"id"`
		Entrypoint string `json:"entrypoint"`
		Bytes      string `json:"bytesBase64"`
		Expected   string `json:"expected"`
	} `json:"rawNegativeVectors"`
}

func fixtures(t *testing.T) (vectors, []byte, model.Model) {
	t.Helper()
	raw, e := os.ReadFile("testdata/vectors.json")
	if e != nil {
		t.Fatal(e)
	}
	sum := sha256.Sum256(raw)
	if hex.EncodeToString(sum[:]) != "206113fe0d07f04ed3989bc63134b01a64defa5994f9d2a195a0584fa6bf255e" {
		t.Fatal("frozen vectors changed")
	}
	var v vectors
	if json.Unmarshal(raw, &v) != nil {
		t.Fatal("invalid vectors")
	}
	b, e := os.ReadFile("../examples/lifecycle.json")
	if e != nil {
		t.Fatal(e)
	}
	m, e := jsonformat.Decode(b)
	if e != nil {
		t.Fatal(e)
	}
	return v, b, m
}
func requestBytes(m, r []byte) []byte {
	return []byte(`{"formatVersion":"lifecycle-scenario-explain-request/v1","model":` + string(m) + `,"report":` + string(r) + `}`)
}
func TestFrozenVectors(t *testing.T) {
	v, modelRaw, _ := fixtures(t)
	if len(v.Cases) != 20 || len(v.Raw) != 18 {
		t.Fatal("incomplete vector inventory")
	}
	seen := map[string]bool{}
	for _, c := range v.Cases {
		t.Run(c.ID, func(t *testing.T) {
			if seen[c.ID] {
				t.Fatal("duplicate vector")
			}
			seen[c.ID] = true
			m, e := jsonformat.Decode(modelRaw)
			if e != nil {
				t.Fatal(e)
			}
			r, e := scenarios.DecodeReport(c.Report)
			if e != nil {
				if c.Expected.Result == "reject" && c.ID == "invalid_report_envelope" {
					if out, err := Execute(requestBytes(modelRaw, c.Report)); err == nil || out != nil {
						t.Fatal("Execute accepted invalid report envelope")
					}
					return
				}
				t.Fatalf("report decode: %v", e)
			}
			if c.ModelChange != nil {
				switch c.ModelChange.Operation {
				case "replace_requirement_text":
					for i := range m.Requirements {
						if m.Requirements[i].ID == c.ModelChange.ID {
							m.Requirements[i].Text = c.ModelChange.Value
						}
					}
				case "rename_guard_and_references":
					for i := range m.Guards {
						if m.Guards[i].ID == c.ModelChange.From {
							m.Guards[i].ID = c.ModelChange.To
						}
					}
					for i := range m.Transitions {
						for j := range m.Transitions[i].Guards {
							if m.Transitions[i].Guards[j] == c.ModelChange.From {
								m.Transitions[i].Guards[j] = c.ModelChange.To
							}
						}
					}
				default:
					t.Fatal("unknown model change")
				}
			}
			beforeModel, _ := json.Marshal(m)
			beforeReport, _ := json.Marshal(r)
			got, e := Explain(m, r)
			afterModel, _ := json.Marshal(m)
			afterReport, _ := json.Marshal(r)
			if !bytes.Equal(beforeModel, afterModel) || !bytes.Equal(beforeReport, afterReport) {
				t.Fatal("input mutated")
			}
			if c.Expected.Result == "reject" {
				if e == nil || !reflect.DeepEqual(got, Explanation{}) {
					t.Fatal("expected complete rejection")
				}
				if c.ModelChange == nil {
					if out, err := Execute(requestBytes(modelRaw, c.Report)); err == nil || out != nil {
						t.Fatal("Execute accepted invalid case")
					}
				}
				return
			}
			if e != nil {
				t.Fatal(e)
			}
			var want Explanation
			if json.Unmarshal(c.Expected.Explanation, &want) != nil {
				t.Fatal("bad expectation")
			}
			if !reflect.DeepEqual(got, want) {
				g, _ := json.Marshal(got)
				w, _ := json.Marshal(want)
				t.Fatalf("literal mismatch\ngot %s\nwant %s", g, w)
			}
			raw, e := Execute(requestBytes(modelRaw, c.Report))
			if e != nil {
				t.Fatal(e)
			}
			var out Explanation
			if json.Unmarshal(raw, &out) != nil || !reflect.DeepEqual(out, want) {
				t.Fatal("Execute differs from frozen expectation")
			}
			again, e := Execute(requestBytes(modelRaw, c.Report))
			if e != nil || !bytes.Equal(raw, again) {
				t.Fatal("nondeterministic output")
			}
		})
	}
	for _, c := range v.Raw {
		t.Run(c.ID, func(t *testing.T) {
			if seen[c.ID] {
				t.Fatal("duplicate vector")
			}
			seen[c.ID] = true
			b, e := base64.StdEncoding.DecodeString(c.Bytes)
			if e != nil || c.Expected != "reject" {
				t.Fatal("bad raw vector")
			}
			switch c.Entrypoint {
			case "DecodeReport":
				if r, e := scenarios.DecodeReport(b); e == nil || !reflect.DeepEqual(r, scenarios.Report{}) {
					t.Fatal("invalid report accepted")
				}
			case "Execute":
				if out, e := Execute(b); e == nil || out != nil {
					t.Fatal("invalid request accepted")
				}
			default:
				t.Fatal("unknown entrypoint")
			}
		})
	}
}
func TestBoundsAndOwnership(t *testing.T) {
	v, _, m := fixtures(t)
	r, e := scenarios.DecodeReport(v.Cases[0].Report)
	if e != nil {
		t.Fatal(e)
	}
	out, e := Explain(m, r)
	if e != nil {
		t.Fatal(e)
	}
	savedSource := *out.SourceDigest
	*r.SourceDigest = strings.Repeat("d", 64)
	if *out.SourceDigest != savedSource {
		t.Fatal("source digest pointer aliases input")
	}
	saved := out.Entries[0].Guards[0].RequirementText
	savedRef := out.Entries[0].Guards[0].Expression.Args[0].Ref
	for i := range m.Guards {
		if m.Guards[i].ID == "G_outcome" {
			m.Guards[i].Expr.Args[0].Ref = "changed-input"
		}
	}
	if out.Entries[0].Guards[0].Expression.Args[0].Ref != savedRef {
		t.Fatal("nested input aliases output")
	}
	m.Requirements[0].Text = "changed"
	r.Entries[0].Expected.Refusals[0] = "changed"
	r.Entries[0].Actual.Refusals[0] = "changed"
	if out.Entries[0].ExpectedRefusals[0] == "changed" || out.Entries[0].ActualRefusals[0] == "changed" || out.Entries[0].Guards[0].RequirementText != saved {
		t.Fatal("input aliases output")
	}
	modelBefore, _ := json.Marshal(m)
	reportBefore, _ := json.Marshal(r)
	out.Entries[0].Guards[0].Expression.Args[0].Ref = "changed-output"
	out.Entries[0].ActualRefusals[0] = "changed-output"
	modelAfter, _ := json.Marshal(m)
	reportAfter, _ := json.Marshal(r)
	if !bytes.Equal(modelBefore, modelAfter) || !bytes.Equal(reportBefore, reportAfter) {
		t.Fatal("output aliases input")
	}
	if raw, err := encode(Explanation{Status: strings.Repeat("x", MaxOutputBytes+1)}); err == nil || raw != nil {
		t.Fatal("oversized output accepted")
	}
	if x, e := Execute(bytes.Repeat([]byte("x"), MaxRequestBytes+1)); e == nil || x != nil {
		t.Fatal("oversized input")
	}
	if _, e := scenarios.DecodeReport([]byte(strings.Repeat("x", scenarios.MaxReportBytes+1))); e == nil {
		t.Fatal("oversized report")
	}
}

func TestStrictRequestLimits(t *testing.T) {
	v, modelRaw, _ := fixtures(t)
	report := []byte(v.Cases[0].Report)
	hugeModel := bytes.Replace(modelRaw, []byte(`"label": "Draft"`), []byte(`"label": "`+strings.Repeat("x", MaxModelBytes)+`"`), 1)
	if out, err := Execute(requestBytes(hugeModel, report)); err == nil || out != nil {
		t.Fatal("oversized model accepted")
	}
	hugeReport := bytes.Replace(report, []byte(`"Case_example"`), []byte(`"`+strings.Repeat("x", scenarios.MaxReportBytes)+`"`), 1)
	if out, err := Execute(requestBytes(modelRaw, hugeReport)); err == nil || out != nil {
		t.Fatal("oversized report accepted")
	}
	deep := []byte(strings.Repeat("[", MaxRequestDepth+1) + "0" + strings.Repeat("]", MaxRequestDepth+1))
	if out, err := Execute(deep); err == nil || out != nil {
		t.Fatal("overdepth request accepted")
	}
	for _, raw := range [][]byte{[]byte(`{"formatVersion":"lifecycle-scenario-explain-request/v1","model":null,"report":null}`), []byte(`{"formatVersion":"lifecycle-scenario-explain-request/v1","model":{},"report":null}`)} {
		if out, err := Execute(raw); err == nil || out != nil {
			t.Fatal("null wrapper value accepted")
		}
	}
}
