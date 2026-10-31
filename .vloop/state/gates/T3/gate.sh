#!/bin/sh
# Gate T3 — V2, the run preflight and doctor for agent.verify = copilot. The
# experimental warning in `vloop run` and in doctor; the refusal of the default
# model.verify before any session; the copilot checks the other Copilot phases
# have — copilot on PATH and runnable, a GitHub login reaching its sessions,
# K1's refusal of a project skill named like one of vloop's — each stopping
# the run (exit 1, no session) and each a doctor row; none of it with
# run.verify = false or agent.verify = claude; agent.work's checks keep their
# words. Fixture runs with the stub claude and the stub copilot.
G=$(cd "$(dirname "$0")" && pwd -P) || { echo "GATE FAIL: no gate folder" >&2; exit 1; }
. "$G/lib.sh"
cd "$top" || fail "repo root"
newbin
VER="agent.verify copilot
model.verify $MODEL"
WARN='! agent.verify is copilot: experimental (docs/guide/configuration.md)'
REFUSE='vloop: agent.verify is copilot but model.verify is the default "sonnet", a Claude alias: set model.verify to a Copilot model id'
doc() { o=$1; shift; (cd "$t/repo" && venv "$@" "$S" doctor >"$o" 2>&1); rc=$?; }
docnl() { o=$1; shift; (cd "$t/repo" && venvnl "$@" "$S" doctor >"$o" 2>&1); rc=$?; }
stripped() { sed 's/^[[:space:]]*//' "$1"; }

# The warning, in the run and in doctor; the run goes ahead.
CFG=$VER; fixture
doc "$t/d1"
[ "$rc" = 0 ] || fail "doctor with agent.verify = copilot and a model exited $rc: $(cat "$t/d1")"
grep -qxF "$WARN" "$t/d1" || fail "doctor does not print '$WARN': $(cat "$t/d1")"
grep '^✓' "$t/d1" | grep -qw copilot || fail "doctor with agent.verify = copilot has no passing copilot row: $(cat "$t/d1")"
grep '^✓' "$t/d1" | grep -qF 'copilot login' || fail "doctor with agent.verify = copilot has no passing copilot login row: $(cat "$t/d1")"
runfix "$t/r1"
[ "$rc" = 0 ] || fail "a run with agent.verify = copilot and a model exited $rc: $(tail -8 "$t/r1")"
grep -qxF "  $WARN" "$t/r1" || fail "vloop run does not print '  $WARN': $(head -15 "$t/r1")"
[ "$(task status T1)" = done ] || fail "the warned run: T1 is $(task status T1), want done"
onecommit "the warned run"
grep -qE '^[[:space:]]*verify: no findings$' "$t/r1" || fail "the warned run does not print 'verify: no findings': $(tail -8 "$t/r1")"
ev "$t/r1" "^session copilot-verify - model=$MODEL\$" >/dev/null || fail "the run's verify session was not the copilot one: $(ev "$t/r1" '^session ')"

# model.verify left at its default: refused before any session.
CFG="agent.verify copilot"; fixture
runfix "$t/r2"
[ "$rc" = 1 ] || fail "agent.verify = copilot with the default model.verify: vloop run exited $rc, want 1: $(tail -5 "$t/r2")"
grep -qxF "$REFUSE" "$t/r2" || fail "the refusal is not the brief's line '$REFUSE': $(cat "$t/r2")"
nosession "$t/r2" "the refused run (default model.verify)"
CFG=; fixture
runfix "$t/r2e" VLOOP_AGENT_VERIFY=copilot
[ "$rc" = 1 ] || fail "VLOOP_AGENT_VERIFY=copilot with the default model.verify: vloop run exited $rc, want 1: $(tail -5 "$t/r2e")"
grep -qxF "$REFUSE" "$t/r2e" || fail "VLOOP_AGENT_VERIFY=copilot: no model.verify refusal: $(cat "$t/r2e")"
nosession "$t/r2e" "the refused run (VLOOP_AGENT_VERIFY)"

# K1: a project skill named like vloop's verify skill.
CFG=$VER; fixture
skill .claude/skills/verify; commitfix
runfix "$t/k1"
[ "$rc" = 1 ] || fail "agent.verify = copilot with .claude/skills/verify: vloop run exited $rc, want 1: $(tail -5 "$t/k1")"
grep -qxF "vloop: .claude/skills/verify has the name of vloop's verify skill; a copilot session may load it instead — rename it" "$t/k1" ||
  fail "vloop run does not refuse .claude/skills/verify with K1's message: $(cat "$t/k1")"
nosession "$t/k1" "the refused run (.claude/skills/verify)"
doc "$t/dk1"
[ "$rc" = 1 ] || fail "doctor with .claude/skills/verify and agent.verify = copilot exited $rc, want 1: $(cat "$t/dk1")"
grep '^✗' "$t/dk1" | grep -qF ".claude/skills/verify has the name of vloop's verify skill" || fail "doctor has no error row for .claude/skills/verify: $(cat "$t/dk1")"
CFG=$VER; fixture
skill .github/skills/review; commitfix
runfix "$t/k2"
[ "$rc" = 1 ] || fail "agent.verify = copilot with .github/skills/review: vloop run exited $rc, want 1: $(tail -5 "$t/k2")"
grep -qxF "vloop: .github/skills/review has the name of vloop's review skill; a copilot session may load it instead — rename it" "$t/k2" ||
  fail "vloop run does not refuse .github/skills/review with K1's message: $(cat "$t/k2")"
nosession "$t/k2" "the refused run (.github/skills/review)"

