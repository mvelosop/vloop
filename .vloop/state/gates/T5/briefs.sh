# Brief helpers for the brief check gates.
GF=.vloop/repo-guidelines.md
n=0
# wbrief <what it is> <contract> <constraints> [<extra binding reference line>]:
# a ready brief in $R that passes every other rule; its path is in $F.
wbrief() {
  n=$((n + 1)); nm=B20260101-$((1000 + n))-case.loop-brief
  F=docs/briefs/$nm.md
  mkdir -p "$R/docs/briefs" || fail mkdir
  cat >"$R/$F" <<B
---
name: $nm
description: A demo
kind: brief
status: ready
created: 2026-01-01
seeds: A demo
depends-on: []
---
# Brief

## What it is

$1

## Why this shape, and what was rejected

Because.

## Binding references

- \`docs/x.md\` — the error contract
${4:-}

## Behaviour contract

An unknown id exits 1.

$2

## Worked example

\`\`\`
\$ demo
ok
\`\`\`

## Out of scope

- one
- two

## Constraints

- Go.
$3

## Shape

2 to 3 tasks.
B
}
# check <out> <brief path>: vloop brief check in $R; output, each line with its
# indentation removed, in $t/<out>; exit in $got.
check() { inr "$B" brief check "$2" </dev/null >"$t/$1.raw" 2>&1; got=$?; sed 's/^ *//' "$t/$1.raw" >"$t/$1"; }
has() { grep -Fxq -- "$2" "$t/$1" || fail "$1: no line '$2' in:
$(cat "$t/$1.raw")"; }
# none <out> <prefix>: no line of <out> starts with "! <prefix>".
none() { grep -Fq -- "! $2" "$t/$1" && fail "$1: an unexpected '$2' warning:
$(cat "$t/$1.raw")"; return 0; }
# basebin: $t/vloop-base, the binary of the plan's base, for "as today".
BASE=dc2642803289b6626124451e2567f80f3795bfc7
basebin() {
  mkdir -p "$t/base" && git -C "$top" archive "$BASE" | tar -x -C "$t/base" || fail "git archive $BASE"
  (cd "$t/base" && go build -buildvcs=false -o "$t/vloop-base" ./cmd/vloop) || fail "build of the base"
}
