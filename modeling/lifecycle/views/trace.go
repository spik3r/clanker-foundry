package views

import (
	"crypto/sha256"
	"fmt"
	"sort"

	"github.com/spik3r/clanker-foundry/modeling/lifecycle/model"
)

type CodeArtifacts struct{ Implementation, DerivedTests []byte }
type artifactDigest struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
}
type location struct {
	ID           string   `json:"id"`
	Kind         string   `json:"kind"`
	Path         string   `json:"semanticPath"`
	Requirements []string `json:"requirements"`
	Guards       []string `json:"guards"`
	Transitions  []string `json:"transitions"`
}

// Trace binds supplied artifact bytes to declared model relationships. Hashes do
// not establish that the bytes were generated correctly or independently tested.
func Trace(m model.Model, id Identity, view Artifacts, code CodeArtifacts) ([]byte, error) {
	m, err := normalized(m, id)
	if err != nil {
		return nil, err
	}
	files := []struct {
		path string
		data []byte
	}{
		{"generated/decision-table.json", view.DecisionTable},
		{"generated/decision.svg", view.SVG},
		{"generated/lifecycle.go", code.Implementation},
		{"generated/model_paths_test.go", code.DerivedTests},
	}
	digests := []artifactDigest{}
	for _, f := range files {
		if len(f.data) == 0 {
			return nil, fmt.Errorf("missing artifact: %s", f.path)
		}
		digests = append(digests, artifactDigest{f.path, fmt.Sprintf("%x", sha256.Sum256(f.data))})
	}
	locations := []location{}
	fresh := func(id, kind string) location {
		return location{ID: id, Kind: kind, Path: "model/" + m.ID + "/" + kind + "/" + id, Requirements: []string{}, Guards: []string{}, Transitions: []string{}}
	}
	guardReq := map[string]string{}
	for _, g := range m.Guards {
		guardReq[g.ID] = g.Requirement
	}
	for _, s := range m.States {
		l := fresh(s.ID, "states")
		for _, t := range m.Transitions {
			if t.From == s.ID || t.To == s.ID {
				l.Transitions = append(l.Transitions, t.ID)
			}
		}
		locations = append(locations, l)
	}
	for _, r := range m.Requirements {
		l := fresh(r.ID, "requirements")
		for _, g := range m.Guards {
			if g.Requirement == r.ID {
				l.Guards = append(l.Guards, g.ID)
			}
		}
		for _, t := range m.Transitions {
			for _, g := range t.Guards {
				if guardReq[g] == r.ID {
					l.Transitions = append(l.Transitions, t.ID)
					break
				}
			}
		}
		locations = append(locations, l)
	}
	for _, g := range m.Guards {
		l := fresh(g.ID, "guards")
		l.Requirements = append(l.Requirements, g.Requirement)
		for _, t := range m.Transitions {
			for _, gid := range t.Guards {
				if gid == g.ID {
					l.Transitions = append(l.Transitions, t.ID)
				}
			}
		}
		locations = append(locations, l)
	}
	for _, t := range m.Transitions {
		l := fresh(t.ID, "transitions")
		l.Guards = append(l.Guards, t.Guards...)
		reqs := map[string]bool{}
		for _, g := range t.Guards {
			reqs[guardReq[g]] = true
		}
		for r := range reqs {
			l.Requirements = append(l.Requirements, r)
		}
		sort.Strings(l.Requirements)
		locations = append(locations, l)
	}
	sort.Slice(locations, func(i, j int) bool { return locations[i].ID < locations[j].ID })
	return encode(struct {
		Schema           string           `json:"schema"`
		Status           string           `json:"status"`
		Identity         Identity         `json:"identity"`
		SourceConvention string           `json:"sourceConvention"`
		Artifacts        []artifactDigest `json:"artifacts"`
		Locations        []location       `json:"locations"`
	}{"lifecycle-trace/v1", intent, id, "Stable semantic-ID paths; no source line numbers or verification evidence.", digests, locations})
}
