# Tertic - Optimized Termux for AI Agents

**Make AI agents accessible to students on Android phones!**

Tertic is a pre-configured, optimized Termux setup that allows students to run AI agents directly on their Android devices without needing a laptop or computer.

## Why Tertic?

- 📱 **No laptop needed** - Run AI agents on your Android phone
- 🚀 **Optimized for mobile** - Lightweight, resource-efficient configuration
- 🎓 **Student-friendly** - Easy setup with step-by-step instructions
- 💻 **Full terminal experience** - Complete command-line AI agent capabilities
- 🆓 **Free & open source** - Built on Termux, no hidden costs

## Features

- Pre-configured Python environment optimized for ARM Android
- Lightweight package manager (uv) for fast installations
- Optimized memory and CPU settings for mobile devices
- Support for popular AI agent frameworks
- Battery-saving optimizations
- Quick-start scripts for common AI agent setups

## Quick Start

### Prerequisites

- Android 7.0 or higher
- At least 2GB RAM (4GB+ recommended)
- 1GB free storage space

### Installation

1. **Download Termux**
   - Go to [Termux GitHub Releases](https://github.com/termux/termux-app/releases)
   - Download the latest version (v0.118.3 or newer)
   - For Android 7+: Download `apt-android-7` variant
   - For Android 5-6: Download `apt-android-5` variant
   - Install the APK

2. **Clone Tertic**
   ```bash
   # In Termux
   pkg update && pkg upgrade
   pkg install git
   git clone https://github.com/yourusername/tertic.git
   cd tertic
   ```

3. **Run Setup**
   ```bash
   bash setup.sh
   ```

4. **Start Coding!**
   ```bash
   # Example: Run a simple AI agent
   python agent.py
   ```

## Supported AI Agents

- OpenAI Agents
- LangChain Agents
- Simple AI Agent
- And more...

## Performance Tips

- Close other apps while running AI agents
- Use a charger for intensive tasks
- Enable high-performance mode in developer options
- Use lightweight models when possible

## Contributing

We welcome contributions! Please see CONTRIBUTING.md for details.

## License

MIT License - See LICENSE file for details

## Support

- 📖 Documentation: [docs/](docs/)
- 💬 Community: [Discord/GitHub Discussions]
- 🐛 Issues: [GitHub Issues](https://github.com/yourusername/tertic/issues)

---

**Made with ❤️ for students everywhere**
