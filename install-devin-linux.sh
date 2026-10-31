#!/usr/bin/env bash

# ==============================================================================
# Tertic - Robust Devin CLI Installer for Termux / Android
# Handles Android DNS restrictions, IPv4 forcing, and PRoot glibc bridge
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

if [ "$IS_TERMUX" = true ]; then
    echo -e "${GREEN}[1/5] Checking Termux base dependencies...${NC}"
    pkg update -y || true
    pkg install -y proot-distro curl tar tmux jq
fi

# 1. Install Ubuntu if not present
echo -e "\n${GREEN}[2/5] Ensuring Ubuntu PRoot container is installed...${NC}"
if ! proot-distro list 2>/dev/null | grep -q "ubuntu.*installed"; then
    echo "Installing Ubuntu container (~300MB)..."
    proot-distro install ubuntu
else
    echo "✓ Ubuntu container is installed."
fi

# 2. Fix Ubuntu DNS, IPv4 forcing, and resolv.conf
echo -e "\n${GREEN}[3/5] Fixing DNS resolution and forcing IPv4 for apt...${NC}"
proot-distro login ubuntu -- bash -c '
    # Remove broken symlinks to systemd-resolved (127.0.0.53)
    rm -f /etc/resolv.conf
    printf "nameserver 1.1.1.1\nnameserver 8.8.8.8\nnameserver 8.8.4.4\n" > /etc/resolv.conf
    
    # Force apt to use IPv4 to avoid hanging on "Connecting to ports.ubuntu.com"
    mkdir -p /etc/apt/apt.conf.d
    echo "Acquire::ForceIPv4 \"true\";" > /etc/apt/apt.conf.d/99force-ipv4
    
    # Ensure /etc/resolv.conf persists on bash startup
    if ! grep -q "nameserver 1.1.1.1" /root/.bashrc 2>/dev/null; then
        echo "printf \"nameserver 1.1.1.1\nnameserver 8.8.8.8\n\" > /etc/resolv.conf 2>/dev/null || true" >> /root/.bashrc
    fi
'

# 3. Test Network inside Ubuntu & Install Tools
echo -e "\n${GREEN}[4/5] Testing network and installing base tools in Ubuntu...${NC}"
NETWORK_OK=false
if proot-distro login ubuntu -- bash -c "apt-get update -o Acquire::ForceIPv4=true -y && apt-get install -y --no-install-recommends curl ca-certificates tar"; then
    NETWORK_OK=true
    echo -e "${GREEN}✓ Ubuntu network and APT working successfully!${NC}"
else
    echo -e "${YELLOW}Warning: Direct APT update encountered issues. Trying alternative mirror...${NC}"
    proot-distro login ubuntu -- bash -c '
        # Fallback to mirror if default ports.ubuntu.com stalls
        sed -i "s|ports.ubuntu.com/ubuntu-ports|mirror.ox.ac.uk/sites/ports.ubuntu.com|g" /etc/apt/sources.list 2>/dev/null || true
        apt-get update -o Acquire::ForceIPv4=true -y || true
        apt-get install -y --no-install-recommends curl ca-certificates tar || true
    '
fi

# 4. Install Devin CLI
echo -e "\n${GREEN}[5/5] Installing Devin CLI binary...${NC}"
# Method A: Try official installer inside Ubuntu
DEVIN_INSTALLED=false
if proot-distro login ubuntu -- bash -c "curl -fsSL https://cli.devin.ai/install.sh | bash"; then
    DEVIN_INSTALLED=true
else
    echo -e "${YELLOW}Official curl-bash failed inside container. Using Termux download bridge...${NC}"
    # Method B: Fetch manifest directly in Termux (where network works) and transfer to Ubuntu
    MANIFEST=$(curl -sSf "https://static.devin.ai/cli/current/manifest.json" || true)
    if [ -n "$MANIFEST" ]; then
        DEVIN_VERSION=$(echo "$MANIFEST" | grep -o '"version"[[:space:]]*:[[:space:]]*"[^"]*"' | head -1 | sed 's/.*"\([^"]*\)"$/\1/')
        ARCH=$(uname -m)
        if [ "$ARCH" = "aarch64" ]; then
            TAR_URL="https://static.devin.ai/cli/${DEVIN_VERSION}/devin-${DEVIN_VERSION}-aarch64-unknown-linux.tar.gz"
        else
            TAR_URL="https://static.devin.ai/cli/${DEVIN_VERSION}/devin-${DEVIN_VERSION}-x86_64-unknown-linux.tar.gz"
        fi
        
        echo -e "${CYAN}Downloading Devin v${DEVIN_VERSION} ($ARCH) in Termux...${NC}"
        curl -fsSL "$TAR_URL" -o "$HOME/devin-bundle.tar.gz"
        
        # Unpack inside Ubuntu
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
        DEVIN_INSTALLED=true
    fi
fi

# 5. Create Seamless Termux Launcher
echo -e "\n${GREEN}[*] Generating ~/.local/bin/devin launcher...${NC}"
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
        if command -v devin >/dev/null 2>&1; then
            devin \"\$@\"
        elif [ -f /root/.local/bin/devin ]; then
            /root/.local/bin/devin \"\$@\"
        elif [ -f /root/.local/share/devin/cli/_versions/current/bin/devin ]; then
            /root/.local/share/devin/cli/_versions/current/bin/devin \"\$@\"
        else
            echo 'Error: Devin executable not found in Ubuntu container.'
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

# Add to PATH in ~/.bashrc
if ! grep -q '\$HOME/\.local/bin' "$HOME/.bashrc" 2>/dev/null; then
    echo 'export PATH="$HOME/.local/bin:$HOME/bin:$PATH"' >> "$HOME/.bashrc"
fi

echo -e "\n${GREEN}==========================================${NC}"
echo -e "${GREEN}   Devin CLI Setup Complete! 🎉           ${NC}"
echo -e "${GREEN}==========================================${NC}\n"
echo -e "Usage:"
echo -e "  ${YELLOW}devin --version${NC}          # Verify version"
echo -e "  ${YELLOW}devin auth login${NC}         # Authenticate Devin"
echo -e "  ${YELLOW}tertic run devin${NC}         # Run in persistent background session"
echo ""
