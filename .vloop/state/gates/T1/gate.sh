#!/bin/sh
# Gate T1 — V1, the schema: verify/v2, embedded and listed. This tree's binary,
# built into a temporary directory, lists verify/v1 and verify/v2; validates
# each report in valid/ as verify/v2 (exit 0) — the worked example's stub
# reports — and refuses each one in invalid/ (exit 1, not the unknown-schema
# exit 2) — the planted reports (a verify/v1 report, no complete, a not_run
# entry without reason) and the rest of the shape: not_run required, an array
# of {line, reason} objects with both non-empty strings, complete a boolean,
# and verify/v1's properties and constraints carried over. schemas/verify.v1.json
# stays byte for byte as at the base (verify.v1.json in this folder).
set -u
fail() { echo "GATE FAIL: $*" >&2; exit 1; }
for v in $(env | sed -n 's/^\(VLOOP_[A-Z0-9_]*\)=.*/\1/p'); do unset "$v"; done
export LC_ALL=C GOTOOLCHAIN=local
G=$(cd "$(dirname "$0")" && pwd -P) || fail "no gate folder"
top=$(pwd -P)
cmp -s "$G/verify.v1.json" "$top/schemas/verify.v1.json" || fail "schemas/verify.v1.json changed: a published schema does not change"
t=$(mktemp -d "${TMPDIR:-/tmp}/gate.XXXXXX") && t=$(cd "$t" && pwd -P) || fail mktemp
trap 'rm -rf "$t"' EXIT
go build -buildvcs=false -o "$t/vloop" ./cmd/vloop || fail "go build of this tree"
mkdir -p "$t/r" && cd "$t/r" || fail mkdir
V() { env -i PATH=/usr/bin:/bin HOME="$t" LC_ALL=C "$t/vloop" "$@"; }
V schema list >"$t/list" 2>&1 || fail "vloop schema list failed: $(cat "$t/list")"
grep -qxF 'verify/v1' "$t/list" || fail "vloop schema list does not name verify/v1: $(tr '\n' ' ' <"$t/list")"
grep -qxF 'verify/v2' "$t/list" || fail "vloop schema list does not name verify/v2: $(tr '\n' ' ' <"$t/list")"
n=0
for f in "$G"/valid/*.json; do
  V schema validate verify/v2 "$f" >"$t/o" 2>&1; c=$?
  [ "$c" = 0 ] || fail "vloop schema validate verify/v2 valid/$(basename "$f") exited $c, want 0: $(cat "$t/o")"
  n=$((n + 1))
done
[ "$n" -ge 6 ] || fail "only $n valid fixtures found"
n=0
for f in "$G"/invalid/*.json; do
  V schema validate verify/v2 "$f" >"$t/o" 2>&1; c=$?
  [ "$c" = 1 ] || fail "vloop schema validate verify/v2 invalid/$(basename "$f") exited $c, want 1: $(cat "$t/o")"
  n=$((n + 1))
done
[ "$n" -ge 15 ] || fail "only $n invalid fixtures found"
# verify/v1 still accepts its own report and refuses a verify/v2 one.
V schema validate verify/v1 "$G/invalid/v1-report.json" >"$t/o" 2>&1 || fail "verify/v1 no longer accepts a verify/v1 report: $(cat "$t/o")"
V schema validate verify/v1 "$G/valid/v2-empty.json" >"$t/o" 2>&1 && fail "verify/v1 accepts a verify/v2 report"
echo "gate T1: ok"
