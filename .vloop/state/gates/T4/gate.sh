#!/bin/sh
# Gate T4 — C4. With agent.review = copilot: doctor and the run preflight
# check copilot (on PATH, --version runs: a problem otherwise, and the run
# starts no session) and print the experimental warning, which stops nothing;
# the run refuses a model.review that comes from the default, with the
# brief's message, before any session; doctor warns of each project skill in
# .claude/skills/ whose name ends in "review". With every agent claude, doctor
# and the run need no copilot and say nothing of it. The configuration guide
# documents copilot for agent.review, copilot.credit-usd, the experimental
# status.
G=$(cd "$(dirname "$0")" && pwd -P) || { echo "GATE FAIL: no gate folder" >&2; exit 1; }
. "$G/lib.sh"
newbin
S=$t/bin/new
WARN='agent.review is copilot: experimental — review sessions are not fenced until B20 (docs/guide/configuration.md)'
REFUSE='vloop: agent.review is copilot but model.review is the default "sonnet", a Claude alias: set model.review to a Copilot model id'
# doc <out> [env...]: vloop doctor in the fixture; $rc its exit code.
doc() { o=$1; shift; (cd "$t/repo" && venv "$@" "$S" doctor >"$o" 2>&1); rc=$?; }
nosession() { [ -f "$t/stub/log" ] && grep -q '^arg -p$' "$t/stub/log" && fail "$1: a session was started"; rm -f "$t/stub/log"; return 0; }

# Copilot configured and working: the warning, exit 0, no problem.
CFG=$COPILOT_CFG; fixture
doc "$t/d1"
[ "$rc" = 0 ] || fail "doctor with agent.review copilot exited $rc: $(cat "$t/d1")"
grep -qxF "! $WARN" "$t/d1" || fail "doctor with agent.review copilot does not print '! $WARN': $(cat "$t/d1")"
grep -q '^✗' "$t/d1" && fail "doctor with a working copilot reports a problem: $(grep '^✗' "$t/d1")"
grep -q '^=== copilot$' "$t/stub/log" && grep -qx 'arg --version' "$t/stub/log" || fail "doctor did not run copilot --version"
rm -f "$t/stub/log"
# copilot missing, or its --version failing: a problem in doctor and in the run.
rm -f "$t/stub/copilot"
doc "$t/d2"
[ "$rc" = 1 ] || fail "doctor with agent.review copilot and no copilot on PATH exited $rc, want 1: $(cat "$t/d2")"
grep '^✗' "$t/d2" | grep -q copilot || fail "doctor with no copilot on PATH names no copilot problem: $(cat "$t/d2")"
(cd "$t/repo" && venv "$S" run "$BRIEF" >"$t/r2" 2>&1); c=$?
[ "$c" = 1 ] || fail "vloop run with no copilot on PATH exited $c, want 1: $(tail -5 "$t/r2")"
grep -q copilot "$t/r2" || fail "vloop run with no copilot on PATH does not say so: $(tail -5 "$t/r2")"
nosession "vloop run with no copilot on PATH"
cp "$G/stub-copilot.sh" "$t/stub/copilot" && chmod +x "$t/stub/copilot" && : >"$t/stub/copilot.noversion"
doc "$t/d3"
[ "$rc" = 1 ] || fail "doctor with a copilot whose --version fails exited $rc, want 1: $(cat "$t/d3")"
grep '^✗' "$t/d3" | grep -q copilot || fail "doctor with a broken copilot names no copilot problem: $(cat "$t/d3")"
(cd "$t/repo" && venv "$S" run "$BRIEF" >"$t/r3" 2>&1); c=$?
[ "$c" = 1 ] || fail "vloop run with a broken copilot exited $c, want 1: $(tail -5 "$t/r3")"
nosession "vloop run with a broken copilot"
rm -f "$t/stub/copilot.noversion"

# The run preflight: the warning, and the run goes on.
(cd "$t/repo" && venv "$S" run "$BRIEF" >"$t/r4" 2>&1); c=$?
[ "$c" = 0 ] || fail "vloop run with agent.review copilot and a model exited $c: $(tail -8 "$t/r4")"
sed 's/^[[:space:]]*//' "$t/r4" | grep -qxF "! $WARN" || fail "the run preflight does not print '! $WARN': $(head -15 "$t/r4")"
grep -q '^=== copilot$' "$t/stub/log" || fail "the run with copilot reviews started no copilot"
rm -f "$t/stub/log"

