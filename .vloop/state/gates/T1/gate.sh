#!/bin/sh
# Gate T1 — W1, the judge's work phase, through `vloop fence check`. Every line
# of the worked example's two tables in phase work, the review phase's `rm`
# line unchanged, the unknown phase's message, and the contract's work rules
# beyond the example (writes, reads, skills, other tools, package installs, the
# shell syntax that stays denied). `vloop fence check` prints `allow` (exit 0)
# or `deny: <reason>` (exit 1), where <reason> is what the hook sends after its
# `vloop fence: ` prefix; bad arguments exit 2. A few review and second-review
# lines pin that those phases decide and word as before. Nothing runs a command
# under judgment: the gate only asks the judge, in a fixture root with a fake
# HOME, never this repository's tree.
set -u
fail() { echo "GATE FAIL: $*" >&2; exit 1; }
for v in $(env | sed -n 's/^\(VLOOP_[A-Z0-9_]*\)=.*/\1/p'); do unset "$v"; done
export LC_ALL=C GOTOOLCHAIN=local
top=$(pwd -P)
t=$(mktemp -d "${TMPDIR:-/tmp}/gate.XXXXXX") && t=$(cd "$t" && pwd -P) || fail mktemp
trap 'rm -rf "$t"' EXIT
mkdir -p "$t/bin" && go build -buildvcs=false -o "$t/bin/new" ./cmd/vloop || fail "go build of this tree"
S=$t/bin/new
R=$t/repo
mkdir -p "$R/src" "$R/.vloop/tmp" "$t/home" "$t/tmp" || fail mkdir
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
wsh_allow() { allow work bash "'$1'" "$1"; allow work powershell "'$1'" "$1"; }
wsh_deny() { deny work bash "'$1'" "$1"; deny work powershell "'$1'" "$1"; }

# The worked example, shell table, phase work.
wsh_allow 'rm src/old.ts'
wsh_allow 'pnpm add nanoid'
wsh_allow 'git diff > .vloop/tmp/diff.txt'
wsh_allow 'echo ok > src/version.txt'
wsh_allow 'ls src > /dev/null'
denyx work bash 'rm -rf dist' 'rm -rf dist is denied by the work fence (Bash(rm -rf:*))' 'rm -rf dist'
denyp work bash 'sh -c "git push origin main"' 'denied by the work fence (Bash(git push:*))' 'sh -c "git push origin main"'
denyp work bash 'env gh pr list' 'denied by the Copilot deny list (gh:*)' 'env gh pr list'
denyx work bash 'curl https://example.com' 'curl https://example.com is denied by the work fence (Bash(curl:*))' 'curl https://example.com'
wsh_deny 'git commit -m x'
denyx work bash 'echo x > .vloop/state/state.json' 'write to .vloop/state/state.json is not allowed in a work session: nothing under .git/ or .vloop/ but .vloop/tmp/' 'echo x > .vloop/state/state.json'
denyx work bash 'echo x > ../outside.txt' 'write to ../outside.txt is not allowed in a work session: it is outside the repository' 'echo x > ../outside.txt'
wsh_deny 'for f in a b; do echo $f; done'
wsh_deny 'echo $(date)'

# The worked example, other tools, phase work.
allow work create src/links/links.service.ts --path src/links/links.service.ts
allow work edit test/app.e2e-spec.ts --path test/app.e2e-spec.ts
allow work create .vloop/tmp/proposal.json --path .vloop/tmp/proposal.json
denyx work create .vloop/state/state.json 'write to .vloop/state/state.json is not allowed in a work session: nothing under .git/ or .vloop/ but .vloop/tmp/' --path .vloop/state/state.json
denyx work edit .git/config 'write to .git/config is not allowed in a work session: nothing under .git/ or .vloop/ but .vloop/tmp/' --path .git/config
denyx work create .vloop/config.toml 'write to .vloop/config.toml is not allowed in a work session: nothing under .git/ or .vloop/ but .vloop/tmp/' --path .vloop/config.toml
denyx work create /etc/hosts 'write to /etc/hosts is not allowed in a work session: it is outside the repository' --path /etc/hosts
allow work view src/app.ts --path src/app.ts
deny work view '~/.copilot/config.json' --path '~/.copilot/config.json'
allow work skill work --path work
denyx work skill vloop-work "skill vloop-work is not vloop's — use /vloop:work" --path vloop-work
denyx work web_fetch '' 'tool web_fetch is not allowed in a work session'
P_SRC=$(printf '*** Begin Patch\n*** Update File: src/app.ts\n@@\n-a\n+b\n*** End Patch\n')
P_DEF=$(printf '*** Begin Patch\n*** Add File: .vloop/defects/D1.md\n+x\n*** End Patch\n')
allow work apply_patch 'Update src/app.ts' "$P_SRC"
deny work apply_patch 'Add .vloop/defects/D1.md' "$P_DEF"

# The review phase's rm line, unchanged; review stays as it was.
denyx review bash 'rm src/old.ts' 'rm src/old.ts is denied by the Copilot deny list (rm:*)' 'rm src/old.ts'
denyx second-review bash 'rm src/old.ts' 'rm src/old.ts is denied by the Copilot deny list (rm:*)' 'rm src/old.ts'
denyx review create src/app.ts 'write to src/app.ts is not allowed in a review session: only .vloop/tmp/ may be written' --path src/app.ts
denyx second-review edit src/app.ts 'write to src/app.ts is not allowed in a second-review session: only .vloop/tmp/ may be written' --path src/app.ts
denyx review bash 'echo ok > src/version.txt' 'write to src/version.txt is not allowed in a review session: only .vloop/tmp/ may be written' 'echo ok > src/version.txt'
denyx review bash 'git push' 'git push is denied by the review fence (Bash(git push:*))' 'git push'
denyx review skill vloop-review "skill vloop-review is not vloop's — use /vloop:review" --path vloop-review
denyx review web_fetch '' 'tool web_fetch is not allowed in a review session'
allow review create .vloop/tmp/verdict.json --path .vloop/tmp/verdict.json
allow review bash 'git diff > .vloop/tmp/diff.txt' 'git diff > .vloop/tmp/diff.txt'

