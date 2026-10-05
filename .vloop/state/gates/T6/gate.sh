#!/bin/sh
# Gate T6 — `vloop metrics --interventions` (brief I6, across briefs): one row
# per kind in kind order then total, by phase in phase order, --json, across a
# workspace with a repo column, and --by task refused. Judges the built binary
# over records it writes itself in temporary repositories; every expected
# count below is computed by hand from the records this gate adds.
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
# rec <summary> <phase> <kind> <automatable> <flags…>: add one record.
rec() {
  s=$1; p=$2; k=$3; a=$4; shift 4
  run intervention add "$s" --phase "$p" --kind "$k" --automatable "$a" --by both "$@" || fail "vloop intervention add '$s' failed: $(cat "$t/err")"
}
BR="--brief B20260101-0900-a.loop-brief"
KINDS="direction decision context-supply halt verification-finding repair carry-forward ceremony"
PHASES="setup design run halt verify close next"
HEAD9="n recommended share other-option adjusted different no-options automatable-yes automatable-partly"
# table <file> <first column> <order> <expected rows file>: check a kind or phase table.
# Expected rows are '<name> <nine values>'; rows not expected must have n 0.
table() {
  f=$1; col=$2; order=$3; want=$4
  [ "$(head -n 1 "$f" | awk '{ $1 = $1; print }')" = "$col $HEAD9" ] || fail "the table's header is '$(head -n 1 "$f")', want '$col $HEAD9'"
  awk 'NR > 1 && NF > 0 { print $1 }' "$f" >"$t/names"
  [ "$(tail -n 1 "$t/names")" = total ] || fail "the last row is not total: $(tr '\n' ' ' <"$t/names")"
  last=0
  for n in $(sed '$d' "$t/names"); do
    i=0; pos=0
    for o in $order; do i=$((i + 1)); [ "$o" = "$n" ] && pos=$i; done
    [ "$pos" -gt 0 ] || fail "a row is named '$n', not a $col"
    [ "$pos" -gt "$last" ] || fail "the rows are not in $col order: $(tr '\n' ' ' <"$t/names")"
    last=$pos
    grep -q "^$n " "$want" || { awk -v r="$n" 'NR > 1 && $1 == r && $2 != 0 { bad = 1 } END { exit bad }' "$f" || fail "the $n row counts records there are none of"; }
  done
  while read -r name vals; do
    got=$(awk -v r="$name" 'NR > 1 && $1 == r { $1 = ""; sub(/^ /, ""); print; exit }' "$f")
    [ "$got" = "$vals" ] || fail "the $name row is '$got', want '$vals'"
  done <"$want"
}

newrepo a
rec "r1 recommended" halt repair partly $BR --option x --option y --option z --recommended 1 --why w --decided-option 1
rec "r2 recommended" halt repair yes $BR --option x --option y --recommended 2 --why w --decided-option 2
rec "r3 other option" verify repair yes $BR --option x --option y --recommended 1 --why w --decided-option 2
rec "r4 adjusted" verify repair no $BR --option x --option y --recommended 1 --why w --decided-option 1 --adjusted
rec "r5 no options" verify repair partly $BR
rec "d1 different, series-level" design decision yes --option x --recommended 1 --why w --decided-other --decided 'something else'
rec "d2 recommended" design decision partly $BR --option x --option y --recommended 2 --why w --decided-option 2
rec "h1 no options" halt halt no $BR
rec "h2 no options" run halt yes $BR

# --- by kind ---
run metrics --interventions || fail "vloop metrics --interventions failed: $(cat "$t/err")"
cp "$t/out" "$t/kind"
printf '%s\n' \
  'decision 2 1 50% 0 0 1 0 1 1' \
  'halt 2 0 n/a 0 0 0 2 1 0' \
  'repair 5 2 50% 1 1 0 1 2 2' \
  'total 9 3 50% 1 1 1 3 4 3' >"$t/want-kind"
