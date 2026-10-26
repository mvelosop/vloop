#!/bin/sh
# T4's stub claude. It answers doctor's two probes; any session prompt (-p) is
# logged to sessions.log beside it and refused with exit 64, so a run that gets
# past its preflight stops at its first session.
dir=$(dirname "$0")
case "$1" in
  --version) echo "2.0.0 (Claude Code)"; exit 0;;
  plugin) echo '[{"id":"vloop@vloop","version":"@PLUGIN_VERSION@","enabled":true}]'; exit 0;;
esac
while [ $# -gt 0 ]; do
  case "$1" in
    -p) echo "$2" >>"$dir/sessions.log"; shift 2;;
    *) shift;;
  esac
done
echo "stub claude: no sessions in this gate" >&2
exit 64
