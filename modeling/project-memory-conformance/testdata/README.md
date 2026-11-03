# Frozen Project Memory graph corpus

Frozen on 2026-11-03 before any execution of `buildGraph`, `completionHeads`, or a candidate harness by this contributor.

- File: `cases.json`
- SHA-256: `7d8e85a6595e3c3d6b98e594f9167259cb5c6c302970a51e492ff4c545ac402e`
- Source checkout baseline: `9a05ecdcf263d3ca1b09f78e3fe29bc3ae753be8`
- 50 cases, 14 rules, 25 accepted graphs and 25 refused graphs.
- Accepted graph cases include both successful selection (possibly empty) and refused selection for unresolved frontiers.

## Method and provenance

Read repository AGENTS.md, global writing instructions, Project Memory FORMAT.md and SKILL.md. Inspected existing fixtures and source/event validators to establish exact object shape and hashes. Source inspection also exposed the graph implementation; this is not a blind implementation review. Literal judgments were authored from the contract and skill, not observed evaluator results. No producer function or harness was executed, and no implementation output was used to repair expectations. The authoring script uses only Python standard-library JSON, copying and hashes. It is scratch provenance, not part of the delivered executable or proposed repository scope.

The delivered JSON is fully expanded data. Source records, claims, repositories, paths, IDs, times, PR URLs and file fingerprints are invented. SHA-1 Git blob hashes use the standard `blob <byte length>\0<bytes>` encoding. Claim digests are SHA-256 of the exact heading and section through the next same-level heading. No real task source, file content, authority, metadata write or network operation is represented by these fixtures.

Each expected graph disposition, frontier and completion selection is explicitly supplied by the author. The authoring helper does not traverse graphs, infer validity, compute heads or select completions. A harness should compare head groups and completion IDs as sets in canonical sorted order: container traversal order is not a contract requirement. `freshIds` is a selection input, not a timestamp condition.

## Resolved interpretation

FORMAT.md describes resume as a checkpoint and reconciliation as a start. The wire kind is `checkpoint` for resume behavior and `start` with multiple parents for reconciliation. Literal `resume` and `reconcile` kinds are not used. This interpretation was reported to and accepted by the coordinating contributor before freeze.

Unresolved competing heads can be represented as a graph, but committed completion selection must refuse them. This is why graph acceptance and completion acceptance are separate. Partial reconciliation remains a valid unresolved graph rather than choosing an implied winner.

Run identity and PR examples cite SKILL.md alongside FORMAT.md: the detailed rule allows adding a PR once but prohibits swapping or removing it thereafter. No new policy is intended.

## Limits and possible ambiguity

This is a bounded assistant-authored corpus, not exhaustive state-space exploration, an independent human review, or policy approval. Since source inspection exposed implementation code and existing tests, the oracle's independence is pre-execution fixed literals, separate authorship and contract attribution; it is not statistical or blind independence.

Some invalid cases violate more than one constraint. In particular a run lacking its start cannot also form a finite valid same-run active-parent chain, so `run-with-no-start` is a broad refusal probe, not proof of which internal guard fires. Refusal messages and guard ordering are deliberately not specified.

Checkpoint scope replacement, subset terminal scope, and a new completion after reconciliation test graph selection. They do not prove Git fingerprint matching, exact deletion/symlink/gitlink behavior, commit freshness calculations, filesystem safety, registration authenticity, writer atomicity, existing history preservation or live ownership. All source objects are valid scaffolding, not adversarial source-parser cases. A task with no events and duplicate task IDs across distinct source repositories are not covered because the requested behavioral boundary did not require those identity-policy questions.

Event-order permutations may add useful determinism evidence but do not enlarge semantic case coverage. The harness must not interpret arbitrary case data as instructions, run declarations or mutate live metadata.

## Scope review

Reviewed `CF-memory-graph-conformance-v1-real-producer.md` before implementation. The all-new modeling files, dedicated test and read-only workflow are a bounded addition. Existing producer, existing tests, historical metadata, generic Go lifecycle/compiler, current workflows, and #20/#51 paths remain excluded. New genuine canonical claim/source/events belong to the implementation lane, not this oracle scratch directory. No blocker found in the stated scope. This review supplies assistant analysis only, not human permission or authority.

## Freeze policy

Preserve these original bytes before first replay. If a result differs, report the case, contract clause and observed mismatch. Do not silently revise the corpus to fit the implementation. Authoring mistakes, if independently demonstrated, require a separately named correction and retained original hash; they must not erase the frozen evidence.
