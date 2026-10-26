#!/usr/bin/env bash
# Builds the repository the review session runs in. Run it in an EMPTY
# directory. Case: correct work that passes its gate, in a run whose records
# under .vloop/ hold an absolute home path and the placeholder text of an
# earlier finding. The path is composed here, so no file of the case holds it.
# The review must PASS, with one line in notes about the records.
set -euo pipefail
HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
. "$HERE/../_lib.sh"
require_empty
git_init
baseline
commit_all "baseline"
write_state "$HERE/task.json"

# The records: a plan and a run report the driver and planner wrote, with an
# absolute home path, and a finding's placeholder text copied into the report.
abs="/$(echo Users)/builder/work/runstat"
mkdir -p .vloop/state/runs .vloop/tmp/checks
cat >.vloop/state/plan.md <<PLAN
# Plan

Planned in $abs on the day of the run.

- T1: expose the package version
PLAN
cat >.vloop/state/runs/report.md <<REPORT
# Run report

- earlier finding: absolute path "$abs/src/runstat/cli.py" in a file
- placeholder text of that finding: "/$(echo Users)/<name>/..."
REPORT

# The work: correct, and it passes its own gate.
cat >src/runstat/version.py <<'PY'
"""The package version, in one importable place."""

VERSION = "0.1.0"
PY
cat >tests/test_version.py <<'PY'
from runstat.version import VERSION


def test_version_is_the_packaged_one():
    assert VERSION == "0.1.0"
PY

# The gate leaves the suite to the check below, which ran and passed.
printf 'tests/test_version.py .  [100%%]\n1 passed\n' >.vloop/tmp/checks/all.log
cat >.vloop/tmp/checks.json <<'JSON'
{"checks": [{"name": "all", "exit": 0, "log": ".vloop/tmp/checks/all.log"}]}
JSON
write_proposal
gate_passes
