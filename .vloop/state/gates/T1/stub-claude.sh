#!/bin/sh
# The B34 gates' stub claude: every phase of their fixtures runs on Copilot,
# so it only answers the probes a preflight may make, and refuses a session.
case "${1:-}" in
  --version) echo "2.0.0 (Claude Code)"; exit 0;;
  plugin) echo '[{"id":"vloop@vloop","version":"@PLUGIN_VERSION@","enabled":true}]'; exit 0;;
esac
echo "stub claude: no session runs on claude in these fixtures" >&2
exit 64
