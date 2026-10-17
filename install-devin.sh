#!/data/data/com.termux/files/usr/bin/bash

set -e

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo "=========================================="
echo "   Installing Devin CLI on Android/Termux  "
echo "=========================================="
echo ""

# Check if running in Termux
if [ ! -d "/data/data/com.termux" ]; then
    echo -e "${RED}Error: This script must be run in Termux${NC}"
    exit 1
fi

echo -e "${GREEN}[1/4] Installing dependencies...${NC}"
pkg install -y curl tar

echo -e "${GREEN}[2/4] Downloading Devin CLI for ARM64...${NC}"
# Devin CLI ARM64 Linux binary for Android
DEVIN_VERSION="latest"
DOWNLOAD_URL="https://static.devin.ai/cli/aarch64-unknown-linux/devin"

# Create bin directory
mkdir -p ~/.local/bin

# Download binary
echo "Downloading from: $DOWNLOAD_URL"
curl -fsSL "$DOWNLOAD_URL" -o ~/.local/bin/devin

if [ ! -f ~/.local/bin/devin ]; then
    echo -e "${RED}Error: Failed to download Devin CLI${NC}"
    exit 1
fi

echo -e "${GREEN}[3/4] Making binary executable...${NC}"
chmod +x ~/.local/bin/devin

echo -e "${GREEN}[4/4] Adding to PATH...${NC}"
# Add to PATH if not already there
if ! grep -q "~/.local/bin" ~/.bashrc; then
    echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc
fi

# Export for current session
export PATH="$HOME/.local/bin:$PATH"

echo ""
echo -e "${GREEN}=========================================="
echo -e "   Devin CLI Installed! 🎉"
echo -e "==========================================${NC}"
echo ""
echo -e "${YELLOW}Next steps:${NC}"
echo "1. Restart Termux or run: export PATH=\"\$HOME/.local/bin:\$PATH\""
echo "2. Login to Devin: devin auth login"
echo "3. Run Devin in a project: cd your-project && devin"
echo ""
echo -e "${YELLOW}Verify installation:${NC}"
echo "  devin --version"
echo ""
