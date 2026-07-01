#!/usr/bin/env bash
set -euo pipefail

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$root"

mkdir -p dist
GOOS=windows GOARCH=amd64 go build -trimpath -ldflags="-s -w -H windowsgui" -o dist/MusicVisualizerPro.exe ./cmd/music-visualizer

(
  cd dist
  rm -f MusicVisualizerPro-Windows.zip
  zip -9 -q MusicVisualizerPro-Windows.zip MusicVisualizerPro.exe
)

echo "Built dist/MusicVisualizerPro.exe"
echo "Packaged dist/MusicVisualizerPro-Windows.zip"
