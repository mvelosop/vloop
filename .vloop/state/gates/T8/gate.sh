#!/bin/sh
# Gate T8 — the cross-platform findings of brief Q9, as seen on this host
# (macOS): the trust lookup in ~/.claude.json compares paths case-insensitively,
# so a trusted key that differs from the repository's path only in case is
# trusted (and another path is not); and the pre-commit hook check finds the
# hooks directory git uses, so a hook in the common hooks directory is reported
# from a linked worktree too. Windows' rules (backslashes, hook presence) are
# pure functions the task's own tests cover; this gate builds for windows.
set -u
fail() { echo "GATE FAIL: $*" >&2; exit 1; }
for v in $(env | sed -n 's/^\(VLOOP_[A-Z0-9_]*\)=.*/\1/p'); do unset "$v"; done
unset NO_COLOR GIT_DIR GIT_WORK_TREE GIT_INDEX_FILE CLAUDE_PLUGIN_ROOT
export LC_ALL=C GIT_CONFIG_GLOBAL=/dev/null GIT_CONFIG_NOSYSTEM=1 GIT_TERMINAL_PROMPT=0
export GIT_AUTHOR_NAME=gate GIT_AUTHOR_EMAIL=gate@example.invalid GIT_COMMITTER_NAME=gate GIT_COMMITTER_EMAIL=gate@example.invalid
G=$(cd "$(dirname "$0")" && pwd -P) || fail "no gate folder"
t=$(mktemp -d "${TMPDIR:-/tmp}/gate.XXXXXX") && t=$(cd "$t" && pwd -P) || fail mktemp
trap 'rm -rf "$t"' EXIT
go build -o "$t/vloop" ./cmd/vloop || fail "go build ./cmd/vloop"
GOOS=linux go build -o "$t/vloop-linux" ./cmd/vloop || fail "GOOS=linux go build ./cmd/vloop"
GOOS=windows go build -o "$t/vloop.exe" ./cmd/vloop || fail "GOOS=windows go build ./cmd/vloop"
B="$t/vloop"
# A claude that answers doctor's probes and runs no session.
mkdir -p "$t/home" "$t/bin" || fail mkdir
cat >"$t/bin/claude" <<'EOF'
#!/bin/sh
case "$1" in
  --version) echo "2.0.0 (Claude Code)"; exit 0;;
  plugin) echo '[]'; exit 0;;
esac
exit 1
EOF
chmod +x "$t/bin/claude" || fail chmod
export HOME="$t/home" PATH="$t/bin:$PATH"
R="$t/repo"
mkdir -p "$R" && git -C "$R" init -q -b main || fail "git init"
git -C "$R" config user.name gate && git -C "$R" config user.email gate@example.invalid || fail "identity"
(cd "$R" && "$B" init </dev/null) >"$t/init.out" 2>&1 || fail "vloop init: $(cat "$t/init.out")"
mkdir -p "$R/docs/briefs" && echo x >"$R/docs/x.md" && cp "$G/brief.md" "$R/docs/briefs/B20260101-0900-demo.loop-brief.md" || fail "brief"
printf '[[check]]\nname = "all"\npaths = ["**"]\nrun = "true"\n' >"$R/.vloop/config.toml" || fail config
git -C "$R" add -A && git -C "$R" commit -q -m fixture || fail "commit"
trust() { printf '{"projects":{"%s":{"hasTrustDialogAccepted":true}}}\n' "$1" >"$HOME/.claude.json"; }
doctor_trust() { (cd "$R" && "$B" doctor) 2>&1 | grep ' trust' | head -1; }

# --- trust: case-insensitive on this host, and not trusting just anything ---
trust "$R"
doctor_trust | grep -q '^✓ trust' || fail "doctor does not trust a repository whose exact path is trusted: $(doctor_trust)"
trust "$(echo "$R" | tr '[:lower:]' '[:upper:]')"
doctor_trust | grep -q '^✓ trust' || fail "doctor does not trust a repository whose trusted key differs only in case (macOS paths are case-insensitive): $(doctor_trust)"
trust "$t/elsewhere"
doctor_trust | grep -q '^✓ trust' && fail "doctor trusts a repository whose path is not in ~/.claude.json"

# --- the pre-commit hook seen from a linked worktree ---
git -C "$R" worktree add -q "$t/wt" -b wt || fail "worktree add"
printf '#!/bin/sh\nexit 0\n' >"$t/bundled-hook" || fail hook
# vloop run's preflight: not trusted here, so it refuses before anything runs.
trust "$t/elsewhere"
(cd "$t/wt" && "$B" run docs/briefs/B20260101-0900-demo.loop-brief.md --plan-only) >"$t/wt0.out" 2>&1
grep -q 'pre-commit hook is active' "$t/wt0.out" && fail "vloop run reports a pre-commit hook in a worktree that has none: $(cat "$t/wt0.out")"
mkdir -p "$(git -C "$t/wt" rev-parse --path-format=absolute --git-path hooks)" || fail "hooks dir"
HOOK="$(git -C "$t/wt" rev-parse --path-format=absolute --git-path hooks)/pre-commit"
cp "$t/bundled-hook" "$HOOK" && chmod +x "$HOOK" || fail "plant the hook"
(cd "$t/wt" && "$B" run docs/briefs/B20260101-0900-demo.loop-brief.md --plan-only) >"$t/wt.out" 2>&1
grep -q 'pre-commit hook is active' "$t/wt.out" || fail "vloop run in a linked worktree does not report the active pre-commit hook git would run: $(cat "$t/wt.out")"
[ -e "$t/wt/.vloop/state/state.json" ] && fail "vloop run planned in an untrusted worktree"
echo "gate T8: ok"
