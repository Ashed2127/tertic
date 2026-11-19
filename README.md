# Tertic

Autonomous multi-agent AI coding system running directly inside Termux on Android. Orchestrates specialized autonomous agents (Coordinator, Planner, Researcher, Developer, Reviewer) to build, test, and deliver software with native speed and memory-safe sequential execution.

---

## What It Is

Tertic turns your Android device into a full-fledged autonomous software development team. It is optimized for mobile hardware (including 4 GB RAM devices), boots in `<15ms`, and supports both free cloud models (OpenRouter, Google Gemini) and 100% offline local models (Ollama).

---

## Setup (One-Line)

Run in Termux:

```bash
curl -sSL https://raw.githubusercontent.com/Ashed2127/tertic/main/install.sh | bash
```

*Pre-configured out-of-the-box with free AI models. Zero API key required to get started.*

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
trtc                     # Interactive multi-agent session
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

| Command | Action |
|---|---|
| `/agents` | View specialized agent roster |
| `/model` | Open arrow-key menu to switch between local Ollama & cloud models |
| `/process` | Toggle step-by-step tool output ON or OFF (clean mode by default) |
| `exit` | Quit session |

### 4. System Utilities

| Command | Description |
|---|---|
| `tertic doctor` | Inspect CPU, RAM profile, and developer tool readiness |
| `tertic optimize` | Apply Android wake-lock, memory caps, and mobile keys |
| `tertic notify "<msg>"` | Send Android push alert/vibration when long tasks finish |

---

## Resource Optimization for Android

- **Sequential Execution**: Models run sequentially by default to prevent OOM on devices with 4 GB RAM.
- **Model Reuse**: Agents share existing loaded models without unnecessary duplicate instances.
- **Tool Validation**: Strict security checks reject empty paths, malformed JSON, and destructive root commands.
- **Dependency Graphs**: Structured task queue automatically resolves task prerequisite order.

---

## License

MIT © [Ashed](https://github.com/Ashed2127)
