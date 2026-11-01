#!/bin/sh
# The gates' stub copilot. It never touches the network. It answers
# --version. For a -p prompt it accepts /vloop:plan <brief> (role plan, the
# brief path in the task's place), /vloop:work <task> (role work),
# /vloop:review <task> (role review), /vloop:verify (role verify, task -) and
# /vloop:gate-review (role gate-review, task -), and appends to events:
#
#   session copilot-<role> <task> model=<model>
#   deny-tool <role> <pattern>   every --deny-tool pattern (--deny-tool=P form,
#                                or --deny-tool P [P...])
#   hooks <role> <n> bash=<yes|no> powershell=<yes|no>
#                                n: the preToolUse entries found in the JSON
#                                files of $COPILOT_HOME/hooks/; bash and
#                                powershell: whether every entry has that form
#   hookfile <role> <phases>     the phases named after "fence hook --phase "
#                                in those files, sorted, space-separated
#                                ("none" when there is none)
#   hook-phase <role> <phase>    the word after --phase in the bash form
#   hook <role> <label> <decision> <exit>
#                                the bash form of the first entry, run as
#                                Copilot runs a hook: sh -c, from the
#                                repository root, one preToolUse payload on
#                                stdin; decision is the permissionDecision of
#                                its stdout, "none" when there is none
#
# A plan session then copies plan.json beside the stub to the plan file and,
# when the driver sent the plan back (.vloop/tmp/plan-problems.md exists),
# appends that file to problems beside the stub after a line
# "--- plan session"; a work session writes <task>.out and a done proposal under .vloop/tmp/;
# a review session writes a passing verdict; a verify session writes a valid
# verify/v1 report with no findings to .vloop/tmp/verify.json; a gate-review
# session writes a valid gate-verdict/v2 PASS with no findings to
# .vloop/tmp/gate-verdict.json. Files beside the stub change what a verify
# session does:
#   copilot.verify-moveref    it first makes a commit (HEAD moves)
#   copilot.verify-readme     it also appends a line to README.md
#   copilot.verify-finding    its report holds one finding
#   copilot.verify-noreport   it writes no report
#   copilot.verify-fail       it exits 1 after its output
# what a plan session does:
#   copilot.plan-moveref      it first makes a commit (HEAD moves)
#   copilot.plan-noplan       it writes no plan
#   copilot.plan-exit3        it exits 3 after its output
# and what a gate-review session does:
#   copilot.gr-moveref        it first makes a commit (HEAD moves)
#   copilot.gr-noverdict      it writes no verdict
#   copilot.gr-invalid        its verdict does not follow gate-verdict/v2
#   copilot.gr-fail-once      its first session writes a valid FAIL verdict
#                             with one wrong-reason finding on T1, later
#                             ones the PASS (gr-count beside the stub counts)
#   copilot.gr-checkgap       its verdict is a FAIL with one check-gap finding
# All roles print JSON Lines like Copilot's: turns, usage checkpoints, one
# tool call denied by the hook (c1, bash) and one completed (c2, view), and a
# result event.
dir=$(dirname "$0")
case "${1:-}" in
  --version) echo "GitHub Copilot CLI 1.0.103"; exit 0;;
esac
prompt="" MODEL="" indeny=0
tmpev="$dir/events.tmp.$$"
: >"$tmpev"
while [ $# -gt 0 ]; do
  if [ "$indeny" = 1 ]; then
    case "$1" in -*) indeny=0;; *) echo "$1" >>"$tmpev"; shift; continue;; esac
  fi
  case "$1" in
    -p) prompt="${2:-}"; shift 2;;
    --model) MODEL="${2:-}"; shift 2;;
    --deny-tool=*) echo "${1#--deny-tool=}" >>"$tmpev"; shift;;
    --deny-tool) indeny=1; shift;;
    *) shift;;
  esac
done
cmd="${prompt%% *}" TASK="${prompt#* }"
[ "$TASK" = "$prompt" ] && TASK=-
case "$cmd" in
  /vloop:plan) ROLE=plan;;
  /vloop:work) ROLE=work;;
  /vloop:review) ROLE=review;;
  /vloop:verify) ROLE=verify;;
  /vloop:gate-review) ROLE=gate-review;;
  *) ROLE="";;
esac
ok=yes
[ -n "$ROLE" ] || ok=no
case "$ROLE" in verify|gate-review) [ "$TASK" != - ] && ok=no;; *) [ "$TASK" = - ] && ok=no;; esac
if [ "$ok" = no ]; then
  echo "stub copilot: unrecognised prompt: $prompt" >&2; rm -f "$tmpev"; exit 64
