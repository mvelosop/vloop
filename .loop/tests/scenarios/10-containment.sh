#!/usr/bin/env bash
# Hard rules 1 and 2: nothing written outside the repo, and nothing persisted
# names the machine — even when a session emits an absolute path, which is the
# case sessions cannot be trusted to prevent.
. "$(dirname "$0")/../lib.sh"
fixture_new
fixture_plan "$PLAN_ONE"
fixture_stub <<STUB
case "\$PHASE" in
  plan) cat > .loop/state/state.json <<'PLANJSON'
$PLAN_ONE
PLANJSON
    ;;
  work) touch "\$TASK.out"
    # A session leaking an absolute path into its own report.
    jq -nc --arg t "\$TASK" --arg p "\$HOME/secret/file.py" \
      '{task:\$t,outcome:"done",summary:("wrote "+\$p),files:[\$p],verified:"ok",notes:"none"}' > .loop/tmp/proposal.json ;;
  review) jq -nc --arg t "\$TASK" '{task:\$t,verdict:"PASS",criteria:[],findings:[],notes:"none"}' > .loop/tmp/verdict.json ;;
esac
STUB
fixture_run docs/briefs/0003-runstat-cli.md
assert_exit 0
assert_contained
grep -q '~/secret/file.py' "$(fx_journal)" \
  && ok "absolute path was masked to ~ in the journal" \
  || bad "journal did not carry the masked path"
finish
