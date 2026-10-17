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

echo -e "${GREEN}[1/8] Updating Termux packages...${NC}"
pkg update -y
pkg upgrade -y

echo -e "${GREEN}[2/8] Installing core dependencies...${NC}"
pkg install -y python git curl wget

echo -e "${GREEN}[3/8] Installing Python development tools...${NC}"
pkg install -y python-pip

echo -e "${GREEN}[4/8] Installing UV (fast Python package manager)...${NC}"
curl -LsSf https://astral.sh/uv/install.sh | sh
export PATH="$HOME/.local/bin:$PATH"

echo -e "${GREEN}[5/8] Configuring Termux storage access...${NC}"
termux-setup-storage

echo -e "${GREEN}[6/8] Setting up Python environment...${NC}"
mkdir -p ~/tertic-env
uv venv ~/tertic-env
source ~/tertic-env/bin/activate

echo -e "${GREEN}[7/8] Installing optimized AI agent dependencies...${NC}"
uv pip install --upgrade pip
uv pip install aiohttp aiosqlite pyyaml python-dotenv
uv pip install rich textual prompt-toolkit
uv pip install openai anthropic
uv pip install langchain langchain-openai

echo -e "${GREEN}[8/8] Applying mobile optimizations...${NC}"
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
echo -e "${YELLOW}Run example agent:${NC}"
echo "  python ~/tertic/examples/simple_agent.py"
echo ""
