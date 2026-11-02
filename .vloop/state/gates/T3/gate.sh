#!/bin/sh
# Gate T3 — R1: `vloop brief review <brief>` on Claude, with the stub claude
# (stub-claude.sh) in fixture repositories (lib.sh). The keys
# agent/model/effort.review-brief; the command's help; every line of the
# worked example on Claude — the FIX run (stderr's reviewing line, the verdict
# block on stdout, the kept report, the folder, its one session/v5 record, the
# stub's prompt and the review-brief fence's settings), the PASS run — and
# every planted failure with the stub claude, each with its exact line and
# exit code; plus the lock taken and released, .vloop/tmp/ cleared before the
# session, the preflight, the model and effort resolution, and that the
# command commits nothing and writes only its folder and .vloop/tmp/.
G=$(cd "$(dirname "$0")" && pwd -P) || { echo "GATE FAIL: no gate folder" >&2; exit 1; }
. "$G/lib.sh"
cd "$top" || fail "repo root"
newbin

# The keys.
CFG=; fixture
get() { k=$1; shift; (cd "$t/repo" && venv "$@" "$S" config get "$k" 2>&1); }
[ "$(get agent.review-brief)" = claude ] || fail "agent.review-brief reads '$(get agent.review-brief)', want the default claude"
[ "$(get model.review-brief)" = opus ] || fail "model.review-brief reads '$(get model.review-brief)', want the default opus"
(cd "$t/repo" && venv "$S" config get effort.review-brief >"$t/eff" 2>&1) || fail "vloop config get effort.review-brief failed: $(cat "$t/eff")"
[ -z "$(cat "$t/eff")" ] || fail "effort.review-brief reads '$(cat "$t/eff")', want no default"
[ "$(get model.review-brief VLOOP_MODEL_REVIEW_BRIEF=sonnet)" = sonnet ] || fail "VLOOP_MODEL_REVIEW_BRIEF does not set model.review-brief"
[ "$(get effort.review-brief VLOOP_EFFORT_REVIEW_BRIEF=high)" = high ] || fail "VLOOP_EFFORT_REVIEW_BRIEF does not set effort.review-brief"
(cd "$t/repo" && venv "$S" config set agent.review-brief codex >"$t/set" 2>&1); c=$?
[ "$c" = 2 ] || fail "vloop config set agent.review-brief codex exited $c, want 2: $(cat "$t/set")"
(cd "$t/repo" && venv "$S" config set agent.review-brief copilot >"$t/set" 2>&1) || fail "vloop config set agent.review-brief copilot failed: $(cat "$t/set")"
[ "$(get agent.review-brief)" = copilot ] || fail "agent.review-brief set to copilot reads '$(get agent.review-brief)'"

# The command and its help.
(cd "$t/repo" && venv "$S" brief --help >"$t/h" 2>&1) || fail "vloop brief --help failed"
grep -qE '^[[:space:]]+review[[:space:]]' "$t/h" || fail "vloop brief --help does not list review: $(cat "$t/h")"
(cd "$t/repo" && venv "$S" brief review --help >"$t/h" 2>&1) || fail "vloop brief review --help failed: $(cat "$t/h")"
grep -qF 'vloop brief review' "$t/h" || fail "vloop brief review --help has no usage line: $(cat "$t/h")"
for c in configuration commands; do
  grep -qF "$( [ $c = commands ] && echo 'vloop brief review' || echo 'agent.review-brief')" "$top/docs/guide/$c.md" || fail "docs/guide/$c.md does not name $( [ $c = commands ] && echo 'vloop brief review' || echo 'agent.review-brief')"
done
for k in model.review-brief effort.review-brief; do grep -qF "$k" "$top/docs/guide/configuration.md" || fail "docs/guide/configuration.md does not name $k"; done

