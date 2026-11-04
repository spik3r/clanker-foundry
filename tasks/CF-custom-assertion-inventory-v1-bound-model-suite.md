---
id: CF-custom-assertion-inventory-v1
status: review
owner: Foundry custom assertion inventory contributor
repositories: [spik3r/clanker-foundry]
---

# Inventory supplied assertions for a custom model and exact suite

## Implementation claim

The coordinator selected this bounded modeling lane at main
f16c94d456e89b7ee0680b91d13a44fb79be5556. Existing scenario assertion inventory
uses the built-in model; custom change and merge rehearsals now produce reports
for other accepted models. Add an offline inventory bound to the original custom
model, suite and report, so authors can inspect absent transition/guard assertion
references. Empty buckets do not establish missing behavioral coverage.

The implementation owner uses isolated branch codex/custom-assertion-inventory-v1.
Live PR search found only PR79, the model-to-builder guidance/pilot lane, with no
write-path overlap. Existing historical inventory records do not reserve this new
package. Recheck live work before publication and stop on conflicting ownership.
The coordinator owns review and merge; this contributor does not merge.

Exact maintained scope:

- modeling/lifecycle/boundinventory/check.go
- modeling/lifecycle/boundinventory/report.go
- modeling/lifecycle/boundinventory/check_test.go
- modeling/lifecycle/boundinventory/contract_test.go
- modeling/lifecycle/boundinventory/testdata/observations.json
- modeling/lifecycle/boundinventory/README.md
- modeling/lifecycle/cmd/modelboundinventory/main.go
- modeling/lifecycle/cmd/modelboundinventory/main_test.go
- This task and its new immutable Project Memory source/events.

Do not edit old inventory, explanation, scenario runner, trace, proposal or
qualification sources, old tests/fixtures, compiler/evaluator, operator routes,
source pins, runtime, CI, global instructions, held artifacts or other tasks.
No filesystem access, processes, providers, reducer execution or workspace manager
belongs in the package. The standalone command reads stdin and writes stdout only.

## Request and native boundaries

Expose Check(raw []byte) ([]byte,error). Require a strict JSON object with exactly
formatVersion lifecycle-bound-assertion-inventory-request/v1 and model, suite,
report fields. Each source is its original UTF-8 JSON text string, not a nested
object or path; null, unknown, missing, duplicate and case-variant keys refuse.

Use native model decoding, normalization and canonical identity. Create an
explicit validation-only native Runner for that model with a callback that panics
if called. Runner.Decode, Digest and DecodeReport supply structural/result checks;
ExplainWithRunner validates the entire report and declared transition/guard links
before inventory mapping. Do not implement another reducer, result comparison,
model validator or guard evaluator.

Require native report SourceDigest nonnull and equal to the exact suite byte hash,
with canonical model and scenario digests matching those inputs. A narrow private
suite/report binder checks case/step order, complete original commands, expected
literals and input flow using the already decoded native values. Existing full
binders are private; do not fabricate trace/catalog requests or change native APIs
to reuse them. Accept only a complete matching report or a valid fail-fast prefix
ending at its first mismatch. Matched refusals retain the prior input, including
an actual null task. Reject altered literals/commands/input, missing/reordered/
extra entries, wrong identities and dishonest completion counts before output.

## Inventory and failures

Emit lifecycle-bound-assertion-inventory/v1, mode example-only, authority none,
executionStatus and reviewStatus not-established, and status caller-supplied-
assertions. Record exact model/suite/report source SHA256, canonical model and
scenario digests, and native matched/completed-case/step observations.

List every normalized transition and each associated guard in sorted native ID
order, even without references. Only matched report entries contribute successful
transition, explicitly named guard-refusal and transition revision-limit references.
List matched invalid-input and illegal-transition assertions separately. Mismatched
entries are excluded, never mapped as successful or refusal assertions. References
carry case/step IDs and zero-based report entry position in report order. Empty
collections remain arrays. No guard-pass inference, requirement satisfaction,
reachability, provenance authentication, approval or coverage percentage is added.

Wrapper bound is 16 MiB; native model and suite each retain 1 MiB bounds and report
8 MiB. Output is fully buffered and bounded at 8 MiB. Native stricter limits still
apply. ErrLimit identifies explicit byte/output bounds; ErrInvalid covers malformed
sources, native structural refusal or unmappable/inconsistent observations;
ErrBinding identifies otherwise valid identity assertion disagreement; ErrBoundary
covers unexpected native/encoding failures. Errors return no partial output and
use fixed diagnostics without source payloads.

