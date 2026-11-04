---
id: CF-provider-reply-boundary-v1
status: in-progress
owner: Codex provider reply boundary contributor
repositories: [spik3r/clanker-foundry]
---

# Reject incomplete-turn text and malformed provider text

## Implementation claim

Owner: Codex provider reply boundary contributor. Branch
`codex/provider-reply-boundary-v1`, reviewed main `249cf74`. The repository owner
requested another batch of selected larger tasks. The lead independently
reproduced these transport failures and explicitly assigned this bounded repair.
The worker may publish the reviewed claim and validated PR; the lead owns
independent review and merge.

The historical provider task still says review but records implemented/reviewed
work and has a completed Project Memory frontier. Do not falsely describe that
record as an explicit claim release. The live PR search found no competing lane;
the lead allocated this new defect repair without rewriting old claims. Held
general recovery conformance, live provider activation and deployment are excluded.

Exact scope:
- `runtime/src/providers.mjs`
- New `runtime/test/provider-reply-boundary.test.mjs`
- `runtime/test/provider-input-parity.test.mjs`, only the missing turn.started line in fakeCliJsonl described below
- `runtime/README.md`
- This task and its immutable Project Memory source/events.

Preserve all other existing test bytes, schemas, evaluator/store/runner/adapters, CLI operator
routing, provider arguments/credentials, timeout/cancellation behavior and byte
limits. Use invented fetch and child-process fixtures; no real provider call.

## Reproduced failures

The CLI parser keeps an earlier turn-completed flag while replacing the final
assistant text with a later message. It accepts assistant text after the terminal
turn, and text from a second turn that never completed. Public-API injected-child
reproductions and a valid complete control confirmed this without a real process.
Both transports also decode malformed raw UTF-8 by replacement rather than
refusing it; escaped lone-surrogate assistant text can cross their text boundary.

## Approved bounded transport contract

Accept one turn.started, assistant/reasoning items inside that active turn, then
one turn.completed with completed assistant text. No item.started counterpart or
item ID is required. A second turn start, terminal without a start, repeated
terminal, or item outside the active turn raises fixed
PROVIDER_CLI_INVALID_OUTPUT without returning any text. Missing completion/text
at EOF remains PROVIDER_INCOMPLETE. Optional thread.started occurs at most once
before the turn. rate_limits.updated metadata may occur before/during/after it.
Keep the exact existing code-mode-disabled item diagnostic inside the active turn;
other tool/error denials retain their existing fixed classifications.

This is Foundry's bounded one-prompt transport contract, not an official complete
Codex event validator. The official noninteractive example shows
thread.started, turn.started, item.completed agent_message, turn.completed and no
mandatory item.started: https://developers.openai.com/codex/noninteractive .
It does not specify every sequencing rule above.

Decode actual CLI/stdout and Responses raw-stream bytes with fatal UTF-8 checks;
reject malformed assistant surrogate text after JSON decoding, with no Unicode
normalization or replacement repair. Preserve valid split multibyte sequences,
paired surrogates and literal U+FFFD. Use fixed PROVIDER_CLI_INVALID_OUTPUT or
PROVIDER_INVALID_RESPONSE for the respective malformed transport/text, without
reflecting payloads. Preserve BOM rejection rather than silently stripping it.

Keep the existing text-only fetch fallback compatible: enforce well-formed string
and byte limits, but document that original byte encoding cannot be verified once
an injected fetch adapter already decoded the text. Do not claim raw-wire proof
for that fallback. Keep cancellation, output limits and successful current fixtures.

## Acceptance and checks

Freeze independently authored boundary tests before production edits and observe
old-code failures plus positive controls. Cover valid example/control metadata,
framing refusals, incomplete EOF, tool/error classifications, malformed raw bytes,
split multibyte positives, assistant surrogate negatives and text-only fallback.
Run affected/full runtime tests on the actual local Node version (24 or later),
mutation controls, syntax, pack, whitespace and full committed-diff Project Memory
validation. Record actual results and limits; require hosted Node 24 CI and lead
review before merge. No live provider result is claimed from injected fixtures.


## Reviewed fixture scope amendment before edit

The independent boundary tests were frozen in `ddc27f0` before production edits.
They and the existing provider control tests pass on the WIP repair. The separate
input-parity test's fakeCliJsonl fixture emits item.completed and turn.completed
without turn.started. That fixture conflicts with the reviewed single-turn
contract; accepting it would also accept the new no-start negative case.

The lead independently read that fixture and approved exactly one change in
`runtime/test/provider-input-parity.test.mjs`: prepend one turn.started JSONL
frame to fakeCliJsonl. Preserve every prompt literal, hash, argument, environment
assertion and other fixture byte. This corrects a success fixture's protocol
framing without weakening the parser or its independent negative expectations.

Checkpoint and abandon the original run without completion because its pinned
claim excluded old tests. Publish this amended claim, obtain lead review, register
that exact source and make a genuine new start for all final paths before editing
the parity fixture. Existing production WIP is an explicit handoff from the first
run. Preserve all original start/freeze/source bytes and keep the new independent
file's SHA-256 unchanged. Final completion belongs to the amended run.
