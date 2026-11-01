#!/bin/sh
# Gate T1 — V1, the judge's verify phase, through `vloop fence check`. Every
# line of the worked example's two tables and its apply_patch lines in phase
# verify, the unchanged lines (review's skill hint, work's rm, gate-review's
# config refusal), the changed unknown-phase message, the help naming verify,
# and the contract's verify rules beyond the example (as in review: writes
# only under .vloop/tmp/, the whole deny list with rm:*, reads, skills, other
# tools, the shell syntax that stays denied). `vloop fence check` prints
# `allow` (exit 0) or `deny: <reason>` (exit 1); bad arguments exit 2. A few
# review, second-review and work lines pin that those phases decide and word as
# before. Nothing runs a command under judgment: the gate only asks the judge,
# in a fixture root with a fake HOME, never this repository's tree.
set -u
fail() { echo "GATE FAIL: $*" >&2; exit 1; }
for v in $(env | sed -n 's/^\(VLOOP_[A-Z0-9_]*\)=.*/\1/p'); do unset "$v"; done
export LC_ALL=C GOTOOLCHAIN=local
G=$(cd "$(dirname "$0")" && pwd -P) || fail "no gate folder"
top=$(pwd -P)
# The fences as they stood when this plan was written: no task changes them.
for f in verify review work; do
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
vsh_allow() { allow verify bash "'$1'" "$1"; allow verify powershell "'$1'" "$1"; }
vsh_deny() { deny verify bash "'$1'" "$1"; deny verify powershell "'$1'" "$1"; }
VW='is not allowed in a verify session: only .vloop/tmp/ may be written'

# The worked example, shell table, phase verify.
vsh_allow 'go build -o .vloop/tmp/v/vloop ./cmd/vloop'
vsh_allow 'git init .vloop/tmp/v/ex1'
vsh_allow './.vloop/tmp/v/vloop --version > .vloop/tmp/v/out.txt'
vsh_allow 'vloop schema validate verify/v1 .vloop/tmp/verify.json'
denyx verify bash 'cd .vloop/tmp/v && ./vloop status > out.txt' "write to out.txt $VW" 'cd .vloop/tmp/v && ./vloop status > out.txt'
denyx verify bash 'echo x > README.md' "write to README.md $VW" 'echo x > README.md'
denyx verify bash 'rm .vloop/tmp/v/out.txt' 'rm .vloop/tmp/v/out.txt is denied by the Copilot deny list (rm:*)' 'rm .vloop/tmp/v/out.txt'
denyx verify bash 'git -C .vloop/tmp/v/ex1 status' 'git -C .vloop/tmp/v/ex1 status is denied by the verify fence (Bash(git -C:*))' 'git -C .vloop/tmp/v/ex1 status'
denyx verify bash 'vloop run' 'vloop run is denied by the verify fence (Bash(vloop run:*))' 'vloop run'
denyx verify bash 'sh -c "git push origin main"' 'git push origin main is denied by the verify fence (Bash(git push:*))' 'sh -c "git push origin main"'
denyx verify bash 'env gh pr list' 'gh pr list is denied by the Copilot deny list (gh:*)' 'env gh pr list'
denyx verify bash 'echo $(date)' 'command substitution cannot be judged' 'echo $(date)'
vsh_deny 'cd .vloop/tmp/v && ./vloop status > out.txt'
vsh_deny 'rm .vloop/tmp/v/out.txt'

