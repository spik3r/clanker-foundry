# Foundry profiles: choosing one and drafting against it

Use this route when the work might fit any registered Foundry profile, not only
`lifecycle/v1`. For the lifecycle handoff steps use the
[Foundry handoff reference](foundry-handoff.md). This reference does not install
anything: it needs an existing trusted Foundry checkout, and an installed skill
does not imply one. If there is none, deliver the brief, name the missing tool
context and do not invent results of checks that did not run.

## The rule for the generated path

On the generated path the agent's output is data: a model, expectation literals,
or a gap report. It is never feature code. Every byte of feature code comes from
the accepted model, a reviewed generator and a pinned toolchain. Do not write
code to cover what a profile cannot express, and do not hide the gap in prose.

## Find the registry

In a trusted Foundry checkout, read `modeling/profiles.json`. Check the checkout's
revision and its `AGENTS.md` first, and do not clone or install anything without
authority. Each entry names:

- `profile` and `generator`, and a one-sentence `summary`;
- `profileDoc`, the schema, and `examples`, all as paths relative to the checkout
  root;
- `emits`, what the generator writes;
- `cannotExpress`, short statements of what the profile cannot say.

`modeling/README.md` explains the registry and lists the profiles.

## Choose a profile from the brief

1. Read each entry's `summary`. Drop profiles that plainly do not match the
   brief.
2. Compare every requirement in the brief with the `cannotExpress` list of each
   remaining profile. A requirement that matches an entry there is a gap for
   that profile.
3. For the best remaining candidate, read its `profileDoc` and schema. The
   profile document is the boundary; the registry lines are a summary of it.
4. Read the example models to see the shape. Do not copy their content as
   answers for a different brief.
5. State which requirements fit, which need work outside the model, and which
   block the proposal. If no profile fits, or a requirement is outside the chosen
   profile, write a [gap report](../assets/gap-report.md). Do not substitute code,
   an invented model field or a prose workaround.

## Draft and check the model

Write the model as the profile document and schema describe. Use the schema in an
editor or validator for a first check; it is an authoring aid, and each schema
lists the rules it cannot express. The authoritative check is the generator
itself, which refuses a bad model without writing anything:

```sh
GEN validate MODEL
```

`GEN` is the generator binary built from the checkout (see `modeling/CONTRACT.md`
for the commands, the exit codes and the identity line). Exit 0 means the model is
accepted; exit 2 means it was refused and stderr says why. Build and run
the generator only where that is authorized. Report what ran and what did not.

An accepted model proves that the model is well formed for the profile. It does
not show that the model means what the person wanted, and the generated output
repeats a wrong model perfectly. Keep the expectation literals separate from the
model, as in the main procedure, and hand the model to the person to review.

## What the agent hands over

One of these, never feature code:

- a model that the generator's `validate` accepted, with the identity line it
  printed, and the requirements it does not cover;
- expectation literals written separately from the model;
- a gap report for every requirement that does not fit.

Expanding a profile, writing a hand-written seam, or leaving a feature out are
decisions for a named person. Present them as options in the gap report; do not
choose one.
