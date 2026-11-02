#!/bin/sh
# Gate T2 — V1, the driver and the close. With the stub claude (stub-claude.sh)
# in fixture repositories (lib.sh) whose one-task run completes, every line of
# the brief's worked example and each planted failure: what the verify step
# prints (stdout's "verify: " line, stderr's warnings), what it keeps as
# <run dir>reports/verify.json, and that the run still exits 0 with its plan
# complete; then `vloop brief close <brief>` with no flag on the kept reports
# (exit 2, the verify block on stdout), and on a kept verify/v1 report the
# base's lines exactly.
G=$(cd "$(dirname "$0")" && pwd -P) || { echo "GATE FAIL: no gate folder" >&2; exit 1; }
. "$G/lib.sh"
cd "$top" || fail "repo root"
newbin

CK='{"example_lines":2,"sentences":3,"surfaces":1}'
F1='{"kind":"example","claim":"`demo` exits 0","observed":"exit 1","where":"Worked example, line 1; bin/demo"}'
F2='{"kind":"surface","claim":"README.md:3: \"demo ignores its arguments\"","observed":"demo reads its first argument","where":"Behaviour contract; README.md:3"}'
N1='{"line":"vloop init","reason":"denied by the verify fence (Bash(vloop init:*))"}'
N2='{"line":"vloop brief review docs/briefs/x.loop-brief.md","reason":"the operator'"'"'s, at close"}'
# v2 <complete> <findings> <not_run>: a verify/v2 report.
v2() { printf '{"schema":"verify/v2","complete":%s,"findings":%s,"not_run":%s,"checked":%s,"notes":"none"}' "$1" "$2" "$3" "$CK"; }
FBLOCK1='  1. example — Worked example, line 1; bin/demo
     brief:    `demo` exits 0
     observed: exit 1'
FBLOCK2='  2. surface — Behaviour contract; README.md:3
     brief:    README.md:3: "demo ignores its arguments"
     observed: demo reads its first argument'
NOTRUN1='  not run: vloop init — denied by the verify fence (Bash(vloop init:*))'
NOTRUN2="  not run: vloop brief review docs/briefs/x.loop-brief.md — the operator's, at close"
STOPPED='  the session stopped before its last pass'
REFUSE='say what you found before merge'
# closewant <name> <expected stdout>: a bare close refuses with exit 2 and
# prints exactly the expected verify block on stdout.
closewant() {
  closebare "$1"
  [ "$crc" = 2 ] || fail "$1: vloop brief close with no flag exited $crc, want 2: $(cat "$t/$1.cerr")"
  grep -qF "$REFUSE" "$t/$1.cerr" || fail "$1: vloop brief close with no flag did not refuse: $(cat "$t/$1.cerr")"
  printf '%s\n' "$2" >"$t/$1.cwant"
  cmp -s "$t/$1.cwant" "$t/$1.cout" || fail "$1: vloop brief close prints:
$(cat "$t/$1.cout")
--- want
$(cat "$t/$1.cwant")"
}
stamp() { basename "$RD"; }

# 1. complete, no findings, nothing not run: the base's line.
runcase e1 "$(v2 true '[]' '[]')"
completed e1; verifylines e1 1; outis e1 'verify: no findings'
nowhere e1 'not run'; nowhere e1 'stopped before its last pass'; nowhere e1 'verify by hand'; kept e1

# 2. complete, one finding: the base's line.
runcase e2 "$(v2 true "[$F1]" '[]')"
completed e2; verifylines e2 1; outis e2 "verify: 1 finding(s) — see ${RD}reports/verify.json"
nowhere e2 'not run'; nowhere e2 'stopped before its last pass'; kept e2
# A kept verify/v1 report prints exactly the base's lines at close.
printf '%s\n' "{\"schema\":\"verify/v1\",\"findings\":[$F1],\"checked\":$CK,\"notes\":\"none\"}" >"$t/repo/${RD}reports/verify.json"
closewant e2v1 "verify of $RID/$(stamp): 1 finding(s)
$FBLOCK1"
printf '%s\n' "{\"schema\":\"verify/v1\",\"findings\":[],\"checked\":$CK,\"notes\":\"none\"}" >"$t/repo/${RD}reports/verify.json"
closewant e2v1e "verify of $RID/$(stamp): no findings"

