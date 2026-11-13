#!/bin/sh
# roost installer — downloads the latest release for this machine, verifies
# its checksum, installs it to ~/.local/bin, then runs `roost init`.
#
#   curl -fsSL https://raw.githubusercontent.com/CemOlmez/roost/main/install.sh | sh
#
# Environment:
#   ROOST_VERSION   install a specific tag (e.g. v0.1.0) instead of the latest
#   ROOST_BIN_DIR   install directory (default: ~/.local/bin)
#   ROOST_NO_INIT=1 install only; don't run `roost init`
#   ROOST_BASE_URL  download from here instead of GitHub Releases (testing)
set -eu

REPO="CemOlmez/roost"
BIN_DIR="${ROOST_BIN_DIR:-$HOME/.local/bin}"

say() { printf '%s\n' "$*"; }
die() { printf 'roost install: %s\n' "$*" >&2; exit 1; }

[ "$(uname -s)" = "Linux" ] || die "roost runs on Linux (on Windows, run this inside WSL Ubuntu)."

case "$(uname -m)" in
  x86_64 | amd64) arch=amd64 ;;
  aarch64 | arm64) arch=arm64 ;;
  armv7l | armv7*) arch=arm ;;
  *) die "unsupported CPU: $(uname -m)" ;;
esac

if [ -n "${ROOST_BASE_URL:-}" ]; then
  base="$ROOST_BASE_URL"
elif [ -n "${ROOST_VERSION:-}" ]; then
  base="https://github.com/$REPO/releases/download/$ROOST_VERSION"
else
  base="https://github.com/$REPO/releases/latest/download"
fi
archive="roost_linux_${arch}.tar.gz"

command -v curl >/dev/null 2>&1 || die "curl is required (sudo apt-get install curl)"
command -v sha256sum >/dev/null 2>&1 || die "sha256sum is required (coreutils)"

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT

say "Downloading roost for linux/$arch..."
curl -fsSL "$base/$archive" -o "$tmp/$archive" || die "download failed: $base/$archive"
curl -fsSL "$base/checksums.txt" -o "$tmp/checksums.txt" || die "could not download checksums.txt"

(cd "$tmp" && grep " $archive\$" checksums.txt | sha256sum -c --quiet -) \
  || die "checksum mismatch — refusing to install"

tar -xzf "$tmp/$archive" -C "$tmp" roost
mkdir -p "$BIN_DIR"
install -m 0755 "$tmp/roost" "$BIN_DIR/roost"
say "Installed $("$BIN_DIR/roost" --version) → $BIN_DIR/roost"

case ":$PATH:" in
  *":$BIN_DIR:"*) ;;
  *)
    # shellcheck disable=SC2016 # written literally into .bashrc on purpose
    printf '\nexport PATH="%s:$PATH"\n' "$BIN_DIR" >>"$HOME/.bashrc"
    say "Added $BIN_DIR to PATH in ~/.bashrc (open a new terminal, or: export PATH=\"$BIN_DIR:\$PATH\")"
    ;;
esac

if [ "${ROOST_NO_INIT:-0}" = 1 ]; then
  say "Next: roost init"
  exit 0
fi

say ""
# stdin is the curl pipe; give init the terminal so it can ask questions.
if [ -r /dev/tty ]; then
  exec "$BIN_DIR/roost" init </dev/tty
else
  say "No terminal available — run: roost init"
fi
