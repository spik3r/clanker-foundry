# Frozen trace observations

A separate assistant contributor wrote `oracle.json` from static inspection of
the v1 contract and shared evaluator before the trace verifier or replay test
was implemented. It did not run or import the evaluator or copy output from a
run. The invented spec digest was calculated independently from the canonical
serialization contract. Contributor separation is not independent-person
attestation, human approval or proof that requirements are correct.

The cases cover an ADMIT event to PENDING and four blocked ASSIGN attempts,
including omitted and false dependency/resource assertions. Inputs are invented.
They describe selected output fields and a full ADMIT event, not all evaluator
behavior. No success ASSIGN, dispatched worker, external state or candidate
model is involved. The SHA-256 pinned by the replay test protects fixture bytes
from silent refresh; a hash does not authenticate authorship or correctness.
