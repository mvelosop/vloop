#!/bin/sh
# Gate T17 — the close (brief "Worked example" and "Real-data check"): the
# worked example line for line on a fresh build in temporary directories, then
# this repository read by the new binary and by v2.0.0-beta.2, built from its
# tag: metrics --json for every consumed brief equal except first-pass and the
# iterations per closed task, and defect list, intervention list and brief list
# equal. The suite's speed is measured once by the operator, not here.
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
go build -o "$t/gendocs" ./cmd/gendocs || fail "go build ./cmd/gendocs"
mkdir -p "$t/b2" && (git archive v2.0.0-beta.2 | tar -xf - -C "$t/b2") || fail "git archive v2.0.0-beta.2"
(cd "$t/b2" && go build -o "$t/vloop-b2" ./cmd/vloop) || fail "build v2.0.0-beta.2"
B="$t/vloop"
OLDHOME=$HOME
mkdir -p "$t/home" "$t/bin" || fail mkdir
cat >"$t/bin/claude" <<'EOF'
#!/bin/sh
case "$1" in
  --version) echo "2.0.0 (Claude Code)"; exit 0;;
  plugin) echo '[]'; exit 0;;
esac
exit 1
EOF
chmod +x "$t/bin/claude" || fail chmod
export HOME="$t/home" PATH="$t/bin:$PATH"
newrepo() {
  R="$t/$1"
  mkdir -p "$R" && git -C "$R" init -q -b main && git -C "$R" commit -q --allow-empty -m "gate base" || fail "git init $1"
  (cd "$R" && "$B" init </dev/null) >"$t/init.out" 2>&1 || fail "vloop init in $1: $(cat "$t/init.out")"
}
run() { (cd "$R" && "$B" "$@") >"$t/out" 2>"$t/err"; }
exact() {
  want=$1; msg=$2; shift 2
  run "$@"; got=$?
  [ "$got" = "$want" ] || fail "vloop $* exited $got, want $want — stderr: $(cat "$t/err")"
  [ "$(cat "$t/err")" = "$msg" ] || fail "vloop $* printed '$(cat "$t/err")' on stderr, want '$msg'"
}
crlf() { awk '{ printf "%s\r\n", $0 }' "$1"; }
allcrlf() { awk '{ if (substr($0, length($0), 1) != "\r") bad++ } END { exit bad > 0 }' "$1"; }

# --- the worked example ---
newrepo w
[ -e /nonexistent ] && fail "/nonexistent exists on this machine"
exact 1 "vloop: -C /nonexistent: no such directory" -C /nonexistent status
exact 1 "vloop: brief not found: docs/briefs/nope.md" run docs/briefs/nope.md
exact 1 "vloop: no such file: missing.md" brief check missing.md
run --bogus; [ $? = 2 ] || fail "vloop --bogus did not exit 2"
exact 1 "vloop: no plan — vloop run <brief> makes one" task show T1
mkdir -p "$R/.vloop/state" && cp "$G/plan.json" "$R/.vloop/state/state.json" || fail "plan"
exact 1 "vloop: no task T99 — vloop task list shows the plan's tasks" task show T99
exact 1 "vloop: no brief nope — vloop brief list shows the briefs" metrics nope
printf '{bad' >"$R/.vloop/state/state.json"
exact 1 "vloop: .vloop/state/state.json is not valid JSON (line 1, column 2)" status
printf '{"version":"x"}' >"$R/.vloop/state/state.json"
run status --json; [ $? = 1 ] || fail "status --json on a plan that is not valid did not exit 1"
jq -e '.error | contains("is not a valid plan — vloop task validate lists the problems")' "$t/out" >/dev/null 2>&1 || fail "status --json on a plan that is not valid printed: $(cat "$t/out")"
rm -f "$R/.vloop/state/state.json"

