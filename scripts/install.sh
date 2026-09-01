#!/usr/bin/env bash
set -euo pipefail

REPO="SalvucciFacundo/novel-tui"
BINARY_NAME="novel-tui"
INSTALL_DIR="/usr/local/bin"

echo "==> Detecting platform and architecture..."
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"

case "$ARCH" in
  x86_64|amd64)
    ARCH="amd64"
    ;;
  aarch64|arm64)
    ARCH="arm64"
    ;;
  i386|i686)
    ARCH="386"
    ;;
  *)
    echo "Error: Unsupported architecture $ARCH" >&2
    exit 1
    ;;
esac

if [ "$OS" != "linux" ]; then
  echo "Error: This install script currently supports Linux." >&2
  exit 1
fi

echo "==> Fetching latest release for ${OS}_${ARCH}..."
LATEST_TAG=$(curl -sSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/' || true)

if [ -z "$LATEST_TAG" ]; then
  LATEST_TAG="v0.1.0"
fi

VERSION="${LATEST_TAG#v}"
TARBALL="${BINARY_NAME}_${VERSION}_${OS}_${ARCH}.tar.gz"
DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${LATEST_TAG}/${TARBALL}"

TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT

echo "==> Downloading ${DOWNLOAD_URL}..."
if curl -sSL -f -o "${TMP_DIR}/${TARBALL}" "$DOWNLOAD_URL"; then
  echo "==> Extracting binary..."
  tar -xzf "${TMP_DIR}/${TARBALL}" -C "$TMP_DIR"
  
  echo "==> Installing to ${INSTALL_DIR}..."
  if [ -w "$INSTALL_DIR" ]; then
    mv "${TMP_DIR}/${BINARY_NAME}" "${INSTALL_DIR}/${BINARY_NAME}"
  else
    sudo mv "${TMP_DIR}/${BINARY_NAME}" "${INSTALL_DIR}/${BINARY_NAME}"
  fi
  chmod +x "${INSTALL_DIR}/${BINARY_NAME}"
  echo "==> Successfully installed ${BINARY_NAME} (${LATEST_TAG}) to ${INSTALL_DIR}/${BINARY_NAME}!"
  echo "Run '${BINARY_NAME}' to start writing."
else
  echo "==> Binary release archive not found directly, falling back to source build via go install..."
  if command -v go >/dev/null 2>&1; then
    go install "github.com/${REPO}/cmd/${BINARY_NAME}@latest"
    echo "==> Successfully installed via 'go install' to $(go env GOPATH)/bin/${BINARY_NAME}"
  else
    echo "Error: Pre-built binary not found and Go is not installed on this system." >&2
    exit 1
  fi
fi
