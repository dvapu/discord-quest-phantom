<#
.SYNOPSIS
    Automated Multi-Platform Build & Packaging Script for Discord Quest Completer
#>
$ErrorActionPreference = "Stop"

$VERSION = "1.0.0"
$PROJECT_ROOT = Resolve-Path (Join-Path $PSScriptRoot "..")
$DIST_DIR = Join-Path $PROJECT_ROOT "dist"
$CMD_PATH = "./cmd/completer"

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
Write-Host " Building Discord Quest Completer v$VERSION" -ForegroundColor Cyan
Write-Host " Using Go compiler: $GO_BIN" -ForegroundColor Cyan
Write-Host "==========================================================" -ForegroundColor Cyan

$WIN_STAGING = Join-Path $DIST_DIR "windows/discord-quest-completer-windows-amd64"
$LINUX_STAGING = Join-Path $DIST_DIR "linux/discord-quest-completer-linux-amd64"

New-Item -ItemType Directory -Path $WIN_STAGING -Force | Out-Null
New-Item -ItemType Directory -Path $LINUX_STAGING -Force | Out-Null

# 1. Build Windows Binary
Write-Host "[1/4] Compiling Windows x64 standalone executable..." -ForegroundColor Yellow
$env:GOOS = "windows"
$env:GOARCH = "amd64"
$env:CGO_ENABLED = "0"
& $GO_BIN build -trimpath -ldflags="-s -w -X main.Version=$VERSION" `
    -o (Join-Path $WIN_STAGING "discord-quest-completer.exe") $CMD_PATH

# 2. Build Linux Binary
Write-Host "[2/4] Cross-compiling Linux x64 static executable..." -ForegroundColor Yellow
$env:GOOS = "linux"
$env:GOARCH = "amd64"
$env:CGO_ENABLED = "0"
& $GO_BIN build -trimpath -ldflags="-s -w -X main.Version=$VERSION" `
    -o (Join-Path $LINUX_STAGING "discord-quest-completer") $CMD_PATH

# 3. Copy supporting assets
Write-Host "[3/4] Copying configurations and documentation..." -ForegroundColor Yellow
$SUPPORT_FILES = @("config.example.json", ".token.example", "LICENSE", "README.txt")
foreach ($file in $SUPPORT_FILES) {
    $src = Join-Path $PROJECT_ROOT $file
    if (Test-Path $src) {
        Copy-Item $src $WIN_STAGING -Force
        Copy-Item $src $LINUX_STAGING -Force
    }
}

# 4. Create Standalone Archives
Write-Host "[4/4] Creating distribution archives in dist/..." -ForegroundColor Yellow
$WIN_ZIP = Join-Path $DIST_DIR "discord-quest-completer-windows-amd64.zip"
$LINUX_TAR = Join-Path $DIST_DIR "discord-quest-completer-linux-amd64.tar.gz"
$LINUX_ZIP = Join-Path $DIST_DIR "discord-quest-completer-linux-amd64.zip"

# Remove existing zip files if present
if (Test-Path $WIN_ZIP) { Remove-Item $WIN_ZIP -Force }
if (Test-Path $LINUX_TAR) { Remove-Item $LINUX_TAR -Force }
if (Test-Path $LINUX_ZIP) { Remove-Item $LINUX_ZIP -Force }

# Create Primary Standalone ZIP Archives (Extract and Run, No Installers)
if (Get-Command 7z -ErrorAction SilentlyContinue) {
    7z a -tzip $WIN_ZIP "$WIN_STAGING\*" | Out-Null
    7z a -tzip $LINUX_ZIP "$LINUX_STAGING\*" | Out-Null
} else {
    Compress-Archive -Path "$WIN_STAGING\*" -DestinationPath $WIN_ZIP -Force
    Compress-Archive -Path "$LINUX_STAGING\*" -DestinationPath $LINUX_ZIP -Force
}

# Create Supplementary POSIX Tarball (.tar.gz) without temporary filename artifacts
if (Get-Command tar -ErrorAction SilentlyContinue) {
    # Preferred: Use native tar (bsdtar) to generate clean .tar.gz in a single step
    & tar -czf $LINUX_TAR -C $LINUX_STAGING .
} elseif (Get-Command 7z -ErrorAction SilentlyContinue) {
    # 7z fallback: Use canonical archive name so gzip header contains correct filename
    $LINUX_TAR_INTERMEDIATE = Join-Path $DIST_DIR "discord-quest-completer-linux-amd64.tar"
    7z a -ttar $LINUX_TAR_INTERMEDIATE "$LINUX_STAGING\*" | Out-Null
    7z a -tgzip $LINUX_TAR $LINUX_TAR_INTERMEDIATE | Out-Null
    if (Test-Path $LINUX_TAR_INTERMEDIATE) { Remove-Item $LINUX_TAR_INTERMEDIATE -Force }
}

Write-Host "==========================================================" -ForegroundColor Green
Write-Host " Build Complete! Artifacts generated in dist/:" -ForegroundColor Green
Get-ChildItem $DIST_DIR | Where-Object { $_.Name -match '\.(zip|tar\.gz)$' } | ForEach-Object {
    Write-Host "  - $($_.Name) ($([math]::Round($_.Length / 1MB, 2)) MB)" -ForegroundColor White
}
Write-Host " Note for Linux users: Run 'chmod +x discord-quest-completer' before first use." -ForegroundColor Yellow
Write-Host "==========================================================" -ForegroundColor Green
