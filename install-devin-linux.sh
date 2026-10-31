#!/usr/bin/env bash

# ==============================================================================
# Tertic - Robust Devin CLI Installer for Termux / Android
# Uses Ubuntu PRoot (glibc) environment and creates seamless wrapper
# ==============================================================================

set -e

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
CYAN='\033[0;36m'
NC='\033[0m'

echo -e "${CYAN}==========================================${NC}"
echo -e "${CYAN}   Installing Devin CLI on Termux (Tertic)${NC}"
echo -e "${CYAN}==========================================${NC}\n"

IS_TERMUX=false
if [ -d "/data/data/com.termux" ]; then
    IS_TERMUX=true
fi

# Ensure bin directories
mkdir -p "$HOME/.local/bin"
mkdir -p "$HOME/bin"

# Remove any old broken direct-termux devin binary
rm -f "$HOME/.local/bin/devin"
rm -f "$HOME/bin/devin"
rm -f "$HOME/devin"

if [ "$IS_TERMUX" = true ]; then
    echo -e "${GREEN}[1/4] Checking Termux base dependencies...${NC}"
    pkg update -y || true
    pkg install -y proot-distro curl tar tmux jq
fi

# 1. Check if Ubuntu container is installed
echo -e "\n${GREEN}[2/4] Ensuring Ubuntu PRoot container exists...${NC}"
PREFIX_PATH="${PREFIX:-/data/data/com.termux/files/usr}"
UBUNTU_ROOTFS="$PREFIX_PATH/var/lib/proot-distro/installed-rootfs/ubuntu"

if [ -d "$UBUNTU_ROOTFS" ]; then
    echo -e "${GREEN}✓ Ubuntu container is installed.${NC}"
else
    echo "Installing Ubuntu container (~300MB)..."
    proot-distro install ubuntu || true
fi

# 2. Fix Ubuntu DNS, IPv4 forcing, and static hosts
echo -e "\n${GREEN}[3/4] Optimizing Ubuntu networking & hosts...${NC}"
proot-distro login ubuntu -- bash -c '
    rm -f /etc/resolv.conf
    printf "nameserver 1.1.1.1\nnameserver 8.8.8.8\nnameserver 8.8.4.4\n" > /etc/resolv.conf
    
    # Inject static IPs for ports.ubuntu.com to eliminate DNS hangs completely
    if ! grep -q "ports.ubuntu.com" /etc/hosts 2>/dev/null; then
        echo "91.189.92.19 ports.ubuntu.com" >> /etc/hosts
        echo "91.189.92.21 ports.ubuntu.com" >> /etc/hosts
    fi

    # Force apt IPv4
    mkdir -p /etc/apt/apt.conf.d
    echo "Acquire::ForceIPv4 \"true\";" > /etc/apt/apt.conf.d/99force-ipv4
'

# 3. Download and Install Devin CLI (via Termux bridge)
echo -e "\n${GREEN}[4/4] Installing Devin CLI into Ubuntu glibc environment...${NC}"

# Fetch manifest in Termux where network is 100% working
echo -e "${CYAN}Fetching latest Devin release information...${NC}"
MANIFEST=$(curl -sSf "https://static.devin.ai/cli/current/manifest.json" || true)

if [ -z "$MANIFEST" ]; then
    # Fallback to known stable version if manifest fails
    DEVIN_VERSION="3000.11.234"
else
    DEVIN_VERSION=$(echo "$MANIFEST" | grep -o '"version"[[:space:]]*:[[:space:]]*"[^"]*"' | head -1 | sed 's/.*"\([^"]*\)"$/\1/')
fi

ARCH=$(uname -m)
if [ "$ARCH" = "aarch64" ]; then
    TAR_URL="https://static.devin.ai/cli/${DEVIN_VERSION}/devin-${DEVIN_VERSION}-aarch64-unknown-linux.tar.gz"
else
    TAR_URL="https://static.devin.ai/cli/${DEVIN_VERSION}/devin-${DEVIN_VERSION}-x86_64-unknown-linux.tar.gz"
