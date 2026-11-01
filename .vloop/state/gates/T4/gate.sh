#!/bin/sh
# Gate T4 — A1, the audit, the guide and the domain pages. The fence audit
# gains a part for the gate-review fence — a heading naming gate review and
# what lies under it, up to the next heading of the same or a higher level
# that does not name it — that classifies every deny rule of
# fence/gate-review.json and every pattern of the adapter's deny list (the
# review list, shell(rm:*) included) by who enforces it in a Copilot
# gate-review session: driver, hook or neither. The configuration guide's
# agent.gate-review row offers copilot and says experimental, the plan row
# still offers claude alone, its prose no longer keeps agent.gate-review on
# claude and covers model.gate-review, its fence hook section names the
# gate-review phase and fence, and K1's hint names /vloop:gate-review. The
# session page says copilot runs gate-review sessions and its Copilot hook
# paragraph serves the gate-review phase with the gate-review fence; the
# glossary's hook entry covers gate review. Read-only: it reads this tree's
# documents and nothing else.
set -u
fail() { echo "GATE FAIL: $*" >&2; exit 1; }
export LC_ALL=C
JQ=$(command -v jq) || fail "jq is not on PATH"
AUDIT=docs/design-notes/fence-audit.md
GUIDE=docs/guide/configuration.md
SESSION=docs/domain/execution/session.md
MODEL=docs/domain/domain-model.md
for f in "$AUDIT" "$GUIDE" "$SESSION" "$MODEL"; do [ -f "$f" ] || fail "$f does not exist"; done
t=$(mktemp -d "${TMPDIR:-/tmp}/gate.XXXXXX") || fail mktemp
trap 'rm -rf "$t"' EXIT

# The audit's gate-review part: from each heading naming gate review, down to
# the next heading of its level or above that does not name it.
awk '
  /^#+[ \t]/ {
    match($0, /^#+/); lv = RLENGTH; g = (tolower($0) ~ /gate[- ]review/)
    if (g && (!p || lv <= top)) { p = 1; top = lv }
    else if (p && !g && lv <= top) { p = 0 }
  }
  p { print }
' "$AUDIT" >"$t/gr.md"
[ -s "$t/gr.md" ] || fail "$AUDIT has no heading naming the gate-review fence or gate-review sessions"
classified() {
  grep -F -- "$1" "$t/gr.md" | grep -qiE 'driver|hook|neither' || fail "$AUDIT's gate-review part does not classify $1 (driver, hook or neither) on its line"
}
"$JQ" -r '.permissions.deny[]' fence/gate-review.json >"$t/rules" || fail "fence/gate-review.json unreadable"
n=0
while IFS= read -r r; do classified "$r"; n=$((n + 1)); done <"$t/rules"
[ "$n" -gt 100 ] || fail "read only $n deny rules from fence/gate-review.json"
# The adapter's deny list as it stands when this plan was written: a
# gate-review session carries all of it, shell(rm:*) included.
for p in 'shell(gh:*)' 'shell(git push:*)' 'shell(git remote:*)' 'shell(git config --global:*)' 'shell(git config --system:*)' \
  'shell(git config --file:*)' 'shell(curl:*)' 'shell(wget:*)' 'shell(ssh:*)' 'shell(scp:*)' 'shell(nc:*)' 'shell(rm:*)' \
  'shell(git commit:*)' 'shell(git reset:*)' 'shell(git checkout:*)' 'shell(git switch:*)' 'shell(git stash:*)' \
  'shell(git branch:*)' 'shell(git tag:*)' '`url`'; do
  classified "$p"
done
grep -qiw neither "$t/gr.md" || fail "$AUDIT's gate-review part classifies no rule as neither — say what covers those, or say none is"

