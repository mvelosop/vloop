$ErrorActionPreference = "Stop"

foreach ($task in @("T1", "T2", "T3", "T4")) {
    & pwsh -NoProfile -File (Join-Path $PSScriptRoot ("..\$task\gate.ps1"))
    if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
}

. (Join-Path $PSScriptRoot "..\T1\harness.ps1")
$h = New-GateHarness -Agents @("claude")
try {
    $env:FAKE_CLAUDE_LIST = '[{"id":"vloop@vloop","version":"0.14.0","scope":"user","enabled":true}]'
    $r = Invoke-GateCommand $h.Exe @("self-update")
    $want = @(
        "installing vloop v0.15.0 (from v0.14.0)…"
        "installed vloop v0.15.0 at $(Format-GatePath $h.Gobin)"
        "plugin claude (user): updated"
        "next: vloop upgrade in each repository that uses vloop, and vloop self-update there for a plugin installed for that repository"
    ) -join "`n"
    Assert-Equal $r.Code 0 "worked example exits 0"
    Assert-Equal $r.Out $want "worked example output line by line"
    Assert-Equal $r.Err "" "worked example stderr"
} finally {
    Remove-Item $h.Temp -Recurse -Force -ErrorAction SilentlyContinue
    $env:Path = $script:OriginalPath
}
