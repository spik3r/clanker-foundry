package boundinventory

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/canonical"
	g "github.com/spik3r/clanker-foundry/modeling/lifecycle/generated"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/jsonformat"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/model"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/scenarioexplain"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/scenarios"
)

// The reports are fabricated consistency observations, not executed evidence.
// Literal bucket membership was handwritten before production. Native APIs below
// validate identities and envelopes only; their callback must never be invoked.
type frozenVector struct {
	ID, Model, Suite, Report string
	Expected                 json.RawMessage
}
type frozenRequest struct {
	FormatVersion string `json:"formatVersion"`
	Model         string `json:"model"`
	Suite         string `json:"suite"`
	Report        string `json:"report"`
}

func frozenSHA(b []byte) string { d := sha256.Sum256(b); return hex.EncodeToString(d[:]) }
func frozenJSON(t *testing.T, v any) []byte {
	t.Helper()
	var b bytes.Buffer
	e := json.NewEncoder(&b)
	e.SetEscapeHTML(false)
	if e.Encode(v) != nil {
		t.Fatal("fixture encoding")
	}
	return b.Bytes()
}
func frozenVectors(t *testing.T) []frozenVector {
	t.Helper()
	b, e := os.ReadFile("testdata/observations.json")
	if e != nil {
		t.Fatal(e)
	}
	if frozenSHA(b) != "b8696125353e92d915f6f7dee7b0d1674f03b513ad660f174dca4b306ff8e563" {
		t.Fatal("independent fixed observations changed")
	}
	var x struct{ Vectors []frozenVector }
	if json.Unmarshal(b, &x) != nil || len(x.Vectors) != 3 {
		t.Fatal("fixtures")
	}
	return x.Vectors
}
func frozenWire(t *testing.T, v frozenVector) []byte {
	return frozenJSON(t, frozenRequest{"lifecycle-bound-assertion-inventory-request/v1", v.Model, v.Suite, v.Report})
}
func frozenNative(t *testing.T, v frozenVector) (model.Model, scenarios.Runner, scenarios.Suite, scenarios.Report) {
	t.Helper()
	m, e := jsonformat.Decode([]byte(v.Model))
	if e != nil {
		t.Fatal(e)
	}
	md, e := canonical.Digest(m)
	if e != nil {
		t.Fatal(e)
	}
	r, e := scenarios.NewRunner(scenarios.Binding{ProfileVersion: "lifecycle/v1", ModelDigest: md}, func(g.Task, g.Command) g.Result {
		t.Fatal("inventory/native proof executed a reducer")
		return g.Result{}
	})
	if e != nil {
		t.Fatal(e)
	}
	s, e := r.Decode([]byte(v.Suite))
	if e != nil {
		t.Fatal("suite", e)
	}
	p, e := r.DecodeReport([]byte(v.Report))
	if e != nil {
		t.Fatal("report", e)
	}
	return m, r, s, p
}
func TestIndependentNativeInventoryControls(t *testing.T) {
	for _, v := range frozenVectors(t) {
		t.Run(v.ID, func(t *testing.T) {
			m, r, s, p := frozenNative(t, v)
			md, _ := canonical.Digest(m)
			sd, e := r.Digest(s)
			if e != nil || p.ModelDigest != md || p.ScenarioDigest != sd || p.SourceDigest == nil || *p.SourceDigest != frozenSHA([]byte(v.Suite)) {
				t.Fatal("fixed identities")
			}
			if md == g.ModelDigest {
				t.Fatal("fixture must remain a nondefault model")
			}
			if _, e := scenarioexplain.ExplainWithRunner(m, p, r); e != nil {
				t.Fatal("native explanation", e)
			}
			var want struct{ Inputs map[string]string }
			if json.Unmarshal(v.Expected, &want) != nil {
				t.Fatal("expected")
			}
			for key, value := range map[string]string{"modelSHA256": frozenSHA([]byte(v.Model)), "suiteSHA256": frozenSHA([]byte(v.Suite)), "reportSHA256": frozenSHA([]byte(v.Report)), "modelDigest": md, "scenarioDigest": sd} {
				if want.Inputs[key] != value {
					t.Fatal("independent source expectation", key)
				}
			}
		})
	}
}
func assertFrozenInventory(t *testing.T, check func([]byte) ([]byte, error)) {
	t.Helper()
	vectors := frozenVectors(t)
	for _, v := range vectors {
		t.Run(v.ID, func(t *testing.T) {
			input := frozenWire(t, v)
			saved := bytes.Clone(input)
			out, e := check(input)
			if e != nil {
				t.Fatal(e)
			}
			var got, want any
			if json.Unmarshal(out, &got) != nil || json.Unmarshal(v.Expected, &want) != nil || !reflect.DeepEqual(got, want) {
				t.Fatalf("fixed inventory differs\ngot%s\nwant%s", out, v.Expected)
			}
			if !bytes.Equal(input, saved) {
				t.Fatal("input mutated")
			}
			again, e := check(input)
			if e != nil || !bytes.Equal(out, again) {
				t.Fatal("repeatability")
			}
			if len(out) > 0 {
				out[0] = '!'
			}
			again, e = check(input)
			if e != nil || !json.Valid(again) {
				t.Fatal("returned storage aliases later call")
			}
		})
	}
	base := vectors[0]
	mutations := []struct {
		name  string
		apply func(map[string]any)
	}{
		{"wrong-source", func(p map[string]any) { p["sourceDigest"] = string(bytes.Repeat([]byte("a"), 64)) }},
		{"missing-source", func(p map[string]any) { p["sourceDigest"] = nil }},
		{"wrong-scenario", func(p map[string]any) { p["scenarioDigest"] = string(bytes.Repeat([]byte("a"), 64)) }},
		{"wrong-model", func(p map[string]any) { p["modelDigest"] = string(bytes.Repeat([]byte("a"), 64)) }},
		{"changed-command", func(p map[string]any) {
			p["entries"].([]any)[0].(map[string]any)["command"].(map[string]any)["actor"] = "Other"
		}},
		{"changed-literal-and-actual", func(p map[string]any) {
			e := p["entries"].([]any)[0].(map[string]any)
			for _, k := range []string{"expected", "actual"} {
				e[k].(map[string]any)["task"].(map[string]any)["revision"] = "17"
			}
		}},
		{"changed-initial", func(p map[string]any) {
			p["entries"].([]any)[0].(map[string]any)["input"].(map[string]any)["id"] = "different-input"
		}},
		{"wrong-flow-after-null-refusal", func(p map[string]any) {
			p["entries"].([]any)[4].(map[string]any)["input"].(map[string]any)["revision"] = "7"
		}},
		{"reordered-steps", func(p map[string]any) { e := p["entries"].([]any); e[3], e[4] = e[4], e[3] }},
		{"missing-suite-step", func(p map[string]any) {
			e := p["entries"].([]any)
			p["entries"] = append(e[:4], e[5:]...)
			p["stepsExecuted"] = float64(6)
		}},
		{"false-completed-count", func(p map[string]any) { p["casesCompleted"] = float64(4) }},
		{"unfinished-matching-prefix", func(p map[string]any) {
			p["entries"] = p["entries"].([]any)[:4]
			p["stepsExecuted"] = float64(4)
			p["casesCompleted"] = float64(3)
		}},
	}
	for _, x := range mutations {
		t.Run(x.name, func(t *testing.T) {
			v := base
			var p map[string]any
			if json.Unmarshal([]byte(v.Report), &p) != nil {
				t.Fatal("report")
			}
			x.apply(p)
			v.Report = string(frozenJSON(t, p))
			if out, e := check(frozenWire(t, v)); e == nil || out != nil {
				t.Fatal("tampered report produced inventory")
			}
		})
	}
	t.Run("actual-undeclared-transition-even-at-mismatch", func(t *testing.T) {
		v := vectors[1]
		var p map[string]any
		_ = json.Unmarshal([]byte(v.Report), &p)
		e := p["entries"].([]any)[4].(map[string]any)
		e["actual"].(map[string]any)["transitionId"] = "T_unlisted"
		e["differences"] = []string{"transitionId", "task"}
		v.Report = string(frozenJSON(t, p))
		if out, e := check(frozenWire(t, v)); e == nil || out != nil {
			t.Fatal("unmapped actual mismatch accepted")
		}
	})
	valid := frozenWire(t, base)
	for _, bad := range [][]byte{nil, []byte("null"), []byte(`{}`), append(bytes.Clone(valid), valid...), bytes.Replace(valid, []byte(`"report":`), []byte(`"REPORT":`), 1), bytes.Replace(valid, []byte(`"report":`), []byte(`"report":null,"report":`), 1)} {
		if out, e := check(bad); e == nil || out != nil {
			t.Fatal("strict wrapper accepted")
		}
	}
	if out, e := check(bytes.Repeat([]byte(" "), (16<<20)+1)); e == nil || out != nil {
		t.Fatal("wrapper bound")
	}
}
