# Scenario explanation: proposed contract

**Reviewed behavior contract.** A validated report can be invented. Every output describes
caller-supplied assertions and model declarations, not evidence that a reducer,
review, actor or test ran. It grants no authority and does not establish
coverage, authenticity or behavioral equivalence. Model text, guard expressions, expected literals, IDs and even
step IDs can contain private information; output must be handled accordingly.
Task, owner, actor, review and acceptance fields are not copied; equivalent
sensitive text can still occur in IDs, expected literals, requirements or expressions.

## Inputs and compatible model

The typed API takes a `model.Model` and `scenarios.Report`. Normalize the entire
model and calculate `canonical.Digest`; require equality with the compiled
`generated.ModelDigest` and the supplied report's model digest. This v1 is tied
to the existing compiled scenario profile; reject mismatched models rather
than relaxing `EncodeReport` or inventing another baseline. Call
`scenarios.EncodeReport` on the complete input report before mapping, and reject
any error, including inconsistency of matched/differences/counts. The serializer
validates shape and internal comparison consistency; it does not authenticate the
report or prove all possible cases were supplied. An additional whole-input
validation pass enforces the mapping rules below before any output is returned.
No call to Transition, Run or Execute and no guard evaluation is permitted.

An additive `scenarios.DecodeReport` may decode the existing exact report JSON
into owned typed data. It must use the package's existing DTOs, decimal revision
and task/command conversion helpers, strictjson and `EncodeReport`, preserving
null/empty distinctions. `EncodeReport` can encode actual refusals as `null`;
the decoder must accept that exact wire case and distinguish it from `[]`,
for example with a decode-only nullable slice DTO. It must not normalize either. Reject missing, duplicate, unknown or case-variant
fields; wrong versions, invalid Unicode, trailing values and depth/size errors.
The decode-only DTO may repeat wire declarations, not comparison logic. Check
the input envelope formatVersion, scenarioVersion, profileVersion, mode and
authority against exact constants before conversion: typed Report does not retain
these fields and EncodeReport would otherwise regenerate them. Required pointer
fields must be present even when null. Reuse the existing task/command
conversions, copy helpers and EncodeReport; preserve every semantic field and
null distinction. Never turn a decode/encode round trip into a source
authenticity claim. If the
existing private helper API cannot support lossless decoding, stop and report the
specific mismatch; do not duplicate the comparison semantics or change old files.

## Actual assertions and declared links

For every report entry, require actual `Mode`, `Authority` and `ModelDigest` to
match example-only, none and the compatible model. An actual successful result
must have a known transition ID and no refusal IDs. A known actual transition
must be declared for the entry's input state and command trigger; otherwise
reject the whole explanation. These are structural consistency checks, not
proof that the transition was evaluated or that its effects were correct.

For an actual refusal, fixed runtime reasons are exact singleton lists:
`INVALID_INPUT`, `ILLEGAL_TRANSITION`, or `REVISION_LIMIT`. Reject mixtures of
fixed and model reasons, duplicates, unknown actual reasons, or a fixed reason
combined with another reason. For `INVALID_INPUT` require empty transition ID
and nil actual task; for `ILLEGAL_TRANSITION` require empty transition ID;
for `REVISION_LIMIT` require a known compatible actual transition. Do not infer
why the runtime reported a fixed reason. Fixed IDs are reserved in explanation
classification; if the model declares a colliding guard ID, reject the model
for this explanation profile rather than silently choose a meaning.

For actual model-guard refusals, require a known actual transition, nonempty
sorted unique reason IDs, and every ID to be a guard declared on that transition.
For each, report its stable ID, declared requirement ID and text, guard expression,
and direct transition ID. These are declared links and supplied refusals, not
facts that the guard is false. Do not claim linked guards or requirements are
complete downstream impact or dynamic coverage. Reject a guard outside the
claimed transition and any unknown actual transition/reason for the whole input.

Preserve each entry's supplied matched flag, actual refusal null/empty distinction,
and its exact comparison differences;
never rewrite a mismatch as success. Expected transition/reason strings are
literal caller expectations: retain them only as bounded mismatch metadata,
without classifying unknown expected reasons as actual model guards. Unknown
expected values are allowed when `EncodeReport` accepts them. Actual structural
inconsistencies remain errors even for mismatch entries; do not mask the reported
mismatch by fabricating model definitions.

## Output, bounds and tests

The expression output wire shape is the versioned `Expression` projection in
`types.go`: all five fields `kind`, `ref`, `value`, `op`, and `args` are present;
all non-args fields are strings and `args` is always a non-null array, including
for leaves. It is a structural projection, not an evaluator.

Pure typed output owns all collections and expressions; no pointer or slice
aliases the input model/report. Provide deterministic compact JSON plus newline
through a bounded stdin/stdout CLI, with no filesystem, network, process, image
or browser operation. Reject complete input on any invalid model/report/link,
without returning a partial typed explanation or starting output before full
validation and encoding. An I/O failure or short write can still leave a prefix
with the caller; report failure rather than claiming atomic stream writes. Keep
source/scenario digests as supplied
assertions and do not imply that they authenticate bytes. Use existing model and
scenario input bounds, with a separately specified finite explanation output cap;
short writes and oversized input/output fail without leaking input in diagnostics.

