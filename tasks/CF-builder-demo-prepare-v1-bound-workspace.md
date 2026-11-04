---
id: CF-builder-demo-prepare-v1
status: in-progress
owner: Foundry builder demo preparation contributor
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
- modeling/lifecycle/builderdemo/README.md
- modeling/lifecycle/cmd/modelbuilderdemo/main.go
- modeling/lifecycle/cmd/modelbuilderdemo/main_test.go
- examples/model-builder-go/preparation.md
- tasks/CF-builder-demo-prepare-v1-bound-workspace.md
- new immutable source/event files under this task ID

No changes to prior models, packet, acceptance, evidence, workflow, compiler,
runtime, preflight, existing CLI routes or CI. No extra-platform work, dependency
installation, candidate implementation/execution, publication, merge or deployment.

## Ownership check

GitHub main, open pull requests and builder/handoff branches were checked on
4 November 2026. Main is the commit above; no open PR was returned. The current
claims were inspected. The coordinator reported no verified active owner for
this additive preparation lane and accepted this boundary before edits. Historical
status labels and branch presence do not prove active ownership. This local claim
does not acquire an external lock; recheck before publication.

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
