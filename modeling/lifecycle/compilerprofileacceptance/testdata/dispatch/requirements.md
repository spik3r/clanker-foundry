# Frozen dispatch requirements

This invented signal model is independent of the baseline workflow. Queue and
ReviewBay accept the same `advance` command but choose T_queue and T_accept
respectively. Queue advancement assigns an absent owner even though the command
is not named start. A present empty owner is retained. ReviewBay pulse is a
self-loop that increments once. Accepted and Discarded are terminal.

Both declared guards on T_queue fail in lexicographic ID order. Invalid inputs
return no Task; a known but illegal state/command pair returns an owned unchanged
Task before guard or overflow evaluation. Guards precede revision overflow.
Success and every valid-input refusal own their Task slice and owner pointer.
The oracle compares complete Results, including nonnil empty success refusals.

Declared effect order is assign_owner_if_absent, set_state, increment_revision
for T_queue; set_state, increment_revision for T_accept, T_pulse, T_discard.
Final-state observations cannot prove the relative order of commuting effects;
the harness must inspect that order separately.

Independent identity derivation: Python json plus hashlib SHA-256 over the model
with formatVersion removed, all four semantic collections sorted by ID, guard
references sorted, object keys sorted, compact UTF-8 and no trailing newline.
There are no all/any expressions in this model. Effect and comparison operand
orders are preserved. JSON keeps HTML characters literal and escapes U+2028/2029.
No Foundry encoder, decoder, generator or reducer supplied expected values.

Frozen canonical model SHA-256: `e4b871a029bb81d19df3428c6834f827986042daf50b0eef84e32faad252c012`.
