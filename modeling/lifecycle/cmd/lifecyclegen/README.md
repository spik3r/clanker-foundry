# lifecyclegen

`lifecyclegen` is the [Foundry generator contract v1](../../../CONTRACT.md)
command for the `lifecycle/v1` profile. It takes arbitrary paths:

```sh
lifecyclegen validate MODEL
lifecyclegen generate MODEL OUT
lifecyclegen check MODEL OUT
```

Any other argument shape exits 2. `generate` creates a new `OUT` (an existing
path is refused) holding `lifecycle.go`, `model_paths_test.go`, `decision.svg`,
`decision-table.json` and `foundry-manifest.json`. The two Go files are the
compiler output with one contract header line prepended. There is no
`trace.json`; the manifest replaces it. Generator id `lifecyclegen`, version
`1.0.0`, toolchain `go1.27.3`.

`modelc`, its fixed-path commands and the committed `generated/` directory are
unchanged.
