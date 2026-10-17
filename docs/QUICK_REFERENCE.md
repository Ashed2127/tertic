# Tertic Quick Reference Card

Print this or save it for quick reference!

## Essential Commands

### Environment
```bash
source ~/tertic-env/bin/activate    # Activate Tertic environment
deactivate                          # Deactivate environment
python --version                    # Check Python version
pip list                            # List installed packages
```

### Running Agents
```bash
python ~/tertic/examples/simple_agent.py       # Simple agent
python ~/tertic/examples/langchain_agent.py    # LangChain agent
```

### API Keys
```bash
export OPENAI_API_KEY='your-key'              # Set OpenAI key
export ANTHROPIC_API_KEY='your-key'           # Set Anthropic key
```

### Termux Basics
```bash
pkg update && pkg upgrade                      # Update Termux
termux-setup-storage                           # Enable storage access
pkg install <package>                          # Install package
```

### File Operations
```bash
ls                                           # List files
cd <directory>                               # Change directory
pwd                                          # Current directory
mkdir <name>                                 # Create directory
nano <file>                                  # Edit file
cat <file>                                   # View file
```

## File Locations

- Tertic: `~/tertic/`
- Environment: `~/tertic-env/`
- Config: `~/.termux/termux.properties`
- API keys: `~/.tertic-config` or `~/.env`

## Keyboard Shortcuts (Termux)

- `Ctrl + C` - Cancel current command
- `Ctrl + D` - Exit Python/shell
- `Ctrl + L` - Clear screen
- `Ctrl + A` - Start of line
- `Ctrl + E` - End of line
- `Ctrl + U` - Delete to start of line
- `Ctrl + K` - Delete to end of line

## Common Issues

**"Command not found"**
```bash
source ~/tertic-env/bin/activate
```

**"Permission denied"**
```bash
chmod +x script.sh
```

**Out of memory**
```bash
# Close other apps
# Reduce max_tokens in your code
```

**Storage permission**
```bash
termux-setup-storage
```

## Python Quick Start

```python
# Hello World
print("Hello Tertic!")

# Input
name = input("Your name: ")
print(f"Hello {name}!")

# Loop
for i in range(5):
    print(i)

# Function
def greet(name):
    return f"Hello {name}!"
```

## Model Recommendations

| Model | Speed | Cost | Best For |
|-------|-------|------|----------|
| gpt-3.5-turbo | Fast | Low | General use |
| gpt-4 | Slow | High | Complex tasks |
| claude-3-haiku | Fast | Low | Quick responses |
| claude-3-opus | Slow | High | Advanced tasks |

## Performance Tips

1. Use Wi-Fi instead of mobile data
2. Close other apps
3. Keep phone charging
4. Use gpt-3.5-turbo instead of gpt-4
5. Set max_tokens to 500 or less
6. Clear caches weekly

## Getting Help

- Check `docs/STUDENT_GUIDE.md`
- Check `docs/PERFORMANCE_TIPS.md`
- Search online: "Termux Python [your question]"
- GitHub Issues: [tertic/issues]

## Resources

- Termux Wiki: https://wiki.termux.com/
- Python Docs: https://docs.python.org/3/
- OpenAI Docs: https://platform.openai.com/docs
- LangChain Docs: https://python.langchain.com/

---

**Keep coding! 💪**
