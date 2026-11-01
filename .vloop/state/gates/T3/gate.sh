#!/bin/sh
# Gate T3 — G2, the run preflight and doctor for agent.gate-review = copilot.
# The experimental warning in `vloop run` and in doctor; the refusal of the
# default model.gate-review before any session (from the file and from
# VLOOP_AGENT_GATE_REVIEW); the copilot checks the other Copilot phases have —
# copilot on PATH and runnable, a GitHub login reaching its sessions, K1's
# refusal of a project skill named like one of vloop's — each stopping the run
# (exit 1, no session) and each a doctor row; the `review` look-alike warning
# with its base condition and words, now shown with only the gate review on
# copilot; the key in use whatever run.verify says; none of it with
# agent.gate-review = claude; agent.work's and agent.verify's checks keep
# their words. Fixture runs with the stub claude and the stub copilot.
G=$(cd "$(dirname "$0")" && pwd -P) || { echo "GATE FAIL: no gate folder" >&2; exit 1; }
. "$G/lib.sh"
cd "$top" || fail "repo root"
newbin
GRC="agent.gate-review copilot
model.gate-review $MODEL"
WARN='! agent.gate-review is copilot: experimental (docs/guide/configuration.md)'
REFUSE='vloop: agent.gate-review is copilot but model.gate-review is the default "sonnet", a Claude alias: set model.gate-review to a Copilot model id'
doc() { o=$1; shift; (cd "$t/repo" && venv "$@" "$S" doctor >"$o" 2>&1); rc=$?; }
docnl() { o=$1; shift; (cd "$t/repo" && venvnl "$@" "$S" doctor >"$o" 2>&1); rc=$?; }

# The warning, in the run and in doctor; the run goes ahead.
CFG=$GRC; fixture
doc "$t/d1"
[ "$rc" = 0 ] || fail "doctor with agent.gate-review = copilot and a model exited $rc: $(cat "$t/d1")"
grep -qxF "$WARN" "$t/d1" || fail "doctor does not print '$WARN': $(cat "$t/d1")"
grep '^✓' "$t/d1" | grep -qw copilot || fail "doctor with agent.gate-review = copilot has no passing copilot row: $(cat "$t/d1")"
grep '^✓' "$t/d1" | grep -qF 'copilot login' || fail "doctor with agent.gate-review = copilot has no passing copilot login row: $(cat "$t/d1")"
runfix "$t/r1"
[ "$rc" = 0 ] || fail "a run with agent.gate-review = copilot and a model exited $rc: $(tail -8 "$t/r1")"
grep -qxF "  $WARN" "$t/r1" || fail "vloop run does not print '  $WARN': $(head -15 "$t/r1")"
[ "$(task status T1)" = done ] || fail "the warned run: T1 is $(task status T1), want done"
onecommit "the warned run"
ev "$t/r1" "^session copilot-gate-review - model=$MODEL\$" >/dev/null || fail "the run's gate-review session was not the copilot one: $(ev "$t/r1" '^session ')"

# The key is in use whatever run.verify says: the gate review has no switch.
CFG="$GRC
run.verify false"; fixture
doc "$t/d2"
grep -qxF "$WARN" "$t/d2" || fail "doctor with run.verify = false does not print '$WARN': $(cat "$t/d2")"
runfix "$t/r2"
[ "$rc" = 0 ] || fail "a run with agent.gate-review = copilot and run.verify = false exited $rc: $(tail -8 "$t/r2")"
grep -qxF "  $WARN" "$t/r2" || fail "with run.verify = false, vloop run does not print '  $WARN'"

# model.gate-review left at its default: refused before any session.
CFG="agent.gate-review copilot"; fixture
runfix "$t/m1"
[ "$rc" = 1 ] || fail "agent.gate-review = copilot with the default model.gate-review: vloop run exited $rc, want 1: $(tail -5 "$t/m1")"
grep -qxF "$REFUSE" "$t/m1" || fail "the refusal is not the brief's line '$REFUSE': $(cat "$t/m1")"
nosession "$t/m1" "the refused run (default model.gate-review)"
CFG=; fixture
runfix "$t/m2" VLOOP_AGENT_GATE_REVIEW=copilot
[ "$rc" = 1 ] || fail "VLOOP_AGENT_GATE_REVIEW=copilot with the default model.gate-review: vloop run exited $rc, want 1: $(tail -5 "$t/m2")"
grep -qxF "$REFUSE" "$t/m2" || fail "VLOOP_AGENT_GATE_REVIEW=copilot: no model.gate-review refusal: $(cat "$t/m2")"
nosession "$t/m2" "the refused run (VLOOP_AGENT_GATE_REVIEW)"

