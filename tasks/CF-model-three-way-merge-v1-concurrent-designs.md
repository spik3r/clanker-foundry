---
id: CF-model-three-way-merge-v1
status: in-progress
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
