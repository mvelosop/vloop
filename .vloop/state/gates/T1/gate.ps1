$ErrorActionPreference = 'Stop'
& go test -count=1 ./.vloop/state/gates/T1
if ($LASTEXITCODE -ne 0) {
  exit $LASTEXITCODE
}