The argument-free CLI accepts bounded stdin, writes the complete result once and
exits 0 for a produced report, 2 for initial input/binding/limit refusal, 3 for short
or failed stdout, and 4 for unexpected boundary or stdin read failure. Ignore
SIGPIPE so closed stdout reports exit 3. A report can contain a valid mismatch;
exit 0 means a completed inventory, not passing replay or reviewed behavior.

## Independent acceptance and completion

Publish this committed public-safe claim and receive coordinator approval before
registering its pinned source and recording genuine Project Memory start. A
separate contributor must freeze literal custom-model suites, supplied reports
and expected inventories before production edits. Label fabricated observations
as consistency fixtures rather than executed evidence. Preserve their bytes.

Independent vectors cover a nondefault model, matched success, multi-guard and
fixed refusals, absent transition/guard buckets, a valid early mismatch prefix,
wrong source/canonical/scenario identities, altered literals/commands/input,
refusal-retains-input continuity, ordering/count tampering and undeclared links.
Builder checks add strict/bound controls, deterministic output, concurrent calls,
owned outputs, no input mutation and actual compiled binary stdout/closed-pipe
behavior. All validation paths must remain nonexecuting.

Run Go1.27.3 focused and full race, vet, generated-artifact drift, gofmt, pack,
installer, whitespace and full committed Project Memory validation. Observe
critical mapping/binding mutants in a separate disposable clone. Obtain both
coordinator and independent exact-source review, publish final evidence, and
hand off hosted checks and merge. No observed result establishes universal
coverage, an execution origin or human approval.


## Implementation result

Product commit c77b38d adds only the new boundinventory package, standalone
modelboundinventory command and their tests/README. Reviewed source
eeb939018ffaf130bb7d6a8ad933fcb26b8e061f includes ordinary main merges through
4a52cf19a026e8f442d13031135a3e9e10bee8fa. Those merges add no changes to the
covered product bytes. Existing inventory/native APIs, qualified source pins,
old tests/fixtures and foundry routes remain unchanged.

Independent fixture commit fedd6553fe9ee66e1da06d3160583f6dbad7125e precedes
production. Its contract and observations hashes remain 02d9e6bfd1727600575771921599216d57b217b64eb01a03de192a25d4a64c05
and b8696125353e92d915f6f7dee7b0d1674f03b513ad660f174dca4b306ff8e563.
The preproduction native-only race control compiled without future Check and
passed before production. Its observations are fabricated consistency fixtures,
not execution evidence. Coordinator and separate exact-source reviews are clean;
both independently ran the new package and actual CLI race checks.

Go1.27.3 full race passed on the unchanged reviewed source: 81 tested packages
and two packages without tests. The checkout remained frozen throughout that
run. Focused race, vet, modelc generated-artifact check, gofmt, pack, installer and
whitespace checks passed. Optional skills-ref was unavailable and was explicitly
skipped by the validator. Full committed Project Memory validation follows the
completion commit; final hosted checks and merge belong to the coordinator.

Four separate disposable-clone mutations compiled and failed frozen behavioral
controls: dropping source/scenario binding, skipping expected literal comparison,
including a mismatched entry in buckets, and omitting the unused transition.
Failures identified incorrect output or accepted tampering, not setup or panic.
Actual compiled CLI output matched all independent fields for the full supplied
five-case/seven-entry report and both mismatch prefixes. All three exited zero
because each produced a complete inventory; supplied matched/count facts do not
establish that those fabricated reports came from execution. Closed stdout
returned exit 3. Output includes the unused transition and empty guard buckets.

Native decoding and model-binding errors remain ErrInvalid. Only successfully
native-decoded source/scenario identity disagreements, including null source
identity, become ErrBinding. Entire-report explanation precedes matched-only
mapping, so undeclared actual links are rejected even at a mismatch, while a
native-valid undeclared expected ID remains an excluded expectation.

The next step is final hosted checks and coordinator merge. Empty reference
buckets are not coverage conclusions, and hashes prove neither provenance nor
approval. This contributor has not merged or assigned another product task.
