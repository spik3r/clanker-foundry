## What changed

<One short paragraph or a tight bullet list. A reviewer should understand the change
in one read.>

## Why

<The problem or ticket this addresses. Link the issue.>

## Behaviour changes

- <Before → after, observable from the outside. "None" is a valid answer — say it explicitly.>

## Testing

- <What was run and the result, e.g. `go test ./...` — pass>
- <What was not verified, and why>

## Migration / deployment

<Steps required beyond merge: migrations, flags, config, order of deploys. "None" if none.>

## Reviewer notes

<Where to start reading, decisions that need scrutiny, known follow-up work.>

---

- [ ] Diff is limited to the stated change
- [ ] Tests cover the new or changed behaviour
- [ ] No secrets, credentials, or personal data in the diff
- [ ] Docs updated where behaviour changed
- [ ] `checklists/code-review.md` walked before requesting review
