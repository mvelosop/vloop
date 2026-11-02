#!/bin/sh
# Gate T2 — R3's judge: the review-brief phase of `vloop fence check` and
# `vloop fence hook`, fence/review-brief.json, and the one deny rule every
# fence gains. The fences are judged as parsed JSON against the copies of the
# base fences in this folder: each of the five loop fences is its base copy
# plus the deny rule Bash(vloop brief review:*) and nothing else; the new
# fence is gate-review's plus Bash(vloop brief check:*) in its allow list and
# Bash(vloop brief review:*) in its deny list (rules compared as sets). Then
# every line of the brief's fence-check worked example, the work line, the
# changed unknown-phase message, the unchanged gate-review skill line, the
# review-brief rules beyond the example, every phase denying a nested brief
# review, the hook answering review-brief payloads, and both commands' help.
# Nothing runs a command under judgment: the gate only asks the judge, in a
# fixture root with a fake HOME, never this repository's tree.
set -u
fail() { echo "GATE FAIL: $*" >&2; exit 1; }
for v in $(env | sed -n 's/^\(VLOOP_[A-Z0-9_]*\)=.*/\1/p'); do unset "$v"; done
export LC_ALL=C GOTOOLCHAIN=local
G=$(cd "$(dirname "$0")" && pwd -P) || fail "no gate folder"
top=$(pwd -P)
JQ=$(command -v jq) || fail "jq is not on PATH"
t=$(mktemp -d "${TMPDIR:-/tmp}/gate.XXXXXX") && t=$(cd "$t" && pwd -P) || fail mktemp
trap 'rm -rf "$t"' EXIT

# The fences, as parsed JSON.
NR='Bash(vloop brief review:*)'
# same <file> <expected file>: equal but for rule order.
same() {
  "$JQ" -e -n --slurpfile a "$1" --slurpfile b "$2" '
    def norm: (.permissions.allow |= sort) | (.permissions.deny |= sort);
    ($a[0] | norm) == ($b[0] | norm)' >/dev/null 2>&1
}
for f in plan work review gate-review verify; do
  [ -f "$top/fence/$f.json" ] || fail "fence/$f.json is missing"
  "$JQ" --arg r "$NR" '.permissions.deny += [$r]' "$G/$f.json" >"$t/want-$f.json" || fail "jq"
  same "$top/fence/$f.json" "$t/want-$f.json" ||
    fail "fence/$f.json is not its base rules plus the one deny rule $NR: allow +$("$JQ" -c -n --slurpfile a "$top/fence/$f.json" --slurpfile b "$t/want-$f.json" '($a[0].permissions.allow - $b[0].permissions.allow)') -$("$JQ" -c -n --slurpfile a "$top/fence/$f.json" --slurpfile b "$t/want-$f.json" '($b[0].permissions.allow - $a[0].permissions.allow)'), deny +$("$JQ" -c -n --slurpfile a "$top/fence/$f.json" --slurpfile b "$t/want-$f.json" '($a[0].permissions.deny - $b[0].permissions.deny)') -$("$JQ" -c -n --slurpfile a "$top/fence/$f.json" --slurpfile b "$t/want-$f.json" '($b[0].permissions.deny - $a[0].permissions.deny)')"
done
[ -f "$top/fence/review-brief.json" ] || fail "fence/review-brief.json is missing"
"$JQ" --arg r "$NR" '.permissions.deny += [$r] | .permissions.allow += ["Bash(vloop brief check:*)"]' "$G/gate-review.json" >"$t/want-rb.json" || fail "jq"
same "$top/fence/review-brief.json" "$t/want-rb.json" ||
  fail "fence/review-brief.json is not the gate-review fence plus Bash(vloop brief check:*) allowed and $NR denied"

go build -buildvcs=false -o "$t/vloop" ./cmd/vloop || fail "go build of this tree"
S=$t/vloop
R=$t/repo
mkdir -p "$R/docs/briefs" "$t/home" "$t/tmp" || fail mkdir
git -C "$R" init -q || fail "git init of the fixture root"
B=docs/briefs/B20990101-0000-x.loop-brief.md
printf -- '---\nstatus: draft\n---\n# Brief\n' >"$R/$B"
cd "$R" || fail "fixture root"

ROOTARG=.
chk() {
  ph=$1 tl=$2; shift 2
  out=$(env -i PATH=/usr/bin:/bin HOME="$t/home" TMPDIR="$t/tmp" LC_ALL=C "$S" fence check --phase "$ph" --root "$ROOTARG" --tool "$tl" "$@" 2>"$t/err"); rc=$?
}
allow() {
  ph=$1 tl=$2; shift 2; chk "$ph" "$tl" "$@"
  [ "$rc" = 0 ] && [ "$out" = allow ] || fail "[$ph] $tl $*: want 'allow' and exit 0, got exit $rc, stdout '$out', stderr '$(cat "$t/err")'"
}
# denyx <phase> <tool> <exact reason> [args...]
denyx() {
  ph=$1 tl=$2 want=$3; shift 3; chk "$ph" "$tl" "$@"
  [ "$rc" = 1 ] && [ "$out" = "deny: $want" ] || fail "[$ph] $tl $*: want 'deny: $want' and exit 1, got exit $rc, stdout '$out', stderr '$(cat "$t/err")'"
}
P=review-brief
OW="is not allowed in a review-brief session: only .vloop/tmp/ may be written"