# run.verify = false: agent.verify = copilot adds no row, warning or refusal.
CFG="agent.verify copilot
run.verify false"; fixture
skill .claude/skills/verify; commitfix
runfix "$t/f1"
[ "$rc" = 0 ] || fail "run.verify = false, agent.verify = copilot, model.verify unset: vloop run exited $rc, want 0: $(tail -8 "$t/f1")"
[ "$(task status T1)" = done ] || fail "run.verify = false: T1 is $(task status T1), want done"
grep -qF 'agent.verify' "$t/f1" && fail "run.verify = false: the run speaks of agent.verify: $(grep -F agent.verify "$t/f1")"
grep -qF "has the name of vloop's" "$t/f1" && fail "run.verify = false: the run refuses a project skill: $(cat "$t/f1")"
ev "$t/f1" 'verify' >/dev/null && fail "run.verify = false: a verify session ran: $(ev "$t/f1" 'verify')"
doc "$t/df1"
[ "$rc" = 0 ] || fail "doctor with run.verify = false and agent.verify = copilot exited $rc: $(cat "$t/df1")"
grep -qF 'agent.verify' "$t/df1" && fail "doctor with run.verify = false speaks of agent.verify: $(grep -F agent.verify "$t/df1")"
grep -qF "has the name of vloop's" "$t/df1" && fail "doctor with run.verify = false reports a project skill: $(cat "$t/df1")"

# agent.verify = claude: none of it, and a project skill named verify stops nothing.
CFG=; fixture
skill .claude/skills/verify; commitfix
runfix "$t/c0"
[ "$rc" = 0 ] || fail "with agent.verify = claude, .claude/skills/verify stopped the run (exit $rc): $(tail -5 "$t/c0")"
grep -qF "has the name of vloop's" "$t/c0" && fail "with agent.verify = claude the run refuses a project skill: $(cat "$t/c0")"
grep -qF 'agent.verify is copilot' "$t/c0" && fail "with agent.verify = claude the run warns of copilot"
doc "$t/dc0"
[ "$rc" = 0 ] || fail "doctor with agent.verify = claude exited $rc: $(cat "$t/dc0")"
grep -qF 'agent.verify' "$t/dc0" && fail "doctor with agent.verify = claude speaks of agent.verify: $(grep -F agent.verify "$t/dc0")"

# No GitHub login reaching the sessions.
CFG=$VER; fixture
runfixnl "$t/l1"
[ "$rc" = 1 ] || fail "agent.verify = copilot with no GitHub login: vloop run exited $rc, want 1: $(tail -5 "$t/l1")"
grep -qF 'vloop: agent.verify is copilot but no GitHub login reaches its sessions' "$t/l1" ||
  fail "the run does not refuse the missing login naming agent.verify: $(cat "$t/l1")"
nosession "$t/l1" "the refused run (no login)"
docnl "$t/dl1"
[ "$rc" = 1 ] || fail "doctor with agent.verify = copilot and no GitHub login exited $rc, want 1: $(cat "$t/dl1")"
grep '^✗' "$t/dl1" | grep -qF 'copilot login' || fail "doctor has no error row for the copilot login: $(cat "$t/dl1")"

# copilot missing, then not runnable.
CFG=$VER; fixture
rm -f "$t/stub/copilot"
runfix "$t/c1"
[ "$rc" = 1 ] || fail "agent.verify = copilot with no copilot on PATH: vloop run exited $rc, want 1: $(tail -5 "$t/c1")"
grep -q copilot "$t/c1" || fail "the refused run names no copilot problem: $(cat "$t/c1")"
nosession "$t/c1" "the refused run (no copilot)"
doc "$t/dc1"
[ "$rc" = 1 ] || fail "doctor with agent.verify = copilot and no copilot on PATH exited $rc, want 1: $(cat "$t/dc1")"
grep '^✗' "$t/dc1" | grep -q copilot || fail "doctor names no copilot problem: $(cat "$t/dc1")"
printf '#!/bin/sh\nexit 3\n' >"$t/stub/copilot" && chmod +x "$t/stub/copilot"
doc "$t/dc2"
[ "$rc" = 1 ] || fail "doctor with a copilot whose --version fails exited $rc, want 1: $(cat "$t/dc2")"
grep '^✗' "$t/dc2" | grep -qF 'copilot --version failed' || fail "doctor does not report copilot --version failing: $(cat "$t/dc2")"
runfix "$t/c2"
[ "$rc" = 1 ] || fail "agent.verify = copilot with a copilot whose --version fails: vloop run exited $rc, want 1: $(tail -5 "$t/c2")"
nosession "$t/c2" "the refused run (copilot --version fails)"

# agent.work's checks keep their words.
CFG="agent.work copilot"; fixture
runfix "$t/w1"
[ "$rc" = 1 ] || fail "agent.work = copilot with the default model.work: vloop run exited $rc, want 1"
grep -qxF 'vloop: agent.work is copilot but model.work is the default "sonnet", a Claude alias: set model.work to a Copilot model id' "$t/w1" ||
  fail "agent.work's refusal changed: $(cat "$t/w1")"
CFG="agent.work copilot
model.work $MODEL"; fixture
doc "$t/w2"
grep -qxF '! agent.work is copilot: experimental (docs/guide/configuration.md)' "$t/w2" || fail "agent.work's doctor warning changed: $(cat "$t/w2")"
grep -qF 'agent.verify' "$t/w2" && fail "doctor with only agent.work = copilot speaks of agent.verify: $(grep -F agent.verify "$t/w2")"
echo "gate T3: ok"
