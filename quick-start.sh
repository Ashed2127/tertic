#!/data/data/com.termux/files/usr/bin/bash

# Quick-start script for Tertic
# Run this after setup.sh to get started quickly

set -e

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${GREEN}Tertic Quick Start${NC}"
echo ""

# Activate environment
if [ -d "$HOME/tertic-env" ]; then
    echo -e "${GREEN}Activating Tertic environment...${NC}"
    source ~/tertic-env/bin/activate
else
    echo -e "${YELLOW}Environment not found. Run setup.sh first.${NC}"
    exit 1
fi

# Load optimizations
if [ -f "$HOME/.tertic-config" ]; then
    source ~/.tertic-config
fi

echo -e "${GREEN}Environment ready!${NC}"
echo ""
echo -e "${YELLOW}Available commands:${NC}"
echo "  python ~/tertic/examples/simple_agent.py  - Run example agent"
echo "  python --version                           - Check Python version"
echo "  pip list                                  - List installed packages"
echo ""
echo -e "${YELLOW}To set up API keys:${NC}"
echo "  export OPENAI_API_KEY='your-key'"
echo "  export ANTHROPIC_API_KEY='your-key'"
echo ""