# FIX, two findings: the worked example's first run.
CFG=; fixture
report "$REPORT_FIX"
mkdir -p "$t/repo/.vloop/tmp" && echo stale >"$t/repo/.vloop/tmp/stale.txt"
rev "$t/fix"
[ "$rc" = 1 ] || fail "a FIX report: exit $rc, want 1: $(cat "$t/fix.err")"
folder "$t/fix" "a FIX report"
errline "$t/fix" "reviewing $B using opus (agent claude)" || fail "a FIX report: no 'reviewing $B using opus (agent claude)' on stderr: $(cat "$t/fix.err")"
{ block FIX "$t/stub/report.json"; echo "kept: $F/brief-review.json"; } >"$t/fix.want"
cmp -s "$t/fix.want" "$t/fix.out" || fail "a FIX report: stdout is not the verdict block:
--- want
$(cat "$t/fix.want")
--- got
$(cat "$t/fix.out")"
grep -q '^vloop:' "$t/fix.err" && fail "a FIX report: a vloop: line on stderr: $(grep '^vloop:' "$t/fix.err")"
grep -qF 'reviewing ' "$t/fix.out" && fail "a FIX report: the reviewing line is on stdout"
[ "$("$JQ" -S . "$t/repo/$F/brief-review.json" 2>/dev/null)" = "$("$JQ" -S . "$t/stub/report.json")" ] || fail "a FIX report: $F/brief-review.json is not the session's report: $(cat "$t/repo/$F/brief-review.json" 2>&1)"
record "a FIX report" claude opus
[ "$(ev "$t/fix" '^session ')" = "session review-brief $B model=opus effort=-" ] || fail "a FIX report: want one session of /vloop:review-brief $B with model opus and no effort, got: $(ev "$t/fix" '^session ')"
[ "$(ev "$t/fix" '^fence ')" = "fence review-brief ok" ] || fail "a FIX report: the session's --settings is not the review-brief fence (gate-review's plus Bash(vloop brief check:*) allowed, Bash(vloop brief review:*) denied): $(ev "$t/fix" '^fence ')"
[ "$(ev "$t/fix" '^stale ')" = "stale review-brief no" ] || fail "a FIX report: .vloop/tmp/ was not cleared before the session"
[ "$(ev "$t/fix" '^lock ')" = "lock review-brief yes" ] || fail "a FIX report: the working tree's lock was not held during the session"
[ -e "$t/repo/.vloop/tmp/.running" ] && fail "a FIX report: the lock .vloop/tmp/.running was not released"
refsame "$t/fix" "a FIX report"
briefsame "a FIX report"
st=$(venv git -C "$t/repo" status --porcelain --untracked-files=all | grep -vE "^\?\? $F/")
[ -z "$st" ] || fail "a FIX report: the command wrote outside its folder and .vloop/tmp/: $st"
[ -e "$t/repo/.vloop/state/state.json" ] || [ -e "$t/repo/.vloop/state/runs" ] && fail "a FIX report: the command wrote plan or run state"

# PASS, no findings.
CFG=; fixture
report "$REPORT_PASS"
rev "$t/pass"
[ "$rc" = 0 ] || fail "a PASS report: exit $rc, want 0: $(cat "$t/pass.err")"
folder "$t/pass" "a PASS report"
printf 'verdict: PASS — 0 finding(s)\nkept: %s/brief-review.json\n' "$F" >"$t/pass.want"
cmp -s "$t/pass.want" "$t/pass.out" || fail "a PASS report: stdout is '$(cat "$t/pass.out")', want '$(cat "$t/pass.want")'"
record "a PASS report" claude opus

# An ASK report exits 1 and prints its verdict.
CFG=; fixture
report "$(printf '%s' "$REPORT_PASS" | "$JQ" -c '.verdict = "ASK" | .findings = [{"kind":"decision","section":"Why this shape","summary":"the model default was not approved"}]')"
rev "$t/ask"
[ "$rc" = 1 ] || fail "an ASK report: exit $rc, want 1: $(cat "$t/ask.err")"
folder "$t/ask" "an ASK report"
{ block ASK "$t/stub/report.json"; echo "kept: $F/brief-review.json"; } >"$t/ask.want"
cmp -s "$t/ask.want" "$t/ask.out" || fail "an ASK report: stdout is '$(cat "$t/ask.out")', want '$(cat "$t/ask.want")'"