# The configuration guide.
row() { grep -E "^\|[[:space:]]*\`$1\`[[:space:]]*\|" "$GUIDE" || fail "$GUIDE has no table row for $1"; }
col() { echo "$1" | awk -F'|' -v n="$2" '{ c = $(n + 1); gsub(/^[ \t]+|[ \t]+$/, "", c); print c }'; }
r=$(row 'agent\.gate-review')
[ "$(col "$r" 2)" = '`claude`' ] || fail "the agent.gate-review row's default is '$(col "$r" 2)', want \`claude\`"
col "$r" 3 | grep -qF '`copilot`' || fail "the agent.gate-review row does not offer \`copilot\`: $r"
col "$r" 3 | grep -qF '`claude`' || fail "the agent.gate-review row no longer offers \`claude\`: $r"
[ "$(col "$r" 4)" = '`VLOOP_AGENT_GATE_REVIEW`' ] || fail "the agent.gate-review row's variable is '$(col "$r" 4)'"
echo "$r" | grep -qi experimental || fail "the agent.gate-review row does not say copilot is experimental: $r"
r=$(row 'agent\.plan')
[ "$(col "$r" 3)" = '`claude`' ] || fail "the agent.plan row offers '$(col "$r" 3)', want \`claude\` alone"
grep -v '^|' "$GUIDE" >"$t/prose"
grep -qE 'agent\.gate-review`?[^.]* stays? on `?claude' "$t/prose" &&
  fail "$GUIDE still says agent.gate-review stays on claude: $(grep -nE 'agent\.gate-review`?[^.]* stays? on' "$t/prose")"
grep -qF 'agent.gate-review' "$t/prose" || fail "$GUIDE's prose (outside the key table) does not cover agent.gate-review"
grep -qF 'model.gate-review' "$t/prose" || fail "$GUIDE's prose does not cover the model.gate-review refusal"
grep -oE -- '--phase <[^>]*>' "$GUIDE" >"$t/phases"
[ -s "$t/phases" ] || fail "$GUIDE no longer shows the --phase list of vloop fence hook and vloop fence check"
grep -vF gate-review "$t/phases" >/dev/null && fail "a --phase list in $GUIDE does not name gate-review: $(grep -vF gate-review "$t/phases" | head -1)"
grep -qiE 'gate-review fence|fence/gate-review\.json' "$GUIDE" || fail "$GUIDE does not say the hook judges gate-review sessions against the gate-review fence"
grep -qF '/vloop:gate-review' "$GUIDE" || fail "$GUIDE's K1 skill hint does not name /vloop:gate-review for a gate-review session"
grep -qF 'vloop fence hook' "$GUIDE" || fail "$GUIDE no longer describes vloop fence hook"

# The domain pages.
grep -iE 'copilot runs ' "$SESSION" >"$t/runs" || fail "$SESSION no longer says which sessions copilot runs"
grep -qiE 'gate-review|gate review' "$t/runs" || fail "$SESSION does not say copilot runs gate-review sessions: $(cat "$t/runs")"
grep -qw verify "$t/runs" || fail "$SESSION no longer says copilot runs verify sessions: $(cat "$t/runs")"
awk '/\*\*The Copilot hook\.\*\*/ { p = 1 } p && /^$/ { exit } p { print }' "$SESSION" >"$t/hook"
[ -s "$t/hook" ] || fail "$SESSION has no 'The Copilot hook.' paragraph"
grep -qiE 'gate-review fence|fence/gate-review\.json' "$t/hook" || fail "$SESSION's Copilot hook paragraph does not say the hook serves gate-review sessions with the gate-review fence"
grep -qiE 'verify fence|fence/verify\.json' "$t/hook" || fail "$SESSION's Copilot hook paragraph no longer names the verify fence"
h=$(grep -E '^\|[[:space:]]*\*\*hook\*\*' "$MODEL") || fail "$MODEL's glossary has no hook entry"
echo "$h" | grep -qiE 'gate-review|gate review' || fail "$MODEL's glossary hook entry does not cover gate-review sessions: $h"
echo "gate T4: ok"
