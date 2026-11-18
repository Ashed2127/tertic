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

It will install all dependencies, build the high-speed agent, and let you choose your AI provider:
- **Option 1: OpenRouter (Recommended)**: 100% free coding models with no strict quotas (`meta-llama/llama-3.3-70b-instruct:free`, `qwen/qwen-2.5-coder-32b-instruct:free`, `deepseek/deepseek-r1:free`).
- **Option 2: Google Gemini Studio**: Direct Google Gemini API key.

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
 Autonomous AI Coding Agent for Termux (OpenRouter & Gemini)

Provider: openrouter | Model: meta-llama/llama-3.3-70b-instruct:free
Interactive Mode (type your instructions or 'exit' to quit)

trtc > create a fullstack react + express app and run tests
```

### 2. Direct Task Execution
```bash
trtc "Build a full-stack Node.js API with SQLite and user auth"
```

---

## Switching Models Anytime

You can switch models or providers anytime using environment variables or in `~/.tertic/config.env`:

```bash
# To switch to Qwen Coder (Free on OpenRouter):
export OPENROUTER_MODEL="qwen/qwen-2.5-coder-32b-instruct:free"

# To switch to DeepSeek R1 (Free on OpenRouter):
export OPENROUTER_MODEL="deepseek/deepseek-r1:free"
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
