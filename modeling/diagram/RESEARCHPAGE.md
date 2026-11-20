# Profile `research-page/v1`

The fixed text of a research page, as a SolidJS component. `diagramgen` turns the
model into `ResearchPage.jsx`, which renders the host's page shell with the
title, purpose, status, bottom line, section headings and method list written in
as string literals. The parts that are interactive or computed (filters,
matrices, tables, metrics, freshness) are not in the model. The model names them
as slots and the host fills them. Selected by
`"profileVersion": "research-page/v1"`; the input format is `diagram-json/v1`,
as for [`dashboard/v1`](DASHBOARD.md).

**This profile is shaped around one host's page outline.** It encodes the
`ResearchPageShell` / `ResearchBottomLine` / `Panel` / `SectionHeader` outline of
one SolidJS app: a bottom line, then Explore, Method and Sources tabs. It is not
a general page generator. Another host with a different outline needs a
different profile, or an adapter module that maps these four names onto its own
components and accepts what it cannot map. Nothing in the profile checks the
host; the adapter contract below is the only coupling.

The idea behind the split: a person or an AI writes the model (data), and the
generator writes the code. The text that every research page of this outline
repeats in the same shape is data. The code that reads data files and filters
them is not, and stays hand-written.

Go validation is authoritative. The
[schema](authoring/research-page-v1.schema.json) is an authoring aid and is
looser in the places its own description lists.

## Input

```json
{
  "formatVersion": "diagram-json/v1",
  "profileVersion": "research-page/v1",
  "id": "EvidenceAtlas",
  "title": "Evidence Atlas",
  "purpose": "Compare measured, sparse, and unmeasured evidence ...",
  "status": "mixed",
  "class": "stat-research-page",
  "bottomLine": {
    "finding": "The atlas records where evidence exists ...",
    "confidence": "Exploratory",
    "limitation": "Sparse and unmeasured rows describe data coverage ...",
    "context": ["Paired deltas are not causal proof. ..."]
  },
  "sections": [
    {"tab": "explore", "items": [
      {"kind": "slot", "id": "coverage"},
      {"kind": "heading", "title": "Regime matrix", "description": "..."},
      {"kind": "panel", "slot": "matrix"}]},
    {"tab": "method", "items": [
      {"kind": "methodPanel", "title": "Method and limitations", "description": "...", "points": ["...", "..."]}]},
    {"tab": "sources", "items": [
      {"kind": "heading", "title": "Sources and provenance", "description": "..."},
      {"kind": "panel", "slot": "provenance"}]}
  ]
}
```

Full example: [evidence-atlas-page.json](examples/evidence-atlas-page.json),
written from the text of the host's `StatisticalResearchPage.jsx`. Its output is
under [examples/generated/evidence-atlas-page](examples/generated/evidence-atlas-page).

Required: `formatVersion`, `profileVersion`, `id`, `title`, `purpose`, `status`,
`bottomLine` (with `finding`, `confidence`, `limitation`) and `sections`.
Optional: `class`, `bottomLine.context`, and `description` on a `heading` or
`methodPanel`.

### Members

| Member | Meaning |
| --- | --- |
| `id` | ASCII identifier `[A-Z][A-Za-z0-9]{0,47}`. The component is exported as `<id>Content`. |
| `title` | page title, 1 to 120 bytes |
| `purpose` | one paragraph under the title, 1 to 600 bytes |
| `status` | one of `supported`, `mixed`, `inconclusive`, `experimental`, `unavailable`, `not-yet-run`, `unknown`: the values the host's `ResearchBottomLine` accepts (its `STATUS_COPY` table). Any other value is refused. |
| `class` | optional CSS class token for the shell, `[a-z][a-z0-9-]{0,63}`; the only presentation the model has |
| `bottomLine.finding`, `.limitation` | 1 to 600 bytes |
| `bottomLine.confidence` | 1 to 120 bytes |
| `bottomLine.context` | optional, 1 to 6 paragraphs of 1 to 600 bytes, shown inside the bottom-line card |
| `sections` | 1 to 3 sections; each `tab` is `explore`, `method` or `sources` and appears at most once |

### Items

A section holds 1 to 24 items, in the order they are shown.

