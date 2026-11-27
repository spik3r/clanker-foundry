package diagram

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/spik3r/clanker-foundry/modeling/diagram/contract"
	"github.com/spik3r/clanker-foundry/modeling/diagram/internal/svgcheck"
)

func TestDiagramStartersReplayAndDetectDrift(t *testing.T) {
	for _, name := range []string{"uml-class", "uml-sequence", "bpmn-process", "architecture"} {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join("..", "..", "templates", "diagrams", name+".json")
			raw, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			res, err := Build(raw)
			if err != nil {
				t.Fatal(err)
			}
			var model map[string]any
			if err := json.Unmarshal(raw, &model); err != nil {
				t.Fatal(err)
			}
			out := filepath.Join(t.TempDir(), "diagram")
			if err := contract.Write(out, res.Files); err != nil {
				t.Fatal(err)
			}
			formatted, err := json.Marshal(model)
			if err != nil {
				t.Fatal(err)
			}
			replay, err := Build(formatted)
			if err != nil {
				t.Fatal(err)
			}
			if drift, err := contract.Drift(replay.Files, out); err != nil || len(drift) > 0 {
				t.Fatalf("formatting-only change caused drift: %v %v", drift, err)
			}
			if err := svgcheck.Inert(res.Files["diagram.svg"]); err != nil {
				t.Fatal(err)
			}
			if name == "bpmn-process" && !bytes.Contains(res.Files["diagram.bpmn"], []byte(`sourceRef="Available" targetRef="Accept"`)) {
				t.Fatal("BPMN export omitted expected gateway branch")
			}
			model["title"] = "A reviewed change"
			changed, err := json.Marshal(model)
			if err != nil {
				t.Fatal(err)
			}
			replay, err = Build(changed)
			if err != nil {
				t.Fatal(err)
			}
			if drift, err := contract.Drift(replay.Files, out); err != nil || len(drift) == 0 {
				t.Fatalf("meaningful change did not report drift: %v %v", drift, err)
			}
		})
	}
}
