# Tertic

**The Autonomous AI Agent Environment for Android & Termux 🚀**

Tertic turns any Android phone into an autonomous AI development machine. It provides automated device profiling, glibc compatibility via PRoot, background crash protection, and optimized tooling for AI coding agents (Devin, Antigravity `agy`, Aider, Claude Code, and more).

## Features

- **Universal Device Optimization:** Auto-detects device RAM and CPU cores to dynamically tune Node/Python memory limits, swap space, and build concurrency for low, mid, and high-end phones.
- **Background Survival:** Keeps agents running safely with Android wake-lock and detached `tmux` sessions so tasks don't terminate when the screen turns off.
- **Agent Ecosystem:** Native runners and glibc bridges for **Devin CLI** and **Antigravity CLI (`agy`)**.
- **Mobile Hardware Bridge:** Android push notifications and vibration alerts via `termux-api` when agents complete tasks.

## Quick Start

1. **In Termux, install Tertic**:
   ```bash
   pkg update && pkg upgrade -y
   pkg install -y git curl
   git clone https://github.com/Ashed2127/tertic.git
   cd tertic
   bash setup.sh
   ```

2. **Run Diagnostics & Auto-Optimization**:
   ```bash
   tertic doctor     # Inspect device profile, RAM tier, and missing tools
   tertic optimize   # Auto-apply memory, swap, keyboard, and wake-lock tweaks
   ```

3. **Install Your AI Agent of Choice**:
   - For **Devin CLI**:
     ```bash
     bash install-devin-linux.sh
     ```
   - For **Antigravity CLI (`agy`)**:
     ```bash
     bash install-antigravity.sh
     ```

4. **Launch Agent in Crash-Proof Background Session**:
   ```bash
   tertic run devin
   # or
   tertic run agy
   ```

## CLI Reference

| Command | Description |
|---|---|
| `tertic doctor` | Analyzes hardware, memory tiers, and agent dependencies |
| `tertic optimize` | Auto-tunes swap, heap limits, and mobile keyboard shortcuts |
| `tertic run <agent>` | Runs the agent in a persistent `tmux` session with wake-lock |
| `tertic notify <title> <msg>` | Sends a native Android notification when tasks finish |
- First setup takes ~10-15 minutes (Ubuntu download)

Check `docs/` for more details.
