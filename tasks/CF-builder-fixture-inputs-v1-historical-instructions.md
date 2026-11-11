---
id: CF-builder-fixture-inputs-v1
status: in-progress
owner: Foundry test fixture maintenance contributor
repositories: [spik3r/clanker-foundry]
---

# Replay frozen builder instructions as test data

## Implementation claim

Owner: Foundry test fixture maintenance contributor. The integration lead
approved this test-only prerequisite to the separately sourced CI scheduling
change. Base main is `30756603f16f3e4fe9216c40389d6d0df024218a`; the first
reviewed source commit is `dd1cb32d9940dcf382b94e28910ffaa262ece3b3`.
Existing CI scheduling edits remain covered only by their original task/start.
This claim covers no workflow, root instructions or other task edits.

Exact scope:

- `modeling/lifecycle/builderdemo/testdata/AGENTS-882b5002.txt`, one exact
  historical root instruction snapshot stored as data, not active instructions
- `modeling/lifecycle/builderdemo/testdata/README.md`, snapshot provenance and
  the distinction between historical test replay and current-source acceptance
- `modeling/lifecycle/builderdemo/prepare_test.go`, only the instruction fixture
  source choice and focused supplementary tests
- `modeling/lifecycle/cmd/modelbuilderdemo/main_test.go`, only the instruction
  fixture source choice and focused supplementary tests
- This task and its new immutable Project Memory source/event files

Both fixture helpers currently read live root `AGENTS.md` although the fixed
15-artifact manifest pins its original bytes. Updating the current CI schedule
wording therefore breaks historical replay. On the same Go 1.27.4 toolchain,
both affected packages pass from the unchanged source archive and fail with
`baseline drift: AGENTS.md` or a blocked CLI result after the instruction edit.
This is a fixture dependency, separate from other lifecycle checks that require
exact Go 1.27.3. The builder README already says legitimate later input edits may
block the old example and require its reviewed baseline.

Extract the one snapshot from existing Git commit
`882b5002bd4dc70b4be0060e5e5668d1faa8bbdd`, whose root instructions have SHA-256
`4574ee25fe89b6658ae237f3f50648ef1e542c487a2c23d7ab2833696f9455eb`.
Independently verify those bytes against the manifest and unchanged base root
instructions. Redirect only the `AGENTS.md` fixture artifact to this snapshot
in both helpers. Keep all other 14 artifact reads and pins, original manifest
bytes and hash, literal acceptance assertions, production builder/preflight
code, directory-boundary tests, public worked-example sources and historical
evidence unchanged. Never refresh a hash to match changed instructions.

Add supplementary tests demonstrating successful historical replay, exact
instruction bytes copied into the prepared output, and refusal with no output
for a changed instruction input. Exercise actual current root instructions too:
when their bytes differ from the frozen snapshot, preparation must remain
blocked; exact original bytes remain eligible. This is a data equality contract,
not an exemption for current instructions or permission to execute a builder.
Keep existing refusal coverage and add a deterministic mutated-input case that
always refuses, even if future current instructions match the historical source.

No production behavior, manifest admission rule, fixture expectation, workflow
job, toolchain pin, dependency, authority, activation or held work changes.
The 38 frozen scheduling tests remain byte-identical with SHA-256
`f63b4a988543b7474fcb0d521007b3d08f0e5486136a1279dc716375c52a6316`.
Publish only a qualified combined draft after local validation. Fresh independent
exact-source review and hosted pinned-toolchain checks remain integration gates;
Ready and merge remain with the integration lead.

## Source review and validation plan

Review this source separately before registration or implementation. Create its
true source commit with the already verified first source commit as parent and
only this new task added, fetch and verify it, register it separately, then record
a genuine start for these exact paths. Do not backfill prerequisite coverage
into the CI task. Preserve both source commits and every existing event byte.

Copy and verify the historical snapshot, then freeze new test expectations and
reproduce failures with the old fixture helpers before redirecting their input.
Keep supplementary assertions unchanged after that reproduction. Compare the
original manifest, all 14 other input pins, original acceptance assertions and
production paths against base. Run both affected package race suites and vet,
format checks, repository validation, frozen scheduling tests and full committed
Project Memory validation for both tasks. Recheck the existing broader local
results on final relevant bytes. Do not present available Go 1.27.4 validation as
full Go 1.27.3 lifecycle acceptance. The unchanged hosted job must supply that
exact-pin result before integration.
