---
id: CF-builder-demo-prepare-v1
status: in-progress
owner: Codex builder preparation boundary correction
repositories: [spik3r/clanker-foundry]
---

# Prepare a private Go builder demonstration

## Implementation claim

Start from reviewed main 882b5002bd4dc70b4be0060e5e5668d1faa8bbdd. Add a bounded
Go preparation step joining the unchanged FirstDesignDraft worked handoff and
frozen Go acceptance with the existing builder input preflight. It prepares a
fresh private workspace and a path-bound handoff draft. It does not dispatch a
builder, compile or run a candidate, establish permission, or implement a generic
workspace service. Existing input, fixture, workflow and evidence bytes stay fixed.

Exact maintained scope:
- modeling/lifecycle/builderdemo/inputs.json
- modeling/lifecycle/builderdemo/prepare.go
- modeling/lifecycle/builderdemo/files.go
- modeling/lifecycle/builderdemo/handoff.go
- modeling/lifecycle/builderdemo/prepare_test.go
- modeling/lifecycle/builderdemo/directory_boundary_test.go
- modeling/lifecycle/builderdemo/README.md
- modeling/lifecycle/cmd/modelbuilderdemo/main.go
- modeling/lifecycle/cmd/modelbuilderdemo/main_test.go
- examples/model-builder-go/preparation.md
- tasks/CF-builder-demo-prepare-v1-bound-workspace.md
- new immutable source/event files under this task ID

No changes to prior models, packet, acceptance, evidence, workflow, compiler,
runtime, preflight, existing CLI routes or CI. No extra-platform work, dependency
installation, candidate implementation/execution, publication, merge or deployment.

The PR #84 correction is owned by the Codex builder preparation boundary
contributor. Reproduce differently cased directory aliases on the native Mac,
preserve the reproduction as an additional regression, and replace lexical
containment with directory identity checks without changing the contract above.
Keep the original frozen inventory and both original acceptance files unchanged.
Production edits are limited to prepare.go and files.go, with the additional
regression, README clarification, this task and immutable lifecycle records.
Publication of the correction is separately coordinated; this claim grants no
publication or execution authority.

## Ownership check

GitHub main, open pull requests and builder/handoff branches were checked on
4 November 2026. Main is the commit above; no open PR was returned. The current
claims were inspected. The coordinator reported no verified active owner for
this additive preparation lane and accepted this boundary before edits. Historical
status labels and branch presence do not prove active ownership. This local claim
does not acquire an external lock; recheck before publication.

At correction start, PR #84 still has head
cd2ca0d7995b3a81a425dd550b9b9197a8e97876 and main remains the baseline above.
It is the only open PR, the worktree is clean, and the coordinator assigned this
reported boundary defect to the correction owner. The prior run is complete
with no conflicting Project Memory frontier.

## Prospective contract

This command serves only the fictional FirstDesignDraft Go pilot. The fixed
source inventory carries original exact hashes from the reviewed baseline, not
hashes accepted from whatever bytes arrive. It includes the original brief,
model, public profile, five frozen acceptance files, public packet, governing
instructions and named handoff guidance. The inventory uses the existing
model-builder-inputs/v1 format and is frozen before production implementation.
Sources must come from an explicit trusted checkout of the pinned baseline.
Legitimate later source changes require separately reviewed new preparation
support, not automatic digest refresh.

Prepare(sourceRoot, workParent, goBinary string) returns a Receipt and error.
All three paths must be absolute, nonempty UTF-8 without control characters.
Root and parent final symlinks are refused; ancestor aliases retain host traversal
semantics. The existing work parent must resolve outside the source root.
The named Go path is recorded, not opened, run or represented as verified.
Arguments and location constraints fail before creating output. Existing trusted,
exclusive input and output parents are required; no hostile-writer, mount,
hardlink, sandbox or atomic snapshot guarantee is made.

Run the existing byte preflight against the fixed source manifest. A blocked
source returns its complete unchanged report and creates no workspace. Do not
search, enumerate source directories, rewrite inputs, refresh pins or fall back.
Read only fixed listed regular files beneath the root, using bounded reads and
checking each copied byte string against its fixed hash. Source replacement or
changed bytes during collection must block without output. Do not copy bytes
merely because an earlier preflight matched.

Only after inputs are collected, create a fresh exclusive 0700 directory named
foundry-builder-* under the explicit work parent. Never reuse an existing directory.
Copy fixed inputs byte-for-byte under inputs/ with their original relative paths.
Copy exactly the five frozen module files into exercise/. Create cache/, tmp/ and
modcache/. Do not create candidate.go or candidate.test. New regular files are
0600 and directories 0700, subject to a stricter umask. Do not change source modes.
Use exclusive writes; no overwrite, cleanup or deletion of previous work.

Write source-manifest.json, source-report.json, exercise-manifest.json,
inputs-report.json, exercise-report.json and handoff.md. Reuse the same fixed
source manifest for checking the copied inputs. The exercise manifest maps only
the five original fixed module hashes to their exercise filenames; its digest is
explicit in the draft. Both destination preflights must match before a prepared
draft is reported. A later I/O error retains an incomplete workspace and its path;
no incomplete result is called prepared. These reports prove listed bytes only.

