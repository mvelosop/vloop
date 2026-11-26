$ErrorActionPreference = 'Stop'

$root = (Get-Location).Path
$temp = Join-Path ([System.IO.Path]::GetTempPath()) ("vloop-gate-t1-" + [guid]::NewGuid().ToString('N'))
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

  $push = 'deny: git push is denied by the work fence (Bash(git push:*))'
  Assert-Fence powershell "pwsh -c 'git push'" $push
  Assert-Fence bash "pwsh -NoProfile -Command `"git push origin main`"" 'deny: git push origin main is denied by the work fence (Bash(git push:*))'
  Assert-Fence powershell "powershell -nop -noni -nol -ep Bypass -Command 'go test ./x'" 'allow'
  Assert-Fence bash "cmd /c go test ./..." 'allow'
  Assert-Fence powershell 'pwsh -File x.ps1' 'deny: pwsh: runs a script, stdin or an encoded command, not a literal -Command string'
  Assert-Fence powershell 'powershell -EncodedCommand ZwBpAHQA' 'deny: powershell: runs a script, stdin or an encoded command, not a literal -Command string'
  Assert-Fence bash 'pwsh -Command -' 'deny: pwsh: runs a script, stdin or an encoded command, not a literal -Command string'
  Assert-Fence powershell 'pwsh x.ps1' 'deny: pwsh: runs a script, stdin or an encoded command, not a literal -Command string'
  Assert-Fence powershell 'pwsh' 'deny: pwsh: runs a script, stdin or an encoded command, not a literal -Command string'
  Assert-Fence powershell 'pwsh -c' 'deny: -Command has no command string'
  Assert-Fence powershell 'pwsh -c $x' 'deny: the -Command string expands and cannot be judged'
  Assert-Fence powershell 'cmd /c git push' $push
  Assert-Fence bash 'cmd /k git push' $push
  Assert-Fence powershell 'cmd /c git status && git push' $push
  Assert-Fence powershell 'cmd /c echo %PATH%' 'deny: cmd: % and ^ cannot be judged'
  Assert-Fence powershell 'cmd /c echo a^&git push' 'deny: cmd: % and ^ cannot be judged'
  Assert-Fence powershell 'cmd /d /q /s /c git status' 'allow'
  Assert-Fence powershell "pwsh -c 'echo x > .vloop/state/x'" 'deny: write to .vloop/state/x is not allowed in a work session: nothing under .git/ or .vloop/ but .vloop/tmp/'
  Assert-Fence powershell 'cmd' 'deny: cmd: runs stdin or a batch file, not a literal /c string'
  Assert-Fence powershell 'C:\Windows\System32\WindowsPowerShell\v1.0\powershell.exe -c ''git push''' $push

  $nested = 'git status'
  1..9 | ForEach-Object { $nested = "pwsh -c $nested" }
  Assert-Fence powershell $nested 'deny: shell wrappers nest too deep to judge'
}
finally {
  Remove-Item -LiteralPath $temp -Recurse -Force
}
