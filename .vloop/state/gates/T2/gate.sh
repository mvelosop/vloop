#!/bin/sh
# Gate T2 — W1 and W2: Copilot work sessions. `vloop fence hook --phase work`
# answers Copilot payloads with the work phase's decisions and reasons, the
# review phases unchanged, an unknown phase the deny line and exit 2.
# agent.work accepts copilot; agent.plan, agent.gate-review and agent.verify
# still do not. A fixture run with agent.work = copilot (the stub copilot, see
# stub-copilot.sh) does the one task in one commit: its work session record is
# phase work, agent copilot, the model set, session/v4 valid, with its
# denials; the stub's argv keeps the deny list but not shell(rm:*); its hooks
# file runs `fence hook --phase work`, and that hook answers the work fence's
# decisions. A copilot review in the same run keeps its argv and phase; a
# Claude work session keeps the work fence; a work session that moves HEAD
# halts the run with exit 9.
G=$(cd "$(dirname "$0")" && pwd -P) || { echo "GATE FAIL: no gate folder" >&2; exit 1; }
. "$G/lib.sh"
cd "$top" || fail "repo root"
newbin
WORK="agent.work copilot
model.work $MODEL"

# The hook, phase work, on its own.
R=$t/hookroot
mkdir -p "$R/src" "$R/.vloop/tmp" "$t/home" "$t/tmp" "$t/stub" || fail mkdir
echo x >"$R/src/app.ts"
snap() { (cd "$t" && find hookroot home -print | LC_ALL=C sort && cat hookroot/src/app.ts); }
before=$(snap)
# hookrun <phase> <payload>: $rc the exit, $dec and $why the decision and
# reason of its one line of stdout.
hookrun() {
  out=$(printf '%s' "$2" | (cd "$R" && venv "$S" fence hook --phase "$1" --root "$R") 2>"$t/err"); rc=$?
  [ "$(printf '%s\n' "$out" | grep -c .)" = 1 ] || fail "[$1] payload $2: want one line on stdout, got '$out'"
  dec=$(printf '%s\n' "$out" | "$JQ" -r '.permissionDecision // "none"' 2>/dev/null) || fail "[$1] payload $2: stdout is not JSON: '$out'"
  why=$(printf '%s\n' "$out" | "$JQ" -r '.permissionDecisionReason // ""')
}
payload() { "$JQ" -cn --arg n "$1" --argjson a "$2" --arg c "$R" '{timestamp: 1700000000000, cwd: $c, toolName: $n, toolArgs: $a}'; }
bash_() { payload bash "$("$JQ" -cn --arg c "$1" '{command: $c, description: "d"}')"; }
file_() { payload "$1" "$("$JQ" -cn --arg p "$2" '{path: $p, file_text: "x"}')"; }
patch_() { payload apply_patch "$("$JQ" -cn --arg p "$1" '{input: $p}')"; }
ok() {
  hookrun "$1" "$2"
  [ "$rc" = 0 ] && [ "$dec" = allow ] || fail "[$1] payload $2: want allow and exit 0, got '$out', exit $rc"
}
# no <phase> <payload> [exact reason]
no() {
  hookrun "$1" "$2"
  [ "$rc" = 0 ] && [ "$dec" = deny ] || fail "[$1] payload $2: want deny and exit 0, got '$out', exit $rc"
  case "$why" in "vloop fence: "?*) ;; *) fail "[$1] payload $2: the reason '$why' does not start 'vloop fence: '";; esac
  [ -z "${3:-}" ] || [ "$why" = "$3" ] || fail "[$1] payload $2: the reason is '$why', want '$3'"
}
ok work "$(bash_ 'rm src/old.ts')"
ok work "$(bash_ 'pnpm add nanoid')"
ok work "$(bash_ 'echo ok > src/version.txt')"
no work "$(bash_ 'rm -rf dist')" 'vloop fence: rm -rf dist is denied by the work fence (Bash(rm -rf:*))'
no work "$(bash_ 'curl https://example.com')" 'vloop fence: curl https://example.com is denied by the work fence (Bash(curl:*))'
no work "$(bash_ 'env gh pr list')"
no work "$(bash_ 'for f in a b; do echo $f; done')"
no work "$(bash_ 'echo x > ../outside.txt')" 'vloop fence: write to ../outside.txt is not allowed in a work session: it is outside the repository'
ok work "$(file_ create src/links/links.service.ts)"
ok work "$(file_ edit "$R/src/app.ts")"
ok work "$(file_ create .vloop/tmp/proposal.json)"
no work "$(file_ create .vloop/state/state.json)" 'vloop fence: write to .vloop/state/state.json is not allowed in a work session: nothing under .git/ or .vloop/ but .vloop/tmp/'
no work "$(file_ edit .git/config)"
no work "$(file_ create /etc/hosts)" 'vloop fence: write to /etc/hosts is not allowed in a work session: it is outside the repository'
ok work "$(file_ view src/app.ts)"
no work "$(file_ view "$t/home/.copilot/config.json")"
ok work "$(patch_ "$(printf '*** Begin Patch\n*** Update File: src/app.ts\n@@\n-x\n+y\n*** End Patch\n')")"
no work "$(patch_ "$(printf '*** Begin Patch\n*** Add File: .vloop/defects/D1.md\n+x\n*** End Patch\n')")"
ok work "$(payload skill '{"skill":"work"}')"
no work "$(payload skill '{"skill":"vloop-work"}')" "vloop fence: skill vloop-work is not vloop's — use /vloop:work"
no work "$(payload web_fetch '{"url":"https://example.com"}')" 'vloop fence: tool web_fetch is not allowed in a work session'
no work "$(payload search_code_subagent '{"query":"x"}')"
for ph in review second-review; do
  no "$ph" "$(bash_ 'rm src/old.ts')" 'vloop fence: rm src/old.ts is denied by the Copilot deny list (rm:*)'
  no "$ph" "$(file_ create src/a.ts)" "vloop fence: write to src/a.ts is not allowed in a $ph session: only .vloop/tmp/ may be written"
  ok "$ph" "$(file_ create .vloop/tmp/verdict.json)"
  no "$ph" "$(payload skill '{"skill":"vloop-review"}')" "vloop fence: skill vloop-review is not vloop's — use /vloop:review"
