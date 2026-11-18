#!/data/data/com.termux/files/usr/bin/bash

set -e

echo "=========================================="
echo "   Tertic - Optimized Termux for Android"
echo "=========================================="
echo ""

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Check if running in Termux
if [ ! -d "/data/data/com.termux" ]; then
    echo -e "${RED}Error: This script must be run in Termux${NC}"
    exit 1
fi

echo -e "${GREEN}[1/5] Updating Termux packages...${NC}"
pkg update -y
pkg upgrade -y

echo -e "${GREEN}[2/5] Installing core dependencies...${NC}"
pkg install -y git curl tar python nodejs golang jq tmux termux-api

echo -e "${GREEN}[3/5] Configuring Termux storage access...${NC}"
# Check if storage is already configured
if [ ! -d "$HOME/storage" ]; then
    termux-setup-storage || true
else
    echo "Storage already configured, skipping..."
fi

echo -e "${GREEN}[4/5] Setting up development environment...${NC}"
# Create bin directories
mkdir -p ~/.local/bin
mkdir -p ~/bin
mkdir -p ~/.tertic

# Add both ~/.local/bin and ~/bin to PATH if not already present
if ! grep -q '\$HOME/\.local/bin' ~/.bashrc 2>/dev/null; then
    echo 'export PATH="$HOME/.local/bin:$HOME/bin:$PATH"' >> ~/.bashrc
fi
export PATH="$HOME/.local/bin:$HOME/bin:$PATH"

echo -e "${GREEN}[5/5] Building Tertic AI Agent & installing CLI...${NC}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Build the Go agent binary for Termux
if command -v go >/dev/null 2>&1; then
    echo "Building native Go agent (tertic-agent)..."
    (cd "$SCRIPT_DIR" && go build -ldflags="-s -w" -o "$SCRIPT_DIR/bin/tertic-agent" cmd/agent/main.go)
fi

cp "$SCRIPT_DIR/bin/tertic" ~/.local/bin/tertic
cp "$SCRIPT_DIR/bin/tertic" ~/bin/tertic 2>/dev/null || true
ln -sf ~/.local/bin/tertic ~/.local/bin/trtc
ln -sf ~/bin/tertic ~/bin/trtc 2>/dev/null || true
chmod +x ~/.local/bin/tertic ~/.local/bin/trtc
chmod +x ~/bin/tertic ~/bin/trtc 2>/dev/null || true

if [ -f "$SCRIPT_DIR/bin/tertic-agent" ]; then
    cp "$SCRIPT_DIR/bin/tertic-agent" ~/.local/bin/tertic-agent
    cp "$SCRIPT_DIR/bin/tertic-agent" ~/bin/tertic-agent 2>/dev/null || true
    chmod +x ~/.local/bin/tertic-agent
    chmod +x ~/bin/tertic-agent 2>/dev/null || true
fi

# Run Tertic optimize to auto-tune based on hardware profile
tertic optimize || bash ~/.local/bin/tertic optimize

echo ""
echo -e "${GREEN}=========================================="
echo -e "   Tertic Setup Complete! 🎉"
echo -e "==========================================${NC}"
echo ""

# Prompt for API Key if not already present
CONFIG_FILE="$HOME/.tertic/config.env"
if ! grep -q "GEMINI_API_KEY" "$CONFIG_FILE" 2>/dev/null; then
    echo -e "${YELLOW}${BOLD}[Setup] Please enter your Google Gemini API Key${NC}"
    echo -e "Get a free key here: ${CYAN}https://aistudio.google.com/app/apikey${NC}"
    read -p "API Key: " USER_KEY
    if [ -n "$USER_KEY" ]; then
        mkdir -p "$HOME/.tertic"
        echo "export GEMINI_API_KEY=\"$USER_KEY\"" >> "$CONFIG_FILE"
        echo -e "${GREEN}✓ Key saved to $CONFIG_FILE${NC}"
    fi
fi

echo ""
echo -e "${GREEN}Start the agent anytime with the short command:${NC}"
echo -e "   ${BOLD}${CYAN}trtc${NC}                 (starts interactive agent session)"
echo -e "   ${BOLD}${CYAN}trtc \"build an app\"${NC}  (runs a task directly)"
echo ""
