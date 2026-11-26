$ErrorActionPreference = 'Stop'
go run .vloop/state/gates/T1/oracle.go
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
