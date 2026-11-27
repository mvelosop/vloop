$ErrorActionPreference = 'Stop'

foreach ($id in @("T1", "T2", "T3")) {
    & pwsh -NoProfile -File ".vloop/state/gates/$id/gate.ps1"
    if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
}

function Assert-True {
    param([bool]$Condition, [string]$Message)
    if (-not $Condition) { throw $Message }
}

function Invoke-Cli {
    param([string]$Binary, [string[]]$Arguments)
    $output = & $Binary @Arguments 2>&1 | Out-String
    $code = $LASTEXITCODE
    Assert-True ($code -eq 0) "vloop $($Arguments -join ' ') exited ${code}: $output"
    return (($output -replace "`r`n", "`n").TrimEnd("`n"))
}

$temp = Join-Path ([System.IO.Path]::GetTempPath()) ("vloop-gate-t4-" + [guid]::NewGuid().ToString("N"))
try {
    New-Item -ItemType Directory -Path $temp | Out-Null
    $binary = Join-Path $temp "vloop.exe"
    & go build -o $binary ./cmd/vloop
    if ($LASTEXITCODE -ne 0) { throw "build failed" }
    $versionLine = Invoke-Cli $binary @("version")
    $match = [regex]::Match($versionLine, "^vloop (\S+)")
    Assert-True $match.Success "vloop version did not print its version: $versionLine"
    $version = $match.Groups[1].Value

    $repo = Join-Path $temp "repo"
    New-Item -ItemType Directory -Path $repo | Out-Null
    & git -C $repo init -q
    $output = Invoke-Cli $binary @("-C", $repo, "init")
    $lines = $output -split "`n"
    $stampIndex = [array]::IndexOf($lines, "wrote .vloop/install.json")
    $guidelinesIndex = [array]::IndexOf($lines, "wrote .vloop/repo-guidelines.md")
    $readmeIndex = [array]::IndexOf($lines, "wrote .vloop/README.md")
    $briefIndex = [array]::FindIndex($lines, [Predicate[string]] { param($line) $line -like "wrote docs/briefs/*loop-brief.md" })
    Assert-True (($stampIndex -ge 0) -and ($guidelinesIndex -gt $stampIndex) -and ($readmeIndex -gt $guidelinesIndex) -and ($briefIndex -gt $readmeIndex)) "worked example init output is out of order:`n$output"

    $readmePath = Join-Path $repo ".vloop/README.md"
    $readmeLines = Get-Content $readmePath -TotalCount 3
    Assert-True ($readmeLines[0] -ceq "# vloop") "worked example first line differs"
    Assert-True ($readmeLines[1] -ceq "") "worked example has no blank second line"
    Assert-True ($readmeLines[2] -ceq "Written by vloop $version. vloop owns this file: ``vloop upgrade`` rewrites it, so edits here are lost.") "worked example ownership line differs"

    Remove-Item $readmePath
    $output = Invoke-Cli $binary @("-C", $repo, "upgrade")
    Assert-True ($output -eq "updated .vloop/README.md`nalready at $version") "restore output differs:`n$output"

    Add-Content -Path $readmePath -Value "local edit" -NoNewline
    $edited = Get-Content -Raw $readmePath
    $output = Invoke-Cli $binary @("-C", $repo, "upgrade", "--dry-run")
    Assert-True ($output -eq "would update .vloop/README.md`nalready at $version") "dry-run output differs:`n$output"
    Assert-True ((Get-Content -Raw $readmePath) -ceq $edited) "worked example dry-run changed the README"

    $output = Invoke-Cli $binary @("-C", $repo, "upgrade")
    Assert-True ($output -eq "updated .vloop/README.md`nalready at $version") "overwrite output differs:`n$output"
    Assert-True (-not (Get-Content -Raw $readmePath).Contains("local edit")) "worked example did not remove the local edit"

    $output = Invoke-Cli $binary @("-C", $repo, "upgrade")
    Assert-True ($output -eq "already at $version") "current README was not silent:`n$output"
}
finally {
    if (Test-Path $temp) { Remove-Item -Recurse -Force $temp }
}