# The model and effort resolve by P-6.
CFG="model.review-brief haiku
effort.review-brief low"; fixture
report "$REPORT_PASS"
rev "$t/cfgm"
[ "$rc" = 0 ] || fail "model.review-brief haiku: exit $rc: $(cat "$t/cfgm.err")"
errline "$t/cfgm" "reviewing $B using haiku (agent claude)" || fail "model.review-brief haiku: no 'reviewing $B using haiku (agent claude)': $(cat "$t/cfgm.err")"
[ "$(ev "$t/cfgm" '^session ')" = "session review-brief $B model=haiku effort=low" ] || fail "model.review-brief haiku, effort low: the session got $(ev "$t/cfgm" '^session ')"
record "model.review-brief haiku" claude haiku
CFG=; fixture
report "$REPORT_PASS"
rev "$t/envm" "$B" VLOOP_MODEL_REVIEW_BRIEF=sonnet VLOOP_EFFORT_REVIEW_BRIEF=high
[ "$rc" = 0 ] || fail "VLOOP_MODEL_REVIEW_BRIEF=sonnet: exit $rc: $(cat "$t/envm.err")"
errline "$t/envm" "reviewing $B using sonnet (agent claude)" || fail "VLOOP_MODEL_REVIEW_BRIEF=sonnet: no 'reviewing $B using sonnet (agent claude)': $(cat "$t/envm.err")"
[ "$(ev "$t/envm" '^session ')" = "session review-brief $B model=sonnet effort=high" ] || fail "VLOOP_*_REVIEW_BRIEF: the session got $(ev "$t/envm" '^session ')"

# Refusals before any session.
CFG=; fixture
report "$REPORT_PASS"
rev "$t/nb" notes.md
[ "$rc" = 2 ] || fail "notes.md: exit $rc, want 2: $(cat "$t/nb.err")"
endswith "$t/nb" 'vloop: not a loop brief: notes.md' "notes.md"
noreview "$t/nb" "notes.md"
rev "$t/ab" docs/briefs/B20990101-0001-y.loop-brief.md
[ "$rc" = 1 ] || fail "an absent brief: exit $rc, want 1: $(cat "$t/ab.err")"
endswith "$t/ab" 'vloop: no such file: docs/briefs/B20990101-0001-y.loop-brief.md' "an absent brief"
noreview "$t/ab" "an absent brief"
(cd "$t/repo" && venv "$S" brief review >"$t/na.out" 2>&1); c=$?
[ "$c" = 2 ] || fail "vloop brief review with no argument exited $c, want 2"
(cd "$t/repo" && venv "$S" brief review "$B" "$B" >"$t/na.out" 2>&1); c=$?
[ "$c" = 2 ] || fail "vloop brief review with two arguments exited $c, want 2"
mkdir -p "$t/plain/docs/briefs" && cp "$G/brief.md" "$t/plain/$B" && venv git -C "$t/plain" init -q || fail "plain repo"
(cd "$t/plain" && venv "$S" brief review "$B" >"$t/pl.out" 2>"$t/pl.err"); c=$?
[ "$c" = 1 ] && grep -qxF 'vloop: not in a vloop repository — run vloop init first' "$t/pl.err" ||
  fail "outside a vloop repository: want 'vloop: not in a vloop repository — run vloop init first' and exit 1, got exit $c: $(cat "$t/pl.err")"
[ -s "$t/stub/events" ] && fail "outside a vloop repository: a session was started"
# A live loop holds the lock.
sleep 300 >/dev/null 2>&1 & SPID=$!
printf '{"pid":%d,"branch":"main","started":"2099-01-01T00:00:00Z","run":"other"}\n' "$SPID" >"$t/repo/.vloop/tmp/.running"
cp "$t/repo/.vloop/tmp/.running" "$t/lock.orig"
rev "$t/lk"
[ "$rc" = 1 ] || fail "a live loop: exit $rc, want 1: $(cat "$t/lk.err")"
sed 's/^[[:space:]]*//' "$t/lk.err" | grep -qF "vloop: a loop is already running in this working tree (pid $SPID, branch 'main', started 2099-01-01T00:00:00Z) — " ||
  fail "a live loop: want vloop run's refusal 'vloop: a loop is already running in this working tree (pid $SPID, …', got: $(cat "$t/lk.err")"
