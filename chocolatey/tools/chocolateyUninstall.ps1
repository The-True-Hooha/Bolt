$ErrorActionPreference = 'Stop'

$toolsDir = "$(Split-Path -parent $MyInvocation.MyCommand.Definition)"
$boltExe  = Join-Path $toolsDir 'bolt.exe'

if (Test-Path $boltExe) {
  Remove-Item -Path $boltExe -Force
  Write-Host "bolt.exe removed from $toolsDir"
}
