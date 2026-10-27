#!/bin/sh
# Gate T2 — H1, `vloop fence hook`. Copilot preToolUse payloads (camelCase:
# toolName, toolArgs, cwd) on stdin, in phases review and second-review: one
# line of JSON on stdout, permissionDecision allow or deny with a reason
# "vloop fence: …", exit 0; a payload it cannot read or an unknown phase is
# the deny line and exit 2. The hook writes nothing: the fixture root and the
# fake HOME are the same before and after.
G=$(cd "$(dirname "$0")" && pwd -P) || { echo "GATE FAIL: no gate folder" >&2; exit 1; }
. "$G/lib.sh"
cd "$top" || fail "repo root"
newbin
R=$t/repo
mkdir -p "$R/src" "$R/.vloop/tmp" "$t/home" "$t/tmp" "$t/stub" || fail mkdir
echo x >"$R/src/app.ts"
snap() { (cd "$t" && find repo home -print | LC_ALL=C sort && cat repo/src/app.ts); }
before=$(snap)
# hookrun <phase> <payload>: $out the hook's stdout, $rc its exit, $dec and
# $why the decision and reason parsed from its one line.
hookrun() {
  out=$(printf '%s' "$2" | (cd "$R" && venv "$S" fence hook --phase "$1" --root "$R") 2>"$t/err"); rc=$?
  lines=$(printf '%s\n' "$out" | grep -c .)
  [ "$lines" = 1 ] || fail "[$1] payload $2: want one line on stdout, got $lines: '$out'"
  dec=$(printf '%s\n' "$out" | "$JQ" -r '.permissionDecision // "none"' 2>/dev/null) || fail "[$1] payload $2: stdout is not JSON: '$out'"
  why=$(printf '%s\n' "$out" | "$JQ" -r '.permissionDecisionReason // ""')
}
# payload <tool> <toolArgs JSON>
payload() { "$JQ" -cn --arg n "$1" --argjson a "$2" --arg c "$R" '{timestamp: 1700000000000, cwd: $c, toolName: $n, toolArgs: $a}'; }
ok() {
  hookrun "$1" "$2"
  [ "$rc" = 0 ] && [ "$dec" = allow ] || fail "[$1] payload $2: want allow and exit 0, got '$out', exit $rc"
}
no() {
  hookrun "$1" "$2"
  [ "$rc" = 0 ] && [ "$dec" = deny ] || fail "[$1] payload $2: want deny and exit 0, got '$out', exit $rc"
  case "$why" in "vloop fence: "?*) ;; *) fail "[$1] payload $2: the reason '$why' does not start 'vloop fence: '";; esac
}
bash_() { payload "$1" "$("$JQ" -cn --arg c "$2" '{command: $c, description: "d"}')"; }
file_() { payload "$1" "$("$JQ" -cn --arg p "$2" '{path: $p, file_text: "x"}')"; }
# The brief and the experiments name no key for apply_patch's patch, so each
# patch goes in three shapes — toolArgs the raw patch string, {input: …} and
# {patch: …} — and every shape must be denied: by its paths when the hook finds
# the patch there, and fail-closed (the deny line, exit 0 or 2) when it does not.
patchno() {
  for a in "$("$JQ" -cn --arg p "$2" '$p')" "$("$JQ" -cn --arg p "$2" '{input: $p}')" "$("$JQ" -cn --arg p "$2" '{patch: $p}')"; do
    hookrun "$1" "$(payload apply_patch "$a")"
    [ "$dec" = deny ] && { [ "$rc" = 0 ] || [ "$rc" = 2 ]; } || fail "[$1] apply_patch with toolArgs $a: want deny, got '$out', exit $rc"
    case "$why" in "vloop fence: "?*) ;; *) fail "[$1] apply_patch with toolArgs $a: the reason '$why' does not start 'vloop fence: '";; esac
  done
}

for ph in review second-review; do
  ok "$ph" "$(bash_ bash 'git diff HEAD')"
  ok "$ph" "$(bash_ bash 'git log --oneline | head -5')"
  ok "$ph" "$(bash_ bash 'git diff > .vloop/tmp/diff.txt')"
  no "$ph" "$(bash_ bash 'sh -c "gh pr list"')"
  no "$ph" "$(bash_ bash 'env gh auth token')"
  no "$ph" "$(bash_ bash 'nohup timeout 5 wget x')"
  no "$ph" "$(bash_ bash 'echo $(gh auth token)')"
  no "$ph" "$(bash_ bash 'git diff > .vloop/state/x')"
  no "$ph" "$(bash_ powershell 'env gh auth token')"
  ok "$ph" "$(bash_ powershell 'git status')"
  ok "$ph" "$(file_ create .vloop/tmp/verdict.json)"
  ok "$ph" "$(file_ create "$R/.vloop/tmp/abs.json")"
  no "$ph" "$(file_ create .vloop/state/state.json)"
  no "$ph" "$(file_ edit src/app.ts)"
  no "$ph" "$(file_ edit "$R/.git/config")"
  ok "$ph" "$(file_ view src/app.ts)"
  no "$ph" "$(file_ view "$t/home/.claude/settings.json")"
  no "$ph" "$(file_ view "$t/home/.copilot/config.json")"
  no "$ph" "$(payload web_fetch '{"url":"https://example.com"}')"
  no "$ph" "$(payload search_code_subagent '{"query":"x"}')"
  patchno "$ph" "$(printf '*** Begin Patch\n*** Update File: .git/config\n@@\n-a\n+b\n*** End Patch\n')"
  patchno "$ph" "$(printf '*** Begin Patch\n*** Add File: .vloop/tmp/a.txt\n+a\n*** Update File: src/app.ts\n@@\n-x\n+y\n*** End Patch\n')"
  patchno "$ph" "$(printf '*** Begin Patch\n*** Update File: .vloop/tmp/a.txt\n*** Move to: src/a.txt\n@@\n-x\n+y\n*** End Patch\n')"
  patchno "$ph" 'no patch lines at all'
  hookrun "$ph" '{"toolName":"bash","cwd":"x"}'
  [ "$dec" = deny ] && { [ "$rc" = 0 ] || [ "$rc" = 2 ]; } || fail "[$ph] a bash payload with no toolArgs: want deny, got '$out', exit $rc"
  no "$ph" "$(payload frobnicate '{}')"
  [ "$why" = "vloop fence: tool frobnicate is not allowed in a $ph session" ] ||
    fail "[$ph] frobnicate: the reason is '$why', want 'vloop fence: tool frobnicate is not allowed in a $ph session'"
done

# Unreadable payloads and an unknown phase: the deny line, exit 2.
bad() {
  hookrun "$1" "$2"
  [ "$rc" = 2 ] && [ "$dec" = deny ] || fail "[$1] $3: want the deny line and exit 2, got '$out', exit $rc"
  case "$why" in "vloop fence: "?*) ;; *) fail "[$1] $3: the reason '$why' does not start 'vloop fence: '";; esac
}
bad review 'not json' 'a payload that is not JSON'
bad review '{}' 'a payload with no toolName'
bad review '{"toolArgs":{"command":"git status"}}' 'a payload with no toolName'
bad work "$(bash_ bash 'git status')" 'an unknown phase'
bad bogus "$(bash_ bash 'git status')" 'an unknown phase'

[ "$(snap)" = "$before" ] || fail "the hook changed the fixture root or the HOME"
echo "gate T2: ok"
