# Agent coordination

Keep concurrent work from duplicating changes or overwriting another worker's edits.
The [skill](SKILL.md) covers shared claims, narrow write scopes, worktree and staging
safety, handoffs, integration review and claim release. It can be used by one agent
working alongside people; it does not launch a team.

```text
Use agent-coordination to check the active work before I start this slice.
Use agent-coordination to hand this claimed task to the next owner.
```

It follows the existing task store and works without other skills. Ownership stays
there; checkpoints and implementation evidence link to it. See the
[adoption guide](../../docs/task-workflows.md) for setup choices.
