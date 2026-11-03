# Shared set-up of the fixture cases; sourced by each scaffold through its own path.
require_empty() { [ -z "$(ls -A)" ] || { echo "the workspace is not empty" >&2; exit 1; }; }
make_tmp() { mkdir -p .vloop/tmp; }
