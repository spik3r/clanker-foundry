---
id: CF-model-three-way-merge-v1
status: review
owner: Byte model merge contributor
repositories: [spik3r/clanker-foundry]
---

# Reconcile three reviewed lifecycle model versions

## Implementation claim

The coordinator assigned this independent local lane on 2026-11-02 after reserving
scenario coverage for another contributor. Main was independently read at
`55cc43870ad8017ecb334536a269e6ac9c49afee`, tree
`3bab0a7b8dc5944e23c3708352ceb890c7e2102d`. All public task records, the live
branch inventory and open PRs were checked. No model-merge reservation was found.
PRs #20 and #5 remain excluded. Existing task statuses and claims do not transfer.
Held compiler-profile acceptance and guide work remain separate. This isolated
worktree uses branch `codex/model-three-way-merge-v1` from that exact main.
This is a local claim, not a shared reservation or published task.

Byte owns the package, wire carrier, CLI and integration. An independent author
will freeze literal cases and expected models/conflicts before merge logic is
implemented. A separate reviewer checks exact final source and validation.
Commit/register this prospective source and record a genuine start before edits.

Exact maintained frontier under `modeling/lifecycle/`:
- `modelmerge/types.go`
- `modelmerge/merge.go`
- `modelmerge/decode.go`
- `modelmerge/merge_test.go`
- `modelmerge/decode_test.go`
- `modelmerge/oracle_test.go`
- `modelmerge/README.md`
- `cmd/modelmerge/main.go`
- `cmd/modelmerge/main_test.go`

Also this task and its own new immutable Project Memory source/events only.
No existing model, compiler, reducer, modelpatch, scenario coverage, runner,
acceptance fixture, profile corpus, runtime, workflow, toolchain, dependency,
standard projection, visual artifact, guide or other task/evidence changes.
No access to denied artifacts, publication, merge, deployment or network action.

## Contract

Add a deterministic three-way merge of complete lifecycle/v1 models. This is an
inspectable proposal, not automatic acceptance, reference repair or code execution.
`Merge(base, left, right model.Model) (Result, error)` validates and normalizes all
three through existing model APIs, requires each to be representable by existing
JSON-v1, and computes existing canonical digests. Input validation is left-to-right
base/left/right, before any conflict resolution. Invalid inputs return zero Result
and fixed `ErrInvalid`; no caller text is copied into errors.

Merge root `id` and `initial` independently. Merge each state, requirement, guard
and transition as one atomic record addressed by (kind, ID). Compare full normalized
values, including presence. If left equals right choose it; else if left equals
base choose right; else if right equals base choose left; otherwise record a conflict.
Do not merge disjoint fields within one record, infer renames, merge text or ASTs,
repair references, reorder effects, or treat equivalent-looking expressions as equal.
Existing Normalize alone determines semantic set order. An absent record differs
from a present empty value. Symmetric left/right choices must produce the same
candidate or conflict locations.

Conflicts contain only kind, ID, field and reason, ordered by kind then ID then
field. Root conflicts use kind model, empty ID, field id or initial. Other fields
are empty. Reasons are `add-add` when base is absent, `remove-change` when one side
is absent, and `change-change` otherwise. A conflicted result contains every
conflict and no partial candidate or candidate digest.

Without conflicts, existing whole-model validation and JSON-v1 size/representation
checks run on the private candidate. Individually valid branches can form an
invalid combination: duplicate global IDs, dispatch collision, dangling references,
limits or reachability, for example. Report status `invalid-combination` with no
candidate/digest and no conflicts; do not silently discard an edit or expose a
partially invalid model. On success report `merged` with an owned normalized
candidate and its canonical digest. No-op success still grants no approval.
All outputs own mutable data; callers and immutable concurrent inputs are untouched.

Result identifies `lifecycle-model-merge/v1`, `lifecycle/v1`, `example-only`,
`authority=none`, base/left/right canonical digests, status, candidate/digest or
null, and an explicit conflict array. Typed source digests are null. There is no
public arbitrary-report encoder that could claim unverified observations.

## Wire and CLI

`Execute(raw []byte) ([]byte, error)` reads exactly formatVersion
`lifecycle-model-merge-json/v1`, base, left and right. Each embedded value is a full
existing JSON-v1 model. Reuse strictjson and jsonformat; reject duplicate, unknown,
case-variant, missing, null, malformed Unicode and trailing data. No default inputs.
Whole input is at most 4 MiB/depth66; each embedded model keeps its existing 1 MiB
limit. Typed input keeps existing bounded IR and the same JSON representability
limit. Output is capped at 2 MiB. Planner-owned input/output limits return fixed
ErrLimit; malformed/unsupported/invalid models return fixed ErrInvalid.

Wire output recomputes the result and records SHA-256 of exact request and embedded
base/left/right bytes, separate from normalized identities. It includes candidate
as a complete normalized existing JSON-v1 document or null. Output is fixed-order
JSON with HTML escaping off and one final LF. Identical request bytes produce
identical output; source formatting can change source hashes without changing
canonical digests or merge decisions. Hashes prove consistency, not authenticity.