fi

echo -e "${CYAN}Downloading Devin v${DEVIN_VERSION} for ${ARCH} in Termux...${NC}"
curl -fsSL "$TAR_URL" -o "$HOME/devin-bundle.tar.gz"

echo -e "${CYAN}Unpacking Devin into Ubuntu container...${NC}"
proot-distro login ubuntu --bind "$HOME:/root/termux-home" -- bash -c "
    mkdir -p /root/.local/share/devin/cli/_versions/${DEVIN_VERSION}
    tar xzf /root/termux-home/devin-bundle.tar.gz -C /root/.local/share/devin/cli/_versions/${DEVIN_VERSION}
    rm -f /root/.local/share/devin/cli/_versions/current
    ln -s ${DEVIN_VERSION} /root/.local/share/devin/cli/_versions/current
    mkdir -p /root/.local/bin
    rm -f /root/.local/bin/devin
    ln -s /root/.local/share/devin/cli/_versions/current/bin/devin /root/.local/bin/devin
    chmod +x /root/.local/share/devin/cli/_versions/${DEVIN_VERSION}/bin/devin
"
rm -f "$HOME/devin-bundle.tar.gz"

# 4. Create Seamless Termux Launcher (~/.local/bin/devin)
echo -e "\n${GREEN}[*] Creating ~/.local/bin/devin launcher...${NC}"
cat > "$HOME/.local/bin/devin" << 'EOF'
#!/usr/bin/env bash
# Tertic - Transparent Devin CLI launcher running via PRoot Ubuntu
CURRENT_DIR="$(pwd)"
WAKE_LOCK_ACTIVE=false

if command -v termux-wake-lock >/dev/null 2>&1; then
    termux-wake-lock
    WAKE_LOCK_ACTIVE=true
fi

proot-distro login ubuntu \
    --bind "$HOME:/root/termux-home" \
    --bind "$CURRENT_DIR:/workspace" \
    -- bash -c "
        export PATH=\"/root/.local/bin:/root/bin:\$PATH\"
        export PAGER=cat
        export GIT_PAGER=cat
        cd /workspace
        if [ -f /root/.local/bin/devin ]; then
            /root/.local/bin/devin \"\$@\"
        elif [ -f /root/.local/share/devin/cli/_versions/current/bin/devin ]; then
            /root/.local/share/devin/cli/_versions/current/bin/devin \"\$@\"
        elif command -v devin >/dev/null 2>&1; then
            devin \"\$@\"
        else
            echo 'Error: Devin binary not found inside Ubuntu container.'
            exit 1
        fi
    " -- "$@"

EXIT_STATUS=$?

if [ "$WAKE_LOCK_ACTIVE" = true ] && command -v termux-wake-unlock >/dev/null 2>&1; then
    termux-wake-unlock
fi

exit $EXIT_STATUS
EOF

chmod +x "$HOME/.local/bin/devin"
cp "$HOME/.local/bin/devin" "$HOME/bin/devin" 2>/dev/null || true

# Prepend ~/.local/bin to PATH in ~/.bashrc
if ! grep -q '\$HOME/\.local/bin' "$HOME/.bashrc" 2>/dev/null; then
    echo 'export PATH="$HOME/.local/bin:$HOME/bin:$PATH"' >> "$HOME/.bashrc"
fi
export PATH="$HOME/.local/bin:$HOME/bin:$PATH"

echo -e "\n${GREEN}==========================================${NC}"
echo -e "${GREEN}   Devin CLI Setup Complete! 🎉           ${NC}"
echo -e "${GREEN}==========================================${NC}\n"
echo -e "Usage:"
echo -e "  ${YELLOW}devin --version${NC}          # Check version"
echo -e "  ${YELLOW}devin auth login${NC}         # Authenticate Devin"
echo -e "  ${YELLOW}tertic run devin${NC}         # Run in persistent background session"
echo ""
