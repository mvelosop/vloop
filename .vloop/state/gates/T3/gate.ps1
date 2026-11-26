$ErrorActionPreference = "Stop"
& pwsh -NoProfile -File .vloop/state/gates/T1/gate.ps1
if ($LASTEXITCODE -ne 0) { throw "T1 gate failed" }

& pwsh -NoProfile -File .vloop/state/gates/T2/gate.ps1
if ($LASTEXITCODE -ne 0) { throw "T2 gate failed" }
