# ui-verification

Checks the browser-visible result of a frontend change: rendering, interaction,
navigation, layout, accessibility, and relevant loading or failure states.

## When to use

- after changing visible UI behaviour;
- reviewing a frontend regression;
- checking repeated, interrupted, or responsive flows;
- verifying theme, density, keyboard, or focus behaviour.

Skip it for backend-only work with no visible browser effect. It does not grant
permission to edit code or mutate production data.

## How it behaves

Uses the project's documented environment and verifies the server serves the intended
revision. Exercises relevant flows in a real browser, collects concise reproduction
evidence, and separates browser observations from build and unit-test results.
If browser access is unavailable, it reports that limit and useful fallback checks.

## Output

Tested environment and revision, covered flows, findings, and unverified cases.
Screenshots and logs exclude private data.

## Example invocations

```text
Use ui-verification to check the new search panel at desktop and narrow widths.
Use ui-verification in a read-only review of Cancel and Back/Forward behaviour.
```

## Installation

See the repository root `README.md`. This skill has no dependency on another skill.
