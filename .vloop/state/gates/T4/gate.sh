# T4 (K4): the two eval cases, the evals README's table and the evals guide.
# The cases' structure, grants and pinned graders by the gate's oracle in a copy
# of the tree; the two documents by their text.
G=.vloop/state/gates/T4
. "$G/lib.sh"
cd "$top" || fail cd

for c in review-brief-planted-faults create-brief-drafts-checked; do
  [ -d "plugin/evals/$c" ] || fail "no eval case plugin/evals/$c"
done
n=0
for d in plugin/evals/*/; do
  case $(basename "$d") in _*) continue ;; esac
  n=$((n + 1))
done
[ "$n" = 21 ] || fail "plugin/evals holds $n cases, want 21"

# plugin/evals/README.md's table of other cases.
grep -qE '^\| `review-brief-planted-faults` \| `/vloop:review-brief` \|' plugin/evals/README.md ||
  fail "plugin/evals/README.md has no table row | \`review-brief-planted-faults\` | \`/vloop:review-brief\` | …"
grep -qE '^\| `create-brief-drafts-checked` \| `/vloop:create-brief` \|' plugin/evals/README.md ||
  fail "plugin/evals/README.md has no table row | \`create-brief-drafts-checked\` | \`/vloop:create-brief\` | …"

# docs/guide/evals.md: the suite size and the suites table.
E=docs/guide/evals.md
grep -qF '21 cases × 3 runs × 2 arms' "$E" || fail "$E does not say 21 cases × 3 runs × 2 arms"
grep -qF '16 cases × 3 runs × 2 arms' "$E" && fail "$E still says 16 cases × 3 runs × 2 arms"
grep -qF '`/vloop:operate` has no suite' "$E" && fail "$E still says /vloop:operate has no suite"
row() { grep -E "^\| \`$1\` \|[^|]*$2[^|]*\|" "$E" >/dev/null || fail "$E's suites table has no row for \`$1\` naming $2"; }
row /vloop:create-brief create-brief-drafts-checked
row /vloop:review-brief review-brief-planted-faults
row /vloop:verify verify-planted-mismatches
row /vloop:operate operate-proposes-options

srccopy
oracle zz_gate_t4_evals_test.go internal/cli
(cd "$t/src" && go test -count=1 -run '^TestGateT4' ./internal/cli >"$t/oracle.out" 2>&1) ||
  fail "the eval cases' oracle failed:
$(tail -40 "$t/oracle.out")"
echo "T4 gate: ok"
