# Project Memory

A dependency-free Node/Git helper for repository-local implementation continuity.
Read the [workflow](SKILL.md) before starting, resuming or finishing a change.

It records append-only start, checkpoint, complete and abandon events in
`.project-memory/tasks.jsonl`. Status shows the latest state and marks old active
work for review. Central backlog tasks still own claims and project lifecycle.

The committed-diff validator rejects rewritten history, stale historical
coverage, unfinished tasks, and completion snapshots that no longer match the
changed files. It handles additions, deletions, renames, modes, symlinks and
literal Unicode paths. Completion records store Git object IDs, not file data.

Use Node 24 and Git. No package installation, network access, secrets, or
repository write token is needed for validation. The helper must run from
inside the target Git repository. New events use schema version 1; historical
unversioned events remain read-only compatibility data.

Run regression tests:

`node --test tests/project-memory.test.mjs`

In Clanker Foundry, the `project-memory` CI job runs those tests on pushes, pull
requests and manual runs. It validates lifecycle evidence for the full committed
pull-request diff using that PR's base SHA. Pushes and manual runs test the helper
without guessing a comparison base or requiring active work to be complete.

Adding this check to CI does not itself make the status required. Repository
rules must separately require the relevant job; never weaken other tests or
change those security settings as a side effect of adopting this helper.
