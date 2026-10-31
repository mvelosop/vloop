#!/bin/sh
# Gate T2 — V1 and V2: Copilot verify sessions. `vloop fence hook --phase
# verify` answers Copilot payloads with the verify phase's decisions and
# reasons, the other phases unchanged, an unknown phase the deny line and exit
# 2. agent.verify accepts copilot; agent.plan and agent.gate-review still do
# not. A fixture run with agent.verify = copilot (the stub copilot, see
# stub-copilot.sh) does the one task in one commit and runs one copilot verify
# session: its record is phase verify, agent copilot, the model set, session/v4
# valid, with its denials; the run folder keeps reports/verify.json; the
# stub's argv carries the whole review deny list, shell(rm:*) included; its
# hooks file runs `fence hook --phase verify`, and that hook answers the verify
# fence's decisions. Copilot work and review in the same run keep their argv
# and phases; with agent.verify = claude, the claude verify session keeps the
# verify fence. Around the session nothing changes: a finding, a failed
# session or a missing report is a warning and exit 0; a changed README.md is
# reverted with the driver's line and the report still kept; a moved HEAD
# halts with exit 9; run.verify = false runs no verify session.
G=$(cd "$(dirname "$0")" && pwd -P) || { echo "GATE FAIL: no gate folder" >&2; exit 1; }
. "$G/lib.sh"
cd "$top" || fail "repo root"
newbin
VER="agent.verify copilot
model.verify $MODEL"

# The hook, phase verify, on its own.
R=$t/hookroot
mkdir -p "$R/src" "$R/.vloop/tmp" "$t/home" "$t/tmp" "$t/stub" || fail mkdir
echo x >"$R/src/app.go"
echo r >"$R/README.md"
snap() { (cd "$t" && find hookroot home -print | LC_ALL=C sort && cat hookroot/src/app.go hookroot/README.md); }
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
VW='is not allowed in a verify session: only .vloop/tmp/ may be written'
ok verify "$(bash_ 'go build -o .vloop/tmp/v/vloop ./cmd/vloop')"
ok verify "$(bash_ './.vloop/tmp/v/vloop --version > .vloop/tmp/v/out.txt')"
no verify "$(bash_ 'cd .vloop/tmp/v && ./vloop status > out.txt')" "vloop fence: write to out.txt $VW"
no verify "$(bash_ 'echo x > README.md')" "vloop fence: write to README.md $VW"
no verify "$(bash_ 'rm .vloop/tmp/v/out.txt')" 'vloop fence: rm .vloop/tmp/v/out.txt is denied by the Copilot deny list (rm:*)'
no verify "$(bash_ 'git -C .vloop/tmp/v/ex1 status')" 'vloop fence: git -C .vloop/tmp/v/ex1 status is denied by the verify fence (Bash(git -C:*))'
no verify "$(bash_ 'vloop run')" 'vloop fence: vloop run is denied by the verify fence (Bash(vloop run:*))'
no verify "$(bash_ 'sh -c "git push origin main"')" 'vloop fence: git push origin main is denied by the verify fence (Bash(git push:*))'
no verify "$(bash_ 'env gh pr list')" 'vloop fence: gh pr list is denied by the Copilot deny list (gh:*)'
no verify "$(bash_ 'echo $(date)')" 'vloop fence: command substitution cannot be judged'
ok verify "$(file_ create .vloop/tmp/verify.json)"
ok verify "$(file_ create "$R/.vloop/tmp/v/x.txt")"
no verify "$(file_ edit internal/x.go)" "vloop fence: write to internal/x.go $VW"
no verify "$(file_ create .vloop/state/state.json)" "vloop fence: write to .vloop/state/state.json $VW"
no verify "$(file_ edit "$R/src/app.go")"
ok verify "$(file_ view src/app.go)"
no verify "$(file_ view "$t/home/.copilot/config.json")"
ok verify "$(patch_ "$(printf '*** Begin Patch\n*** Add File: .vloop/tmp/v/x.txt\n+x\n*** End Patch\n')")"
no verify "$(patch_ "$(printf '*** Begin Patch\n*** Update File: src/app.go\n@@\n-x\n+y\n*** End Patch\n')")" "vloop fence: write to src/app.go $VW"
ok verify "$(payload skill '{"skill":"verify"}')"
no verify "$(payload skill '{"skill":"vloop-verify"}')" "vloop fence: skill vloop-verify is not vloop's — use /vloop:verify"
no verify "$(payload web_fetch '{"url":"https://example.com"}')" 'vloop fence: tool web_fetch is not allowed in a verify session'
no verify "$(payload search_code_subagent '{"query":"x"}')" 'vloop fence: tool search_code_subagent is not allowed in a verify session'
for ph in review second-review; do
  no "$ph" "$(bash_ 'rm src/old.ts')" 'vloop fence: rm src/old.ts is denied by the Copilot deny list (rm:*)'
  no "$ph" "$(file_ create src/a.ts)" "vloop fence: write to src/a.ts is not allowed in a $ph session: only .vloop/tmp/ may be written"
  ok "$ph" "$(file_ create .vloop/tmp/verdict.json)"
  no "$ph" "$(payload skill '{"skill":"vloop-verify"}')" "vloop fence: skill vloop-verify is not vloop's — use /vloop:review"
