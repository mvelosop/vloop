$ErrorActionPreference = 'Stop'
foreach ($gate in @(
  '.vloop/state/gates/T1/gate.ps1',
  '.vloop/state/gates/T2/gate.ps1',
  '.vloop/state/gates/T3/gate.ps1',
  '.vloop/state/gates/T4/gate.ps1',
  '.vloop/state/gates/T5/gate.ps1'
)) {
  & $gate
  if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
}
go run .vloop/state/gates/T6/oracle.go
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
