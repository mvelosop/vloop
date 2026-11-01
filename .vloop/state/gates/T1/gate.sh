#!/bin/sh
# Gate T1 — G1, the judge's gate-review phase, through `vloop fence check` and
# `vloop fence hook` (they share the judge). Every line of the worked example's
# two tables and its apply_patch lines in phase gate-review, the unchanged
# lines (review's skill hint, verify's rm, work's rm, agent.plan's config
# refusal), the changed unknown-phase message, both commands' help naming
# gate-review, and the contract's gate-review rules beyond the example (as in
# review: writes only under .vloop/tmp/, the whole deny list with rm:*, reads,
# skills, other tools, the shell syntax that stays denied). Then the hook:
# each Copilot payload answered with the same decision and the reason
# `vloop fence: <message>`, exit 0; an unknown phase the deny line and exit 2.
# A few review, second-review, work and verify lines pin that those phases
# decide and word as before. Nothing runs a command under judgment: the gate
# only asks the judge, in a fixture root with a fake HOME, never this
# repository's tree.
set -u
fail() { echo "GATE FAIL: $*" >&2; exit 1; }
for v in $(env | sed -n 's/^\(VLOOP_[A-Z0-9_]*\)=.*/\1/p'); do unset "$v"; done
export LC_ALL=C GOTOOLCHAIN=local
G=$(cd "$(dirname "$0")" && pwd -P) || fail "no gate folder"
top=$(pwd -P)
JQ=$(command -v jq) || fail "jq is not on PATH"
# The fences as they stood when this plan was written: no task changes them.
for f in gate-review review verify work; do
  cmp -s "$G/$f.json" "$top/fence/$f.json" || fail "fence/$f.json differs from the fence this plan was written against: no task may change a fence's rules"
done
t=$(mktemp -d "${TMPDIR:-/tmp}/gate.XXXXXX") && t=$(cd "$t" && pwd -P) || fail mktemp
trap 'rm -rf "$t"' EXIT
mkdir -p "$t/bin" && go build -buildvcs=false -o "$t/bin/new" ./cmd/vloop || fail "go build of this tree"
S=$t/bin/new
R=$t/repo
mkdir -p "$R" "$t/home" "$t/tmp" || fail mkdir
git -C "$R" init -q || fail "git init of the fixture root"
cd "$R" || fail "fixture root"

# chk <phase> <tool> [--path p] [command]: stdout in $out, exit in $rc.
chk() {
  ph=$1 tl=$2; shift 2
  out=$(env -i PATH=/usr/bin:/bin HOME="$t/home" TMPDIR="$t/tmp" LC_ALL=C "$S" fence check --phase "$ph" --root "$R" --tool "$tl" "$@" 2>"$t/err"); rc=$?
}
allow() {
  ph=$1 tl=$2 lb=$3; shift 3; chk "$ph" "$tl" "$@"
  [ "$rc" = 0 ] && [ "$out" = allow ] || fail "[$ph] $tl $lb: want 'allow' and exit 0, got exit $rc, stdout '$out', stderr '$(cat "$t/err")'"
}
deny() {
  ph=$1 tl=$2 lb=$3; shift 3; chk "$ph" "$tl" "$@"
  [ "$rc" = 1 ] || fail "[$ph] $tl $lb: want exit 1 (deny), got exit $rc, stdout '$out', stderr '$(cat "$t/err")'"
  case "$out" in "deny: "?*) ;; *) fail "[$ph] $tl $lb: want one line 'deny: <reason>', got '$out'";; esac
}
# denyx <phase> <tool> <label> <exact reason> [args...]
denyx() {
  ph=$1 tl=$2 lb=$3 want=$4; shift 4; deny "$ph" "$tl" "$lb" "$@"
  [ "$out" = "deny: $want" ] || fail "[$ph] $tl $lb: want 'deny: $want', got '$out'"
}
# denyp <phase> <tool> <label> <text the reason must hold> [args...]
denyp() {
  ph=$1 tl=$2 lb=$3 want=$4; shift 4; deny "$ph" "$tl" "$lb" "$@"
  printf '%s\n' "$out" | grep -qF -- "$want" || fail "[$ph] $tl $lb: the reason '$out' does not hold '$want'"
}
GR=gate-review
gsh_allow() { allow $GR bash "'$1'" "$1"; allow $GR powershell "'$1'" "$1"; }
gsh_deny() { deny $GR bash "'$1'" "$1"; deny $GR powershell "'$1'" "$1"; }
GW='is not allowed in a gate-review session: only .vloop/tmp/ may be written'

