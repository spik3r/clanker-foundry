# Lifecycle JSON authoring contract

[lifecycle-v1.schema.json](lifecycle-v1.schema.json) is a draft-07 JSON Schema
for editors and AI-assisted drafting of `lifecycle-json/v1` with profile
`lifecycle/v1`. Point a draft-07-compatible editor at this local file; its `$id`
is an identifier, not a network endpoint. The independently specified
[cases.json](cases.json) contains a small complete valid model and examples of
structural and semantic errors. The existing [model](../examples/lifecycle.json)
is a larger example.

A schema pass means only that the parsed JSON value satisfies this authoring
shape. Submit the original bytes to `jsonformat.Decode`, which performs strict
JSON parsing, semantic validation and normalization; this is the authority for
acceptance. The schema cannot reliably detect duplicate object keys after a parser
has discarded them, malformed Unicode, trailing data, the one-MiB source limit,
expression depth/node budgets, or byte-based text limits. It also leaves graph
reachability, uniqueness, references, operand types/arity, required nonblank text,
and effect ordering to Go. A schema-valid document can fail Go validation. Do not
use a schema result as review, execution authority, or proof of correctness.

Authors should use unique ASCII element IDs, link the initial state, guards and
requirements to actual declarations, give every nonterminal an outgoing transition,
make every state reachable, and include a terminal state. Guard expressions use
only the supported references and operators. Effects are ordered: optional
`assign_owner_if_absent`, followed by `set_state` and `increment_revision`.
Consult the [lifecycle guide](../README.md) for the complete profile and review
workflow. This contract describes the current example, not UML or SysML support.

The schema and model case checks run in separate CI jobs. To run locally, install
the repository's pinned test-only Ajv dependency with `npm ci --prefix runtime
--ignore-scripts`, then run `node --test modeling/lifecycle/authoring/schema.test.cjs`
from the repository root and `go test ./authoring` from `modeling/lifecycle` using
its pinned Go toolchain. Node and Ajv are not dependencies of normal Go model
operations. The case corpus is finite; it detects covered drift, not equivalence
of every possible input. Update schema, cases and this guidance when the underlying
format/profile changes, retaining their explicit versions.
