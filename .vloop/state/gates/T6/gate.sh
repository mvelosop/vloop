#!/bin/sh
# Gate T6 — G1: run.git-config-ignore compares a pattern as git reports keys:
# the text before the first "." and after the last "." lower-cased, a
# subsection between kept as given. `Core.*` is refused exactly as `core.*`
# is, naming the pattern as given (exit 2); `Branch.Main.Foo-Bar` is accepted
# and ignores branch.Main.foo-bar during a run, but not branch.main.foo-bar,
# which halts the run with exit 9 as today.
G=$(cd "$(dirname "$0")" && pwd -P) || { echo "GATE FAIL: no gate folder" >&2; exit 1; }
. "$G/lib.sh"
# cset <want exit> <pattern> <want stderr or "">: vloop config set in $R.
cset() {
  inr "$B" config set run.git-config-ignore "$2" </dev/null >"$t/set.out" 2>"$t/set.err"; c=$?
  [ "$c" = "$1" ] || fail "vloop config set run.git-config-ignore '$2': exit $c, want $1: $(cat "$t/set.err")"
  if [ -n "$3" ]; then
    [ "$(cat "$t/set.err")" = "$3" ] || fail "vloop config set run.git-config-ignore '$2' printed
$(cat "$t/set.err")
want
$3"
  fi
}
MSG='— the guard must see it'
newrepo cfg
cp "$R/.vloop/config.toml" "$t/config.before"
cset 2 'Core.*' "vloop: run.git-config-ignore: Core.* would ignore core.* $MSG"
cset 2 'core.*' "vloop: run.git-config-ignore: core.* would ignore core.* $MSG"
cmp -s "$R/.vloop/config.toml" "$t/config.before" || fail "a refused pattern changed .vloop/config.toml"
cset 0 'Branch.Main.Foo-Bar' ''

# runwith <repo> <key>: a run whose work session sets <key>, under the ignore
# pattern Branch.Main.Foo-Bar; output in $t/<repo>.run, exit in $got.
runwith() {
  runrepo "$1" T1
  cset 0 'Branch.Main.Foo-Bar' ''
  commit "ignore pattern"
  printf 'git config %s x\n' "$2" >"$t/stub/work-T1.sh" || fail work.sh
  vrun "$1.run"
  rm -f "$t/stub/work-T1.sh"
}
runwith sub branch.Main.foo-bar
[ "$got" = 0 ] || fail "Branch.Main.Foo-Bar does not ignore branch.Main.foo-bar: the run exited $got:
$(tail -n 15 "$t/sub.run")"
runwith low branch.main.foo-bar
[ "$got" = 9 ] || fail "Branch.Main.Foo-Bar ignored branch.main.foo-bar, whose subsection differs: the run exited $got, want 9:
$(tail -n 15 "$t/low.run")"
grep -Fq 'branch.main.foo-bar' "$t/low.run" || fail "the halt does not name branch.main.foo-bar:
$(tail -n 15 "$t/low.run")"
echo "gate T6: ok"
