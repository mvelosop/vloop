#!/bin/sh
# Gate T6 — the run's cost while it runs (brief R5, the lines): one line per
# session on stdout and in run.log, with its time, cost and the run's spend
# against the ceiling; the 50% and 80% warnings once each on stderr and run.log.
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

# Every session costs $1.10 against a $4 ceiling; the work session takes 2 s.
CONFIG='run.cost-ceiling = 4
' runrepo c
cat >"$S/script.sh" <<'EOF'
STUB_COST=1.10
[ "$PHASE" = work ] && STUB_SLEEP=2
EOF
vrun c.out; got=$?
[ "$got" = 0 ] || fail "the stubbed run exited $got, want 0 — $(tail -5 "$t/c.out.err")"
RD=$(ls -d "$R"/.vloop/state/runs/"$N"/*/ | head -1)
# The expected line per session, in session order, from the records: phase,
# task, the session's cost and the run's running total after it.
for f in "$RD"sessions/*.json; do jq -c '{phase, task, cost_usd}' "$f"; done | jq -rs '
  reduce .[] as $s ({sum: 0, out: []};
    .sum += $s.cost_usd
    | .out += [(if $s.phase == "gate-review" then "gate review" else $s.phase end)
        + (if $s.task then " " + $s.task else "" end)
        + "\t" + ($s.cost_usd * 100 | round / 100 | tostring) + "\t" + (.sum * 100 | round / 100 | tostring)])
  | .out[]' >"$t/expect" || fail "records"
[ "$(wc -l <"$t/expect" | tr -d ' ')" = 4 ] || fail "want four session records (plan, gate review, work, review): $(cat "$t/expect")"
money() { printf '%.2f' "$1"; }
grep -E '^   (plan|gate review|work|review)( T[0-9]+)? · ' "$t/c.out" >"$t/lines"
i=0
while IFS="$(printf '\t')" read -r who cost sum; do
  i=$((i + 1))
  got_line=$(sed -n "${i}p" "$t/lines")
  re="^   $who · [0-9]+m[0-5][0-9]s · \\\$$(money "$cost") · run \\\$$(money "$sum") / \\\$4\\.00\$"
  echo "$got_line" | grep -Eq "$re" || fail "session line $i is '$got_line', want '   $who · <m>m<ss>s · \$$(money "$cost") · run \$$(money "$sum") / \$4.00'"
  grep -Eq "$re" "$RD/run.log" || fail "run.log has no line '   $who · … · \$$(money "$cost") · run \$$(money "$sum") / \$4.00'"
done <"$t/expect"
[ "$(wc -l <"$t/lines" | tr -d ' ')" = 4 ] || fail "want one line per session on stdout, got: $(cat "$t/lines")"
grep -Eq '^   work T1 · 0m0[2-4]s · ' "$t/lines" || fail "the work session slept 2 s but its line says: $(grep 'work T1' "$t/lines")"
# 50% and 80% of the ceiling, warned once each, on stderr and in run.log.
grep '^   cost: ' "$t/c.out.err" >"$t/warn"
w50="   cost: \$2.20 spent, 50% of the \$4.00 ceiling"
w80="   cost: \$3.30 spent, 80% of the \$4.00 ceiling"
[ "$(cat "$t/warn")" = "$(printf '%s\n%s' "$w50" "$w80")" ] || fail "want exactly the two warnings '$w50' and '$w80' on stderr, got: $(cat "$t/warn")"
grep -Fxq "$w50" "$RD/run.log" && grep -Fxq "$w80" "$RD/run.log" || fail "run.log does not carry the two cost warnings"
grep -q '^   cost: ' "$t/c.out" && fail "a cost warning went to stdout"
echo "gate T6: ok"
