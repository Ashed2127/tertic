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
echo "This will try to install glibc-runner for native performance."
echo "If unavailable, it will fall back to proot-distro/Ubuntu."
echo ""
echo "Note: glibc-runner is ~50MB (fast), Ubuntu is ~500MB (slower)."
echo ""

# Check if running in Termux
if [ ! -d "/data/data/com.termux" ]; then
    echo -e "${RED}Error: This script must be run in Termux${NC}"
    exit 1
fi

echo -e "${GREEN}[*] Attempting to install glibc-repo...${NC}"
if pkg install -y glibc-repo 2>/dev/null; then
    echo -e "${GREEN}[*] Installing glibc-runner...${NC}"
    pkg install -y glibc-runner

    echo -e "${GREEN}[*] Installing Devin CLI via glibc-runner...${NC}"
    # Run the official Devin installer inside glibc environment
    grun -- bash -c "curl -fsSL https://cli.devin.ai/install.sh | bash"

    echo -e "${GREEN}[*] Creating shortcut to run Devin...${NC}"
    # Create a wrapper script to easily run Devin via glibc-runner
    cat > ~/devin << 'EOF'
#!/data/data/com.termux/files/usr/bin/bash
grun -- devin "$@"
EOF
    chmod +x ~/devin

    USE_GLIBC_RUNNER=true
else
    echo -e "${YELLOW}glibc-repo not available, falling back to proot-distro/Ubuntu${NC}"
    echo -e "${YELLOW}This is slower but more compatible. Consider updating Termux from GitHub releases.${NC}"

    echo -e "${GREEN}[*] Installing proot-distro...${NC}"
    pkg install -y proot-distro

    echo -e "${GREEN}[*] Installing Ubuntu in proot-distro...${NC}"
    # Reset Ubuntu to ensure clean DNS configuration
    proot-distro reset ubuntu --overwrite 2>/dev/null || proot-distro install ubuntu 2>/dev/null || true
    echo "Ubuntu setup complete."

    echo -e "${GREEN}[*] Installing curl and Devin CLI in Ubuntu...${NC}"
    # Configure DNS in Ubuntu and install
    proot-distro login ubuntu -- bash -c "echo 'nameserver 8.8.8.8' > /etc/resolv.conf && echo 'nameserver 8.8.4.4' >> /etc/resolv.conf && apt update && apt install -y curl && curl -fsSL https://cli.devin.ai/install.sh | bash"

    echo -e "${GREEN}[*] Creating shortcut to run Devin...${NC}"
    # Create a wrapper script to easily run Devin from Ubuntu
    cat > ~/devin << 'EOF'
#!/data/data/com.termux/files/usr/bin/bash
proot-distro login ubuntu -- devin "$@"
EOF
    chmod +x ~/devin

    USE_GLIBC_RUNNER=false
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
    echo "Devin runs via glibc-runner at native speed (no proot overhead)."
    echo "This provides glibc compatibility without the slowdown!"
else
    echo -e "${YELLOW}Note:${NC}"
    echo "Devin runs inside Ubuntu via proot-distro."
    echo "For better performance, update Termux from GitHub releases."
    echo "Latest version: https://github.com/termux/termux-app/releases"
fi
echo ""
