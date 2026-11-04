# FirstDesignDraft handed to a Go builder

This worked exercise connects an unchanged model blueprint to a fresh agent that
writes Go code. It does not substitute generator output for agent implementation.
The result is a fictional, pure typed Submit function, not a product, lifecycle
engine, standards adapter or production acceptance decision.

- [Implementation workflow](../../workflows/model-driven-implementation.md)
- [Worked packet](packet.md): source pins, decisions, trace, concrete contract,
  independent acceptance provenance, bounded plan and return requirements.
- [Frozen acceptance](acceptance/README.md): 30 literal cases, three isolation
  tests and the exact native-uint64 Go interface.
- [Observed evidence](evidence.md): source identities and actual separate-agent results.
- [Seven-control rubric](acceptance/rubric.md): separate fresh-agent observations;
  Go test passes do not establish control outcomes.

The five acceptance files were independently authored from only the original
brief and public Go profile, then frozen in commit
`0016283944c0bb868e2d3edb2ca94e935a421dba` before packet or builder work.
The packet was independently reviewed before fresh builder dispatch. Candidate
implementation, compile-only evidence, authorized execution and control results
must be reported separately; no pending step is a pass.

Use an explicitly located trusted repository checkout. `examples/` is not
Stow-installed, so installing the workflow does not provide this example's model,
brief or tests. Before dispatch the coordinator creates a separate private,
fully bound packet with exact accessible paths and verified hashes, copies the
frozen module inputs unchanged, and supplies trusted edit/build authority.
Only candidate.go and named build products/caches may be written. Candidate
execution needs separate exact-source review and authority; `go test -c` compiles,
while even `go test -run '^$'` executes initialization.

The optional view remains uninspected. Q_use and visual approval remain open.
D_scope, D_actor, D_owner, D_terminal and D_failures are fictional fixture choices,
not human product approvals. A narrowly authorized local experiment does not
answer those real-use questions. No persistence, authentication, JSON API,
input validator, CLI, runtime, dependencies, external systems, publication,
merge or deployment are included.
