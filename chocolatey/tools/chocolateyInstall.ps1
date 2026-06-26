$ErrorActionPreference = 'Stop'

$toolsDir   = "$(Split-Path -parent $MyInvocation.MyCommand.Definition)"
$version    = '0.1.5'
$packageName = 'bolt-fm'
$url64       = "https://github.com/The-True-Hooha/Bolt/releases/download/v$version/bolt-fm_${version}_windows_amd64.zip"

# update after each release
$checksum64  = 'bdf3696175e75357f8ae45dd579e4d9e6d49507a42a23468e429e78e75946f61'

$packageArgs = @{
  packageName   = $packageName
  unzipLocation = $toolsDir
  url64bit      = $url64
  checksum64    = $checksum64
  checksumType64 = 'sha256'
}

Install-ChocolateyZipPackage @packageArgs

# Rename extracted exe to bolt.exe for convenience
$extracted = Join-Path $toolsDir "bolt-fm_${version}_windows_amd64.exe"
$target    = Join-Path $toolsDir 'bolt-fm.exe'
if (Test-Path $extracted) {
  Copy-Item -Path $extracted -Destination $target -Force
}