done
hookrun plan "$(bash_ 'ls')"
[ "$rc" = 2 ] && [ "$dec" = deny ] || fail "fence hook --phase plan: want the deny line and exit 2, got '$out', exit $rc"
printf '%s' "$why" | grep -qF 'unknown phase "plan": want review, second-review or work' ||
  fail "fence hook --phase plan: the reason '$why' does not hold 'unknown phase \"plan\": want review, second-review or work'"
[ "$(snap)" = "$before" ] || fail "the hook changed the fixture root or the HOME"

# The agent keys.
CFG=; fixture
get() { k=$1; shift; (cd "$t/repo" && venv "$@" "$S" config get "$k" 2>&1); }
set_() { (cd "$t/repo" && venv "$S" config set "$1" "$2" >"$t/set.out" 2>&1); }
set_ agent.work copilot || fail "vloop config set agent.work copilot failed: $(cat "$t/set.out")"
[ "$(get agent.work)" = copilot ] || fail "agent.work set to copilot reads '$(get agent.work)'"
set_ agent.work claude || fail "vloop config set agent.work claude failed: $(cat "$t/set.out")"
[ "$(get agent.work VLOOP_AGENT_WORK=copilot)" = copilot ] || fail "VLOOP_AGENT_WORK=copilot does not set agent.work: '$(get agent.work VLOOP_AGENT_WORK=copilot)'"
set_ agent.work codex; c=$?
[ "$c" = 2 ] || fail "vloop config set agent.work codex exited $c, want 2"
for k in agent.plan agent.gate-review agent.verify; do
  set_ "$k" copilot; c=$?
  [ "$c" = 2 ] || fail "vloop config set $k copilot exited $c, want 2 (it accepts claude only): $(cat "$t/set.out")"
