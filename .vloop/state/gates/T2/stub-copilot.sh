#!/bin/sh
# The gates' stub copilot. It never touches the network. Every invocation
# appends to log, in a form that does not depend on the binary that started
# it: a header, each argument on its own line, the working directory, and the
# full environment sorted by name. It answers --version (failing when the file
# copilot.noversion exists). For a -p prompt it accepts only /vloop:review
# <task>, and before playing the review it records what it was given:
#
#   home <state>        COPILOT_HOME: "unset", "missing", or "empty"/"used"
#                       for a directory with nothing / something in it; it then
#                       leaves a file there, so a later session handed the same
#                       home finds it used
#   plugin <state>      whether --plugin-dir holds skills/review/SKILL.md
#   guidelines <state>  "found" when a file under a directory listed in
#                       COPILOT_CUSTOM_INSTRUCTIONS_DIRS (comma or colon
#                       separated) holds the fixture's lesson, else "absent"
#   deny <pattern>      every --deny-tool pattern, either form
#                       (--deny-tool=P, or --deny-tool P [P...] up to the next
#                       argument that starts with -)
#   probe <verdict> <kind> <subject>
#                       for each probe below, "denied" when a deny pattern
#                       matches it and "allowed" when none does
#
# Matching models Copilot's Kind(argument) patterns: a bare kind ("shell",
# "url", "write") matches every use of that kind; shell(<stem>:*) matches the
# command <stem> alone or followed by a space; shell(<command>) matches that
# exact command only. Any other pattern matches nothing.
#
# It then writes a passing verdict and prints JSON Lines: three
# assistant.turn_start events, two session.usage_checkpoint events (the last at
# 14.2 credits; none when copilot.nocheckpoint exists) and a result event with
# exitCode 0 (the content of copilot.exitcode when it exists),
# sessionDurationMs 18000 and a codeChanges list holding an absolute path (no
# result event when copilot.noresult exists). It exits 0.
dir=$(dirname "$0")
{
  echo "=== copilot"
  for a in "$@"; do printf 'arg %s\n' "$a"; done
  printf 'cwd %s\n' "$(pwd -P)"
  env | LC_ALL=C sort | sed 's/^/env /'
} >>"$dir/log"
case "${1:-}" in
  --version) [ -f "$dir/copilot.noversion" ] && { echo "copilot: broken" >&2; exit 1; }; echo "GitHub Copilot CLI 1.0.101"; exit 0;;
esac
prompt="" plugin="" pats="$dir/pats.$$" indeny=0
: >"$pats"
while [ $# -gt 0 ]; do
  if [ "$indeny" = 1 ]; then
    case "$1" in -*) indeny=0;; *) printf '%s\n' "$1" >>"$pats"; shift; continue;; esac
  fi
  case "$1" in
    -p) prompt="${2:-}"; shift 2;;
    --plugin-dir) plugin="${2:-}"; shift 2;;
    --plugin-dir=*) plugin="${1#--plugin-dir=}"; shift;;
    --deny-tool=*) printf '%s\n' "${1#--deny-tool=}" >>"$pats"; shift;;
    --deny-tool) indeny=1; shift;;
    *) shift;;
  esac
done
cmd="${prompt%% *}" TASK="${prompt#* }"
if [ "$cmd" != /vloop:review ] || [ -z "$TASK" ] || [ "$TASK" = "$prompt" ]; then
  echo "stub copilot: unrecognised prompt: $prompt" >&2; rm -f "$pats"; exit 64
fi
{
  if [ -z "${COPILOT_HOME+x}" ]; then echo "home unset"
  elif [ ! -d "$COPILOT_HOME" ]; then echo "home missing"
  elif [ -z "$(ls -A "$COPILOT_HOME")" ]; then echo "home empty"
  else echo "home used"; fi
  if [ -n "$plugin" ] && [ -f "$plugin/skills/review/SKILL.md" ]; then echo "plugin ok"; else echo "plugin missing"; fi
  g=absent
  for d in $(printf '%s' "${COPILOT_CUSTOM_INSTRUCTIONS_DIRS:-}" | tr ',:' '  '); do
    [ -d "$d" ] && grep -rqF -- "- A lesson of this repository, for the gate." "$d" 2>/dev/null && g=found
  done
  echo "guidelines $g"
  sed 's/^/deny /' "$pats"
} >>"$dir/log"
[ -d "${COPILOT_HOME:-/nonexistent}" ] && echo "a stub session was here" >"$COPILOT_HOME/stub-session-marker"

# matches <kind> <subject>: whether any deny pattern matches.
matches() {
  while IFS= read -r p; do
    case "$p" in
      "$1") return 0;;
    esac
    [ "$1" = shell ] || continue
    case "$p" in
      'shell('*':*)') s=${p#shell(}; s=${s%:\*)}
        case "$2" in "$s"|"$s "*) return 0;; esac;;
      'shell('*')') s=${p#shell(}; s=${s%)}
        [ "$2" = "$s" ] && return 0;;
    esac
  done <"$pats"
  return 1
}
probe() {
  if matches "$1" "$2"; then v=denied; else v=allowed; fi
  printf 'probe %s %s %s\n' "$v" "$1" "$2" >>"$dir/log"
}
while IFS= read -r c; do
  [ -n "$c" ] && probe shell "$c"
done <<'EOF'
gh pr view 1
gh api repos/o/r
gh auth status
git push origin main
git push
git remote add other https://example.invalid/r.git
git remote -v
git config --global user.name x
git config --system core.editor x
git config --file ../outside.cfg a.b c
curl https://example.invalid
wget https://example.invalid
ssh host.invalid
scp a host.invalid:b
nc host.invalid 80
rm -rf build
rm x
git commit -m x
git reset --hard
git checkout main
git switch main
git stash
git branch x
git tag v1
git diff
git diff HEAD
git status --short
git log -1
git show HEAD
vloop task show T1
printf x
cat .vloop/tmp/proposal.json
jq . .vloop/tmp/proposal.json
go test ./...
ls
grep -rn x .
EOF
probe url https://example.invalid/page
probe write .vloop/tmp/verdict.json
rm -f "$pats"

mkdir -p .vloop/tmp
printf '{"schema":"verdict/v1","task":"%s","verdict":"PASS","criteria":[],"findings":[],"notes":"none"}\n' "$TASK" >.vloop/tmp/verdict.json
echo "stub copilot review session" >&2
here=$(pwd -P)
printf '{"type":"session.start","data":{"sessionId":"stub-session"}}\n'
for i in 1 2 3; do
  printf '{"type":"assistant.turn_start","data":{"turnId":"%s"}}\n' "$i"
  if [ "$i" = 2 ] && [ ! -f "$dir/copilot.nocheckpoint" ]; then
    printf '{"type":"session.usage_checkpoint","data":{"totalNanoAiu":5100000000}}\n'
  fi
done
[ -f "$dir/copilot.nocheckpoint" ] || printf '{"type":"session.usage_checkpoint","data":{"totalNanoAiu":14200000000}}\n'
if [ ! -f "$dir/copilot.noresult" ]; then
  ec=$(cat "$dir/copilot.exitcode" 2>/dev/null || echo 0)
  printf '{"type":"result","sessionId":"stub-session","exitCode":%s,"usage":{"premiumRequests":1,"totalApiDurationMs":12000,"sessionDurationMs":18000,"codeChanges":{"linesAdded":1,"linesRemoved":0,"filesModified":["%s/%s.out"]}}}\n' "$ec" "$here" "$TASK"
fi
exit 0
