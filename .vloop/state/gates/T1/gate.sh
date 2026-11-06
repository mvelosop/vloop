#!/bin/sh
# Gate T1 — Z1, the kit's fixtures without ' in a path. Judged from outside the
# task's reach: the starting point's testdata/ and gen.sh (56bf0155, read with
# git archive), the Python of d7b9cbb5 (read with git show), and the vloop
# built from this tree. The edited gen.sh and the kit's tests are not run here
# (P-8); the go check runs the tests.
G=.vloop/state/gates/T1
. "$G/lib.sh"
START=56bf0155
PYC=d7b9cbb5
K=tools/copilot-acceptance/kit/testdata

# toREPO DIR: every directory named "it's repo" under DIR renamed @REPO@.
toREPO() {
  find "$1" -depth -type d -name "it's repo" | while IFS= read -r d; do
    mv "$d" "$(dirname "$d")/@REPO@" || exit 1
  done || fail "rename under $1"
}

# 1. No path under tools/copilot-acceptance holds ' (the working tree).
find tools/copilot-acceptance -name "*'*" >"$t/quoted" || fail "find tools/copilot-acceptance"
[ ! -s "$t/quoted" ] || fail "paths under tools/copilot-acceptance still hold ':
$(cat "$t/quoted")"

# 2. Every file of testdata/ but gen.sh keeps its bytes, and its path with each
#    component "it's repo" read as @REPO@.
mkdir -p "$t/start" || fail mkdir
git archive "$START" "$K" | tar -x -C "$t/start" || fail "git archive $START $K"
cp -R "$t/start/$K" "$t/want" && rm -f "$t/want/gen.sh" || fail "copy the starting point's testdata"
toREPO "$t/want"
cp -R "$K" "$t/have" && rm -f "$t/have/gen.sh" || fail "copy the working tree's testdata"
diff -r -x .DS_Store -x __pycache__ "$t/want" "$t/have" >"$t/diff" ||
  fail "testdata/ (gen.sh aside) is not the starting point's with it's repo read as @REPO@:
$(head -n 40 "$t/diff")"

# 3. The starting point's gen.sh, on d7b9cbb5's Python and the vloop built
#    here, from a copy of the starting point's testdata/, writes goldens that,
#    with it's repo read as @REPO@, equal the working tree's, file for file.
mkdir -p "$t/py" || fail mkdir
for f in report.py hook.py bypass.sh; do
  git show "$PYC:tools/copilot-acceptance/$f" >"$t/py/$f" || fail "git show $PYC:tools/copilot-acceptance/$f"
done
command -v python3 >/dev/null || fail "python3 is needed to run the reference Python"
go build -buildvcs=false -o "$t/vloop" ./cmd/vloop >"$t/build.out" 2>&1 ||
  fail "go build ./cmd/vloop: $(tail -n 5 "$t/build.out")"
(cd "$t/elsewhere" && HOME="$t/elsewhere" sh "$t/start/$K/gen.sh" "$t/py" "$t/vloop" "$t/gen") >"$t/gen.out" 2>&1 ||
  fail "the starting point's gen.sh failed: $(tail -n 10 "$t/gen.out")"
toREPO "$t/gen"
diff -r -x .DS_Store "$t/gen" "$K/golden" >"$t/diff" ||
  fail "the goldens are not what the starting point's gen.sh writes, with it's repo read as @REPO@:
$(head -n 40 "$t/diff")"

echo "gate T1: ok"
