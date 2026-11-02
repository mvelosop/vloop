#!/bin/sh
# Gate T4 — R3 on Copilot: `vloop brief review` with agent.review-brief =
# copilot, the stub copilot (stub-copilot.sh) in fixture repositories
# (lib.sh). The worked example's Copilot run — the experimental warning, the
# reviewing line, the verdict block, exit 1, a session/v5 record with agent
# copilot, the deny list with shell(rm:*), a hooks file running only
# `fence hook --phase review-brief`, whose hook answers the review-brief
# decisions — and its preflight: the default-model refusal, the login
# refusal, a skill clash, copilot missing. vloop doctor lists the key with
# the other Copilot keys; vloop run's preflight ignores it; with
# agent.review-brief = claude a clash stops nothing.
G=$(cd "$(dirname "$0")" && pwd -P) || { echo "GATE FAIL: no gate folder" >&2; exit 1; }
. "$G/lib.sh"
cd "$top" || fail "repo root"
newbin
CP="agent.review-brief copilot
model.review-brief $MODEL"
WARN='! agent.review-brief is copilot: experimental (docs/guide/configuration.md)'

# The worked example's Copilot run.
CFG=$CP; fixture
report "$REPORT_FIX"
rev "$t/cp"
[ "$rc" = 1 ] || fail "a copilot FIX report: exit $rc, want 1: $(cat "$t/cp.err")"
folder "$t/cp" "a copilot FIX report"
errline "$t/cp" "$WARN" || fail "a copilot review: no '  $WARN' on stderr: $(cat "$t/cp.err")"
errline "$t/cp" "reviewing $B using $MODEL (agent copilot)" || fail "a copilot review: no 'reviewing $B using $MODEL (agent copilot)': $(cat "$t/cp.err")"
{ block FIX "$t/stub/report.json"; echo "kept: $F/brief-review.json"; } >"$t/cp.want"
cmp -s "$t/cp.want" "$t/cp.out" || fail "a copilot FIX report: stdout is not the verdict block:
--- want
$(cat "$t/cp.want")
--- got
$(cat "$t/cp.out")"
record "a copilot review" copilot "$MODEL"
"$JQ" -e '.credits == 8.4 and .cost_source == "estimated" and (.permission_denials | length) == 1' "$t/repo/$F/sessions/001-review-brief.json" >/dev/null ||
  fail "a copilot review: the record's credits, cost_source and denials are $("$JQ" -c '{credits, cost_source, permission_denials}' "$t/repo/$F/sessions/001-review-brief.json")"
[ "$(ev "$t/cp" '^session ')" = "session copilot-review-brief $B model=$MODEL" ] || fail "a copilot review: want one copilot session of /vloop:review-brief $B with model $MODEL, got: $(ev "$t/cp" '^session ')"
for p in 'shell(rm:*)' 'shell(gh:*)' 'shell(git push:*)' 'shell(git commit:*)' 'shell(curl:*)' 'url'; do
  ev "$t/cp" '^deny-tool review-brief ' | grep -qxF "deny-tool review-brief $p" ||
    fail "a copilot review: the argv lost --deny-tool=$p: $(ev "$t/cp" '^deny-tool ')"
done
ev "$t/cp" '^hooks review-brief ' | grep -qxE 'hooks review-brief [1-9][0-9]* bash=yes powershell=yes' ||
  fail "a copilot review: no preToolUse hook with both command forms in COPILOT_HOME/hooks/: $(ev "$t/cp" '^hooks ')"
[ "$(ev "$t/cp" '^hookfile ')" = "hookfile review-brief review-brief" ] || fail "a copilot review: the hooks file does not run 'fence hook --phase review-brief' alone: $(ev "$t/cp" '^hookfile ')"
for want in 'check allow 0' 'cat allow 0' 'report allow 0' 'briefw deny 0' 'commit deny 0' 'rmbrief deny 0' 'skillrb allow 0' 'skillother deny 0' 'nested deny 0' 'webfetch deny 0'; do
  ev "$t/cp" "^hook review-brief ${want%% *} " | grep -qxF "hook review-brief $want" ||
    fail "a copilot review: its hook answered '$(ev "$t/cp" "^hook review-brief ${want%% *} ")', want 'hook review-brief $want'"
done
refsame "$t/cp" "a copilot review"
briefsame "a copilot review"

