#!/bin/sh
# Gate T3 — H2 and K1. With agent.review = copilot, a fixture run whose stub
# copilot reads the hooks file of its own COPILOT_HOME and runs its bash form
# as Copilot would, with payloads on stdin: the hook is this tree's vloop,
# judging phase review against the repository root, in both command forms; the
# deny list stays on the argv. The stub prints tool events with two denials
# among other calls: the session record's permission_denials lists exactly
# those, in order, masked; with none it is []. K1: a project skill directory
# named like one of vloop's skills stops `vloop run` (exit 1, the brief's
# message, no session) and is an error row in doctor when the review agent is
# copilot; renamed, the run starts; with claude reviews it stops nothing.
G=$(cd "$(dirname "$0")" && pwd -P) || { echo "GATE FAIL: no gate folder" >&2; exit 1; }
. "$G/lib.sh"
cd "$top" || fail "repo root"
newbin
COP="agent.review copilot
model.review $MODEL"

# H2: the hook, the deny list, the denials.
CFG=$COP; fixture
runfix "$t/r1"
[ "$rc" = 0 ] || fail "a run with copilot reviews exited $rc: $(tail -8 "$t/r1")"
n=$(ev "$t/r1" '^session copilot-review ' | wc -l | tr -d ' ')
[ "$n" -ge 2 ] || fail "the run started $n copilot review sessions, want one per task (2): $(cat "$t/r1.ev")"
[ "$(ev "$t/r1" '^hooks ' | grep -cxE 'hooks [1-9][0-9]* bash=yes powershell=yes')" = "$n" ] ||
  fail "not every copilot session found a preToolUse hook with both the bash and powershell forms in COPILOT_HOME/hooks/: $(ev "$t/r1" '^hooks ')"
[ "$(ev "$t/r1" '^hook-phase ' | grep -cx 'hook-phase review')" = "$n" ] || fail "the hook is not run with --phase review: $(ev "$t/r1" '^hook-phase ')"
for want in 'shgh deny 0' 'envgh deny 0' 'diff allow 0' 'state-abs deny 0' 'tmp-abs allow 0' 'tmp-rel allow 0' 'webfetch deny 0'; do
  got=$(ev "$t/r1" "^hook ${want%% *} ")
  [ "$(printf '%s\n' "$got" | grep -cx "hook $want")" = "$n" ] || fail "the session's hook answered '$got', want 'hook $want' in every session"
done
for p in 'shell(gh:*)' 'shell(git push:*)' 'shell(rm:*)' url; do
  [ "$(ev "$t/r1" '^deny-tool ' | grep -cxF "deny-tool $p")" -ge "$n" ] || fail "the deny list on the argv lost $p: $(ev "$t/r1" '^deny-tool ' | sort -u)"
done
recs=$(phaserecs "$t/r1" review)
[ -n "$recs" ] || fail "no review session record"
want='[{"tool_name":"bash","tool_use_id":"c1","message":"Denied by preToolUse hook: vloop fence: gh is denied (seen from ~/notes)"},{"tool_name":"web_fetch","tool_use_id":"c4","message":"Permission to access this URL was denied."}]'
for f in $recs; do
  [ "$("$JQ" -r .agent "$f")" = copilot ] || fail "$(basename "$f"): a review session record of agent $("$JQ" -r .agent "$f")"
  "$JQ" -e --argjson w "$want" '.permission_denials == $w' "$f" >/dev/null ||
    fail "$(basename "$f"): permission_denials is $("$JQ" -c .permission_denials "$f"), want $want"
  sc=$("$JQ" -r .schema "$f")
  (cd "$t/repo" && venv "$S" schema validate "$sc" "$f" >"$t/val" 2>&1) || fail "$(basename "$f") is not valid $sc: $(cat "$t/val")"
done
grep -rqF "$t/home" "$t/r1.sessions" && fail "a session record keeps the HOME path unmasked"
CFG=$COP; fixture
: >"$t/stub/copilot.nodenials"
runfix "$t/r2"
[ "$rc" = 0 ] || fail "a run with copilot reviews and no denials exited $rc: $(tail -8 "$t/r2")"
for f in $(phaserecs "$t/r2" review); do
  "$JQ" -e '.permission_denials == []' "$f" >/dev/null || fail "$(basename "$f"): with no denials permission_denials is $("$JQ" -c .permission_denials "$f"), want []"
done

# K1: the preflight refusal and doctor's error row.
MSG="has the name of vloop's review skill; a copilot session may load it instead — rename it"
skill() { mkdir -p "$t/repo/$1" && printf -- '---\nname: %s\ndescription: a project skill\n---\nBody.\n' "$(basename "$1")" >"$t/repo/$1/SKILL.md"; }
nosession() { [ -s "$1.ev" ] && fail "$2: a session was started: $(head -3 "$1.ev")"; return 0; }
CFG=$COP; fixture
skill .claude/skills/review; commitfix
runfix "$t/k1"
[ "$rc" = 1 ] || fail "vloop run with .claude/skills/review and copilot reviews exited $rc, want 1: $(tail -5 "$t/k1")"
grep -qxF "vloop: .claude/skills/review $MSG" "$t/k1" || fail "vloop run does not print 'vloop: .claude/skills/review $MSG': $(cat "$t/k1")"
nosession "$t/k1" "the refused run"
(cd "$t/repo" && venv "$S" doctor >"$t/d1" 2>&1); c=$?
[ "$c" = 1 ] || fail "doctor with .claude/skills/review and copilot reviews exited $c, want 1: $(cat "$t/d1")"
grep '^✗' "$t/d1" | grep -qF ".claude/skills/review $MSG" || fail "doctor has no error row for .claude/skills/review $MSG: $(cat "$t/d1")"
CFG=$COP; fixture
skill .github/skills/plan; commitfix
runfix "$t/k2"
[ "$rc" = 1 ] || fail "vloop run with .github/skills/plan and copilot reviews exited $rc, want 1"
grep -qxF "vloop: .github/skills/plan has the name of vloop's plan skill; a copilot session may load it instead — rename it" "$t/k2" ||
  fail "vloop run does not refuse .github/skills/plan with the brief's message: $(cat "$t/k2")"
nosession "$t/k2" "the refused run"
# Renamed, it starts; the warning for names ending in review stays.
CFG=$COP; fixture
skill .claude/skills/repo-review; commitfix
runfix "$t/k3"
[ "$rc" = 0 ] || fail "vloop run with .claude/skills/repo-review exited $rc, want 0: $(tail -8 "$t/k3")"
(cd "$t/repo" && venv "$S" doctor >"$t/d3" 2>&1); c=$?
[ "$c" = 0 ] || fail "doctor with .claude/skills/repo-review exited $c: $(cat "$t/d3")"
grep -qxF "! .claude/skills/repo-review may be loaded instead of /vloop:review by a copilot session" "$t/d3" ||
  fail "doctor dropped the warning for .claude/skills/repo-review: $(cat "$t/d3")"
# Claude reviews: the same skill stops nothing.
CFG=; fixture
skill .claude/skills/review; commitfix
runfix "$t/k4"
[ "$rc" = 0 ] || fail "with claude reviews, .claude/skills/review stopped the run (exit $rc): $(tail -5 "$t/k4")"
grep -qF "$MSG" "$t/k4" && fail "with claude reviews the run refuses .claude/skills/review"
echo "gate T3: ok"
