$ErrorActionPreference = 'Stop'

$root = (Get-Location).Path
$temp = Join-Path ([System.IO.Path]::GetTempPath()) ("vloop-gate-t2-" + [guid]::NewGuid().ToString('N'))
New-Item -ItemType Directory -Path $temp | Out-Null
try {
  $binary = Join-Path $temp 'vloop.exe'
  & go build -o $binary ./cmd/vloop
  if ($LASTEXITCODE -ne 0) { throw 'failed to build vloop' }

  function Assert-Fence {
    param([string]$Tool, [string]$Command, [string]$Expected)
    $output = (& $binary fence check --phase work --root $root --tool $Tool $Command 2>&1 | Out-String).Trim()
    $code = $LASTEXITCODE
    $wantedCode = if ($Expected -eq 'allow') { 0 } else { 1 }
    if ($code -ne $wantedCode -or $output -ne $Expected) {
      throw "[$Tool] $Command`nexpected ($wantedCode) $Expected`nactual   ($code) $output"
    }
  }

  foreach ($name in @(
    'Invoke-Expression', 'iex', 'Invoke-Command', 'icm', 'Start-Process', 'saps', 'start',
    'Start-Job', 'sajb', 'Start-ThreadJob', 'Invoke-Item', 'ii'
  )) {
    Assert-Fence powershell "$name git push" "deny: $name is not allowed"
  }
  Assert-Fence powershell "Get-Content a | Invoke-Expression" 'deny: Invoke-Expression is not allowed'
  Assert-Fence bash "pwsh -c 'INVOKE-ITEM a'" 'deny: INVOKE-ITEM is not allowed'
  Assert-Fence powershell "C:\Windows\System32\WindowsPowerShell\v1.0\powershell.exe -c `"IEX 'git push'`"" 'deny: IEX is not allowed'

  $block = 'deny: script blocks cannot be judged — use Where-Object <Property> -like <value>, ForEach-Object -MemberName <Property> or Select-Object -ExpandProperty <Property>'
  Assert-Fence powershell 'Get-ChildItem | ForEach-Object { $_.Name }' $block
  Assert-Fence powershell '& {git push}' $block
  Assert-Fence powershell '@{ Name = "value" }' $block
  Assert-Fence powershell 'pwsh -c ''& {git push}''' $block
  Assert-Fence powershell "Get-ChildItem | ForEach-Object -MemberName Name" 'allow'
  Assert-Fence powershell "Get-ChildItem | Where-Object Name -like '*.go'" 'allow'
  Assert-Fence powershell "Get-ChildItem | Select-Object -ExpandProperty Name" 'allow'
  Assert-Fence powershell "Write-Output '{ literal }'" 'allow'
  Assert-Fence bash '{ git push; }' 'deny: git push is denied by the work fence (Bash(git push:*))'
}
finally {
  Remove-Item -LiteralPath $temp -Recurse -Force
}
