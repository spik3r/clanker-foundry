#!/usr/bin/env bash
# Run from modeling/lifecycle; stdout is the review HTML.
set -euo pipefail
export GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off GOWORK=off GOENV=off GOFLAGS=
test "$(go env GOVERSION)" = go1.27.3
test -f tutorial/model-v2.json
scratch=$(mktemp -d)
trap 'rm -rf -- "$scratch"' EXIT
root=$PWD
go build -o "$scratch/foundryreview" ./cmd/foundryreview
go build -o "$scratch/modelc" ./cmd/modelc
go build -o "$scratch/acceptance-v1" ./cmd/modelacceptance
mkdir "$scratch/candidate"
cp -R . "$scratch/candidate/"
(
 cd "$scratch/candidate"
 cp tutorial/model-v2.json examples/lifecycle.json
 "$scratch/modelc" validate >&2
 "$scratch/modelc" generate >&2
 "$scratch/modelc" check >&2
 go build -o "$scratch/acceptance-v2" ./cmd/modelacceptance
)
request() {
 python3 - "$1" "$2" "$3" <<'PY'
import json, pathlib, sys
manifest, model, suite = (pathlib.Path(p).read_bytes().decode("utf-8") for p in sys.argv[1:])
print(json.dumps(dict(formatVersion="lifecycle-acceptance-authoring-request/v1", mode="replay", manifest=manifest, model=model, suite=suite)))
PY
}
request tutorial/manifest-v1.json examples/lifecycle.json tutorial/suite-v1.json |
 "$scratch/acceptance-v1" > "$scratch/before.json"
request tutorial/manifest-v2.json tutorial/model-v2.json tutorial/suite-v2.json |
 "$scratch/acceptance-v2" > "$scratch/after.json"
cd "$root"
"$scratch/foundryreview" --tutorial "$scratch/before.json" "$scratch/after.json"
