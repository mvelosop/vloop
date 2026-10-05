#!/bin/sh
# Gate T3 — `vloop intervention set` for recommended, decided and adjusted,
# validated against the record's option count, re-deriving agreement in the same
# write; agreement and options are not settable (brief I3 and the worked
# example). Judges the built binary in a temporary repository.
set -u
fail() { echo "GATE FAIL: $*" >&2; exit 1; }
for v in $(env | sed -n 's/^\(VLOOP_[A-Z0-9_]*\)=.*/\1/p'); do unset "$v"; done
unset NO_COLOR GIT_DIR GIT_WORK_TREE GIT_INDEX_FILE CLAUDE_PLUGIN_ROOT
export LC_ALL=C TZ=UTC GIT_CONFIG_GLOBAL=/dev/null GIT_CONFIG_NOSYSTEM=1 GIT_TERMINAL_PROMPT=0
export GIT_AUTHOR_NAME=gate GIT_AUTHOR_EMAIL=gate@example.invalid GIT_COMMITTER_NAME=gate GIT_COMMITTER_EMAIL=gate@example.invalid
t=$(mktemp -d "${TMPDIR:-/tmp}/gate.XXXXXX") && t=$(cd "$t" && pwd -P) || fail mktemp
trap 'rm -rf "$t"' EXIT
go build -o "$t/vloop" ./cmd/vloop || fail "go build ./cmd/vloop"
GOOS=linux go build -o "$t/vloop-linux" ./cmd/vloop || fail "GOOS=linux go build ./cmd/vloop"
GOOS=windows go build -o "$t/vloop.exe" ./cmd/vloop || fail "GOOS=windows go build ./cmd/vloop"
B="$t/vloop"
# newrepo <name>: a git repository under $t with vloop init done; sets R.
newrepo() {
  R="$t/$1"
  mkdir -p "$R" && git -C "$R" init -q -b main && git -C "$R" commit -q --allow-empty -m "gate base" || fail "git init $1"
  (cd "$R" && "$B" init </dev/null) >"$t/init.out" 2>&1 || fail "vloop init in $1: $(cat "$t/init.out")"
}
# run <args>: vloop in $R; stdout in $t/out, stderr in $t/err; returns its exit code.
run() { (cd "$R" && "$B" "$@") >"$t/out" 2>"$t/err"; }

newrepo r
run intervention add "T3's gate failed on a path typo" --brief B1 --phase halt --kind repair --automatable partly --by both \
  --trigger 'vloop run exited 2' --done 'gate replaced' --context 'T3 in run B20260101-0900-a' \
  --option 'replace the gate with vloop task verify' --option 'reset T3 and retry' --option 'abandon the brief' \
  --recommended 1 --why 'the gate, not the work, is wrong' --decided-option 1 --decided 'replaced as recommended' \
  || fail "vloop intervention add with three options failed: $(cat "$t/err")"
F="$R/$(cat "$t/out")"; ID=$(basename "$F" .md)
[ -f "$F" ] || fail "add printed $(cat "$t/out") but wrote no such file"
run intervention add "a record without options" --phase run --kind ceremony --automatable yes --by operator \
  || fail "vloop intervention add without options failed: $(cat "$t/err")"
F0="$R/$(cat "$t/out")"; ID0=$(basename "$F0" .md)
[ -f "$F0" ] || fail "add printed $(cat "$t/out") but wrote no such file"

fm() { awk 'NR == 1 { next } /^---$/ { exit } { print }' "$1"; }
body() { awk 'n >= 2 { print } /^---$/ { n++ }' "$1"; }
# field <file> <key>: the frontmatter value of key.
field() { fm "$1" | sed -n "s/^$2: //p"; }
# set_ok <label> <field> <value> <want agreement> [<want adjusted line: yes|no>]
set_ok() {
  body "$F" >"$t/body.before"
  run intervention set "$ID" "$2" "$3" || fail "$1: vloop intervention set $ID $2 $3 failed: $(cat "$t/err")"
  [ "$(field "$F" agreement)" = "$4" ] || fail "$1: after set $2 $3, agreement is '$(field "$F" agreement)', want '$4'"
  body "$F" >"$t/body.after"
  cmp -s "$t/body.before" "$t/body.after" || fail "$1: set $2 $3 changed the record's body"
  run intervention list || fail "$1: vloop intervention list rejects the record after set $2 $3: $(cat "$t/err")"
  run intervention list --json || fail "$1: list --json failed"
  jq -e --arg id "$ID" --arg a "$4" '[.[] | select(.id == $id)][0].agreement == $a' "$t/out" >/dev/null 2>&1 || fail "$1: list --json does not read agreement $4 after set $2 $3"
  case "${5:-}" in
    yes) [ "$(field "$F" adjusted)" = true ] || fail "$1: after set $2 $3 the frontmatter has no adjusted: true" ;;
    no) [ "$(field "$F" adjusted)" = true ] && fail "$1: after set $2 $3 the frontmatter still says adjusted: true" ;;
  esac
  return 0
}
# set_refused <label> <file> <id> <field> <value> [<exact message>]: refused,
# file untouched; with a quoted message, exit 2 and that message exactly.
set_refused() {
  cp "$2" "$t/before"
  run intervention set "$3" "$4" "$5"; c=$?
  [ "$c" -ne 0 ] || fail "$1: vloop intervention set $3 $4 $5 was accepted"
  if [ -n "${6:-}" ]; then
    [ "$c" -eq 2 ] || fail "$1: vloop intervention set $3 $4 $5 exited $c, want 2 (stderr: $(cat "$t/err"))"
    [ "$(cat "$t/err")" = "vloop: $6" ] || fail "$1: set $4 printed '$(cat "$t/err")', want 'vloop: $6'"
  fi
  cmp -s "$2" "$t/before" || fail "$1: a refused set $4 $5 changed the file"
}

[ "$(field "$F" decided)" = 1 ] && [ "$(field "$F" agreement)" = recommended ] || fail "the starting record is not decided 1, agreement recommended"
set_ok "decided 2" decided 2 other-option
[ "$(field "$F" decided)" = 2 ] || fail "set decided 2 did not write decided: 2"
set_ok "adjusted true" adjusted true adjusted yes
set_ok "adjusted false" adjusted false other-option no
set_ok "recommended 2" recommended 2 recommended
[ "$(field "$F" recommended)" = 2 ] || fail "set recommended 2 did not write recommended: 2"
set_ok "decided other" decided other different
[ "$(field "$F" decided)" = other ] || fail "set decided other did not write decided: other"
set_ok "decided 1" decided 1 other-option
set_ok "decided 2 again" decided 2 recommended

set_refused "agreement" "$F" "$ID" agreement recommended 'agreement is derived — set decided, adjusted or recommended instead'
set_refused "options" "$F" "$ID" options 2 'options are recorded with the intervention, not set'
set_refused "decided beyond the options" "$F" "$ID" decided 4
set_refused "recommended beyond the options" "$F" "$ID" recommended 4
set_refused "adjusted not a boolean" "$F" "$ID" adjusted maybe
set_refused "decided on a record without options" "$F0" "$ID0" decided 1
set_refused "agreement on a record without options" "$F0" "$ID0" agreement recommended 'agreement is derived — set decided, adjusted or recommended instead'

# the fields of v1 stay settable
run intervention set "$ID" phase verify || fail "set phase verify failed: $(cat "$t/err")"
[ "$(field "$F" phase)" = verify ] || fail "set phase verify did not write phase: verify"
[ "$(field "$F" agreement)" = recommended ] || fail "set phase changed the agreement"
echo "gate T3: ok"
