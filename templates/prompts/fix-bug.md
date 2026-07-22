# Prompt template: fix a bug

Copy this block, fill in the placeholders, and delete any section that does not apply.
Pair with `workflows/bug-investigation.md` when the cause is not yet known.

```text
Use engineer in repair mode. Follow the bug investigation workflow.

## Symptom
<what goes wrong, observed from the outside: error message, wrong output, bad state>

## Expected behaviour
<what should happen instead>

## Reproduction
1. <step>
2. <step>
3. <step>

Reproduces: <always | sometimes: rate or conditions>

## Environment
- Version / commit: <sha or tag>
- Runtime: <OS, language version, dependencies that matter>
- Config: <flags or settings that differ from defaults>

## Evidence so far
- Logs / traces: <excerpts or links>
- First seen: <date, release, or commit>
- Suspected area: <paths or "unknown">

## Constraints
- <behaviour that must not change>
- Out of scope: <known adjacent problems to leave alone>

## Done means
- Reproduction steps now produce the expected behaviour
- A regression test covers the failure
- Relevant tests pass: <command>
```

## Notes

- If the bug cannot be reproduced, say so in the prompt and ask for an evidence-gathering pass first.
- Include the actual error text, not a paraphrase.
