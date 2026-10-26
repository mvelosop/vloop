#!/bin/sh
# Gate T2 — R2: a brief named any way. In a local clone of this repository
# (its committed history and records; nothing is written to this tree), for
# B8, N = B20261002-2135-vloop-v1-security, given as the bare name N, the
# .loop-brief name and the path docs/briefs/N.loop-brief.md:
#   - `intervention add --brief`, `defect add --brief` and `intervention set
#     <id> brief` write brief: N.loop-brief whatever form they were given;
#   - an intervention and a defect planted with a bare `brief: N` (as records
#     written before this brief were) count;
#   - `vloop metrics <form> --json` gives equal "defects" and "interventions"
#     in all three forms, each grown by exactly the records added here over the
#     .loop-brief form's counts before them (threaded, never hardcoded);
#   - `intervention list --brief` and `defect list --brief` list the same
#     records in all three forms, the planted bare ones among them;
#   - `defect add` still refuses a brief that is not a loop brief, and one that
#     does not exist, writing nothing.
set -u
fail() { echo "GATE FAIL: $*" >&2; exit 1; }
for v in $(env | sed -n 's/^\(VLOOP_[A-Z0-9_]*\)=.*/\1/p'); do unset "$v"; done
unset GIT_DIR GIT_WORK_TREE GIT_INDEX_FILE GOFLAGS GOBIN
export LC_ALL=C GIT_CONFIG_GLOBAL=/dev/null GIT_CONFIG_NOSYSTEM=1 GIT_TERMINAL_PROMPT=0 GOTOOLCHAIN=local GOPROXY=off
JQ=$(command -v jq) || fail "jq is not on PATH"
t=$(mktemp -d "${TMPDIR:-/tmp}/gate.XXXXXX") && t=$(cd "$t" && pwd -P) || fail mktemp
trap 'chmod -R u+w "$t" 2>/dev/null; rm -rf "$t"' EXIT
V="$t/bin/vloop"
mkdir -p "$t/bin" && go build -buildvcs=false -o "$V" ./cmd/vloop || fail "go build of this tree"
git clone -q . "$t/c" || fail "git clone of this repository"
cd "$t/c" || fail cd
git config user.name Gate && git config user.email gate@example.invalid || fail "git config"
export HOME="$t/home"
mkdir -p "$HOME"
N=B20261002-2135-vloop-v1-security
FORMS="$N $N.loop-brief docs/briefs/$N.loop-brief.md"
[ -f "docs/briefs/$N.loop-brief.md" ] || fail "fixture: no docs/briefs/$N.loop-brief.md"
want="$N.loop-brief"
# briefof <record>: the brief: field of a record's frontmatter, unquoted.
briefof() { sed -n 's/^brief:[[:space:]]*//p' "$1" | head -1 | sed 's/^"\(.*\)"$/\1/'; }
metrics() { "$V" metrics "$1" --json >"$t/m.json" 2>"$t/m.err" || fail "vloop metrics $1 exits non-zero: $(cat "$t/m.err")"; }
metrics "$want"
d0=$("$JQ" '.defects.total' "$t/m.json") i0=$("$JQ" '.interventions.total' "$t/m.json")
case "$d0$i0" in *null*|"") fail "vloop metrics $want --json has no defects.total or interventions.total";; esac

# add, in every form, writes the .loop-brief name.
n=0
for f in $FORMS; do
  n=$((n + 1))
  p=$("$V" intervention add "gate r2 intervention $n" --brief "$f" --kind repair --phase halt --by operator --automatable no 2>"$t/e") ||
    fail "intervention add --brief $f exits non-zero: $(cat "$t/e")"
  [ -f "$p" ] || fail "intervention add --brief $f printed '$p', not a file"
  [ "$(briefof "$p")" = "$want" ] || fail "intervention add --brief $f wrote brief: $(briefof "$p"), want $want"
  last=$p
  p=$("$V" defect add "gate r2 defect $n" --brief "$f" --found-by operator 2>"$t/e") ||
    fail "defect add --brief $f exits non-zero: $(cat "$t/e")"
  [ -f "$p" ] || fail "defect add --brief $f printed '$p', not a file"
  [ "$(briefof "$p")" = "$want" ] || fail "defect add --brief $f wrote brief: $(briefof "$p"), want $want"
