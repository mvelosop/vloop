#!/bin/sh
# Gate T5 — `vloop intervention show <id>` (brief I5): the record, then links
# for the ids in its Context section in order of first appearance — a task from
# the latest committed plan of the record's brief, a run's folders, a defect, a
# commit, another intervention, and an id that resolves to nothing — in text and
# --json; an unknown id refused. Judges the built binary in a temporary
# repository built from the fixtures beside this gate.
set -u
fail() { echo "GATE FAIL: $*" >&2; exit 1; }
for v in $(env | sed -n 's/^\(VLOOP_[A-Z0-9_]*\)=.*/\1/p'); do unset "$v"; done
unset NO_COLOR GIT_DIR GIT_WORK_TREE GIT_INDEX_FILE CLAUDE_PLUGIN_ROOT
export LC_ALL=C TZ=UTC GIT_CONFIG_GLOBAL=/dev/null GIT_CONFIG_NOSYSTEM=1 GIT_TERMINAL_PROMPT=0
export GIT_AUTHOR_NAME=gate GIT_AUTHOR_EMAIL=gate@example.invalid GIT_COMMITTER_NAME=gate GIT_COMMITTER_EMAIL=gate@example.invalid
G=$(cd "$(dirname "$0")" && pwd -P) || fail "no gate folder"
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
# A commit to link to, by a short sha.
echo widget >"$R/widget.txt" && git -C "$R" add widget.txt && git -C "$R" commit -q -m "the subject line for show" || fail "commit widget"
SHA=$(git -C "$R" rev-parse --short=12 HEAD) || fail "rev-parse"
# Two plan commits for the brief, then a later non-plan change to the plan: the
# title comes from the latest [vloop] plan commit.
mkdir -p "$R/.vloop/state" || fail mkdir
cp "$G/plans/plan1.json" "$R/.vloop/state/state.json" && git -C "$R" add -A && git -C "$R" commit -q -m "[vloop] plan B20260101-0900-a" || fail "plan commit 1"
cp "$G/plans/plan2.json" "$R/.vloop/state/state.json" && git -C "$R" add -A && git -C "$R" commit -q -m "[vloop] plan B20260101-0900-a" || fail "plan commit 2"
cp "$G/plans/drift.json" "$R/.vloop/state/state.json" && git -C "$R" add -A && git -C "$R" commit -q -m "a later change to the plan" || fail "drift commit"
# A run folder, a defect, another intervention.
mkdir -p "$R/.vloop/state/runs/B20260101-0900-a/20260101-090000" && echo "run" >"$R/.vloop/state/runs/B20260101-0900-a/20260101-090000/run.log" || fail "run folder"
mkdir -p "$R/.vloop/defects" && cp "$G"/defects/D*.md "$R/.vloop/defects/" || fail "defect fixture"
DEF=D20260101-0905-the-widget-gate-pinned-a-path
run defect list || fail "the defect fixture does not list: $(cat "$t/err")"
run intervention add "the earlier repair of the widget" --phase run --kind repair --automatable no --by operator \
  || fail "vloop intervention add failed: $(cat "$t/err")"
OTHER=$(basename "$(cat "$t/out")" .md)
CTX="T3 in run B20260101-0900-a; see $DEF and commit $SHA, after $OTHER; not D20990101-0000-nope."
run intervention add "T3's gate failed on a path typo" --brief B20260101-0900-a.loop-brief --phase halt --kind repair \
  --automatable partly --by both --trigger 'vloop run exited 2' --done 'gate replaced' --context "$CTX" \
  --option 'replace the gate with vloop task verify' --option 'reset T3 and retry' \
  --recommended 1 --why 'the gate, not the work, is wrong' --decided-option 1 --decided 'replaced as recommended' \
  || fail "vloop intervention add with options and context failed: $(cat "$t/err")"
ID=$(basename "$(cat "$t/out")" .md)

# --- text ---
run intervention show "$ID" || fail "vloop intervention show $ID failed: $(cat "$t/err")"
cp "$t/out" "$t/show"
for s in "$ID" "T3's gate failed on a path typo" "agreement" "recommended" "replace the gate with vloop task verify" "reset T3 and retry" "the gate, not the work, is wrong" "replaced as recommended" "Context" "vloop run exited 2"; do
  grep -Fq "$s" "$t/show" || fail "vloop intervention show does not print '$s'"
