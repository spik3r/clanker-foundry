# rawinput: raw Task and Command JSON for the lifecycle reducer

An isolated, offline demo. It turns two JSON byte slices into the typed `Task` and
`Command` values the [generated reducer](../generated/lifecycle.go) takes, then calls
`Transition` once. It is not a production API, an HTTP endpoint or a second
lifecycle engine, and it adds no lifecycle rule. Results stay `example-only` with
`authority=none`.

```go
result, err := rawinput.Transition(taskJSON, commandJSON)
```

- `err == nil`: both inputs were structurally valid and `result` is exactly what the
  reducer returned, including a refusal, `ILLEGAL_TRANSITION`, `REVISION_LIMIT` or
  the reducer's own `INVALID_INPUT` for a typed-invalid value.
- `err != nil`: a raw input was rejected and the reducer was **not** called.
  `result` has the shape the reducer gives typed invalid input (`OK=false`,
  `Task=nil`, `Refusals=["INVALID_INPUT"]`), and `err` is an `*InputError` naming
  the input (`task` or `command`) and a `Reason`. It holds no input text.

The task is decoded first. If it is rejected the command is not examined. Both are
decoded before the reducer runs, so a bad command blocks the call even when the
task is fine.

## Wire format

Strict JSON, UTF-8. Every key below is required. Keys are case-sensitive.

| Input | Keys |
|---|---|
| Task | `id`, `state`, `revision`, `outcome`, `acceptance`, `owner` |
| Command | `type`, `actor`, `expectedRevision`, `review` |
| Review | `taskId`, `taskRevision`, `modelDigest`, `taskDigest`, `reviewer`, `decision` |

Revisions (`revision`, `expectedRevision`, `taskRevision`) are JSON **strings** in
canonical decimal: `0`, or digits with no leading zero, no sign, no space and only
ASCII digits, at most 18446744073709551615. A string like `"01"`, `"-1"`, `"1e3"` or
`"18446744073709551616"` is rejected with `invalid_revision`. A JSON number there is
a type error (`strict_json`). Strings avoid the precision loss JSON numbers have
above 2^53.

## Null versus empty

| Field | `null` | empty value | Both reach the reducer as |
|---|---|---|---|
| `owner` | no owner (`nil`) | `""` is a present, empty owner | `nil` vs pointer to `""` |
| `acceptance` | no list (`nil`) | `[]` is a present, empty list | `nil` vs empty non-nil slice |
| `review` | no review (`nil`) | `{}` is rejected (keys are required) | `nil` vs rejected |

The reducer treats a nil and an empty acceptance list alike (both fail
`G_acceptance`) but keeps them distinct in the unchanged task it returns. A present
empty owner is not absent: `start` refuses it with `G_ownerCompatible` when the
actor differs, and `approve` refuses it with `G_owner`. `null` is rejected for
every field not listed above, and so is a missing key.

## Rejected before the reducer (`strict_json` unless noted)

Empty, malformed or trailing JSON; duplicate keys (also after `\u` escape decoding);
unknown or case-folded keys; missing keys; wrong types; a top-level value that is not
an object; invalid UTF-8; lone or misordered surrogate escapes; input over the size
limit; nesting over the depth limit; invalid or overflowing revisions
(`invalid_revision`). These checks come from the existing
[`strictjson`](../strictjson/decode.go) decoder plus the revision parser here.

Left to the reducer, so they come back as ordinary results with `err == nil`:
unknown state or command type, blank ID, over-long text, more than 32 criteria,
blank criteria, malformed digests and unsupported decisions. Guard precedence and
sorting are the reducer's: `ILLEGAL_TRANSITION` first, then guard failures sorted,
then `REVISION_LIMIT` only when every guard passed. A guard or limit refusal carries
the transition ID (for example `T_approve`); only `ILLEGAL_TRANSITION` leaves it empty.

## Limits

| Limit | Value | Why |
|---|---|---|
| `MaxTaskBytes` | 262144 | outcome 4096 + 32 criteria of 1024 bytes, each up to 6x when written as `\u` escapes |
| `MaxCommandBytes` | 16384 | a few 256-byte strings and two digests, escaped |
| `MaxDepth` | 4 | top-level object is depth 0; no valid input needs more than 2 |

These bound the decoder only. Content limits stay in the reducer. Whitespace padding
up to the size limit is accepted; one byte more is rejected. Deep nesting is
rejected without unbounded recursion.

## Oracle cases special-03 and special-04

[`examples/oracle.json`](../examples/oracle.json) lists both as `pending` because
no raw adapter existed. That file is unchanged.
[`rawinputacceptance`](../rawinputacceptance/oracle_test.go) reads the two cases
from it, feeds their `rawTask` and `rawCommand` strings verbatim, and checks the
catalogued outcome (`ok=false`, no task, no transition, `INVALID_INPUT`). Note that
special-04's `rawTask` is `{}`, which is not a valid Task here (keys are required),
so that case is rejected on the task. A separate case pairs a valid task with
`not json` to show a command rejection.

## Tests and limits of the claim

`rawinputacceptance/expected.json` holds 318 literal expected outcomes written by an
agent that did not read this code or the reducer source, and committed before the
adapter existed. After the first run three corrections were made, because the
author lacked the model file (`G_owner` on approve) and the contract did not specify
the transition ID on refusals: the ID rule for all guard and limit refusals, the
absent-owner approve case (adds `G_owner`), and the empty-owner approve case (a
refusal, not a success). The commit history shows both steps.
`adapter_test.go` checks that rejected input never reaches the reducer, that the
rejected result shape equals the reducer's own typed-invalid result, and that valid
requests return the reducer's result unchanged.

This proves consistency of invented example data, not authentication. Actor,
reviewer and digests are supplied assertions. The adapter keeps no state, opens no
files and uses no network. A hostile caller can still send many requests; no rate
limit exists here.