done
# set brief, given the bare name, writes the .loop-brief name.
id=$(basename "$last" .md)
"$V" intervention set "$id" brief "$N" >"$t/o" 2>&1 || fail "intervention set $id brief $N exits non-zero: $(cat "$t/o")"
[ "$(briefof "$last")" = "$want" ] || fail "intervention set $id brief $N wrote brief: $(briefof "$last"), want $want"

# Records already written with the bare name.
p=$("$V" intervention add "gate r2 planted bare intervention" --brief "$want" --kind repair --phase halt --by operator --automatable no 2>"$t/e") || fail "planting: $(cat "$t/e")"
sed "s/^brief:.*/brief: $N/" "$p" >"$t/x" && cat "$t/x" >"$p" || fail "planting"
pi=$(basename "$p" .md)
p=$("$V" defect add "gate r2 planted bare defect" --brief "$want" --found-by operator 2>"$t/e") || fail "planting: $(cat "$t/e")"
sed "s/^brief:.*/brief: $N/" "$p" >"$t/x" && cat "$t/x" >"$p" || fail "planting"
pd=$(basename "$p" .md)
[ "$(briefof "$p")" = "$N" ] || fail "planting: the defect's brief is not bare"

# The metrics, in every form.
dw=$((d0 + 4)) iw=$((i0 + 4))
for f in $FORMS; do
  metrics "$f"
  d=$("$JQ" '.defects.total' "$t/m.json") i=$("$JQ" '.interventions.total' "$t/m.json")
  [ "$d" = "$dw" ] || fail "vloop metrics $f: defects.total $d, want $dw ($d0 before, 4 recorded here, one with a bare brief:)"
  [ "$i" = "$iw" ] || fail "vloop metrics $f: interventions.total $i, want $iw ($i0 before, 4 recorded here, one with a bare brief:)"
done
prev=""
for f in $FORMS; do
  metrics "$f"; cur=$("$JQ" -c '{defects, interventions}' "$t/m.json")
  [ -z "$prev" ] || [ "$cur" = "$prev" ] || fail "vloop metrics $f gives other defects or interventions than the form before it: $cur vs $prev"
  prev=$cur
done

# The lists, in every form.
previ="" prevd=""
for f in $FORMS; do
  li=$("$V" intervention list --brief "$f" --json 2>"$t/e") || fail "intervention list --brief $f exits non-zero: $(cat "$t/e")"
  ld=$("$V" defect list --brief "$f" --json 2>"$t/e") || fail "defect list --brief $f exits non-zero: $(cat "$t/e")"
  ids=$(printf '%s' "$li" | "$JQ" -r '[.[].id] | sort | join(" ")') || fail "intervention list --brief $f --json is not a JSON list"
  dds=$(printf '%s' "$ld" | "$JQ" -r '[.[].id] | sort | join(" ")') || fail "defect list --brief $f --json is not a JSON list"
  case " $ids " in *" $pi "*) ;; *) fail "intervention list --brief $f omits $pi, whose brief: is the bare $N";; esac
  case " $dds " in *" $pd "*) ;; *) fail "defect list --brief $f omits $pd, whose brief: is the bare $N";; esac
  [ "$(printf '%s\n' $ids | wc -l | tr -d ' ')" = "$iw" ] || fail "intervention list --brief $f lists other than the $iw records the metrics count"
  [ -z "$previ" ] || [ "$ids" = "$previ" ] || fail "intervention list --brief $f lists other records than the form before it"
  [ -z "$prevd" ] || [ "$dds" = "$prevd" ] || fail "defect list --brief $f lists other records than the form before it"
  previ=$ids prevd=$dds
done

# Refusals, as today.
before=$(ls .vloop/defects | wc -l)
for f in docs/briefs/B20260929-1222-initial-setup-for-vloop-cli.architect-brief.md B20991231-0000-no-such-brief B20991231-0000-no-such-brief.loop-brief; do
  "$V" defect add "gate r2 refused" --brief "$f" >"$t/o" 2>&1 && fail "defect add --brief $f exits 0; a name that is not an existing loop brief is refused"
done
[ "$(ls .vloop/defects | wc -l)" = "$before" ] || fail "a refused defect add wrote a record"
echo "gate T2: ok"
