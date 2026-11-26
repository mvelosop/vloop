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

$temp = Join-Path ([System.IO.Path]::GetTempPath()) ("vloop-gate-t1-" + [guid]::NewGuid().ToString("N"))
try {
    New-Item -ItemType Directory -Path $temp | Out-Null
    $binary = Join-Path $temp "vloop.exe"
    & go build -o $binary ./cmd/vloop
    if ($LASTEXITCODE -ne 0) { throw "build failed" }

    $versionLine = Invoke-Cli $binary @("version")
    $versionMatch = [regex]::Match($versionLine, "^vloop (\S+)")
    Assert-True $versionMatch.Success "vloop version did not print its version: $versionLine"
    $version = $versionMatch.Groups[1].Value

    $english = Join-Path $temp "english"
    New-Item -ItemType Directory -Path $english | Out-Null
    & git -C $english init -q
    if ($LASTEXITCODE -ne 0) { throw "git init failed" }
    $readmePath = Join-Path $english ".vloop/README.md"
    New-Item -ItemType Directory -Path (Split-Path $readmePath) | Out-Null
    Set-Content -Path $readmePath -Value "local text" -NoNewline

    $output = Invoke-Cli $binary @("-C", $english, "init", "--language", "en")
    $lines = $output -split "`n"
    $stampIndex = [array]::IndexOf($lines, "wrote .vloop/install.json")
    $guidelinesIndex = [array]::IndexOf($lines, "wrote .vloop/repo-guidelines.md")
    $readmeIndex = [array]::IndexOf($lines, "wrote .vloop/README.md")
    $briefIndex = [array]::FindIndex($lines, [Predicate[string]] { param($line) $line -like "wrote docs/briefs/*loop-brief.md" })
    Assert-True (($stampIndex -ge 0) -and ($guidelinesIndex -gt $stampIndex) -and ($readmeIndex -gt $guidelinesIndex) -and ($briefIndex -gt $readmeIndex)) "README output is not between the install files and starter brief:`n$output"

    $readme = Get-Content -Raw $readmePath
    Assert-True (-not $readme.Contains("local text")) "init kept or merged an existing .vloop/README.md"
    foreach ($heading in @("# vloop", "## What vloop is", "## How to use it", "## What is in this folder", "## More")) {
        Assert-True $readme.Contains($heading) "English README lacks $heading"
    }
    $ownership = "Written by vloop $version. vloop owns this file: ``vloop upgrade`` rewrites it, so edits here are lost."
    Assert-True $readme.Contains($ownership) "English README lacks the versioned ownership line"
    foreach ($text in @("command-line tool", "Claude Code", "GitHub Copilot CLI", "gate", "independent review", "one commit per task", "macOS", "Linux", "Windows", "go install", "github.com/mvelosop/vloop/cmd/vloop@latest", "vloop doctor", "/vloop:create-brief", "vloop run <brief>", "vloop self-update", "vloop upgrade", "config.toml", "config.local.toml", "install.json", "repo-guidelines.md", "README.md", "defects/", "interventions/", "state/", "tmp/", "https://github.com/mvelosop/vloop")) {
        Assert-True $readme.Contains($text) "English README lacks $text"
    }
    $wordCount = @($readme -split "\s+" | Where-Object { $_.Length -gt 0 }).Count
    Assert-True ($wordCount -le 400) "README has $wordCount words"
    Assert-True (-not $readme.Contains("`r")) "README does not use LF line endings"

    $existingGuidelines = Join-Path $temp "existing-guidelines"
    New-Item -ItemType Directory -Path $existingGuidelines | Out-Null
    & git -C $existingGuidelines init -q
    New-Item -ItemType Directory -Path (Join-Path $existingGuidelines ".vloop") | Out-Null
    Set-Content -Path (Join-Path $existingGuidelines ".vloop/repo-guidelines.md") -Value "# Mine`n" -NoNewline
    $keptOutput = Invoke-Cli $binary @("-C", $existingGuidelines, "init")
    Assert-True (-not $keptOutput.Contains("repo-guidelines.md")) "init printed for existing guidelines"
    $keptLines = $keptOutput -split "`n"
    $keptStamp = [array]::IndexOf($keptLines, "wrote .vloop/install.json")
    $keptReadme = [array]::IndexOf($keptLines, "wrote .vloop/README.md")
    $keptBrief = [array]::FindIndex($keptLines, [Predicate[string]] { param($line) $line -like "wrote docs/briefs/*loop-brief.md" })
    Assert-True (($keptReadme -gt $keptStamp) -and ($keptBrief -gt $keptReadme)) "README output is misplaced when guidelines already exist"

    $spanish = Join-Path $temp "spanish"
    New-Item -ItemType Directory -Path $spanish | Out-Null
    & git -C $spanish init -q
    Invoke-Cli $binary @("-C", $spanish, "init", "--language", "es") | Out-Null
    $spanishReadme = Get-Content -Raw (Join-Path $spanish ".vloop/README.md")
    foreach ($heading in @("# vloop", "## Qué es vloop", "## Cómo se usa", "## Qué hay en esta carpeta", "## Más")) {
        Assert-True $spanishReadme.Contains($heading) "Spanish README lacks $heading"
    }
    $spanishOwnership = "Escrito por vloop $version. Este archivo es de vloop: ``vloop upgrade`` lo reescribe, así que lo que se edite aquí se pierde."
    Assert-True $spanishReadme.Contains($spanishOwnership) "Spanish README lacks the ownership line"
    Assert-True $spanishReadme.Contains("https://github.com/mvelosop/vloop") "Spanish README lacks the repository URL"
}
finally {
    if (Test-Path $temp) { Remove-Item -Recurse -Force $temp }
}
