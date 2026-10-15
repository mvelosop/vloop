#!/bin/sh
# Gate T2 — K2: vloop brief check no longer reads "## Pinning tests" or
# "## External formats" from .vloop/repo-guidelines.md and never prints a
# must-change list: or external format: line; every other line, and the exit
# code, are what the plan's base prints. vloop init's guidelines file is the
# title and its one sentence, in en and es. doctor's guidelines line is as the
# base prints it.
G=$(cd "$(dirname "$0")" && pwd -P) || { echo "GATE FAIL: no gate folder" >&2; exit 1; }
. "$G/lib.sh"
basebin
BB=$t/vloop-base
nonblank() { grep -v '^[[:space:]]*$' "$1"; }

# --- init's guidelines file, en and es, against the base's first two lines
for lang in en es; do
  newrepo "n$lang" --language "$lang"
  N=$R/.vloop/repo-guidelines.md
  R=$t/b$lang
  mkdir -p "$R" && git -C "$R" init -q -b main || fail "git init $R"
  inr "$BB" init --language "$lang" </dev/null >"$R.init" 2>&1 || fail "base vloop init: $(cat "$R.init")"
  nonblank "$R/.vloop/repo-guidelines.md" | head -n 2 >"$t/want-$lang"
  nonblank "$N" >"$t/got-$lang"
  cmp -s "$t/want-$lang" "$t/got-$lang" || fail "vloop init --language $lang: the guidelines file is not the title and its one sentence:
$(cat "$N")
want exactly (blank lines aside):
$(cat "$t/want-$lang")"
done

# --- brief check with a guidelines file that has both sections
newrepo k2
cat >"$R/.vloop/repo-guidelines.md" <<'GL'
# Lessons of this repository

This file holds this repository's lessons, and every loop session receives it.

## Pinning tests

- when `config key`: `TestConfigList`, `TestConfigGet`

## External formats

- `manifest`: `docs/manifest-spec.md`
GL
briefdoc "$RUN" ready | sed 's/^An unknown id exits 1\.$/An unknown id exits 1. A new config key changes the manifest./' >"$R/$BRIEF" || fail brief
grep -q 'config key changes the manifest' "$R/$BRIEF" || fail "the fixture brief has no contract sentence to match"
commit fixtures
for b in base:"$BB" new:"$B"; do
  n=${b%%:*} bin=${b#*:}
  inr "$bin" brief check "$BRIEF" </dev/null >"$t/chk-$n" 2>&1; echo $? >"$t/chk-$n.exit"
  inr "$bin" brief check "$BRIEF" --json </dev/null >"$t/chkj-$n" 2>&1; echo $? >"$t/chkj-$n.exit"
  inr "$bin" doctor </dev/null >"$t/doc-$n" 2>&1
done
grep -q 'must-change list:' "$t/chk-base" && grep -q 'external format:' "$t/chk-base" ||
  fail "harness: the base does not warn on the fixture, so the comparison proves nothing: $(cat "$t/chk-base")"
grep -q -e 'must-change list:' -e 'external format:' "$t/chk-new" && fail "brief check still prints a retired warning:
$(cat "$t/chk-new")"
grep -q -e 'must-change list' -e 'external format' "$t/chkj-new" && fail "brief check --json still carries a retired warning:
$(cat "$t/chkj-new")"
retire "$t/chk-base" >"$t/chk-want"
cmp -s "$t/chk-want" "$t/chk-new" || fail "brief check's other lines changed:
$(diff "$t/chk-want" "$t/chk-new")"
cmp -s "$t/chk-base.exit" "$t/chk-new.exit" || fail "brief check exits $(cat "$t/chk-new.exit"), the base $(cat "$t/chk-base.exit")"
cmp -s "$t/chkj-base.exit" "$t/chkj-new.exit" || fail "brief check --json exits $(cat "$t/chkj-new.exit"), the base $(cat "$t/chkj-base.exit")"
grep 'guidelines' "$t/doc-base" >"$t/dg-base"; grep 'guidelines' "$t/doc-new" >"$t/dg-new"
[ -s "$t/dg-base" ] || fail "harness: the base's doctor prints no guidelines line"
cmp -s "$t/dg-base" "$t/dg-new" || fail "doctor's guidelines line changed:
base: $(cat "$t/dg-base")
now:  $(cat "$t/dg-new")"
echo "gate T2: ok"
