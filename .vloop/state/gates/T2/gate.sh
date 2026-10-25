#!/bin/sh
# Gate T2 — C1's session and C2's denials. In a fixture run with
# agent.review = copilot and an explicit model.review, every review session is
# a copilot session (the stub in stub-copilot.sh) and every other session is
# claude's, exactly as in the same run with every phase on claude. For each
# review session: -p "/vloop:review <task>", --model, --effort only when
# effort.review is set, --plugin-dir the extracted plugin the claude sessions
# get, --no-ask-user, --output-format json, --allow-all-tools,
# --disable-builtin-mcps, and none of --allow-all-paths, --add-dir, --sandbox,
# --no-sandbox, --allow-all, --yolo; the repository root as working
# directory; the environment a claude review session gets plus COPILOT_HOME
# (an empty directory under .vloop/tmp/, never reused, never ~/.copilot) and,
# with guidelines, COPILOT_CUSTOM_INSTRUCTIONS_DIRS naming a directory under
# .vloop/tmp/ that holds them. The stub applies the --deny-tool patterns it was
# given to every C2 form and to commands a review needs, and each verdict is
# checked here. Last, B18's comparison gate (b18/, byte for byte as B18 ran
# it) still passes: claude sessions are unchanged.
G=$(cd "$(dirname "$0")" && pwd -P) || { echo "GATE FAIL: no gate folder" >&2; exit 1; }
. "$G/lib.sh"
newbin

# The same two-task run, all on claude and with copilot reviews.
fixture; runfix "$t/cl" $CARRY
[ "$rc" = 0 ] || fail "the all-claude run exited $rc: $(tail -8 "$t/cl")"
CFG=$COPILOT_CFG fixture; runfix "$t/cop" $CARRY
[ "$rc" = 0 ] || fail "the run with agent.review copilot exited $rc: $(tail -8 "$t/cop")"
L=$t/cop.log
s=$(sessions "$L" copilot | tr '\n' ';')
[ "$s" = "/vloop:review T1;/vloop:review T2;" ] || fail "copilot sessions started with -p '$s', want the reviews of T1 and T2"
sessions "$L" claude | grep -q '^/vloop:review' && fail "a review session started on claude with agent.review copilot"
for p in "/vloop:plan $BRIEF" /vloop:gate-review "/vloop:work T1" "/vloop:work T2" /vloop:verify; do
  sessions "$L" claude | grep -qxF "$p" || fail "no claude session started with -p '$p'"
done

# Claude sessions: what claude received is what it receives in the all-claude
# run, every session but the reviews.
noreview() {
  awk 'function flush() { if (inb && !drop) printf "%s", buf; buf = ""; inb = 0; drop = 0 }
    /^=== / { flush(); if ($0 == "=== claude") { inb = 1; buf = $0 "\n" } next }
    inb { buf = buf $0 "\n"; if ($0 ~ /^arg \/vloop:review /) drop = 1 }
    END { flush() }' "$1"
}
noreview "$t/cl.log" >"$t/cl.claude"; noreview "$L" >"$t/cop.claude"
diff "$t/cl.claude" "$t/cop.claude" >"$t/claude.diff" ||
  fail "with copilot reviews, what the claude sessions received differs from the all-claude run (< all claude, > copilot reviews):
$(head -30 "$t/claude.diff")"

