#!/bin/bash
set -e

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

echo_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 1. Detect OS and Architecture
OS="$(uname -s)"
ARCH="$(uname -m)"

case "$OS" in
    Linux)
        OS_TYPE="linux"
        ;;
    Darwin)
        OS_TYPE="darwin"
        ;;
    *)
        echo_error "Unsupported operating system: $OS"
        exit 1
        ;;
esac

case "$ARCH" in
    x86_64)
        ARCH_TYPE="amd64"
        ;;
    aarch64|arm64)
        ARCH_TYPE="arm64"
        ;;
    *)
        echo_error "Unsupported architecture: $ARCH"
        exit 1
        ;;
esac

BINARY_NAME="mihosh-${OS_TYPE}-${ARCH_TYPE}"
INSTALL_DIR="/usr/local/bin"
TARGET_NAME="mihosh"

echo_info "Detected System: $OS_TYPE $ARCH_TYPE"

# 2. Get Latest Version
echo_info "Checking latest version..."
LATEST_RELEASE_URL="https://api.github.com/repos/aimony/mihosh/releases/latest"
VERSION=$(curl -s $LATEST_RELEASE_URL | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

if [ -z "$VERSION" ]; then
    echo_error "Failed to fetch latest version info."
    exit 1
fi

echo_info "Latest version: $VERSION"

# 3. Construct Download URL
DOWNLOAD_URL="https://github.com/aimony/mihosh/releases/download/${VERSION}/${BINARY_NAME}"

# 4. Download and Install
TMP_FILE="/tmp/${BINARY_NAME}"

echo_info "Downloading ${BINARY_NAME} from ${DOWNLOAD_URL}..."
if curl -L -o "$TMP_FILE" --fail "$DOWNLOAD_URL"; then
    echo_info "Download successful."
else
    echo_error "Download failed. Please check your internet connection or if the asset exists for your architecture."
    exit 1
fi

chmod +x "$TMP_FILE"

echo_info "Installing to ${INSTALL_DIR}/${TARGET_NAME}..."
if [ -w "$INSTALL_DIR" ]; then
    mv "$TMP_FILE" "${INSTALL_DIR}/${TARGET_NAME}"
else
    echo_info "Sudo permission required to install to ${INSTALL_DIR}"
    sudo mv "$TMP_FILE" "${INSTALL_DIR}/${TARGET_NAME}"
fi

echo_info "Installation completed successfully!"
echo_info "Run 'mihosh' to start."