# K1: a project skill named like vloop's gate-review skill, or another of vloop's.
CFG=$GRC; fixture
skill .claude/skills/gate-review; commitfix
runfix "$t/k1"
[ "$rc" = 1 ] || fail "agent.gate-review = copilot with .claude/skills/gate-review: vloop run exited $rc, want 1: $(tail -5 "$t/k1")"
grep -qxF "vloop: .claude/skills/gate-review has the name of vloop's gate-review skill; a copilot session may load it instead — rename it" "$t/k1" ||
  fail "vloop run does not refuse .claude/skills/gate-review with K1's message: $(cat "$t/k1")"
nosession "$t/k1" "the refused run (.claude/skills/gate-review)"
doc "$t/dk1"
[ "$rc" = 1 ] || fail "doctor with .claude/skills/gate-review and agent.gate-review = copilot exited $rc, want 1: $(cat "$t/dk1")"
grep '^✗' "$t/dk1" | grep -qF ".claude/skills/gate-review has the name of vloop's gate-review skill" || fail "doctor has no error row for .claude/skills/gate-review: $(cat "$t/dk1")"
CFG=$GRC; fixture
skill .github/skills/verify; commitfix
runfix "$t/k2"
[ "$rc" = 1 ] || fail "agent.gate-review = copilot with .github/skills/verify: vloop run exited $rc, want 1: $(tail -5 "$t/k2")"
grep -qxF "vloop: .github/skills/verify has the name of vloop's verify skill; a copilot session may load it instead — rename it" "$t/k2" ||
  fail "vloop run does not refuse .github/skills/verify with K1's message: $(cat "$t/k2")"
nosession "$t/k2" "the refused run (.github/skills/verify)"

# The review look-alike warning: its base condition and words, now with only
# the gate review on copilot; no new warning for names ending in gate-review.
CFG=$GRC; fixture
skill .claude/skills/my-review; skill .claude/skills/my-gate-review; commitfix
doc "$t/dl"
[ "$rc" = 0 ] || fail "doctor with .claude/skills/my-review and agent.gate-review = copilot exited $rc: $(cat "$t/dl")"
grep '^!' "$t/dl" | grep -qF '.claude/skills/my-review may be loaded instead of /vloop:review by a copilot session' ||
  fail "doctor with only agent.gate-review = copilot has no '.claude/skills/my-review may be loaded instead of /vloop:review by a copilot session' warning: $(cat "$t/dl")"
grep -qF '/vloop:gate-review by a copilot session' "$t/dl" && fail "doctor warns of a look-alike gate-review skill name: $(grep -F '/vloop:gate-review' "$t/dl")"
runfix "$t/rl"
[ "$rc" = 0 ] || fail "a look-alike review skill stopped the run (exit $rc): $(tail -5 "$t/rl")"
grep -qxF '  ! .claude/skills/my-review may be loaded instead of /vloop:review by a copilot session' "$t/rl" ||
  fail "vloop run does not warn '  ! .claude/skills/my-review may be loaded instead of /vloop:review by a copilot session': $(head -15 "$t/rl")"

# agent.gate-review = claude: none of it, and a project skill named gate-review stops nothing.
CFG=; fixture
skill .claude/skills/gate-review; commitfix
runfix "$t/c0"
[ "$rc" = 0 ] || fail "with agent.gate-review = claude, .claude/skills/gate-review stopped the run (exit $rc): $(tail -5 "$t/c0")"
grep -qF "has the name of vloop's" "$t/c0" && fail "with agent.gate-review = claude the run refuses a project skill: $(cat "$t/c0")"
grep -qF 'agent.gate-review' "$t/c0" && fail "with agent.gate-review = claude the run speaks of agent.gate-review: $(grep -F agent.gate-review "$t/c0")"
doc "$t/dc0"
[ "$rc" = 0 ] || fail "doctor with agent.gate-review = claude exited $rc: $(cat "$t/dc0")"
grep -qF 'agent.gate-review' "$t/dc0" && fail "doctor with agent.gate-review = claude speaks of agent.gate-review: $(grep -F agent.gate-review "$t/dc0")"
grep -qF 'copilot login' "$t/dc0" && fail "doctor with every agent on claude has a copilot login row: $(cat "$t/dc0")"

