#!/usr/bin/env bash
set -euo pipefail

# Push the Muse export commit to github.com/armanster111/muse.
# Create an empty "muse" repository on GitHub first, then run this script.

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
MUSE_COMMIT="${MUSE_COMMIT:-80fa0b1}"
REMOTE_URL="${MUSE_REMOTE:-https://github.com/armanster111/muse.git}"

cd "$ROOT"

if ! git cat-file -e "${MUSE_COMMIT}^{commit}" 2>/dev/null; then
  echo "Muse export commit ${MUSE_COMMIT} not found in this clone."
  exit 1
fi

echo "Pushing Muse project (${MUSE_COMMIT}) to ${REMOTE_URL} ..."
git push "${REMOTE_URL}" "${MUSE_COMMIT}:main"
echo "Done. Muse lives at ${REMOTE_URL%.git}"
