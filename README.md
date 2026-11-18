# Tertic

**The Autonomous AI Agent Environment for Android & Termux 🚀**

Tertic turns any Android phone into an autonomous AI development machine. It provides automated device profiling, glibc compatibility via PRoot, background crash protection, and optimized tooling for AI coding agents (Devin, Antigravity `agy`, Aider, Claude Code, and more).

## Features

- **Universal Device Optimization:** Auto-detects device RAM and CPU cores to dynamically tune Node/Python memory limits, swap space, and build concurrency for low, mid, and high-end phones.
- **Background Survival:** Keeps agents running safely with Android wake-lock and detached `tmux` sessions so tasks don't terminate when the screen turns off.
- **Agent Ecosystem:** Native runners and glibc bridges for **Devin CLI** and **Antigravity CLI (`agy`)**.
- **Mobile Hardware Bridge:** Android push notifications and vibration alerts via `termux-api` when agents complete tasks.

## Installation on Termux ⚡

Just run this one-line installer in Termux:

```bash
curl -sSL https://raw.githubusercontent.com/Ashed2127/tertic/main/install.sh | bash
```

It will install all dependencies, build the high-speed agent, and prompt you to paste your Google AI (Gemini) API Key.

---

## Usage (`trtc`)

Just like `devin` or `agy`, simply type **`trtc`**:

### 1. Interactive Agent Mode
```bash
trtc
```
Starts the full interactive session with live tool calling:
```
  _____         _   _      
 |_   _|__ _ __| |_(_) ___ 
   | |/ _ \ '__| __| |/ __|
   | |  __/ |  | |_| | (__ 
   |_|\___|_|   \__|_|\___|
 Autonomous AI Coding Agent for Termux (Gemini Flash)

Interactive Agent Mode (type your instructions or 'exit' to quit)

trtc > create a react vite app and test it with a python backend
```

### 2. Direct Task Execution
```bash
trtc "Build a full-stack Node.js API with SQLite and user auth"
```

---

## Environment Optimization & Diagnostics

| Command | Description |
|---|---|
| `trtc` | Starts the native autonomous AI coding agent |
| `trtc "task"` | Directly executes a coding task |
| `trtc doctor` | Hardware diagnostics & memory profile inspection |
| `trtc optimize` | Auto-tunes RAM heap caps, swap, and mobile keyboard |
| `trtc run <cmd>` | Runs external heavy agents in crash-proof tmux daemon |
- First setup takes ~10-15 minutes (Ubuntu download)

Check `docs/` for more details.
