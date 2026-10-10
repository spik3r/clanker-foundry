---
name: ui-verification
description: Verify visible browser behaviour after UI changes or during a scoped frontend review. Use for rendering, interaction, layout, loading, error, navigation, accessibility, and theme checks. Do not use for backend-only changes with no browser-visible effect, and do not treat this skill as permission to edit code or mutate production data.
---

# UI Verification

Exercise the real interface and report observed behaviour separately from build and
test results.

## Establish the target

1. Read the repository's run, test, and browser instructions.
2. Identify the changed surface, its callers, expected behaviour, and relevant states.
3. Record the revision or working-tree state, browser, viewport, and environment tested.
4. Verify an existing server serves the intended checkout before reusing it. Otherwise
   start the repository's documented local workflow when authorised.
5. Use local fixtures or approved test accounts. Do not create real transactions,
   send messages, change production data, or share private data just to verify a UI.

Do not assume a framework, start command, port, route, account, or deployment target.
A live server can be healthy while serving the wrong revision.

## Verify the affected flows

Select cases according to the change rather than performing a fixed ritual:

- Confirm the expected content renders and loading indicators settle.
- Exercise the changed action and verify its visible result, not just its click handler.
- Check relevant empty, slow, error, disabled, and retry states.
- Repeat the action; check duplicate submissions and stale or out-of-order responses.
- Interrupt where relevant: Cancel, Close, navigation, Back/Forward, and reload.
  Verify dismissed UI stays dismissed and state returns to a coherent screen.
- Check common and narrow viewport sizes for overlap, clipping, unreachable controls,
  and unexpected scrolling.
- For style changes, check supported theme, contrast, accent, and density settings.
  Include already-mounted components when settings can change at runtime.
- For interaction changes, check keyboard access, focus order, visible focus, labels,
  and focus restoration after dismissal.
- Inspect the browser console and relevant network failures. Distinguish pre-existing
  noise from errors introduced by the change.

Test only the states that bear on the requested scope. Record important omitted states.

## Evidence

For each material finding, record:

- environment and revision;
- route or surface and initial state;
- concise reproduction steps;
- expected and observed behaviour;
- relevant screenshot, log, or failing request when it helps establish the issue.

Keep secrets, session tokens, and personal data out of saved evidence. Inspect screenshots
before sharing them. Recheck affected flows after later edits, rebuilds, or rebases.

## When browser access is unavailable

Use meaningful fallbacks such as targeted tests, compilation, static checks, or a local
HTTP response. State exactly what each establishes. None proves that the rendered
interface works, and none should be reported as a visual or browser pass.

## Output

Report the tested target, covered flows, findings, and verification limits. Distinguish
passed, failed, blocked, and not-run cases. If the review is read-only, return defects
and reproduction steps without editing the implementation.

Stop only servers or sessions started for this task and no longer needed; preserve
pre-existing services and other workers' sessions.
