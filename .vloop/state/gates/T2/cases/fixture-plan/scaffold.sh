#!/usr/bin/env bash
set -euo pipefail
HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
. "$HERE/../_lib.sh"
require_empty
make_tmp
mkdir -p .vloop/state
cp "$HERE/state.json" .vloop/state/state.json

