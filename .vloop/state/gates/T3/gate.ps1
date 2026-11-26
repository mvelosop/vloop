$ErrorActionPreference = 'Stop'

$root = (Get-Location).Path
$temp = Join-Path ([System.IO.Path]::GetTempPath()) ("vloop-gate-t3-" + [guid]::NewGuid().ToString('N'))
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

  Assert-Fence powershell '(Get-Content README.md | Measure-Object -Line).Lines' 'allow'
  Assert-Fence powershell '(Get-Content a).Name | Measure-Object' 'allow'
  Assert-Fence powershell '(Get-Item a).Name.Value2 | Select-Object -First 1' 'allow'
  Assert-Fence powershell '(git push).Length' 'deny: git push is denied by the work fence (Bash(git push:*))'
  Assert-Fence powershell '(Get-Content a).Trim()' 'deny: parentheses cannot be judged here — (<command>).<Property> is read; or pipe into Measure-Object or Select-Object -ExpandProperty'
  Assert-Fence powershell 'Write-Output (Get-Content a)' 'deny: parentheses cannot be judged here — (<command>).<Property> is read; or pipe into Measure-Object or Select-Object -ExpandProperty'
  Assert-Fence powershell '@(git status)' 'deny: parentheses cannot be judged here — (<command>).<Property> is read; or pipe into Measure-Object or Select-Object -ExpandProperty'

  Assert-Fence powershell '$env:GOOS=''linux''; go build ./...' 'allow'
  Assert-Fence powershell '$env:GOOS = "linux"; go vet ./...' 'allow'
  Assert-Fence powershell '$env:_NAME=linux; go test ./x' 'allow'
  Assert-Fence powershell '$env:GOOS=$x; go build ./...' 'deny: command word $env:GOOS=$x starts with $ and cannot be judged'
  Assert-Fence powershell '$env:1GOOS=linux; go test ./x' 'deny: command word $env:1GOOS=linux starts with $ and cannot be judged'
  Assert-Fence powershell '$env:GOOS=(linux); go test ./x' 'deny: command word $env:GOOS=(linux) starts with $ and cannot be judged'

  Assert-Fence powershell 'Write-Output "$(git push)"' 'deny: command substitution cannot be judged — run the inner command on its own'
  Assert-Fence powershell 'Write-Output "a`tb"' 'deny: a backtick cannot be judged — write the command on one line, without escapes'
  Assert-Fence bash '(Get-Content README.md).Count' 'deny: subshells and process substitution cannot be judged'
  Assert-Fence bash '$env:GOOS=linux; go build ./...' 'deny: command word $env:GOOS=linux starts with $ and cannot be judged'
  Assert-Fence bash 'Write-Output "$(git push)"' 'deny: command substitution cannot be judged'
  Assert-Fence bash 'Write-Output "a`tb"' 'deny: command substitution cannot be judged'
}
finally {
  Remove-Item -LiteralPath $temp -Recurse -Force
}