noreview "$t/lk" "a live loop"
cmp -s "$t/lock.orig" "$t/repo/.vloop/tmp/.running" || fail "a live loop: its lock was changed or removed"
kill "$SPID" 2>/dev/null; SPID=
# The preflight: claude missing, the folder untrusted.
CFG=; fixture
report "$REPORT_PASS"
rm -f "$t/stub/claude"
rev "$t/nc"
[ "$rc" = 1 ] || fail "claude missing: exit $rc, want 1: $(cat "$t/nc.err")"
grep -qxF '  ✗ claude claude is not on PATH' "$t/nc.err" || fail "claude missing: want '  ✗ claude claude is not on PATH' on stderr: $(cat "$t/nc.err")"
grep -qxF 'vloop: preflight failed — 1 problem(s), nothing has run' "$t/nc.err" || fail "claude missing: want 'vloop: preflight failed — 1 problem(s), nothing has run': $(cat "$t/nc.err")"
[ -e "$t/repo/.vloop/state/brief-reviews" ] && fail "claude missing: a review folder was made"
CFG=; fixture
report "$REPORT_PASS"
rm -f "$t/home/.claude.json"
rev "$t/nt"
[ "$rc" = 1 ] || fail "an untrusted folder: exit $rc, want 1: $(cat "$t/nt.err")"
grep -q '^  ✗ trust ' "$t/nt.err" || fail "an untrusted folder: want a '  ✗ trust …' line on stderr: $(cat "$t/nt.err")"
grep -qxF 'vloop: preflight failed — 1 problem(s), nothing has run' "$t/nt.err" || fail "an untrusted folder: want 'vloop: preflight failed — 1 problem(s), nothing has run': $(cat "$t/nt.err")"
noreview "$t/nt" "an untrusted folder"

# Planted failures after the session; each line names the folder.
# planted <label> <stub flag file|-> <report json|-> <exit> <line, with @F for the folder>
planted() {
  CFG=; fixture
  [ "$3" = - ] || report "$3"
  [ "$2" = - ] || : >"$t/stub/$2"
  rev "$t/$1"
  [ "$rc" = "$4" ] || fail "$1: exit $rc, want $4: $(cat "$t/$1.err")"
  folder "$t/$1" "$1"
  want=$(printf '%s' "$5" | sed "s#@F#$F#")
  endswith "$t/$1" "$want" "$1"
  grep -c '^vloop:' "$t/$1.err" | grep -qx 1 || fail "$1: want one vloop: line, got: $(grep '^vloop:' "$t/$1.err")"
  briefsame "$1"
  [ -e "$t/repo/.vloop/tmp/.running" ] && fail "$1: the lock was not released"
  [ "$(ev "$t/$1" '^session ' | grep -c .)" = 1 ] || fail "$1: want one session, got: $(ev "$t/$1" '^session ')"
  return 0
}
SCH="vloop: the brief review's .vloop/tmp/brief-review.json does not follow brief-review/v1 — see @F"
planted noreport - - 7 'vloop: the brief review left no .vloop/tmp/brief-review.json — see @F'
[ -e "$t/repo/$F/brief-review.json" ] && fail "noreport: a report was kept"
record "noreport" claude opus
planted verdictok - "$(printf '%s' "$REPORT_FIX" | "$JQ" -c '.verdict = "OK"')" 7 "$SCH"
[ -e "$t/repo/$F/brief-review.json" ] && fail "verdict OK: an invalid report was kept as $F/brief-review.json"
planted fixempty - "$(printf '%s' "$REPORT_FIX" | "$JQ" -c '.findings = []')" 7 "$SCH"
[ -e "$t/repo/$F/brief-review.json" ] && fail "FIX with no findings: an invalid report was kept"
planted othername - "$(printf '%s' "$REPORT_FIX" | "$JQ" -c '.brief = "B20990101-0001-y.loop-brief"')" 7 "$SCH"
[ -e "$t/repo/$F/brief-review.json" ] && fail "a report naming another brief was kept"
planted exit3 claude.exit3 "$REPORT_FIX" 7 'vloop: brief review session failed (claude exited 3) — see @F'
planted gitcfg claude.gitconfig "$REPORT_FIX" 9 'vloop: the brief review changed .git/config (gate.planted) — nothing was committed; restore it, then re-run'
planted briefw claude.brief "$REPORT_FIX" 9 "vloop: the brief review changed $B; it was restored"
planted moveref claude.moveref "$REPORT_FIX" 9 'vloop: REFS MOVED review-brief — the brief review changed git refs; nothing was committed'
# Refs moved and the brief changed: the refs line, the brief restored.
CFG=; fixture
report "$REPORT_FIX"; : >"$t/stub/claude.moveref"; : >"$t/stub/claude.brief"
rev "$t/both"
[ "$rc" = 9 ] || fail "refs moved and the brief changed: exit $rc, want 9: $(cat "$t/both.err")"
endswith "$t/both" 'vloop: REFS MOVED review-brief — the brief review changed git refs; nothing was committed' "refs moved and the brief changed"
briefsame "refs moved and the brief changed"
echo "gate T3: ok"
