#!/usr/bin/env bash
# ==============================================================================
# Tertic - Antigravity (agy) CLI Installer for Termux / Android
# ==============================================================================

set -e

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
CYAN='\033[0;36m'
NC='\033[0m'

echo -e "${CYAN}==========================================${NC}"
echo -e "${CYAN}   Installing Antigravity (agy) on Termux ${NC}"
echo -e "${CYAN}==========================================${NC}\n"

# Ensure dependencies
pkg update -y
pkg install -y proot-distro curl tmux termux-api jq

echo -e "\n${GREEN}[1/4] Setting up Ubuntu glibc environment...${NC}"
if ! proot-distro list | grep -q "ubuntu.*installed"; then
    echo "Installing Ubuntu rootfs (glibc compatibility layer)..."
    proot-distro install ubuntu
else
    echo "Ubuntu environment already installed."
fi

echo -e "\n${GREEN}[2/4] Provisioning agent dependencies inside Ubuntu...${NC}"
proot-distro login ubuntu -- bash -c '
    apt-get update -y
    apt-get install -y --no-install-recommends \
        curl git ca-certificates python3 python3-pip nodejs npm ripgrep fzf
'

echo -e "\n${GREEN}[3/4] Installing Antigravity CLI...${NC}"
proot-distro login ubuntu -- bash -c '
    mkdir -p ~/.local/bin
    # Check if agy installer script or binary is accessible
    if command -v agy >/dev/null 2>&1; then
        echo "agy already installed inside Ubuntu."
    else
        echo "Downloading & configuring agy..."
        # If user provides custom agy download or official curl URL:
        if [ -f /root/termux-home/.gemini/antigravity-cli/bin/agy ]; then
            cp /root/termux-home/.gemini/antigravity-cli/bin/agy ~/.local/bin/agy
            chmod +x ~/.local/bin/agy
        fi
    fi
'

echo -e "\n${GREEN}[4/4] Creating Termux launcher wrapper (~/.local/bin/agy)...${NC}"
mkdir -p "$HOME/.local/bin"

cat > "$HOME/.local/bin/agy" << 'EOF'
#!/data/data/com.termux/files/usr/bin/bash
# Wrapper to run Antigravity CLI seamlessly from Termux
CURRENT_DIR="$(pwd)"
WAKE_LOCK_ACTIVE=false

if command -v termux-wake-lock >/dev/null 2>&1; then
    termux-wake-lock
    WAKE_LOCK_ACTIVE=true
fi

# Run agy inside proot ubuntu mapping current directory
proot-distro login ubuntu --bind "$HOME:/root/termux-home" --bind "$CURRENT_DIR:/workspace" -- bash -c "
    export PATH=\"\$HOME/.local/bin:\$PATH\"
    export PAGER=cat
    export GIT_PAGER=cat
    cd /workspace
    if command -v agy >/dev/null 2>&1; then
        agy \"\$@\"
    elif [ -f /root/termux-home/.local/bin/agy ]; then
        /root/termux-home/.local/bin/agy \"\$@\"
    else
        echo 'Antigravity (agy) CLI binary not found in PATH.'
        echo 'Please place the agy binary into ~/.local/bin/ or configure your token.'
    fi
" -- "$@"

if [ "$WAKE_LOCK_ACTIVE" = true ] && command -v termux-wake-unlock >/dev/null 2>&1; then
    termux-wake-unlock
fi
EOF

chmod +x "$HOME/.local/bin/agy"

# Ensure ~/.local/bin is in PATH
if ! grep -q '\$HOME/\.local/bin' "$HOME/.bashrc" 2>/dev/null; then
    echo 'export PATH="$HOME/.local/bin:$PATH"' >> "$HOME/.bashrc"
fi

echo -e "\n${GREEN}==========================================${NC}"
echo -e "${GREEN}   Antigravity CLI setup finished! 🎉     ${NC}"
echo -e "${GREEN}==========================================${NC}\n"
echo -e "Usage:"
echo -e "  ${YELLOW}agy --help${NC}                       # Test CLI options"
echo -e "  ${YELLOW}tertic run agy${NC}                   # Run inside persistent tmux session"
echo -e "  ${YELLOW}agy --dangerously-skip-permissions${NC} # Autonomous mode for mobile"
echo ""
