#!/bin/bash
set -e

# ========================
# CONFIGURABLE VARIABLES
# ========================
VERSION="v2.0.6"

# ========================
# COLORS FOR UI
# ========================
GREEN="\033[1;32m"
YELLOW="\033[1;33m"
RED="\033[1;31m"
BLUE="\033[1;34m"
RESET="\033[0m"

# ========================
# DETECT OS AND ARCH
# ========================
echo -e "${BLUE}Detecting system...${RESET}"
OS="$(uname -s)"
ARCH="$(uname -m)"
URL=""

case "$OS" in
    Linux)
        case "$ARCH" in
            x86_64)
                URL="https://github.com/pytgcalls/ntgcalls/releases/download/$VERSION/ntgcalls.linux-x86_64-static_libs.zip"
                ;;
            aarch64|arm64)
                URL="https://github.com/pytgcalls/ntgcalls/releases/download/$VERSION/ntgcalls.linux-arm64-static_libs.zip"
                ;;
            *)
                echo -e "${RED}Unsupported architecture: $ARCH${RESET}"
                exit 1
                ;;
        esac
        ;;
    Darwin)
        if [[ "$ARCH" == "arm64" ]]; then
            URL="https://github.com/pytgcalls/ntgcalls/releases/download/$VERSION/ntgcalls.macos-arm64-static_libs.zip"
        else
            echo -e "${RED}Unsupported architecture: $ARCH${RESET}"
            exit 1
        fi
        ;;
    MINGW*|MSYS*|CYGWIN*)
        URL="https://github.com/pytgcalls/ntgcalls/releases/download/$VERSION/ntgcalls.windows-x86_64-static_libs.zip"
        ;;
    *)
        echo -e "${RED}Unsupported OS: $OS${RESET}"
        exit 1
        ;;
esac

echo -e "${GREEN}System detected: $OS $ARCH${RESET}"
echo -e "${GREEN}Download URL: $URL${RESET}"

# ========================
# DOWNLOAD AND EXTRACT
# ========================
echo -e "${BLUE}Downloading ntgcalls...${RESET}"
curl -sSL -o ntgcalls.zip "$URL"
echo -e "${GREEN}Download complete: ntgcalls.zip${RESET}"

echo -e "${BLUE}Extracting files...${RESET}"
mkdir -p tmp
unzip -q ntgcalls.zip -d tmp
echo -e "${GREEN}Extraction complete.${RESET}"

# ========================
# COPY HEADER
# ========================
echo -e "${BLUE}Copying ntgcalls.h...${RESET}"
cp "tmp/include/ntgcalls.h" "ntgcalls/"
echo -e "${GREEN}Header copied to ntgcalls/${RESET}"

# ========================
# MOVE LIBRARY FILE
# ========================
echo -e "${BLUE}Moving library file...${RESET}"
LIB_FILE=$(find "tmp/lib" -type f | head -n1)
LIB_NAME=$(basename "$LIB_FILE")
mv "$LIB_FILE" "./$LIB_NAME"
echo -e "${GREEN}Library moved: $LIB_NAME${RESET}"

# ========================
# CLEANUP
# ========================
echo -e "${BLUE}Cleaning up temporary files...${RESET}"
rm -rf ntgcalls.zip tmp
echo -e "${GREEN}Cleanup done.${RESET}"

echo -e "${YELLOW}All done! ntgcalls is ready to use.${RESET}"
