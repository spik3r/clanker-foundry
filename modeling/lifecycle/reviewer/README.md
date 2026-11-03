# Reviewing intake alongside requirements trace

The intake records decisions, unresolved questions, author claims and exact-byte
artifact identities; the requirements trace records requirement-to-model and
requirement-to-acceptance declarations. Neither records which intake decision or
question a requirement addresses. A human can use the [worksheet](worksheet.md)
to state and scrutinize those relationships while retaining the separate evidence
identities, gaps and unrun checks. The [worked example](example.md) uses an
invented intake and a small selection of the existing example trace.

To inspect the example from `modeling/lifecycle`, use Go 1.27.3 and an offline
Go cache:

```sh
GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off go run ./cmd/modelintake < reviewer/example-intake.json
GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off go run ./cmd/modeltrace validate
GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off go run ./cmd/modeltrace check
GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off go test ./reviewer
```

These commands inspect or test local artifacts; they do not authenticate their
origin or perform a human review. `modeltrace validate` reports valid gaps;
`check` compares generated trace views, and neither executes acceptance cases.
The unit test verifies the worked example's stated identities and selected
relationships against the existing APIs. It is not an automated review or a
complete acceptance suite. Reports and worksheets can contain sensitive material;
handle them under the same access rules as their inputs. Consult the
[intake contract](../intake/README.md) and [trace contract](../traceability/README.md)
for their exact limits and meanings.
