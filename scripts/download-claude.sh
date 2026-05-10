#!/usr/bin/env bash
# Downloads the claude binary from npm into ~/.angelix/bin/claude.
# Usage: ./scripts/download-claude.sh [version]
# If version is omitted, the latest published version is used.
set -euo pipefail

DEST_DIR="$HOME/.angelix/bin"
DEST="$DEST_DIR/claude"

# Detect platform and architecture.
OS="$(uname -s)"
ARCH="$(uname -m)"

case "$OS" in
  Darwin)
    case "$ARCH" in
      arm64)  PKG_SUFFIX="darwin-arm64" ;;
      x86_64) PKG_SUFFIX="darwin-x64"   ;;
      *)      echo "Unsupported macOS architecture: $ARCH" >&2; exit 1 ;;
    esac
    ;;
  Linux)
    # Detect musl (Alpine) vs glibc.
    if ldd /bin/sh 2>&1 | grep -q musl; then
      MUSL_SUFFIX="-musl"
    else
      MUSL_SUFFIX=""
    fi
    case "$ARCH" in
      aarch64|arm64) PKG_SUFFIX="linux-arm64${MUSL_SUFFIX}" ;;
      x86_64)        PKG_SUFFIX="linux-x64${MUSL_SUFFIX}"   ;;
      *)             echo "Unsupported Linux architecture: $ARCH" >&2; exit 1 ;;
    esac
    ;;
  *)
    echo "Unsupported OS: $OS" >&2
    exit 1
    ;;
esac

PKG_NAME="@anthropic-ai/claude-code-${PKG_SUFFIX}"

# Resolve version: use argument if provided, otherwise fetch latest from registry.
if [ $# -ge 1 ] && [ -n "$1" ]; then
  VERSION="$1"
else
  echo "Fetching latest claude-code version..."
  VERSION="$(curl -fsSL "https://registry.npmjs.org/@anthropic-ai/claude-code/latest" | grep -o '"version":"[^"]*"' | head -1 | cut -d'"' -f4)"
  if [ -z "$VERSION" ]; then
    echo "Failed to fetch latest version from npm registry." >&2
    exit 1
  fi
fi

echo "Downloading claude ${VERSION} for ${PKG_SUFFIX}..."

TARBALL_URL="https://registry.npmjs.org/${PKG_NAME}/-/$(echo "$PKG_NAME" | sed 's|@anthropic-ai/||')-${VERSION}.tgz"

TMPDIR="$(mktemp -d)"
trap 'rm -rf "$TMPDIR"' EXIT

curl -fsSL "$TARBALL_URL" -o "$TMPDIR/claude.tgz"
tar -xzf "$TMPDIR/claude.tgz" -C "$TMPDIR" package/claude

mkdir -p "$DEST_DIR"
cp "$TMPDIR/package/claude" "$DEST"
chmod 755 "$DEST"

echo "claude ${VERSION} installed to ${DEST}"
