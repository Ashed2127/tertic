# Tertic

Autonomous multi-agent AI coding system running natively inside Termux on Android. Orchestrates specialized autonomous agents (Coordinator, Planner, Researcher, Developer, Reviewer) to build, test, and deliver software with native speed and memory-safe sequential execution.

---

## What It Is

Tertic turns your Android device into a full-fledged autonomous software development team. It is optimized for mobile hardware (including 4 GB RAM devices), boots in `<15ms`, and supports both free cloud models (OpenRouter, Google Gemini) and 100% offline local models (Ollama).

Tertic features a modern terminal interface with **Text, Voice, and Hybrid input modes**, human-readable tool execution, adaptive theme support, and seamless Termux Android speech recognition.

---

## Setup (One-Line)

Run in Termux:

```bash
curl -sSL https://raw.githubusercontent.com/Ashed2127/tertic/main/install.sh | bash
```

*Pre-configured out-of-the-box with free AI models. Zero API key required to get started.*

---

## Terminal Experience & Modes

### 1. Three Input Modes
- **Text Mode** (Default): Fast, keyboard-based terminal interaction with multiline (`"""` or trailing `\`), slash suggestions, and command history.
- **Voice Mode**: Speak your instructions instead of typing. Automatically integrates with Termux:API (`termux-speech-to-text`) or local whisper engines. Review and edit transcription before execution.
- **Hybrid Mode**: Alternate seamlessly between voice and keyboard input within the same session.

### 2. Modern UI & Theme System
- **Minimalist Aesthetic**: Inspired by modern developer tools (Devin, Claude Code, Gemini CLI).
- **Themes**: Switch between `dark` (default cyan/purple), `light`, `monochrome` (plain ASCII, respects `NO_COLOR`), and `high-contrast`.
- **Human-Readable Tool Activity**: Real-time action indicators (`Reading: main.go`, `Writing: app.js`, `Testing & verifying...`) instead of raw JSON dumps.
- **Markdown & Streaming Output**: Renders syntax code blocks, headings, bold text, and lists with smooth output.

---

## Multi-Agent Architecture

Tertic coordinates five specialized autonomous agents through a shared, resource-aware harness:

1. **Coordinator Agent**: Evaluates user objectives, routes work, integrates outputs, and performs final verification.
2. **Planner Agent**: Decomposes complex goals into atomic structured tasks with dependencies and explicit acceptance criteria.
3. **Researcher Agent**: Inspects codebase architecture and reads relevant files before implementation.
4. **Developer Agent**: Writes code, modifies files, and executes build/test commands via validated tools.
5. **Reviewer Agent**: Audits generated code against acceptance criteria, verifies tests, and sends failed tasks back to the developer with actionable feedback.

---

## Commands & Usage (`trtc`)

### 1. Launching Tasks

```bash
trtc                     # Interactive premium multi-agent session
trtc "build a react app" # Autonomous multi-agent execution
trtc --single-agent "task" # Single-agent direct execution
```

### 2. Multi-Agent CLI Commands

| Command | Action |
|---|---|
| `trtc agents` | List all available specialized agents and roles |
| `trtc agents status` | Inspect agent mesh status and allowed tool permissions |
| `trtc models` | List configured model providers assigned to each agent |
| `trtc config` | Display active runtime and memory constraints |

### 3. Inside Interactive Session

| Command | Function |
|---|---|
| `/help` | Show available interactive commands |
| `/model` | Switch between Cloud and Local Offline models |
| `/models` | List configured models |
| `/agents` | Show available specialized agents & roles |
| `/status` | Show session & speech engine status |
| `/clear` | Clear terminal screen and redraw welcome view |
| `/history` | Show previous interactions & conversation memory |
| `/context` | Show project directory context |
| `/config` | Display active model configuration |
| `/verbose` | Toggle detailed execution output & previews |
| `/theme` | Change terminal color theme (`dark`, `light`, `monochrome`, `high-contrast`) |
| `/voice` | Activate voice input mode |
| `/text` | Switch to keyboard text input mode |
| `/input` | Configure input mode (Text / Voice / Hybrid) |
| `/listen` | Start voice recording immediately |
| `/stop` | Stop recording and generate transcription |
| `/exit` | Exit Tertic |

### 4. System Utilities

| Command | Description |
|---|---|
| `tertic doctor` | Inspect CPU, RAM profile, and developer tool readiness |
| `tertic optimize` | Apply Android wake-lock, memory caps, and mobile keys |
| `tertic notify "<msg>"` | Send Android push alert/vibration when long tasks finish |

---

## Voice Input Setup for Termux

To enable voice input on Android:
1. Install Termux:API in your terminal:
   ```bash
   pkg install termux-api
   ```
2. Install the **Termux:API** APK companion app (from F-Droid).
3. Grant **Microphone** permission to Termux:API in Android settings.
4. Launch `trtc` and type `/voice` or `/listen`.

---

## Resource Optimization for Android

- **Sequential Execution**: Models run sequentially by default to prevent OOM on devices with 4 GB RAM.
- **Model Reuse**: Agents share existing loaded models without unnecessary duplicate instances.
- **Tool Validation**: Strict security checks reject empty paths, malformed JSON, and destructive root commands.
- **Dependency Graphs**: Structured task queue automatically resolves task prerequisite order.

---

## License

MIT © [Ashed](https://github.com/Ashed2127)
