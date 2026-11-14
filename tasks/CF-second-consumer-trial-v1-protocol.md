---
id: CF-second-consumer-trial-v1
status: completed
owner: Claude pipeline coordinator
repositories: [spik3r/clanker-foundry]
---

# Write the protocol for a second consumer change by another operator

## Implementation claim

Owner: Claude pipeline coordinator. The repository owner asked for the
recommended next steps to be carried out. One of them is milestone 3 in the
[current direction](../docs/current-direction.md): a second change to the
consumer made by someone other than the session that built the tools. That
session must not run the trial itself.

Scope: a trial protocol and result template under
`docs/trials/second-consumer-change/`, and a link from the
[generator pipeline plan](../docs/generator-pipeline-plan.md). It names two
tasks against the consumer's existing generated screens, what the operator may
read, what to time and record, and what counts as success, a stop, or a gap. It
records a real-terminal session run by the coordinating session through a
terminal multiplexer, stating that it is not a person's check. Also this task
record and new immutable Project Memory files.

Exclusions: running the trial, changing any generator or the consumer, naming
an operator, or setting a value threshold, which is the owner's decision.

Checks: the repository validator, whitespace check and committed Project
Memory validation.
