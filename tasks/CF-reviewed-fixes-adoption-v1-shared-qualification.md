---
id: CF-reviewed-fixes-adoption-v1
status: done
owner: astra-foundry-reliability-integration
repositories: [spik3r/clanker-foundry]
---

# Adopt and qualify three reviewed Foundry fixes

This task retains the original byte-exact adoption and bounded test correction
as attributed history. The selected successor claim remains unchanged; its scope,
twelve-path coverage, source and history restrictions still apply. The Integration
closeout below records the accepted merged outcome and remaining platform limits.

## Successor test-correction claim

Owner and integration owner: `astra-foundry-reliability-integration`, continuing
the same task and receiving-owner responsibility. The coordinator assigned this
bounded correction after independent hosted review; the owner accepts it.
No new task identity, ownership transfer or unrelated scope is introduced.

Verified current main: `c08604c5f9eebd412af02e95fb06d7846a9f51e2`, tree
`d45d6adae50015117d38bb221d77cfa879847b67`. Actual published adoption head:
`e80a74a478206466087eac099ed21a064baae30a`, tree
`486508815693e299dec29b125008d453348d9d03`, with sole parent and original
canonical source `9f312b3e6297daa78979e22d49a2f52399020a71`.
Continue branch `astra/reviewed-fixes-adoption-v1` and
[draft PR96](https://github.com/spik3r/clanker-foundry/pull/96). Refresh these
identities and shared ownership before publication; reconcile material changes.

The sole writable product path is `tests/test-project-setup.sh`, preserving mode
`100755`. Correct only the two trace-absence assertions and the Git lookup/export
site identified below. Accept only grep status 1 as absence in a readable trace,
including a legitimate empty file. Reject a forbidden match (status 0), missing
or unreadable input, and every other grep/read error. Capture the status without
letting a negated command bypass failure handling. Preserve the positive
lazy-fetch control and actual Git-command checks. Separate the Git lookup
assignment from export so lookup failure propagates before the fixture continues.
Keep Bash 3.2-compatible syntax; this is not native-platform qualification.

The accepted eleven-blob table below remains the exact original adoption
inventory at `e80a74a...`. After independent acceptance, this successor permits
replacing only its `tests/test-project-setup.sh` blob with the bounded correction.
The other ten product blobs and every product mode must remain byte-for-byte
unchanged. No production script, workflow, dependency or other test may change.

The only other mutable project file is this same canonical task. New immutable
source/start/completion records may be added only under
`.project-memory/tasks/CF-reviewed-fixes-adoption-v1/`; preserve every existing
source, event, task record and legacy ledger. Do not amend prior events or import
old unpublished contributor metadata. No profile-CRLF payload or PR94 state
change is included. No workflow weakening, local tool installation, local Go
aggregate, provider/client action, promotion, merge or deployment is included.

Fresh lifecycle coverage must name all twelve original non-metadata PR paths:

- `tasks/CF-reviewed-fixes-adoption-v1-shared-qualification.md`
- `.github/workflows/validate.yml`
- `docs/getting-started.md`
- `docs/project-shared-base.md`
- `scripts/project-setup.sh`
- `scripts/stow.sh`
- `scripts/validate.sh`
- `tests/project-memory-workflow.test.mjs`
- `tests/test-install.sh`
- `tests/test-project-setup.sh`
- `tests/test-stow-preflight.sh`
- `tests/test-validate.sh`

This coverage scope does not authorize edits to the ten frozen product files.
The new terminal completion must fingerprint all twelve final paths because the
latest frontier must cover the full diff against main; the previous completion
cannot supply missing coverage. Retain its original source and causal history.
At preparation, this successor is a source proposal only: no successor pin,
registration, new run or maintained correction has been created.

## Successor validation and publication sequence

1. Independently review the exact successor task text, one-product write scope,
   twelve-path coverage and preserved inventory. After acceptance, refresh main,
   PR head and ownership. Publish only the reviewed canonical-task delta in a
   source commit whose sole parent is the actual accepted PR head. Keep PR96
   draft; this intermediate source-only checkpoint is not final qualified coverage.
2. Read back and independently verify the actual new source SHA, tree, sole parent
   and exact task bytes before registration. Materialize real ancestry and
   register that actual pin with source identity `spik3r/clanker-foundry`.
   Explicitly select the new source and `Successor test-correction claim` when
   recording a genuine new start with code identity `spik3r/clanker-foundry`,
   PR96 and all twelve paths. The helper must link the start to the prior terminal
   completion. No maintained product edit may precede this start.
3. Apply only the bounded test correction. Run the corrected full shared-base
   suite and relevant clean, readable nonmatching/empty trace controls, independent
   raw/fallback forbidden-record mutations, missing/unreadable/read-error controls
   and injected Git lookup-status propagation with both admitted Git versions,
   2.52.0 and 2.47.3. Retain qualification on the actual hosted Git version in step 5
   and record that version. If the stronger assertions expose a production behavior
   gap on an admitted version, retain the failure and stop this test-only scope
   for a separate scope/compatibility decision. Do not weaken assertions or silently
   change `scripts/project-setup.sh`. Run pack validation, Bash syntax and whitespace.
   Preserve existing positive controls and disclose unavailable tools and native
   platforms. Test in invented isolated homes; do not install tools or run a local
   Go aggregate.
4. Freeze the exact corrected test/task diff and observed results for independent
   correction review. After acceptance, stage named final paths, generate genuine
   completion with all twelve final fingerprints, stage only the new additive
   metadata, and commit. Validate the full committed diff against fresh actual
   main with explicit code/source identities `spik3r/clanker-foundry`; obtain
   independent committed-coverage review before corrected product publication.
5. Publish the accepted correction on the same draft, retaining both original and
   successor source pins. Update its description to distinguish original exact
   adoption from the later test-only correction. Verify actual remote head/tree,
   complete diff and current PR base/merge checkout. Obtain fresh existing push
   and PR workflows on that exact final head, including actual ShellCheck, real
   Stow/full installer/shared-base, pack/Markdown, lifecycle/committed evidence
   and every other existing workflow. Require successful corrected ShellCheck
   execution without suppressing diagnostics or changing the advisory job policy.
   Inspect real logs and skips; old-head green results do not qualify a new head.
6. Keep the task in review and PR96 draft pending independent exact-head hosted
   acceptance and a separate promotion/normal-merge decision. Preserve both source
   pins and all immutable history in ancestry; do not squash or rebase it away.
   Any later canonical edit requires genuine fresh coverage and requalification.
   Native macOS/BSD/Bash 3.2, Windows and actual client loading remain unclaimed.

## Published adoption and review findings

The original byte-exact adoption is published at `e80a74a...`, not at the different
local review commit `55db6d00db9c56c08f1ee0c1b33a43c681163946`. Its task and genuine
source/start/completion evidence are preserved. Independent review accepted the
actual source, full tree, complete diff and chronology.

All ten original-head push/PR workflows completed successfully. Logs verify real
Stow 2.4.1, full installer, shared-base and PR committed coverage at push checkout
`e80a74a478206466087eac099ed21a064baae30a` and PR merge checkout
`a882b80dfd6583137f10833508cfee0a778b31e0`. Optional skills-ref was absent and
explicitly skipped. Workflow-level success did not override the advisory
ShellCheck failures at original test lines 403/406 (SC2251) and 414 (SC2155).

Independent whole-suite mutation checks proved that forbidden `child_start`
records in the two traces did not fail the original suite. Separate controls
confirmed that the lookup/export form masked injected lookup status 73. Disposable
corrected controls accepted readable nonmatching and empty traces, rejected
forbidden records and read errors, and propagated lookup failure. Those controls
inform this proposal; they are not a maintained correction or new-head acceptance.
No production defect was demonstrated. The original hosted results remain
attributed evidence, but the assertion gap prevents promotion until corrected.

## Original integration claim (historical)

Owner and integration owner: `astra-foundry-reliability-integration`. Accept the
coordinator's bounded handoff for a distinct shared integration and qualification
outcome: adopt the reviewed installer preflight, literal Markdown filename
validation and strict `.gitmodules` CRLF representation fixes onto a branch based
on current main without changing their eleven accepted product blobs or modes.
The original local implementation work and its history remain intact.

Verified base: `c08604c5f9eebd412af02e95fb06d7846a9f51e2`, tree
`d45d6adae50015117d38bb221d77cfa879847b67`.
Intended source and integration branch: `astra/reviewed-fixes-adoption-v1`.
Canonical destination: `tasks/CF-reviewed-fixes-adoption-v1-shared-qualification.md`.
This records an actual coordinator decision and receiving-owner acceptance. At
preparation, it is still an unpublished source proposal awaiting exact independent
review, not a shared reservation. No source pin, branch, PR, lifecycle registration,
start or completion exists for this adoption. Publish only after that review.

Refresh main, canonical tasks, open work and branch availability immediately before
publication, and verify the shared claim after publication. Reuse a matching
shared adoption task if one appears; this ID is not reserved by the scratch copy.
Stop and coordinate any new conflicting work rather than assuming exclusivity.

## Original ownership and qualification sequence decision (historical)

The coordinator accepted the completed local handoffs from Foundry installer
contributor (`CF-global-migration-preflight-v1`), Foundry pack-validation contributor
(`CF-pack-markdown-path-v1`) and Foundry project setup implementation
(`CF-project-metadata-crlf-v1`) for the bounded reviewed outcomes below, not for
all historical claims or qualifications. `astra-foundry-reliability-integration`
accepts responsibility for their byte-exact shared adoption, current integration
checks, truthful evidence and later hosted qualification. This does not transfer
any other live owner, claim or outcome, or declare a historical claim expired.

Current-main task and open-work checks found no matching shared adoption task or
branch. The nearby canonical records remain intact and retain their owners:

- `CF-001`: preserve lifecycle behavior, permissions and all existing workflow
  assertions. This adoption adds only the accepted separate Markdown regression
  step and its exact assertion; it does not take over lifecycle implementation.
- `CF-002`: preserve the literal pointer match and its ShellCheck exception.
  The adopted validator hunks concern literal Markdown filenames only.
- `CF-project-shared-base-v1`: preserve the landed setup and shared-base CI work.
  The accepted follow-on changes only the bounded metadata representation check,
  its regressions and guide; it does not reopen or close the original outcome.
- `CF-readme-navigation-v1`: preserve the landed navigation and instructions.
  The only adopted getting-started change documents the bounded installer preflight.
- `CF-agent-model-duplicate-keys`: its exact installer scope is the separate
  invocation of `tests/test-agent-model.sh`. That invocation and every existing
  installer byte are preserved. The adoption adds only
  `bash "$repo_dir/tests/test-stow-preflight.sh"` immediately after `repo_dir` is
  set; removing that line reproduces the exact main installer blob. This distinct
  hunk and outcome do not take over the resolver claim or depend on its release.

[PR95](https://github.com/spik3r/clanker-foundry/pull/95) is now merged normally
as `c08604c5f9eebd412af02e95fb06d7846a9f51e2`, retaining reviewed head
`0cea6d1b15331786965ab1bfeb60d0c2f6bce513` and its source parent
`d0a019051cd4d1225fac9c93fee41240fe2532ed` in main ancestry. This integrated
metadata-only closeout changes its resolver task and adds three immutable
evidence files; it changes none of the eleven adoption paths. The task now records
the completed duplicate/malformed/NUL mapping outcome as done, and only that
completed scope is released. Preserve the exact newly landed task and evidence,
all prior history, original owner provenance and recorded native acceptance.
The adoption does not claim the separate profile-Markdown CRLF work. Main-push
checks concern that merged source and do not qualify this adoption candidate.
[PR94](https://github.com/spik3r/clanker-foundry/pull/94) is path-disjoint, remains
draft and is excluded from this adoption; any integration of it is a separate
coordinator decision. No visible competing adoption outcome was found;
these checks do not establish the absence of unseen private work.

The installer claim expressly kept local preparation on hold until ownership,
supported-platform/full-installer qualification and publication authority were
resolved. A draft PR is publication. The coordinator has now explicitly accepted
this prospective qualification-order decision for the distinct adoption:

After independent review of this exact source, permit publication of the new
canonical source, followed by genuine bounded integration and an independently
reviewed **draft PR solely to obtain exact-candidate GNU Stow/Linux qualification
through the existing hosted workflow**. This resolves the earlier local-only
hold only for that sequence; it does not rewrite the original claim or treat it
as having already allowed publication. The existing hosted GNU Stow installation
and full checks are authorized for this purpose. Full real-Stow installer and
shared-base acceptance remains required before promotion, merge or release.
No check, assertion, workflow selector, security setting or protection rule may
be weakened, and no new workflow change beyond the accepted delta is included.

Native macOS/BSD/Bash 3.2, Windows and actual client loading are unclaimed unless
separately brought into scope and executed. Preserve Bash 3.2-compatible source.
Those platforms are not universal draft gates for these three changes. The
separate resolver's explicit native acceptance and owner scope remain untouched.

## Immutable local provenance

These are **unpublished local objects**, not remote commit links or new task
sources. The original three source pins were reported absent by GitHub commit
reads. Tree equivalence and green lifecycle CI do not prove their remote existence.

- Installer: claim `ba71edf949afb410c50d63aa52299f187a981084`; implementation
  `542fafc6a7a825c380b8048147e862d135460931`.
- Markdown paths: claim `aee2d1790f78ff163ee3d6d4270869e36180b98b`; implementation
  `735621d982a2757ac9e2abaa0e2a8b4e5dba884f`.
- CRLF metadata: claim `63df9167698be3141ad5ec589d5331e036ab0090`; implementation
  `5b723b60d77f37e2a8fc1fbcbfbe78e51797e18b`.
- Accepted review-only composition: `8ee650d59d742f4add0475304cff4311f15dc4e3`,
  tree `e74fa18bd165688c637d91526fc9ec86314cb8f4`; earlier two-fix review merge
  `2116ba33af19c944f1228060e7bd48227847b5ac`.
- Retained eight-commit bundle SHA-256:
  `9a74ab5b9b5e8c00a3f9d2e5aeea5310da8bc110efb72b9c54f72453e0fde394`.

Retain all original commits, events, task records and transfer artifacts unchanged
outside the new adoption branch. Do not recreate their history with replacement
SHAs, rewrite their events, fabricate ancestry or imply the original objects were
published. No supported exact-object transport was verified in the preparation
environment. This adoption reuses reviewed bytes in genuinely new shared work;
it does not represent the old local commit chain as published.

Prior independent local qualification passed pack validation, 92 Markdown path
checks, 12 recording-seam preflight cases, resolver and shared-base suites,
143 lifecycle/workflow and 91 conformance tests, 230 CRLF controls and 12 interaction
controls. These are attributed historical results at the local composition, not
new execution on an adoption candidate. The Stow recording seam proves ordering,
not real GNU Stow acceptance. At that original preparation stage, full
installer/native/client/hosted acceptance of the adoption remained outstanding.

## Original adoption inventory (historical)

The original adoption used these eleven exact product blobs and modes. The
successor permits only its named test correction; the other ten remain frozen:

| Path | Mode | Accepted blob |
| --- | --- | --- |
| `.github/workflows/validate.yml` | `100644` | `52484b793f6c262d91fa647c65303f86c952b1fd` |
| `docs/getting-started.md` | `100644` | `d80b771ce89d1bc7bdd22666a82346888df9ccd1` |
| `docs/project-shared-base.md` | `100644` | `991a51343b86e5313d48c0518f908f09802555ba` |
| `scripts/project-setup.sh` | `100755` | `7f5f830cbb12b445444d1199def295dce600b87c` |
| `scripts/stow.sh` | `100755` | `2e1abf65100e8ac04c04d1ce94d0cd542c776bed` |
| `scripts/validate.sh` | `100755` | `67d44f33fb7f048053c30b1802117caa9c50f21b` |
| `tests/project-memory-workflow.test.mjs` | `100644` | `c2d80545b71c22ea0ec1323a541fb6e73e01a753` |
| `tests/test-install.sh` | `100755` | `af703148bc4826894350449bb0438e49e4ce8f80` |
| `tests/test-project-setup.sh` | `100755` | `bc89d016a742c100bc083ed873742cac3f33682a` |
| `tests/test-stow-preflight.sh` | `100755` | `e964d86329763b9bef5f4fdeaaa8645e628c4aa6` |
| `tests/test-validate.sh` | `100755` | `1395316d9d7f29c0ea510df3d0d842e2065accb0` |

The only other writable paths are this new canonical task and its own newly
generated immutable v2 source/event files under
`.project-memory/tasks/CF-reviewed-fixes-adoption-v1/`. Name those generated files
explicitly when staging. Do not invent their pins, UUIDs, times or validation
results now. Include this task record and all eleven product paths in the genuine
integration run's exact coverage, including unchanged source-claim text as needed
for the full diff against main.

Do not import the three local task records or their nine source/event JSON files.
None is in the verified base; this omission deletes nothing from main. Preserve
all main task records, source snapshots, events and legacy ledger bytes and modes.
If main later contains any of those local records, stop and reconcile preservation
and ownership before proceeding. No unrelated main path may change.

The accepted workflow addition is only the dedicated Markdown regression step,
with its matching assertion. Retain main's shared-base step and every existing
trigger, permission, check and assertion. No further CI change is needed.

## Behavior acceptance

- Installer: reject edited canonical copies and occupied migration backups before
  Stow effects, preserve diagnostics and synthetic files, keep late recheck and
  migration ordering, and pass the maintained full installer with real GNU Stow.
  Do not claim transactionality, rollback or concurrent-writer safety.
- Markdown: preserve literal tracked/unignored filenames, NUL-delimited discovery,
  real parent directories and option-safe reads. Retain ignored-file behavior,
  CF-002's literal pointer check and current Markdown parsing policy.
- CRLF: preserve raw equality first; allow only exact one-way LF-to-CRLF expansion
  of a stage-0 regular blob without CR/NUL. Preserve EOF, byte/mode checks,
  fail-closed command errors, filter/FSMonitor suppression, no fetch and no writes.
  Do not claim general normalization or Windows support.
- Preserve the original adopted blobs/modes as attributed inventory. For the
  successor, verify the separately reviewed corrected test blob, the ten frozen
  product blobs/modes and every other main entry against the current base. Review
  the complete diff and all ownership scopes. No further product change may be
  silently folded into the adoption or its bounded successor.

## Original lifecycle and publication sequence (historical)

1. Independently review this exact active-claim proposal and its recorded handoff,
   receiving-owner acceptance, disjoint scopes and hosted-first decision. Refresh
   main and visible shared work before publication; stop for any material change.
   Preserve the integrated resolver closeout and draft PR94 without action.
2. Publish a new claim-only source commit based on freshly verified main through
   the supported repository tools. Read back its actual server SHA, tree, parent
   and task bytes, confirm the shared claim and independently verify the source.
   Use the actual server object as the new source pin, even if its SHA differs
   from a local draft. No event may refer to an unverified or guessed pin.
3. Materialize that exact published source in a separate integration checkout.
   Register it with explicit code/source identity `spik3r/clanker-foundry` and
   record a genuine current start before applying the eleven accepted product
   files. Do not merge or cherry-pick the original local claim/event histories.
4. Run relevant current checks with invented HOME/XDG/CODEX and isolated Git
   configuration. Run pack, Markdown, focused preflight, resolver, shared-base,
   lifecycle/workflow and conformance suites, shell syntax and whitespace checks.
   Record unavailable real Stow, ShellCheck, skills-ref or native checks honestly.
   No real profile access or local/global package installation is authorized.
   The unchanged hosted job's GNU Stow install is permitted by the decision above.
5. Finish task-result text, obtain independent exact-diff review, stage only the
   named final paths, record a genuine completion for the current local integration
   work and observed checks, commit its generated evidence, and validate the full
   committed diff against actual main with both explicit repository identities.
   Local completion does not mark the adoption done or the installer accepted.
6. Publish the reviewed integration on top of the verified source pin and open
   the specifically authorized draft. Verify actual remote commit, tree, parent,
   task source and immutable events. Server-selected commit metadata must never be
   represented as exact publication of a different local commit.
7. Obtain exact-candidate results through the existing hosted workflow. Inspect
   logs proving real Stow installation and execution of pack, dedicated Markdown,
   full installer and shared-base tests, plus lifecycle/workflow and PR committed
   lifecycle checks. Preserve other applicable workflows and required checks.
   Inspect actual checkout SHAs: push tests the branch head; PR tests its synthetic
   merge with the then-current base. Record both identities and step results;
   Actions head_sha or an overall green badge alone is insufficient.
8. Keep the task in review and the PR draft until actual acceptance and a separate
   promotion/merge decision. Record hosted evidence in a genuine follow-on run if
   canonical task text changes; preserve prior events and refresh full coverage
   for all paths changed against main. Recheck the resulting exact head and PR
   merge. Never amend a completed event or relabel older runs as new evidence.
   Close only after the adopted outcome's required integration checks pass.

If a later merge is authorized, use a normal history-preserving merge that
retains the actual published canonical source commit and integration chronology.
Do not squash, rebase or otherwise rewrite that source history. Before merging,
verify that the source pin is an ancestor of the reviewed PR head. After merging,
verify the actual merge commit and that the same pin is an ancestor of main,
with the reviewed source/event bytes preserved. If repository policy prevents
that merge method, stop for a coordinator decision without changing protections.

For committed validation use the actual base and explicit identities:
`node skills/project-memory/scripts/project-memory.mjs validate --base ACTUAL_BASE
--repository spik3r/clanker-foundry --source-repository spik3r/clanker-foundry`.
No synthetic event-only rerun replaces real adoption or validation work.

## Hosted feasibility and remaining limits

The existing workflow installs real GNU Stow on `ubuntu-latest` and has no draft
exclusion. A prior draft's [retained installer/shared-base job](https://github.com/spik3r/clanker-foundry/actions/runs/42440385361/job/128608155615)
checked out `86f3794d9bf2dd62d36144aeeb92bdfcb45a340d`, installed Stow 2.4.1 and
ran those suites. That is route feasibility, not acceptance of these fixes. Its
older push job lacked the later shared-base step; inspect actual steps on both
new push and PR runs rather than borrowing an old green result.

ShellCheck remains advisory; optional skills-ref absence is a disclosed skip.
Configured required statuses must be checked independently before any promotion
or merge. No protection change, release, deployment, provider execution, actual
client load, global installation, resolver implementation, profile CRLF work,
or PR94 state change belongs to this task.

## Original local integration result

The actual published claim source is
[`9f312b3e6297daa78979e22d49a2f52399020a71`](https://github.com/spik3r/clanker-foundry/commit/9f312b3e6297daa78979e22d49a2f52399020a71),
whose sole parent is the verified main recorded above. Its exact source and
registration were independently reviewed. A genuine integration start under that
source preceded adoption of all eleven product paths. Their bytes and modes match
the accepted table unchanged. The original three local task records and nine
local source/event files were not imported. Every current-main task/evidence
record, including the integrated resolver closeout, is preserved.

Observed checks on the adopted code passed on Linux, Bash 5.2.37 and Node 24.22.1,
using invented temporary HOME, XDG_CONFIG_HOME, CODEX_HOME and Git configuration:

- `bash scripts/validate.sh`: pack conventions, shell syntax and Markdown links.
- `bash tests/test-validate.sh`: 92 literal Markdown-path checks.
- `bash tests/test-stow-preflight.sh`: 12 recording-seam ordering cases. These do
  not establish real GNU Stow acceptance.
- `bash tests/test-agent-model.sh`: existing resolver regressions.
- `bash tests/test-project-setup.sh`: the full maintained setup, guide-snippet,
  strict representation, filter, promisor and failure controls on Git 2.52.0
  and Git 2.47.3.
- `node --test tests/project-memory.test.mjs tests/project-memory-workflow.test.mjs`:
  143 passed, with zero failures or skips.
- `node --test tests/project-memory-conformance.test.mjs`: 91 passed, with zero
  failures or skips. The separate conformance `check` also passed.
- Explicit Bash syntax checks for the ten maintained shell entrypoints and
  whitespace checks passed.

GNU Stow is unavailable locally, so the full installer was not run. ShellCheck
and optional skills-ref are unavailable; optional skills-ref was explicitly
skipped by the pack validator. No tool was installed. Native macOS/BSD/Bash 3.2,
Windows and actual client loading remain untested. These current local results
do not replace exact-candidate hosted installer/shared-base qualification.

## Local successor correction result

The independently verified successor source is
[`6e824dc675245794dbabb9118d9348966f45abc1`](https://github.com/spik3r/clanker-foundry/commit/6e824dc675245794dbabb9118d9348966f45abc1),
with actual adoption `e80a74a...` as its sole parent. Its exact registration was
independently checked. A genuine new twelve-path start explicitly selected this
source, the successor claim and PR96, linked to the prior terminal completion,
and preceded the maintained correction.

Only `tests/test-project-setup.sh` changed as a product, at the two trace guards
and the assignment/export site. Its corrected local blob is
`69a0509b3edb31e63bec28d1970ceb984c793c6c`, mode `100755`. Each trace guard now
accepts only grep status 1 and fails on matches or inspection errors. The Git
lookup assignment precedes export. The other ten product blobs/modes, production
scripts, workflows and all old source/events remain unchanged. The original
eleven-blob table remains attributed adoption history, not the corrected test's
new fingerprint.

Observed local checks on Linux and Bash 5.2.37 passed:

- The corrected full `bash tests/test-project-setup.sh` suite passed separately
  with Git 2.52.0 and Git 2.47.3. Ordinary-Git positive controls were retained;
  no outer no-lazy-fetch or no-replacement environment suppressed those controls.
  No production behavior gap was exposed on either admitted version.
- Four disposable whole-suite mutations injected forbidden `child_start` records
  into raw and fallback traces independently, once per admitted Git version.
  Every mutated suite exited 1 at the intended guard, with grep status 0, before
  the strict-metadata and final PASS messages.
- Twenty-eight isolated trace probes exercised the exact maintained guard blocks
  in both Git environments: readable nonmatching and empty traces were accepted;
  forbidden records, missing paths, directories, genuinely unreadable files and
  injected grep status 3 were rejected. All four unreadable-file controls actually
  produced grep status 2 with permission-denied diagnostics; these are distinct
  from missing-path/directory errors and injected status controls.
- Four lookup probes exercised the exact assignment/export sequence in both Git
  environments. Successful lookup continued; injected failure exited 73 before
  continuation. These isolated controls do not claim a real Git lookup failure
  occurred in the clean suites.
- Pack validation, explicit syntax checks for all ten maintained shell
  entrypoints and whitespace checks passed. Pack validation was also checked
  with the final task text.

These checks use invented isolated HOME/XDG/CODEX/TMPDIR values and Git
configuration. Full-suite mutations and isolated Bash probes are distinguished
above; they are not identical forms of coverage. ShellCheck, GNU Stow and optional
skills-ref remain unavailable locally. Optional skills-ref was explicitly skipped
by the pack validator. No tool was installed and no local Go aggregate ran.
Native macOS/BSD/Bash 3.2, Windows and actual client loading remain unqualified.
Fresh actual hosted Git, ShellCheck, Stow/full installer/shared-base and all other
existing workflows remain required on the final corrected head.

## Precompletion state (historical)

The bounded correction and factual task update are local and ready for exact-diff
review. PR96 remains draft at the published successor source; the correction has
not been published. No new completion or implementation commit has been recorded.
Independent correction review must precede genuine completion for all twelve
final fingerprints, committed validation and independent coverage review.
Corrected publication then requires fresh exact-head/current-base hosted
qualification, including successful ShellCheck and the actual hosted Git version.
No old-head or source-only green result qualifies this correction. Any later
production behavior gap on an admitted version still stops this test-only scope
for a separate compatibility decision, without weakening assertions or changing
production scripts. Promotion and normal merge remain separate decisions;
preserve both source pins, immutable history and all unrelated scopes.

## Integration closeout

[PR96](https://github.com/spik3r/clanker-foundry/pull/96) merged normally at
`2026-11-09T02:52:39Z` as
[`a43f03b75a58f6c4636f6482f189f9d946b517c8`](https://github.com/spik3r/clanker-foundry/commit/a43f03b75a58f6c4636f6482f189f9d946b517c8),
with accepted tree `529d150b80db1bb3973dd1ea45fdbfceb22ceb20`. Its ordered parents
are prior main `c08604c5f9eebd412af02e95fb06d7846a9f51e2` and actual corrected head
[`8ad9765bd480f65afa9013a73b5f562e8c6cf848`](https://github.com/spik3r/clanker-foundry/commit/8ad9765bd480f65afa9013a73b5f562e8c6cf848).
The actual merged tree, all 1,238 leaves/modes and source ancestry were independently
verified. No squash, rebase or history rewrite occurred.

The corrected head's sole parent is actual successor source
[`6e824dc675245794dbabb9118d9348966f45abc1`](https://github.com/spik3r/clanker-foundry/commit/6e824dc675245794dbabb9118d9348966f45abc1),
whose parent is original adoption `e80a74a478206466087eac099ed21a064baae30a`,
whose parent is original source
[`9f312b3e6297daa78979e22d49a2f52399020a71`](https://github.com/spik3r/clanker-foundry/commit/9f312b3e6297daa78979e22d49a2f52399020a71).
Both published source pins and every prior immutable record remain in main.
The unpublished contributor chain remains local provenance, not published ancestry.

At that integrated main, all twelve fingerprints of final correction completion
`b3a8439f-1e6b-4835-b9b4-98b52371319b` match the accepted tree. The corrected
`tests/test-project-setup.sh` blob is `69a0509b3edb31e63bec28d1970ceb984c793c6c`,
mode `100755`; the other ten products and every product mode retain the original
adoption inventory. The original and successor source snapshots, both starts,
both completions, all 586 task/evidence entries inherited from the pre-adoption
base and unrelated content are preserved. The correction changed only test
assertions and lookup-status handling; no production defect was demonstrated.

### Accepted hosted qualification before merge

Independent review accepted all ten push/PR workflows and fourteen actual jobs
at corrected head `8ad9765...` and current-base PR checkout
`60288afbbb95eb0a18111c5b188ecb14cc796d0b`. That checkout has the accepted tree
and ordered parents `c08604c5...` then `8ad9765...`.
[Push validation](https://github.com/spik3r/clanker-foundry/actions/runs/42551074667)
and [PR validation](https://github.com/spik3r/clanker-foundry/actions/runs/42551079397)
executed successful ShellCheck, real GNU Stow/full installer, resolver, pack,
Markdown and strict shared-base checks on hosted Git 2.56.0.
[PR committed validation](https://github.com/spik3r/clanker-foundry/actions/runs/42551079397/job/128946576824)
passed against the stated base with twelve covered paths and four new events,
using the sole latest completion rather than a union with the old completion.
Conformance, schema, orchestration/Docker and existing Go workflow checks also
passed. These are executed premerge results, not runs on the later merge commit.

### Accepted qualification on actual merged main

All five ordinary main-push workflows and seven jobs completed successfully on
actual merge `a43f03b75a58f6c4636f6482f189f9d946b517c8`, attempt 1:

- [Validate](https://github.com/spik3r/clanker-foundry/actions/runs/42553305809).
- [Project Memory conformance](https://github.com/spik3r/clanker-foundry/actions/runs/42553305843).
- [Model authoring contract](https://github.com/spik3r/clanker-foundry/actions/runs/42553305779).
- [Orchestration](https://github.com/spik3r/clanker-foundry/actions/runs/42553305806).
- [Model lifecycle Go](https://github.com/spik3r/clanker-foundry/actions/runs/42553305770).

Independent inspection of every job log confirmed the actual merge checkout,
executed steps and results: ShellCheck 0.11.0; real GNU Stow 2.4.1; full installer,
resolver, twelve preflight cases and strict shared-base checks on Git 2.56.0;
pack and 92 Markdown-path cases; 143 lifecycle/workflow tests; 91 conformance tests
plus their check; two schema tests; 307 runtime tests plus Docker isolation; and
the existing Go 1.27.3 race tests, vet, model validation and generated-artifact
checks. No failure was hidden by advisory ShellCheck policy.

The committed-PR-validation step was correctly skipped on this push event. The
accepted premerge PR run supplies its actual committed-coverage result; the skip
is not another successful coverage execution. Optional skills-ref absence remains
an explicit pack-tool skip. No workflow rerun or setting change was needed.

### Completed scope and remaining limits

The integrated installer preflight, literal Markdown filename and strict
`.gitmodules` CRLF adoption, including its bounded test correction, is recorded
as done. The owner remains `astra-foundry-reliability-integration` as provenance.
Effective shared release of only this completed scope occurs when this canonical
closeout is independently reviewed and integrated. This does not release or
acquire profile-CRLF, PR94, resolver or any other owner's outcome.

Historical local limits and failures remain attributed above. The original
advisory ShellCheck failures and false-pass assertions were corrected rather
than erased. The intermediate source-only
[PR committed-evidence failure](https://github.com/spik3r/clanker-foundry/actions/runs/42546893048/job/128933711891)
was caused by the changed task lacking fresh final completion evidence; it is
not a failure of the accepted corrected head or merged main. No original event
has been rewritten to claim later execution.

This closeout changes only task metadata. Its publication still requires a genuine
new run under the unchanged published successor claim, all twelve final
task/product fingerprints, current committed validation and fresh exact-head/
current-base hosted qualification. All eleven product blobs/modes and old
evidence must remain unchanged; prior green results alone do not qualify a new
canonical head. Any later integration must preserve both source pins and all
immutable history through a normal merge.

Native macOS/BSD/Bash 3.2, Windows and actual client loading remain unqualified.
No release artifact, deployment, global installation or provider/client action
is claimed. The separate resolver's native obligation and unrelated ownership
remain untouched. Any production behavior gap exposed on an admitted Git version
still requires a separate compatibility decision, without weakening assertions
or silently broadening the product scope.
