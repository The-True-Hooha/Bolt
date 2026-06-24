#Requires -Version 5.1
<#
.SYNOPSIS
    Installs Bolt — blazingly fast terminal file manager.
.DESCRIPTION
    Downloads the latest prebuilt release from GitHub by default.
    Pass -FromSource to build from source instead (requires Go).
.PARAMETER InstallDir
    Target installation directory. Defaults to %APPDATA%\bolt\bin
.PARAMETER FromSource
    Build from source instead of downloading a prebuilt binary.
.EXAMPLE
    irm https://raw.githubusercontent.com/The-True-Hooha/Bolt/master/install.ps1 | iex
.EXAMPLE
    .\install.ps1 -FromSource
#>
param(
    [string]$InstallDir = (Join-Path $env:APPDATA "bolt\bin"),
    [switch]$FromSource
)

$ErrorActionPreference = "Stop"
$repo    = "https://github.com/The-True-Hooha/Bolt"
$apiBase = "https://api.github.com/repos/The-True-Hooha/Bolt"

function Write-Ok   { param($msg) Write-Host "  [OK] $msg" -ForegroundColor Green  }
function Write-Info { param($msg) Write-Host "  --> $msg"  -ForegroundColor Cyan   }
function Write-Fail { param($msg) Write-Host "  [X] $msg"  -ForegroundColor Red; exit 1 }

Write-Host ""
Write-Host "  ⚡ Bolt installer" -ForegroundColor Yellow
Write-Host ""

New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
$dest = Join-Path $InstallDir "bolt.exe"

if ($FromSource) {
    if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
        Write-Fail "Go not found. Install from https://go.dev/dl/ or omit -FromSource to download prebuilt."
    }
    Write-Info "Go $(go version) found"

    $buildDir = if ($PSScriptRoot) { $PSScriptRoot } else { $PWD.Path }
    if (-not (Test-Path (Join-Path $buildDir "go.mod"))) {
        $tmp = Join-Path $env:TEMP "bolt-src-$(Get-Random)"
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

    Copy-Item (Join-Path $buildDir "bolt.exe") $dest -Force
} else {
    Write-Info "fetching latest release info..."
    try {
        $release = Invoke-RestMethod "$apiBase/releases/latest"
    } catch {
        Write-Fail "could not reach GitHub API: $_"
    }

    $tag    = $release.tag_name
    $asset  = $release.assets | Where-Object { $_.name -like "*windows_amd64*" } | Select-Object -First 1
    if (-not $asset) {
        Write-Fail "no windows_amd64 asset in release $tag — try -FromSource"
    }

    $url = $asset.browser_download_url
    Write-Info "downloading $($asset.name) ($tag)"
    Invoke-WebRequest -Uri $url -OutFile $dest -UseBasicParsing
    Write-Ok "downloaded"
}

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
