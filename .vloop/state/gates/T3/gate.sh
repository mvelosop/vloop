#!/bin/sh
# Gate T3 — P2, the run preflight and doctor for agent.plan = copilot. The
# experimental warning in `vloop run` and in doctor; the refusal of the
# default model.plan before any session (from the file and from
# VLOOP_AGENT_PLAN); the copilot checks the other Copilot phases have —
# copilot on PATH and runnable, a GitHub login reaching its sessions, K1's
# refusal of a project skill named like one of vloop's — each stopping the run
# (exit 1, no session) and each a doctor row; the key in use also on a resume
# that will not plan; none of it with agent.plan = claude; agent.work's and
# agent.gate-review's checks keep their words. Fixture runs with the stub
# claude and the stub copilot.
G=$(cd "$(dirname "$0")" && pwd -P) || { echo "GATE FAIL: no gate folder" >&2; exit 1; }
. "$G/lib.sh"
cd "$top" || fail "repo root"
newbin
PLC="agent.plan copilot
model.plan $MODEL"
WARN='! agent.plan is copilot: experimental (docs/guide/configuration.md)'
REFUSE='vloop: agent.plan is copilot but model.plan is the default "opus", a Claude alias: set model.plan to a Copilot model id'
doc() { o=$1; shift; (cd "$t/repo" && venv "$@" "$S" doctor >"$o" 2>&1); rc=$?; }
docnl() { o=$1; shift; (cd "$t/repo" && venvnl "$@" "$S" doctor >"$o" 2>&1); rc=$?; }

# The warning, in the run and in doctor; the run goes ahead.
CFG=$PLC; fixture
doc "$t/d1"
[ "$rc" = 0 ] || fail "doctor with agent.plan = copilot and a model exited $rc: $(cat "$t/d1")"
grep -qxF "$WARN" "$t/d1" || fail "doctor does not print '$WARN': $(cat "$t/d1")"
grep '^✓' "$t/d1" | grep -qw copilot || fail "doctor with agent.plan = copilot has no passing copilot row: $(cat "$t/d1")"
grep '^✓' "$t/d1" | grep -qF 'copilot login' || fail "doctor with agent.plan = copilot has no passing copilot login row: $(cat "$t/d1")"
runfix "$t/r1"
[ "$rc" = 0 ] || fail "a run with agent.plan = copilot and a model exited $rc: $(tail -8 "$t/r1")"
grep -qxF "  $WARN" "$t/r1" || fail "vloop run does not print '  $WARN': $(head -15 "$t/r1")"
[ "$(task status T1)" = done ] || fail "the warned run: T1 is $(task status T1), want done"
onecommit "the warned run"
ev "$t/r1" "^session copilot-plan $BRIEF model=$MODEL\$" >/dev/null || fail "the run's plan session was not the copilot one: $(ev "$t/r1" '^session ')"

# model.plan left at its default: refused before any session.
CFG="agent.plan copilot"; fixture
runfix "$t/m1"
[ "$rc" = 1 ] || fail "agent.plan = copilot with the default model.plan: vloop run exited $rc, want 1: $(tail -5 "$t/m1")"
grep -qxF "$REFUSE" "$t/m1" || fail "the refusal is not the brief's line '$REFUSE': $(cat "$t/m1")"
nosession "$t/m1" "the refused run (default model.plan)"
CFG=; fixture
runfix "$t/m2" VLOOP_AGENT_PLAN=copilot
[ "$rc" = 1 ] || fail "VLOOP_AGENT_PLAN=copilot with the default model.plan: vloop run exited $rc, want 1: $(tail -5 "$t/m2")"
grep -qxF "$REFUSE" "$t/m2" || fail "VLOOP_AGENT_PLAN=copilot: no model.plan refusal: $(cat "$t/m2")"
nosession "$t/m2" "the refused run (VLOOP_AGENT_PLAN)"