# The worked example, shell table, phase gate-review.
gsh_allow 'vloop task gate T1'
gsh_allow 'vloop task show T1 --json'
gsh_allow 'cat .vloop/state/gates/T1/check.sh'
gsh_allow "printf 'hello\\n' | grep -E '^hello\$'"
gsh_allow 'vloop schema validate gate-verdict/v2 .vloop/tmp/gate-verdict.json'
denyx $GR bash 'cd .vloop/tmp/g && ls > out.txt' "write to out.txt $GW" 'cd .vloop/tmp/g && ls > out.txt'
denyx $GR bash 'echo x > .vloop/state/gates/T1/check.sh' "write to .vloop/state/gates/T1/check.sh $GW" 'echo x > .vloop/state/gates/T1/check.sh'
denyx $GR bash 'rm .vloop/tmp/g/out.txt' 'rm .vloop/tmp/g/out.txt is denied by the Copilot deny list (rm:*)' 'rm .vloop/tmp/g/out.txt'
denyx $GR bash "vloop task verify T1 'true' --reason x" 'vloop task verify T1 true --reason x is denied by the gate-review fence (Bash(vloop task verify:*))' "vloop task verify T1 'true' --reason x"
denyx $GR bash 'git -C .vloop/tmp/g status' 'git -C .vloop/tmp/g status is denied by the gate-review fence (Bash(git -C:*))' 'git -C .vloop/tmp/g status'
denyx $GR bash 'sh -c "git push origin main"' 'git push origin main is denied by the gate-review fence (Bash(git push:*))' 'sh -c "git push origin main"'
denyx $GR bash 'env gh pr list' 'gh pr list is denied by the Copilot deny list (gh:*)' 'env gh pr list'
denyx $GR bash 'echo $(date)' 'command substitution cannot be judged' 'echo $(date)'
gsh_deny 'cd .vloop/tmp/g && ls > out.txt'
gsh_deny 'rm .vloop/tmp/g/out.txt'
gsh_deny "vloop task verify T1 'true' --reason x"

# The worked example, other tools, phase gate-review.
allow $GR create .vloop/tmp/gate-verdict.json --path .vloop/tmp/gate-verdict.json
denyx $GR edit .vloop/state/state.json "write to .vloop/state/state.json $GW" --path .vloop/state/state.json
denyx $GR create src/app.go "write to src/app.go $GW" --path src/app.go
allow $GR view .vloop/state/gates/T1/check.sh --path .vloop/state/gates/T1/check.sh
denyx $GR view '~/.copilot/config.json' 'read of ~/.copilot/config.json is not allowed: it is under ~/.claude, ~/.copilot or ~/.config/gh' --path '~/.copilot/config.json'
allow $GR skill gate-review --path gate-review
denyx $GR skill vloop-gate-review "skill vloop-gate-review is not vloop's — use /vloop:gate-review" --path vloop-gate-review
denyx $GR web_fetch '' 'tool web_fetch is not allowed in a gate-review session'
P_ADD=$(printf '*** Begin Patch\n*** Add File: .vloop/tmp/gate-verdict.json\n+x\n*** End Patch\n')
P_GATE=$(printf '*** Begin Patch\n*** Update File: .vloop/state/gates/T1/check.sh\n@@\n-a\n+b\n*** End Patch\n')
allow $GR apply_patch 'Add .vloop/tmp/gate-verdict.json' "$P_ADD"
denyx $GR apply_patch 'Update .vloop/state/gates/T1/check.sh' "write to .vloop/state/gates/T1/check.sh $GW" "$P_GATE"

