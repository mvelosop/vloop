#!/usr/bin/env bash
# An index row and the file it points at must not drift apart.
#
# `.loop/todo/` grew both halves of this within a day of being created. One
# entry had every item discharged and still read `status: open`, in a directory
# whose index is titled "Open work". And a row summarising an entry's
# `waiting-on` as "Nothing — briefed" pointed at a file whose own frontmatter
# said "interim fix ready to plan; residue open".
#
# The second one recurred while the entry describing it was being written:
# widening it changed its `waiting-on`, the row kept the previous wording, and
# both versions read perfectly well. The drift was a clause. That is why the
# index quotes verbatim instead of summarising, and why every case below plants
# a tree that a reader would call fine.
. "$(dirname "$0")/../lib.sh"

CHECK="$(cd "$(dirname "$0")/.." && pwd)/check-todo.sh"
TREE=""
cleanup_tree() { [[ -n "$TREE" && "$TREE" == */looptodo.* ]] && rm -rf "$TREE"; }
trap 'cleanup_tree; fixture_cleanup' EXIT

# Two entries, one open and one closed, correctly indexed. Every case below is
# this tree with exactly one thing moved.
plant() {                       # plant [<waiting-on as the ROW states it>]
  cleanup_tree
  TREE="$(mktemp -d "${TMPDIR:-/tmp}/looptodo.XXXXXX")"
  mkdir -p "$TREE/.loop/todo"
  local row="${1:-A decision on the shape.}"
  cat >"$TREE/.loop/todo/TD20260101-0900-open-one.todo.md" <<'MD'
---
name: TD20260101-0900-open-one
description: an entry that is still open
status: open
created: 2026-01-01
source: a scenario
waiting-on: A decision on the shape.
---
# Open one
MD
  cat >"$TREE/.loop/todo/TD20260101-1000-closed-one.todo.md" <<'MD'
---
name: TD20260101-1000-closed-one
description: an entry that has been discharged
status: closed
created: 2026-01-01
closed: 2026-01-02
closed-by: a brief
source: a scenario
---
# Closed one
MD
  cat >"$TREE/.loop/todo/TODO.md" <<MD
# Open work — index

## Open

| Entry | Source | Waiting on |
| --- | --- | --- |
| [TD20260101-0900-open-one](TD20260101-0900-open-one.todo.md) | a scenario | $row |

## Closed

| Entry | Closed | Discharged by |
| --- | --- | --- |
| [TD20260101-1000-closed-one](TD20260101-1000-closed-one.todo.md) | 2026-01-02 | a brief |

## Status

Prose after the tables, which the section reader must not swallow.
MD
}

run() { bash "$CHECK" "$TREE" >"$TREE/out.log" 2>&1; }
show() { sed 's/^/      /' "$TREE/out.log"; }

note "── a consistent directory passes ──"
plant
if run; then ok "consistent index passes"; else bad "a correct tree failed"; show; fi

note "── the drift that started this: the row paraphrases ──"
# Reads fine. Says the same thing. Is not the same string.
plant "A decision on its shape."
if run; then bad "a paraphrased row passed — this is the exact drift the rule exists for"; show
else ok "paraphrased row fails"
  grep -q 'verbatim' "$TREE/out.log" && ok "log: names the verbatim rule" || bad "log does not mention verbatim"
  grep -q 'TD20260101-0900-open-one' "$TREE/out.log" && ok "log: names the entry" || bad "log does not name the entry"
fi

note "── an entry discharged but still marked open ──"
plant
sed -i.bak 's/^status: closed/status: open/' "$TREE/.loop/todo/TD20260101-1000-closed-one.todo.md"
if run; then bad "an entry whose status contradicts its section passed"; show
else ok "status/section mismatch fails"
  grep -q 'not listed under' "$TREE/out.log" && ok "log: names the missing section" \
    || bad "log does not say which section it is missing from"
fi

note "── a closed entry that still says what it awaits ──"
# Written out in full rather than appended: the first version of this case
# appended the line after the body, where the frontmatter parser never looks,
# so it asserted nothing and said "ok". A fixture can be broken too.
plant
cat >"$TREE/.loop/todo/TD20260101-1000-closed-one.todo.md" <<'MD'
---
name: TD20260101-1000-closed-one
description: an entry that has been discharged
status: closed
created: 2026-01-01
closed: 2026-01-02
closed-by: a brief
waiting-on: something
source: a scenario
---
# Closed one
MD
if run; then bad "a closed entry carrying waiting-on passed"; show
else ok "closed-with-waiting-on fails"
  grep -q 'what discharged it' "$TREE/out.log" && ok "log: says what a closed entry carries instead" \
    || bad "log does not explain the rule"
fi

note "── a closed entry with no record of what discharged it ──"
plant
sed -i.bak '/^closed-by:/d' "$TREE/.loop/todo/TD20260101-1000-closed-one.todo.md"
if run; then bad "a closed entry with no closed-by passed"; show
else ok "closed without closed-by fails"; fi

note "── a file nobody indexed, and a row pointing at nothing ──"
plant
cp "$TREE/.loop/todo/TD20260101-0900-open-one.todo.md" \
   "$TREE/.loop/todo/TD20260101-1100-orphan.todo.md"
if run; then bad "an unindexed entry passed — invisible to anyone reading the index"; show
else ok "unindexed entry fails"; fi

plant
rm "$TREE/.loop/todo/TD20260101-1000-closed-one.todo.md"
if run; then bad "a row pointing at a missing file passed"; show
else ok "dangling row fails"
  grep -q 'does not exist' "$TREE/out.log" && ok "log: names the dangling row" || bad "log does not name it"
fi

note "── absence is not failure, and an installed copy is not ours to police ──"
plant
rm -rf "$TREE/.loop/todo"
if run; then ok "no .loop/todo/ skips cleanly"; else bad "a tree with no todo directory failed"; show; fi

plant
sed -i.bak 's/^status: closed/status: open/' "$TREE/.loop/todo/TD20260101-1000-closed-one.todo.md"
: >"$TREE/.loop/.installed"
if run; then ok "an installed copy is skipped despite a real defect"
else bad "an installed copy was policed"; show; fi

finish
