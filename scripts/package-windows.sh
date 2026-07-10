#!/usr/bin/env bash
set -euo pipefail

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$root"

mkdir -p dist
GOOS=windows GOARCH=amd64 go build -trimpath -ldflags="-s -w -H windowsgui" -o dist/Muse.exe ./cmd/muse

if [[ -n "${SIGN_CERT:-}" && -f "${SIGN_CERT}" ]]; then
  echo "Signing with SIGN_CERT..."
  if command -v osslsigncode >/dev/null 2>&1; then
    osslsigncode sign -pkcs12 "$SIGN_CERT" -pass "${SIGN_PASSWORD:-}" \
      -n "Muse" -i "https://github.com/armanster111/muse" \
      -in dist/Muse.exe -out dist/Muse-signed.exe
    mv dist/Muse-signed.exe dist/Muse.exe
  else
    echo "osslsigncode not found — skip Linux cross-sign; use sign-windows.bat on Windows."
  fi
fi

(
  cd dist
  rm -f Muse-Windows.zip
  zip -9 -q Muse-Windows.zip Muse.exe
)

echo "Built dist/Muse.exe"
echo "Packaged dist/Muse-Windows.zip"