# An unknown phase: the message, exit 2.
for ph in plan bogus; do
  chk "$ph" bash 'ls'
  [ "$rc" = 2 ] || fail "fence check --phase $ph: want exit 2, got $rc, stdout '$out', stderr '$(cat "$t/err")'"
  grep -qF "unknown phase \"$ph\": want review, second-review or work" "$t/err" ||
    fail "fence check --phase $ph: stderr does not hold 'unknown phase \"$ph\": want review, second-review or work': '$(cat "$t/err")'"
done

# Writes in work: anywhere under the root but .git/ and .vloop/ outside .vloop/tmp/.
for tl in create edit; do
  allow work "$tl" README.md --path README.md
  allow work "$tl" "<root>/src/abs.ts" --path "$R/src/abs.ts"
  allow work "$tl" src/../lib/x.go --path src/../lib/x.go
  allow work "$tl" .vloop/tmp/a/b.json --path .vloop/tmp/a/b.json
  allow work "$tl" .github/workflows/ci.yml --path .github/workflows/ci.yml
  allow work "$tl" /dev/null --path /dev/null
  deny work "$tl" .git/HEAD --path .git/HEAD
  deny work "$tl" "<root>/.git/hooks/pre-commit" --path "$R/.git/hooks/pre-commit"
  deny work "$tl" .vloop/tmp/../state/x --path .vloop/tmp/../state/x
  deny work "$tl" .vloop/defects/D1.md --path .vloop/defects/D1.md
  deny work "$tl" .vloop/interventions/I1.md --path .vloop/interventions/I1.md
  deny work "$tl" .vloop/repo-guidelines.md --path .vloop/repo-guidelines.md
  deny work "$tl" .vloop/install.json --path .vloop/install.json
  deny work "$tl" .vloop/state/gates/T1/gate.sh --path .vloop/state/gates/T1/gate.sh
  deny work "$tl" ../sibling/x --path ../sibling/x
  deny work "$tl" src/../../x --path src/../../x
  deny work "$tl" '~/.claude/settings.json' --path '~/.claude/settings.json'
  deny work "$tl" "\$HOME/notes.txt" --path "$t/home/notes.txt"
  deny work "$tl" "(no path)"
done
wsh_deny 'echo x >> .git/config'
wsh_deny 'echo x > .vloop/defects/D2.md'
wsh_deny 'cat a > ../x'
wsh_allow 'cat a >> src/b.txt'
wsh_allow 'go test ./... > .vloop/tmp/test.log'
# Reads in work: as in review.
for tl in view read_file grep rg glob; do
  allow work "$tl" .vloop/state/state.json --path .vloop/state/state.json
  deny work "$tl" '~/.claude/settings.json' --path '~/.claude/settings.json'
  deny work "$tl" '$HOME/.config/gh/hosts.yml' --path "$t/home/.config/gh/hosts.yml"
  deny work "$tl" '$HOME/.copilot/config.json' --path "$t/home/.copilot/config.json"
done
# Skills: vloop's only.
allow work skill review --path review
denyx work skill my-work "skill my-work is not vloop's — use /vloop:work" --path my-work
# Everything else.
for tl in task explore_subagent frobnicate; do
  denyx work "$tl" '' "tool $tl is not allowed in a work session"
done

# Shell in work: the work fence's rules and the work deny-list stems, through
# the same unwrapping; rm of a file is allowed, recursive rm is not.
wsh_allow 'rm -f src/a.ts src/b.ts'
wsh_allow 'npm install'
wsh_allow 'pnpm test'
wsh_allow 'node scripts/x.js'
wsh_allow 'go test ./...'
wsh_allow 'mkdir -p src/links && touch src/links/a.ts'
wsh_allow 'git status && git diff HEAD'
denyp work bash 'pip install requests' 'denied by the work fence (Bash(pip install:*))' 'pip install requests'
denyp work bash 'rm -r dist' 'Bash(rm -r:*)' 'rm -r dist'
denyp work bash 'env rm -fr dist' 'Bash(rm -fr:*)' 'env rm -fr dist'
denyp work bash 'sh -c "rm -R x"' 'Bash(rm -R:*)' 'sh -c "rm -R x"'
denyp work bash 'git reset --hard' 'Bash(git reset:*)' 'git reset --hard'
denyp work bash 'vloop task set T1 done' 'Bash(vloop task set:*)' 'vloop task set T1 done'
denyp work bash 'nohup wget x' 'Bash(wget:*)' 'nohup wget x'
denyp work bash 'ssh host.invalid' 'denied by the Copilot deny list (ssh:*)' 'ssh host.invalid'
denyp work bash 'git remote -v' 'denied by the Copilot deny list (git remote:*)' 'git remote -v'
denyp work bash 'git config --global user.name x' 'git config' 'git config --global user.name x'
denyp work bash 'find . -name x -delete' 'find' 'find . -name x -delete'
wsh_deny 'sh script.sh'
wsh_deny 'while true; do ls; done'
wsh_deny 'if true; then ls; fi'
wsh_deny '(cd src && ls)'
wsh_deny 'echo `date`'
wsh_deny 'eval "ls"'
wsh_deny 'source x.sh'
wsh_deny '$CMD x'
wsh_deny "echo 'unterminated"
echo "gate T1: ok"
