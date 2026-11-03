// Independent pre-adoption oracle. Candidate and suite literals predate the
// exporter. No Build, Verify, compiler.Generate or runner output authored these
// expectations. Source inventory is pinned to the independently inspected main.
package qualification

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/artifactbundle"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/canonical"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/jsonformat"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
)

const adoptionSourceRevision = "3d6d17188cc2723cfe04649d0f8b4220a37b1afe"
const adoptionCandidateSHA = "9674eb213352232f354c2c338709baf6a3230a599ef89845e1cfac5525d4c6e1"
const adoptionSuiteSHA = "53b5ceec6b68719243853dae5285c13d5c25df89cb203a16279f3cacd2f989f1"
const adoptionModelDigest = "a9ad2dfb926a24a026c8565f0d929a5b803bec43f6eb7342e196830791eccc7c"

// Calculated independently by serializing the original scenario DTO field order,
// excluding formatVersion and final LF; never from an exported execution report.
const adoptionScenarioDigest = "7a50d74f611ddf854504eceaa34d85a42bb5e56777b80ab84332921d167f3975"

type adoptionSourcePin struct {
	name string
	size int
	sha  string
}

var adoptionExpectedSources = []adoptionSourcePin{
	{"go.mod", 71, "cc128fce13193d3066bdeb08cca9db627fb6c3120967c3beff254a514f2a5fac"},
	{"model/ir.go", 9385, "13f4ad2980b1ca523b07bdbc8ebeb0e0300f8df81faa8dcd1cdec5ccef007aeb"},
	{"canonical/canonical.go", 1946, "2c3cba3fd53adf8ca5a4cc2ce3d96b51457742c47b6ab4fcedba219e19001a06"},
	{"strictjson/decode.go", 4236, "a72ded4af99ffba5f5cca2931c6cace072634ee2e42eb00cd9a5988c691f1ee2"},
	{"compiler/generate.go", 12315, "e17610a27748a5d3baedf4f59c2183cb184f7e2681b5532c6f9d8d3d1eda6a38"},
	{"scenarios/types.go", 5904, "147b12bd1726fec156ba08ec5666a9a362e1563115151ae57bf9face70f86eec"},
	{"scenarios/decode.go", 7047, "e6119229899d3084820d72a777d1771037f871b14554630067b7da022cdedc15"},
	{"scenarios/run.go", 10114, "46d41c5d27d87f71f7191f18d182a99c3ca491126afb601599067692c3379dde"},
	{"cmd/modelplay/main.go", 1613, "925ec202c74aba5d50eb2bfb5dc54ea22387ccec90e969f1a7ebe1ecbd9a0c63"},
	{"scenarios/runner.go", 4005, "e85b2146d7c1d16269ff886f6c67084de5e263db8b01a7676b8c7491ed8fb59c"},
	{"scenarios/report_decode.go", 3626, "f14816a19faacf707e23b01d5d16b26be8e8c03a9f5f9af801029b89fb2ceddf"},
}