done
for v in VLOOP_AGENT_PLAN VLOOP_AGENT_GATE_REVIEW VLOOP_AGENT_VERIFY; do
  CFG=; fixture
  runfix "$t/p-$v" "$v=copilot"
  [ "$rc" != 0 ] || fail "vloop run with $v=copilot exited 0: it must be refused"
  nosession "$t/p-$v" "the run with $v=copilot"
done

# A run with copilot work sessions.
CFG=$WORK; fixture
runfix "$t/r1"
[ "$rc" = 0 ] || fail "a run with agent.work = copilot exited $rc: $(tail -8 "$t/r1")"
[ "$(task status T1)" = done ] || fail "T1 is $(task status T1) after the run, want done"
[ "$(venv git -C "$t/repo" log --all --format=%H -- T1.out | wc -l | tr -d ' ')" = 1 ] ||
  fail "want one commit carrying T1.out, got: $(venv git -C "$t/repo" log --all --oneline -- T1.out)"
grep -qxF 'T1 by copilot' "$t/repo/T1.out" || fail "T1.out was not written by the copilot work session: $(cat "$t/repo/T1.out" 2>&1)"
[ "$(ev "$t/r1" '^session copilot-work ' | grep -cxF "session copilot-work T1 model=$MODEL")" = 1 ] ||
  fail "want one copilot work session of T1 with model $MODEL: $(ev "$t/r1" '^session ')"
ev "$t/r1" '^session work ' >/dev/null && fail "a claude work session ran with agent.work = copilot: $(ev "$t/r1" '^session work ')"
ev "$t/r1" '^session review T1 ' >/dev/null || fail "the review of T1 did not run on claude: $(ev "$t/r1" '^session ')"
ev "$t/r1" '^fence review T1 ok$' >/dev/null || fail "the claude review session did not run under the review fence: $(ev "$t/r1" '^fence ')"
for p in 'shell(gh:*)' 'shell(git push:*)' 'shell(git commit:*)' 'shell(curl:*)' 'url'; do
  ev "$t/r1" "^deny-tool work " | grep -qxF "deny-tool work $p" || fail "the work session's argv lost --deny-tool=$p: $(ev "$t/r1" '^deny-tool work ')"
done
ev "$t/r1" "^deny-tool work " | grep -qxF 'deny-tool work shell(rm:*)' && fail "the work session's argv holds --deny-tool=shell(rm:*)"
ev "$t/r1" '^hooks work ' | grep -qxE 'hooks work [1-9][0-9]* bash=yes powershell=yes' ||
  fail "the work session found no preToolUse hook with both command forms in COPILOT_HOME/hooks/: $(ev "$t/r1" '^hooks ')"
ev "$t/r1" '^hookfile work ' | grep -qxF 'hookfile work work=yes review=no' ||
  fail "the work session's hooks file does not run 'fence hook --phase work' alone: $(ev "$t/r1" '^hookfile ')"
ev "$t/r1" '^hook-phase work ' | grep -qxF 'hook-phase work work' || fail "the work session's hook is not run with --phase work: $(ev "$t/r1" '^hook-phase ')"
echo "$WORKHOOK" | while IFS= read -r want; do
  ev "$t/r1" "^hook work ${want%% *} " | grep -qxF "hook work $want" ||
    fail "the work session's hook answered '$(ev "$t/r1" "^hook work ${want%% *} ")', want 'hook work $want'"
done || exit 1
recs=$(phaserecs "$t/r1" work)
[ "$(echo "$recs" | grep -c .)" = 1 ] || fail "want one work session record, got: $recs"
f=$recs
[ "$("$JQ" -r .agent "$f")" = copilot ] && [ "$("$JQ" -r .model "$f")" = "$MODEL" ] && [ "$("$JQ" -r .task "$f")" = T1 ] ||
  fail "the work session record is $("$JQ" -c '{phase, agent, model, task}' "$f"), want phase work, agent copilot, model $MODEL, task T1"