# A resume that will not plan: the key is still in use.
CFG=; fixture
(cd "$t/repo" && venv "$S" run --plan-only "$BRIEF" >"$t/po" 2>&1) || fail "vloop run --plan-only with every agent on claude failed: $(tail -5 "$t/po")"
rm -f "$t/stub/events"
[ "$(task status T1)" = pending ] || fail "after --plan-only, T1 is $(task status T1), want pending"
runfix "$t/z1" VLOOP_AGENT_PLAN=copilot
[ "$rc" = 1 ] || fail "a resume with VLOOP_AGENT_PLAN=copilot and the default model.plan exited $rc, want 1: $(tail -5 "$t/z1")"
grep -qxF "$REFUSE" "$t/z1" || fail "a resume with VLOOP_AGENT_PLAN=copilot: no model.plan refusal: $(cat "$t/z1")"
nosession "$t/z1" "the refused resume"
runfix "$t/z2" VLOOP_AGENT_PLAN=copilot VLOOP_MODEL_PLAN="$MODEL"
[ "$rc" = 0 ] || fail "a resume with agent.plan = copilot and a model exited $rc: $(tail -8 "$t/z2")"
grep -qxF "  $WARN" "$t/z2" || fail "a resume with agent.plan = copilot does not print '  $WARN': $(head -15 "$t/z2")"
ev "$t/z2" '^session (plan|copilot-plan) ' >/dev/null && fail "the resume planned again: $(ev "$t/z2" '^session ')"
[ "$(task status T1)" = done ] || fail "the resume: T1 is $(task status T1), want done"

# K1: a project skill named like vloop's plan skill, or another of vloop's.
CFG=$PLC; fixture
skill .claude/skills/plan; commitfix
runfix "$t/k1"
[ "$rc" = 1 ] || fail "agent.plan = copilot with .claude/skills/plan: vloop run exited $rc, want 1: $(tail -5 "$t/k1")"
grep -qxF "vloop: .claude/skills/plan has the name of vloop's plan skill; a copilot session may load it instead — rename it" "$t/k1" ||
  fail "vloop run does not refuse .claude/skills/plan with K1's message: $(cat "$t/k1")"
nosession "$t/k1" "the refused run (.claude/skills/plan)"
doc "$t/dk1"
[ "$rc" = 1 ] || fail "doctor with .claude/skills/plan and agent.plan = copilot exited $rc, want 1: $(cat "$t/dk1")"
grep '^✗' "$t/dk1" | grep -qF ".claude/skills/plan has the name of vloop's plan skill" || fail "doctor has no error row for .claude/skills/plan: $(cat "$t/dk1")"
CFG=$PLC; fixture
skill .github/skills/work; commitfix
runfix "$t/k2"
[ "$rc" = 1 ] || fail "agent.plan = copilot with .github/skills/work: vloop run exited $rc, want 1: $(tail -5 "$t/k2")"
grep -qxF "vloop: .github/skills/work has the name of vloop's work skill; a copilot session may load it instead — rename it" "$t/k2" ||
  fail "vloop run does not refuse .github/skills/work with K1's message: $(cat "$t/k2")"
nosession "$t/k2" "the refused run (.github/skills/work)"

# agent.plan = claude: none of it, and a project skill named plan stops nothing.
CFG=; fixture
skill .claude/skills/plan; commitfix
runfix "$t/c0"
[ "$rc" = 0 ] || fail "with agent.plan = claude, .claude/skills/plan stopped the run (exit $rc): $(tail -5 "$t/c0")"
grep -qF "has the name of vloop's" "$t/c0" && fail "with agent.plan = claude the run refuses a project skill: $(cat "$t/c0")"
grep -qF 'agent.plan' "$t/c0" && fail "with agent.plan = claude the run speaks of agent.plan: $(grep -F agent.plan "$t/c0")"
doc "$t/dc0"
[ "$rc" = 0 ] || fail "doctor with agent.plan = claude exited $rc: $(cat "$t/dc0")"
grep -qF 'agent.plan' "$t/dc0" && fail "doctor with agent.plan = claude speaks of agent.plan: $(grep -F agent.plan "$t/dc0")"
grep -qF 'copilot login' "$t/dc0" && fail "doctor with every agent on claude has a copilot login row: $(cat "$t/dc0")"

