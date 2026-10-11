#!/bin/sh
# Gate T10 — vloop metrics' `instructions` line (brief L11): for a brief, at the
# commit its delivered lines are measured to, the words each session kind
# starts with (its skill + .vloop/repo-guidelines.md + CLAUDE.md, as wc -w
# counts them, 0 when absent; plugin/skills/<kind>/SKILL.md, else the shell
# loop's .claude/skills/loop-<kind>/SKILL.md for plan, work and review, else
# n/a) and the guidelines' rule count; the line after `interventions`, the
# optional `instructions` object in --json, frozen by brief close. And the
# real-data check in this repository.
G=$(cd "$(dirname "$0")" && pwd -P) || { echo "GATE FAIL: no gate folder" >&2; exit 1; }
. "$G/lib.sh"
th() { awk -v n="$1" 'BEGIN { s = ""; while (n >= 1000) { s = sprintf(",%03d", n % 1000) s; n = int(n / 1000) }; print n s }'; }
# endof <repo> <run id>: the latest run commit of the run id, either layout.
endof() {
  git -C "$1" log --all --format='%ct %H %s' | awk -v a="[vloop] run $2/" -v b="[loop] run $2/" '
    { s = $0; sub(/^[^ ]+ [^ ]+ /, "", s) }
    (index(s, a) == 1 || index(s, b) == 1) && $1 >= best { best = $1; h = $2 }
    END { print h }'
}
# words <repo> <commit> <path>: wc -w of the file at the commit, 0 when absent.
words() { if git -C "$1" cat-file -e "$2:$3" 2>/dev/null; then git -C "$1" show "$2:$3" | wc -w | tr -d ' '; else echo 0; fi; }
# skill <repo> <commit> <kind dir> <shell loop name or ->: the kind's skill path, or nothing.
skill() {
  if git -C "$1" cat-file -e "$2:plugin/skills/$3/SKILL.md" 2>/dev/null; then echo "plugin/skills/$3/SKILL.md"
  elif [ "$4" != - ] && git -C "$1" cat-file -e "$2:.claude/skills/loop-$4/SKILL.md" 2>/dev/null; then echo ".claude/skills/loop-$4/SKILL.md"
  fi
}
# kind <repo> <commit> <kind dir> <shell loop name or ->: the kind's figure, or n/a.
kind() {
  p=$(skill "$1" "$2" "$3" "$4")
  if [ -z "$p" ]; then echo n/a; return; fi
  th $(($(words "$1" "$2" "$p") + $(words "$1" "$2" CLAUDE.md) + $(words "$1" "$2" .vloop/repo-guidelines.md)))
}
# expected <repo> <run id> <rules>: the instructions line.
expected() {
  e=$(endof "$1" "$2"); [ -n "$e" ] || fail "no run commit for $2 in $1"
  printf ' instructions  plan %s · gate review %s · work %s · review %s words · guidelines %s rules\n' \
    "$(kind "$1" "$e" plan plan)" "$(kind "$1" "$e" gate-review -)" "$(kind "$1" "$e" work work)" "$(kind "$1" "$e" review review)" "$3"
}
after() { awk 'p { print; exit } /^ interventions / { p = 1 }' "$1"; }

# --- the worked example: a plan skill of 100 words, a work skill of 50, the seed ---
runrepo m
mkdir -p "$R/plugin/skills/plan" "$R/plugin/skills/work" || fail mkdir
wn() { i=0; while [ $i -lt "$1" ]; do printf 'word '; i=$((i + 1)); done; echo; }
wn 100 >"$R/plugin/skills/plan/SKILL.md" && wn 50 >"$R/plugin/skills/work/SKILL.md" || fail skills
cp "$G/seed.md" "$R/.vloop/repo-guidelines.md" || fail seed
commit "skills and guidelines"
vrun m.run
[ "$got" = 0 ] || fail "vloop run: exit $got: $(tail -n 20 "$t/m.run")"
e=$(endof "$R" "$RUN")
cw=$(words "$R" "$e" CLAUDE.md) gw=$(words "$R" "$e" .vloop/repo-guidelines.md)
[ "$gw" -gt 0 ] && [ "$cw" -gt 0 ] || fail "fixture: CLAUDE.md or the seed is not at the run's end commit"
want=" instructions  plan $(th $((100 + cw + gw))) · gate review n/a · work $(th $((50 + cw + gw))) · review n/a words · guidelines 6 rules"
inr "$B" metrics "$RUN" </dev/null >"$t/m.txt" 2>&1 || fail "vloop metrics $RUN: $(cat "$t/m.txt")"
[ "$(after "$t/m.txt")" = "$want" ] || fail "after the interventions line want
'$want', got
'$(after "$t/m.txt")' in:
$(cat "$t/m.txt")"
inr "$B" metrics "$RUN" --json </dev/null >"$t/m.json" 2>"$t/m.err" || fail "vloop metrics --json: $(cat "$t/m.err")"
wj=$("$JQ" -cnS --argjson p $((100 + cw + gw)) --argjson w $((50 + cw + gw)) '{plan: $p, gate_review: null, work: $w, review: null, guidelines_rules: 6}')
[ "$("$JQ" -cS .instructions "$t/m.json")" = "$wj" ] || fail "--json instructions want $wj, got $("$JQ" -cS .instructions "$t/m.json")"
"$B" schema validate metrics/v2 "$t/m.json" >"$t/m.valid" 2>&1 || fail "--json is not metrics/v2: $(cat "$t/m.valid")"
inr "$B" brief close "$BRIEF" --no-findings </dev/null >"$t/close" 2>&1 || fail "vloop brief close: $(cat "$t/close")"
snap="$R/.vloop/state/metrics/$RUN.json"
[ -f "$snap" ] || fail "brief close froze no $snap"
[ "$("$JQ" -cS .instructions "$snap")" = "$wj" ] || fail "the snapshot's instructions want $wj, got $("$JQ" -cS .instructions "$snap")"

# --- the real-data check, in this repository ---
for b in B20261011-2100-b13-docs-onboarding B20260929-1804-vloop-skeleton-config-briefs; do
  (cd "$top" && "$B" metrics "$b") </dev/null >"$t/$b.txt" 2>&1 || fail "vloop metrics $b: $(cat "$t/$b.txt")"
  w=$(expected "$top" "$b" 0)
  [ "$(after "$t/$b.txt")" = "$w" ] || fail "$b: want
'$w', got
'$(after "$t/$b.txt")'"
done
case "$(after "$t/B20260929-1804-vloop-skeleton-config-briefs.txt")" in *'· gate review n/a ·'*) ;; *) fail "B1's gate review is not n/a" ;; esac
for f in "$top"/docs/briefs/*.loop-brief.md; do
  awk 'NR == 1 && $0 == "---" { p = 1; next } p && $0 == "---" { exit } p && $0 == "status: consumed" { f = 1 } END { exit !f }' "$f" || continue
  b=$(basename "$f" .loop-brief.md)
  (cd "$top" && "$B" metrics "$b") </dev/null >"$t/all.txt" 2>&1 || fail "vloop metrics $b exits non-zero: $(tail -n 5 "$t/all.txt")"
  grep -q '^ instructions  plan ' "$t/all.txt" || fail "vloop metrics $b has no instructions line"
done
echo "gate T10: ok"