fi
here=$(pwd -P)
{
  echo "session copilot-$ROLE $TASK model=$MODEL"
  sed "s/^/deny-tool $ROLE /" "$tmpev"
} >>"$dir/events"
rm -f "$tmpev"

phases=$(cat "${COPILOT_HOME:-/nonexistent}"/hooks/*.json 2>/dev/null | grep -oE 'fence hook --phase [a-z-]+' | sed 's/^fence hook --phase //' | LC_ALL=C sort -u | tr '\n' ' ' | sed 's/ $//')
entries=$(cat "${COPILOT_HOME:-/nonexistent}"/hooks/*.json 2>/dev/null |
  jq -c '.. | objects | select(has("preToolUse")) | .preToolUse | if type == "array" then .[] else . end | objects' 2>/dev/null)
n=0; [ -n "$entries" ] && n=$(printf '%s\n' "$entries" | grep -c .)
hb=no; hp=no
if [ "$n" -gt 0 ]; then
  [ "$(printf '%s\n' "$entries" | jq -s 'all(.bash | type == "string" and length > 0)')" = true ] && hb=yes
  [ "$(printf '%s\n' "$entries" | jq -s 'all(.powershell | type == "string" and length > 0)')" = true ] && hp=yes
fi
{
  echo "hooks $ROLE $n bash=$hb powershell=$hp"
  echo "hookfile $ROLE ${phases:-none}"
} >>"$dir/events"
hookcmd=$(printf '%s\n' "$entries" | head -1 | jq -r '.bash // empty' 2>/dev/null)
if [ -n "$hookcmd" ]; then
  echo "hook-phase $ROLE $(printf '%s\n' "$hookcmd" | awk '{ for (i = 1; i < NF; i++) if ($i == "--phase") { v = $(i + 1); gsub(/["'\'']/, "", v); print v; exit } }')" >>"$dir/events"
  # hook <label> <toolName> <toolArgs JSON>
  hook() {
    p=$(jq -cn --arg n "$2" --argjson a "$3" --arg c "$here" '{timestamp: 1, cwd: $c, toolName: $n, toolArgs: $a}')
    out=$(printf '%s' "$p" | sh -c "$hookcmd" 2>/dev/null); x=$?
    d=$(printf '%s\n' "$out" | head -1 | jq -r '.permissionDecision // "none"' 2>/dev/null)
    echo "hook $ROLE $1 ${d:-none} $x" >>"$dir/events"
  }
  bashargs() { jq -cn --arg c "$1" '{command: $c, description: "d"}'; }
  fileargs() { jq -cn --arg p "$1" '{path: $p, file_text: "x"}'; }
  case "$ROLE" in
    work)
      hook rmfile bash "$(bashargs 'rm src/old.ts')"
      hook rmrf bash "$(bashargs 'rm -rf dist')"
      hook srcrel create "$(fileargs src/a.ts)"
      hook stateabs create "$(fileargs "$here/.vloop/state/x.json")"
      ;;
    review)
      hook shgh bash "$(bashargs 'sh -c "gh pr list"')"
      hook diff bash "$(bashargs 'git diff HEAD')"
      hook rmfile bash "$(bashargs 'rm src/old.ts')"
      hook srcrel create "$(fileargs src/a.ts)"
      hook tmprel create "$(fileargs .vloop/tmp/y.json)"
      ;;
    verify)
      hook build bash "$(bashargs 'go build -o .vloop/tmp/v/vloop ./cmd/vloop')"
      hook cdout bash "$(bashargs 'cd .vloop/tmp/v && ./vloop status > out.txt')"
      hook rmtmp bash "$(bashargs 'rm .vloop/tmp/v/out.txt')"
      hook gitc bash "$(bashargs 'git -C .vloop/tmp/v/ex1 status')"
      hook shpush bash "$(bashargs 'sh -c "git push origin main"')"
      hook envgh bash "$(bashargs 'env gh pr list')"
      hook srcrel edit "$(fileargs internal/x.go)"
      hook readme bash "$(bashargs 'echo x > README.md')"
      hook tmprel create "$(fileargs .vloop/tmp/verify.json)"
      hook tmpabs create "$(fileargs "$here/.vloop/tmp/v/x.txt")"
      hook staterel create "$(fileargs .vloop/state/state.json)"
      hook viewsrc view "$(fileargs src/app.go)"
      hook skillv skill '{"skill":"verify"}'
      hook skillother skill '{"skill":"vloop-verify"}'
      hook webfetch web_fetch '{"url":"https://example.invalid"}'
      ;;
    plan)
      hook gate bash "$(bashargs 'vloop task gate T1')"
      hook validate bash "$(bashargs 'vloop schema validate state/v2 .vloop/state/state.json')"
      hook statewrite bash "$(bashargs "printf '{}' > .vloop/state/state.json")"
      hook gatewrite bash "$(bashargs "echo 'exit 1' > .vloop/state/gates/T1/gate.sh")"
      hook tmpwrite bash "$(bashargs 'echo x > .vloop/tmp/p/out.txt')"
      hook journal bash "$(bashargs 'echo x > .vloop/state/journals/a.md')"
      hook srcwrite bash "$(bashargs 'echo x > src/app.go')"
      hook cdf bash "$(bashargs 'cd .vloop/state/gates/T1 && echo x > f')"
      hook rmgate bash "$(bashargs 'rm .vloop/state/gates/T1/old.sh')"
      hook shgate bash "$(bashargs 'sh .vloop/state/gates/T1/gate.sh')"
      hook taskdrop bash "$(bashargs 'vloop task drop T2')"
      hook shpush bash "$(bashargs 'sh -c "git push origin main"')"
      hook envgh bash "$(bashargs 'env gh pr list')"
      hook subst bash "$(bashargs 'echo $(date)')"
      hook staterel create "$(fileargs .vloop/state/state.json)"
      hook oracle create "$(fileargs .vloop/state/gates/T1/oracle.test.ts)"
      hook gateabs create "$(fileargs "$here/.vloop/state/gates/T1/x.sh")"
      hook notes create "$(fileargs .vloop/tmp/notes.md)"
      hook planmd edit "$(fileargs .vloop/state/plan.md)"
      hook config create "$(fileargs .vloop/config.toml)"
      hook gatesdir create "$(fileargs .vloop/state/gates)"
      hook viewbrief view "$(fileargs docs/briefs/b.loop-brief.md)"
      hook skillp skill '{"skill":"plan"}'
      hook skillother skill '{"skill":"vloop-plan"}'
      hook webfetch web_fetch '{"url":"https://example.invalid"}'
      ;;
    gate-review)
      hook gate bash "$(bashargs 'vloop task gate T1')"
      hook show bash "$(bashargs 'vloop task show T1 --json')"
      hook cdout bash "$(bashargs 'cd .vloop/tmp/g && ls > out.txt')"
      hook gateedit bash "$(bashargs 'echo x > .vloop/state/gates/T1/check.sh')"
      hook rmtmp bash "$(bashargs 'rm .vloop/tmp/g/out.txt')"
      hook taskverify bash "$(bashargs "vloop task verify T1 'true' --reason x")"
      hook gitc bash "$(bashargs 'git -C .vloop/tmp/g status')"
      hook shpush bash "$(bashargs 'sh -c "git push origin main"')"
      hook envgh bash "$(bashargs 'env gh pr list')"
      hook verdictrel create "$(fileargs .vloop/tmp/gate-verdict.json)"
      hook verdictabs create "$(fileargs "$here/.vloop/tmp/g/x.txt")"
      hook staterel edit "$(fileargs .vloop/state/state.json)"
      hook srcrel create "$(fileargs src/app.go)"
      hook viewgate view "$(fileargs .vloop/state/gates/T1/check.sh)"
      hook skillg skill '{"skill":"gate-review"}'
      hook skillother skill '{"skill":"vloop-gate-review"}'
      hook webfetch web_fetch '{"url":"https://example.invalid"}'
      ;;
  esac
fi

mkdir -p .vloop/tmp
xit=0
case "$ROLE" in
  plan)
    [ -f "$dir/copilot.plan-moveref" ] && git commit -q --allow-empty -m "moved by the stub plan session"
    if [ ! -f "$dir/copilot.plan-noplan" ]; then mkdir -p .vloop/state && cp "$dir/plan.json" .vloop/state/state.json; fi
    if [ -f .vloop/tmp/plan-problems.md ]; then echo "--- plan session" >>"$dir/problems"; cat .vloop/tmp/plan-problems.md >>"$dir/problems"; fi
    [ -f "$dir/copilot.plan-exit3" ] && xit=3
    denied="vloop task drop T2 is denied by the plan fence (Bash(vloop task drop:*))"
    ;;
  work)
    echo "$TASK by copilot" >>"$TASK.out"
    printf '{"schema":"proposal/v1","task":"%s","outcome":"done","summary":"made %s","files":[],"verified":"ok","notes":"none"}\n' "$TASK" "$TASK" >.vloop/tmp/proposal.json
    denied="rm -rf dist is denied by the work fence (Bash(rm -rf:*))"
    ;;
  review)
    printf '{"schema":"verdict/v1","task":"%s","verdict":"PASS","criteria":[],"findings":[],"notes":"copilot"}\n' "$TASK" >.vloop/tmp/verdict.json
    denied="gh pr list is denied by the Copilot deny list (gh:*)"
    ;;
  verify)
    [ -f "$dir/copilot.verify-moveref" ] && git commit -q --allow-empty -m "moved by the stub verify session"
    [ -f "$dir/copilot.verify-readme" ] && echo "written by the stub verify session" >>README.md
    if [ ! -f "$dir/copilot.verify-noreport" ]; then
      if [ -f "$dir/copilot.verify-finding" ]; then
        echo '{"schema":"verify/v1","findings":[{"kind":"example","claim":"demo prints ok","observed":"demo printed nothing","where":"Worked example, line 1"}],"checked":{"example_lines":1,"sentences":0,"surfaces":0},"notes":"copilot"}' >.vloop/tmp/verify.json
      else
        echo '{"schema":"verify/v1","findings":[],"checked":{"example_lines":1,"sentences":0,"surfaces":0},"notes":"copilot"}' >.vloop/tmp/verify.json
      fi
    fi
    [ -f "$dir/copilot.verify-fail" ] && xit=1
    denied="rm .vloop/tmp/v/out.txt is denied by the Copilot deny list (rm:*)"
    ;;
  gate-review)
    [ -f "$dir/copilot.gr-moveref" ] && git commit -q --allow-empty -m "moved by the stub gate-review session"
    n=$(( $(cat "$dir/gr-count" 2>/dev/null || echo 0) + 1 )); echo "$n" >"$dir/gr-count"
    v='{"schema":"gate-verdict/v2","verdict":"PASS","tasks":[{"task":"T1","verdict":"PASS","findings":[]}],"notes":"copilot"}'
    [ -f "$dir/copilot.gr-fail-once" ] && [ "$n" = 1 ] &&
      v='{"schema":"gate-verdict/v2","verdict":"FAIL","tasks":[{"task":"T1","verdict":"FAIL","findings":[{"summary":"the gate passes on a stub","kind":"wrong-reason"}]}],"notes":"copilot round one"}'
    [ -f "$dir/copilot.gr-checkgap" ] &&
      v='{"schema":"gate-verdict/v2","verdict":"FAIL","tasks":[{"task":"T1","verdict":"FAIL","findings":[{"summary":"no check runs the tests","kind":"check-gap"}]}],"notes":"copilot"}'
    [ -f "$dir/copilot.gr-invalid" ] && v='{"schema":"gate-verdict/v2","verdict":"MAYBE","notes":"copilot"}'
    [ -f "$dir/copilot.gr-noverdict" ] || printf '%s\n' "$v" >.vloop/tmp/gate-verdict.json
    denied="vloop task verify T1 true --reason x is denied by the gate-review fence (Bash(vloop task verify:*))"
    ;;
esac
echo "stub copilot $ROLE session" >&2
printf '{"type":"session.start","data":{"sessionId":"stub-session"}}\n'
printf '{"type":"assistant.turn_start","data":{"turnId":"1"}}\n'
printf '{"type":"session.usage_checkpoint","data":{"totalNanoAiu":5100000000}}\n'
printf '{"type":"tool.execution_start","data":{"toolCallId":"c1","toolName":"bash","arguments":{"command":"x"}}}\n'
printf '{"type":"tool.execution_complete","data":{"toolCallId":"c1","success":false,"error":{"code":"denied","message":"Denied by preToolUse hook: vloop fence: %s"}}}\n' "$denied"
printf '{"type":"tool.execution_start","data":{"toolCallId":"c2","toolName":"view","arguments":{"path":"docs/x.md"}}}\n'
printf '{"type":"tool.execution_complete","data":{"toolCallId":"c2","success":true}}\n'
printf '{"type":"assistant.turn_start","data":{"turnId":"2"}}\n'
printf '{"type":"session.usage_checkpoint","data":{"totalNanoAiu":8400000000}}\n'
printf '{"type":"result","sessionId":"stub-session","exitCode":%d,"usage":{"premiumRequests":1,"totalApiDurationMs":12000,"sessionDurationMs":18000}}\n' "$xit"
exit "$xit"
