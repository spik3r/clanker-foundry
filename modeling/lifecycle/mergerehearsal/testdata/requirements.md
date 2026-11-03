# Invented parcel dispatch requirements

This fixed example models a small sequential parcel rule, not a real shipping
service, identity system, approval, task claim or execution authority.

A parcel starts Raw. `packParcel` moves Raw to Packed. `dispatchParcel` moves
Packed to terminal Dispatched. Both transitions require the command's expected
revision to equal the parcel revision; their effects set the named destination
and then increment revision exactly once. They never assign or change an owner,
outcome or acceptance list. Commands at the terminal state refuse unchanged.

The base design permits any nonblank pack or dispatch actor. The left design
changes only `G_packActor` to require literal actor `packer`. The right design
changes only `G_dispatchActor` to require literal actor `dispatcher`. The merged
design retains both independently specified changes. Each rule is witnessed at
its own transition: guest pack succeeds in base/right and refuses in left/merged;
guest dispatch succeeds in base/left and refuses in right/merged. No single
combined-only witness is claimed for these separate transition changes.

The happy path packs as packer, dispatches as dispatcher, then refuses another
pack at terminal Dispatched. Separate cases retain the full unchanged task on
guard refusal, distinguish a present owner `keeper` from null, reject whitespace
pack actor, reject a stale revision, retain MaxUint64 on REVISION_LIMIT, and
return a null task for unknown-trigger INVALID_INPUT. Refusals are sorted guard
IDs; invalid input precedes transition selection, guards precede revision overflow.

A competing design requires pack actor `auditor`; merging it with left conflicts
on exactly the atomic `G_packActor` record. The invalid-combination branches each
add one valid transition with a different semantic ID from Raw on `inspectParcel`:
left targets Packed, right targets Dispatched. Their union has no atomic conflict
but duplicates the state/trigger dispatch pair and must expose no candidate.
Both failure statuses must stop the rehearsal before generation or replay.

A deliberately wrong expectation says happy_pack changes the outcome to
`Intentionally wrong expectation`. The real transition preserves `Ship <parcel>`
and must remain a fail-fast task mismatch. Separately
removing each strict guard from disposable generated code must compile and fail
at its named guest witness against the unchanged merged suite. No output-derived
expectation repair is permitted.

These models, suites, full merge and replay expectations were authored by a
separate assistant from the workflow implementer before implementation. Source
and semantic hashes were computed directly from these literal documents using
the public canonical contracts in Python; no implementation output supplied
expected behavior. This separation does not authenticate human authorship, prove
requirements correct, or confer approval.
