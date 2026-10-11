#!/bin/sh
# Gate T13 — close (brief: Real-data check, Real-data check L11): with the
# seed in a temporary repository, vloop brief check on copies of four consumed
# briefs set to ready warns naming the tests their recorded defects say they
# omitted; vloop metrics on this repository prints, for B13 and B1, the
# instructions figures git gives at the commit their delivered lines are
# measured to, and exits 0 with an instructions line for every consumed brief.
G=$(cd "$(dirname "$0")" && pwd -P) || { echo "GATE FAIL: no gate folder" >&2; exit 1; }
. "$G/lib.sh"
TAIL='(.vloop/repo-guidelines.md)'
P='must-change list: the contract mentions'

newrepo real
cp "$G/seed.md" "$R/.vloop/repo-guidelines.md" || fail seed
mkdir -p "$R/docs/briefs"
for b in B20261003-2049-gate-model B20261004-1434-interventions-options B20261008-2144-b12-field-reliability B20261011-2100-b13-docs-onboarding; do
  awk '!d && $0 == "status: consumed" { print "status: ready"; d = 1; next } { print }' "$top/docs/briefs/$b.loop-brief.md" >"$R/docs/briefs/$b.loop-brief.md" || fail "copy $b"
  inr "$B" brief check "docs/briefs/$b.loop-brief.md" </dev/null 2>&1 | sed 's/^ *//' >"$t/$b"
done
has() { grep -Fxq -- "! $P $2 $TAIL" "$t/$1" || fail "$1: no warning '$P $2 $TAIL' in:
$(cat "$t/$1")"; }
has B20261003-2049-gate-model '"schemas/" but Constraints do not name TestWorkedExampleB2Commands, TestSchemasAreEmbedded, TestWorkedExampleB6RealData'
has B20261004-1434-interventions-options '"schemas/" but Constraints do not name TestSchemasAreEmbedded, TestWorkedExampleB6RealData'
has B20261008-2144-b12-field-reliability '"config key" but Constraints do not name TestConfigListDefaults, TestConfigListJSONOrderAndNull, TestDefaultsEveryRow'
has B20261011-2100-b13-docs-onboarding '"closing lines" but Constraints do not name TestInitPointsAtBothWaysToLoadThePlugin'

th() { awk -v n="$1" 'BEGIN { s = ""; while (n >= 1000) { s = sprintf(",%03d", n % 1000) s; n = int(n / 1000) }; print n s }'; }
endof() {
  git -C "$top" log --all --format='%ct %H %s' | awk -v a="[vloop] run $1/" -v b="[loop] run $1/" '
    { s = $0; sub(/^[^ ]+ [^ ]+ /, "", s) }
    (index(s, a) == 1 || index(s, b) == 1) && $1 >= best { best = $1; h = $2 }
    END { print h }'
}
words() { if git -C "$top" cat-file -e "$1:$2" 2>/dev/null; then git -C "$top" show "$1:$2" | wc -w | tr -d ' '; else echo 0; fi; }
after() { awk 'p { print; exit } /^ interventions / { p = 1 }' "$1"; }
# plan figure: the brief's formula, the plan skill + CLAUDE.md at the end commit.
for bs in B20261011-2100-b13-docs-onboarding:plugin/skills/plan/SKILL.md B20260929-1804-vloop-skeleton-config-briefs:.claude/skills/loop-plan/SKILL.md; do
  b=${bs%%:*} s=${bs#*:}
  e=$(endof "$b"); [ -n "$e" ] || fail "no run commit for $b"
  git -C "$top" cat-file -e "$e:$s" || fail "fixture: no $s at $e"
  want=$(th $(($(words "$e" "$s") + $(words "$e" CLAUDE.md) + $(words "$e" .vloop/repo-guidelines.md))))
  (cd "$top" && "$B" metrics "$b") </dev/null >"$t/$b.txt" 2>&1 || fail "vloop metrics $b: $(cat "$t/$b.txt")"
  case "$(after "$t/$b.txt")" in " instructions  plan $want · "*) ;; *) fail "$b: the instructions line does not give plan $want: '$(after "$t/$b.txt")'" ;; esac
done
case "$(after "$t/B20260929-1804-vloop-skeleton-config-briefs.txt")" in *' · gate review n/a · '*) ;; *) fail "B1's gate review is not n/a" ;; esac
for f in "$top"/docs/briefs/*.loop-brief.md; do
  awk 'NR == 1 && $0 == "---" { p = 1; next } p && $0 == "---" { exit } p && $0 == "status: consumed" { f = 1 } END { exit !f }' "$f" || continue
  b=$(basename "$f" .loop-brief.md)
  (cd "$top" && "$B" metrics "$b") </dev/null >"$t/all.txt" 2>&1 || fail "vloop metrics $b exits non-zero: $(tail -n 5 "$t/all.txt")"
  grep -q '^ instructions  plan ' "$t/all.txt" || fail "vloop metrics $b has no instructions line"
done
echo "gate T13: ok"