# model.review from the default: refused before any session, in the
# worked example's form (agent from the environment) and from the file.
CFG=; fixture
(cd "$t/repo" && venv VLOOP_AGENT_REVIEW=copilot "$S" run "$BRIEF" >"$t/o5" 2>"$t/e5"); c=$?
[ "$c" = 1 ] || fail "VLOOP_AGENT_REVIEW=copilot vloop run with the default model.review exited $c, want 1: $(cat "$t/e5")"
grep -qxF "$REFUSE" "$t/e5" || fail "VLOOP_AGENT_REVIEW=copilot vloop run with the default model.review: stderr has no line '$REFUSE': $(cat "$t/e5")"
nosession "the refused run"
CFG="agent.review copilot"; fixture
(cd "$t/repo" && venv "$S" run "$BRIEF" >"$t/o6" 2>"$t/e6"); c=$?
[ "$c" = 1 ] || fail "agent.review copilot in the file and the default model.review: vloop run exited $c, want 1"
grep -qxF "$REFUSE" "$t/e6" || fail "agent.review copilot in the file, model.review default: stderr has no line '$REFUSE': $(cat "$t/e6")"
nosession "the refused run (file)"
# A model.review that is set is not the default.
(cd "$t/repo" && venv VLOOP_MODEL_REVIEW="$MODEL" "$S" run "$BRIEF" >"$t/o7" 2>&1); c=$?
[ "$c" = 0 ] || fail "agent.review copilot with VLOOP_MODEL_REVIEW set: vloop run exited $c: $(tail -8 "$t/o7")"
grep -qF "$REFUSE" "$t/o7" && fail "a model.review from the environment was refused as the default"
rm -f "$t/stub/log"

# Skill collisions: each project skill whose name ends in "review".
CFG=$COPILOT_CFG; fixture
for s in loop-review code-review reviewer; do
  mkdir -p "$t/repo/.claude/skills/$s" && printf -- '---\nname: %s\ndescription: a project skill\n---\nBody.\n' "$s" >"$t/repo/.claude/skills/$s/SKILL.md"
done
venv git -C "$t/repo" add -A && venv git -C "$t/repo" commit -q -m skills || fail "commit skills"
doc "$t/d8"
[ "$rc" = 0 ] || fail "doctor with colliding skills exited $rc: $(cat "$t/d8")"
for s in loop-review code-review; do
  grep -qxF "! .claude/skills/$s may be loaded instead of /vloop:review by a copilot session" "$t/d8" ||
    fail "doctor does not warn '! .claude/skills/$s may be loaded instead of /vloop:review by a copilot session': $(cat "$t/d8")"
done
grep -q 'skills/reviewer' "$t/d8" && fail "doctor warns of .claude/skills/reviewer, whose name does not end in review"
grep -qxF "! $WARN" "$t/d8" || fail "doctor with colliding skills dropped the experimental warning"

# Every agent claude: no copilot needed, nothing said of it.
CFG=; fixture
rm -f "$t/stub/copilot"
for s in loop-review; do
  mkdir -p "$t/repo/.claude/skills/$s" && printf -- '---\nname: %s\ndescription: a project skill\n---\nBody.\n' "$s" >"$t/repo/.claude/skills/$s/SKILL.md"
done
venv git -C "$t/repo" add -A && venv git -C "$t/repo" commit -q -m skills || fail "commit skills"
doc "$t/d9"
[ "$rc" = 0 ] || fail "doctor with every agent claude and no copilot exited $rc: $(cat "$t/d9")"
grep -qi copilot "$t/d9" && fail "doctor with every agent claude speaks of copilot: $(grep -i copilot "$t/d9")"
(cd "$t/repo" && venv "$S" run "$BRIEF" >"$t/r9" 2>&1); c=$?
[ "$c" = 0 ] || fail "an all-claude run with no copilot on PATH exited $c: $(tail -8 "$t/r9")"
grep -qi copilot "$t/r9" && fail "an all-claude run speaks of copilot: $(grep -i copilot "$t/r9")"

# The configuration guide.
GUIDE=$top/docs/guide/configuration.md
row() { grep -E "^\|[[:space:]]*\`$1\`[[:space:]]*\|" "$GUIDE" || fail "docs/guide/configuration.md has no table row for $1"; }
col() { echo "$1" | awk -F'|' -v n="$2" '{ c = $(n + 1); gsub(/^[ \t]+|[ \t]+$/, "", c); print c }'; }
r=$(row 'agent\.review')
col "$r" 3 | grep -q '`copilot`' || fail "the agent.review row's valid values do not name \`copilot\`: $r"
col "$r" 3 | grep -q '`claude`' || fail "the agent.review row's valid values do not name \`claude\`: $r"
for p in plan work 'gate-review' verify; do
  r=$(row "agent\\.$p")
  col "$r" 3 | grep -q copilot && fail "the agent.$p row offers copilot: $r"
done
r=$(row 'copilot\.credit-usd')
[ "$(col "$r" 2)" = '`0.0066667`' ] || fail "the copilot.credit-usd row's default is '$(col "$r" 2)', want \`0.0066667\`: $r"
[ "$(col "$r" 4)" = '`VLOOP_COPILOT_CREDIT_USD`' ] || fail "the copilot.credit-usd row's variable is '$(col "$r" 4)', want \`VLOOP_COPILOT_CREDIT_USD\`: $r"
grep -qi experimental "$GUIDE" && grep -q B20 "$GUIDE" || fail "docs/guide/configuration.md does not say copilot reviews are experimental until B20"
echo "gate T4: ok"
