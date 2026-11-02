#!/bin/sh
# Gate T4 — the close: every gate of the plan at once on the final tree (T1 the
# plan skill, T2 the gate review and its ceiling, T3 the eval case and the
# catalogues), and nothing the brief leaves unchanged has changed: schemas/*.json
# byte for byte as at the base (schemas/ in this folder), no schema added or
# removed; the other six skills' SKILL.md byte for byte as at the base (skills/
# in this folder); no mention of the URL shortener in the two skills or the two
# catalogues.
set -u
fail() { echo "GATE FAIL: $*" >&2; exit 1; }
export LC_ALL=C
G=$(cd "$(dirname "$0")" && pwd -P) || fail "no gate folder"
top=$(pwd -P)
GS=$(dirname "$G")
[ "$(cd "$G/schemas" && ls)" = "$(cd "$top/schemas" && ls)" ] || fail "a schema was added or removed under schemas/"
for f in "$G"/schemas/*.json; do
  cmp -s "$f" "$top/schemas/$(basename "$f")" || fail "schemas/$(basename "$f") changed: no schema changes"
done
for f in "$G"/skills/*/SKILL.md; do
  k=$(basename "$(dirname "$f")")
  cmp -s "$f" "$top/plugin/skills/$k/SKILL.md" || fail "plugin/skills/$k/SKILL.md changed: only the plan and gate-review skills change"
done
for f in plugin/skills/plan/SKILL.md plugin/skills/gate-review/SKILL.md plugin/evals/README.md docs/guide/evals.md; do
  grep -qi 'shortener' "$top/$f" && fail "$f mentions the URL shortener"
done
for id in T1 T2 T3; do
  sh "$GS/$id/gate.sh" || fail "gate $id fails on the final tree"
done
echo "gate T4: ok"
