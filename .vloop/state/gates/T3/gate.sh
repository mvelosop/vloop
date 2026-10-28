#!/bin/sh
# Gate T3 — W2, the run preflight and doctor for agent.work = copilot. The
# experimental warning in `vloop run` and in doctor; the refusal of the default
# model.work before any session; the copilot checks Copilot review has —
# copilot on PATH and runnable, a GitHub login reaching its sessions, K1's
# refusal of a project skill named like one of vloop's — each stopping the run
# (exit 1, no session) and each a doctor row; and none of it with agent.work =
# claude. Fixture runs with the stub claude and the stub copilot.
G=$(cd "$(dirname "$0")" && pwd -P) || { echo "GATE FAIL: no gate folder" >&2; exit 1; }
. "$G/lib.sh"
cd "$top" || fail "repo root"
newbin
WORK="agent.work copilot
model.work $MODEL"
WARN='! agent.work is copilot: experimental (docs/guide/configuration.md)'
doc() { o=$1; shift; (cd "$t/repo" && venv "$@" "$S" doctor >"$o" 2>&1); rc=$?; }
docnl() { o=$1; shift; (cd "$t/repo" && venvnl "$@" "$S" doctor >"$o" 2>&1); rc=$?; }
stripped() { sed 's/^[[:space:]]*//' "$1"; }

# The warning, in the run and in doctor; the run goes ahead.
CFG=$WORK; fixture
doc "$t/d1"
[ "$rc" = 0 ] || fail "doctor with agent.work = copilot and a model exited $rc: $(cat "$t/d1")"
grep -qxF "$WARN" "$t/d1" || fail "doctor does not print '$WARN': $(cat "$t/d1")"
grep '^✓' "$t/d1" | grep -qw copilot || fail "doctor with agent.work = copilot has no passing copilot row: $(cat "$t/d1")"
runfix "$t/r1"
[ "$rc" = 0 ] || fail "a run with agent.work = copilot and a model exited $rc: $(tail -8 "$t/r1")"
stripped "$t/r1" | grep -qxF "$WARN" || fail "vloop run does not print '  $WARN': $(head -15 "$t/r1")"
ev "$t/r1" "^session copilot-work T1 model=$MODEL\$" >/dev/null || fail "the run's work session was not the copilot one: $(ev "$t/r1" '^session ')"

# model.work left at its default: refused before any session.
CFG="agent.work copilot"; fixture
runfix "$t/r2"
[ "$rc" = 1 ] || fail "agent.work = copilot with the default model.work: vloop run exited $rc, want 1: $(tail -5 "$t/r2")"
grep -qxF 'vloop: agent.work is copilot but model.work is the default "sonnet", a Claude alias: set model.work to a Copilot model id' "$t/r2" ||
  fail "the refusal is not the brief's line: $(cat "$t/r2")"
nosession "$t/r2" "the refused run (default model.work)"
CFG=; fixture
runfix "$t/r2e" VLOOP_AGENT_WORK=copilot
[ "$rc" = 1 ] || fail "VLOOP_AGENT_WORK=copilot with the default model.work: vloop run exited $rc, want 1: $(tail -5 "$t/r2e")"
grep -qF 'agent.work is copilot but model.work is the default "sonnet"' "$t/r2e" || fail "VLOOP_AGENT_WORK=copilot: no model.work refusal: $(cat "$t/r2e")"
nosession "$t/r2e" "the refused run (VLOOP_AGENT_WORK)"

# K1: a project skill named like vloop's work skill.
CFG=$WORK; fixture
skill .claude/skills/work; commitfix
runfix "$t/k1"
[ "$rc" = 1 ] || fail "agent.work = copilot with .claude/skills/work: vloop run exited $rc, want 1: $(tail -5 "$t/k1")"
grep -qxF "vloop: .claude/skills/work has the name of vloop's work skill; a copilot session may load it instead — rename it" "$t/k1" ||
  fail "vloop run does not refuse .claude/skills/work with K1's message: $(cat "$t/k1")"
