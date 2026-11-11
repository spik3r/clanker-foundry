---
id: CF-runtime-cli-bounded-read-v1
status: proposed
owner: Foundry runtime CLI maintenance contributor
repositories: [spik3r/clanker-foundry]
---

# Bound runtime CLI input reads

## Implementation claim

Base: `origin/main` at `30756603f16f3e4fe9216c40389d6d0df024218a`.
This is a proposed, bounded follow-up for the trusted local CLI. Review this
claim before registering a Project Memory source or starting implementation.
Recheck main, live claims and open pull requests before implementation.

`runtime/src/cli.mjs` checks a JSON input file's size with `fstatSync`, then
calls `readFileSync(fd)` and checks the returned length. A regular file can grow
between the stat and the read. In that case the CLI may read more than its
1,048,576-byte limit before rejecting it. A finite local probe with an owned
scratch file observed a zero-byte stat followed by a 2 MiB read after growth.
This is a pre-existing P3 resource-bound issue, not a remote exploit or a
blocker for the separate CLI special-file refusal pull request.

Limit descriptor reads to at most 1,048,577 bytes, including the byte needed
to detect an over-limit file. Keep the regular-file check, existing open flags,
descriptor cleanup, strict UTF-8 decoding, JSON parsing and redacted CLI error
behavior. Reject over-limit data without reading the rest of the file. Keep the
public CLI interface stable.

Implementation scope is `runtime/src/cli.mjs` and `runtime/test/cli.test.mjs`,
plus this canonical claim and new immutable Project Memory source and event
files required by the repository workflow. Tests should cover an empty file,
short reads, exactly 1,048,576 bytes, one byte above the limit, growth after
the stat, a read failure with descriptor cleanup, and multibyte UTF-8 near the
limit. Use finite owned files and bounded subprocesses; do not use an unbounded
producer or a blocking special file as a fixture.

Pull request #110 also touches `runtime/src/cli.mjs`. Its FIFO open behavior is
outside this claim. Do not copy, replace or merge its changes here. Review the
same-file interaction against the then-current base before publishing an
implementation pull request.

Acceptance: run the focused CLI tests, `npm test --prefix runtime`,
`scripts/validate.sh`, `git diff --check`, and the Project Memory validation
required for the committed implementation diff. Record observed results in a
completion event. This claim grants no authority to merge, deploy, activate an
orchestration namespace, dispatch workers or change repository protection.
