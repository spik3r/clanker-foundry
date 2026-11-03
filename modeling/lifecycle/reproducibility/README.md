# Clean regeneration test

`go test ./reproducibility` copies the lifecycle module to a temporary directory,
removes the five compiler outputs and three requirements-trace views, then uses
Go 1.27.3 offline to regenerate them twice and compare each file byte for byte
with the checkout. It corrupts each output in turn and requires its existing
check command to fail, restores the bytes, and runs the frozen Go acceptance and
authoring cases against the rebuilt module. A final full-file comparison catches
changed or added module files, including schema and frozen expectations. The
original checkout is never written by the test.

The existing authoring CI job separately checks the JSON Schema with pinned Ajv;
the Go authoring cases check the semantic decoder. Neither schema acceptance nor
trace links prove semantic correctness. This test compares against the supplied
checkout, so it does not authenticate it or establish independent authorship.
It uses temporary test subprocesses only; it adds no product runner.
