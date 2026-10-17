#!/data/data/com.termux/files/usr/bin/bash

# Advanced setup for Tertic
# Installs additional AI agent frameworks and tools

set -e

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${GREEN}Tertic Advanced Setup${NC}"
echo ""

# Check if basic setup is done
if [ ! -d "$HOME/tertic-env" ]; then
    echo -e "${YELLOW}Basic environment not found. Run setup.sh first.${NC}"
    exit 1
fi

source ~/tertic-env/bin/activate

echo -e "${GREEN}Installing additional packages...${NC}"

# Install LangChain (full)
echo "Installing LangChain..."
pip install langchain langchain-community

# Install Anthropic
echo "Installing Anthropic SDK..."
pip install anthropic

# Install additional tools
echo "Installing additional tools..."
pip install httpx websockets

# Install debugging tools
echo "Installing debugging tools..."
pip install ipdb

echo -e "${GREEN}Installing system tools...${NC}"
pkg install -y tmux ripgrep fzf

echo -e "${GREEN}Configuring tmux...${NC}"
cat > ~/.tmux.conf << 'EOF'
# Tertic tmux configuration
set -g default-terminal "screen-256color"
set -g history-limit 10000
set -g mouse on
set -g status-bg black
set -g status-fg green
EOF

echo ""
echo -e "${GREEN}Advanced setup complete!${NC}"
echo ""
echo -e "${YELLOW}New tools available:${NC}"
echo "  tmux      - Terminal multiplexer (tmux new -s session)"
echo "  rg        - Fast search (ripgrep)"
echo "  fzf       - Fuzzy finder"
echo ""
echo -e "${YELLOW}New Python packages:${NC}"
echo "  langchain - Full LangChain framework"
echo "  anthropic - Claude API support"
echo "  ipdb      - Python debugger"
echo ""
