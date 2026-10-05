#!/bin/sh
# Gate T10 — run folders that order right (brief Q7). A new run folder is named
# in UTC whatever the local zone: a stubbed one-task vloop run under a zone far
# from UTC must name its folder with a UTC stamp between the UTC clock before
# and after the run. Existing folders keep their names and are ordered by their
# records' timestamps: two folders a DST fall-back named out of order (the later
# run's local name sorts first) must be read earlier-run first, which metrics
# --json shows in the order of the missing session records it lists.
set -u
fail() { echo "GATE FAIL: $*" >&2; exit 1; }
for v in $(env | sed -n 's/^\(VLOOP_[A-Z0-9_]*\)=.*/\1/p'); do unset "$v"; done
unset NO_COLOR GIT_DIR GIT_WORK_TREE GIT_INDEX_FILE CLAUDE_PLUGIN_ROOT TZ
export LC_ALL=C GIT_CONFIG_GLOBAL=/dev/null GIT_CONFIG_NOSYSTEM=1 GIT_TERMINAL_PROMPT=0
export GIT_AUTHOR_NAME=gate GIT_AUTHOR_EMAIL=gate@example.invalid GIT_COMMITTER_NAME=gate GIT_COMMITTER_EMAIL=gate@example.invalid
G=$(cd "$(dirname "$0")" && pwd -P) || fail "no gate folder"
t=$(mktemp -d "${TMPDIR:-/tmp}/gate.XXXXXX") && t=$(cd "$t" && pwd -P) || fail mktemp
trap 'rm -rf "$t"' EXIT
go build -o "$t/vloop" ./cmd/vloop || fail "go build ./cmd/vloop"
GOOS=linux go build -o "$t/vloop-linux" ./cmd/vloop || fail "GOOS=linux go build ./cmd/vloop"
GOOS=windows go build -o "$t/vloop.exe" ./cmd/vloop || fail "GOOS=windows go build ./cmd/vloop"
B="$t/vloop"
mkdir -p "$t/home" "$t/stub" || fail mkdir
cp "$G/claude" "$t/stub/claude" && chmod +x "$t/stub/claude" && cp "$G/plan.json" "$t/stub/plan.json" || fail "stub claude"
export HOME="$t/home" PATH="$t/stub:$PATH"
N=B20260101-0900-demo
# runrepo <name>: a vloop repository on main with a ready brief and a check,
# committed, trusted by the fake home.
runrepo() {
  R="$t/$1"
  mkdir -p "$R" && git -C "$R" init -q -b main || fail "git init $1"
  git -C "$R" config user.name gate && git -C "$R" config user.email gate@example.invalid || fail "identity"
  printf '{"projects":{"%s":{"hasTrustDialogAccepted":true}}}\n' "$R" >"$HOME/.claude.json"
  (cd "$R" && "$B" init </dev/null) >"$t/init.out" 2>&1 || fail "vloop init in $1: $(cat "$t/init.out")"
  printf '[[check]]\nname = "all"\npaths = ["**"]\nrun = "true"\n' >"$R/.vloop/config.toml" || fail config
  mkdir -p "$R/docs/briefs" && echo x >"$R/docs/x.md" && cp "$G/brief.md" "$R/docs/briefs/$N.loop-brief.md" || fail brief
  git -C "$R" add -A && git -C "$R" commit -q -m harness || fail "commit $1"
}

# --- a new run folder is named in UTC ---
runrepo utc
before=$(date -u +%Y%m%d-%H%M%S)
(cd "$R" && TZ=Asia/Kolkata "$B" run "docs/briefs/$N.loop-brief.md") >"$t/run.out" 2>&1 || fail "the stubbed run failed: $(tail -15 "$t/run.out")"
after=$(date -u +%Y%m%d-%H%M%S)
folders=$(ls "$R/.vloop/state/runs/$N")
[ "$(echo "$folders" | wc -l | tr -d ' ')" = 1 ] || fail "expected one run folder, found: $folders"
stamp=$(echo "$folders" | cut -c1-15)
echo "$stamp" | grep -q '^[0-9]\{8\}-[0-9]\{6\}$' || fail "the run folder '$folders' is not named <YYYYMMDD-HHMMSS>"
if [ "$stamp" \< "$before" ] || [ "$stamp" \> "$after" ]; then
  fail "the run folder is named $folders under TZ=Asia/Kolkata, but the UTC clock read $before to $after: it is not named in UTC"
fi

# --- existing folders ordered by their records' timestamps ---
runrepo order
for f in 20261025-015000 20261025-011000; do
  mkdir -p "$R/.vloop/state/runs/$N/$f" && cp -R "$G/runs/$f/." "$R/.vloop/state/runs/$N/$f/" || fail "fixture folder $f"
done
(cd "$R" && "$B" metrics "$N" --json) >"$t/m.json" 2>"$t/err" || fail "metrics --json on the fixture folders failed: $(cat "$t/err")"
got=$(jq -r '.records.missing | map("\(.task)/\(.iteration)") | join(" ")' "$t/m.json") || fail "metrics --json has no records.missing list: $(head -c 300 "$t/m.json")"
[ "$got" = "T1/1 T2/2" ] || fail "the folders are not read in the order of their records' timestamps: missing records '$got', want 'T1/1 T2/2' (folder 20261025-015000 ran at 00:50Z, before 20261025-011000 at 01:10Z)"
ls "$R/.vloop/state/runs/$N" | tr '\n' ' ' | grep -q '^20261025-011000 20261025-015000 $' || fail "the existing run folders were renamed: $(ls "$R/.vloop/state/runs/$N")"
echo "gate T10: ok"