func adoptionSHA(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }
func adoptionPinnedSources(t *testing.T) []artifactbundle.File {
	t.Helper()
	out := []artifactbundle.File{}
	for _, p := range adoptionExpectedSources {
		b, e := os.ReadFile(filepath.Join("..", p.name))
		if e != nil {
			t.Fatal(e)
		}
		if len(b) != p.size || adoptionSHA(b) != p.sha {
			t.Fatal("current-main source identity drift", p.name)
		}
		out = append(out, artifactbundle.File{Name: p.name, Data: b})
	}
	return out
}
func TestAdoptionFrozenSourceAndInputIdentities(t *testing.T) {
	if SourceRevision != "1e88dbebb4ab13873e2a4159dc928dd1f5f73e20" {
		t.Fatal("source revision relabeled")
	}
	if len(candidate) != 8769 || adoptionSHA([]byte(candidate)) != adoptionCandidateSHA || len(suite) != 6656 || adoptionSHA([]byte(suite)) != adoptionSuiteSHA {
		t.Fatal("original candidate/suite literal changed")
	}
	m, e := jsonformat.Decode([]byte(candidate))
	if e != nil {
		t.Fatal(e)
	}
	d, e := canonical.Digest(m)
	if e != nil || d != adoptionModelDigest {
		t.Fatal("candidate semantic identity changed", d, e)
	}
	want := []string{}
	for _, p := range adoptionExpectedSources {
		want = append(want, p.name)
	}
	got := adoptionLegacySourceNames(t)
	sort.Strings(got)
	sort.Strings(want)
	if !reflect.DeepEqual(got, want) {
		t.Fatal("source names differ from independent eleven-file inventory", got)
	}
	// SourceNames must not expose mutable pin state.
	exposed := adoptionLegacySourceNames(t)
	exposed[0] = "corrupted caller list"
	fresh := adoptionLegacySourceNames(t)
	sort.Strings(fresh)
	if !reflect.DeepEqual(fresh, want) {
		t.Fatal("source list aliases implementation")
	}
	adoptionPinnedSources(t)
}
func adoptionFiles(t *testing.T, raw []byte) []artifactbundle.File {
	t.Helper()
	f, e := artifactbundle.Verify(raw, artifactbundle.QualificationProfile)
	if e != nil {
		t.Fatal(e)
	}
	return f
}
func adoptionMetadata(t *testing.T, files []artifactbundle.File) {
	t.Helper()
	var raw []byte
	for _, f := range files {
		if f.Name == "qualification.json" {
			raw = f.Data
		}
	}
	var got map[string]any
	if e := json.Unmarshal(raw, &got); e != nil {
		t.Fatal(e)
	}
	want := map[string]any{"format": "lifecycle-json/v1", "profile": "lifecycle/v1", "generator": "lifecycle-go/v1", "toolchain": "go1.27.3", "modelDigest": adoptionModelDigest, "sourceRevision": adoptionSourceRevision, "suiteStatus": "unvalidated and unexecuted"}
	if !reflect.DeepEqual(got, want) {
		t.Fatal("identity or suite status promoted/altered", got)
	}
}
func TestAdoptionDeterministicCompleteOpaqueBundle(t *testing.T) {
	src := adoptionPinnedSources(t)
	a, e := adoptionLegacyBuild([]byte(candidate), []byte(suite), src)
	if e != nil {
		t.Fatal(e)
	}
	b, e := adoptionLegacyBuild([]byte(candidate), []byte(suite), src)
	if e != nil || !bytes.Equal(a, b) {
		t.Fatal("nondeterministic bundle", e)
	}
	if e = Verify(a, adoptionCandidateSHA, adoptionSuiteSHA); e != nil {
		t.Fatal(e)
	}
	files := adoptionFiles(t, a)
	if len(files) != 17 {
		t.Fatal("generic Verify must return17payloads,excludingmanifest", len(files))
	}
	z, e := zip.NewReader(bytes.NewReader(a), int64(len(a)))
	if e != nil || len(z.File) != 18 {
		t.Fatal("archive must contain17payloads+manifest", e)
	}
	manifestCount := 0
	for _, f := range z.File {
		if f.Name == "manifest.json" {
			manifestCount++
		}
	}
	if manifestCount != 1 {
		t.Fatal("missing or duplicate manifest")
	}
	want := []string{"generated/lifecycle.go", "generated/model_paths_test.go", "candidate.json", "suite.json", "qualification.json", "GUIDE.txt"}
	for _, p := range adoptionExpectedSources {
		want = append(want, p.name)
	}
	sort.Strings(want)
	got := []string{}
	byName := map[string][]byte{}
	for _, f := range files {
		got = append(got, f.Name)
		byName[f.Name] = f.Data
	}
	sort.Strings(got)
	if !reflect.DeepEqual(got, want) {
		t.Fatal("payload inventory drift", got)
	}
	if !bytes.Equal(byName["candidate.json"], []byte(candidate)) || !bytes.Equal(byName["suite.json"], []byte(suite)) {
		t.Fatal("candidate or independent suite rewritten")
	}
	for _, p := range adoptionExpectedSources {
		if len(byName[p.name]) != p.size || adoptionSHA(byName[p.name]) != p.sha {
			t.Fatal("archive source differs", p.name)
		}
	}
	adoptionMetadata(t, files)
	if !bytes.Contains(byName["GUIDE.txt"], []byte("no review or execution authority")) {
		t.Fatal("guide omitted authority boundary")
	}
	// Even malformed suite text remains exact opaque data; export success is not
	// scenario validation, a passing suite, or authority to execute anything.
	malformed := []byte("{ malformed suite\n")
	raw, e := adoptionLegacyBuild([]byte(candidate), malformed, src)
	if e != nil {
		t.Fatal(e)
	}
	if e = Verify(raw, adoptionCandidateSHA, adoptionSHA(malformed)); e != nil {
		t.Fatal(e)
	}
	opaque := adoptionFiles(t, raw)
	adoptionMetadata(t, opaque)
	found := false
	for _, f := range opaque {
		if f.Name == "suite.json" {
			found = bytes.Equal(f.Data, malformed)
		}
	}
	if !found {
		t.Fatal("opaque malformed suite changed")
	}
}
func TestAdoptionRefusesStalePinsMetadataAndDomainExtras(t *testing.T) {
	src := adoptionPinnedSources(t)
	for _, name := range []string{"scenarios/runner.go", "scenarios/report_decode.go"} {
		filtered := []artifactbundle.File{}
		for _, f := range src {
			if f.Name != name {
				filtered = append(filtered, f)
			}
		}
		out, e := adoptionLegacyBuild([]byte(candidate), []byte(suite), filtered)
		if e == nil || out != nil {
			t.Fatal("missing current runner dependency accepted", name)
		}
	}
	for i := range src {
		bad := append([]artifactbundle.File{}, src...)
		bad[i].Data = bytes.Clone(src[i].Data)
		bad[i].Data[0] ^= 1
		out, e := adoptionLegacyBuild([]byte(candidate), []byte(suite), bad)
		if e == nil || out != nil {
			t.Fatal("stale source bytes accepted", bad[i].Name)
		}
	}
	raw, e := adoptionLegacyBuild([]byte(candidate), []byte(suite), src)
	if e != nil {
		t.Fatal(e)
	}
	files := adoptionFiles(t, raw)
	for _, field := range []string{"sourceRevision", "suiteStatus", "modelDigest"} {
		bad := append([]artifactbundle.File{}, files...)
		for i := range bad {
			if bad[i].Name == "qualification.json" {
				var id map[string]any
				if e = json.Unmarshal(bad[i].Data, &id); e != nil {
					t.Fatal(e)
				}
				switch field {
				case "sourceRevision":
					id[field] = "3312a090afac526ddeda392dbd31ed49fa0e04ca"
				case "suiteStatus":
					id[field] = "executed and approved"
				case "modelDigest":
					id[field] = strings.Repeat("0", 64)
				}
				bad[i].Data, e = json.Marshal(id)
				if e != nil {
					t.Fatal(e)
				}
				bad[i].Data = append(bad[i].Data, '\n')
			}
		}
		tampered, e := artifactbundle.Build(artifactbundle.QualificationProfile, bad)
		if e != nil {
			t.Fatal(e)
		}
		if Verify(tampered, adoptionCandidateSHA, adoptionSuiteSHA) == nil {
			t.Fatal("self-consistent archive metadata tamper accepted", field)
		}
	}
	extra := append(append([]artifactbundle.File{}, files...), artifactbundle.File{Name: "extra.txt", Data: []byte("no authority")})
	tampered, e := artifactbundle.Build(artifactbundle.QualificationProfile, extra)
	if e != nil {
		t.Fatal(e)
	}
	if Verify(tampered, adoptionCandidateSHA, adoptionSuiteSHA) == nil {
		t.Fatal("domain extra accepted")
	}
	for _, name := range []string{"scenarios/runner.go", "scenarios/report_decode.go"} {
		missing := []artifactbundle.File{}
		for _, f := range files {
			if f.Name != name {
				missing = append(missing, f)
			}
		}
		b, e := artifactbundle.Build(artifactbundle.QualificationProfile, missing)
		if e != nil {
			t.Fatal(e)
		}
		if Verify(b, adoptionCandidateSHA, adoptionSuiteSHA) == nil {
			t.Fatal("domain missing source accepted", name)
		}
	}
	if Verify(raw, strings.Repeat("0", 64), adoptionSuiteSHA) == nil || Verify(raw, adoptionCandidateSHA, strings.Repeat("0", 64)) == nil {
		t.Fatal("wrong external input identities accepted")
	}
}

