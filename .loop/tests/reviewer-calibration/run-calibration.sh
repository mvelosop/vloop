#!/usr/bin/env bash
#
# Reviewer calibration: does a real review session reject work that passes its
# gate but violates its acceptance criteria?
#
#   .loop/tests/reviewer-calibration/run-calibration.sh [case ...]
#
# Two runs of the loop produced 21 work/review pairs and zero rejections. That
# is either a reviewer with nothing to catch or a reviewer that cannot catch —
# and no amount of clean runs distinguishes those. This plants the defect
# instead of waiting for one.
#
# Each case sets up an isolated repo, plants a defective implementation, writes
# the proposal a work session would have written, and invokes ONE real review
# session. The work session is deliberately absent: the question is entirely
# about the reviewer's discrimination.
#
# **Every case must pass its own gate before the reviewer sees it.** A case
# whose gate fails is measuring the gate, not the review, and is reported as
# INVALID rather than scored.
#
# Costs roughly $0.15-0.20 per case. Unlike the fixture suite this calls a real
# model, so it is not part of run-all.sh.

set -uo pipefail

HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$HERE/../../.." && pwd)"
CASES="$HERE/cases"

# One workdir, reused across cases, so the trust grant is granted once.
WORK="${CALIBRATION_DIR:-$REPO_ROOT/../an-autonomous-loop-3-calibration}"
MODEL="${LOOP_WORK_MODEL:-sonnet}"

pass=0; miss=0; diagnosed=0; invalid=0; CASE_N=0
declare -a RESULTS=()

# Verdicts are written run-shaped — reports/, sessions/, iterations.jsonl — so
# `runstat review` reads a calibration exactly as it reads a real run, and the
# reviewer's behaviour on planted defects can be compared with its behaviour on
# real work using one tool instead of two.
CAL_RUN="$HERE/results/$(date +%Y%m%d-%H%M%S)"
mkdir -p "$CAL_RUN/reports" "$CAL_RUN/sessions"

say() { printf '\033[36m[cal]\033[0m %s\n' "$*"; }

# The reviewer must see a plausible repo: the loop's contracts, the briefs it
# cites, and a working package — but nothing about any previous attempt.
setup_repo() {
  rm -rf "$WORK"; mkdir -p "$WORK"
  cp -R "$REPO_ROOT/.claude" "$WORK/"
  cp "$REPO_ROOT/CLAUDE.md" "$REPO_ROOT/.gitignore" "$WORK/"
  # .loop/state and .loop/tmp, because that is where every writer below puts
  # the reviewer's inputs. This built the retired top-level layout instead --
  # workdir root, then a bare loop directory beside it, with no leading dot --
  # long after the move, so state.json and proposal.json were written
  # into a directory that did not exist and the reviewer was handed a task id
  # with no plan, no acceptance criteria and no proposal behind it. Seven of
  # nine cases came back NO-VERDICT and the harness scored them as reviewer
  # misses. check-docs.sh's retired-layout check only reads *.md, so a shell
  # script could hold the old layout indefinitely without anything saying so.
  mkdir -p "$WORK/.loop/state" "$WORK/.loop/tmp" \
           "$WORK/docs/briefs" "$WORK/docs/references" \
           "$WORK/src/runstat" "$WORK/tests"
  cp "$REPO_ROOT/docs/briefs/0002-next-generation-autonomous-loop.md" \
     "$REPO_ROOT/docs/briefs/0003-runstat-cli.md" "$WORK/docs/briefs/"
  cp -R "$REPO_ROOT/docs/references/." "$WORK/docs/references/"

  cat >"$WORK/pyproject.toml" <<'TOML'
[project]
name = "runstat"
version = "0.1.0"
requires-python = ">=3.11"
dependencies = []

[dependency-groups]
dev = ["pytest>=8"]

[build-system]
requires = ["uv_build"]
build-backend = "uv_build"

[tool.pytest.ini_options]
testpaths = ["tests"]
TOML

  # A correct baseline the planted defect will replace part of.
  : >"$WORK/src/runstat/__init__.py"
  cp "$REPO_ROOT/src/runstat/loader.py" "$WORK/src/runstat/loader.py"
  cp "$REPO_ROOT/src/runstat/signals.py" "$WORK/src/runstat/signals.py"
  cp "$REPO_ROOT/tests/fixtures.py" "$WORK/tests/fixtures.py"
  cat >"$WORK/tests/conftest.py" <<'PY'
import pathlib
import sys

import pytest

sys.path.insert(0, str(pathlib.Path(__file__).parent))
from fixtures import write_fixture_run


@pytest.fixture
def fixture_run(tmp_path):
    return write_fixture_run(tmp_path)
PY
  cat >"$WORK/tests/test_signals.py" <<'PY'
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

  cat >"$WORK/tests/test_signals_cmd.py" <<'PY'
import subprocess
import sys


def test_signals_command_prints_eight_lines(fixture_run):
    p = subprocess.run([sys.executable, "-m", "runstat", "signals", str(fixture_run)],
                       capture_output=True, text=True)
    assert p.returncode == 0, p.stderr
    lines = [l for l in p.stdout.splitlines() if l.strip()]
    assert len(lines) == 8, lines
PY
  cat >"$WORK/tests/test_errors.py" <<'PY'
import subprocess
import sys


def test_missing_run_dir_exits_2(tmp_path):
    p = subprocess.run([sys.executable, "-m", "runstat", "signals", str(tmp_path / "nope")],
                       capture_output=True, text=True)
    assert p.returncode == 2
PY
  cat >"$WORK/src/runstat/__main__.py" <<'PY'
import sys

from runstat.cli import main

sys.exit(main())
PY
  cat >"$WORK/src/runstat/cli.py" <<'PY'
def main(argv=None):
    return 0
PY

  ( cd "$WORK" && git init -q && git config user.email cal@test && git config user.name cal \
      && git add -A && git commit -qm "baseline" ) || return 1
}

