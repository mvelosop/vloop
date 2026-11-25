# Shared harness for B39's gates (the same file in every gate folder). The
# sourcing gate sets G, its own folder. The vloop built from this tree, every
# fixture repository, the stub agents and the fake home live under $t, a
# temporary directory removed on exit; nothing touches this repository's tree,
# refs, .vloop/ or .gitignore, nor the operator's home. No VLOOP_* variable of
# the caller reaches a vloop run here.
set -u
fail() { printf 'GATE FAIL: %s\n' "$*" >&2; exit 1; }
for v in $(env | sed -n 's/^\(VLOOP_[A-Za-z0-9_]*\)=.*/\1/p'); do unset "$v"; done
unset GOFLAGS GOBIN GIT_DIR GIT_WORK_TREE GIT_INDEX_FILE GIT_CEILING_DIRECTORIES NO_COLOR \
  COPILOT_GITHUB_TOKEN GH_TOKEN GITHUB_TOKEN
export LC_ALL=C GIT_CONFIG_GLOBAL=/dev/null GIT_CONFIG_NOSYSTEM=1 GIT_TERMINAL_PROMPT=0
[ -f go.mod ] && [ -d cmd/vloop ] || fail "run from the repository root"
command -v python3 >/dev/null || fail "python3 is needed to parse JSON"
# The Go caches stay the caller's, also once HOME is the fake one below.
GOCACHE=$(go env GOCACHE) && GOMODCACHE=$(go env GOMODCACHE) && GOPATH=$(go env GOPATH) || fail "go env"
export GOCACHE GOMODCACHE GOPATH
t=$(mktemp -d "${TMPDIR:-/tmp}/gate.XXXXXX") && t=$(cd "$t" && pwd -P) || fail mktemp
trap 'rm -rf "$t"' EXIT
go build -buildvcs=false -o "$t/vloop" ./cmd/vloop >"$t/build.out" 2>&1 ||
  fail "go build ./cmd/vloop: $(tail -n 5 "$t/build.out")"
mkdir -p "$t/home" "$t/stub" "$t/cstub" || fail mkdir
# From here on the home is the fake one, and git looks for no repository above $t.
export HOME="$t/home" USERPROFILE="$t/home" XDG_CONFIG_HOME="$t/home/.config" APPDATA="$t/home" \
  GIT_CEILING_DIRECTORIES="$t"

BRIEF=docs/briefs/B20260101-0900-demo.loop-brief.md
EMDASH=$(printf '\342\200\224')

# jtrue FILE EXPR MSG: EXPR, a Python expression over d (FILE parsed as JSON),
# is true; else the gate fails with MSG and the file.
jtrue() {
  python3 -I -c 'import json,sys; d=json.load(open(sys.argv[1])); sys.exit(0 if eval(sys.argv[2]) else 1)' "$1" "$2" 2>"$t/py.err" ||
    fail "$3
$(cat "$1")
$(cat "$t/py.err")"
}

# The stub claude: logs every argument line to claude.log, answers --version
# and plugin list (with this binary's version), and for a -p prompt sets PHASE
# and sources script.sh beside it when there is one, then prints a result.
cat >"$t/stub/claude" <<'EOF'
#!/bin/sh
dir=$(dirname "$0")
echo "$*" >> "$dir/claude.log"
case "$1" in
  --version) echo "2.0.0 (Claude Code)"; exit 0;;
  plugin) printf '[{"id":"vloop@vloop","version":"%s","enabled":true}]\n' "$(cat "$dir/plugin-version")"; exit 0;;
esac
prompt=""; MODEL=""
while [ $# -gt 0 ]; do
  case "$1" in
    -p) prompt="$2"; shift 2;;
    --model) MODEL="$2"; shift 2;;
    *) shift;;
  esac
done
case "${prompt%% *}" in
  /vloop:plan) PHASE=plan;; /vloop:work) PHASE=work;; /vloop:review) PHASE=review;;
  /vloop:gate-review) PHASE=gate-review;; /vloop:verify) PHASE=verify;;
  *) echo "stub claude: unrecognised prompt: $prompt" >&2; exit 64;;
esac
if [ -f "$dir/script.sh" ]; then . "$dir/script.sh"; fi
printf '{"type":"result","subtype":"success","is_error":false,"duration_ms":1000,"num_turns":3,"total_cost_usd":0.1,"session_id":"stub","permission_denials":[],"modelUsage":{"%s":{"costUSD":0.1}}}\n' "$MODEL"
EOF
# The stub copilot, in its own folder: logs each argument on a line "arg ...",
# answers --version, and for anything else prints a session that did nothing.
cat >"$t/cstub/copilot" <<'EOF'
#!/bin/sh
dir=$(dirname "$0")
for a in "$@"; do printf 'arg %s\n' "$a"; done >> "$dir/copilot.log"
case "$1" in
  --version) echo "GitHub Copilot CLI 1.0.101"; exit 0;;