func adoptionAssertReportJSON(t *testing.T, raw []byte, wantRaw string) {
	t.Helper()
	var got, want any
	if e := json.Unmarshal(raw, &got); e != nil {
		t.Fatal("exported report is not JSON", e)
	}
	if e := json.Unmarshal([]byte(wantRaw), &want); e != nil {
		t.Fatal("invalid frozen literal", e)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("FROZEN_EXPORTED_REPORT_MISMATCH: complete report differs\ngot %s\nwant %s", raw, wantRaw)
	}
}

// Called by the existing exported-module harness on actual stdout, never by
// exporter logic. Full envelope, tasks, commands, reviews and comparisons matter.
func assertFrozenCandidateReport(t *testing.T, raw []byte) {
	t.Helper()
	adoptionAssertReportJSON(t, raw, adoptionExpectedCandidateReport)
}
func assertFrozenMismatchReport(t *testing.T, raw []byte) {
	t.Helper()
	adoptionAssertReportJSON(t, raw, adoptionExpectedMismatchReport)
}

// Literal projection of preexisting independent suite expectations, not observed runner output.
const adoptionExpectedCandidateReport = `{
  "formatVersion": "lifecycle-scenario-report/v1",
  "scenarioVersion": "lifecycle-scenarios/v1",
  "profileVersion": "lifecycle/v1",
  "mode": "example-only",
  "authority": "none",
  "modelDigest": "a9ad2dfb926a24a026c8565f0d929a5b803bec43f6eb7342e196830791eccc7c",
  "scenarioDigest": "7a50d74f611ddf854504eceaa34d85a42bb5e56777b80ab84332921d167f3975",
  "sourceDigest": "53b5ceec6b68719243853dae5285c13d5c25df89cb203a16279f3cacd2f989f1",
  "matched": true,
  "casesCompleted": 4,
  "stepsExecuted": 6,
  "entries": [
    {
      "caseId": "missing_owner_refuses",
      "stepId": "prepare_refused",
      "input": {
        "id": "task-1",
        "state": "Draft",
        "revision": "0",
        "outcome": "An invented outcome",
        "acceptance": [
          "An invented criterion"
        ],
        "owner": null
      },
      "command": {
        "type": "prepare",
        "actor": "owner-1",
        "expectedRevision": "0",
        "review": null
      },
      "expected": {
        "ok": false,
        "transitionId": "T_prepare",
        "task": {
          "id": "task-1",
          "state": "Draft",
          "revision": "0",
          "outcome": "An invented outcome",
          "acceptance": [
            "An invented criterion"
          ],
          "owner": null
        },
        "refusals": [
          "G_owner"
        ]
      },
      "actual": {
        "mode": "example-only",
        "authority": "none",
        "modelDigest": "a9ad2dfb926a24a026c8565f0d929a5b803bec43f6eb7342e196830791eccc7c",
        "ok": false,
        "transitionId": "T_prepare",
        "task": {
          "id": "task-1",
          "state": "Draft",
          "revision": "0",
          "outcome": "An invented outcome",
          "acceptance": [
            "An invented criterion"
          ],
          "owner": null
        },
        "refusals": [
          "G_owner"
        ]
      },
      "matched": true,
      "differences": []
    },
    {
      "caseId": "named_owner_succeeds",
      "stepId": "prepare_allowed",
      "input": {
        "id": "task-1",
        "state": "Draft",
        "revision": "0",
        "outcome": "An invented outcome",
        "acceptance": [
          "An invented criterion"
        ],
        "owner": "owner-1"
      },
      "command": {
        "type": "prepare",
        "actor": "owner-1",
        "expectedRevision": "0",
        "review": null
      },
      "expected": {
        "ok": true,
        "transitionId": "T_prepare",
        "task": {
          "id": "task-1",
          "state": "Ready",
          "revision": "1",
          "outcome": "An invented outcome",
          "acceptance": [
            "An invented criterion"
          ],
          "owner": "owner-1"
        },
        "refusals": []
      },
      "actual": {
        "mode": "example-only",
        "authority": "none",
        "modelDigest": "a9ad2dfb926a24a026c8565f0d929a5b803bec43f6eb7342e196830791eccc7c",
        "ok": true,
        "transitionId": "T_prepare",
        "task": {
          "id": "task-1",
          "state": "Ready",
          "revision": "1",
          "outcome": "An invented outcome",
          "acceptance": [
            "An invented criterion"
          ],
          "owner": "owner-1"
        },
        "refusals": []
      },
      "matched": true,
      "differences": []
    },
    {
      "caseId": "rework_replay",
      "stepId": "request_changes",
      "input": {
        "id": "review-example",
        "state": "InReview",
        "revision": "3",
        "outcome": "Deliver an example",
        "acceptance": [
          "First criterion"
        ],
        "owner": "worker"
      },
      "command": {
        "type": "requestChanges",
        "actor": "observer",
        "expectedRevision": "3",
        "review": {
          "taskId": "review-example",
          "taskRevision": "3",
          "modelDigest": "a9ad2dfb926a24a026c8565f0d929a5b803bec43f6eb7342e196830791eccc7c",
          "taskDigest": "29b49f551bf9020677e1e306946a40530bd9d40b45f0be31ae81bb89d800b300",
          "reviewer": "observer",
          "decision": "changesRequested"
        }
      },
      "expected": {
        "ok": true,
        "transitionId": "T_requestChanges",
        "task": {
          "id": "review-example",
          "state": "InProgress",
          "revision": "4",
          "outcome": "Deliver an example",
          "acceptance": [
            "First criterion"
          ],
          "owner": "worker"
        },
        "refusals": []
      },
      "actual": {
        "mode": "example-only",
        "authority": "none",
        "modelDigest": "a9ad2dfb926a24a026c8565f0d929a5b803bec43f6eb7342e196830791eccc7c",
        "ok": true,
        "transitionId": "T_requestChanges",
        "task": {
          "id": "review-example",
          "state": "InProgress",
          "revision": "4",
          "outcome": "Deliver an example",
          "acceptance": [
            "First criterion"
          ],
          "owner": "worker"
        },
        "refusals": []
      },
      "matched": true,
      "differences": []
    },
    {
      "caseId": "rework_replay",
      "stepId": "resubmit",
      "input": {
        "id": "review-example",
        "state": "InProgress",
        "revision": "4",
        "outcome": "Deliver an example",
        "acceptance": [
          "First criterion"
        ],
        "owner": "worker"
      },
      "command": {
        "type": "requestReview",
        "actor": "worker",
        "expectedRevision": "4",
        "review": null
      },
      "expected": {
        "ok": true,
        "transitionId": "T_requestReview",
        "task": {
          "id": "review-example",
          "state": "InReview",
          "revision": "5",
          "outcome": "Deliver an example",
          "acceptance": [
            "First criterion"
          ],
          "owner": "worker"
        },
        "refusals": []
      },
      "actual": {
        "mode": "example-only",
        "authority": "none",
        "modelDigest": "a9ad2dfb926a24a026c8565f0d929a5b803bec43f6eb7342e196830791eccc7c",
        "ok": true,
        "transitionId": "T_requestReview",
        "task": {
          "id": "review-example",
          "state": "InReview",
          "revision": "5",
          "outcome": "Deliver an example",
          "acceptance": [
            "First criterion"
          ],
          "owner": "worker"
        },
        "refusals": []
      },
      "matched": true,
      "differences": []
    },
    {
      "caseId": "rework_replay",
      "stepId": "stale_review",
      "input": {
        "id": "review-example",
        "state": "InReview",
        "revision": "5",
        "outcome": "Deliver an example",
        "acceptance": [
          "First criterion"
        ],
        "owner": "worker"
      },
      "command": {
        "type": "approve",
        "actor": "observer",
        "expectedRevision": "5",
        "review": {
          "taskId": "review-example",
          "taskRevision": "3",
          "modelDigest": "a9ad2dfb926a24a026c8565f0d929a5b803bec43f6eb7342e196830791eccc7c",
          "taskDigest": "29b49f551bf9020677e1e306946a40530bd9d40b45f0be31ae81bb89d800b300",
          "reviewer": "observer",
          "decision": "approved"
        }
      },
      "expected": {
        "ok": false,
        "transitionId": "T_approve",
        "task": {
          "id": "review-example",
          "state": "InReview",
          "revision": "5",
          "outcome": "Deliver an example",
          "acceptance": [
            "First criterion"
          ],
          "owner": "worker"
        },
        "refusals": [
          "G_reviewContent",
          "G_reviewRevision"
        ]
      },
      "actual": {
        "mode": "example-only",
        "authority": "none",
        "modelDigest": "a9ad2dfb926a24a026c8565f0d929a5b803bec43f6eb7342e196830791eccc7c",
        "ok": false,
        "transitionId": "T_approve",
        "task": {
          "id": "review-example",
          "state": "InReview",
          "revision": "5",
          "outcome": "Deliver an example",
          "acceptance": [
            "First criterion"
          ],
          "owner": "worker"
        },
        "refusals": [
          "G_reviewContent",
          "G_reviewRevision"
        ]
      },
      "matched": true,
      "differences": []
    },
    {
      "caseId": "content_mismatch",
      "stepId": "approve",
      "input": {
        "id": "review-example",
        "state": "InReview",
        "revision": "3",
        "outcome": "Deliver an example",
        "acceptance": [
          "Revised criterion"
        ],
        "owner": "worker"
      },
      "command": {
        "type": "approve",
        "actor": "observer",
        "expectedRevision": "3",
        "review": {
          "taskId": "review-example",
          "taskRevision": "3",
          "modelDigest": "a9ad2dfb926a24a026c8565f0d929a5b803bec43f6eb7342e196830791eccc7c",
          "taskDigest": "29b49f551bf9020677e1e306946a40530bd9d40b45f0be31ae81bb89d800b300",
          "reviewer": "observer",
          "decision": "approved"
        }
      },
      "expected": {
        "ok": false,
        "transitionId": "T_approve",
        "task": {
          "id": "review-example",
          "state": "InReview",
          "revision": "3",
          "outcome": "Deliver an example",
          "acceptance": [
            "Revised criterion"
          ],
          "owner": "worker"
        },
        "refusals": [
          "G_reviewContent"
        ]
      },
      "actual": {
        "mode": "example-only",
        "authority": "none",
        "modelDigest": "a9ad2dfb926a24a026c8565f0d929a5b803bec43f6eb7342e196830791eccc7c",
        "ok": false,
        "transitionId": "T_approve",
        "task": {
          "id": "review-example",
          "state": "InReview",
          "revision": "3",
          "outcome": "Deliver an example",
          "acceptance": [
            "Revised criterion"
          ],
          "owner": "worker"
        },
        "refusals": [
          "G_reviewContent"
        ]
      },
      "matched": true,
      "differences": []
    }
  ]
}`