[ "$("$JQ" -r .schema "$f")" = session/v4 ] || fail "the work session record's schema is $("$JQ" -r .schema "$f"), want session/v4"
(cd "$t/repo" && venv "$S" schema validate session/v4 "$f" >"$t/val" 2>&1) || fail "the work session record is not valid session/v4: $(cat "$t/val")"
"$JQ" -e '.credits == 8.4 and .cost_source == "estimated"' "$f" >/dev/null || fail "the work session record's credits and cost_source are $("$JQ" -c '{credits, cost_source}' "$f"), want 8.4 and estimated"
"$JQ" -e '(.permission_denials | length) == 1 and .permission_denials[0].tool_name == "bash" and (.permission_denials[0].message | contains("rm -rf dist is denied by the work fence"))' "$f" >/dev/null ||
  fail "the work session record's permission_denials are $("$JQ" -c .permission_denials "$f"), want the one bash denial"

# Copilot work and copilot review in one run: the review keeps its argv and phase.
CFG="$WORK
agent.review copilot
model.review $MODEL"; fixture
runfix "$t/r2"
[ "$rc" = 0 ] || fail "a run with copilot work and copilot review exited $rc: $(tail -8 "$t/r2")"
[ "$(task status T1)" = done ] || fail "copilot work and review: T1 is $(task status T1), want done"
ev "$t/r2" '^session copilot-review T1 ' >/dev/null || fail "no copilot review session: $(ev "$t/r2" '^session ')"
for p in 'shell(rm:*)' 'shell(gh:*)' 'shell(git push:*)' 'url'; do
  ev "$t/r2" "^deny-tool review " | grep -qxF "deny-tool review $p" || fail "the copilot review session's argv lost --deny-tool=$p"
done
ev "$t/r2" "^deny-tool work " | grep -qxF 'deny-tool work shell(rm:*)' && fail "with copilot reviews too, the work session's argv holds --deny-tool=shell(rm:*)"
ev "$t/r2" '^hookfile review ' | grep -qxF 'hookfile review work=no review=yes' || fail "the review session's hooks file is not the review phase's: $(ev "$t/r2" '^hookfile ')"
ev "$t/r2" '^hook-phase review ' | grep -qxF 'hook-phase review review' || fail "the review session's hook is not run with --phase review"
for want in 'shgh deny 0' 'diff allow 0' 'rmfile deny 0' 'srcrel deny 0' 'tmprel allow 0'; do
  ev "$t/r2" "^hook review ${want%% *} " | grep -qxF "hook review $want" ||
    fail "the review session's hook answered '$(ev "$t/r2" "^hook review ${want%% *} ")', want 'hook review $want'"
done
ev "$t/r2" '^hook-phase work ' | grep -qxF 'hook-phase work work' || fail "with copilot reviews too, the work session's hook is not run with --phase work"

# Claude work sessions: unchanged, under the work fence.
CFG=; fixture
runfix "$t/r3"
[ "$rc" = 0 ] || fail "a run with claude work exited $rc: $(tail -8 "$t/r3")"
ev "$t/r3" '^session work T1 model=sonnet$' >/dev/null || fail "no claude work session of T1 with model sonnet: $(ev "$t/r3" '^session ')"
ev "$t/r3" '^fence work T1 ok$' >/dev/null || fail "the claude work session did not run under the work fence: $(ev "$t/r3" '^fence ')"
ev "$t/r3" '^session copilot-' >/dev/null && fail "a copilot session ran with agent.work = claude"

# A copilot work session that moves HEAD: the run halts, exit 9.
CFG=$WORK; fixture
: >"$t/stub/copilot.moveref"
runfix "$t/r4"
[ "$rc" = 9 ] || fail "a copilot work session that moved HEAD: vloop run exited $rc, want 9: $(tail -8 "$t/r4")"
echo "gate T2: ok"
