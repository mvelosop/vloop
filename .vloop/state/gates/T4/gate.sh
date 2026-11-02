#!/bin/sh
# Gate T4 — the docs. Each file is read as paragraphs (blank-line separated, a
# `#` heading its own paragraph), each joined into one line with runs of
# whitespace collapsed to one space, so the wrapping is free and the words
# are not. The literal texts are the files in this folder:
#   docs/guide/configuration.md: conf.txt in the paragraph that held
#     conf-old.txt, and conf-old.txt gone.
#   docs/domain/execution/plan.md: plan.txt is the paragraph right after the
#     first one under `# Plan` (the one ending "Part of [execution](...)").
#   docs/domain/execution/session.md: guards.txt right after the sentence
#     "what a session changed among them is restored (…)." in the guards
#     paragraph; reviewbrief.txt right after "with no task." in the paragraph
#     beginning "A **review-brief** session is started by".
#   docs/design-notes/fence-audit.md: note.txt is the first paragraph under
#     the gate-review audit's heading and under the plan audit's heading
#     (italic allowed), and the file without those two notes is the base's
#     (fence-audit.base.md in this folder), whitespace collapsed.
set -u
fail() { echo "GATE FAIL: $*" >&2; exit 1; }
export LC_ALL=C
G=$(cd "$(dirname "$0")" && pwd -P) || fail "no gate folder"
top=$(pwd -P)
t=$(mktemp -d "${TMPDIR:-/tmp}/gate.XXXXXX") || fail mktemp
trap 'rm -rf "$t"' EXIT
# paras <file>: one collapsed paragraph per line.
paras() {
  awk '
    function out() { if (p != "") { gsub(/[ \t]+/, " ", p); sub(/^ /, "", p); sub(/ $/, "", p); print p }; p = "" }
    /^[ \t]*$/ { out(); next }
    /^#/ { out(); p = $0; out(); next }
    { p = (p == "" ? $0 : p " " $0) }
    END { out() }
  ' "$1"
}
txt() { tr -s ' \n\t' '   ' <"$G/$1" | sed 's/ $//'; }
CONF=$(txt conf.txt); OLD=$(txt conf-old.txt); PLAN=$(txt plan.txt)
GUARDS=$(txt guards.txt); RB=$(txt reviewbrief.txt); NOTE=$(txt note.txt)

# configuration.md
F=$top/docs/guide/configuration.md
paras "$F" >"$t/conf"
grep -qF -- "$OLD" "$t/conf" && fail "docs/guide/configuration.md still says: $OLD"
grep -F -- '**In a plan session the hook judges against the plan fence**' "$t/conf" | grep -qF -- "$CONF" ||
  fail "docs/guide/configuration.md's plan fence bullet does not say: $CONF"

# plan.md
F=$top/docs/domain/execution/plan.md
paras "$F" >"$t/plan"
p=$(awk 'f == 1 { print; exit } /^A brief decomposed into tasks/ { f = 1 }' "$t/plan")
[ "$p" = "$PLAN" ] || fail "docs/domain/execution/plan.md: the paragraph after the first one is not the brief's:
$p
--- want
$PLAN"

# session.md
F=$top/docs/domain/execution/session.md
paras "$F" >"$t/session"
grep -F -- 'The driver checks anyway' "$t/session" | grep -qF -- "what a session changed among them is restored (a review session that changes the work it judges has it reverted and fails). $GUARDS" ||
  fail "docs/domain/execution/session.md's guards paragraph does not say, right after the restored sentence: $GUARDS"
grep -F -- 'A **review-brief** session is started by' "$t/session" | grep -qF -- "with no task. $RB" ||
  fail "docs/domain/execution/session.md's review-brief paragraph does not say, right after 'with no task.': $RB"

# fence-audit.md
F=$top/docs/design-notes/fence-audit.md
paras "$F" >"$t/audit"
for h in '# Fence audit: who enforces each rule in a Copilot gate-review session' '# Fence audit: who enforces each rule in a Copilot plan session'; do
  grep -qxF -- "$h" "$t/audit" || fail "docs/design-notes/fence-audit.md lost its heading: $h"
  p=$(awk -v h="$h" 'f == 1 { print; exit } $0 == h { f = 1 }' "$t/audit" | sed -e 's/^[*_]//' -e 's/[*_]$//')
  [ "$p" = "$NOTE" ] || fail "docs/design-notes/fence-audit.md: the first paragraph under '$h' is not the note:
$p
--- want
$NOTE"
done
[ "$(grep -cF -- "$NOTE" "$t/audit")" = 2 ] || fail "docs/design-notes/fence-audit.md: want the note exactly twice"
grep -vF -- "$NOTE" "$t/audit" >"$t/audit.rest"
paras "$G/fence-audit.base.md" >"$t/audit.base"
cmp -s "$t/audit.base" "$t/audit.rest" || fail "docs/design-notes/fence-audit.md changed outside its two notes: the audit stays as recorded
$(diff "$t/audit.base" "$t/audit.rest" | head -10)"
echo "gate T4: ok"
