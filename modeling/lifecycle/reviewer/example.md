# Worked example: input and revision questions

This invented, partial worksheet is a reading aid. The source intake is
[example-intake.json](example-intake.json); it deliberately supplies no observed
artifact bytes. The trace refers to the committed [requirements](../traceability/requirements.json),
[model](../examples/lifecycle.json) and [catalog](../examples/oracle.json).
Only the two selected requirements are discussed; this is not a complete review.

## Identities

| Identity | SHA-256 |
|---|---|
| Exact intake request bytes | `312ea20f1c5eca4d281a09451097767de28c2c69cda3729a6e139b72d25f0779` |
| Exact requirements source bytes | `8f50fa3dbaf95ce1a5530a8256ecb42572ab96bb2ebb475177b1d07d446040d3` |
| Trace normalized requirements | `c951e96de70fc9be968b01395cfc459e562ebd0711e15ec7999543d3bf873bd3` |
| Exact model source bytes | `a60916ca68771b28ee48712288decad7eb741d6a18e1cbc95b2a196eb10d12d8` |
| Trace canonical model | `948630cb7857199443b59b92d9b1ab0fadf2bf7a5bf8125f7a7e767db83958ed` |
| Exact acceptance catalog bytes | `bb886d8a8d63a00f6b6a225c9b4855c306e5cfd0dc45a4f262d410842076acae` |

The intake declares the requirements and model hashes and invented author
labels. Both bindings report `missing-observation`; equality with separately
read bytes above is a manual comparison, not an intake observation. The suite
binding is `missing-declaration`. The model-to-requirements assertion reports
`matches-declaration`; suite-to-requirements and suite-to-model report
`missing-assertion`. These statuses neither validate the model against the
requirements nor establish that a suite exists. The catalog is separate from a
scenario suite. Exact source hashes and normalized identities have different
purposes and must not be substituted for one another.

## Human-declared relationships

These mappings are illustrative reviewer statements, not links supplied by the
intake or inferred by either tool:

| Intake item | Requirement | Stated reason / remaining question |
|---|---|---|
| `D_input_boundary` | `UR-010` | Both describe rejection before lifecycle rules. |
| `Q_input_evidence` | `UR-010` | Which independently executed cases establish this for the exact model and adapter? Unresolved. |
| `D_revision` | `UR-012` | Both describe one revision increment per accepted command. |
| `Q_revision_evidence` | `UR-012` | Which case should be explicitly linked? Unresolved. |

## Existing trace declarations

| Requirement | Selected model links | Acceptance cases as catalogued | Trace gaps |
|---|---|---|---|
| `UR-010` | none | `special-03` pending; `special-04` pending | `no-executable-acceptance`; `no-model-link` |
| `UR-012` | `T_approve`, `T_prepare`, `T_requestChanges`, `T_requestReview`, `T_start` | none | `no-acceptance-link` |

The catalog labels are declarations, not execution results. The raw-input
adapter has separate acceptance tests, but this catalog still marks those two
cases pending; this worksheet neither equates them nor asserts a passing run.
The trace links validate reference existence, not correctness or satisfaction.

## Recorded checks and outstanding work

| Check for this example | Recorded status |
|---|---|
| Intake report | `not run` as review evidence |
| Requirements trace | `not run` as review evidence; rows above describe the committed trace's declarations |
| Independent acceptance or scenario execution | `not run` |
| Human review or approval | `not performed` |

A maintainer may run the commands in the [reviewer guide](README.md), record
actual outcomes and artifacts, and resolve the two questions. The example and
its regression tests make no claim that this human review occurred.
