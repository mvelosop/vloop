#!/usr/bin/env bash
# Builds the repository the review session runs in. Run it in an EMPTY
# directory. Planted defect: a correct signals command plus a --json flag the brief excludes.
set -euo pipefail
HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
. "$HERE/../_lib.sh"
require_empty
git_init
baseline
commit_all "baseline"
write_state "$HERE/task.json"
# The planted work is left UNCOMMITTED: the review reads `git diff HEAD`.
bash "$HERE/plant.sh"
write_proposal
gate_passes