# Unchanged, on the same repository.
denyx review skill vloop-gate-review "skill vloop-gate-review is not vloop's — use /vloop:review" --path vloop-gate-review
denyx verify bash "'rm x'" 'rm x is denied by the Copilot deny list (rm:*)' 'rm x'
allow work bash "'rm src/old.go'" 'rm src/old.go'
mkdir -p "$t/cfg" && git -C "$t/cfg" init -q && (cd "$t/cfg" && env -i PATH=/usr/bin:/bin HOME="$t/home" TMPDIR="$t/tmp" LC_ALL=C "$S" init >"$t/init.out" 2>&1) ||
  fail "vloop init of the config fixture: $(cat "$t/init.out")"
c=$(cd "$t/cfg" && env -i PATH=/usr/bin:/bin HOME="$t/home" TMPDIR="$t/tmp" LC_ALL=C "$S" config set agent.plan copilot 2>&1); x=$?
[ "$x" = 2 ] && [ "$c" = 'vloop: invalid value "copilot" for agent.plan: want one of claude' ] ||
  fail "vloop config set agent.plan copilot: want exit 2 and 'vloop: invalid value \"copilot\" for agent.plan: want one of claude', got exit $x, '$c'"
# The other phases decide and word as before.
for ph in review second-review; do
  denyx "$ph" bash 'rm src/old.ts' 'rm src/old.ts is denied by the Copilot deny list (rm:*)' 'rm src/old.ts'
  denyx "$ph" create src/app.ts "write to src/app.ts is not allowed in a $ph session: only .vloop/tmp/ may be written" --path src/app.ts
  denyx "$ph" bash 'git push' 'git push is denied by the review fence (Bash(git push:*))' 'git push'
  denyx "$ph" web_fetch '' "tool web_fetch is not allowed in a $ph session"
  allow "$ph" create .vloop/tmp/verdict.json --path .vloop/tmp/verdict.json
done
denyx verify bash 'vloop run' 'vloop run is denied by the verify fence (Bash(vloop run:*))' 'vloop run'
denyx verify create README.md 'write to README.md is not allowed in a verify session: only .vloop/tmp/ may be written' --path README.md
denyx verify skill vloop-verify "skill vloop-verify is not vloop's — use /vloop:verify" --path vloop-verify
denyx work bash 'rm -rf dist' 'rm -rf dist is denied by the work fence (Bash(rm -rf:*))' 'rm -rf dist'
denyx work skill vloop-work "skill vloop-work is not vloop's — use /vloop:work" --path vloop-work
allow work edit src/app.ts --path src/app.ts

# Changed: an unknown phase, its message, exit 2.
for ph in plan bogus gatereview; do
  chk "$ph" bash 'ls'
  [ "$rc" = 2 ] || fail "fence check --phase $ph: want exit 2, got $rc, stdout '$out', stderr '$(cat "$t/err")'"
  grep -qxF "vloop: unknown phase \"$ph\": want review, second-review, work, verify or gate-review" "$t/err" ||
    fail "fence check --phase $ph: stderr is not 'vloop: unknown phase \"$ph\": want review, second-review, work, verify or gate-review': '$(cat "$t/err")'"
done

# The help names gate-review among the phases: usage, --phase flag, long help.
for sub in check hook; do
  env -i PATH=/usr/bin:/bin HOME="$t/home" LC_ALL=C "$S" fence "$sub" --help >"$t/help" 2>&1 || fail "vloop fence $sub --help failed: $(cat "$t/help")"
  grep -E '^[[:space:]]*vloop fence '"$sub"' --phase' "$t/help" | grep -qF gate-review || fail "vloop fence $sub --help: the usage line does not name gate-review: $(grep -E 'vloop fence '"$sub" "$t/help")"
  grep -E '^[[:space:]]*--phase' "$t/help" | grep -qF gate-review || fail "vloop fence $sub --help: the --phase flag does not name gate-review: $(grep -E -- '--phase' "$t/help")"
  sed -n '1,/^Usage:/p' "$t/help" | grep -qiE 'gate-review|gate review' || fail "vloop fence $sub --help: the long help does not name gate-review"
