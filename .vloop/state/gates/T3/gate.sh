#!/bin/sh
# Gate T3 — C4, the docs: the configuration guide's work-fence sentence, the
# fence audit's hook bullet and its work-session paragraph, and `vloop fence
# check --help` built from this tree. Each text is read with its lines joined
# and its whitespace collapsed, so an edit wrapped in the file's style passes;
# needles.tsv holds the text that must stand (with the words around it, so it
# sits where the brief puts it) and the text that must be gone.
set -u
fail() { echo "GATE FAIL: $*" >&2; exit 1; }
G=.vloop/state/gates/T3
unset GOFLAGS GOBIN
export LC_ALL=C
[ -f go.mod ] && [ -d cmd/vloop ] || fail "run from the repository root"
t=$(mktemp -d "${TMPDIR:-/tmp}/gate.XXXXXX") || fail mktemp
trap 'rm -rf "$t"' EXIT
go build -buildvcs=false -o "$t/vloop" ./cmd/vloop >"$t/build.out" 2>&1 || fail "go build ./cmd/vloop: $(tail -n 5 "$t/build.out")"
"$t/vloop" fence check --help >"$t/help.raw" 2>&1 || fail "vloop fence check --help exited non-zero: $(cat "$t/help.raw")"
tab=$(printf '\t')
n=0 bad=0
while IFS="$tab" read -r file text; do
  case "$file" in '' | '#'*) continue ;; esac
  n=$((n + 1))
  want=present
  case "$file" in '!'*) want=absent file=${file#!} ;; esac
  src=$file
  [ "$file" = help ] && src=$t/help.raw
  [ -f "$src" ] || fail "$file is missing"
  tr -s '[:space:]' ' ' <"$src" >"$t/joined"
  if grep -qF -- "$text" "$t/joined"; then
    [ "$want" = present ] || { bad=$((bad + 1)); echo "GATE MISMATCH $file still holds: $text" >&2; }
  else
    [ "$want" = absent ] || { bad=$((bad + 1)); echo "GATE MISMATCH $file lacks: $text" >&2; }
  fi
done <"$G/needles.tsv"
[ "$n" -gt 0 ] || fail "no needles read"
[ "$bad" = 0 ] || fail "$bad of $n doc checks failed"
echo "gate T3: ok"
