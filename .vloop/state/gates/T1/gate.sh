#!/bin/sh
# Gate T1 — P1, the judge's plan phase, through `vloop fence check` and
# `vloop fence hook` (they share the judge). Every line of the worked example's
# two tables and its apply_patch lines in phase plan, the unchanged lines
# (gate-review's state.json write, review's skill hint, work's rm), the changed
# unknown-phase message, both commands' help naming plan, and the contract's
# plan rules beyond the example (as in review, with the plan write rule:
# .vloop/state/state.json, under .vloop/state/gates/ and under .vloop/tmp/,
# the directories themselves not; the whole deny list with rm:*; reads;
# skills; other tools; the shell syntax that stays denied). Then the hook:
# each Copilot payload answered with the same decision and the reason
# `vloop fence: <message>`, exit 0; an unknown phase the deny line and exit 2.
# A few review, second-review, work, verify and gate-review lines pin that
# those phases decide and word as before. Nothing runs a command under
# judgment: the gate only asks the judge, in a fixture root with a fake HOME,
# never this repository's tree.
set -u
fail() { echo "GATE FAIL: $*" >&2; exit 1; }
for v in $(env | sed -n 's/^\(VLOOP_[A-Z0-9_]*\)=.*/\1/p'); do unset "$v"; done
export LC_ALL=C GOTOOLCHAIN=local
G=$(cd "$(dirname "$0")" && pwd -P) || fail "no gate folder"
top=$(pwd -P)
JQ=$(command -v jq) || fail "jq is not on PATH"
# The fences as they stood when this plan was written: no task changes them.
for f in plan gate-review review verify work; do
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

