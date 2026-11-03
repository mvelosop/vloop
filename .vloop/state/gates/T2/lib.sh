# Shared harness for B35's gates (the same file in T1 and T2). The sourcing
# gate sets G, its own folder. The binary and the fence's root live under $t,
# a temporary directory removed on exit; nothing touches this repository's
# tree or refs, and nothing runs a command it judges: `vloop fence check`
# only prints allow or deny.
set -u
fail() { echo "GATE FAIL: $*" >&2; exit 1; }
for v in $(env | sed -n 's/^\(VLOOP_[A-Z0-9_]*\)=.*/\1/p'); do unset "$v"; done
unset GOFLAGS GOBIN
export LC_ALL=C
top=$(pwd -P)
[ -f "$top/go.mod" ] && [ -d "$top/cmd/vloop" ] || fail "run from the repository root"
t=$(mktemp -d "${TMPDIR:-/tmp}/gate.XXXXXX") && t=$(cd "$t" && pwd -P) || fail mktemp
trap 'rm -rf "$t"' EXIT

# The fences as they stood when this plan was written: B35 changes no rule.
for f in "$G"/fence/*.json; do
  cmp -s "$f" "$top/fence/$(basename "$f")" || fail "fence/$(basename "$f") differs from the fence this plan was written against: B35 changes no fence's rules"
done

mkdir -p "$t/bin" "$t/root" || fail mkdir
go build -buildvcs=false -o "$t/bin/vloop" ./cmd/vloop >"$t/build.out" 2>&1 || fail "go build ./cmd/vloop: $(tail -n 5 "$t/build.out")"
S=$t/bin/vloop

# check FILE: every line of FILE that is not blank or a comment is
# PHASE<TAB>COMMAND<TAB>WANT, WANT being `allow` (exit 0, stdout `allow`) or
# `deny: <reason>` (exit 1, stdout exactly that line). COMMAND is passed to
# `vloop fence check --tool bash` as one argument, byte for byte. Every
# mismatch is reported; the gate fails if any line does.
check() {
  n=0 bad=0
  tab=$(printf '\t')
  while IFS="$tab" read -r phase cmd want; do
    case "$phase" in '' | '#'*) continue ;; esac
    n=$((n + 1))
    got=$("$S" fence check --root "$t/root" --phase "$phase" --tool bash "$cmd" 2>"$t/err")
    rc=$?
    wantrc=1
    [ "$want" = allow ] && wantrc=0
    if [ "$got" != "$want" ] || [ "$rc" != "$wantrc" ]; then
      bad=$((bad + 1))
      printf 'GATE MISMATCH [%s] %s\n  want (exit %s): %s\n  got  (exit %s): %s %s\n' "$phase" "$cmd" "$wantrc" "$want" "$rc" "$got" "$(cat "$t/err")" >&2
    fi
  done <"$1"
  [ "$n" -gt 0 ] || fail "no cases read from $1"
  [ "$bad" = 0 ] || fail "$bad of $n cases in $1 answered otherwise"
}

# hook PHASE COMMAND WANT: the Copilot hook's answer for one bash call, WANT
# being the JSON line it must print (and exit 0).
hook() {
  payload=$(printf '%s' "$2" | sed 's/\\/\\\\/g; s/"/\\"/g')
  got=$(printf '{"toolName":"bash","toolArgs":{"command":"%s"}}' "$payload" | "$S" fence hook --phase "$1" --root "$t/root" 2>"$t/err")
  rc=$?
  [ "$rc" = 0 ] && [ "$got" = "$3" ] || fail "the hook, $1 phase, for '$2': want $3 (exit 0), got (exit $rc) $got $(cat "$t/err")"
}
