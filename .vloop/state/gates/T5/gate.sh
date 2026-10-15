#!/bin/sh
# Gate T5 — P1: go run ./tools/release prepare <version>, its checks passed,
# creates release/<version> from main with one commit "Release <version>" that
# changes only plugin.json's version, pushes the branch, opens a PR titled
# "Release <version>" whose body is notes from the latest version tag to main,
# and prints the two lines. It never pushes main, tags or merges.
G=$(cd "$(dirname "$0")" && pwd -P) || { echo "GATE FAIL: no gate folder" >&2; exit 1; }
. "$G/lib.sh"
xbuild
buildrel
mkrel

# What the PR body must be: the tool's own notes from the latest tag to main.
rel notes v0.11.0 main
[ "$got" = 0 ] || fail "notes v0.11.0 main: $(said)"
cp "$t/out" "$t/body.want"
main0=$(git -C "$O" rev-parse refs/heads/main) || fail "origin main"
git -C "$O" for-each-ref --format='%(refname) %(objectname)' refs/tags >"$t/otags0"
wmain0=$(git -C "$W" rev-parse refs/heads/main)
: >"$t/stub/gh.log"

rel prepare v0.12.0
printf '%s\n' "prepared release/v0.12.0: https://github.example/mvelosop/vloop/pull/7" \
  "squash-merge it, then: go run ./tools/release publish v0.12.0" >"$t/tail.want"
[ "$got" = 0 ] || fail "prepare v0.12.0: want exit 0; $(said)"
tailis "$t/tail.want"

# The branch on origin: one commit on main, plugin.json's version and nothing else.
rb=$(git -C "$O" rev-parse -q --verify refs/heads/release/v0.12.0) || fail "origin has no release/v0.12.0; $(said)"
[ "$(git -C "$O" rev-parse "$rb^")" = "$main0" ] && [ "$(git -C "$O" rev-list --count "$main0..$rb")" = 1 ] ||
  fail "release/v0.12.0 is not one commit on main: $(git -C "$O" log --oneline "$main0..$rb")"
[ "$(git -C "$O" log -1 --format=%s "$rb")" = "Release v0.12.0" ] || fail "the release commit's subject is '$(git -C "$O" log -1 --format=%s "$rb")', want 'Release v0.12.0'"
printf '1\t1\tplugin/.claude-plugin/plugin.json\n' >"$t/numstat.want"
git -C "$O" diff --numstat "$main0" "$rb" >"$t/numstat"
cmp -s "$t/numstat.want" "$t/numstat" || fail "the release commit changes more than plugin.json's version line: $(cat "$t/numstat")"
git -C "$O" show "$main0:plugin/.claude-plugin/plugin.json" >"$t/pj.old" && git -C "$O" show "$rb:plugin/.claude-plugin/plugin.json" >"$t/pj.new" || fail "plugin.json"
[ "$("$JQ" -r .version "$t/pj.new")" = 0.12.0 ] || fail "plugin.json's version on the branch is $("$JQ" -r .version "$t/pj.new"), want 0.12.0"
[ "$("$JQ" -cS 'del(.version)' "$t/pj.old")" = "$("$JQ" -cS 'del(.version)' "$t/pj.new")" ] || fail "plugin.json changed beyond its version"

# Nothing else written: origin's main and tags, the clone's main.
[ "$(git -C "$O" rev-parse refs/heads/main)" = "$main0" ] || fail "prepare moved origin's main"
git -C "$O" for-each-ref --format='%(refname) %(objectname)' refs/tags >"$t/otags1"
cmp -s "$t/otags0" "$t/otags1" || fail "prepare changed origin's tags: $(diff "$t/otags0" "$t/otags1")"
[ "$(git -C "$W" rev-parse refs/heads/main)" = "$wmain0" ] || fail "prepare committed on main"
[ -z "$(git -C "$W" tag -l 'v0.12*' | grep -vx v0.12)" ] || fail "prepare tagged: $(git -C "$W" tag -l 'v0.12*')"

# The PR.
[ "$(grep -c '^pr create' "$t/stub/gh.log")" = 1 ] || fail "want one gh pr create: $(cat "$t/stub/gh.log")"
! grep -Eq '^(pr merge|release create)' "$t/stub/gh.log" || fail "prepare merged or released: $(cat "$t/stub/gh.log")"
[ "$(cat "$t/stub/pr-title" 2>/dev/null)" = "Release v0.12.0" ] || fail "the PR's title is '$(cat "$t/stub/pr-title" 2>/dev/null)', want 'Release v0.12.0'"
[ -f "$t/stub/pr-body" ] && [ "$(cat "$t/stub/pr-body")" = "$(cat "$t/body.want")" ] || fail "the PR's body is not notes v0.11.0 main: want
$(cat "$t/body.want")
got
$(cat "$t/stub/pr-body" 2>/dev/null)"
echo "gate T5: ok"
