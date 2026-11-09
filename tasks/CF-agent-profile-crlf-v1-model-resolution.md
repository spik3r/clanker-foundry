---
id: CF-agent-profile-crlf-v1
status: done
owner: astra-foundry-profile-crlf
repositories: [spik3r/clanker-foundry]
---

# Resolve shipped profile tiers with CRLF Markdown

## Implementation claim

Owner: `astra-foundry-profile-crlf`, accepting the coordinator's bounded
assignment for this new outcome. Integration responsibility stays with the
Foundry coordinator. This active-form record is a proposal until independently
reviewed and published through the shared canonical source. Its scratch copy
does not reserve the ID or branch, register a source, or start implementation.

Intended branch: `astra/agent-profile-crlf-v1`, from verified main
`c08604c5f9eebd412af02e95fb06d7846a9f51e2`, tree
`d45d6adae50015117d38bb221d77cfa879847b67`. Recheck shared main, canonical IDs,
branch availability and live overlapping work immediately before publication.
If the base, ownership or scope changes, stop for reconciliation and review the
changed source; do not silently rebase the reviewed claim.

Outcome: the existing resolver returns the same declared alias when a shipped
agent profile uses LF or CRLF Markdown. Preserve its current interface,
backtick-quoted tier grammar and all model mapping behavior. Correct only the
profile record's terminal-CR representation before the existing tier matcher.
Do not introduce a general YAML parser or another quoting policy.

### Exact write scope

The only product paths writable after accepted source publication, actual-pin
verification, registration and a genuine start are:

- `scripts/agent-model.sh`, mode `100755`; base blob
  `fbe0220287efaaefc6dda65b4a778c0c2bfb3c86`.
- `tests/test-agent-model.sh`, mode `100755`; base blob
  `d3c7828b9363051477c8cbb26c14ba766e5c7401`.

The only additional maintained paths are this canonical task,
`tasks/CF-agent-profile-crlf-v1-model-resolution.md`, mode `100644`, and new
immutable source snapshots and genuine events for this task under
`.project-memory/tasks/CF-agent-profile-crlf-v1/`. Preserve the executable modes
of both product files. Record any later task status/result changes under genuine
fresh lifecycle coverage. No prewritten event IDs, timestamps or completions.

No other maintained path is included. Do not edit `scripts/validate.sh`,
`scripts/stow.sh`, `scripts/project-setup.sh`, installer tests, guides, profiles,
model maps, defaults, global instructions, runtime, modeling, workflows,
installed files or user configuration. Preserve every old task, source snapshot,
event and legacy ledger byte and mode. Do not copy local-only evidence from any
previous lane into new shared history or backfill execution evidence.

Draft PR94 and its producer/model work remain excluded. This task does not
authorize changing PR94. The eleven-path adoption assigned to
`astra-foundry-reliability-integration` remains separate; it excludes both product
paths above and this outcome. Its installer entrypoint change preserves the
existing resolver test invocation. Recheck any newer shared claim before editing.
Do not repackage installer, literal-Markdown-path, project-metadata CRLF,
global-import or space-path work.

### Ownership and provenance

