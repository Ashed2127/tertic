# Devin CLI Setup Guide for Android

This guide helps you set up Devin AI CLI on your Android phone using Tertic.

## What is Devin CLI?

Devin CLI is a local AI coding agent that runs in your terminal. It can:
- Write and fix code
- Review your codebase
- Answer questions about your code
- Automate development tasks
- Run in your local environment

## Prerequisites

- Android 7.0 or higher
- Termux installed (see README)
- A Devin account with CLI access
- Internet connection

## Installation Steps

### 1. Install Tertic

```bash
pkg update && pkg upgrade
pkg install git curl
git clone https://github.com/Ashed2127/tertic.git
cd tertic
bash setup.sh
```

### 2. Install Devin CLI

```bash
bash install-devin.sh
```

This downloads the ARM64 Linux binary and installs it to `~/.local/bin/devin`.

### 3. Restart Termux

Swipe Termux away from recent apps and open it again.

### 4. Login to Devin

```bash
devin auth login
```

This will open a browser where you can authenticate with your Devin account.

### 5. Test Devin

```bash
devin --version
```

You should see the version number.

## Using Devin CLI

### Basic Usage

Navigate to any project directory and run:

```bash
cd your-project
devin
```

### Example Workflow

1. **Create a test project**:
   ```bash
   python ~/tertic/examples/devin_workflow.py
   cd devin-test-project
   ```

2. **Start Devin**:
   ```bash
   devin
   ```

3. **Ask Devin to help**:
   ```
   Add a function to calculate the sum of two numbers
   ```

4. **Devin will**:
   - Read your code
   - Write the new function
   - Show you the changes
   - Ask for confirmation

### Common Devin Commands

Inside the Devin CLI session:

- Type your request in natural language
- Devin will show you what it plans to do
- Press Enter to confirm or type to modify
- Use `/help` for slash commands
- Use `/exit` to quit

## Tips for Android

### Performance

- Close other apps while using Devin
- Use Wi-Fi instead of mobile data
- Keep your phone charging for long sessions
- Use lighter models when possible

### Terminal Tips

- Use `Ctrl+C` to cancel Devin's current action
- Use `Ctrl+D` to exit Devin
- Termux keyboard shortcuts work as expected
- The extra keys bar helps with common keys

### Common Issues

**"devin: command not found"**
```bash
export PATH="$HOME/.local/bin:$PATH"
```

**"Authentication failed"**
```bash
devin auth login
```

**Devin is slow**
- Check your internet connection
- Close other apps
- Reduce task complexity

## Example Projects to Try

### 1. Fix a Bug

```bash
cd your-project
devin
# Ask: "Find and fix any bugs in this code"
```

### 2. Add a Feature

```bash
cd your-project
devin
# Ask: "Add error handling to the main function"
```

### 3. Review Code

```bash
cd your-project
devin
# Ask: "Review this code and suggest improvements"
```

### 4. Learn Codebase

```bash
cd your-project
devin
# Ask: "Explain how this code works"
```

## Advanced Usage

### Configuration

Devin CLI can be configured with a config file. Create `~/.config/devin/config.yaml`:

```yaml
model: claude-3-5-sonnet
legacy_terminal: false
```

### MCP Servers

Connect to MCP servers for additional capabilities:

```bash
devin mcp add <server-name>
```

### Skills

Use custom skills for specific workflows:

```bash
devin skill invoke <skill-name>
```

## Troubleshooting

### Binary Download Fails

If the download fails, try manually:

```bash
curl -fsSL https://static.devin.ai/cli/aarch64-unknown-linux/devin -o ~/.local/bin/devin
chmod +x ~/.local/bin/devin
```

### Permission Denied

```bash
chmod +x ~/.local/bin/devin
```

### Login Issues

Make sure you have a Devin account with CLI access enabled. Check your account settings at https://devin.ai

## Resources

- Devin CLI Docs: https://docs.devin.ai/cli
- Devin Website: https://devin.ai
- Community: https://github.com/cognition-ai/devin

---

**Happy coding with Devin on Android! 🚀**
