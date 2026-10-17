#!/data/data/com.termux/files/usr/bin/bash

# Quick-start script for Tertic
# Run this after setup.sh to get started quickly

set -e

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${GREEN}Tertic Quick Start${NC}"
echo ""

# Load optimizations
if [ -f "$HOME/.tertic-config" ]; then
    source ~/.tertic-config
fi

# Add to PATH
export PATH="$HOME/.local/bin:$PATH"

echo -e "${GREEN}Environment ready!${NC}"
echo ""
echo -e "${YELLOW}Check Devin CLI:${NC}"
if [ -f "$HOME/devin" ]; then
    echo "  devin --version"
    ~/devin --version
else
    echo "  Devin CLI not installed. Run: bash ~/tertic/install-devin-linux.sh"
fi
echo ""
echo -e "${YELLOW}Available commands:${NC}"
echo "  devin                  - Start Devin CLI"
echo "  devin auth login       - Login to Devin"
echo "  python                 - Run Python"
echo "  node                   - Run Node.js"
echo ""
echo -e "${YELLOW}For more tools:${NC}"
echo "  bash ~/tertic/advanced-setup.sh"
echo ""
