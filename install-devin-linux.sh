#!/data/data/com.termux/files/usr/bin/bash

set -e

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo "=========================================="
echo "   Installing Devin CLI via Ubuntu on Android"
echo "=========================================="
echo ""
echo "This will install Ubuntu inside Termux using proot-distro,"
echo "then install Devin CLI in that Ubuntu environment."
echo ""

# Check if running in Termux
if [ ! -d "/data/data/com.termux" ]; then
    echo -e "${RED}Error: This script must be run in Termux${NC}"
    exit 1
fi

echo -e "${GREEN}[1/5] Installing proot-distro...${NC}"
pkg install -y proot-distro

echo -e "${GREEN}[2/5] Installing Ubuntu in proot-distro...${NC}"
# Install Ubuntu (this will take a while)
proot-distro install ubuntu

echo -e "${GREEN}[3/5] Installing curl in Ubuntu...${NC}"
proot-distro login ubuntu -- bash -c "apt update && apt install -y curl"

echo -e "${GREEN}[4/5] Installing Devin CLI in Ubuntu...${NC}"
# Run the official Devin installer inside Ubuntu
proot-distro login ubuntu -- bash -c "curl -fsSL https://cli.devin.ai/install.sh | bash"

echo -e "${GREEN}[5/5] Creating shortcut to run Devin...${NC}"
# Create a wrapper script to easily run Devin from Ubuntu
cat > ~/devin << 'EOF'
#!/data/data/com.termux/files/usr/bin/bash
proot-distro login ubuntu -- devin "$@"
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
echo "Devin runs inside Ubuntu via proot-distro."
echo "This provides a full glibc Linux environment on Android!"
echo ""