done
ok work "$(bash_ 'rm src/old.ts')"
no work "$(bash_ 'rm -rf dist')" 'vloop fence: rm -rf dist is denied by the work fence (Bash(rm -rf:*))'
ok work "$(file_ create src/links/a.ts)"
for ph in plan gate-review; do
  hookrun "$ph" "$(bash_ 'ls')"
  [ "$rc" = 2 ] && [ "$dec" = deny ] || fail "fence hook --phase $ph: want the deny line and exit 2, got '$out', exit $rc"
  printf '%s' "$why" | grep -qF "unknown phase \"$ph\": want review, second-review, work or verify" ||
    fail "fence hook --phase $ph: the reason '$why' does not hold 'unknown phase \"$ph\": want review, second-review, work or verify'"
done
[ "$(snap)" = "$before" ] || fail "the hook changed the fixture root or the HOME"

# The agent keys.
CFG=; fixture
get() { k=$1; shift; (cd "$t/repo" && venv "$@" "$S" config get "$k" 2>&1); }
set_() { (cd "$t/repo" && venv "$S" config set "$1" "$2" >"$t/set.out" 2>&1); }
set_ agent.verify copilot || fail "vloop config set agent.verify copilot failed: $(cat "$t/set.out")"
[ "$(get agent.verify)" = copilot ] || fail "agent.verify set to copilot reads '$(get agent.verify)'"
set_ agent.verify claude || fail "vloop config set agent.verify claude failed: $(cat "$t/set.out")"
[ "$(get agent.verify VLOOP_AGENT_VERIFY=copilot)" = copilot ] || fail "VLOOP_AGENT_VERIFY=copilot does not set agent.verify: '$(get agent.verify VLOOP_AGENT_VERIFY=copilot)'"
set_ agent.verify codex; c=$?
[ "$c" = 2 ] || fail "vloop config set agent.verify codex exited $c, want 2"
for k in agent.plan agent.gate-review; do
  set_ "$k" copilot; c=$?
  [ "$c" = 2 ] && grep -qxF "vloop: invalid value \"copilot\" for $k: want one of claude" "$t/set.out" ||
    fail "vloop config set $k copilot: want exit 2 and 'vloop: invalid value \"copilot\" for $k: want one of claude', got exit $c: $(cat "$t/set.out")"
done
for v in VLOOP_AGENT_PLAN VLOOP_AGENT_GATE_REVIEW; do
  CFG=; fixture
  runfix "$t/p-$v" "$v=copilot"
  [ "$rc" != 0 ] || fail "vloop run with $v=copilot exited 0: it must be refused"
  nosession "$t/p-$v" "the run with $v=copilot"
done