# chk <phase> <tool> [--path p] [command]: stdout in $out, exit in $rc. The
# worked example's lines run with --root . (its own form); the rest, which
# also give absolute paths, with the fixture root's physical path.
ROOTARG=.
chk() {
  ph=$1 tl=$2; shift 2
  out=$(env -i PATH=/usr/bin:/bin HOME="$t/home" TMPDIR="$t/tmp" LC_ALL=C "$S" fence check --phase "$ph" --root "$ROOTARG" --tool "$tl" "$@" 2>"$t/err"); rc=$?
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
P=plan
psh_allow() { allow $P bash "'$1'" "$1"; allow $P powershell "'$1'" "$1"; }
psh_deny() { deny $P bash "'$1'" "$1"; deny $P powershell "'$1'" "$1"; }
PW='is not allowed in a plan session: only .vloop/state/state.json, .vloop/state/gates/ and .vloop/tmp/ may be written'

# The worked example, shell table, phase plan.
psh_allow 'vloop task gate T1'
psh_allow 'vloop schema validate state/v2 .vloop/state/state.json'
psh_allow 'mkdir -p .vloop/state/gates/T1'
psh_allow "printf '{}' > .vloop/state/state.json"
psh_allow "echo 'exit 1' > .vloop/state/gates/T1/gate.sh"
psh_allow 'echo x > .vloop/tmp/p/out.txt'
denyx $P bash 'journal' "write to .vloop/state/journals/a.md $PW" 'echo x > .vloop/state/journals/a.md'
denyx $P bash 'src' "write to src/app.go $PW" 'echo x > src/app.go'
denyx $P bash 'cd gates' "write to f $PW" 'cd .vloop/state/gates/T1 && echo x > f'
denyx $P bash 'rm gate' 'rm .vloop/state/gates/T1/old.sh is denied by the Copilot deny list (rm:*)' 'rm .vloop/state/gates/T1/old.sh'
denyx $P bash 'sh gate' 'sh: runs a script or stdin, not a literal -c string' 'sh .vloop/state/gates/T1/gate.sh'
denyx $P bash 'task drop' 'vloop task drop T2 is denied by the plan fence (Bash(vloop task drop:*))' 'vloop task drop T2'
denyx $P bash 'sh -c push' 'git push origin main is denied by the plan fence (Bash(git push:*))' 'sh -c "git push origin main"'
denyx $P bash 'env gh' 'gh pr list is denied by the Copilot deny list (gh:*)' 'env gh pr list'
denyx $P bash 'subst' 'command substitution cannot be judged' 'echo $(date)'
psh_deny 'echo x > .vloop/state/journals/a.md'
psh_deny 'cd .vloop/state/gates/T1 && echo x > f'
psh_deny 'rm .vloop/state/gates/T1/old.sh'
psh_deny 'vloop task drop T2'

# The worked example, other tools, phase plan.
allow $P create .vloop/state/state.json --path .vloop/state/state.json
allow $P create .vloop/state/gates/T1/oracle.test.ts --path .vloop/state/gates/T1/oracle.test.ts
allow $P create .vloop/tmp/notes.md --path .vloop/tmp/notes.md
denyx $P edit .vloop/state/plan.md "write to .vloop/state/plan.md $PW" --path .vloop/state/plan.md
denyx $P create .vloop/config.toml "write to .vloop/config.toml $PW" --path .vloop/config.toml
denyx $P create .vloop/state/gates "write to .vloop/state/gates $PW" --path .vloop/state/gates
allow $P view docs/briefs/b.loop-brief.md --path docs/briefs/b.loop-brief.md
denyx $P view '~/.copilot/config.json' 'read of ~/.copilot/config.json is not allowed: it is under ~/.claude, ~/.copilot or ~/.config/gh' --path '~/.copilot/config.json'
allow $P skill plan --path plan
denyx $P skill vloop-plan "skill vloop-plan is not vloop's — use /vloop:plan" --path vloop-plan
denyx $P web_fetch '' 'tool web_fetch is not allowed in a plan session'
P_ADD=$(printf '*** Begin Patch\n*** Add File: .vloop/state/gates/T1/gate.sh\n+exit 1\n*** End Patch\n')
P_README=$(printf '*** Begin Patch\n*** Update File: README.md\n@@\n-a\n+b\n*** End Patch\n')
allow $P apply_patch 'Add .vloop/state/gates/T1/gate.sh' "$P_ADD"
denyx $P apply_patch 'Update README.md' "write to README.md $PW" "$P_README"

# Unchanged, on the same repository.
denyx gate-review create .vloop/state/state.json 'write to .vloop/state/state.json is not allowed in a gate-review session: only .vloop/tmp/ may be written' --path .vloop/state/state.json
denyx review skill vloop-plan "skill vloop-plan is not vloop's — use /vloop:review" --path vloop-plan
allow work bash "'rm src/old.go'" 'rm src/old.go'
# The other phases decide and word as before.
for ph in review second-review; do
  denyx "$ph" bash 'rm src/old.ts' 'rm src/old.ts is denied by the Copilot deny list (rm:*)' 'rm src/old.ts'
  denyx "$ph" create src/app.ts "write to src/app.ts is not allowed in a $ph session: only .vloop/tmp/ may be written" --path src/app.ts
  denyx "$ph" create .vloop/state/state.json "write to .vloop/state/state.json is not allowed in a $ph session: only .vloop/tmp/ may be written" --path .vloop/state/state.json
  denyx "$ph" create .vloop/state/gates/T1/g.sh "write to .vloop/state/gates/T1/g.sh is not allowed in a $ph session: only .vloop/tmp/ may be written" --path .vloop/state/gates/T1/g.sh
  denyx "$ph" bash 'git push' 'git push is denied by the review fence (Bash(git push:*))' 'git push'
  denyx "$ph" web_fetch '' "tool web_fetch is not allowed in a $ph session"
  allow "$ph" create .vloop/tmp/verdict.json --path .vloop/tmp/verdict.json
done
denyx verify bash 'vloop run' 'vloop run is denied by the verify fence (Bash(vloop run:*))' 'vloop run'
denyx verify create .vloop/state/state.json 'write to .vloop/state/state.json is not allowed in a verify session: only .vloop/tmp/ may be written' --path .vloop/state/state.json
denyx verify skill vloop-verify "skill vloop-verify is not vloop's — use /vloop:verify" --path vloop-verify
denyx gate-review create .vloop/state/gates/T1/g.sh 'write to .vloop/state/gates/T1/g.sh is not allowed in a gate-review session: only .vloop/tmp/ may be written' --path .vloop/state/gates/T1/g.sh
denyx gate-review bash 'task verify' 'vloop task verify T1 is denied by the gate-review fence (Bash(vloop task verify:*))' 'vloop task verify T1'
denyx gate-review skill vloop-plan "skill vloop-plan is not vloop's — use /vloop:gate-review" --path vloop-plan
denyx work bash 'rm -rf dist' 'rm -rf dist is denied by the work fence (Bash(rm -rf:*))' 'rm -rf dist'
denyx work skill vloop-work "skill vloop-work is not vloop's — use /vloop:work" --path vloop-work
allow work edit src/app.ts --path src/app.ts

# Changed: an unknown phase, its message, exit 2.
for ph in deploy bogus plans; do
  chk "$ph" bash 'ls'
  [ "$rc" = 2 ] || fail "fence check --phase $ph: want exit 2, got $rc, stdout '$out', stderr '$(cat "$t/err")'"
  grep -qxF "vloop: unknown phase \"$ph\": want review, second-review, work, verify, gate-review or plan" "$t/err" ||
    fail "fence check --phase $ph: stderr is not 'vloop: unknown phase \"$ph\": want review, second-review, work, verify, gate-review or plan': '$(cat "$t/err")'"
done

# The help names plan among the phases: usage, --phase flag, long help.
for sub in check hook; do
  env -i PATH=/usr/bin:/bin HOME="$t/home" LC_ALL=C "$S" fence "$sub" --help >"$t/help" 2>&1 || fail "vloop fence $sub --help failed: $(cat "$t/help")"
  grep -E '^[[:space:]]*vloop fence '"$sub"' --phase' "$t/help" | grep -qE '[^a-z-]plan([^a-z-]|$)' || fail "vloop fence $sub --help: the usage line does not name plan: $(grep -E 'vloop fence '"$sub" "$t/help")"
  grep -E '^[[:space:]]*--phase' "$t/help" | grep -qE '[^a-z-]plan([^a-z-]|$)' || fail "vloop fence $sub --help: the --phase flag does not name plan: $(grep -E -- '--phase' "$t/help")"
  sed -n '1,/^Usage:/p' "$t/help" | grep -qiE '(^|[^a-z-])plan([^a-z-]|$)' || fail "vloop fence $sub --help: the long help does not name plan"
done

ROOTARG=$R
# Writes in plan: .vloop/state/state.json, under .vloop/state/gates/, under .vloop/tmp/.
for tl in create edit; do
  allow $P "$tl" .vloop/state/state.json --path .vloop/state/state.json
  allow $P "$tl" "<root>/.vloop/state/state.json" --path "$R/.vloop/state/state.json"
  allow $P "$tl" .vloop/state/gates/T1/gate.sh --path .vloop/state/gates/T1/gate.sh
  allow $P "$tl" .vloop/state/gates/T12/a/b/fixture.json --path .vloop/state/gates/T12/a/b/fixture.json
  allow $P "$tl" "<root>/.vloop/state/gates/T2/x.sh" --path "$R/.vloop/state/gates/T2/x.sh"
  allow $P "$tl" .vloop/tmp/a/b.json --path .vloop/tmp/a/b.json
  allow $P "$tl" "<root>/.vloop/tmp/p/x.txt" --path "$R/.vloop/tmp/p/x.txt"
  allow $P "$tl" /dev/null --path /dev/null
  denyx $P "$tl" README.md "write to README.md $PW" --path README.md
  denyx $P "$tl" .vloop/state/gates/ "write to .vloop/state/gates/ $PW" --path .vloop/state/gates/
  denyx $P "$tl" .vloop/tmp "write to .vloop/tmp $PW" --path .vloop/tmp
  deny $P "$tl" .vloop/state --path .vloop/state
  deny $P "$tl" .vloop/state/state.json.bak --path .vloop/state/state.json.bak
  deny $P "$tl" .vloop/state/state.json/x --path .vloop/state/state.json/x
  deny $P "$tl" .vloop/state/gatesx/T1/a --path .vloop/state/gatesx/T1/a
  deny $P "$tl" .vloop/state/gates/../plan.md --path .vloop/state/gates/../plan.md
  deny $P "$tl" .vloop/tmp/../config.toml --path .vloop/tmp/../config.toml
  deny $P "$tl" .vloop/state/runs/r/x.json --path .vloop/state/runs/r/x.json
  deny $P "$tl" .vloop/state/journals/j.md --path .vloop/state/journals/j.md
  deny $P "$tl" .vloop/state/metrics.json --path .vloop/state/metrics.json
  deny $P "$tl" "<root>/src/abs.go" --path "$R/src/abs.go"
  deny $P "$tl" .git/config --path .git/config
  deny $P "$tl" .vloop/defects/D1.md --path .vloop/defects/D1.md
  deny $P "$tl" ../sibling/.vloop/state/state.json --path ../sibling/.vloop/state/state.json
  deny $P "$tl" '~/.claude/settings.json' --path '~/.claude/settings.json'
  deny $P "$tl" "(no path)"
done
psh_allow 'vloop task gate T1 > .vloop/tmp/gate-T1.log'
psh_allow 'jq . .vloop/state/state.json > .vloop/state/gates/T1/plan-copy.json'
psh_allow 'ls .vloop/state/gates > /dev/null'
psh_deny 'echo x >> .git/config'
psh_deny 'cat a > src/b.txt'
psh_deny 'cat a > ../x'
psh_deny 'echo {} > .vloop/state/plan.md'
psh_deny 'echo x > .vloop/config.toml'
# Reads in plan: as in review.
for tl in view read_file grep rg glob; do
  allow $P "$tl" .vloop/state/state.json --path .vloop/state/state.json
  allow $P "$tl" docs/briefs/b.md --path docs/briefs/b.md
  deny $P "$tl" '~/.claude/settings.json' --path '~/.claude/settings.json'
  deny $P "$tl" '$HOME/.config/gh/hosts.yml' --path "$t/home/.config/gh/hosts.yml"
  deny $P "$tl" '$HOME/.copilot/config.json' --path "$t/home/.copilot/config.json"
done
# Skills: vloop's only.
allow $P skill review --path review
denyx $P skill my-plan "skill my-plan is not vloop's — use /vloop:plan" --path my-plan
# Everything else.
for tl in task explore_subagent frobnicate; do
  denyx $P "$tl" '' "tool $tl is not allowed in a plan session"
done

# Shell in plan: the plan fence's rules and the whole deny list, rm:*
# included, through the same unwrapping.
psh_allow 'git status && git diff HEAD'
psh_allow 'vloop task validate'
psh_allow 'vloop config get shell'
denyp $P bash 'rm -f .vloop/tmp/a' 'denied by the Copilot deny list (rm:*)' 'rm -f .vloop/tmp/a'
denyp $P bash 'env rm x' 'rm:*' 'env rm x'
denyp $P bash 'curl https://example.com' 'curl' 'curl https://example.com'
denyp $P bash 'nohup wget x' 'wget' 'nohup wget x'
denyp $P bash 'git commit -m x' 'git commit' 'git commit -m x'
denyp $P bash 'ssh host.invalid' 'denied by the Copilot deny list (ssh:*)' 'ssh host.invalid'
denyp $P bash 'git remote -v' 'denied by the Copilot deny list (git remote:*)' 'git remote -v'
denyx $P bash 'task set' 'vloop task set T1 done is denied by the plan fence (Bash(vloop task set:*))' 'vloop task set T1 done'
denyx $P bash 'task verify' 'vloop task verify T1 is denied by the plan fence (Bash(vloop task verify:*))' 'vloop task verify T1'
denyx $P bash 'vloop run' 'vloop run is denied by the plan fence (Bash(vloop run:*))' 'vloop run'
denyp $P bash 'sh -c "gh auth token"' 'denied by the Copilot deny list (gh:*)' 'sh -c "gh auth token"'
psh_deny 'for f in a b; do echo $f; done'
psh_deny 'while true; do ls; done'
psh_deny 'if true; then ls; fi'
psh_deny '(cd src && ls)'
psh_deny 'echo `date`'
psh_deny 'eval "ls"'
psh_deny 'source x.sh'
psh_deny 'bash .vloop/state/gates/T1/gate.sh'
psh_deny '$CMD x'
psh_deny "echo 'unterminated"

# The hook, phase plan: the same judge behind Copilot's payloads.
H=$t/hookroot
mkdir -p "$H/src" "$H/.vloop/tmp" "$H/.vloop/state/gates/T1" || fail mkdir
echo x >"$H/src/app.go"
echo c >"$H/.vloop/state/gates/T1/gate.sh"
snap() { (cd "$t" && find hookroot home -print | LC_ALL=C sort && cat hookroot/src/app.go hookroot/.vloop/state/gates/T1/gate.sh); }
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
ok $P "$(bash_ 'vloop task gate T1')"
ok $P "$(bash_ 'vloop schema validate state/v2 .vloop/state/state.json')"
ok $P "$(bash_ "printf '{}' > .vloop/state/state.json")"
ok $P "$(bash_ "echo 'exit 1' > .vloop/state/gates/T1/gate.sh")"
no $P "$(bash_ 'echo x > .vloop/state/journals/a.md')" "vloop fence: write to .vloop/state/journals/a.md $PW"
no $P "$(bash_ 'cd .vloop/state/gates/T1 && echo x > f')" "vloop fence: write to f $PW"
no $P "$(bash_ 'rm .vloop/state/gates/T1/old.sh')" 'vloop fence: rm .vloop/state/gates/T1/old.sh is denied by the Copilot deny list (rm:*)'
no $P "$(bash_ 'sh .vloop/state/gates/T1/gate.sh')" 'vloop fence: sh: runs a script or stdin, not a literal -c string'
no $P "$(bash_ 'vloop task drop T2')" 'vloop fence: vloop task drop T2 is denied by the plan fence (Bash(vloop task drop:*))'
no $P "$(bash_ 'sh -c "git push origin main"')" 'vloop fence: git push origin main is denied by the plan fence (Bash(git push:*))'
no $P "$(bash_ 'env gh pr list')" 'vloop fence: gh pr list is denied by the Copilot deny list (gh:*)'
no $P "$(bash_ 'echo $(date)')" 'vloop fence: command substitution cannot be judged'
ok $P "$(file_ create .vloop/state/state.json)"
ok $P "$(file_ create .vloop/state/gates/T1/oracle.test.ts)"
ok $P "$(file_ create "$H/.vloop/state/gates/T1/x.sh")"
ok $P "$(file_ create .vloop/tmp/notes.md)"
no $P "$(file_ edit .vloop/state/plan.md)" "vloop fence: write to .vloop/state/plan.md $PW"
no $P "$(file_ create .vloop/config.toml)" "vloop fence: write to .vloop/config.toml $PW"
no $P "$(file_ create .vloop/state/gates)" "vloop fence: write to .vloop/state/gates $PW"
no $P "$(file_ edit "$H/src/app.go")"
ok $P "$(file_ view docs/briefs/b.loop-brief.md)"
no $P "$(file_ view "$t/home/.copilot/config.json")"
ok $P "$(patch_ "$P_ADD")"
no $P "$(patch_ "$P_README")" "vloop fence: write to README.md $PW"
ok $P "$(payload skill '{"skill":"plan"}')"
no $P "$(payload skill '{"skill":"vloop-plan"}')" "vloop fence: skill vloop-plan is not vloop's — use /vloop:plan"
no $P "$(payload web_fetch '{"url":"https://example.com"}')" 'vloop fence: tool web_fetch is not allowed in a plan session'
no $P "$(payload search_code_subagent '{"query":"x"}')" 'vloop fence: tool search_code_subagent is not allowed in a plan session'
# The hook's other phases answer as before.
for ph in review second-review; do
  no "$ph" "$(bash_ 'rm src/old.ts')" 'vloop fence: rm src/old.ts is denied by the Copilot deny list (rm:*)'
  no "$ph" "$(file_ create src/a.ts)" "vloop fence: write to src/a.ts is not allowed in a $ph session: only .vloop/tmp/ may be written"
  ok "$ph" "$(file_ create .vloop/tmp/verdict.json)"
  no "$ph" "$(payload skill '{"skill":"vloop-plan"}')" "vloop fence: skill vloop-plan is not vloop's — use /vloop:review"
done
no verify "$(bash_ 'vloop run')" 'vloop fence: vloop run is denied by the verify fence (Bash(vloop run:*))'
no gate-review "$(file_ create .vloop/state/state.json)" 'vloop fence: write to .vloop/state/state.json is not allowed in a gate-review session: only .vloop/tmp/ may be written'
ok work "$(bash_ 'rm src/old.ts')"
no work "$(bash_ 'rm -rf dist')" 'vloop fence: rm -rf dist is denied by the work fence (Bash(rm -rf:*))'
ok work "$(file_ create src/links/a.ts)"
for ph in deploy bogus; do
  hookrun "$ph" "$(bash_ 'ls')"
  [ "$rc" = 2 ] && [ "$dec" = deny ] || fail "fence hook --phase $ph: want the deny line and exit 2, got '$hout', exit $rc"
  printf '%s' "$why" | grep -qF "unknown phase \"$ph\": want review, second-review, work, verify, gate-review or plan" ||
    fail "fence hook --phase $ph: the reason '$why' does not hold 'unknown phase \"$ph\": want review, second-review, work, verify, gate-review or plan'"
done
[ "$(snap)" = "$before" ] || fail "the hook changed the fixture root or the HOME"
echo "gate T1: ok"
