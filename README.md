# Tertic 🚀

**Autonomous AI Coding Agent for Android & Termux**

Tertic turns your Android phone into an autonomous AI development machine. It runs directly inside Termux with native speed (<15ms boot, minimal RAM), zero-setup free AI models, live motion indicators, and persistent background execution.

---

## ⚡ Install on Termux (One-Line)

Open Termux and run:

```bash
curl -sSL https://raw.githubusercontent.com/Ashed2127/tertic/main/install.sh | bash
```

*Pre-configured with free OpenRouter AI models out-of-the-box. Zero API key setup required to get started.*

---

## 💻 Usage (`trtc`)

Just type **`trtc`** to launch:

### 1. Interactive Agent Mode
```bash
trtc
```
```text
  _____         _   _      
 |_   _|__ _ __| |_(_) ___ 
   | |/ _ \ '__| __| |/ __|
   | |  __/ |  | |_| | (__ 
   |_|\___|_|   \__|_|\___|
 Autonomous AI Coding Agent for Termux

trtc ❯ create a full-stack react + express app with sqlite
```

### 2. Direct Task Execution
```bash
trtc "build an express api with tests and git commit"
```

---

## ✨ Features

- **⚡ Blazing Fast**: Single native Go binary (`<15ms` startup, `<15MB` RAM).
- **🤖 Local & Cloud AI Options**:
  - **Local Offline (Ollama)**: 100% offline & unlimited running `qwen2.5-coder:0.5b` (397MB), `deepseek-coder:1.3b` (776MB), `qwen2.5-coder:1.5b` (986MB), `deepseek-r1:1.5b` (1.1GB), `gemma:2b` (1.7GB), or `qwen2.5-coder:3b` (1.9GB).
  - **Cloud APIs**: Free OpenRouter coding models or Google Gemini.
- **✨ Live Experience**: Real-time rotating spinners and motion emoji indicators (`🧠 Thinking...`, `💻 Running bash`, `📝 Writing file`).
- **🛡️ Android Survival**: Auto wake-lock prevents Android from sleeping the CPU during long coding tasks.
- **🔄 Arrow Key Menu**: Interactive setup menu with ↑/↓ arrow keys to pick models or switch providers.

---

## 🛠️ CLI Reference

| Command | Description |
|---|---|
| `trtc` | Start interactive AI agent |
| `trtc "<task>"` | Execute a coding task directly |
| `trtc doctor` | Hardware diagnostics & memory inspection |
| `trtc optimize` | Auto-tune memory limits & mobile keyboard |
| `trtc notify "<msg>"` | Send native Android vibration / push notification |

---

## 📜 License
MIT © [Ashed](https://github.com/Ashed2127)