done
awk 'h && /^[ \t]*$/ { exit } h { print } /^[ \t]*links:?[ \t]*$/ { h = 1 }' "$t/show" >"$t/links"
[ -s "$t/links" ] || fail "vloop intervention show prints no 'links' heading followed by link lines"
awk '{ print $1 }' "$t/links" >"$t/refs"
printf '%s\n' T3 B20260101-0900-a "$DEF" "$SHA" "$OTHER" D20990101-0000-nope >"$t/want"
cmp -s "$t/refs" "$t/want" || fail "the links are '$(tr '\n' ' ' <"$t/refs")', want one per id in the Context section in order of first appearance: '$(tr '\n' ' ' <"$t/want")'"
# line <ref> <text> <message>
line() { awk -v r="$1" '$1 == r' "$t/links" | grep -Fq "$2" || fail "$3"; }
line T3 "Fix the widget gate" "T3's link is not the title from the latest [vloop] plan commit of the record's brief: $(awk '$1 == "T3"' "$t/links")"
awk '$1 == "T3"' "$t/links" | grep -Fq "drifted" && fail "T3's link takes its title from a commit that is not a [vloop] plan commit"
line B20260101-0900-a "20260101-090000" "the run id's link does not name its run folder: $(awk '$1 == "B20260101-0900-a"' "$t/links")"
line "$DEF" "the widget gate pinned a path" "the defect's link is not its summary"
line "$SHA" "the subject line for show" "the commit's link is not its subject"
line "$OTHER" "the earlier repair of the widget" "the other intervention's link is not its summary"
line D20990101-0000-nope "(not found)" "an id that resolves to nothing does not print (not found)"

# --- JSON ---
run intervention show "$ID" --json || fail "vloop intervention show --json failed: $(cat "$t/err")"
cp "$t/out" "$t/show.json"
jq -e --arg id "$ID" '.id == $id and .agreement == "recommended" and (.options | length) == 2' "$t/show.json" >/dev/null 2>&1 || fail "show --json is not the record's JSON form"
jq -e --arg d "$DEF" --arg s "$SHA" --arg o "$OTHER" '[.links[].ref] == ["T3", "B20260101-0900-a", $d, $s, $o, "D20990101-0000-nope"]' "$t/show.json" >/dev/null 2>&1 \
  || fail "show --json's links are not the Context ids in order of first appearance: $(jq -c '[.links[]?.ref]' "$t/show.json" 2>/dev/null)"
jq -e 'all(.links[]; (.kind | type) == "string" and has("title"))' "$t/show.json" >/dev/null 2>&1 || fail "a link in show --json lacks kind or title"
jq -e --arg d "$DEF" --arg s "$SHA" --arg o "$OTHER" '
  (.links[] | select(.ref == "T3") | .title) == "Fix the widget gate" and
  (.links[] | select(.ref == $d) | .title) == "the widget gate pinned a path" and
  (.links[] | select(.ref == $s) | .title) == "the subject line for show" and
  (.links[] | select(.ref == $o) | .title) == "the earlier repair of the widget"' "$t/show.json" >/dev/null 2>&1 \
  || fail "show --json's link titles are not the task title, defect summary, commit subject and intervention summary"

# --- an unknown id ---
run intervention show I20990101-0000-none; c=$?
[ "$c" -eq 1 ] || fail "show of an unknown id exited $c, want 1"
[ "$(cat "$t/err")" = "vloop: no intervention I20990101-0000-none" ] || fail "show of an unknown id printed '$(cat "$t/err")', want 'vloop: no intervention I20990101-0000-none'"
run intervention show I20990101-0000-none --json; c=$?
[ "$c" -eq 1 ] || fail "show --json of an unknown id exited $c, want 1"
jq -e '.error | contains("no intervention I20990101-0000-none")' "$t/out" >/dev/null 2>&1 || fail "show --json of an unknown id does not print {\"error\": …}: $(cat "$t/out")"
echo "gate T5: ok"
