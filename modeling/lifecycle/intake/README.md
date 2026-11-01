# Design intake and artifact identity handoff

This offline example records the original intent, explicit decisions and their
rationales, and unanswered questions. It also records separate caller-declared
identities for a requirements artifact, a model artifact and a scenario suite.
It does not create, validate, inspect, execute, package, or approve any of them.
It is independent of the held workflow acceptance, review-pack and export lanes.

From `modeling/lifecycle`, with Go 1.27.3 and offline modules:

```sh
GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off go run ./cmd/modelintake < request.json
GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off go test ./intake ./cmd/modelintake
```

The CLI accepts one UTF-8 JSON value on stdin, no arguments or paths. All fields
shown are required; each artifact declaration and each observation may be null.
The example uses `abc` only to demonstrate hashing: it is not a requirements file.

```json
{
  "formatVersion": "lifecycle-intake-json/v1",
  "intent": "Capture the intended outcome before authoring.",
  "decisions": [{"id":"D_scope","text":"Use the lifecycle example.","rationale":"Keep this bounded."}],
  "questions": [{"id":"Q_owner","text":"Who owns the next decision?"}],
  "artifacts": {
    "requirements": {"sha256":"ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad","author":"claimed requirements author"},
    "model": null,
    "scenarioSuite": null
  },
  "dependencies": {"modelRequirements":null,"suiteRequirements":null,"suiteModel":null},
  "observed": {"requirements":"abc","model":null,"scenarioSuite":null}
}
```

An observation is the **literal UTF-8 artifact text**, encoded as a JSON string,
not a pathname, digest, embedded JSON object, or base64. The digest is SHA-256
of the decoded string's exact UTF-8 bytes, including whitespace and final newline.
Only supply an observation actually read from the artifact being compared. The
API and CLI perform no filesystem reads of artifact paths and cannot tell whether
an observation is current or authentic. Non-UTF-8 artifacts cannot be observed by
this v1 transport. Source input, original intent, author labels and observations
may be sensitive; the report repeats intent, decisions, questions and author
labels, but not observed artifact text. Keep requests and reports appropriately
private. The report's request digest covers the exact input bytes.

For each of the three fixed roles the report contains the declaration, the observed
hash or null, and exactly one status:

- `missing-declaration`: neither declaration nor observation was supplied.
- `undeclared`: observation supplied without a declaration.
- `missing-observation`: declaration supplied but no observation supplied.
- `digest-mismatch`: both supplied, but observed bytes differ from the declaration.
- `matches-declaration`: the observed bytes match the declared hash.

The required `dependencies` object records three independent *assertions*: the
requirements digest the model was authored against, and the requirements and
model digests the scenario suite was authored against. Each can be null when
unknown. The three fixed `links` in the report compare those assertions to the
corresponding declared artifact hashes. Link statuses are `missing-assertion`,
`missing-artifact-declaration` (either end absent), `digest-mismatch`, and
`matches-declaration`. These are caller statements: the tool does not inspect an
artifact to discover or verify its actual dependencies. A link match also does
not imply its artifact observations match; check both sets of statuses.

A match is a byte-consistency check, not evidence of independent authorship,
format or model validity, test execution, a passing test, review, or approval.
Declared author names are unverified labels and may coincide. Mismatches and
missing items are valid reports, not successful qualification. Observe all three
current artifacts and carry the resulting report forward for human interpretation;
never silently replace a declaration to make it match. Exit 0 means a report was
produced, even when it reports gaps; exit 2 means invalid request or input failure,
and exit 3 means output failed or was short. Errors produce no report, although a
failed output write can leave a partial byte stream.

The format rejects duplicate, missing, unknown and case-variant keys, null where
not explicitly allowed, trailing data, invalid Unicode, blank required text,
invalid IDs and SHA-256 values, and excess bounds. Intent is limited to 16384
UTF-8 bytes; decision text, rationale and question text to 4096 bytes each;
author labels to 256 bytes. Up to 64 decisions and 64 unanswered questions are
allowed; IDs are globally unique across these lists and use the model profile's
ASCII identifier rule. Each observation is at most 1 MiB; the whole request is at
most 4 MiB. Decisions and questions appear sorted by ID; bindings appear in fixed
requirements/model/scenario-suite order. Given identical input bytes, output is
byte-identical. Changing whitespace in the request changes its request hash.
