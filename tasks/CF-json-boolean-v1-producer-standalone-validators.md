---
id: CF-json-boolean-v1-producer
status: done
owner: Foundry JSON validator implementation lane
repositories: [spik3r/clanker-foundry]
depends_on: []
---

# Generate standalone JSON boolean validators

## Implementation claim

The repository owner approved publication of this bounded producer draft.
This is a proposed producer claim, not a merged canonical task, consumer adoption
or permission to merge or deploy. This remote source commit precedes fresh
implementation events bound to this reviewed claim. Earlier local preparation
remains separate; its unpublished commits and events are not represented as
remote history.

The implementation lane owns this isolated frontier, subject to independent
exact-source review before merge:

- `modeling/jsonvalidator/go.mod`
- `modeling/jsonvalidator/toolchain.json`
- `modeling/jsonvalidator/README.md`
- `modeling/jsonvalidator/PROFILE.md`
- `modeling/jsonvalidator/cmd/jsonvalidator/main.go`
- `modeling/jsonvalidator/cmd/jsonvalidator/main_test.go`
- `modeling/jsonvalidator/cmd/jsonvalidator/profile_test.go`
- `modeling/jsonvalidator/testdata/sample/model.json`
- `modeling/jsonvalidator/testdata/sample/cases.json`
- `modeling/jsonvalidator/testdata/sample/validator.go`
- `modeling/jsonvalidator/testdata/sample/validator.ts`
- `modeling/jsonvalidator/verify/acceptance_test.go`
- `modeling/jsonvalidator/verify/typescript.mjs`
- `.github/workflows/json-validator-go.yml`
- this task record and its new immutable Project Memory source/event files.

Current main and the visible open PR inventory were checked before this draft
claim. No competing claim on this new module or new workflow was found. Recheck
current ownership and pending paths before publication. The inspection is not a
lock or knowledge of unpublished work.

Do not edit the existing lifecycle module, runtime, root npm package or lockfile,
shared instructions, setup templates, existing workflows, other claims or prior
evidence. No consumer source, model, corpus, private task reference or raw private
log belongs in this public producer. No managed dispatch or consumer claim is
created by this task.

## Scope and interface

Build a standard-library Go compiler in a separate module. Its proposed profile
is `json-boolean/v1`, distinct from `lifecycle/v1`. The CLI accepts a model and
a new absent output directory. Outputs are standalone Go package `validator`
with `Validate(any) bool`, and TypeScript `validate(unknown): boolean`. Both
include the exact input model SHA-256. There is no consumer-specific function
name, runtime npm dependency, network request or executable model content.

The profile supports strings, booleans with optional boolean constants, arrays
and objects. Declared properties explicitly state required or optional presence;
extra properties are allowed and not traversed. The input contract is
already-parsed ordinary JSON. No first-error API, type guard, coercion, defaults,
numeric constraints, unions, recursion, nullability or cross-field semantics are
introduced. Strict compiler intake limits and refusal behavior remain documented.

The implementation is adapted from reviewed generic scratch compiler mechanics.
Public fixtures are synthetic and consumer-independent. New fixture expectations
must be reviewed independently before calling the producer accepted.

## Acceptance

- The generic profile and fixed exports are documented and independently reviewed.
- Go and TypeScript agree with the frozen synthetic acceptance cases and the
  property-presence edge cases, including declared `__proto__` names.
- Unsupported model shapes, duplicate keys, invalid field names and resource
  limits refuse without emitted source.
- Repeated generation is byte-identical and matches checked golden output.
  Existing output directories and symlinks are refused without overwriting.
- Genuine Go test, race and vet checks, TypeScript execution and the repository
  validator pass against the final source.
- Real source/start/completion evidence covers the committed diff and preserves
  existing Project Memory bytes.
- Public-source review finds no private consumer material. Hosted CI and final
  publication status are reported separately; local success is not merge or
  deployment approval.

## Validation

Use Go 1.27.3 and Node 24. Run from this module:

- `GOTOOLCHAIN=local GOPROXY=off go test -race -count=1 ./...`
- `GOTOOLCHAIN=local GOPROXY=off go vet ./...`
- `node verify/typescript.mjs`

Run `bash scripts/validate.sh` at the repository root. Stage only the named
frontier, record the actual checks with Project Memory, commit the new immutable
events, then validate against the actual `origin/main` with code/source identity
`spik3r/clanker-foundry`. Any unavailable check remains explicit.

## Integration

This producer has no consumer adoption or production activation gate within its
claim. A consumer must separately own its model, immutable producer pin,
generated-file provenance, API integration, regeneration/drift checks and actual
application/host acceptance. The root shadow npm package is unchanged.

Publication is authorized as a draft proposal only. Independent exact-source
review, actual hosted checks and all applicable approvals precede merge. No
deployment is authorized by this task.

## Integrated producer outcome (2026-11-09)

The preceding claim records the original draft-only scope at source
94908bab8edb672af4119abfe604574322c48714. It remains historical evidence,
not retroactive merge authority. Later independent review, hosted acceptance
and separate approval preceded the normal integration of
[PR94](https://github.com/spik3r/clanker-foundry/pull/94).

- Accepted head: e2b682c749f9e75e02c4d9f2d2b777f947335cca.
- Actual merge: 6606c5d073815cc19f4745d0f280e4ff8a1aa78f, at
  2026-11-09T04:27:05Z.
- Ordered parents: a43f03b75a58f6c4636f6482f189f9d946b517c8, then the
  accepted head.
- Integrated tree: 7268a25bdfe7b2e065fcb8af0414b2b88c2c4274, identical
  to the independently accepted head and current-base PR test tree.

All twelve accepted-head/current-base workflows and sixteen jobs passed before
merge. All six actual-main workflows and eight jobs then passed on the actual
merge above. The [JSON validator run](https://github.com/spik3r/clanker-foundry/actions/runs/42561897046)
checked Go 1.27.3 formatting, race tests and vet, all 46 saved synthetic
TypeScript cases, and generated Go/TypeScript byte equality.
[Repository validation](https://github.com/spik3r/clanker-foundry/actions/runs/42561897033)
passed the pack, path, installer/shared-base and lifecycle/workflow checks.
The PR-only committed-evidence step passed before merge against the actual PR
base; its skipped main-push step is not counted as another execution.

A read-only check at later main
b5c6c8760fb4f9778480f4b5ab73912b5fb99322, tree
bda6f5cfc0733a93dc4d057ec896963130b4f203, found every final producer
file and all fifteen completion fingerprints unchanged. Original source,
implementation commits, six lifecycle events and the registered snapshot remain
intact. These are source-preservation checks, not new behavioral executions.

Only this completed standalone producer outcome is released when this canonical
closeout is integrated. The original owner remains recorded as provenance.
No other task, runtime authority or consumer claim is released. Consumer models,
immutable pins, generated provenance, application/host acceptance and activation
remain separately owned and qualified.

Optional skills-ref was unavailable in the retained producer validation. Node
execution is not browser or consumer acceptance; CLI and emitted Go child
processes were not race-instrumented. The finite synthetic and separately
reported compiler checks do not prove universal language equivalence or native
client/provider acceptance. No deployment or new producer release is asserted.