# A run with a copilot verify session.
CFG=$VER; fixture
runfix "$t/r1"
[ "$rc" = 0 ] || fail "a run with agent.verify = copilot exited $rc: $(tail -8 "$t/r1")"
[ "$(task status T1)" = done ] || fail "T1 is $(task status T1) after the run, want done"
onecommit "the run with agent.verify = copilot"
grep -qE '^[[:space:]]*verify: no findings$' "$t/r1" || fail "the run does not print 'verify: no findings': $(tail -8 "$t/r1")"
[ "$(ev "$t/r1" '^session copilot-verify ' | grep -cxF "session copilot-verify - model=$MODEL")" = 1 ] ||
  fail "want one copilot verify session with model $MODEL: $(ev "$t/r1" '^session ')"
ev "$t/r1" '^session verify ' >/dev/null && fail "a claude verify session ran with agent.verify = copilot: $(ev "$t/r1" '^session verify ')"
ev "$t/r1" '^fence work T1 ok$' >/dev/null || fail "the claude work session did not run under the work fence: $(ev "$t/r1" '^(fence|session) ')"
ev "$t/r1" '^fence review T1 ok$' >/dev/null || fail "the claude review session did not run under the review fence: $(ev "$t/r1" '^(fence|session) ')"
for p in 'shell(rm:*)' 'shell(gh:*)' 'shell(git push:*)' 'shell(git commit:*)' 'shell(curl:*)' 'url'; do
  ev "$t/r1" "^deny-tool verify " | grep -qxF "deny-tool verify $p" || fail "the verify session's argv lost --deny-tool=$p: $(ev "$t/r1" '^deny-tool verify ')"
done
ev "$t/r1" '^hooks verify ' | grep -qxE 'hooks verify [1-9][0-9]* bash=yes powershell=yes' ||
  fail "the verify session found no preToolUse hook with both command forms in COPILOT_HOME/hooks/: $(ev "$t/r1" '^hooks ')"
ev "$t/r1" '^hookfile verify ' | grep -qxF 'hookfile verify verify' ||
  fail "the verify session's hooks file does not run 'fence hook --phase verify' alone: $(ev "$t/r1" '^hookfile ')"
ev "$t/r1" '^hook-phase verify ' | grep -qxF 'hook-phase verify verify' || fail "the verify session's hook is not run with --phase verify: $(ev "$t/r1" '^hook-phase ')"
echo "$VERIFYHOOK" | while IFS= read -r want; do
  ev "$t/r1" "^hook verify ${want%% *} " | grep -qxF "hook verify $want" ||
    fail "the verify session's hook answered '$(ev "$t/r1" "^hook verify ${want%% *} ")', want 'hook verify $want'"
done || exit 1
recs=$(phaserecs "$t/r1" verify)
[ "$(echo "$recs" | grep -c .)" = 1 ] || fail "want one verify session record, got: $recs"
f=$recs
[ "$("$JQ" -r .agent "$f")" = copilot ] && [ "$("$JQ" -r .model "$f")" = "$MODEL" ] ||
  fail "the verify session record is $("$JQ" -c '{phase, agent, model}' "$f"), want phase verify, agent copilot, model $MODEL"
[ "$("$JQ" -r .schema "$f")" = session/v4 ] || fail "the verify session record's schema is $("$JQ" -r .schema "$f"), want session/v4"
(cd "$t/repo" && venv "$S" schema validate session/v4 "$f" >"$t/val" 2>&1) || fail "the verify session record is not valid session/v4: $(cat "$t/val")"
"$JQ" -e '.credits == 8.4 and .cost_source == "estimated"' "$f" >/dev/null || fail "the verify session record's credits and cost_source are $("$JQ" -c '{credits, cost_source}' "$f"), want 8.4 and estimated"
"$JQ" -e '(.permission_denials | length) == 1 and .permission_denials[0].tool_name == "bash" and (.permission_denials[0].message | contains("rm .vloop/tmp/v/out.txt is denied by the Copilot deny list (rm:*)"))' "$f" >/dev/null ||
  fail "the verify session record's permission_denials are $("$JQ" -c .permission_denials "$f"), want the one bash denial"
