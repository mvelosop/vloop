#!/bin/sh
# Gate T1 — H1, the judge: `vloop fence check`. Every line of the worked
# example's two tables, then each wrapper, separator, refused form, redirection
# and path rule of the behaviour contract, judged in a fixture root with a fake
# HOME. An allow prints exactly "allow" and exits 0; a deny prints a line
# starting "deny: " and exits 1; bad arguments exit 2. Nothing runs a command
# under judgment: the gate only asks the judge.
G=$(cd "$(dirname "$0")" && pwd -P) || { echo "GATE FAIL: no gate folder" >&2; exit 1; }
. "$G/lib.sh"
cd "$top" || fail "repo root"
newbin
R=$t/repo
mkdir -p "$R/src" "$R/.vloop/tmp" "$t/home" "$t/tmp" "$t/stub" || fail mkdir
# chk <phase> <tool> [--path p] [command]: the judge's stdout in $out, exit in $rc.
chk() {
  ph=$1 tl=$2; shift 2
  out=$(cd "$R" && venv "$S" fence check --phase "$ph" --root "$R" --tool "$tl" "$@" 2>"$t/err"); rc=$?
}
# allow/deny <phase> <tool> <label> [args...]
allow() {
  ph=$1 tl=$2 lb=$3; shift 3; chk "$ph" "$tl" "$@"
  [ "$rc" = 0 ] && [ "$out" = allow ] || fail "[$ph] $tl $lb: want 'allow' and exit 0, got exit $rc, stdout '$out', stderr '$(cat "$t/err")'"
}
deny() {
  ph=$1 tl=$2 lb=$3; shift 3; chk "$ph" "$tl" "$@"
  [ "$rc" = 1 ] || fail "[$ph] $tl $lb: want exit 1 (deny), got exit $rc, stdout '$out', stderr '$(cat "$t/err")'"
  first=$(printf '%s\n' "$out" | head -1)
  case "$first" in "deny: "?*) ;; *) fail "[$ph] $tl $lb: want a line 'deny: <reason>', got '$out'";; esac
}
# denyx <phase> <tool> <label> <text the reason must hold> [args...]
denyx() {
  ph=$1 tl=$2 lb=$3 want=$4; shift 4
  deny "$ph" "$tl" "$lb" "$@"
  printf '%s\n' "$out" | head -1 | grep -qF -- "$want" || fail "[$ph] $tl $lb: the deny reason '$out' does not name '$want'"
}
# sh_allow/sh_deny <command> [text]: the bash tool in phase review, and the
# same command through the powershell tool and in phase second-review.
sh_allow() { for ph in review second-review; do allow "$ph" bash "'$1'" "$1"; done; allow review powershell "'$1'" "$1"; }
sh_deny() {
  for ph in review second-review; do
    if [ -n "${2:-}" ]; then denyx "$ph" bash "'$1'" "$2" "$1"; else deny "$ph" bash "'$1'" "$1"; fi
  done
  deny review powershell "'$1'" "$1"
}

# The worked example, shell table.
sh_allow 'git diff 1a2b3c..HEAD -- src/'
sh_allow 'git log --oneline | head -5'
sh_allow 'vloop task show T3'
sh_allow 'git diff > .vloop/tmp/diff.txt'
sh_allow 'ls src > /dev/null'
sh_deny 'sh -c "gh pr list"' gh
sh_deny 'env gh auth token' gh
sh_deny 'env -i PATH=/bin git push origin main' 'git push'
sh_deny "bash -lc 'git push'" 'git push'
sh_deny '/usr/bin/curl https://example.com' curl
sh_deny 'nohup timeout 5 wget x' wget
sh_deny 'git status && xargs rm -rf' rm
sh_deny 'echo $(gh auth token)'
sh_deny 'eval "git push"' eval
sh_deny 'git diff > .vloop/state/x' .vloop/state/x
sh_deny 'sh -c "$CMD"'
sh_deny "echo 'unterminated"

# The worked example, other tools.
allow review create .vloop/tmp/verdict.json --path .vloop/tmp/verdict.json
deny review create .vloop/state/state.json --path .vloop/state/state.json
deny review edit src/app.ts --path src/app.ts
allow review view src/app.ts --path src/app.ts
deny review view '~/.claude/settings.json' --path '~/.claude/settings.json'
deny review web_fetch ''
deny review search_code_subagent ''
allow review skill review --path review
denyx review skill vloop-review "skill vloop-review is not vloop's — use /vloop:review" --path vloop-review
denyx review frobnicate '' 'tool frobnicate is not allowed in a review session'