done

# Writes in gate-review: only under .vloop/tmp/, as in review.
for tl in create edit; do
  allow $GR "$tl" .vloop/tmp/a/b.json --path .vloop/tmp/a/b.json
  allow $GR "$tl" "<root>/.vloop/tmp/g/x.txt" --path "$R/.vloop/tmp/g/x.txt"
  allow $GR "$tl" /dev/null --path /dev/null
  denyx $GR "$tl" README.md "write to README.md $GW" --path README.md
  deny $GR "$tl" "<root>/src/abs.go" --path "$R/src/abs.go"
  deny $GR "$tl" .vloop/tmp/../state/x --path .vloop/tmp/../state/x
  deny $GR "$tl" .vloop/state/gates/T1/check.sh --path .vloop/state/gates/T1/check.sh
  deny $GR "$tl" .git/config --path .git/config
  deny $GR "$tl" .vloop/defects/D1.md --path .vloop/defects/D1.md
  deny $GR "$tl" ../sibling/x --path ../sibling/x
  deny $GR "$tl" '~/.claude/settings.json' --path '~/.claude/settings.json'
  deny $GR "$tl" "(no path)"
done
gsh_allow 'vloop task gate T1 > .vloop/tmp/gate-T1.log'
gsh_allow 'ls .vloop/state/gates > /dev/null'
gsh_deny 'echo x >> .git/config'
gsh_deny 'cat a > src/b.txt'
gsh_deny 'cat a > ../x'
gsh_deny 'echo {} > .vloop/state/state.json'
# Reads in gate-review: as in review.
for tl in view read_file grep rg glob; do
  allow $GR "$tl" .vloop/state/state.json --path .vloop/state/state.json
  allow $GR "$tl" docs/briefs/b.md --path docs/briefs/b.md
  deny $GR "$tl" '~/.claude/settings.json' --path '~/.claude/settings.json'
  deny $GR "$tl" '$HOME/.config/gh/hosts.yml' --path "$t/home/.config/gh/hosts.yml"
  deny $GR "$tl" '$HOME/.copilot/config.json' --path "$t/home/.copilot/config.json"
done
# Skills: vloop's only.
allow $GR skill review --path review
denyx $GR skill my-gate-review "skill my-gate-review is not vloop's — use /vloop:gate-review" --path my-gate-review
# Everything else.
for tl in task explore_subagent frobnicate; do
  denyx $GR "$tl" '' "tool $tl is not allowed in a gate-review session"
done

# Shell in gate-review: the gate-review fence's rules and the whole deny list,
# rm:* included, through the same unwrapping.
gsh_allow 'git status && git diff HEAD'
gsh_allow 'go vet ./...'
gsh_allow 'mkdir -p .vloop/tmp/g'
denyp $GR bash 'rm -f .vloop/tmp/a' 'denied by the Copilot deny list (rm:*)' 'rm -f .vloop/tmp/a'
denyp $GR bash 'env rm x' 'rm:*' 'env rm x'
denyp $GR bash 'curl https://example.com' 'curl' 'curl https://example.com'
denyp $GR bash 'nohup wget x' 'wget' 'nohup wget x'
denyp $GR bash 'git commit -m x' 'git commit' 'git commit -m x'
denyp $GR bash 'ssh host.invalid' 'denied by the Copilot deny list (ssh:*)' 'ssh host.invalid'
denyp $GR bash 'git remote -v' 'denied by the Copilot deny list (git remote:*)' 'git remote -v'
denyx $GR bash 'vloop task set T1 done' 'vloop task set T1 done is denied by the gate-review fence (Bash(vloop task set:*))' 'vloop task set T1 done'
denyx $GR bash 'vloop run' 'vloop run is denied by the gate-review fence (Bash(vloop run:*))' 'vloop run'
denyp $GR bash 'sh -c "gh auth token"' 'denied by the Copilot deny list (gh:*)' 'sh -c "gh auth token"'
gsh_deny 'for f in a b; do echo $f; done'
gsh_deny 'while true; do ls; done'
gsh_deny 'if true; then ls; fi'
gsh_deny '(cd src && ls)'
gsh_deny 'echo `date`'
gsh_deny 'eval "ls"'
gsh_deny 'source x.sh'
gsh_deny '$CMD x'
gsh_deny "echo 'unterminated"

