#!/usr/bin/env bash
#
# The interventions index: one table row per record in .vloop/interventions/.
#
#   tools/interventions-index.sh           print the table
#   tools/interventions-index.sh --write   rewrite it between the index markers
#   tools/interventions-index.sh --check   exit 1 if the committed index is stale
#
# Rows sort by phase in lifecycle order, then kind, then id — which is also
# chronological, since the id starts with I<YYYYMMDD-HHMM>. A record whose
# phase, kind, automatable, by or agreement is outside the vocabulary fails the
# run (a v1 record has no agreement and reads as no-options): an index
# that silently sorts a typo to the top is worse than none.
#
# Plain bash and POSIX awk, so it runs on macOS's BSD tools and on Linux.

set -euo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/.."

DIR=".vloop/interventions"
INDEX="$DIR/README-interventions.md"
BEGIN_MARK="<!-- index:begin -->"
END_MARK="<!-- index:end -->"

table() {
  printf '| Phase | Kind | Automatable | Agreement | Id |\n| --- | --- | --- | --- | --- |\n'
  local f
  for f in "$DIR"/I*.md; do
    [[ -e "$f" ]] || continue
    awk -v file="${f##*/}" '
      NR == 1 && $0 == "---" { fm = 1; next }
      fm && $0 == "---"      { exit }
      fm {
        k = $0; sub(/:.*/, "", k)
        v = $0; sub(/^[^:]*:[ \t]*/, "", v)
        a[k] = v
      }
      END {
        n = split("setup design run halt verify close next", ph, " ")
        for (i = 1; i <= n; i++) order[ph[i]] = i
        split("direction decision context-supply halt verification-finding repair carry-forward ceremony", kk, " ")
        for (i in kk) kinds[kk[i]] = 1
        split("yes partly no", aa, " ")
        for (i in aa) autos[aa[i]] = 1
        split("operator assistant both", bb, " ")
        for (i in bb) bys[bb[i]] = 1
        split("recommended other-option adjusted different no-options", ag, " ")
        for (i in ag) agrees[ag[i]] = 1
        if (!("agreement" in a)) a["agreement"] = "no-options"
        bad = ""
        if (!(a["phase"] in order))      bad = bad " phase \"" a["phase"] "\""
        if (!(a["kind"] in kinds))       bad = bad " kind \"" a["kind"] "\""
        if (!(a["automatable"] in autos)) bad = bad " automatable \"" a["automatable"] "\""
        if (!(a["by"] in bys))           bad = bad " by \"" a["by"] "\""
        if (!(a["agreement"] in agrees)) bad = bad " agreement \"" a["agreement"] "\""
        if (a["id"] "" == "")            bad = bad " no id"
        if (bad != "") { print file ":" bad > "/dev/stderr"; exit 2 }
        printf "%d\t%s\t%s\t%s\t%s\t%s\t%s\n", order[a["phase"]], a["phase"], a["kind"], a["automatable"], a["agreement"], a["id"], file
      }' "$f"
  done | sort -t "$(printf '\t')" -k1,1n -k3,3 -k6,6 |
    awk -F '\t' '{ printf "| %s | %s | %s | %s | [%s](%s) |\n", $2, $3, $4, $5, $6, $7 }'
}

# The index file with its marked block replaced by a fresh table.
rendered() {
  local t
  t="$(mktemp "${TMPDIR:-/tmp}/interventions-index.XXXXXX")"
  table >"$t"
  awk -v b="$BEGIN_MARK" -v e="$END_MARK" -v t="$t" '
    $0 == b { print; while ((getline l < t) > 0) print l; skip = 1; next }
    $0 == e { skip = 0 }
    !skip   { print }' "$INDEX"
  rm -f "$t"
}

case "${1:-}" in
  "")      table ;;
  --write)
    [[ -f "$INDEX" ]] || { echo "no index: $INDEX" >&2; exit 1; }
    out="$(mktemp "${TMPDIR:-/tmp}/interventions-index.XXXXXX")"
    rendered >"$out" && mv "$out" "$INDEX"
    echo "wrote $INDEX" ;;
  --check)
    [[ -f "$INDEX" ]] || { echo "no index: $INDEX" >&2; exit 1; }
    if rendered | cmp -s - "$INDEX"; then echo "index up to date"
    else echo "index is stale — run tools/interventions-index.sh --write" >&2; exit 1; fi ;;
  *) echo "usage: tools/interventions-index.sh [--write | --check]" >&2; exit 2 ;;
esac
