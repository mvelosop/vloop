$ErrorActionPreference = 'Stop'
$text = Get-Content -Raw docs/guide/configuration.md
$section = [regex]::Match($text, '(?ms)^### Copilot model ids\s*(.*?)(?=^### |\z)')
if (!$section.Success) { throw 'Copilot model ids section is missing' }
$body = $section.Groups[1].Value
if ($body -notmatch 'vloop models copilot') { throw 'guide does not use vloop models copilot' }
if ($body -notmatch 'policy') { throw 'guide does not explain the policy column' }
if ($body -notmatch 'plan') { throw 'guide does not explain the plan line' }
if ($body -notmatch 'config set') { throw 'guide does not say config set checks ids' }
if ($body -notmatch 'doctor') { throw 'guide does not say doctor checks ids' }
if ($body -notmatch '--kinds') { throw 'guide has no --kinds example' }
if ($body -match 'api\.githubcopilot\.com') { throw 'guide still embeds the raw Copilot API in the section' }
if ($body -match '(?m)^\s*(curl|Invoke-RestMethod)\b') { throw 'guide still contains the manual API request' }