# The worked example, other tools, phase verify.
allow verify create .vloop/tmp/verify.json --path .vloop/tmp/verify.json
denyx verify edit internal/x.go "write to internal/x.go $VW" --path internal/x.go
denyx verify create .vloop/state/state.json "write to .vloop/state/state.json $VW" --path .vloop/state/state.json
allow verify view src/app.go --path src/app.go
denyx verify view '~/.copilot/config.json' 'read of ~/.copilot/config.json is not allowed: it is under ~/.claude, ~/.copilot or ~/.config/gh' --path '~/.copilot/config.json'
allow verify skill verify --path verify
denyx verify skill vloop-verify "skill vloop-verify is not vloop's — use /vloop:verify" --path vloop-verify
denyx verify web_fetch '' 'tool web_fetch is not allowed in a verify session'
P_ADD=$(printf '*** Begin Patch\n*** Add File: .vloop/tmp/v/x.txt\n+x\n*** End Patch\n')
P_SRC=$(printf '*** Begin Patch\n*** Update File: src/app.go\n@@\n-a\n+b\n*** End Patch\n')
allow verify apply_patch 'Add .vloop/tmp/v/x.txt' "$P_ADD"
denyx verify apply_patch 'Update src/app.go' "write to src/app.go $VW" "$P_SRC"

# Unchanged, on the same repository.
denyx review skill vloop-verify "skill vloop-verify is not vloop's — use /vloop:review" --path vloop-verify
allow work bash "'rm src/old.go'" 'rm src/old.go'
mkdir -p "$t/cfg" && git -C "$t/cfg" init -q && (cd "$t/cfg" && env -i PATH=/usr/bin:/bin HOME="$t/home" TMPDIR="$t/tmp" LC_ALL=C "$S" init >"$t/init.out" 2>&1) ||
  fail "vloop init of the config fixture: $(cat "$t/init.out")"
c=$(cd "$t/cfg" && env -i PATH=/usr/bin:/bin HOME="$t/home" TMPDIR="$t/tmp" LC_ALL=C "$S" config set agent.gate-review copilot 2>&1); x=$?
[ "$x" = 2 ] && [ "$c" = 'vloop: invalid value "copilot" for agent.gate-review: want one of claude' ] ||
  fail "vloop config set agent.gate-review copilot: want exit 2 and 'vloop: invalid value \"copilot\" for agent.gate-review: want one of claude', got exit $x, '$c'"
# The other phases decide and word as before.
for ph in review second-review; do
  denyx "$ph" bash 'rm src/old.ts' 'rm src/old.ts is denied by the Copilot deny list (rm:*)' 'rm src/old.ts'
  denyx "$ph" create src/app.ts "write to src/app.ts is not allowed in a $ph session: only .vloop/tmp/ may be written" --path src/app.ts
  denyx "$ph" bash 'git push' 'git push is denied by the review fence (Bash(git push:*))' 'git push'
  denyx "$ph" web_fetch '' "tool web_fetch is not allowed in a $ph session"
  allow "$ph" create .vloop/tmp/verdict.json --path .vloop/tmp/verdict.json
done
denyx work bash 'rm -rf dist' 'rm -rf dist is denied by the work fence (Bash(rm -rf:*))' 'rm -rf dist'
denyx work skill vloop-work "skill vloop-work is not vloop's — use /vloop:work" --path vloop-work
allow work edit src/app.ts --path src/app.ts

# Changed: an unknown phase, its message, exit 2.
for ph in plan gate-review bogus; do
  chk "$ph" bash 'ls'
  [ "$rc" = 2 ] || fail "fence check --phase $ph: want exit 2, got $rc, stdout '$out', stderr '$(cat "$t/err")'"
  grep -qxF "vloop: unknown phase \"$ph\": want review, second-review, work or verify" "$t/err" ||
    fail "fence check --phase $ph: stderr is not 'vloop: unknown phase \"$ph\": want review, second-review, work or verify': '$(cat "$t/err")'"
done

# The help names verify among the phases: usage, --phase flag, long help.
for sub in check hook; do
  env -i PATH=/usr/bin:/bin HOME="$t/home" LC_ALL=C "$S" fence "$sub" --help >"$t/help" 2>&1 || fail "vloop fence $sub --help failed: $(cat "$t/help")"
  grep -E '^[[:space:]]*vloop fence '"$sub"' --phase' "$t/help" | grep -qw verify || fail "vloop fence $sub --help: the usage line does not name verify: $(grep -E 'vloop fence '"$sub" "$t/help")"
  grep -E '^[[:space:]]*--phase' "$t/help" | grep -qw verify || fail "vloop fence $sub --help: the --phase flag does not name verify: $(grep -E -- '--phase' "$t/help")"
  sed -n '1,/^Usage:/p' "$t/help" | grep -qw verify || fail "vloop fence $sub --help: the long help does not name verify"
