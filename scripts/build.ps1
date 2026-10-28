<#
.SYNOPSIS
    Automated Multi-Platform Build & Packaging Script for Discord Quest Phantom.

.DESCRIPTION
    Produces release-ready archives in dist/:
      - discord-quest-phantom-windows-amd64.zip
      - discord-quest-phantom-linux-amd64.tar.gz
      - discord-quest-phantom-linux-arm64.tar.gz

    Windows batch launchers are re-encoded to CRLF before zipping so cmd.exe
    never truncates their commands (the "window flashes and closes" bug).
#>
$ErrorActionPreference = "Stop"

$VERSION = "1.1.2"
$PROJECT_ROOT = Resolve-Path (Join-Path $PSScriptRoot "..")
$DIST_DIR = Join-Path $PROJECT_ROOT "dist"
$CMD_PATH = "./cmd/completer/"

# Locate Go binary
$GO_BIN = "go"
if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
    if (Test-Path "C:\Program Files\Go\bin\go.exe") {
        $GO_BIN = "C:\Program Files\Go\bin\go.exe"
    } else {
        throw "Go compiler not found in PATH or 'C:\Program Files\Go\bin\go.exe'"
    }
}

Write-Host "==========================================================" -ForegroundColor Cyan
Write-Host " Building Discord Quest Phantom v$VERSION" -ForegroundColor Cyan
Write-Host " Using Go compiler: $GO_BIN" -ForegroundColor Cyan
Write-Host "==========================================================" -ForegroundColor Cyan

$WIN_STAGING = Join-Path $DIST_DIR "windows/discord-quest-phantom-windows-amd64"
$LINUX_AMD64_STAGING = Join-Path $DIST_DIR "linux/discord-quest-phantom-linux-amd64"
$LINUX_ARM64_STAGING = Join-Path $DIST_DIR "linux/discord-quest-phantom-linux-arm64"

foreach ($dir in @($WIN_STAGING, $LINUX_AMD64_STAGING, $LINUX_ARM64_STAGING)) {
    New-Item -ItemType Directory -Path $dir -Force | Out-Null
}

$LDFLAGS = "-s -w -X main.Version=$VERSION"

# 1. Windows x64
Write-Host "[1/5] Compiling Windows x64 executable..." -ForegroundColor Yellow
$env:GOOS = "windows"; $env:GOARCH = "amd64"; $env:CGO_ENABLED = "0"
& $GO_BIN build -trimpath -ldflags="$LDFLAGS" -o (Join-Path $WIN_STAGING "discord-quest-phantom.exe") $CMD_PATH

# 2. Linux x64
Write-Host "[2/5] Cross-compiling Linux x64 static executable..." -ForegroundColor Yellow
$env:GOOS = "linux"; $env:GOARCH = "amd64"; $env:CGO_ENABLED = "0"
& $GO_BIN build -trimpath -ldflags="$LDFLAGS" -o (Join-Path $LINUX_AMD64_STAGING "discord-quest-phantom") $CMD_PATH

# 3. Linux arm64
Write-Host "[3/5] Cross-compiling Linux arm64 static executable..." -ForegroundColor Yellow
$env:GOOS = "linux"; $env:GOARCH = "arm64"; $env:CGO_ENABLED = "0"
& $GO_BIN build -trimpath -ldflags="$LDFLAGS" -o (Join-Path $LINUX_ARM64_STAGING "discord-quest-phantom") $CMD_PATH

# 4. Copy supporting assets
Write-Host "[4/5] Copying configurations and documentation..." -ForegroundColor Yellow

$COMMON_FILES = @("README.md", "LICENSE", "config.example.json", ".token.example")
$WINDOWS_ONLY_FILES = @("run.cmd", "setup_token.cmd")

foreach ($file in $COMMON_FILES) {
    $src = Join-Path $PROJECT_ROOT $file
    if (-not (Test-Path $src)) { continue }
    Copy-Item $src $WIN_STAGING -Force
    Copy-Item $src $LINUX_AMD64_STAGING -Force
    Copy-Item $src $LINUX_ARM64_STAGING -Force
}

foreach ($file in $WINDOWS_ONLY_FILES) {
    $src = Join-Path $PROJECT_ROOT $file
    if (Test-Path $src) { Copy-Item $src $WIN_STAGING -Force }
}

# Force CRLF on every staged batch file, then verify it byte-by-byte.
Get-ChildItem -Path $WIN_STAGING -Filter "*.cmd" | ForEach-Object {
    $text = [System.IO.File]::ReadAllText($_.FullName)
    $crlf = $text -replace "`r`n", "`n" -replace "`n", "`r`n"
    $bytes = (New-Object System.Text.UTF8Encoding($false)).GetBytes($crlf)
    [System.IO.File]::WriteAllBytes($_.FullName, $bytes)

    # A bare LF (not preceded by CR) breaks cmd.exe parsing.
    $raw = [System.IO.File]::ReadAllBytes($_.FullName)
    for ($i = 0; $i -lt $raw.Length; $i++) {
        if ($raw[$i] -eq 10 -and ($i -eq 0 -or $raw[$i - 1] -ne 13)) {
            throw "Bare LF found in $($_.Name) at byte offset $i - cmd.exe would corrupt this script."
        }
    }
    Write-Host "  CRLF verified: $($_.Name)" -ForegroundColor DarkGray
}

# 5. Create archives
Write-Host "[5/5] Creating distribution archives in dist/..." -ForegroundColor Yellow

$WIN_ZIP   = Join-Path $DIST_DIR "discord-quest-phantom-windows-amd64.zip"
$AMD64_TAR = Join-Path $DIST_DIR "discord-quest-phantom-linux-amd64.tar.gz"
$ARM64_TAR = Join-Path $DIST_DIR "discord-quest-phantom-linux-arm64.tar.gz"

foreach ($f in @($WIN_ZIP, $AMD64_TAR, $ARM64_TAR)) {
    if (Test-Path $f) { Remove-Item $f -Force }
}

if (Get-Command 7z -ErrorAction SilentlyContinue) {
    7z a -tzip $WIN_ZIP "$WIN_STAGING\*" | Out-Null
} else {
    Compress-Archive -Path "$WIN_STAGING\*" -DestinationPath $WIN_ZIP -Force
}

if (Get-Command tar -ErrorAction SilentlyContinue) {
    & tar -czf $AMD64_TAR -C $LINUX_AMD64_STAGING .
    & tar -czf $ARM64_TAR -C $LINUX_ARM64_STAGING .
} else {
    Write-Warning "tar not found - Linux tarballs were skipped."
}

Write-Host "==========================================================" -ForegroundColor Green
Write-Host " Build Complete! Artifacts generated in dist/:" -ForegroundColor Green
Get-ChildItem $DIST_DIR | Where-Object { $_.Name -match '^discord-quest-phantom-.*\.(zip|tar\.gz)$' } | ForEach-Object {
    Write-Host "  - $($_.Name) ($([math]::Round($_.Length / 1MB, 2)) MB)" -ForegroundColor White
}
Write-Host " Note for Linux users: run 'chmod +x discord-quest-phantom' before first use." -ForegroundColor Yellow
Write-Host "==========================================================" -ForegroundColor Green
