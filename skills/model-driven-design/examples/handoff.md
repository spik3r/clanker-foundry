# Worked first-design handoff: submit a draft

**Proposal only.** This is an invented example, not an approved product or a record
of a real person's authorization. It ends at human review; its suite is unexecuted.

## What is proposed

The [brief](brief.md) defines a pure Draft-to-Ready submit operation, four guards,
unchanged content/owner on success except state and revision, and explicit refusal
behavior. [model.json](model.json) expresses those rules in lifecycle/v1. The model
covers no persistence, authentication, UI, concurrency, external effect or real
ownership claim. Ready means only the example terminal state.

The five D_* decisions are fictional inputs, not conclusions produced by validation.
**Q_use remains unanswered:** who may accept this for real use, and what review
evidence must that person see? Meaning review, visual inspection and execution or
integration authority are not established. Those gates remain even if every
mechanical check passes.

## Identities and expectations

[inventory.json](inventory.json) freezes exact brief/model/suite/manifest/intake
bytes, a separately calculated model canonical identity and an independent suite
semantic identity. The model proposal is bound to the brief's exact source bytes;
the independently authored [suite](suite.json) is bound to the same brief and to
the model's semantic identity. The intake's artifact dependency hashes refer to
source bytes, not that semantic digest.

The fixture author read the brief and public profile/transport contract plus the
stable model interface and supplied semantic binding; it did not read model
expressions, compiler/generated code or observed behavior. It asked about an
underspecified refusal transition ID, which was clarified in the brief before
freeze. The [manifest](manifest.json) records the suite hash and nine case intents.
This is independent assistant authorship evidence, not a human approval or a proof
that every requirement is correct or completely tested. Expectations are proposals
for separate semantic review and later authorized execution.

## What the handoff check establishes

From the unmodified reviewed checkout's `modeling/lifecycle`,
`go test -race -count=1 ./authoringhandoff` checks these frozen files through existing
profile, strict model, intake, suite and check-only acceptance APIs. It generates
code/views in memory and checks syntax, declared table contents and artifact links.
It checks drift/invalid/stale controls without repairing their inputs. A panic
sentinel prevents candidate execution in this check.

The test's pass is only declaration and generated-artifact consistency evidence.
It is not a behavioral replay, diagram viewing, human meaning review or permission.
The retained intake question proves that model acceptance and unresolved decisions
can coexist; no aggregate approved/ready state is invented.

To prepare actual draft review files in an authorized disposable checkout, follow
[the Foundry reference](../references/foundry-handoff.md). Record the resulting
file hashes, view-inspection result and actual commands separately. No generated
SVG is committed here and no visual-review result is asserted by this document.

## Next owner/action

The responsible person must review the requirements/model and independent literals,
answer Q_use and decide whether the next bounded step is authorized. A viewer must
inspect the generated diagram and table through an allowed method. If execution is
then authorized, use a separately generated candidate project or explicitly composed
custom-model runner—not a default example binary with substituted digest metadata.
No publication, integration or deployment authority follows from this handoff.