# The worked example, line for line.
allow $P bash 'vloop brief check .vloop/tmp/rb/B20990101-0000-x.loop-brief.md'
allow $P bash "cat $B"
allow $P create --path .vloop/tmp/brief-review.json
denyx $P create "write to $B $OW" --path "$B"
denyx $P bash 'git commit -m x is denied by the review-brief fence (Bash(git commit:*))' 'git commit -m x'
denyx $P bash "rm $B is denied by the Copilot deny list (rm:*)" "rm $B"
allow $P skill --path review-brief
denyx $P skill "skill vloop-review-brief is not vloop's — use /vloop:review-brief" --path vloop-review-brief
denyx $P bash "vloop brief review $B is denied by the review-brief fence (Bash(vloop brief review:*))" "vloop brief review $B"
denyx work bash "vloop brief review $B is denied by the work fence (Bash(vloop brief review:*))" "vloop brief review $B"
# The changed unknown-phase message.
chk deploy bash ls
[ "$rc" = 2 ] || fail "fence check --phase deploy: want exit 2, got $rc ('$out')"
grep -qxF 'vloop: unknown phase "deploy": want review, second-review, work, verify, gate-review, plan or review-brief' "$t/err" ||
  fail "fence check --phase deploy: want 'vloop: unknown phase \"deploy\": want review, second-review, work, verify, gate-review, plan or review-brief' on stderr, got '$(cat "$t/err")'"
# Unchanged.
denyx gate-review skill "skill vloop-review-brief is not vloop's — use /vloop:gate-review" --path vloop-review-brief

# Every phase denies a nested brief review through its own fence.
for ph in plan:plan review:review second-review:review verify:verify gate-review:gate-review work:work review-brief:review-brief; do
  denyx "${ph%%:*}" bash "vloop brief review x.loop-brief.md is denied by the ${ph#*:} fence (Bash(vloop brief review:*))" 'vloop brief review x.loop-brief.md'
  denyx "${ph%%:*}" bash "vloop brief review $B is denied by the ${ph#*:} fence (Bash(vloop brief review:*))" "env vloop brief review $B"
done

# review-brief beyond the example: decided as gate-review, with its own name.
allow $P bash 'vloop schema validate brief-review/v1 .vloop/tmp/brief-review.json'
allow $P bash "mkdir -p .vloop/tmp/rb && cp $B .vloop/tmp/rb/B20990101-0000-x.loop-brief.md"
allow $P bash 'git log --oneline -5'
allow $P bash 'vloop brief check docs/briefs/B20990101-0000-x.loop-brief.md'
allow $P view --path internal/x.go
denyx $P bash 'vloop run is denied by the review-brief fence (Bash(vloop run:*))' 'vloop run'
denyx $P bash 'vloop task verify T1 is denied by the review-brief fence (Bash(vloop task verify:*))' 'vloop task verify T1'
denyx $P bash 'vloop brief close x is denied by the review-brief fence (Bash(vloop brief close:*))' 'vloop brief close x'
denyx $P bash 'vloop brief new x is denied by the review-brief fence (Bash(vloop brief new:*))' 'vloop brief new x'
denyx $P bash 'gh pr list is denied by the Copilot deny list (gh:*)' 'sh -c "gh pr list"'
denyx $P bash 'rm .vloop/tmp/rb/x.md is denied by the Copilot deny list (rm:*)' 'rm .vloop/tmp/rb/x.md'
denyx $P bash "write to $B $OW" "echo x >> $B"
denyx $P create "write to src/a.go $OW" --path src/a.go
denyx $P edit "write to .vloop/state/state.json $OW" --path .vloop/state/state.json
denyx $P create "write to .vloop/state/brief-reviews/x.json $OW" --path .vloop/state/brief-reviews/x.json
denyx $P web_fetch 'tool web_fetch is not allowed in a review-brief session'
# Other phases word as before.
denyx gate-review create "write to $B is not allowed in a gate-review session: only .vloop/tmp/ may be written" --path "$B"
denyx review skill "skill vloop-plan is not vloop's — use /vloop:review" --path vloop-plan
allow work bash 'rm src/old.go'
denyx verify bash 'vloop run is denied by the verify fence (Bash(vloop run:*))' 'vloop run'

