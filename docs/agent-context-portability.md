# Comparing agent context at observable boundaries

Two agents given the same instruction file can still act differently. A file path,
model alias or matching prompt hash tells you something about configuration or
chosen bytes; it does not establish what a client or service actually loaded,
which instructions took precedence, or how a model will behave. Treat missing
information as unknown rather than filling it in from a configuration file.

## What Foundry configures

The repository's `AGENTS.md` is its intended project instruction source. The
repo's `CLAUDE.md` imports it and `GEMINI.md` points to it. The global wiring
script uses symlinks for Codex, opencode and Gemini and an import for Claude;
these are different mechanisms. A pointer or successful setup check does not
observe a client's effective context. Loading and precedence depend on the
actual client, version, launch directory and its configuration.

Agent profiles use tiers resolved by `scripts/agent-model.sh` through the shared
model map and, when present, a local override. The profile does not itself force
the host model. Record a resolved alias only as a configuration observation; it
need not identify an exact service-side model version or prove what was used.
See [agent profiles and model selection](../agents/README.md).

## What the offline test observes

`runtime/test/provider-input-parity.test.mjs` supplies invented requests to both
existing provider adapters using injected fake transports and an invented key.
It compares the explicit Responses user text and CLI stdin bytes separately to
independently frozen literals. One request omits instruction-file content; another
includes invented `AGENTS.md` content as JSON task data. Inclusion in that data
does not establish that the client loaded or trusted it as an instruction. The
checks cover Unicode and escaping, the current input-size boundary and rejection
before either fake transport, plus deliberate changes that the byte comparison
must reject. They do not invoke a provider or client, test model outputs, or
establish equivalent effective context.

The adapters expose additional observable settings: the Responses request sets an
empty tools array and `store: false`; the CLI invocation uses an empty private
working directory, flags, and a selected child environment. It copies `PATH`,
`HOME`, `CODEX_HOME`, `TMPDIR`, `LANG`, `LC_ALL` and `TERM` when present. Values
can influence a real client: for example, `PATH` affects executable lookup and
`HOME` or `CODEX_HOME` can identify client state. The fake test uses invented
values and does not launch a process. Flags and an empty directory are not proof
of a sandbox, actual rule loading, or compliance by a real client.

## A bounded comparison

1. Use a small invented task and explicitly list any invented file content being
   supplied. Record which client boundary you can observe and whether each file
   was omitted or included as task data. Do not collect private instructions,
   credentials, hidden prompts or personal paths for this comparison.
2. Compare the exact observable bytes independently against a reviewed literal;
   comparing two captured inputs only to one another can miss the same error in
   both. A hash can help identify chosen bytes, but equality is not evidence of
   provenance or that a client read or followed them. Consider whether even an
   equality-revealing hash is appropriate to retain.
3. Record known client and version, configured model alias, tool availability,
   relevant environment variable names and launch settings without secret values.
   Mark actual model/version, account settings, global and project instructions,
   hidden service or client context, context window and task history unknown when
   they have not been observed. Do not infer a universal effective-prompt export.
4. Run deterministic parser or transition checks on the same fixed inputs where
   appropriate. Separately evaluate model responses against explicit acceptance
   criteria, recognizing that model selection, context and probabilistic choices
   can vary even when visible inputs match.

The fake-only parity test is a regression check for this narrow transport boundary.
It does not authenticate an instruction source, certify security or prove equal
agent behavior.
