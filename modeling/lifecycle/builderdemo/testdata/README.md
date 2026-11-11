# Historical instruction test data

`AGENTS-882b5002.txt` is an exact copy of root `AGENTS.md` from commit
`882b5002bd4dc70b4be0060e5e5668d1faa8bbdd`. Its SHA-256 is
`4574ee25fe89b6658ae237f3f50648ef1e542c487a2c23d7ab2833696f9455eb`, matching
`guidance-2` in the unchanged [input manifest](../inputs.json). It is fixture
data, not current agent instructions. Do not refresh its bytes or hash when
current instructions change.

The builderdemo and modelbuilderdemo tests use this snapshot only for the
manifest's root instruction artifact. The other fourteen fixture inputs still
come from their existing paths and retain their original pins. All original
acceptance assertions remain in place.

This keeps historical replay separate from current-source acceptance. The
production preparer still checks exactly the original manifest, so changed
current instructions remain blocked. Supplementary API and CLI tests replay the
snapshot, verify the copied output bytes, and check actual current instructions
against the original bytes. They also always exercise a changed copy that must
be refused without source or output writes. If current instructions return to
the exact historical bytes, they are eligible under the same original rule.

No production fallback, refreshed manifest, builder execution or new authority
is added. The [fixed-input contract](../README.md#fixed-inputs-not-refreshed-pins)
still applies to callers.
