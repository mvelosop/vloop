#!/bin/sh
# Gate T1 — vloop init writes an empty .vloop/repo-guidelines.md (brief L2):
# `wrote .vloop/repo-guidelines.md` right after `wrote .vloop/install.json`;
# the template has `## Pinning tests` and `## External formats` and no rule
# outside an HTML comment; its prose follows `language`, its headings do not;
# an existing file is kept untouched and nothing is printed for it.
G=$(cd "$(dirname "$0")" && pwd -P) || { echo "GATE FAIL: no gate folder" >&2; exit 1; }
. "$G/lib.sh"
GF=.vloop/repo-guidelines.md
# template <repo> <init output>: the checks every fresh template passes.
template() {
  f="$t/$1/$GF"
  [ -f "$f" ] && [ -s "$f" ] || fail "$1: vloop init wrote no $GF:
$(cat "$t/$2")"
  nxt=$(awk 'p { print; exit } $0 == "wrote .vloop/install.json" { p = 1 }' "$t/$2")
  [ "$nxt" = "wrote $GF" ] || fail "$1: the line after 'wrote .vloop/install.json' is '$nxt', not 'wrote $GF':
$(cat "$t/$2")"
  [ "$(grep -Fxc "wrote $GF" "$t/$2")" = 1 ] || fail "$1: 'wrote $GF' is not printed exactly once"
  for h in '## Pinning tests' '## External formats'; do
    grep -Fxq -- "$h" "$f" || fail "$1: the template has no line '$h':
$(cat "$f")"
  done
  # A rule is a line starting `- ` outside an HTML comment (brief L9).
  rules=$(awk '
    { line = $0; out = "" ; i = 1
      while (i <= length(line)) {
        if (inc) { j = index(substr(line, i), "-->"); if (j == 0) { i = length(line) + 1 } else { inc = 0; i += j + 2 } }
        else { j = index(substr(line, i), "<!--"); if (j == 0) { out = out substr(line, i); i = length(line) + 1 } else { out = out substr(line, i, j - 1); inc = 1; i += j + 3 } }
      }
      if (substr(out, 1, 2) == "- " && substr(line, 1, 2) == "- ") n++ }
    END { print n + 0 }' "$f")
  [ "$rules" = 0 ] || fail "$1: the template has $rules rule line(s) outside an HTML comment; its examples must be commented out:
$(cat "$f")"
}

newrepo en
template en en.init
R="$t/es0"
mkdir -p "$R/docs" && git -C "$R" init -q -b main || fail "git init es0"
git -C "$R" config user.name gate && git -C "$R" config user.email gate@example.invalid || fail identity
printf 'x\n' >"$R/docs/x.md"
inr "$B" init --language es </dev/null >"$t/es.init" 2>&1 || fail "vloop init --language es: $(cat "$t/es.init")"
template es0 es.init
cmp -s "$t/en/$GF" "$t/es0/$GF" && fail "the template's prose does not follow language: the es file equals the en file"

# An existing file is kept, byte for byte, and nothing is printed for it.
R="$t/kept"
mkdir -p "$R/.vloop" "$R/docs" && git -C "$R" init -q -b main || fail "git init kept"
git -C "$R" config user.name gate && git -C "$R" config user.email gate@example.invalid || fail identity
printf '# Mine\n\n- keep me (D20261011-1455)\n' >"$R/$GF" && cp "$R/$GF" "$t/mine" || fail fixture
inr "$B" init </dev/null >"$t/kept.init" 2>&1; got=$?
[ "$got" = 0 ] || fail "vloop init with an existing $GF: exit $got: $(cat "$t/kept.init")"
cmp -s "$R/$GF" "$t/mine" || fail "vloop init changed an existing $GF:
$(cat "$R/$GF")"
grep -Fq "$GF" "$t/kept.init" && fail "vloop init printed a line for an existing $GF:
$(cat "$t/kept.init")"
grep -Fxq 'wrote .vloop/install.json' "$t/kept.init" || fail "vloop init with an existing $GF did not write install.json: $(cat "$t/kept.init")"
echo "gate T1: ok"
