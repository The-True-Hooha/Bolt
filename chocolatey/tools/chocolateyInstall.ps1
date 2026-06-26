$ErrorActionPreference = 'Stop'

$toolsDir   = "$(Split-Path -parent $MyInvocation.MyCommand.Definition)"
$version    = '0.1.4'
$packageName = 'bolt-fm'
$url64       = "https://github.com/The-True-Hooha/Bolt/releases/download/v$version/bolt-fm_${version}_windows_amd64.zip"

# update after each release
$checksum64  = 'ecb6b4f80c4d99b52dd023a524c3d343b24174d91298a02609b3b5db41d8555f'

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
