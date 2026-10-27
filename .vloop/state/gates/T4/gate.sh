#!/bin/sh
# Gate T4 — S1's keys and doctor, and the reworded Copilot warning.
# run.second-review (off, record, feedback, on-pass, both-pass; default off),
# agent.second-review (claude, copilot; default claude), model.second-review
# (default sonnet), effort.second-review: each read from the file and from its
# VLOOP_ variable, and an invalid value refused with exit 2. Doctor: nothing of
# the second review when off; when on, a row naming mode, agent and model, the
# no-independence warning when agent and model match the primary's, and the
# copilot checks for a copilot second reviewer. The experimental warning in
# its new words, for agent.review and agent.second-review. A copilot second
# review with the default model is refused before any session, and K1's
# refusal applies to agent.second-review when the second review is in use.
G=$(cd "$(dirname "$0")" && pwd -P) || { echo "GATE FAIL: no gate folder" >&2; exit 1; }
. "$G/lib.sh"
cd "$top" || fail "repo root"
newbin
get() { k=$1; shift; (cd "$t/repo" && venv "$@" "$S" config get "$k" 2>&1); }
set_() { (cd "$t/repo" && venv "$S" config set "$1" "$2" >"$t/set.out" 2>&1); }
doc() { o=$1; shift; (cd "$t/repo" && venv "$@" "$S" doctor >"$o" 2>&1); rc=$?; }
nosession() { [ -s "$1.ev" ] && fail "$2: a session was started: $(head -3 "$1.ev")"; return 0; }
SAME="! run.second-review: the second reviewer is the primary's agent and model — it adds no independence"

# The keys.
CFG=; fixture
[ "$(get run.second-review)" = off ] || fail "run.second-review defaults to '$(get run.second-review)', want off"
[ "$(get agent.second-review)" = claude ] || fail "agent.second-review defaults to '$(get agent.second-review)', want claude"
[ "$(get model.second-review)" = sonnet ] || fail "model.second-review defaults to '$(get model.second-review)', want sonnet"
[ -z "$(get effort.second-review)" ] || fail "effort.second-review has a default: '$(get effort.second-review)'"
for m in record feedback on-pass both-pass off; do
  set_ run.second-review "$m" || fail "vloop config set run.second-review $m failed: $(cat "$t/set.out")"
  [ "$(get run.second-review)" = "$m" ] || fail "run.second-review set to $m reads '$(get run.second-review)'"
done
set_ run.second-review always; c=$?
[ "$c" = 2 ] || fail "vloop config set run.second-review always exited $c, want 2"
set_ agent.second-review copilot || fail "vloop config set agent.second-review copilot failed: $(cat "$t/set.out")"
set_ agent.second-review codex; c=$?
[ "$c" = 2 ] || fail "vloop config set agent.second-review codex exited $c, want 2"
set_ effort.second-review high || fail "vloop config set effort.second-review high failed: $(cat "$t/set.out")"
set_ effort.second-review extreme; c=$?
[ "$c" = 2 ] || fail "vloop config set effort.second-review extreme exited $c, want 2"
set_ model.second-review some-model || fail "vloop config set model.second-review failed: $(cat "$t/set.out")"
[ "$(get run.second-review VLOOP_RUN_SECOND_REVIEW=both-pass)" = both-pass ] || fail "VLOOP_RUN_SECOND_REVIEW does not set run.second-review"
[ "$(get agent.second-review VLOOP_AGENT_SECOND_REVIEW=claude)" = claude ] || fail "VLOOP_AGENT_SECOND_REVIEW does not set agent.second-review"
[ "$(get model.second-review VLOOP_MODEL_SECOND_REVIEW=env-model)" = env-model ] || fail "VLOOP_MODEL_SECOND_REVIEW does not set model.second-review"
[ "$(get effort.second-review VLOOP_EFFORT_SECOND_REVIEW=low)" = low ] || fail "VLOOP_EFFORT_SECOND_REVIEW does not set effort.second-review"

# Doctor, second review off: nothing of it.
CFG=; fixture
doc "$t/d0"
[ "$rc" = 0 ] || fail "doctor in a plain fixture exited $rc: $(cat "$t/d0")"
grep -qiE 'second[- ]review' "$t/d0" && fail "doctor with run.second-review off speaks of the second review: $(grep -iE 'second[- ]review' "$t/d0")"
# On, with a model of its own: a row naming mode, agent and model; no warning.
CFG="run.second-review feedback
model.second-review $SECOND"; fixture
doc "$t/d1"
[ "$rc" = 0 ] || fail "doctor with a second reviewer exited $rc: $(cat "$t/d1")"
grep -F 'second-review' "$t/d1" | grep -F feedback | grep -F claude | grep -qF "$SECOND" ||
  fail "doctor has no row naming the second review's mode (feedback), agent (claude) and model ($SECOND): $(cat "$t/d1")"
