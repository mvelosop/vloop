$ErrorActionPreference = 'Stop'

$temp = Join-Path ([System.IO.Path]::GetTempPath()) ("vloop-gate-t4-" + [guid]::NewGuid().ToString('N'))
New-Item -ItemType Directory -Path $temp | Out-Null
try {
  $binary = Join-Path $temp 'vloop.exe'
  & go build -o $binary ./cmd/vloop
  if ($LASTEXITCODE -ne 0) { throw 'failed to build vloop' }
  $help = (& $binary fence check --help 2>&1 | Out-String).Trim()
  if ($LASTEXITCODE -ne 0) { throw 'fence check --help failed' }

  function Assert-Text {
    param([string]$Text, [string]$Needle, [string]$Surface)
    $flat = ($Text -replace '\s+', ' ')
    if (-not $flat.Contains($Needle)) { throw "$Surface is missing: $Needle" }
  }

  Assert-Text $help 'pwsh and powershell -Command' 'vloop fence check --help'
  Assert-Text $help 'cmd /c' 'vloop fence check --help'
  Assert-Text $help 'pwsh''s eval-like commands and script blocks are refused' 'vloop fence check --help'

  $guide = Get-Content -Raw docs/guide/configuration.md
  Assert-Text $guide 'pwsh' 'docs/guide/configuration.md'
  Assert-Text $guide 'pwsh and powershell -Command' 'docs/guide/configuration.md'
  Assert-Text $guide 'cmd /c' 'docs/guide/configuration.md'
  Assert-Text $guide 'script blocks' 'docs/guide/configuration.md'
  Assert-Text $guide 'Where-Object <Property> -like <value>' 'docs/guide/configuration.md'
  Assert-Text $guide '(<command>).<Property>' 'docs/guide/configuration.md'
  Assert-Text $guide '$env:<NAME>' 'docs/guide/configuration.md'
  Assert-Text $guide 'one line, without escapes' 'docs/guide/configuration.md'

  $audit = Get-Content -Raw docs/design-notes/fence-audit.md
  Assert-Text $audit 'pwsh' 'docs/design-notes/fence-audit.md'
  Assert-Text $audit 'powershell -Command' 'docs/design-notes/fence-audit.md'
  Assert-Text $audit 'cmd /c' 'docs/design-notes/fence-audit.md'
  Assert-Text $audit 'Invoke-Expression' 'docs/design-notes/fence-audit.md'
  Assert-Text $audit 'script block' 'docs/design-notes/fence-audit.md'
  Assert-Text $audit 'driver' 'docs/design-notes/fence-audit.md'

  $commands = Get-Content -Raw docs/guide/commands.md
  Assert-Text $commands 'pwsh and powershell -Command' 'docs/guide/commands.md'
  Assert-Text $commands 'cmd /c' 'docs/guide/commands.md'
  Assert-Text $commands 'pwsh''s eval-like commands and script blocks are refused' 'docs/guide/commands.md'
}
finally {
  Remove-Item -LiteralPath $temp -Recurse -Force
}
