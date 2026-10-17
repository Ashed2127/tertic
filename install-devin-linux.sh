#!/data/data/com.termux/files/usr/bin/bash

set -e

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo "=========================================="
echo "   Installing Devin CLI on Termux"
echo "=========================================="
echo ""
echo "This will try multiple installation methods."
echo ""

# Check if running in Termux
if [ ! -d "/data/data/com.termux" ]; then
    echo -e "${RED}Error: This script must be run in Termux${NC}"
    exit 1
fi

echo -e "${GREEN}[*] Method 1: Installing directly in Termux...${NC}"
# Try installing directly in Termux first
pkg install -y curl
if curl -fsSL https://cli.devin.ai/install.sh | bash; then
    echo -e "${GREEN}✓ Devin CLI installed successfully in Termux!${NC}"
    cat > ~/devin << 'EOF'
#!/data/data/com.termux/files/usr/bin/bash
devin "$@"
EOF
    chmod +x ~/devin
    USE_GLIBC_RUNNER=false
else
    echo -e "${YELLOW}✗ Direct installation failed, trying glibc-runner...${NC}"

    echo -e "${GREEN}[*] Method 2: Installing via glibc-runner...${NC}"
    if pkg install -y glibc-repo 2>/dev/null; then
        pkg install -y glibc-runner
        if grun -- bash -c "curl -fsSL https://cli.devin.ai/install.sh | bash"; then
            echo -e "${GREEN}✓ Devin CLI installed via glibc-runner!${NC}"
            cat > ~/devin << 'EOF'
#!/data/data/com.termux/files/usr/bin/bash
grun -- devin "$@"
EOF
            chmod +x ~/devin
            USE_GLIBC_RUNNER=true
        else
            echo -e "${RED}✗ glibc-runner installation failed${NC}"
            exit 1
        fi
    else
        echo -e "${RED}✗ glibc-repo not available. Please update Termux from GitHub releases.${NC}"
        echo -e "${YELLOW}https://github.com/termux/termux-app/releases${NC}"
        exit 1
    fi
fi

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
if [ "$USE_GLIBC_RUNNER" = true ]; then
    echo -e "${YELLOW}Note:${NC}"
    echo "Devin runs via glibc-runner at native speed."
else
    echo -e "${YELLOW}Note:${NC}"
    echo "Devin runs directly in Termux."
fi
echo ""
