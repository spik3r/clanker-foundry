---
id: CF-consumer-shadow-v1
status: review
owner: Astra shared consumer surface
repositories: [spik3r/clanker-foundry]
depends_on: []
---

# Versioned read-only consumer surface

## Implementation claim

Astra owns this bounded producer slice on `codex/consumer-shadow-v1`, based on
main `6c5f72278b0765fa9ea4df6907e0f93e053edb5e`. The repository owner requested
one reusable implementation with thin consumers, rather than copies of generic
contracts and state transitions. Start with explicit read-only shadow inspection.
This public-safe task contains no consumer task records, private backlog or logs.

Inspected main, completed producer claims, and open PR5. That PR touches lifecycle
CI/docs, not the paths below. Prior runtime implementation claims reserve no
further editing. Recheck before publication; this observation is not a lock.
The exact prospective scope is:

- New root `package.json` and `package-lock.json`.
- New `runtime/src/consumer-shadow-v1.mjs`.
- New `runtime/test/consumer-shadow-v1.test.mjs`.
- New `docs/consumer-shadow-v1.md`.
- This task record and new immutable Project Memory source/event files only.

Do not edit existing runtime modules, schemas, helpers, workflows, claims or
historical evidence. Other contributors retain their ownership. Commit and
register this genuine canonical source, then record a start before maintained
implementation edits.

## Bounded outcome

Provide a private installable root package with only the versioned subpath
`@clanker-foundry/orchestration/shadow/v1`. Keep one implementation of contracts,
canonical hashing, source import and pure lifecycle guards by calling the existing
Foundry modules. Package their unchanged dependency closure, including the
Project Memory state helper at its original path, rather than copying it.

Use a narrow facade accepting bounded JSON text before existing canonicalization.
Return explicitly hypothetical shadow results, not admissions or authoritative
state. A reviewed-source preview must preserve the importer's genuine source,
exactly-one-claim, identity and pin checks. Transition inspection must delegate
to the existing evaluator, not duplicate its guards. Caller-supplied context
is not authenticated authority. Reject malformed, unknown, duplicate-key,
oversized/deep input and unsupported surface versions without side effects.

Package only exact required source/schema/helper/documentation paths; exclude
task records, evidence history, databases, credentials, test fixtures and all
operational modules. No store, runner, provider, CLI or dispatch entrypoint is
exported. Keep the root package private and dependency versions consistent with
the existing reviewed runtime lock. No registry publication is authorized.

## Acceptance and validation

- The actual packed inventory is bounded and public-safe, including automatic
  package metadata/README inclusion. No private inputs enter logs or fixtures.
- Install the packed artifact into a clean temporary consumer and exercise its
  public versioned subpath; no source checkout or runtime dependency-directory
  fallback may mask packaging defects.
- Unsupported subpaths refuse. Import performs no filesystem writes, network,
  provider, subprocess, database or worker activity. Package exports are an API
  boundary, not an OS sandbox.
- Invented fixtures prove facade/direct shared-engine result and error parity,
  deterministic replay, unchanged inputs, source/claim/pin mismatches, rejected
  transitions and bounded parsing. Do not fabricate a claim to pass an importer.
- Run focused tests, full Node24 runtime and Project Memory/workflow regressions,
  pack validation, whitespace and genuine full-diff committed lifecycle checks.
  Distinguish missing optional/local-host checks and hosted CI from observed passes.
- Obtain fresh-context independent exact-diff review before final draft handoff.

## Integration and exclusions

Consumers pin the reviewed producer commit/artifact and retain their canonical
repository, claim, policy, source and security gates. Shadow output cannot replace
current ownership or activate another namespace. Do not silently synchronize a
consumer's distinct lifecycle helper or copy its private source into this producer.

No runtime namespace, authority database, persistent access, provider, service,
automatic dispatch, managed writable execution, existing task migration, deployment,
merge or permission/policy change is included. Source draft publication is the
intended review handoff after verification; runtime/consumer activation remains a
separate explicitly authorized decision.

## Local implementation and observed validation

The narrow facade exports two JSON-text inspection functions with explicit
shadow-only/authority-none results. Existing importer and evaluator implementations,
schemas, helpers and operational modules remain byte-for-byte unchanged. The
root private package includes their exact pure dependency closure at original
paths; it exposes no store, runner, provider, CLI or dispatcher.

Cloud Linux Node 24.22.1 passed all 319 runtime and Project Memory/workflow
regressions, zero failures or skips. All eight new tests passed independently,
including an actual tarball packed from the producer and installed with npm ci
into a clean temporary consumer using the pinned dependency cache in offline
mode. Tests checked actual package inventory, non-symlink installation, unsupported
subpaths, inert import, bounded parsing and direct shared-engine result/error parity.
Syntax, pack validator, relative links and whitespace checks passed.

The optional skills-ref tool is unavailable. Installer regression reached the
missing GNU Stow prerequisite on this cloud host; it is not reported as passing.
No actual Docker or live provider test was run because this slice changes no
operational module and exposes neither capability. Hosted CI is separate and
pending final source publication.

Independent exact-source review found no material facade/packaging defect and
confirmed the eight tests. Documentation explicitly preserves unknown inventory,
external/security and scope refusals; simulated PENDING is not conflict-free or
claimed. Caught errors can contain private input and must remain private.

The guide includes a boundary diagram and staged next modeling slice: versioned
transition/guard trace manifest, contract/test links, generated diagrams and
executable replay fixtures. Diagrams remain explanatory until their semantics
are reviewed and machine-checked; no standards-conformance claim is made.

Implementation edits are finished for review. Next: record genuine completion,
validate the committed full diff, publish the reviewed draft source and check its
exact-head CI. The owner must review a pinned consumer adapter separately; no
consumer adoption, namespace activation, merge or deployment is claimed here.
