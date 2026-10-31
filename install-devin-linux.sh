#!/usr/bin/env bash

# ==============================================================================
# Tertic - Robust Devin CLI Installer for Termux / Android
# Uses Ubuntu PRoot (glibc) as primary reliable method with direct/glibc fallback
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

# Ensure bin path
mkdir -p "$HOME/.local/bin"
mkdir -p "$HOME/bin"

install_via_proot() {
    echo -e "${GREEN}[*] Installing via PRoot Ubuntu (glibc compatible environment)...${NC}"
    
    if [ "$IS_TERMUX" = true ]; then
        pkg update -y
        pkg install -y proot-distro curl tmux tar
    fi

    # Install Ubuntu if not installed
    if ! proot-distro list 2>/dev/null | grep -q "ubuntu.*installed"; then
        echo -e "${YELLOW}Setting up Ubuntu container (~300MB)...${NC}"
        proot-distro install ubuntu
    else
        echo -e "${GREEN}✓ Ubuntu container already present.${NC}"
    fi

    # Fix DNS inside Ubuntu container if Termux resolv.conf exists
    RESOLV_BIND=""
    if [ -f "/data/data/com.termux/files/usr/etc/resolv.conf" ]; then
        RESOLV_BIND="--bind /data/data/com.termux/files/usr/etc/resolv.conf:/etc/resolv.conf"
    fi

    echo -e "${GREEN}[*] Provisioning Ubuntu and installing Devin CLI...${NC}"
    proot-distro login ubuntu $RESOLV_BIND -- bash -c '
        set -e
        apt-get update -y
        apt-get install -y --no-install-recommends curl ca-certificates git python3 nodejs
        echo "Running official Devin installer inside Ubuntu..."
        curl -fsSL https://cli.devin.ai/install.sh | bash
    '

    echo -e "${GREEN}[*] Creating seamless Termux wrapper for 'devin'...${NC}"

    # Wrapper script that maps working directory & home
    cat > "$HOME/.local/bin/devin" << 'EOF'
#!/usr/bin/env bash
CURRENT_DIR="$(pwd)"
RESOLV_BIND=""
if [ -f "/data/data/com.termux/files/usr/etc/resolv.conf" ]; then
    RESOLV_BIND="--bind /data/data/com.termux/files/usr/etc/resolv.conf:/etc/resolv.conf"
fi

if command -v termux-wake-lock >/dev/null 2>&1; then
    termux-wake-lock
fi

proot-distro login ubuntu $RESOLV_BIND \
    --bind "$HOME:/root/termux-home" \
    --bind "$CURRENT_DIR:/workspace" \
    -- bash -c "
        export PATH=\"/root/.local/bin:/root/bin:\$PATH\"
        export PAGER=cat
        export GIT_PAGER=cat
        cd /workspace
        devin \"\$@\"
    "
EOF
    chmod +x "$HOME/.local/bin/devin"
    # Also link to ~/bin/devin for fallback
    cp "$HOME/.local/bin/devin" "$HOME/bin/devin" 2>/dev/null || true
    chmod +x "$HOME/bin/devin" 2>/dev/null || true

    echo -e "${GREEN}✓ Devin CLI successfully installed via Ubuntu PRoot!${NC}"
}

# Execute installation
install_via_proot

# Ensure PATH in ~/.bashrc
if ! grep -q '\$HOME/\.local/bin' "$HOME/.bashrc" 2>/dev/null; then
    echo 'export PATH="$HOME/.local/bin:$HOME/bin:$PATH"' >> "$HOME/.bashrc"
fi

echo -e "\n${GREEN}==========================================${NC}"
echo -e "${GREEN}   Devin CLI Installation Complete! 🎉    ${NC}"
echo -e "${GREEN}==========================================${NC}\n"
echo -e "${YELLOW}Usage:${NC}"
echo "  devin --version"
echo "  devin auth login"
echo "  tertic run devin"
echo ""
