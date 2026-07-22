# Code review checklist

Walk this before requesting review (self-review) and during review of others' work.
Order matches how a review should read: correctness first, style last. Every "no"
needs a comment in the review or a reason it does not apply.

## Intent and scope

- [ ] The diff does what the PR description says — nothing more
- [ ] Unrelated changes are split out (refactors, renames, formatting)
- [ ] The change is small enough to review in one sitting; if not, it is stacked or split

## Correctness

- [ ] Happy path traced by reading; edge cases named and handled (empty, null, boundary, concurrent)
- [ ] Errors are handled at the right level — not swallowed, not panicked past the caller's ability to act
- [ ] Failure modes considered: what happens when the dependency is down, slow, or returns garbage
- [ ] State changes are atomic where they must be; partial failure leaves no corrupt state
- [ ] Concurrency: shared state is guarded; no races, no deadlocks introduced
- [ ] Backward compatibility: public APIs, stored data, and migrations still work for existing consumers

## Tests

- [ ] New behaviour has tests; changed behaviour has updated tests
- [ ] Tests assert behaviour, not implementation details
- [ ] A regression test exists for every bug fix
- [ ] The suite was run and passes — the reviewer can see the command and result
- [ ] Coverage of the risky paths is real, not a number (check the error branches)

## Security and data

- [ ] `checklists/security.md` walked for anything touching auth, input, secrets, or personal data
- [ ] No secrets, credentials, or personal data in the diff or in test fixtures
- [ ] New external input is validated; new output is encoded

## Design

- [ ] The change fits the codebase's existing patterns and boundaries; departures are justified in the PR
- [ ] No new abstraction without a current second user; no dependency added without justification
- [ ] Public interfaces stay stable, or the breakage is deliberate and documented
- [ ] Complexity is proportional to the problem — the simplest change that works won

## Readability

- [ ] Names explain intent at call sites
- [ ] Comments explain why, not what; none restate the code
- [ ] Dead code is deleted, not commented out
- [ ] Diff reads top to bottom without needing the author's voice in your head

## Documentation and communication

- [ ] User-visible changes have doc updates (README, help text, changelog as the project uses)
- [ ] Migrations, flags, and deploy steps are written in the PR
- [ ] PR description stands alone: what, why, behaviour changes, testing — understandable in one read

## Reviewer's closing questions

- [ ] Would I be able to debug this at 3 a.m. from the logs it produces?
- [ ] If this breaks, how is it detected and how is it rolled back?
- [ ] Am I approving because it is right, or because arguing is tiring? (Accuracy over agreement.)
