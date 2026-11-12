# Integrate a generated master/detail browser

Foundry already generates a bounded responsive browser. Use the
[profile](PROFILE.md) to choose fields and breakpoints and the
[compiler guide](README.md) to generate it. Do not hand-edit emitted Go files.

The public [Flexgo responsive reference](https://github.com/spik3r/flexgo/tree/547738cdaaaed3a5cfd22c937bbf16aa01159ff0/example/responsive)
shows the same composition pattern: a list and detail beside each other when
space permits, one active pane on a small terminal, and a resize hint below the
minimum. It is a reference app, not a new Flexgo component API to import.

## What the application owns

```text
reviewed JSON model ── compiler ── generated recordbrowser package
                                        ▲
application snapshots + terminal size ──┤
                                        ▼
              wide: list | detail   narrow: active pane   tiny: resize
```

1. Supply typed records through `New` or `Replace`. Use stable IDs, not list
   indices, when records may reorder. The generated browser copies and validates
   snapshots; the application keeps ownership of fetching and actions.
2. Keep the returned Bubble Tea model between updates. Recreating it on every
   resize would discard identity, focus and offsets. Store the returned value
   from `Replace`, including after loading/error transitions.
3. Choose width **and height** breakpoints from the content. An unusually wide
   but short terminal still uses one pane when it is below the wide height.
4. Keep leaf output bounded by terminal cells. The generated backend sanitises
   external text and clips it; the generic Flexgo example does not promise
   sanitisation of arbitrary control sequences.

The model selects layout and data presentation; it cannot add commands, network
access, arbitrary imports or editable fields. There is no framework-wide
breakpoint policy or new public API in this integration.

## A small screen must still explain how to leave

At the minimum 24-column supported width, a populated list shows:

```text
enter open ? help q quit
```

Details show `esc back ? help q quit`. Wider footers add movement and pane-switch
hints only when the complete line fits. Help shows close/scroll rather than quit,
because it captures other keys. Empty, loading and failed snapshots show only
help and quit. The full key list is available through `?`.

A small desk can hold either the index or an open book, not both. Enter opens the
book and Esc returns to the index. The bookmark belongs to the reader, not the
desk, so changing the desk size must not select a different record.

## Scope and a deliberate difference from the reference

Foundry preserves selected ID and focus across resizing. Its documented offset
contract **clamps stored offsets** at usable sizes, including during loading and
failure. Growing then shrinking must not resurrect an old clamped detail offset.
Tiny sizes freeze offsets. The newer Flexgo reference instead retains scroll
intent and clamps only while rendering. Do not copy that different state policy
into generated consumers without a reviewed contract change.

Neither renderer wraps arbitrary prose. Snapshot replacement, removed IDs,
loading/error recovery and captive help are Foundry-specific responsibilities;
the static Flexgo example uses indices and has no reload lifecycle.

## Rehearse before wiring a consumer

Using the synthetic generated demo, select a later record, open its details,
resize wide → narrow → tiny → wide, then return to the list and repeat. Test a
same-ID reordered snapshot, a removed selected ID, and loading/error transitions.
Check both minimum dimensions and non-ASCII text. Do not feed private scanner
output into public test fixtures.

The generated acceptance suite exercises baseline and changed JSON models,
repeated navigation, resize bounds, unchanged rendering state and footer hints.
Run the commands in the compiler guide. Deterministic tests do not substitute
for an interactive terminal check or downstream application trial.
