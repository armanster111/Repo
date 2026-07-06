#!/usr/bin/env bash
set -euo pipefail

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$root"

mkdir -p dist
GOOS=windows GOARCH=amd64 go build -trimpath -ldflags="-s -w -H windowsgui" -o dist/MusicVisualizerPro.exe ./cmd/music-visualizer

if [[ -n "${SIGN_CERT:-}" && -f "${SIGN_CERT}" ]]; then
  echo "Signing with SIGN_CERT..."
  if command -v osslsigncode >/dev/null 2>&1; then
    osslsigncode sign -pkcs12 "$SIGN_CERT" -pass "${SIGN_PASSWORD:-}" \
      -n "Music Visualizer Pro" -i "https://github.com/armanster111/Repo" \
      -in dist/MusicVisualizerPro.exe -out dist/MusicVisualizerPro-signed.exe
    mv dist/MusicVisualizerPro-signed.exe dist/MusicVisualizerPro.exe
  else
    echo "osslsigncode not found — skip Linux cross-sign; use sign-windows.bat on Windows."
  fi
fi

(
  cd dist
  rm -f MusicVisualizerPro-Windows.zip
  zip -9 -q MusicVisualizerPro-Windows.zip MusicVisualizerPro.exe
)

echo "Built dist/MusicVisualizerPro.exe"
echo "Packaged dist/MusicVisualizerPro-Windows.zip"