# Every wrapper the contract unwraps, alone and stacked.
sh_deny 'command gh pr list' gh
sh_deny 'exec git push' 'git push'
sh_deny 'nice -n 5 git push' 'git push'
sh_deny 'nice git push' 'git push'
sh_deny 'time git push' 'git push'
sh_deny 'timeout -s KILL 5 curl x' curl
sh_deny 'timeout 5s git push' 'git push'
sh_deny 'stdbuf -oL git push' 'git push'
sh_deny 'xargs -0 -n 1 git push' 'git push'
sh_deny 'env -u HOME git push' 'git push'
sh_deny 'env - git push' 'git push'
sh_deny 'env FOO=bar git push' 'git push'
sh_deny 'FOO=1 BAR=2 git push' 'git push'
sh_deny "zsh -c 'git push'" 'git push'
sh_deny "dash -c 'git push'" 'git push'
sh_deny "ksh -c 'git push'" 'git push'
sh_deny "bash -e -x -c 'git push'" 'git push'
sh_deny "sh -c 'env sh -c \"gh x\"'" gh
sh_deny 'env nohup nice -n 3 timeout 5 stdbuf -o0 gh x' gh
sh_deny './bin/gh pr list' gh
# Separators.
sh_deny 'git status; git push' 'git push'
sh_deny 'git status || git push' 'git push'
sh_deny 'git log | git push' 'git push'
sh_deny 'git status & git push' 'git push'
sh_deny "$(printf 'git status\ngit push')" 'git push'
# Refused before any matching.
sh_deny 'cat <(gh x)'
sh_deny 'tee >(gh x)'
sh_deny 'echo `gh auth token`'
sh_deny 'source x.sh'
sh_deny '. x.sh'
sh_deny '$CMD push'
sh_deny '${CMD} push'
sh_deny 'sh -c'
sh_deny 'bash -c "git $SUB"'
# Redirections are writes.
sh_deny 'git diff >> .vloop/state/x' .vloop/state/x
sh_deny 'git diff >| .vloop/state/x' .vloop/state/x
sh_deny 'git diff &> .vloop/state/x' .vloop/state/x
sh_deny 'git diff >.vloop/state/x' .vloop/state/x
sh_deny 'git diff > src/out.txt' src/out.txt
# The fence's own matching (`*` any run, `:*` a prefix) and the deny-list stems.
sh_deny 'git diff --stat --output=x' 'git diff'
sh_deny 'find . -name x -delete' find
sh_deny 'git -C . push' 'git -C'
sh_deny 'git config --global user.name x' 'git config'
sh_deny 'ssh host.invalid' ssh
sh_deny 'scp a host.invalid:b' scp
sh_deny 'nc host.invalid 80' nc
sh_deny 'rm x' rm
sh_deny 'git remote -v' 'git remote'
# What must stay allowed: quoted separators and words, harmless wrappers.
sh_allow 'git log --grep "a;gh x"'
sh_allow "grep -rn 'git push' ."
sh_allow 'echo "sh -c gh"'
sh_allow 'git status && git diff HEAD'
sh_allow 'vloop task show T1 | jq .'
sh_allow 'ls src 2> /dev/null'
sh_allow 'cat .vloop/tmp/a > .vloop/tmp/b'
sh_allow 'env git status'
sh_allow "sh -c 'git status'"
sh_allow '/usr/bin/git log -1'

# Writes: resolved against the root, cleaned, allowed only under .vloop/tmp/.
for ph in review second-review; do
  for tl in create edit; do
    allow "$ph" "$tl" .vloop/tmp/x --path .vloop/tmp/x
    allow "$ph" "$tl" "<root>/.vloop/tmp/abs.json" --path "$R/.vloop/tmp/abs.json"
    deny "$ph" "$tl" .vloop/tmp/../state/x --path .vloop/tmp/../state/x
    deny "$ph" "$tl" ../outside --path ../outside
    deny "$ph" "$tl" /etc/x --path /etc/x
    deny "$ph" "$tl" "<root>/src/app.ts" --path "$R/src/app.ts"
    deny "$ph" "$tl" .git/config --path .git/config
  done
done
# Reads: allowed, except the global Claude, Copilot and GitHub CLI directories.
for tl in view read_file grep rg grep_search glob file_search; do
  allow review "$tl" src/app.ts --path src/app.ts
  allow review "$tl" .vloop/state/state.json --path .vloop/state/state.json
  deny review "$tl" '~/.claude/settings.json' --path '~/.claude/settings.json'
  deny review "$tl" '$HOME/.claude/settings.json' --path "$t/home/.claude/settings.json"
  deny review "$tl" '~/.copilot/config.json' --path '~/.copilot/config.json'
  deny review "$tl" '$HOME/.copilot/config.json' --path "$t/home/.copilot/config.json"
  deny review "$tl" '~/.config/gh/hosts.yml' --path '~/.config/gh/hosts.yml'
  deny second-review "$tl" '$HOME/.config/gh/hosts.yml' --path "$t/home/.config/gh/hosts.yml"
done
# Skills: vloop's only.
for ph in review second-review; do
  allow "$ph" skill review --path review
  allow "$ph" skill verify --path verify
  denyx "$ph" skill repo-review "skill repo-review is not vloop's — use /vloop:review" --path repo-review
done
# Everything else.
for tl in web_fetch task explore_subagent; do deny second-review "$tl" ''; done
denyx second-review frobnicate '' 'tool frobnicate is not allowed in a second-review session'

# Bad arguments exit 2.
out=$(cd "$R" && venv "$S" fence check --phase bogus --root "$R" --tool bash 'git status' 2>&1); rc=$?
[ "$rc" = 2 ] || fail "fence check --phase bogus: want exit 2, got $rc: $out"
out=$(cd "$R" && venv "$S" fence check --phase review --root "$R" 'git status' 2>&1); rc=$?
[ "$rc" = 2 ] || fail "fence check with no --tool: want exit 2, got $rc: $out"
echo "gate T1: ok"