// Literal projection of preexisting independent suite expectations, not observed runner output.
const adoptionExpectedMismatchReport = `{
  "formatVersion": "lifecycle-scenario-report/v1",
  "scenarioVersion": "lifecycle-scenarios/v1",
  "profileVersion": "lifecycle/v1",
  "mode": "example-only",
  "authority": "none",
  "modelDigest": "a9ad2dfb926a24a026c8565f0d929a5b803bec43f6eb7342e196830791eccc7c",
  "scenarioDigest": "ca41db0c38b2cce79911fec148bdcbe76951c598bca107a2c91d7afeef630f35",
  "sourceDigest": "47c01272e0c67252d68d3dd71796c14397211577fd66f53e6deba21d828b3760",
  "matched": false,
  "casesCompleted": 0,
  "stepsExecuted": 1,
  "entries": [
    {
      "caseId": "missing_owner_refuses",
      "stepId": "prepare_refused",
      "input": {
        "id": "task-1",
        "state": "Draft",
        "revision": "0",
        "outcome": "An invented outcome",
        "acceptance": [
          "An invented criterion"
        ],
        "owner": null
      },
      "command": {
        "type": "prepare",
        "actor": "owner-1",
        "expectedRevision": "0",
        "review": null
      },
      "expected": {
        "ok": false,
        "transitionId": "T_prepare",
        "task": {
          "id": "task-1",
          "state": "Draft",
          "revision": "0",
          "outcome": "An invented outcome",
          "acceptance": [
            "An invented criterion"
          ],
          "owner": null
        },
        "refusals": [
          "G_actor"
        ]
      },
      "actual": {
        "mode": "example-only",
        "authority": "none",
        "modelDigest": "a9ad2dfb926a24a026c8565f0d929a5b803bec43f6eb7342e196830791eccc7c",
        "ok": false,
        "transitionId": "T_prepare",
        "task": {
          "id": "task-1",
          "state": "Draft",
          "revision": "0",
          "outcome": "An invented outcome",
          "acceptance": [
            "An invented criterion"
          ],
          "owner": null
        },
        "refusals": [
          "G_owner"
        ]
      },
      "matched": false,
      "differences": [
        "refusals"
      ]
    }
  ]
}`

// These original eleven-pin/17-payload expectations now explicitly exercise the
// supported historical policy. Semantic fixtures and full report literals above
// remain byte-for-byte unchanged; current policy is tested in profilecatalog.
func adoptionLegacySourceNames(t *testing.T) []string {
	t.Helper()
	names, err := SourceNamesForRevision(adoptionSourceRevision)
	if err != nil {
		t.Fatal(err)
	}
	return names
}
func adoptionLegacyBuild(candidate, suite []byte, sources []artifactbundle.File) ([]byte, error) {
	return BuildForRevision(adoptionSourceRevision, candidate, suite, sources)
}