grep -qxF "$SAME" "$t/d1" && fail "doctor warns of no independence for a second reviewer with its own model"
# On, matching the primary: the warning.
CFG="run.second-review on-pass"; fixture
doc "$t/d2"
[ "$rc" = 0 ] || fail "doctor with a second reviewer matching the primary exited $rc: $(cat "$t/d2")"
grep -qxF "$SAME" "$t/d2" || fail "doctor does not print '$SAME': $(cat "$t/d2")"
CFG="agent.review copilot
model.review $MODEL
run.second-review record
agent.second-review copilot
model.second-review $MODEL"; fixture
doc "$t/d3"
grep -qxF "$SAME" "$t/d3" || fail "doctor does not warn when both reviewers are copilot $MODEL: $(cat "$t/d3")"

# The experimental warning, new words; copilot checks for the second reviewer.
CFG="agent.review copilot
model.review $MODEL"; fixture
doc "$t/d4"
[ "$rc" = 0 ] || fail "doctor with copilot reviews exited $rc: $(cat "$t/d4")"
grep -qxF '! agent.review is copilot: experimental (docs/guide/configuration.md)' "$t/d4" ||
  fail "doctor does not print '! agent.review is copilot: experimental (docs/guide/configuration.md)': $(cat "$t/d4")"
grep -qi 'not fenced' "$t/d4" && fail "doctor still says copilot reviews are not fenced"
runfix "$t/r4"
[ "$rc" = 0 ] || fail "a run with copilot reviews exited $rc: $(tail -8 "$t/r4")"
sed 's/^[[:space:]]*//' "$t/r4" | grep -qxF '! agent.review is copilot: experimental (docs/guide/configuration.md)' ||
  fail "the run preflight does not print the new experimental warning: $(head -15 "$t/r4")"
grep -qi 'not fenced' "$t/r4" && fail "the run still says copilot reviews are not fenced"
CFG="run.second-review record
agent.second-review copilot
model.second-review $MODEL"; fixture
doc "$t/d5"
[ "$rc" = 0 ] || fail "doctor with a copilot second reviewer exited $rc: $(cat "$t/d5")"
grep -qxF '! agent.second-review is copilot: experimental (docs/guide/configuration.md)' "$t/d5" ||
  fail "doctor does not print '! agent.second-review is copilot: experimental (docs/guide/configuration.md)': $(cat "$t/d5")"
rm -f "$t/stub/copilot"
doc "$t/d6"
[ "$rc" = 1 ] || fail "doctor with a copilot second reviewer and no copilot on PATH exited $rc, want 1: $(cat "$t/d6")"
grep '^✗' "$t/d6" | grep -q copilot || fail "doctor with a copilot second reviewer and no copilot names no copilot problem: $(cat "$t/d6")"

# A copilot second review with the default model: refused before any session.
CFG="run.second-review feedback
agent.second-review copilot"; fixture
runfix "$t/r7"
[ "$rc" = 1 ] || fail "a copilot second review with the default model.second-review: vloop run exited $rc, want 1: $(tail -5 "$t/r7")"
grep -qF model.second-review "$t/r7" || fail "the refusal does not name model.second-review: $(cat "$t/r7")"
nosession "$t/r7" "the refused run"

# K1 for the second reviewer: in use, refused; off, it starts.
skill() { mkdir -p "$t/repo/$1" && printf -- '---\nname: %s\ndescription: a project skill\n---\nBody.\n' "$(basename "$1")" >"$t/repo/$1/SKILL.md"; }
CFG="run.second-review record
agent.second-review copilot
model.second-review $MODEL"; fixture
skill .claude/skills/review; commitfix
runfix "$t/r8"
[ "$rc" = 1 ] || fail "a copilot second reviewer with .claude/skills/review: vloop run exited $rc, want 1: $(tail -5 "$t/r8")"
grep -qxF "vloop: .claude/skills/review has the name of vloop's review skill; a copilot session may load it instead — rename it" "$t/r8" ||
  fail "a copilot second reviewer with .claude/skills/review: no K1 message: $(cat "$t/r8")"
nosession "$t/r8" "the refused run"
CFG="agent.second-review copilot
model.second-review $MODEL"; fixture
skill .claude/skills/review; commitfix
runfix "$t/r9"
[ "$rc" = 0 ] || fail "run.second-review off with agent.second-review copilot and .claude/skills/review: vloop run exited $rc, want 0: $(tail -5 "$t/r9")"
echo "gate T4: ok"