nosession "$t/k1" "the refused run (.claude/skills/work)"
doc "$t/dk1"
[ "$rc" = 1 ] || fail "doctor with .claude/skills/work and agent.work = copilot exited $rc, want 1: $(cat "$t/dk1")"
grep '^✗' "$t/dk1" | grep -qF ".claude/skills/work has the name of vloop's work skill" || fail "doctor has no error row for .claude/skills/work: $(cat "$t/dk1")"
CFG=$WORK; fixture
skill .github/skills/review; commitfix
runfix "$t/k2"
[ "$rc" = 1 ] || fail "agent.work = copilot with .github/skills/review: vloop run exited $rc, want 1: $(tail -5 "$t/k2")"
grep -qxF "vloop: .github/skills/review has the name of vloop's review skill; a copilot session may load it instead — rename it" "$t/k2" ||
  fail "vloop run does not refuse .github/skills/review with K1's message: $(cat "$t/k2")"
nosession "$t/k2" "the refused run (.github/skills/review)"
# With claude work sessions, the same skill stops nothing and nothing warns.
CFG=; fixture
skill .claude/skills/work; commitfix
runfix "$t/k3"
[ "$rc" = 0 ] || fail "with agent.work = claude, .claude/skills/work stopped the run (exit $rc): $(tail -5 "$t/k3")"
grep -qF "has the name of vloop's" "$t/k3" && fail "with agent.work = claude the run refuses a project skill: $(cat "$t/k3")"
grep -qF 'agent.work is copilot' "$t/k3" && fail "with agent.work = claude the run warns of copilot"
doc "$t/dk3"
[ "$rc" = 0 ] || fail "doctor with agent.work = claude exited $rc: $(cat "$t/dk3")"
grep -qF 'agent.work' "$t/dk3" && fail "doctor with agent.work = claude speaks of agent.work: $(grep -F agent.work "$t/dk3")"

# No GitHub login reaching the sessions.
CFG=$WORK; fixture
runfixnl "$t/l1"
[ "$rc" = 1 ] || fail "agent.work = copilot with no GitHub login: vloop run exited $rc, want 1: $(tail -5 "$t/l1")"
grep -qF 'vloop: agent.work is copilot but no GitHub login reaches its sessions' "$t/l1" ||
  fail "the run does not refuse the missing login naming agent.work: $(cat "$t/l1")"
nosession "$t/l1" "the refused run (no login)"
docnl "$t/dl1"
[ "$rc" = 1 ] || fail "doctor with agent.work = copilot and no GitHub login exited $rc, want 1: $(cat "$t/dl1")"
grep '^✗' "$t/dl1" | grep -qF 'copilot login' || fail "doctor has no error row for the copilot login: $(cat "$t/dl1")"

# copilot missing, then not runnable.
CFG=$WORK; fixture
rm -f "$t/stub/copilot"
runfix "$t/c1"
[ "$rc" = 1 ] || fail "agent.work = copilot with no copilot on PATH: vloop run exited $rc, want 1: $(tail -5 "$t/c1")"
grep -q copilot "$t/c1" || fail "the refused run names no copilot problem: $(cat "$t/c1")"
nosession "$t/c1" "the refused run (no copilot)"
doc "$t/dc1"
[ "$rc" = 1 ] || fail "doctor with agent.work = copilot and no copilot on PATH exited $rc, want 1: $(cat "$t/dc1")"
grep '^✗' "$t/dc1" | grep -q copilot || fail "doctor names no copilot problem: $(cat "$t/dc1")"
printf '#!/bin/sh\nexit 3\n' >"$t/stub/copilot" && chmod +x "$t/stub/copilot"
doc "$t/dc2"
[ "$rc" = 1 ] || fail "doctor with a copilot whose --version fails exited $rc, want 1: $(cat "$t/dc2")"
grep '^✗' "$t/dc2" | grep -qF 'copilot --version failed' || fail "doctor does not report copilot --version failing: $(cat "$t/dc2")"
runfix "$t/c2"
[ "$rc" = 1 ] || fail "agent.work = copilot with a copilot whose --version fails: vloop run exited $rc, want 1: $(tail -5 "$t/c2")"
echo "gate T3: ok"
