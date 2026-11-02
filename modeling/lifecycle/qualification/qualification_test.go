package qualification

import (
	"bytes"
	"context"
	"errors"
	"github.com/spik3r/clanker-foundry/modeling/lifecycle/artifactbundle"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

// Frozen before exporter implementation. Independently calculated canonical model
// digest a9ad2dfb926a24a026c8565f0d929a5b803bec43f6eb7342e196830791eccc7c.
const candidate = "{\n  \"formatVersion\": \"lifecycle-json/v1\",\n  \"profileVersion\": \"lifecycle/v1\",\n  \"id\": \"TaskLifecycleExample\",\n  \"initial\": \"Draft\",\n  \"states\": [\n    {\n      \"id\": \"Draft\",\n      \"label\": \"Draft\",\n      \"terminal\": false\n    },\n    {\n      \"id\": \"Ready\",\n      \"label\": \"Ready\",\n      \"terminal\": false\n    },\n    {\n      \"id\": \"InProgress\",\n      \"label\": \"In progress\",\n      \"terminal\": false\n    },\n    {\n      \"id\": \"InReview\",\n      \"label\": \"In review\",\n      \"terminal\": false\n    },\n    {\n      \"id\": \"Done\",\n      \"label\": \"Done\",\n      \"terminal\": true\n    }\n  ],\n  \"requirements\": [\n    {\n      \"id\": \"R_revision\",\n      \"text\": \"Command matches the current revision.\"\n    },\n    {\n      \"id\": \"R_outcome\",\n      \"text\": \"An outcome is nonblank.\"\n    },\n    {\n      \"id\": \"R_acceptance\",\n      \"text\": \"Acceptance criteria are supplied.\"\n    },\n    {\n      \"id\": \"R_actor\",\n      \"text\": \"The named starting actor is nonblank.\"\n    },\n    {\n      \"id\": \"R_ownerCompatible\",\n      \"text\": \"Start assigns only absent ownership or matches the existing owner.\"\n    },\n    {\n      \"id\": \"R_owner\",\n      \"text\": \"An owner is present.\"\n    },\n    {\n      \"id\": \"R_ownerActor\",\n      \"text\": \"The owner requests review.\"\n    },\n    {\n      \"id\": \"R_reviewer\",\n      \"text\": \"The reviewer is nonblank.\"\n    },\n    {\n      \"id\": \"R_reviewTask\",\n      \"text\": \"Review refers to this task.\"\n    },\n    {\n      \"id\": \"R_reviewRevision\",\n      \"text\": \"Review refers to the current revision.\"\n    },\n    {\n      \"id\": \"R_reviewModel\",\n      \"text\": \"Review refers to the accepted model.\"\n    },\n    {\n      \"id\": \"R_reviewContent\",\n      \"text\": \"Review refers to current task content.\"\n    },\n    {\n      \"id\": \"R_reviewActor\",\n      \"text\": \"The acting reviewer matches the review.\"\n    },\n    {\n      \"id\": \"R_independent\",\n      \"text\": \"The reviewer differs from the owner.\"\n    },\n    {\n      \"id\": \"R_approved\",\n      \"text\": \"Approval is explicit.\"\n    },\n    {\n      \"id\": \"R_changesRequested\",\n      \"text\": \"Changes are explicitly requested.\"\n    }\n  ],\n  \"guards\": [\n    {\n      \"id\": \"G_revision\",\n      \"requirement\": \"R_revision\",\n      \"expr\": {\n        \"op\": \"eq\",\n        \"args\": [\n          {\n            \"ref\": \"task.revision\"\n          },\n          {\n            \"ref\": \"command.expectedRevision\"\n          }\n        ]\n      }\n    },\n    {\n      \"id\": \"G_outcome\",\n      \"requirement\": \"R_outcome\",\n      \"expr\": {\n        \"op\": \"nonempty\",\n        \"args\": [\n          {\n            \"ref\": \"task.outcome\"\n          }\n        ]\n      }\n    },\n    {\n      \"id\": \"G_acceptance\",\n      \"requirement\": \"R_acceptance\",\n      \"expr\": {\n        \"op\": \"nonempty\",\n        \"args\": [\n          {\n            \"ref\": \"task.acceptance\"\n          }\n        ]\n      }\n    },\n    {\n      \"id\": \"G_actor\",\n      \"requirement\": \"R_actor\",\n      \"expr\": {\n        \"op\": \"nonempty\",\n        \"args\": [\n          {\n            \"ref\": \"command.actor\"\n          }\n        ]\n      }\n    },\n    {\n      \"id\": \"G_ownerCompatible\",\n      \"requirement\": \"R_ownerCompatible\",\n      \"expr\": {\n        \"op\": \"any\",\n        \"args\": [\n          {\n            \"op\": \"absent\",\n            \"args\": [\n              {\n                \"ref\": \"task.owner\"\n              }\n            ]\n          },\n          {\n            \"op\": \"eq\",\n            \"args\": [\n              {\n                \"ref\": \"task.owner\"\n              },\n              {\n                \"ref\": \"command.actor\"\n              }\n            ]\n          }\n        ]\n      }\n    },\n    {\n      \"id\": \"G_owner\",\n      \"requirement\": \"R_owner\",\n      \"expr\": {\n        \"op\": \"nonempty\",\n        \"args\": [\n          {\n            \"ref\": \"task.owner\"\n          }\n        ]\n      }\n    },\n    {\n      \"id\": \"G_ownerActor\",\n      \"requirement\": \"R_ownerActor\",\n      \"expr\": {\n        \"op\": \"eq\",\n        \"args\": [\n          {\n            \"ref\": \"task.owner\"\n          },\n          {\n            \"ref\": \"command.actor\"\n          }\n        ]\n      }\n    },\n    {\n      \"id\": \"G_reviewer\",\n      \"requirement\": \"R_reviewer\",\n      \"expr\": {\n        \"op\": \"nonempty\",\n        \"args\": [\n          {\n            \"ref\": \"review.reviewer\"\n          }\n        ]\n      }\n    },\n    {\n      \"id\": \"G_reviewTask\",\n      \"requirement\": \"R_reviewTask\",\n      \"expr\": {\n        \"op\": \"eq\",\n        \"args\": [\n          {\n            \"ref\": \"review.taskId\"\n          },\n          {\n            \"ref\": \"task.id\"\n          }\n        ]\n      }\n    },\n    {\n      \"id\": \"G_reviewRevision\",\n      \"requirement\": \"R_reviewRevision\",\n      \"expr\": {\n        \"op\": \"eq\",\n        \"args\": [\n          {\n            \"ref\": \"review.taskRevision\"\n          },\n          {\n            \"ref\": \"task.revision\"\n          }\n        ]\n      }\n    },\n    {\n      \"id\": \"G_reviewModel\",\n      \"requirement\": \"R_reviewModel\",\n      \"expr\": {\n        \"op\": \"eq\",\n        \"args\": [\n          {\n            \"ref\": \"review.modelDigest\"\n          },\n          {\n            \"ref\": \"model.digest\"\n          }\n        ]\n      }\n    },\n    {\n      \"id\": \"G_reviewContent\",\n      \"requirement\": \"R_reviewContent\",\n      \"expr\": {\n        \"op\": \"eq\",\n        \"args\": [\n          {\n            \"ref\": \"review.taskDigest\"\n          },\n          {\n            \"ref\": \"task.digest\"\n          }\n        ]\n      }\n    },\n    {\n      \"id\": \"G_reviewActor\",\n      \"requirement\": \"R_reviewActor\",\n      \"expr\": {\n        \"op\": \"eq\",\n        \"args\": [\n          {\n            \"ref\": \"review.reviewer\"\n          },\n          {\n            \"ref\": \"command.actor\"\n          }\n        ]\n      }\n    },\n    {\n      \"id\": \"G_independent\",\n      \"requirement\": \"R_independent\",\n      \"expr\": {\n        \"op\": \"ne\",\n        \"args\": [\n          {\n            \"ref\": \"review.reviewer\"\n          },\n          {\n            \"ref\": \"task.owner\"\n          }\n        ]\n      }\n    },\n    {\n      \"id\": \"G_approved\",\n      \"requirement\": \"R_approved\",\n      \"expr\": {\n        \"op\": \"eq\",\n        \"args\": [\n          {\n            \"ref\": \"review.decision\"\n          },\n          {\n            \"string\": \"approved\"\n          }\n        ]\n      }\n    },\n    {\n      \"id\": \"G_changesRequested\",\n      \"requirement\": \"R_changesRequested\",\n      \"expr\": {\n        \"op\": \"eq\",\n        \"args\": [\n          {\n            \"ref\": \"review.decision\"\n          },\n          {\n            \"string\": \"changesRequested\"\n          }\n        ]\n      }\n    }\n  ],\n  \"transitions\": [\n    {\n      \"id\": \"T_prepare\",\n      \"from\": \"Draft\",\n      \"trigger\": \"prepare\",\n      \"to\": \"Ready\",\n      \"guards\": [\n        \"G_revision\",\n        \"G_outcome\",\n        \"G_acceptance\",\n        \"G_owner\"\n      ],\n      \"effects\": [\n        {\n          \"op\": \"set_state\"\n        },\n        {\n          \"op\": \"increment_revision\"\n        }\n      ]\n    },\n    {\n      \"id\": \"T_start\",\n      \"from\": \"Ready\",\n      \"trigger\": \"start\",\n      \"to\": \"InProgress\",\n      \"guards\": [\n        \"G_revision\",\n        \"G_outcome\",\n        \"G_acceptance\",\n        \"G_actor\",\n        \"G_ownerCompatible\"\n      ],\n      \"effects\": [\n        {\n          \"op\": \"assign_owner_if_absent\"\n        },\n        {\n          \"op\": \"set_state\"\n        },\n        {\n          \"op\": \"increment_revision\"\n        }\n      ]\n    },\n    {\n      \"id\": \"T_requestReview\",\n      \"from\": \"InProgress\",\n      \"trigger\": \"requestReview\",\n      \"to\": \"InReview\",\n      \"guards\": [\n        \"G_revision\",\n        \"G_outcome\",\n        \"G_acceptance\",\n        \"G_owner\",\n        \"G_ownerActor\"\n      ],\n      \"effects\": [\n        {\n          \"op\": \"set_state\"\n        },\n        {\n          \"op\": \"increment_revision\"\n        }\n      ]\n    },\n    {\n      \"id\": \"T_approve\",\n      \"from\": \"InReview\",\n      \"trigger\": \"approve\",\n      \"to\": \"Done\",\n      \"guards\": [\n        \"G_revision\",\n        \"G_outcome\",\n        \"G_acceptance\",\n        \"G_owner\",\n        \"G_reviewer\",\n        \"G_reviewTask\",\n        \"G_reviewRevision\",\n        \"G_reviewModel\",\n        \"G_reviewContent\",\n        \"G_reviewActor\",\n        \"G_independent\",\n        \"G_approved\"\n      ],\n      \"effects\": [\n        {\n          \"op\": \"set_state\"\n        },\n        {\n          \"op\": \"increment_revision\"\n        }\n      ]\n    },\n    {\n      \"id\": \"T_requestChanges\",\n      \"from\": \"InReview\",\n      \"trigger\": \"requestChanges\",\n      \"to\": \"InProgress\",\n      \"guards\": [\n        \"G_revision\",\n        \"G_outcome\",\n        \"G_acceptance\",\n        \"G_owner\",\n        \"G_reviewer\",\n        \"G_reviewTask\",\n        \"G_reviewRevision\",\n        \"G_reviewModel\",\n        \"G_reviewContent\",\n        \"G_reviewActor\",\n        \"G_independent\",\n        \"G_changesRequested\"\n      ],\n      \"effects\": [\n        {\n          \"op\": \"set_state\"\n        },\n        {\n          \"op\": \"increment_revision\"\n        }\n      ]\n    }\n  ]\n}\n"

// Corrected fixture freeze: only duplicate step IDs were renamed after preflight
// review; all task, command, review, and expected values remain unchanged.
// Earlier suite SHA-256 ffdacd35a797d001edc9cacc502ae119c448f921242f287acc73b93fb3a47ba9;
// corrected suite SHA-256 53b5ceec6b68719243853dae5285c13d5c25df89cb203a16279f3cacd2f989f1.
const suite = "{\n  \"formatVersion\": \"lifecycle-scenarios-json/v1\",\n  \"scenarioVersion\": \"lifecycle-scenarios/v1\",\n  \"profileVersion\": \"lifecycle/v1\",\n  \"modelDigest\": \"a9ad2dfb926a24a026c8565f0d929a5b803bec43f6eb7342e196830791eccc7c\",\n  \"id\": \"qualification_candidate_v1\",\n  \"cases\": [\n    {\n      \"id\": \"missing_owner_refuses\",\n      \"initial\": {\n        \"id\": \"task-1\",\n        \"state\": \"Draft\",\n        \"revision\": \"0\",\n        \"outcome\": \"An invented outcome\",\n        \"acceptance\": [\n          \"An invented criterion\"\n        ],\n        \"owner\": null\n      },\n      \"steps\": [\n        {\n          \"id\": \"prepare_refused\",\n          \"command\": {\n            \"type\": \"prepare\",\n            \"actor\": \"owner-1\",\n            \"expectedRevision\": \"0\",\n            \"review\": null\n          },\n          \"expected\": {\n            \"ok\": false,\n            \"transitionId\": \"T_prepare\",\n            \"task\": {\n              \"id\": \"task-1\",\n              \"state\": \"Draft\",\n              \"revision\": \"0\",\n              \"outcome\": \"An invented outcome\",\n              \"acceptance\": [\n                \"An invented criterion\"\n              ],\n              \"owner\": null\n            },\n            \"refusals\": [\n              \"G_owner\"\n            ]\n          }\n        }\n      ]\n    },\n    {\n      \"id\": \"named_owner_succeeds\",\n      \"initial\": {\n        \"id\": \"task-1\",\n        \"state\": \"Draft\",\n        \"revision\": \"0\",\n        \"outcome\": \"An invented outcome\",\n        \"acceptance\": [\n          \"An invented criterion\"\n        ],\n        \"owner\": \"owner-1\"\n      },\n      \"steps\": [\n        {\n          \"id\": \"prepare_allowed\",\n          \"command\": {\n            \"type\": \"prepare\",\n            \"actor\": \"owner-1\",\n            \"expectedRevision\": \"0\",\n            \"review\": null\n          },\n          \"expected\": {\n            \"ok\": true,\n            \"transitionId\": \"T_prepare\",\n            \"task\": {\n              \"id\": \"task-1\",\n              \"state\": \"Ready\",\n              \"revision\": \"1\",\n              \"outcome\": \"An invented outcome\",\n              \"acceptance\": [\n                \"An invented criterion\"\n              ],\n              \"owner\": \"owner-1\"\n            },\n            \"refusals\": []\n          }\n        }\n      ]\n    },\n    {\n      \"id\": \"rework_replay\",\n      \"initial\": {\n        \"id\": \"review-example\",\n        \"state\": \"InReview\",\n        \"revision\": \"3\",\n        \"outcome\": \"Deliver an example\",\n        \"acceptance\": [\n          \"First criterion\"\n        ],\n        \"owner\": \"worker\"\n      },\n      \"steps\": [\n        {\n          \"id\": \"request_changes\",\n          \"command\": {\n            \"type\": \"requestChanges\",\n            \"actor\": \"observer\",\n            \"expectedRevision\": \"3\",\n            \"review\": {\n              \"taskId\": \"review-example\",\n              \"taskRevision\": \"3\",\n              \"modelDigest\": \"a9ad2dfb926a24a026c8565f0d929a5b803bec43f6eb7342e196830791eccc7c\",\n              \"taskDigest\": \"29b49f551bf9020677e1e306946a40530bd9d40b45f0be31ae81bb89d800b300\",\n              \"reviewer\": \"observer\",\n              \"decision\": \"changesRequested\"\n            }\n          },\n          \"expected\": {\n            \"ok\": true,\n            \"transitionId\": \"T_requestChanges\",\n            \"task\": {\n              \"id\": \"review-example\",\n              \"state\": \"InProgress\",\n              \"revision\": \"4\",\n              \"outcome\": \"Deliver an example\",\n              \"acceptance\": [\n                \"First criterion\"\n              ],\n              \"owner\": \"worker\"\n            },\n            \"refusals\": []\n          }\n        },\n        {\n          \"id\": \"resubmit\",\n          \"command\": {\n            \"type\": \"requestReview\",\n            \"actor\": \"worker\",\n            \"expectedRevision\": \"4\",\n            \"review\": null\n          },\n          \"expected\": {\n            \"ok\": true,\n            \"transitionId\": \"T_requestReview\",\n            \"task\": {\n              \"id\": \"review-example\",\n              \"state\": \"InReview\",\n              \"revision\": \"5\",\n              \"outcome\": \"Deliver an example\",\n              \"acceptance\": [\n                \"First criterion\"\n              ],\n              \"owner\": \"worker\"\n            },\n            \"refusals\": []\n          }\n        },\n        {\n          \"id\": \"stale_review\",\n          \"command\": {\n            \"type\": \"approve\",\n            \"actor\": \"observer\",\n            \"expectedRevision\": \"5\",\n            \"review\": {\n              \"taskId\": \"review-example\",\n              \"taskRevision\": \"3\",\n              \"modelDigest\": \"a9ad2dfb926a24a026c8565f0d929a5b803bec43f6eb7342e196830791eccc7c\",\n              \"taskDigest\": \"29b49f551bf9020677e1e306946a40530bd9d40b45f0be31ae81bb89d800b300\",\n              \"reviewer\": \"observer\",\n              \"decision\": \"approved\"\n            }\n          },\n          \"expected\": {\n            \"ok\": false,\n            \"transitionId\": \"T_approve\",\n            \"task\": {\n              \"id\": \"review-example\",\n              \"state\": \"InReview\",\n              \"revision\": \"5\",\n              \"outcome\": \"Deliver an example\",\n              \"acceptance\": [\n                \"First criterion\"\n              ],\n              \"owner\": \"worker\"\n            },\n            \"refusals\": [\n              \"G_reviewContent\",\n              \"G_reviewRevision\"\n            ]\n          }\n        }\n      ]\n    },\n    {\n      \"id\": \"content_mismatch\",\n      \"initial\": {\n        \"id\": \"review-example\",\n        \"state\": \"InReview\",\n        \"revision\": \"3\",\n        \"outcome\": \"Deliver an example\",\n        \"acceptance\": [\n          \"Revised criterion\"\n        ],\n        \"owner\": \"worker\"\n      },\n      \"steps\": [\n        {\n          \"id\": \"approve\",\n          \"command\": {\n            \"type\": \"approve\",\n            \"actor\": \"observer\",\n            \"expectedRevision\": \"3\",\n            \"review\": {\n              \"taskId\": \"review-example\",\n              \"taskRevision\": \"3\",\n              \"modelDigest\": \"a9ad2dfb926a24a026c8565f0d929a5b803bec43f6eb7342e196830791eccc7c\",\n              \"taskDigest\": \"29b49f551bf9020677e1e306946a40530bd9d40b45f0be31ae81bb89d800b300\",\n              \"reviewer\": \"observer\",\n              \"decision\": \"approved\"\n            }\n          },\n          \"expected\": {\n            \"ok\": false,\n            \"transitionId\": \"T_approve\",\n            \"task\": {\n              \"id\": \"review-example\",\n              \"state\": \"InReview\",\n              \"revision\": \"3\",\n              \"outcome\": \"Deliver an example\",\n              \"acceptance\": [\n                \"Revised criterion\"\n              ],\n              \"owner\": \"worker\"\n            },\n            \"refusals\": [\n              \"G_reviewContent\"\n            ]\n          }\n        }\n      ]\n    }\n  ]\n}\n"

func sources(t *testing.T) []artifactbundle.File {
	t.Helper()
	var out []artifactbundle.File
	for _, name := range SourceNames() {
		b, e := os.ReadFile(filepath.Join("..", name))
		if e != nil {
			t.Fatal(e)
		}
		out = append(out, artifactbundle.File{Name: name, Data: b})
	}
	return out
}
func TestReproducibilityAndTamper(t *testing.T) {
	src := sources(t)
	a, e := Build([]byte(candidate), []byte(suite), src)
	if e != nil {
		t.Fatal(e)
	}
	b, e := Build([]byte(candidate), []byte(suite), src)
	if e != nil || !bytes.Equal(a, b) {
		t.Fatal("nondeterministic", e)
	}
	if e = Verify(a, hash([]byte(candidate)), hash([]byte(suite))); e != nil {
		t.Fatal(e)
	}
	for i := range src {
		copyFiles := append([]artifactbundle.File(nil), src...)
		copyFiles[i].Data = append([]byte(nil), src[i].Data...)
		copyFiles[i].Data[0] ^= 1
		if _, e = Build([]byte(candidate), []byte(suite), copyFiles); e == nil {
			t.Fatalf("accepted source tamper %s", src[i].Name)
		}
	}
	files, e := artifactbundle.Verify(a, artifactbundle.QualificationProfile)
	if e != nil {
		t.Fatal(e)
	}
	for i := range files {
		changed := append([]artifactbundle.File(nil), files...)
		changed[i].Data = append([]byte(nil), files[i].Data...)
		changed[i].Data = append(changed[i].Data, 'x')
		raw, e := artifactbundle.Build(artifactbundle.QualificationProfile, changed)
		if e != nil {
			t.Fatal(e)
		}
		if Verify(raw, hash([]byte(candidate)), hash([]byte(suite))) == nil {
			t.Fatalf("accepted tamper %s", files[i].Name)
		}
	}
	if _, e = Build([]byte(candidate), []byte(suite), src[:len(src)-1]); e == nil {
		t.Fatal("missing source")
	}
	if _, e = Build([]byte(candidate), []byte(suite), append(src, artifactbundle.File{Name: "extra", Data: []byte("x")})); e == nil {
		t.Fatal("extra source")
	}
	if _, e = Build([]byte("{"), []byte(suite), src); e == nil {
		t.Fatal("malformed model")
	}

	for _, tc := range []struct{ name, sourceName string }{
		{"oversized name", strings.Repeat("x", maxSourceName+1)},
		{"unknown short name", "not-an-allowed-path"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			modified := append([]artifactbundle.File(nil), src...)
			modified[0].Name = tc.sourceName
			got, err := Build([]byte(candidate), []byte(suite), modified)
			if err == nil || got != nil {
				t.Fatal("invalid source name returned archive")
			}
		})
	}

	// The suite remains literal and unvalidated until the exported runner is invoked.
	invalidSuite, e := Build([]byte(candidate), []byte("{"), src)
	if e != nil || Verify(invalidSuite, hash([]byte(candidate)), hash([]byte("{"))) != nil {
		t.Fatal("bounded invalid suite should remain unvalidated", e)
	}
	if _, e = Build([]byte(candidate), bytes.Repeat([]byte("x"), maxInput+1), src); e == nil {
		t.Fatal("oversized suite accepted")
	}
	missing := append([]artifactbundle.File(nil), files[1:]...)
	missingRaw, e := artifactbundle.Build(artifactbundle.QualificationProfile, missing)
	if e != nil || Verify(missingRaw, hash([]byte(candidate)), hash([]byte(suite))) == nil {
		t.Fatal("missing project artifact accepted")
	}
	extra := append(append([]artifactbundle.File(nil), files...), artifactbundle.File{Name: "extra.txt", Data: []byte("x")})
	extraRaw, e := artifactbundle.Build(artifactbundle.QualificationProfile, extra)
	if e != nil || Verify(extraRaw, hash([]byte(candidate)), hash([]byte(suite))) == nil {
		t.Fatal("extra project artifact accepted")
	}
	if Verify(a, strings.Repeat("0", 64), hash([]byte(suite))) == nil {
		t.Fatal("wrong expected candidate hash accepted")
	}
	if Verify(a, hash([]byte(candidate)), strings.Repeat("0", 64)) == nil {
		t.Fatal("wrong expected suite hash accepted")
	}
}