esac
printf '{"type":"session.start","data":{"sessionId":"stub"}}\n'
printf '{"type":"assistant.turn_start","data":{"turnId":"1"}}\n'
printf '{"type":"session.usage_checkpoint","data":{"totalNanoAiu":1000000000}}\n'
printf '{"type":"result","sessionId":"stub","exitCode":0,"usage":{"premiumRequests":1,"sessionDurationMs":1000}}\n'
EOF
chmod +x "$t/stub/claude" "$t/cstub/copilot" || fail chmod
"$t/vloop" version --json >"$t/version.json" 2>"$t/version.err" || fail "vloop version --json: $(cat "$t/version.err")"
python3 -I -c 'import json,sys; print(json.load(open(sys.argv[1]))["version"])' "$t/version.json" >"$t/stub/plugin-version" ||
  fail "vloop version --json has no version"
# The stub claude is first on PATH for every gate but the one that removes it.
PATH="$t/stub:$PATH"
export PATH

# commitall DIR MSG: commit everything in the fixture repository DIR.
commitall() { git -C "$1" add -A && git -C "$1" commit -q -m "$2" || fail "commit in $1"; }

# newrepo NAME: a git repository $t/NAME on main, an identity in its own
# config, one commit holding docs/x.md; sets R.
newrepo() {
  R=$t/$1
  mkdir -p "$R/docs" && git -C "$R" init -q -b main && git -C "$R" config user.name gate &&
    git -C "$R" config user.email gate@example.com && git -C "$R" config commit.gpgsign false &&
    printf 'x\n' >"$R/docs/x.md" || fail "newrepo $1"
  commitall "$R" init
}

# writebrief DIR: the ready demo brief (it passes vloop brief check) at $BRIEF.
writebrief() {
  mkdir -p "$1/docs/briefs" || fail mkdir
  cat >"$1/$BRIEF" <<'EOF'
---
name: B20260101-0900-demo.loop-brief
description: Build a demo
kind: brief
status: ready
created: 2026-01-01
seeds: A run that builds a demo
depends-on: []
---
# Brief

## What it is

A thing.

## Why this shape, and what was rejected

Because.

## Binding references

- `docs/x.md` — the error contract

## Behaviour contract

An unknown id exits 1.

## Worked example

```
$ demo
ok
```

## Out of scope

- one
- two

## Constraints

- Go.

## Shape

2 to 3 tasks.
EOF
}

# CHECKCFG is a shared config holding nothing but one check.
CHECKCFG='[[check]]
name = "all"
paths = ["**"]
run = "true"
'

# initrepo NAME: newrepo, vloop init, the shared config set to CHECKCFG, the
# demo brief, all committed; sets R.
initrepo() {
  newrepo "$1"
  (cd "$R" && "$t/vloop" init) >"$t/init.out" 2>&1 || fail "vloop init in $1: $(cat "$t/init.out")"
  printf '%s' "$CHECKCFG" >"$R/.vloop/config.toml" || fail "config in $1"
  writebrief "$R"
  commitall "$R" "init and brief"
}

# trust DIR: the fake home's .claude.json trusts DIR, besides those it trusted.
trust() {
  printf '%s\n' "$1" >>"$t/trusted"
  python3 -I -c 'import json,sys
ps=[l.rstrip("\n") for l in open(sys.argv[1]) if l.strip()]
json.dump({"projects":{p:{"hasTrustDialogAccepted":True} for p in ps}}, open(sys.argv[2],"w"))' "$t/trusted" "$HOME/.claude.json" ||
    fail "trust $1"
}

# runin DIR ARGS: vloop run in DIR; stdout in $t/out, stderr in $t/err, the
# exit code in rc.
runin() { d=$1; shift; (cd "$d" && "$t/vloop" "$@") >"$t/out" 2>"$t/err"; rc=$?; }

# want NAME RC OUT ERR: the last runin exited RC, and its stdout and stderr
# equal the files OUT and ERR byte for byte (/dev/null for nothing).
want() {
  [ "$rc" = "$2" ] || fail "$1: exit $rc, want $2
stdout:
$(cat "$t/out")
stderr:
$(cat "$t/err")"
  cmp -s "$t/out" "$3" || fail "$1: stdout differs.
got:
$(cat "$t/out")
want:
$(cat "$3")"
  cmp -s "$t/err" "$4" || fail "$1: stderr differs.
got:
$(cat "$t/err")
want:
$(cat "$4")"
}

# wantrc NAME RC: the last runin exited RC.
wantrc() {
  [ "$rc" = "$2" ] || fail "$1: exit $rc, want $2
stdout:
$(cat "$t/out")
stderr:
$(cat "$t/err")"
}

# hasline NAME FILE LINE: FILE holds LINE as a whole line.
hasline() {
  grep -qxF -- "$3" "$2" || fail "$1: no line
$3
in:
$(cat "$2")"
}

# noline NAME FILE LINE: FILE does not hold LINE as a whole line.
noline() {
  ! grep -qxF -- "$3" "$2" || fail "$1: unexpected line
$3
in:
$(cat "$2")"
}

# lit FILE TEXT: FILE holds TEXT followed by a newline.
lit() { printf '%s\n' "$2" >"$1"; }

# rows NAME FILE: the doctor rows of FILE named NAME (the second field), one per line.
rows() { awk -v n="$1" '$2 == n' "$2"; }