# The hook, phase review-brief.
H=$t/hookroot
mkdir -p "$H/docs/briefs" "$H/.vloop/tmp" || fail mkdir
cp "$R/$B" "$H/$B" || fail cp
snap() { (cd "$t" && find hookroot home -print | LC_ALL=C sort && cat "hookroot/$B"); }
before=$(snap)
hookrun() {
  hout=$(printf '%s' "$2" | (cd "$H" && env -i PATH=/usr/bin:/bin HOME="$t/home" TMPDIR="$t/tmp" LC_ALL=C "$S" fence hook --phase "$1" --root "$H") 2>"$t/err"); rc=$?
  [ "$(printf '%s\n' "$hout" | grep -c .)" = 1 ] || fail "[hook $1] payload $2: want one line on stdout, got '$hout'"
  dec=$(printf '%s\n' "$hout" | "$JQ" -r '.permissionDecision // "none"' 2>/dev/null) || fail "[hook $1] payload $2: stdout is not JSON: '$hout'"
  why=$(printf '%s\n' "$hout" | "$JQ" -r '.permissionDecisionReason // ""')
}
payload() { "$JQ" -cn --arg n "$1" --argjson a "$2" --arg c "$H" '{timestamp: 1700000000000, cwd: $c, toolName: $n, toolArgs: $a}'; }
bash_() { payload bash "$("$JQ" -cn --arg c "$1" '{command: $c, description: "d"}')"; }
file_() { payload "$1" "$("$JQ" -cn --arg p "$2" '{path: $p, file_text: "x"}')"; }
ok() {
  hookrun "$1" "$2"
  [ "$rc" = 0 ] && [ "$dec" = allow ] || fail "[hook $1] payload $2: want allow and exit 0, got '$hout', exit $rc"
}
no() {
  hookrun "$1" "$2"
  [ "$rc" = 0 ] && [ "$dec" = deny ] && [ "$why" = "$3" ] || fail "[hook $1] payload $2: want deny '$3' and exit 0, got '$hout', exit $rc"
}
ok $P "$(bash_ 'vloop brief check .vloop/tmp/rb/B20990101-0000-x.loop-brief.md')"
ok $P "$(bash_ "cat $B")"
ok $P "$(file_ create .vloop/tmp/brief-review.json)"
ok $P "$(file_ create "$H/.vloop/tmp/rb/x.md")"
no $P "$(file_ create "$B")" "vloop fence: write to $B $OW"
hookrun $P "$(file_ edit "$H/$B")"
[ "$rc" = 0 ] && [ "$dec" = deny ] || fail "[hook $P] an edit of the brief by its absolute path: want deny and exit 0, got '$hout', exit $rc"
no $P "$(bash_ 'git commit -m x')" 'vloop fence: git commit -m x is denied by the review-brief fence (Bash(git commit:*))'
no $P "$(bash_ "rm $B")" "vloop fence: rm $B is denied by the Copilot deny list (rm:*)"
no $P "$(bash_ "vloop brief review $B")" "vloop fence: vloop brief review $B is denied by the review-brief fence (Bash(vloop brief review:*))"
ok $P "$(payload skill '{"skill":"review-brief"}')"
no $P "$(payload skill '{"skill":"vloop-review-brief"}')" "vloop fence: skill vloop-review-brief is not vloop's — use /vloop:review-brief"
no $P "$(payload web_fetch '{"url":"https://example.invalid"}')" 'vloop fence: tool web_fetch is not allowed in a review-brief session'
no work "$(bash_ "vloop brief review $B")" "vloop fence: vloop brief review $B is denied by the work fence (Bash(vloop brief review:*))"
no gate-review "$(payload skill '{"skill":"vloop-review-brief"}')" "vloop fence: skill vloop-review-brief is not vloop's — use /vloop:gate-review"
for ph in deploy bogus; do
  hookrun "$ph" "$(bash_ 'ls')"
  [ "$rc" = 2 ] && [ "$dec" = deny ] || fail "fence hook --phase $ph: want the deny line and exit 2, got '$hout', exit $rc"
  printf '%s' "$why" | grep -qF "unknown phase \"$ph\": want review, second-review, work, verify, gate-review, plan or review-brief" ||
    fail "fence hook --phase $ph: the reason '$why' does not hold the new unknown-phase message"
done
[ "$(snap)" = "$before" ] || fail "the hook changed the fixture root or the HOME"

# Both commands' help name the phase: usage line and --phase flag.
for c in check hook; do
  "$S" fence "$c" --help >"$t/help" 2>&1 || fail "vloop fence $c --help failed"
  grep -E "^[[:space:]]*vloop fence $c --phase <[^>]*>" "$t/help" | grep -qF 'review-brief' ||
    fail "vloop fence $c --help: the usage line does not name review-brief: $(grep -F "vloop fence $c --phase" "$t/help")"
  grep -E '^[[:space:]]*--phase ' "$t/help" | grep -qF 'review-brief' ||
    fail "vloop fence $c --help: the --phase flag does not name review-brief: $(grep -F -- '--phase ' "$t/help")"
done
echo "gate T2: ok"
