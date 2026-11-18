#!/data/data/com.termux/files/usr/bin/bash
# ==============================================================================
# One-Line Direct Installer for Tertic on Termux
# Usage: curl -sSL https://raw.githubusercontent.com/Ashed2127/tertic/main/install.sh | bash
# ==============================================================================

set -e

# Terminal Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
CYAN='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m'

echo -e "${CYAN}${BOLD}"
echo "  _____         _   _      "
echo " |_   _|__ _ __| |_(_) ___ "
echo "   | |/ _ \ '__| __| |/ __|"
echo "   | |  __/ |  | |_| | (__ "
echo "   |_|\___|_|   \__|_|\___|"
echo " Autonomous AI Agent for Termux (Google Gemini Inside)"
echo -e "${NC}"

if [ ! -d "/data/data/com.termux" ]; then
    echo -e "${RED}Notice: Designed for Termux on Android. Proceeding for Linux environment...${NC}"
fi

echo -e "${GREEN}[1/3] Installing dependencies (Go, Node, Git, Tmux)...${NC}"
if command -v pkg >/dev/null 2>&1; then
    pkg update -y
    pkg install -y git curl golang nodejs python tmux termux-api jq || true
fi

INSTALL_DIR="$HOME/.tertic-src"
mkdir -p "$INSTALL_DIR"
mkdir -p "$HOME/.local/bin"
mkdir -p "$HOME/bin"

echo -e "${GREEN}[2/3] Downloading latest Tertic...${NC}"
if [ -d "$INSTALL_DIR/.git" ]; then
    (cd "$INSTALL_DIR" && git reset --hard && git pull)
else
    git clone https://github.com/Ashed2127/tertic.git "$INSTALL_DIR"
fi

echo -e "${GREEN}[3/3] Building native agent binary...${NC}"
(cd "$INSTALL_DIR" && go build -ldflags="-s -w" -o "$INSTALL_DIR/bin/tertic-agent" cmd/agent/main.go)

# Link binaries
cp "$INSTALL_DIR/bin/tertic" "$HOME/.local/bin/tertic"
cp "$INSTALL_DIR/bin/tertic" "$HOME/bin/tertic" 2>/dev/null || true
ln -sf "$HOME/.local/bin/tertic" "$HOME/.local/bin/trtc"
ln -sf "$HOME/bin/tertic" "$HOME/bin/trtc" 2>/dev/null || true
chmod +x "$HOME/.local/bin/tertic" "$HOME/.local/bin/trtc"
chmod +x "$HOME/bin/tertic" "$HOME/bin/trtc" 2>/dev/null || true

cp "$INSTALL_DIR/bin/tertic-agent" "$HOME/.local/bin/tertic-agent"
cp "$INSTALL_DIR/bin/tertic-agent" "$HOME/bin/tertic-agent" 2>/dev/null || true
chmod +x "$HOME/.local/bin/tertic-agent"
chmod +x "$HOME/bin/tertic-agent" 2>/dev/null || true

# Add to path in bashrc
if ! grep -q '\$HOME/\.local/bin' "$HOME/.bashrc" 2>/dev/null; then
    echo 'export PATH="$HOME/.local/bin:$HOME/bin:$PATH"' >> "$HOME/.bashrc"
fi
export PATH="$HOME/.local/bin:$HOME/bin:$PATH"

# Run optimize
"$HOME/.local/bin/tertic" optimize || true

echo ""
echo -e "${GREEN}${BOLD}Tertic installed successfully! 🎉${NC}"
echo ""

# Ask user for API Key immediately right after install if not already present
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
echo -e "${GREEN}You are all set! Start the agent anytime by simply typing:${NC}"
echo -e "  ${BOLD}${CYAN}trtc${NC}                 (starts interactive agent session)"
echo -e "  ${BOLD}${CYAN}trtc \"build an app\"${NC}  (runs a task directly)"
echo ""
