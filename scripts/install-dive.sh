#!/usr/bin/env bash
# Install a wagoodman/dive release binary into .dive-bin/dive.
set -euo pipefail

VERSION="${1:?usage: install-dive.sh <version e.g. 0.13.1>}"

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DEST="${ROOT}/.dive-bin"
mkdir -p "$DEST"

OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"
case "$ARCH" in
  x86_64) ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  *)
    echo "unsupported architecture: $ARCH" >&2
    exit 1
    ;;
esac

if [ "$OS" != "linux" ] && [ "$OS" != "darwin" ]; then
  echo "unsupported OS: $OS" >&2
  exit 1
fi

ARCHIVE="dive_${VERSION}_${OS}_${ARCH}.tar.gz"
URL="https://github.com/wagoodman/dive/releases/download/v${VERSION}/${ARCHIVE}"
TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT

echo "Downloading dive v${VERSION} (${OS}/${ARCH})..."
curl -fsSL "$URL" -o "${TMP}/${ARCHIVE}"
tar -xzf "${TMP}/${ARCHIVE}" -C "$TMP"

BIN="${TMP}/dive"
if [ ! -x "$BIN" ]; then
  echo "dive binary not found in archive" >&2
  exit 1
fi

install -m 0755 "$BIN" "${DEST}/dive"
"${DEST}/dive" --version
