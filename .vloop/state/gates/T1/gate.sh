# T1 (K1): the two brief skills in the plugin — the worked example's first
# table, the moved files, /vloop:operate's description, the budget's list and
# ceilings with the padded-skill failure, the fence's decisions on the two names,
# the extracted plugin's templates and doctor's clash line for review-brief.
G=.vloop/state/gates/T1
. "$G/lib.sh"
cd "$top" || fail cd

# The worked example's first table, on the tree.
got=$(listing plugin/skills)
[ "$got" = "create-brief gate-review operate plan review review-brief verify work" ] ||
  fail "ls plugin/skills: '$got', want 'create-brief gate-review operate plan review review-brief verify work'"
[ -d plugin/skills/create-brief ] || fail "plugin/skills/create-brief is not a folder"
got=$(listing plugin/skills/create-brief)
[ "$got" = "SKILL.md guidelines.md templates" ] || fail "ls plugin/skills/create-brief: '$got', want 'SKILL.md guidelines.md templates'"
got=$(listing plugin/skills/create-brief/templates)
[ "$got" = "brief-en.md brief-es.md" ] || fail "ls plugin/skills/create-brief/templates: '$got', want 'brief-en.md brief-es.md'"
got=$(listing plugin/skills/review-brief)
[ "$got" = "SKILL.md" ] || fail "ls plugin/skills/review-brief: '$got', want 'SKILL.md'"
if hits=$(grep -rliE 'vloop-(create|review)-brief|experimental|not yet in the plugin' plugin/skills); then
  fail "files under plugin/skills still name the project skills or the experimental label: $hits"
fi
grep -q '/vloop:review-brief' plugin/skills/create-brief/SKILL.md || fail "create-brief's SKILL.md does not name /vloop:review-brief"
grep -q '/vloop:operate' plugin/skills/create-brief/SKILL.md || fail "create-brief's SKILL.md does not name /vloop:operate"
grep -q '/vloop:create-brief' plugin/skills/review-brief/SKILL.md || fail "review-brief's SKILL.md does not name /vloop:create-brief"
grep -qF '../create-brief/guidelines.md' plugin/skills/review-brief/SKILL.md ||
  fail "review-brief's SKILL.md does not point at ../create-brief/guidelines.md"
[ "$(sed -n 2p plugin/skills/create-brief/SKILL.md)" = "name: create-brief" ] || fail "line 2 of create-brief's SKILL.md is not 'name: create-brief'"
[ "$(sed -n 2p plugin/skills/review-brief/SKILL.md)" = "name: review-brief" ] || fail "line 2 of review-brief's SKILL.md is not 'name: review-brief'"

# Each description is the original's, its experimental prefix dropped (the
# first letter of what is left may be capitalised).
for s in create review; do
  want=$(git -C "$top" show "$BASE:.claude/skills/vloop-$s-brief/SKILL.md" | sed -n 3p |
    sed 's/^description: Experimental, not yet in the plugin — /description: /' | tr 'A-Z' 'a-z') || fail "git show of the original $s-brief"
  case $want in "description: experimental"*) fail "the original $s-brief description did not drop its prefix" ;; esac
  got=$(sed -n 3p "plugin/skills/$s-brief/SKILL.md" | tr 'A-Z' 'a-z')
  [ "$got" = "$want" ] || fail "line 3 of $s-brief's SKILL.md is not the original description without its experimental prefix: $(sed -n 3p "plugin/skills/$s-brief/SKILL.md")"
done

# guidelines.md and the templates move byte for byte.
for f in guidelines.md templates/brief-en.md templates/brief-es.md; do
  git -C "$top" show "$BASE:.claude/skills/vloop-create-brief/$f" >"$t/orig" || fail "git show of the original $f"
  cmp -s "$t/orig" "plugin/skills/create-brief/$f" || fail "plugin/skills/create-brief/$f is not .claude/skills/vloop-create-brief/$f byte for byte"
done

# /vloop:operate: its description's last sentence, and nothing else.
git -C "$top" show "$BASE:plugin/skills/operate/SKILL.md" |
  sed "s|Writing briefs is not this skill's job\.\$|Writing briefs is /vloop:create-brief's.|" >"$t/operate.want" || fail "git show of operate"
grep -qF "Writing briefs is /vloop:create-brief's." "$t/operate.want" || fail "the gate's expected operate SKILL.md was not built"
cmp -s "$t/operate.want" plugin/skills/operate/SKILL.md ||
  fail "plugin/skills/operate/SKILL.md is not the base's with its description ending \"Writing briefs is /vloop:create-brief's.\" and nothing else changed"

# The budget (the padded-skill failure) and doctor's clash line, by the gate's
# oracles in a copy of the tree.
srccopy
oracle zz_gate_t1_budget_test.go internal/budget
oracle zz_gate_t1_doctor_test.go internal/cli
(cd "$t/src" && go test -count=1 -run '^TestGateT1' ./internal/budget ./internal/cli >"$t/oracle.out" 2>&1) ||
  fail "the budget or doctor oracle failed:
$(tail -40 "$t/oracle.out")"

# The fence's decisions on the two names.
newbin
out=$("$S" fence check --phase review --root . --tool skill --path review-brief 2>&1); rc=$?
[ "$rc" = 0 ] && [ "$out" = "allow" ] || fail "fence check review-brief: exit $rc, '$out', want allow, exit 0"
out=$("$S" fence check --phase review --root . --tool skill --path vloop-review-brief 2>&1); rc=$?
[ "$rc" = 1 ] && [ "$out" = "deny: skill vloop-review-brief is not vloop's — use /vloop:review" ] ||
  fail "fence check vloop-review-brief: exit $rc, '$out', want the deny line, exit 1"

# The extracted plugin holds create-brief's templates.
freshrepo
p=$(cd "$t/repo" && venv "$S" plugin path 2>"$t/pp.err") || fail "vloop plugin path: $(cat "$t/pp.err")"
got=$(cd "$t/repo" && listing "$p/skills/create-brief/templates")
[ "$got" = "brief-en.md brief-es.md" ] || fail "the extracted plugin's create-brief/templates: '$got', want 'brief-en.md brief-es.md'"
echo "T1 gate: ok"
