# Model change review

This offline tool helps review an AI-drafted change to a lifecycle model. It
compares normalized design content and shows which **declared links** lead from
changed requirements, guards and states to transitions in each model. It does
not evaluate either model or claim that omitted transitions behave the same.

Use it between drafting and accepting a model revision. Review the changed fields
and both sets of links, inspect the actual before/after source, then decide whether
to accept the design and explicitly update/re-run relevant acceptance evidence.
Nothing here accepts a design, updates expectations, rebinds review/scenario
digests, generates code, schedules work or grants execution authority.

## Try an editorial change

From `modeling/lifecycle`, use the existing exact Go 1.27.3 toolchain. This creates
only an in-memory candidate and prints a report; the tracked model stays unchanged:

```sh
python3 -c '
import copy, json, sys
before = json.load(open("examples/lifecycle.json"))
after = copy.deepcopy(before)
after["requirements"][0]["text"] += " Clarification for design review."
json.dump({"formatVersion":"lifecycle-model-diff-json/v1",
           "before":before, "after":after, "bindings":[]}, sys.stdout)
' | GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off go run ./cmd/modeldiff
```

The result lists the changed requirement and linked transitions in both models.
Even this text-only change sets `changed` and `reviewRequired` to true: existing
canonical identity intentionally includes reviewed requirements and state labels.
Those flags do **not** distinguish behavior change from editorial change. A false
flag is not approval or a proof of equivalent behavior.

The CLI takes no arguments: bounded stdin to buffered stdout. Exit 0 means a valid
comparison, even when the models differ or bindings do not match. Exit 2 rejects
invalid input, an unsupported toolchain or report limits, with no stdout. Exit 3
means a failed/short output write; stdout might then be incomplete and is not a
successful report. Diagnostics use fixed categories, not input text. Reports omit
model text, but IDs and metadata may still be private; handle stdout accordingly.

## Typed and wire boundaries

`Compare(before, after model.Model, bindings []Binding) (Report, error)` reuses
`model.Normalize` and `canonical.Digest`. It accepts the existing finite
`lifecycle/v1` semantic profile, not a new model language or validator. Inputs are
not mutated; returned collections are owned copies. Concurrent callers must keep
their own input values immutable.

`Execute(raw []byte) ([]byte, error)` accepts exactly this JSON request shape:

```text
{formatVersion: "lifecycle-model-diff-json/v1",
 before: <complete lifecycle-json/v1 document>,
 after:  <complete lifecycle-json/v1 document>,
 bindings: [{id: <asserted scenario ID>, modelDigest: <asserted 64-hex digest>}]}
```

All keys are mandatory. Bindings may be `[]`, not null. Unknown/case-variant,
missing/duplicate keys, trailing data, invalid UTF-8/surrogates and wrong model
versions refuse. Both models must validate completely before any report is
returned. No arbitrary paths, hooks, network, installations or filesystem outputs.

Bounds: request 3 MiB/depth 66; each embedded JSON model 1 MiB/depth 64 and all
existing semantic bounds; 128 unique bindings; buffered report 1 MiB. Binding IDs
use `[A-Za-z][A-Za-z0-9_]{0,47}`, case-sensitive, in a separate namespace from
model IDs. Digests are exactly 64 lowercase hex characters. Typed nil bindings
mean an empty list. Typed models retain semantic bounds without a JSON roundtrip.

## Changes and declared links

Changes sort lexically by `(kind, id)`, not by source order. Kinds are `model`,
`state`, `requirement`, `guard`, `transition`. The root model alone has empty ID.
An element changing kind or identity is a removal and addition, not an inferred
rename. Types are `added`, `removed`, `modified`; added/removed fields are `[]`.
For modified elements, field names sort lexically:

| Kind | Compared fields |
|---|---|
| model | id, initial |
| state | label, terminal |
| requirement | text |
| guard | requirement, expr |
| transition | from, trigger, to, guards, effects |

Existing normalization ignores collection/guard-set and pure all/any operand
ordering. It does not ignore effect order or invent behavioral equivalences.
Unsupported effect orders refuse rather than appearing as valid differences.

`linkedTransitions` lists the **before** graph first, then **after**, with IDs
sorted within each. A transition gets every applicable sorted unique cause:

- `transition:ID` when that transition changed or was added/removed;
- `state:ID` for a changed endpoint state;
- `guard:ID` for a changed referenced guard;
- `requirement:ID` for a changed requirement referenced by one of its guards;
- `model:` for a model ID/initial-state change, on every transition in that graph.

This is a finite direct-link projection, not graph-cycle closure or runtime data
flow. Removed links remain visible through the before graph. A transition absent
from this list is **not** proven unaffected. Global model-binding invalidation
still applies whenever canonical identity changes, regardless of linked entries.

Bindings are only caller-supplied assertions, never inspected scenario cases.
Results sort by ID and use exactly:

- `matches-after`: asserted digest equals the candidate, including equal models;
- `matches-before-only`: equals the prior model but not the candidate;
- `matches-neither`: equals neither.

These are digest comparisons, not verified scenario compatibility, execution,
coverage, review validity or authentication. The existing scenario runner remains
bound to its compiled model and is not weakened to accept a candidate.

## Identity and output

Report format is `lifecycle-model-diff-report/v1`, profile `lifecycle/v1`, mode
`example-only`, authority `none`. Canonical before/after digests identify accepted
normalized content. Typed Compare has no source bytes: request and both embedded
source digests are null. Execute records hashes of the exact input bytes and of
each extracted JSON **value**, excluding surrounding whitespace. These are not
the original model-file hashes. Formatting can change source identity while
canonical identity remains equal. Outer-only whitespace changes request identity
but not the embedded values.

Report fields follow the declared `Report` struct order; arrays are never null,
fields/causes are sorted, JSON uses no HTML escaping and ends with one newline.
No clock, random value or ambient machine path enters output. Execute validates,
compares and serializes internally; there is no public API accepting a separately
modified Report as a validated observation. Report values remain ordinary data
and can be forged like any JSON. Hashes do not authenticate review, binary or code.

## Verification and scope

Independent handwritten expected changes/causes/binding statuses freeze before
comparison implementation. Tests cover both graphs, all applicable causes,
normalization, identity/editorial changes, aliases, strict inputs, bounds,
deterministic bytes and I/O failures. Deliberate omission mutants must fail the
independent expectations. Existing generated code, original oracles and scenario
fixtures remain separate, unchanged gates.

This is an additive local follow-up to PR11 and the scenario runner, not an
independently mergeable production feature. No standard-conformance, runtime
migration or hostile-host invocation security is claimed. See the genuine
[task](../../../tasks/CF-model-impact-v1-review-diff.md) for scope and lifecycle.
