#!/usr/bin/env bash
#
# The loop's gate. Run this before and after any change to .loop/run.sh.
#
#   .loop/tests/run-all.sh              every scenario
#   .loop/tests/run-all.sh 03 07        only those matching
#
# Every scenario builds a throwaway repo with a scripted `claude` on PATH, so
# the suite is free, offline and deterministic. Nothing here reaches a model.

set -uo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")" || exit 1

for t in jq git; do
  command -v "$t" >/dev/null 2>&1 || { echo "missing required tool: $t" >&2; exit 1; }
done

pass=0
fail=0
failed=()

for s in scenarios/*.sh; do
  if [[ $# -gt 0 ]]; then
    match=0
    for pat in "$@"; do [[ "$s" == *"$pat"* ]] && match=1; done
    [[ $match -eq 1 ]] || continue
  fi
  printf '\n\033[1m%s\033[0m\n' "$(basename "${s%.sh}")"
  if bash "$s"; then pass=$((pass + 1)); else fail=$((fail + 1)); failed+=("$(basename "$s")"); fi
done

# Free and offline like the scenarios, so it belongs in the same gate.
# Briefs marked "ready to plan" are checked structurally; anything else is
# skipped. Free and offline, so it belongs in the same gate.
if compgen -G "../../docs/briefs/*.md" >/dev/null 2>&1; then
  printf '\n\033[1mcheck-brief\033[0m\n'
  brief_out="$(cd ../.. && .loop/check-brief.sh .loop/brief-template.md docs/briefs/*.md 2>&1)"
  # Nothing retires a brief (item 8 of B20260924-1947 refuses to PLAN from one,
  # it does not stamp it), so this repo's own already-run briefs -- including
  # this run's own -- report "already run" forever and check-brief.sh must
  # keep exiting 1 for them: that is the whole point of the check. Weakening
  # or exempting it here would be exactly what item 8 exists to prevent doing
  # inside check-brief.sh itself. Any OTHER problem it reports still fails
  # this gate.
  real_problems="$(grep '✗' <<<"$brief_out" | grep -vc 'already run')"
  if [[ "$real_problems" -eq 0 ]]; then pass=$((pass + 1))
  else printf '%s\n' "$brief_out"; fail=$((fail + 1)); failed+=("check-brief.sh"); fi
fi

printf '\n\033[1mcheck-docs\033[0m\n'
if bash ./check-docs.sh; then pass=$((pass + 1))
else fail=$((fail + 1)); failed+=("check-docs.sh"); fi

# Free and offline like the rest. Skips itself in an installed copy and in a
# tree with no .loop/todo/, so it is safe to run unconditionally here.
printf '\n\033[1mcheck-todo\033[0m\n'
if bash ./check-todo.sh; then pass=$((pass + 1))
else fail=$((fail + 1)); failed+=("check-todo.sh"); fi

printf '\n────────────────────────\n'
if [[ $fail -eq 0 ]]; then
  printf '\033[32m%d passed\033[0m\n' "$pass"
  # Reviewer calibration is deliberately not run here: it calls a real model
  # and costs money, where everything above is free and offline.
  printf '\nreviewer calibration is separate and NOT run here — it calls a real\n'
  printf 'model (~$1.50 for all nine): .loop/tests/reviewer-calibration/run-calibration.sh\n'
  exit 0
fi
printf '\033[32m%d passed\033[0m, \033[31m%d failed\033[0m\n' "$pass" "$fail"
printf '  %s\n' "${failed[@]}"
exit 1
