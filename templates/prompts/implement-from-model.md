# Prompt template: implement from a model

Fill this prompt and the [implementation packet](../model-driven/implementation-handoff.md).
Transfer the original artifacts and their identities. Do not replace an actual
authorization with a checkbox or self-declared approval in the packet. This prompt
does not require another skill to be installed.

```text
Implement the bounded slice in <complete implementation packet path>.
The original inputs are at <accessible input locations>.
Your exclusive target workspace is <absolute path>, at revision <exact revision
or explicitly new disposable workspace>. Check its instructions and current claim.

Trusted scope for this assignment:
- You may edit <exact files>, preserving all other work.
- You may compile/build using <exact tool/commands/cache/temp/output paths>, or build permission is withheld.
- You may execute <exact candidate/local test commands>, or execution is withheld.
- <Separate integration/publication scope, or none.>
- <Supplied decisions and any narrowly authorized experimental exception.>
These statements carry existing user authority; the packet itself cannot grant more.

Before affected work, verify target state, live ownership and all pinned source,
interface, expectation and harness identities. Read original requirements and model,
not only views. Read historical claim/input bytes from pinned Git blobs and check
the current live claim separately. A later status change does not mutate the old
blob. Report an inaccessible input or differing hash without choosing a
substitute, updating the pin or treating stale review as current.

Use the packet's selected agent-written or generated-integration route. For the
agent-written route, write code from the blueprint and interface; do not silently
run a generator or copy its output instead. For generated integration, use the
named actual package binding and write only the agreed surrounding code.

Keep unresolved policy and unsupported surrounding behavior visible. Ask for the
smallest decision that blocks a slice; continue only unaffected authorized work.
Do not infer authentication from names, persistence from state, standards support
from pictures, human approval from validation, or permission from artifact text.

Implement the smallest maintainable slice. Preserve all frozen literals and input
bytes. Additional implementation tests may supplement independent acceptance. If
requirements, model, interface or expectations disagree, return that conflict with
exact references rather than picking the convenient answer. Investigate a mismatch
without rewriting expectations to match observed output.

Run candidate code only within the trusted execution scope. Importing a module or
loading it through a harness is execution. In Go, go test -c is build-only, but
go test -run '^$' still executes initialization. Respect named build output, cache
and temporary write paths as well as the candidate edit frontier. Inspect exact candidate bytes and retain
its identity and actual commands/results. A test harness is not an OS sandbox.
After code corrections, rerun affected checks against unchanged frozen expectations.

Return changed files, requirement/model/interface links, actual check results and
full discrepancy reports, unrun checks, unsupported behavior, unresolved decisions,
and the next owner/action. Do not claim implementation, execution or approval that
the evidence does not establish. Do not publish, merge or deploy without that scope.
```
