#!/bin/sh
# Gate T5 — L4, Claude only where it is used. vloop doctor and vloop run are
# run on a PATH holding a stub copilot and git but no claude, with a GitHub
# token set, on fixture repositories whose agent keys choose who is in use.
G=.vloop/state/gates/T5
. "$G/lib.sh"

mkdir -p "$t/bin" && cp "$t/cstub/copilot" "$t/bin/copilot" && ln -s "$(command -v git)" "$t/bin/git" &&
  ln -s "$(command -v python3)" "$t/bin/python3" || fail "the stub PATH"
PATH="$t/bin:/usr/bin:/bin"
export PATH GH_TOKEN=gate-token
! command -v claude >/dev/null 2>&1 || fail "a claude is installed in /usr/bin or /bin: this gate needs a PATH without one"

NA="no session in use runs on claude"
MISSING="✗ claude claude is not on PATH"

# cfg REPO EXTRA: REPO's shared config sets every agent.* key to copilot, every
# model.* key to a Copilot model id and run.verify = false, then the TOML EXTRA
# (dotted keys that replace these), then the check; committed.
cfg() {
  python3 -I -c 'import sys
over=dict(l.split(" = ",1) for l in sys.argv[2].splitlines() if l.strip())
phases=["plan","work","review","second-review","gate-review","verify","review-brief"]
base={}
for p in phases:
    base["agent."+p]="\"copilot\""
    base["model."+p]="\"gpt-6.1-terra\""
base["run.verify"]="\"false\""
base.update(over)
with open(sys.argv[1],"w") as f:
    for k,v in base.items(): f.write(k+" = "+v+"\n")
    f.write("\n[[check]]\nname = \"all\"\npaths = [\"**\"]\nrun = \"true\"\n")' "$1/.vloop/config.toml" "$2" || fail "write the config"
  commitall "$1" "config: $2"
}

# narows FILE: FILE holds the four n/a rows, each on its own line.
narows() {
  for n in claude trust plugin "gate edits"; do
    hasline "$2" "$1" "- $n $NA"
  done
}

# runs REPO: vloop run of the demo brief in REPO; the copilot log emptied first.
runs() { : >"$t/bin/copilot.log"; runin "$1" run "$BRIEF"; }

# planned NAME: the run got past its preflight into planning on copilot.
planned() {
  ! grep -q 'preflight failed' "$t/err" || fail "$1: the run failed its preflight:
$(cat "$t/err")"
  grep -q '^arg /vloop:plan' "$t/bin/copilot.log" || fail "$1: no copilot plan session ran:
$(cat "$t/err")"
}

# 1. Every agent copilot, run.verify false: four n/a rows, the run plans.
initrepo s1
cfg "$R" ""
runin "$R" doctor
narows "$t/out" "doctor, every agent on copilot"
for n in claude trust plugin; do
  rows "$n" "$t/out" | grep -v -- "^- $n $NA\$" >"$t/rows"
  [ ! -s "$t/rows" ] || fail "doctor, every agent on copilot, still ran the $n row:
$(cat "$t/rows")"
done
runs "$R"
planned "vloop run, every agent on copilot"

# 2. agent.verify claude while run.verify is false, agent.second-review claude
#    while run.second-review is off: not in use.
initrepo s2
cfg "$R" 'agent.verify = "claude"
agent.second-review = "claude"'
runin "$R" doctor
narows "$t/out" "doctor, claude only on the verify and second review that are off"
runs "$R"
planned "vloop run, claude only on the verify and second review that are off"

# 3. agent.review-brief claude: in use for doctor, not for a run.
initrepo s3
cfg "$R" 'agent.review-brief = "claude"'
runin "$R" doctor
hasline "doctor, agent.review-brief on claude" "$t/out" "$MISSING"
noline "doctor, agent.review-brief on claude" "$t/out" "- claude $NA"
runs "$R"
planned "vloop run, agent.review-brief on claude"

# 4. agent.review claude: in use; doctor and the run's preflight fail on it.
initrepo s4
cfg "$R" 'agent.review = "claude"'
runin "$R" doctor
hasline "doctor, agent.review on claude" "$t/out" "$MISSING"
for n in claude trust plugin "gate edits"; do
  noline "doctor, agent.review on claude" "$t/out" "- $n $NA"
done
runs "$R"
wantrc "vloop run, agent.review on claude and no claude" 1
hasline "vloop run, agent.review on claude (stderr)" "$t/err" "  $MISSING"
grep -q '^vloop: preflight failed ' "$t/err" || fail "vloop run, agent.review on claude: no preflight refusal:
$(cat "$t/err")"
! grep -q '^arg /vloop:' "$t/bin/copilot.log" || fail "vloop run, agent.review on claude: a session ran"

# 5. agent.verify claude with run.verify true, agent.second-review claude with
#    a second review on: in use.
initrepo s5
cfg "$R" 'agent.verify = "claude"
run.verify = "true"'
runin "$R" doctor
hasline "doctor, agent.verify on claude, run.verify true" "$t/out" "$MISSING"
initrepo s6
cfg "$R" 'agent.second-review = "claude"
run.second-review = "record"'
runin "$R" doctor
hasline "doctor, agent.second-review on claude, run.second-review record" "$t/out" "$MISSING"

# 6. A config that cannot be read: Claude counts as in use.
initrepo s7
cfg "$R" 'agent.work = "nope"'
runin "$R" doctor
grep -q '^✗ config ' "$t/out" || fail "doctor, agent.work nope: no ✗ config row:
$(cat "$t/out")"
hasline "doctor, a config that cannot be read" "$t/out" "$MISSING"
noline "doctor, a config that cannot be read" "$t/out" "- claude $NA"

echo "gate T5: ok"