newrepo drafts
run brief new second || fail "brief new: $(cat "$t/err")"
(cd "$R" && "$B" brief check docs/briefs/*.loop-brief.md) >"$t/out" 2>&1 || fail "brief check on drafts only did not exit 0: $(cat "$t/out")"
grep -qx 'nothing checked: 2 draft brief(s) skipped' "$t/out" || fail "brief check on drafts only printed: $(cat "$t/out")"

mkdir -p "$t/outside" || fail mkdir
(cd "$t/outside" && "$B" plugin path) >"$t/out" 2>"$t/err"; got=$?
[ "$got" = 1 ] && [ "$(cat "$t/err")" = "vloop: not in a vloop repository — run vloop init first" ] || fail "plugin path outside any repository: exit $got, '$(cat "$t/err")'"
[ -e "$t/outside/.vloop" ] && fail "plugin path outside any repository created .vloop/"
(cd "$t/outside" && "$B" metrics) >"$t/out" 2>"$t/err"; got=$?
[ "$got" = 1 ] && grep -q '^vloop: .* is not a git repository$' "$t/err" || fail "metrics outside a git repository: exit $got, '$(cat "$t/err")'"

R="$t/w"
BRIEF=$(cd "$R/docs/briefs" && ls *.loop-brief.md | head -1) || fail "no brief"
run defect add "a CRLF defect" --found-by operator --brief "${BRIEF%.md}" || fail "defect add: $(cat "$t/err")"
D=$(cat "$t/out"); crlf "$R/$D" >"$t/d" && cp "$t/d" "$R/$D" || fail "make the defect CRLF"
run defect set "$(basename "$D" .md)" status fixed || fail "defect set on a CRLF defect: $(cat "$t/err")"
grep -q '^status: fixed' "$R/$D" && allcrlf "$R/$D" || fail "the CRLF defect does not say status: fixed with every line CRLF"
run intervention add "a BOM and CRLF intervention" --phase run --kind repair --automatable no --by operator \
  --option one --option two --recommended 1 --why because --decided-option 1 || fail "intervention add: $(cat "$t/err")"
I=$(cat "$t/out"); { printf '\357\273\277'; crlf "$R/$I"; } >"$t/i" && cp "$t/i" "$R/$I" || fail "make the intervention BOM + CRLF"
run intervention set "$(basename "$I" .md)" phase verify || fail "intervention set on a BOM + CRLF record: $(cat "$t/err")"
head -c 3 "$R/$I" | od -An -tx1 | tr -d ' \n' | grep -q '^efbbbf$' && allcrlf "$R/$I" || fail "intervention set lost the BOM or the CRLF endings"
run intervention set "$(basename "$I" .md)" decided 5; [ $? = 2 ] || fail "intervention set <id> decided 5 did not exit 2"

mkdir -p "$t/gd" || fail mkdir
(cd "$t/gd" && "$t/gendocs" --help) >/dev/null 2>&1 && fail "cmd/gendocs --help was not refused"
[ -e "$t/gd/--help" ] && fail "cmd/gendocs --help wrote a file named --help"

(cd "$R" && "$B" doctor) >"$t/doctor" 2>&1
grep -q '^✓ plugin the plugin is supplied by vloop run (--plugin-dir); ' "$t/doctor" || fail "doctor with no marketplace plugin: $(grep ' plugin' "$t/doctor")"
"$B" --help >"$t/help" 2>&1
grep -qi 'brief' "$t/help" && grep -Fq 'vloop init' "$t/help" || fail "vloop --help does not say what a brief is and to start with vloop init"

# --- the real-data check: this repository, read by both binaries ---
cd "$top" || fail "cd"
"$B" brief list --json >"$t/briefs.json" || fail "brief list --json on this repository"
jq -r '.[] | select(.status == "consumed") | .path' "$t/briefs.json" >"$t/consumed"
[ "$(wc -l <"$t/consumed" | tr -d ' ')" -ge 10 ] || fail "fewer than ten consumed briefs listed: $(cat "$t/consumed")"
: >"$t/changed"
while read -r b; do
  "$t/vloop-b2" metrics --json "$b" >"$t/m.old" 2>"$t/err" || fail "v2.0.0-beta.2 metrics --json $b: $(cat "$t/err")"
  "$B" metrics --json "$b" >"$t/m.new" 2>"$t/err" || fail "metrics --json $b: $(cat "$t/err")"
  for f in m.old m.new; do jq -S 'del(.tasks.first_pass, .iterations_per_closed)' "$t/$f" >"$t/$f.cmp" || fail "metrics --json $b is not JSON"; done
  cmp -s "$t/m.old.cmp" "$t/m.new.cmp" || fail "metrics --json for $b differs from v2.0.0-beta.2's beyond first-pass and convergence: $(diff "$t/m.old.cmp" "$t/m.new.cmp" | head -8)"
  o=$(jq -c '[.tasks.first_pass, .iterations_per_closed]' "$t/m.old"); n=$(jq -c '[.tasks.first_pass, .iterations_per_closed]' "$t/m.new")
  [ "$o" = "$n" ] || echo "$b: first-pass, iterations per closed $o -> $n" >>"$t/changed"
done <"$t/consumed"
for c in "defect list" "intervention list" "brief list"; do
  # shellcheck disable=SC2086
  "$t/vloop-b2" $c >"$t/l.old" 2>&1 || fail "v2.0.0-beta.2 $c failed: $(cat "$t/l.old")"
  # shellcheck disable=SC2086
  "$B" $c >"$t/l.new" 2>&1 || fail "vloop $c failed on this repository: $(cat "$t/l.new")"
  cmp -s "$t/l.old" "$t/l.new" || fail "vloop $c reads this repository's records differently from v2.0.0-beta.2: $(diff "$t/l.old" "$t/l.new" | head -6)"
done
[ -s "$t/changed" ] && { echo "changed numbers (for the run record):"; cat "$t/changed"; }
HOME=$OLDHOME
echo "gate T17: ok"