[ -f "$t/r1.report" ] || fail "the run folder holds no reports/verify.json"
(cd "$t/repo" && venv "$S" schema validate verify/v1 "$t/r1.report" >"$t/val" 2>&1) || fail "reports/verify.json is not valid verify/v1: $(cat "$t/val")"
"$JQ" -e '.notes == "copilot" and (.findings | length) == 0' "$t/r1.report" >/dev/null || fail "reports/verify.json is not the copilot session's report: $(cat "$t/r1.report")"

# Copilot work, review and verify in one run: each keeps its argv and phase.
CFG="$VER
agent.work copilot
model.work $MODEL
agent.review copilot
model.review $MODEL"; fixture
runfix "$t/r2"
[ "$rc" = 0 ] || fail "a run with copilot work, review and verify exited $rc: $(tail -8 "$t/r2")"
[ "$(task status T1)" = done ] || fail "copilot work, review and verify: T1 is $(task status T1), want done"
for r in work review verify; do
  ev "$t/r2" "^session copilot-$r " >/dev/null || fail "no copilot $r session: $(ev "$t/r2" '^session ')"
  ev "$t/r2" "^hookfile $r " | grep -qxF "hookfile $r $r" || fail "the copilot $r session's hooks file is not its phase's: $(ev "$t/r2" '^hookfile ')"
done
ev "$t/r2" '^deny-tool review ' | sed 's/^deny-tool review //' | LC_ALL=C sort >"$t/dt.review"
ev "$t/r2" '^deny-tool verify ' | sed 's/^deny-tool verify //' | LC_ALL=C sort >"$t/dt.verify"
ev "$t/r2" '^deny-tool work ' | sed 's/^deny-tool work //' | LC_ALL=C sort >"$t/dt.work"
[ -s "$t/dt.review" ] && cmp -s "$t/dt.review" "$t/dt.verify" || fail "the verify session's --deny-tool list differs from the review session's: review '$(tr '\n' ' ' <"$t/dt.review")', verify '$(tr '\n' ' ' <"$t/dt.verify")'"
grep -qxF 'shell(rm:*)' "$t/dt.work" && fail "the copilot work session's argv holds --deny-tool=shell(rm:*)"
grep -vxF 'shell(rm:*)' "$t/dt.review" | cmp -s - "$t/dt.work" || fail "the copilot work session's --deny-tool list is not the review list without shell(rm:*)"
for want in 'shgh deny 0' 'diff allow 0' 'rmfile deny 0' 'srcrel deny 0' 'tmprel allow 0'; do
  ev "$t/r2" "^hook review ${want%% *} " | grep -qxF "hook review $want" ||
    fail "the review session's hook answered '$(ev "$t/r2" "^hook review ${want%% *} ")', want 'hook review $want'"
done
for want in 'rmfile allow 0' 'rmrf deny 0' 'srcrel allow 0' 'stateabs deny 0'; do
  ev "$t/r2" "^hook work ${want%% *} " | grep -qxF "hook work $want" ||
    fail "the work session's hook answered '$(ev "$t/r2" "^hook work ${want%% *} ")', want 'hook work $want'"
done
[ -f "$t/r2.report" ] || fail "copilot work, review and verify: the run folder holds no reports/verify.json"

# Claude verify sessions: unchanged, under the verify fence.
CFG=; fixture
runfix "$t/r3"
[ "$rc" = 0 ] || fail "a run with claude verify exited $rc: $(tail -8 "$t/r3")"
ev "$t/r3" '^session verify - model=sonnet$' >/dev/null || fail "no claude verify session with model sonnet: $(ev "$t/r3" '^session ')"
ev "$t/r3" '^fence verify - ok$' >/dev/null || fail "the claude verify session did not run under the verify fence: $(ev "$t/r3" '^fence ')"
ev "$t/r3" '^session copilot-' >/dev/null && fail "a copilot session ran with every agent on claude"
"$JQ" -e '.notes == "claude"' "$t/r3.report" >/dev/null 2>&1 || fail "the claude verify session's report was not kept: $(cat "$t/r3.report" 2>&1)"

