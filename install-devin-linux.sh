#!/data/data/com.termux/files/usr/bin/bash

set -e

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo "=========================================="
echo "   Installing Devin CLI via glibc-runner"
echo "=========================================="
echo ""
echo "This will install glibc-runner for native glibc support,"
echo "then install Devin CLI to run at native speed."
echo ""
echo "Note: glibc-runner is ~50MB and takes 2-3 minutes to install."
echo "This is MUCH faster than proot-distro/Ubuntu (2x-10x speedup)."
echo ""

# Check if running in Termux
if [ ! -d "/data/data/com.termux" ]; then
    echo -e "${RED}Error: This script must be run in Termux${NC}"
    exit 1
fi

echo -e "${GREEN}[1/4] Installing glibc-repo...${NC}"
pkg install -y glibc-repo

echo -e "${GREEN}[2/4] Installing glibc-runner...${NC}"
pkg install -y glibc-runner

echo -e "${GREEN}[3/4] Installing Devin CLI via glibc-runner...${NC}"
# Run the official Devin installer inside glibc environment
grun -- bash -c "curl -fsSL https://cli.devin.ai/install.sh | bash"

echo -e "${GREEN}[4/4] Creating shortcut to run Devin...${NC}"
# Create a wrapper script to easily run Devin via glibc-runner
cat > ~/devin << 'EOF'
#!/data/data/com.termux/files/usr/bin/bash
grun -- devin "$@"
EOF
chmod +x ~/devin

# Add to PATH if not already there
if ! grep -q "export PATH=\"\$HOME:\$PATH\"" ~/.bashrc; then
    echo 'export PATH="$HOME:$PATH"' >> ~/.bashrc
fi

echo ""
echo -e "${GREEN}=========================================="
echo -e "   Devin CLI Installation Complete! 🎉"
echo -e "==========================================${NC}"
echo ""
echo -e "${YELLOW}Usage:${NC}"
echo "  devin --version           # Check version"
echo "  devin auth login          # Login to Devin"
echo "  devin                    # Start Devin in current directory"
echo ""
echo -e "${YELLOW}Note:${NC}"
echo "Devin runs via glibc-runner at native speed (no proot overhead)."
echo "This provides glibc compatibility without the slowdown!"
echo ""
