#!/bin/sh
# Gate T1 — R2: the brief-review/v1 report schema and the session/v5 record,
# embedded and served by `vloop schema` like the others. Built from this tree;
# every fixture is this folder's: valid/ must validate (exit 0), invalid/ must
# be refused (exit 1, not 2: the schema is known). br-* fixtures are checked
# against brief-review/v1, session-v5-* against session/v5. session/v4 is a
# published schema and must not change: its embedded document equals the copy
# taken when this plan was written, and it still refuses a review-brief record.
set -u
fail() { echo "GATE FAIL: $*" >&2; exit 1; }
for v in $(env | sed -n 's/^\(VLOOP_[A-Z0-9_]*\)=.*/\1/p'); do unset "$v"; done
export LC_ALL=C GOTOOLCHAIN=local
G=$(cd "$(dirname "$0")" && pwd -P) || fail "no gate folder"
JQ=$(command -v jq) || fail "jq is not on PATH"
t=$(mktemp -d "${TMPDIR:-/tmp}/gate.XXXXXX") && t=$(cd "$t" && pwd -P) || fail mktemp
trap 'rm -rf "$t"' EXIT
go build -buildvcs=false -o "$t/vloop" ./cmd/vloop || fail "go build of this tree"
S=$t/vloop
mkdir -p "$t/work" || fail mkdir
cd "$t/work" || fail cd

"$S" schema list >"$t/list" 2>&1 || fail "vloop schema list failed: $(cat "$t/list")"
for n in brief-review/v1 session/v5 session/v4 session/v3 gate-verdict/v2 verify/v1; do
  grep -qxF "$n" "$t/list" || fail "vloop schema list does not list $n: $(tr '\n' ' ' <"$t/list")"
done
"$S" schema show brief-review/v1 >"$t/br.json" 2>&1 || fail "vloop schema show brief-review/v1 failed: $(cat "$t/br.json")"
"$JQ" -e 'type == "object"' "$t/br.json" >/dev/null 2>&1 || fail "vloop schema show brief-review/v1 is not a JSON document"

# session/v4 is unchanged.
"$S" schema show session/v4 >"$t/s4.json" 2>&1 || fail "vloop schema show session/v4 failed"
[ "$("$JQ" -S . "$t/s4.json")" = "$("$JQ" -S . "$G/session.v4.json")" ] || fail "session/v4 changed: a published schema does not change"
"$S" schema validate session/v4 "$G/s4-review-brief.json" >"$t/out" 2>&1; rc=$?
[ "$rc" = 1 ] || fail "session/v4 accepted (exit $rc) a record with phase review-brief: $(cat "$t/out")"

sch() { case "$(basename "$1")" in br-*) echo brief-review/v1;; session-v5-*) echo session/v5;; *) fail "fixture $1 has no schema";; esac; }
for f in "$G"/valid/*.json; do
  n=$(sch "$f")
  "$S" schema validate "$n" "$f" >"$t/out" 2>&1; rc=$?
  [ "$rc" = 0 ] || fail "valid/$(basename "$f") must validate against $n, got exit $rc: $(cat "$t/out")"
done
for f in "$G"/invalid/*.json; do
  n=$(sch "$f")
  "$S" schema validate "$n" "$f" >"$t/out" 2>&1; rc=$?
  [ "$rc" = 1 ] || fail "invalid/$(basename "$f") must be refused by $n with exit 1, got exit $rc: $(cat "$t/out")"
done
echo "gate T1: ok"
