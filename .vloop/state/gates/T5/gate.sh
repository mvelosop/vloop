#!/bin/sh
# Gate T5 — the close: every gate of the plan at once on the final tree (T1
# the schema, T2 the doctor and the preflight with every V1 line and the
# planted running plan, T3 the plan phase's guard with every V2 line and its
# planted cases, T4 the docs), and no fence changed: fence/*.json byte for
# byte as at the base (fence/ in this folder), with no fence file added or
# removed.
set -u
fail() { echo "GATE FAIL: $*" >&2; exit 1; }
G=$(cd "$(dirname "$0")" && pwd -P) || fail "no gate folder"
top=$(pwd -P)
GS=$(dirname "$G")
[ "$(cd "$G/fence" && ls)" = "$(cd "$top/fence" && ls)" ] || fail "a fence file was added or removed under fence/"
for f in "$G"/fence/*.json; do
  cmp -s "$f" "$top/fence/$(basename "$f")" || fail "fence/$(basename "$f") changed: no task changes a fence"
done
for id in T1 T2 T3 T4; do
  sh "$GS/$id/gate.sh" || fail "gate $id fails on the final tree"
done
echo "gate T5: ok"