| Kind | Members besides `kind` | Emits |
| --- | --- | --- |
| `slot` | `id` | the host's content, inserted bare |
| `panel` | `slot` | the host's content inside `<Panel>` |
| `heading` | `title`, optional `description` | `<SectionHeader title description />` |
| `methodPanel` | `title`, optional `description`, `points` (1 to 12) | `<Panel>` with a `<SectionHeader>` and a `<ul>` of `<li>` |
| `prose` | `paragraphs` (1 to 6) | one `<p>` per paragraph |

A member that the kind does not use is refused, even when empty. A `slot` item
names its slot with `id`, a `panel` item with `slot`; both define one slot.

Slot ids match `[a-z][A-Za-z0-9]{0,31}` and are unique across the whole model
(slots and panels, every tab). The names `constructor`, `hasOwnProperty`,
`isPrototypeOf`, `propertyIsEnumerable`, `toLocaleString`, `toString` and
`valueOf` are refused, because `props.slots.<id>` would find an inherited
function on a plain object.

## Strict decoding

As for the dashboard profile (the shared strict scanner): unknown members (a case
variant such as `Title` is unknown), duplicate JSON keys, `null`, an object or
array where a scalar belongs, trailing data after the model, invalid UTF-8 or
broken surrogate escapes, nesting deeper than 16 and input larger than 512 KiB
are refused. `class`, `context` and `description`, when present, must not be
empty. There is no explicit default spelling of an absent member.

## Validation

- `formatVersion` and `profileVersion` are exactly the values above.
- `id`, `status`, `class` and every slot id follow the patterns above.
- Text is plain: not blank, no control characters, no U+2028, U+2029, U+FFFE or
  U+FFFF, valid UTF-8. Titles and `confidence` are 1 to 120 bytes. Every other
  text (purpose, finding, limitation, context, descriptions, points,
  paragraphs) is 1 to 600 bytes. Text is plain text: there is no markup, no
  link and no line break.
- Counts: 1 to 3 sections, 1 to 24 items per section, 0 to 6 context
  paragraphs, 1 to 12 points, 1 to 6 paragraphs.
- Tabs are unique. Slot ids are unique.

## Canonical encoding

Input to the model digest (`canonicalSHA256`): compact JSON, no trailing
newline, no HTML escaping, members in this fixed order:

- model: `formatVersion`, `profileVersion`, `id`, `title`, `purpose`, `status`,
  `class` (only when present), `bottomLine`, `sections`.
- bottomLine: `finding`, `confidence`, `limitation`, `context` (only when
  present).
- section: `tab`, `items`.
- item: `kind`, then those of `id`, `slot`, `title`, `description`, `points`,
  `paragraphs` that the kind uses.

Sections are sorted into the order `explore`, `method`, `sources`, which is the
order the host's tabs show them. Item order, point order, paragraph order and
context order are significant and kept. JSON whitespace, member order in the
source and section order do not change the digest. Any text, id, status, class,
item, slot or order-within-a-list change does.

## Output files

| File | Content |
| --- | --- |
| `ResearchPage.jsx` | the component |
| `slots.json` | the slot ids the host must supply, in output order (explore, method, sources; items in order), as a two-space indented JSON array with a trailing newline. `[]` when there are none. |
| `model.json` | the canonical model, two-space indented, trailing newline |
| `foundry-manifest.json` | integrity record: generator `diagramgen` 1.4.0, toolchain constant `go1.27.4` |

`ResearchPage.jsx` carries the contract's generated-file marker as its first
line: `// Code generated by diagramgen 1.4.0 from model <canonicalSHA256>. DO NOT EDIT.`

### ResearchPage.jsx

- One default export, `export default function <Id>Content(props)`.
- The only module it imports from is `"../researchPageParts"`, the host adapter.
  It imports only the names it uses, from `Panel`, `ResearchBottomLine`,
  `ResearchPageShell` and `SectionHeader`. It imports nothing from `solid-js`
  at present, because a static list needs no `For` or `Show`; `solid-js` is the
  only other module an output may import from.
