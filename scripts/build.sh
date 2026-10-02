#!/usr/bin/env bash
set -euo pipefail

VERSION="1.0.0"
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DIST_DIR="${PROJECT_ROOT}/dist"
CMD_PATH="./cmd/completer"

echo "=== Building Discord Quest Completer v${VERSION} ==="

WIN_STAGING="${DIST_DIR}/windows/discord-quest-completer-windows-amd64"
LINUX_STAGING="${DIST_DIR}/linux/discord-quest-completer-linux-amd64"

mkdir -p "${WIN_STAGING}"
mkdir -p "${LINUX_STAGING}"

echo "[1/4] Compiling Windows x64 standalone executable..."
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags="-s -w -X main.Version=${VERSION}" \
    -o "${WIN_STAGING}/discord-quest-completer.exe" "${CMD_PATH}"

echo "[2/4] Compiling Linux x64 static executable..."
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags="-s -w -X main.Version=${VERSION}" \
    -o "${LINUX_STAGING}/discord-quest-completer" "${CMD_PATH}"
chmod +x "${LINUX_STAGING}/discord-quest-completer"

echo "[3/4] Copying configurations and documentation..."
for f in config.example.json .token.example LICENSE README.txt; do
    if [ -f "${PROJECT_ROOT}/${f}" ]; then
        cp "${PROJECT_ROOT}/${f}" "${WIN_STAGING}/"
        cp "${PROJECT_ROOT}/${f}" "${LINUX_STAGING}/"
    fi
done

echo "[4/4] Creating distribution archives in dist/..."
cd "${WIN_STAGING}"
zip -rq "${DIST_DIR}/discord-quest-completer-windows-amd64.zip" .

cd "${LINUX_STAGING}"
tar -czf "${DIST_DIR}/discord-quest-completer-linux-amd64.tar.gz" .
zip -rq "${DIST_DIR}/discord-quest-completer-linux-amd64.zip" .

echo "=== Build Complete! Artifacts in dist/ ==="
ls -lh "${DIST_DIR}"/*.zip "${DIST_DIR}"/*.tar.gz 2>/dev/null || true