# The hook, phase gate-review: the same judge behind Copilot's payloads.
H=$t/hookroot
mkdir -p "$H/src" "$H/.vloop/tmp" "$H/.vloop/state/gates/T1" || fail mkdir
echo x >"$H/src/app.go"
echo c >"$H/.vloop/state/gates/T1/check.sh"
snap() { (cd "$t" && find hookroot home -print | LC_ALL=C sort && cat hookroot/src/app.go hookroot/.vloop/state/gates/T1/check.sh); }
before=$(snap)
# hookrun <phase> <payload>: $rc the exit, $dec and $why the decision and
# reason of its one line of stdout.
hookrun() {
  hout=$(printf '%s' "$2" | (cd "$H" && env -i PATH=/usr/bin:/bin HOME="$t/home" TMPDIR="$t/tmp" LC_ALL=C "$S" fence hook --phase "$1" --root "$H") 2>"$t/err"); rc=$?
  [ "$(printf '%s\n' "$hout" | grep -c .)" = 1 ] || fail "[hook $1] payload $2: want one line on stdout, got '$hout'"
  dec=$(printf '%s\n' "$hout" | "$JQ" -r '.permissionDecision // "none"' 2>/dev/null) || fail "[hook $1] payload $2: stdout is not JSON: '$hout'"
  why=$(printf '%s\n' "$hout" | "$JQ" -r '.permissionDecisionReason // ""')
}
payload() { "$JQ" -cn --arg n "$1" --argjson a "$2" --arg c "$H" '{timestamp: 1700000000000, cwd: $c, toolName: $n, toolArgs: $a}'; }
bash_() { payload bash "$("$JQ" -cn --arg c "$1" '{command: $c, description: "d"}')"; }
file_() { payload "$1" "$("$JQ" -cn --arg p "$2" '{path: $p, file_text: "x"}')"; }
patch_() { payload apply_patch "$("$JQ" -cn --arg p "$1" '{input: $p}')"; }
ok() {
  hookrun "$1" "$2"
  [ "$rc" = 0 ] && [ "$dec" = allow ] || fail "[hook $1] payload $2: want allow and exit 0, got '$hout', exit $rc"
}
# no <phase> <payload> [exact reason]
no() {
  hookrun "$1" "$2"
  [ "$rc" = 0 ] && [ "$dec" = deny ] || fail "[hook $1] payload $2: want deny and exit 0, got '$hout', exit $rc"
  case "$why" in "vloop fence: "?*) ;; *) fail "[hook $1] payload $2: the reason '$why' does not start 'vloop fence: '";; esac
  [ -z "${3:-}" ] || [ "$why" = "$3" ] || fail "[hook $1] payload $2: the reason is '$why', want '$3'"
}
ok $GR "$(bash_ 'vloop task gate T1')"
ok $GR "$(bash_ 'vloop schema validate gate-verdict/v2 .vloop/tmp/gate-verdict.json')"
no $GR "$(bash_ 'cd .vloop/tmp/g && ls > out.txt')" "vloop fence: write to out.txt $GW"
no $GR "$(bash_ 'echo x > .vloop/state/gates/T1/check.sh')" "vloop fence: write to .vloop/state/gates/T1/check.sh $GW"
no $GR "$(bash_ 'rm .vloop/tmp/g/out.txt')" 'vloop fence: rm .vloop/tmp/g/out.txt is denied by the Copilot deny list (rm:*)'
no $GR "$(bash_ "vloop task verify T1 'true' --reason x")" 'vloop fence: vloop task verify T1 true --reason x is denied by the gate-review fence (Bash(vloop task verify:*))'
no $GR "$(bash_ 'git -C .vloop/tmp/g status')" 'vloop fence: git -C .vloop/tmp/g status is denied by the gate-review fence (Bash(git -C:*))'
no $GR "$(bash_ 'sh -c "git push origin main"')" 'vloop fence: git push origin main is denied by the gate-review fence (Bash(git push:*))'
no $GR "$(bash_ 'env gh pr list')" 'vloop fence: gh pr list is denied by the Copilot deny list (gh:*)'
no $GR "$(bash_ 'echo $(date)')" 'vloop fence: command substitution cannot be judged'
ok $GR "$(file_ create .vloop/tmp/gate-verdict.json)"
ok $GR "$(file_ create "$H/.vloop/tmp/g/x.txt")"
no $GR "$(file_ edit .vloop/state/state.json)" "vloop fence: write to .vloop/state/state.json $GW"
no $GR "$(file_ create src/app.go)" "vloop fence: write to src/app.go $GW"
no $GR "$(file_ edit "$H/src/app.go")"
ok $GR "$(file_ view .vloop/state/gates/T1/check.sh)"
no $GR "$(file_ view "$t/home/.copilot/config.json")"
ok $GR "$(patch_ "$P_ADD")"
no $GR "$(patch_ "$P_GATE")" "vloop fence: write to .vloop/state/gates/T1/check.sh $GW"
ok $GR "$(payload skill '{"skill":"gate-review"}')"
no $GR "$(payload skill '{"skill":"vloop-gate-review"}')" "vloop fence: skill vloop-gate-review is not vloop's — use /vloop:gate-review"
no $GR "$(payload web_fetch '{"url":"https://example.com"}')" 'vloop fence: tool web_fetch is not allowed in a gate-review session'
no $GR "$(payload search_code_subagent '{"query":"x"}')" 'vloop fence: tool search_code_subagent is not allowed in a gate-review session'
# The hook's other phases answer as before.
for ph in review second-review; do
  no "$ph" "$(bash_ 'rm src/old.ts')" 'vloop fence: rm src/old.ts is denied by the Copilot deny list (rm:*)'
  no "$ph" "$(file_ create src/a.ts)" "vloop fence: write to src/a.ts is not allowed in a $ph session: only .vloop/tmp/ may be written"
  ok "$ph" "$(file_ create .vloop/tmp/verdict.json)"
  no "$ph" "$(payload skill '{"skill":"vloop-gate-review"}')" "vloop fence: skill vloop-gate-review is not vloop's — use /vloop:review"
done
no verify "$(bash_ 'vloop run')" 'vloop fence: vloop run is denied by the verify fence (Bash(vloop run:*))'
no verify "$(file_ create src/a.ts)" 'vloop fence: write to src/a.ts is not allowed in a verify session: only .vloop/tmp/ may be written'
ok work "$(bash_ 'rm src/old.ts')"
no work "$(bash_ 'rm -rf dist')" 'vloop fence: rm -rf dist is denied by the work fence (Bash(rm -rf:*))'
ok work "$(file_ create src/links/a.ts)"
for ph in plan bogus; do
  hookrun "$ph" "$(bash_ 'ls')"
  [ "$rc" = 2 ] && [ "$dec" = deny ] || fail "fence hook --phase $ph: want the deny line and exit 2, got '$hout', exit $rc"
  printf '%s' "$why" | grep -qF "unknown phase \"$ph\": want review, second-review, work, verify or gate-review" ||
    fail "fence hook --phase $ph: the reason '$why' does not hold 'unknown phase \"$ph\": want review, second-review, work, verify or gate-review'"
done
[ "$(snap)" = "$before" ] || fail "the hook changed the fixture root or the HOME"
echo "gate T1: ok"