done

# Writes in verify: only under .vloop/tmp/, as in review.
for tl in create edit; do
  allow verify "$tl" .vloop/tmp/a/b.json --path .vloop/tmp/a/b.json
  allow verify "$tl" "<root>/.vloop/tmp/v/x.txt" --path "$R/.vloop/tmp/v/x.txt"
  allow verify "$tl" /dev/null --path /dev/null
  denyx verify "$tl" README.md "write to README.md $VW" --path README.md
  deny verify "$tl" "<root>/src/abs.go" --path "$R/src/abs.go"
  deny verify "$tl" .vloop/tmp/../state/x --path .vloop/tmp/../state/x
  deny verify "$tl" .git/config --path .git/config
  deny verify "$tl" .vloop/defects/D1.md --path .vloop/defects/D1.md
  deny verify "$tl" ../sibling/x --path ../sibling/x
  deny verify "$tl" '~/.claude/settings.json' --path '~/.claude/settings.json'
  deny verify "$tl" "(no path)"
done
vsh_allow 'go test ./... > .vloop/tmp/test.log'
vsh_allow 'ls src > /dev/null'
vsh_deny 'echo x >> .git/config'
vsh_deny 'cat a > src/b.txt'
vsh_deny 'cat a > ../x'
# Reads in verify: as in review.
for tl in view read_file grep rg glob; do
  allow verify "$tl" .vloop/state/state.json --path .vloop/state/state.json
  allow verify "$tl" docs/briefs/b.md --path docs/briefs/b.md
  deny verify "$tl" '~/.claude/settings.json' --path '~/.claude/settings.json'
  deny verify "$tl" '$HOME/.config/gh/hosts.yml' --path "$t/home/.config/gh/hosts.yml"
  deny verify "$tl" '$HOME/.copilot/config.json' --path "$t/home/.copilot/config.json"
done
# Skills: vloop's only.
allow verify skill review --path review
denyx verify skill my-verify "skill my-verify is not vloop's — use /vloop:verify" --path my-verify
# Everything else.
for tl in task explore_subagent frobnicate; do
  denyx verify "$tl" '' "tool $tl is not allowed in a verify session"
done

# Shell in verify: the verify fence's rules and the whole deny list, rm:*
# included, through the same unwrapping.
vsh_allow 'git status && git diff HEAD'
vsh_allow 'go vet ./...'
vsh_allow 'mkdir -p .vloop/tmp/v'
denyp verify bash 'rm -f .vloop/tmp/a' 'denied by the Copilot deny list (rm:*)' 'rm -f .vloop/tmp/a'
denyp verify bash 'env rm x' 'rm:*' 'env rm x'
denyp verify bash 'curl https://example.com' 'curl' 'curl https://example.com'
denyp verify bash 'nohup wget x' 'wget' 'nohup wget x'
denyp verify bash 'git commit -m x' 'git commit' 'git commit -m x'
denyp verify bash 'ssh host.invalid' 'denied by the Copilot deny list (ssh:*)' 'ssh host.invalid'
denyp verify bash 'git remote -v' 'denied by the Copilot deny list (git remote:*)' 'git remote -v'
denyp verify bash 'vloop task set T1 done' 'denied by the verify fence (Bash(vloop task set:*))' 'vloop task set T1 done'
denyp verify bash 'sh -c "gh auth token"' 'denied by the Copilot deny list (gh:*)' 'sh -c "gh auth token"'
vsh_deny 'for f in a b; do echo $f; done'
vsh_deny 'while true; do ls; done'
vsh_deny 'if true; then ls; fi'
vsh_deny '(cd src && ls)'
vsh_deny 'echo `date`'
vsh_deny 'eval "ls"'
vsh_deny 'source x.sh'
vsh_deny '$CMD x'
vsh_deny "echo 'unterminated"
echo "gate T1: ok"
