---
name: knowledge-note
description: Create or update a connected note in a personal knowledge base, vault, or Obsidian workspace. Use when the user asks to capture, write up, organise, or improve notes about a topic, system, project, investigation, meeting, or decision; to add a note to a vault; or to bring notes into the vault's established structure. Follow existing conventions, links, metadata, and indexes; do not use for standalone documents outside a knowledge base.
---

# Knowledge Note

## Purpose

Create notes that are concise, findable, connected to their context, and consistent
with the knowledge base that owns them.

## Use when

- creating, updating, or reorganising notes in a vault or personal knowledge base;
- capturing an investigation, system overview, convention, reference, or decision;
- adding links and index entries so a new note is not orphaned.

## Do not use when

- producing a standalone report, document, or temporary plan outside a knowledge base;
- recording a project-level architecture decision: use `adr`;
- preserving task state between agent sessions: use `checkpoint` or `context-offload`.

## Default access and modes

Edit mode for the named vault only. **Create** adds a note and relevant index links.
**Update** revises an existing note without discarding useful history. **Organise** moves
or connects notes only after confirming the vault's conventions and links.

## Workflow

1. Confirm the vault root and requested outcome. Read applicable vault instructions,
   target-folder siblings, indexes, and templates before writing.
2. Select the existing folder and note shape. Reuse its naming, metadata, tags, link,
   and index patterns; do not invent a top-level taxonomy without user direction.
3. Distil the source into a one- or two-sentence summary, concise body, evidence or
   canonical-source links, explicit unknowns, and related notes.
4. Add the note to its parent index, map of content, or other established navigation
   surface. Create a hub only when the topic has enough distinct notes to justify one.
5. Use `references/note-conventions.md` and the bundled templates as fallbacks when
   the vault has no established convention.

## Validation

Check metadata syntax, file path, date, internal links, index entries, and consistency
with sibling notes. Distinguish evidence from assumptions; do not copy a canonical
source into a competing note when a concise summary and link will do.

## Output contract

Return:

- **Summary** — note purpose and location;
- **Evidence** — source material and canonical links used;
- **Changes** — created or updated notes and indexes;
- **Risks** — assumptions, unresolved links, or taxonomy uncertainty;
- **Next action** — only when additional capture or review is needed.

## Stop or hand back

Stop when the vault root, intended audience, or target taxonomy cannot be safely
inferred. Ask for the smallest clarifying decision rather than scattering notes or
creating a competing structure.

## Safeguards

Distil rather than transcribe. Preserve existing notes and links, keep indexes slim,
avoid secrets or sensitive personal data, and treat the vault's local conventions as
the source of truth.
