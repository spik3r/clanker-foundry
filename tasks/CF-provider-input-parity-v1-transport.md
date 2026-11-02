---
id: CF-provider-input-parity-v1
status: in-progress
owner: Codex / offline provider-input portability contributor
repositories: [spik3r/clanker-foundry]
depends_on: [existing provider transport APIs at reviewed Foundry main]
---

# Compare explicit provider inputs without claiming identical agents

## Prospective implementation claim

This public-safe claim was independently reviewed at the earlier main and
refreshed for the local start on Foundry main
`463847f07664768b25e8740ed821857d8dc22f23`, tree
`8ad5dbbd3afe056fdaab986de77d1a652cea219c`. The proposed paths and task ID
are absent from that tree; the existing provider owner's source and test paths
remain excluded. The permitted open PRs #5 and #36 have no overlapping declared
frontier. PR #20 and held publication branches were intentionally not inspected;
this is not proof an unseen claim or unadvertised branch is free. A local claim
is not a shared reservation. Recheck current main and all permitted active claims
with the coordinator before publication; stop on an unresolved overlap. Do not
inspect denied diagnostics or protected visual work.

Owner: Codex offline provider-input portability contributor, subject to the
coordinator's assignment and separate review. Proposed implementation branch:
`codex/provider-input-parity-v1`. The bounded maintained frontier is only:

- New `runtime/test/provider-input-parity.test.mjs`.
- New `docs/agent-context-portability.md`.
- This new public-safe task record at
  `tasks/CF-provider-input-parity-v1.md` and this task's own new immutable
  `.project-memory/tasks/CF-provider-input-parity-v1/` source and events.

No changes to `runtime/src/providers.mjs`, existing provider tests, runtime
contracts, model maps, instruction files, installer or configuration. No new
product runner, provider integration, credential handling, dependency, network
call, live model invocation, installation, deployment or worker activation.
Use only invented public-safe fixture data and private temporary directories.
Preserve all other owners' claims and files, including the active provider
implementation owner's existing source and test paths.

After independent scope review and coordinator clearance, commit this local claim,
register its exact source revision and claim section, and record a genuine start
before maintained implementation edits. A separate contributor freezes invented
inputs, exact expected prompt bytes and negative-control outcomes before test or
guide implementation. The provider implementation output is not the acceptance
oracle. Preserve the freeze and immutable event history. Assistant separation is
not independent-person attestation or formal approval. Publication requires fresh
ownership review and independent exact-source review; local start is not publication.

## Problem and boundary

The existing `createResponsesProvider` and `createCodexCliProvider` both call
`buildPrompt` on the supplied request, but their transports and host contexts
are different. The Responses adapter sends the explicit user text, with no tools
and no stored response; it does not itself discover a repository's instruction
files. The CLI path uses a private empty working directory and configured
feature restrictions, and receives its explicit prompt on stdin. Do not infer
from that stdin what additional internal client instructions, global files,
hidden context or model settings were actually loaded.

The existing `runtime/test/providers.test.mjs` exercises each transport's
protocol and error cases independently, but does not compare their explicit prompt
bytes with each other and a literal independent expectation. This task adds that
narrow regression boundary; it does not establish equality of model inputs beyond
what the fake transport surfaces can observe. A shared `AGENTS.md`, identical
explicit text or matching hash does not prove rule loading, instruction
compliance, output equality, model identity, authentication, provenance or safety.

## Frozen offline test contract

Use the existing public provider factory functions with both an injected fake fetch
and an injected fake spawn, and an invented API key only. Capture the Responses
request's user text and the CLI child's stdin as UTF-8 bytes. Never call a real provider, a real Codex CLI,
or the network. Any temp CLI working directory must be private, empty and
resolved as required by the existing factory. A fake uses invented text/events;
no real credentials, user instruction files, account data or private content.

For each same plain JSON request object, compare both captured explicit prompt
byte sequences to the independently authored literal bytes, not to a value
recomputed using `buildPrompt`, `JSON.stringify` or the other transport. Freeze
expected JSON property order and byte encoding for each fixture. The test does
not impose canonical serialization on arbitrary semantically equivalent objects.
At minimum cover:

1. An invented request with an empty file list and no instruction-file fixture;
   neither captured explicit payload contains the omitted fixture marker.
2. An invented request explicitly containing an `AGENTS.md`-named file with a
   recognizable instruction-like marker. Both explicit payloads contain exactly
   the literal serialized data. The marker is task data in this boundary; the
   test does not assert it is a trusted or loaded provider instruction.
3. Quotes, newlines, non-ASCII and supplementary Unicode in task/file data, with
   exact bytes and order checked. The valid request remains unchanged after each
   fake invocation. Whitespace or a one-byte difference must be detected.
4. A JSON serialization error and an input whose serialized JSON exceeds the
   existing 1 MiB bound. Assert the existing `INVALID_INPUT` or `INPUT_TOO_LARGE`
   error respectively and zero fake fetch/spawn invocations, using valid private
   CLI cwd and never weakening source limits.
5. Check the Responses request's existing `tools: []` and `store: false`, and
   the CLI's existing relevant arguments, without asserting that command flags
   are a sandbox or that the real client actually obeyed/loaded a file.
6. Document that the CLI adapter selectively copies `PATH`, `HOME`,
   `CODEX_HOME`, `TMPDIR`, `LANG`, `LC_ALL` and `TERM` when present. Do not record
   actual values, personal paths, credentials or private configuration. A fake
   spawn observes adapter options only; inherited variables and matching flags do
   not establish the effective context of a real client or identical behavior.

The fake provider result is used only to exercise capture and protocol plumbing;
no generated result becomes an expected prompt. Include a targeted negative
control showing a one-byte altered explicit prompt is rejected by the comparison,
not a generic failure. Report these as offline transport-boundary tests, not
cross-model evaluations, live authentication, or proof of equal behaviour.

## Portability guide and verification

The new guide should distinguish configured instruction paths, directly observed
explicit payload bytes and unknown effective provider/client context. Explain
Foundry's symlink/import differences, model tier resolution and local overrides,
client/project instructions, permissions, available tools, model/version,
context window and task history as sources of variation. Label inference and
unobserved client behavior; do not state that a symlink, model alias, instruction
hash, prompt hash or self-report proves a file was read or followed. Do not
invent a universal effective-prompt export or store credentials, raw private
prompts, hidden system data or personal paths in examples or logs. Hashes may
reveal equality and must not be treated as proof of provenance.

Show a bounded, manual comparison method using invented inputs and observable
metadata, with missing client/model/context information marked unknown. Distinguish
deterministic parser/transition checks from probabilistic model choices. No
instruction precedence changes or model pin changes are proposed.

Validate with Node 24 offline: run the new targeted fake-only test and the
existing `runtime/test/providers.test.mjs`, the normal applicable runtime/pack
checks if their already installed dependencies are available without network,
format and whitespace checks, and full committed-diff Project Memory validation.
Report exact versions, test counts, missing dependencies and any unrun checks; no
live provider or client result can be inferred. Seek separate exact-source review
before a single coherent draft PR, if permitted. Do not approve, merge, deploy,
activate workers, access denied diagnostics or claim identical agent behaviour.
