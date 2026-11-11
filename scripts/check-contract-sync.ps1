param([string]$CoreRepo = "..\KINETOVELA")
$ErrorActionPreference = "Stop"
function Get-NormalizedHash([string]$Path) {
  $bytes = [System.IO.File]::ReadAllBytes($Path)
  $stream = New-Object System.IO.MemoryStream
  for ($i = 0; $i -lt $bytes.Length; $i++) {
    if ($bytes[$i] -eq 13 -and ($i + 1) -lt $bytes.Length -and $bytes[$i + 1] -eq 10) { continue }
    $stream.WriteByte($bytes[$i])
  }
  $sha = [System.Security.Cryptography.SHA256]::Create()
  try { $hash = $sha.ComputeHash($stream.ToArray()) }
  finally { $sha.Dispose(); $stream.Dispose() }
  return -join ($hash | ForEach-Object { $_.ToString("x2") })
}
$manifest = Get-Content (Join-Path $PSScriptRoot "..\contracts\sync-manifest.json") -Raw | ConvertFrom-Json
$failed = $false
foreach ($entry in $manifest.files) {
  $src = Join-Path $CoreRepo ("contracts/" + $entry.path)
  $dst = Join-Path $PSScriptRoot ("../contracts/" + $entry.path)
  if (-not (Test-Path $src)) { Write-Error "missing source: $src"; $failed = $true; continue }
  $a = Get-NormalizedHash $src
  $b = Get-NormalizedHash $dst
  if ($a -ne $b) { Write-Output "drift: $($entry.path)"; $failed = $true }
}
if ($failed) { exit 1 }
Write-Output "contract sync check passed"
