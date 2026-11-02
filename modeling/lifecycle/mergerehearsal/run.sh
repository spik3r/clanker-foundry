#!/usr/bin/env bash
set -euo pipefail
if [ "$#" -ne 0 ]; then
  printf '%s\n' 'merge rehearsal: no arguments accepted' >&2
  exit 2
fi
here="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd -P)"
exec python3 -I "$here/run.py"