# No GitHub login reaching the sessions.
CFG=$PLC; fixture
runfixnl "$t/l1"
[ "$rc" = 1 ] || fail "agent.plan = copilot with no GitHub login: vloop run exited $rc, want 1: $(tail -5 "$t/l1")"
grep -qF 'vloop: agent.plan is copilot but no GitHub login reaches its sessions' "$t/l1" ||
  fail "the run does not refuse the missing login naming agent.plan: $(cat "$t/l1")"
nosession "$t/l1" "the refused run (no login)"
docnl "$t/dl1"
[ "$rc" = 1 ] || fail "doctor with agent.plan = copilot and no GitHub login exited $rc, want 1: $(cat "$t/dl1")"
grep '^✗' "$t/dl1" | grep -qF 'copilot login' || fail "doctor has no error row for the copilot login: $(cat "$t/dl1")"

# copilot missing, then not runnable.
CFG=$PLC; fixture
rm -f "$t/stub/copilot"
runfix "$t/c1"
[ "$rc" = 1 ] || fail "agent.plan = copilot with no copilot on PATH: vloop run exited $rc, want 1: $(tail -5 "$t/c1")"
grep -q copilot "$t/c1" || fail "the refused run names no copilot problem: $(cat "$t/c1")"
nosession "$t/c1" "the refused run (no copilot)"
doc "$t/dc1"
[ "$rc" = 1 ] || fail "doctor with agent.plan = copilot and no copilot on PATH exited $rc, want 1: $(cat "$t/dc1")"
grep '^✗' "$t/dc1" | grep -q copilot || fail "doctor names no copilot problem: $(cat "$t/dc1")"
printf '#!/bin/sh\nexit 3\n' >"$t/stub/copilot" && chmod +x "$t/stub/copilot"
doc "$t/dc2"
[ "$rc" = 1 ] || fail "doctor with a copilot whose --version fails exited $rc, want 1: $(cat "$t/dc2")"
grep '^✗' "$t/dc2" | grep -qF 'copilot --version failed' || fail "doctor does not report copilot --version failing: $(cat "$t/dc2")"
runfix "$t/c2"
[ "$rc" = 1 ] || fail "agent.plan = copilot with a copilot whose --version fails: vloop run exited $rc, want 1: $(tail -5 "$t/c2")"
nosession "$t/c2" "the refused run (copilot --version fails)"

# agent.work's and agent.gate-review's checks keep their words.
CFG="agent.work copilot"; fixture
runfix "$t/w1"
[ "$rc" = 1 ] || fail "agent.work = copilot with the default model.work: vloop run exited $rc, want 1"
grep -qxF 'vloop: agent.work is copilot but model.work is the default "sonnet", a Claude alias: set model.work to a Copilot model id' "$t/w1" ||
  fail "agent.work's refusal changed: $(cat "$t/w1")"
CFG="agent.work copilot
model.work $MODEL"; fixture
doc "$t/w2"
grep -qxF '! agent.work is copilot: experimental (docs/guide/configuration.md)' "$t/w2" || fail "agent.work's doctor warning changed: $(cat "$t/w2")"
grep -qF 'agent.plan' "$t/w2" && fail "doctor with only agent.work = copilot speaks of agent.plan: $(grep -F agent.plan "$t/w2")"
CFG="agent.gate-review copilot"; fixture
runfix "$t/g1"
[ "$rc" = 1 ] && grep -qxF 'vloop: agent.gate-review is copilot but model.gate-review is the default "sonnet", a Claude alias: set model.gate-review to a Copilot model id' "$t/g1" ||
  fail "agent.gate-review's refusal changed (exit $rc): $(cat "$t/g1")"
echo "gate T3: ok"
