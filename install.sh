#!/usr/bin/env bash
# roost installer.
#   curl -fsSL https://raw.githubusercontent.com/CemOlmez/roost/main/install.sh | bash
# or, from a clone:  ./install.sh
set -euo pipefail

REPO="${ROOST_REPO:-https://github.com/CemOlmez/roost.git}"
DEST="${ROOST_DIR:-$HOME/.local/share/roost}"
BIN="$HOME/.local/bin"

here="$(cd "$(dirname "${BASH_SOURCE[0]:-.}")" 2>/dev/null && pwd || true)"
if [[ -n "$here" && -x "$here/bin/roost" ]]; then
  DEST="$here"                       # running from a clone: use it in place
elif [[ -d "$DEST/.git" ]]; then
  git -C "$DEST" pull --ff-only -q
else
  command -v git >/dev/null || { sudo apt-get update -qq && sudo apt-get install -y -qq git; }
  git clone -q "$REPO" "$DEST"
fi

mkdir -p "$BIN"
ln -sf "$DEST/bin/roost" "$BIN/roost"

if [[ ":$PATH:" != *":$BIN:"* ]]; then
  echo 'export PATH="$HOME/.local/bin:$PATH"' >>"$HOME/.bashrc"
  export PATH="$BIN:$PATH"
fi

echo "roost installed → $BIN/roost"
echo
exec "$BIN/roost" init </dev/tty
