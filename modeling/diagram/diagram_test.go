package diagram

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spik3r/clanker-foundry/modeling/diagram/contract"
)

var exampleNames = []string{"order", "parallel", "rework"}

// TestExamplesMatchGenerated regenerates every example and compares bytes with
// the files committed under examples/generated.
func TestExamplesMatchGenerated(t *testing.T) {
	for _, name := range exampleNames {
		t.Run(name, func(t *testing.T) {
			raw, err := os.ReadFile(filepath.Join("examples", name+".json"))
			if err != nil {
				t.Fatal(err)
			}
			res, err := Build(raw)
			if err != nil {
				t.Fatal(err)
			}
			drift, err := contract.Drift(res.Files, filepath.Join("examples", "generated", name))
			if err != nil {
				t.Fatal(err)
			}
			if len(drift) > 0 {
				t.Fatalf("committed output differs from regenerated output (run: go run ./cmd/diagramgen generate examples/%s.json <new dir>):\n%s", name, strings.Join(drift, "\n"))
			}
			for _, want := range []string{"diagram.svg", "diagram.bpmn", "model.json", "foundry-manifest.json"} {
				if _, ok := res.Files[want]; !ok {
					t.Errorf("output lacks %s", want)
				}
			}
			if len(res.Files) != 4 {
				t.Errorf("%d output files, want 4", len(res.Files))
			}
		})
	}
}

func TestManifestAndIdentity(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("examples", "order.json"))
	if err != nil {
		t.Fatal(err)
	}
	res, err := Build(raw)
	if err != nil {
		t.Fatal(err)
	}
	var id struct{ Contract, Profile, SourceSHA256, CanonicalSHA256 string }
	if err := json.Unmarshal([]byte(res.Identity), &id); err != nil {
		t.Fatal(err)
	}
	if id.Contract != "foundry-generator/v1" || id.Profile != "process-diagram/v1" || id.SourceSHA256 != contract.HexSum(raw) {
		t.Errorf("identity %+v", id)
	}
	var m struct {
		Profile   string
		Generator struct{ ID, Version, Toolchain string }
		Model     struct{ CanonicalSHA256 string }
		Files     []struct {
			Path   string
			SHA256 string
			Bytes  int
		}
	}
	if err := json.Unmarshal(res.Files[contract.ManifestName], &m); err != nil {
		t.Fatal(err)
	}
	if m.Generator.ID != "diagramgen" || m.Generator.Version != "1.0.0" || m.Generator.Toolchain != "go1.27.4" || m.Profile != id.Profile {
		t.Errorf("manifest %+v", m)
	}
	if m.Model.CanonicalSHA256 != id.CanonicalSHA256 {
		t.Errorf("manifest and identity digests differ")
	}
	want := []string{"diagram.bpmn", "diagram.svg", "model.json"}
	if len(m.Files) != len(want) {
		t.Fatalf("manifest lists %d files", len(m.Files))
	}
	for i, f := range m.Files {
		if f.Path != want[i] || f.SHA256 != contract.HexSum(res.Files[f.Path]) || f.Bytes != len(res.Files[f.Path]) {
			t.Errorf("manifest entry %+v", f)
		}
	}
	if !strings.HasSuffix(string(res.Files["model.json"]), "}\n") || !strings.Contains(string(res.Files["model.json"]), "\n  \"id\"") {
		t.Errorf("model.json is not indented with a trailing newline")
	}
}

func TestSourceFormattingDoesNotChangeOutput(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("examples", "order.json"))
	if err != nil {
		t.Fatal(err)
	}
	var v any
	if err := json.Unmarshal(raw, &v); err != nil {
		t.Fatal(err)
	}
	compact, _ := json.Marshal(v) // also reorders object members
	a, err := Build(raw)
	if err != nil {
		t.Fatal(err)
	}
	b, err := Build(compact)
	if err != nil {
		t.Fatal(err)
	}
	for name, data := range a.Files {
		if string(data) != string(b.Files[name]) {
			t.Errorf("%s differs for a reformatted source", name)
		}
	}
	if a.Identity == b.Identity {
		t.Errorf("identity lines should differ in sourceSHA256")
	}
}

func TestSelect(t *testing.T) {
	if got := Supported(); len(got) != 1 || got[0] != "process-diagram/v1" {
		t.Errorf("supported profiles %v", got)
	}
	for _, raw := range []string{
		``, `[]`, `{}`, `{"profileVersion":"class-diagram/v1"}`, `{"profileVersion":7}`, `not json`,
		`{"profileVersion":"process-diagram/v1"}`, // selected, then refused by the profile
	} {
		_, err := Build([]byte(raw))
		var me contract.ModelError
		if err == nil {
			t.Errorf("Build(%q) accepted", raw)
		} else if !errors.As(err, &me) {
			t.Errorf("Build(%q) error %v is not a model error", raw, err)
		}
	}
}
