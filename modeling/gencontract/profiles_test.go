package gencontract

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The profile registry (../profiles.json) tells an author which profiles
// exist, where each is documented and what it cannot express. These checks keep
// it consistent with the generator registry and with the files it names.

type profileEntry struct {
	Profile       string   `json:"profile"`
	Generator     string   `json:"generator"`
	Summary       string   `json:"summary"`
	ProfileDoc    string   `json:"profileDoc"`
	Schema        string   `json:"schema"`
	Examples      []string `json:"examples"`
	Emits         []string `json:"emits"`
	CannotExpress []string `json:"cannotExpress"`
}

type profileRegistry struct {
	RegistryVersion string         `json:"registryVersion"`
	Profiles        []profileEntry `json:"profiles"`
}

type generatorEntry struct {
	ID      string `json:"id"`
	Profile string `json:"profile"`
	Module  string `json:"module"`
	Package string `json:"package"`
	Model   string `json:"model"`
}

func decodeStrict(t *testing.T, path string, v any) {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		t.Fatalf("%s: %v", path, err)
	}
	if dec.More() {
		t.Fatalf("%s: data after the registry object", path)
	}
}

func requireRegularFile(t *testing.T, root, rel string) string {
	t.Helper()
	if rel == "" || filepath.IsAbs(rel) || strings.Contains(rel, "\\") {
		t.Fatalf("path %q must be a relative path with / separators", rel)
	}
	for _, part := range strings.Split(rel, "/") {
		if part == "" || part == "." || part == ".." {
			t.Fatalf("path %q has an empty, . or .. segment", rel)
		}
	}
	full := filepath.Join(root, filepath.FromSlash(rel))
	info, err := os.Lstat(full)
	if err != nil {
		t.Fatalf("%s: %v", rel, err)
	}
	if !info.Mode().IsRegular() {
		t.Fatalf("%s is not a regular file", rel)
	}
	return full
}

func TestProfileRegistryIsConsistent(t *testing.T) {
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	var profiles profileRegistry
	decodeStrict(t, "../profiles.json", &profiles)
	if profiles.RegistryVersion != "foundry-profile-registry/v1" {
		t.Fatalf("registryVersion is %q", profiles.RegistryVersion)
	}
	if len(profiles.Profiles) == 0 {
		t.Fatal("registry lists no profiles")
	}

	var generators struct {
		RegistryVersion string           `json:"registryVersion"`
		Generators      []generatorEntry `json:"generators"`
	}
	decodeStrict(t, "generators.json", &generators)
	byID := map[string]generatorEntry{}
	for _, g := range generators.Generators {
		byID[g.ID] = g
	}

	// Every registered generator except the harness's own reference generator
	// has at least one profile entry; a generator may serve several profiles.
	entriesFor := map[string]int{}
	seenProfile := map[string]bool{}
	for _, e := range profiles.Profiles {
		entriesFor[e.Generator]++
		if seenProfile[e.Profile] {
			t.Errorf("profile %q is listed twice", e.Profile)
		}
		seenProfile[e.Profile] = true
	}
	for _, g := range generators.Generators {
		switch {
		case g.ID == "refgen":
			if entriesFor[g.ID] != 0 {
				t.Errorf("refgen is the harness reference and must not have a profile entry")
			}
		case entriesFor[g.ID] < 1:
			t.Errorf("generator %q has no profile entry", g.ID)
		}
	}

	// The profile the conformance harness drives must be one the generator
	// serves.
	for _, g := range generators.Generators {
		if g.ID == "refgen" {
			continue
		}
		served := false
		for _, e := range profiles.Profiles {
			if e.Generator == g.ID && e.Profile == g.Profile {
				served = true
			}
		}
		if !served {
			t.Errorf("generator %q is registered for profile %q, which has no profile entry for it", g.ID, g.Profile)
		}
	}

	for _, e := range profiles.Profiles {
		t.Run(e.Profile, func(t *testing.T) {
			_, ok := byID[e.Generator]
			if !ok {
				t.Fatalf("generator %q is not in generators.json", e.Generator)
			}
			if strings.TrimSpace(e.Summary) == "" {
				t.Error("summary is empty")
			}
			if len(e.Emits) == 0 {
				t.Error("emits is empty")
			}
			for _, s := range e.Emits {
				if strings.TrimSpace(s) == "" {
					t.Error("emits has an empty entry")
				}
			}
			if len(e.CannotExpress) == 0 {
				t.Error("cannotExpress is empty")
			}
			for _, s := range e.CannotExpress {
				if strings.TrimSpace(s) == "" {
					t.Error("cannotExpress has an empty entry")
				}
			}
			requireRegularFile(t, root, e.ProfileDoc)
			requireRegularFile(t, root, e.Schema)
			if len(e.Examples) == 0 {
				t.Fatal("no examples listed")
			}
			for _, rel := range e.Examples {
				raw, err := os.ReadFile(requireRegularFile(t, root, rel))
				if err != nil {
					t.Fatal(err)
				}
				if !json.Valid(raw) {
					t.Fatalf("%s is not valid JSON", rel)
				}
				var doc map[string]json.RawMessage
				if err := json.Unmarshal(raw, &doc); err != nil {
					t.Fatalf("%s: %v", rel, err)
				}
				key := "profileVersion"
				if _, ok := doc[key]; !ok {
					key = "profile"
				}
				var declared string
				if err := json.Unmarshal(doc[key], &declared); err != nil {
					t.Fatalf("%s: %s is missing or not a string", rel, key)
				}
				if declared != e.Profile {
					t.Fatalf("%s declares profile %q, want %q", rel, declared, e.Profile)
				}
			}
		})
	}
}
