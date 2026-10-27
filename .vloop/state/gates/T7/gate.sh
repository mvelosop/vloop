#!/bin/sh
# Gate T7 — A1, the audit and the docs. The fence audit classifies every deny
# rule of fence/review.json and every Copilot deny-list pattern by who enforces
# it (driver, hook or neither); the configuration guide has a row for each S1
# key with its default and variable, names the modes and the hook, and no
# longer says copilot reviews are unfenced; the domain pages carry the second
# review, dissent and the hook; both operator playbooks read the dissents at
# close. Read-only: it reads this tree's documents and nothing else.
set -u
fail() { echo "GATE FAIL: $*" >&2; exit 1; }
export LC_ALL=C
JQ=$(command -v jq) || fail "jq is not on PATH"
AUDIT=docs/design-notes/fence-audit.md
GUIDE=docs/guide/configuration.md
[ -f "$AUDIT" ] || fail "$AUDIT does not exist"
# classified <text>: a line of the audit holds the text and a classification.
classified() {
  grep -F -- "$1" "$AUDIT" | grep -qiE 'driver|hook|neither' || fail "$AUDIT does not classify $1 (driver, hook or neither) on its line"
}
"$JQ" -r '.permissions.deny[]' fence/review.json >"${TMPDIR:-/tmp}/gate-t7-rules.$$" || fail "fence/review.json unreadable"
trap 'rm -f "${TMPDIR:-/tmp}/gate-t7-rules.$$"' EXIT
n=0
while IFS= read -r r; do classified "$r"; n=$((n + 1)); done <"${TMPDIR:-/tmp}/gate-t7-rules.$$"
[ "$n" -gt 50 ] || fail "read only $n deny rules from fence/review.json"
# The adapter's deny list as it stood when this plan was written.
for p in 'shell(gh:*)' 'shell(git push:*)' 'shell(git remote:*)' 'shell(git config --global:*)' 'shell(git config --system:*)' \
  'shell(git config --file:*)' 'shell(curl:*)' 'shell(wget:*)' 'shell(ssh:*)' 'shell(scp:*)' 'shell(nc:*)' 'shell(rm:*)' \
  'shell(git commit:*)' 'shell(git reset:*)' 'shell(git checkout:*)' 'shell(git switch:*)' 'shell(git stash:*)' \
  'shell(git branch:*)' 'shell(git tag:*)' '`url`'; do
  classified "$p"
done
grep -qiw neither "$AUDIT" || fail "$AUDIT classifies no rule as neither — say what covers those, or say none is"

# The configuration guide.
row() { grep -E "^\|[[:space:]]*\`$1\`[[:space:]]*\|" "$GUIDE" || fail "$GUIDE has no table row for $1"; }
col() { echo "$1" | awk -F'|' -v n="$2" '{ c = $(n + 1); gsub(/^[ \t]+|[ \t]+$/, "", c); print c }'; }
r=$(row 'run\.second-review')
[ "$(col "$r" 2)" = '`off`' ] || fail "the run.second-review row's default is '$(col "$r" 2)', want \`off\`"
for m in off record feedback on-pass both-pass; do col "$r" 3 | grep -qF "\`$m\`" || fail "the run.second-review row's valid values lack \`$m\`: $r"; done
[ "$(col "$r" 4)" = '`VLOOP_RUN_SECOND_REVIEW`' ] || fail "the run.second-review row's variable is '$(col "$r" 4)'"
r=$(row 'agent\.second-review')
[ "$(col "$r" 2)" = '`claude`' ] || fail "the agent.second-review row's default is '$(col "$r" 2)', want \`claude\`"
col "$r" 3 | grep -qF '`copilot`' || fail "the agent.second-review row does not offer \`copilot\`: $r"
[ "$(col "$r" 4)" = '`VLOOP_AGENT_SECOND_REVIEW`' ] || fail "the agent.second-review row's variable is '$(col "$r" 4)'"
r=$(row 'model\.second-review')
[ "$(col "$r" 2)" = '`sonnet`' ] || fail "the model.second-review row's default is '$(col "$r" 2)', want \`sonnet\`"
[ "$(col "$r" 4)" = '`VLOOP_MODEL_SECOND_REVIEW`' ] || fail "the model.second-review row's variable is '$(col "$r" 4)'"
r=$(row 'effort\.second-review')
[ "$(col "$r" 4)" = '`VLOOP_EFFORT_SECOND_REVIEW`' ] || fail "the effort.second-review row's variable is '$(col "$r" 4)'"
grep -qF 'vloop fence hook' "$GUIDE" || fail "$GUIDE does not describe vloop fence hook"
grep -qF 'vloop fence check' "$GUIDE" || fail "$GUIDE does not describe vloop fence check"
grep -qi 'experimental' "$GUIDE" || fail "$GUIDE no longer says copilot reviews are experimental"
grep -qiE 'not fenced|no fence|until the fence lands' "$GUIDE" && fail "$GUIDE still says copilot reviews are not fenced: $(grep -inE 'not fenced|no fence|until the fence lands' "$GUIDE")"
grep -qi 'dissent' "$GUIDE" || fail "$GUIDE does not say what a dissent is"

# The domain pages.
grep -qi 'second review' docs/domain/domain-model.md && grep -qi 'dissent' docs/domain/domain-model.md || fail "docs/domain/domain-model.md lacks the second review or dissent"
grep -qF 'second-review' docs/domain/execution/session.md && grep -qi 'hook' docs/domain/execution/session.md || fail "docs/domain/execution/session.md lacks the second-review kind or the hook"
grep -qiE 'second review|second_review' docs/domain/measurement/metrics.md || fail "docs/domain/measurement/metrics.md lacks the second review"
grep -qi 'second review' docs/domain/execution/task.md || fail "docs/domain/execution/task.md does not say what the second review's findings do to a task's notes"

# The playbooks read the dissents at close.
for s in plugin/skills/operate/SKILL.md .claude/skills/vloop-operator/SKILL.md; do
  grep -qi 'dissent' "$s" || fail "$s does not read the dissents at close"
done
echo "gate T7: ok"
