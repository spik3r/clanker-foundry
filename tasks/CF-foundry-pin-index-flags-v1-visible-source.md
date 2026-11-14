---
id: CF-foundry-pin-index-flags-v1
status: in-progress
owner: Codex pinned-source maintenance coordinator
repositories: [spik3r/clanker-foundry]
---

# Refuse hidden tracked changes in pinned generator source

## Implementation claim

The owner authorized several hours of bounded maintenance with independent
review, publication and merging after checks. Main at `65a7f4e`, live claims
and open PRs were reviewed. The separate project-setup repair owns Bash baseline
checks; runtime workspace-budget work is also independent.

Foundry's pin check accepts modified tracked modeling files when Git marks them
assume-unchanged or skip-worktree. Both failures were reproduced in disposable
local Git fixtures: HEAD matched the pin and status reported no changes while
the registry file's bytes differed. Status alone cannot certify these files.

Before trusting status, inspect tracked modeling entries and refuse either
visibility flag without clearing it or refreshing the index. Keep unrelated
flags and changes outside modeling allowed. The conservative check requires an
ordinary full modeling checkout; sparse checkouts using skip-worktree are not
certified. This closes a known status blind spot, not a hostile-host guarantee.

Use NUL-delimited metadata. Refuse incomplete metadata, including a listing
that reaches the existing capture limit, rather than trusting a possibly
truncated prefix. Preserve revision matching, untracked-file checks, ignored-file
policy and the explicit development-only --unpinned escape.

Scope: `modeling/foundrygen/exec.go`, new
`modeling/foundrygen/pin_flags_test.go`, `modeling/foundrygen/README.md`, this
task and new immutable Project Memory files. No generator, output, runtime or
project-setup changes.

Checks: show regressions fail before the repair for both flags on clean and
modified files; verify flags, index bytes and source bytes remain unchanged;
exercise outside-modeling and ordinary-checkout controls; test complete and
capture-limit metadata. Run gofmt, vet, full foundrygen race tests, repository
validation, whitespace and committed Project Memory checks. Independently review
the final commit and await all hosted checks before merging.