table "$t/kind" kind "$KINDS" "$t/want-kind"
run metrics --interventions --by kind || fail "vloop metrics --interventions --by kind failed: $(cat "$t/err")"
cmp -s "$t/out" "$t/kind" || fail "--by kind does not print the default table"

# --- by phase ---
run metrics --interventions --by phase || fail "vloop metrics --interventions --by phase failed: $(cat "$t/err")"
cp "$t/out" "$t/phase"
printf '%s\n' \
  'design 2 1 50% 0 0 1 0 1 1' \
  'run 1 0 n/a 0 0 0 1 1 0' \
  'halt 3 2 100% 0 0 0 1 1 1' \
  'verify 3 0 0% 1 1 0 1 1 1' \
  'total 9 3 50% 1 1 1 3 4 3' >"$t/want-phase"
table "$t/phase" phase "$PHASES" "$t/want-phase"

# --- JSON ---
run metrics --interventions --json || fail "vloop metrics --interventions --json failed: $(cat "$t/err")"
jq -e '([.. | strings] + [.. | objects | keys[]]) | index("repair") != null' "$t/out" >/dev/null 2>&1 || fail "metrics --interventions --json does not print the rows as JSON: $(head -c 300 "$t/out")"
run metrics --interventions --by phase --json || fail "vloop metrics --interventions --by phase --json failed: $(cat "$t/err")"
jq -e '([.. | strings] + [.. | objects | keys[]]) | index("verify") != null' "$t/out" >/dev/null 2>&1 || fail "metrics --interventions --by phase --json does not print the phase rows as JSON"

# --- --by task is refused ---
run metrics --interventions --by task; c=$?
[ "$c" -eq 2 ] || fail "metrics --interventions --by task exited $c, want 2"
[ "$(cat "$t/err")" = "vloop: --by takes kind or phase with --interventions" ] || fail "metrics --interventions --by task printed '$(cat "$t/err")', want 'vloop: --by takes kind or phase with --interventions'"

# --- across a workspace ---
newrepo b
rec "b1 recommended" halt repair yes --option x --option y --recommended 1 --why w --decided-option 1
rec "b2 other option" close ceremony partly --option x --option y --recommended 1 --why w --decided-option 2
printf '%s\n' '[[repo]]' 'path = "a"' 'name = "a"' '' '[[repo]]' 'path = "b"' 'name = "b"' >"$t/ws.toml"
R="$t/a"
run metrics --workspace "$t/ws.toml" --interventions || fail "vloop metrics --workspace <file> --interventions failed: $(cat "$t/err")"
cp "$t/out" "$t/ws"
[ "$(head -n 1 "$t/ws" | awk '{ $1 = $1; print }')" = "repo kind $HEAD9" ] || fail "the workspace table's header is '$(head -n 1 "$t/ws")', want 'repo kind $HEAD9'"
# wsrow <repo> <kind> <nine values>
wsrow() {
  got=$(awk -v r="$1" -v k="$2" 'NR > 1 && $1 == r && $2 == k { $1 = ""; $2 = ""; sub(/^  /, ""); print; exit }' "$t/ws")
  [ "$got" = "$3" ] || fail "the workspace row $1 $2 is '$got', want '$3'"
}
wsrow a decision '2 1 50% 0 0 1 0 1 1'
wsrow a halt '2 0 n/a 0 0 0 2 1 0'
wsrow a repair '5 2 50% 1 1 0 1 2 2'
wsrow b repair '1 1 100% 0 0 0 0 1 0'
wsrow b ceremony '1 0 0% 1 0 0 0 0 1'
tot=$(awk 'NF > 0 { l = $0 } END { print l }' "$t/ws" | awk '{ for (i = 1; i <= NF; i++) if ($i == "total") { s = ""; for (j = i + 1; j <= NF; j++) s = s (s == "" ? "" : " ") $j; print s; exit } }')
[ "$tot" = "11 4 50% 2 1 1 3 5 4" ] || fail "the workspace's last row is not a total over both repositories: '$(awk 'NF > 0 { l = $0 } END { print l }' "$t/ws")'"
echo "gate T6: ok"