# No GitHub login reaching the sessions.
CFG=$GRC; fixture
runfixnl "$t/l1"
[ "$rc" = 1 ] || fail "agent.gate-review = copilot with no GitHub login: vloop run exited $rc, want 1: $(tail -5 "$t/l1")"
grep -qF 'vloop: agent.gate-review is copilot but no GitHub login reaches its sessions' "$t/l1" ||
  fail "the run does not refuse the missing login naming agent.gate-review: $(cat "$t/l1")"
nosession "$t/l1" "the refused run (no login)"
docnl "$t/dl1"
[ "$rc" = 1 ] || fail "doctor with agent.gate-review = copilot and no GitHub login exited $rc, want 1: $(cat "$t/dl1")"
grep '^✗' "$t/dl1" | grep -qF 'copilot login' || fail "doctor has no error row for the copilot login: $(cat "$t/dl1")"

# copilot missing, then not runnable.
CFG=$GRC; fixture
rm -f "$t/stub/copilot"
runfix "$t/c1"
[ "$rc" = 1 ] || fail "agent.gate-review = copilot with no copilot on PATH: vloop run exited $rc, want 1: $(tail -5 "$t/c1")"
grep -q copilot "$t/c1" || fail "the refused run names no copilot problem: $(cat "$t/c1")"
nosession "$t/c1" "the refused run (no copilot)"
doc "$t/dc1"
[ "$rc" = 1 ] || fail "doctor with agent.gate-review = copilot and no copilot on PATH exited $rc, want 1: $(cat "$t/dc1")"
grep '^✗' "$t/dc1" | grep -q copilot || fail "doctor names no copilot problem: $(cat "$t/dc1")"
printf '#!/bin/sh\nexit 3\n' >"$t/stub/copilot" && chmod +x "$t/stub/copilot"
doc "$t/dc2"
[ "$rc" = 1 ] || fail "doctor with a copilot whose --version fails exited $rc, want 1: $(cat "$t/dc2")"
grep '^✗' "$t/dc2" | grep -qF 'copilot --version failed' || fail "doctor does not report copilot --version failing: $(cat "$t/dc2")"
runfix "$t/c2"
[ "$rc" = 1 ] || fail "agent.gate-review = copilot with a copilot whose --version fails: vloop run exited $rc, want 1: $(tail -5 "$t/c2")"
nosession "$t/c2" "the refused run (copilot --version fails)"

# agent.work's and agent.verify's checks keep their words.
CFG="agent.work copilot"; fixture
runfix "$t/w1"
[ "$rc" = 1 ] || fail "agent.work = copilot with the default model.work: vloop run exited $rc, want 1"
grep -qxF 'vloop: agent.work is copilot but model.work is the default "sonnet", a Claude alias: set model.work to a Copilot model id' "$t/w1" ||
  fail "agent.work's refusal changed: $(cat "$t/w1")"
CFG="agent.work copilot
model.work $MODEL"; fixture
doc "$t/w2"
grep -qxF '! agent.work is copilot: experimental (docs/guide/configuration.md)' "$t/w2" || fail "agent.work's doctor warning changed: $(cat "$t/w2")"
grep -qF 'agent.gate-review' "$t/w2" && fail "doctor with only agent.work = copilot speaks of agent.gate-review: $(grep -F agent.gate-review "$t/w2")"
CFG="agent.verify copilot"; fixture
runfix "$t/v1"
[ "$rc" = 1 ] && grep -qxF 'vloop: agent.verify is copilot but model.verify is the default "sonnet", a Claude alias: set model.verify to a Copilot model id' "$t/v1" ||
  fail "agent.verify's refusal changed (exit $rc): $(cat "$t/v1")"
echo "gate T3: ok"
