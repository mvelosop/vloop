$ErrorActionPreference = 'Stop'
go run .vloop/state/gates/T2/oracle.go
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