- It renders `<ResearchPageShell>` with `class`, `title`, `purpose`, `status`,
  `freshness`, `provenance`, a `summary` prop holding `<ResearchBottomLine>`, and
  one prop per section (`explore`, `method`, `sources`) holding a fragment of
  that section's items. A tab with no section gets no prop.
- Model text is never JSX source. Every text child is `{"..."}` and every
  attribute that holds model text is `name={"..."}`, a JavaScript string
  literal. A literal escapes `"`, `\`, control characters, U+2028, U+2029 and
  `<`, `>` and `&` (as `\u` escapes). Braces, backticks and `${` are inert
  inside a literal and are left as written. The same escaper writes the
  dashboard's Solid output.
- Slots: `{renderSlot(props.slots.<id>)}`, where `renderSlot` returns the value,
  or calls it first when it is a function. A slot value is therefore a JSX
  element or a function returning one. Each slot id is referenced exactly once.
  `renderSlot` is emitted only when the model has a slot. `props.slots` must be
  an object when the model has a slot; a missing id renders nothing.
- Pass-through: `props.freshness` goes to the shell, `props.provenance` to the
  shell, `props.metrics` to `ResearchBottomLine`, and `props.class` is appended
  to the model's `class` (or is the whole class when the model has none). These
  are computed, so the model cannot hold them. Nothing else of `props` is read.
- No `innerHTML`, event handler, `fetch`, URL, `eval`, `style` attribute,
  `href`, `src`, `document` or `window`. Nothing is computed in the browser by
  the generated code beyond the string concatenation for the class and the
  `renderSlot` call.

## Adapter contract

The host supplies `researchPageParts` in the parent of the output directory, next to it rather than inside it, so `diagramgen check` sees only generated files. It must export
four names. The generator does not check them; a host that renames or changes
them breaks the page at build time or run time.

| Export | Called with |
| --- | --- |
| `ResearchPageShell` | `class`, `title`, `purpose`, `status` (a status string), `freshness`, `provenance`, `summary`, and `explore`, `method`, `sources` when the model has that section. Each of those is a JSX value. |
| `ResearchBottomLine` | `status`, `finding`, `confidence`, `metrics`, `limitation`, and the context paragraphs as children |
| `Panel` | children only |
| `SectionHeader` | `title` and, when the model has one, `description` |

In the host this was written for, the adapter is a re-export of the existing
components, with `Panel` and the shell free to add the host's own classes:

```js
export { default as ResearchPageShell } from "../components/ResearchPageShell";
export { default as ResearchBottomLine } from "../components/ResearchBottomLine";
export { default as Panel } from "../components/Panel";
export { default as SectionHeader } from "../components/SectionHeader";
```

Presentation that the model cannot carry has to live in the adapter or in the
host's style sheet, keyed on the shell `class`. For example the host's
hand-written page gives each panel a `stat-effect-panel` class and the context
paragraph a `stat-research-context` class; the generated page emits plain
`<Panel>` and `<p>`, so the adapter's `Panel` would add the panel class and the
style sheet would reach the context paragraphs through the shell class.

A page using the output provides the slots and the computed props:

```jsx
<EvidenceAtlasContent
  freshness={freshnessText()}
  metrics={metrics()}
  provenance={provenance()}
  slots={{ coverage: () => <AtlasCoverage ... />, matrix: () => <Matrix ... /> }}
/>
```

Passing a function (rather than an element built at the call site) lets Solid
create the content inside the tab that shows it.

## What it cannot express

- a page outline other than the host's: summary (bottom line), explore, method,
  sources, in that fixed shell
- anything interactive or computed: filters, segmented controls, matrices,
  tables, charts, URL state, loading states or data from files. These are the
  slots, and stay hand-written
- freshness, metrics or provenance values: they are props, filled at run time
- more than one section per tab, other tabs, a custom tab label or a default tab
- presentation: no colours, classes beyond one shell class, panel classes,
  tones, spacing or icons
- links, images, emphasis, line breaks or any markup inside text
- a heading with actions (the host's `SectionHeader` accepts children; the model
  has none), page actions, a breadcrumb or group label, or the host's loading and
  error states
- nested lists, numbered lists, or a list outside a `methodPanel`
- checking that the host's shell, slots and adapter agree with each other: the
  output is correct for the contract above and nothing more
