# Independently frozen proposal-replay literals

These eight files are invented acceptance inputs and complete expected reports for
the first-proposal replay bridge. They are not production artifacts, approvals,
authenticated actor identities, or evidence that any replay has run.

## Requirements written before execution

**ParcelSeal (A):** Packed moves to terminal Sealed on seal / T_seal only when
G_actor (nonblank actor), G_outcome (nonblank outcome), and G_revision (exact
revision) all hold. Success assigns the actor only when owner is null, then
changes state and increments revision once. It preserves ID, outcome and ordered
acceptance. There is deliberately no acceptance or owner/actor equality guard.

The four cases require: nil-owner assignment followed by terminal refusal;
present-empty owner and empty-list preservation; all failed guards sorted,
followed by a narrower refusal without advancement; and MaxUint64 refusal before
effects. Null acceptance remains null, and present empty owner is not null.

**CuratorRelease (B):** Queued moves to terminal Released on release / T_release
only when G_acceptance (nonempty criteria), G_curator (actor exactly curator),
and G_revision hold. Only state and revision change. There is deliberately no
outcome guard, owner assignment, or owner/actor equality rule.

The four cases require: exact arithmetic above 2^53, duplicate criterion order
and nil-owner preservation followed by terminal refusal; preservation of a
different present owner and whitespace outcome; sorted multi-guard refusal then
a narrower refusal retaining null acceptance; and unknown-trigger INVALID_INPUT
with null result task followed by success from the unchanged initial task.

Selected guard and revision-limit refusals retain the selected transition ID.
Terminal known-trigger refusal has an empty transition ID and unchanged task.
Invalid input has an empty transition ID and null returned task. A matched
refusal does not advance the scenario's current task. These follow the baseline
profile contract, not a new bridge policy.

## Exact identities and report use

Each JSON file is compact UTF-8 with one final LF. Revisions are decimal strings,
never JavaScript numbers. All keys, nulls, empty arrays, ordered criteria and
Unicode text are intentional. Source hashes include the final LF.

expected.json contains both full legacy lifecycle-scenario-report/v1 objects,
not a projection or just success counts. Every entry's input, command, expected
result, complete actual result envelope, matched flag and empty differences
array is specified. Model A has four cases / six entries; B has four / seven.
All entries are expected matches, including refused transitions. SourceDigest
is the SHA-256 of the original suite bytes. reportSourceSha256 hashes the compact
report object in its supplied field order plus LF. There is no bound-runner
identitySource field and no invented proposal-wrapper receipt expectation.

A model digest hashes normalized canonical model JSON: sorted map keys and
ID collections, sorted transition guards, unchanged effects, no final LF.
The source guard lists are intentionally reversed relative to canonical order.
A scenario digest hashes the documented fixed-order semantic object, excluding
formatVersion and without final LF. Manifests bind the exact suite source hash.
Cross-pairing these model/suite/manifest sets must not become a candidate replay;
the bridge's separate tests must demonstrate refusal and zero process launches.

## Authorship, freeze and limits

The independent fixture author is /root/review_current_prs, a separate assistant
lane from the bridge implementer. The author selected these policies and literal
tasks/results before bridge implementation existed, using baseline public
contracts at 28739a21b126dfc0a793622ce46593ec7559d379:
model/ir.go; profilecatalog/README.md; jsonformat/json.go; canonical/canonical.go;
scenarios README/types/decode/report_decode and report-encoder shapes;
acceptanceauthor/README.md; the lifecycle README's runtime/refusal contract; and
the first-design brief's documented transition-ID refusal rules.

No compiler generation, reducer Transition, scenarios Run/Execute, modelplay or
proposal replay was called to derive these expectations. No existing generated
report or behavior fixture was copied. In-memory construction expanded the
written literal outcomes into the report DTO; it did not interpret guards or
invoke a reducer. The author had no usable native executor. The implementation
owner may mechanically materialize the author's exact frozen base64 payloads;
such copying is not an independent runtime rerun by the author.

Source and semantic hashes were computed independently in a pure JavaScript
SHA-256 implementation over UTF-8, checked against the standard empty, abc and
abcdbcdecdefdefgefghfghighijhijkijkljklmklmnlmnomnopnopq vectors. Canonical
serialization was constructed from the published contracts, not generated
output. Git blob IDs use SHA-1 over the Git blob header and exact bytes.
Subsequent strict Decode/Digest/check-only validation may confirm this metadata
with a panic-only callback; it must not regenerate behavioral expectations.
Record any discrepancy explicitly rather than quietly refreshing literals.

This freeze establishes separately authored expectations, not mathematical
proof of model correctness, human review, or independence from all prior project
context. Actual candidate replay, independent source review, mutation checks,
toolchain qualification and integration remain later evidence. Suggested
behavioral controls include removing A's owner assignment, ignoring B's curator
guard, changing revision advancement, and replaying the wrong candidate model.
Keep failing observations and any justified future oracle amendment historical.
No denied PR20 artifact, private consumer, deployment, worker or remote action
is involved.
