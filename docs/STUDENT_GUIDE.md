# Tertic Student Guide

## Welcome! 🎓

This guide will help you set up and use AI agents on your Android phone using Tertic. No laptop needed!

## What You'll Need

- An Android phone (Android 7.0 or higher)
- At least 2GB RAM (4GB recommended)
- 1GB free storage space
- Internet connection

## Step 1: Install Termux

### Download Termux

1. Open your phone's browser
2. Go to: https://github.com/termux/termux-app/releases
3. Download the latest version:
   - **Android 7+**: Download the file ending in `apt-android-7.apk`
   - **Android 5-6**: Download the file ending in `apt-android-5.apk`
4. Install the APK (you may need to enable "Install from unknown sources" in settings)

### Open Termux

- Tap the Termux icon to open it
- You'll see a terminal with a `$` prompt

## Step 2: Install Tertic

### Option A: Quick Install (Recommended)

```bash
# Update Termux first
pkg update && pkg upgrade

# Install git
pkg install git

# Clone Tertic
git clone https://github.com/yourusername/tertic.git
cd tertic

# Run setup
bash setup.sh
```

### Option B: Manual Install

If the quick install doesn't work, follow these steps:

```bash
# 1. Update packages
pkg update && pkg upgrade

# 2. Install basic tools
pkg install python git curl wget

# 3. Install UV (fast package manager)
curl -LsSf https://astral.sh/uv/install.sh | sh

# 4. Create Python environment
mkdir -p ~/tertic-env
~/.local/bin/uv venv ~/tertic-env
source ~/tertic-env/bin/activate

# 5. Install dependencies
uv pip install aiohttp aiosqlite pyyaml rich openai
```

## Step 3: Configure Termux (Important!)

### Enable Storage Access

```bash
termux-setup-storage
```
Tap "Allow" when prompted.

### Configure Termux Settings

```bash
# Create config directory
mkdir -p ~/.termux

# Copy optimized config
cp ~/tertic/termux.properties ~/.termux/

# Restart Termux for changes to take effect
```

### Restart Termux

- Swipe Termux away from recent apps
- Open it again

## Step 4: Run Your First AI Agent!

### Activate the Environment

```bash
source ~/tertic-env/bin/activate
```

### Try the Local Agent (No API Key Needed)

```bash
python ~/tertic/examples/simple_agent.py
```
Select option 2 for the local agent.

### Try with OpenAI (Optional)

If you have an OpenAI API key:

```bash
# Set your API key
export OPENAI_API_KEY='your-api-key-here'

# Run the agent
python ~/tertic/examples/simple_agent.py
```
Select option 1 for the OpenAI agent.

## Tips for Better Performance

### Save Battery 🔋

- Close other apps while running agents
- Lower screen brightness
- Keep your phone charging during long tasks
- Use lighter AI models when possible

### Make It Faster ⚡

- Enable high-performance mode:
  - Go to Settings → Developer Options
  - Enable "High performance mode"
- Clear Termux cache occasionally:
  ```bash
  pkg clean
  ```

### Common Issues

**Issue**: "Command not found"
- **Solution**: Make sure you activated the environment:
  ```bash
  source ~/tertic-env/bin/activate
  ```

**Issue**: Installation fails
- **Solution**: Update Termux and try again:
  ```bash
  pkg update && pkg upgrade
  ```

**Issue**: Out of memory
- **Solution**: Close other apps, or use a lighter AI model

## Next Steps

### Learn Python

```bash
# Start Python
python

# Try this:
print("Hello from Tertic!")
for i in range(5):
    print(f"Count: {i}")
```

### Create Your Own Agent

Create a new file:

```bash
nano my_agent.py
```

Add this code:

```python
from rich.console import Console
console = Console()
console.print("[bold green]My Custom Agent![/bold green]")
console.input("What's your name? ")
```

Save and run:
```bash
python my_agent.py
```

## Need Help?

- Check the main README.md
- Look at more examples in the `examples/` folder
- Search online for "Termux Python tutorial"

## You Did It! 🎉

You now have AI agents running on your Android phone! Keep learning and building.

---

**Remember**: Every expert was once a beginner. Keep coding! 💪
