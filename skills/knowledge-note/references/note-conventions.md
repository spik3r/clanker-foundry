# Knowledge-base note conventions

Use these only when the target vault has no stronger local conventions. Existing files,
instructions, and sibling notes take precedence.

## Placement and naming

- Put the note in an existing topic, project, area, archive, or inbox folder.
- Reuse sibling naming style. Use a descriptive, stable name; avoid dates in a title
  unless the vault uses daily or meeting notes.
- Do not create a new top-level folder or classification scheme without direction.

## Metadata

Use the fields already used by the vault. When none exist, use only the metadata that
helps discovery:

```yaml
---
created: YYYY-MM-DD
type: reference
aliases: [Readable alternate name]
tags: [topic]
---
```

## Note shape

- Start with a short `> summary` blockquote.
- Distil the topic; link canonical sources instead of duplicating them.
- Mark an important unknown with `❓` and a caution with `⚠️` only when the vault uses
  these markers or no convention exists.
- End with **Related** or **See also** links.

## Navigation

Use the vault's link style. Add the note to its parent index, map of content, or
equivalent navigation surface. Keep an index as a small hub of links and descriptions;
move detail into content notes.

## Verification

- Metadata parses and dates are current.
- Internal links resolve and have readable aliases where the vault uses them.
- Tags and note type match siblings.
- The note is reachable from an existing index or hub.
