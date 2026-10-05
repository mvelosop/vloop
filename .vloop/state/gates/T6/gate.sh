#!/bin/sh
# Gate T6 — a plan that is not valid JSON, or not a valid plan, is said so with
# the next step (brief Q3): status and status --json exit 1 on either, with the
# error under --json; task validate and schema validate name the file, line and
# column, with no empty field. A valid plan still works.
set -u
fail() { echo "GATE FAIL: $*" >&2; exit 1; }
for v in $(env | sed -n 's/^\(VLOOP_[A-Z0-9_]*\)=.*/\1/p'); do unset "$v"; done
unset NO_COLOR GIT_DIR GIT_WORK_TREE GIT_INDEX_FILE CLAUDE_PLUGIN_ROOT
export LC_ALL=C GIT_CONFIG_GLOBAL=/dev/null GIT_CONFIG_NOSYSTEM=1 GIT_TERMINAL_PROMPT=0
export GIT_AUTHOR_NAME=gate GIT_AUTHOR_EMAIL=gate@example.invalid GIT_COMMITTER_NAME=gate GIT_COMMITTER_EMAIL=gate@example.invalid
G=$(cd "$(dirname "$0")" && pwd -P) || fail "no gate folder"
t=$(mktemp -d "${TMPDIR:-/tmp}/gate.XXXXXX") && t=$(cd "$t" && pwd -P) || fail mktemp
trap 'rm -rf "$t"' EXIT
go build -o "$t/vloop" ./cmd/vloop || fail "go build ./cmd/vloop"
GOOS=linux go build -o "$t/vloop-linux" ./cmd/vloop || fail "GOOS=linux go build ./cmd/vloop"
GOOS=windows go build -o "$t/vloop.exe" ./cmd/vloop || fail "GOOS=windows go build ./cmd/vloop"
B="$t/vloop"
mkdir -p "$t/home" "$t/bin" && printf '#!/bin/sh\nexit 1\n' >"$t/bin/claude" && chmod +x "$t/bin/claude" || fail "fake home"
export HOME="$t/home" PATH="$t/bin:$PATH"
R="$t/r"
mkdir -p "$R" && git -C "$R" init -q -b main && git -C "$R" commit -q --allow-empty -m "gate base" || fail "git init"
(cd "$R" && "$B" init </dev/null) >"$t/init.out" 2>&1 || fail "vloop init: $(cat "$t/init.out")"
mkdir -p "$R/.vloop/state" || fail mkdir
P=.vloop/state/state.json
run() { (cd "$R" && "$B" "$@") >"$t/out" 2>"$t/err"; }
exact() {
  want=$1; msg=$2; shift 2
  run "$@"; got=$?
  [ "$got" = "$want" ] || fail "vloop $* exited $got, want $want — stderr: $(cat "$t/err")"
  [ "$(cat "$t/err")" = "$msg" ] && [ "$(wc -l <"$t/err" | tr -d ' ')" = 1 ] || fail "vloop $* printed '$(cat "$t/err")' on stderr, want exactly '$msg'"
}
# jsonerr <substring>: status --json exits 1 and prints {"error": "...<substring>..."} on stdout.
jsonerr() {
  run status --json; got=$?
  [ "$got" = 1 ] || fail "vloop status --json exited $got on a broken plan, want 1 — stdout: $(cat "$t/out")"
  jq -e --arg s "$1" '(.error | type) == "string" and (.error | contains($s))' "$t/out" >/dev/null 2>&1 \
    || fail "vloop status --json did not print {\"error\": \"...$1...\"}: $(cat "$t/out")"
}
# located <line> <col> <args...>: the command exits 1 and prints "<path>: not valid JSON (line l, column c)".
located() {
  l=$1; c=$2; shift 2
  run "$@"; got=$?
  [ "$got" = 1 ] || fail "vloop $* exited $got on a plan that is not JSON, want 1"
  cat "$t/out" "$t/err" >"$t/both"
  grep -Fq "$P: not valid JSON (line $l, column $c)" "$t/both" || fail "vloop $* does not print '$P: not valid JSON (line $l, column $c)': $(cat "$t/both")"
  grep -q ': :' "$t/both" && fail "vloop $* prints an empty field: $(cat "$t/both")"
  :
}

# --- not JSON, on line 1 ---
printf '{bad' >"$R/$P"
exact 1 "vloop: $P is not valid JSON (line 1, column 2)" status
jsonerr "$P is not valid JSON (line 1, column 2)"
located 1 2 task validate
located 1 2 schema validate state/v2 "$P"

# --- not JSON, on line 3 ---
printf '{\n  "schema": "state/v2",\n  "run_id": bad\n}\n' >"$R/$P"
exact 1 "vloop: $P is not valid JSON (line 3, column 13)" status
located 3 13 task validate

# --- JSON, but not a plan ---
printf '{"version":"x"}' >"$R/$P"
exact 1 "vloop: $P is not a valid plan — vloop task validate lists the problems" status
jsonerr "is not a valid plan — vloop task validate lists the problems"
run task validate; [ $? = 1 ] || fail "task validate on a plan that fails its schema did not exit 1"

# --- a valid plan still works ---
cp "$G/plan.json" "$R/$P" || fail "plan fixture"
run status || fail "status on a valid plan failed: $(cat "$t/err")"
run status --json || fail "status --json on a valid plan failed: $(cat "$t/err")"
jq -e '.total == 2' "$t/out" >/dev/null || fail "status --json on a valid plan does not count its two tasks: $(cat "$t/out")"
echo "gate T6: ok"
