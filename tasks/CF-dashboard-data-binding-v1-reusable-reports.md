---
id: CF-dashboard-data-binding-v1
status: in-progress
owner: Foundry report binding worker
repositories: [spik3r/clanker-foundry]
---

# Reuse a reviewed dashboard with fresh report data

## Implementation claim

Implement bounded dashboard data binding in the existing diagram module. A valid
reviewed dashboard/v1 model fixes columns, labels, chart choices and layout. A
separate strictly decoded payload pins its exact source bytes and replaces every
dataset's rows and every stat's value/note. The existing validator checks the
result, and the existing generator renders it. Add a `diagramgen bind TEMPLATE
DATA` command that writes only the materialized model to stdout, a public-safe
example, contract documentation and independent literal tests.

This addresses recurring reports where an author should supply changing data,
not regenerate layout or HTML. It neither infers data nor measures token savings.
No expressions, file inclusion, network input, new rendering profile, live data,
consumer integration, arbitrary field mutation or pipeline authority changes.

Validate with diagram race tests and vet, repository validation, bounded CLI
refusal tests and Project Memory committed-diff validation. Obtain independent
review before publication.
