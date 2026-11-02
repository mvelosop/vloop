#!/bin/sh
# Gate T4 — A1, the audit, the guide and the domain pages. The fence audit
# gains a part for a Copilot plan session — a heading naming the plan session
# or the plan fence and what lies under it, up to the next heading of the
# same or a higher level that does not name it — that classifies every deny
# rule of fence/plan.json, its allowed writes and every pattern of the
# adapter's deny list (the review list, shell(rm:*) included) by who enforces
# it in a Copilot plan session: driver, hook or neither. The configuration
# guide's agent.plan row offers copilot and says experimental, its prose no
# longer keeps agent.plan on claude and covers model.plan, its fence hook
# section names the plan phase and fence, and K1's hint names /vloop:plan.
# The session page says copilot runs plan sessions and its Copilot hook
# paragraph serves the plan phase with the plan fence; the glossary's hook
# entry covers the plan phase. Read-only: it reads this tree's documents and
# nothing else.
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
# plan as a word: not "planner", "plans", "planning" or "explanation".
W='(^|[^a-z-])plan([^a-z-]|$)'

# The audit's plan part: from each heading naming plan, down to the next
# heading of its level or above that does not name it.
awk -v w="$W" '
  /^#+[ \t]/ {
    match($0, /^#+/); lv = RLENGTH; g = (tolower($0) ~ w)
    if (g && (!p || lv <= top)) { p = 1; top = lv }
    else if (p && !g && lv <= top) { p = 0 }
  }
  p { print }
' "$AUDIT" >"$t/pl.md"
[ -s "$t/pl.md" ] || fail "$AUDIT has no heading naming the plan fence or plan sessions"
classified() {
  grep -F -- "$1" "$t/pl.md" | grep -qiE 'driver|hook|neither' || fail "$AUDIT's plan part does not classify $1 (driver, hook or neither) on its line"
}
"$JQ" -r '.permissions.deny[]' fence/plan.json >"$t/rules" || fail "fence/plan.json unreadable"
n=0
while IFS= read -r r; do classified "$r"; n=$((n + 1)); done <"$t/rules"
[ "$n" -ge 90 ] || fail "read only $n deny rules from fence/plan.json, which holds 90 (this brief does not change it)"
# The plan fence's allowed writes.
for w in .vloop/state/state.json .vloop/state/gates/ .vloop/tmp/; do
  classified "$w"
done
# The adapter's deny list as it stands when this plan was written: a plan
# session carries all of it, shell(rm:*) included.
for p in 'shell(gh:*)' 'shell(git push:*)' 'shell(git remote:*)' 'shell(git config --global:*)' 'shell(git config --system:*)' \
  'shell(git config --file:*)' 'shell(curl:*)' 'shell(wget:*)' 'shell(ssh:*)' 'shell(scp:*)' 'shell(nc:*)' 'shell(rm:*)' \
  'shell(git commit:*)' 'shell(git reset:*)' 'shell(git checkout:*)' 'shell(git switch:*)' 'shell(git stash:*)' \
  'shell(git branch:*)' 'shell(git tag:*)' '`url`'; do
  classified "$p"
done
grep -qiw neither "$t/pl.md" || fail "$AUDIT's plan part classifies no rule as neither — say what covers those, or say none is"

# The configuration guide.
row() { grep -E "^\|[[:space:]]*\`$1\`[[:space:]]*\|" "$GUIDE" || fail "$GUIDE has no table row for $1"; }
col() { echo "$1" | awk -F'|' -v n="$2" '{ c = $(n + 1); gsub(/^[ \t]+|[ \t]+$/, "", c); print c }'; }
r=$(row 'agent\.plan')
[ "$(col "$r" 2)" = '`claude`' ] || fail "the agent.plan row's default is '$(col "$r" 2)', want \`claude\`"
col "$r" 3 | grep -qF '`copilot`' || fail "the agent.plan row does not offer \`copilot\`: $r"
col "$r" 3 | grep -qF '`claude`' || fail "the agent.plan row no longer offers \`claude\`: $r"
[ "$(col "$r" 4)" = '`VLOOP_AGENT_PLAN`' ] || fail "the agent.plan row's variable is '$(col "$r" 4)'"
echo "$r" | grep -qi experimental || fail "the agent.plan row does not say copilot is experimental: $r"
grep -v '^|' "$GUIDE" >"$t/prose"
grep -qE 'agent\.plan`?[^.]* stays? on `?claude' "$t/prose" &&
  fail "$GUIDE still says agent.plan stays on claude: $(grep -nE 'agent\.plan`?[^.]* stays? on' "$t/prose")"
grep -qF 'agent.plan' "$t/prose" || fail "$GUIDE's prose (outside the key table) does not cover agent.plan"
grep -qF 'agent.plan is copilot: experimental' "$t/prose" || fail "$GUIDE's prose does not give the agent.plan warning"
grep -qF 'model.plan is the default "opus"' "$t/prose" || fail "$GUIDE's prose does not give the model.plan refusal"
grep -oE -- '--phase <[^>]*>' "$GUIDE" >"$t/phases"
[ -s "$t/phases" ] || fail "$GUIDE no longer shows the --phase list of vloop fence hook and vloop fence check"
grep -vE '[<|]plan[|>]' "$t/phases" >/dev/null && fail "a --phase list in $GUIDE does not name plan: $(grep -vE '[<|]plan[|>]' "$t/phases" | head -1)"
grep -qiE 'plan fence|fence/plan\.json' "$GUIDE" || fail "$GUIDE does not say the hook judges plan sessions against the plan fence"
grep -qF '/vloop:plan' "$GUIDE" || fail "$GUIDE's K1 skill hint does not name /vloop:plan for a plan session"
grep -qF 'vloop fence hook' "$GUIDE" || fail "$GUIDE no longer describes vloop fence hook"

# The domain pages.
grep -iE 'copilot runs ' "$SESSION" >"$t/runs" || fail "$SESSION no longer says which sessions copilot runs"
grep -qiE "$W|every session kind|every kind|all session kinds" "$t/runs" || fail "$SESSION does not say copilot runs plan sessions: $(cat "$t/runs")"
awk '/\*\*The Copilot hook\.\*\*/ { p = 1 } p && /^$/ { exit } p { print }' "$SESSION" >"$t/hook"
[ -s "$t/hook" ] || fail "$SESSION has no 'The Copilot hook.' paragraph"
grep -qiE 'plan fence|fence/plan\.json' "$t/hook" || fail "$SESSION's Copilot hook paragraph does not say the hook serves plan sessions with the plan fence"
grep -qiE 'gate-review fence|fence/gate-review\.json' "$t/hook" || fail "$SESSION's Copilot hook paragraph no longer names the gate-review fence"
h=$(grep -E '^\|[[:space:]]*\*\*hook\*\*' "$MODEL") || fail "$MODEL's glossary has no hook entry"
echo "$h" | grep -qiE 'plan fence|plan session' || fail "$MODEL's glossary hook entry does not cover plan sessions: $h"
echo "gate T4: ok"