# run.verify = false: no verify session.
CFG="$VER
run.verify false"; fixture
runfix "$t/r4"
[ "$rc" = 0 ] || fail "run.verify = false with agent.verify = copilot exited $rc: $(tail -8 "$t/r4")"
[ "$(task status T1)" = done ] || fail "run.verify = false: T1 is $(task status T1), want done"
ev "$t/r4" 'verify' >/dev/null && fail "run.verify = false: a verify session ran: $(ev "$t/r4" 'verify')"
[ -f "$t/r4.report" ] && fail "run.verify = false: the run folder holds reports/verify.json"

# A finding is reported, not a halt.
CFG=$VER; fixture
: >"$t/stub/copilot.verify-finding"
runfix "$t/r5"
[ "$rc" = 0 ] || fail "a copilot verify session with a finding: vloop run exited $rc, want 0: $(tail -8 "$t/r5")"
[ "$(task status T1)" = done ] || fail "a verify finding: T1 is $(task status T1), want done"
grep -qE 'verify: 1 finding\(s\) — see .*reports/verify\.json' "$t/r5" || fail "the run does not report the one finding: $(tail -8 "$t/r5")"
"$JQ" -e '(.findings | length) == 1' "$t/r5.report" >/dev/null 2>&1 || fail "the finding's report was not kept: $(cat "$t/r5.report" 2>&1)"

# No report, then a failed session: a warning, never a halt.
CFG=$VER; fixture
: >"$t/stub/copilot.verify-noreport"
runfix "$t/r6"
[ "$rc" = 0 ] || fail "a copilot verify session with no report: vloop run exited $rc, want 0: $(tail -8 "$t/r6")"
grep -qF 'vloop: the verify session left no valid .vloop/tmp/verify.json — verify by hand' "$t/r6" || fail "no warning for the missing report: $(tail -8 "$t/r6")"
[ -f "$t/r6.report" ] && fail "a verify session with no report: the run folder holds reports/verify.json"
CFG=$VER; fixture
: >"$t/stub/copilot.verify-fail"
runfix "$t/r7"
[ "$rc" = 0 ] || fail "a failed copilot verify session: vloop run exited $rc, want 0: $(tail -8 "$t/r7")"
[ "$(task status T1)" = done ] || fail "a failed verify session: T1 is $(task status T1), want done"
grep -qF 'vloop: the verify session failed to run — verify by hand' "$t/r7" || fail "no warning for the failed verify session: $(tail -8 "$t/r7")"

# A copilot verify session that writes README.md: reverted, the report kept.
CFG=$VER; fixture
: >"$t/stub/copilot.verify-readme"
runfix "$t/r8"
[ "$rc" = 0 ] || fail "a copilot verify session that wrote README.md: vloop run exited $rc, want 0: $(tail -8 "$t/r8")"
grep -qF 'vloop: the verify session changed README.md — reverted' "$t/r8" || fail "no 'vloop: the verify session changed README.md — reverted' line: $(tail -8 "$t/r8")"
[ "$(cat "$t/repo/README.md")" = "the fixture's readme" ] || fail "README.md was not reverted: $(cat "$t/repo/README.md")"
[ -f "$t/r8.report" ] || fail "a verify session that wrote README.md: reports/verify.json was not kept"

# A copilot verify session that moves HEAD: the run halts, exit 9.
CFG=$VER; fixture
: >"$t/stub/copilot.verify-moveref"
runfix "$t/r9"
[ "$rc" = 9 ] || fail "a copilot verify session that moved HEAD: vloop run exited $rc, want 9: $(tail -8 "$t/r9")"
grep -qF 'REFS MOVED verify — a verify session changed git refs; nothing was committed' "$t/r9" ||
  fail "no 'REFS MOVED verify — a verify session changed git refs; nothing was committed': $(tail -8 "$t/r9")"
echo "gate T2: ok"