C2="gh pr view 1
gh api repos/o/r
gh auth status
git push origin main
git push
git remote add other https://example.invalid/r.git
git remote -v
git config --global user.name x
git config --system core.editor x
git config --file ../outside.cfg a.b c
curl https://example.invalid
wget https://example.invalid
ssh host.invalid
scp a host.invalid:b
nc host.invalid 80
rm -rf build
rm x
git commit -m x
git reset --hard
git checkout main
git switch main
git stash
git branch x
git tag v1"
NEEDED="git diff
git diff HEAD
git status --short
git log -1
git show HEAD
vloop task show T1
printf x
cat .vloop/tmp/proposal.json
jq . .vloop/tmp/proposal.json
go test ./...
ls
grep -rn x ."
# review <log> <task> <name>: the checks every copilot review session passes.
review() {
  b=$(block "$1" copilot "/vloop:review $2")
  [ -n "$b" ] || fail "$3: no copilot session started with -p '/vloop:review $2'"
  [ "$(argafter "$b" -p)" = "/vloop:review $2" ] || fail "$3: -p is not followed by '/vloop:review $2'"
  [ "$(argafter "$b" --model)" = "$MODEL" ] || fail "$3: --model '$(argafter "$b" --model)', want $MODEL"
  [ "$(argafter "$b" --output-format)" = json ] || fail "$3: --output-format '$(argafter "$b" --output-format)', want json"
  for f in --no-ask-user --allow-all-tools --disable-builtin-mcps; do
    echo "$b" | grep -qx "arg $f" || fail "$3: no $f"
  done
  bad=$(echo "$b" | grep -E '^arg (--allow-all-paths|--add-dir|--sandbox|--no-sandbox|--allow-all|--yolo|--allow-all-urls|--allow-url|--config-dir)(=.*)?$') &&
    fail "$3: started with $bad"
  echo "$b" | grep -q '^arg --append-system-prompt-file' && fail "$3: given claude's --append-system-prompt-file"
  cw=$(block "$1" claude "/vloop:work $2")
  pd=$(argafter "$b" --plugin-dir)
  [ -n "$pd" ] && [ "$pd" = "$(argafter "$cw" --plugin-dir)" ] || fail "$3: --plugin-dir '$pd', the claude sessions' '$(argafter "$cw" --plugin-dir)'"
  echo "$b" | grep -qx 'plugin ok' || fail "$3: --plugin-dir holds no skills/review/SKILL.md when the session starts"
  echo "$b" | grep -qx "cwd $t/repo" || fail "$3: did not start in the repository root: $(echo "$b" | grep '^cwd ')"
  echo "$b" | grep -qx 'home empty' || fail "$3: COPILOT_HOME is $(echo "$b" | grep '^home ') when the session starts, want an empty directory"
  h=$(echo "$b" | sed -n 's/^env COPILOT_HOME=//p')
  case "$h" in "$t/repo/.vloop/tmp/"?*) ;; *) fail "$3: COPILOT_HOME '$h' is not under the repository's .vloop/tmp/";; esac
  for c in $(echo "$b" | sed -n 's/^env COPILOT_CUSTOM_INSTRUCTIONS_DIRS=//p' | tr ',:' '  '); do
    case "$c" in "$t/repo/.vloop/tmp/"?*) ;; *) fail "$3: COPILOT_CUSTOM_INSTRUCTIONS_DIRS names '$c', not under the repository's .vloop/tmp/";; esac
  done
  echo "$C2" | while IFS= read -r c; do
    echo "$b" | grep -qxF "probe denied shell $c" || { echo "GATE FAIL: $3: the deny patterns do not stop '$c' ($(echo "$b" | grep -F "probe allowed shell $c"))" >&2; exit 1; }
  done || exit 1
  echo "$NEEDED" | while IFS= read -r c; do
    echo "$b" | grep -qxF "probe allowed shell $c" || { echo "GATE FAIL: $3: a deny pattern stops '$c', which a review needs" >&2; exit 1; }
  done || exit 1
  echo "$b" | grep -qx 'probe denied url https://example.invalid/page' || fail "$3: the deny patterns do not stop the url tool"
  echo "$b" | grep -qx 'probe allowed write .vloop/tmp/verdict.json' || fail "$3: a deny pattern stops the file tools writing .vloop/tmp/verdict.json"
}
for k in T1 T2; do review "$L" "$k" "review $k"; done
# The environment: a claude review session's, plus the copilot variables.
for k in T1 T2; do
  envof "$(block "$t/cl.log" claude "/vloop:review $k")" | grep -Ev '^env COPILOT_(HOME|CUSTOM_INSTRUCTIONS_DIRS)=' >"$t/env.cl"
  envof "$(block "$L" copilot "/vloop:review $k")" | grep -Ev '^env COPILOT_(HOME|CUSTOM_INSTRUCTIONS_DIRS)=' >"$t/env.cop"
  [ -s "$t/env.cl" ] || fail "no environment logged for the all-claude review of $k"
  diff "$t/env.cl" "$t/env.cop" >"$t/env.diff" ||
    fail "review $k: the copilot session's environment, but for COPILOT_HOME and COPILOT_CUSTOM_INSTRUCTIONS_DIRS, differs from a claude review's (< claude, > copilot):
$(head -20 "$t/env.diff")"
done
[ -e "$t/home/.copilot" ] && fail "the run created ~/.copilot"
grep -qx 'guidelines absent' "$L" || fail "without .vloop/repo-guidelines.md a copilot session found the lesson"

# Guidelines: handed over through COPILOT_CUSTOM_INSTRUCTIONS_DIRS.
CFG=$COPILOT_CFG fixture plan.json guidelines; runfix "$t/gl" $CARRY
[ "$rc" = 0 ] || fail "the run with guidelines exited $rc: $(tail -8 "$t/gl")"
review "$t/gl.log" T1 "review T1 with guidelines"
b=$(block "$t/gl.log" copilot "/vloop:review T1")
echo "$b" | grep -qx 'guidelines found' ||
  fail "with .vloop/repo-guidelines.md, no directory in COPILOT_CUSTOM_INSTRUCTIONS_DIRS ('$(echo "$b" | sed -n 's/^env COPILOT_CUSTOM_INSTRUCTIONS_DIRS=//p')') holds them"

# Effort.
CFG="$COPILOT_CFG
effort.review high" fixture; runfix "$t/ef" $CARRY
[ "$rc" = 0 ] || fail "the run with effort.review high exited $rc: $(tail -8 "$t/ef")"
review "$t/ef.log" T1 "review T1 with an effort"
[ "$(argafter "$(block "$t/ef.log" copilot "/vloop:review T1")" --effort)" = high ] || fail "effort.review high: no --effort high"
block "$L" copilot "/vloop:review T1" | grep -q '^arg --effort' && fail "an --effort with effort.review unset"

# B18's comparison gate, unchanged.
sh "$G/b18/gate.sh" >"$t/b18.out" 2>&1 || fail "B18's comparison gate fails: $(tail -5 "$t/b18.out")"
echo "gate T2: ok"