`modelmerge` accepts no arguments or paths. It reads bounded stdin and buffers
output before any write. Exit 0 covers all valid reports, including conflicts and
invalid combinations; callers must inspect status. Exit 2 means rejected/read
failure and empty stdout; exit 3 means failed/short stdout write. Handle real
closed pipes as write errors. No writes, subprocesses, fetching, code generation,
scenario execution, evidence rebinding or automatic continuation is performed.

## Acceptance and verification

Freeze independently authored literal cases before merge logic: no-op, one-sided,
identical two-sided, independent records/root fields, set-order normalization,
conflicting replacements/additions/deletion versus change, coordinated removals,
semantic-invalid combinations and both-side symmetry. Test whole candidate and
complete conflict list, not only status. Preserve independent expected bytes.
Own tests cover invalid inputs, cyclic/deep expressions, source/output limits,
strict JSON keys/Unicode/nulls, determinism, no aliases, parallel immutable use,
CLI arguments/read/short-write/closed-pipe behavior and no partial output.

Run actual Go 1.27.3 offline: focused tests, full race suite, vet, modelc check;
gofmt; repository validator; Git whitespace checks; committed full-diff Project
Memory validation against verified main. Require independent exact-source review.
Use disposable mutants which compile and fail relevant frozen assertions; do not
count compilation errors as behavioral kills. Record unavailable checks honestly.
Completion remains a local handoff; publication requires separate authorization.


## Local result and independent review

Claim source `7e580a426b905a6707b727986dce52159068cb22` and the genuine start
preceded maintained edits. API shapes were committed separately, then an independent
author froze fourteen invented model/conflict vectors at `30fa851`. The unchanged
oracle SHA-256 is
`00c46bc02a46795b36c0c56383f7a640c0d6145bf73fcf303da68e2354489cbd`.
Candidate digest expectations were independently computed with Python JSON/SHA-256,
not the merger. Fixture self-checks ran before merge implementation, using a
panic-only stub which those self-checks never invoked.

The additive typed merger and stdin CLI are implemented. Independent tests compare
whole literal candidates and all conflicts in both branch orientations. Three
compiling disposable mutants (discard changed right, misclassify conflicts, accept
an invalid combination) fail the unchanged behavioral oracle. This is finite
assistant-authored acceptance evidence, not human approval or exhaustive proof.

Exact-source review found an output-policy defect in `a5e01431`: inherited raw JSON
preserved HTML escapes in expression literals. Repair `d8dde7519a66c9cca19b42afceea8da04e52998e`
decodes only the existing adapter's validated output and re-encodes the complete
envelope; it does not change the shared adapter or apply unsafe byte replacement.
The reviewer's unchanged `<tag>&` probe and the new Unicode/literal-backslash
roundtrip test pass. Independent re-review of that exact source found no remaining
blocker; its bounded focused race run passed. The frozen oracle did not change.

Final observed checks use actual Go 1.27.3, disabled module network and ambient Go
workspace/config/flags, an authorized workspace cache/temp directory, GOMAXPROCS=2:
- `go test -p 1 -race -count=1 ./...` passes on the branch: 49 tested packages;
  modelc has no test files. `go vet -p 1 ./...` and `go run ./cmd/modelc check` pass.
- The same full race/vet/drift checks pass in a disposable current-main union:
  50 tested packages plus modelc. Current main was read at
  `1c57dd3f6b6489f6f690917366e8e803ad610076`, tree
  `feb030b57a23db4ec53984414ebedee81a54802f`. The union used local `03cd2bb`, whose
  complete tree exactly equals that verified main tree, and applied this branch's
  non-overlapping diff. The maintained branch was not rebased or published.
- The actual stdin CLI no-op command reports merged with the unchanged canonical
  digest and authority none. All three valid CLI statuses, input failures, short
  writes and a real closed stdout pipe are tested.
- Repository validator, gofmt and whitespace checks pass. Optional skills-ref is
  unavailable. All 143 Project Memory regressions pass. Full committed-diff
  lifecycle coverage is validated at closeout against the reviewed base.
- Every pre-existing base path is byte-preserved; only the claimed new files and
  this task/evidence are changed. No private data or new dependencies were used.

Environmental failures were not counted as passes. The default home Go cache was
read-only; a writable cache was selected. A full build initially exhausted /tmp;
only this lane's own cache was moved to the workspace. Its external temporary
build directory then needed a valid local Git repository for unchanged VCS-stamping
checks. Later concurrent aggregate runs under host resource pressure hit several
existing and new closed-pipe deadlines. Those runs failed; the unchanged full
suites and independent focused race subsequently passed with bounded serial
execution. No check, deadline, policy or expected result was weakened.

The task remains a local review handoff. No push, PR, merge, deployment, worker
activation, scenario execution or evidence rebinding was performed by this tool or
lane. #20, #5, standard projections and other ownership remain untouched. Publication
and integration require the coordinator's applicable authority and fresh checks.
