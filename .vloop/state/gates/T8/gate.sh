#!/bin/sh
# Gate T8 — the skills and the operate eval case (brief I8). The shipped
# operate skill, as the built binary extracts it, and this repository's
# operator skill name the recording flags; the eval case
# operate-proposes-options exists in the shape every eval case has, its
# scaffold builds a repository whose run halted on a blocked task, and the evals
# index lists it. What the skills say about real options, recommending and
# never padding is the review's to rule on; this gate checks they name the means.
set -u
fail() { echo "GATE FAIL: $*" >&2; exit 1; }
for v in $(env | sed -n 's/^\(VLOOP_[A-Z0-9_]*\)=.*/\1/p'); do unset "$v"; done
unset NO_COLOR GIT_DIR GIT_WORK_TREE GIT_INDEX_FILE CLAUDE_PLUGIN_ROOT
export LC_ALL=C TZ=UTC GIT_CONFIG_GLOBAL=/dev/null GIT_CONFIG_NOSYSTEM=1 GIT_TERMINAL_PROMPT=0
export GIT_AUTHOR_NAME=gate GIT_AUTHOR_EMAIL=gate@example.invalid GIT_COMMITTER_NAME=gate GIT_COMMITTER_EMAIL=gate@example.invalid
ROOT=$(pwd -P)
t=$(mktemp -d "${TMPDIR:-/tmp}/gate.XXXXXX") && t=$(cd "$t" && pwd -P) || fail mktemp
trap 'rm -rf "$t"' EXIT
mkdir -p "$t/bin" || fail mkdir
go build -o "$t/bin/vloop" ./cmd/vloop || fail "go build ./cmd/vloop"
GOOS=linux go build -o "$t/vloop-linux" ./cmd/vloop || fail "GOOS=linux go build ./cmd/vloop"
GOOS=windows go build -o "$t/vloop.exe" ./cmd/vloop || fail "GOOS=windows go build ./cmd/vloop"
B="$t/bin/vloop"

# --- the shipped operate skill, as the binary extracts it ---
R="$t/p"
mkdir -p "$R" && git -C "$R" init -q -b main && git -C "$R" commit -q --allow-empty -m "gate base" || fail "git init"
(cd "$R" && "$B" init </dev/null) >"$t/init.out" 2>&1 || fail "vloop init: $(cat "$t/init.out")"
P=$(cd "$R" && "$B" plugin path 2>"$t/err") || fail "vloop plugin path failed: $(cat "$t/err")"
case "$P" in /*) ;; *) P="$R/$P" ;; esac
S="$P/skills/operate/SKILL.md"
[ -f "$S" ] || fail "the extracted plugin has no skills/operate/SKILL.md"
for w in 'vloop intervention add' '--option' '--recommended' '--why' '--context' '--decided-option' '--decided-other' '--decided'; do
  grep -Fq -- "$w" "$S" || fail "the shipped operate skill does not name $w"
done
grep -qi 'three' "$S" || fail "the shipped operate skill does not bound the options at three"

# --- this repository's operator skill, step 6 ---
O="$ROOT/.claude/skills/vloop-operator/SKILL.md"
for w in 'vloop intervention add' '--option' '--recommended' '--why' '--context' 'vloop intervention migrate'; do
  grep -Fq -- "$w" "$O" || fail ".claude/skills/vloop-operator/SKILL.md does not name $w"
done

# --- the eval case ---
C="$ROOT/plugin/evals/operate-proposes-options"
[ -d "$C" ] || fail "no eval case plugin/evals/operate-proposes-options"
grep -Eq '^name: *"?operate-proposes-options"?$' "$C/case.yaml" || fail "case.yaml does not name the case operate-proposes-options"
grep -Eq 'scaffold_script: *"?scaffold.sh"?' "$C/case.yaml" || fail "case.yaml does not declare scaffold.sh"
[ "$(head -n 1 "$C/prompt.md")" = "---" ] || fail "prompt.md has no frontmatter"
awk 'NR > 1 && /^---$/ { exit } /^max_turns:/ { m = 1 } /^allowed_tools:/ { a = 1 } END { exit !(m && a) }' "$C/prompt.md" || fail "prompt.md's frontmatter lacks max_turns or allowed_tools"
grep -q '^/vloop:operate' "$C/prompt.md" || fail "prompt.md does not invoke /vloop:operate"
n=0
for g in "$C"/graders/*.md; do
  [ -f "$g" ] || continue
  awk 'NR == 1 && $0 != "---" { exit 1 } NR > 1 && /^---$/ { exit } /^type:/ { ty = 1 } /^weight:/ { w = 1 } END { exit !(ty && w) }' "$g" || fail "grader $(basename "$g") lacks a frontmatter type or weight"
  n=$((n + 1))
done
[ "$n" -ge 3 ] || fail "the case has $n graders; it is graded on options with a recommendation and reason, on not amending the gate before the operator answers, and on no straw option"
mkdir -p "$t/s" || fail mkdir
(cd "$t/s" && PATH="$t/bin:$PATH" bash "$C/scaffold.sh") >"$t/scaffold.out" 2>&1 || fail "the case's scaffold.sh failed in an empty directory: $(tail -5 "$t/scaffold.out")"
[ -f "$t/s/.vloop/state/state.json" ] || fail "the scaffold leaves no .vloop/state/state.json"
jq -e 'any(.tasks[]; .status == "blocked")' "$t/s/.vloop/state/state.json" >/dev/null 2>&1 || fail "the scaffold's run is not halted: no task is blocked"
grep -Fq 'operate-proposes-options' "$ROOT/plugin/evals/README.md" || fail "plugin/evals/README.md does not list operate-proposes-options"
echo "gate T8: ok"
