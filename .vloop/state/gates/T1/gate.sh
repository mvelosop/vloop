#!/bin/sh
# Gate T1 — the .git/config guard names keys and ignores editor bookkeeping
# (brief R1), on stubbed vloop runs in temporary repositories: an editor's
# branch.<name>.vscode-merge-base is ignored; core.fsmonitor halts, exit 9,
# named; at most three keys, sorted, then how many more; a changed value, a
# removed key and an include directive halt; the base gate is named as the gate
# of its task; run.git-config-ignore from the file and the environment adds
# patterns; config set refuses a protected pattern, exit 2, and a protected
# pattern in the environment or the file refuses the run, exit 1.
set -u
fail() { echo "GATE FAIL: $*" >&2; exit 1; }
for v in $(env | sed -n 's/^\(VLOOP_[A-Z0-9_]*\)=.*/\1/p'); do unset "$v"; done
unset NO_COLOR GIT_DIR GIT_WORK_TREE GIT_INDEX_FILE CLAUDE_PLUGIN_ROOT
export LC_ALL=C GIT_CONFIG_GLOBAL=/dev/null GIT_CONFIG_NOSYSTEM=1 GIT_TERMINAL_PROMPT=0
export GIT_AUTHOR_NAME=gate GIT_AUTHOR_EMAIL=gate@example.invalid GIT_COMMITTER_NAME=gate GIT_COMMITTER_EMAIL=gate@example.invalid
G=$(cd "$(dirname "$0")" && pwd -P) || fail "no gate folder"
top=$(pwd -P)
t=$(mktemp -d "${TMPDIR:-/tmp}/gate.XXXXXX") && t=$(cd "$t" && pwd -P) || fail mktemp
trap 'rm -rf "$t"' EXIT
go build -o "$t/vloop" ./cmd/vloop || fail "go build ./cmd/vloop"
GOOS=linux go build -o "$t/vloop-linux" ./cmd/vloop || fail "GOOS=linux go build ./cmd/vloop"
GOOS=windows go build -o "$t/vloop.exe" ./cmd/vloop || fail "GOOS=windows go build ./cmd/vloop"
B="$t/vloop"
# Go keeps its caches under HOME unless told; keep the real ones across the change.
GOMODCACHE=$(go env GOMODCACHE) GOCACHE=$(go env GOCACHE) GOPATH=$(go env GOPATH) || fail "go env"
export GOMODCACHE GOCACHE GOPATH GOPROXY=off
mkdir -p "$t/Users/gateuser" || fail mkdir
export HOME="$t/Users/gateuser"
N=B20260101-0900-demo
BRIEF=docs/briefs/$N.loop-brief.md
trust() { printf '{"projects":{"%s":{"hasTrustDialogAccepted":true}}}\n' "$R" >"$HOME/.claude.json"; }
# runrepo <name> [<plan>]: a vloop repository on main with a ready brief and a
# check (after $CONFIG, when set), committed; its own stub claude playing
# <plan> (plan.json by default) from $t/<name>.stub.
runrepo() {
  R="$t/$1" S="$t/$1.stub"
  mkdir -p "$R" "$S" && git -C "$R" init -q -b main || fail "git init $1"
  cp "$G/claude" "$S/claude" && chmod +x "$S/claude" && cp "$G/${2:-plan.json}" "$S/plan.json" || fail "stub for $1"
  git -C "$R" config user.name gate && git -C "$R" config user.email gate@example.invalid || fail "identity"
  trust
  (cd "$R" && "$B" init </dev/null) >"$t/init.out" 2>&1 || fail "vloop init in $1: $(cat "$t/init.out")"
  printf '%s[[check]]\nname = "all"\npaths = ["**"]\nrun = "true"\n' "${CONFIG:-}" >"$R/.vloop/config.toml" || fail config
  mkdir -p "$R/docs/briefs" && echo x >"$R/docs/x.md" && cp "$G/brief.md" "$R/$BRIEF" || fail brief
  git -C "$R" add -A && git -C "$R" commit -q -m harness || fail "commit $1"
}
# vrun <out> [VAR=value...]: vloop run <brief> in $R with its stub on PATH;
# stdout in $t/<out>, stderr in $t/<out>.err; returns vloop's exit code.
vrun() {
  o=$1; shift
  trust
  (cd "$R" && env PATH="$S:$PATH" "$@" "$B" run "$BRIEF") >"$t/$o" 2>"$t/$o.err"
}
# records <glob>: the session records of the brief's run folders matching <glob>, one path per line.
records() { ls "$R"/.vloop/state/runs/"$N"/*/sessions/$1 2>/dev/null; }

# halt9 <out> <keys>: the run halted with exit 9, naming exactly <keys> for work, and committed no T1.
halt9() {
  [ "$got" = 9 ] || fail "$1: exit $got, want 9 — $(tail -5 "$t/$1.err")"
  want="vloop: work changed .git/config ($2) — nothing was committed; restore it, then re-run"
  grep -Fxq "$want" "$t/$1.err" || fail "$1: stderr has no line '$want': $(grep 'git/config' "$t/$1.err")"
  git -C "$R" -c core.fsmonitor=false log --all --format=%s | grep -q '^\[vloop\] T1:' && fail "$1: T1 was committed after .git/config changed"
  return 0
}
done0() {
  [ "$got" = 0 ] || fail "$1: exit $got, want 0 (the run continues) — $(tail -5 "$t/$1.err")"
  [ "$(cd "$R" && "$B" status --json | jq -r .status)" = complete ] || fail "$1: the plan is not complete"
}

# --- an editor's bookkeeping key is ignored ---
runrepo vs
cat >"$S/script.sh" <<'EOF'
stub_after() { [ "$PHASE" = work ] && printf '[branch "%s"]\n\tvscode-merge-base = refs/heads/main\n' "$(git rev-parse --abbrev-ref HEAD)" >>.git/config; return 0; }
EOF
vrun vs.out; got=$?
done0 vs.out
git -C "$R" config --get "branch.$N.vscode-merge-base" >/dev/null || fail "the fixture did not write branch.$N.vscode-merge-base"

# --- a core key halts and is named ---
runrepo fsm
cat >"$S/script.sh" <<'EOF'
stub_after() { [ "$PHASE" = work ] && printf '[core]\n\tfsmonitor = false\n' >>.git/config; return 0; }
EOF
vrun fsm.out; got=$?
halt9 fsm.out core.fsmonitor

# --- at most three keys, sorted, and how many more; ignored keys are not among them ---
runrepo many
cat >"$S/script.sh" <<'EOF'
stub_after() {
  [ "$PHASE" = work ] || return 0
  printf '[branch "%s"]\n\tvscode-merge-base = refs/heads/main\n' "$(git rev-parse --abbrev-ref HEAD)" >>.git/config
  printf '[alias]\n\te = x\n\tc = x\n\ta = x\n\td = x\n\tb = x\n' >>.git/config
}
EOF
vrun many.out; got=$?
halt9 many.out "alias.a, alias.b, alias.c and 2 more"

# --- a changed value, a removed key and an include directive halt ---
runrepo chg
cat >"$S/script.sh" <<'EOF'
stub_after() {
  [ "$PHASE" = work ] || return 0
  git config user.email other@example.invalid && git config --unset user.name
  printf '[include]\n\tpath = extra.cfg\n' >>.git/config
}
EOF
vrun chg.out; got=$?
halt9 chg.out "include.path, user.email, user.name"

# --- a base gate that changes .git/config is named as the gate of its task ---
runrepo gate
jq '.tasks[0].verify = "printf '"'"'[core]\\n\\tfsmonitor = false\\n'"'"' >> .git/config; test -f T1.out"' "$G/plan.json" >"$S/plan.json" || fail "gate plan"
vrun gate.out; got=$?
[ "$got" = 9 ] || fail "a base gate that changed .git/config: exit $got, want 9 — $(tail -5 "$t/gate.out.err")"
want="vloop: the gate of T1 changed .git/config (core.fsmonitor) — nothing was committed; restore it, then re-run"
grep -Fxq "$want" "$t/gate.out.err" || fail "a base gate that changed .git/config: stderr has no line '$want': $(grep 'git/config' "$t/gate.out.err")"

# --- run.git-config-ignore from the config file and from the environment ---
CONFIG='run.git-config-ignore = ["gitlens.*"]
' runrepo gl
cat >"$S/script.sh" <<'EOF'
stub_after() { [ "$PHASE" = work ] && git config gitlens.x 1; return 0; }
EOF
vrun gl.out; got=$?
done0 gl.out
runrepo glenv
cat >"$S/script.sh" <<'EOF'
stub_after() { [ "$PHASE" = work ] && git config zed.panel.width 3; return 0; }
EOF
vrun glenv.out VLOOP_RUN_GIT_CONFIG_IGNORE='gitlens.*,zed.*'; got=$?
done0 glenv.out

# --- config set refuses a protected pattern, exit 2, and writes nothing ---
runrepo cfg
cp "$R/.vloop/config.toml" "$t/cfg.before"
# refuse <value> <section> [<pattern>]: <pattern> is the offending pattern the
# line names — <value> itself unless <value> is a list.
refuse() {
  (cd "$R" && "$B" config set run.git-config-ignore "$1") >"$t/out" 2>"$t/err"; got=$?
  want="vloop: run.git-config-ignore: ${3:-$1} would ignore $2.* — the guard must see it"
  [ "$got" = 2 ] || fail "config set run.git-config-ignore '$1': exit $got, want 2 — $(cat "$t/err")"
  [ "$(cat "$t/err")" = "$want" ] || fail "config set run.git-config-ignore '$1' printed '$(cat "$t/err")', want '$want'"
  cmp -s "$t/cfg.before" "$R/.vloop/config.toml" || fail "config set run.git-config-ignore '$1' changed the config file"
}
refuse 'core.*' core
refuse '*' core
refuse 'c*' core
refuse 'filter.lfs.*' filter
refuse 'i*' include
refuse 'includeif.*' includeif
refuse 'gitlens.*,core.*' core 'core.*'
for p in 'gitlens.*' 'branch.*.vscode-merge-base'; do
  (cd "$R" && "$B" config set run.git-config-ignore "$p") >"$t/out" 2>"$t/err" || fail "config set run.git-config-ignore '$p' was refused: $(cat "$t/err")"
  (cd "$R" && "$B" config get run.git-config-ignore) >"$t/out" 2>"$t/err" || fail "config get run.git-config-ignore: $(cat "$t/err")"
  grep -Fq "$p" "$t/out" || fail "config get run.git-config-ignore after setting '$p' printed: $(cat "$t/out")"
done

# --- a protected pattern in the environment or the file refuses the run, exit 1 ---
runrepo envp
vrun envp.out VLOOP_RUN_GIT_CONFIG_IGNORE='gitlens.*,core.*'; got=$?
[ "$got" = 1 ] || fail "vloop run with VLOOP_RUN_GIT_CONFIG_IGNORE='gitlens.*,core.*': exit $got, want 1"
grep -Fxq "vloop: run.git-config-ignore: core.* would ignore core.* — the guard must see it" "$t/envp.out.err" \
  || fail "vloop run with a protected pattern in the environment printed: $(cat "$t/envp.out.err")"
grep -q '/vloop:plan' "$S/argv.log" 2>/dev/null && fail "vloop run planned with a protected pattern in the environment"
printf 'run.git-config-ignore = ["filter.*"]\n[[check]]\nname = "all"\npaths = ["**"]\nrun = "true"\n' >"$R/.vloop/config.toml" || fail config
git -C "$R" commit -qam "protected pattern" || fail "commit the config"
vrun envf.out; got=$?
[ "$got" = 1 ] || fail "vloop run with run.git-config-ignore = [\"filter.*\"] in the config file: exit $got, want 1"
grep -Fxq "vloop: run.git-config-ignore: filter.* would ignore filter.* — the guard must see it" "$t/envf.out.err" \
  || fail "vloop run with a protected pattern in the config file printed: $(cat "$t/envf.out.err")"
echo "gate T1: ok"