// qualificationOutput caps only this trusted fixed-fixture test's observations.
// It is not a resource sandbox for caller-provided programs.
type qualificationOutput struct {
	bytes.Buffer
	overflow bool
}

func (w *qualificationOutput) Write(p []byte) (int, error) {
	if len(p) > (1<<20)-w.Len() {
		w.overflow = true
		return 0, errors.New("qualification test output limit")
	}
	return w.Buffer.Write(p)
}

func TestCandidateProjectRun(t *testing.T) {
	if runtime.Version() != compilerToolchain() {
		t.Skip("requires actual Go 1.27.3")
	}
	raw, e := Build([]byte(candidate), []byte(suite), sources(t))
	if e != nil {
		t.Fatal(e)
	}
	files, e := artifactbundle.Verify(raw, artifactbundle.QualificationProfile)
	if e != nil {
		t.Fatal(e)
	}
	parent := t.TempDir()
	// Keep the existing ambient-VCS regression: no exported command may need
	// the malformed unrelated ancestor's Git metadata.
	for _, name := range []string{".git/objects", ".git/refs", "project", "home", "cache", "tmp", "gopath"} {
		if e = os.MkdirAll(filepath.Join(parent, name), 0700); e != nil {
			t.Fatal(e)
		}
	}
	if e = os.WriteFile(filepath.Join(parent, ".git", "HEAD"), []byte("ref: refs/heads/main\n"), 0600); e != nil {
		t.Fatal(e)
	}
	if e = os.WriteFile(filepath.Join(parent, ".git", "config"), []byte("[broken\n"), 0600); e != nil {
		t.Fatal(e)
	}
	dir := filepath.Join(parent, "project")
	goBinary := filepath.Join(runtime.GOROOT(), "bin", "go")
	if !filepath.IsAbs(goBinary) {
		t.Fatal("Go executable must be absolute")
	}
	env := []string{
		"PATH=" + filepath.Dir(goBinary), "HOME=" + filepath.Join(parent, "home"),
		"GOCACHE=" + filepath.Join(parent, "cache"), "TMPDIR=" + filepath.Join(parent, "tmp"), "GOPATH=" + filepath.Join(parent, "gopath"),
		"GOTOOLCHAIN=local", "GOWORK=off", "GOENV=off", "GOFLAGS=", "GOPROXY=off", "GOSUMDB=off", "CGO_ENABLED=0", "LC_ALL=C", "GIT_CONFIG_NOSYSTEM=1",
	}
	ctx, cancel := context.WithTimeout(context.Background(), 180*time.Second)
	defer cancel()
	runProgram := func(t *testing.T, executable string, args []string, input string) (int, []byte, []byte) {
		t.Helper()
		phase, stop := context.WithTimeout(ctx, 90*time.Second)
		defer stop()
		cmd := exec.CommandContext(phase, executable, args...)
		cmd.Dir, cmd.Env, cmd.Stdin = dir, env, strings.NewReader(input)
		cmd.WaitDelay = time.Second
		var out, diag qualificationOutput
		cmd.Stdout, cmd.Stderr = &out, &diag
		err := cmd.Run()
		if phase.Err() != nil || out.overflow || diag.overflow {
			t.Fatalf("test process deadline/output limit: %v", phase.Err())
		}
		if err == nil {
			return 0, out.Bytes(), diag.Bytes()
		}
		var exit *exec.ExitError
		if errors.As(err, &exit) {
			return exit.ExitCode(), out.Bytes(), diag.Bytes()
		}
		t.Fatalf("test process failed to start/wait: %v", err)
		return -1, nil, nil
	}
	gitBinary, e := exec.LookPath("git")
	if e != nil {
		t.Fatal(e)
	}
	status, _, diagnostic := runProgram(t, gitBinary, []string{"status", "--short"}, "")
	if status == 0 || !bytes.Contains(diagnostic, []byte("bad config")) {
		t.Fatal("expected constructed ancestor config to be invalid")
	}
	for _, f := range files {
		p := filepath.Join(dir, filepath.FromSlash(f.Name))
		if e = os.MkdirAll(filepath.Dir(p), 0700); e != nil {
			t.Fatal(e)
		}
		if e = os.WriteFile(p, f.Data, 0600); e != nil {
			t.Fatal(e)
		}
	}
	status, version, _ := runProgram(t, goBinary, []string{"env", "GOVERSION"}, "")
	if status != 0 || strings.TrimSpace(string(version)) != compilerToolchain() {
		t.Fatal("wrong Go executable")
	}
	if status, _, _ = runProgram(t, goBinary, []string{"test", "-buildvcs=false", "./generated"}, ""); status != 0 {
		t.Fatalf("generated structural tests exit %d", status)
	}
	binary := filepath.Join(parent, "modelplay-test-bin")
	build := func(t *testing.T) {
		t.Helper()
		if code, _, _ := runProgram(t, goBinary, []string{"build", "-buildvcs=false", "-o", binary, "./cmd/modelplay"}, ""); code != 0 {
			t.Fatalf("candidate build exit %d", code)
		}
	}
	run := func(t *testing.T, input string) (int, []byte) {
		t.Helper()
		code, out, _ := runProgram(t, binary, nil, input)
		return code, out
	}
	build(t)
	code, output := run(t, suite)
	if code != 0 {
		t.Fatalf("candidate suite exit %d", code)
	}
	assertFrozenCandidateReport(t, output)
	wrongBinding := strings.Replace(suite, "a9ad2dfb926a24a026c8565f0d929a5b803bec43f6eb7342e196830791eccc7c", "b9ad2dfb926a24a026c8565f0d929a5b803bec43f6eb7342e196830791eccc7c", 1)
	if code, out := run(t, wrongBinding); code != 2 || len(out) != 0 {
		t.Fatalf("wrong binding exit/output %d/%d", code, len(out))
	}
	wrongExpected := strings.Replace(suite, "G_owner", "G_actor", 1)
	code, output = run(t, wrongExpected)
	if code != 1 {
		t.Fatalf("wrong expectation exit %d", code)
	}
	assertFrozenMismatchReport(t, output)
	if code, out := run(t, "{"); code != 2 || len(out) != 0 {
		t.Fatalf("malformed suite exit/output %d/%d", code, len(out))
	}
	codePath := filepath.Join(dir, "generated", "lifecycle.go")
	original, e := os.ReadFile(codePath)
	if e != nil {
		t.Fatal(e)
	}
	mutations := []struct{ name, old, new string }{
		{"guard omitted", `return nonempty(refString("task.owner", task, command, currentDigest))`, `return true`},
		{"always refuse", `return Result{Mode: "example-only", Authority: "none", ModelDigest: ModelDigest, OK: true, TransitionID: id, Task: &next, Refusals: []string{}}`, `return refused(task, id, []string{"MUTANT"})`},
	}
	for _, m := range mutations {
		t.Run(m.name, func(t *testing.T) {
			if bytes.Count(original, []byte(m.old)) != 1 {
				t.Fatal("mutation target not unique")
			}
			mutated := bytes.Replace(original, []byte(m.old), []byte(m.new), 1)
			if e := os.WriteFile(codePath, mutated, 0600); e != nil {
				t.Fatal(e)
			}
			build(t)
			if code, _ := run(t, suite); code != 1 {
				t.Fatalf("mutant exit %d", code)
			}
		})
	}
	if e = os.WriteFile(codePath, original, 0600); e != nil {
		t.Fatal(e)
	}
	build(t)
	code, output = run(t, suite)
	if code != 0 {
		t.Fatalf("restored candidate suite exit %d", code)
	}
	assertFrozenCandidateReport(t, output)
	for _, f := range files {
		b, e := os.ReadFile(filepath.Join(dir, filepath.FromSlash(f.Name)))
		if e != nil || !bytes.Equal(b, f.Data) {
			t.Fatalf("exported payload changed: %s", f.Name)
		}
	}
}
func compilerToolchain() string { return "go1.27.3" }

func TestRejectWrongToolchain(t *testing.T) {
	if runtime.Version() == compilerToolchain() {
		t.Skip("wrong-toolchain case requires a different Go version")
	}
	if _, err := Build([]byte(candidate), []byte(suite), sources(t)); err == nil {
		t.Fatal("accepted unsupported generation toolchain")
	}
}