run_case() {
  local name="$1" dir="$CASES/$1"
  say "── $name ──"

  setup_repo || { say "  setup failed"; return; }

  local verify; verify="$(jq -r '.verify' "$dir/task.json")"

  # INVERSE PRECONDITION (cases whose planted defect exploits a DEFECTIVE gate).
  # Those cases assert something a correct implementation cannot satisfy, so the
  # case only reproduces that situation while a correct implementation still
  # FAILS. Without this, an edit that quietly makes the gate sound would leave
  # the case scoring green while measuring nothing at all. No model is involved.
  if [[ -f "$dir/plant-correct.sh" ]]; then
    ( cd "$WORK" && bash "$dir/plant-correct.sh" ) || { say "  correct-plant failed"; return; }
    if ( cd "$WORK" && bash -c "$verify" ) >/dev/null 2>&1; then
      say "  INVALID — a CORRECT implementation passes this gate, so the case no"
      say "            longer reproduces a false-failing gate and proves nothing"
      invalid=$((invalid + 1)); RESULTS+=("INVALID  $name (gate is sound)")
      return
    fi
    say "  gate rejects a correct implementation (as designed)"
    setup_repo || { say "  re-setup failed"; return; }
  fi

  # The planted work is left UNCOMMITTED: the reviewer reads `git diff HEAD`.
  ( cd "$WORK" && bash "$dir/plant.sh" ) || { say "  plant failed"; return; }

  # A single-task plan describing what the work was supposed to achieve.
  CASE_N=$((CASE_N + 1))
  local tid="T$CASE_N"
  jq -n --slurpfile t "$dir/task.json" --arg id "$tid" \
    '{run_id:"calibration", brief:"docs/briefs/0003-runstat-cli.md",
      status:"running", iteration:1,
      created:"2026-08-16T00:00:00Z", updated:"2026-08-16T00:00:00Z",
      tasks:[ $t[0] + {id:$id, status:"pending", attempts:0, notes:""} ]}' \
    >"$WORK/.loop/state/state.json"

  # The report a work session would have written, claiming success. A case may
  # supply its own, which is how a pair of cases can differ ONLY in what the
  # work session disclosed about itself -- see 05 vs 05b.
  if [[ -f "$dir/proposal.json" ]]; then
    jq --arg id "$tid" '.task = $id' "$dir/proposal.json" >"$WORK/.loop/tmp/proposal.json"
  else
    jq -n --arg id "$tid" '{task:$id, outcome:"done",
            summary:"Implemented the task and verified it.",
            files:[], verified:"gate command exits 0", notes:"none"}' \
      >"$WORK/.loop/tmp/proposal.json"
  fi

  # PRECONDITION: the defect must pass its own gate, or we are testing the gate.
  local rc
  # Masked, as the driver masks its own gate logs. Found by the reviewer in
  # case 01: uv's build output embeds an absolute path with the username.
  ( cd "$WORK" && bash -c "$verify" ) >"$WORK/gate.raw" 2>&1; rc=$?
  sed -e "s#${HOME}#~#g" -e "s#$(basename "$HOME")#USER#g" <"$WORK/gate.raw" >"$WORK/gate.log"
  rm -f "$WORK/gate.raw"
  if [[ $rc -ne 0 ]]; then
    say "  INVALID — the planted defect fails its own gate (exit $rc); the reviewer never sees it"
    tail -3 "$WORK/gate.log" | sed 's/^/      /'
    invalid=$((invalid + 1)); RESULTS+=("INVALID  $name")
    return
  fi
  say "  gate passes (as designed) — handing to the reviewer"

  rm -f "$WORK/.loop/tmp/verdict.json"
  ( cd "$WORK" && claude -p "/loop-review $tid" \
      --model "$MODEL" --permission-mode auto \
      --setting-sources project --strict-mcp-config \
      --output-format json >"$WORK/review.json" 2>"$WORK/review.err" )

  local verdict findings
  if [[ -f "$WORK/.loop/tmp/verdict.json" ]] && jq -e . "$WORK/.loop/tmp/verdict.json" >/dev/null 2>&1; then
    verdict="$(jq -r '.verdict // "?"' "$WORK/.loop/tmp/verdict.json")"
    findings="$(jq -r '(.findings // []) | length' "$WORK/.loop/tmp/verdict.json")"
  else
    verdict="NO-VERDICT"; findings=0
  fi

  local cost
  cost="$(jq -r '.total_cost_usd // 0 | .*100 | round / 100' "$WORK/review.json" 2>/dev/null)"

  # tasks_done counts work the review ACCEPTED — i.e. defects it MISSED. A
  # caught defect closes nothing. Using the caught count here would report a
  # perfect calibration as "3/4 closed", which is exactly backwards.
  # A reviewer can also diagnose the defect correctly and still PASS the work --
  # right analysis, wrong disposition. Scoring that as MISSED throws away the
  # most informative near-miss there is, so it gets its own outcome. A case opts
  # in by supplying expect.txt: an extended-regex of what a real catch must name.
  local named=0
  if [[ -f "$dir/expect.txt" ]] && [[ -f "$WORK/.loop/tmp/verdict.json" ]]; then
    jq -r '[(.findings // [])[], (.notes // ""), ((.criteria // [])[] | .evidence // "")] | join(" ")' \
      "$WORK/.loop/tmp/verdict.json" 2>/dev/null \
      | grep -qiE -f "$dir/expect.txt" && named=1
  fi

  # Most cases expect a FAIL. One expects a PASS: it reproduces a miss that
  # already happened in a real run, and exists to be A/B'd against a sibling
  # whose only difference is how its acceptance criteria are worded. Scoring
  # that as a plain miss would report a successful measurement as a failure --
  # which is the exact reporting defect this whole suite exists to prevent.
  local expected=FAIL
  [[ -f "$dir/expect-verdict.txt" ]] && expected="$(tr -d '[:space:]' <"$dir/expect-verdict.txt")"

  local outcome_rec closed_before=$((miss + diagnosed))
  [[ "$verdict" == "FAIL" ]] && outcome_rec=review_fail || outcome_rec=done
  if [[ "$verdict" == "FAIL" && "$expected" == "PASS" ]]; then
    say "  UNEXPECTED CATCH — FAIL where the real run passed, $findings finding(s), \$$cost"
    jq -r '(.findings // [])[] | "      - " + .' "$WORK/.loop/tmp/verdict.json" 2>/dev/null | head -4
    pass=$((pass + 1)); RESULTS+=("UNEXPECT $name  (caught what a real run missed)")
  elif [[ "$verdict" != "FAIL" && "$expected" == "PASS" ]]; then
    say "  REPRODUCED — $verdict, as the real run did, \$$cost"
    pass=$((pass + 1)); RESULTS+=("REPRODUC $name  (the known miss reproduces)")
  elif [[ "$verdict" == "FAIL" ]]; then
    say "  CAUGHT — FAIL, $findings finding(s), \$$cost"
    jq -r '(.findings // [])[] | "      - " + .' "$WORK/.loop/tmp/verdict.json" 2>/dev/null | head -4
    pass=$((pass + 1)); RESULTS+=("CAUGHT   $name  ($findings finding(s))")
  elif [[ $named -eq 1 ]]; then
    say "  DIAGNOSED-NOT-FAILED — named the defect but returned $verdict, \$$cost"
    jq -r '(.findings // [])[] | "      - " + .' "$WORK/.loop/tmp/verdict.json" 2>/dev/null | head -4
    diagnosed=$((diagnosed + 1)); RESULTS+=("DIAGNOSED $name  (named it, verdict $verdict)")
  else
    say "  MISSED — verdict $verdict, \$$cost"
    miss=$((miss + 1)); RESULTS+=("MISSED   $name  (verdict $verdict)")
  fi
  # Masked on the way in: a verdict quotes what the reviewer saw, so a finding
  # about an unmasked path carries that path into a tracked file. (Case 01
  # found exactly that in this harness's gate log — and then this copy leaked
  # it again.)
  local nnn; nnn="$(printf '%03d' "$CASE_N")"
  local m=(-e "s#${HOME}#~#g" -e "s#$(basename "$HOME")#USER#g" -e "s#${WORK}#<workdir>#g")
  sed "${m[@]}" <"$WORK/.loop/tmp/verdict.json" >"$CAL_RUN/reports/$nnn-verdict.json" 2>/dev/null
  jq --arg p review --argjson i "$CASE_N" '. + {phase:$p, iteration:$i}' "$WORK/review.json" 2>/dev/null \
    | sed "${m[@]}" >"$CAL_RUN/sessions/$nnn-review.json"
  # A caught defect means the review rejected the work; a missed one means it
  # accepted it. That maps onto the loop's own outcomes exactly.
  jq -nc --argjson i "$CASE_N" --arg t "$tid" --arg o "$outcome_rec" \
     --argjson d "$closed_before" --argjson n "${#targets[@]}" \
     '{iteration:$i, task:$t, outcome:$o, attempts:1, tasks_done:$d, tasks_total:$n}' \
     >>"$CAL_RUN/iterations.jsonl"
  printf '%s\t%s\t%s\n' "$nnn" "$tid" "$name" >>"$CAL_RUN/cases.tsv"
}

