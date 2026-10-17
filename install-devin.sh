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

echo -e "${GREEN}[2/4] Downloading Devin CLI official installer...${NC}"
# Download the official install script to home directory (tmp might have issues on Android)
INSTALL_SCRIPT="$HOME/devin-install.sh"
curl -fsSL https://cli.devin.ai/install.sh -o "$INSTALL_SCRIPT"

if [ ! -f "$INSTALL_SCRIPT" ]; then
    echo -e "${RED}Error: Failed to download install script${NC}"
    exit 1
fi

echo -e "${GREEN}[3/4] Patching installer for Android...${NC}"
# Patch the installer to recognize Android as Linux
# Use a different delimiter for sed to avoid escaping issues
sed -i 's|os=$(uname -s)|os=$(uname -s | sed "s/Android/Linux/")|' "$INSTALL_SCRIPT"

echo -e "${GREEN}[4/4] Running patched installer...${NC}"
# Run the patched installer
bash "$INSTALL_SCRIPT"

# Clean up
rm -f "$INSTALL_SCRIPT"

echo ""
echo -e "${GREEN}=========================================="
echo -e "   Devin CLI Installation Complete! 🎉"
echo -e "==========================================${NC}"
echo ""
echo -e "${YELLOW}Next steps:${NC}"
echo "1. Restart Termux"
echo "2. Verify installation: devin --version"
echo "3. Login to Devin: devin auth login"
echo "4. Run Devin in a project: cd your-project && devin"
echo ""