The handoff draft begins with an explicit initial read boundary. It identifies
all input paths and hashes, exact candidate-only edit frontier, named build
products/caches and a fully expanded offline compile-only command represented
as a program, argument array, working directory and environment, not executable
shell interpolation. It retains Q_use and visual approval as open, original
pilot evidence/authority as historical, and current claim, semantic review,
independence, tool version and edit/build authority as coordinator checks. It is
not ready for dispatch until the coordinator independently supplies those facts
and grants. It never claims permission from matching bytes. No instruction to
execute tests or a candidate is included. It repeats that go test -run still
executes initializers; only proposed go test -c is build-only. No tool is run.

Receipt fields: formatVersion model-builder-demo-preparation/v1; outcome
prepared-draft, blocked or incomplete; workspace and handoffPath (null before
creation); sourceManifestSHA256; sourceReport, inputsReport, exerciseReport as
full JSON reports or null when unrun; executionStatus not-run; reviewStatus
not-established; authority none. The public API exposes ErrInvalid and ErrBlocked.
A nil error means prepared-draft only. Invalid arguments yield no meaningful
receipt. Blocked collection uses ErrBlocked. Other failures are incomplete.

Standalone CLI accepts exactly prepare --root ABS --work-parent ABS --go ABS,
with the three flag/value pairs in any order, no duplicate/default/extra flags
and no stdin consumption. Exit 0 delivers a prepared draft receipt, 1 a blocked
receipt, 2 invalid arguments/locations, 3 failed/short stdout delivery, 4 an
incomplete preparation or internal failure. A rejected invocation emits no report.
If output fails after workspace creation, diagnostics retain its exact path so
an operator can find the incomplete delivery. Nothing is silently retried or
removed. This is a new standalone command, not a competing execution route.

## Evidence sequence

Commit the task, register its pinned source and record a real start before edits.
Have a separate author freeze original source hashes and literal preparation/CLI
acceptance before production code. Then implement and test on the native installed
Go toolchain, obtain independent exact-source review, and rehearse preparation
against unchanged real inputs without writing or executing a candidate. Validate
repository rules, Project Memory chronology, native tests, vet and formatting.
Record observed limits and keep a local full-history handoff without publication.


## Local qualification and retained limits

The source claim was committed at c6c957f046a3082416e757dacd0b7e954c4baf7e.
A genuine start preceded all implementation work. A separate acceptance author
verified the 15 source hashes against baseline main, retained the eight original
packet pins and froze the inventory and both test files in
3bab7afbe6eb70a2d53d2f740672a2b0b69ed01b before production source was created.
Their hashes remain unchanged:

- inputs.json: 6a87f4db69a9c43376b405aca0d1252a662a3a4cdc7496329f2b41fac2896032
- prepare_test.go: c5c262e666b4e3758964351e41fb1f6f518dbdcbf9379cf19958ba43fc6f7273
- main_test.go: 831c90ff2fc576c1234b627d8856348142dd98ed6851cf66296dcc27725922ba

Independent contract and exact-source reviews found no remaining implementation
blocker. The final handoff explicitly distinguishes historical packet authority
from current coordinator checks and specifies a replacing command environment.
The walkthrough was corrected to check the native Go version under offline/local
settings, and to name the module working directory for test commands.

A real-input preparation rehearsal produced three matched native reports and a
private 26-file workspace: 15 source copies, five frozen module copies and six
handoff/evidence files. All original source bytes remained unchanged. A separate
negative source with one added model newline returned exit 1, reported only that
source mismatch and created no workspace. An actual closed stdout pipe returned
exit 3 and retained the completed workspace path in its diagnostic. An independent
receiving-coordinator reviewer checked only named inputs and found no contradiction
between the brief, model, interface, 30 literal cases and isolation assertions.
This is a reviewed preparation draft, not a fresh builder dispatch or approval.

Native Go 1.27.3 checks passed: 87 tested module packages under race, with two more
compiling without tests; full-module vet; formatting; model validation and generated
drift; and the 12 new preparation/CLI test entry points. Permission-error cases ran
as a non-root user. The existing Unix-socket fixture was explicitly excluded because
this host previously denied its creation; no permission workaround was attempted.
The first full race run failed the unchanged regeneration test on temporary-directory
Git discovery. A dedicated valid local Git boundary for scratch corrected that
environment issue; the focused rerun and repeated full suite passed without changing
source, tests or Go's VCS checks. The failed log is retained separately.

Repository validation and walkthrough shell syntax passed. All 234 Project Memory
regressions passed, and its graph check matched 50 cases and 598 orderings with zero
failed cases. GNU Stow and optional skills-ref are unavailable here; installer and
optional skills-ref checks are not qualified. No runtime dependency installation,
runtime/Docker rerun, extra-platform qualification or hosted CI is claimed.

Concurrent replacement and post-creation I/O errors still lack deterministic fault
injection. Source review is not an executed failure probe, atomic snapshot,
hostile-writer defense, filesystem provenance check or sandbox guarantee. The
preparer never verifies tool identity, live claims or historical author exposure.
Before any later dispatch, the coordinator must supply the actual builder, current
instructions/claim and exact trusted grants, verify the toolchain and formatter,
bind compiler/standard-library reads and check the exclusive exercise frontier.
The proposed Go build may read matching package files beyond a prior manifest.
Q_use, visual approval and real-use acceptance remain open.

No new model-builder candidate was written, compiled or executed. Repository tests
exercise their existing compiler/replay fixtures separately. Original pilot sources,
tests and evidence remain byte-preserved. Main was still 882b5002 and no open PR
was returned at the final source check. This additive branch remains local; no
publication, merge or deployment occurred. Committed coverage and selected-ref
full-history restoration are verified after the implementation commit.