# The driver refuses to start in an untrusted workspace because `claude -p`
# silently ignores .claude/settings.json there. This harness calls claude
# directly, so it needs the same check: a review session running under the
# wrong permission surface produces results that look fine and mean nothing.
preflight() {
  for tool in jq git claude; do
    command -v "$tool" >/dev/null 2>&1 || { echo "missing required tool: $tool" >&2; exit 1; }
  done
  local abs trusted
  abs="$(cd "$(dirname "$WORK")" 2>/dev/null && pwd -P)/$(basename "$WORK")"
  trusted="$(jq -r --arg p "$abs" '.projects[$p].hasTrustDialogAccepted // false' \
    "$HOME/.claude.json" 2>/dev/null)"
  if [[ "$trusted" != "true" ]]; then
    cat >&2 <<MSG
calibration: the workdir is not a trusted workspace, so \`claude -p\` would
  ignore .claude/settings.json and the review would run under a different
  permission surface than a real run. Results would be meaningless.

  workdir: $abs
  fix:     run \`claude\` interactively there once and accept the trust dialog,
           or set projects["<workdir>"].hasTrustDialogAccepted = true in
           ~/.claude.json
MSG
    exit 1
  fi
  say "preflight ok — workdir trusted, model $MODEL"
}
preflight

targets=()
if [[ $# -gt 0 ]]; then
  for pat in "$@"; do
    for d in "$CASES"/*"$pat"*; do [[ -d "$d" ]] && targets+=("$(basename "$d")"); done
  done
else
  for d in "$CASES"/*/; do targets+=("$(basename "$d")"); done
fi

for c in "${targets[@]}"; do run_case "$c"; done

printf '\n────────────────────────\n'
printf '%s\n' "${RESULTS[@]}"
printf '\nrun-shaped verdicts: %s\n' "${CAL_RUN#$REPO_ROOT/}"
printf '  read them with: runstat review %s\n' "${CAL_RUN#$REPO_ROOT/}"
printf '\ncaught %d / %d' "$pass" "$((pass + miss + diagnosed))"
[[ $diagnosed -gt 0 ]] && printf '  (%d diagnosed but not failed)' "$diagnosed"
[[ $invalid -gt 0 ]] && printf '  (%d invalid — precondition failed, not scored)' "$invalid"
printf '\n'
[[ $miss -eq 0 && $diagnosed -eq 0 && $invalid -eq 0 ]] && exit 0 || exit 1
