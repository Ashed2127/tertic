# Tertic

**Optimized Termux setup for coding on Android phones.**

⚠️ **Important**: Devin CLI's official binary is compiled for Linux (glibc) and won't run on Android (bionic). This project provides an optimized Termux environment for development, but Devin CLI itself is not currently supported on Android.

## What Tertic Provides

- Optimized Termux configuration for coding
- Development tools (git, nodejs, python, tmux, etc.)
- Terminal utilities (ripgrep, fzf, neovim)
- Mobile-friendly settings and keyboard shortcuts

## Setup

1. **Download Termux**: https://github.com/termux/termux-app/releases (get the `apt-android-7` APK for Android 7+)

2. **In Termux, run**:
   ```bash
   pkg update && pkg upgrade
   pkg install git curl
   git clone https://github.com/Ashed2127/tertic.git
   cd tertic
   bash setup.sh
   ```

3. **Optional: Advanced tools**:
   ```bash
   bash advanced-setup.sh
   ```

## Available Tools

After setup, you'll have:
- **Languages**: Python, Node.js
- **Editors**: Nano, Neovim
- **Tools**: Git, curl, wget, tar
- **Advanced (optional)**: tmux, ripgrep, fzf, htop, black, prettier

## Why Devin CLI Doesn't Work on Android

Devin CLI is compiled for Linux with glibc. Android uses bionic, a different C library. While both use ARM64, the binary won't run due to library incompatibilities.

**Possible future solutions:**
- Devin team releases an Android-specific build
- Someone compiles Devin CLI from source for Android
- Use a Linux container on Android (UserLAnd, Proot)

## Alternatives

While waiting for Android support, you can:
- Use Tertic for local development with standard tools
- Access cloud-based Devin from your phone's browser
- Use other AI coding assistants that support Android

Check `docs/` for more details.
