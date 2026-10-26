#!/bin/sh
# Gate T1 — R1: a nested git repository is its own root (C-2). In a scratch
# vloop repository (git, `vloop init` by this tree's binary, committed):
#   - .vloop/tmp/w, a git repository: `vloop config set language es` there
#     writes .vloop/tmp/w/.vloop/config.toml and leaves the outer config as
#     committed (the worked example's first line; today it edits the outer one);
#   - .vloop/tmp/plain, no .git: `vloop config get language` reads the outer
#     repository's en, as before;
#   - .vloop/tmp/sub, whose .git is a file (git init --separate-git-dir): its
#     own root too;
#   - a directory holding .vloop/ under a git repository with no .vloop/: the
#     directory is the root.
# Then the documents: C-2 in docs/domain/domain-model.md reads the brief's rule
# and the old one is gone there and in docs/domain/platform/platform-context.md,
# which still cites C-2; the verify skill's "Build into a scratch directory"
# section has a scratch made a repository with git init, citing C-2.
set -u
fail() { echo "GATE FAIL: $*" >&2; exit 1; }
for v in $(env | sed -n 's/^\(VLOOP_[A-Z0-9_]*\)=.*/\1/p'); do unset "$v"; done
unset GIT_DIR GIT_WORK_TREE GIT_INDEX_FILE GOFLAGS GOBIN
export LC_ALL=C GIT_CONFIG_GLOBAL=/dev/null GIT_CONFIG_NOSYSTEM=1 GIT_TERMINAL_PROMPT=0 GOTOOLCHAIN=local GOPROXY=off
t=$(mktemp -d "${TMPDIR:-/tmp}/gate.XXXXXX") && t=$(cd "$t" && pwd -P) || fail mktemp
trap 'chmod -R u+w "$t" 2>/dev/null; rm -rf "$t"' EXIT
V="$t/bin/vloop"
mkdir -p "$t/bin" && go build -buildvcs=false -o "$V" ./cmd/vloop || fail "go build of this tree"
export HOME="$t/home"
mkdir -p "$HOME" || fail mkdir

# The scratch vloop repository.
R="$t/repo"
mkdir -p "$R" && git -C "$R" init -q -b main && git -C "$R" config user.name Gate &&
  git -C "$R" config user.email gate@example.invalid && git -C "$R" config commit.gpgsign false || fail "git init"
(cd "$R" && "$V" init >"$t/init.out" 2>&1) || fail "vloop init: $(cat "$t/init.out")"
git -C "$R" add -A && git -C "$R" commit -q -m init || fail "commit"
lang() { sed -n 's/^[[:space:]]*language[[:space:]]*=[[:space:]]*"\([a-z]*\)".*/\1/p' "$1" | head -1; }
outer() {
  [ -z "$(git -C "$R" diff -- .vloop/config.toml)" ] || fail "$1: the outer .vloop/config.toml changed: $(git -C "$R" diff -- .vloop/config.toml | grep '^[-+][^-+]' | head -3)"
  [ "$(lang "$R/.vloop/config.toml")" != es ] || fail "$1: the outer config holds language = es"
}

# A nested git repository.
mkdir -p "$R/.vloop/tmp/w" && git -C "$R/.vloop/tmp/w" init -q || fail "git init w"
(cd "$R/.vloop/tmp/w" && "$V" config set language es >"$t/o" 2>&1) || fail "config set in .vloop/tmp/w exits non-zero: $(cat "$t/o")"
outer "config set language es in the nested repository .vloop/tmp/w"
[ -f "$R/.vloop/tmp/w/.vloop/config.toml" ] || fail "config set in .vloop/tmp/w wrote no .vloop/tmp/w/.vloop/config.toml"
[ "$(lang "$R/.vloop/tmp/w/.vloop/config.toml")" = es ] || fail ".vloop/tmp/w/.vloop/config.toml does not hold language = \"es\""
g=$(cd "$R/.vloop/tmp/w" && "$V" config get language 2>&1) || fail "config get in .vloop/tmp/w: $g"
[ "$g" = es ] || fail "config get language in .vloop/tmp/w prints '$g', want es"

# No nearer .git: the outer repository, as before.
mkdir -p "$R/.vloop/tmp/plain"
g=$(cd "$R/.vloop/tmp/plain" && "$V" config get language 2>&1) || fail "config get in .vloop/tmp/plain exits non-zero: $g"
[ "$g" = en ] || fail "config get language in .vloop/tmp/plain prints '$g', want en (the outer repository's)"

# A .git file counts as .git.
git init -q --separate-git-dir "$t/sub.git" "$R/.vloop/tmp/sub" || fail "git init --separate-git-dir"
[ -f "$R/.vloop/tmp/sub/.git" ] || fail "fixture: .vloop/tmp/sub/.git is not a file"
(cd "$R/.vloop/tmp/sub" && "$V" config set language es >"$t/o" 2>&1) || fail "config set in .vloop/tmp/sub exits non-zero: $(cat "$t/o")"
outer "config set language es under a .git file (.vloop/tmp/sub)"
[ "$(lang "$R/.vloop/tmp/sub/.vloop/config.toml" 2>/dev/null)" = es ] || fail "config set under a .git file wrote no language = \"es\" to .vloop/tmp/sub/.vloop/config.toml"

# The start directory holding .vloop/ under a git repository without one.
G="$t/g"
mkdir -p "$G/a/.vloop" && git -C "$G" init -q || fail "git init g"
(cd "$G/a" && "$V" config set language es >"$t/o" 2>&1) || fail "config set in g/a exits non-zero: $(cat "$t/o")"
[ "$(lang "$G/a/.vloop/config.toml" 2>/dev/null)" = es ] || fail "a start directory holding .vloop/ is not its own root: g/a/.vloop/config.toml lacks language = \"es\""
[ ! -e "$G/.vloop" ] || fail "config set in g/a wrote the enclosing git repository's .vloop/"

# The documents.
flat() { tr '\n' ' ' <"$1" | tr -s ' \t' '  '; }
DM=docs/domain/domain-model.md PC=docs/domain/platform/platform-context.md SK=plugin/skills/verify/SKILL.md
c2=$(flat "$DM" | sed -n 's/.*\(\*\*C-2\*\*.*\)\*\*C-3\*\*.*/\1/p')
[ -n "$c2" ] || fail "$DM: no C-2 entry followed by C-3"
case "$c2" in *'nearest ancestor (the start directory included) holding `.vloop/` or `.git`, whichever is found first; else the start directory.'*) ;;
  *) fail "$DM: C-2 does not read the brief's rule: $c2";; esac
case "$c2" in *'Every path vloop prints is relative to it, with `/`.'*) ;; *) fail "$DM: C-2 lost 'Every path vloop prints is relative to it, with \`/\`.'";; esac
for f in "$DM" "$PC"; do
  case "$(flat "$f")" in *'holding `.vloop/`, else `.git`'*) fail "$f still states the old rule: holding \`.vloop/\`, else \`.git\`";; esac
done
grep -qF 'C-2' "$PC" || fail "$PC no longer cites C-2"
sec=$(awk '/^## 2\. Build into a scratch directory/ { on = 1; next } on && /^## / { exit } on' "$SK")
[ -n "$sec" ] || fail "$SK has no section '## 2. Build into a scratch directory'"
printf '%s\n' "$sec" | grep -qF 'git init' || fail "$SK's scratch section does not say to make a scratch a repository with git init"
printf '%s\n' "$sec" | grep -qF 'C-2' || fail "$SK's scratch section does not cite C-2"
echo "gate T1: ok"
