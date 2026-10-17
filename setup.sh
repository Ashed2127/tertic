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
pkg install -y git curl tar python nodejs

echo -e "${GREEN}[3/5] Configuring Termux storage access...${NC}"
termux-setup-storage

echo -e "${GREEN}[4/5] Setting up development environment...${NC}"
# Create bin directory
mkdir -p ~/.local/bin

# Add to PATH
if ! grep -q "~/.local/bin" ~/.bashrc; then
    echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc
fi

echo -e "${GREEN}[5/5] Applying mobile optimizations...${NC}"
# Create optimized config
cat > ~/.tertic-config << 'EOF'
# Tertic Configuration
export PYTHONUNBUFFERED=1
export PYTHONDONTWRITEBYTECODE=1
export EDITOR=nano
export PAGER=less
EOF

# Add to bashrc
if ! grep -q "tertic-config" ~/.bashrc; then
    echo "source ~/.tertic-config" >> ~/.bashrc
fi

echo ""
echo -e "${GREEN}=========================================="
echo -e "   Tertic Setup Complete! 🎉"
echo -e "==========================================${NC}"
echo ""
echo -e "${YELLOW}Next steps:${NC}"
echo "1. Restart Termux"
echo "2. Run: bash ~/tertic/advanced-setup.sh (for more tools)"
echo "3. Start coding!"
echo ""
echo -e "${YELLOW}Available tools:${NC}"
echo "  python, nodejs, git, nano"
echo ""
echo -e "${YELLOW}Optional tools (advanced-setup.sh):${NC}"
echo "  tmux, ripgrep, fzf, neovim, black, prettier"
echo ""
