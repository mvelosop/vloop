#!/usr/bin/env bash
# Shared set-up for the eval scaffolds. Sourced, never run. Every function works
# in the current directory, which must be empty.

EVALS="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

die() { echo "scaffold: $*" >&2; exit 1; }

# Refuse anything but an empty directory: a scaffold rewrites what it finds.
require_empty() {
  [ -z "$(ls -A .)" ] || die "run this in an empty directory"
}

git_init() {
  git init -q -b main . && git config user.email eval@test && git config user.name eval
}

# The runstat baseline: a correct loader and signals module, the worked-example
# fixture, and the tests that pin them. A planted defect replaces part of it.
baseline() {
  mkdir -p docs/briefs src/runstat tests .vloop/state .vloop/tmp
  cp "$EVALS/_brief.md" docs/briefs/runstat-cli.md
  cat >pyproject.toml <<'TOML'
[project]
name = "runstat"
version = "0.1.0"
requires-python = ">=3.9"

[tool.pytest.ini_options]
testpaths = ["tests"]
TOML
  printf '.pytest_cache/\n__pycache__/\n.vloop/tmp/\n' >.gitignore
  : >src/runstat/__init__.py
  cat >src/runstat/loader.py <<'PY'
"""Load a run directory: iterations.jsonl and sessions/*.json."""

import json
from dataclasses import dataclass
from pathlib import Path


class RunError(Exception):
    pass


@dataclass
class Run:
    iterations: list
    sessions: list


def load_run(path):
    d = Path(path)
    if not d.is_dir():
        raise RunError(f"run directory not found: {d}")
    try:
        log = d / "iterations.jsonl"
        iterations = [json.loads(l) for l in log.read_text().splitlines() if l.strip()] if log.exists() else []
        sessions = [json.loads(p.read_text()) for p in sorted((d / "sessions").glob("*.json"))]
    except ValueError as exc:
        raise RunError(f"malformed run: {exc}")
    return Run(iterations, sessions)
PY
  cat >src/runstat/signals.py <<'PY'
"""The eight run-level signals from section 7 of the brief."""


def compute_signals(run):
    """Return the signals for a loaded run."""
    records = run.iterations
    last = records[-1] if records else {}
    done = last.get("tasks_done", 0)
    streak = 0
    for r in reversed(records):
        if r["outcome"] == "done":
            break
        streak += 1
    return {
        "iterations": len(records),
        "tasks_done": done,
        "tasks_total": last.get("tasks_total", 0),
        "iterations_per_closed": len(records) / done if done else 0.0,
        "gate_failures": sum(1 for r in records if r["outcome"] == "gate_fail"),
        "review_rejections": sum(1 for r in records if r["outcome"] == "review_fail"),
        "attempts_burned": sum(1 for r in records if r["outcome"] != "done"),
        "no_progress_streak": streak,
        "estimated_spend": sum(s.get("total_cost_usd", 0) for s in run.sessions),
    }


def format_signals(s):
    """The eight key/value pairs, in the brief's order."""
    return [
        ("iterations", str(s["iterations"])),
        ("tasks closed", f"{s['tasks_done']}/{s['tasks_total']}"),
        ("iterations per closed", f"{s['iterations_per_closed']:.2f}"),
        ("gate failures", str(s["gate_failures"])),
        ("review rejections", str(s["review_rejections"])),
        ("attempts burned", str(s["attempts_burned"])),
        ("no-progress streak", str(s["no_progress_streak"])),
        ("estimated spend", f"${s['estimated_spend']:.2f}"),
    ]
PY
  cat >src/runstat/__main__.py <<'PY'
import sys

from runstat.cli import main

sys.exit(main())
PY
  cat >src/runstat/cli.py <<'PY'
def main(argv=None):
    return 0
PY
  cat >tests/fixtures.py <<'PY'
import json


def write_fixture_run(root):
    """The brief's worked example: 3 iterations, 2 of 8 tasks done, $4.08."""
    run = root / "run"
    (run / "sessions").mkdir(parents=True)
    rows = [
        {"iteration": 1, "task": "T1", "outcome": "done", "attempts": 1, "tasks_done": 1, "tasks_total": 8},
        {"iteration": 2, "task": "T2", "outcome": "gate_fail", "attempts": 1, "tasks_done": 1, "tasks_total": 8},
        {"iteration": 3, "task": "T2", "outcome": "done", "attempts": 2, "tasks_done": 2, "tasks_total": 8},
    ]
    (run / "iterations.jsonl").write_text("".join(json.dumps(r) + "\n" for r in rows))
    for i in range(6):
        (run / "sessions" / f"{i + 1:03d}.json").write_text(json.dumps({"total_cost_usd": 0.68}))
    return run
PY
  cat >tests/conftest.py <<'PY'
import os
import pathlib
import sys

import pytest

sys.path.insert(0, str(pathlib.Path(__file__).parent))
sys.path.insert(0, str(pathlib.Path(__file__).parent.parent / "src"))
# Subprocesses run `python -m runstat`, which pytest's own sys.path does not reach.
os.environ["PYTHONPATH"] = str(pathlib.Path(__file__).parent.parent / "src")
from fixtures import write_fixture_run


@pytest.fixture
def fixture_run(tmp_path):
    return write_fixture_run(tmp_path)
PY
  cat >tests/test_signals.py <<'PY'
from runstat.loader import load_run
from runstat.signals import compute_signals


def test_signals_for_worked_example(fixture_run):
    s = compute_signals(load_run(fixture_run))
    assert s["iterations"] == 3
    assert s["tasks_done"] == 2 and s["tasks_total"] == 8
    assert abs(s["iterations_per_closed"] - 1.5) < 1e-9
    assert s["gate_failures"] == 1
    assert s["review_rejections"] == 0
    assert s["attempts_burned"] == 1
    assert s["no_progress_streak"] == 0
    assert abs(s["estimated_spend"] - 4.08) < 1e-9
PY
  cat >tests/test_signals_cmd.py <<'PY'
import subprocess
import sys


def test_signals_command_prints_eight_lines(fixture_run):
    p = subprocess.run([sys.executable, "-m", "runstat", "signals", str(fixture_run)],
                       capture_output=True, text=True)
    assert p.returncode == 0, p.stderr
    lines = [l for l in p.stdout.splitlines() if l.strip()]
    assert len(lines) == 8, lines
PY
  cat >tests/test_errors.py <<'PY'
import subprocess
import sys


def test_missing_run_dir_exits_2(tmp_path):
    p = subprocess.run([sys.executable, "-m", "runstat", "signals", str(tmp_path / "nope")],
                       capture_output=True, text=True)
    assert p.returncode == 2
PY
}