The prior `CF-agent-model-duplicate-keys` canonical record is now `done` and
explicitly releases only the completed duplicate/malformed/NUL mapping outcome.
[PR95](https://github.com/spik3r/clanker-foundry/pull/95) merged normally as
`c08604c5f9eebd412af02e95fb06d7846a9f51e2`, retaining source
`d0a019051cd4d1225fac9c93fee41240fe2532ed` and all previous evidence in ancestry.
The closeout did not implement, assign or reserve this profile-CRLF task. This
new claim neither edits nor inherits authority from the old record.

The unchanged unadopted proposal for this ID had SHA-256
`54d28610bb8eb63bdc6186763de4aac11eb538774486f41f9a9469138e30d523`.
This claim keeps its two-file outcome and safeguards, updates the verified base
and released prerequisite, records receiving-owner acceptance, and specifies the
actual-source publication sequence. The old proposal and its assessment remain
unchanged historical evidence, not canonical active claims or new test results.

Fresh shared-state inspection found no matching canonical ID or branch and no
visible competing CRLF owner. Only draft PR94 was open and its changed paths were
disjoint. The receiving owner of the separate eleven-path adoption explicitly
confirmed the disjoint outcome and paths. These observations are not an atomic
reservation and cannot exclude unseen private work. Resolve any newly discovered
conflict through the coordinator before publication or edits.

### Reproduced defect and fixed inputs

The retained assessment ran at earlier main
`8b9d9617a24d0666aba7231068d74e0d0c6d6b55`, tree
`d380cf2e7f95ef0b7432e382222336c75b9732b3`. Both product blobs, the shared map
and all four shipped profiles are unchanged at this claim's base. That evidence
is a prior reproduction, not execution against this new merge commit.

The normal invocation `scripts/agent-model.sh claude explorer-researcher`
returned `haiku` for the shipped LF profile and exited 1 with empty stdout and a
missing-tier diagnostic for CRLF. The documented Codex builder invocation failed
instead of returning `gpt-5.6-sol`. All eight client/profile combinations passed
with LF, failed with CRLF, and passed again in unchanged LF controls. The prior
maintained suite passed on LF; with CRLF it failed exactly its eight shipped
profile/default checks while its invented LF fixture checks did not fail.

The prior proof used separate clean disposable Git checkouts at one exact commit.
Before the CRLF checkout, invented `.git/info/attributes` specified
`agents/*.md text eol=crlf`. The resolver and map remained LF and byte-identical,
no local model override existed, and both worktrees had empty Git status. Raw
byte comparisons proved each CRLF profile was exactly the corresponding LF
profile with LF expanded to CRLF. Preserve that original proof; obtain new
evidence for the corrected resolver only after the claim gates below.

Awk removes the LF record separator but leaves CR. The existing tier matcher
anchors the closing backtick at the absolute record end and therefore fails.
Handling one terminal CR in the Markdown record must not normalize model-map
bytes. Mapping-file CRLF rejection is intentional and remains required.

### Acceptance and regression plan

1. Exercise each actual shipped profile in both LF and CRLF form with both
   `claude` and `codex`. Every success must exit 0, write only the exact alias
   plus newline to stdout, and leave stderr empty. Expected pairs are:
   `explorer-researcher`: `haiku` / `gpt-5.6-terra`;
   `architect-planner`: `sonnet` / `gpt-5.6-terra`;
   `reviewer-verifier`: `sonnet` / `gpt-5.6-terra`;
   `builder-developer`: `opus` / `gpt-5.6-sol`.
   Derive fixture representations from the real tracked declarations, not
   substitute declarations invented to pass. Do not modify tracked profiles.
2. Run the new maintained regression against a retained exact-base resolver and
   show the expected CRLF failures, then against the correction and show both
   representations passing. Keep baseline and candidate paths, hashes, commands,
   exit codes and raw output streams. Do not call a red control a product failure
   or hide an unrelated failure among the expected negatives.
3. Repeat the normal eight invocations in separate clean Git-built LF and CRLF
   checkouts at one corrected implementation identity. Set the local attributes
   before checkout. Assert identical HEAD/tree, clean status, no local override,
   byte-identical resolver/map and exact one-way LF-to-CRLF profile comparisons.
   Retain unchanged LF controls after the CRLF runs. This proves representation
   handling without changing a declaration or relying on a dirty checkout.
4. Preserve every existing regression and its asserted outcome: duplicate keys
   in either map and in either order, unrelated/identical duplicates, malformed
   and empty maps, NUL in every tested position, validation before output,
   data-only parsing, unknown/missing selections, local overrides and shared
   fallback. Keep CRLF mapping-file rejection, including with a valid CRLF
   profile; do not loosen its assertion or normalize its input. Check ordinary
   LF local override and partial-local fallback with CRLF profiles too.
5. Keep the exact profile grammar. Add bounded negative controls for unquoted,
   single/double-quoted, unclosed or extra-backtick declarations and suffix text,
   including trailing spaces or two terminal CR bytes. They must not newly
   resolve an alias. Do not globally strip CR, trim whitespace, accept another
   prefix or change tier matching to make the positive tests pass.
6. All execution uses invented HOME, XDG_CONFIG_HOME, CODEX_HOME and TMPDIR
   directories, isolated Git configuration and disposable fixture repositories.
   Never read or alter real user configuration, installed profiles or local
   overrides, launch a client/agent, select a real user's model, or call a model
   provider. No network request is needed for resolver behavior tests.

### Available validation and limits

After the genuine start and final product edits, run the maintained resolver
suite, red/green and clean-checkout controls above, and `bash -n` on both changed
shell files. Run `bash scripts/validate.sh` on the ordinary LF source after each
maintained change. Before committing, run `git diff --check` and
`git diff --cached --check`; inspect the complete named-path staged diff.

Use Node 24 and Git 2.43 or later for Project Memory. After the genuine completion
and commit, run `node skills/project-memory/scripts/project-memory.mjs validate
--base BASE --repository spik3r/clanker-foundry --source-repository
spik3r/clanker-foundry`, replacing `BASE` with the verified actual PR base commit.
Do not guess a fallback base or use a stale remote-tracking name. Inspect status
and prove all pre-existing metadata, tasks and out-of-scope products are preserved.
Recheck validation against a changed integration base when necessary.

The current cloud tools are Node 24.22.1, Git 2.52.0, Bash 5.2.37 and mawk 1.3.4
on Linux. GNU Stow, ShellCheck and skills-ref are unavailable. Do not install them
or substitute a fake executable. Run additional already available relevant
checks within this scope; record unavailable checks explicitly. Local focused
checks are not full installer acceptance. Existing hosted workflows may qualify
the exact published source under the coordinator's later publication assignment;
their current workflow/tool installation is not permission for local installation.

Run and inspect all applicable existing hosted checks on the actual published
head and current-base PR checkout before integration. Verify checkout identity
and relevant step results, including resolver execution through the existing
installer entrypoint and committed lifecycle validation. Do not equate a workflow
label, an old head's green result, or source-only CI with new-code acceptance.
Keep new-main checks separate from premerge results.

This bounded change does not promise native Windows/macOS/Bash 3.2 behavior,
whole-pack CRLF validation, client loading or provider behavior. Prior resolver
native evidence covers its older unchanged mapping safeguards, not this new
profile correction. Report native/platform checks as not run unless fresh,
authorized exact-source evidence actually exists. Any additional platform
qualification needs its own agreed scope and execution environment.

### Publication, lifecycle and integration order

1. Obtain independent acceptance of these exact source bytes and the ownership
   conclusion. Refresh the shared base, IDs, branches and overlapping work;
   reconcile any change before publication. First publish only the reviewed
   canonical task on the intended branch from the verified base through the
   authorized route. Shared publication, not this private file, reserves the ID.
   Recheck for a conflicting shared claim immediately afterward.
2. Read back the actual server commit, parent, tree, exact task blob/mode/bytes
   and all changed paths. Have that real canonical source independently reviewed
   before any product edit. Publication can assign a different commit identity
   from a local preparation; never register the local or guessed SHA as if it
   were published. Materialize and verify the actual Git objects on the cloud
   machine, preserving actual source parentage and chronology.
3. Register that actual reviewed source using the same-repository identity and
   `tasks/CF-agent-profile-crlf-v1-model-resolution.md`. Independently verify the
   printed revision, source snapshot and complete `Implementation claim` digest.
   Inspect status. Record a genuine start on `astra/agent-profile-crlf-v1` with
   exactly the two product paths and this task path before implementation edits.
   Source/event metadata is created only by the helper for this task.
4. Make the bounded two-file change, run the required available checks, and stage
   only the intended final product/task paths. Record a genuine completion with
   actual results, limits and final staged blob/mode fingerprints. Commit the new
   immutable source/events with the implementation. Validate the committed diff
   against the actual base and independently review the complete exact diff and
   evidence before further publication. Do not rewrite old records to create
   coverage. Any post-completion edit needs a new genuine run and checks.
5. Publish later implementation work only under the coordinator's next assignment.
   Open any eventual PR as draft. Qualify the exact remote head/current-base
   integration, then let the coordinator decide any Ready/merge transition under
   current authorization. No workflow weakening, force update, bypass,
   deployment, runtime activation or unrelated PR action is included.
6. Same-repository source validation requires the actual published source pin to
   remain in history. Use a normal history-preserving merge, never squash/rebase
   away the pin. Verify its ancestry before and after integration. Record actual
   integration evidence and release only this completed scope under fresh
   lifecycle coverage. Preserve all earlier source/events and legacy history.

Initial preparation ends with the proposed source, exact diff, hashes, ownership
evidence and test plan for independent review. It does not itself authorize
source publication, shared reservation, registration, start, product edits or
PR transitions. Later execution and integration require the coordinator's
explicit assignment; this canonical text does not grant authority.

## Local implementation result

The two-file correction is ready for independent implementation review, not
integrated or published as a product change. It uses the actual published task
source `e26d8c0452a41f1b8cd41fc2554d641d53561e90`, based on main
`c08604c5f9eebd412af02e95fb06d7846a9f51e2`. Registration and a genuine start
preceded implementation edits.

The resolver removes one terminal CR from the copied Markdown tier record before
the existing matcher. The mapping reader and remainder of the resolver are
unchanged. The maintained test adds both representations of the four actual
shipped profiles, exact success-output checks, malformed tier controls and CRLF
profile override/fallback/map-refusal checks. All previous mapping safeguard
assertions and tracked-default checks remain intact.

Observed on Linux with Bash 5.2.37, mawk 1.3.4, Git 2.52.0 and Node 24.22.1:

- The original maintained suite passes against both baseline and correction.
- The expanded maintained suite passes against the correction. Against the
  unchanged exact-base resolver it fails exactly 13 new checks: eight CRLF
  defaults, three CRLF-profile override/fallback successes and two mapping-error
  diagnostics reached only after a valid CRLF profile is recognized.
- Separate clean Git LF/CRLF checkouts reproduce all eight baseline CRLF
  failures and all eight corrected CRLF successes. Both versions pass eight LF
  calls before and after the CRLF controls. The original suite fails exactly
  its eight default checks on baseline CRLF; the new suite passes in both
  corrected checkouts. All four checkouts remain clean, with identical
  resolver/map bytes across representations and exact LF-to-CRLF profile bytes.
- Twenty malformed/unchanged-grammar controls pass, including two terminal CR
  bytes and an internal CR that must not be stripped. Mapping-file CRLF rejection,
  duplicate/malformed/NUL refusal, data-only parsing and validation before output
  remain covered by the maintained suite.
- LF pack validation and whitespace checks pass. Both changed shell files pass
  syntax checks through the pack validator. Optional skills-ref is unavailable.

The tested product blobs, both mode `100755`, are
`d8dcca5316ee9d37b29baa7c5defe302dc77bf81` for the resolver and
`215c00de00838ec8f3190e6ffa37aed39f4946f2` for its test. Genuine completion records
the final staged product/task fingerprints; full committed lifecycle validation
follows that commit.

GNU Stow and ShellCheck are unavailable locally. No full installer, native
macOS/BSD-awk/Bash 3.2, Windows, client/provider or new-product hosted acceptance
is claimed. Independent implementation review and separately assigned draft
publication/hosted qualification remain required before integration. All earlier
task/history and out-of-scope products remain unchanged.

## Integration closeout

[PR98](https://github.com/spik3r/clanker-foundry/pull/98) merged normally as
[`d368cc4ad2739fc47a439b592898577da387752f`](https://github.com/spik3r/clanker-foundry/commit/d368cc4ad2739fc47a439b592898577da387752f),
with tree `01e8046e1239b0038f0b994c049feaa0b1d8e18b`. Its ordered parents are
prior main `6606c5d073815cc19f4745d0f280e4ff8a1aa78f` and accepted profile head
[`53b1992e76fa687c772b0c8a6c8e7389b7e659a6`](https://github.com/spik3r/clanker-foundry/commit/53b1992e76fa687c772b0c8a6c8e7389b7e659a6).
The actual canonical source
[`e26d8c0452a41f1b8cd41fc2554d641d53561e90`](https://github.com/spik3r/clanker-foundry/commit/e26d8c0452a41f1b8cd41fc2554d641d53561e90),
its snapshot and both original lifecycle events remain in normal merge ancestry.
No squash, rebase or history rewrite occurred.

Fresh closeout inspection at main
`6d3998bd88e4624e6e67de556d8f144127a39bd1`, tree
`c99823d04630954f3d9846e35fcec40b3c0b7dcb`, confirms that source, accepted head
and PR98 merge remain ancestors. The two integrated product blobs remain
`d8dcca5316ee9d37b29baa7c5defe302dc77bf81` and
`215c00de00838ec8f3190e6ffa37aed39f4946f2`, both mode `100755`, as recorded
above. The original completion `b4399f3f-7004-4a12-bf28-ad60185ba02a` and its
three fingerprints are preserved. This closeout does not edit either product.

### Accepted hosted qualification

Independent review accepted twelve push/PR workflows and sixteen jobs at head
`53b1992e76fa687c772b0c8a6c8e7389b7e659a6` and current-base test merge
`bbc9610ae00cc58a3fb10dccc4f2ae11fbbaf169`. The
[PR validation](https://github.com/spik3r/clanker-foundry/actions/runs/42564547760)
ran real GNU Stow/full installer and resolver checks, pack and Markdown-path
checks, strict metadata/shared-base regressions, and successful ShellCheck.
Its [committed-evidence job](https://github.com/spik3r/clanker-foundry/actions/runs/42564547760/job/128988089494)
validated three changed paths and two new events against actual base
`6606c5d073815cc19f4745d0f280e4ff8a1aa78f`.

All six separate push workflows and eight jobs passed on actual merged main
`d368cc4ad2739fc47a439b592898577da387752f`, attempt 1:

- [Validation](https://github.com/spik3r/clanker-foundry/actions/runs/42566773474).
- [Project Memory conformance](https://github.com/spik3r/clanker-foundry/actions/runs/42566773553).
- [Model authoring](https://github.com/spik3r/clanker-foundry/actions/runs/42566773489).
- [Orchestration](https://github.com/spik3r/clanker-foundry/actions/runs/42566773466).
- [Model lifecycle Go](https://github.com/spik3r/clanker-foundry/actions/runs/42566773505).
- [JSON validator Go](https://github.com/spik3r/clanker-foundry/actions/runs/42566773527).

Every actual-main checkout and complete job log was inspected. Executed results
include GNU Stow 2.4.1/full installer and resolver checks, ShellCheck 0.11.0,
92 Markdown-path cases, 143 lifecycle/workflow tests, 91 conformance tests and
the graph check, two schema tests, 307 runtime tests and Docker-inspected
isolation. Both Go workflows passed their race/vet/formatting and generated
checks; the JSON-validator workflow also passed 46 synthetic TypeScript cases.
The twelve installer-ordering seam cases remain distinct from real Stow tests.
The pull-request-only committed-evidence step was skipped on push; only the
premerge PR job above supplies that committed-coverage execution. Optional
skills-ref was absent and skipped. These are retained hosted Linux results,
not newly run checks for this metadata-only closeout.

### Completed scope and remaining limits

The bounded profile-record terminal-CR correction is integrated and recorded as
done. The owner remains `astra-foundry-profile-crlf` as provenance. Effective
shared release of only this completed outcome occurs when this canonical
closeout is independently reviewed and integrated. The historical claim and
local implementation result above retain their original phase-specific wording.
No older resolver, reliability-integration or other owner's scope is reopened,
acquired or released.

This task-only closeout requires a genuine new run under the existing published
source, final fingerprints for this task and both unchanged product paths,
committed validation against the actual base, and fresh applicable exact-head/
current-base hosted qualification. Prior green results do not qualify a new
canonical head. Later publication starts as draft; integration must preserve
the original source pin and every immutable record through a normal merge.

Native macOS/BSD awk/Bash 3.2, Windows, whole-pack CRLF, installed-client loading
and provider behavior remain unqualified. The accepted task excluded those
claims; this closeout does not convert Linux checks into native qualification.
No local installation, release artifact, deployment or client/provider action
is claimed. Later platform work requires its own agreed scope and environment.
