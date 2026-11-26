$ErrorActionPreference = 'Stop'

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

$temp = Join-Path ([System.IO.Path]::GetTempPath()) ("vloop-gate-t2-" + [guid]::NewGuid().ToString("N"))
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
    Invoke-Cli $binary @("-C", $repo, "init", "--language", "es") | Out-Null
    $readmePath = Join-Path $repo ".vloop/README.md"

    Remove-Item $readmePath -ErrorAction SilentlyContinue
    $output = Invoke-Cli $binary @("-C", $repo, "upgrade")
    Assert-True ($output -eq "updated .vloop/README.md`nalready at $version") "missing same-version README output was:`n$output"
    $restored = Get-Content -Raw $readmePath
    Assert-True $restored.Contains("## Qué es vloop") "restored README ignored resolved Spanish language"
    Assert-True $restored.Contains("Escrito por vloop $version.") "restored README lacks the current version"

    $output = Invoke-Cli $binary @("-C", $repo, "upgrade")
    Assert-True ($output -eq "already at $version") "current README printed extra output:`n$output"

    Add-Content -Path $readmePath -Value "local edit" -NoNewline
    $edited = Get-Content -Raw $readmePath
    $output = Invoke-Cli $binary @("-C", $repo, "upgrade", "--dry-run")
    Assert-True ($output -eq "would update .vloop/README.md`nalready at $version") "dry-run output was:`n$output"
    Assert-True ((Get-Content -Raw $readmePath) -ceq $edited) "dry-run changed the edited README"

    $output = Invoke-Cli $binary @("-C", $repo, "upgrade")
    Assert-True ($output -eq "updated .vloop/README.md`nalready at $version") "edited same-version README output was:`n$output"
    Assert-True (-not (Get-Content -Raw $readmePath).Contains("local edit")) "upgrade kept the local edit"

    $stampPath = Join-Path $repo ".vloop/install.json"
    $stamp = Get-Content -Raw $stampPath | ConvertFrom-Json
    $stamp.version = "0.14.0"
    $stampJson = $stamp | ConvertTo-Json -Compress
    Set-Content -Path $stampPath -Value $stampJson -NoNewline
    Add-Content -Path $readmePath -Value "break me" -NoNewline
    $beforeBreaking = Get-Content -Raw $readmePath
    $stampBeforeBreaking = Get-Content -Raw $stampPath
    $breaking = & $binary -C $repo upgrade 2>&1 | Out-String
    $breakingCode = $LASTEXITCODE
    $breaking = ($breaking -replace "`r`n", "`n").TrimEnd("`n")
    Assert-True ($breakingCode -eq 1) "breaking upgrade did not refuse: $breaking"
    Assert-True $breaking.Contains("would update .vloop/README.md") "breaking upgrade omitted the README change"
    Assert-True ((Get-Content -Raw $readmePath) -ceq $beforeBreaking) "refused breaking upgrade changed the README"
    Assert-True ((Get-Content -Raw $stampPath) -ceq $stampBeforeBreaking) "refused breaking upgrade changed the install stamp"

    $output = Invoke-Cli $binary @("-C", $repo, "upgrade", "--yes")
    Assert-True $output.Contains("updated .vloop/README.md") "approved breaking upgrade did not update the README"
    Assert-True (-not (Get-Content -Raw $readmePath).Contains("break me")) "approved upgrade kept the edited README"
}
finally {
    if (Test-Path $temp) { Remove-Item -Recurse -Force $temp }
}
