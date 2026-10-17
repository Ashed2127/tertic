#!/data/data/com.termux/files/usr/bin/bash

set -e

echo "=========================================="
echo "   Tertic - Termux AI Agent Optimizer   "
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

echo -e "${GREEN}[1/7] Updating Termux packages...${NC}"
pkg update -y
pkg upgrade -y

echo -e "${GREEN}[2/7] Installing core dependencies...${NC}"
pkg install -y python git curl wget

echo -e "${GREEN}[3/7] Installing Python development tools...${NC}"
pkg install -y python-pip

echo -e "${GREEN}[4/7] Configuring Termux storage access...${NC}"
termux-setup-storage

echo -e "${GREEN}[5/7] Setting up Python environment...${NC}"
mkdir -p ~/tertic-env
python -m venv ~/tertic-env
source ~/tertic-env/bin/activate

echo -e "${GREEN}[6/7] Installing optimized AI agent dependencies...${NC}"
pip install --upgrade pip
pip install aiohttp aiosqlite pyyaml python-dotenv
pip install rich prompt-toolkit

echo -e "${GREEN}[7/7] Applying mobile optimizations...${NC}"
# Create optimized config
cat > ~/.tertic-config << 'EOF'
# Tertic Configuration
export PYTHONUNBUFFERED=1
export PYTHONDONTWRITEBYTECODE=1
export OMP_NUM_THREADS=4
export MALLOC_TRIM_THRESHOLD_=65536
EOF

# Add to bashrc
if ! grep -q "tertic-config" ~/.bashrc; then
    echo "source ~/.tertic-config" >> ~/.bashrc
fi

echo ""
echo -e "${GREEN}=========================================="
echo -e "   Installation Complete! 🎉"
echo -e "==========================================${NC}"
echo ""
echo -e "${YELLOW}Next steps:${NC}"
echo "1. Restart Termux"
echo "2. Activate environment: source ~/tertic-env/bin/activate"
echo "3. Check installation: python --version"
echo ""
echo -e "${YELLOW}Run example agent (no API key needed):${NC}"
echo "  python ~/tertic/examples/simple_agent.py"
echo ""
echo -e "${YELLOW}To use OpenAI/Anthropic (optional):${NC}"
echo "  pip install openai anthropic"
echo "  export OPENAI_API_KEY='your-key'"
echo ""