The independently authored literal vectors and explicit expected output were
frozen before mapping; they cover single/multiple guards, each runtime refusal
category,
expected mismatch with an unknown literal reason, unknown actual reason,
unknown or mismatched transition, mixed fixed/model refusal, stale model digest,
sensitive caller literal escaping, altered-model-text digest rejection, and
invalid/oversized/alias cases. Test deterministic
bytes, ownership on both success and failure, strict report round trip and
malformed null/missing/case/duplicate/Unicode/depth/decimal input. Establish all
expected fields from the written contract, not generated output. Keep model and
scenario fixture data invented and public-safe. Complete only after lead and
independent review and actual Go1.27.3 validation.


## Lead ABI and transport clarification

The public typed API is `Explain(model.Model, scenarios.Report)
(Explanation, error)`. Additive `scenarios.DecodeReport([]byte) (Report,
error)` retains the existing report wire contract. `Execute([]byte)
([]byte, error)` accepts exactly one wrapper with required non-null fields:
`formatVersion` = `lifecycle-scenario-explain-request/v1`, `model` = existing
JSON-v1 model, and `report` = existing report JSON. Unknown, duplicate,
case-variant and missing wrapper fields or a null required value fail. Wrapper
maximum is 10 MiB and depth 66; embedded model is at most 1 MiB/depth 64,
report is at most scenarios.MaxReportBytes (8 MiB)/depth 64, and output is at
most 8 MiB. Counts must be present integer JSON numbers in the existing report's
int range; reject fractional or exponent notation, negative invalid counts,
and overflow; do not coerce strings or null.
Output is compact deterministic JSON plus LF, buffered before any write.

Fixed Explanation.Status is `caller-supplied-assertions`. Entries, Guards and
Expression.Args are non-null arrays. Differences and ExpectedRefusals retain
the report's required non-null arrays; ActualRefusals preserves null versus [];
SourceDigest preserves an owned nullable pointer. `Execute` creates a fresh
explanation only; no public arbitrary-Explanation encoder is exposed.

The CLI takes no arguments or file paths and uses stdin/stdout only. Exit 0
means assertions were explained, including a mismatch; exit 2 is malformed,
binding, link, toolchain or limit failure with empty stdout; exit 3 is a failed
or short stdout write. Diagnostics are fixed and private-safe. A failed write
may have emitted a prefix. The CLI requires exact `compiler.ToolchainVersion`;
Go 1.27.3 is required for final qualification. Decoder and mapping stay pure
and never launch a process or render a browser.

Model guard IDs colliding with fixed reason IDs are rejected globally before
mapping. `ILLEGAL_TRANSITION` and `REVISION_LIMIT` classification describes
only the caller assertion and declared structural compatibility: it does not
prove a transition was evaluated, was illegal, overflow occurred or any effect
was applied. Altering sensitive model text is a negative digest-mismatch case;
only caller-supplied literals in an otherwise compatible report test escaping.


## Current-main adoption and default-report boundary

This adopts the existing local explainer, not the held archive/reviewpack stack.
The fresh task pins the original source and literal vectors; mapping, decoder and
original tests are unchanged. Scenario conversion/encoding internals have since
gained the separately reviewed model-bound Runner, so the adoption tests qualify
those current helpers rather than claiming an unchanged dependency closure.

`DecodeReport` accepts only the original baseline report wire contract. It rejects
`lifecycle-bound-scenario-report/v1` and the `identitySource` field separately;
it never strips these fields or substitutes a requested model identity. Changed
models, including the tutorial candidate, remain unsupported by this v1 explainer.

A typed `scenarios.Report` does not carry its wire envelope or origin. An otherwise
compatible typed value can be invented or come from a baseline-bound callback;
`Explain` cannot distinguish those cases. It accepts structural assertions, not
an authenticated execution receipt. For example, a structurally consistent report
may claim success without applying the declared effects. Explanation does not
check effects, invoke the callback or promote that assertion into observed truth.

The independently frozen integration uses the existing four-case/thirteen-step
suite through the existing scenario engine, then checks complete literal
explanations, guard projections, report roundtrip and the strict wrapper. It also
checks invented reports, null/empty refusals, stale model identity and separate
bound-version/unknown-field rejections. Bound report encoding is tested with a
callback that must never run. Neither original nor integration expectations are
refreshed from explanation output.

The CLI now ignores SIGPIPE so a real closed stdout returns the documented exit 3
with the fixed output diagnostic. A trusted subprocess regression has a deadline
and finite WaitDelay; it is not a general process runner or a process-tree cleanup
guarantee. Existing short-writer tests remain.

## Reuse the existing report workflow

From the module directory with actual Go 1.27.3, run:

```sh
GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off go test -race -count=1 ./scenarioexplain ./scenarios ./cmd/modelexplain
```

For an inspectable example, compose the existing commands; no new orchestration
entrypoint is needed. The Python wrapper preserves the exact model/report bytes
instead of text-mode newline conversion. All files below are example artifacts
in a new local temporary directory, not authenticated execution evidence:

```sh
export GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off
work=$(mktemp -d)
go run ./cmd/modelplay < scenarios/testdata/lifecycle-v1.json > "$work/report.json"
python3 - "$work/report.json" > "$work/request.json" <<'PY'
from pathlib import Path
import sys
model = Path("examples/lifecycle.json").read_bytes()
report = Path(sys.argv[1]).read_bytes()
sys.stdout.buffer.write(
    b'{"formatVersion":"lifecycle-scenario-explain-request/v1","model":'
    + model + b',"report":' + report + b'}\n'
)
PY
go run ./cmd/modelexplain < "$work/request.json" > "$work/explanation.json"
```

The resulting status is still `caller-supplied-assertions`. Source/scenario digests
are copied claims, not authentication. Keep real private model text and reports
out of public artifacts. This workflow performs no browser rendering, deployment,
worker activation or UML/SysML conformance check.
