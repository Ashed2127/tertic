#!/data/data/com.termux/files/usr/bin/bash

# Advanced setup for Tertic
# Installs additional development tools for better coding experience

set -e

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${GREEN}Tertic Advanced Setup${NC}"
echo ""

echo -e "${GREEN}Installing additional development tools...${NC}"

# Install Python tools
echo "Installing Python tools..."
pkg install -y python-pip

# Install Node.js for web development
echo "Installing Node.js..."
pkg install -y nodejs

# Install useful CLI tools
echo "Installing CLI tools..."
pkg install -y tmux ripgrep fzf neovim htop

echo -e "${GREEN}Configuring tmux...${NC}"
cat > ~/.tmux.conf << 'EOF'
# Tertic tmux configuration
set -g default-terminal "screen-256color"
set -g history-limit 10000
set -g mouse on
set -g status-bg black
set -g status-fg green
set -g status-left "Tertic > "
set -g status-right "%H:%M "
EOF

echo -e "${GREEN}Installing common Python packages...${NC}"
# Don't upgrade pip in Termux (it's managed by Termux)
pip install black pylint mypy requests

echo -e "${GREEN}Installing common Node.js packages...${NC}"
npm install -g prettier eslint typescript

echo ""
echo -e "${GREEN}Advanced setup complete!${NC}"
echo ""
echo -e "${YELLOW}New tools available:${NC}"
echo "  tmux      - Terminal multiplexer (tmux new -s coding)"
echo "  rg        - Fast code search (ripgrep)"
echo "  fzf       - Fuzzy file finder"
echo "  nvim      - Neovim editor"
echo "  htop      - System monitor"
echo ""
echo -e "${YELLOW}New Python tools:${NC}"
echo "  black     - Code formatter"
echo "  pylint    - Code linter"
echo "  mypy      - Type checker"
echo ""
echo -e "${YELLOW}New Node.js tools:${NC}"
echo "  prettier  - Code formatter"
echo "  eslint    - JavaScript linter"
echo "  typescript - TypeScript compiler"
echo ""
echo -e "${YELLOW}Usage with tmux:${NC}"
echo "  tmux new -s coding    # Start tmux session"
echo "  Ctrl+B then d         # Detach from tmux"
echo "  tmux attach -t coding # Reattach to session"
echo ""
