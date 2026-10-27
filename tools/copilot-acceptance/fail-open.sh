#!/bin/sh
# Does a preToolUse hook that means "deny" stop the call when it answers late or
# not in JSON? One scratch repo and COPILOT_HOME per case; the prompt asks for
# exactly one shell call, `touch marker`. The marker existing afterwards means
# Copilot ran the call.
#
#   fail-open.sh <model> [workdir]
#
# Needs copilot and gh (logged in: a fresh COPILOT_HOME logs in through gh).
set -u
model=${1:?usage: fail-open.sh <model> [workdir]}
work=${2:-$(mktemp -d)}
mkdir -p "$work"
copilot --version | head -1

# case name | hook command | timeoutSec
cases='deny-json|printf %s "{\"permissionDecision\":\"deny\",\"permissionDecisionReason\":\"blocked by test\"}"|10
exit-2|echo blocked >&2; exit 2|10
exit-1|exit 1|10
log-line-then-deny|echo "checking command..."; printf %s "{\"permissionDecision\":\"deny\",\"permissionDecisionReason\":\"blocked by test\"}"|10
not-json|echo deny|10
slow-deny|sleep 8; printf %s "{\"permissionDecision\":\"deny\",\"permissionDecisionReason\":\"blocked by test\"}"|2'

printf '\n| case | hook | marker created (call ran) | denial message |\n|---|---|---|---|\n' >"$work/summary.md"
echo "$cases" | while IFS='|' read -r name cmd timeout; do
  d=$work/$name
  rm -rf "$d"; mkdir -p "$d/repo" "$d/home/hooks"
  git -C "$d/repo" init -q
  printf '%s\n' "$cmd" >"$d/hook.sh"
  cat >"$d/home/hooks/test.json" <<JSON
{"version": 1, "hooks": {"preToolUse": [{"type": "command", "bash": "sh '$d/hook.sh'", "timeoutSec": $timeout}]}}
JSON
  (cd "$d/repo" && COPILOT_HOME=$d/home copilot -p "Run exactly this shell command, once, as one tool call: touch marker. If it is denied, do not retry and do not try another way." \
    --model "$model" --no-ask-user --output-format json --allow-all-tools --disable-builtin-mcps \
    >"$d/out.jsonl" 2>"$d/err.txt")
  ran=no; [ -e "$d/repo/marker" ] && ran=YES
  msg=$(grep -o '"error":{"message":"[^}]*"code":"denied"' "$d/out.jsonl" | head -1 | sed 's/^"error":{"message":"//; s/","code":"denied"$//; s|"[^"]*hooks/test.json\\"|<hooks file>|')
  printf '| %s | `%s` (timeoutSec %s) | %s | %s |\n' "$name" "$cmd" "$timeout" "$ran" "$msg" >>"$work/summary.md"
  echo "$name: ran=$ran"
done
cat "$work/summary.md"
