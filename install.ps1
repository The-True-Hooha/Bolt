#Requires -Version 5.1
<#
.SYNOPSIS
    Installs Bolt — blazingly fast terminal file manager.
.DESCRIPTION
    Builds Bolt from source and adds it to the current user's PATH
    without requiring administrator privileges.
.PARAMETER InstallDir
    Target installation directory. Defaults to %APPDATA%\bolt\bin
.EXAMPLE
    irm https://raw.githubusercontent.com/The-True-Hooha/Bolt/master/install.ps1 | iex
#>
param(
    [string]$InstallDir = (Join-Path $env:APPDATA "bolt\bin")
)

$ErrorActionPreference = "Stop"
$repo = "https://github.com/The-True-Hooha/Bolt"

function Write-Ok   { param($msg) Write-Host "  [OK] $msg" -ForegroundColor Green  }
function Write-Info { param($msg) Write-Host "  --> $msg"  -ForegroundColor Cyan   }
function Write-Fail { param($msg) Write-Host "  [X] $msg"  -ForegroundColor Red; exit 1 }

Write-Host ""
Write-Host "  ⚡ Bolt installer" -ForegroundColor Yellow
Write-Host ""

if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
    Write-Fail "Go is not installed. Download from https://go.dev/dl/"
}
$goVer = (go version)
Write-Info "$goVer found"

$buildDir = $PSScriptRoot
if (-not (Test-Path (Join-Path $buildDir "go.mod"))) {
    $tmp = Join-Path $env:TEMP "bolt-install-$(Get-Random)"
    Write-Info "cloning $repo"
    git clone --depth 1 $repo $tmp
    $buildDir = $tmp
} else {
    Write-Info "building from current directory"
}

Write-Info "building bolt..."
Push-Location $buildDir
go build -ldflags="-s -w" -o bolt.exe .
if ($LASTEXITCODE -ne 0) { Write-Fail "build failed" }
Pop-Location
Write-Ok "build complete"

New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
$dest = Join-Path $InstallDir "bolt.exe"
Copy-Item (Join-Path $buildDir "bolt.exe") $dest -Force
Write-Ok "installed to $dest"

$scope   = [System.EnvironmentVariableTarget]::User
$current = [Environment]::GetEnvironmentVariable("PATH", $scope)
$parts   = $current -split ";" | Where-Object { $_ -ne "" }

if ($parts -notcontains $InstallDir) {
    $newPath = ($parts + $InstallDir) -join ";"
    [Environment]::SetEnvironmentVariable("PATH", $newPath, $scope)
    Write-Ok "added $InstallDir to User PATH"
    Write-Host ""
    Write-Host "  Restart your terminal for PATH to take effect." -ForegroundColor Yellow
    Write-Host "  Or run in this session:" -ForegroundColor Yellow
    Write-Host "    `$env:PATH += `";$InstallDir`"" -ForegroundColor Gray
} else {
    Write-Ok "$InstallDir already in PATH"
}

Write-Host ""
Write-Ok "bolt installed! run: bolt --version"
Write-Host ""