# Refusals, in the preflight's order, each with nothing run.
# refused <label> <line> [revnl]
refused() {
  [ "$rc" = 1 ] || fail "$1: exit $rc, want 1: $(cat "$t/$1.err")"
  endswith "$t/$1" "$2" "$1"
  noreview "$t/$1" "$1"
}
CFG="agent.review-brief copilot"; fixture
report "$REPORT_FIX"
rev "$t/defmodel"
refused defmodel 'vloop: agent.review-brief is copilot but model.review-brief is the default "opus", a Claude alias: set model.review-brief to a Copilot model id'
CFG=$CP; fixture
report "$REPORT_FIX"
revnl "$t/nologin"
refused nologin 'vloop: agent.review-brief is copilot but no GitHub login reaches its sessions: set COPILOT_GITHUB_TOKEN, GH_TOKEN or GITHUB_TOKEN, or run gh auth login'
CFG=$CP; fixture
report "$REPORT_FIX"
skill .claude/skills/review-brief
revnl "$t/clash"
refused clash "vloop: .claude/skills/review-brief has the name of vloop's review-brief skill; a copilot session may load it instead — rename it"
CFG=$CP; fixture
report "$REPORT_FIX"
skill .github/skills/plan
rev "$t/clash2"
refused clash2 "vloop: .github/skills/plan has the name of vloop's plan skill; a copilot session may load it instead — rename it"
CFG=$CP; fixture
report "$REPORT_FIX"
rm -f "$t/stub/copilot"
rev "$t/nocp"
[ "$rc" = 1 ] || fail "copilot missing: exit $rc, want 1: $(cat "$t/nocp.err")"
grep -qE '^  ✗ copilot [^l]' "$t/nocp.err" || fail "copilot missing: want a '  ✗ copilot …' line on stderr: $(cat "$t/nocp.err")"
grep -qxF 'vloop: preflight failed — 1 problem(s), nothing has run' "$t/nocp.err" || fail "copilot missing: want 'vloop: preflight failed — 1 problem(s), nothing has run': $(cat "$t/nocp.err")"
noreview "$t/nocp" "copilot missing"
# With agent.review-brief = claude, a project skill named review-brief stops nothing.
CFG=; fixture
report "$REPORT_PASS"
skill .claude/skills/review-brief
venv git -C "$t/repo" add -A && venv git -C "$t/repo" commit -q -m skill || fail "fixture commit"
revnl "$t/clcl"
[ "$rc" = 0 ] || fail "agent.review-brief claude with .claude/skills/review-brief: exit $rc, want 0: $(cat "$t/clcl.err")"
grep -qF 'copilot' "$t/clcl.err" && fail "agent.review-brief claude: stderr names copilot: $(cat "$t/clcl.err")"
[ "$(ev "$t/clcl" '^session ')" = "session review-brief $B model=opus effort=-" ] || fail "agent.review-brief claude: want the claude session, got: $(ev "$t/clcl" '^session ')"

# vloop doctor lists the key with the other Copilot keys.
CFG=$CP; fixture
skill .claude/skills/review-brief
(cd "$t/repo" && venv "$S" doctor >"$t/doc" 2>&1)
grep -qxF "$WARN" "$t/doc" || fail "doctor: no '$WARN' row: $(cat "$t/doc")"
grep -qxF '✓ copilot' "$t/doc" || fail "doctor: no '✓ copilot' row: $(cat "$t/doc")"
grep -qxF '✓ copilot login' "$t/doc" || fail "doctor: no '✓ copilot login' row: $(cat "$t/doc")"
grep -qxF "✗ .claude/skills/review-brief has the name of vloop's review-brief skill; a copilot session may load it instead — rename it" "$t/doc" ||
  fail "doctor: no clash row for .claude/skills/review-brief: $(cat "$t/doc")"
(cd "$t/repo" && venvnl "$S" doctor >"$t/doc" 2>&1)
grep -q '^✗ copilot login ' "$t/doc" || fail "doctor with no login: no '✗ copilot login …' row: $(cat "$t/doc")"

# vloop run's preflight ignores agent.review-brief: no warning, no refusal —
# the run of a ready brief (runbrief.md) reaches its stub plan session,
# which exits 3, as at the base.
CFG="agent.review-brief copilot"; fixture
skill .claude/skills/review-brief
rm -f "$t/stub/copilot"
RB=docs/briefs/B20260101-0900-demo.loop-brief.md
cp "$G/runbrief.md" "$t/repo/$RB" && echo x >"$t/repo/docs/x.md" || fail "run brief"
venv git -C "$t/repo" add -A && venv git -C "$t/repo" commit -q -m skill || fail "fixture commit"
(cd "$t/repo" && venvnl "$S" run "$RB" >"$t/run" 2>&1); c=$?
[ "$c" = 1 ] || fail "vloop run with only agent.review-brief = copilot: exit $c, want 1 from the stub plan session: $(tail -8 "$t/run")"
grep -qF '(claude exited 3)' "$t/run" || fail "vloop run with only agent.review-brief = copilot did not reach the plan session: $(tail -8 "$t/run")"
grep -qiF 'copilot' "$t/run" && fail "vloop run with only agent.review-brief = copilot names copilot: $(grep -iF copilot "$t/run")"
grep -qF 'review-brief' "$t/run" && fail "vloop run names review-brief: $(grep -F review-brief "$t/run")"
echo "gate T4: ok"
