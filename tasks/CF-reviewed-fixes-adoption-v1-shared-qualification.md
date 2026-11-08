---
id: CF-reviewed-fixes-adoption-v1
status: in-progress
owner: astra-foundry-reliability-integration
repositories: [spik3r/clanker-foundry]
---

# Adopt and qualify three reviewed Foundry fixes

## Integration claim

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

## Ownership and qualification sequence decision

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
not real GNU Stow acceptance. Full installer/native/client/hosted acceptance of
this adoption candidate remains outstanding.

## Exact write scope

Adopt only these eleven product paths, with the exact Git blob and mode shown:

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
- Verify every adopted blob/mode against the table and every other main entry
  against the current base. Review the complete diff and all existing ownership
  scopes, not just the newly added task. A product change requires new scope/review;
  it cannot be silently folded into this byte-exact adoption.

## Honest lifecycle and publication sequence

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

## Current state

Coordinator adoption and receiving-owner acceptance are recorded for this bounded
outcome. The active canonical claim is proposed for independent review before
publication; the scratch copy reserves nothing. No integration has started and
no current hosted acceptance exists. The original artifacts remain accepted only
within their recorded limits. Next: independently review these exact source bytes,
then refresh shared state and, only if still clear, publish and verify the actual
new canonical source before registration and a genuine start.
