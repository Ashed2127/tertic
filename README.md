# Tertic

**Run Devin AI CLI on your Android phone - yes, it's possible! 🚀**

Tertic sets up Ubuntu inside Termux using proot-distro, providing a full glibc Linux environment where Devin CLI runs perfectly.

## How It Works

- **Termux** provides the Android terminal
- **proot-distro** runs Ubuntu inside Termux
- **Ubuntu** has glibc (like regular Linux)
- **Devin CLI** runs in that Ubuntu environment

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

3. **Install Devin CLI**:
   ```bash
   bash install-devin-linux.sh
   ```
   This installs Ubuntu (~500MB) and Devin CLI inside it.

4. **Run Devin**:
   ```bash
   cd your-project
   devin
   ```

## What You Get

- **Devin CLI** - Full AI coding agent on your phone!
- **Ubuntu environment** - Full Linux compatibility
- **Development tools** - Python, Node.js, git, etc.
- **Optimized Termux** - Mobile-friendly configuration

## Performance Tips

- Use Wi-Fi instead of mobile data
- Close other apps while using Devin
- Keep your phone charging for long sessions
- First setup takes ~10-15 minutes (Ubuntu download)

Check `docs/` for more details.