commit_all() { git add -A && git commit -qm "$1"; }

# A task file from the calibration lacks what state/v1 requires of a task; add
# it, and wrap the task in a plan. $1 = task.json, $2 = kind, $3 = the brief.
write_state() {
  python3 - "$1" "${2:-feature}" "${3:-docs/briefs/runstat-cli.md}" <<'PY'
import json, subprocess, sys
task = json.load(open(sys.argv[1]))
task.setdefault("kind", sys.argv[2])
task.setdefault("references", [])
task.update(status="pending", attempts=0, notes="")
base = subprocess.check_output(["git", "rev-parse", "HEAD"], text=True).strip()
plan = {"schema": "state/v1", "run_id": "eval", "brief": sys.argv[3],
        "base": base, "branch": "vloop/eval", "status": "running", "iteration": 1,
        "created": "2026-10-01T00:00:00Z", "updated": "2026-10-01T00:00:00Z",
        "shell": "bash", "tasks": [task]}
json.dump(plan, open(".vloop/state/state.json", "w"), indent=2)
PY
}

# The report a work session would have written, claiming success. $1 = a
# proposal file lacking `schema`, or empty for the bland default.
write_proposal() {
  python3 - "${1:-}" <<'PY'
import json, sys
p = json.load(open(sys.argv[1])) if sys.argv[1] else {
    "task": "T1", "outcome": "done", "summary": "Implemented the task and verified it.",
    "files": [], "verified": "gate command exits 0", "notes": "none"}
json.dump({"schema": "proposal/v1", **p}, open(".vloop/tmp/proposal.json", "w"), indent=2)
PY
}

verify_of() { python3 -c 'import json; print(json.load(open(".vloop/state/state.json"))["tasks"][0]["verify"])'; }

# The planted defect must pass its own gate, or the case measures the gate.
gate_passes() {
  bash -c "$(verify_of)" >/dev/null 2>&1 || die "the planted defect fails its own gate: the case is invalid"
}

# Cases whose defect exploits a defective gate only reproduce that situation
# while a CORRECT implementation still fails the gate. $1 = the script that
# installs it. Run in a copy, so the repository is untouched.
gate_rejects_correct() {
  local verify copy; verify="$(verify_of)"; copy="$(mktemp -d)"
  git ls-files -z | xargs -0 tar -cf - | tar -xf - -C "$copy"
  ( cd "$copy" && bash "$1" && ! bash -c "$verify" >/dev/null 2>&1 ) || { rm -rf "$copy"; die "a correct implementation passes this gate: the case no longer reproduces a false-failing gate"; }
  rm -rf "$copy"
}

