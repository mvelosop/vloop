#!/bin/sh
# Gate T4 — A1, the audit, the guide and the domain pages. The fence audit
# gains a part for the verify fence — from its first heading naming "verify"
# on — that classifies every deny rule of fence/verify.json and every pattern
# of the adapter's deny list (the review list, shell(rm:*) included) by who
# enforces it in a Copilot verify session: driver, hook or neither. The
# configuration guide's agent.verify row offers copilot and says experimental,
# the plan and gate-review rows still offer claude alone, its prose no longer
# keeps agent.verify on claude and covers model.verify, its fence hook section
# names the verify phase and fence, and K1's hint names /vloop:verify. The
# session page says copilot runs verify sessions, names verify in its list of
# fence files and in its Copilot hook paragraph; the glossary's hook entry
# covers verify. Read-only: it reads this tree's documents and nothing else.
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

# The audit's verify part: from its first heading naming verify to the end.
awk 'p || (/^#/ && tolower($0) ~ /verify/) { p = 1; print }' "$AUDIT" >"$t/verify.md"
[ -s "$t/verify.md" ] || fail "$AUDIT has no heading naming the verify fence or verify sessions"
classified() {
  grep -F -- "$1" "$t/verify.md" | grep -qiE 'driver|hook|neither' || fail "$AUDIT's verify part does not classify $1 (driver, hook or neither) on its line"
}
"$JQ" -r '.permissions.deny[]' fence/verify.json >"$t/rules" || fail "fence/verify.json unreadable"
n=0
while IFS= read -r r; do classified "$r"; n=$((n + 1)); done <"$t/rules"
[ "$n" -gt 100 ] || fail "read only $n deny rules from fence/verify.json"
# The adapter's deny list as it stands when this plan was written: a verify
# session carries all of it, shell(rm:*) included.
for p in 'shell(gh:*)' 'shell(git push:*)' 'shell(git remote:*)' 'shell(git config --global:*)' 'shell(git config --system:*)' \
  'shell(git config --file:*)' 'shell(curl:*)' 'shell(wget:*)' 'shell(ssh:*)' 'shell(scp:*)' 'shell(nc:*)' 'shell(rm:*)' \
  'shell(git commit:*)' 'shell(git reset:*)' 'shell(git checkout:*)' 'shell(git switch:*)' 'shell(git stash:*)' \
  'shell(git branch:*)' 'shell(git tag:*)' '`url`'; do
  classified "$p"
done
grep -qiw neither "$t/verify.md" || fail "$AUDIT's verify part classifies no rule as neither — say what covers those, or say none is"

# The configuration guide.
row() { grep -E "^\|[[:space:]]*\`$1\`[[:space:]]*\|" "$GUIDE" || fail "$GUIDE has no table row for $1"; }
col() { echo "$1" | awk -F'|' -v n="$2" '{ c = $(n + 1); gsub(/^[ \t]+|[ \t]+$/, "", c); print c }'; }
r=$(row 'agent\.verify')
[ "$(col "$r" 2)" = '`claude`' ] || fail "the agent.verify row's default is '$(col "$r" 2)', want \`claude\`"
col "$r" 3 | grep -qF '`copilot`' || fail "the agent.verify row does not offer \`copilot\`: $r"
col "$r" 3 | grep -qF '`claude`' || fail "the agent.verify row no longer offers \`claude\`: $r"
[ "$(col "$r" 4)" = '`VLOOP_AGENT_VERIFY`' ] || fail "the agent.verify row's variable is '$(col "$r" 4)'"
echo "$r" | grep -qi experimental || fail "the agent.verify row does not say copilot is experimental: $r"
for k in plan gate-review; do
  r=$(row "agent\\.$k")
  [ "$(col "$r" 3)" = '`claude`' ] || fail "the agent.$k row offers '$(col "$r" 3)', want \`claude\` alone"
done
grep -v '^|' "$GUIDE" >"$t/prose"
grep -qE 'agent\.verify`?[^.]* stays? on `?claude' "$t/prose" && fail "$GUIDE still says agent.verify stays on claude: $(grep -nE 'agent\.verify`?[^.]* stays? on' "$t/prose")"
grep -qF 'agent.verify' "$t/prose" || fail "$GUIDE's prose (outside the key table) does not cover agent.verify"
grep -qF 'model.verify' "$t/prose" || fail "$GUIDE's prose does not cover the model.verify refusal"
grep -oE -- '--phase <[^>]*>' "$GUIDE" >"$t/phases"
[ -s "$t/phases" ] || fail "$GUIDE no longer shows the --phase list of vloop fence hook and vloop fence check"
grep -v verify "$t/phases" >/dev/null && fail "a --phase list in $GUIDE does not name verify: $(grep -v verify "$t/phases" | head -1)"
grep -qiE 'verify fence|fence/verify\.json' "$GUIDE" || fail "$GUIDE does not say the hook judges verify sessions against the verify fence"
grep -qF '/vloop:verify' "$GUIDE" || fail "$GUIDE's K1 skill hint does not name /vloop:verify for a verify session"
grep -qF 'vloop fence hook' "$GUIDE" || fail "$GUIDE no longer describes vloop fence hook"

# The domain pages.
grep -iE 'copilot runs ' "$SESSION" >"$t/runs" || fail "$SESSION no longer says which sessions copilot runs"
grep -qw verify "$t/runs" || fail "$SESSION does not say copilot runs verify sessions: $(cat "$t/runs")"
grep -i 'one file per phase' "$SESSION" >"$t/list" || fail "$SESSION's list of the fence files (one file per phase) is gone"
grep -qw verify "$t/list" || fail "$SESSION's list of the fence files does not name the verify fence: $(cat "$t/list")"
awk '/\*\*The Copilot hook\.\*\*/ { p = 1 } p && /^$/ { exit } p { print }' "$SESSION" >"$t/hook"
[ -s "$t/hook" ] || fail "$SESSION has no 'The Copilot hook.' paragraph"
grep -qiE 'verify fence|fence/verify\.json' "$t/hook" || fail "$SESSION's Copilot hook paragraph does not say the hook serves verify sessions with the verify fence"
h=$(grep -E '^\|[[:space:]]*\*\*hook\*\*' "$MODEL") || fail "$MODEL's glossary has no hook entry"
echo "$h" | grep -qi verify || fail "$MODEL's glossary hook entry does not cover verify sessions: $h"
echo "gate T4: ok"