# 3. complete, no findings, one line not run.
runcase e3 "$(v2 true '[]' "[$N1]")"
completed e3; verifylines e3 1; outis e3 "verify: no findings; 1 example line(s) not run — see ${RD}reports/verify.json"
nowhere e3 'stopped before its last pass'; nowhere e3 'verify by hand'; kept e3
closewant e3 "verify of $RID/$(stamp): no findings; 1 example line(s) not run
$NOTRUN1"

# 4. complete, two findings, two lines not run.
runcase e4 "$(v2 true "[$F1,$F2]" "[$N1,$N2]")"
completed e4; verifylines e4 1; outis e4 "verify: 2 finding(s); 2 example line(s) not run — see ${RD}reports/verify.json"
nowhere e4 'stopped before its last pass'; kept e4
closewant e4 "verify of $RID/$(stamp): 2 finding(s); 2 example line(s) not run
$FBLOCK1
$FBLOCK2
$NOTRUN1
$NOTRUN2"

# 5. incomplete, exit 0: the line, then the stopped-early warning.
runcase e5 "$(v2 false '[]' '[]')"
completed e5; verifylines e5 1; outis e5 'verify: no findings'
erris e5 "vloop: the verify session stopped before its last pass — verify the rest by hand; see ${RD}reports/verify.json"
nowhere e5 'failed to run'; nowhere e5 'not run'; kept e5
closewant e5 "verify of $RID/$(stamp): no findings
$STOPPED"

# 6. incomplete, one finding, then the session exits 3: the line and the
# base's failed-to-run warning, not the stopped-early one; the report kept.
runcase e6 "$(v2 false "[$F1]" '[]')" claude.verify-exit3
completed e6; verifylines e6 1; outis e6 "verify: 1 finding(s) — see ${RD}reports/verify.json"
erris e6 "vloop: the verify session failed to run — verify by hand; see ${RD}run.log"
nowhere e6 'stopped before its last pass'; nowhere e6 'left no valid'; kept e6
closewant e6 "verify of $RID/$(stamp): 1 finding(s)
$FBLOCK1
$STOPPED"

# 6b. the same report after a session that left no record (nothing on stdout,
# exit 0): the same two lines.
runcase e6n "$(v2 false "[$F1]" '[]')" claude.verify-norecord
completed e6n; verifylines e6n 1; outis e6n "verify: 1 finding(s) — see ${RD}reports/verify.json"
erris e6n "vloop: the verify session failed to run — verify by hand; see ${RD}run.log"
nowhere e6n 'stopped before its last pass'; nowhere e6n 'left no valid'; kept e6n

# Planted failures.
LEFT='vloop: the verify session left no valid .vloop/tmp/verify.json — verify by hand'
# planted <name> <report> [flag]: exit 0, the left-no-valid line, no
# "verify: " line, nothing kept.
planted() {
  runcase "$@"
  completed "$1"; verifylines "$1" 0; erris "$1" "$LEFT"; nowhere "$1" 'failed to run'; notkept "$1"
}
planted p1 "{\"schema\":\"verify/v1\",\"findings\":[],\"checked\":$CK,\"notes\":\"none\"}"
planted p2 '{"schema":"verify/v2","findings":[],"not_run":[],"checked":'"$CK"',"notes":"none"}'
planted p3 "$(v2 true '[]' '[{"line":"vloop init"}]')"
# No report, then exit 3: the failed-to-run warning alone, nothing kept.
runcase p4 - claude.verify-exit3
completed p4; verifylines p4 0
erris p4 "vloop: the verify session failed to run — verify by hand; see ${RD}run.log"
nowhere p4 'left no valid'; notkept p4
# An invalid report after a failed session: the failed-to-run warning alone.
runcase p5 "{\"schema\":\"verify/v1\",\"findings\":[],\"checked\":$CK,\"notes\":\"none\"}" claude.verify-exit3
completed p5; verifylines p5 0
erris p5 "vloop: the verify session failed to run — verify by hand; see ${RD}run.log"
nowhere p5 'left no valid'; notkept p5
echo "gate T2: ok"
