#!/bin/sh
# Gate T4 — A1, the audit, the guide and the domain pages. The fence audit
# gains a part for the work fence — from its first heading naming "work" on —
# that classifies every deny rule of fence/work.json and every work deny-list
# pattern (the review list without shell(rm:*)) by who enforces it in a Copilot
# work session: driver, hook or neither. The configuration guide's agent.work
# row offers copilot and says experimental, the plan, gate-review and verify
# rows still offer claude alone, and the guide says the hook judges work
# sessions against the work fence. The session page no longer limits copilot
# to review sessions and names the work fence; the glossary's hook entry covers
# work. Read-only: it reads this tree's documents and nothing else.
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

# The audit's work part: from its first heading naming work to the end.
awk 'p || (/^#/ && tolower($0) ~ /work/) { p = 1; print }' "$AUDIT" >"$t/work.md"
[ -s "$t/work.md" ] || fail "$AUDIT has no heading naming the work fence or work sessions"
classified() {
  grep -F -- "$1" "$t/work.md" | grep -qiE 'driver|hook|neither' || fail "$AUDIT's work part does not classify $1 (driver, hook or neither) on its line"
}
"$JQ" -r '.permissions.deny[]' fence/work.json >"$t/rules" || fail "fence/work.json unreadable"
n=0
while IFS= read -r r; do classified "$r"; n=$((n + 1)); done <"$t/rules"
[ "$n" -gt 100 ] || fail "read only $n deny rules from fence/work.json"
# The adapter's deny list for work sessions as it stands when this plan was written.
for p in 'shell(gh:*)' 'shell(git push:*)' 'shell(git remote:*)' 'shell(git config --global:*)' 'shell(git config --system:*)' \
  'shell(git config --file:*)' 'shell(curl:*)' 'shell(wget:*)' 'shell(ssh:*)' 'shell(scp:*)' 'shell(nc:*)' \
  'shell(git commit:*)' 'shell(git reset:*)' 'shell(git checkout:*)' 'shell(git switch:*)' 'shell(git stash:*)' \
  'shell(git branch:*)' 'shell(git tag:*)' '`url`'; do
  classified "$p"
done
grep -qiw neither "$t/work.md" || fail "$AUDIT's work part classifies no rule as neither — say what covers those, or say none is"

# The configuration guide.
row() { grep -E "^\|[[:space:]]*\`$1\`[[:space:]]*\|" "$GUIDE" || fail "$GUIDE has no table row for $1"; }
col() { echo "$1" | awk -F'|' -v n="$2" '{ c = $(n + 1); gsub(/^[ \t]+|[ \t]+$/, "", c); print c }'; }
r=$(row 'agent\.work')
[ "$(col "$r" 2)" = '`claude`' ] || fail "the agent.work row's default is '$(col "$r" 2)', want \`claude\`"
col "$r" 3 | grep -qF '`copilot`' || fail "the agent.work row does not offer \`copilot\`: $r"
col "$r" 3 | grep -qF '`claude`' || fail "the agent.work row no longer offers \`claude\`: $r"
[ "$(col "$r" 4)" = '`VLOOP_AGENT_WORK`' ] || fail "the agent.work row's variable is '$(col "$r" 4)'"
echo "$r" | grep -qi experimental || fail "the agent.work row does not say copilot is experimental: $r"
for k in plan gate-review verify; do
  r=$(row "agent\\.$k")
  [ "$(col "$r" 3)" = '`claude`' ] || fail "the agent.$k row offers '$(col "$r" 3)', want \`claude\` alone"
done
grep -v '^|' "$GUIDE" | grep -qF 'agent.work' || fail "$GUIDE's prose (outside the key table) does not cover agent.work"
grep -qiE 'work fence|fence/work\.json' "$GUIDE" || fail "$GUIDE does not say the hook judges work sessions against the work fence"
grep -qF 'vloop fence hook' "$GUIDE" || fail "$GUIDE no longer describes vloop fence hook"

# The domain pages.
grep -qi 'review sessions only' "$SESSION" && fail "$SESSION still says copilot runs review sessions only: $(grep -in 'review sessions only' "$SESSION")"
grep -qiE 'work fence|fence/work\.json' "$SESSION" || fail "$SESSION does not say the hook serves work sessions with the work fence"
h=$(grep -E '^\|[[:space:]]*\*\*hook\*\*' "$MODEL") || fail "$MODEL's glossary has no hook entry"
echo "$h" | grep -qi work || fail "$MODEL's glossary hook entry does not cover work sessions: $h"
echo "gate T4: ok"
