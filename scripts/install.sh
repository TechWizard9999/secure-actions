#!/bin/sh
set -e

REPO="Kota-Karthik/secure-actions"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$ARCH" in
  x86_64) ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  *) echo "Unsupported architecture: $ARCH" >&2; exit 1 ;;
esac

case "$OS" in
  linux|darwin) ;;
  *) echo "Unsupported OS: $OS" >&2; exit 1 ;;
esac

# Get latest release tag
if [ -z "$VERSION" ]; then
  VERSION=$(curl -sL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | cut -d'"' -f4)
  if [ -z "$VERSION" ]; then
    echo "Failed to determine latest version" >&2
    exit 1
  fi
fi

VERSION_NUM="${VERSION#v}"
ARCHIVE="secure-actions_${VERSION_NUM}_${OS}_${ARCH}.tar.gz"
URL="https://github.com/${REPO}/releases/download/${VERSION}/${ARCHIVE}"

echo "Installing secure-actions ${VERSION} (${OS}/${ARCH})..."

TMPDIR=$(mktemp -d)
trap 'rm -rf "$TMPDIR"' EXIT

curl -sL "$URL" -o "${TMPDIR}/${ARCHIVE}"
tar -xzf "${TMPDIR}/${ARCHIVE}" -C "$TMPDIR"

if [ -w "$INSTALL_DIR" ]; then
  mv "${TMPDIR}/secure-actions" "${INSTALL_DIR}/secure-actions"
else
  sudo mv "${TMPDIR}/secure-actions" "${INSTALL_DIR}/secure-actions"
fi

chmod +x "${INSTALL_DIR}/secure-actions"

echo "Installed secure-actions to ${INSTALL_DIR}/secure-actions"
echo ""
echo "Next steps:"
echo "  1. Start MongoDB:"
echo "     docker run -d --name secure-actions-mongo \\"
echo "       -p 27018:27017 \\"
echo "       -v ~/.secure-actions/mongo:/data/db \\"
echo "       mongo:8"
echo ""
echo "  2. Register the MCP server in your Claude Code config:"
echo "     \"secure-actions\": {"
echo "       \"type\": \"stdio\","
echo "       \"command\": \"${INSTALL_DIR}/secure-actions\","
echo "       \"env\": { \"MONGO_URI\": \"mongodb://localhost:27018\" }"
echo "     }"
echo ""
echo "A master encryption key will be auto-generated at ~/.secure-actions/master.key on first run."
