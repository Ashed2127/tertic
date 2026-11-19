# Tertic

Autonomous AI coding agent running directly inside Termux on Android. Executes tasks end-to-end without clutter—creating files, running commands, and debugging code with native speed.

---

## What It Is

Tertic turns your Android phone into an autonomous development environment (like Devin or Antigravity CLI). It boots in `<15ms`, uses minimal memory, and supports both free cloud models (OpenRouter, Gemini) and 100% offline local models (Ollama).

---

## Setup (One-Line)

Run in Termux:

```bash
curl -sSL https://raw.githubusercontent.com/Ashed2127/tertic/main/install.sh | bash
```

*Ready to use out-of-the-box with free cloud AI models. No API key required.*

---

## Features & Commands

### 1. Agent Commands (`trtc`)

Launch by typing `trtc`:

```bash
trtc                  # Interactive agent prompt
trtc "build a todo app" # Run a task directly
```

### 2. Inside the Agent Session

| Command | Action |
|---|---|
| `/model` | Open arrow-key menu to switch between local Ollama & cloud models |
| `/process` | Toggle step-by-step tool output ON or OFF (clean mode by default) |
| `exit` | Quit session |

### 3. Utility Commands

| Command | Description |
|---|---|
| `tertic doctor` | Inspect CPU, RAM, and developer tool readiness |
| `tertic optimize` | Apply Android wake-lock, memory caps, and mobile keys |
| `tertic notify "<msg>"` | Send Android vibration / alert when a task finishes |

---

## License

MIT © [Ashed](https://github.com/Ashed2127)
